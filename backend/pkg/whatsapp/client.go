package whatsapp

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

// Client wrapper do whatsmeow.Client
type Client struct {
	wa       *whatsmeow.Client
	device   *store.Device
	handler  EventHandler
	status   Status
	qrCode   string
	killChan chan bool
	mu       sync.RWMutex
}

// NewClient cria novo cliente WhatsApp
func NewClient(device *store.Device) *Client {
	// Configurar propriedades do dispositivo
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()
	store.DeviceProps.Os = proto.String("Zyntra")

	wa := whatsmeow.NewClient(device, nil)

	client := &Client{
		wa:       wa,
		device:   device,
		status:   StatusDisconnected,
		killChan: make(chan bool, 1),
	}

	wa.AddEventHandler(client.handleEvent)

	return client
}

// SetEventHandler define o handler de eventos
func (c *Client) SetEventHandler(handler EventHandler) {
	c.handler = handler
}

// Connect inicia a conexao com WhatsApp
// Retorna canal de eventos QR se precisar de autenticacao
func (c *Client) Connect(ctx context.Context) (<-chan QREvent, error) {
	c.mu.Lock()
	c.status = StatusConnecting
	c.mu.Unlock()

	// Se ja tem sessao, conectar direto
	if c.wa.Store.ID != nil {
		go c.connectExisting()
		return nil, nil
	}

	// Novo login - precisa de QR code
	qrChan := make(chan QREvent, 10)
	go c.handleQRFlow(qrChan)
	return qrChan, nil
}

func (c *Client) handleQRFlow(qrChan chan<- QREvent) {
	defer close(qrChan)

	waQRChan, err := c.wa.GetQRChannel(context.Background())
	if err != nil {
		if err == whatsmeow.ErrQRStoreContainsID {
			log.Printf("[WhatsApp] Already logged in, connecting...")
			c.connectExisting()
			return
		}
		log.Printf("[WhatsApp] Failed to get QR channel: %v", err)
		return
	}

	if err := c.wa.Connect(); err != nil {
		log.Printf("[WhatsApp] Failed to connect: %v", err)
		return
	}

	for evt := range waQRChan {
		switch evt.Event {
		case "code":
			image, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			if err != nil {
				log.Printf("[WhatsApp] Failed to generate QR image: %v", err)
				continue
			}
			base64QR := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)

			c.mu.Lock()
			c.qrCode = base64QR
			c.status = StatusQRCode
			c.mu.Unlock()

			qrChan <- QREvent{
				Event:  "code",
				Code:   evt.Code,
				Base64: base64QR,
			}

			if c.handler != nil {
				c.handler.OnQRCode(QREvent{Event: "code", Code: evt.Code, Base64: base64QR})
			}

		case "timeout":
			log.Printf("[WhatsApp] QR timeout")
			c.mu.Lock()
			c.status = StatusDisconnected
			c.qrCode = ""
			c.mu.Unlock()

			qrChan <- QREvent{Event: "timeout"}

			if c.handler != nil {
				c.handler.OnDisconnected()
			}
			c.wa.Disconnect()
			return

		case "success":
			log.Printf("[WhatsApp] QR pairing successful")
			c.mu.Lock()
			c.status = StatusConnected
			c.qrCode = ""
			c.mu.Unlock()

			qrChan <- QREvent{Event: "success"}
		}
	}

	c.keepAliveLoop()
}

func (c *Client) connectExisting() {
	log.Printf("[WhatsApp] Connecting existing session...")

	if err := c.wa.Connect(); err != nil {
		log.Printf("[WhatsApp] Failed to connect: %v", err)
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()
		return
	}

	c.keepAliveLoop()
}

func (c *Client) keepAliveLoop() {
	for {
		select {
		case <-c.killChan:
			log.Printf("[WhatsApp] Kill signal received")
			c.wa.Disconnect()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// Disconnect desconecta do WhatsApp
func (c *Client) Disconnect() error {
	select {
	case c.killChan <- true:
	default:
	}

	c.mu.Lock()
	c.status = StatusDisconnected
	c.qrCode = ""
	c.mu.Unlock()

	return nil
}

// Logout faz logout e remove sessao
func (c *Client) Logout(ctx context.Context) error {
	return c.wa.Logout(ctx)
}

// Status retorna status atual
func (c *Client) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.wa.IsConnected() && c.wa.IsLoggedIn() {
		return StatusConnected
	}
	return c.status
}

// IsConnected verifica se esta conectado
func (c *Client) IsConnected() bool {
	return c.wa.IsConnected() && c.wa.IsLoggedIn()
}

// GetQRCode retorna QR code atual
func (c *Client) GetQRCode() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.qrCode
}

// GetJID retorna JID do dispositivo
func (c *Client) GetJID() string {
	if c.wa.Store.ID != nil {
		return c.wa.Store.ID.String()
	}
	return ""
}

// GetPhone retorna numero de telefone
func (c *Client) GetPhone() string {
	if c.wa.Store.ID != nil {
		return JIDToPhone(*c.wa.Store.ID)
	}
	return ""
}

// SendText envia mensagem de texto
func (c *Client) SendText(ctx context.Context, to, content string) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("client not connected")
	}

	jid := PhoneToJID(to)
	msg := &waProto.Message{
		Conversation: proto.String(content),
	}

	resp, err := c.wa.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return resp.ID, nil
}

// SendImage envia imagem
func (c *Client) SendImage(ctx context.Context, to string, data []byte, caption, mimeType string) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("client not connected")
	}

	jid := PhoneToJID(to)

	uploaded, err := c.wa.Upload(ctx, data, whatsmeow.MediaImage)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String(mimeType),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}

	resp, err := c.wa.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send image: %w", err)
	}

	return resp.ID, nil
}

// SendDocument envia documento
func (c *Client) SendDocument(ctx context.Context, to string, data []byte, filename, caption, mimeType string) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("client not connected")
	}

	jid := PhoneToJID(to)

	uploaded, err := c.wa.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}

	msg := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String(mimeType),
			FileName:      proto.String(filename),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}

	resp, err := c.wa.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send document: %w", err)
	}

	return resp.ID, nil
}

// GetContactName busca nome do contato
func (c *Client) GetContactName(jid types.JID) string {
	ctx := context.Background()
	contactInfo, err := c.wa.Store.Contacts.GetContact(ctx, jid)
	if err == nil {
		if contactInfo.FullName != "" {
			return contactInfo.FullName
		}
		if contactInfo.BusinessName != "" {
			return contactInfo.BusinessName
		}
		if contactInfo.PushName != "" {
			return contactInfo.PushName
		}
	}
	if jid.User != "" {
		return "+" + jid.User
	}
	return jid.String()
}

func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		log.Printf("[WhatsApp] Connected")
		c.mu.Lock()
		c.status = StatusConnected
		c.qrCode = ""
		c.mu.Unlock()

		if c.handler != nil {
			c.handler.OnConnected(c.GetPhone(), c.GetJID())
		}

	case *events.PairSuccess:
		log.Printf("[WhatsApp] PairSuccess: JID=%s", v.ID.String())
		c.mu.Lock()
		c.status = StatusConnected
		c.qrCode = ""
		c.mu.Unlock()

		if c.handler != nil {
			c.handler.OnConnected(JIDToPhone(v.ID), v.ID.String())
		}

	case *events.Disconnected:
		log.Printf("[WhatsApp] Disconnected")
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()

		if c.handler != nil {
			c.handler.OnDisconnected()
		}

	case *events.LoggedOut:
		log.Printf("[WhatsApp] LoggedOut")
		c.mu.Lock()
		c.status = StatusDisconnected
		c.mu.Unlock()

		if c.handler != nil {
			c.handler.OnLoggedOut()
		}

	case *events.Message:
		if c.handler != nil {
			event := c.parseMessage(v)
			if event != nil {
				c.handler.OnMessage(*event)
			}
		}

	case *events.Receipt:
		if c.handler != nil {
			receiptType := ReceiptTypeDelivered
			if v.Type == types.ReceiptTypeRead {
				receiptType = ReceiptTypeRead
			}
			c.handler.OnReceipt(ReceiptEvent{
				MessageIDs: v.MessageIDs,
				ChatJID:    v.Chat.String(),
				SenderJID:  v.Sender.String(),
				Type:       receiptType,
				Timestamp:  v.Timestamp,
			})
		}
	}
}

func (c *Client) parseMessage(evt *events.Message) *MessageEvent {
	senderJID := c.resolveJID(evt.Info.Sender)
	chatJID := c.resolveJID(evt.Info.Chat)

	event := &MessageEvent{
		ID:         evt.Info.ID,
		ChatJID:    chatJID.String(),
		SenderJID:  senderJID.String(),
		SenderName: c.GetContactName(senderJID),
		IsFromMe:   evt.Info.IsFromMe,
		Timestamp:  evt.Info.Timestamp,
		RawMessage: evt.Message,
	}

	if evt.Message.GetConversation() != "" {
		event.Content = evt.Message.GetConversation()
	} else if evt.Message.GetExtendedTextMessage() != nil {
		event.Content = evt.Message.GetExtendedTextMessage().GetText()
	} else if evt.Message.GetImageMessage() != nil {
		event.MediaType = MediaTypeImage
		event.Content = evt.Message.GetImageMessage().GetCaption()
	} else if evt.Message.GetVideoMessage() != nil {
		event.MediaType = MediaTypeVideo
		event.Content = evt.Message.GetVideoMessage().GetCaption()
	} else if evt.Message.GetAudioMessage() != nil {
		event.MediaType = MediaTypeAudio
	} else if evt.Message.GetDocumentMessage() != nil {
		event.MediaType = MediaTypeDocument
		event.Content = evt.Message.GetDocumentMessage().GetFileName()
	} else {
		return nil
	}

	return event
}

func (c *Client) resolveJID(jid types.JID) types.JID {
	if jid.Server == "lid" {
		ctx := context.Background()
		pnJID, err := c.wa.Store.LIDs.GetPNForLID(ctx, jid)
		if err == nil && !pnJID.IsEmpty() {
			return pnJID
		}
	}
	return jid
}

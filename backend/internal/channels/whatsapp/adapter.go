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
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"

	"github.com/zyntra/backend/internal/ports"
)

// Adapter implementa ports.Channel para WhatsApp
type Adapter struct {
	client       *whatsmeow.Client
	device       *store.Device
	inboxID      string
	handler      ports.ChannelEventHandler
	status       ports.ChannelStatus
	qrCode       string
	killChan     chan bool
	mu           sync.RWMutex
}

// NewAdapter cria novo adapter WhatsApp
func NewAdapter(device *store.Device, inboxID string) *Adapter {
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()
	store.DeviceProps.Os = proto.String("Zyntra")

	client := whatsmeow.NewClient(device, waLog.Noop)

	adapter := &Adapter{
		client:   client,
		device:   device,
		inboxID:  inboxID,
		status:   ports.ChannelStatusDisconnected,
		killChan: make(chan bool, 1),
	}

	client.AddEventHandler(adapter.handleEvent)

	return adapter
}

// Type retorna o tipo do canal
func (a *Adapter) Type() ports.ChannelType {
	return ports.ChannelTypeWhatsApp
}

// Connect inicia a conexao
func (a *Adapter) Connect(ctx context.Context, inboxID string) error {
	a.mu.Lock()
	a.inboxID = inboxID
	a.status = ports.ChannelStatusConnecting
	a.mu.Unlock()

	go a.startConnection()
	return nil
}

func (a *Adapter) startConnection() {
	log.Printf("[WhatsApp] Starting connection for inbox %s", a.inboxID)

	if a.client.Store.ID == nil {
		log.Printf("[WhatsApp] New login, starting QR code flow for %s", a.inboxID)
		a.handleQRFlow()
	} else {
		a.connectExisting()
	}

	a.keepAliveLoop()
}

func (a *Adapter) handleQRFlow() {
	qrChan, err := a.client.GetQRChannel(context.Background())
	if err != nil {
		if err == whatsmeow.ErrQRStoreContainsID {
			log.Printf("[WhatsApp] Already logged in for %s", a.inboxID)
			a.connectExisting()
			return
		}
		log.Printf("[WhatsApp] Failed to get QR channel: %v", err)
		return
	}

	if err := a.client.Connect(); err != nil {
		log.Printf("[WhatsApp] Failed to connect: %v", err)
		return
	}

	for evt := range qrChan {
		switch evt.Event {
		case "code":
			image, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
			if err != nil {
				log.Printf("[WhatsApp] Failed to generate QR image: %v", err)
				continue
			}
			base64QR := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)

			a.mu.Lock()
			a.qrCode = base64QR
			a.status = ports.ChannelStatusQRCode
			a.mu.Unlock()

			if a.handler != nil {
				a.handler.OnQRCode(a.inboxID, evt.Code, base64QR)
			}

		case "timeout":
			log.Printf("[WhatsApp] QR timeout for %s", a.inboxID)
			a.mu.Lock()
			a.status = ports.ChannelStatusDisconnected
			a.qrCode = ""
			a.mu.Unlock()

			if a.handler != nil {
				a.handler.OnDisconnected(a.inboxID)
			}
			a.client.Disconnect()
			return

		case "success":
			log.Printf("[WhatsApp] QR pairing successful for %s", a.inboxID)
			a.mu.Lock()
			a.status = ports.ChannelStatusConnected
			a.qrCode = ""
			a.mu.Unlock()
		}
	}
}

func (a *Adapter) connectExisting() {
	log.Printf("[WhatsApp] Connecting existing session for %s", a.inboxID)

	if err := a.client.Connect(); err != nil {
		log.Printf("[WhatsApp] Failed to connect: %v", err)
		a.mu.Lock()
		a.status = ports.ChannelStatusDisconnected
		a.mu.Unlock()
		return
	}
}

func (a *Adapter) keepAliveLoop() {
	for {
		select {
		case <-a.killChan:
			log.Printf("[WhatsApp] Kill signal received for %s", a.inboxID)
			a.client.Disconnect()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// Disconnect encerra a conexao
func (a *Adapter) Disconnect(ctx context.Context) error {
	select {
	case a.killChan <- true:
	default:
	}

	a.mu.Lock()
	a.status = ports.ChannelStatusDisconnected
	a.qrCode = ""
	a.mu.Unlock()

	return nil
}

// Status retorna o status atual
func (a *Adapter) Status() ports.ChannelStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.client.IsConnected() && a.client.IsLoggedIn() {
		return ports.ChannelStatusConnected
	}
	return a.status
}

// SendText envia mensagem de texto
func (a *Adapter) SendText(ctx context.Context, to, content string) (string, error) {
	if !a.client.IsConnected() {
		return "", fmt.Errorf("client not connected")
	}

	jid := phoneToJID(to)
	msg := &waProto.Message{
		Conversation: proto.String(content),
	}

	resp, err := a.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return resp.ID, nil
}

// SendMedia envia mensagem com midia
func (a *Adapter) SendMedia(ctx context.Context, to string, media ports.Media) (string, error) {
	if !a.client.IsConnected() {
		return "", fmt.Errorf("client not connected")
	}

	jid := phoneToJID(to)

	var msg *waProto.Message

	switch media.Type {
	case ports.MediaTypeImage:
		uploaded, err := a.client.Upload(ctx, media.Data, whatsmeow.MediaImage)
		if err != nil {
			return "", fmt.Errorf("failed to upload image: %w", err)
		}
		msg = &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				Caption:       proto.String(media.Caption),
				Mimetype:      proto.String(media.MimeType),
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(media.Data))),
			},
		}

	case ports.MediaTypeDocument:
		uploaded, err := a.client.Upload(ctx, media.Data, whatsmeow.MediaDocument)
		if err != nil {
			return "", fmt.Errorf("failed to upload document: %w", err)
		}
		msg = &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				Caption:       proto.String(media.Caption),
				Mimetype:      proto.String(media.MimeType),
				FileName:      proto.String(media.FileName),
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(media.Data))),
			},
		}

	default:
		return "", fmt.Errorf("unsupported media type: %s", media.Type)
	}

	resp, err := a.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send media: %w", err)
	}

	return resp.ID, nil
}

// SetEventHandler define o handler de eventos
func (a *Adapter) SetEventHandler(handler ports.ChannelEventHandler) {
	a.handler = handler
}

// GetQRCode retorna QR code atual
func (a *Adapter) GetQRCode() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.qrCode
}

// GetJID retorna o JID do dispositivo conectado
func (a *Adapter) GetJID() string {
	if a.client.Store.ID != nil {
		return a.client.Store.ID.String()
	}
	return ""
}

// GetPhoneNumber retorna o numero de telefone
func (a *Adapter) GetPhoneNumber() string {
	if a.client.Store.ID != nil {
		return jidToPhone(*a.client.Store.ID)
	}
	return ""
}

// Logout faz logout do dispositivo
func (a *Adapter) Logout(ctx context.Context) error {
	return a.client.Logout(ctx)
}

func (a *Adapter) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		log.Printf("[WhatsApp] Connected for %s", a.inboxID)
		a.mu.Lock()
		a.status = ports.ChannelStatusConnected
		a.qrCode = ""
		a.mu.Unlock()

		if a.handler != nil {
			a.handler.OnConnected(a.inboxID, a.GetPhoneNumber())
		}

	case *events.PairSuccess:
		log.Printf("[WhatsApp] PairSuccess for %s: JID=%s", a.inboxID, v.ID.String())
		a.mu.Lock()
		a.status = ports.ChannelStatusConnected
		a.qrCode = ""
		a.mu.Unlock()

		if a.handler != nil {
			a.handler.OnConnected(a.inboxID, jidToPhone(v.ID))
		}

	case *events.Disconnected:
		log.Printf("[WhatsApp] Disconnected for %s", a.inboxID)
		a.mu.Lock()
		a.status = ports.ChannelStatusDisconnected
		a.mu.Unlock()

		if a.handler != nil {
			a.handler.OnDisconnected(a.inboxID)
		}

	case *events.LoggedOut:
		log.Printf("[WhatsApp] LoggedOut for %s", a.inboxID)
		a.mu.Lock()
		a.status = ports.ChannelStatusDisconnected
		a.mu.Unlock()

		if a.handler != nil {
			a.handler.OnDisconnected(a.inboxID)
		}

	case *events.Message:
		if a.handler != nil {
			event := a.parseMessage(v)
			if event != nil {
				a.handler.OnMessage(*event)
			}
		}

	case *events.Receipt:
		if a.handler != nil {
			status := ports.MessageStatusDelivered
			if v.Type == types.ReceiptTypeRead {
				status = ports.MessageStatusRead
			}
			for _, id := range v.MessageIDs {
				a.handler.OnStatusUpdate(a.inboxID, id, status)
			}
		}
	}
}

func (a *Adapter) parseMessage(evt *events.Message) *ports.IncomingEvent {
	senderJID := a.resolveJID(evt.Info.Sender)
	chatJID := a.resolveJID(evt.Info.Chat)

	contactName := a.getContactName(senderJID)

	event := &ports.IncomingEvent{
		InboxID:     a.inboxID,
		SourceID:    evt.Info.ID,
		ContactID:   chatJID.String(),
		ContactName: contactName,
		IsFromMe:    evt.Info.IsFromMe,
		Type:        ports.EventTypeMessage,
		Timestamp:   evt.Info.Timestamp,
	}

	if evt.Message.GetConversation() != "" {
		event.Content = evt.Message.GetConversation()
	} else if evt.Message.GetExtendedTextMessage() != nil {
		event.Content = evt.Message.GetExtendedTextMessage().GetText()
	} else if evt.Message.GetImageMessage() != nil {
		event.MediaType = ports.MediaTypeImage
		event.Content = evt.Message.GetImageMessage().GetCaption()
	} else if evt.Message.GetVideoMessage() != nil {
		event.MediaType = ports.MediaTypeVideo
		event.Content = evt.Message.GetVideoMessage().GetCaption()
	} else if evt.Message.GetAudioMessage() != nil {
		event.MediaType = ports.MediaTypeAudio
	} else if evt.Message.GetDocumentMessage() != nil {
		event.MediaType = ports.MediaTypeDocument
		event.Content = evt.Message.GetDocumentMessage().GetFileName()
	} else {
		return nil
	}

	return event
}

func (a *Adapter) resolveJID(jid types.JID) types.JID {
	if jid.Server == "lid" {
		ctx := context.Background()
		pnJID, err := a.client.Store.LIDs.GetPNForLID(ctx, jid)
		if err == nil && !pnJID.IsEmpty() {
			return pnJID
		}
	}
	return jid
}

func (a *Adapter) getContactName(jid types.JID) string {
	ctx := context.Background()
	contactInfo, err := a.client.Store.Contacts.GetContact(ctx, jid)
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

// Helper functions
func phoneToJID(phone string) types.JID {
	user := phone
	if len(user) > 0 && user[0] == '+' {
		user = user[1:]
	}
	return types.NewJID(user, types.DefaultUserServer)
}

func jidToPhone(jid types.JID) string {
	return "+" + jid.User
}

// Store gerencia armazenamento de sessoes WhatsApp
type Store struct {
	Container *sqlstore.Container
}

// NewStore cria novo store
func NewStore(databaseURL string) (*Store, error) {
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "pgx", databaseURL, waLog.Noop)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}
	return &Store{Container: container}, nil
}

// GetDevice obtem ou cria device
func (s *Store) GetDevice(ctx context.Context, jid string) (*store.Device, error) {
	if jid != "" {
		parsedJID, err := types.ParseJID(jid)
		if err == nil {
			device, err := s.Container.GetDevice(ctx, parsedJID)
			if err == nil && device != nil {
				return device, nil
			}
		}
	}
	return s.Container.NewDevice(), nil
}

// Close fecha o store
func (s *Store) Close() error {
	return nil
}

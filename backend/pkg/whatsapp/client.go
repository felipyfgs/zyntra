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
	waLog "go.mau.fi/whatsmeow/util/log"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

// Client wraps whatsmeow.Client with additional functionality (wuzapi MyClient pattern)
type Client struct {
	WAClient       *whatsmeow.Client
	eventHandlerID uint32
	connectionID   string
	handler        EventHandler
	mu             sync.RWMutex
	connected      bool
}

// NewClient creates a new WhatsApp client (following wuzapi pattern)
func NewClient(device *store.Device, connectionID string, handler EventHandler) *Client {
	// Set device properties BEFORE creating client (like wuzapi does)
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()
	store.DeviceProps.Os = proto.String("Zyntra")

	// Create whatsmeow client
	client := whatsmeow.NewClient(device, waLog.Noop)

	c := &Client{
		WAClient:     client,
		connectionID: connectionID,
		handler:      handler,
	}

	// Register event handler and save ID (like wuzapi)
	c.eventHandlerID = client.AddEventHandler(c.handleEvent)

	return c
}

// StartClient handles the full connection flow (following wuzapi startClient pattern)
// This runs in a goroutine and handles QR codes, connection, and keep-alive
func (c *Client) StartClient(ctx context.Context, killChan <-chan bool) {
	log.Printf("[WhatsApp] Starting client for connection %s", c.connectionID)

	if c.WAClient.Store.ID == nil {
		// No ID stored, new login - need QR code
		log.Printf("[WhatsApp] No stored session, starting QR code flow for %s", c.connectionID)

		qrChan, err := c.WAClient.GetQRChannel(ctx)
		if err != nil {
			if err == whatsmeow.ErrQRStoreContainsID {
				log.Printf("[WhatsApp] Already logged in for %s", c.connectionID)
				c.connectExisting(ctx, killChan)
				return
			}
			log.Printf("[WhatsApp] Failed to get QR channel for %s: %v", c.connectionID, err)
			return
		}

		// Must connect to generate QR codes
		if err := c.WAClient.Connect(); err != nil {
			log.Printf("[WhatsApp] Failed to connect for %s: %v", c.connectionID, err)
			return
		}

		// Process QR codes (like wuzapi - in same goroutine)
		for evt := range qrChan {
			log.Printf("[WhatsApp] QR event: %s for %s", evt.Event, c.connectionID)

			switch evt.Event {
			case "code":
				// Generate QR code image as base64 (like wuzapi)
				image, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
				if err != nil {
					log.Printf("[WhatsApp] Failed to generate QR image: %v", err)
					continue
				}
				base64QR := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)

				log.Printf("[WhatsApp] New QR code for %s (timeout: %v)", c.connectionID, evt.Timeout)

				if c.handler != nil {
					c.handler.OnQRCode(QRCodeEvent{
						ConnectionID: c.connectionID,
						Code:         evt.Code,
						Base64Image:  base64QR,
					})
				}

			case "timeout":
				log.Printf("[WhatsApp] QR timeout for %s - killing channel", c.connectionID)
				if c.handler != nil {
					c.handler.OnQRTimeout(c.connectionID)
				}
				c.WAClient.Disconnect()
				return

			case "success":
				log.Printf("[WhatsApp] QR pairing successful for %s", c.connectionID)
				// Clear QR code and mark as connected
				if c.handler != nil {
					jid := ""
					if c.WAClient.Store.ID != nil {
						jid = c.WAClient.Store.ID.String()
					}
					c.handler.OnPairSuccess(PairSuccessEvent{
						ConnectionID: c.connectionID,
						JID:          jid,
					})
				}
			}
		}
	} else {
		// Already logged in, just connect
		c.connectExisting(ctx, killChan)
		return
	}

	// Keep-alive loop (like wuzapi)
	c.keepAliveLoop(killChan)
}

// connectExisting connects an already paired device
func (c *Client) connectExisting(ctx context.Context, killChan <-chan bool) {
	log.Printf("[WhatsApp] Connecting existing session for %s", c.connectionID)

	if err := c.WAClient.Connect(); err != nil {
		log.Printf("[WhatsApp] Failed to connect existing session for %s: %v", c.connectionID, err)
		return
	}

	c.keepAliveLoop(killChan)
}

// keepAliveLoop keeps the client alive until killed (like wuzapi)
func (c *Client) keepAliveLoop(killChan <-chan bool) {
	log.Printf("[WhatsApp] Starting keep-alive loop for %s", c.connectionID)

	for {
		select {
		case <-killChan:
			log.Printf("[WhatsApp] Received kill signal for %s", c.connectionID)
			c.WAClient.Disconnect()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// Connect initiates the connection (simple version for backward compatibility)
func (c *Client) Connect(ctx context.Context) error {
	return c.WAClient.Connect()
}

// Disconnect closes the connection to WhatsApp
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.WAClient.Disconnect()
	c.connected = false
}

// IsConnected returns the connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.WAClient.IsConnected()
}

// IsLoggedIn returns whether the client is logged in
func (c *Client) IsLoggedIn() bool {
	return c.WAClient.IsLoggedIn()
}

// GetJID returns the JID of the connected device
func (c *Client) GetJID() string {
	if c.WAClient.Store.ID != nil {
		return c.WAClient.Store.ID.String()
	}
	return ""
}

// GetPhoneNumber returns the phone number of the connected device
func (c *Client) GetPhoneNumber() string {
	if c.WAClient.Store.ID != nil {
		return JIDToPhone(*c.WAClient.Store.ID)
	}
	return ""
}

// SendTextMessage sends a text message
func (c *Client) SendTextMessage(ctx context.Context, to string, text string) (*Message, error) {
	if !c.WAClient.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	jid := PhoneToJID(to)

	msg := &waProto.Message{
		Conversation: proto.String(text),
	}

	resp, err := c.WAClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &Message{
		ID:           resp.ID,
		ConnectionID: c.connectionID,
		ChatJID:      jid.String(),
		SenderJID:    c.GetJID(),
		Direction:    DirectionOutbound,
		Content:      text,
		Status:       MessageSent,
		Timestamp:    resp.Timestamp,
		CreatedAt:    time.Now(),
	}, nil
}

// SendImageMessage sends an image message
func (c *Client) SendImageMessage(ctx context.Context, to string, imageData []byte, caption string, mimeType string) (*Message, error) {
	if !c.WAClient.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	jid := PhoneToJID(to)

	uploaded, err := c.WAClient.Upload(ctx, imageData, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
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
			FileLength:    proto.Uint64(uint64(len(imageData))),
		},
	}

	resp, err := c.WAClient.SendMessage(ctx, jid, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send image: %w", err)
	}

	return &Message{
		ID:           resp.ID,
		ConnectionID: c.connectionID,
		ChatJID:      jid.String(),
		SenderJID:    c.GetJID(),
		Direction:    DirectionOutbound,
		Content:      caption,
		MediaType:    "image",
		Status:       MessageSent,
		Timestamp:    resp.Timestamp,
		CreatedAt:    time.Now(),
	}, nil
}

// Logout logs out and removes the device from WhatsApp
func (c *Client) Logout(ctx context.Context) error {
	return c.WAClient.Logout(ctx)
}

// handleEvent processes events from whatsmeow (like wuzapi myEventHandler)
func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		c.mu.Lock()
		c.connected = true
		c.mu.Unlock()

		log.Printf("[WhatsApp] Connected event for %s", c.connectionID)

		if c.handler != nil {
			c.handler.OnConnected(ConnectionEvent{
				ConnectionID: c.connectionID,
				Status:       StatusConnected,
				Phone:        c.GetPhoneNumber(),
			})
		}

	case *events.PairSuccess:
		log.Printf("[WhatsApp] PairSuccess for %s: JID=%s", c.connectionID, v.ID.String())

		if c.handler != nil {
			c.handler.OnPairSuccess(PairSuccessEvent{
				ConnectionID: c.connectionID,
				JID:          v.ID.String(),
				BusinessName: v.BusinessName,
				Platform:     v.Platform,
			})
		}

	case *events.Disconnected:
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

		log.Printf("[WhatsApp] Disconnected event for %s", c.connectionID)

		if c.handler != nil {
			c.handler.OnDisconnected(ConnectionEvent{
				ConnectionID: c.connectionID,
				Status:       StatusDisconnected,
			})
		}

	case *events.Message:
		if c.handler != nil {
			msg := c.parseMessage(v)
			if msg != nil {
				c.handler.OnMessage(MessageEvent{
					ConnectionID: c.connectionID,
					Message:      msg,
				})
			}
		}

	case *events.Receipt:
		if c.handler != nil {
			status := MessageDelivered
			if v.Type == types.ReceiptTypeRead {
				status = MessageRead
			}

			for _, id := range v.MessageIDs {
				c.handler.OnMessageStatus(MessageEvent{
					ConnectionID: c.connectionID,
					MessageID:    id,
					Status:       status,
				})
			}
		}

	case *events.LoggedOut:
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

		log.Printf("[WhatsApp] LoggedOut event for %s", c.connectionID)

		if c.handler != nil {
			c.handler.OnDisconnected(ConnectionEvent{
				ConnectionID: c.connectionID,
				Status:       StatusDisconnected,
			})
		}
	}
}

// parseMessage converts a whatsmeow message event to our Message type
func (c *Client) parseMessage(evt *events.Message) *Message {
	msg := &Message{
		ID:           evt.Info.ID,
		ConnectionID: c.connectionID,
		ChatJID:      evt.Info.Chat.String(),
		SenderJID:    evt.Info.Sender.String(),
		Direction:    DirectionInbound,
		Status:       MessageDelivered,
		Timestamp:    evt.Info.Timestamp,
		CreatedAt:    time.Now(),
	}

	if evt.Info.IsFromMe {
		msg.Direction = DirectionOutbound
	}

	if evt.Message.GetConversation() != "" {
		msg.Content = evt.Message.GetConversation()
	} else if evt.Message.GetExtendedTextMessage() != nil {
		msg.Content = evt.Message.GetExtendedTextMessage().GetText()
	} else if evt.Message.GetImageMessage() != nil {
		msg.MediaType = "image"
		msg.Content = evt.Message.GetImageMessage().GetCaption()
	} else if evt.Message.GetVideoMessage() != nil {
		msg.MediaType = "video"
		msg.Content = evt.Message.GetVideoMessage().GetCaption()
	} else if evt.Message.GetAudioMessage() != nil {
		msg.MediaType = "audio"
	} else if evt.Message.GetDocumentMessage() != nil {
		msg.MediaType = "document"
		msg.Content = evt.Message.GetDocumentMessage().GetFileName()
	} else {
		return nil
	}

	return msg
}

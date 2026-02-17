package ports

import (
	"context"
	"time"
)

// ChannelType tipo do canal
type ChannelType string

const (
	ChannelTypeWhatsApp ChannelType = "whatsapp"
	ChannelTypeTelegram ChannelType = "telegram"
	ChannelTypeAPI      ChannelType = "api"
)

// ChannelStatus status de conexao do canal
type ChannelStatus string

const (
	ChannelStatusDisconnected ChannelStatus = "disconnected"
	ChannelStatusConnecting   ChannelStatus = "connecting"
	ChannelStatusQRCode       ChannelStatus = "qr_code"
	ChannelStatusConnected    ChannelStatus = "connected"
)

// Channel interface que todo canal deve implementar
type Channel interface {
	// Type retorna o tipo do canal
	Type() ChannelType

	// Connect inicia a conexao do canal
	Connect(ctx context.Context, inboxID string) error

	// Disconnect encerra a conexao
	Disconnect(ctx context.Context) error

	// Status retorna o status atual
	Status() ChannelStatus

	// SendText envia mensagem de texto
	SendText(ctx context.Context, to, content string) (sourceID string, err error)

	// SendMedia envia mensagem com midia
	SendMedia(ctx context.Context, to string, media Media) (sourceID string, err error)

	// SetEventHandler define o handler de eventos
	SetEventHandler(handler ChannelEventHandler)

	// GetQRCode retorna QR code atual (se aplicavel)
	GetQRCode() string
}

// ChannelEventHandler interface para processar eventos do canal
type ChannelEventHandler interface {
	OnMessage(event IncomingEvent)
	OnStatusUpdate(inboxID, sourceID string, status MessageStatus)
	OnQRCode(inboxID, qrCode, base64Image string)
	OnConnected(inboxID, phone string)
	OnDisconnected(inboxID string)
}

// IncomingEvent evento recebido de qualquer canal
type IncomingEvent struct {
	InboxID     string
	SourceID    string
	ContactID   string
	ContactName string
	IsFromMe    bool
	Type        EventType
	Content     string
	MediaURL    string
	MediaType   MediaType
	Timestamp   time.Time
	RawPayload  map[string]interface{}
}

// EventType tipo de evento
type EventType string

const (
	EventTypeMessage      EventType = "message"
	EventTypeStatusUpdate EventType = "status_update"
	EventTypeTyping       EventType = "typing"
	EventTypeQRCode       EventType = "qr_code"
	EventTypeConnected    EventType = "connected"
	EventTypeDisconnected EventType = "disconnected"
)

// Media midia a ser enviada
type Media struct {
	Type     MediaType
	URL      string
	Data     []byte
	MimeType string
	Caption  string
	FileName string
}

// MediaType tipo de midia
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeSticker  MediaType = "sticker"
)

// MessageStatus status de entrega
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// ChannelFactory cria instancias de canais
type ChannelFactory interface {
	Create(channelType ChannelType, config map[string]interface{}) (Channel, error)
}

// ChannelManager gerencia multiplos canais
type ChannelManager interface {
	Get(inboxID string) (Channel, error)
	Register(inboxID string, channel Channel) error
	Unregister(inboxID string) error
	GetAll() map[string]Channel
}

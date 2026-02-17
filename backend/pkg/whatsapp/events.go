package whatsapp

import "time"

// Status do cliente WhatsApp
type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusQRCode       Status = "qr_code"
	StatusConnected    Status = "connected"
)

// QREvent evento do fluxo de QR code
type QREvent struct {
	Event  string // "code", "timeout", "success"
	Code   string // Codigo QR (apenas quando Event == "code")
	Base64 string // Imagem base64 do QR
}

// MessageEvent mensagem recebida do WhatsApp
type MessageEvent struct {
	ID          string
	ChatJID     string
	SenderJID   string
	SenderName  string
	Content     string
	MediaType   MediaType
	MediaURL    string
	IsFromMe    bool
	Timestamp   time.Time
	RawMessage  interface{}
}

// MediaType tipo de midia
type MediaType string

const (
	MediaTypeNone     MediaType = ""
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeSticker  MediaType = "sticker"
)

// ReceiptEvent evento de recebimento/leitura
type ReceiptEvent struct {
	MessageIDs []string
	ChatJID    string
	SenderJID  string
	Type       ReceiptType
	Timestamp  time.Time
}

// ReceiptType tipo de recibo
type ReceiptType string

const (
	ReceiptTypeDelivered ReceiptType = "delivered"
	ReceiptTypeRead      ReceiptType = "read"
)

// ConnectionEvent evento de conexao
type ConnectionEvent struct {
	Status Status
	Phone  string
	JID    string
}

// EventHandler interface para processar eventos do cliente
type EventHandler interface {
	OnMessage(event MessageEvent)
	OnReceipt(event ReceiptEvent)
	OnQRCode(event QREvent)
	OnConnected(phone, jid string)
	OnDisconnected()
	OnLoggedOut()
}

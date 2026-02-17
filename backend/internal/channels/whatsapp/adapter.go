package whatsapp

import (
	"context"
	"fmt"

	wapkg "github.com/zyntra/backend/pkg/whatsapp"
	"github.com/zyntra/backend/internal/ports"
)

// Adapter implementa ports.Channel usando pkg/whatsapp
type Adapter struct {
	client  *wapkg.Client
	inboxID string
	handler ports.ChannelEventHandler
}

// NewAdapter cria novo adapter
func NewAdapter(client *wapkg.Client, inboxID string) *Adapter {
	adapter := &Adapter{
		client:  client,
		inboxID: inboxID,
	}

	// Conectar eventos do cliente ao adapter
	client.SetEventHandler(adapter)

	return adapter
}

// Type retorna o tipo do canal
func (a *Adapter) Type() ports.ChannelType {
	return ports.ChannelTypeWhatsApp
}

// Connect inicia a conexao
func (a *Adapter) Connect(ctx context.Context, inboxID string) error {
	a.inboxID = inboxID
	_, err := a.client.Connect(ctx)
	return err
}

// Disconnect encerra a conexao
func (a *Adapter) Disconnect(ctx context.Context) error {
	return a.client.Disconnect()
}

// Status retorna o status atual
func (a *Adapter) Status() ports.ChannelStatus {
	switch a.client.Status() {
	case wapkg.StatusConnected:
		return ports.ChannelStatusConnected
	case wapkg.StatusConnecting:
		return ports.ChannelStatusConnecting
	case wapkg.StatusQRCode:
		return ports.ChannelStatusQRCode
	default:
		return ports.ChannelStatusDisconnected
	}
}

// SendText envia mensagem de texto
func (a *Adapter) SendText(ctx context.Context, to, content string) (string, error) {
	return a.client.SendText(ctx, to, content)
}

// SendMedia envia mensagem com midia
func (a *Adapter) SendMedia(ctx context.Context, to string, media ports.Media) (string, error) {
	switch media.Type {
	case ports.MediaTypeImage:
		return a.client.SendImage(ctx, to, media.Data, media.Caption, media.MimeType)
	case ports.MediaTypeDocument:
		return a.client.SendDocument(ctx, to, media.Data, media.FileName, media.Caption, media.MimeType)
	default:
		return "", fmt.Errorf("unsupported media type: %s", media.Type)
	}
}

// SetEventHandler define o handler de eventos
func (a *Adapter) SetEventHandler(handler ports.ChannelEventHandler) {
	a.handler = handler
}

// GetQRCode retorna QR code atual
func (a *Adapter) GetQRCode() string {
	return a.client.GetQRCode()
}

// GetJID retorna JID do dispositivo
func (a *Adapter) GetJID() string {
	return a.client.GetJID()
}

// GetPhone retorna numero de telefone
func (a *Adapter) GetPhone() string {
	return a.client.GetPhone()
}

// Logout faz logout do dispositivo
func (a *Adapter) Logout(ctx context.Context) error {
	return a.client.Logout(ctx)
}

// Client retorna o cliente subjacente
func (a *Adapter) Client() *wapkg.Client {
	return a.client
}

// ========== Implementacao de wapkg.EventHandler ==========

// OnMessage processa mensagem recebida
func (a *Adapter) OnMessage(event wapkg.MessageEvent) {
	if a.handler == nil {
		return
	}

	a.handler.OnMessage(ports.IncomingEvent{
		InboxID:     a.inboxID,
		SourceID:    event.ID,
		ContactID:   event.ChatJID,
		ContactName: event.SenderName,
		IsFromMe:    event.IsFromMe,
		Type:        ports.EventTypeMessage,
		Content:     event.Content,
		MediaType:   convertMediaType(event.MediaType),
		Timestamp:   event.Timestamp,
	})
}

// OnReceipt processa recibo de entrega/leitura
func (a *Adapter) OnReceipt(event wapkg.ReceiptEvent) {
	if a.handler == nil {
		return
	}

	status := ports.MessageStatusDelivered
	if event.Type == wapkg.ReceiptTypeRead {
		status = ports.MessageStatusRead
	}

	for _, id := range event.MessageIDs {
		a.handler.OnStatusUpdate(a.inboxID, id, status)
	}
}

// OnQRCode processa evento de QR code
func (a *Adapter) OnQRCode(event wapkg.QREvent) {
	if a.handler == nil {
		return
	}

	a.handler.OnQRCode(a.inboxID, event.Code, event.Base64)
}

// OnConnected processa evento de conexao
func (a *Adapter) OnConnected(phone, jid string) {
	if a.handler == nil {
		return
	}

	a.handler.OnConnected(a.inboxID, phone)
}

// OnDisconnected processa evento de desconexao
func (a *Adapter) OnDisconnected() {
	if a.handler == nil {
		return
	}

	a.handler.OnDisconnected(a.inboxID)
}

// OnLoggedOut processa evento de logout
func (a *Adapter) OnLoggedOut() {
	if a.handler == nil {
		return
	}

	a.handler.OnDisconnected(a.inboxID)
}

// ========== Helpers ==========

func convertMediaType(mt wapkg.MediaType) ports.MediaType {
	switch mt {
	case wapkg.MediaTypeImage:
		return ports.MediaTypeImage
	case wapkg.MediaTypeVideo:
		return ports.MediaTypeVideo
	case wapkg.MediaTypeAudio:
		return ports.MediaTypeAudio
	case wapkg.MediaTypeDocument:
		return ports.MediaTypeDocument
	case wapkg.MediaTypeSticker:
		return ports.MediaTypeSticker
	default:
		return ""
	}
}

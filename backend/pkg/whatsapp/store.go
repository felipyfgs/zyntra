package whatsapp

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Store gerencia armazenamento de sessoes WhatsApp
type Store struct {
	Container *sqlstore.Container
}

// NewStore cria novo store conectado ao banco de dados
func NewStore(databaseURL string) (*Store, error) {
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "pgx", databaseURL, waLog.Noop)
	if err != nil {
		return nil, fmt.Errorf("failed to create whatsapp store: %w", err)
	}
	return &Store{Container: container}, nil
}

// GetDevice obtem device existente ou cria novo
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

// GetAllDevices retorna todos os devices salvos
func (s *Store) GetAllDevices(ctx context.Context) ([]*store.Device, error) {
	return s.Container.GetAllDevices(ctx)
}

// DeleteDevice remove um device
func (s *Store) DeleteDevice(ctx context.Context, jid types.JID) error {
	device, err := s.Container.GetDevice(ctx, jid)
	if err != nil {
		return err
	}
	if device == nil {
		return nil
	}
	return device.Delete(ctx)
}

// Close fecha a conexao com o banco
func (s *Store) Close() error {
	// sqlstore.Container nao tem metodo Close
	return nil
}

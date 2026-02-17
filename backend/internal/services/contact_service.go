package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zyntra/backend/internal/domain"
	"github.com/zyntra/backend/internal/repository"
)

// ContactService servico de contatos
type ContactService struct {
	contactRepo      *repository.ContactRepository
	contactInboxRepo *repository.ContactInboxRepository
}

// NewContactService cria novo servico
func NewContactService(
	contactRepo *repository.ContactRepository,
	contactInboxRepo *repository.ContactInboxRepository,
) *ContactService {
	return &ContactService{
		contactRepo:      contactRepo,
		contactInboxRepo: contactInboxRepo,
	}
}

// Create cria um contato
func (s *ContactService) Create(ctx context.Context, req domain.CreateContactRequest) (*domain.Contact, error) {
	contact := &domain.Contact{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Email:            req.Email,
		PhoneNumber:      req.PhoneNumber,
		CustomAttributes: req.CustomAttributes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return contact, nil
}

// GetByID busca contato por ID
func (s *ContactService) GetByID(ctx context.Context, id string) (*domain.Contact, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	if contact == nil {
		return nil, fmt.Errorf("contact not found")
	}
	return contact, nil
}

// GetByPhone busca contato por telefone
func (s *ContactService) GetByPhone(ctx context.Context, phone string) (*domain.Contact, error) {
	return s.contactRepo.GetByPhone(ctx, phone)
}

// List lista contatos
func (s *ContactService) List(ctx context.Context, limit, offset int) ([]*domain.Contact, error) {
	return s.contactRepo.GetAll(ctx, limit, offset)
}

// Search busca contatos
func (s *ContactService) Search(ctx context.Context, term string, limit int) ([]*domain.Contact, error) {
	return s.contactRepo.Search(ctx, term, limit)
}

// Update atualiza um contato
func (s *ContactService) Update(ctx context.Context, id string, req domain.UpdateContactRequest) (*domain.Contact, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil || contact == nil {
		return nil, fmt.Errorf("contact not found")
	}

	if req.Name != nil {
		contact.Name = *req.Name
	}
	if req.Email != nil {
		contact.Email = *req.Email
	}
	if req.PhoneNumber != nil {
		contact.PhoneNumber = *req.PhoneNumber
	}
	if req.AvatarURL != nil {
		contact.AvatarURL = *req.AvatarURL
	}
	if req.CustomAttributes != nil {
		for k, v := range req.CustomAttributes {
			if contact.CustomAttributes == nil {
				contact.CustomAttributes = make(map[string]interface{})
			}
			contact.CustomAttributes[k] = v
		}
	}

	contact.UpdatedAt = time.Now()

	if err := s.contactRepo.Update(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return contact, nil
}

// Delete remove um contato
func (s *ContactService) Delete(ctx context.Context, id string) error {
	return s.contactRepo.Delete(ctx, id)
}

// GetWithInboxes busca contato com suas identidades por canal
func (s *ContactService) GetWithInboxes(ctx context.Context, id string) (*domain.ContactWithInboxes, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil || contact == nil {
		return nil, fmt.Errorf("contact not found")
	}

	contactInboxes, _ := s.contactInboxRepo.GetByContactID(ctx, id)

	return &domain.ContactWithInboxes{
		Contact:        *contact,
		ContactInboxes: derefContactInboxes(contactInboxes),
	}, nil
}

func derefContactInboxes(list []*domain.ContactInbox) []domain.ContactInbox {
	result := make([]domain.ContactInbox, len(list))
	for i, ci := range list {
		result[i] = *ci
	}
	return result
}

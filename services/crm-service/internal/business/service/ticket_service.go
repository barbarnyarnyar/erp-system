package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
)

type ServiceTicketService struct {
	ticketRepo domain.ServiceTicketRepository
	publisher  domain.EventPublisher
}

func NewServiceTicketService(ticketRepo domain.ServiceTicketRepository, publisher domain.EventPublisher) *ServiceTicketService {
	return &ServiceTicketService{
		ticketRepo: ticketRepo,
		publisher:  publisher,
	}
}

func (s *ServiceTicketService) CreateServiceTicket(ctx context.Context, customerID, title, description, priority string) (*domain.ServiceTicket, error) {
	id := utils.NewID("ticket")
	ticket := &domain.ServiceTicket{
		ID:          id,
		CustomerID:  customerID,
		Title:       title,
		Description: description,
		Status:      "OPEN",
		Priority:    priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmServiceTicketCreated, id, domain.ServiceTicketCreatedEvent{
		TicketID:   id,
		CustomerID: customerID,
		Title:      title,
		Priority:   priority,
		Timestamp:  time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmServiceTicketCreated, err)
	}

	return ticket, nil
}

func (s *ServiceTicketService) GetServiceTicket(ctx context.Context, id string) (*domain.ServiceTicket, error) {
	return s.ticketRepo.GetByID(ctx, id)
}

func (s *ServiceTicketService) ListServiceTickets(ctx context.Context) ([]domain.ServiceTicket, error) {
	return s.ticketRepo.List(ctx)
}

func (s *ServiceTicketService) UpdateServiceTicket(ctx context.Context, id string, status, priority string) (*domain.ServiceTicket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := ticket.Status
	ticket.Status = status
	ticket.Priority = priority
	ticket.UpdatedAt = time.Now()

	err = s.ticketRepo.Update(ctx, ticket)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmServiceTicketUpdated, id, domain.ServiceTicketUpdatedEvent{
		TicketID:  id,
		Status:    status,
		Priority:  priority,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmServiceTicketUpdated, err)
	}

	if oldStatus != status {
		if status == "RESOLVED" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmServiceTicketResolved, id, domain.ServiceTicketResolvedEvent{
				TicketID:  id,
				Timestamp: time.Now(),
			}); err != nil {
				utils.LogPublishErr("crm-service", domain.TopicCrmServiceTicketResolved, err)
			}
		} else if status == "ESCALATED" {
			if err := s.publisher.Publish(ctx, domain.TopicCrmServiceTicketEscalated, id, domain.ServiceTicketEscalatedEvent{
				TicketID:  id,
				Timestamp: time.Now(),
			}); err != nil {
				utils.LogPublishErr("crm-service", domain.TopicCrmServiceTicketEscalated, err)
			}
		}
	}

	return ticket, nil
}

func (s *ServiceTicketService) DeleteServiceTicket(ctx context.Context, id string) error {
	return s.ticketRepo.Delete(ctx, id)
}

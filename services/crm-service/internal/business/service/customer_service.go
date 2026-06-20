package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type CustomerService struct {
	customerRepo domain.CustomerRepository
	publisher    domain.EventPublisher
}

func NewCustomerService(customerRepo domain.CustomerRepository, publisher domain.EventPublisher) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		publisher:    publisher,
	}
}

func (s *CustomerService) CreateCustomer(ctx context.Context, companyName, contactName, email, phone, category, parentCustomerID string) (*domain.CustomerProfile, error) {
	id := utils.NewID("cust")
	cust := &domain.CustomerProfile{
		ID:                 id,
		LegalEntityID:      "default_entity_id",
		CustomerCode:       "CODE-" + id[:8],
		CompanyName:        companyName,
		AccountManagerHrID: "default_manager_id",
		Status:             domain.CustomerStatusACTIVE,
		CreditLimit:        decimal.NewFromInt(50000),
		Currency:           "USD",
		Version:            1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}


	err := s.customerRepo.Create(ctx, cust)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmCustomerCreated, id, domain.CustomerCreatedEvent{
		CustomerID:  id,
		CompanyName: companyName,
		ContactName: contactName,
		Email:       email,
		Timestamp:   time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmCustomerCreated, err)
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmCustomerActivated, id, domain.CustomerActivatedEvent{
		CustomerID: id,
		Timestamp:  time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmCustomerActivated, err)
	}

	return cust, nil
}

func (s *CustomerService) GetCustomer(ctx context.Context, id string) (*domain.CustomerProfile, error) {
	return s.customerRepo.GetByID(ctx, id)
}

func (s *CustomerService) ListCustomers(ctx context.Context) ([]domain.CustomerProfile, error) {
	return s.customerRepo.List(ctx)
}

func (s *CustomerService) UpdateCustomer(ctx context.Context, id string, companyName, contactName, email, phone, status, category string) (*domain.CustomerProfile, error) {
	statusEnum := domain.CustomerStatus(status)
	if !statusEnum.IsValid() {
		return nil, fmt.Errorf("invalid customer status: %s", status)
	}

	cust, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := cust.Status
	cust.CompanyName = companyName
	cust.Status = statusEnum
	cust.UpdatedAt = time.Now()

	err = s.customerRepo.Update(ctx, cust)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicCrmCustomerUpdated, id, domain.CustomerUpdatedEvent{
		CustomerID:  id,
		CompanyName: companyName,
		Status:      status,
		Timestamp:   time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmCustomerUpdated, err)
	}

	if oldStatus != statusEnum {
		if statusEnum == domain.CustomerStatusACTIVE {
			if err := s.publisher.Publish(ctx, domain.TopicCrmCustomerActivated, id, domain.CustomerActivatedEvent{
				CustomerID: id,
				Timestamp:  time.Now(),
			}); err != nil {
				utils.LogPublishErr("crm-service", domain.TopicCrmCustomerActivated, err)
			}
		} else if statusEnum == domain.CustomerStatusINACTIVE {
			if err := s.publisher.Publish(ctx, domain.TopicCrmCustomerDeactivated, id, domain.CustomerDeactivatedEvent{
				CustomerID: id,
				Timestamp:  time.Now(),
			}); err != nil {
				utils.LogPublishErr("crm-service", domain.TopicCrmCustomerDeactivated, err)
			}
		}
	}

	return cust, nil
}

func (s *CustomerService) DeleteCustomer(ctx context.Context, id string) error {
	return s.customerRepo.Delete(ctx, id)
}

package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type QuoteLineItemInput struct {
	ProductID string          `json:"product_id"`
	Quantity  int             `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
}

type QuoteService struct {
	quoteRepo     domain.QuoteRepository
	quoteItemRepo domain.QuoteLineItemRepository
	publisher     domain.EventPublisher
}

func NewQuoteService(
	quoteRepo domain.QuoteRepository,
	quoteItemRepo domain.QuoteLineItemRepository,
	publisher domain.EventPublisher,
) *QuoteService {
	return &QuoteService{
		quoteRepo:     quoteRepo,
		quoteItemRepo: quoteItemRepo,
		publisher:     publisher,
	}
}

func (s *QuoteService) CreateQuote(ctx context.Context, customerID, title string, validUntil time.Time, items []QuoteLineItemInput) (*domain.Quote, error) {
	quoteID := utils.NewID("q")
	total := decimal.Zero

	for _, it := range items {
		subtotal := decimal.NewFromInt(int64(it.Quantity)).Mul(it.UnitPrice)
		total = total.Add(subtotal)
	}

	quote := &domain.Quote{
		ID:          quoteID,
		CustomerID:  customerID,
		Title:       title,
		ValidUntil:  validUntil,
		Status:      "DRAFT",
		TotalAmount: total,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.quoteRepo.Create(ctx, quote)
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		itemID := utils.NewID("qi")
		item := &domain.QuoteLineItem{
			ID:        itemID,
			QuoteID:   quoteID,
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			UnitPrice: it.UnitPrice,
		}
		_ = s.quoteItemRepo.Create(ctx, item)
	}

	return quote, nil
}

func (s *QuoteService) GetQuote(ctx context.Context, id string) (*domain.Quote, error) {
	return s.quoteRepo.GetByID(ctx, id)
}

func (s *QuoteService) ListQuotes(ctx context.Context) ([]domain.Quote, error) {
	return s.quoteRepo.List(ctx)
}

func (s *QuoteService) UpdateQuote(ctx context.Context, id string, status string) (*domain.Quote, error) {
	quote, err := s.quoteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	quote.Status = status
	quote.UpdatedAt = time.Now()

	err = s.quoteRepo.Update(ctx, quote)
	if err != nil {
		return nil, err
	}

	return quote, nil
}

func (s *QuoteService) DeleteQuote(ctx context.Context, id string) error {
	return s.quoteRepo.Delete(ctx, id)
}

func (s *QuoteService) SendQuote(ctx context.Context, id string) (*domain.Quote, error) {
	quote, err := s.quoteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	quote.Status = "SENT"
	quote.UpdatedAt = time.Now()
	_ = s.quoteRepo.Update(ctx, quote)

	// Publish Email Sent Event
	emailID := utils.NewID("email")
	if err := s.publisher.Publish(ctx, domain.TopicCrmEmailSent, emailID, domain.EmailSentEvent{
		EmailID:    emailID,
		CampaignID: "quote_dispatch",
		Recipient:  "customer_quote_inbox",
		Timestamp:  time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmEmailSent, err)
	}

	return quote, nil
}

func (s *QuoteService) OpenEmail(ctx context.Context, emailID string) error {
	if err := s.publisher.Publish(ctx, domain.TopicCrmEmailOpened, emailID, domain.EmailOpenedEvent{
		EmailID:   emailID,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmEmailOpened, err)
		return err
	}
	return nil
}

func (s *QuoteService) ClickEmail(ctx context.Context, emailID string, url string) error {
	if err := s.publisher.Publish(ctx, domain.TopicCrmEmailClicked, emailID, domain.EmailClickedEvent{
		EmailID:   emailID,
		URL:       url,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("crm-service", domain.TopicCrmEmailClicked, err)
		return err
	}
	return nil
}

package service_test

import (
	sharedtesting "erp-system/shared/testing"
	"context"
	"testing"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
	"github.com/erp-system/crm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func TestQuoteService_All(t *testing.T) {
	quoteRepo := memory.NewQuoteRepository()
	quoteItemRepo := memory.NewQuoteLineItemRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewQuoteService(quoteRepo, quoteItemRepo, pub)

	ctx := context.Background()

	// 1. Create Quote
	items := []service.QuoteLineItemInput{
		{
			ProductID: "prod_1",
			Quantity:  3,
			UnitPrice: decimal.NewFromInt(40),
		},
	}
	validUntil := time.Now().AddDate(0, 0, 7)
	quote, err := svc.CreateQuote(ctx, "cust_1", "Intro Proposal", validUntil, items)
	if err != nil {
		t.Fatalf("failed to create quote: %v", err)
	}
	if quote.CustomerID != "cust_1" {
		t.Errorf("expected customer ID 'cust_1', got %q", quote.CustomerID)
	}
	expectedTotal := decimal.NewFromInt(120) // 3 * 40
	if !quote.TotalAmount.Equal(expectedTotal) {
		t.Errorf("expected total amount %s, got %s", expectedTotal, quote.TotalAmount)
	}

	// Verify line item was created
	lines, err := quoteItemRepo.ListByQuoteID(ctx, quote.ID)
	if err != nil {
		t.Fatalf("failed to list quote lines: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(lines))
	}
	if lines[0].ProductID != "prod_1" {
		t.Errorf("expected product ID 'prod_1', got %q", lines[0].ProductID)
	}

	// 2. Get Quote
	fetched, err := svc.GetQuote(ctx, quote.ID)
	if err != nil {
		t.Fatalf("failed to get quote: %v", err)
	}
	if fetched.ID != quote.ID {
		t.Errorf("expected quote ID %q, got %q", quote.ID, fetched.ID)
	}

	// 3. List Quotes
	list, err := svc.ListQuotes(ctx)
	if err != nil {
		t.Fatalf("failed to list quotes: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected list length 1, got %d", len(list))
	}

	// 4. Update Quote
	updated, err := svc.UpdateQuote(ctx, quote.ID, "APPROVED")
	if err != nil {
		t.Fatalf("failed to update quote: %v", err)
	}
	if updated.Status != "APPROVED" {
		t.Errorf("expected status 'APPROVED', got %q", updated.Status)
	}

	// 5. Send Quote
	pub.Events = nil
	sent, err := svc.SendQuote(ctx, quote.ID)
	if err != nil {
		t.Fatalf("failed to send quote: %v", err)
	}
	if sent.Status != "SENT" {
		t.Errorf("expected status 'SENT', got %q", sent.Status)
	}
	foundSent := false
	var emailID string
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmEmailSent {
			foundSent = true
			emailID = ev.Key
		}
	}
	if !foundSent {
		t.Errorf("expected email sent event to be published")
	}

	// 6. Open Email
	pub.Events = nil
	err = svc.OpenEmail(ctx, emailID)
	if err != nil {
		t.Fatalf("failed to open email: %v", err)
	}
	foundOpened := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmEmailOpened {
			foundOpened = true
		}
	}
	if !foundOpened {
		t.Errorf("expected email opened event to be published")
	}

	// 7. Click Email
	pub.Events = nil
	err = svc.ClickEmail(ctx, emailID, "http://example.com/quote")
	if err != nil {
		t.Fatalf("failed to click email: %v", err)
	}
	foundClicked := false
	for _, ev := range pub.Events {
		if ev.Topic == domain.TopicCrmEmailClicked {
			foundClicked = true
		}
	}
	if !foundClicked {
		t.Errorf("expected email clicked event to be published")
	}

	// 8. Delete Quote
	err = svc.DeleteQuote(ctx, quote.ID)
	if err != nil {
		t.Fatalf("failed to delete quote: %v", err)
	}

	// Verify deletion
	_, err = svc.GetQuote(ctx, quote.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted quote, got nil")
	}
}

func TestQuoteService_Errors(t *testing.T) {
	quoteRepo := memory.NewQuoteRepository()
	quoteItemRepo := memory.NewQuoteLineItemRepository()
	pub := &sharedtesting.MockPublisher{}
	svc := service.NewQuoteService(quoteRepo, quoteItemRepo, pub)

	ctx := context.Background()

	_, err := svc.UpdateQuote(ctx, "non-existent", "APPROVED")
	if err == nil {
		t.Errorf("expected error updating non-existent quote, got nil")
	}

	_, err = svc.SendQuote(ctx, "non-existent")
	if err == nil {
		t.Errorf("expected error sending non-existent quote, got nil")
	}
}

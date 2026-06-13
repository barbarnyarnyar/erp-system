package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/erp-system/crm-service/internal/business/domain"
	"gorm.io/gorm"
)

// ==========================================
// Customer SQL Repository
// ==========================================

type SQLCustomerRepository struct {
	db *gorm.DB
}

func NewSQLCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &SQLCustomerRepository{db: db}
}

func (r *SQLCustomerRepository) Create(ctx context.Context, customer *domain.CustomerProfile) error {
	db := GetDB(ctx, r.db)
	entity := FromCustomerProfileDomain(customer)
	return db.Create(entity).Error
}

func (r *SQLCustomerRepository) GetByID(ctx context.Context, id string) (*domain.CustomerProfile, error) {
	db := GetDB(ctx, r.db)
	var entity CustomerProfile
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("customer profile not found: %s", id)
		}
		return nil, err
	}
	return ToCustomerProfileDomain(&entity), nil
}

func (r *SQLCustomerRepository) List(ctx context.Context) ([]domain.CustomerProfile, error) {
	db := GetDB(ctx, r.db)
	var entities []CustomerProfile
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.CustomerProfile, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToCustomerProfileDomain(&e))
	}
	return list, nil
}

func (r *SQLCustomerRepository) Update(ctx context.Context, customer *domain.CustomerProfile) error {
	db := GetDB(ctx, r.db)
	entity := FromCustomerProfileDomain(customer)
	return db.Save(entity).Error
}

func (r *SQLCustomerRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&CustomerProfile{}, "id = ?", id).Error
}

// ==========================================
// Lead SQL Repository
// ==========================================

type SQLLeadRepository struct {
	db *gorm.DB
}

func NewSQLLeadRepository(db *gorm.DB) domain.LeadRepository {
	return &SQLLeadRepository{db: db}
}

func (r *SQLLeadRepository) Create(ctx context.Context, lead *domain.Lead) error {
	db := GetDB(ctx, r.db)
	entity := FromLeadDomain(lead)
	return db.Create(entity).Error
}

func (r *SQLLeadRepository) GetByID(ctx context.Context, id string) (*domain.Lead, error) {
	db := GetDB(ctx, r.db)
	var entity Lead
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("lead not found: %s", id)
		}
		return nil, err
	}
	return ToLeadDomain(&entity), nil
}

func (r *SQLLeadRepository) List(ctx context.Context) ([]domain.Lead, error) {
	db := GetDB(ctx, r.db)
	var entities []Lead
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Lead, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToLeadDomain(&e))
	}
	return list, nil
}

func (r *SQLLeadRepository) Update(ctx context.Context, lead *domain.Lead) error {
	db := GetDB(ctx, r.db)
	entity := FromLeadDomain(lead)
	return db.Save(entity).Error
}

func (r *SQLLeadRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&Lead{}, "id = ?", id).Error
}

// ==========================================
// Opportunity SQL Repository
// ==========================================

type SQLOpportunityRepository struct {
	db *gorm.DB
}

func NewSQLOpportunityRepository(db *gorm.DB) domain.OpportunityRepository {
	return &SQLOpportunityRepository{db: db}
}

func (r *SQLOpportunityRepository) Create(ctx context.Context, opp *domain.Opportunity) error {
	db := GetDB(ctx, r.db)
	entity := FromOpportunityDomain(opp)
	return db.Create(entity).Error
}

func (r *SQLOpportunityRepository) GetByID(ctx context.Context, id string) (*domain.Opportunity, error) {
	db := GetDB(ctx, r.db)
	var entity Opportunity
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("opportunity not found: %s", id)
		}
		return nil, err
	}
	return ToOpportunityDomain(&entity), nil
}

func (r *SQLOpportunityRepository) List(ctx context.Context) ([]domain.Opportunity, error) {
	db := GetDB(ctx, r.db)
	var entities []Opportunity
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Opportunity, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToOpportunityDomain(&e))
	}
	return list, nil
}

func (r *SQLOpportunityRepository) Update(ctx context.Context, opp *domain.Opportunity) error {
	db := GetDB(ctx, r.db)
	entity := FromOpportunityDomain(opp)
	return db.Save(entity).Error
}

func (r *SQLOpportunityRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&Opportunity{}, "id = ?", id).Error
}

// ==========================================
// Sales Order SQL Repositories
// ==========================================

type SQLSalesOrderRepository struct {
	db *gorm.DB
}

func NewSQLSalesOrderRepository(db *gorm.DB) domain.SalesOrderRepository {
	return &SQLSalesOrderRepository{db: db}
}

func (r *SQLSalesOrderRepository) Create(ctx context.Context, order *domain.SalesOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromSalesOrderDomain(order)
	return db.Create(entity).Error
}

func (r *SQLSalesOrderRepository) GetByID(ctx context.Context, id string) (*domain.SalesOrder, error) {
	db := GetDB(ctx, r.db)
	var entity SalesOrder
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("sales order not found: %s", id)
		}
		return nil, err
	}
	return ToSalesOrderDomain(&entity), nil
}

func (r *SQLSalesOrderRepository) List(ctx context.Context) ([]domain.SalesOrder, error) {
	db := GetDB(ctx, r.db)
	var entities []SalesOrder
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.SalesOrder, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToSalesOrderDomain(&e))
	}
	return list, nil
}

func (r *SQLSalesOrderRepository) Update(ctx context.Context, order *domain.SalesOrder) error {
	db := GetDB(ctx, r.db)
	entity := FromSalesOrderDomain(order)
	return db.Save(entity).Error
}

func (r *SQLSalesOrderRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&SalesOrder{}, "id = ?", id).Error
}

type SQLSalesOrderLineRepository struct {
	db *gorm.DB
}

func NewSQLSalesOrderLineRepository(db *gorm.DB) domain.SalesOrderLineRepository {
	return &SQLSalesOrderLineRepository{db: db}
}

func (r *SQLSalesOrderLineRepository) Create(ctx context.Context, item *domain.SalesOrderLine) error {
	db := GetDB(ctx, r.db)
	entity := FromSalesOrderLineDomain(item)
	return db.Create(entity).Error
}

func (r *SQLSalesOrderLineRepository) ListByOrderID(ctx context.Context, orderID string) ([]domain.SalesOrderLine, error) {
	db := GetDB(ctx, r.db)
	var entities []SalesOrderLine
	err := db.Find(&entities, "sales_order_id = ?", orderID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.SalesOrderLine, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToSalesOrderLineDomain(&e))
	}
	return list, nil
}

// ==========================================
// Quote SQL Repositories
// ==========================================

type SQLQuoteRepository struct {
	db *gorm.DB
}

func NewSQLQuoteRepository(db *gorm.DB) domain.QuoteRepository {
	return &SQLQuoteRepository{db: db}
}

func (r *SQLQuoteRepository) Create(ctx context.Context, quote *domain.Quote) error {
	db := GetDB(ctx, r.db)
	entity := FromQuoteDomain(quote)
	return db.Create(entity).Error
}

func (r *SQLQuoteRepository) GetByID(ctx context.Context, id string) (*domain.Quote, error) {
	db := GetDB(ctx, r.db)
	var entity Quote
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("quote not found: %s", id)
		}
		return nil, err
	}
	return ToQuoteDomain(&entity), nil
}

func (r *SQLQuoteRepository) List(ctx context.Context) ([]domain.Quote, error) {
	db := GetDB(ctx, r.db)
	var entities []Quote
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Quote, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToQuoteDomain(&e))
	}
	return list, nil
}

func (r *SQLQuoteRepository) Update(ctx context.Context, quote *domain.Quote) error {
	db := GetDB(ctx, r.db)
	entity := FromQuoteDomain(quote)
	return db.Save(entity).Error
}

func (r *SQLQuoteRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&Quote{}, "id = ?", id).Error
}

type SQLQuoteLineItemRepository struct {
	db *gorm.DB
}

func NewSQLQuoteLineItemRepository(db *gorm.DB) domain.QuoteLineItemRepository {
	return &SQLQuoteLineItemRepository{db: db}
}

func (r *SQLQuoteLineItemRepository) Create(ctx context.Context, item *domain.QuoteLineItem) error {
	db := GetDB(ctx, r.db)
	entity := FromQuoteLineItemDomain(item)
	return db.Create(entity).Error
}

func (r *SQLQuoteLineItemRepository) ListByQuoteID(ctx context.Context, quoteID string) ([]domain.QuoteLineItem, error) {
	db := GetDB(ctx, r.db)
	var entities []QuoteLineItem
	err := db.Find(&entities, "quote_id = ?", quoteID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.QuoteLineItem, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToQuoteLineItemDomain(&e))
	}
	return list, nil
}

// ==========================================
// Price Book SQL Repositories
// ==========================================

type SQLPriceBookHeaderRepository struct {
	db *gorm.DB
}

func NewSQLPriceBookHeaderRepository(db *gorm.DB) domain.PriceBookHeaderRepository {
	return &SQLPriceBookHeaderRepository{db: db}
}

func (r *SQLPriceBookHeaderRepository) Create(ctx context.Context, priceBook *domain.PriceBookHeader) error {
	db := GetDB(ctx, r.db)
	entity := FromPriceBookHeaderDomain(priceBook)
	return db.Create(entity).Error
}

func (r *SQLPriceBookHeaderRepository) GetByID(ctx context.Context, id string) (*domain.PriceBookHeader, error) {
	db := GetDB(ctx, r.db)
	var entity PriceBookHeader
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("price book not found: %s", id)
		}
		return nil, err
	}
	return ToPriceBookHeaderDomain(&entity), nil
}

func (r *SQLPriceBookHeaderRepository) List(ctx context.Context) ([]domain.PriceBookHeader, error) {
	db := GetDB(ctx, r.db)
	var entities []PriceBookHeader
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.PriceBookHeader, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToPriceBookHeaderDomain(&e))
	}
	return list, nil
}

func (r *SQLPriceBookHeaderRepository) Update(ctx context.Context, priceBook *domain.PriceBookHeader) error {
	db := GetDB(ctx, r.db)
	entity := FromPriceBookHeaderDomain(priceBook)
	return db.Save(entity).Error
}

func (r *SQLPriceBookHeaderRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&PriceBookHeader{}, "id = ?", id).Error
}

type SQLPriceBookEntryRepository struct {
	db *gorm.DB
}

func NewSQLPriceBookEntryRepository(db *gorm.DB) domain.PriceBookEntryRepository {
	return &SQLPriceBookEntryRepository{db: db}
}

func (r *SQLPriceBookEntryRepository) Create(ctx context.Context, item *domain.PriceBookEntry) error {
	db := GetDB(ctx, r.db)
	entity := FromPriceBookEntryDomain(item)
	return db.Create(entity).Error
}

func (r *SQLPriceBookEntryRepository) ListByPriceBookID(ctx context.Context, priceBookID string) ([]domain.PriceBookEntry, error) {
	db := GetDB(ctx, r.db)
	var entities []PriceBookEntry
	err := db.Find(&entities, "price_book_id = ?", priceBookID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.PriceBookEntry, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToPriceBookEntryDomain(&e))
	}
	return list, nil
}

// ==========================================
// Service Ticket SQL Repository
// ==========================================

type SQLServiceTicketRepository struct {
	db *gorm.DB
}

func NewSQLServiceTicketRepository(db *gorm.DB) domain.ServiceTicketRepository {
	return &SQLServiceTicketRepository{db: db}
}

func (r *SQLServiceTicketRepository) Create(ctx context.Context, ticket *domain.ServiceTicket) error {
	db := GetDB(ctx, r.db)
	entity := FromServiceTicketDomain(ticket)
	return db.Create(entity).Error
}

func (r *SQLServiceTicketRepository) GetByID(ctx context.Context, id string) (*domain.ServiceTicket, error) {
	db := GetDB(ctx, r.db)
	var entity ServiceTicket
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("service ticket not found: %s", id)
		}
		return nil, err
	}
	return ToServiceTicketDomain(&entity), nil
}

func (r *SQLServiceTicketRepository) List(ctx context.Context) ([]domain.ServiceTicket, error) {
	db := GetDB(ctx, r.db)
	var entities []ServiceTicket
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.ServiceTicket, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToServiceTicketDomain(&e))
	}
	return list, nil
}

func (r *SQLServiceTicketRepository) Update(ctx context.Context, ticket *domain.ServiceTicket) error {
	db := GetDB(ctx, r.db)
	entity := FromServiceTicketDomain(ticket)
	return db.Save(entity).Error
}

func (r *SQLServiceTicketRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&ServiceTicket{}, "id = ?", id).Error
}

// ==========================================
// Campaign SQL Repository
// ==========================================

type SQLCampaignRepository struct {
	db *gorm.DB
}

func NewSQLCampaignRepository(db *gorm.DB) domain.CampaignRepository {
	return &SQLCampaignRepository{db: db}
}

func (r *SQLCampaignRepository) Create(ctx context.Context, campaign *domain.Campaign) error {
	db := GetDB(ctx, r.db)
	entity := FromCampaignDomain(campaign)
	return db.Create(entity).Error
}

func (r *SQLCampaignRepository) GetByID(ctx context.Context, id string) (*domain.Campaign, error) {
	db := GetDB(ctx, r.db)
	var entity Campaign
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("campaign not found: %s", id)
		}
		return nil, err
	}
	return ToCampaignDomain(&entity), nil
}

func (r *SQLCampaignRepository) List(ctx context.Context) ([]domain.Campaign, error) {
	db := GetDB(ctx, r.db)
	var entities []Campaign
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.Campaign, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToCampaignDomain(&e))
	}
	return list, nil
}

func (r *SQLCampaignRepository) Update(ctx context.Context, campaign *domain.Campaign) error {
	db := GetDB(ctx, r.db)
	entity := FromCampaignDomain(campaign)
	return db.Save(entity).Error
}

func (r *SQLCampaignRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&Campaign{}, "id = ?", id).Error
}

// ==========================================
// OpportunityStageHistory SQL Repository
// ==========================================

type SQLOpportunityStageHistoryRepository struct {
	db *gorm.DB
}

func NewSQLOpportunityStageHistoryRepository(db *gorm.DB) domain.OpportunityStageHistoryRepository {
	return &SQLOpportunityStageHistoryRepository{db: db}
}

func (r *SQLOpportunityStageHistoryRepository) Create(ctx context.Context, osh *domain.OpportunityStageHistory) error {
	db := GetDB(ctx, r.db)
	entity := FromOpportunityStageHistoryDomain(osh)
	return db.Create(entity).Error
}

func (r *SQLOpportunityStageHistoryRepository) ListByOpportunityID(ctx context.Context, opportunityID string) ([]domain.OpportunityStageHistory, error) {
	db := GetDB(ctx, r.db)
	var entities []OpportunityStageHistory
	err := db.Order("changed_at asc").Find(&entities, "opportunity_id = ?", opportunityID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.OpportunityStageHistory, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToOpportunityStageHistoryDomain(&e))
	}
	return list, nil
}

// ==========================================
// CustomerInteraction SQL Repository
// ==========================================

type SQLCustomerInteractionRepository struct {
	db *gorm.DB
}

func NewSQLCustomerInteractionRepository(db *gorm.DB) domain.CustomerInteractionRepository {
	return &SQLCustomerInteractionRepository{db: db}
}

func (r *SQLCustomerInteractionRepository) Create(ctx context.Context, ci *domain.CustomerInteraction) error {
	db := GetDB(ctx, r.db)
	entity := FromCustomerInteractionDomain(ci)
	return db.Create(entity).Error
}

func (r *SQLCustomerInteractionRepository) GetByID(ctx context.Context, id string) (*domain.CustomerInteraction, error) {
	db := GetDB(ctx, r.db)
	var entity CustomerInteraction
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("customer interaction not found: %s", id)
		}
		return nil, err
	}
	return ToCustomerInteractionDomain(&entity), nil
}

func (r *SQLCustomerInteractionRepository) ListByCustomerID(ctx context.Context, customerID string) ([]domain.CustomerInteraction, error) {
	db := GetDB(ctx, r.db)
	var entities []CustomerInteraction
	err := db.Find(&entities, "customer_id = ?", customerID).Error
	if err != nil {
		return nil, err
	}
	list := make([]domain.CustomerInteraction, 0, len(entities))
	for _, e := range entities {
		list = append(list, *ToCustomerInteractionDomain(&e))
	}
	return list, nil
}

func (r *SQLCustomerInteractionRepository) Delete(ctx context.Context, id string) error {
	db := GetDB(ctx, r.db)
	return db.Delete(&CustomerInteraction{}, "id = ?", id).Error
}

package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/erp-system/crm-service/internal/business/domain"
)

// ==========================================
// Customer Memory Repository
// ==========================================

type CustomerRepository struct {
	mu        sync.RWMutex
	customers map[string]domain.CustomerProfile
}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{
		customers: make(map[string]domain.CustomerProfile),
	}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *domain.CustomerProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.customers[customer.ID] = *customer
	return nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id string) (*domain.CustomerProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.customers[id]
	if !ok {
		return nil, fmt.Errorf("customer profile not found: %s", id)
	}
	return &c, nil
}

func (r *CustomerRepository) List(ctx context.Context) ([]domain.CustomerProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.CustomerProfile, 0, len(r.customers))
	for _, c := range r.customers {
		list = append(list, c)
	}
	return list, nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *domain.CustomerProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.customers[customer.ID]; !ok {
		return fmt.Errorf("customer profile not found: %s", customer.ID)
	}
	r.customers[customer.ID] = *customer
	return nil
}

func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.customers, id)
	return nil
}

// ==========================================
// Lead Memory Repository
// ==========================================

type LeadRepository struct {
	mu    sync.RWMutex
	leads map[string]domain.Lead
}

func NewLeadRepository() *LeadRepository {
	return &LeadRepository{
		leads: make(map[string]domain.Lead),
	}
}

func (r *LeadRepository) Create(ctx context.Context, lead *domain.Lead) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leads[lead.ID] = *lead
	return nil
}

func (r *LeadRepository) GetByID(ctx context.Context, id string) (*domain.Lead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	l, ok := r.leads[id]
	if !ok {
		return nil, fmt.Errorf("lead not found: %s", id)
	}
	return &l, nil
}

func (r *LeadRepository) List(ctx context.Context) ([]domain.Lead, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Lead, 0, len(r.leads))
	for _, l := range r.leads {
		list = append(list, l)
	}
	return list, nil
}

func (r *LeadRepository) Update(ctx context.Context, lead *domain.Lead) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.leads[lead.ID]; !ok {
		return fmt.Errorf("lead not found: %s", lead.ID)
	}
	r.leads[lead.ID] = *lead
	return nil
}

func (r *LeadRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.leads, id)
	return nil
}

// ==========================================
// Opportunity Memory Repository
// ==========================================

type OpportunityRepository struct {
	mu            sync.RWMutex
	opportunities map[string]domain.Opportunity
}

func NewOpportunityRepository() *OpportunityRepository {
	return &OpportunityRepository{
		opportunities: make(map[string]domain.Opportunity),
	}
}

func (r *OpportunityRepository) Create(ctx context.Context, opp *domain.Opportunity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.opportunities[opp.ID] = *opp
	return nil
}

func (r *OpportunityRepository) GetByID(ctx context.Context, id string) (*domain.Opportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.opportunities[id]
	if !ok {
		return nil, fmt.Errorf("opportunity not found: %s", id)
	}
	return &o, nil
}

func (r *OpportunityRepository) List(ctx context.Context) ([]domain.Opportunity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Opportunity, 0, len(r.opportunities))
	for _, o := range r.opportunities {
		list = append(list, o)
	}
	return list, nil
}

func (r *OpportunityRepository) Update(ctx context.Context, opp *domain.Opportunity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.opportunities[opp.ID]; !ok {
		return fmt.Errorf("opportunity not found: %s", opp.ID)
	}
	r.opportunities[opp.ID] = *opp
	return nil
}

func (r *OpportunityRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.opportunities, id)
	return nil
}

// ==========================================
// Sales Order Memory Repositories
// ==========================================

type SalesOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]domain.SalesOrder
}

func NewSalesOrderRepository() *SalesOrderRepository {
	return &SalesOrderRepository{
		orders: make(map[string]domain.SalesOrder),
	}
}

func (r *SalesOrderRepository) Create(ctx context.Context, order *domain.SalesOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = *order
	return nil
}

func (r *SalesOrderRepository) GetByID(ctx context.Context, id string) (*domain.SalesOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[id]
	if !ok {
		return nil, fmt.Errorf("sales order not found: %s", id)
	}
	return &o, nil
}

func (r *SalesOrderRepository) List(ctx context.Context) ([]domain.SalesOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.SalesOrder, 0, len(r.orders))
	for _, o := range r.orders {
		list = append(list, o)
	}
	return list, nil
}

func (r *SalesOrderRepository) Update(ctx context.Context, order *domain.SalesOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.orders[order.ID]; !ok {
		return fmt.Errorf("sales order not found: %s", order.ID)
	}
	r.orders[order.ID] = *order
	return nil
}

func (r *SalesOrderRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.orders, id)
	return nil
}

type SalesOrderLineRepository struct {
	mu    sync.RWMutex
	items map[string]domain.SalesOrderLine
}

func NewSalesOrderLineRepository() *SalesOrderLineRepository {
	return &SalesOrderLineRepository{
		items: make(map[string]domain.SalesOrderLine),
	}
}

func (r *SalesOrderLineRepository) Create(ctx context.Context, item *domain.SalesOrderLine) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = *item
	return nil
}

func (r *SalesOrderLineRepository) ListByOrderID(ctx context.Context, orderID string) ([]domain.SalesOrderLine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.SalesOrderLine
	for _, it := range r.items {
		if it.SalesOrderID == orderID {
			list = append(list, it)
		}
	}
	return list, nil
}

// ==========================================
// Quote Memory Repositories
// ==========================================

type QuoteRepository struct {
	mu     sync.RWMutex
	quotes map[string]domain.Quote
}

func NewQuoteRepository() *QuoteRepository {
	return &QuoteRepository{
		quotes: make(map[string]domain.Quote),
	}
}

func (r *QuoteRepository) Create(ctx context.Context, quote *domain.Quote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quotes[quote.ID] = *quote
	return nil
}

func (r *QuoteRepository) GetByID(ctx context.Context, id string) (*domain.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q, ok := r.quotes[id]
	if !ok {
		return nil, fmt.Errorf("quote not found: %s", id)
	}
	return &q, nil
}

func (r *QuoteRepository) List(ctx context.Context) ([]domain.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Quote, 0, len(r.quotes))
	for _, q := range r.quotes {
		list = append(list, q)
	}
	return list, nil
}

func (r *QuoteRepository) Update(ctx context.Context, quote *domain.Quote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.quotes[quote.ID]; !ok {
		return fmt.Errorf("quote not found: %s", quote.ID)
	}
	r.quotes[quote.ID] = *quote
	return nil
}

func (r *QuoteRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.quotes, id)
	return nil
}

type QuoteLineItemRepository struct {
	mu    sync.RWMutex
	items map[string]domain.QuoteLineItem
}

func NewQuoteLineItemRepository() *QuoteLineItemRepository {
	return &QuoteLineItemRepository{
		items: make(map[string]domain.QuoteLineItem),
	}
}

func (r *QuoteLineItemRepository) Create(ctx context.Context, item *domain.QuoteLineItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = *item
	return nil
}

func (r *QuoteLineItemRepository) ListByQuoteID(ctx context.Context, quoteID string) ([]domain.QuoteLineItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.QuoteLineItem
	for _, it := range r.items {
		if it.QuoteID == quoteID {
			list = append(list, it)
		}
	}
	return list, nil
}

// ==========================================
// Price Book Memory Repositories
// ==========================================

type PriceBookHeaderRepository struct {
	mu         sync.RWMutex
	priceBooks map[string]domain.PriceBookHeader
}

func NewPriceBookHeaderRepository() *PriceBookHeaderRepository {
	return &PriceBookHeaderRepository{
		priceBooks: make(map[string]domain.PriceBookHeader),
	}
}

func (r *PriceBookHeaderRepository) Create(ctx context.Context, priceBook *domain.PriceBookHeader) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.priceBooks[priceBook.ID] = *priceBook
	return nil
}

func (r *PriceBookHeaderRepository) GetByID(ctx context.Context, id string) (*domain.PriceBookHeader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pl, ok := r.priceBooks[id]
	if !ok {
		return nil, fmt.Errorf("price book not found: %s", id)
	}
	return &pl, nil
}

func (r *PriceBookHeaderRepository) List(ctx context.Context) ([]domain.PriceBookHeader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.PriceBookHeader, 0, len(r.priceBooks))
	for _, pl := range r.priceBooks {
		list = append(list, pl)
	}
	return list, nil
}

func (r *PriceBookHeaderRepository) Update(ctx context.Context, priceBook *domain.PriceBookHeader) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.priceBooks[priceBook.ID]; !ok {
		return fmt.Errorf("price book not found: %s", priceBook.ID)
	}
	r.priceBooks[priceBook.ID] = *priceBook
	return nil
}

func (r *PriceBookHeaderRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.priceBooks, id)
	return nil
}

type PriceBookEntryRepository struct {
	mu    sync.RWMutex
	items map[string]domain.PriceBookEntry
}

func NewPriceBookEntryRepository() *PriceBookEntryRepository {
	return &PriceBookEntryRepository{
		items: make(map[string]domain.PriceBookEntry),
	}
}

func (r *PriceBookEntryRepository) Create(ctx context.Context, item *domain.PriceBookEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = *item
	return nil
}

func (r *PriceBookEntryRepository) ListByPriceBookID(ctx context.Context, priceBookID string) ([]domain.PriceBookEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []domain.PriceBookEntry
	for _, it := range r.items {
		if it.PriceBookID == priceBookID {
			list = append(list, it)
		}
	}
	return list, nil
}

// ==========================================
// Service Ticket Memory Repository
// ==========================================

type ServiceTicketRepository struct {
	mu      sync.RWMutex
	tickets map[string]domain.ServiceTicket
}

func NewServiceTicketRepository() *ServiceTicketRepository {
	return &ServiceTicketRepository{
		tickets: make(map[string]domain.ServiceTicket),
	}
}

func (r *ServiceTicketRepository) Create(ctx context.Context, ticket *domain.ServiceTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tickets[ticket.ID] = *ticket
	return nil
}

func (r *ServiceTicketRepository) GetByID(ctx context.Context, id string) (*domain.ServiceTicket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tickets[id]
	if !ok {
		return nil, fmt.Errorf("service ticket not found: %s", id)
	}
	return &t, nil
}

func (r *ServiceTicketRepository) List(ctx context.Context) ([]domain.ServiceTicket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.ServiceTicket, 0, len(r.tickets))
	for _, t := range r.tickets {
		list = append(list, t)
	}
	return list, nil
}

func (r *ServiceTicketRepository) Update(ctx context.Context, ticket *domain.ServiceTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tickets[ticket.ID]; !ok {
		return fmt.Errorf("service ticket not found: %s", ticket.ID)
	}
	r.tickets[ticket.ID] = *ticket
	return nil
}

func (r *ServiceTicketRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tickets, id)
	return nil
}

// ==========================================
// Campaign Memory Repository
// ==========================================

type CampaignRepository struct {
	mu        sync.RWMutex
	campaigns map[string]domain.Campaign
}

func NewCampaignRepository() *CampaignRepository {
	return &CampaignRepository{
		campaigns: make(map[string]domain.Campaign),
	}
}

func (r *CampaignRepository) Create(ctx context.Context, campaign *domain.Campaign) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.campaigns[campaign.ID] = *campaign
	return nil
}

func (r *CampaignRepository) GetByID(ctx context.Context, id string) (*domain.Campaign, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.campaigns[id]
	if !ok {
		return nil, fmt.Errorf("campaign not found: %s", id)
	}
	return &c, nil
}

func (r *CampaignRepository) List(ctx context.Context) ([]domain.Campaign, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.Campaign, 0, len(r.campaigns))
	for _, c := range r.campaigns {
		list = append(list, c)
	}
	return list, nil
}

func (r *CampaignRepository) Update(ctx context.Context, campaign *domain.Campaign) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.campaigns[campaign.ID]; !ok {
		return fmt.Errorf("campaign not found: %s", campaign.ID)
	}
	r.campaigns[campaign.ID] = *campaign
	return nil
}

func (r *CampaignRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.campaigns, id)
	return nil
}

// ==========================================
// OpportunityStageHistory Memory Repository
// ==========================================

type OpportunityStageHistoryRepository struct {
	mu        sync.RWMutex
	histories map[string]domain.OpportunityStageHistory
}

func NewOpportunityStageHistoryRepository() *OpportunityStageHistoryRepository {
	return &OpportunityStageHistoryRepository{
		histories: make(map[string]domain.OpportunityStageHistory),
	}
}

func (r *OpportunityStageHistoryRepository) Create(ctx context.Context, osh *domain.OpportunityStageHistory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.histories[osh.ID] = *osh
	return nil
}

func (r *OpportunityStageHistoryRepository) ListByOpportunityID(ctx context.Context, opportunityID string) ([]domain.OpportunityStageHistory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.OpportunityStageHistory, 0)
	for _, h := range r.histories {
		if h.OpportunityID == opportunityID {
			list = append(list, h)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ChangedAt.Before(list[j].ChangedAt)
	})
	return list, nil
}

// ==========================================
// CustomerInteraction Memory Repository
// ==========================================

type CustomerInteractionRepository struct {
	mu           sync.RWMutex
	interactions map[string]domain.CustomerInteraction
}

func NewCustomerInteractionRepository() *CustomerInteractionRepository {
	return &CustomerInteractionRepository{
		interactions: make(map[string]domain.CustomerInteraction),
	}
}

func (r *CustomerInteractionRepository) Create(ctx context.Context, ci *domain.CustomerInteraction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.interactions[ci.ID] = *ci
	return nil
}

func (r *CustomerInteractionRepository) GetByID(ctx context.Context, id string) (*domain.CustomerInteraction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ci, ok := r.interactions[id]
	if !ok {
		return nil, fmt.Errorf("customer interaction not found: %s", id)
	}
	return &ci, nil
}

func (r *CustomerInteractionRepository) ListByCustomerID(ctx context.Context, customerID string) ([]domain.CustomerInteraction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]domain.CustomerInteraction, 0)
	for _, ci := range r.interactions {
		if ci.CustomerID == customerID {
			list = append(list, ci)
		}
	}
	return list, nil
}

func (r *CustomerInteractionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.interactions, id)
	return nil
}

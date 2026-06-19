package sql

import (
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CustomerProfile struct {
	ID                 string         `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID      string         `gorm:"type:varchar(255);index"`
	CustomerCode       string         `gorm:"type:varchar(255);uniqueIndex:idx_cust_code"`
	CompanyName        string         `gorm:"type:varchar(255)"`
	AccountManagerHrID string         `gorm:"type:varchar(255);index"`
	Status             string         `gorm:"type:varchar(50)"`
	CreditLimit        decimal.Decimal `gorm:"type:numeric(18,4)"`
	Currency           string         `gorm:"type:varchar(10)"`
	ContactName        *string        `gorm:"type:varchar(255)"`
	Email              *string        `gorm:"type:varchar(255)"`
	Phone              *string        `gorm:"type:varchar(50)"`
	Category           *string        `gorm:"type:varchar(100)"`
	ParentCustomerID   *string        `gorm:"type:varchar(255);index"`
	Version            int            `gorm:"type:int;default:1"`
	CreatedAt          time.Time      `gorm:"index"`
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

func (CustomerProfile) TableName() string {
	return "crm_customers"
}

func ToCustomerProfileDomain(c *CustomerProfile) *domain.CustomerProfile {
	if c == nil {
		return nil
	}
	var delAt *time.Time
	if c.DeletedAt.Valid {
		delAt = &c.DeletedAt.Time
	}
	return &domain.CustomerProfile{
		ID:                 c.ID,
		LegalEntityID:      c.LegalEntityID,
		CustomerCode:       c.CustomerCode,
		CompanyName:        c.CompanyName,
		AccountManagerHrID: c.AccountManagerHrID,
		Status:             domain.CustomerStatus(c.Status),
		CreditLimit:        c.CreditLimit,
		Currency:           c.Currency,
		ContactName:        c.ContactName,
		Email:              c.Email,
		Phone:              c.Phone,
		Category:           c.Category,
		ParentCustomerID:   c.ParentCustomerID,
		Version:            c.Version,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
		DeletedAt:          delAt,
	}
}

func FromCustomerProfileDomain(c *domain.CustomerProfile) *CustomerProfile {
	if c == nil {
		return nil
	}
	var delAt gorm.DeletedAt
	if c.DeletedAt != nil {
		delAt = gorm.DeletedAt{Time: *c.DeletedAt, Valid: true}
	}
	return &CustomerProfile{
		ID:                 c.ID,
		LegalEntityID:      c.LegalEntityID,
		CustomerCode:       c.CustomerCode,
		CompanyName:        c.CompanyName,
		AccountManagerHrID: c.AccountManagerHrID,
		Status:             string(c.Status),
		CreditLimit:        c.CreditLimit,
		Currency:           c.Currency,
		ContactName:        c.ContactName,
		Email:              c.Email,
		Phone:              c.Phone,
		Category:           c.Category,
		ParentCustomerID:   c.ParentCustomerID,
		Version:            c.Version,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
		DeletedAt:          delAt,
	}
}

type Lead struct {
	ID         string    `gorm:"primaryKey;type:varchar(255)"`
	FirstName  string    `gorm:"type:varchar(255)"`
	LastName   string    `gorm:"type:varchar(255)"`
	Company    string    `gorm:"type:varchar(255)"`
	Email      string    `gorm:"type:varchar(255)"`
	Phone      string    `gorm:"type:varchar(50)"`
	Status     string    `gorm:"type:varchar(50)"`
	Score      int       `gorm:"type:int"`
	Source     string    `gorm:"type:varchar(100)"`
	CampaignID *string   `gorm:"type:varchar(255);index"`
	CreatedAt  time.Time `gorm:"index"`
	UpdatedAt  time.Time
}

func (Lead) TableName() string {
	return "crm_leads"
}

func ToLeadDomain(l *Lead) *domain.Lead {
	if l == nil {
		return nil
	}
	return &domain.Lead{
		ID:         l.ID,
		FirstName:  l.FirstName,
		LastName:   l.LastName,
		Company:    l.Company,
		Email:      l.Email,
		Phone:      l.Phone,
		Status:     l.Status,
		Score:      l.Score,
		Source:     l.Source,
		CampaignID: l.CampaignID,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}
}

func FromLeadDomain(l *domain.Lead) *Lead {
	if l == nil {
		return nil
	}
	return &Lead{
		ID:         l.ID,
		FirstName:  l.FirstName,
		LastName:   l.LastName,
		Company:    l.Company,
		Email:      l.Email,
		Phone:      l.Phone,
		Status:     l.Status,
		Score:      l.Score,
		Source:     l.Source,
		CampaignID: l.CampaignID,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}
}

type Opportunity struct {
	ID          string          `gorm:"primaryKey;type:varchar(255)"`
	CustomerID  string          `gorm:"type:varchar(255);index"`
	Title       string          `gorm:"type:varchar(255)"`
	Value       decimal.Decimal `gorm:"type:numeric(18,4)"`
	Status      string          `gorm:"type:varchar(50)"`
	Stage       string          `gorm:"type:varchar(50)"`
	Probability decimal.Decimal `gorm:"type:numeric(5,4)"`
	CreatedAt   time.Time       `gorm:"index"`
	UpdatedAt   time.Time
}

func (Opportunity) TableName() string {
	return "crm_opportunities"
}

func ToOpportunityDomain(o *Opportunity) *domain.Opportunity {
	if o == nil {
		return nil
	}
	return &domain.Opportunity{
		CustomerID:  o.CustomerID,
		ID:          o.ID,
		Title:       o.Title,
		Value:       o.Value,
		Status:      o.Status,
		Stage:       domain.OpportunityStage(o.Stage),
		Probability: o.Probability,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

func FromOpportunityDomain(o *domain.Opportunity) *Opportunity {
	if o == nil {
		return nil
	}
	return &Opportunity{
		CustomerID:  o.CustomerID,
		ID:          o.ID,
		Title:       o.Title,
		Value:       o.Value,
		Status:      o.Status,
		Stage:       string(o.Stage),
		Probability: o.Probability,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

type OpportunityStageHistory struct {
	ID            string `gorm:"primaryKey;type:varchar(255)"`
	OpportunityID string `gorm:"type:varchar(255);index"`
	Stage         string `gorm:"type:varchar(50)"`
	ChangedAt     time.Time
	ChangedBy     string `gorm:"type:varchar(255)"`
}

func (OpportunityStageHistory) TableName() string {
	return "crm_opportunity_stage_history"
}

func ToOpportunityStageHistoryDomain(h *OpportunityStageHistory) *domain.OpportunityStageHistory {
	if h == nil {
		return nil
	}
	return &domain.OpportunityStageHistory{
		ID:            h.ID,
		OpportunityID: h.OpportunityID,
		Stage:         domain.OpportunityStage(h.Stage),
		ChangedAt:     h.ChangedAt,
		ChangedBy:     h.ChangedBy,
	}
}

func FromOpportunityStageHistoryDomain(h *domain.OpportunityStageHistory) *OpportunityStageHistory {
	if h == nil {
		return nil
	}
	return &OpportunityStageHistory{
		ID:            h.ID,
		OpportunityID: h.OpportunityID,
		Stage:         string(h.Stage),
		ChangedAt:     h.ChangedAt,
		ChangedBy:     h.ChangedBy,
	}
}

type SalesOrder struct {
	ID              string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID   string          `gorm:"type:varchar(255);index"`
	CustomerID      string          `gorm:"type:varchar(255);index"`
	PriceBookID     string          `gorm:"type:varchar(255);index"`
	OrderNumber     string          `gorm:"type:varchar(255);uniqueIndex:idx_so_num"`
	Status          string          `gorm:"type:varchar(50)"`
	TotalGrossValue decimal.Decimal `gorm:"type:numeric(18,4)"`
	TotalTaxValue   decimal.Decimal `gorm:"type:numeric(18,4)"`
	Version         int             `gorm:"type:int;default:1"`
	CreatedAt       time.Time       `gorm:"index"`
	UpdatedAt       time.Time
}

func (SalesOrder) TableName() string {
	return "crm_sales_orders"
}

func ToSalesOrderDomain(s *SalesOrder) *domain.SalesOrder {
	if s == nil {
		return nil
	}
	return &domain.SalesOrder{
		ID:              s.ID,
		LegalEntityID:   s.LegalEntityID,
		CustomerID:      s.CustomerID,
		PriceBookID:     s.PriceBookID,
		OrderNumber:     s.OrderNumber,
		Status:          domain.SalesOrderState(s.Status),
		TotalGrossValue: s.TotalGrossValue,
		TotalTaxValue:   s.TotalTaxValue,
		Version:         s.Version,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

func FromSalesOrderDomain(s *domain.SalesOrder) *SalesOrder {
	if s == nil {
		return nil
	}
	return &SalesOrder{
		ID:              s.ID,
		LegalEntityID:   s.LegalEntityID,
		CustomerID:      s.CustomerID,
		PriceBookID:     s.PriceBookID,
		OrderNumber:     s.OrderNumber,
		Status:          string(s.Status),
		TotalGrossValue: s.TotalGrossValue,
		TotalTaxValue:   s.TotalTaxValue,
		Version:         s.Version,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

type SalesOrderLine struct {
	ID                     string          `gorm:"primaryKey;type:varchar(255)"`
	SalesOrderID           string          `gorm:"type:varchar(255);index"`
	MaterialID             string          `gorm:"type:varchar(255);index"`
	LineSequence           int             `gorm:"type:int"`
	QuantityOrdered        decimal.Decimal `gorm:"type:numeric(14,4)"`
	QuantityShipped        decimal.Decimal `gorm:"type:numeric(14,4)"`
	UnitSellPrice          decimal.Decimal `gorm:"type:numeric(18,4)"`
	DiscountApplied        decimal.Decimal `gorm:"type:numeric(5,4)"`
	NetLineAmount          decimal.Decimal `gorm:"type:numeric(18,4)"`
	AppliedStrategyVersion *int            `gorm:"type:int"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (SalesOrderLine) TableName() string {
	return "crm_sales_order_lines"
}

func ToSalesOrderLineDomain(l *SalesOrderLine) *domain.SalesOrderLine {
	if l == nil {
		return nil
	}
	return &domain.SalesOrderLine{
		ID:                     l.ID,
		SalesOrderID:           l.SalesOrderID,
		MaterialID:             l.MaterialID,
		LineSequence:           l.LineSequence,
		QuantityOrdered:        l.QuantityOrdered,
		QuantityShipped:        l.QuantityShipped,
		UnitSellPrice:          l.UnitSellPrice,
		DiscountApplied:        l.DiscountApplied,
		NetLineAmount:          l.NetLineAmount,
		AppliedStrategyVersion: l.AppliedStrategyVersion,
		CreatedAt:              l.CreatedAt,
		UpdatedAt:              l.UpdatedAt,
	}
}

func FromSalesOrderLineDomain(l *domain.SalesOrderLine) *SalesOrderLine {
	if l == nil {
		return nil
	}
	return &SalesOrderLine{
		ID:                     l.ID,
		SalesOrderID:           l.SalesOrderID,
		MaterialID:             l.MaterialID,
		LineSequence:           l.LineSequence,
		QuantityOrdered:        l.QuantityOrdered,
		QuantityShipped:        l.QuantityShipped,
		UnitSellPrice:          l.UnitSellPrice,
		DiscountApplied:        l.DiscountApplied,
		NetLineAmount:          l.NetLineAmount,
		AppliedStrategyVersion: l.AppliedStrategyVersion,
		CreatedAt:              l.CreatedAt,
		UpdatedAt:              l.UpdatedAt,
	}
}

type Quote struct {
	ID             string          `gorm:"primaryKey;type:varchar(255)"`
	CustomerID     string          `gorm:"type:varchar(255);index"`
	Title          string          `gorm:"type:varchar(255)"`
	ValidUntil     time.Time       `gorm:"index"`
	Status         string          `gorm:"type:varchar(50)"`
	TotalAmount    decimal.Decimal `gorm:"type:numeric(18,4)"`
	OpportunityID  *string         `gorm:"type:varchar(255);index"`
	CreatedAt      time.Time       `gorm:"index"`
	UpdatedAt      time.Time
}

func (Quote) TableName() string {
	return "crm_quotes"
}

func ToQuoteDomain(q *Quote) *domain.Quote {
	if q == nil {
		return nil
	}
	return &domain.Quote{
		CustomerID:    q.CustomerID,
		ID:            q.ID,
		Title:         q.Title,
		ValidUntil:    q.ValidUntil,
		Status:        q.Status,
		TotalAmount:   q.TotalAmount,
		OpportunityID: q.OpportunityID,
		CreatedAt:     q.CreatedAt,
		UpdatedAt:     q.UpdatedAt,
	}
}

func FromQuoteDomain(q *domain.Quote) *Quote {
	if q == nil {
		return nil
	}
	return &Quote{
		CustomerID:    q.CustomerID,
		ID:            q.ID,
		Title:         q.Title,
		ValidUntil:    q.ValidUntil,
		Status:        q.Status,
		TotalAmount:   q.TotalAmount,
		OpportunityID: q.OpportunityID,
		CreatedAt:     q.CreatedAt,
		UpdatedAt:     q.UpdatedAt,
	}
}

type QuoteLineItem struct {
	ID        string          `gorm:"primaryKey;type:varchar(255)"`
	QuoteID   string          `gorm:"type:varchar(255);index"`
	ProductID string          `gorm:"type:varchar(255);index"`
	Quantity  int             `gorm:"type:int"`
	UnitPrice decimal.Decimal `gorm:"type:numeric(18,4)"`
}

func (QuoteLineItem) TableName() string {
	return "crm_quote_line_items"
}

func ToQuoteLineItemDomain(l *QuoteLineItem) *domain.QuoteLineItem {
	if l == nil {
		return nil
	}
	return &domain.QuoteLineItem{
		ID:        l.ID,
		QuoteID:   l.QuoteID,
		ProductID: l.ProductID,
		Quantity:  l.Quantity,
		UnitPrice: l.UnitPrice,
	}
}

func FromQuoteLineItemDomain(l *domain.QuoteLineItem) *QuoteLineItem {
	if l == nil {
		return nil
	}
	return &QuoteLineItem{
		ID:        l.ID,
		QuoteID:   l.QuoteID,
		ProductID: l.ProductID,
		Quantity:  l.Quantity,
		UnitPrice: l.UnitPrice,
	}
}

type PriceBookHeader struct {
	ID            string    `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID string    `gorm:"type:varchar(255);index"`
	PriceBookCode string    `gorm:"type:varchar(255);uniqueIndex:idx_pb_code"`
	Name          string    `gorm:"type:varchar(255)"`
	Type          string    `gorm:"type:varchar(50)"`
	StartDate     time.Time `gorm:"index"`
	EndDate       time.Time `gorm:"index"`
	IsActive      bool      `gorm:"type:boolean"`
	CreatedAt     time.Time `gorm:"index"`
	UpdatedAt     time.Time
}

func (PriceBookHeader) TableName() string {
	return "crm_price_books"
}

func ToPriceBookHeaderDomain(h *PriceBookHeader) *domain.PriceBookHeader {
	if h == nil {
		return nil
	}
	return &domain.PriceBookHeader{
		ID:            h.ID,
		LegalEntityID: h.LegalEntityID,
		PriceBookCode: h.PriceBookCode,
		Name:          h.Name,
		Type:          domain.PriceBookType(h.Type),
		StartDate:     h.StartDate,
		EndDate:       h.EndDate,
		IsActive:      h.IsActive,
		CreatedAt:     h.CreatedAt,
		UpdatedAt:     h.UpdatedAt,
	}
}

func FromPriceBookHeaderDomain(h *domain.PriceBookHeader) *PriceBookHeader {
	if h == nil {
		return nil
	}
	return &PriceBookHeader{
		ID:            h.ID,
		LegalEntityID: h.LegalEntityID,
		PriceBookCode: h.PriceBookCode,
		Name:          h.Name,
		Type:          string(h.Type),
		StartDate:     h.StartDate,
		EndDate:       h.EndDate,
		IsActive:      h.IsActive,
		CreatedAt:     h.CreatedAt,
		UpdatedAt:     h.UpdatedAt,
	}
}

type PriceBookEntry struct {
	ID                    string          `gorm:"primaryKey;type:varchar(255)"`
	PriceBookID           string          `gorm:"type:varchar(255);index"`
	MaterialID            string          `gorm:"type:varchar(255);index"`
	UnitListPrice         decimal.Decimal `gorm:"type:numeric(18,4)"`
	MinQuantityThreshold decimal.Decimal `gorm:"type:numeric(14,4)"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (PriceBookEntry) TableName() string {
	return "crm_price_book_entries"
}

func ToPriceBookEntryDomain(e *PriceBookEntry) *domain.PriceBookEntry {
	if e == nil {
		return nil
	}
	return &domain.PriceBookEntry{
		ID:                    e.ID,
		PriceBookID:           e.PriceBookID,
		MaterialID:            e.MaterialID,
		UnitListPrice:         e.UnitListPrice,
		MinQuantityThreshold: e.MinQuantityThreshold,
		CreatedAt:             e.CreatedAt,
		UpdatedAt:             e.UpdatedAt,
	}
}

func FromPriceBookEntryDomain(e *domain.PriceBookEntry) *PriceBookEntry {
	if e == nil {
		return nil
	}
	return &PriceBookEntry{
		ID:                    e.ID,
		PriceBookID:           e.PriceBookID,
		MaterialID:            e.MaterialID,
		UnitListPrice:         e.UnitListPrice,
		MinQuantityThreshold: e.MinQuantityThreshold,
		CreatedAt:             e.CreatedAt,
		UpdatedAt:             e.UpdatedAt,
	}
}

type PricingStrategy struct {
	ID                  string          `gorm:"primaryKey;type:varchar(255)"`
	PriceBookID         string          `gorm:"type:varchar(255);index"`
	EvaluationRule      string          `gorm:"type:varchar(50)"`
	StrategyVersion     int             `gorm:"type:int"`
	ModifierPercentage  decimal.Decimal `gorm:"type:numeric(5,4)"`
	ConfigurationMatrix string          `gorm:"type:jsonb"`
	IsActive            bool            `gorm:"type:boolean"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (PricingStrategy) TableName() string {
	return "crm_pricing_strategies"
}

func ToPricingStrategyDomain(s *PricingStrategy) *domain.PricingStrategy {
	if s == nil {
		return nil
	}
	return &domain.PricingStrategy{
		ID:                  s.ID,
		PriceBookID:         s.PriceBookID,
		EvaluationRule:      domain.StrategyEvaluationRule(s.EvaluationRule),
		StrategyVersion:     s.StrategyVersion,
		ModifierPercentage:  s.ModifierPercentage,
		ConfigurationMatrix: s.ConfigurationMatrix,
		IsActive:            s.IsActive,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}

func FromPricingStrategyDomain(s *domain.PricingStrategy) *PricingStrategy {
	if s == nil {
		return nil
	}
	var matrix string
	if s.ConfigurationMatrix != nil {
		if str, ok := s.ConfigurationMatrix.(string); ok {
			matrix = str
		}
	}
	return &PricingStrategy{
		ID:                  s.ID,
		PriceBookID:         s.PriceBookID,
		EvaluationRule:      string(s.EvaluationRule),
		StrategyVersion:     s.StrategyVersion,
		ModifierPercentage:  s.ModifierPercentage,
		ConfigurationMatrix: matrix,
		IsActive:            s.IsActive,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}

type ServiceTicket struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	CustomerID  string    `gorm:"type:varchar(255);index"`
	Title       string    `gorm:"type:varchar(255)"`
	Description string    `gorm:"type:text"`
	Status      string    `gorm:"type:varchar(50)"`
	Priority    string    `gorm:"type:varchar(50)"`
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time
}

func (ServiceTicket) TableName() string {
	return "crm_service_tickets"
}

func ToServiceTicketDomain(t *ServiceTicket) *domain.ServiceTicket {
	if t == nil {
		return nil
	}
	return &domain.ServiceTicket{
		CustomerID:  t.CustomerID,
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func FromServiceTicketDomain(t *domain.ServiceTicket) *ServiceTicket {
	if t == nil {
		return nil
	}
	return &ServiceTicket{
		CustomerID:  t.CustomerID,
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

type Campaign struct {
	ID        string          `gorm:"primaryKey;type:varchar(255)"`
	Name      string          `gorm:"type:varchar(255)"`
	Type      string          `gorm:"type:varchar(50)"`
	Status    string          `gorm:"type:varchar(50)"`
	Budget    decimal.Decimal `gorm:"type:numeric(18,4)"`
	CreatedAt time.Time       `gorm:"index"`
	UpdatedAt time.Time
}

func (Campaign) TableName() string {
	return "crm_campaigns"
}

func ToCampaignDomain(c *Campaign) *domain.Campaign {
	if c == nil {
		return nil
	}
	return &domain.Campaign{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		Status:    c.Status,
		Budget:    c.Budget,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func FromCampaignDomain(c *domain.Campaign) *Campaign {
	if c == nil {
		return nil
	}
	return &Campaign{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		Status:    c.Status,
		Budget:    c.Budget,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type CustomerInteraction struct {
	ID              string    `gorm:"primaryKey;type:varchar(255)"`
	CustomerID      string    `gorm:"type:varchar(255);index"`
	Type            string    `gorm:"type:varchar(50)"`
	Subject         string    `gorm:"type:varchar(255)"`
	Description     string    `gorm:"type:text"`
	InteractionDate time.Time `gorm:"index"`
	CreatedBy       string    `gorm:"type:varchar(255)"`
	CreatedAt       time.Time
}

func (CustomerInteraction) TableName() string {
	return "crm_customer_interactions"
}

func ToCustomerInteractionDomain(i *CustomerInteraction) *domain.CustomerInteraction {
	if i == nil {
		return nil
	}
	return &domain.CustomerInteraction{
		CustomerID:      i.CustomerID,
		ID:              i.ID,
		Type:            i.Type,
		Subject:         i.Subject,
		Description:     i.Description,
		InteractionDate: i.InteractionDate,
		CreatedBy:       i.CreatedBy,
		CreatedAt:       i.CreatedAt,
	}
}

func FromCustomerInteractionDomain(i *domain.CustomerInteraction) *CustomerInteraction {
	if i == nil {
		return nil
	}
	return &CustomerInteraction{
		CustomerID:      i.CustomerID,
		ID:              i.ID,
		Type:            i.Type,
		Subject:         i.Subject,
		Description:     i.Description,
		InteractionDate: i.InteractionDate,
		CreatedBy:       i.CreatedBy,
		CreatedAt:       i.CreatedAt,
	}
}

type BillingTrigger struct {
	ID                          string          `gorm:"primaryKey;type:varchar(255)"`
	LegalEntityID               string          `gorm:"type:varchar(255);index"`
	SalesOrderID                string          `gorm:"type:varchar(255);index"`
	SourceDeliveryDocumentID    string          `gorm:"type:varchar(255);index"`
	BillableAmount              decimal.Decimal `gorm:"type:numeric(18,4)"`
	TaxAmount                   decimal.Decimal `gorm:"type:numeric(18,4)"`
	Status                      string          `gorm:"type:varchar(50);index"`
	TriggeredAt                 time.Time       `gorm:"primaryKey;index"` // Partition Coordinator
	ProcessedAt                 *time.Time      `gorm:"index"`
}

func (BillingTrigger) TableName() string {
	return "crm_billing_triggers"
}

func ToBillingTriggerDomain(b *BillingTrigger) *domain.BillingTrigger {
	if b == nil {
		return nil
	}
	return &domain.BillingTrigger{
		ID:                       b.ID,
		LegalEntityID:            b.LegalEntityID,
		SalesOrderID:             b.SalesOrderID,
		SourceDeliveryDocumentID: b.SourceDeliveryDocumentID,
		BillableAmount:           b.BillableAmount,
		TaxAmount:                b.TaxAmount,
		Status:                   domain.BillingTriggerStatus(b.Status),
		TriggeredAt:              b.TriggeredAt,
		ProcessedAt:              b.ProcessedAt,
	}
}

func FromBillingTriggerDomain(b *domain.BillingTrigger) *BillingTrigger {
	if b == nil {
		return nil
	}
	return &BillingTrigger{
		ID:                       b.ID,
		LegalEntityID:            b.LegalEntityID,
		SalesOrderID:             b.SalesOrderID,
		SourceDeliveryDocumentID: b.SourceDeliveryDocumentID,
		BillableAmount:           b.BillableAmount,
		TaxAmount:                b.TaxAmount,
		Status:                   string(b.Status),
		TriggeredAt:              b.TriggeredAt,
		ProcessedAt:              b.ProcessedAt,
	}
}

type TransactionalOutbox struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType   string    `gorm:"type:varchar(255);index"`
	AggregateID string    `gorm:"type:varchar(255);index"`
	Payload     string    `gorm:"type:jsonb"`
	Status      string    `gorm:"type:varchar(50);index"`
	RetryCount  int       `gorm:"type:int;default:0"`
	CreatedAt   time.Time `gorm:"index"`
}

func (TransactionalOutbox) TableName() string {
	return "crm_transactional_outbox"
}

func ToOutboxDomain(o *TransactionalOutbox) *domain.TransactionalOutbox {
	if o == nil {
		return nil
	}
	return &domain.TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     o.Payload,
		Status:      domain.OutboxStatus(o.Status),
		RetryCount:  o.RetryCount,
		CreatedAt:   o.CreatedAt,
	}
}

func FromOutboxDomain(o *domain.TransactionalOutbox) *TransactionalOutbox {
	if o == nil {
		return nil
	}
	var payloadStr string
	if o.Payload != nil {
		if s, ok := o.Payload.(string); ok {
			payloadStr = s
		}
	}
	return &TransactionalOutbox{
		ID:          o.ID,
		EventType:   o.EventType,
		AggregateID: o.AggregateID,
		Payload:     payloadStr,
		Status:      string(o.Status),
		RetryCount:  o.RetryCount,
		CreatedAt:   o.CreatedAt,
	}
}

type KafkaEventInbox struct {
	AttemptCount     int       `gorm:"type:integer;default:0;not null"`
	EventID          string    `gorm:"primaryKey;type:varchar(255)"`
	EventType        string    `gorm:"type:varchar(255)"`
	ProcessedAt      time.Time `gorm:"index"`
	ProcessingStatus string    `gorm:"type:varchar(50)"`
	Payload          string    `gorm:"type:jsonb"`
}

func (KafkaEventInbox) TableName() string {
	return "crm_kafka_event_inbox"
}

func ToInboxDomain(i *KafkaEventInbox) *domain.KafkaEventInbox {
	if i == nil {
		return nil
	}
	return &domain.KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: domain.EventProcessingStatus(i.ProcessingStatus),
		Payload:          i.Payload,
	}
}

func FromInboxDomain(i *domain.KafkaEventInbox) *KafkaEventInbox {
	if i == nil {
		return nil
	}
	var payloadStr string
	if i.Payload != nil {
		if s, ok := i.Payload.(string); ok {
			payloadStr = s
		}
	}
	return &KafkaEventInbox{
		AttemptCount:     i.AttemptCount,
		EventID:          i.EventID,
		EventType:        i.EventType,
		ProcessedAt:      i.ProcessedAt,
		ProcessingStatus: string(i.ProcessingStatus),
		Payload:          payloadStr,
	}
}

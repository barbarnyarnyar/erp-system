package domain

import "context"

type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, id string) (*Customer, error)
	List(ctx context.Context) ([]Customer, error)
	Update(ctx context.Context, customer *Customer) error
	Delete(ctx context.Context, id string) error
}

type LeadRepository interface {
	Create(ctx context.Context, lead *Lead) error
	GetByID(ctx context.Context, id string) (*Lead, error)
	List(ctx context.Context) ([]Lead, error)
	Update(ctx context.Context, lead *Lead) error
	Delete(ctx context.Context, id string) error
}

type OpportunityRepository interface {
	Create(ctx context.Context, opportunity *Opportunity) error
	GetByID(ctx context.Context, id string) (*Opportunity, error)
	List(ctx context.Context) ([]Opportunity, error)
	Update(ctx context.Context, opportunity *Opportunity) error
	Delete(ctx context.Context, id string) error
}

type SalesOrderRepository interface {
	Create(ctx context.Context, order *SalesOrder) error
	GetByID(ctx context.Context, id string) (*SalesOrder, error)
	List(ctx context.Context) ([]SalesOrder, error)
	Update(ctx context.Context, order *SalesOrder) error
	Delete(ctx context.Context, id string) error
}

type SalesOrderItemRepository interface {
	Create(ctx context.Context, item *SalesOrderItem) error
	ListByOrderID(ctx context.Context, orderID string) ([]SalesOrderItem, error)
}

type QuoteRepository interface {
	Create(ctx context.Context, quote *Quote) error
	GetByID(ctx context.Context, id string) (*Quote, error)
	List(ctx context.Context) ([]Quote, error)
	Update(ctx context.Context, quote *Quote) error
	Delete(ctx context.Context, id string) error
}

type QuoteLineItemRepository interface {
	Create(ctx context.Context, item *QuoteLineItem) error
	ListByQuoteID(ctx context.Context, quoteID string) ([]QuoteLineItem, error)
}

type PriceListRepository interface {
	Create(ctx context.Context, priceList *PriceList) error
	GetByID(ctx context.Context, id string) (*PriceList, error)
	List(ctx context.Context) ([]PriceList, error)
	Update(ctx context.Context, priceList *PriceList) error
	Delete(ctx context.Context, id string) error
}

type PriceListItemRepository interface {
	Create(ctx context.Context, item *PriceListItem) error
	ListByPriceListID(ctx context.Context, priceListID string) ([]PriceListItem, error)
}

type ServiceTicketRepository interface {
	Create(ctx context.Context, ticket *ServiceTicket) error
	GetByID(ctx context.Context, id string) (*ServiceTicket, error)
	List(ctx context.Context) ([]ServiceTicket, error)
	Update(ctx context.Context, ticket *ServiceTicket) error
	Delete(ctx context.Context, id string) error
}

type CampaignRepository interface {
	Create(ctx context.Context, campaign *Campaign) error
	GetByID(ctx context.Context, id string) (*Campaign, error)
	List(ctx context.Context) ([]Campaign, error)
	Update(ctx context.Context, campaign *Campaign) error
	Delete(ctx context.Context, id string) error
}

type OpportunityStageHistoryRepository interface {
	Create(ctx context.Context, osh *OpportunityStageHistory) error
	ListByOpportunityID(ctx context.Context, opportunityID string) ([]OpportunityStageHistory, error)
}


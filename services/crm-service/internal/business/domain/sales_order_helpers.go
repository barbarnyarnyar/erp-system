package domain

import (
	"errors"
	"time"
)

type SalesOrderStatus string

const (
	SalesOrderStatusDraft     SalesOrderStatus = "DRAFT"
	SalesOrderStatusConfirmed SalesOrderStatus = "CONFIRMED"
	SalesOrderStatusShipped   SalesOrderStatus = "SHIPPED"
	SalesOrderStatusDelivered SalesOrderStatus = "DELIVERED"
	SalesOrderStatusCancelled SalesOrderStatus = "CANCELLED"
)

func (s SalesOrderStatus) IsValid() bool {
	switch s {
	case SalesOrderStatusDraft, SalesOrderStatusConfirmed, SalesOrderStatusShipped,
		SalesOrderStatusDelivered, SalesOrderStatusCancelled:
		return true
	}
	return false
}

var (
	ErrOrderNotConfirmable = errors.New("sales order is not in a confirmable state")
	ErrOrderNotFound       = errors.New("sales order not found")
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrCustomerNotActive   = errors.New("customer is not active")
	ErrOrderHasNoItems     = errors.New("sales order has no items")
	ErrInvalidItemQuantity = errors.New("sales order item has invalid quantity")
)

func (o *SalesOrder) CanConfirm() bool {
	return o != nil && o.Status == string(SalesOrderStatusDraft)
}

func (o *SalesOrder) MarkConfirmed(at time.Time) {
	o.Status = string(SalesOrderStatusConfirmed)
	o.UpdatedAt = at
}

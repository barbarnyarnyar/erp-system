package domain

import (
	"errors"
	"time"
)

type SalesOrderStatus = SalesOrderState

const (
	SalesOrderStatusDraft     SalesOrderStatus = "DRAFT"
	SalesOrderStatusConfirmed SalesOrderStatus = "CONFIRMED"
	SalesOrderStatusShipped   SalesOrderStatus = "SHIPPED"
	SalesOrderStatusDelivered SalesOrderStatus = "DELIVERED"
	SalesOrderStatusCancelled SalesOrderStatus = "CANCELLED"
)

var (
	ErrOrderNotConfirmable = errors.New("sales order is not in a confirmable state")
	ErrOrderNotFound       = errors.New("sales order not found")
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrCustomerNotActive   = errors.New("customer is not active")
	ErrOrderHasNoItems     = errors.New("sales order has no items")
	ErrInvalidItemQuantity = errors.New("sales order item has invalid quantity")
)

func (o *SalesOrder) CanConfirm() bool {
	return o != nil && o.Status == SalesOrderStatusDraft
}

func (o *SalesOrder) MarkConfirmed(at time.Time) {
	o.Status = SalesOrderStatusConfirmed
	o.UpdatedAt = at
}

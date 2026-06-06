package service

import (
	"log"
	"context"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type CostingService struct {
	costRepo  domain.CostingRecordRepository
	poRepo    domain.ProductionOrderRepository
	compRepo  domain.BOMComponentRepository
	publisher domain.EventPublisher
}

func NewCostingService(
	costRepo domain.CostingRecordRepository,
	poRepo domain.ProductionOrderRepository,
	compRepo domain.BOMComponentRepository,
	publisher domain.EventPublisher,
) *CostingService {
	return &CostingService{
		costRepo:  costRepo,
		poRepo:    poRepo,
		compRepo:  compRepo,
		publisher: publisher,
	}
}

func (s *CostingService) GetCostingRecord(ctx context.Context, poID string) (*domain.CostingRecord, error) {
	return s.costRepo.GetByProductionOrderID(ctx, poID)
}

func (s *CostingService) RunMRP(ctx context.Context) error {
	orders, err := s.poRepo.List(ctx)
	if err != nil {
		return err
	}

	for _, po := range orders {
		if po.Status == "PLANNED" {
			components, _ := s.compRepo.ListByBOMID(ctx, po.BomID)
			for _, comp := range components {
				qtyNeeded := decimal.NewFromInt(int64(po.Quantity)).Mul(comp.Quantity).Mul(decimal.NewFromFloat(1.0).Add(comp.WasteFactor))
				if err := s.publisher.Publish(ctx, domain.TopicMfgMaterialRequired, comp.ComponentProductID, domain.MaterialRequiredEvent{
					ProductID:  comp.ComponentProductID,
					Quantity:   qtyNeeded,
					RequiredBy: po.ScheduledDate,
					Timestamp:  time.Now(),
				}); err != nil {
					log.Printf("ERROR: failed to publish event %s: %v", domain.TopicMfgMaterialRequired, err)
				}
			}
		}
	}
	return nil
}

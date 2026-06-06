package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type QualityService struct {
	qualityRepo domain.QualityInspectionRepository
	nonConfRepo domain.NonConformanceRepository
	woRepo      domain.WorkOrderRepository
	publisher   domain.EventPublisher
	prodSvc     *ProductionService
}

func NewQualityService(
	qualityRepo domain.QualityInspectionRepository,
	nonConfRepo domain.NonConformanceRepository,
	woRepo domain.WorkOrderRepository,
	publisher domain.EventPublisher,
) *QualityService {
	return &QualityService{
		qualityRepo: qualityRepo,
		nonConfRepo: nonConfRepo,
		woRepo:      woRepo,
		publisher:   publisher,
	}
}

func (s *QualityService) SetProductionService(prodSvc *ProductionService) {
	s.prodSvc = prodSvc
}

func (s *QualityService) RecordQualityInspection(ctx context.Context, workOrderID string, inspectorID string, result string, remarks string) (*domain.QualityInspection, error) {
	id := fmt.Sprintf("qi_%d", time.Now().UnixNano())
	qi := &domain.QualityInspection{
		ID:            id,
		WorkOrderID:   workOrderID,
		InspectorID:   inspectorID,
		Result:        result,
		Remarks:       remarks,
		InspectedAt:   time.Now(),
	}

	err := s.qualityRepo.Create(ctx, qi)
	if err != nil {
		return nil, err
	}

	if result == "FAIL" {
		ncID := fmt.Sprintf("nc_%d", time.Now().UnixNano())
		nc := &domain.NonConformance{
			ID:            ncID,
			InspectionID:  id,
			Description:   fmt.Sprintf("Failed quality inspection: %s", remarks),
			Severity:      "MEDIUM",
			Status:        "OPEN",
		}
		_ = s.nonConfRepo.Create(ctx, nc)

		// Publish Quality Inspection Failed Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgQualityInspectionFailed, id, domain.QualityInspectionFailedEvent{
			InspectionID: id,
			WorkOrderID:  workOrderID,
			InspectorID:  inspectorID,
			Remarks:      remarks,
			Timestamp:    time.Now(),
		})

		// Publish Non-Conformance Detected Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgQualityNonConformanceDetected, ncID, domain.QualityNonConformanceDetectedEvent{
			NonConformanceID: ncID,
			InspectionID:     id,
			Severity:         "MEDIUM",
			Description:      nc.Description,
			Timestamp:        time.Now(),
		})

		// Publish Material Wasted Event (using mock quantities/reasons)
		wo, _ := s.woRepo.GetByID(ctx, workOrderID)
		_ = s.publisher.Publish(ctx, domain.TopicMfgMaterialWasted, workOrderID, domain.MaterialWastedEvent{
			ProductionOrderID: wo.ProductionOrderID,
			ProductID:         "component_wasted",
			Quantity:          decimal.NewFromFloat(1.0),
			Reason:            fmt.Sprintf("Inspection Fail: %s", remarks),
			Timestamp:         time.Now(),
		})

	} else if result == "PASS" {
		// Publish Quality Inspection Passed Event
		_ = s.publisher.Publish(ctx, domain.TopicMfgQualityInspectionPassed, id, domain.QualityInspectionPassedEvent{
			InspectionID: id,
			WorkOrderID:  workOrderID,
			InspectorID:  inspectorID,
			Timestamp:    time.Now(),
		})

		if s.prodSvc != nil {
			wo, _ := s.woRepo.GetByID(ctx, workOrderID)
			workOrders, _ := s.woRepo.ListByProductionOrderID(ctx, wo.ProductionOrderID)
			allComplete := true
			for _, w := range workOrders {
				if w.Status != "COMPLETED" {
					allComplete = false
					break
				}
				ins, err := s.qualityRepo.GetByWorkOrderID(ctx, w.ID)
				if err != nil || ins.Result != "PASS" {
					allComplete = false
					break
				}
			}

			if allComplete {
				_, _ = s.prodSvc.CompleteProductionOrder(ctx, wo.ProductionOrderID)
			}
		}
	}

	return qi, nil
}

func (s *QualityService) ListQualityInspections(ctx context.Context) ([]domain.QualityInspection, error) {
	return s.qualityRepo.List(ctx)
}

func (s *QualityService) GetQualityInspection(ctx context.Context, id string) (*domain.QualityInspection, error) {
	return s.qualityRepo.GetByID(ctx, id)
}

func (s *QualityService) UpdateQualityInspection(ctx context.Context, id string, result string, remarks string) (*domain.QualityInspection, error) {
	qi, err := s.qualityRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	qi.Result = result
	qi.Remarks = remarks

	err = s.qualityRepo.Update(ctx, qi)
	if err != nil {
		return nil, err
	}
	return qi, nil
}

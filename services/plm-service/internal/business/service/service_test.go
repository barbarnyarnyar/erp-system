package service_test

import (
	"context"
	"testing"

	"github.com/erp-system/plm-service/internal/business/domain"
	"github.com/erp-system/plm-service/internal/business/service"
	"github.com/erp-system/plm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

func TestPlmService(t *testing.T) {
	matRepo := memory.NewMemoryMaterialMasterRepo()
	hdrRepo := memory.NewMemoryBomHeaderRepo()
	lineRepo := memory.NewMemoryBomLineRepo()
	ecoRepo := memory.NewMemoryEngineeringChangeOrderRepo()
	publisher := &MockPublisher{}

	matSvc := service.NewMaterialService(matRepo, publisher)
	bomSvc := service.NewBomService(hdrRepo, lineRepo, matRepo, publisher)
	changeSvc := service.NewEngineeringChangeService(ecoRepo, matRepo, publisher)

	ctx := context.Background()

	// 1. Test CreateMaterial
	mat1, err := matSvc.CreateMaterial(ctx, "tenant-1", "SKU-001", "Steel Pipe", "EA", domain.ProcurementTypeBUY)
	if err != nil {
		t.Fatalf("failed to create material: %v", err)
	}
	if mat1.Sku != "SKU-001" {
		t.Errorf("expected SKU-001, got %s", mat1.Sku)
	}

	mat2, err := matSvc.CreateMaterial(ctx, "tenant-1", "SKU-002", "Steel Joint", "EA", domain.ProcurementTypeMAKE)
	if err != nil {
		t.Fatalf("failed to create material 2: %v", err)
	}

	// 2. Test UpdateTechnicalSpecs
	mat1, err = matSvc.UpdateTechnicalSpecs(ctx, mat1.ID, `{"diameter": "2 inches"}`)
	if err != nil {
		t.Fatalf("failed to update technical specs: %v", err)
	}
	if mat1.TechnicalSpecifications != `{"diameter": "2 inches"}` {
		t.Errorf("expected specifications update, got %v", mat1.TechnicalSpecifications)
	}

	// 3. Test TransitionStatus
	mat1, err = matSvc.TransitionStatus(ctx, mat1.ID, domain.MaterialStatusACTIVE)
	if err != nil {
		t.Fatalf("failed to transition status: %v", err)
	}

	// 4. Test EstablishBomHeader
	bomLines := []service.BomLineInput{
		{
			ComponentMaterialID: mat1.ID,
			SequenceNumber:      10,
			QuantityRequired:    decimal.NewFromInt(2),
			Uom:                 "EA",
			ScrapPercentage:     decimal.NewFromFloat(0.05),
		},
	}
	bh, err := bomSvc.EstablishBomHeader(ctx, "tenant-1", mat2.ID, "REV-1.0", bomLines)
	if err != nil {
		t.Fatalf("failed to establish BOM: %v", err)
	}
	if bh.Status != domain.BomStatusDRAFT {
		t.Errorf("expected DRAFT status, got %v", bh.Status)
	}

	// 5. Test ReleaseBom
	bh, err = bomSvc.ReleaseBom(ctx, bh.ID)
	if err != nil {
		t.Fatalf("failed to release BOM: %v", err)
	}
	if bh.Status != domain.BomStatusRELEASED {
		t.Errorf("expected RELEASED, got %v", bh.Status)
	}

	// 6. Test ExplodeBillOfMaterials
	graph, err := bomSvc.ExplodeBillOfMaterials(ctx, bh.ID, 5)
	if err != nil {
		t.Fatalf("failed to explode BOM: %v", err)
	}
	if len(graph.Components) != 1 {
		t.Errorf("expected 1 component node, got %d", len(graph.Components))
	}

	// 7. Test InitiateChangeRequest
	eco, err := changeSvc.InitiateChangeRequest(ctx, "tenant-1", mat2.ID, "emp-101", "Design update", "Need to change dimensions")
	if err != nil {
		t.Fatalf("failed to initiate ECO: %v", err)
	}
	if eco.Status != domain.EcoStatusDRAFT {
		t.Errorf("expected DRAFT, got %v", eco.Status)
	}

	// 8. Test ProcessApprovalAction
	eco, err = changeSvc.ProcessApprovalAction(ctx, eco.ID, "emp-202", domain.EcoStatusAPPROVED)
	if err != nil {
		t.Fatalf("failed to process approval: %v", err)
	}
	if eco.Status != domain.EcoStatusAPPROVED {
		t.Errorf("expected APPROVED, got %v", eco.Status)
	}
}

package service_test

import (
	"context"
	"testing"

	sharedtesting "erp-system/shared/testing"
	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/erp-system/eam-service/internal/business/service"
	"github.com/erp-system/eam-service/internal/data/memory"
)

func TestEamService(t *testing.T) {
	facRepo := memory.NewMemoryFacilityRepo()
	eqRepo := memory.NewMemoryEquipmentRepo()
	woRepo := memory.NewMemoryMaintenanceWorkOrderRepo()
	schRepo := memory.NewMemoryPreventativeScheduleRepo()
	publisher := &sharedtesting.MockPublisher{}

	eqSvc := service.NewEquipmentService(facRepo, eqRepo, publisher)
	maintSvc := service.NewMaintenanceService(woRepo, eqRepo, schRepo, publisher)

	ctx := context.Background()

	// 1. Test CreateFacility
	f, err := eqSvc.CreateFacility(ctx, "tenant-1", "Houston Plant", "100 Industrial Blvd")
	if err != nil {
		t.Fatalf("failed to create facility: %v", err)
	}
	if f.Name != "Houston Plant" {
		t.Errorf("expected Houston Plant, got %s", f.Name)
	}

	// 2. Test RegisterEquipment
	eq, err := eqSvc.RegisterEquipment(ctx, "tenant-1", f.ID, "EQ-001", "Turbine A", "SN-9988")
	if err != nil {
		t.Fatalf("failed to register equipment: %v", err)
	}
	if eq.Status != domain.EquipmentStatusONLINE {
		t.Errorf("expected ONLINE status, got %v", eq.Status)
	}

	// 3. Test UpdateEquipmentStatus
	eq, err = eqSvc.UpdateEquipmentStatus(ctx, eq.ID, domain.EquipmentStatusOFFLINE_BROKEN)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	if eq.Status != domain.EquipmentStatusOFFLINE_BROKEN {
		t.Errorf("expected OFFLINE_BROKEN, got %v", eq.Status)
	}

	// 4. Test FileMachineIncident
	wo, err := maintSvc.FileMachineIncident(ctx, "tenant-1", eq.ID, "emp-101", "Turbine failure", domain.WorkOrderPriorityCRITICAL)
	if err != nil {
		t.Fatalf("failed to file incident: %v", err)
	}
	if wo.Status != domain.WorkOrderStatusOPEN {
		t.Errorf("expected OPEN status, got %v", wo.Status)
	}

	// 5. Test RouteToTechnician
	wo, err = maintSvc.RouteToTechnician(ctx, wo.ID, "emp-202")
	if err != nil {
		t.Fatalf("failed to route: %v", err)
	}
	if wo.Status != domain.WorkOrderStatusASSIGNED {
		t.Errorf("expected ASSIGNED, got %v", wo.Status)
	}

	// 6. Test TransitionToActiveState
	wo, err = maintSvc.TransitionToActiveState(ctx, wo.ID)
	if err != nil {
		t.Fatalf("failed to transition to active: %v", err)
	}
	if wo.Status != domain.WorkOrderStatusIN_PROGRESS {
		t.Errorf("expected IN_PROGRESS, got %v", wo.Status)
	}

	// 7. Test FinalizeResolution
	wo, err = maintSvc.FinalizeResolution(ctx, wo.ID, "Replaced gasket")
	if err != nil {
		t.Fatalf("failed to finalize resolution: %v", err)
	}
	if wo.Status != domain.WorkOrderStatusRESOLVED {
		t.Errorf("expected RESOLVED, got %v", wo.Status)
	}
}

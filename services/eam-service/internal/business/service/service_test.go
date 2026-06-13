package service_test

import (
	"context"
	"testing"

	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/erp-system/eam-service/internal/business/service"
	"github.com/erp-system/eam-service/internal/data/sql"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (
	*gorm.DB,
	domain.FacilityRepository,
	domain.EquipmentRepository,
	domain.MaintenanceWorkOrderRepository,
	domain.PreventativeScheduleRepository,
	domain.TelemetryIngestBufferRepository,
	domain.TransactionalOutboxRepository,
	domain.KafkaEventInboxRepository,
) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.Facility{},
		&sql.Equipment{},
		&sql.MaintenanceWorkOrder{},
		&sql.PreventativeSchedule{},
		&sql.TelemetryIngestBuffer{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	facRepo := sql.NewSQLFacilityRepository(db)
	eqRepo := sql.NewSQLEquipmentRepository(db)
	woRepo := sql.NewSQLMaintenanceWorkOrderRepository(db)
	schRepo := sql.NewSQLPreventativeScheduleRepository(db)
	bufRepo := sql.NewSQLTelemetryIngestBufferRepository(db)
	outboxRepo := sql.NewSQLTransactionalOutboxRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	return db, facRepo, eqRepo, woRepo, schRepo, bufRepo, outboxRepo, inboxRepo
}

func TestEamService(t *testing.T) {
	db, facRepo, eqRepo, woRepo, schRepo, bufRepo, outboxRepo, inboxRepo := setupTestDB(t)

	reliableSvc := service.NewReliableMessagingService(db, inboxRepo, outboxRepo)
	eqSvc := service.NewEquipmentService(db, facRepo, eqRepo, reliableSvc)
	maintSvc := service.NewMaintenanceService(db, woRepo, eqRepo, schRepo, reliableSvc)
	telSvc := service.NewTelemetryIngestionService(db, bufRepo)

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

	// 3. Test UpdateEquipmentStatus (without tx)
	eq, err = eqSvc.UpdateEquipmentStatus(ctx, nil, eq.ID, domain.EquipmentStatusOFFLINE_BROKEN)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	if eq.Status != domain.EquipmentStatusOFFLINE_BROKEN {
		t.Errorf("expected OFFLINE_BROKEN, got %v", eq.Status)
	}

	// Verify that an outbox event eam.machine.offline was written
	outboxes, err := outboxRepo.GetUnsent(ctx, 10)
	if err != nil {
		t.Fatalf("failed to fetch outbox messages: %v", err)
	}
	foundOffline := false
	for _, out := range outboxes {
		if out.EventType == domain.TopicEamMachineOffline && out.AggregateID == eq.ID {
			foundOffline = true
			break
		}
	}
	if !foundOffline {
		t.Error("expected eam.machine.offline outbox event to be queued")
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

	// Verify that equipment status is now ONLINE again
	eq, err = eqRepo.GetByID(ctx, eq.ID)
	if err != nil {
		t.Fatalf("failed to fetch equipment: %v", err)
	}
	if eq.Status != domain.EquipmentStatusONLINE {
		t.Errorf("expected equipment status to be ONLINE after resolution, got %v", eq.Status)
	}

	// Verify that an outbox event eam.machine.online was written
	outboxes, err = outboxRepo.GetUnsent(ctx, 10)
	if err != nil {
		t.Fatalf("failed to fetch outbox messages: %v", err)
	}
	foundOnline := false
	for _, out := range outboxes {
		if out.EventType == domain.TopicEamMachineOnline && out.AggregateID == eq.ID {
			foundOnline = true
			break
		}
	}
	if !foundOnline {
		t.Error("expected eam.machine.online outbox event to be queued")
	}

	// 8. Test Telemetry Ingestion and Draining
	err = telSvc.QueueSensorMetrics(ctx, "tenant-1", eq.ID, "temp_c", decimal.NewFromFloat(85.4))
	if err != nil {
		t.Fatalf("failed to queue telemetry metrics: %v", err)
	}

	// Ingested metric should be in the database
	list, err := bufRepo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list telemetry buffer: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 telemetry metric, got %d", len(list))
	}

	// Flush the metrics using the service
	flushedIds, err := telSvc.FlushStagedMetricsToTimeSeriesStore(ctx, nil, 10)
	if err != nil {
		t.Fatalf("failed to flush metrics: %v", err)
	}
	if len(flushedIds) != 1 {
		t.Errorf("expected 1 flushed metric ID, got %d", len(flushedIds))
	}

	// Telemetry buffer should now be empty
	list, err = bufRepo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list telemetry buffer: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected telemetry buffer to be empty, got %d", len(list))
	}
}

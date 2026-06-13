package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/eam-service/internal/business/domain"
	"github.com/erp-system/eam-service/internal/business/service"
	kafkago "github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader      *kafkago.Reader
	publisher   domain.EventPublisher
	reliableSvc service.ReliableMessagingService
	eqSvc       *service.EquipmentService
	maintSvc    *service.MaintenanceService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	reliableSvc service.ReliableMessagingService,
	eqSvc *service.EquipmentService,
	maintSvc *service.MaintenanceService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicScmAssetReceived,
		domain.TopicFmAssetCapitalized,
		domain.TopicHrEmployeeCreated,
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:      reader,
		publisher:   publisher,
		reliableSvc: reliableSvc,
		eqSvc:       eqSvc,
		maintSvc:    maintSvc,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for eam-service...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer due to context cancellation...")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error reading message: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			log.Printf("[EAM-CONSUMER] Received event on topic %s, key %s", msg.Topic, string(msg.Key))
			if err := c.handleMessage(ctx, msg.Topic, msg.Value); err != nil {
				log.Printf("[EAM-CONSUMER] Failed to process event %s: %v", msg.Topic, err)
			}
		}
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicScmAssetReceived:
		var ev struct {
			EventID       string `json:"event_id"`
			LegalEntityID string `json:"legal_entity_id"`
			SerialNumber  string `json:"serial_number"`
			Manufacturer  string `json:"manufacturer"`
			Timestamp     string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("[EAM-CONSUMER] Auto-registering equipment for received asset: Serial: %s, Manufacturer: %s", ev.SerialNumber, ev.Manufacturer)
			// Auto register under default/dummy facility
			_, err := c.eqSvc.RegisterEquipment(txCtx, ev.LegalEntityID, "fac_default", "TAG-"+ev.SerialNumber[:6], "Asset "+ev.SerialNumber, ev.SerialNumber)
			return err
		})

	case domain.TopicFmAssetCapitalized:
		var ev struct {
			EventID          string `json:"event_id"`
			LegalEntityID    string `json:"legal_entity_id"`
			FinancialAssetID string `json:"financial_asset_id"`
			AssetTag         string `json:"asset_tag"`
			Timestamp        string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("[EAM-CONSUMER] Linking financial asset %s to equipment with tag %s", ev.FinancialAssetID, ev.AssetTag)
			all, err := c.eqSvc.FetchTargetTenantAssets(txCtx, ev.LegalEntityID, domain.EquipmentStatusONLINE)
			if err != nil {
				allOffline, _ := c.eqSvc.FetchTargetTenantAssets(txCtx, ev.LegalEntityID, domain.EquipmentStatusOFFLINE_BROKEN)
				all = append(all, allOffline...)
				allInMaint, _ := c.eqSvc.FetchTargetTenantAssets(txCtx, ev.LegalEntityID, domain.EquipmentStatusIN_MAINTENANCE)
				all = append(all, allInMaint...)
			}
			for _, eq := range all {
				if eq.AssetTag == ev.AssetTag {
					_, err = c.eqSvc.AssociateFinancialAsset(txCtx, eq.ID, ev.FinancialAssetID)
					return err
				}
			}
			return nil
		})

	case domain.TopicHrEmployeeCreated:
		var ev struct {
			EventID       string `json:"event_id"`
			LegalEntityID string `json:"legal_entity_id"`
			EmployeeID    string `json:"employee_id"`
			ExplicitRole  string `json:"explicit_role"`
			Timestamp     string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("[EAM-CONSUMER] Syncing employee %s as EAM Technician", ev.EmployeeID)
			return nil
		})
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

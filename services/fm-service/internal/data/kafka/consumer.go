package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

type DeadLetterMessage struct {
	OriginalTopic string      `json:"original_topic"`
	OriginalKey   string      `json:"original_key,omitempty"`
	Payload       interface{} `json:"payload"`
	Error         string      `json:"error"`
	FailedAt      time.Time   `json:"failed_at"`
	ServiceName   string      `json:"service_name"`
}

const (
	TopicHrEmployeeCreatedDeadLetter       = domain.TopicHrEmployeeCreated + ".dead-letter"
	TopicHrPayrollProcessedDeadLetter      = domain.TopicHrPayrollProcessed + ".dead-letter"
	TopicHrExpenseSubmittedDeadLetter      = domain.TopicHrExpenseSubmitted + ".dead-letter"
	TopicScmPurchaseOrderCreatedDeadLetter = domain.TopicScmPurchaseOrderCreated + ".dead-letter"
	TopicScmInventoryValuedDeadLetter      = domain.TopicScmInventoryValued + ".dead-letter"
	TopicCrmOrderConfirmedDeadLetter       = domain.TopicCrmOrderConfirmed + ".dead-letter"
	TopicCrmCustomerCreatedDeadLetter      = domain.TopicCrmCustomerCreated + ".dead-letter"
	TopicMfgProductionCompletedDeadLetter  = domain.TopicMfgProductionCompleted + ".dead-letter"
	TopicMfgMaterialConsumedDeadLetter     = domain.TopicMfgMaterialConsumed + ".dead-letter"
	TopicPrjProjectCreatedDeadLetter       = domain.TopicPrjProjectCreated + ".dead-letter"
	TopicPrjTimeLoggedDeadLetter           = domain.TopicPrjTimeLogged + ".dead-letter"
	TopicPrjExpenseIncurredDeadLetter      = domain.TopicPrjExpenseIncurred + ".dead-letter"

	defaultLegalEntityID = "00000000-0000-0000-0000-000000000000"
)

// KafkaConsumer listens to external microservice events and updates the financial records
type KafkaConsumer struct {
	reader    *kafka.Reader
	publisher domain.EventPublisher
	gl        *service.GeneralLedgerService
	ap        *service.AccountsPayableService
	ar        *service.AccountsReceivableService
	cash      *service.CashManagementService
	budget    *service.BudgetingService
}

// NewKafkaConsumer initializes the Kafka consumer with a list of topics
func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	gl *service.GeneralLedgerService,
	ap *service.AccountsPayableService,
	ar *service.AccountsReceivableService,
	cash *service.CashManagementService,
	budget *service.BudgetingService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicHrEmployeeCreated,
		domain.TopicHrPayrollProcessed,
		domain.TopicHrExpenseSubmitted,
		domain.TopicScmPurchaseOrderCreated,
		domain.TopicScmInvoiceReceived,
		domain.TopicScmInventoryValued,
		domain.TopicCrmOrderConfirmed,
		domain.TopicCrmCustomerCreated,
		domain.TopicMfgProductionCompleted,
		domain.TopicMfgMaterialConsumed,
		domain.TopicPrjProjectCreated,
		domain.TopicPrjTimeLogged,
		domain.TopicPrjExpenseIncurred,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:    reader,
		publisher: publisher,
		gl:        gl,
		ap:        ap,
		ar:        ar,
		cash:      cash,
		budget:    budget,
	}
}

// Start runs the message consumption loop in the background
func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer...")
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

			log.Printf("Received event on topic %s, key %s", msg.Topic, string(msg.Key))
			if err := c.handleMessage(ctx, msg.Topic, msg.Value); err != nil {
				log.Printf("Failed to process event %s: %v", msg.Topic, err)
				c.publishToDLQ(ctx, msg.Topic, string(msg.Key), msg.Value, err)
			}
		}
	}
}

func (c *KafkaConsumer) publishToDLQ(ctx context.Context, topic string, key string, value []byte, err error) {
	dlqMsg := DeadLetterMessage{
		OriginalTopic: topic,
		OriginalKey:   key,
		Payload:       string(value),
		Error:         err.Error(),
		FailedAt:      time.Now(),
		ServiceName:   "fm-service",
	}
	dlqTopic := topic + ".dead-letter"
	if dlqErr := c.publisher.Publish(ctx, dlqTopic, key, dlqMsg); dlqErr != nil {
		log.Printf("ERROR: failed to publish DLQ message for topic %s: %v", topic, dlqErr)
	} else {
		log.Printf("ERROR: consumer handler failed for topic %s: %v — sent to DLQ topic %s", topic, err, dlqTopic)
	}
}

func (c *KafkaConsumer) getOrCreateAccount(ctx context.Context, accNum, name, accType string) (*domain.ChartOfAccounts, error) {
	acc, err := c.gl.GetAccountByCode(ctx, defaultLegalEntityID, accNum)
	if err == nil {
		return acc, nil
	}
	return c.gl.CreateAccount(ctx, defaultLegalEntityID, accNum, name, accType)
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicHrEmployeeCreated:
		var ev domain.EmployeeCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Create payroll liability account for this specific employee
		accNum := "2120-" + ev.EmployeeID[:8]
		accName := "Payroll Liability - " + ev.FirstName + " " + ev.LastName
		_, err := c.getOrCreateAccount(ctx, accNum, accName, "LIABILITY")
		return err

	case domain.TopicHrPayrollProcessed:
		var ev domain.PayrollProcessedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Salaries Expense, Credit Payroll Liability Control
		salariesExpenseAcc, err := c.getOrCreateAccount(ctx, "6010-001", "Salaries Expense", "EXPENSE")
		if err != nil {
			return err
		}
		payrollLiabilityAcc, err := c.getOrCreateAccount(ctx, "2120-001", "Payroll Liabilities Control", "LIABILITY")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             salariesExpenseAcc.ID,
				AmountFunctional:      ev.TotalGross,
				AmountTransactional:   ev.TotalGross,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             payrollLiabilityAcc.ID,
				AmountFunctional:      ev.TotalGross.Neg(),
				AmountTransactional:   ev.TotalGross.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "HR", "PAY-"+ev.PayrollID, ev.Timestamp, lines)
		return err

	case domain.TopicHrExpenseSubmitted:
		var ev domain.ExpenseSubmittedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Travel & Expense, Credit Accounts Payable - Reimbursement
		expenseAcc, err := c.getOrCreateAccount(ctx, "6020-001", "Travel & Entertainment Expense", "EXPENSE")
		if err != nil {
			return err
		}
		payableAcc, err := c.getOrCreateAccount(ctx, "2110-001", "Accounts Payable - Reimbursement", "LIABILITY")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             expenseAcc.ID,
				AmountFunctional:      ev.Amount,
				AmountTransactional:   ev.Amount,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             payableAcc.ID,
				AmountFunctional:      ev.Amount.Neg(),
				AmountTransactional:   ev.Amount.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "HR", "EXP-"+ev.ExpenseID, ev.Timestamp, lines)
		return err

	case domain.TopicScmPurchaseOrderCreated:
		var ev domain.PurchaseOrderCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Inventory in Transit, Credit Accrued Accounts Payable
		invAssetAcc, err := c.getOrCreateAccount(ctx, "1210-001", "Inventory in Transit/Accrued", "ASSET")
		if err != nil {
			return err
		}
		accruedPayableAcc, err := c.getOrCreateAccount(ctx, "2110-002", "Accrued Accounts Payable", "LIABILITY")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             invAssetAcc.ID,
				AmountFunctional:      ev.TotalAmount,
				AmountTransactional:   ev.TotalAmount,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             accruedPayableAcc.ID,
				AmountFunctional:      ev.TotalAmount.Neg(),
				AmountTransactional:   ev.TotalAmount.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "SCM", "PO-LIAB-"+ev.PurchaseOrderID, ev.Timestamp, lines)
		return err

	case domain.TopicScmInvoiceReceived:
		var ev domain.InvoiceReceivedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		_, err := c.ap.CreateVendorBill(ctx, defaultLegalEntityID, ev.VendorID, ev.InvoiceNo, ev.POID, ev.DueDate, ev.TotalAmount, decimal.Zero)
		return err

	case domain.TopicScmInventoryValued:
		var ev domain.InventoryValuedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Raw Materials Inventory, Credit Cost of Goods Sold - Adjustments
		invAssetAcc, err := c.getOrCreateAccount(ctx, "1200-001", "Raw Materials Inventory", "ASSET")
		if err != nil {
			return err
		}
		invAdjAcc, err := c.getOrCreateAccount(ctx, "5010-001", "Cost of Goods Sold - Inventory Adjustments", "EXPENSE")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             invAssetAcc.ID,
				AmountFunctional:      ev.TotalValue,
				AmountTransactional:   ev.TotalValue,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             invAdjAcc.ID,
				AmountFunctional:      ev.TotalValue.Neg(),
				AmountTransactional:   ev.TotalValue.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "SCM", "INV-VAL-"+ev.LocationID, ev.Timestamp, lines)
		return err

	case domain.TopicCrmOrderConfirmed:
		var ev domain.SalesOrderConfirmedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Generate customer invoice
		_, err := c.ar.CreateInvoice(ctx, defaultLegalEntityID, ev.CustomerID, ev.SalesOrderID, ev.TotalAmount, decimal.Zero, ev.Timestamp.AddDate(0, 1, 0))
		return err

	case domain.TopicCrmCustomerCreated:
		var ev domain.CustomerCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Create customer AR account
		accNum := "1100-" + ev.CustomerID[:8]
		accName := "AR - Customer " + ev.CustomerName
		_, err := c.getOrCreateAccount(ctx, accNum, accName, "ASSET")
		return err

	case domain.TopicMfgProductionCompleted:
		var ev domain.ProductionCompletedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Finished Goods, Credit WIP
		fgAcc, err := c.getOrCreateAccount(ctx, "1220-001", "Finished Goods Inventory", "ASSET")
		if err != nil {
			return err
		}
		wipAcc, err := c.getOrCreateAccount(ctx, "1230-001", "Work in Progress", "ASSET")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             fgAcc.ID,
				AmountFunctional:      ev.TotalValuation,
				AmountTransactional:   ev.TotalValuation,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             wipAcc.ID,
				AmountFunctional:      ev.TotalValuation.Neg(),
				AmountTransactional:   ev.TotalValuation.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "MFG", "MFG-COMP-"+ev.ProductionOrderID, ev.Timestamp, lines)
		return err

	case domain.TopicMfgMaterialConsumed:
		var ev domain.MaterialConsumedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit WIP, Credit Raw Materials Inventory
		wipAcc, err := c.getOrCreateAccount(ctx, "1230-001", "Work in Progress", "ASSET")
		if err != nil {
			return err
		}
		rmAcc, err := c.getOrCreateAccount(ctx, "1200-001", "Raw Materials Inventory", "ASSET")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             wipAcc.ID,
				AmountFunctional:      ev.TotalCost,
				AmountTransactional:   ev.TotalCost,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             rmAcc.ID,
				AmountFunctional:      ev.TotalCost.Neg(),
				AmountTransactional:   ev.TotalCost.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "MFG", "MFG-CONS-"+ev.ProductionOrderID, ev.Timestamp, lines)
		return err

	case domain.TopicPrjProjectCreated:
		var ev domain.ProjectCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Create project expense account
		accNum := "6030-" + ev.ProjectID[:8]
		accName := "Project Expense - " + ev.ProjectName
		_, err := c.getOrCreateAccount(ctx, accNum, accName, "EXPENSE")
		return err

	case domain.TopicPrjTimeLogged:
		var ev domain.TimeLoggedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Unbilled Receivables, Credit Project Revenue
		unbilledAcc, err := c.getOrCreateAccount(ctx, "1110-001", "Unbilled Receivables", "ASSET")
		if err != nil {
			return err
		}
		revAcc, err := c.getOrCreateAccount(ctx, "4010-001", "Project Revenue", "REVENUE")
		if err != nil {
			return err
		}

		val := ev.HoursLogged.Mul(ev.BillableRate)
		lines := []domain.UniversalJournalLine{
			{
				AccountID:             unbilledAcc.ID,
				AmountFunctional:      val,
				AmountTransactional:   val,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             revAcc.ID,
				AmountFunctional:      val.Neg(),
				AmountTransactional:   val.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "PRJ", "PRJ-TIME-"+ev.TimeLogID, ev.Timestamp, lines)
		return err

	case domain.TopicPrjExpenseIncurred:
		var ev domain.ProjectExpenseIncurredEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Project Expenses Control, Credit Accounts Payable - Control
		expenseAcc, err := c.getOrCreateAccount(ctx, "6030-001", "Project Expenses Control", "EXPENSE")
		if err != nil {
			return err
		}
		payableAcc, err := c.getOrCreateAccount(ctx, "2110-001", "Accounts Payable - Control", "LIABILITY")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             expenseAcc.ID,
				AmountFunctional:      ev.Amount,
				AmountTransactional:   ev.Amount,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             payableAcc.ID,
				AmountFunctional:      ev.Amount.Neg(),
				AmountTransactional:   ev.Amount.Neg(),
				CurrencyTransactional: "USD",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, defaultLegalEntityID, "PRJ", "PRJ-EXP-"+ev.ExpenseID, ev.Timestamp, lines)
		return err
	}

	return nil
}

// Close stops the reader
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}


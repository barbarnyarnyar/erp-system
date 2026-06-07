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
	TopicCrmSalesOrderConfirmedDeadLetter  = domain.TopicCrmSalesOrderConfirmed + ".dead-letter"
	TopicCrmCustomerCreatedDeadLetter      = domain.TopicCrmCustomerCreated + ".dead-letter"
	TopicMfgProductionCompletedDeadLetter  = domain.TopicMfgProductionCompleted + ".dead-letter"
	TopicMfgMaterialConsumedDeadLetter     = domain.TopicMfgMaterialConsumed + ".dead-letter"
	TopicPrjProjectCreatedDeadLetter       = domain.TopicPrjProjectCreated + ".dead-letter"
	TopicPrjTimeLoggedDeadLetter           = domain.TopicPrjTimeLogged + ".dead-letter"
	TopicPrjExpenseIncurredDeadLetter      = domain.TopicPrjExpenseIncurred + ".dead-letter"
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
		domain.TopicCrmSalesOrderConfirmed,
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

func (c *KafkaConsumer) getOrCreateAccount(ctx context.Context, accNum, name, accType, parentID, currency string) (*domain.Account, error) {
	acc, err := c.gl.GetAccountByNumber(ctx, accNum)
	if err == nil {
		return acc, nil
	}
	return c.gl.CreateAccount(ctx, accNum, name, accType, parentID, currency)
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
		_, err := c.getOrCreateAccount(ctx, accNum, accName, "LIABILITY", "", "USD")
		return err

	case domain.TopicHrPayrollProcessed:
		var ev domain.PayrollProcessedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Salaries Expense, Credit Payroll Liability Control
		salariesExpenseAcc, err := c.getOrCreateAccount(ctx, "6010-001", "Salaries Expense", "EXPENSE", "", "USD")
		if err != nil {
			return err
		}
		payrollLiabilityAcc, err := c.getOrCreateAccount(ctx, "2120-001", "Payroll Liabilities Control", "LIABILITY", "", "USD")
		if err != nil {
			return err
		}

		lines := []domain.JournalEntryLine{
			{
				AccountID:    salariesExpenseAcc.ID,
				DebitAmount:  ev.TotalGross,
				CreditAmount: decimal.Zero,
				Description:  "Payroll Gross Salaries",
			},
			{
				AccountID:    payrollLiabilityAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: ev.TotalGross,
				Description:  "Payroll Liabilities Credit",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "PAY-"+ev.PayrollID, "Record payroll processed entries", lines)
		return err

	case domain.TopicHrExpenseSubmitted:
		var ev domain.ExpenseSubmittedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Travel & Expense, Credit Accounts Payable - Reimbursement
		expenseAcc, err := c.getOrCreateAccount(ctx, "6020-001", "Travel & Entertainment Expense", "EXPENSE", "", "USD")
		if err != nil {
			return err
		}
		payableAcc, err := c.getOrCreateAccount(ctx, "2110-001", "Accounts Payable - Reimbursement", "LIABILITY", "", "USD")
		if err != nil {
			return err
		}

		lines := []domain.JournalEntryLine{
			{
				AccountID:    expenseAcc.ID,
				DebitAmount:  ev.Amount,
				CreditAmount: decimal.Zero,
				Description:  ev.Description,
			},
			{
				AccountID:    payableAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: ev.Amount,
				Description:  "Employee Reimbursement Liability: " + ev.EmployeeID,
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "EXP-"+ev.ExpenseID, "Process employee expense reimbursement", lines)
		return err

	case domain.TopicScmPurchaseOrderCreated:
		var ev domain.PurchaseOrderCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Inventory in Transit, Credit Accrued Accounts Payable
		invAssetAcc, err := c.getOrCreateAccount(ctx, "1210-001", "Inventory in Transit/Accrued", "ASSET", "", "USD")
		if err != nil {
			return err
		}
		accruedPayableAcc, err := c.getOrCreateAccount(ctx, "2110-002", "Accrued Accounts Payable", "LIABILITY", "", "USD")
		if err != nil {
			return err
		}

		lines := []domain.JournalEntryLine{
			{
				AccountID:    invAssetAcc.ID,
				DebitAmount:  ev.TotalAmount,
				CreditAmount: decimal.Zero,
				Description:  "Accrued PO Inventory: " + ev.PONumber,
			},
			{
				AccountID:    accruedPayableAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: ev.TotalAmount,
				Description:  "Accrued PO AP Liability: " + ev.PONumber,
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "PO-LIAB-"+ev.PurchaseOrderID, "Create AP liability for PO "+ev.PONumber, lines)
		return err

	case domain.TopicScmInvoiceReceived:
		var ev domain.InvoiceReceivedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		lines := []domain.VendorBillLine{
			{
				ID:          "vbl_" + ev.InvoiceNo,
				Description: "SCM Vendor Invoice " + ev.InvoiceNo,
				Quantity:    1,
				UnitPrice:   ev.TotalAmount,
				LineTotal:   ev.TotalAmount,
			},
		}
		_, err := c.ap.CreateVendorBill(ctx, ev.VendorID, ev.InvoiceNo, ev.POID, ev.Timestamp, ev.DueDate, ev.TotalAmount, lines)
		return err

	case domain.TopicScmInventoryValued:
		var ev domain.InventoryValuedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Raw Materials Inventory, Credit Cost of Goods Sold - Adjustments
		invAssetAcc, err := c.getOrCreateAccount(ctx, "1200-001", "Raw Materials Inventory", "ASSET", "", "USD")
		if err != nil {
			return err
		}
		invAdjAcc, err := c.getOrCreateAccount(ctx, "5010-001", "Cost of Goods Sold - Inventory Adjustments", "EXPENSE", "", "USD")
		if err != nil {
			return err
		}

		lines := []domain.JournalEntryLine{
			{
				AccountID:    invAssetAcc.ID,
				DebitAmount:  ev.TotalValue,
				CreditAmount: decimal.Zero,
				Description:  "Inventory Valuation adjustment",
			},
			{
				AccountID:    invAdjAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: ev.TotalValue,
				Description:  "Inventory Valuation contra-account",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "INV-VAL-"+ev.LocationID, "Update inventory accounts valuation", lines)
		return err

	case domain.TopicCrmSalesOrderConfirmed:
		var ev domain.SalesOrderConfirmedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Generate customer invoice
		lines := []domain.InvoiceLine{
			{
				Description: "CRM Completed Sale: " + ev.SalesOrderID,
				Quantity:    1,
				UnitPrice:   ev.TotalAmount,
				LineTotal:   ev.TotalAmount,
			},
		}
		_, err := c.ar.CreateInvoice(ctx, ev.CustomerID, ev.Timestamp, ev.Timestamp.AddDate(0, 1, 0), lines)
		return err

	case domain.TopicCrmCustomerCreated:
		var ev domain.CustomerCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Create customer AR account
		accNum := "1100-" + ev.CustomerID[:8]
		accName := "AR - Customer " + ev.CustomerName
		_, err := c.getOrCreateAccount(ctx, accNum, accName, "ASSET", "", "USD")
		return err

	case domain.TopicMfgProductionCompleted:
		var ev domain.ProductionCompletedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Finished Goods, Credit WIP
		fgAcc, err := c.getOrCreateAccount(ctx, "1220-001", "Finished Goods Inventory", "ASSET", "", "USD")
		if err != nil {
			return err
		}
		wipAcc, err := c.getOrCreateAccount(ctx, "1230-001", "Work in Progress", "ASSET", "", "USD")
		if err != nil {
			return err
		}

		lines := []domain.JournalEntryLine{
			{
				AccountID:    fgAcc.ID,
				DebitAmount:  ev.TotalValuation,
				CreditAmount: decimal.Zero,
				Description:  "MFG Production Completed: " + ev.ProductionOrderID,
			},
			{
				AccountID:    wipAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: ev.TotalValuation,
				Description:  "Transfer WIP to Finished Goods",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "MFG-COMP-"+ev.ProductionOrderID, "MFG Update Finished Goods and WIP", lines)
		return err

	case domain.TopicMfgMaterialConsumed:
		var ev domain.MaterialConsumedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit WIP, Credit Raw Materials Inventory
		wipAcc, err := c.getOrCreateAccount(ctx, "1230-001", "Work in Progress", "ASSET", "", "USD")
		if err != nil {
			return err
		}
		rmAcc, err := c.getOrCreateAccount(ctx, "1200-001", "Raw Materials Inventory", "ASSET", "", "USD")
		if err != nil {
			return err
		}

		lines := []domain.JournalEntryLine{
			{
				AccountID:    wipAcc.ID,
				DebitAmount:  ev.TotalCost,
				CreditAmount: decimal.Zero,
				Description:  "Material consumption cost: " + ev.ProductID,
			},
			{
				AccountID:    rmAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: ev.TotalCost,
				Description:  "Deduct Raw Material for MFG PO: " + ev.ProductionOrderID,
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "MFG-CONS-"+ev.ProductionOrderID, "MFG Material Consumption", lines)
		return err

	case domain.TopicPrjProjectCreated:
		var ev domain.ProjectCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Create project expense account
		accNum := "6030-" + ev.ProjectID[:8]
		accName := "Project Expense - " + ev.ProjectName
		_, err := c.getOrCreateAccount(ctx, accNum, accName, "EXPENSE", "", "USD")
		return err

	case domain.TopicPrjTimeLogged:
		var ev domain.TimeLoggedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Unbilled Receivables, Credit Project Revenue
		unbilledAcc, err := c.getOrCreateAccount(ctx, "1110-001", "Unbilled Receivables", "ASSET", "", "USD")
		if err != nil {
			return err
		}
		revAcc, err := c.getOrCreateAccount(ctx, "4010-001", "Project Revenue", "REVENUE", "", "USD")
		if err != nil {
			return err
		}

		val := ev.HoursLogged.Mul(ev.BillableRate)
		lines := []domain.JournalEntryLine{
			{
				AccountID:    unbilledAcc.ID,
				DebitAmount:  val,
				CreditAmount: decimal.Zero,
				Description:  "Unbilled revenue for Project: " + ev.ProjectID,
			},
			{
				AccountID:    revAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: val,
				Description:  "Recognize project revenue",
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "PRJ-TIME-"+ev.TimeLogID, "Record project billable time entries", lines)
		return err

	case domain.TopicPrjExpenseIncurred:
		var ev domain.ProjectExpenseIncurredEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		// Debit Project Expenses Control, Credit Accounts Payable - Control
		expenseAcc, err := c.getOrCreateAccount(ctx, "6030-001", "Project Expenses Control", "EXPENSE", "", "USD")
		if err != nil {
			return err
		}
		payableAcc, err := c.getOrCreateAccount(ctx, "2110-001", "Accounts Payable - Control", "LIABILITY", "", "USD")
		if err != nil {
			return err
		}

		lines := []domain.JournalEntryLine{
			{
				AccountID:    expenseAcc.ID,
				DebitAmount:  ev.Amount,
				CreditAmount: decimal.Zero,
				Description:  ev.Description + " (Project: " + ev.ProjectID + ")",
			},
			{
				AccountID:    payableAcc.ID,
				DebitAmount:  decimal.Zero,
				CreditAmount: ev.Amount,
				Description:  "Project Expense AP Liability: " + ev.ExpenseID,
			},
		}
		_, err = c.gl.CreateJournalEntry(ctx, "PRJ-EXP-"+ev.ExpenseID, "Record project expenses incurred", lines)
		return err
	}

	return nil
}

// Close stops the reader
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

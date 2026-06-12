package sql

import (
	"encoding/json"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)



// Payment GORM struct
type Payment struct {
	ID            string          `gorm:"primaryKey"`
	InvoiceID     *string         `gorm:"index"`
	BillID        *string         `gorm:"index"`
	BankAccountID *string         `gorm:"index"`
	PaymentNumber string
	PaymentDate   time.Time
	Amount        decimal.Decimal `gorm:"type:numeric(18,4)"`
	PaymentMethod string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Invoice     *ArInvoice     `gorm:"foreignKey:InvoiceID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	VendorBill  *ApVendorBill  `gorm:"foreignKey:BillID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	BankAccount *BankAccount `gorm:"foreignKey:BankAccountID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainPayment(d *domain.Payment) *Payment {
	if d == nil {
		return nil
	}
	return &Payment{
		ID:            d.ID,
		InvoiceID:     d.InvoiceID,
		BillID:        d.BillID,
		BankAccountID: d.BankAccountID,
		PaymentNumber: d.PaymentNumber,
		PaymentDate:   d.PaymentDate,
		Amount:        d.Amount,
		PaymentMethod: d.PaymentMethod,
		Status:        d.Status,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func ToDomainPayment(dbModel *Payment) *domain.Payment {
	if dbModel == nil {
		return nil
	}
	return &domain.Payment{
		ID:            dbModel.ID,
		InvoiceID:     dbModel.InvoiceID,
		BillID:        dbModel.BillID,
		BankAccountID: dbModel.BankAccountID,
		PaymentNumber: dbModel.PaymentNumber,
		PaymentDate:   dbModel.PaymentDate,
		Amount:        dbModel.Amount,
		PaymentMethod: dbModel.PaymentMethod,
		Status:        dbModel.Status,
		CreatedAt:     dbModel.CreatedAt,
		UpdatedAt:     dbModel.UpdatedAt,
	}
}

// Budget GORM struct
type Budget struct {
	ID              string          `gorm:"primaryKey"`
	AccountID       string          `gorm:"index"`
	CostCenterID    *string         `gorm:"index"`
	FiscalYear      int
	Period          int
	AllocatedAmount decimal.Decimal `gorm:"type:numeric(18,4)"`
	SpentAmount     decimal.Decimal `gorm:"type:numeric(18,4)"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Account    ChartOfAccounts `gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	CostCenter *CostCenter     `gorm:"foreignKey:CostCenterID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainBudget(d *domain.Budget) *Budget {
	if d == nil {
		return nil
	}
	return &Budget{
		ID:              d.ID,
		AccountID:       d.AccountID,
		CostCenterID:    d.CostCenterID,
		FiscalYear:      d.FiscalYear,
		Period:          d.Period,
		AllocatedAmount: d.AllocatedAmount,
		SpentAmount:     d.SpentAmount,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func ToDomainBudget(dbModel *Budget) *domain.Budget {
	if dbModel == nil {
		return nil
	}
	return &domain.Budget{
		ID:              dbModel.ID,
		AccountID:       dbModel.AccountID,
		CostCenterID:    dbModel.CostCenterID,
		FiscalYear:      dbModel.FiscalYear,
		Period:          dbModel.Period,
		AllocatedAmount: dbModel.AllocatedAmount,
		SpentAmount:     dbModel.SpentAmount,
		CreatedAt:       dbModel.CreatedAt,
		UpdatedAt:       dbModel.UpdatedAt,
	}
}



// TaxRate GORM struct
type TaxRate struct {
	ID       string          `gorm:"primaryKey"`
	Code     string          `gorm:"uniqueIndex"`
	Name     string
	Rate     decimal.Decimal `gorm:"type:numeric(18,4)"`
	IsActive bool
}

func FromDomainTaxRate(d *domain.TaxRate) *TaxRate {
	if d == nil {
		return nil
	}
	return &TaxRate{
		ID:       d.ID,
		Code:     d.Code,
		Name:     d.Name,
		Rate:     d.Rate,
		IsActive: d.IsActive,
	}
}

func ToDomainTaxRate(dbModel *TaxRate) *domain.TaxRate {
	if dbModel == nil {
		return nil
	}
	return &domain.TaxRate{
		ID:       dbModel.ID,
		Code:     dbModel.Code,
		Name:     dbModel.Name,
		Rate:     dbModel.Rate,
		IsActive: dbModel.IsActive,
	}
}

// CurrencyRate GORM struct
type CurrencyRate struct {
	ID            string          `gorm:"primaryKey"`
	FromCurrency  string
	ToCurrency    string
	Rate          decimal.Decimal `gorm:"type:numeric(18,4)"`
	EffectiveDate time.Time
}

func FromDomainCurrencyRate(d *domain.CurrencyRate) *CurrencyRate {
	if d == nil {
		return nil
	}
	return &CurrencyRate{
		ID:            d.ID,
		FromCurrency:  d.FromCurrency,
		ToCurrency:    d.ToCurrency,
		Rate:          d.Rate,
		EffectiveDate: d.EffectiveDate,
	}
}

func ToDomainCurrencyRate(dbModel *CurrencyRate) *domain.CurrencyRate {
	if dbModel == nil {
		return nil
	}
	return &domain.CurrencyRate{
		ID:            dbModel.ID,
		FromCurrency:  dbModel.FromCurrency,
		ToCurrency:    dbModel.ToCurrency,
		Rate:          dbModel.Rate,
		EffectiveDate: dbModel.EffectiveDate,
	}
}

// FiscalYear GORM struct
type FiscalYear struct {
	ID        string    `gorm:"primaryKey"`
	Year      int       `gorm:"uniqueIndex"`
	StartDate time.Time
	EndDate   time.Time
	IsClosed  bool
}

func FromDomainFiscalYear(d *domain.FiscalYear) *FiscalYear {
	if d == nil {
		return nil
	}
	return &FiscalYear{
		ID:        d.ID,
		Year:      d.Year,
		StartDate: d.StartDate,
		EndDate:   d.EndDate,
		IsClosed:  d.IsClosed,
	}
}

func ToDomainFiscalYear(dbModel *FiscalYear) *domain.FiscalYear {
	if dbModel == nil {
		return nil
	}
	return &domain.FiscalYear{
		ID:        dbModel.ID,
		Year:      dbModel.Year,
		StartDate: dbModel.StartDate,
		EndDate:   dbModel.EndDate,
		IsClosed:  dbModel.IsClosed,
	}
}

// CostCenter GORM struct
type CostCenter struct {
	ID          string `gorm:"primaryKey"`
	Code        string `gorm:"uniqueIndex"`
	Name        string
	Description string
	ManagerID   *string
	IsActive    bool
}

func FromDomainCostCenter(d *domain.CostCenter) *CostCenter {
	if d == nil {
		return nil
	}
	return &CostCenter{
		ID:          d.ID,
		Code:        d.Code,
		Name:        d.Name,
		Description: d.Description,
		ManagerID:   d.ManagerID,
		IsActive:    d.IsActive,
	}
}

func ToDomainCostCenter(dbModel *CostCenter) *domain.CostCenter {
	if dbModel == nil {
		return nil
	}
	return &domain.CostCenter{
		ID:          dbModel.ID,
		Code:        dbModel.Code,
		Name:        dbModel.Name,
		Description: dbModel.Description,
		ManagerID:   dbModel.ManagerID,
		IsActive:    dbModel.IsActive,
	}
}

// BankAccount GORM struct
type BankAccount struct {
	ID            string          `gorm:"primaryKey"`
	LegalEntityID string          `gorm:"index"`
	AccountNumber string          `gorm:"uniqueIndex"`
	Currency      string
	LiquidBalance decimal.Decimal `gorm:"type:numeric(18,4)"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	LegalEntity LegalEntity `gorm:"foreignKey:LegalEntityID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainBankAccount(d *domain.BankAccount) *BankAccount {
	if d == nil {
		return nil
	}
	return &BankAccount{
		ID:            d.ID,
		LegalEntityID: d.LegalEntityID,
		AccountNumber: d.AccountNumber,
		Currency:      d.Currency,
		LiquidBalance: d.LiquidBalance,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func ToDomainBankAccount(dbModel *BankAccount) *domain.BankAccount {
	if dbModel == nil {
		return nil
	}
	return &domain.BankAccount{
		ID:            dbModel.ID,
		LegalEntityID: dbModel.LegalEntityID,
		AccountNumber: dbModel.AccountNumber,
		Currency:      dbModel.Currency,
		LiquidBalance: dbModel.LiquidBalance,
		CreatedAt:     dbModel.CreatedAt,
		UpdatedAt:     dbModel.UpdatedAt,
	}
}

// CustomerCredit GORM struct
type CustomerCredit struct {
	ID             string          `gorm:"primaryKey"`
	CustomerID     string          `gorm:"uniqueIndex"`
	CreditLimit    decimal.Decimal `gorm:"type:numeric(18,4)"`
	CurrentBalance decimal.Decimal `gorm:"type:numeric(18,4)"`
	IsOnHold       bool
	UpdatedAt      time.Time
}

func FromDomainCustomerCredit(d *domain.CustomerCredit) *CustomerCredit {
	if d == nil {
		return nil
	}
	return &CustomerCredit{
		ID:             d.ID,
		CustomerID:     d.CustomerID,
		CreditLimit:    d.CreditLimit,
		CurrentBalance: d.CurrentBalance,
		IsOnHold:       d.IsOnHold,
		UpdatedAt:      d.UpdatedAt,
	}
}

func ToDomainCustomerCredit(dbModel *CustomerCredit) *domain.CustomerCredit {
	if dbModel == nil {
		return nil
	}
	return &domain.CustomerCredit{
		ID:             dbModel.ID,
		CustomerID:     dbModel.CustomerID,
		CreditLimit:    dbModel.CreditLimit,
		CurrentBalance: dbModel.CurrentBalance,
		IsOnHold:       dbModel.IsOnHold,
		UpdatedAt:      dbModel.UpdatedAt,
	}
}

// BankStatement GORM struct
type BankStatement struct {
	ID            string          `gorm:"primaryKey"`
	BankAccountID string          `gorm:"index"`
	StatementDate time.Time
	EndingBalance decimal.Decimal `gorm:"type:numeric(18,4)"`
	IsReconciled  bool

	BankAccount BankAccount `gorm:"foreignKey:BankAccountID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainBankStatement(d *domain.BankStatement) *BankStatement {
	if d == nil {
		return nil
	}
	return &BankStatement{
		ID:            d.ID,
		BankAccountID: d.BankAccountID,
		StatementDate: d.StatementDate,
		EndingBalance: d.EndingBalance,
		IsReconciled:  d.IsReconciled,
	}
}

func ToDomainBankStatement(dbModel *BankStatement) *domain.BankStatement {
	if dbModel == nil {
		return nil
	}
	return &domain.BankStatement{
		ID:            dbModel.ID,
		BankAccountID: dbModel.BankAccountID,
		StatementDate: dbModel.StatementDate,
		EndingBalance: dbModel.EndingBalance,
		IsReconciled:  dbModel.IsReconciled,
	}
}

// BankStatementLine GORM struct
type BankStatementLine struct {
	ID              string          `gorm:"primaryKey"`
	StatementID     string          `gorm:"index"`
	TransactionDate time.Time
	Description     string
	Amount          decimal.Decimal `gorm:"type:numeric(18,4)"`
	IsMatched       bool

	BankStatement BankStatement `gorm:"foreignKey:StatementID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func FromDomainBankStatementLine(d *domain.BankStatementLine) *BankStatementLine {
	if d == nil {
		return nil
	}
	return &BankStatementLine{
		ID:              d.ID,
		StatementID:     d.StatementID,
		TransactionDate: d.TransactionDate,
		Description:     d.Description,
		Amount:          d.Amount,
		IsMatched:       d.IsMatched,
	}
}

func ToDomainBankStatementLine(dbModel *BankStatementLine) *domain.BankStatementLine {
	if dbModel == nil {
		return nil
	}
	return &domain.BankStatementLine{
		ID:              dbModel.ID,
		StatementID:     dbModel.StatementID,
		TransactionDate: dbModel.TransactionDate,
		Description:     dbModel.Description,
		Amount:          dbModel.Amount,
		IsMatched:       dbModel.IsMatched,
	}
}



// ChartOfAccounts GORM struct
type ChartOfAccounts struct {
	ID            string             `gorm:"primaryKey"`
	LegalEntityID string             `gorm:"uniqueIndex:idx_entity_account"`
	AccountCode   string             `gorm:"uniqueIndex:idx_entity_account"`
	AccountName   string
	Type          domain.AccountType `gorm:"type:varchar(50)"`
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time

	LegalEntity LegalEntity `gorm:"foreignKey:LegalEntityID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainChartOfAccounts(d *domain.ChartOfAccounts) *ChartOfAccounts {
	if d == nil {
		return nil
	}
	return &ChartOfAccounts{
		ID:            d.ID,
		LegalEntityID: d.LegalEntityID,
		AccountCode:   d.AccountCode,
		AccountName:   d.AccountName,
		Type:          d.Type,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func ToDomainChartOfAccounts(dbModel *ChartOfAccounts) *domain.ChartOfAccounts {
	if dbModel == nil {
		return nil
	}
	return &domain.ChartOfAccounts{
		ID:            dbModel.ID,
		LegalEntityID: dbModel.LegalEntityID,
		AccountCode:   dbModel.AccountCode,
		AccountName:   dbModel.AccountName,
		Type:          dbModel.Type,
		IsActive:      dbModel.IsActive,
		CreatedAt:     dbModel.CreatedAt,
		UpdatedAt:     dbModel.UpdatedAt,
	}
}

// LegalEntity GORM struct
type LegalEntity struct {
	ID                    string    `gorm:"primaryKey"`
	CompanyCode           string    `gorm:"uniqueIndex"`
	CompanyName           string
	FunctionalCurrency    string
	TaxRegistrationNumber string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func FromDomainLegalEntity(d *domain.LegalEntity) *LegalEntity {
	if d == nil {
		return nil
	}
	return &LegalEntity{
		ID:                    d.ID,
		CompanyCode:           d.CompanyCode,
		CompanyName:           d.CompanyName,
		FunctionalCurrency:    d.FunctionalCurrency,
		TaxRegistrationNumber: d.TaxRegistrationNumber,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}
}

func ToDomainLegalEntity(dbModel *LegalEntity) *domain.LegalEntity {
	if dbModel == nil {
		return nil
	}
	return &domain.LegalEntity{
		ID:                    dbModel.ID,
		CompanyCode:           dbModel.CompanyCode,
		CompanyName:           dbModel.CompanyName,
		FunctionalCurrency:    dbModel.FunctionalCurrency,
		TaxRegistrationNumber: dbModel.TaxRegistrationNumber,
		CreatedAt:             dbModel.CreatedAt,
		UpdatedAt:             dbModel.UpdatedAt,
	}
}

// UniversalJournalEntry GORM struct
type UniversalJournalEntry struct {
	ID               string             `gorm:"primaryKey"`
	LegalEntityID    string             `gorm:"index"`
	SourceModule     string
	SourceDocumentID string
	PostingDate      time.Time
	FinancialPeriod  string
	Status           domain.LedgerState `gorm:"type:varchar(50)"`
	CreatedAt        time.Time
	UpdatedAt        time.Time

	LegalEntity LegalEntity `gorm:"foreignKey:LegalEntityID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainUniversalJournalEntry(d *domain.UniversalJournalEntry) *UniversalJournalEntry {
	if d == nil {
		return nil
	}
	return &UniversalJournalEntry{
		ID:               d.ID,
		LegalEntityID:    d.LegalEntityID,
		SourceModule:     d.SourceModule,
		SourceDocumentID: d.SourceDocumentID,
		PostingDate:      d.PostingDate,
		FinancialPeriod:  d.FinancialPeriod,
		Status:           d.Status,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

func ToDomainUniversalJournalEntry(dbModel *UniversalJournalEntry) *domain.UniversalJournalEntry {
	if dbModel == nil {
		return nil
	}
	return &domain.UniversalJournalEntry{
		ID:               dbModel.ID,
		LegalEntityID:    dbModel.LegalEntityID,
		SourceModule:     dbModel.SourceModule,
		SourceDocumentID: dbModel.SourceDocumentID,
		PostingDate:      dbModel.PostingDate,
		FinancialPeriod:  dbModel.FinancialPeriod,
		Status:           dbModel.Status,
		CreatedAt:        dbModel.CreatedAt,
		UpdatedAt:        dbModel.UpdatedAt,
	}
}

// UniversalJournalLine GORM struct
type UniversalJournalLine struct {
	ID                    string          `gorm:"primaryKey"`
	JournalEntryID        string          `gorm:"index"`
	AccountID             string          `gorm:"index"`
	AmountFunctional      decimal.Decimal `gorm:"type:numeric(18,4)"`
	AmountTransactional   decimal.Decimal `gorm:"type:numeric(18,4)"`
	CurrencyTransactional string
	TrackingDimensions    []byte `gorm:"type:jsonb"` // Marshalled json of interface{}

	JournalEntry UniversalJournalEntry `gorm:"foreignKey:JournalEntryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Account      ChartOfAccounts       `gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainUniversalJournalLine(d *domain.UniversalJournalLine) *UniversalJournalLine {
	if d == nil {
		return nil
	}
	tdBytes, _ := json.Marshal(d.TrackingDimensions)
	return &UniversalJournalLine{
		ID:                    d.ID,
		JournalEntryID:        d.JournalEntryID,
		AccountID:             d.AccountID,
		AmountFunctional:      d.AmountFunctional,
		AmountTransactional:   d.AmountTransactional,
		CurrencyTransactional: d.CurrencyTransactional,
		TrackingDimensions:    tdBytes,
	}
}

func ToDomainUniversalJournalLine(dbModel *UniversalJournalLine) *domain.UniversalJournalLine {
	if dbModel == nil {
		return nil
	}
	var td interface{}
	if len(dbModel.TrackingDimensions) > 0 {
		_ = json.Unmarshal(dbModel.TrackingDimensions, &td)
	}
	return &domain.UniversalJournalLine{
		ID:                    dbModel.ID,
		JournalEntryID:        dbModel.JournalEntryID,
		AccountID:             dbModel.AccountID,
		AmountFunctional:      dbModel.AmountFunctional,
		AmountTransactional:   dbModel.AmountTransactional,
		CurrencyTransactional: dbModel.CurrencyTransactional,
		TrackingDimensions:    td,
	}
}

// DepreciationScheduleLine GORM struct
type DepreciationScheduleLine struct {
	ID                 string          `gorm:"primaryKey"`
	FixedAssetID       string          `gorm:"index"`
	FiscalYear         int
	PeriodNumber       int
	DepreciationAmount decimal.Decimal `gorm:"type:numeric(18,4)"`
	IsPosted           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time

	FixedAsset CapitalAsset `gorm:"foreignKey:FixedAssetID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainDepreciationScheduleLine(d *domain.DepreciationScheduleLine) *DepreciationScheduleLine {
	if d == nil {
		return nil
	}
	return &DepreciationScheduleLine{
		ID:                 d.ID,
		FixedAssetID:       d.FixedAssetID,
		FiscalYear:         d.FiscalYear,
		PeriodNumber:       d.PeriodNumber,
		DepreciationAmount: d.DepreciationAmount,
		IsPosted:           d.IsPosted,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
	}
}

func ToDomainDepreciationScheduleLine(dbModel *DepreciationScheduleLine) *domain.DepreciationScheduleLine {
	if dbModel == nil {
		return nil
	}
	return &domain.DepreciationScheduleLine{
		ID:                 dbModel.ID,
		FixedAssetID:       dbModel.FixedAssetID,
		FiscalYear:         dbModel.FiscalYear,
		PeriodNumber:       dbModel.PeriodNumber,
		DepreciationAmount: dbModel.DepreciationAmount,
		IsPosted:           dbModel.IsPosted,
		CreatedAt:          dbModel.CreatedAt,
		UpdatedAt:          dbModel.UpdatedAt,
	}
}

// CapitalAsset GORM struct
type CapitalAsset struct {
	ID                      string            `gorm:"primaryKey"`
	LegalEntityID           string            `gorm:"index"`
	AssetTag                string
	EamEquipmentID          *string           `gorm:"index"`
	AcquisitionCost         decimal.Decimal   `gorm:"type:numeric(18,4)"`
	AccumulatedDepreciation decimal.Decimal   `gorm:"type:numeric(18,4)"`
	UsefulLifeMonths        int
	CapitalizationDate      time.Time
	Status                  domain.AssetState `gorm:"type:varchar(50)"`
	CreatedAt               time.Time
	UpdatedAt               time.Time

	LegalEntity LegalEntity `gorm:"foreignKey:LegalEntityID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainCapitalAsset(d *domain.CapitalAsset) *CapitalAsset {
	if d == nil {
		return nil
	}
	return &CapitalAsset{
		ID:                      d.ID,
		LegalEntityID:           d.LegalEntityID,
		AssetTag:                d.AssetTag,
		EamEquipmentID:          d.EamEquipmentID,
		AcquisitionCost:         d.AcquisitionCost,
		AccumulatedDepreciation: d.AccumulatedDepreciation,
		UsefulLifeMonths:        d.UsefulLifeMonths,
		CapitalizationDate:      d.CapitalizationDate,
		Status:                  d.Status,
		CreatedAt:               d.CreatedAt,
		UpdatedAt:               d.UpdatedAt,
	}
}

func ToDomainCapitalAsset(dbModel *CapitalAsset) *domain.CapitalAsset {
	if dbModel == nil {
		return nil
	}
	return &domain.CapitalAsset{
		ID:                      dbModel.ID,
		LegalEntityID:           dbModel.LegalEntityID,
		AssetTag:                dbModel.AssetTag,
		EamEquipmentID:          dbModel.EamEquipmentID,
		AcquisitionCost:         dbModel.AcquisitionCost,
		AccumulatedDepreciation: dbModel.AccumulatedDepreciation,
		UsefulLifeMonths:        dbModel.UsefulLifeMonths,
		CapitalizationDate:      dbModel.CapitalizationDate,
		Status:                  dbModel.Status,
		CreatedAt:               dbModel.CreatedAt,
		UpdatedAt:               dbModel.UpdatedAt,
	}
}

// KafkaEventInbox GORM struct
type KafkaEventInbox struct {
	EventID          string                       `gorm:"primaryKey"`
	EventType        string
	ProcessedAt      time.Time
	ProcessingStatus domain.EventProcessingStatus `gorm:"type:varchar(50)"`
	Payload          []byte                       `gorm:"type:jsonb"` // Marshalled json of interface{}
}

func FromDomainKafkaEventInbox(d *domain.KafkaEventInbox) *KafkaEventInbox {
	if d == nil {
		return nil
	}
	pBytes, _ := json.Marshal(d.Payload)
	return &KafkaEventInbox{
		EventID:          d.EventID,
		EventType:        d.EventType,
		ProcessedAt:      d.ProcessedAt,
		ProcessingStatus: d.ProcessingStatus,
		Payload:          pBytes,
	}
}

func ToDomainKafkaEventInbox(dbModel *KafkaEventInbox) *domain.KafkaEventInbox {
	if dbModel == nil {
		return nil
	}
	var p interface{}
	if len(dbModel.Payload) > 0 {
		_ = json.Unmarshal(dbModel.Payload, &p)
	}
	return &domain.KafkaEventInbox{
		EventID:          dbModel.EventID,
		EventType:        dbModel.EventType,
		ProcessedAt:      dbModel.ProcessedAt,
		ProcessingStatus: dbModel.ProcessingStatus,
		Payload:          p,
	}
}

// TransactionalOutbox GORM struct
type TransactionalOutbox struct {
	ID          string              `gorm:"primaryKey"`
	EventType   string
	AggregateID string
	Payload     []byte              `gorm:"type:jsonb"` // Marshalled json of interface{}
	Status      domain.OutboxStatus `gorm:"type:varchar(50)"`
	CreatedAt   time.Time
}

func FromDomainTransactionalOutbox(d *domain.TransactionalOutbox) *TransactionalOutbox {
	if d == nil {
		return nil
	}
	pBytes, _ := json.Marshal(d.Payload)
	return &TransactionalOutbox{
		ID:          d.ID,
		EventType:   d.EventType,
		AggregateID: d.AggregateID,
		Payload:     pBytes,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
	}
}

func ToDomainTransactionalOutbox(dbModel *TransactionalOutbox) *domain.TransactionalOutbox {
	if dbModel == nil {
		return nil
	}
	var p interface{}
	if len(dbModel.Payload) > 0 {
		_ = json.Unmarshal(dbModel.Payload, &p)
	}
	return &domain.TransactionalOutbox{
		ID:          dbModel.ID,
		EventType:   dbModel.EventType,
		AggregateID: dbModel.AggregateID,
		Payload:     p,
		Status:      dbModel.Status,
		CreatedAt:   dbModel.CreatedAt,
	}
}

// ApVendorBill GORM struct
type ApVendorBill struct {
	ID              string               `gorm:"primaryKey"`
	LegalEntityID   string               `gorm:"index"`
	BillNumber      string
	VendorID        string
	PurchaseOrderID string
	TotalAmount     decimal.Decimal      `gorm:"type:numeric(18,4)"`
	TaxAmount       decimal.Decimal      `gorm:"type:numeric(18,4)"`
	DueDate         time.Time
	Status          domain.PaymentStatus `gorm:"type:varchar(50)"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	LegalEntity LegalEntity `gorm:"foreignKey:LegalEntityID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainApVendorBill(d *domain.ApVendorBill) *ApVendorBill {
	if d == nil {
		return nil
	}
	return &ApVendorBill{
		ID:              d.ID,
		LegalEntityID:   d.LegalEntityID,
		BillNumber:      d.BillNumber,
		VendorID:        d.VendorID,
		PurchaseOrderID: d.PurchaseOrderID,
		TotalAmount:     d.TotalAmount,
		TaxAmount:       d.TaxAmount,
		DueDate:         d.DueDate,
		Status:          d.Status,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func ToDomainApVendorBill(dbModel *ApVendorBill) *domain.ApVendorBill {
	if dbModel == nil {
		return nil
	}
	return &domain.ApVendorBill{
		ID:              dbModel.ID,
		LegalEntityID:   dbModel.LegalEntityID,
		BillNumber:      dbModel.BillNumber,
		VendorID:        dbModel.VendorID,
		PurchaseOrderID: dbModel.PurchaseOrderID,
		TotalAmount:     dbModel.TotalAmount,
		TaxAmount:       dbModel.TaxAmount,
		DueDate:         dbModel.DueDate,
		Status:          dbModel.Status,
		CreatedAt:       dbModel.CreatedAt,
		UpdatedAt:       dbModel.UpdatedAt,
	}
}

// ArInvoice GORM struct
type ArInvoice struct {
	ID            string               `gorm:"primaryKey"`
	LegalEntityID string               `gorm:"index"`
	InvoiceNumber string
	CustomerID    string
	SalesOrderID  string
	TotalAmount   decimal.Decimal      `gorm:"type:numeric(18,4)"`
	TaxAmount     decimal.Decimal      `gorm:"type:numeric(18,4)"`
	DueDate         time.Time
	Status          domain.PaymentStatus `gorm:"type:varchar(50)"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	LegalEntity LegalEntity `gorm:"foreignKey:LegalEntityID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func FromDomainArInvoice(d *domain.ArInvoice) *ArInvoice {
	if d == nil {
		return nil
	}
	return &ArInvoice{
		ID:            d.ID,
		LegalEntityID: d.LegalEntityID,
		InvoiceNumber: d.InvoiceNumber,
		CustomerID:    d.CustomerID,
		SalesOrderID:  d.SalesOrderID,
		TotalAmount:   d.TotalAmount,
		TaxAmount:     d.TaxAmount,
		DueDate:       d.DueDate,
		Status:        d.Status,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

func ToDomainArInvoice(dbModel *ArInvoice) *domain.ArInvoice {
	if dbModel == nil {
		return nil
	}
	return &domain.ArInvoice{
		ID:            dbModel.ID,
		LegalEntityID: dbModel.LegalEntityID,
		InvoiceNumber: dbModel.InvoiceNumber,
		CustomerID:    dbModel.CustomerID,
		SalesOrderID:  dbModel.SalesOrderID,
		TotalAmount:   dbModel.TotalAmount,
		TaxAmount:     dbModel.TaxAmount,
		DueDate:       dbModel.DueDate,
		Status:        dbModel.Status,
		CreatedAt:     dbModel.CreatedAt,
		UpdatedAt:     dbModel.UpdatedAt,
	}
}

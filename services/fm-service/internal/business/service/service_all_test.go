package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/erp-system/fm-service/internal/business/service"
	"github.com/erp-system/fm-service/internal/data/memory"
	"github.com/shopspring/decimal"
)

func TestTaxService_All(t *testing.T) {
	repo := memory.NewMemoryTaxRateRepo()
	svc := service.NewTaxService(repo)
	ctx := context.Background()

	// 1. Validation error: empty code
	_, err := svc.CreateTaxRate(ctx, "", "VAT", decimal.NewFromFloat(0.15))
	if err == nil {
		t.Error("expected error for empty tax code, got nil")
	}

	// 2. Success path
	rate, err := svc.CreateTaxRate(ctx, "VAT15", "VAT 15%", decimal.NewFromFloat(0.15))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rate.ID != "tax_VAT15" || rate.Code != "VAT15" || !rate.Rate.Equal(decimal.NewFromFloat(0.15)) {
		t.Errorf("unexpected tax rate values: %+v", rate)
	}

	// 3. GetTaxRate
	retrieved, err := svc.GetTaxRate(ctx, "tax_VAT15")
	if err != nil {
		t.Fatalf("unexpected error getting tax rate: %v", err)
	}
	if retrieved.Name != "VAT 15%" {
		t.Errorf("expected VAT 15%%, got %s", retrieved.Name)
	}

	// Get non-existent
	_, err = svc.GetTaxRate(ctx, "tax_non_existent")
	if err == nil {
		t.Error("expected error for non-existent tax rate, got nil")
	}

	// 4. ListTaxRates
	list, err := svc.ListTaxRates(ctx)
	if err != nil {
		t.Fatalf("unexpected error listing tax rates: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 tax rate in list, got %d", len(list))
	}
}

func TestBudgetingService_All(t *testing.T) {
	budgets := memory.NewMemoryBudgetRepo()
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(budgets, outbox)

	svc := service.NewBudgetingService(budgets, accounts, entries, outbox, tm)
	ctx := context.Background()

	// 1. Input validations
	_, err := svc.CreateBudget(ctx, "", "", 2026, 6, decimal.NewFromInt(1000))
	if err == nil {
		t.Error("expected error for empty accountID")
	}
	_, err = svc.CreateBudget(ctx, "acc_1", "", 0, 6, decimal.NewFromInt(1000))
	if err == nil {
		t.Error("expected error for non-positive year")
	}
	_, err = svc.CreateBudget(ctx, "acc_1", "", 2026, 0, decimal.NewFromInt(1000))
	if err == nil {
		t.Error("expected error for period < 1")
	}
	_, err = svc.CreateBudget(ctx, "acc_1", "", 2026, 13, decimal.NewFromInt(1000))
	if err == nil {
		t.Error("expected error for period > 12")
	}

	// 2. Successful budget creation
	costCenter := "cc_marketing"
	bud, err := svc.CreateBudget(ctx, "acc_1", costCenter, 2026, 6, decimal.NewFromInt(10000))
	if err != nil {
		t.Fatalf("unexpected error creating budget: %v", err)
	}
	if bud.AccountID != "acc_1" || *bud.CostCenterID != costCenter || bud.FiscalYear != 2026 || bud.Period != 6 {
		t.Errorf("unexpected budget values: %+v", bud)
	}

	// List Budgets
	buds, err := svc.ListBudgets(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buds) != 1 {
		t.Errorf("expected 1 budget, got %d", len(buds))
	}

	// Check and Track budget expense (exceeded)
	err = svc.CheckAndTrackBudgetExpense(ctx, "acc_1", decimal.NewFromInt(12000), 2026, 6)
	if err != nil {
		t.Fatalf("unexpected error tracking expense: %v", err)
	}

	// Verify budget is updated
	updatedBud, err := budgets.GetByAccountAndPeriod(ctx, "acc_1", 2026, 6)
	if err != nil {
		t.Fatalf("unexpected error getting budget: %v", err)
	}
	if !updatedBud.SpentAmount.Equal(decimal.NewFromInt(12000)) {
		t.Errorf("expected spent amount to be 12000, got %s", updatedBud.SpentAmount)
	}

	// Check outbox has BudgetExceeded and BudgetUpdated events
	pendingEvents, err := outbox.GetPending(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pendingEvents) != 3 {
		t.Errorf("expected 3 events in outbox, got %d", len(pendingEvents))
	}

	// Test CheckAndTrackBudgetExpense with non-existent budget (should just ignore and return nil)
	err = svc.CheckAndTrackBudgetExpense(ctx, "acc_non_existent", decimal.NewFromInt(500), 2026, 12)
	if err != nil {
		t.Errorf("expected nil error for non-existent budget tracking, got %v", err)
	}

	// GetBudgetVsActualReport - Account not found
	_, err = svc.GetBudgetVsActualReport(ctx, "acc_1", 2026)
	if err == nil {
		t.Error("expected error for non-existent account in report, got nil")
	}

	// Create Account and get report
	acc := &domain.ChartOfAccounts{
		ID:            "acc_1",
		LegalEntityID: "legal_123",
		AccountCode:   "1100",
		AccountName:   "Marketing Expense",
		Type:          domain.AccountTypeEXPENSE,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_ = accounts.Create(ctx, acc)

	// Since we dynamic balance based on entries, let's create a posted entry for acc_1 to simulate spent amount
	draftEntry := &domain.UniversalJournalEntry{
		ID:            "je_budget_spent",
		LegalEntityID: "legal_123",
		PostingDate:   time.Now(),
		Status:        domain.LedgerStatePOSTED,
	}
	lines := []domain.UniversalJournalLine{
		{AccountID: "acc_1", AmountFunctional: decimal.NewFromInt(12000)},
		{AccountID: "acc_offset", AmountFunctional: decimal.NewFromInt(-12000)},
	}
	_ = entries.Create(ctx, draftEntry, lines)

	report, err := svc.GetBudgetVsActualReport(ctx, "acc_1", 2026)
	if err != nil {
		t.Fatalf("unexpected error getting report: %v", err)
	}
	if report["account_number"] != "1100" || !report["budget_amount"].(decimal.Decimal).Equal(decimal.NewFromInt(10000)) || !report["actual_spent"].(decimal.Decimal).Equal(decimal.NewFromInt(12000)) {
		t.Errorf("unexpected report content: %+v", report)
	}
}

func TestAccountsPayableService_All(t *testing.T) {
	bills := memory.NewMemoryApVendorBillRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(bills, outbox)

	svc := service.NewAccountsPayableService(bills, outbox, tm)
	ctx := context.Background()

	// MatchPurchaseOrder
	matched, err := svc.MatchPurchaseOrder(ctx, "b1", "po1", "gr1")
	if err != nil || !matched {
		t.Errorf("expected PO match to return true, nil; got %t, %v", matched, err)
	}

	_, err = svc.MatchPurchaseOrder(ctx, "", "po1", "gr1")
	if err == nil {
		t.Error("expected error for empty match IDs, got nil")
	}

	// CreateVendorBill
	poID := "po_123"
	bill, err := svc.CreateVendorBill(ctx, "legal_123", "supplier_1", "BILL-100", poID, time.Now().AddDate(0, 0, 30), decimal.NewFromInt(150), decimal.Zero)
	if err != nil {
		t.Fatalf("unexpected error creating bill: %v", err)
	}
	if bill.VendorID != "supplier_1" || bill.PurchaseOrderID != poID || bill.BillNumber != "BILL-100" || bill.Status != domain.PaymentStatusOPEN {
		t.Errorf("unexpected bill values: %+v", bill)
	}

	// ListVendorBills
	list, err := svc.ListVendorBills(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 bill, got %d", len(list))
	}

	// GetVendorBill
	retrieved, err := svc.GetVendorBill(ctx, bill.ID)
	if err != nil {
		t.Fatalf("unexpected error getting bill: %v", err)
	}
	if retrieved.ID != bill.ID {
		t.Errorf("unexpected retrieved bill ID: %s", retrieved.ID)
	}
}

func TestAccountsReceivableService_All(t *testing.T) {
	invoices := memory.NewMemoryArInvoiceRepo()
	credits := memory.NewMemoryCustomerCreditRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(invoices, outbox)

	svc := service.NewAccountsReceivableService(invoices, credits, outbox, tm)
	ctx := context.Background()

	// CheckCustomerCredit
	allowed, err := svc.CheckCustomerCredit(ctx, "cust_1", decimal.NewFromInt(5000))
	if err != nil || !allowed {
		t.Errorf("expected credit check true, got %t, %v", allowed, err)
	}
	allowed, err = svc.CheckCustomerCredit(ctx, "cust_1", decimal.NewFromInt(150000))
	if err != nil || allowed {
		t.Errorf("expected credit check false, got %t, %v", allowed, err)
	}

	// CreateInvoice
	inv, err := svc.CreateInvoice(ctx, "legal_123", "cust_1", "so_123", decimal.NewFromInt(200), decimal.Zero, time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("unexpected error creating invoice: %v", err)
	}
	if inv.CustomerID != "cust_1" || !inv.TotalAmount.Equal(decimal.NewFromInt(200)) || inv.Status != domain.PaymentStatusOPEN {
		t.Errorf("unexpected invoice fields: %+v", inv)
	}

	// ListInvoices
	list, err := svc.ListInvoices(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 invoice, got %d", len(list))
	}

	// GetInvoice
	retrieved, err := svc.GetInvoice(ctx, inv.ID)
	if err != nil {
		t.Fatalf("unexpected error getting invoice: %v", err)
	}
	if retrieved.ID != inv.ID {
		t.Errorf("unexpected retrieved invoice ID: %s", retrieved.ID)
	}

	// UpdateInvoice
	fields := map[string]interface{}{
		"status": "PAID",
	}
	updated, err := svc.UpdateInvoice(ctx, inv.ID, fields)
	if err != nil {
		t.Fatalf("unexpected error updating invoice: %v", err)
	}
	if updated.Status != "PAID" {
		t.Errorf("expected status PAID, got %s", updated.Status)
	}

	// UpdateInvoice non-existent
	_, err = svc.UpdateInvoice(ctx, "non_existent_inv", fields)
	if err == nil {
		t.Error("expected error for updating non-existent invoice, got nil")
	}

	// SendInvoice
	ok, err := svc.SendInvoice(ctx, inv.ID)
	if err != nil || !ok {
		t.Errorf("expected SendInvoice successful, got %t, %v", ok, err)
	}
	sentInv, _ := svc.GetInvoice(ctx, inv.ID)
	if sentInv.Status != domain.PaymentStatusOPEN {
		t.Errorf("expected status OPEN, got %s", sentInv.Status)
	}

	// SendInvoice non-existent
	_, err = svc.SendInvoice(ctx, "non_existent_inv")
	if err == nil {
		t.Error("expected error for sending non-existent invoice, got nil")
	}

	// MarkInvoiceOverdue
	err = svc.MarkInvoiceOverdue(ctx, inv.ID)
	if err != nil {
		t.Fatalf("unexpected error marking overdue: %v", err)
	}
	overdueInv, _ := svc.GetInvoice(ctx, inv.ID)
	if overdueInv.Status != domain.PaymentStatusOPEN {
		t.Errorf("expected status OPEN, got %s", overdueInv.Status)
	}

	// MarkInvoiceOverdue non-existent
	err = svc.MarkInvoiceOverdue(ctx, "non_existent_inv")
	if err == nil {
		t.Error("expected error for non-existent invoice in mark overdue, got nil")
	}

	// DeleteInvoice
	err = svc.DeleteInvoice(ctx, inv.ID)
	if err != nil {
		t.Fatalf("unexpected error deleting invoice: %v", err)
	}
}

func TestCashManagementService_All(t *testing.T) {
	payments := memory.NewMemoryPaymentRepo()
	invoices := memory.NewMemoryArInvoiceRepo()
	statements := memory.NewMemoryBankStatementRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(payments, invoices, outbox)

	svc := service.NewCashManagementService(payments, invoices, statements, outbox, tm)
	ctx := context.Background()

	// GetBankStatement - missing stmt
	_, _, err := svc.GetBankStatement(ctx, "stmt_1")
	if err == nil {
		t.Error("expected error getting missing statement, got nil")
	}

	// RecordPayment - Successful with InvoiceID
	invRepo := memory.NewMemoryArInvoiceRepo()
	credits := memory.NewMemoryCustomerCreditRepo()
	invSvc := service.NewAccountsReceivableService(invRepo, credits, outbox, tm)
	inv, _ := invSvc.CreateInvoice(ctx, "legal_123", "cust_1", "so_123", decimal.NewFromInt(100), decimal.Zero, time.Now().AddDate(0, 0, 10))

	// Update svc with the same invoice repo
	svc = service.NewCashManagementService(payments, invRepo, statements, outbox, tm)

	pay, err := svc.RecordPayment(ctx, inv.ID, "bill_1", "bank_1", decimal.NewFromInt(100), "WIRE")
	if err != nil {
		t.Fatalf("unexpected error recording payment: %v", err)
	}
	if pay.Amount.Equal(decimal.NewFromInt(100)) == false || *pay.InvoiceID != inv.ID || pay.Status != "COMPLETED" {
		t.Errorf("unexpected payment values: %+v", pay)
	}

	// Verify invoice is marked paid
	paidInv, _ := invRepo.GetByID(ctx, inv.ID)
	if paidInv.Status != domain.PaymentStatusPAID {
		t.Errorf("expected invoice status to be PAID, got %s", paidInv.Status)
	}

	// ListPayments
	payList, err := svc.ListPayments(ctx)
	if err != nil {
		t.Fatalf("unexpected error listing payments: %v", err)
	}
	if len(payList) != 1 {
		t.Errorf("expected 1 payment in list, got %d", len(payList))
	}

	// GetPayment
	retrievedPay, err := svc.GetPayment(ctx, pay.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrievedPay.ID != pay.ID {
		t.Errorf("retrieved payment ID mismatch")
	}

	// ReconcileBankStatement
	err = svc.ReconcileBankStatement(ctx, "stmt_123")
	if err != nil {
		t.Errorf("expected nil error for reconcile statement, got %v", err)
	}

	// GetCashFlowForecast
	forecast, err := svc.GetCashFlowForecast(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if forecast["forecast_period_months"] != 3 || !forecast["net_cash_flow"].(decimal.Decimal).Equal(decimal.NewFromInt(45000)) {
		t.Errorf("unexpected forecast values: %+v", forecast)
	}

	// RecordPayment failing - invalid invoice ID triggers error inside transaction, verify outbox failed event
	_, err = svc.RecordPayment(ctx, "non_existent_inv", "", "", decimal.NewFromInt(100), "CASH")
	if err == nil {
		t.Error("expected error for non-existent invoice payment recording, got nil")
	}

	// Verify outbox contains payment failed event
	pendingEvents, _ := outbox.GetPending(ctx, 50)
	hasFailedEvent := false
	for _, ev := range pendingEvents {
		if ev.EventType == string(domain.TopicFmPaymentFailed) {
			hasFailedEvent = true
			break
		}
	}
	if !hasFailedEvent {
		t.Error("expected to find a PaymentFailed event in outbox after transaction rollback")
	}
}

func TestGeneralLedgerService_All(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)

	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)
	ctx := context.Background()

	// 1. CreateAccount validation
	_, err := svc.CreateAccount(ctx, "legal_123", "", "Cash", "ASSET")
	if err == nil {
		t.Error("expected error for empty account code")
	}
	_, err = svc.CreateAccount(ctx, "legal_123", "1000", "Cash", "INVALID_TYPE")
	if err == nil {
		t.Error("expected error for invalid account type")
	}

	// 2. CreateAccount success
	accA, err := svc.CreateAccount(ctx, "legal_123", "1000", "Cash", "ASSET")
	if err != nil {
		t.Fatalf("failed to create accA: %v", err)
	}
	accB, err := svc.CreateAccount(ctx, "legal_123", "4000", "Revenue", "REVENUE")
	if err != nil {
		t.Fatalf("failed to create accB: %v", err)
	}

	// GetAccount
	retAcc, err := svc.GetAccount(ctx, accA.ID)
	if err != nil || retAcc.AccountCode != "1000" {
		t.Errorf("error getting account by ID: %v", err)
	}

	// GetAccountByCode
	retAccNum, err := svc.GetAccountByCode(ctx, "legal_123", "4000")
	if err != nil || retAccNum.ID != accB.ID {
		t.Errorf("error getting account by code: %v", err)
	}

	// ListAccounts
	accList, err := svc.ListAccounts(ctx)
	if err != nil || len(accList) != 2 {
		t.Errorf("error listing accounts: %v", err)
	}

	// UpdateAccount
	updatedAcc, err := svc.UpdateAccount(ctx, accA.ID, "Petty Cash", "ASSET", true)
	if err != nil {
		t.Fatalf("failed to update account: %v", err)
	}
	if updatedAcc.AccountName != "Petty Cash" {
		t.Errorf("unexpected updated account name: %s", updatedAcc.AccountName)
	}

	// UpdateAccount invalid type
	_, err = svc.UpdateAccount(ctx, accA.ID, "Petty Cash", "INVALID_TYPE", true)
	if err == nil {
		t.Error("expected error updating with invalid type, got nil")
	}

	// GetAccountBalance
	bal, err := svc.GetAccountBalance(ctx, accA.ID)
	if err != nil || !bal.IsZero() {
		t.Errorf("expected zero balance, got %s, %v", bal, err)
	}

	// 3. CreateJournalEntry validation
	linesEmpty := []domain.UniversalJournalLine{}
	_, err = svc.CreateJournalEntry(ctx, "legal_123", "FM", "JE-1", time.Now(), linesEmpty)
	if err == nil {
		t.Error("expected error for empty lines in journal entry")
	}

	linesUnbalanced := []domain.UniversalJournalLine{
		{AccountID: accA.ID, AmountFunctional: decimal.NewFromInt(100), AmountTransactional: decimal.NewFromInt(100), CurrencyTransactional: "USD"},
		{AccountID: accB.ID, AmountFunctional: decimal.NewFromInt(-120), AmountTransactional: decimal.NewFromInt(-120), CurrencyTransactional: "USD"},
	}
	_, err = svc.CreateJournalEntry(ctx, "legal_123", "FM", "JE-2", time.Now(), linesUnbalanced)
	if err == nil {
		t.Error("expected error for unbalanced journal entry")
	}

	// 4. CreateJournalEntry success
	linesBalanced := []domain.UniversalJournalLine{
		{AccountID: accA.ID, AmountFunctional: decimal.NewFromInt(100), AmountTransactional: decimal.NewFromInt(100), CurrencyTransactional: "USD"},
		{AccountID: accB.ID, AmountFunctional: decimal.NewFromInt(-100), AmountTransactional: decimal.NewFromInt(-100), CurrencyTransactional: "USD"},
	}
	entry, err := svc.CreateJournalEntry(ctx, "legal_123", "FM", "JE-3", time.Now(), linesBalanced)
	if err != nil {
		t.Fatalf("unexpected error creating journal entry: %v", err)
	}
	if entry.Status != domain.LedgerStatePOSTED {
		t.Errorf("expected Posted status, got %v", entry.Status)
	}

	// Verify balances updated: Petty Cash is ASSET (Debit increase, so 100), Revenue is REVENUE (Credit increase, so 100)
	balA, _ := svc.GetAccountBalance(ctx, accA.ID)
	balB, _ := svc.GetAccountBalance(ctx, accB.ID)
	if !balA.Equal(decimal.NewFromInt(100)) || !balB.Equal(decimal.NewFromInt(100)) {
		t.Errorf("balances incorrect. Cash: %s, Revenue: %s", balA, balB)
	}

	// GetTrialBalance
	tb, err := svc.GetTrialBalance(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tb["total_debits"].(decimal.Decimal).Equal(decimal.NewFromInt(100)) || !tb["total_credits"].(decimal.Decimal).Equal(decimal.NewFromInt(100)) {
		t.Errorf("TB unbalanced: %+v", tb)
	}

	// GetBalanceSheet
	bs, err := svc.GetBalanceSheet(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bs["total_assets"].(decimal.Decimal).Equal(decimal.NewFromInt(100)) {
		t.Errorf("BS assets wrong: %+v", bs)
	}

	// GetIncomeStatement
	is, err := svc.GetIncomeStatement(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !is["total_revenue"].(decimal.Decimal).Equal(decimal.NewFromInt(100)) {
		t.Errorf("IS revenue wrong: %+v", is)
	}

	// GetCashFlow
	cf, err := svc.GetCashFlow(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cf["total_inflows"].(decimal.Decimal).Equal(decimal.NewFromInt(100)) {
		t.Errorf("CF inflow wrong: %+v", cf)
	}

	// 5. ReverseJournalEntry
	revEntry, err := svc.ReverseJournalEntry(ctx, entry.ID)
	if err != nil {
		t.Fatalf("unexpected error reversing: %v", err)
	}
	if revEntry.Status != domain.LedgerStatePOSTED {
		t.Errorf("expected reversing entry status Posted, got %v", revEntry.Status)
	}

	// Original entry should now be REVERSED
	orig, _, _ := svc.GetJournalEntry(ctx, entry.ID)
	if orig.Status != domain.LedgerStateREVERSED {
		t.Errorf("original entry status not reversed properly: %+v", orig)
	}

	// Balances should be back to 0
	balA, _ = svc.GetAccountBalance(ctx, accA.ID)
	balB, _ = svc.GetAccountBalance(ctx, accB.ID)
	if !balA.IsZero() || !balB.IsZero() {
		t.Errorf("balances not zero after reversal: Cash %s, Revenue %s", balA, balB)
	}

	// Try reversing again (should fail)
	_, err = svc.ReverseJournalEntry(ctx, entry.ID)
	if err == nil || err.Error() != "journal entry is already reversed" {
		t.Errorf("expected error for already reversed entry, got %v", err)
	}

	// ListJournalEntries
	entryList, err := svc.ListJournalEntries(ctx)
	if err != nil || len(entryList) != 2 { // original + reversing entry
		t.Errorf("unexpected journal entry list: %v, len: %d", err, len(entryList))
	}

	// DeleteJournalEntry
	err = svc.DeleteJournalEntry(ctx, entry.ID)
	if err != nil {
		t.Errorf("unexpected error deleting entry: %v", err)
	}

	// DeleteAccount
	err = svc.DeleteAccount(ctx, accA.ID)
	if err != nil {
		t.Errorf("unexpected error deleting account: %v", err)
	}
}

func TestGeneralLedgerService_UpdateJournalEntry_ValidationsAndErrors(t *testing.T) {
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(accounts, entries, outbox)

	svc := service.NewGeneralLedgerService(accounts, entries, outbox, tm)
	ctx := context.Background()

	// Create accounts
	accA, _ := svc.CreateAccount(ctx, "legal_123", "1000", "Cash", "ASSET")
	accB, _ := svc.CreateAccount(ctx, "legal_123", "4000", "Revenue", "REVENUE")

	// Update validations
	_, err := svc.UpdateJournalEntry(ctx, "je_non_existent", "", "", "", time.Now(), []domain.UniversalJournalLine{})
	if err == nil || err.Error() != "a journal entry must have at least 2 lines" {
		t.Errorf("expected error for too few lines, got %v", err)
	}

	linesUnbalanced := []domain.UniversalJournalLine{
		{AccountID: accA.ID, AmountFunctional: decimal.NewFromInt(10), AmountTransactional: decimal.NewFromInt(10), CurrencyTransactional: "USD"},
		{AccountID: accB.ID, AmountFunctional: decimal.NewFromInt(-20), AmountTransactional: decimal.NewFromInt(-20), CurrencyTransactional: "USD"},
	}
	_, err = svc.UpdateJournalEntry(ctx, "je_non_existent", "", "", "", time.Now(), linesUnbalanced)
	if err == nil {
		t.Errorf("expected error for unbalanced lines, got %v", err)
	}

	// Fetching non-existent entry with balanced lines
	linesBalanced := []domain.UniversalJournalLine{
		{AccountID: accA.ID, AmountFunctional: decimal.NewFromInt(100), AmountTransactional: decimal.NewFromInt(100), CurrencyTransactional: "USD"},
		{AccountID: accB.ID, AmountFunctional: decimal.NewFromInt(-100), AmountTransactional: decimal.NewFromInt(-100), CurrencyTransactional: "USD"},
	}
	_, err = svc.UpdateJournalEntry(ctx, "je_non_existent", "legal_123", "FM", "doc_p", time.Now(), linesBalanced)
	if err == nil {
		t.Error("expected error for updating non-existent journal entry, got nil")
	}
}

func TestLegalEntityService_All(t *testing.T) {
	repo := memory.NewMemoryLegalEntityRepo()
	tm := memory.NewMemoryTransactionManager(repo)
	svc := service.NewLegalEntityService(repo, tm)
	ctx := context.Background()

	// 1. Validation error: empty code
	_, err := svc.CreateLegalEntity(ctx, "", "Corp DE", "EUR", "DE123456789")
	if err == nil {
		t.Error("expected error for empty company code, got nil")
	}

	// 2. Validation error: invalid currency code
	_, err = svc.CreateLegalEntity(ctx, "CORP_DE", "Corp DE", "EURO", "DE123456789")
	if err == nil {
		t.Error("expected error for invalid currency length, got nil")
	}

	// 3. Success path
	le, err := svc.CreateLegalEntity(ctx, "CORP_DE", "Corp DE", "EUR", "DE123456789")
	if err != nil {
		t.Fatalf("failed to create legal entity: %v", err)
	}
	if le.CompanyCode != "CORP_DE" {
		t.Errorf("expected code CORP_DE, got %s", le.CompanyCode)
	}

	// 4. GetByID
	fetched, err := svc.GetLegalEntity(ctx, le.ID)
	if err != nil || fetched.CompanyName != "Corp DE" {
		t.Errorf("failed to fetch legal entity by ID: %v", err)
	}

	// 5. GetByCode
	fetchedCode, err := svc.GetLegalEntityByCode(ctx, "CORP_DE")
	if err != nil || fetchedCode.ID != le.ID {
		t.Errorf("failed to fetch legal entity by code: %v", err)
	}

	// 6. List
	list, err := svc.ListLegalEntities(ctx)
	if err != nil || len(list) != 1 {
		t.Errorf("unexpected list result: %v", err)
	}
}

func TestCapitalAssetService_All(t *testing.T) {
	assetRepo := memory.NewMemoryCapitalAssetRepo()
	lineRepo := memory.NewMemoryDepreciationScheduleLineRepo()
	accounts := memory.NewMemoryChartOfAccountsRepo()
	entries := memory.NewMemoryUniversalJournalEntryRepo()
	outbox := memory.NewMemoryTransactionalOutboxRepo()
	tm := memory.NewMemoryTransactionManager(assetRepo, lineRepo, accounts, entries, outbox)

	svc := service.NewCapitalAssetService(assetRepo, lineRepo, accounts, entries, outbox, tm)
	ctx := context.Background()

	// 1. Capitalization validations
	_, err := svc.CapitalizeAsset(ctx, "", "EQ-001", decimal.NewFromInt(1200), 12, nil)
	if err == nil {
		t.Error("expected error for empty legal entity ID")
	}
	_, err = svc.CapitalizeAsset(ctx, "legal_123", "EQ-001", decimal.Zero, 12, nil)
	if err == nil {
		t.Error("expected error for zero acquisition cost")
	}

	// 2. Capitalize success
	asset, err := svc.CapitalizeAsset(ctx, "legal_123", "EQ-001", decimal.NewFromInt(1200), 12, nil)
	if err != nil {
		t.Fatalf("unexpected error capitalizing asset: %v", err)
	}
	if asset.Status != domain.AssetStateACTIVE {
		t.Errorf("expected asset to be active, got %s", asset.Status)
	}

	// 3. Generate Schedule
	lines, err := svc.GenerateDepreciationSchedule(ctx, asset.ID)
	if err != nil || len(lines) != 12 {
		t.Fatalf("failed to generate depreciation schedule: %v (len: %d)", err, len(lines))
	}
	expectedDepAmt := decimal.NewFromInt(100)
	if !lines[0].DepreciationAmount.Equal(expectedDepAmt) {
		t.Errorf("expected monthly depreciation of 100, got %s", lines[0].DepreciationAmount)
	}

	// 4. Post Monthly Depreciation
	err = svc.PostMonthlyStraightLineDepreciation(ctx, "legal_123", lines[0].FiscalYear, lines[0].PeriodNumber)
	if err != nil {
		t.Fatalf("failed to post monthly depreciation: %v", err)
	}

	// Verify asset accumulated depreciation is updated
	updatedAsset, _ := svc.GetAsset(ctx, asset.ID)
	if !updatedAsset.AccumulatedDepreciation.Equal(expectedDepAmt) {
		t.Errorf("expected accumulated depreciation of 100, got %s", updatedAsset.AccumulatedDepreciation)
	}

	// Verify the schedule line is marked as posted
	lines, _ = lineRepo.GetByAssetID(ctx, asset.ID)
	var postedCount int
	for _, l := range lines {
		if l.IsPosted {
			postedCount++
		}
	}
	if postedCount != 1 {
		t.Errorf("expected exactly 1 posted line, got %d", postedCount)
	}
}

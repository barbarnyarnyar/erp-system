package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type CapitalAssetService struct {
	assetRepo domain.CapitalAssetRepository
	lineRepo  domain.DepreciationScheduleLineRepository
	accounts  domain.ChartOfAccountsRepository
	entries   domain.UniversalJournalEntryRepository
	outbox    domain.TransactionalOutboxRepository
	tm        domain.TransactionManager
}

func NewCapitalAssetService(
	assetRepo domain.CapitalAssetRepository,
	lineRepo domain.DepreciationScheduleLineRepository,
	accounts domain.ChartOfAccountsRepository,
	entries domain.UniversalJournalEntryRepository,
	outbox domain.TransactionalOutboxRepository,
	tm domain.TransactionManager,
) *CapitalAssetService {
	return &CapitalAssetService{
		assetRepo: assetRepo,
		lineRepo:  lineRepo,
		accounts:  accounts,
		entries:   entries,
		outbox:    outbox,
		tm:        tm,
	}
}

func (s *CapitalAssetService) getOrCreateAccount(ctx context.Context, legalEntityID, code, name, accType string) (*domain.ChartOfAccounts, error) {
	acc, err := s.accounts.GetByCode(ctx, legalEntityID, code)
	if err == nil {
		return acc, nil
	}

	typeEnum := domain.AccountType(accType)
	newAcc := &domain.ChartOfAccounts{
		ID:            utils.NewID("acc"),
		LegalEntityID: legalEntityID,
		AccountCode:   code,
		AccountName:   name,
		Type:          typeEnum,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.accounts.Create(ctx, newAcc); err != nil {
		return nil, err
	}
	return newAcc, nil
}

func (s *CapitalAssetService) CapitalizeAsset(
	ctx context.Context,
	legalEntityID string,
	assetTag string,
	acquisitionCost decimal.Decimal,
	usefulLifeMonths int,
	equipmentID *string,
) (*domain.CapitalAsset, error) {
	if legalEntityID == "" || assetTag == "" || usefulLifeMonths <= 0 {
		return nil, errors.New("legal entity ID, asset tag, and a positive useful life are required")
	}
	if acquisitionCost.IsZero() || acquisitionCost.IsNegative() {
		return nil, errors.New("acquisition cost must be positive")
	}

	asset := &domain.CapitalAsset{
		ID:                      utils.NewID("asset"),
		LegalEntityID:           legalEntityID,
		AssetTag:                assetTag,
		EamEquipmentID:          equipmentID,
		AcquisitionCost:         acquisitionCost,
		AccumulatedDepreciation: decimal.Zero,
		UsefulLifeMonths:        usefulLifeMonths,
		CapitalizationDate:      time.Now(),
		Status:                  domain.AssetStateACTIVE,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Save asset
		if err := s.assetRepo.Create(txCtx, asset); err != nil {
			return err
		}

		// Create General Ledger entries for the capitalization
		// Debit: Fixed Asset Account (1500-001)
		assetAcc, err := s.getOrCreateAccount(txCtx, legalEntityID, "1500-001", "Fixed Assets - Equipment", "ASSET")
		if err != nil {
			return err
		}
		// Credit: Accounts Payable Clearing Account (2110-999)
		offsetAcc, err := s.getOrCreateAccount(txCtx, legalEntityID, "2110-999", "AP Clearing - Fixed Assets", "LIABILITY")
		if err != nil {
			return err
		}

		lines := []domain.UniversalJournalLine{
			{
				AccountID:             assetAcc.ID,
				AmountFunctional:      acquisitionCost,
				AmountTransactional:   acquisitionCost,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             offsetAcc.ID,
				AmountFunctional:      acquisitionCost.Neg(),
				AmountTransactional:   acquisitionCost.Neg(),
				CurrencyTransactional: "USD",
			},
		}

		entryID := utils.NewID("je")
		entry := &domain.UniversalJournalEntry{
			ID:               entryID,
			LegalEntityID:    legalEntityID,
			SourceModule:     "FM",
			SourceDocumentID: asset.ID,
			PostingDate:      time.Now(),
			FinancialPeriod:  time.Now().Format("2006-01"),
			Status:           domain.LedgerStatePOSTED,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		for i := range lines {
			lines[i].ID = utils.NewID("jel")
			lines[i].JournalEntryID = entryID
		}

		if err := s.entries.Create(txCtx, entry, lines); err != nil {
			return err
		}

		// Write event to transactional outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   "fm.asset.capitalized",
			AggregateID: asset.ID,
			Payload: map[string]interface{}{
				"asset_id":         asset.ID,
				"asset_tag":        asset.AssetTag,
				"acquisition_cost": asset.AcquisitionCost,
				"timestamp":        time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *CapitalAssetService) GenerateDepreciationSchedule(ctx context.Context, assetID string) ([]domain.DepreciationScheduleLine, error) {
	asset, err := s.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, err
	}

	lines, err := s.lineRepo.GetByAssetID(ctx, assetID)
	if err == nil && len(lines) > 0 {
		return lines, nil // Already generated
	}

	months := asset.UsefulLifeMonths
	depAmount := asset.AcquisitionCost.Div(decimal.NewFromInt(int64(months)))

	scheduleLines := make([]domain.DepreciationScheduleLine, months)
	startDate := asset.CapitalizationDate
	for i := 0; i < months; i++ {
		targetDate := startDate.AddDate(0, i, 0)
		scheduleLines[i] = domain.DepreciationScheduleLine{
			ID:                 utils.NewID("dsl"),
			FixedAssetID:       assetID,
			FiscalYear:         targetDate.Year(),
			PeriodNumber:       int(targetDate.Month()),
			DepreciationAmount: depAmount,
			IsPosted:           false,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
	}

	err = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.lineRepo.CreateMany(txCtx, scheduleLines)
	})
	if err != nil {
		return nil, err
	}

	return scheduleLines, nil
}

func (s *CapitalAssetService) PostMonthlyStraightLineDepreciation(ctx context.Context, legalEntityID string, fiscalYear, periodNumber int) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		lines, err := s.lineRepo.GetUnpostedByPeriod(txCtx, fiscalYear, periodNumber)
		if err != nil {
			return err
		}

		if len(lines) == 0 {
			return nil
		}

		var totalDepreciation decimal.Decimal
		var processedLines []domain.DepreciationScheduleLine

		for _, l := range lines {
			asset, err := s.assetRepo.GetByID(txCtx, l.FixedAssetID)
			if err != nil {
				continue
			}

			if asset.LegalEntityID != legalEntityID {
				continue
			}

			totalDepreciation = totalDepreciation.Add(l.DepreciationAmount)

			// Update schedule line
			l.IsPosted = true
			l.UpdatedAt = time.Now()
			if err := s.lineRepo.Update(txCtx, &l); err != nil {
				return err
			}

			// Update asset accumulated depreciation
			asset.AccumulatedDepreciation = asset.AccumulatedDepreciation.Add(l.DepreciationAmount)
			if asset.AccumulatedDepreciation.GreaterThanOrEqual(asset.AcquisitionCost) {
				asset.Status = domain.AssetStateDISPOSED // Mark fully depreciated/inactive
			}
			asset.UpdatedAt = time.Now()
			if err := s.assetRepo.Update(txCtx, asset); err != nil {
				return err
			}

			processedLines = append(processedLines, l)
		}

		if totalDepreciation.IsZero() {
			return nil
		}

		// Post consolidated depreciation journal entry
		// Debit: Depreciation Expense (6040-001)
		depExpenseAcc, err := s.getOrCreateAccount(txCtx, legalEntityID, "6040-001", "Depreciation Expense", "EXPENSE")
		if err != nil {
			return err
		}
		// Credit: Accumulated Depreciation (1590-001)
		accumDepAcc, err := s.getOrCreateAccount(txCtx, legalEntityID, "1590-001", "Accumulated Depreciation", "ASSET")
		if err != nil {
			return err
		}

		jeLines := []domain.UniversalJournalLine{
			{
				AccountID:             depExpenseAcc.ID,
				AmountFunctional:      totalDepreciation,
				AmountTransactional:   totalDepreciation,
				CurrencyTransactional: "USD",
			},
			{
				AccountID:             accumDepAcc.ID,
				AmountFunctional:      totalDepreciation.Neg(),
				AmountTransactional:   totalDepreciation.Neg(),
				CurrencyTransactional: "USD",
			},
		}

		entryID := utils.NewID("je")
		entry := &domain.UniversalJournalEntry{
			ID:               entryID,
			LegalEntityID:    legalEntityID,
			SourceModule:     "FM",
			SourceDocumentID: entryID, // Self reference
			PostingDate:      time.Now(),
			FinancialPeriod:  fmt.Sprintf("%04d-%02d", fiscalYear, periodNumber),
			Status:           domain.LedgerStatePOSTED,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		for i := range jeLines {
			jeLines[i].ID = utils.NewID("jel")
			jeLines[i].JournalEntryID = entryID
		}

		if err := s.entries.Create(txCtx, entry, jeLines); err != nil {
			return err
		}

		return nil
	})
}

func (s *CapitalAssetService) GetAsset(ctx context.Context, id string) (*domain.CapitalAsset, error) {
	return s.assetRepo.GetByID(ctx, id)
}

func (s *CapitalAssetService) ListAssets(ctx context.Context) ([]domain.CapitalAsset, error) {
	return s.assetRepo.List(ctx)
}

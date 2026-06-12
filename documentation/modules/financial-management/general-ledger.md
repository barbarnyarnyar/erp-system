# General Ledger

Central repository for financial accounts and journal entries with double-entry balance validation and real-time balance tracking.

## Chart of Accounts Structure

### Account Model (ChartOfAccounts)
```go
type ChartOfAccounts struct {
	ID            string      `json:"id"`
	LegalEntityID string      `json:"legal_entity_id"`
	AccountCode   string      `json:"account_code"` // Unique combination with legal_entity_id
	AccountName   string      `json:"account_name"`
	Type          AccountType `json:"type"` // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	IsActive      bool        `json:"is_active"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}
```

### Account Types and Balance Behavior

| Type | Normal Side | Debit Effect | Credit Effect |
|------|------------|--------------|---------------|
| ASSET | Debit | Increase | Decrease |
| LIABILITY | Credit | Decrease | Increase |
| EQUITY | Credit | Decrease | Increase |
| REVENUE | Credit | Decrease | Increase |
| EXPENSE | Debit | Increase | Decrease |

### Constraints
Chart of accounts enforces multi-tenant segmentation. There is a composite unique constraint on `(legal_entity_id, account_code)`. Duplicate account codes under the same legal entity are rejected at the database level.

---

## Journal Entry System

### Journal Entry Models

```go
type UniversalJournalEntry struct {
	ID               string      `json:"id"`
	LegalEntityID    string      `json:"legal_entity_id"`    // Strict multi-tenant partition shield
	SourceModule     string      `json:"source_module"`      // "SCM", "CRM", "HR", "PRJ", "FM"
	SourceDocumentID string      `json:"source_document_id"` // Primitive reference to external systems
	PostingDate      time.Time   `json:"posting_date"`
	FinancialPeriod  string      `json:"financial_period"` // Format: "YYYY-MM"
	Status           LedgerState `json:"status"` // DRAFT, POSTED, REVERSED
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type UniversalJournalLine struct {
	ID                    string          `json:"id"`
	JournalEntryID        string          `json:"journal_entry_id"`
	AccountID             string          `json:"account_id"`
	AmountFunctional      decimal.Decimal `json:"amount_functional"`      // Functional currency value
	AmountTransactional   decimal.Decimal `json:"amount_transactional"`   // Original currency value before conversion
	CurrencyTransactional string          `json:"currency_transactional"` // ISO 4217 code of origin transaction
	TrackingDimensions    interface{}     `json:"tracking_dimensions"`
}
```

### Entry Creation Rules
1. Minimum 2 lines per entry.
2. The sum of `amount_functional` across all lines must equal zero (validated with `decimal.Decimal`).
3. All account IDs must exist.
4. Status defaults to `"POSTED"`.
5. Account balances are calculated dynamically as the sum of all posted journal lines referencing that account.
6. An outbox event `fm.account.balance.changed` is published per line.

### Entry Reversal
1. Validates that the entry is not already `REVERSED`.
2. Creates a new entry with swapped functional and transactional amount signs.
3. Original entry status is set to `REVERSED`.

---

## Account Management

### Create Account
```http
POST /api/v1/accounts
Content-Type: application/json

{
  "legal_entity_id": "le_1234567890",
  "account_code": "1100",
  "account_name": "Cash - Operating",
  "type": "ASSET"
}
```

Response: 
```json
{
  "data": {
    "id": "acc_1234567890",
    "legal_entity_id": "le_1234567890",
    "account_code": "1100",
    "account_name": "Cash - Operating",
    "type": "ASSET",
    "is_active": true,
    "created_at": "2026-06-13T02:00:00Z",
    "updated_at": "2026-06-13T02:00:00Z"
  }
}
```

### Validation Rules
- `legal_entity_id` is required and must match an existing Legal Entity.
- `account_code` is required and must be unique under the given legal entity.
- `account_name` is required.
- `type` is required and must be one of: `ASSET`, `LIABILITY`, `EQUITY`, `REVENUE`, `EXPENSE`.

---

## Financial Reports

All reports gather live aggregations from the ledger lines.

### Balance Sheet
Groups accounts by type (ASSET/LIABILITY/EQUITY). REVENUE and EXPENSE accounts are excluded.

### Income Statement
Groups accounts by type (REVENUE/EXPENSE) and calculates `net_income = total_revenue - total_expense`.

### Cash Flow
Aggregates inflows and outflows from ASSET accounts having "cash" or "bank" (case-insensitive) in their name.

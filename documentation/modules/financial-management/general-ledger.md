# General Ledger

Central repository for financial accounts and journal entries with double-entry balance validation and real-time balance tracking.

## Implementation Status

The GL is fully functional for single-currency accounting. All data is in-memory — no PostgreSQL connection.

## Chart of Accounts Structure

### Account Model
```go
type Account struct {
    ID            string          `json:"id"`
    AccountNumber string          `json:"account_number"`
    Name          string          `json:"name"`
    Type          string          `json:"type"`     // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
    ParentID      *string         `json:"parent_id"`
    Balance       decimal.Decimal `json:"balance"`
    Currency      string          `json:"currency"` // stored per account, no conversion logic
    IsActive      bool            `json:"is_active"`
    CreatedAt     time.Time       `json:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at"`
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

### Hierarchy
Accounts support a parent-child structure via `ParentID`. The code does NOT enforce hierarchy rules (e.g., no check preventing posting to parent accounts).

### Standard Account Number Ranges (Convention)

| Range | Type | Example |
|-------|------|---------|
| 1000-1999 | ASSET | 1100 - Cash, 1200 - Accounts Receivable, 1400 - Fixed Assets |
| 2000-2999 | LIABILITY | 2100 - Accounts Payable, 2200 - Accrued Expenses |
| 3000-3999 | EQUITY | 3100 - Share Capital, 3200 - Retained Earnings |
| 4000-4999 | REVENUE | 4100 - Product Sales, 4200 - Service Revenue |
| 5000-9999 | EXPENSE | 5100 - COGS, 6100 - Operating Expenses |

These are naming conventions only — the code does not enforce any range-to-type mapping.

## Journal Entry System

### Journal Entry Model
```go
type JournalEntry struct {
    ID          string    `json:"id"`
    Reference   string    `json:"reference"`
    Date        time.Time `json:"date"`
    Description string    `json:"description"`
    Status      string    `json:"status"`     // POSTED (upon creation) or REVERSED
    CreatedBy   string    `json:"created_by"`
    ReversedBy  *string   `json:"reversed_by"` // ID of reversing entry
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type JournalEntryLine struct {
    ID           string          `json:"id"`
    EntryID      string          `json:"entry_id"`
    AccountID    string          `json:"account_id"`
    DebitAmount  decimal.Decimal `json:"debit_amount"`
    CreditAmount decimal.Decimal `json:"credit_amount"`
    Description  string          `json:"description"`
}
```

### Entry Creation Rules
1. Minimum 2 lines per entry
2. Total debits must equal total credits (validated with `decimal.Decimal.Equal`)
3. All account IDs must exist in the repository
4. Status is always set to `"POSTED"` — no draft state
5. Account balances are updated atomically during creation:
   - For debit-increase accounts (ASSET/EXPENSE): `Balance += debits - credits`
   - For credit-increase accounts (LIABILITY/EQUITY/REVENUE): `Balance -= debits + credits`
6. `fin.account.balance.changed` event published per line

### Entry Reversal
1. Validates entry is not already REVERSED
2. Creates new entry with swapped debit/credit amounts
3. Reference prefixed with `"REV-"`
4. Original entry status set to REVERSED
5. `ReversedBy` field links to reversing entry ID

## Account Management

### Create Account
```http
POST /api/v1/accounts
Content-Type: application/json

{
  "account_number": "1100",
  "name": "Cash - Operating",
  "type": "ASSET",
  "parent_id": "",
  "currency": "USD"
}
```

Response: `{"data": {"id": "acc_1234567890", "account_number": "1100", ...}}`

### Validation Rules
- `account_number` is required (string)
- `name` is required
- `type` is required (free-form string — no enum validation)
- Duplicate account number is NOT validated (no uniqueness check in service)
- `currency` is stored but never used in conversions

## Trial Balance

```http
GET /api/v1/reports/trial-balance
```

Returns:
```json
{
  "balances": {
    "Cash - Operating": {"debit": "50000.00"},
    "Accounts Payable": {"credit": "15000.00"}
  },
  "total_debits": "50000.00",
  "total_credits": "15000.00"
}
```

Logic: Account types ASSET and EXPENSE → debit side. All others → credit side. Inactive accounts excluded.

## Balance Sheet

```http
GET /api/v1/reports/balance-sheet
```

Returns:
```json
{
  "assets": {"Cash - Operating": "50000.00"},
  "total_assets": "50000.00",
  "liabilities": {"Accounts Payable": "15000.00"},
  "total_liabilities": "15000.00",
  "equity": {"Retained Earnings": "35000.00"},
  "total_equity": "35000.00"
}
```

Logic: Groups accounts by type (ASSET/LIABILITY/EQUITY). REVENUE and EXPENSE accounts are excluded.

## Known Limitations

| Gap | Detail |
|-----|--------|
| No account code validation | Type (ASSET/etc.) is a free string — no enum enforcement |
| No account code uniqueness | Duplicate account numbers not rejected |
| No posting rules | No check for leaf-vs-parent account posting |
| No draft/approval workflow | All entries POSTED immediately |
| Currency is decorative | Stored per account but no conversion, no rate table usage |
| Income statement is stub | Returns hardcoded message — no revenue/expense aggregation |
| Cash flow is stub | Returns hardcoded message — no logic |
| No period closing | No fiscal year close, no retained earnings transfer |
| In-memory only | All data lost on service restart |
| No pagination | List endpoints return all records |

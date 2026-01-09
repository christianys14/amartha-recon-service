package transaction

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

const (
	transactionTypeDebit  TransactionType = "DEBIT"
	transactionTypeCredit TransactionType = "CREDIT"
)

type (
	TransactionType string

	Transaction struct {
		ID              uint64          `db:"id"`
		TransactionID   string          `db:"transaction_id"`
		TerminalRRN     string          `db:"terminal_rrn"`
		Amount          decimal.Decimal `db:"amount"`
		TransactionType TransactionType `db:"transaction_type"`
		BankCode        string          `db:"bank_code"`
		TransactionTime time.Time       `db:"transaction_time"`
		UpdatedAt       time.Time       `db:"updated_at"`
	}

	Criteria struct {
		StartDate time.Time
		EndDate   time.Time
	}

	Repository interface {
		FindDistinctBankCode(ctx context.Context) ([]*Transaction, error)
		FindTransaction(ctx context.Context, tc *Criteria) ([]*Transaction, error)
	}
)

func (t *Transaction) IsDebit() bool {
	return t.TransactionType == transactionTypeDebit
}

func (t *Transaction) IsCredit() bool {
	return t.TransactionType == transactionTypeCredit
}

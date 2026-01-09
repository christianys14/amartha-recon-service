package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

type (
	LoanEntity struct {
		ID        uint64          `db:"id" json:"id,omitempty"`
		Status    string          `db:"status" json:"status,omitempty"`
		UserID    string          `db:"user_id" json:"user_id,omitempty"`
		DueDate   time.Time       `db:"due_date" json:"due_date,omitempty"`
		Amount    decimal.Decimal `db:"amount" json:"amount,omitempty"`
		CreatedAt time.Time       `db:"created_at" json:"created_at,omitempty"`
		Version   int             `db:"version" json:"version,omitempty"`
		UpdatedAt time.Time       `db:"updated_at" json:"updated_at,omitempty"`
		Statuses  []string        `json:"statuses,omitempty"`
	}

	LoanEntityUpdate struct {
		IDs    []uint64 `db:"id" json:"id,omitempty"`
		Status string   `db:"status" json:"status,omitempty"`
	}

	LoanRepository interface {
		SaveLoans(ctx context.Context, tx *sql.Tx, loanEntity ...*LoanEntity) error

		FindLoans(ctx context.Context, loanEntity *LoanEntity) ([]*LoanEntity, error)

		UpdateLoan(ctx context.Context, loanEntity *LoanEntityUpdate) error
	}
)

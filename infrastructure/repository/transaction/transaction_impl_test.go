package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNewTransactionRepository(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := NewTransactionRepository(sqlxDB)
	assert.NotNil(t, repo)
}

func TestTransactionRepository_FindDistinctBankCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewTransactionRepository(sqlxDB)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"bank_code"}).
			AddRow("BANK_A").
			AddRow("BANK_B")

		mock.ExpectQuery(queryDistinctBank).WillReturnRows(rows)

		result, err := repo.FindDistinctBankCode(ctx)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "BANK_A", result[0].BankCode)
		assert.Equal(t, "BANK_B", result[1].BankCode)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery(queryDistinctBank).WillReturnError(errors.New("db error"))

		result, err := repo.FindDistinctBankCode(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTransactionRepository_FindTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewTransactionRepository(sqlxDB)

	ctx := context.Background()
	tc := &Criteria{
		StartDate: time.Now().AddDate(0, 0, -1),
		EndDate:   time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "transaction_id", "terminal_rrn", "amount", "transaction_type", "bank_code", "transaction_time", "updated_at"}).
			AddRow(1, "TX001", "RRN001", 1000.0, "DEBIT", "BANK_A", time.Now(), time.Now())

		mock.ExpectQuery("WHERE date\\(transaction_time\\) >= \\? AND date\\(transaction_time\\) <= \\?").
			WithArgs(tc.StartDate, tc.EndDate).
			WillReturnRows(rows)

		result, err := repo.FindTransaction(ctx, tc)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "TX001", result[0].TransactionID)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery("WHERE date\\(transaction_time\\) >= \\? AND date\\(transaction_time\\) <= \\?").
			WithArgs(tc.StartDate, tc.EndDate).
			WillReturnError(errors.New("db error"))

		result, err := repo.FindTransaction(ctx, tc)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestTransaction_HelperMethods(t *testing.T) {
	t.Run("IsDebit", func(t *testing.T) {
		tr := &Transaction{TransactionType: transactionTypeDebit}
		assert.True(t, tr.IsDebit())
		assert.False(t, tr.IsCredit())
	})

	t.Run("IsCredit", func(t *testing.T) {
		tr := &Transaction{TransactionType: transactionTypeCredit}
		assert.True(t, tr.IsCredit())
		assert.False(t, tr.IsDebit())
	})

	t.Run("OtherType", func(t *testing.T) {
		tr := &Transaction{TransactionType: "OTHER"}
		assert.False(t, tr.IsDebit())
		assert.False(t, tr.IsCredit())
	})
}

package transaction

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
)

const (
	queryDistinctBank    = "select distinct bank_code from transactions"
	queryFindTransaction = "select id, transaction_id, terminal_rrn, amount, transaction_type, bank_code, transaction_time, updated_at FROM transactions "
)

type transactionRepository struct {
	masterConnection *sqlx.DB
}

func NewTransactionRepository(connectionDB *sqlx.DB) Repository {
	return &transactionRepository{masterConnection: connectionDB}
}

func (t *transactionRepository) FindDistinctBankCode(ctx context.Context) ([]*Transaction, error) {
	var transactions []*Transaction
	if err := t.masterConnection.SelectContext(ctx, &transactions, queryDistinctBank); err != nil {
		log.Println("error when selecting distinct bank codes -> ", err)
		return nil, err
	}

	return transactions, nil
}

func (t *transactionRepository) FindTransaction(ctx context.Context, tc *Criteria) ([]*Transaction, error) {
	queryParams := []interface{}{
		tc.StartDate,
		tc.EndDate,
	}
	queryFull := queryFindTransaction + "WHERE date(transaction_time) >= ? AND date(transaction_time) <= ?"

	var transactions []*Transaction
	if err := t.masterConnection.SelectContext(ctx, &transactions, queryFull, queryParams...); err != nil {
		log.Println("error when selecting find transaction -> ", err)
		return nil, err
	}

	return transactions, nil
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	queryInsert = `
		INSERT INTO loan (status, user_id, due_date, amount, created_at, version, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	querySelect = `
		SELECT id, status, user_id, due_date, amount, created_at, version, updated_at 
		FROM loan WHERE TRUE
	`

	queryUpdate = `
		UPDATE loan SET
	`

	queryUpdateWhere = `
		version = version + 1,
		updated_at = now()
	WHERE 
		id IN
	`
)

var (
	ErrorFromDBLoan = errors.New("error from database")
	ErrorNoRows     = errors.New("no rows loan")
)

type loanRepository struct {
	connectionDB *sql.DB
}

func NewLoanRepository(connectionDB *sql.DB) LoanRepository {
	return &loanRepository{
		connectionDB: connectionDB,
	}
}

func (l *loanRepository) SaveLoans(
	ctx context.Context,
	db *sql.Tx,
	loanEntity ...*LoanEntity) error {
	var slice = make([]map[string]interface{}, len(loanEntity))

	for idx, value := range loanEntity {
		slice[idx] = make(map[string]interface{})

		slice[idx]["status"] = value.Status
		slice[idx]["user_id"] = value.UserID
		slice[idx]["due_date"] = value.DueDate
		slice[idx]["amount"] = value.Amount
		slice[idx]["created_at"] = value.CreatedAt
		slice[idx]["version"] = value.Version
		slice[idx]["updated_at"] = value.UpdatedAt
	}

	statement, err := db.PrepareContext(ctx, queryInsert)
	if err != nil {
		log.Println("unidentified error from database when prepare -> ", err)
		return ErrorFromDBLoan
	}
	defer statement.Close()

	for _, entry := range slice {
		_, errExecContext := statement.ExecContext(
			ctx,
			entry["status"],
			entry["user_id"],
			entry["due_date"],
			entry["amount"],
			entry["created_at"],
			entry["version"],
			entry["updated_at"],
		)

		if errExecContext != nil {
			log.Println("unidentified error from database when exec -> ", errExecContext)
			return errExecContext
		}
	}

	return nil
}

func (l *loanRepository) FindLoans(
	ctx context.Context,
	loanEntity *LoanEntity) ([]*LoanEntity, error) {
	queryWhere, parameters := builderWhere(loanEntity)
	queryFull := querySelect + queryWhere + " AND status IN" + "(" + buildWhereIn(len(loanEntity.Statuses)) + ")"

	for _, sts := range loanEntity.Statuses {
		parameters = append(parameters, sts)
	}

	var amount sql.NullFloat64

	res, err := l.connectionDB.QueryContext(ctx, queryFull, parameters...)

	if err != nil {
		log.Println("unidentified error from database when query context -> ", err)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorNoRows
		}

		return nil, ErrorFromDBLoan
	}

	var data []*LoanEntity
	for res.Next() {
		var r LoanEntity
		var dueDate, createdAt, updatedAt string

		errScan := res.Scan(
			&r.ID, &r.Status,
			&r.UserID, &dueDate,
			&amount, &createdAt,
			&r.Version, &updatedAt,
		)

		if errScan != nil {
			log.Println("unidentified error from database when scan -> ", errScan)
			return nil, ErrorFromDBLoan
		}

		amt := func() decimal.Decimal {
			if amount.Valid {
				return decimal.NewFromFloat(amount.Float64)
			}

			return decimal.NewFromFloat(float64(0))
		}()

		parsedDueDate, _ := time.Parse("2006-01-02", dueDate)
		parsedCreatedAt, _ := time.Parse("2006-01-02", createdAt)
		parsedUpdatedAt, _ := time.Parse("2006-01-02", updatedAt)

		r.Amount = amt
		r.Statuses = nil
		r.DueDate = parsedDueDate
		r.CreatedAt = parsedCreatedAt
		r.UpdatedAt = parsedUpdatedAt

		data = append(data, &r)
	}

	return data, nil
}

func (l *loanRepository) UpdateLoan(
	ctx context.Context,
	loanEntityUpdate *LoanEntityUpdate) error {
	querySet, parameters := builderUpdate(loanEntityUpdate)
	for _, id := range loanEntityUpdate.IDs {
		parameters = append(parameters, id)
	}

	queryFull := queryUpdate + querySet + queryUpdateWhere + "(" + buildWhereIn(len(loanEntityUpdate.IDs)) + ")"
	result, err := l.connectionDB.ExecContext(ctx, queryFull, parameters...)

	if err != nil {
		log.Println("unidentified error from database when exec -> ", err)
		return ErrorFromDBLoan
	}

	_, err = result.RowsAffected()

	if err != nil {
		log.Println("unidentified error from database when rowsAffected -> ", err)
		return ErrorFromDBLoan
	}

	return nil
}

func builderWhere(loanEntity *LoanEntity) (string, []interface{}) {
	var sb strings.Builder
	var parameters []interface{}

	if loanEntity.ID != 0 {
		sb.WriteString("AND id = ? ")
		parameters = append(parameters, loanEntity.ID)
	}

	if loanEntity.UserID != "" {
		sb.WriteString("AND user_id = ? ")
		parameters = append(parameters, loanEntity.UserID)
	}

	if !loanEntity.DueDate.IsZero() {
		sb.WriteString("AND due_date < ? ")
		parameters = append(parameters, loanEntity.DueDate)
	}

	return sb.String(), parameters
}

func builderUpdate(loanEntity *LoanEntityUpdate) (string, []interface{}) {
	var sb strings.Builder
	var parameters []interface{}

	if loanEntity.Status != "" {
		sb.WriteString("status = ?, ")
		parameters = append(parameters, loanEntity.Status)
	}

	return sb.String(), parameters
}

func buildWhereIn(n int) string {
	return strings.Trim(strings.Repeat("?,", n), ",")
}

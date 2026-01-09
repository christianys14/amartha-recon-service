package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_loanRepository_SaveLoans(t *testing.T) {
	dateRandom := time.Date(2024, 6, 23, 21, 0, 0, 0, time.UTC)
	le := []*LoanEntity{
		{
			ID:        uint64(1),
			Status:    "PENDING",
			UserID:    "CUSTOMER01",
			DueDate:   dateRandom,
			Amount:    decimal.NewFromFloat(float64(120000)),
			CreatedAt: dateRandom,
			Version:   0,
			UpdatedAt: dateRandom,
		},
	}

	ctx := context.Background()

	type args struct {
		ctx        context.Context
		loanEntity []*LoanEntity
	}
	tests := []struct {
		name      string
		args      args
		sqlErr    error
		sqlResult driver.Result
		want      bool
	}{
		{
			name: "given the happy case," +
				"when saveLoans," +
				"then return nil",
			args: args{
				ctx:        ctx,
				loanEntity: le,
			},
			sqlResult: sqlmock.NewResult(10, 1),
		},
		{
			name: "given the negative case because exec context," +
				"when saveLoans," +
				"then return error",
			args: args{
				ctx:        ctx,
				loanEntity: le,
			},
			sqlErr: sql.ErrTxDone,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				db, mock, err := sqlmock.New()

				if err != nil {
					t.Errorf("DB Connection Error LoanRepositoryImpl.SaveLoans() error = %v", err)
				}
				defer db.Close()

				mock.ExpectBegin().WillReturnError(nil)

				defer func() {
					if err := mock.ExpectationsWereMet(); err != nil {
						assert.Fail(t, "there were unfulfilled expectations", err.Error())
					}
				}()

				if tt.sqlErr != nil {
					mock.
						ExpectPrepare(regexp.QuoteMeta(queryInsert)).
						ExpectExec().
						WithArgs(
							"PENDING",
							"CUSTOMER01",
							dateRandom,
							decimal.NewFromFloat(float64(120000)),
							dateRandom,
							0,
							dateRandom,
						).
						WillReturnError(tt.sqlErr)
				}

				if tt.sqlResult != nil {
					mock.
						ExpectPrepare(regexp.QuoteMeta(queryInsert)).
						ExpectExec().
						WithArgs(
							"PENDING",
							"CUSTOMER01",
							dateRandom,
							decimal.NewFromFloat(float64(120000)),
							dateRandom,
							0,
							dateRandom,
						).
						WillReturnResult(tt.sqlResult)
				}

				l := NewLoanRepository(db)
				tx, _ := db.Begin()

				err = l.SaveLoans(tt.args.ctx, tx, tt.args.loanEntity...)
				if (err != nil) != tt.want {
					t.Errorf(
						"LoanRepositoryImpl.SaveLoans() error = %v, wantErr %v",
						err, tt.want)
					return
				}
			})
	}
}

func Test_loanRepository_FindLoan(t *testing.T) {
	dateRandom := time.Date(2024, 6, 23, 21, 0, 0, 0, time.UTC)

	var data []*LoanEntity
	le := LoanEntity{
		ID:        uint64(10),
		Status:    "PENDING",
		UserID:    "customer01",
		DueDate:   dateRandom,
		Amount:    decimal.NewFromFloat(float64(120000)),
		CreatedAt: dateRandom,
		Version:   1,
		UpdatedAt: dateRandom,
		Statuses:  nil,
	}
	data = append(data, &le)

	type args struct {
		loanEntity *LoanEntity
	}
	tests := []struct {
		name    string
		args    args
		sqlErr  error
		sqlRows *sqlmock.Rows
		want    []*LoanEntity
		wantErr bool
	}{
		{
			name: "given happy case," +
				"when findLoan," +
				"then return the result from db",
			args: args{
				loanEntity: &LoanEntity{
					ID:       uint64(122),
					Status:   "PENDING11",
					UserID:   "customer012",
					DueDate:  dateRandom,
					Amount:   decimal.NewFromFloat(float64(110000)),
					Statuses: []string{"PENDING"},
				},
			},
			sqlRows: sqlmock.NewRows(
				[]string{
					"id",
					"status",
					"user_id",
					"due_date",
					"amount",
					"created_at",
					"version",
					"updated_at",
				}).
				AddRow(
					le.ID,
					le.Status,
					le.UserID,
					le.DueDate,
					le.Amount,
					le.CreatedAt,
					le.Version,
					le.UpdatedAt,
				),
			want: data,
		},
		{
			name: "given negative case sql no rows," +
				"when findLoan," +
				"then return error",
			args: args{
				loanEntity: &LoanEntity{
					ID:       uint64(122),
					Status:   "PENDING11",
					UserID:   "customer012",
					DueDate:  dateRandom,
					Amount:   decimal.NewFromFloat(float64(110000)),
					Statuses: []string{"PENDING", "CLOSED"},
				},
			},
			sqlErr:  sql.ErrNoRows,
			want:    nil,
			wantErr: true,
		},
		{
			name: "given negative case sql tx done," +
				"when findLoan," +
				"then return error",
			args: args{
				loanEntity: &LoanEntity{
					ID:       uint64(122),
					Status:   "PENDING11",
					UserID:   "customer012",
					DueDate:  dateRandom,
					Amount:   decimal.NewFromFloat(float64(110000)),
					Statuses: []string{"PENDING", "CLOSED"},
				},
			},
			sqlErr:  sql.ErrTxDone,
			want:    nil,
			wantErr: true,
		},
		{
			name: "given happy case but the format amount is invalid from db (edge case)," +
				"when findLoan," +
				"then return the result from db",
			args: args{
				loanEntity: &LoanEntity{
					ID:       uint64(122),
					Status:   "PENDING11",
					UserID:   "customer012",
					DueDate:  dateRandom,
					Amount:   decimal.NewFromFloat(float64(110000)),
					Statuses: []string{"PENDING", "CLOSED"},
				},
			},
			sqlRows: sqlmock.NewRows(
				[]string{
					"id",
					"status",
					"user_id",
					"due_date",
					"amount",
					"created_at",
					"version",
					"updated_at",
				}).
				AddRow(
					le.ID,
					le.Status,
					le.UserID,
					le.DueDate,
					nil,
					le.CreatedAt,
					le.Version,
					le.UpdatedAt,
				),
			want: []*LoanEntity{
				{
					ID:        le.ID,
					Status:    le.Status,
					UserID:    le.UserID,
					DueDate:   le.DueDate,
					Amount:    decimal.NewFromFloat(float64(0)),
					CreatedAt: le.CreatedAt,
					Version:   le.Version,
					UpdatedAt: le.UpdatedAt,
				},
			},
		},
		{
			name: "given negative case because scan," +
				"when findLoan," +
				"then return error",
			args: args{
				loanEntity: &LoanEntity{
					ID:       uint64(122),
					Status:   "PENDING11",
					UserID:   "customer012",
					DueDate:  dateRandom,
					Amount:   decimal.NewFromFloat(float64(110000)),
					Statuses: []string{"PENDING", "CLOSED"},
				},
			},
			sqlRows: sqlmock.NewRows(
				[]string{
					"id",
					"status",
					"user_id",
					"due_date",
					"amount",
					"created_at",
					"version",
					"updated_at",
				}).
				AddRow(
					nil,
					nil,
					nil,
					nil,
					nil,
					le.CreatedAt,
					le.Version,
					le.UpdatedAt,
				),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Errorf("DB Connection Error LoanRepositoryImpl.FindLoans() error = %v", err)
				}
				defer db.Close()

				defer func() {
					if err := mock.ExpectationsWereMet(); err != nil {
						assert.Fail(t, "there were unfulfilled expectations", err.Error())
					}
				}()

				if tt.sqlErr != nil {
					mock.ExpectQuery(regexp.QuoteMeta(querySelect)).
						WillReturnError(tt.sqlErr)
				}

				if tt.sqlRows != nil {
					mock.ExpectQuery(regexp.QuoteMeta(querySelect)).
						WillReturnRows(tt.sqlRows)
				}

				store := NewLoanRepository(db)
				got, err := store.FindLoans(context.Background(), tt.args.loanEntity)

				if (err != nil) != tt.wantErr {
					t.Errorf(
						"LoanRepositoryImpl.FindLoans() error = %v, wantErr %v", err,
						tt.wantErr)
					return
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("LoanRepositoryImpl.FindLoans() = %v, want %v", got, tt.want)
				}
			})
	}
}

func Test_loanRepository_UpdateLoan(t *testing.T) {
	le := LoanEntityUpdate{
		IDs:    []uint64{10, 20},
		Status: "PAID",
	}

	type args struct {
		loanEntity *LoanEntityUpdate
	}
	tests := []struct {
		name      string
		args      args
		sqlErr    error
		sqlResult driver.Result
		want      *LoanEntityUpdate
		wantErr   bool
	}{
		{
			name: "given happy case," +
				"when updateLoan," +
				"then return nil",
			args: args{
				loanEntity: &le,
			},
			sqlResult: sqlmock.NewResult(1, 1),
			want:      &le,
		},
		{
			name: "given negative case because RowsAffected," +
				"when updateLoan," +
				"then return error",
			args: args{
				loanEntity: &le,
			},
			sqlResult: sqlmock.NewErrorResult(sql.ErrConnDone),
			wantErr:   true,
		},
		{
			name: "given negative case because execContext," +
				"when updateLoan," +
				"then return error",
			args: args{
				loanEntity: &le,
			},
			sqlErr:  sql.ErrTxDone,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Errorf("DB Connection Error LoanRepositoryImpl.UpdateLoan() error = %v", err)
				}
				defer db.Close()

				defer func() {
					if err := mock.ExpectationsWereMet(); err != nil {
						assert.Fail(t, "there were unfulfilled expectations", err.Error())
					}
				}()

				querySet, _ := builderUpdate(tt.args.loanEntity)
				queryFull := queryUpdate + querySet + queryUpdateWhere + "(" + buildWhereIn(len(tt.args.loanEntity.IDs)) + ")"

				if tt.sqlErr != nil {
					mock.ExpectExec(regexp.QuoteMeta(queryFull)).
						WithArgs(
							"PAID",
							uint64(10),
							uint64(20),
						).
						WillReturnError(tt.sqlErr)
				}

				if tt.sqlResult != nil {
					mock.ExpectExec(regexp.QuoteMeta(queryFull)).
						WithArgs(
							"PAID",
							uint64(10),
							uint64(20),
						).
						WillReturnResult(tt.sqlResult)
				}

				store := NewLoanRepository(db)
				err = store.UpdateLoan(context.Background(), tt.args.loanEntity)

				if (err != nil) != tt.wantErr {
					t.Errorf(
						"LoanRepositoryImpl.UpdateLoan() error = %v, wantErr %v", err,
						tt.wantErr)
					return
				}
			})
	}
}

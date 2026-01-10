package recon_test

import (
	"amartha-recon-service/application/recon"
	"amartha-recon-service/mocks"
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestService_Proceed(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now.Add(24 * time.Hour)

	t.Run("error max rows transaction", func(t *testing.T) {
		cfg := mocks.NewConfiguration(t)
		cfg.On("GetInt", "max.rows.transactions").Return(int64(1))
		svc := recon.NewService(cfg, nil)

		file := recon.NewUploadFile(
			[]recon.TransactionUploadFile{
				{TransactionID: "1"},
				{TransactionID: "2"},
			},
			nil,
			startDate,
			endDate,
		)

		res, err := svc.Proceed(ctx, file)
		assert.Error(t, err)
		assert.Equal(t, recon.ErrorMaxRows, err)
		assert.Empty(t, res.ResultReconciliation)
	})

	t.Run("error max rows bank", func(t *testing.T) {
		cfg := mocks.NewConfiguration(t)
		cfg.On("GetInt", "max.rows.transactions").Return(int64(100))
		cfg.On("GetInt", "max.rows.bank").Return(int64(1))
		svc := recon.NewService(cfg, nil)

		file := recon.NewUploadFile(
			[]recon.TransactionUploadFile{
				{TransactionID: "1"},
			},
			[]recon.BankStatementUploadFile{
				{UniqueID: "B1"},
				{UniqueID: "B2"},
			},
			startDate,
			endDate,
		)

		res, err := svc.Proceed(ctx, file)
		assert.Error(t, err)
		assert.Equal(t, recon.ErrorMaxRows, err)
		assert.Empty(t, res.ResultReconciliation)
	})

	t.Run("success with filtering and matches", func(t *testing.T) {
		cfg := mocks.NewConfiguration(t)
		cfg.On("GetInt", "max.rows.transactions").Return(int64(100))
		cfg.On("GetInt", "max.rows.bank").Return(int64(100))
		cfg.On("GetInt", "max.chunk").Return(int64(10))
		svc := recon.NewService(cfg, nil)

		file := recon.NewUploadFile(
			[]recon.TransactionUploadFile{
				{
					TransactionID:   "TX1",
					Amount:          decimal.NewFromInt(1000),
					BankCode:        "BANK1",
					TransactionTime: now,
				},
				{
					TransactionID:   "TX2",
					Amount:          decimal.NewFromInt(2000),
					BankCode:        "BANK1",
					TransactionTime: now.Add(-48 * time.Hour), // Out of range
				},
				{
					TransactionID:   "TX3",
					Amount:          decimal.NewFromInt(3000),
					BankCode:        "BANK1",
					TransactionTime: now,
				},
			},
			[]recon.BankStatementUploadFile{
				{
					UniqueID: "TX1",
					Amount:   decimal.NewFromInt(1000),
					BankCode: "BANK1",
					Date:     now,
				},
				{
					UniqueID: "TX3",
					Amount:   decimal.NewFromInt(3500), // Discrepancy
					BankCode: "BANK1",
					Date:     now,
				},
				{
					UniqueID: "BANK_ONLY",
					Amount:   decimal.NewFromInt(500),
					BankCode: "BANK1",
					Date:     now,
				},
			},
			startDate,
			endDate,
		)

		res, err := svc.Proceed(ctx, file)
		assert.NoError(t, err)
		assert.Len(t, res.ResultReconciliation, 1)

		reconRes := res.ResultReconciliation[0]
		assert.Equal(t, "BANK1", reconRes.BankCode)
		assert.Equal(t, 2, reconRes.TotalNumberOfTransactions)                           // TX1, TX3 (TX2 filtered out)
		assert.Equal(t, 1, reconRes.TotalNumberOfMatchesTransactions)                    // TX1
		assert.Equal(t, 1, reconRes.TotalNumberOfUnmatchedTransactions)                  // TX3
		assert.True(t, decimal.NewFromInt(500).Equal(reconRes.TotalAmountDiscrepancies)) // 3500 - 3000 = 500

		assert.Len(t, reconRes.ResultReconciliationDetails.TransactionMismatched, 1)
		assert.Equal(t, "TX3", reconRes.ResultReconciliationDetails.TransactionMismatched[0].TransactionID)

		assert.Len(t, reconRes.ResultReconciliationDetails.BankStatementMismatched, 1)
		assert.Equal(t, "BANK_ONLY", reconRes.ResultReconciliationDetails.BankStatementMismatched[0].UniqueID)
	})

	t.Run("success with multiple banks and chunking", func(t *testing.T) {
		cfg := mocks.NewConfiguration(t)
		cfg.On("GetInt", "max.rows.transactions").Return(int64(100))
		cfg.On("GetInt", "max.rows.bank").Return(int64(100))
		cfg.On("GetInt", "max.chunk").Return(int64(2))
		svc := recon.NewService(cfg, nil)

		file := recon.NewUploadFile(
			[]recon.TransactionUploadFile{
				{TransactionID: "B1TX1", Amount: decimal.NewFromInt(100), BankCode: "BANK1", TransactionTime: now},
				{TransactionID: "B1TX2", Amount: decimal.NewFromInt(200), BankCode: "BANK1", TransactionTime: now},
				{TransactionID: "B2TX1", Amount: decimal.NewFromInt(300), BankCode: "BANK2", TransactionTime: now},
			},
			[]recon.BankStatementUploadFile{
				{UniqueID: "B1TX1", Amount: decimal.NewFromInt(100), BankCode: "BANK1", Date: now},
				{UniqueID: "B1TX2", Amount: decimal.NewFromInt(200), BankCode: "BANK1", Date: now},
				{UniqueID: "B2TX1", Amount: decimal.NewFromInt(300), BankCode: "BANK2", Date: now},
			},
			startDate,
			endDate,
		)

		res, err := svc.Proceed(ctx, file)
		assert.NoError(t, err)
		assert.Len(t, res.ResultReconciliation, 2)
		assert.Equal(t, "BANK1", res.ResultReconciliation[0].BankCode)
		assert.Equal(t, "BANK2", res.ResultReconciliation[1].BankCode)
	})

	t.Run("success with no bank entries for a code", func(t *testing.T) {
		cfg := mocks.NewConfiguration(t)
		cfg.On("GetInt", "max.rows.transactions").Return(int64(100))
		cfg.On("GetInt", "max.rows.bank").Return(int64(100))
		cfg.On("GetInt", "max.chunk").Return(int64(2))
		svc := recon.NewService(cfg, nil)

		file := recon.NewUploadFile(
			[]recon.TransactionUploadFile{
				{TransactionID: "B1TX1", Amount: decimal.NewFromInt(100), BankCode: "BANK1", TransactionTime: now},
			},
			[]recon.BankStatementUploadFile{
				{UniqueID: "B2TX1", Amount: decimal.NewFromInt(300), BankCode: "BANK2", Date: now},
			},
			startDate,
			endDate,
		)

		res, err := svc.Proceed(ctx, file)
		assert.NoError(t, err)
		assert.Len(t, res.ResultReconciliation, 2)
	})

	t.Run("NewUploadFile", func(t *testing.T) {
		uf := recon.NewUploadFile(nil, nil, startDate, endDate)
		assert.NotNil(t, uf)
	})
}

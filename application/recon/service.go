package recon

import (
	"amartha-recon-service/configuration"
	"amartha-recon-service/infrastructure/repository/transaction"
	"context"
	"errors"
)

var (
	ErrorMaxRows error = errors.New("file yang diupload terlalu besar")
)

type (
	service struct {
		cfg        configuration.Configuration
		repository transaction.Repository
	}

	Service interface {
		Proceed(ctx context.Context, file *uploadFile) error
	}
)

func NewService(
	cfg configuration.Configuration,
	repository transaction.Repository) Service {
	return &service{
		cfg:        cfg,
		repository: repository,
	}
}

func (s service) Proceed(ctx context.Context, file *uploadFile) error {
	lengthTransaction := len(file.transactionFile)
	maxRowsTransaction := int(s.cfg.GetInt("max.rows.transactions"))

	if lengthTransaction > maxRowsTransaction {
		return ErrorMaxRows
	}

	lengthBank := len(file.bankFile)
	maxRowsBank := int(s.cfg.GetInt("max.rows.bank"))

	if lengthBank > maxRowsBank {
		return ErrorMaxRows
	}

	maxLen := lengthTransaction
	if lengthBank > maxLen {
		maxLen = lengthBank
	}

	// 1. Group transactions and bank statements by BankCode
	transactionsByBank := make(map[string][]TransactionUploadFile)
	for _, tx := range file.transactionFile {
		transactionsByBank[tx.BankCode] = append(transactionsByBank[tx.BankCode], tx)
	}

	bankByBank := make(map[string][]BankStatementUploadFile)
	for _, b := range file.bankFile {
		bankByBank[b.BankCode] = append(bankByBank[b.BankCode], b)
	}

	// 2. Identify all unique bank codes
	uniqueBanks := make(map[string]struct{})
	for code := range transactionsByBank {
		uniqueBanks[code] = struct{}{}
	}

	for code := range bankByBank {
		uniqueBanks[code] = struct{}{}
	}

	// 3. Create a channel to collect results and use a WaitGroup to manage goroutines
	maxChunk := int(s.cfg.GetInt("max.chunk"))
	resultsChan := make(chan ResultReconciliation)
	totalWorkers := 0

	for bankCode := range uniqueBanks {
		txs := transactionsByBank[bankCode]
		banks := bankByBank[bankCode]

		// Determine the largest slice to calculate chunk size
		maxLen := len(txs)
		if len(banks) > maxLen {
			maxLen = len(banks)
		}

		if maxLen == 0 {
			continue
		}

		// Calculate how many rows per chunk
		chunkSize := (maxLen + maxChunk - 1) / maxChunk

		for i := 0; i < maxLen; i += chunkSize {
			end := i + chunkSize

			// Safe slicing for transactions
			txEnd := end
			if txEnd > len(txs) {
				txEnd = len(txs)
			}

			var txChunk []TransactionUploadFile
			if i < len(txs) {
				txChunk = txs[i:txEnd]
			}

			// Safe slicing for bank statements
			bankEnd := end
			if bankEnd > len(banks) {
				bankEnd = len(banks)
			}

			var bankChunk []BankStatementUploadFile
			if i < len(banks) {
				bankChunk = banks[i:bankEnd]
			}

			totalWorkers++
			go func(tc []TransactionUploadFile, bc []BankStatementUploadFile) {
				resultsChan <- s.reconcile(tc, bc)
			}(txChunk, bankChunk)
		}
	}

	var finalResults []ResultReconciliation
	for i := 0; i < totalWorkers; i++ {
		res := <-resultsChan
		finalResults = append(finalResults, res)
	}

	return nil
}

func (s service) reconcile(txs []TransactionUploadFile, banks []BankStatementUploadFile) ResultReconciliation {
	result := ResultReconciliation{
		ResultReconciliationDetails: ResultReconciliationDetails{
			TransactionMismatched:   []TransactionUploadFile{},
			BankStatementMismatched: []BankStatementUploadFile{},
		},
	}

	// 1. Masukkan data bank ke dalam Map untuk pencarian cepat (O(1))
	// Gunakan UniqueID atau RRN sebagai key
	bankMap := make(map[string]BankStatementUploadFile)
	for _, b := range banks {
		bankMap[b.UniqueID] = b
	}

	// 2. Iterasi transaksi sistem dan cari di map bank
	matchedBankIDs := make(map[string]bool)
	for _, tx := range txs {
		bankEntry, found := bankMap[tx.TransactionID]

		if found {
			matchedBankIDs[tx.TransactionID] = true
			if tx.Amount.Equal(bankEntry.Amount) {
				result.TotalNumberOfMatchesTransactions++
			} else {
				// Jika ID cocok tapi nominal beda: hitung selisih absolut
				diff := tx.Amount.Sub(bankEntry.Amount).Abs()
				result.TotalAmountDiscrepancies = result.TotalAmountDiscrepancies.Add(diff)

				// Masukkan ke mismatched karena nominal tidak sama persis
				result.TotalNumberOfUnmatchedTransactions++
				result.ResultReconciliationDetails.TransactionMismatched =
					append(result.ResultReconciliationDetails.TransactionMismatched, tx)
			}
		} else {
			// Jika ID tidak ditemukan sama sekali di bank
			result.TotalNumberOfUnmatchedTransactions++
			result.ResultReconciliationDetails.TransactionMismatched =
				append(result.ResultReconciliationDetails.TransactionMismatched, tx)
		}
	}

	// 3. Cari data bank yang TIDAK ADA di transaksi sistem
	for _, b := range banks {
		if !matchedBankIDs[b.UniqueID] {
			result.ResultReconciliationDetails.BankStatementMismatched =
				append(result.ResultReconciliationDetails.BankStatementMismatched, b)
		}
	}

	result.TotalNumberOfTransactions = len(txs)
	return result
}

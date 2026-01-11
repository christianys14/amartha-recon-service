package recon

import (
	"amartha-recon-service/configuration"
	"amartha-recon-service/infrastructure/repository/transaction"
	"context"
	"errors"
	"sort"
)

var (
	ErrorMaxRows = errors.New("file yang diupload terlalu besar")
)

type (
	service struct {
		cfg        configuration.Configuration
		repository transaction.Repository
	}

	Service interface {
		Proceed(ctx context.Context, file *UploadFile) (ShowResultReconciliation, error)
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

func (s *service) Proceed(ctx context.Context, file *UploadFile) (ShowResultReconciliation, error) {
	lengthTransaction := len(file.transactionFile)
	maxRowsTransaction := int(s.cfg.GetInt("max.rows.transactions"))

	if lengthTransaction > maxRowsTransaction {
		return ShowResultReconciliation{}, ErrorMaxRows
	}

	lengthBank := len(file.bankFile)
	maxRowsBank := int(s.cfg.GetInt("max.rows.bank"))

	if lengthBank > maxRowsBank {
		return ShowResultReconciliation{}, ErrorMaxRows
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
			go func(tx []TransactionUploadFile, bx []BankStatementUploadFile, bc string) {
				resultsChan <- s.reconcile(tx, bx, bc)
			}(txChunk, bankChunk, bankCode)
		}
	}

	var finalResults []ResultReconciliation
	for i := 0; i < totalWorkers; i++ {
		res := <-resultsChan
		finalResults = append(finalResults, res)
	}

	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].BankCode < finalResults[j].BankCode
	})

	return s.showResultReconciliation(finalResults), nil
}

func (s *service) reconcile(
	txs []TransactionUploadFile,
	banks []BankStatementUploadFile,
	bc string) ResultReconciliation {
	result := ResultReconciliation{
		ResultReconciliationDetails: ResultReconciliationDetails{
			TransactionMismatched:   []TransactionUploadFile{},
			BankStatementMismatched: []BankStatementUploadFile{},
		},
	}

	// 1. Store bank data into a Map for fast lookup (O(1))
	// Use UniqueID or RRN as the key
	bankMap := make(map[string]BankStatementUploadFile)
	for _, b := range banks {
		bankMap[b.UniqueID] = b
	}

	// 2. Iterate through system transactions and look them up in the bank map
	matchedBankIDs := make(map[string]bool)
	for _, tx := range txs {
		bankEntry, found := bankMap[tx.TransactionID]

		if found {
			matchedBankIDs[tx.TransactionID] = true
			if tx.Amount.Equal(bankEntry.Amount) {
				result.TotalNumberOfMatchesTransactions++
			} else {
				// If ID matches but amount differs: calculate absolute discrepancy
				diff := tx.Amount.Sub(bankEntry.Amount).Abs()
				result.TotalAmountDiscrepancies = result.TotalAmountDiscrepancies.Add(diff)

				// Add to mismatched because the amount is not an exact match
				result.TotalNumberOfUnmatchedTransactions++
				result.ResultReconciliationDetails.TransactionMismatched =
					append(result.ResultReconciliationDetails.TransactionMismatched, tx)
			}
		} else {
			// If ID is not found in the bank data at all
			result.TotalNumberOfUnmatchedTransactions++
			result.ResultReconciliationDetails.TransactionMismatched =
				append(result.ResultReconciliationDetails.TransactionMismatched, tx)
		}
	}

	// 3. Find bank data that DOES NOT EXIST in system transactions
	for _, b := range banks {
		if !matchedBankIDs[b.UniqueID] {
			result.ResultReconciliationDetails.BankStatementMismatched =
				append(result.ResultReconciliationDetails.BankStatementMismatched, b)
		}
	}

	result.TotalNumberOfTransactions = len(txs)
	result.BankCode = bc
	return result
}

func (s *service) showResultReconciliation(finalResult []ResultReconciliation) ShowResultReconciliation {
	mergedMap := make(map[string]*ResultReconciliation)

	for _, fr := range finalResult {
		if existing, ok := mergedMap[fr.BankCode]; ok {
			existing.TotalNumberOfTransactions += fr.TotalNumberOfTransactions
			existing.TotalNumberOfMatchesTransactions += fr.TotalNumberOfMatchesTransactions
			existing.TotalNumberOfUnmatchedTransactions += fr.TotalNumberOfUnmatchedTransactions
			existing.TotalAmountDiscrepancies = existing.TotalAmountDiscrepancies.Add(fr.TotalAmountDiscrepancies)

			existing.ResultReconciliationDetails.TransactionMismatched = append(
				existing.ResultReconciliationDetails.TransactionMismatched,
				fr.ResultReconciliationDetails.TransactionMismatched...,
			)

			existing.ResultReconciliationDetails.BankStatementMismatched = append(
				existing.ResultReconciliationDetails.BankStatementMismatched,
				fr.ResultReconciliationDetails.BankStatementMismatched...,
			)
		} else {
			// Create a copy to avoid mutating original slice elements if needed
			item := fr
			mergedMap[fr.BankCode] = &item
		}
	}

	var result []ResultReconciliation
	for _, v := range mergedMap {
		result = append(result, *v)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].BankCode < result[j].BankCode
	})

	return ShowResultReconciliation{
		ResultReconciliation: result,
	}
}

package recon

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	UploadFile struct {
		transactionFile []TransactionUploadFile
		bankFile        []BankStatementUploadFile
		startDate       time.Time
		endDate         time.Time
	}

	TransactionUploadFile struct {
		TransactionID   string          `json:"transaction_id"`
		TerminalRRN     string          `json:"terminal_rrn"`
		Amount          decimal.Decimal `json:"amount"`
		TransactionType string          `json:"transaction_type"`
		BankCode        string          `json:"bank_code"`
		TransactionTime time.Time       `json:"transaction_time"`
	}

	BankStatementUploadFile struct {
		UniqueID string          `json:"unique_id"`
		Amount   decimal.Decimal `json:"amount"`
		Date     time.Time       `json:"date"`
		BankCode string          `json:"bank_code"`
	}

	ResultReconciliation struct {
		TotalNumberOfTransactions          int                         `json:"total_number_of_transactions"`
		TotalNumberOfMatchesTransactions   int                         `json:"total_number_of_matches_transactions"`
		TotalNumberOfUnmatchedTransactions int                         `json:"total_number_of_unmatched_transactions"`
		ResultReconciliationDetails        ResultReconciliationDetails `json:"result_reconciliation_details"`
		TotalAmountDiscrepancies           decimal.Decimal             `json:"total_amount_discrepancies"`
		BankCode                           string                      `json:"bank_code"`
	}

	ResultReconciliationDetails struct {
		TransactionMismatched   []TransactionUploadFile   `json:"transaction_mismatched"`
		BankStatementMismatched []BankStatementUploadFile `json:"bank_statement_mismatched"`
	}

	ShowResultReconciliation struct {
		ResultReconciliation []ResultReconciliation `json:"result_reconciliation"`
	}
)

func NewUploadFile(
	transactionFile []TransactionUploadFile,
	bankFile []BankStatementUploadFile,
	startDate, endDate time.Time) *UploadFile {
	return &UploadFile{
		transactionFile: transactionFile,
		bankFile:        bankFile,
		startDate:       startDate,
		endDate:         endDate,
	}
}

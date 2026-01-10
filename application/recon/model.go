package recon

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	uploadFile struct {
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
)

func NewUploadFile(
	transactionFile []TransactionUploadFile,
	bankFile []BankStatementUploadFile,
	startDate, endDate time.Time) *uploadFile {
	return &uploadFile{
		transactionFile: transactionFile,
		bankFile:        bankFile,
		startDate:       startDate,
		endDate:         endDate,
	}
}

package http

import (
	"amartha-recon-service/application/recon"
	"amartha-recon-service/common"
	constant2 "amartha-recon-service/constant"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type (
	controller struct {
		service recon.Service
	}

	Controller interface {
		Proceed(w http.ResponseWriter, r *http.Request)
	}
)

func NewController(service recon.Service) Controller {
	return &controller{service: service}
}

func (c *controller) Proceed(w http.ResponseWriter, r *http.Request) {
	startDate := r.FormValue("start_date")
	startDateParse, err := time.Parse(time.DateOnly, startDate)
	if err != nil {
		log.Printf("error parsing startDate: %v", err)
		common.ToErrorResponse(w,
			constant2.HttpRc[constant2.ValusIsMismatach],
			constant2.HttpRcDescription[constant2.ValusIsMismatach],
		)
		return
	}

	endDate := r.FormValue("end_date")
	endDateParse, err := time.Parse(time.DateOnly, endDate)
	if err != nil {
		log.Printf("error parsing endDate: %v", err)
		common.ToErrorResponse(w,
			constant2.HttpRc[constant2.ValusIsMismatach],
			constant2.HttpRcDescription[constant2.ValusIsMismatach],
		)
		return
	}

	fileSystem, _, err := r.FormFile("system")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			log.Printf("client did not send file system")
		}

		common.ToErrorResponse(w,
			constant2.HttpRc[constant2.Validation],
			constant2.HttpRcDescription[constant2.Validation],
		)
		log.Printf("error reading fileSystem: %v", err)
		return
	}
	defer fileSystem.Close()

	fileBank, _, err := r.FormFile("bank")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			log.Printf("client did not send file bank")
		}
		common.ToErrorResponse(w,
			constant2.HttpRc[constant2.Validation],
			constant2.HttpRcDescription[constant2.Validation],
		)
		log.Printf("error reading fileBank: %v", err)
		return
	}
	defer fileBank.Close()

	readerFileSystem := csv.NewReader(fileSystem)
	transactionUploadFiles, err := parseTransactionsFromCSV(readerFileSystem)
	if err != nil {
		common.ToErrorResponse(w,
			constant2.HttpRc[constant2.Validation],
			constant2.HttpRcDescription[constant2.Validation],
		)
		log.Printf("error parsing fileSystem: %v", err)
		return
	}

	readerBank := csv.NewReader(fileBank)
	bankStatementUploadFiles, err := parseBankFromCSV(readerBank)
	if err != nil {
		common.ToErrorResponse(w,
			constant2.HttpRc[constant2.Validation],
			constant2.HttpRcDescription[constant2.Validation],
		)
		log.Printf("error parsing fileBank: %v", err)
		return
	}

	uploadFile := recon.NewUploadFile(transactionUploadFiles, bankStatementUploadFiles, startDateParse, endDateParse)
	response, err := c.service.Proceed(r.Context(), uploadFile)
	if err != nil {
		common.ToErrorResponse(w,
			constant2.HttpRc[constant2.GeneralError],
			constant2.HttpRcDescription[constant2.GeneralError],
		)
		log.Printf("error invoke service: %v", err)
		return
	}

	common.ToSuccessResponse(w, nil, response)
}

func parseTransactionsFromCSV(reader *csv.Reader) ([]recon.TransactionUploadFile, error) {
	// Skip header
	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	var transactions []recon.TransactionUploadFile
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		transactions = append(transactions, parseTransactionRow(row))
	}

	return transactions, nil
}

func parseTransactionRow(row []string) recon.TransactionUploadFile {
	tfs := recon.TransactionUploadFile{}

	if len(row) > 0 {
		tfs.TransactionID = row[0]
	}

	if len(row) > 1 {
		tfs.TerminalRRN = row[1]
	}

	if len(row) > 2 {
		if amount, err := decimal.NewFromString(row[2]); err == nil {
			tfs.Amount = amount
		}
	}

	if len(row) > 3 {
		tfs.TransactionType = row[3]
	}

	if len(row) > 4 {
		tfs.BankCode = row[4]
	}

	if len(row) > 5 {
		if dt, err := time.Parse(time.DateTime, row[5]); err == nil {
			tfs.TransactionTime = dt
		}
	}

	return tfs
}

func parseBankFromCSV(reader *csv.Reader) ([]recon.BankStatementUploadFile, error) {
	// Skip header
	if _, err := reader.Read(); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	var bankStatementUploadFiles []recon.BankStatementUploadFile
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		bankStatementUploadFiles = append(bankStatementUploadFiles, parseBankRow(row))
	}

	return bankStatementUploadFiles, nil
}

func parseBankRow(row []string) recon.BankStatementUploadFile {
	bsu := recon.BankStatementUploadFile{}

	if len(row) > 0 {
		bsu.UniqueID = row[0]
	}

	if len(row) > 1 {
		if amount, err := decimal.NewFromString(row[1]); err == nil {
			bsu.Amount = amount
		}
	}

	if len(row) > 2 {
		if dt, err := time.Parse(time.DateTime, row[2]); err == nil {
			bsu.Date = dt
		}
	}

	if len(row) > 3 {
		bsu.BankCode = row[3]
	}

	return bsu
}

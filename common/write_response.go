package common

import (
	constant2 "amartha-recon-service/constant"
	"encoding/json"
	"log"
	"net/http"
)

const (
	contentType = "Content-type"
	application = "application/json"
)

type BillingResponse struct {
	Rc         string      `json:"rc,omitempty"`
	Message    string      `json:"message,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func NewBillingResponse(
	rc string,
	message string,
	pagination interface{},
	data interface{}) *BillingResponse {
	return &BillingResponse{
		Rc:         rc,
		Message:    message,
		Pagination: pagination,
		Data:       data,
	}
}

func responseWrite(rw http.ResponseWriter, data interface{}, statusCode int) {
	responseByte, err := json.Marshal(data)
	if err != nil {
		log.Println("error during encode responseWrite", err)
	}

	rw.WriteHeader(statusCode)
	_, err = rw.Write(responseByte)
}

func ToSuccessResponse(writer http.ResponseWriter, pagination interface{}, data interface{}) {
	rc := constant2.HttpRc[constant2.Success]
	rcDesc := constant2.HttpRcDescription[constant2.Success]
	httpRes := constant2.BillingCodeToHttpCode[rc]

	responseWrite(
		writer,
		NewBillingResponse(rc, rcDesc, pagination, data),
		httpRes,
	)
}

func ToErrorResponse(writer http.ResponseWriter, rc, rcDesc string) {
	httpRes := constant2.BillingCodeToHttpCode[rc]

	responseWrite(
		writer,
		NewBillingResponse(rc, rcDesc, nil, nil),
		httpRes,
	)
}

package handler

import (
	"encoding/xml"
	"net/http"
	"strconv"
)

type S3Responser interface {
	Send(http.ResponseWriter)
}

type S3ErrorResponse struct {
	httpStatus int
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`
	Message    string   `xml:"Message"`
	Resource   string   `xml:"Resource"`
	RequestId  string   `xml:"RequestId"`
}

func NewS3ErrorResponse(status int, code, message, resource, requestId string) S3Responser {
	return &S3ErrorResponse{
		httpStatus: status,
		Code:       code,
		Message:    message,
		Resource:   resource,
		RequestId:  requestId,
	}
}

func FormatS3Response(resp S3Responser, typ string) ([]byte, error) {
	if typ == "application/xml" {
		return xml.MarshalIndent(resp, "", "  ")
	} else {
		return nil, nil
	}
}

func (resp *S3ErrorResponse) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(resp.httpStatus)
	w.Write(body)
}

type S3NilResponse struct{
	status int
}

func NewS3NilResponse(status int) S3Responser {
	return &S3NilResponse{status}
}

func (resp *S3NilResponse) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(resp.status)
}


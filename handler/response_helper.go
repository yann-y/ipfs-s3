package handler

import (
	"fmt"
	"github.com/yann-y/ipfs-s3/context"
	"net/http"
)

//func WrapErrorResponseForRequest(r *http.Request, code, resource string) ResponseFormatter {
//	message, ok := ErrorMessage(code)
//	if !ok {
//		panic(fmt.Sprintf("invalid error code: %s", code))
//	}
//
//	return NewErrorResponse(
//		code,
//		message,
//		resource,
//		context.Get(r, "req_id").(string),
//	)
//}

func WrapS3ErrorResponseForRequest(status int, r *http.Request, code, resource string) S3Responser {
	message, ok := ErrorMessage(code)
	if !ok {
		panic(fmt.Sprintf("invalid error code: %s", code))
	}

	return NewS3ErrorResponse(
		status,
		code,
		message,
		resource,
		context.Get(r, "req_id").(string),
	)
}

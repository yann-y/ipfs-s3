package bucket

import (
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/mux"
	// "github.com/yann-y/ipfs-s3/mongodb/dao"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"net/http"
)

func DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		context.Set(r, "response", resp)
		// resp.Send(w)
	}()

	bucket := mux.Vars(r)["bucket"]

	err = db.ActiveService().DeleteBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket)
		} else if err == db.BucketNotEmptyError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusConflict, r, "BucketNotEmpty", "/"+bucket)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}
	resp = handler.NewS3NilResponse(http.StatusNoContent)
	return
}

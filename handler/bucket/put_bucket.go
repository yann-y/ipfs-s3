package bucket

import (
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"
	"net/http"
	"time"
)

func PutBucketHandler(w http.ResponseWriter, r *http.Request) {
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

	bucketBean := &db.Bucket{
		BucketName: bucket,
		// TODO: owner is always 1
		UserID:     "1",
		ACL:        1,
		CreateTime: time.Now().Unix(),
	}

	err = db.ActiveService().PutBucket(bucketBean)
	if err != nil {
		if err == db.BucketExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusConflict, r, "BucketAlreadyExists", "/"+bucket)
		} else if err == db.TooManyBucketError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusConflict, r, "TooManyBuckets", "/"+bucket)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}
	resp = handler.NewS3NilResponse(http.StatusOK)
	return
}

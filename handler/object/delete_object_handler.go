package object

import (
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"
	"net/http"
)

func DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser
	deleteMarker := false

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		if deleteMarker == true {
			w.Header().Set("x-amz-delete-marker", "true")
		}
		// resp.Send(w)
		context.Set(r, "response", resp)
	}()

	bucket := mux.Vars(r)["bucket"]
	version := mux.Vars(r)["versionId"]
	objectName := mux.Vars(r)["object"]

	bucketObject, err := db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket+objectName)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}

	objectPath := "/" + bucket + "/" + objectName
	if err = db.ActiveService().DeleteObject(bucket, objectName, version, bucketObject.VersionEnabled); err != nil {
		if err == db.ObjectNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchKey", objectPath)
		} else if err == db.ObjectDeletedError {
			deleteMarker = true
			resp = handler.NewS3NilResponse(http.StatusNoContent)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", objectPath)
		}
		return
	}
	resp = handler.NewS3NilResponse(http.StatusNoContent)
	return
}

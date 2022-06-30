package object

import (
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/internal/hash"
	"github.com/yann-y/ipfs-s3/internal/storage"
	"github.com/yann-y/ipfs-s3/mux"
	// "github.com/yann-y/ipfs-s3/utils/md5"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func parseContentLength(req *http.Request) (int64, error) {
	contentLen := req.Header.Get("Content-Length")
	if contentLen == "" {
		return -1, errors.New("Content length missing")
	}

	size, err := strconv.ParseInt(contentLen, 10, 64)
	if err != nil || size < 0 {
		return -1, errors.New("Content length invalid")
	}

	return size, nil
}

func validateContentLength(req *http.Request) (int64, error) {
	size, err := parseContentLength(req)
	if err != nil {
		return -1, err
	}

	if size > 5*1024*1024*1024 {
		return -1, fmt.Errorf("Content length %d too large, please use multiupload", size)
	}
	return size, nil
}

type S3PutObjectResponse struct {
	etag string
}

func (resp *S3PutObjectResponse) Send(w http.ResponseWriter) {
	w.Header().Set("ETag", resp.etag)
	w.WriteHeader(http.StatusOK)
}

func PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		// resp.Send(w)
		context.Set(r, "response", resp)
	}()

	objectName := mux.Vars(r)["object"]

	bucket := mux.Vars(r)["bucket"]

	var bucketObject *db.Bucket
	bucketObject, err = db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket+objectName)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}

	var size int64
	size, err = validateContentLength(r)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidParameter", "/")
		return
	}

	objectPath := bucket + "/" + objectName
	// keyMD5sum := md5.Sum(objectPath)

	var cid, objectMd5 string
	reader, err := hash.NewReader(r.Body, size, "", "", -1)
	if err != nil {
		fmt.Errorf("%v", err)
	}
	defer r.Body.Close()
	cid, objectMd5, err = storage.FS.PutObject(reader)
	//Cid, objectMd5, err = fs.PutObject(bucket, size, r.Body, context.Get(r, "req_id").(string))

	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}
	if size <= 0 {
		size = reader.ReadSize()
	}
	objectMeta := &db.Object{
		ObjectName:   objectName,
		Cid:          cid,
		Etag:         fmt.Sprintf("%s%s%s", "\"", objectMd5, "\""),
		Bucket:       bucketObject.BucketName,
		Size:         size,
		LastModified: time.Now().Unix(),
	}

	if err = db.ActiveService().PutObject(objectMeta, bucketObject.VersionEnabled); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}
	resp = &S3PutObjectResponse{objectMd5}
	return
}

package object

import (
	"errors"
	"fmt"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/fs"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"
	"net/http"
	"strconv"
	"time"
)

func parsePartNumber(r *http.Request) (int, error) {
	partNumber := r.Form.Get("partNumber")
	if partNumber == "" {
		return -1, errors.New("miss parameter partNumber")
	}
	partNumberI, err := strconv.Atoi(partNumber)
	if err != nil {
		return -1, err
	}
	return partNumberI, nil
}

var (
	AbortedUploadError    = errors.New("upload has been aborted")
	InvalidPartOrderError = errors.New("invalid part order")
)

func validateUploadId(uploadId string) error {
	upload, err := db.ActiveService().GetUpload(uploadId)
	if err != nil {
		return err
	}
	if upload.IsAbort == true {
		return AbortedUploadError
	}
	return nil
}

//func validatePartNumber(uploadId string, number int) error {
//	lastPartId, err := db.ActiveService().GetLastUploadPart(uploadId)
//	if err != nil {
//		if err == db.UploadPartNotExistError {
//			if number != 1 {
//				return InvalidPartOrderError
//			} else {
//				return nil
//			}
//		}
//		return err
//	}
//	if number != lastPartId + 1 {
//		return InvalidPartOrderError
//	}
//	return nil
//}

type S3UploadPartResponse struct {
	etag string
}

func (resp *S3UploadPartResponse) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Length", "0")
	w.Header().Set("ETag", resp.etag)
	w.WriteHeader(http.StatusOK)
}

// TODO: 增加对Expect参数支持,详见http://docs.aws.amazon.com/AmazonS3/latest/API/mpUploadUploadPart.html
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		// resp.Send(w)
		context.Set(r, "response", resp)
	}()

	if err = r.ParseForm(); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/")
		return
	}

	objectName := mux.Vars(r)["object"]
	bucket := mux.Vars(r)["bucket"]

	_, err = db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket+objectName)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}

	var size int64
	size, err = parseContentLength(r)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidParameter", "/")
		return
	}

	objectPath := bucket + "/" + objectName

	var uploadId string
	uploadId = r.Form.Get("uploadId")
	if uploadId == "" {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidParameter", "/")
		return
	}

	var partNumber int
	partNumber, err = parsePartNumber(r)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidParameter", "/")
		return
	}

	if err = validateUploadId(uploadId); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchUpload", "/")
		return
	}

	// 根据S3的API定义,这里可能会返回InvalidPart和InvalidPartOrder两种错误
	// InvalidPart:
	// InvalidPartOrder: 上传的part顺序错误
	//if err = validatePartNumber(uploadId, partNumber); err != nil {
	//	resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidPartOrder", "/")
	//	return
	//}

	var (
		fid     string
		partMd5 string
	)
	fid, partMd5, err = fs.PutObject(bucket, size, r.Body, context.Get(r, "req_id").(string))
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}

	part := &db.UploadPart{
		UploadID:     uploadId,
		Fid:          fid,
		Number:       partNumber,
		Size:         size,
		LastModified: time.Now().Unix(),
		Etag:         fmt.Sprintf("%s%s%s", "\"", partMd5, "\""),
	}

	if err = db.ActiveService().PutUploadPart(part); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}
	resp = &S3UploadPartResponse{partMd5}
	return
}

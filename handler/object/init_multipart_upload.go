package object

import (
	"encoding/xml"
	"github.com/satori/go.uuid"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type S3InitMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadId string   `xml:"UploadId"`
}

//func (res *S3InitMultipartUploadResult) ContentType() string {
//	return "application/xml"
//}
//
//func (res *S3InitMultipartUploadResult) ContentBody() ([]byte, error) {
//	return handler.FormatS3Response(res, res.ContentType())
//}

func (resp *S3InitMultipartUploadResult) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapS3InitMultipartUploadResponse(bucket, key, uploadId string) *S3InitMultipartUploadResult {
	return &S3InitMultipartUploadResult{
		Bucket:   bucket,
		Key:      key,
		UploadId: uploadId,
	}
}

func InitMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	objectName := mux.Vars(r)["object"]

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		context.Set(r, "response", resp)
		// resp.Send(w)
	}()

	bucket := mux.Vars(r)["bucket"]

	bucketObject, err := db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket+objectName)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}

	uploadId := strings.Trim(strings.Replace(uuid.NewV4().String(), "-", "", -1), "\"")
	upload := &db.UploadInfo{
		UploadID:  uploadId,
		StartTime: time.Now().Unix(),
		Bucket:    bucket,
		Object:    objectName,
		UserID:    bucketObject.UserID,
		IsAbort:   false,
		Meta:      "",
	}
	if err = db.ActiveService().InitMultipartUpload(upload); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", bucket+"/"+objectName)
		return
	}
	resp = wrapS3InitMultipartUploadResponse(bucket, objectName, uploadId)
	return
}

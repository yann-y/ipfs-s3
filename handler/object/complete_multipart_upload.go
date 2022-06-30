package object

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/internal/storage"
	"github.com/yann-y/ipfs-s3/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	InvalidPartError = errors.New("invalid part")
)

type PartEntry struct {
	PartNumber int
	ETag       string
}

type CompleteMultipartUploadPara struct {
	Part []*PartEntry
}

type CompleteMultipartUploadResult struct {
	Location string
	Bucket   string
	Key      string
	ETag     string
}

func (resp *CompleteMultipartUploadResult) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapCompleteMultipartUploadResponse(loc, bucket, key, etag string) handler.S3Responser {
	return &CompleteMultipartUploadResult{
		Location: loc,
		Bucket:   bucket,
		Key:      key,
		ETag:     etag,
	}
}

func parseCompleteMultipartUploadParam(req *http.Request) (*CompleteMultipartUploadPara, error) {
	para := CompleteMultipartUploadPara{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(body, &para); err != nil {
		return nil, err
	}
	return &para, nil
}

func validateCompleteParts(uploadedParts []*db.UploadPart, requestedParts *CompleteMultipartUploadPara) error {
	// 根据S3 api,CompleteMultipartUpload请求中的PartNumber必须以升序排列, 否则返回InvalidPartOrder错误
	last := -1
	for _, part := range requestedParts.Part {
		if part.PartNumber <= last {
			return InvalidPartOrderError
		}
	}

	// 检查CompleteMultipartUpload请求中的Part是否有效:
	// 1. 与服务端的PartNumber必须完全吻合
	// 2. 与服务端的Etag必须完全吻合
	if len(uploadedParts) != len(requestedParts.Part) {
		return InvalidPartError
	}

	for i, clientPart := range requestedParts.Part {
		serverPart := uploadedParts[i]
		if clientPart.PartNumber != serverPart.Number {
			return InvalidPartError
		}
		if clientPart.ETag != strings.Trim(serverPart.Etag, "\"") {
			return InvalidPartError
		}
	}
	return nil
}

func CompleteMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		context.Set(r, "response", resp)
		// resp.Send(w)
	}()

	objectName := mux.Vars(r)["object"]
	if err = r.ParseForm(); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/")
		return
	}

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

	objectPath := bucket + "/" + objectName

	var uploadId string
	uploadId = r.Form.Get("uploadId")
	if uploadId == "" {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidParameter", "/")
		return
	}

	if err = validateUploadId(uploadId); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchUpload", "/")
		return
	}

	// parse complete multiupload part body
	var completeReq *CompleteMultipartUploadPara
	completeReq, err = parseCompleteMultipartUploadParam(r)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}

	var uploadedParts []*db.UploadPart
	uploadedParts, err = db.ActiveService().ListUploadAllParts(uploadId)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}

	if err = validateCompleteParts(uploadedParts, completeReq); err != nil {
		if err == InvalidPartOrderError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidPartOrder", "/"+objectPath)
		} else if err == InvalidPartError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidPart", "/"+objectPath)
		}
		return
	}

	partSize := uploadedParts[0].Size
	size := int64(0)
	Cids, etags := "", ""

	// 如果part数量过多,Cids可能会很长,此处进行了优化:
	// 将拼接后的Cids再写入底层GalaxyFS,得到的Cid再作为该对象的最终file id
	// 这样在读入的时候需要判断出对象是通过普通上传还是Multiupload方式上传,因此需要在数据库记录该信息
	// TODO: 将客户端上传的Part的Etag拼接起来计算而得的MD5作为object的md5
	for _, part := range uploadedParts {
		size = size + part.Size
		Cids = Cids + part.Cid + ","
		etags = etags + part.Etag
	}

	hash := md5.Sum([]byte(etags))
	objectEtag := hex.EncodeToString(hash[:])

	// 去掉Cids最后的','
	Cids = strings.TrimSuffix(Cids, ",")

	var objectCid string
	objectCid, _, err = storage.FS.PutObject(strings.NewReader(Cids))
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}

	// keyHash := md5.Sum([]byte(objectPath))
	newObject := &db.Object{
		ObjectName:      objectName,
		Size:            size,
		Etag:            fmt.Sprintf("%s%s%s", "\"", objectEtag, "\""),
		Bucket:          bucket,
		Cid:             objectCid,
		MultipartUpload: true,
		UploadId:        uploadId,
		PartSize:        partSize,
		LastModified:    time.Now().Unix(),
	}

	if err = db.ActiveService().PutObjectFromMultipartUpload(newObject, bucketObject.VersionEnabled); err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+objectPath)
		return
	}
	objectLocation := r.Host + "/" + objectName
	resp = wrapCompleteMultipartUploadResponse(objectLocation, bucket, objectName, objectEtag)
	return
}

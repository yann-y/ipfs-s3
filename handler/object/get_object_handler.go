package object

import (
	"errors"
	"fmt"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/internal/storage"
	"github.com/yann-y/ipfs-s3/mux"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func parseRangeHeader(r *http.Request) (int64, int64, error) {
	rangeV := r.Header.Get("Range")
	if rangeV == "" {
		return -1, math.MaxInt64, nil
	}
	splits := strings.Split(rangeV, "=")
	if len(splits) != 2 || splits[0] != "bytes" {
		return -1, -1, errors.New("invalid range bytes")
	}
	rangeOffsets := strings.Split(splits[1], "-")
	if len(rangeOffsets) != 2 {
		return -1, -1, errors.New("invalid range bytes")
	}
	start, err := strconv.ParseInt(rangeOffsets[0], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	// range maybe '123-'
	if rangeOffsets[1] == "" {
		return start, math.MaxInt64, nil
	}
	end, err := strconv.ParseInt(rangeOffsets[1], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	return start, end, nil
}

func max(a, b int64) int64 {
	if a < b {
		return b
	}
	return a
}

func min(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

type SimpleS3Response struct {
	size     int64
	etag     string
	contents []io.Reader
}

func (resp *SimpleS3Response) Send(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Content-Length", strconv.FormatInt(resp.size, 10))
	w.Header().Set("ETag", resp.etag)
	w.WriteHeader(http.StatusOK)
	for _, r := range resp.contents {
		io.Copy(w, r)
	}
}

func wrapSimpleS3Response(size int64, etag string, contents []io.Reader) *SimpleS3Response {
	return &SimpleS3Response{
		size:     size,
		etag:     etag,
		contents: contents,
	}
}

func GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		context.Set(r, "response", resp)
		// resp.Send(w)
	}()

	var rangeStart, rangeEnd int64
	rangeStart, rangeEnd, err = parseRangeHeader(r)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidRangeHeader", "/")
		return
	}

	bucket := mux.Vars(r)["bucket"]
	version := mux.Vars(r)["versionId"]
	objectName := mux.Vars(r)["object"]
	objectPath := fmt.Sprintf("%s/%s", bucket, objectName)

	bucketObject, err := db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", objectPath)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", objectPath)
		}
		return
	}

	var object *db.Object
	// object, err = db.GetObject(objectPath, version, bucketObject.VersionEnabled)
	object, err = db.ActiveService().GetObject(bucket, objectName, version, bucketObject.VersionEnabled)
	if err != nil {
		if err == db.ObjectNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchKey", objectPath)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", objectPath)
		}
		return
	}

	if object.Size == 0 {
		resp = handler.NewS3NilResponse(http.StatusOK)
		return
	}

	partSize := object.PartSize
	if rangeStart < 0 {
		rangeStart = 0
	}
	if rangeEnd > object.Size {
		rangeEnd = object.Size
	}

	var reader io.Reader
	reader, err = storage.FS.GetObject(object.Cid)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", objectPath)
		return
	}

	// ?????????????????????multipart upload????????????
	// ????????????????????????part???Cid,?????????CidArray
	contents := make([]io.Reader, 0)
	if object.MultipartUpload {
		Cids := string(handler.StreamToByte(reader))
		CidArray := strings.Split(Cids, ",")
		for i, Cid := range CidArray {
			partRangeStart := int64(i) * partSize
			partRangeEnd := int64(i+1) * partSize
			// ???????????????range??????part????????????,?????????part,????????????
			if rangeStart > partRangeEnd || rangeEnd < partRangeStart {
				continue
			}
			reader, err = storage.FS.GetObject(Cid)
			if err != nil {
				resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", objectPath)
				return
			}
			//reader.Seek(max(0, rangeStart-partRangeStart), io.SeekStart)
			//lm := &io.LimitedReader{R: reader, N: min(rangeEnd, partRangeEnd) - partRangeStart}
			//contents = append(contents, lm)
		}
	} else {
		//reader.Seek(rangeStart, io.SeekStart)
		//lm := &io.LimitedReader{R: reader, N: rangeEnd - rangeStart}
		//contents = append(contents, lm)
	}
	resp = wrapSimpleS3Response(rangeEnd-rangeStart, object.Etag, contents)
	return
}

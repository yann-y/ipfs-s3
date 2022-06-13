package object

import (
	"encoding/xml"
	"errors"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"
	"net/http"
	"strconv"
	"time"
)

func validateMaxParts(smax string) (int, error) {
	imax, err := strconv.Atoi(smax)
	if err != nil {
		return -1, err
	}

	if imax < 1 || imax > 1000 {
		return -1, errors.New("max-parts invalid, not in range 1~1000")
	}
	return imax, nil
}

type PartContent struct {
	PartNumber   int    `xml:"PartNumber"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
}

type ListPartsResult struct {
	XMLName              xml.Name       `xml:"ListPartsResult"`
	Bucket               string         `xml:"Bucket"`
	Key                  string         `xml:"Key"`
	UploadId             string         `xml:"UploadId"`
	StorageClass         string         `xml:"StorageClass"`
	PartNumberMarker     int            `xml:"PartNumberMarker"`
	NextPartNumberMarker int            `xml:"NextPartNumberMarker"`
	MaxParts             int            `xml:"MaxParts"`
	IsTruncated          bool           `xml:"IsTruncated"`
	Parts                []*PartContent `xml:"Part"`
}

func (resp *ListPartsResult) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapListPartsResponse(bucket, object string, items []*PartContent, params *listPartsParams, truncated bool, next int) handler.S3Responser {
	return &ListPartsResult{
		Bucket:   bucket,
		Key:      object,
		UploadId: params.uploadId,
		// TODO: 支持其他类型的class
		StorageClass:         "STANDARD",
		PartNumberMarker:     params.marker,
		NextPartNumberMarker: next,
		MaxParts:             params.maxParts,
		IsTruncated:          truncated,
		Parts:                items,
	}
}

//TODO: 支持"encoding-type"参数
type listPartsParams struct {
	uploadId string
	maxParts int
	marker   int
}

func parseListParams(r *http.Request) (*listPartsParams, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	uploadId := r.Form.Get("uploadId")
	if uploadId == "" {
		return nil, errors.New("uploadId missing")
	}

	max := r.Form.Get("max-parts")
	if max == "" {
		max = "1000"
	}

	maxI, err := validateMaxParts(max)
	if err != nil {
		return nil, err
	}

	marker := r.Form.Get("part-number-marker")
	if marker == "" {
		marker = "0"
	}

	iMarker, err := strconv.Atoi(marker)
	if err != nil {
		return nil, err
	}

	return &listPartsParams{
		marker:   iMarker,
		maxParts: maxI,
		uploadId: uploadId,
	}, nil
}

func ListPartsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		// resp.Send(w)
		context.Set(r, "response", resp)
	}()

	var params *listPartsParams
	params, err = parseListParams(r)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidArgument", "/")
		return
	}

	bucket := mux.Vars(r)["bucket"]

	var items []*db.UploadPart
	items, err = db.ActiveService().ListUploadParts(params.uploadId, params.marker, params.maxParts+3)
	if err != nil {
		if err == db.UploadNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchUpload", "/"+bucket)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}
	truncated, nextPartMaker := false, 0
	if len(items) > params.maxParts {
		truncated = true
		nextPartMaker = items[params.maxParts].Number
		items = items[:params.maxParts]
	}

	parts := make([]*PartContent, 0)
	for _, item := range items {
		part := &PartContent{
			PartNumber:   item.Number,
			LastModified: time.Unix(item.LastModified, 0).Format(time.RFC3339),
			Size:         item.Size,
			ETag:         item.Etag,
		}
		parts = append(parts, part)
	}
	object := mux.Vars(r)["object"]
	resp = wrapListPartsResponse(bucket, object, parts, params, truncated, nextPartMaker)
	return
}

package bucket

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"

	"github.com/yann-y/ipfs-s3/db"
	"net/http"
	"strconv"
	"strings"
)

func maxKeyValid(smax string) (int, error) {
	imax, err := strconv.Atoi(smax)
	if err != nil {
		return -1, err
	}

	if imax < 1 || imax > 1000 {
		return -1, errors.New("max-keys invalid, not in range 1~1000")
	}
	return imax, nil
}

type Prefix struct {
	Prefix string
}

type GetBucketResultV2 struct {
	XMLName               xml.Name             `xml:"ListBucketResult"`
	Name                  string               `xml:"Name"`
	Prefix                string               `xml:"Prefix"`
	Delimiter             string               `xml:"Delimiter"`
	KeyCount              int                  `xml:"KeyCount"`
	MaxKeys               int                  `xml:"MaxKeys"`
	IsTruncated           bool                 `xml:"IsTruncated"`
	NextContinuationToken string               `xml:"NextContinuationToken"`
	StartAfter            string               `xml:"StartAfter"`
	Contents              []*db.ListObjectItem `xml:"Contents"`
	CommonPrefixes        []*Prefix            `xml:"CommonPrefixes"`
}

func (res *GetBucketResultV2) ContentType() string {
	return "application/xml"
}

func (res *GetBucketResultV2) ContentBody() ([]byte, error) {
	body, err := handler.FormatS3Response(res, res.ContentType())
	return body, err
}

func (resp *GetBucketResultV2) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

type GetBucketResultV1 struct {
	XMLName        xml.Name             `xml:"ListBucketResult"`
	Xmlns          string               `xml:"xmlns,attr"`
	Name           string               `xml:"Name"`
	Prefix         string               `xml:"Prefix"`
	Delimiter      string               `xml:"Delimiter"`
	MaxKeys        int                  `xml:"MaxKeys"`
	Marker         string               `xml:"Marker"`
	NextMarker     string               `xml:"NextMarker"`
	IsTruncated    bool                 `xml:"IsTruncated"`
	Contents       []*db.ListObjectItem `xml:"Contents"`
	CommonPrefixes []*Prefix            `xml:"CommonPrefixes"`
}

func (resp *GetBucketResultV1) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	body = []byte(xml.Header + string(body))
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapGetBucketResponseV2(bucket string, items []*db.ListObjectItem, prefixes []*Prefix, queries *queryParams, truncated bool, continuation string) handler.S3Responser {
	return &GetBucketResultV2{
		Name:                  bucket,
		Prefix:                queries.prefix,
		Delimiter:             queries.delimiter,
		KeyCount:              len(items) + len(prefixes),
		MaxKeys:               queries.maxKeys,
		IsTruncated:           truncated,
		NextContinuationToken: continuation,
		StartAfter:            queries.startAfter,
		Contents:              items,
		CommonPrefixes:        prefixes,
	}
}

func wrapGetBucketResponseV1(bucket string, items []*db.ListObjectItem, prefixes []*Prefix, queries *queryParams, truncated bool, nextMarker string) handler.S3Responser {
	return &GetBucketResultV1{
		Name:           bucket,
		Prefix:         queries.prefix,
		Delimiter:      queries.delimiter,
		MaxKeys:        queries.maxKeys,
		IsTruncated:    truncated,
		Marker:         queries.marker,
		NextMarker:     nextMarker,
		Contents:       items,
		CommonPrefixes: prefixes,
		Xmlns:          "http://s3.amazonaws.com/doc/2006-03-01/",
	}
}

type queryParams struct {
	prefix       string
	delimiter    string
	maxKeys      int
	marker       string
	listType     string
	startAfter   string
	continuation string
}

func parseQueryString(r *http.Request) (*queryParams, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	listType := r.Form.Get("list-type")
	if listType == "" {
		listType = "1"
	}
	if listType != "1" && listType != "2" {
		err := fmt.Errorf("unsupported get bucket list type %s", listType)
		return nil, err
	}

	max := r.Form.Get("max-keys")
	if max == "" {
		max = "100"
	}

	maxI, err := maxKeyValid(max)
	if err != nil {
		return nil, err
	}

	var marker string
	var startAfter string
	var nextContinuation string
	if listType == "1" {
		marker = r.Form.Get("marker")
	} else {
		// 根据S3 API定义:start-after一般只用于第一次请求
		// 如果响应被截断,后续的请求需要携带continuation-token, start-after将被视为无效
		startAfter = r.Form.Get("start-after")
		nextContinuation = r.Form.Get("continuation-token")
		marker = nextContinuation
		if marker == "" {
			marker = startAfter
		}
	}
	return &queryParams{
		prefix:       r.Form.Get("prefix"),
		delimiter:    r.Form.Get("delimiter"),
		marker:       marker,
		maxKeys:      maxI,
		listType:     listType,
		continuation: nextContinuation,
		startAfter:   startAfter,
	}, nil
}

func GetBucketHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		context.Set(r, "response", resp)
		//resp.Send(w)
	}()

	var queries *queryParams
	queries, err = parseQueryString(r)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidArgument", "/")
		return
	}

	bucket := mux.Vars(r)["bucket"]

	var items []*db.ListObjectItem
	items, err = db.ActiveService().ListObjectsInBucket(bucket, queries.prefix, queries.delimiter, queries.marker, queries.maxKeys+3)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}
	truncated, continuation := false, queries.continuation
	if len(items) > queries.maxKeys {
		truncated = true
		continuation = items[queries.maxKeys].Key
		items = items[:queries.maxKeys]
	}

	commonPrefixes := make([]*Prefix, 0)
	objects := make([]*db.ListObjectItem, 0)
	for _, item := range items {
		// common prefixes process
		if queries.delimiter != "" && strings.HasSuffix(item.Key, queries.delimiter) && queries.prefix != item.Key {
			commonPrefixes = append(commonPrefixes, &Prefix{item.Key})
		} else {
			objects = append(objects, item)
		}
	}

	if queries.listType == "1" {
		resp = wrapGetBucketResponseV1(bucket, objects, commonPrefixes, queries, truncated, continuation)
	} else {
		resp = wrapGetBucketResponseV2(bucket, objects, commonPrefixes, queries, truncated, continuation)
	}
	return
}

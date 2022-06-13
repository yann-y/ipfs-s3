package bucket

import (
	"encoding/xml"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"
	"net/http"
	"strconv"
)

// func maxKeyValid(smax string) (int, error) {
// 	imax, err := strconv.Atoi(smax)
//         if err != nil {
// 		return -1, err
// 	}
//
// 	if imax < 1 || imax > 1000 {
// 		return -1, errors.New("max-keys invalid, not in range 1~1000")
//         }
// 	return imax, nil
// }
//
// type Prefix struct {
// 	Prefix string
// }

type initiator struct {
	ID          string
	DisplayName string
}

type Upload struct {
	Key          string
	UploadId     string
	Initiator    initiator
	Owner        *db.User
	StorageClass string
	Initiated    string
}

type ListMultipartUploadsResult struct {
	XMLName            xml.Name `xml:"ListMultipartUploadsResult"`
	Bucket             string   `xml:"Bucket"`
	KeyMarker          string   `xml:"KeyMarker"`
	UploadIdMarker     string   `xml:"UploadIdMarker"`
	NextKeyMarker      string   `xml:"NextKeyMarker"`
	NextUploadIdMarker string   `xml:"NextUploadIdMarker"`
	MaxUploads         int      `xml:"MaxUploads"`
	IsTruncated        bool     `xml:"IsTruncated"`
	Uploads            []Upload `xml:"Upload"`
}

func (res *ListMultipartUploadsResult) ContentType() string {
	return "application/xml"
}

func (res *ListMultipartUploadsResult) ContentBody() ([]byte, error) {
	body, err := handler.FormatS3Response(res, res.ContentType())
	return body, err
}

func (resp *ListMultipartUploadsResult) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapListMultipartUploadsResponse(bucket string) handler.S3Responser {
	owner := &db.User{
		ID:          "75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a",
		DisplayName: "dingkai",
	}
	initor := initiator{
		ID:          "75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a",
		DisplayName: "dingkai",
	}
	upload := Upload{
		Key:          "my-divisor",
		UploadId:     "XMgbGlrZSBlbHZpbmcncyBub3QgaGF2aW5nIG11Y2ggbHVjaw",
		Initiator:    initor,
		Owner:        owner,
		StorageClass: "STANDARD",
		Initiated:    "2010-11-10T20:48:33.000Z",
	}
	uploads := make([]Upload, 0)
	uploads = append(uploads, upload)
	return &ListMultipartUploadsResult{
		Bucket:             bucket,
		KeyMarker:          "hello",
		UploadIdMarker:     "hello",
		NextKeyMarker:      "hello",
		NextUploadIdMarker: "hello",
		MaxUploads:         10,
		IsTruncated:        false,
		Uploads:            uploads,
	}
}

type listuploadsQueryParams struct {
	prefix       string
	delimiter    string
	maxKeys      int
	marker       string
	listType     string
	startAfter   string
	continuation string
}

// func parseQueryString(r *http.Request) (*queryParams, error) {
// 	if err := r.ParseForm(); err != nil {
// 		return nil, err
// 	}
//
// 	listType := r.Form.Get("list-type")
// 	if listType == "" {
// 		listType = "1"
// 	}
// 	if listType != "1" && listType != "2" {
// 		err := fmt.Errorf("unsupported get bucket list type %s", listType)
// 		return nil, err
// 	}
//
// 	max := r.Form.Get("max-keys")
// 	if max == "" {
// 		max = "100"
// 	}
//
// 	maxI, err := maxKeyValid(max)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var marker string
// 	var startAfter string
// 	var nextContinuation string
// 	if listType == "1" {
// 		marker = r.Form.Get("marker")
// 	} else {
// 		// 根据S3 API定义:start-after一般只用于第一次请求
// 		// 如果响应被截断,后续的请求需要携带continuation-token, start-after将被视为无效
// 		startAfter = r.Form.Get("start-after")
// 		nextContinuation = r.Form.Get("continuation-token")
// 		marker = nextContinuation
// 		if marker == "" {
// 			marker = startAfter
// 		}
// 	}
// 	return &queryParams{
// 		prefix: r.Form.Get("prefix"),
// 		delimiter: r.Form.Get("delimiter"),
// 		marker: marker,
// 		maxKeys: maxI,
// 		listType: listType,
// 		continuation: nextContinuation,
// 		startAfter: startAfter,
// 	}, nil
// }

func GetBucketUploadsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		context.Set(r, "response", resp)
	}()

	// var queries *queryParams
	// queries, err = parseQueryString(r)
	// if err != nil {
	// 	resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "InvalidArgument", "/")
	// 	return
	// }

	bucket := mux.Vars(r)["bucket"]

	// var items []*db.ListObjectItem
	// items, err = db.ActiveService().ListObjectsInBucket(bucket, queries.prefix, queries.delimiter, queries.marker, queries.maxKeys + 3)
	// if err != nil {
	// 	if err == db.BucketNotExistError {
	// 		resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket)
	// 	} else {
	// 		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
	// 	}
	// 	return
	// }
	// truncated, continuation := false, queries.continuation
	// if len(items) > queries.maxKeys {
	// 	truncated = true
	// 	continuation = items[queries.maxKeys].Key
	// 	items = items[:queries.maxKeys]
	// }

	// commonPrefixes := make([]*Prefix, 0)
	// objects := make([]*db.ListObjectItem, 0)
	// for _, item := range items {
	// 	// common prefixes process
	// 	if queries.delimiter != "" && strings.HasSuffix(item.Key, queries.delimiter) && queries.prefix != item.Key {
	// 		commonPrefixes = append(commonPrefixes, &Prefix{item.Key})
	// 	} else {
	// 		objects = append(objects, item)
	// 	}
	// }

	resp = wrapListMultipartUploadsResponse(bucket)
	return
}

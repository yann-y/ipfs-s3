package bucket

import (
	"encoding/xml"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"

	// "github.com/yann-y/ipfs-s3/mongodb/dao"
	// "github.com/yann-y/ipfs-s3/mongodb/bean"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"net/http"
	"strconv"
	"time"
)

// type User struct {
//	ID string
//	DisplayName string
//}

type Bucket struct {
	XMLName      xml.Name `xml:"Bucket"`
	Name         string
	CreationDate string
}

type ListAllMyBucketsResult struct {
	XMLName xml.Name  `xml:"ListAllMyBucketsResult"`
	Owner   *db.User  `xml:"Owner"`
	Buckets []*Bucket `xml:"Buckets"`
}

func (res *ListAllMyBucketsResult) ContentType() string {
	return "application/xml"
}

func (res *ListAllMyBucketsResult) ContentBody() ([]byte, error) {
	return handler.FormatS3Response(res, res.ContentType())
}

func (resp *ListAllMyBucketsResult) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapListAllMyBucketsResponse(me *db.User, buckets []*Bucket) handler.S3Responser {
	return &ListAllMyBucketsResult{
		Owner:   me,
		Buckets: buckets,
	}
}

func ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		// resp.Send(w)
		context.Set(r, "response", resp)
	}()

	var items []*db.Bucket
	items, err = db.ActiveService().ListUserBuckets("1")
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/")
		return
	}
	buckets := make([]*Bucket, 0)
	for _, item := range items {
		bucket := &Bucket{
			Name:         item.BucketName,
			CreationDate: time.Unix(item.CreateTime, 0).Format(time.RFC3339),
		}
		buckets = append(buckets, bucket)
	}
	me := &db.User{
		ID:          "12345",
		DisplayName: "fake user",
	}
	resp = wrapListAllMyBucketsResponse(me, buckets)
	return
}

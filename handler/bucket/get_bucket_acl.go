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

type Grantee struct {
	Namespace string   `xml:"xmlns:xsi,attr"`
	Type      string   `xml:"xsi:type,attr"`
	URI       string   `xml:"URI,omitempty"`
	User      *db.User `xml:"Owner"`
}

type Grant struct {
	Grantee    *Grantee `xml:"Grantee"`
	Permission string   `xml:"Permission"`
}

type ControlList struct {
	Grants []*Grant `xml:"Grant"`
}

type BucketACLResult struct {
	XMLName      xml.Name       `xml:"AccessControlPolicy"`
	Owner        *db.User       `xml:"Owner"`
	ControlLists []*ControlList `xml:"AccessControlList"`
}

func (resp *BucketACLResult) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapS3BucketACLResponse(acl string) *BucketACLResult {
	owner := &db.User{
		ID:          "12345",
		DisplayName: "fake user",
	}
	controList := &ControlList{
		Grants: make([]*Grant, 0),
	}
	grantee := &Grantee{
		Namespace: "http://www.w3.org/2001/XMLSchema-instance",
		Type:      "CanonicalUser",
		User:      owner,
	}
	controList.Grants = append(controList.Grants, &Grant{grantee, acl})
	return &BucketACLResult{
		Owner:        owner,
		ControlLists: []*ControlList{controList},
	}
}

// GetBucketACLHandler返回假的权限控制:FULL_CONTROL
func GetBucketACLHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		context.Set(r, "response", resp)
		// resp.Send(w)
	}()

	bucket := mux.Vars(r)["bucket"]

	_, err = db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}

	resp = wrapS3BucketACLResponse("FULL_CONTROL")
	return
}

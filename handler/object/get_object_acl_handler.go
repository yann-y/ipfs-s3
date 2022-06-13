package object

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

type ObjectACLResult struct {
	etag         string
	XMLName      xml.Name       `xml:"AccessControlPolicy"`
	Owner        *db.User       `xml:"Owner"`
	ControlLists []*ControlList `xml:"AccessControlList"`
}

func (resp *ObjectACLResult) Send(w http.ResponseWriter) {
	body, _ := xml.MarshalIndent(resp, "", " ")
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Header().Set("ETag", resp.etag)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func wrapS3ObjectACLResponse(acl, etag string) *ObjectACLResult {
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
	return &ObjectACLResult{
		etag:         etag,
		Owner:        owner,
		ControlLists: []*ControlList{controList},
	}
}

// GetObjectACLHandler返回假的权限控制:FULL_CONTROL
func GetObjectACLHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var resp handler.S3Responser

	objectName := mux.Vars(r)["object"]
	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		}
		// resp.Send(w)
		context.Set(r, "response", resp)
	}()

	bucket := mux.Vars(r)["bucket"]

	objectPath := bucket + "/" + objectName

	var bucketObject *db.Bucket
	bucketObject, err = db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", objectPath)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", objectPath)
		}
		return
	}

	var object *db.Object
	object, err = db.ActiveService().GetObject(bucket, objectName, "", bucketObject.VersionEnabled)
	if err != nil {
		if err == db.ObjectNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchKey", objectPath)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", objectPath)
		}
		return
	}
	resp = wrapS3ObjectACLResponse("FULL_CONTROL", object.Etag)
	return
}

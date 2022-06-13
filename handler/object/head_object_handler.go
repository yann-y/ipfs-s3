package object

import (
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/mux"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HeadObjectHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	status := http.StatusOK
	var size int64
	var lastModified string
	var contentType string
	var metaUid string
	var metaGid string
	var metaMode string
	var metaMtime string

	objectName := mux.Vars(r)["object"]

	defer func() {
		if err != nil {
			context.Set(r, "req_error", gerror.NewGError(err))
		} else {
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Last-Modified", lastModified)
			w.Header().Set("x-amz-meta-uid", metaUid)
			w.Header().Set("x-amz-meta-gid", metaGid)
			w.Header().Set("x-amz-meta-mode", metaMode)
			w.Header().Set("x-amz-meta-mtime", metaMtime)
		}
		w.WriteHeader(status)
	}()

	bucket := mux.Vars(r)["bucket"]

	var bucketObject *db.Bucket
	bucketObject, err = db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			status = http.StatusNotFound
		} else {
			status = http.StatusInternalServerError
		}
		return
	}
	// objectPath := bucket + "/" + objectName
	var object *db.Object
	object, err = db.ActiveService().GetObject(bucket, objectName, "", bucketObject.VersionEnabled)
	if err != nil {
		if err == db.ObjectNotExistError {
			status = http.StatusNotFound
		} else {
			status = http.StatusInternalServerError
		}
		return
	}

	if object.DeleteMarker {
		status = http.StatusNotFound
		err = db.ObjectNotExistError
		return
	}

	// is directory
	if strings.HasSuffix(objectName, "/") {
		contentType = "application/x-directory"
	}
	metaMtime = strconv.FormatInt(object.LastModified, 10)
	metaUid = "0"
	metaGid = "0"
	metaMode = "493"
	size = object.Size
	lastModified = time.Unix(object.LastModified, 0).Format(time.RFC3339)
}

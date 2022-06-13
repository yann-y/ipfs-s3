package bucket

import (
	"encoding/xml"
	"github.com/yann-y/ipfs-s3/context"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/gerror"
	"github.com/yann-y/ipfs-s3/handler"
	"github.com/yann-y/ipfs-s3/mux"
	"io/ioutil"
	"net/http"
)

type BucketVersionConfig struct {
	Status    string
	MfaDelete string
}

func parseVersionConfig(input []byte) (*BucketVersionConfig, error) {
	var c BucketVersionConfig
	err := xml.Unmarshal(input, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func PutBucketVersionHandler(w http.ResponseWriter, r *http.Request) {
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		return
	}

	versionConfig, err := parseVersionConfig(body)
	if err != nil {
		resp = handler.WrapS3ErrorResponseForRequest(http.StatusBadRequest, r, "BadRequest", "/"+bucket)
		return
	}

	versionEnabled := false
	if versionConfig.Status == "Enabled" {
		versionEnabled = true
	} else if versionConfig.Status == "Suspended" {
		versionEnabled = false
	}

	bucketBean, err := db.ActiveService().GetBucket(bucket)
	if err != nil {
		if err == db.BucketNotExistError {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "NoSuchBucket", "/"+bucket)
		} else {
			resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
		}
		return
	}
	if bucketBean.VersionEnabled != versionEnabled {
		bucketBean.VersionEnabled = versionEnabled
		err = db.ActiveService().UpdateBucket(bucketBean)
		if err != nil {
			if err == db.BucketNotExistError {
				resp = handler.WrapS3ErrorResponseForRequest(http.StatusNotFound, r, "BucketNotExists", "/"+bucket)
			} else {
				resp = handler.WrapS3ErrorResponseForRequest(http.StatusInternalServerError, r, "InternalError", "/"+bucket)
			}
			return
		}
	}
	resp = handler.NewS3NilResponse(http.StatusOK)
	return
}

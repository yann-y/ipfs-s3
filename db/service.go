package db

import (
	"errors"
)

var (
	TooManyBucketError = errors.New("too many buckets")
	BucketExistError = errors.New("bucket already exist")
	BucketNotEmptyError = errors.New("bucket not empty")
	BucketNotExistError = errors.New("bucket not exist")

	ObjectNotExistError = errors.New("object not exist")
	ObjectDeletedError  = errors.New("object has been deleted")

	UploadNotExistError = errors.New("upload not exist")
	UploadPartNotExistError = errors.New("upload part not exist")
)

type ObjectService interface {
	GetObject(bucket, object, version string, versionEnabled bool) (*Object, error)
	PutObject(object *Object, versionEnabled bool) error
	DeleteObject(bucket, object, version string, versionEnabled bool) error
	PutObjectFromMultipartUpload(object *Object, versionEnabled bool) error
}

type User struct {
     ID string
     DisplayName string
}

type ListObjectItem struct {
	Key  string
	ETag string
	Size int64
	LastModified string
	StorageClass string
	Owner        *User
}

type ListObjectItems []*ListObjectItem
func (items ListObjectItems) Len() int { return len(items) }
func (items ListObjectItems) Less(i, j int) bool { return items[i].Key < items[j].Key }
func (items ListObjectItems) Swap(i, j int) { items[i], items[j] = items[j], items[i] }

type BucketService interface {
	GetBucket(name string) (*Bucket, error)
	PutBucket(bucket *Bucket) error
	DeleteBucket(name string) error
	UpdateBucket(bucket *Bucket) error
	ListUserBuckets(uid string) ([]*Bucket, error)
	ListObjectsInBucket(bucketName, prefix, delimiter, marker string, max int) ([]*ListObjectItem, error)
}

type UploadPartService interface {
	InitMultipartUpload(uploadInfo *UploadInfo) error
	GetUpload(uploadId string) (*UploadInfo, error)
	SetUploadAborted(uploadId string) error
	ListUploadAllParts(uploadId string) ([]*UploadPart, error)
	ListUploadParts(uploadId string, marker int, max int) ([]*UploadPart, error)
	PutUploadPart(part *UploadPart) error
}

type Service interface {
	ObjectService
	BucketService
	UploadPartService
}

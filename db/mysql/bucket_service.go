package mysql

import (
	"bytes"
	"fmt"
	"github.com/yann-y/ipfs-s3/db"
	"sort"
	"strings"
	"time"
)

func (ms *mysqlService) PutBucket(bucket *db.Bucket) error {

	bucketId, err := ms.AllocateID()
	if err != nil {
		return err
	}

	tx := ms.DB.Begin()

	count := 0
	err = tx.Model(&MysqlBucket{}).Where(&MysqlBucket{Bucket: db.Bucket{UserID: bucket.UserID}}).Set("gorm:query_option", "FOR UPDATE").Count(&count).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if count >= 10 {
		tx.Rollback()
		err = db.TooManyBucketError
		return err
	}

	// note: if we use tidb, if bucket already exist, it will return error until tx.Commit
	err = tx.Create(toMysqlBucket(bucketId, bucket)).Error
	if err != nil {
		tx.Rollback()
		if KeyDuplicatedError(err) {
			return db.BucketExistError
		}
		return err
	}

	err = tx.Commit().Error
	if KeyDuplicatedError(err) {
		return db.BucketExistError
	}
	return err
}

func (ms *mysqlService) ListUserBuckets(uid string) ([]*db.Bucket, error) {

	buckets := make([]*MysqlBucket, 0)
	res := make([]*db.Bucket, 0)
	result := ms.DB.Where("user_id = ?", uid).Find(&buckets)
	for _, b := range buckets {
		res = append(res, toBucket(b))
	}
	return res, result.Error
}

func (ms *mysqlService) getBucket(name string) (*MysqlBucket, error) {

	bucket := &MysqlBucket{Bucket: db.Bucket{BucketName: name}}
	result := ms.DB.First(&bucket)
	if result.RecordNotFound() {
		return nil, db.BucketNotExistError
	}
	return bucket, ms.DB.Error
}

func (ms *mysqlService) GetBucket(name string) (*db.Bucket, error) {

	mysqlBucket, err := ms.getBucket(name)
	if err != nil {
		return nil, err
	}
	return toBucket(mysqlBucket), nil
}

func (ms *mysqlService) ListObjectsInBucket(bucketName, prefix, delimiter, marker string, max int) ([]*db.ListObjectItem, error) {

	bucket, err := ms.getBucket(bucketName)
	if err != nil {
		return nil, err
	}

	scope := ms.DB.Raw("call list_objects(?,?,?,?,?)", bucket.BucketID, prefix, delimiter, max, marker)
	if scope.Error != nil {
		return nil, scope.Error
	}
	rows, err := scope.Rows()
	if err != nil {
		return nil, err
	}

	object_names := make([]string, 0, 16)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		object_names = append(object_names, name)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	objects := make([]*db.ListObjectItem, 0, 16)
	var sqlBuff bytes.Buffer
	sqlBuff.WriteString("select object_name, etag, size, last_modified from mysql_objects where ")
	firstNormalObject := true
	for _, name := range object_names {
		if delimiter != "" && strings.HasSuffix(name, delimiter) && prefix != name {
			item := &db.ListObjectItem{
				Key: name,
			}
			objects = append(objects, item)
		} else {
			if !firstNormalObject {
				sqlBuff.WriteString(" or ")
			}
			firstNormalObject = false
			sqlBuff.WriteString(fmt.Sprintf("object_name = '%s'", name))
		}
	}
	if len(object_names) > len(objects) {
		scope = ms.DB.Raw(sqlBuff.String())
		if scope.Error != nil {
			return nil, scope.Error
		}
		rows, err := scope.Rows()
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			item := db.ListObjectItem{StorageClass: "STANDARD"}
			var lastModified int64
			err := rows.Scan(&item.Key, &item.ETag, &item.Size, &lastModified)
			if err != nil {
				return nil, err
			}
			item.LastModified = time.Unix(lastModified, 0).Format(time.RFC3339)
			objects = append(objects, &item)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		sort.Sort(db.ListObjectItems(objects))
	}
	return objects, nil
}

func (ms *mysqlService) DeleteBucket(name string) error {

	tx := ms.DB.Begin()

	bucket := MysqlBucket{Bucket: db.Bucket{BucketName: name}}
	var err error

	// bucket exist?
	result := tx.First(&bucket)
	if result.RecordNotFound() {
		tx.Rollback()
		err = db.BucketNotExistError
		return err
	}

	// Objects exist?
	result = tx.Where("bucket_id = ?", bucket.BucketID).First(&MysqlObject{})
	if !result.RecordNotFound() {
		tx.Rollback()
		err = db.BucketNotEmptyError
		return err
	}

	// UploadInfo exist?
	result = tx.Where("bucket_id = ?", bucket.BucketID).First(&MysqlUploadInfo{})
	if !result.RecordNotFound() {
		tx.Rollback()
		err = db.BucketNotEmptyError
		return err
	}

	// Delete Bucket
	err = tx.Delete(&bucket).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (ms *mysqlService) UpdateBucket(bkt *db.Bucket) error {
	return nil
}

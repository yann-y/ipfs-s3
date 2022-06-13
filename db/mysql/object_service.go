package mysql

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/yann-y/ipfs-s3/db"
	"github.com/yann-y/ipfs-s3/utils/md5"
	"time"
)

type Action int

const (
	INVALID Action = iota
	INSERT
	DUPLICATE_KEY
	UPDATE
	DONE
)

func (ms *mysqlService) GetObject(bucket, object, version string, versionEnabled bool) (*db.Object, error) {

	keyMD5sum := md5.Sum(object)
	keyMD5High := md5.MD5High(keyMD5sum)
	keyMD5Low := md5.MD5Low(keyMD5sum)

	mysqlObject := &MysqlObject{
		KeyMd5High: keyMD5High,
		KeyMd5Low:  keyMD5Low,
		Object:     db.Object{Bucket: bucket},
	}

	result := ms.DB.First(mysqlObject)
	if result.RecordNotFound() {
		return nil, db.ObjectNotExistError
	}

	return toObject(mysqlObject), result.Error
}

func (ms *mysqlService) putObjectInTx(object *MysqlObject, tx *gorm.DB) (*gorm.DB, error) {
	action := INSERT
	var err error
	for {
		switch action {
		case INSERT:
			action, err = ms.insertMeta(tx, object)
			if err != nil {
				return tx, err
			}
		// TODO: may have bugs
		case DUPLICATE_KEY:
			err = tx.Commit().Error
			if err != nil {
				return tx, err
			}
			tx = ms.DB.Begin()
			if tx.Error != nil {
				return tx, tx.Error
			}
			object.ConflictFlag++
			action = INSERT
		case UPDATE:
			action, err = ms.updateMeta(tx, object)
			if err != nil {
				return tx, err
			}
		case DONE:
			return tx, nil
		default:
			return tx, errors.New("invalid database action")
		}
	}
}

func (ms *mysqlService) PutObjectFromMultipartUpload(object *db.Object, versionEnable bool) error {

	mysqlBucket, err := ms.getBucket(object.Bucket)
	if err != nil {
		return err
	}

	id, err := ms.AllocateID()
	if err != nil {
		return err
	}

	mysqlObject := toMysqlObject(id, mysqlBucket.BucketID, object)

	tx := ms.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	tx, err = ms.putObjectInTx(mysqlObject, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除所有的UploadPart
	if err = tx.Where("upload_id = ?", object.UploadId).Delete(db.UploadPart{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除UploadInfo
	if err = tx.Where("upload_id = ?", object.UploadId).Delete(MysqlUploadInfo{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (ms *mysqlService) PutObject(object *db.Object, versionEnable bool) error {

	mysqlBucket, err := ms.getBucket(object.Bucket)
	if err != nil {
		return err
	}

	id, err := ms.AllocateID()
	if err != nil {
		return err
	}

	mysqlObject := toMysqlObject(id, mysqlBucket.BucketID, object)

	tx := ms.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	tx, err = ms.putObjectInTx(mysqlObject, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (ms *mysqlService) insertMeta(tx *gorm.DB, meta *MysqlObject) (Action, error) {

	index := &ObjectList{
		BucketID:   meta.BucketID,
		ObjectName: meta.ObjectName,
	}
	err := tx.Create(index).Error
	if err != nil && !KeyDuplicatedError(err) {
		return INVALID, err
	}

	err = tx.Create(meta).Error
	if err != nil {
		// if object already exist
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				target := &MysqlObject{
					KeyMd5High:   meta.KeyMd5High,
					KeyMd5Low:    meta.KeyMd5Low,
					ConflictFlag: meta.ConflictFlag,
				}
				existing := tx.First(target)
				if existing.Error != nil {
					return INVALID, existing.Error
				}
				if existing.RecordNotFound() {
					return INSERT, nil
				}
				if target.BucketID == meta.BucketID && target.ObjectName == meta.ObjectName {
					return UPDATE, nil
				}
				return DUPLICATE_KEY, nil
			}
		} else {
			return INVALID, err
		}
	}
	return DONE, nil
}

func (ms *mysqlService) updateMeta(tx *gorm.DB, meta *MysqlObject) (Action, error) {

	existingObj := &MysqlObject{
		KeyMd5High:   meta.KeyMd5High,
		KeyMd5Low:    meta.KeyMd5Low,
		ConflictFlag: meta.ConflictFlag,
	}
	err := tx.First(existingObj).Error

	if err != nil && tx.RecordNotFound() {
		return INSERT, nil
	}
	if err != nil {
		return INVALID, tx.Error
	}

	historyObj := &HistoryObject{
		ObjectID:   existingObj.ObjectID,
		KeyMd5High: existingObj.KeyMd5High,
		KeyMd5Low:  existingObj.KeyMd5Low,
		// Md5High:       existingObj.Md5High,
		// Md5Low:        existingObj.Md5Low,
		Object: db.Object{
			Fid:      existingObj.Fid,
			Meta:     existingObj.Meta,
			Size:     existingObj.Size,
			Etag:     existingObj.Etag,
			Bucket:   existingObj.Bucket,
			PartSize: existingObj.PartSize,
			// mysql存储暂不支持version
			Version:         "",
			DeleteMarker:    false,
			ObjectName:      existingObj.ObjectName,
			LastModified:    existingObj.LastModified,
			MultipartUpload: existingObj.MultipartUpload,
		},
		// TODO: 去掉BucketID
		BucketID:      existingObj.BucketID,
		HistoryTime:   time.Now().Unix(),
		DigestVersion: existingObj.DigestVersion,
		DeleteHint:    2,
	}
	err = tx.Create(historyObj).Error
	if err != nil {
		return INVALID, err
	}

	// 为什么重新再生成一个?
	//newMeta := &MysqlObject{
	//	ObjectName:    meta.ObjectName,
	//	BucketID:      meta.BucketID,
	//	Fid:           meta.Fid,
	//	ObjectID:      meta.ObjectID,
	//	Size:          meta.Size,
	//	Md5High:       meta.Md5High,
	//	Md5Low:        meta.Md5Low,
	//	KeyMd5High:    meta.KeyMd5High,
	//	KeyMd5Low:     meta.KeyMd5Low,
	//	ConflictFlag:  meta.ConflictFlag,
	//	LastModified:  meta.LastModified,
	//	DigestVersion: meta.DigestVersion,
	//}
	err = tx.Save(meta).Error
	if err != nil {
		return INVALID, err
	}
	return DONE, nil
}

func (ms *mysqlService) DeleteObject(bucket, object, version string, versionEnabled bool) error {

	// fullPath := bucket + "/" + object
	keyMD5sum := md5.Sum(object)
	keyMD5High := md5.MD5High(keyMD5sum)
	keyMD5Low := md5.MD5Low(keyMD5sum)

	var err error
	tx := ms.DB.Begin()

	obj := &MysqlObject{
		KeyMd5High: keyMD5High,
		KeyMd5Low:  keyMD5Low,
		Object:     db.Object{Bucket: bucket},
	}

	result := tx.Set("gorm:query_option", "FOR UPDATE").First(obj)
	if result.RecordNotFound() {
		tx.Rollback()
		return db.ObjectNotExistError
	}
	if err = result.Error; err != nil {
		tx.Rollback()
		return err
	}

	historyObj := &HistoryObject{
		ObjectID:   obj.ObjectID,
		KeyMd5High: obj.KeyMd5High,
		KeyMd5Low:  obj.KeyMd5Low,
		// Md5High:       obj.Md5High,
		// Md5Low:        obj.Md5Low,
		Object: db.Object{
			Fid:      obj.Fid,
			Meta:     obj.Meta,
			Size:     obj.Size,
			Etag:     obj.Etag,
			Bucket:   obj.Bucket,
			PartSize: obj.PartSize,
			// mysql存储暂不支持version
			Version:         "",
			DeleteMarker:    false,
			ObjectName:      obj.ObjectName,
			LastModified:    obj.LastModified,
			MultipartUpload: obj.MultipartUpload,
		},
		HistoryTime:   time.Now().Unix(),
		DeleteHint:    1,
		DigestVersion: obj.DigestVersion,
	}

	if err = tx.Create(historyObj).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Delete(obj).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Delete(&ObjectList{
		BucketID:   obj.BucketID,
		ObjectName: obj.ObjectName,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

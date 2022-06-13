package mysql

import (
	"database/sql"
	"github.com/yann-y/ipfs-s3/db"
)

func (ms *mysqlService) InitMultipartUpload(uploadInfo *db.UploadInfo) error {
	bucket, err := ms.getBucket(uploadInfo.Bucket)
	if err != nil {
		return err
	}

	return ms.DB.Create(toMysqlUpload(bucket.BucketID, uploadInfo)).Error
}

func (ms *mysqlService) GetUpload(uploadId string) (*db.UploadInfo, error) {
	upload := &MysqlUploadInfo{
		UploadInfo: db.UploadInfo{UploadID: uploadId},
	}
	result := ms.DB.First(upload)
	if result.RecordNotFound() {
		return nil, db.UploadNotExistError
	}
	return toUpload(upload), result.Error
}

func (ms *mysqlService) SetUploadAborted(uploadId string) error {
	upload := &MysqlUploadInfo{
		UploadInfo: db.UploadInfo{UploadID: uploadId},
	}
	res := ms.DB.Model(upload).Update("is_abort", 1)
	if res.Error != nil {
		return res.Error
	}
	if res.RecordNotFound() {
		return db.UploadNotExistError
	}
	return nil
}

func (ms *mysqlService) ListUploadAllParts(uploadId string) ([]*db.UploadPart, error) {
	scope := ms.DB.Raw("select fid, number, size, last_modified, etag from upload_parts where upload_id = ? order by number", uploadId)
	if scope.Error != nil {
		return nil, scope.Error
	}

	rows, err := scope.Rows()
	if err != nil {
		return nil, err
	}

	parts := make([]*db.UploadPart, 0)
	for rows.Next() {
		var part db.UploadPart
		if err := rows.Scan(&part.Fid, &part.Number, &part.Size, &part.LastModified, &part.Etag); err != nil {
			return nil, err
		}
		parts = append(parts, &part)
	}
	return parts, nil
}

func (ms *mysqlService) ListUploadParts(uploadId string, marker int, max int) ([]*db.UploadPart, error) {

	scope := ms.DB.Raw("select number, size, last_modified, etag from upload_parts where upload_id = ? and number > ? order by number limit ?", uploadId, marker, max)
	if scope.Error != nil {
		return nil, scope.Error
	}

	rows, err := scope.Rows()
	if err != nil {
		return nil, err
	}

	parts := make([]*db.UploadPart, 0)
	for rows.Next() {
		var part db.UploadPart
		if err := rows.Scan(&part.Number, &part.Size, &part.LastModified, &part.Etag); err != nil {
			return nil, err
		}
		parts = append(parts, &part)
	}
	return parts, nil
}

func (ms *mysqlService) GetLastUploadPart(uploadId string) (int, error) {
	scope := ms.DB.Raw("select max(number) from upload_parts where upload_id = ? order by number", uploadId)
	if scope.Error != nil {
		return -1, scope.Error
	}

	rows, err := scope.Rows()
	if err != nil {
		return -1, err
	}

	for rows.Next() {
		var number sql.NullInt64
		if err := rows.Scan(&number); err != nil {
			return -1, err
		}
		if number.Valid {
			return int(number.Int64), nil
		}
		return 0, nil
	}
	return 0, nil
}

func (ms *mysqlService) PutUploadPart(part *db.UploadPart) error {
	return ms.DB.Create(part).Error
}

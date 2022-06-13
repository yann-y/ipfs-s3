package db

type Bucket struct {
	BucketName     string `bson:"_id"     gorm:"type:varchar(64);not null;primary_key"`
	UserID         string `bson:"user_id" gorm:"type:varchar(64);not null;index:IDX_Owner"`
	ACL            int8   `bson:"acl"     gorm:"type:tinyint(3);unsigned;not null"`
	CreateTime     int64  `bson:"create_time"     gorm:"type:bigint(20);unsigned;not null"`
	VersionEnabled bool   `bson:"version_enabled" gorm:"type:tinyint(1);unsigned;not null"`
}

type ObjectVersion struct {
	ObjectName string `bson:"_id"`
	Version    string `bson:"version"`
}

type Object struct {
	Etag            string `bson:"etag"   gorm:"type:varchar(64);NOT NULL"`
	Bucket          string `bson:"bucket" gorm:"type:varchar(64);NOT NULL"`
	Size            int64  `bson:"size"   gorm:"type:bigint(20);unsigned;NOT NULL"`
	LastModified    int64  `bson:"last_modified" gorm:"type:bigint(20);unsigned;NOT NULL;index:IDX_BucketID_LastModified"`
	ObjectName      string `bson:"object_name"   gorm:"type:varchar(512);NOT NULL"`
	Fid             string `bson:"fid"           gorm:"type:varchar(256)"`
	Meta            string `bson:"meta"          gorm:"type:varchar(512);NOT NULL"`
	MultipartUpload bool   `bson:"is_multipart_upload" gorm:"type:boolean"`
	UploadId        string `bson:"upload_id"           gorm:"type:varchar(64)"`
	PartSize        int64  `bson:"part_size"           gorm:"type:bigint"`
	Version         string `bson:"version"             gorm:"type:varchar(64)"`
	DeleteMarker    bool   `bson:"delete_marker"       gorm:"type:boolean"`
}

type UploadInfo struct {
	UploadID   string `bson:"_id" gorm:"type:varchar(64);not null;primary_key"`
	StartTime  int64  `bson:"start_time" gorm:"type:bigint(20);unsigned;NOT NULL"`
	Bucket     string `bson:"bucket" gorm:"type:varchar(64);NOT NULL"`
	Object     string `bson:"object" gorm:"type:varchar(512);NOT NULL"`
	UserID     string `bson:"user"   gorm:"type:varchar(64);not null;index:IDX_Owner"`
	IsAbort    bool   `bson:"is_abort" gorm:"type:boolean;unsigned;not null"`
	Meta       string `bson:"meta" gorm:"type:varchar(512);NOT NULL"`
}

type UploadPart struct {
	UploadID     string `bson:"upload_id" gorm:"type:varchar(64);not null;primary_key"`
	Fid          string `bson:"fid" gorm:"type:varchar(64);not null"`
	Number       int    `bson:"number" gorm:"type:bigint(20);NOT NULL;primary_key"`
	Size         int64  `bson:"size" gorm:"type:bigint(20);unsigned;NOT NULL"`
	LastModified int64  `bson:"last_modified" gorm:"type:bigint(20);unsigned;NOT NULL"`
	Etag         string `bson:"etag" gorm:"type:varchar(64);not null"`
}

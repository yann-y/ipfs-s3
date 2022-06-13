package mongodb

import (
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

func (ms *mongoService) GetObject(bucket, objectName, version string, versionEnabled bool) (*db.Object, error) {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil, err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	// 如果开启version且version未设置(为"")
	// 首先从object_versions获取object最新版本
	if versionEnabled && version == "" {
		var objectVersion db.ObjectVersion
		collection := session.DB("galaxy_s3_gateway").C("object_versions")
		err := collection.Find(bson.M{"_id": bucket + "/" + objectName}).One(&objectVersion)
		if err != nil {
			if err == mgo.ErrNotFound {
				return nil, db.ObjectNotExistError
			}
			return nil, err
		}
		version = objectVersion.Version
	}

	var object db.Object
	collection := session.DB("galaxy_s3_gateway").C("objects")
	err = collection.Find(bson.M{"bucket": bucket, "object_name": objectName, "version": version}).One(&object)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, db.ObjectNotExistError
		}
		return nil, err
	}
	return &object, nil
}

func (ms *mongoService) PutObjectFromMultipartUpload(object *db.Object, versionEnabled bool) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	// version功能开启,对象会被生成全局唯一version id
	// 同时在object_versions表中记录当前version id
	// 首先在objects表中插入对象记录
	// 接下来判断version功能开启的话,还需要插入object_versions表
	// 以记录object的当前version
	if versionEnabled {
		currVersion := strings.Replace(uuid.NewV4().String(), "-", "", -1)
		object.Version = currVersion
	} else {
		object.Version = ""
	}
	collection := session.DB("galaxy_s3_gateway").C("objects")
	if err := collection.Insert(object); err != nil {
		return err
	}
	if versionEnabled {
		objVersion := &db.ObjectVersion{
			ObjectName: object.Bucket + "/" + object.ObjectName,
			Version:    object.Version,
		}
		collection = session.DB("galaxy_s3_gateway").C("object_versions")
		if _, err := collection.UpsertId(objVersion.ObjectName, objVersion); err != nil {
			return err
		}
	}

	collection = session.DB("galaxy_s3_gateway").C("upload_parts")

	// 接下来删除upload_parts和upload_info, 即使出错了也没关系,会遗留一些垃圾数据
	// 前面保证了对象数据已经正确写入了object表
	// 但是在listparts的时候可能会出现该parts
	// 可以通过运维手段清理这些垃圾数据
	collection.RemoveAll(bson.M{"upload_id": object.UploadId})

	collection = session.DB("galaxy_s3_gateway").C("upload_infos")
	collection.RemoveId(object.UploadId)
	return nil
}

func (ms *mongoService) PutObject(object *db.Object, versionEnabled bool) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)
	// version功能开启,对象会被生成全局唯一version id
	// 同时在object_versions表中记录当前version id
	// 首先在objects表中插入对象记录
	// 接下来判断version功能开启的话,还需要插入object_versions表
	// 以记录object的当前version
	if versionEnabled {
		currVersion := strings.Replace(uuid.NewV4().String(), "-", "", -1)
		object.Version = currVersion
	} else {
		object.Version = ""
	}
	// 这里使用Upsert,因为可能会出现同名对象重复上传
	collection := session.DB("galaxy_s3_gateway").C("objects")
	_, err = collection.Upsert(bson.M{"bucket": object.Bucket, "object_name": object.ObjectName, "version": object.Version}, object)
	if err != nil {
		return err
	}
	if versionEnabled {
		objVersion := &db.ObjectVersion{
			ObjectName: object.Bucket + "/" + object.ObjectName,
			Version:    object.Version,
		}
		collection = session.DB("galaxy_s3_gateway").C("object_versions")
		if _, err := collection.UpsertId(objVersion.ObjectName, objVersion); err != nil {
			return err
		}
	}
	return nil
}

func (ms *mongoService) DeleteObject(bucket, object, version string, versionEnabled bool) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	if versionEnabled && version == "" {
		var objectVersion db.ObjectVersion
		collection := session.DB("galaxy_s3_gateway").C("object_versions")
		err := collection.Find(bson.M{"_id": bucket + "/" + object}).One(&objectVersion)
		if err != nil {
			// 如果version未找到,可能是由于此前bucket未开启versioning功能
			// 此时不返回错误,而是将version设置为""
			if err == mgo.ErrNotFound {
				version = ""
			} else {
				return err
			}
			// 代表删除最新版本的
		} else {
			version = objectVersion.Version
		}
	}

	collection := session.DB("galaxy_s3_gateway").C("objects")
	// query := collection.Find(bson.M{"object_name": object, "version": version, "delete_marker": false})
	query := collection.Find(bson.M{"bucket": bucket, "object_name": object, "version": version})
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"delete_marker": true}},
		ReturnNew: false,
	}
	var res db.Object
	_, err = query.Apply(change, &res)
	if err != nil {
		if err == mgo.ErrNotFound {
			return db.ObjectNotExistError
		}
		return err
	}

	// 如果对象已经被删除
	if res.DeleteMarker == true {
		return db.ObjectDeletedError
	}
	return nil
}

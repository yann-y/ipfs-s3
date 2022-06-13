package mongodb

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (ms *mongoService) InitMultipartUpload(uploadInfo *db.UploadInfo) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("upload_infos")
	if err := collection.Insert(uploadInfo); err != nil {
		return err
	}
	return nil
}

func (ms *mongoService) GetUpload(uploadId string) (*db.UploadInfo, error) {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil, err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("upload_infos")
	var upload db.UploadInfo
	err = collection.Find(bson.M{"_id": uploadId}).One(&upload)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, db.UploadNotExistError
		}
		return nil, err
	}
	return &upload, nil
}

func (ms *mongoService) SetUploadAborted(uploadId string) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("upload_infos")
	err = collection.UpdateId(uploadId, bson.M{"$set": bson.M{"is_abort": true}})
	if err == mgo.ErrNotFound {
		return db.BucketNotExistError
	}
	return err
}

func (ms *mongoService) ListUploadAllParts(uploadId string) ([]*db.UploadPart, error) {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil, err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("upload_parts")

	res := make([]*db.UploadPart, 0)
	err = collection.Find(bson.M{"upload_id": uploadId}).All(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ms *mongoService) ListUploadParts(uploadId string, marker int, max int) ([]*db.UploadPart, error) {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil, err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)
	collection := session.DB("galaxy_s3_gateway").C("upload_parts")

	res := make([]*db.UploadPart, 0)
	err = collection.Find(bson.M{"upload_id": uploadId, "number": bson.M{"$gt": marker}}).Sort("number").Limit(max).All(&res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ms *mongoService) GetLastUploadPart(uploadId string) (int, error) {
	//scope := db.DB.Raw("select max(number) from upload_parts where upload_id = ? order by number", uploadId)
	//if scope.Error != nil {
	//	return -1, scope.Error
	//}

	//rows, err := scope.Rows()
	//if err != nil {
	//	return -1, err
	//}

	//for rows.Next() {
	//	var number sql.NullInt64
	//	if err := rows.Scan(&number); err != nil {
	//		return -1, err
	//	}
	//	if number.Valid {
	//		return int(number.Int64), nil
	//	}
	//	return 0, nil
	//}
	//return 0, nil
	return 0, nil
}

func (ms *mongoService) PutUploadPart(part *db.UploadPart) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)
	collection := session.DB("galaxy_s3_gateway").C("upload_parts")
	if err := collection.Insert(part); err != nil {
		return err
	}
	return nil
}

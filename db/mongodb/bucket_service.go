package mongodb

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strings"
	"time"
)

func (ms *mongoService) PutBucket(bucket *db.Bucket) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("buckets")
	err = collection.Insert(bucket)
	if mgo.IsDup(err) {
		return db.BucketExistError
	}
	return err
}

func (ms *mongoService) ListUserBuckets(uid string) ([]*db.Bucket, error) {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil, err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	buckets := make([]*db.Bucket, 0)
	collection := session.DB("galaxy_s3_gateway").C("buckets")
	iter := collection.Find(bson.M{"user_id": uid}).Iter()
	err = iter.All(&buckets)
	if err != nil {
		return nil, err
	}
	return buckets, nil
}

func (ms *mongoService) GetBucket(name string) (*db.Bucket, error) {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil, err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("buckets")

	result := db.Bucket{}
	err = collection.Find(bson.M{"_id": name}).One(&result)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, db.BucketNotExistError
		} else {
			return nil, err
		}
	}
	return &result, nil
}

func (ms *mongoService) DeleteBucket(name string) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("objects")

	objCnt, err := collection.Find(bson.M{"bucket_id": name}).Limit(1).Count()
	if err != nil {
		return err
	}
	if objCnt > 0 {
		return db.BucketNotEmptyError
	}

	collection = session.DB("galaxy_s3_gateway").C("upload_info")
	uploadCnt, err := collection.Find(bson.M{"bucket_id": name}).Limit(1).Count()
	if err != nil {
		return err
	}
	if uploadCnt > 0 {
		return db.BucketNotEmptyError
	}

	collection = session.DB("galaxy_s3_gateway").C("buckets")
	err = collection.Remove(bson.M{"_id": name})
	if err != nil {
		if err == mgo.ErrNotFound {
			return db.BucketNotExistError
		}
		return nil
	}
	return nil
}

//type ListObjectItem struct {
//	Key  string
//	ETag string
//	Size int64
//	LastModified string
//	StorageClass string
//}
//
//type ListObjectItems []*ListObjectItem

//func (items ListObjectItems) Len() int { return len(items) }
//func (items ListObjectItems) Less(i, j int) bool { return items[i].Key < items[j].Key }
//func (items ListObjectItems) Swap(i, j int) { items[i], items[j] = items[j], items[i] }

func (ms *mongoService) ListObjectsInBucket(bucketName, prefix, delimiter, marker string, max int) ([]*db.ListObjectItem, error) {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil, err
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("objects")

	items := make([]*db.ListObjectItem, 0)
	var object db.Object
	prefixReg := "^" + prefix

	for len(items) < max {
		query := collection.Find(bson.M{"bucket": bucketName, "delete_marker": false, "object_name": bson.M{"$gt": marker, "$regex": bson.RegEx{prefixReg, "i"}}}).Sort("object_name").Limit(max)
		cnt, err := query.Count()
		if err != nil {
			return nil, err
		}

		// 如果所有已经被遍历完成
		if cnt == 0 {
			break
		}

		iter := query.Iter()
		for iter.Next(&object) {
			// 判断对象是否需要被折叠,例如object name为"test/a"
			// 而prefix为test, delimiter为/,则该对象需要被折叠成"test/"
			// 如果prefix为test/, delimiter为/, 则test/a就无需被折叠,但是test/a/b就需要被折叠成test/a/
			// 如果prefix为test/, delimiter为/, 对象test/该如何处理呢?,目前就将test/当做普通对象处理,不知是否妥当
			// if delimiter != "" && (!strings.HasSuffix(prefix, delimiter)) && strings.Index(object.ObjectName, delimiter) != -1 {
			if delimiter != "" && prefix != object.ObjectName && strings.Index(object.ObjectName[len(prefix):], delimiter) != -1 {
				folded := object.ObjectName[:strings.Index(object.ObjectName[len(prefix):], delimiter)+1+len(prefix)]
				foldedExist := false
				// 需要判断是否已经被折叠过,如果是,则无需再次填充
				// 例如,根据上面的定义,test/a和test/b都会被折叠成为test/
				// 显然test/不能被返回两次
				for _, e := range items {
					if e.Key == folded {
						foldedExist = true
						break
					}
				}
				if !foldedExist {
					item := db.ListObjectItem{
						Key: folded,
					}
					items = append(items, &item)
				}
			} else {
				// 如果对象存在多版本,那么对象名可能会重复
				objectExist := false
				for _, e := range items {
					if e.Key == object.ObjectName {
						objectExist = true
						break
					}
				}
				if !objectExist {
					me := &db.User{
						ID:          "12345",
						DisplayName: "fake user",
					}
					item := db.ListObjectItem{
						Key:          object.ObjectName,
						StorageClass: "STANDARD",
						ETag:         object.Etag,
						LastModified: time.Unix(object.LastModified, 0).Format(time.RFC3339),
						// LastModified: time.Unix(object.LastModified, 0).Format("Mon, 02 Jan 2006 15:04:05 GMT"),
						Size:  object.Size,
						Owner: me,
					}
					items = append(items, &item)
				}
			}
			marker = object.ObjectName
		}
		iter.Close()
	}
	sort.Sort(db.ListObjectItems(items))
	return items, nil
}

func (ms *mongoService) UpdateBucket(bucket *db.Bucket) error {
	session, err := ms.sessions.GetSession(ms.servers)
	if err != nil {
		return nil
	}
	defer ms.sessions.ReturnSession(session, SessionOK)

	collection := session.DB("galaxy_s3_gateway").C("buckets")
	err = collection.UpdateId(bucket.BucketName, bucket)
	if err == mgo.ErrNotFound {
		return db.BucketNotExistError
	}
	return err
}

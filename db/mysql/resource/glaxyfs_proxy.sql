CREATE TABLE `ID` (
  `ID` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '全局自增ID',
  `value` int unsigned NOT NULL COMMENT '',
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='全局自增ID';

CREATE TABLE `Bucket` (
  `BucketID` bigint(20) unsigned NOT NULL COMMENT '桶ID',
  `BucketName` varchar(64) NOT NULL COMMENT '桶名称',
  `KeyMD5High` bigint(20) NOT NULL COMMENT '桶名称KeyMD5High',
  `ProductID` varchar(64) NOT NULL COMMENT '拥有者',
  `ACL` tinyint(3) unsigned NOT NULL COMMENT '桶访问控制符',
  `CreateTime` bigint(20) unsigned NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`BucketName`),
  KEY `IDX_Owner` (`ProductID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='桶信息表';

CREATE TABLE `Object` (
  `ObjectID` bigint(20) unsigned NOT NULL COMMENT '对象ID',
  `KeyMD5High` bigint(20) NOT NULL COMMENT '桶名和对象名的摘要，用于快速定位和保证唯一性',
  `KeyMD5Low` bigint(20) NOT NULL COMMENT '桶名和对象名的摘要,用于快速定位和保证唯一性',
  `MD5High` bigint(20) NOT NULL COMMENT '对象MD5值',
  `MD5Low` bigint(20) NOT NULL COMMENT '对象MD5值',
  `ConflictFlag` tinyint(3) unsigned NOT NULL COMMENT '冲突计数，用于解决key的md5冲突',
  `BucketID` bigint(20) unsigned NOT NULL COMMENT '所属桶号',
  `Size` bigint(20) unsigned NOT NULL COMMENT '文件大小',
  `LastModified` bigint(20) unsigned NOT NULL COMMENT '最后修改时间',
  `ObjectName` varchar(1000) NOT NULL COMMENT '对象名称,桶内唯一',
  `DocIDList` mediumblob ,
  `DigestVersion` smallint(5) NOT NULL COMMENT '摘要版本信息,区分对象的类型',
  `Meta` varchar(4096) NOT NULL COMMENT '用户自定义响应头集合',
  PRIMARY KEY (`KeyMD5High`,`KeyMD5Low`,`ConflictFlag`),
  KEY `Idx_MD5High` (`MD5High`),
  KEY `IDX_BucketID_LastModified` (`BucketID`,`LastModified`),
  KEY `IDX_ObjectID` (ObjectID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='对象信息表';

CREATE TABLE `HistoryObject` (
  `ObjectID` bigint(20) unsigned NOT NULL COMMENT '对象ID',
  `ObjectName` varchar(1000) NOT NULL COMMENT '对象名称，桶内唯一',
  `DocIDList` mediumblob,
  `KeyMD5High` bigint(20) NOT NULL COMMENT '桶名和对象名的摘要，用于快速定位和保证唯一性',
  `KeyMD5Low` bigint(20) NOT NULL COMMENT '桶名和对象名的摘要，用于快速定位和保证唯一性',
  `MD5High` bigint(20) NOT NULL COMMENT '对象MD5值',
  `MD5Low` bigint(20) NOT NULL COMMENT '对象MD5值',
  `BucketID` bigint(20) unsigned NOT NULL COMMENT '所属桶号',
  `Size` bigint(20) unsigned NOT NULL COMMENT '文件大小',
  `LastModified` bigint(20) unsigned NOT NULL COMMENT '最后修改时间',
  `HistoryTime` bigint(20) unsigned NOT NULL COMMENT '进入History表的时间，主要用做清理用',
  `DigestVersion` smallint(5) NOT NULL COMMENT '摘要版本信息，区分摘要的类型',
  `DeleteHint` tinyint(3) unsigned NOT NULL COMMENT '删除标志位 0 表示覆盖操作导致历史版本，1表示删除操作导致历史版本',
  `Meta` varchar(4096) NOT NULL COMMENT '用户自定义响应头集合',
  PRIMARY KEY (`KeyMD5High`,`KeyMD5Low`,`ObjectID`),
  KEY `IDX_HTime` (`HistoryTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='历史对象信息表，历史版本数据存放用。';

CREATE TABLE `ObjectIndex` (
  `BucketID` bigint(20) unsigned NOT NULL COMMENT '桶ID',
  `ObjectName` varchar(1000) COLLATE utf8_bin NOT NULL COMMENT '对象名称',
   PRIMARY KEY (`BucketID`,`ObjectName`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin ROW_FORMAT=DYNAMIC COMMENT='对象名索引';

CREATE TABLE `UploadPart` (
  `UploadID` varchar(64) NOT NULL COMMENT '所属上传SessionID',
  `DocID` varchar(64) NOT NULL COMMENT '对应底层存储的DocID',
  `Sequence` smallint(5) unsigned NOT NULL COMMENT '分块序号',
  `Size` bigint(20) unsigned NOT NULL COMMENT '本分块长度',
  `LastModify` bigint(20) unsigned NOT NULL COMMENT '最后更改时间',
  `ETag` varchar(32) NOT NULL COMMENT '分块ETag，一般为分块MD5',
  PRIMARY KEY (`UploadID`,`Sequence`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='多块上传映射表';

CREATE TABLE `UploadInfo` (
  `UploadID` varchar(64) NOT NULL COMMENT '所属上传SessionID',
  `StartTime` bigint(20) unsigned NOT NULL COMMENT '开始时间',
  `BucketID` bigint(20) unsigned NOT NULL COMMENT '桶ID',
  `ObjectName` varchar(1000) NOT NULL COMMENT '对象名称，桶内唯一',
  `ProductID` varchar(64) NOT NULL COMMENT '用户账号',
  `IsAbort` tinyint(4) NOT NULL COMMENT 'Abort标记位',
  `Meta` varchar(4096) NOT NULL COMMENT '用户自定义响应头集合',
  PRIMARY KEY (`UploadID`),
  KEY `IDX_Owner` (`ProductID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='多块上传信息表';

CREATE TABLE `AbandonUploadPart` (
  `DocID` varchar(64) NOT NULL COMMENT '对应的Ceph ObjectName',
  `CTime` bigint(20) unsigned NOT NULL COMMENT '废弃时间',
  PRIMARY KEY (`DocID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='废弃的上传分块表，主要存储被覆盖，abort的上传分块。';

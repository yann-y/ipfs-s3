#/bin/bash

# 目录变量
PROJECT_ROOT_DIR=`dirname $(readlink -f $0)`"/../.."
if [ -z ${PROJECT_ROOT_DIR} ]; then
    echo "PROJECT_ROOT_DIR empty" 
    exit -1
fi

GIT_VERSION_STR=`(git describe --all 2> /dev/null)`
if [ $? -ne 0 ]; then
    VERSION="unknown"
else
    OLD_IFS="$IFS"
    IFS="/"
    arr=($GIT_VERSION_STR)
    IFS="$OLD_IFS"
    VERSION=${arr[1]}
fi

BUILD_DIR=${PROJECT_ROOT_DIR}/./build
BOSS_RPM_SOURCE=${BUILD_DIR}/./boss-${VERSION}
mkdir -p ${BOSS_RPM_SOURCE}
rm -rf ${BOSS_RPM_SOURCE}/./*

# mkdir
mkdir -p ${BOSS_RPM_SOURCE}/etc/rc.d/init.d
mkdir -p ${BOSS_RPM_SOURCE}/lib64
mkdir -p ${BOSS_RPM_SOURCE}/usr/bin

# copy all file to dest
# now centos 7, not centos 6
cp -d ${PROJECT_ROOT_DIR}/scripts/centos6.0/boss.monitor  ${BOSS_RPM_SOURCE}/etc/rc.d/init.d
cp -d ${PROJECT_ROOT_DIR}/scripts/centos6.0/boss.dataserver  ${BOSS_RPM_SOURCE}/etc/rc.d/init.d
cp -d ${PROJECT_ROOT_DIR}/scripts/centos6.0/boss.iscsd  ${BOSS_RPM_SOURCE}/etc/rc.d/init.d

cp -d ${BUILD_DIR}/build/lib/*.so ${BOSS_RPM_SOURCE}/lib64/
cp -d ${BUILD_DIR}/build/bin/* ${BOSS_RPM_SOURCE}/usr/bin/
cp -d ${PROJECT_ROOT_DIR}/shell/boss_destroy_cluster.sh ${BOSS_RPM_SOURCE}/usr/bin/

# jemalloc
cp -d ${PROJECT_ROOT_DIR}/deps/build/lib/libjemalloc.so.2 ${BOSS_RPM_SOURCE}/lib64/

# protobuf
cp -d ${PROJECT_ROOT_DIR}/deps/build/lib/libprotobuf.so.9.0.1 ${BOSS_RPM_SOURCE}/lib64/
cp -d ${PROJECT_ROOT_DIR}/deps/build/lib/libprotobuf-lite.so.9.0.1 ${BOSS_RPM_SOURCE}/lib64/

# compress bzip2

pushd ${BUILD_DIR}
tar jcvf boss-${VERSION}".tar.bz2" boss-${VERSION}
popd

# mk rpm package
RPMBUILD_ROOT=/root/rpmbuild/
rm -rf ${RPMBUILD_ROOT}/./*
mkdir -p ${RPMBUILD_ROOT}/SPECS
mkdir -p ${RPMBUILD_ROOT}/SOURCES

cp ${PROJECT_ROOT_DIR}/rpmbuild/boss_el6.spec ${RPMBUILD_ROOT}/SPECS/
cp ${BUILD_DIR}/boss-${VERSION}".tar.bz2" ${RPMBUILD_ROOT}/SOURCES/

OS_VERSION=`uname -r|awk -F \. '{print $(NF-1)}'`
ARCH_VERSION=`uname -r|awk -F \. '{print $NF}'`

echo s/RELEASE_VERSION/${VERSION}/g
sed -i "s/RELEASE_VERSION/${VERSION}/g" ${RPMBUILD_ROOT}/SPECS/boss_el6.spec
sed -i "s/RELEASE_OS/${OS_VERSION}/g" ${RPMBUILD_ROOT}/SPECS/boss_el6.spec
sed -i "s/RELEASE_ARCH/${ARCH_VERSION}/g" ${RPMBUILD_ROOT}/SPECS/boss_el6.spec

pushd ${RPMBUILD_ROOT}
rpmbuild -ba SPECS/boss_el6.spec
popd




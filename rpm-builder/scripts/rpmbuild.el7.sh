#/bin/bash

# 目录变量
PROJECT_ROOT_DIR=`dirname $(readlink -f $0)`
if [ -z ${PROJECT_ROOT_DIR} ]; then
    echo "PROJECT_ROOT_DIR empty" 
    exit -1
fi

# GIT_VERSION_STR=`(git describe --all 2> /dev/null)`
# if [ $? -ne 0 ]; then
#     VERSION="unknown"
# else
#     OLD_IFS="$IFS"
#     IFS="/"
#     arr=($GIT_VERSION_STR)
#     IFS="$OLD_IFS"
#     VERSION=${arr[1]}
# fi

VERSION=0.1.0
BINDIR=/home/vagrant/go/project/src/github.com/yann-y/ipfs-s3

BUILD_DIR=${PROJECT_ROOT_DIR}/./build
GALAXYS3_GATEWAY_RPM_SOURCE=${BUILD_DIR}/./github.com/yann-y/ipfs-s3-${VERSION}

echo $GALAXYS3_GATEWAY_RPM_SOURCE

mkdir -p ${GALAXYS3_GATEWAY_RPM_SOURCE}
rm -rf ${GALAXYS3_GATEWAY_RPM_SOURCE}/./*

# mkdir
mkdir -p ${GALAXYS3_GATEWAY_RPM_SOURCE}/lib/systemd/system
mkdir -p ${GALAXYS3_GATEWAY_RPM_SOURCE}/usr/bin
mkdir -p ${GALAXYS3_GATEWAY_RPM_SOURCE}/usr/scripts/github.com/yann-y/ipfs-s3
# 
# # copy all file to dest
cp -d ${PROJECT_ROOT_DIR}/../service/s3-gateway.service  ${GALAXYS3_GATEWAY_RPM_SOURCE}/lib/systemd/system/

cp -d $BINDIR/build/github.com/yann-y/ipfs-s3-${VERSION}/bin/github.com/yann-y/ipfs-s3 ${GALAXYS3_GATEWAY_RPM_SOURCE}/usr/bin/
cp -d $BINDIR/build/github.com/yann-y/ipfs-s3-${VERSION}/bin/run.sh ${GALAXYS3_GATEWAY_RPM_SOURCE}/usr/scripts/github.com/yann-y/ipfs-s3/
cp -d $BINDIR/build/github.com/yann-y/ipfs-s3-${VERSION}/bin/github.com/yann-y/ipfs-s3.cfg ${GALAXYS3_GATEWAY_RPM_SOURCE}/usr/scripts/github.com/yann-y/ipfs-s3/

# compress bzip2

pushd ${BUILD_DIR}
tar jcvf github.com/yann-y/ipfs-s3-${VERSION}".tar.bz2" github.com/yann-y/ipfs-s3-${VERSION}
popd

# 
# mk rpm package
RPMBUILD_ROOT=/root/rpmbuild/
mkdir -p ${RPMBUILD_ROOT}/SPECS
mkdir -p ${RPMBUILD_ROOT}/SOURCES

cp ${PROJECT_ROOT_DIR}/../spec/galaxy_s3_gateway_el7.spec ${RPMBUILD_ROOT}/SPECS/
cp ${BUILD_DIR}/github.com/yann-y/ipfs-s3-${VERSION}".tar.bz2" ${RPMBUILD_ROOT}/SOURCES/
rm -rf ${BUILD_DIR}

 
OS_VERSION=`uname -r|awk -F \. '{print $(NF-1)}'`
ARCH_VERSION=`uname -r|awk -F \. '{print $NF}'`

 
sed -i "s/RELEASE_VERSION/${VERSION}/g" ${RPMBUILD_ROOT}/SPECS/galaxy_s3_gateway_el7.spec
sed -i "s/RELEASE_OS/${OS_VERSION}/g" ${RPMBUILD_ROOT}/SPECS/galaxy_s3_gateway_el7.spec
sed -i "s/RELEASE_ARCH/${ARCH_VERSION}/g" ${RPMBUILD_ROOT}/SPECS/galaxy_s3_gateway_el7.spec

pushd ${RPMBUILD_ROOT}
rpmbuild -ba SPECS/galaxy_s3_gateway_el7.spec
popd

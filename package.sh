DIR=`pwd`

version=`cat VERSION`
make
rm -rf $DIR/dist

sed -i "s/GALAXYS3_GATEWAY_VERSION/${version}/g" rpm-builder/scripts/rpmbuild.el7.sh
sed -i "s#GALAXYS3_GATEWAY_BIN_DIR#${DIR}#g" rpm-builder/scripts/rpmbuild.el7.sh

cd $DIR/rpm-builder/scripts/
bash rpmbuild.el7.sh

cd $DIR
rm -rf build

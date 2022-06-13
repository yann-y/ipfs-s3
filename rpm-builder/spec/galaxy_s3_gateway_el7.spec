# Introduction section
%define name github.com/yann-y/ipfs-s3
%define version RELEASE_VERSION
%define release RELEASE_OS
%define arch RELEASE_ARCH

Name: %{name}
Summary: Galaxy AWS-S3 Gateway
Version: %{version}
Release: %{release}
BuildArch: %{arch}
# Group: System Environment/Libraries

# company info
vendor: HangZhou Galaxy Ltd.
Packager: tracymacding <tracymacding@gmail.com>
# URL: homepage  
# Copyright:
License: GPLv2

# spcefic source files, use %{version}
# service config
SOURCE: %{name}-%{version}.tar.bz2 

# pathes section:no patches 
# spec Requirements 
# see subpackage

# spec Provides 
# Provides:
# see subpackage

# spec Conflicts
# Conflicts:

# spec Obsoletes
# Obsoletes:

%description
Galaxy S3 Gateway is s3 front-end from galaxy LTD. 

# prep section
%prep
%setup -q

# build section
# now need no build section, because we package the binaries
# %build

# install section
# this is important, here not use make install, use install
%install

# mkdir 
mkdir -p ${RPM_BUILD_ROOT}/lib/systemd/system
mkdir -p ${RPM_BUILD_ROOT}/usr/bin
mkdir -p ${RPM_BUILD_ROOT}/usr/script/github.com/yann-y/ipfs-s3

# copy all file to dest
install -m 0755 -t ${RPM_BUILD_ROOT}/lib/systemd/system lib/systemd/system/*
install -m 0755 -t ${RPM_BUILD_ROOT}/usr/bin usr/bin/*
install -m 0755 -t ${RPM_BUILD_ROOT}/usr/script/github.com/yann-y/ipfs-s3 usr/scripts/github.com/yann-y/ipfs-s3/*

# clean section
%clean
#rm -rf $RPM_BUILD_ROOT

# files section
# list the files to go into the binary RPM, along with defined file attributes
# see subpackage

#%post -p /sbin/ldconfig
#%postun -p /sbin/ldconfig

# create Subpackages, name derive to galaxyfs-%{packagename}
# all subpackages section should provide name

#%package -n github.com/yann-y/ipfs-s3
#Summary: Galaxy AWS-S3 gateway

%description -n github.com/yann-y/ipfs-s3
Galaxy AWS-S3 gateway

# Requires: galaxyfs-common >= %{version}
# Requires: curl
# Requires: rrdtool
# Requires: net-snmp-libs

%pre -n github.com/yann-y/ipfs-s3
if [ ! -d "/opt/galaxy" ]; then
     mkdir -m 755 /opt/galaxy
fi

if [ ! -d "/opt/galaxy/galaxy-s3-gw" ]; then
     mkdir -m 755 -p /opt/galaxy/galaxy-s3-gw
fi
if [ ! -d "/opt/galaxy/galaxy-s3-gw/bin" ]; then
     mkdir -m 755 -p /opt/galaxy/galaxy-s3-gw/bin
fi

%files -n github.com/yann-y/ipfs-s3
%defattr(755,root,root)
/lib/systemd/system/s3-gateway.service
/usr/bin/github.com/yann-y/ipfs-s3
/usr/script/github.com/yann-y/ipfs-s3/run.sh
/usr/script/github.com/yann-y/ipfs-s3/github.com/yann-y/ipfs-s3.cfg

%post -n github.com/yann-y/ipfs-s3
cp -d /usr/script/github.com/yann-y/ipfs-s3/run.sh /opt/galaxy/galaxy-s3-gw/bin/run.sh
cp -d /usr/script/github.com/yann-y/ipfs-s3/github.com/yann-y/ipfs-s3.cfg /opt/galaxy/galaxy-s3-gw/bin/github.com/yann-y/ipfs-s3.cfg
cp -d /usr/bin/github.com/yann-y/ipfs-s3 /opt/galaxy/galaxy-s3-gw/bin/github.com/yann-y/ipfs-s3
systemctl enable s3-gateway
#/sbin/chkconfig metanode on

%preun -n github.com/yann-y/ipfs-s3
if [ "$1" = 0 ]; then
  systemctl stop s3-gateway > /dev/null 2>&1
  systemctl disable s3-gateway
fi

%changelog

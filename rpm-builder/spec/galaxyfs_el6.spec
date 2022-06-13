# Introduction section
%define name boss
%define version RELEASE_VERSION
%define release RELEASE_OS
%define arch RELEASE_ARCH

Name: %{name}
Summary: BlueOcean Storage System
Version: %{version}
Release: %{release}
BuildArch: %{arch}
# Group: System Environment/Libraries

# company info
vendor: Shanghai Xiaoyun Ltd.
Packager: chengyu <yu.cheng@shxiaoyun.com.cn>
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
BlueOcean Storage System is a large distributed storage system from shxiaoyun LTD. 

# prep section
%prep
%setup -q

# build section
# now need no build section, because we package the binaries
# %build

# install section
# this is important, here not use make install, use install
%install
rm -rf $RPM_BUILD_ROOT

# mkdir 
mkdir -p ${RPM_BUILD_ROOT}/etc/rc.d/init.d
mkdir -p ${RPM_BUILD_ROOT}/lib64
mkdir -p ${RPM_BUILD_ROOT}/usr/bin

# copy all file to dest
install -m 0755 -t ${RPM_BUILD_ROOT}/etc/rc.d/init.d etc/rc.d/init.d/*
install -m 0755 -t ${RPM_BUILD_ROOT}/lib64 lib64/*
install -m 0755 -t ${RPM_BUILD_ROOT}/usr/bin usr/bin/*

# clean section
%clean
rm -rf $RPM_BUILD_ROOT

# files section
# list the files to go into the binary RPM, along with defined file attributes
# see subpackage

#%post -p /sbin/ldconfig
#%postun -p /sbin/ldconfig

# create Subpackages, name derive to boss-%{packagename}
# all subpackages section should provide name

# boss-common : libs (include jemalloc, protobuf)
%package -n boss-common
Summary: BlueOcean Storage System libs

%description -n boss-common
BOSS shared lib

# provides : maybe no need
%files -n boss-common
%defattr(755,root,root)
/lib64/*

%post -n boss-common
ln -s /lib64/libjemalloc.so.2 /lib64/libjemalloc.so
ln -s  /lib64/libprotobuf-lite.so.9.0.1 /lib64/libprotobuf-lite.so
ln -s  /lib64/libprotobuf-lite.so.9.0.1 /lib64/libprotobuf-lite.so.9
ln -s  /lib64/libprotobuf.so.9.0.1 /lib64/libprotobuf.so
ln -s  /lib64/libprotobuf.so.9.0.1 /lib64/libprotobuf.so.9
/sbin/ldconfig

%postun -n boss-common
rm -f /lib64/libjemalloc.so
rm -f /lib64/libprotobuf-lite.so
rm -f /lib64/libprotobuf-lite.so.9
rm -f /lib64/libprotobuf.so
rm -f /lib64/libprotobuf.so.9

# boss-mon : boss monitor bin and systemd service file
%package -n boss-mon
Summary: BlueOcean Storage System monitor

%description -n boss-mon
BOSS monitor

Requires: boss-common >= %{version}
Requires: curl
Requires: rrdtool
Requires: net-snmp-libs

%pre -n boss-mon
if [ ! -d "/opt/boss" ]; then
     mkdir -m 755 /opt/boss
fi
if [ ! -d "/opt/boss/mon" ]; then
     mkdir -m 755 /opt/boss/mon
fi
if [ ! -d "/opt/boss/log" ]; then
     mkdir -m 755 /opt/boss/log
fi

%files -n boss-mon
%defattr(755,root,root)
/etc/rc.d/init.d/boss.monitor
/usr/bin/boss_admin
/usr/bin/boss_destroy_cluster.sh
/usr/bin/boss_iscsi
/usr/bin/boss_mon

%post -n boss-mon
/sbin/chkconfig boss.monitor on

%preun -n boss-mon
if [ "$1" = 0 ]; then
  /sbin/service boss.monitor stop > /dev/null 2>&1
  /sbin/chkconfig boss.monitor off
fi


# boss-ds : boss dataserver bin and systemd service file
%package -n boss-ds
Summary: BlueOcean Storage System dataserver

%description -n boss-ds
BOSS dataserver

Requires: boss-common >= %{version}
Requires: libaio 

%pre -n boss-ds
if [ ! -d "/opt/boss" ]; then
     mkdir -m 755 /opt/boss
fi
if [ ! -d "/opt/boss/log" ]; then
     mkdir -m 755 /opt/boss/log
fi
if [ ! -d "/opt/boss/ds" ]; then
     mkdir -m 755 /opt/boss/ds
fi

%files -n boss-ds
%defattr(755,root,root)
/etc/rc.d/init.d/boss.dataserver
/etc/rc.d/init.d/boss.iscsid
/usr/bin/boss_ds
/usr/bin/boss_iscsid
%doc

%post -n boss-ds
/sbin/chkconfig boss.dataserver on
/sbin/chkconfig boss.iscsid on

%preun -n boss-ds
if [ "$1" = 0 ]; then
  /sbin/service boss.dataserver stop > /dev/null 2>&1
  /sbin/chkconfig boss.dataserver off

  /sbin/service boss.iscsid stop > /dev/null 2>&1
  /sbin/chkconfig boss.iscsid off
fi

# boss-client : boss client bin
%package -n boss-client
Summary: BlueOcean Storage System client 

%description -n boss-client
BOSS dataserver

Requires: boss-common >= %{version}

%pre -n boss-client
if [ ! -d "/opt/boss" ]; then
     mkdir -m 755 /opt/boss
fi
if [ ! -d "/opt/boss/log" ]; then
     mkdir -m 755 /opt/boss/log
fi
if [ ! -d "/opt/boss/cli" ]; then
     mkdir -m 755 /opt/boss/cli
fi
if [ ! -d "/opt/boss/cli/map" ]; then
     mkdir -m 777 /opt/boss/cli/map
fi

%files -n boss-client
%defattr(755,root,root)
/usr/bin/boss_ds_check_disk
/usr/bin/boss_ds_check_md5
/usr/bin/boss_ds_check_replica
/usr/bin/boss_config
/usr/bin/boss_debug
/usr/bin/boss_fsck
/usr/bin/boss_key
/usr/bin/boss_ping
/usr/bin/compute_cube
/usr/bin/container_create
/usr/bin/container_delete
/usr/bin/container_list
/usr/bin/files_create
/usr/bin/fio-boss
/usr/bin/map_create
/usr/bin/map_deploy
/usr/bin/map_isolate
/usr/bin/map_rebalance
/usr/bin/map_show
/usr/bin/object_delete
/usr/bin/object_get
/usr/bin/object_list
/usr/bin/object_put
/usr/bin/objects_create
/usr/bin/object_stat
/usr/bin/object_xattr_get
/usr/bin/object_xattr_put
/usr/bin/pool_list
/usr/bin/recyclebin_clear
/usr/bin/snap_clone
/usr/bin/snap_create
/usr/bin/snap_delete
/usr/bin/snap_list
/usr/bin/snap_recover
#/usr/bin/sshpass
/usr/bin/volume_copy
/usr/bin/volume_copyin
/usr/bin/volume_copyout
/usr/bin/volume_create
/usr/bin/volume_delete
/usr/bin/volume_info
/usr/bin/volume_list
/usr/bin/volume_resize
/usr/bin/volume_umount

%changelog




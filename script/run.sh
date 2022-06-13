#!/bin/bash

RUNBIN="${BASH_SOURCE-$0}"
RUNBIN="$(dirname "${RUNBIN}")"
BINDIR="$(cd "${RUNBIN}"; pwd)"

# parse gateway config file
config_file="/opt/galaxy/galaxy-s3-gw/bin/github.com/yann-y/ipfs-s3.cfg"
if [ ! -f "$config_file" ]; then
    echo 's3 gateway config file '/opt/galaxy/galaxy-s3-gw/bin/github.com/yann-y/ipfs-s3.cfg' not exist'
    exit 1
fi

while read line; do
    if [[ ${line:0:1} != "#" ]]; then
        eval "$line"
    fi
done < $config_file

mkdir -p $log_dir

cd $BINDIR && nohup ./github.com/yann-y/ipfs-s3 -gfs_zk_addr=$zookeeper -log_dir=$log_dir -logtostderr=false -mongodb_addr=$mongodb_address -port=$listen_port -gfs_scheduler_path=$scheduler_zk_path > nohup.out 2>&1 &

PID=$!
sleep 1
echo "start github.com/yann-y/ipfs-s3 finished, PID=$PID"
echo "checking if $PID is running..."
sleep 2
kill -0 $PID > /dev/null 2>&1
if [ $? -eq 0 ]
then
	echo "$PID is running, start github.com/yann-y/ipfs-s3 success."
	exit 0
else
	echo "start github.com/yann-y/ipfs-s3 failed."
	exit 1
fi

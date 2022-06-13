version='0.1.0'

git clone https://git.coding.net/Mr-x/galaxy-release.git

mkdir -p roles/github.com/yann-y/ipfs-s3/files
cp galaxy-release/github.com/yann-y/ipfs-s3/$version/github.com/yann-y/ipfs-s3.tar.gz  roles/github.com/yann-y/ipfs-s3/files/
cp ~/Downloads/go1.7.3.linux-amd64.tar.gz  roles/github.com/yann-y/ipfs-s3/files/

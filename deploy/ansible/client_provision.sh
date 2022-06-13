#! /bin/bash

echo -e  'y\n' | ssh-keygen -q -t rsa -N "" -f ~/.ssh/id_rsa
# wget http://dl.fedoraproject.org/pub/epel/7/x86_64/e/epel-release-7-8.noarch.rpm
wget https://mirrors.tuna.tsinghua.edu.cn/epel//7Server/x86_64/e/epel-release-7-9.noarch.rpm
sudo rpm -ivh epel-release-7-9.noarch.rpm
sudo yum --enablerepo=epel -y install sshpass

/usr/bin/sshpass -p vagrant /usr/bin/ssh-copy-id -i ~/.ssh/id_rsa.pub -o StrictHostKeyChecking=no 192.168.100.100

sudo easy_install -i http://pypi.douban.com/simple/ pip
mkdir ~/.pip
cp pip.conf ~/.pip/
sudo mkdir /root/.pip
sudo cp pip.conf /root/.pip/
sudo yum -y install git
sudo yum -y install python-devel
sudo yum -y install openssl-devel
sudo yum -y install libffi-devel
sudo pip install ansible==1.9.0.1 -i http://pypi.douban.com/simple/  --trusted-host pypi.douban.com

# start ansible-playbook
ansible all -m ping
ansible-playbook github.com/yann-y/ipfs-s3.yml --become-user=vagrant

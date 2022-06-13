### github.com/yann-y/ipfs-s3自动化部署

#### 说明

github.com/yann-y/ipfs-s3自动化部署用于自动在Vagrant虚拟机上安装github.com/yann-y/ipfs-s3及其相关依赖，主要包括：

* github.com/yann-y/ipfs-s3: 默认安装路径/home/$account/galaxy_s3_gateway
* mongodb: 通过service方式安装和启动

#### 步骤

1. 环境准备: cd deploy/ansilbe & bash prepare.sh, 这将准备好自动部署依赖的所有代码/第三方依赖等
2. 启动vagrant虚拟机: cd deploy/ & vagrant up --no-provision, 根据配置自动启动github.com/yann-y/ipfs-s3-01
3. provision vagrant虚拟机: cd deploy/ansilbe & vagrant provision github.com/yann-y/ipfs-s3-01

#### 注意事项

* 在MacOS下启动vagrant虚拟机时设置虚拟机ip的时候发现ip最后一个部分超过300好像就会启动失败

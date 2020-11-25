### 1. 基本环境配置
查看内核版本
```shell script
cat /proc/version
```
本实验所有的服务器使用的均为Redhat操作系统。
```shell script
sudo vim /etc/resolv.conf
```
添加以下内容：
>nameserver 8.8.8.8 #google域名服务器
>
>nameserver 8.8.4.4 #google域名服务器
```shell script
sudo yum -y install git wget curl zsh tree
```
### 2. 安装oh-my-zsh
```shell script
sudo curl -L https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh | sh 
chsh -s /bin/zsh 
cat /etc/shells # 查看当前所有的
```
eixt之后再重新ssh登录，否则source ~/.zshrc会报错。
### 3. 安装go
```shell script
wget https://studygolang.com/dl/golang/go1.15.4.linux-amd64.tar.gz
sudo tar -zxvf go1.15.4.linux-amd64.tar.gz -C /opt
mkdir ~/Go
sudo vim ~/.zshrc
```
输入以下内容：
>export GOPATH=~/Go
>
>export GOROOT=/opt/go
>
>export GOTOOLS=$GOROOT/pkg/tool
>
>export GOPROXY=https://goproxy.io
>
>export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

```shell script
source ~/.zshrc
```
### 4. 下载MyRaft
```shell script
cd ~/Go
git clone https://github.com/yunxiao3/MyRaft.git
```
### 5. 安装goleveldb
```shell script
export GO111MODULE=off
go get github.com/syndtr/goleveldb/leveldb
```
### 6.安装protobuf
```shell script
git clone https://github.com/golang/protobuf.git $GOPATH/src/github.com/golang/protobuf
cd $GOPATH/src/github.com/golang/protobuf
go install ./proto
go install ./protoc-gen-go
```
### 7. 安装grpc
```shell script
git clone https://github.com/grpc/grpc-go.git $GOPATH/src/google.golang.org/grpc
git clone https://github.com/golang/net.git $GOPATH/src/golang.org/x/net
git clone https://github.com/golang/text.git $GOPATH/src/golang.org/x/text
git clone https://github.com/google/go-genproto.git $GOPATH/src/google.golang.org/genproto
git clone https://github.com/golang/sys.git $GOPATH/src/golang.org/x/sys
git clone https://github.com/protocolbuffers/protobuf-go.git $GOPATH/src/google.golang.org/protobuf
cd $GOPATH/src/
go install google.golang.org/grpc
```
安装grpc的时候报错缺少哪些依赖包，git clone相应的依赖包。
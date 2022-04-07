# kubebuilder创建简单crd实践
# Prerequisites
* go version go1.16.4 linux/arm64
* docker version 20.10.9
* kubectl version v1.19.1
* kubebuilder version 3.2.0
* Access to a Kubernetes v1.19.1 cluster
# 创建项目
创建项目可以[参考](https://book.kubebuilder.io/quick-start.html)
```bash
$ mkdir -p $GOPATH/hellocrd
$ cd $GOPATH/hellocrd
$ kubebuilder init --domain freelizhun.com --repo github.com/freelizhun/hellocrd

# 以下均输入y
$ kubebuilder create api --group myapp --version v1 --kind Hello

# 创建crd等配置文件
$ make manifests
```

# 运行测试
1.开一个终端
```bash
$ mkdir -p $GOPATH
$ cd $GOPATH
$ git git clone https://github.com/freelizhun/hellocrd.git
$ cd hellocrd
$ go mod download
$ go mod tidy

# 先创建crd对象
$ kubectl apply -f config/crd/bases/myapp.freelizhun.com_hellos.yaml
$ go run main.go
```
2.另外再开一个终端
```bash
# 创建hello资源对象
$ kubectl apply -f config/samples/myapp_v1_hello.yaml
```
# 发布到k8s集群
1.构建镜像
```bash
$ mkdir -p $GOPATH
$ cd $GOPATH
$ git git clone https://github.com/freelizhun/hellocrd.git
$ cd hellocrd
$ go mod download
$ go mod tidy

# 编译与构建镜像
$ go build -o bin/manager main.go
$ docker build -t regitryserver:4000/hello:latest -f Dockerfile.test .
```
2.发布
```bash
# 先创建crd对象
$ kubectl apply -f config/crd/bases/myapp.freelizhun.com_hellos.yaml

# 绑定权限
$ kubectl create ns system
$ kubectl apply -f deploy/role.yaml
$ kubectl apply -f deploy/role_binding.yaml
$ kubectl apply -f deploy/service_account.yaml

$ kubectl apply -f deploy/manager.yaml

# 创建hello资源对象进行测试
$ kubectl apply -f deploy/myapp_v1_hello.yaml
```
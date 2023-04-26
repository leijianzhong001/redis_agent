# Current Operator version
# “？=”表示如果该变量没有被赋值，则赋予等号后的值
VERSION ?= v1.0.2

# Image URL to use all building/pushing image targets
IMG ?= registry.cn-hangzhou.aliyuncs.com/leijianzhong/redis_agent:$(VERSION)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
# 如果 go env GOBIN 为空，则从 $GOPATH/bin下查找可执行文件
ifeq (,$(shell go env GOBIN))
# GOBIN=/root/go/bin
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: go-build

# 这里的意思是将main.go编译为 bin/redis-agent
go-build: fmt vet
	go build -o bin/redis-agent main.go

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
# Go vet 命令在编写代码时非常有用。它可以帮助您检测应用程序中任何可疑、异常或无用的代码。
vet:
	go vet ./...

# Build the docker image
docker-build:
	docker build -t ${IMG} .

# Push the docker image
docker-push:
	docker push ${IMG}

# $(MAKEFILE_LIST) => Makefile
# $(lastword $(MAKEFILE_LIST)) => Makefile
# $(abspath $(lastword $(MAKEFILE_LIST))) => /root/redis-operator/Makefile
# dirname $(abspath $(lastword $(MAKEFILE_LIST))) /root/redis-operator
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
# go-get-tool will 'go get' any package $2 and install it to $1.
# 定义一个自定义函数
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef



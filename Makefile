.PHONY: default all prepare build product clean

PWD=$(shell pwd)
PRJNAME := "puller"
ImgName := "bbklab/puller"

# Used to populate version variable in main package.
VERSION=$(shell cat VERSION.txt)
BUILD_TIME=$(shell date -u +%Y-%m-%d_%H:%M:%S_%Z)
PKG := "github.com/Dataman-Cloud/puller"
gopkgs=$(shell go list ./... | grep -v "integration-test")
gitCommit=$(shell git rev-parse --short HEAD)
gitDirty=$(shell git status --porcelain --untracked-files=no)
GIT_COMMIT=$(gitCommit)
ifneq ($(gitDirty),)  # ---> gitDirty != ""
GIT_COMMIT=$(gitCommit)-dirty
endif
# FIXME deprecated
# BUILD_FLAGS=-X $(PKG)/version.version=$(VERSION) -X $(PKG)/version.gitCommit=$(GIT_COMMIT) -X $(PKG)/version.buildAt=$(BUILD_TIME) -w -s

default: binary

all: integration-test push

product: clean binary 

prod: product

prepare:
	mkdir -p bin/
	mkdir -p product/

gocheck: gometalinter

# use gometalinter to replace all of above go linters
gometalinter:
	docker run --name gocheck-puller --rm \
		-v $(PWD):/go/src/${PKG}:ro \
		bbklab/gometalinter:latest \
		gometalinter --skip=integration-test --skip=vendor --skip=assets \
			--deadline=120s \
			--disable-all \
			--enable=gofmt --enable=golint \
			--enable=vet --enable=goconst \
			/go/src/${PKG}/...
	echo "  --- Gometalinter Passed!"

binary: prepare
ifeq (${NOLINT},)   # NOLINT == ""
	@make gocheck
endif
ifeq (${ENV_CIRCLECI}, true)
	@make host-build
else
	@make docker-build
endif

product: image

image: binary
	docker build --force-rm -t $(ImgName):$(gitCommit) -f Dockerfile .
	docker tag $(ImgName):$(gitCommit) $(ImgName):latest
	echo "Puller Image Built!"

push: product
	docker push $(ImgName):$(gitCommit)
	docker push $(ImgName):latest
	echo "Product Pushed!"

#
# docker-build (build binary via docker container)
#
docker-build:
	docker run --rm \
		--name buildpuller \
		-w /go/src/${PKG} \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-v $(PWD):/go/src/${PKG}:rw \
		golang:1.10-alpine \
		sh -c 'go build -a -ldflags "${BUILD_FLAGS}" -o bin/puller ${PKG}/main/'
	echo "Binary Built!"

#
# host-build (direct build binary by using system installed golang)
#  mainly used for CI env
#
host-build:
	env CGO_ENABLED=0 GOOS=linux go build -a -ldflags "${BUILD_FLAGS}" -o bin/puller ${PKG}/main/
	echo "Binary Built!"

# update dep
#  - HTTP[S]_PROXY:  golang use proxy
#  - ALL_PROXY:   git use proxy
update-dep:
	cd ${GOPATH}/src/$(PKG) && \
		env \
		HTTP_PROXY=socks5://127.0.0.1:1080 HTTPS_PROXY=socks5://127.0.0.1:1080 \
		ALL_PROXY=socks5://127.0.0.1:1080 \
		dep ensure -v

#
# clean up outdated
# 
clean:
	rm -fv  bin/puller
	rm -fv  product/*

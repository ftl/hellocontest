VERSION_NUMBER=$(shell git describe --tags | sed -E 's#v##' | sed -E 's#-#_#g')
BINARY_NAME ?= hellocontest
DESTDIR ?=
INSTALL_DIR ?= /usr/bin

all: clean test build

clean:
	go clean
	rm -f ${BINARY_NAME}

deps:
	go get -v -t -d ./...

generate:
	go generate ./core/pb

version_number:
	@echo ${VERSION_NUMBER}

test:
	go test -v -timeout=30s ./...

build:
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -v -ldflags "-linkmode external -extldflags \"${LDFLAGS}\" -X main.version=${VERSION_NUMBER}" -o ${BINARY_NAME}

run: build
	./${BINARY_NAME}

install:
	mkdir -p ${DESTDIR}${INSTALL_DIR}
	cp ./${BINARY_NAME} ${DESTDIR}${INSTALL_DIR}/${BINARY_NAME}

uninstall:
	rm ${INSTALL_DIR}/${BINARY_NAME}

checkout_latest:
	git checkout `git tag --sort=committerdate | tail -1`

latest: clean checkout_latest test build

debpkg:
	sed -i -E "s#!THE_VERSION!#${VERSION_NUMBER}#" ./.debpkg/DEBIAN/control
	dpkg-deb --build ./.debpkg .
	git restore ./.debpkg/DEBIAN/control

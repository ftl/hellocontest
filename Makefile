BINARY_NAME = hellocontest
VERSION_NUMBER ?= $(shell git describe --tags | sed -E 's#v##')

DESTDIR ?=
BINDIR ?= /usr/bin
SHAREDIR ?= /usr/share

all: clean test build

clean:
	go clean
	rm -f ${BINARY_NAME}

deps:
	go get -v -t -d ./...

generate:
	go generate ./core/pb
	go generate ./core/session

version_number:
	@echo ${VERSION_NUMBER}

test:
	go test -v -timeout=30s ./...

build:
	go build -trimpath -buildmode=pie -mod=readonly -modcacherw -v -ldflags "-linkmode external -extldflags \"${LDFLAGS}\" -X main.version=${VERSION_NUMBER}" -o ${BINARY_NAME}

run: build
	./${BINARY_NAME}

install:
	mkdir -p ${DESTDIR}${BINDIR}
	cp ./${BINARY_NAME} ${DESTDIR}${BINDIR}/${BINARY_NAME}
	mkdir -p ${DESTDIR}${SHAREDIR}/applications
	cp ./.assets/${BINARY_NAME}.desktop ${DESTDIR}${SHAREDIR}/applications/${BINARY_NAME}.desktop

uninstall:
	rm ${DESTDIR}${BINDIR}/${BINARY_NAME}

checkout_latest:
	git checkout `git tag --sort=committerdate | tail -1`

latest: clean checkout_latest test build

debpkg:
	sed -i -E "s#!THE_VERSION!#${VERSION_NUMBER}#" ./.debpkg/DEBIAN/control
	mkdir -p ./.debpkg${BINDIR}
	cp ./${BINARY_NAME} ./.debpkg${BINDIR}/${BINARY_NAME}
	mkdir -p ./.debpkg${SHAREDIR}/applications
	cp ./.assets/${BINARY_NAME}.desktop ./.debpkg${SHAREDIR}/applications/${BINARY_NAME}.desktop
	dpkg-deb --build ./.debpkg .
	git restore ./.debpkg/DEBIAN/control

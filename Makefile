BINARY_NAME = hellocontest
VERSION_NUMBER ?= $(shell git describe --tags | sed -E 's#v##')

ARCH = x86_64
DESTDIR ?=
BINDIR ?= /usr/bin
SHAREDIR ?= /usr/share
APPIMAGETOOL ?= appimagetool

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

config:
	xdg-open ~/.config/hamradio/hellocontest.json

install:
	mkdir -p ${DESTDIR}${BINDIR}
	cp ./${BINARY_NAME} ${DESTDIR}${BINDIR}/${BINARY_NAME}
	mkdir -p ${DESTDIR}${SHAREDIR}/applications
	cp ./.assets/${BINARY_NAME}.desktop ${DESTDIR}${SHAREDIR}/applications/${BINARY_NAME}.desktop
	mkdir -p ${DESTDIR}${SHAREDIR}/icons/hicolor/256x256/apps/
	cp ./.assets/${BINARY_NAME}-256x256.png ${DESTDIR}${SHAREDIR}/icons/hicolor/256x256/apps/${BINARY_NAME}.png
	mkdir -p ${DESTDIR}${SHAREDIR}/icons/hicolor/128x128/apps/
	cp ./.assets/${BINARY_NAME}-128x128.png ${DESTDIR}${SHAREDIR}/icons/hicolor/128x128/apps/${BINARY_NAME}.png
	mkdir -p ${DESTDIR}${SHAREDIR}/icons/hicolor/48x48/apps/
	cp ./.assets/${BINARY_NAME}-48x48.png ${DESTDIR}${SHAREDIR}/icons/hicolor/48x48/apps/${BINARY_NAME}.png
	mkdir -p ${DESTDIR}${SHAREDIR}/icons/hicolor/32x32/apps/
	cp ./.assets/${BINARY_NAME}-32x32.png ${DESTDIR}${SHAREDIR}/icons/hicolor/32x32/apps/${BINARY_NAME}.png
	mkdir -p ${DESTDIR}${SHAREDIR}/icons/hicolor/16x16/apps/
	cp ./.assets/${BINARY_NAME}-16x16.png ${DESTDIR}${SHAREDIR}/icons/hicolor/16x16/apps/${BINARY_NAME}.png
	mkdir -p ${DESTDIR}${SHAREDIR}/icons/hicolor/scalable/apps/
	cp ./.assets/${BINARY_NAME}.svg ${DESTDIR}${SHAREDIR}/icons/hicolor/scalable/apps/${BINARY_NAME}.svg

uninstall:
	rm -f ${DESTDIR}${BINDIR}/${BINARY_NAME}
	rm -f ${DESTDIR}${SHAREDIR}/applications/${BINARY_NAME}.desktop
	rm -f ${DESTDIR}${SHAREDIR}/icons/hicolor/256x256/apps/${BINARY_NAME}.png
	rm -f ${DESTDIR}${SHAREDIR}/icons/hicolor/128x128/apps/${BINARY_NAME}.png
	rm -f ${DESTDIR}${SHAREDIR}/icons/hicolor/48x48/apps/${BINARY_NAME}.png
	rm -f ${DESTDIR}${SHAREDIR}/icons/hicolor/32x32/apps/${BINARY_NAME}.png
	rm -f ${DESTDIR}${SHAREDIR}/icons/hicolor/16x16/apps/${BINARY_NAME}.png
	rm -f ${DESTDIR}${SHAREDIR}/icons/hicolor/scalable/apps/${BINARY_NAME}.svg

update_icons:
	gtk-update-icon-cache ${DESTDIR}${SHAREDIR}/icons/hicolor

checkout_latest:
	git checkout `git tag --sort=committerdate | tail -1`

latest: clean checkout_latest test build

debpkg:
	sed -i -E "s#!THE_VERSION!#${VERSION_NUMBER}#" ./.debpkg/DEBIAN/control
	mkdir -p ./.debpkg${BINDIR}
	cp ./${BINARY_NAME} ./.debpkg${BINDIR}/${BINARY_NAME}
	chmod a+x ./.debpkg${BINDIR}/${BINARY_NAME}
	mkdir -p ./.debpkg${SHAREDIR}/applications
	cp ./.assets/${BINARY_NAME}.desktop ./.debpkg${SHAREDIR}/applications/${BINARY_NAME}.desktop
	mkdir -p ./.debpkg${SHAREDIR}/icons/hicolor/256x256/apps/
	cp ./.assets/${BINARY_NAME}-256x256.png ./.debpkg${SHAREDIR}/icons/hicolor/256x256/apps/${BINARY_NAME}.png
	mkdir -p ./.debpkg${SHAREDIR}/icons/hicolor/128x128/apps/
	cp ./.assets/${BINARY_NAME}-128x128.png ./.debpkg${SHAREDIR}/icons/hicolor/128x128/apps/${BINARY_NAME}.png
	mkdir -p ./.debpkg${SHAREDIR}/icons/hicolor/48x48/apps/
	cp ./.assets/${BINARY_NAME}-48x48.png ./.debpkg${SHAREDIR}/icons/hicolor/48x48/apps/${BINARY_NAME}.png
	mkdir -p ./.debpkg${SHAREDIR}/icons/hicolor/32x32/apps/
	cp ./.assets/${BINARY_NAME}-32x32.png ./.debpkg${SHAREDIR}/icons/hicolor/32x32/apps/${BINARY_NAME}.png
	mkdir -p ./.debpkg${SHAREDIR}/icons/hicolor/16x16/apps/
	cp ./.assets/${BINARY_NAME}-16x16.png ./.debpkg${SHAREDIR}/icons/hicolor/16x16/apps/${BINARY_NAME}.png
	mkdir -p ./.debpkg${SHAREDIR}/icons/hicolor/scalable/apps/
	cp ./.assets/${BINARY_NAME}.svg ./.debpkg${SHAREDIR}/icons/hicolor/scalable/apps/${BINARY_NAME}.svg
	dpkg-deb --build ./.debpkg .
	git restore ./.debpkg/DEBIAN/control

prepare_appimage:
	mkdir -p ./.appimage${BINDIR}
	cp ./${BINARY_NAME} ./.appimage${BINDIR}/${BINARY_NAME}
	chmod a+x ./.appimage${BINDIR}/${BINARY_NAME}
	cp ./.assets/${BINARY_NAME}.desktop ./.appimage/${BINARY_NAME}.desktop
	cp ./.assets/${BINARY_NAME}-256x256.png ./.appimage/${BINARY_NAME}.png
	cp ./.assets/${BINARY_NAME}.svg ./.appimage/${BINARY_NAME}.svg
	
appimage: prepare_appimage
	env ARCH=${ARCH} ${APPIMAGETOOL} .appimage ${BINARY_NAME}-${VERSION_NUMBER}-${ARCH}.AppImage

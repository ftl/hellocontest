BINARY_NAME = hellocontest
VERSION_NUMBER ?= $(shell git describe --tags | sed -E 's#v##')

ARCH = x86_64
DESTDIR ?=
BINDIR ?= /usr/bin
SHAREDIR ?= /usr/share
QMAKE ?= /usr/bin/qmake6
QT_PLUGINS = $(shell ${QMAKE} -query QT_INSTALL_PLUGINS)
APPDIR = ./.appimage-build/AppDir
APPIMAGE_TOOLS = ./.appimage-build/tools
LINUXDEPLOY_VERSION ?= 1-alpha-20251107-1
LINUXDEPLOY_PLUGIN_QT_VERSION ?= 1-alpha-20250213-1
APPIMAGETOOL_VERSION ?= 1.9.1
LINUXDEPLOY ?= ${APPIMAGE_TOOLS}/linuxdeploy-${ARCH}.AppImage
LINUXDEPLOY_PLUGIN_QT ?= ${APPIMAGE_TOOLS}/linuxdeploy-plugin-qt-${ARCH}.AppImage
APPIMAGETOOL ?= ${APPIMAGE_TOOLS}/appimagetool-${ARCH}.AppImage

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
	go test -timeout=30s ./core/...

build:
	go build -trimpath -buildmode=pie -ldflags "-s -w -X main.version=${VERSION_NUMBER}" -o ${BINARY_NAME}

run: build
	./${BINARY_NAME}

screenshots: build
	./${BINARY_NAME} screenshots

copy_screenshots_config:
	cp -f ~/.config/com.thecodingflow/hellocontest.conf ./script/screenshots_config.conf

keybindings: build
	./${BINARY_NAME} keybindings

install_deps:
	go install github.com/mappu/miqt

config:
	xdg-open ~/.config/hamradio/hellocontest.json

config_backup:
	mkdir -p ./.config_backup
	cp -f ~/.config/hamradio/hellocontest.* ./.config_backup

config_restore:
	mkdir -p ~/.config/hamradio
	cp -f ./.config_backup/*.* ~/.config/hamradio

config_clear:
	rm -f ~/.config/hamradio/hellocontest.*
	rm -f ~/.config/com.thecodingflow/hellocontest.conf

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

# the plugin must sit next to linuxdeploy, this is how "--plugin qt" finds it
${LINUXDEPLOY}:
	mkdir -p ${APPIMAGE_TOOLS}
	wget -nv -O $@ https://github.com/linuxdeploy/linuxdeploy/releases/download/${LINUXDEPLOY_VERSION}/linuxdeploy-${ARCH}.AppImage
	chmod a+x $@

${LINUXDEPLOY_PLUGIN_QT}:
	mkdir -p ${APPIMAGE_TOOLS}
	wget -nv -O $@ https://github.com/linuxdeploy/linuxdeploy-plugin-qt/releases/download/${LINUXDEPLOY_PLUGIN_QT_VERSION}/linuxdeploy-plugin-qt-${ARCH}.AppImage
	chmod a+x $@

${APPIMAGETOOL}:
	mkdir -p ${APPIMAGE_TOOLS}
	wget -nv -O $@ https://github.com/AppImage/appimagetool/releases/download/${APPIMAGETOOL_VERSION}/appimagetool-${ARCH}.AppImage
	chmod a+x $@

appimage_tools: ${LINUXDEPLOY} ${LINUXDEPLOY_PLUGIN_QT} ${APPIMAGETOOL}

# linuxdeploy collects the Qt6 libraries and the Qt plugins that the binary needs. An
# AppImage must bring them with it, because the target system has no Qt6 installation.
# The tools are AppImages themselves, APPIMAGE_EXTRACT_AND_RUN runs them without FUSE.
prepare_appimage: appimage_tools
	rm -rf ${APPDIR}
	mkdir -p ${APPDIR}
	env APPIMAGE_EXTRACT_AND_RUN=1 QMAKE=${QMAKE} ${LINUXDEPLOY} \
		--appdir ${APPDIR} \
		--executable ./${BINARY_NAME} \
		--desktop-file ./.assets/${BINARY_NAME}.desktop \
		--icon-file ./.assets/${BINARY_NAME}-256x256.png \
		--icon-filename ${BINARY_NAME} \
		--plugin qt
	# the Qt plugin deploys no platform theme, therefore the application follows neither
	# the colors nor the fonts of the desktop. The theme comes from the build system, so
	# that it fits to the Qt version in the AppImage.
	mkdir -p ${APPDIR}/usr/plugins/platformthemes
	cp ${QT_PLUGINS}/platformthemes/libqgtk3.so ${APPDIR}/usr/plugins/platformthemes/
	cp ./.assets/appimage-qt-hook.sh ${APPDIR}/apprun-hooks/linuxdeploy-plugin-qt-hook.sh

appimage: prepare_appimage
	env APPIMAGE_EXTRACT_AND_RUN=1 ARCH=${ARCH} ${APPIMAGETOOL} ${APPDIR} ${BINARY_NAME}-${VERSION_NUMBER}-${ARCH}.AppImage

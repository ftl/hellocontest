[![CI-Build](https://github.com/ftl/hellocontest/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/ftl/hellocontest/actions/workflows/ci.yml)


# Hello Contest
A simple amateur radio contest log for Linux with a strong bias towards CW on HF. *Hello Contest* is written in Go using the Qt6 mapping [miqt](https://github.com/mappu/miqt) for the UI.

![main window with qso data](https://github.com/ftl/hellocontest/blob/master/docs/screenshots/main_window_filled.png?raw=true)

Here are some highlights:
* Enter your contacts simple and fast using the keyboard.
* Use the popular ["enter sends message"](https://github.com/ftl/hellocontest/wiki/Main-Window#enter-sends-message-aka-esm) method to enter your contacts.
* Show the current rate of QSOs, points, and multis in a [comprehensive graphic](https://github.com/ftl/hellocontest/wiki/QSO-Rate).
* Calculate your points, multis, and [score](https://github.com/ftl/hellocontest/wiki/Score) both per band and overall. The calculation is done using the [conval](https://github.com/ftl/conval) library. You can select the contest definition in the [settings dialog](https://github.com/ftl/hellocontest/wiki/Contest-Settings) (File > Settings).
* Send and receive QTCs (for example in the Worked All Europe DX Contest).
* Export the logbook as [Cabrillo](https://wwrof.org/cabrillo/), [ADIF](http://adif.org), CSV, or call history file.
* Get additional information about the entered callsign from the [DXCC](http://www.country-files.com) and [super check partial](http://supercheckpartial.com) databases, or a call history file.
* Use a call history file from a former contest to predict the exchange for the currently entered callsign.
* Use a dx cluster or a local CW skimmer and show the spotted stations in a [spot list](https://github.com/ftl/hellocontest/wiki/Spots).
* Define different [CW macros](https://github.com/ftl/hellocontest/wiki/Main-Window#cw-macros) for the running and the search&pounce working mode.
* Connect to your transceiver through the [Hamlib network protocol](https://github.com/Hamlib/Hamlib) to keep the band and mode information in sync.
* Use the [Hamlib daemon](https://github.com/Hamlib/Hamlib) or the [cwdaemon](https://github.com/acerion/cwdaemon) to transmit CW macros.
* Show the currently worked station on [F5UII's HamDXMap](https://dxmap.f5uii.net/).

You can find the detailed documentation of all features in the [wiki](https://github.com/ftl/hellocontest/wiki).

## Install
See also the [installation](https://github.com/ftl/hellocontest/wiki/Installation) wiki page for more details.

### AppImage
Download the AppImage of the latest release [here](https://github.com/ftl/hellocontest/releases/latest/).

### Debian, Ubuntu, etc.
Download the Debian package of the latest release [here](https://github.com/ftl/hellocontest/releases/latest/).

### Arch
The latest release of *Hello Contest* is available as AUR package: [hellocontest](https://aur.archlinux.org/packages/hellocontest).

## Build

Build *Hello Contest* using the included Makefile by simply running

```
make
```

The Qt6 build requires the Qt6 development package at compile time — `qt6-base-dev` on Debian/Ubuntu or `qt6-qtbase-devel` on Fedora. Running the resulting binary requires the matching runtime packages. See the [miqt's README](https://github.com/mappu/miqt#linux-native) for more details about the required dependencies.


### Protobuf
*Hello Contest* uses Google's [protocol buffers](https://developers.google.com/protocol-buffers/) to define the data format of the log data stored on disk. The proto definition of the data format resides in the `core/pb` package. This package also contains the generated Go code to access the binary logbook data according to the proto definition. If you make any changes to the proto definition, you need to regenerate this code. The code generation is done using Google's `protoc` compiler for protocol buffers (see Google's [documentation on protocol buffers](https://developers.google.com/protocol-buffers/) for more information about how to install this tool). To run the code generation simply execute

```
make generate
```

This will generate the Go code to access the binary data in the logbook files into the `core/pb` package.

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)

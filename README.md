[![CI-Build](https://github.com/ftl/hellocontest/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/ftl/hellocontest/actions/workflows/ci.yml)


# Hello Contest
A simple amateur radio contest log for Linux, written in Go using [gotk3](https://github.com/gotk3) for the UI. The main focus is on CW contests on HF, but it should also work for SSB or RTTY contests.

![main window with qso data](https://github.com/ftl/hellocontest/blob/master/docs/screenshots/main_window_filled.png?raw=true)

Here are some highlights:
* Enter your contacts simple and fast using the keyboard.
* Use the popular ["enter sends message"](https://github.com/ftl/hellocontest/wiki/Main-Window#enter-sends-message-aka-esm) method to enter your contacts.
* Show the current rate of QSOs, points, and multis in a [comprehensive graphic](https://github.com/ftl/hellocontest/wiki/QSO-Rate).
* Calculate your points, multis, and [score](https://github.com/ftl/hellocontest/wiki/Score) both per band and overall. The calculation is done using the [conval](https://github.com/ftl/conval) library. You can select the contest definition in the [settings dialog](https://github.com/ftl/hellocontest/wiki/Contest-Settings) (File > Settings).
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

The following libraries are required:

* libgtk-3-0
* libgtk-3-dev
* libpango-1.0-0
* libpango1.0-dev
* libpangocairo-1.0-0

### gtk+3.0
To build the software on your system with the gotk3 library, you may need to set a tag with the version number of gtk+3.0 that is installed on your system:

```
# find out the version number
pkg-config --modversion gtk+-3.0

# build Hello Contest (example for gtk+ 3.22.30)
go build -tags gtk_3_22
```

### Protobuf
*Hello Contest* uses Google's [protocol buffers](https://developers.google.com/protocol-buffers/) to define the data format of the log data stored on disk. The proto definition of the data format resides in the `core/pb` package. This package also contains the generated Go code to access the binary logbook data according to the proto definition. If you make any changes to the proto definition, you need to regenerate this code. The code generation is done using Google's `protoc` compiler for protocol buffers (see Google's [documentation on protocol buffers](https://developers.google.com/protocol-buffers/) for more information about how to install this tool). To run the code generation simply execute

```
make generate
```

This will generate the Go code to access the binary data in the logbook files into the `core/pb` package.

### Glade
The UI is defined using a Glade file. This file is automatically integrated into the executable by the Go compiler, using Go's `embed` package (new in 1.16).

## Known Issues
In combination with wayland, the "new contest" dialog does not work properly *sometimes*: you cannot select anything and the entry field for the contest name does not respond to any input. I found no hint so far, what the causes this behavior. If you have problems running *Hello Contest* in combination with wayland, please switch to X11. If you have any hints, what could be the cause of this behavior, please don't hesitate to contact me.

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)

# Hello Contest
A simple HF contest log for Linux, written in Go using [gotk3](https://github.com/gotk3) for the UI. As I'm using this only for CW contests on HF, the program might not be very useful for other modes or higher frequencies. Anyway, here are some highlights:

* Calculate points, multis, and score per band and overall. The calculation rules can be changed in the configuration file.
* Export the logbook in the [Cabrillo](https://wwrof.org/cabrillo/) or [ADIF](http://adif.org) format.
* Get additional information about the entered callsign from the [DXCC](http://www.country-files.com) and [super check partial](http://supercheckpartial.com) databases.
* Use the [cwdaemon](https://github.com/acerion/cwdaemon) to transmit CW macros.
* Define different macros for running and search&pounce working mode.
* Connect to your transceiver through [hamlib](https://github.com/Hamlib/Hamlib) to keep the band and mode information in sync.

I use this little project mainly as training ground to learn how to develop a desktop application in Go and to improve my Go-Fu.

## Build

### gtk+3.0
To build the software on your system with the gotk3 library, you may need to set a tag with the version number of gtk+3.0 that is installed on your system:

```
# find out the version number
pkg-config --modversion gtk+-3.0

# build hellocontest (example for gtk+ 3.22.30)
go build -tags gtk_3_22
```

### Protobuf
Hellocontest uses Google's [protocol buffers](https://developers.google.com/protocol-buffers/) to define the data format of the log data stored on disk. The proto definition of the data format resides in the `core/pb` package. This package also contains the generated Go code to access the binary logbook data according to the proto definition. If you make any changes to the proto definition, you need to regenerate this code. The code generation is done using Google's `protoc` compiler for protocol buffers (see Google's [documentation on protocol buffers](https://developers.google.com/protocol-buffers/) for more information about how to install this tool). To run the code generation simply execute 

```
go generate ./core/pb
```

This will generate the Go code to access the binary data in the logbook files into the `core/pb` package.

### Glade
The UI is defined using a Glade file. This file is integrated into the executable by the tool [go-bindata](https://github.com/kevinburke/go-bindata). To integrate a new version of the glade file into the executable, run

```
go generate ./ui/glade
```

This wil generate the related Go-code into the package `ui/glade`.

## Disclaimer
I develop this software for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome). 

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)

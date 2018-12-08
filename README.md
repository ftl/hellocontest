# Hello Contest
A simple HF contest log for Linux, written in Go using [gotk3](https://github.com/gotk3) for the UI.

I use this little project mainly as training ground to learn how to develop a desktop application in Go and to improve my Go-Fu.

## Disclaimer
I develop this software for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome). 

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)

## Build

### gtk+3.0
To build the software on your system with the gotk3 library, you need to set a tag with the version number of gtk+3.0 that is installed on your system:

```
# find out the version number
pkg-config --modversion gtk+-3.0

# build hellocontest (example for gtk+ 3.22.30)
go build -tags gtk_3_22
```

### Protobuf
To generate the Go-code related to Protobuf, use the following command:

```
go generate ./...
```

This will generate the Go-code into the packge `pb`.

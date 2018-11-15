# Hello Contest
A simple HF contest log for Linux, written in Go using [gotk3](https://github.com/gotk3) for the UI.

I use this little project mainly as training ground to learn how to develop a desktop application in Go and to improve my Go-Fu.

## Disclaimer
I develop this software for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome). 

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)

## Build

### Protobuf
To generate the Go-code related to Protobuf, use the following command:

```
go generate ./...
```

This will generate the Go-code into the packge `pb`.

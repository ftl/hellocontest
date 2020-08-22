// Code generated by go-bindata. DO NOT EDIT.
// sources:
// contest.glade (28.654kB)

package glade

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes  []byte
	info   os.FileInfo
	digest [sha256.Size]byte
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _contestGlade = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x5d\x51\x6f\xdb\x38\x12\x7e\xef\xaf\xe0\xe9\x80\x45\x0f\x8b\x24\x8d\xd3\xec\x1d\x6e\x6d\x2d\x12\x6f\xd3\x5d\x6c\xd3\x62\xe3\xdc\xb5\x6f\x02\x2d\x8d\x25\x36\x14\xa9\x25\xa9\xd8\xbe\x5f\x7f\xa0\x64\x27\x76\x2c\xc9\xa2\x2c\xdb\x92\xe3\x97\xa2\x96\xf5\x51\x9c\x99\x6f\x38\xc3\xe1\x28\xee\xfe\x32\x09\x29\x7a\x04\x21\x09\x67\x3d\xeb\xfc\xf4\x9d\x85\x80\xb9\xdc\x23\xcc\xef\x59\xff\xb9\xbf\x39\xf9\x97\xf5\x8b\xfd\xa6\xfb\xb7\x93\x13\xf4\x11\x18\x08\xac\xc0\x43\x63\xa2\x02\xe4\x53\xec\x01\xba\x38\xed\x74\x4e\x3b\xe8\xe4\xc4\x7e\xd3\x25\x4c\x81\x18\x61\x17\xec\x37\x08\x75\x05\xfc\x15\x13\x01\x12\x51\x32\xec\x59\xbe\x7a\xf8\xd1\x7a\x7e\xd0\xc5\x69\xe7\x9d\x75\x96\xdc\xc7\x87\xdf\xc1\x55\xc8\xa5\x58\xca\x9e\xf5\x51\x3d\x5c\x79\xdf\x63\xa9\x42\x60\xca\x42\xc4\xeb\x59\x32\x02\xf0\x16\x2e\x6a\x14\x42\xdd\x48\xf0\x08\x84\x9a\x22\x86\x43\xe8\x59\x94\x8f\x41\x58\xf6\x65\xf7\x6c\xfe\x45\xf6\x7d\x71\x14\xe9\xfb\x7e\x7a\xb7\xee\xc6\x47\x4c\x63\xb0\xec\xce\xfb\x75\x37\x4a\x05\x91\x43\x98\x2b\x20\x9d\xde\xf9\x3a\x40\x84\x7d\x58\x02\xbc\x98\x4b\xf7\x2c\xd5\x49\x8e\x7a\xa2\x88\x12\x17\x2b\xc2\xd9\x57\xc2\x3c\x3e\x4e\xb5\x14\x62\x32\xff\x9c\xfd\x54\x17\x33\x67\xc4\xdd\x58\x5a\xf6\x0d\xa6\x12\xd6\xcd\x32\x80\x49\x84\x99\x67\xd9\xf7\x22\x5e\x7b\xb3\x22\x8a\x82\x85\x94\xc0\x4c\x52\xac\xf0\x90\x42\xcf\x9a\x82\xb4\xec\xdf\x80\x52\x8e\xfa\x9c\x29\x90\x6a\x65\x18\x37\x20\xd4\x43\x6a\x1a\xcd\xc7\x18\x62\x31\x9b\xbf\x7e\x08\xc5\x2e\x04\x9c\x7a\x20\xce\x66\x80\xb3\x04\xb1\x88\x7e\xba\x7b\x45\x53\xd7\x7c\x92\xea\x46\x70\xae\xf4\x87\xf9\xad\x19\xc6\x26\x92\x0c\x29\x64\x0b\x5b\x45\x9b\x59\x18\x2e\x08\x30\x95\x58\xce\xb2\x1f\x41\x28\xe2\x62\x9a\x09\x5c\x12\x2c\x5b\xb8\x5b\x60\xf1\x35\x16\xcf\xc6\xd7\x17\xac\x45\x4c\x05\x29\xab\x4a\x9a\x3d\xe9\xfc\x89\xff\xae\x20\x9c\xcd\x1c\x58\x7c\x43\xf4\x9c\x5e\x00\x2b\xce\x7e\x13\x09\x32\x97\x15\x3c\x04\x9a\x49\x6c\x47\x4f\xdb\x64\xa8\x58\x82\x13\x33\x0f\x04\x25\xac\x8c\x14\x8b\xbe\x21\xe3\x61\xb8\x6a\xdf\x22\x15\xcf\xd6\xce\x14\x97\xa3\xe1\x8d\xb4\xbc\xa9\xa6\x51\x2e\x67\x8a\x05\x5b\xe5\xce\x67\x18\xe7\x08\xb7\xb1\x80\x75\x08\x99\x35\x46\x4a\x2b\xfb\x33\x8c\x4f\x4f\x4f\xab\x0c\x60\x48\xa6\xd9\x20\x0b\x81\x25\xeb\xdb\x02\x6b\xd4\x63\xa9\x2f\x11\xb0\x96\x9a\x4a\x4f\xfd\x55\xd9\x6a\x80\x1f\xe1\x4a\xb6\xd4\x5a\x7a\xf2\x08\xcb\x03\x35\xd8\x00\x22\x2c\xb0\xe2\x62\xd9\x72\x72\x7e\x59\x9b\xef\xbc\xd9\x96\xdb\x3b\xbd\x3f\x4c\x22\x2e\x54\x1f\x0f\x05\xa1\x94\x37\x5b\x59\x26\x69\x49\x2a\x17\x9a\x0b\x76\xa0\x1e\x50\x64\xd3\xab\x5f\x7f\xbf\x39\x38\x7b\x6a\xa1\x0e\xd4\x96\x65\x56\xb3\x4e\xb3\x0d\xba\x77\xe6\xff\x19\x13\xd5\x6c\x15\x99\x70\x5e\x4b\xd3\x7c\xa6\xe7\x03\x73\x40\xd9\x80\xcc\x9b\xab\x6f\xa6\x7f\x03\x1a\xb5\x70\x33\xad\xa7\xdd\x92\xcd\x74\x8e\x86\x37\xd2\xf2\xa6\x9a\x46\xb5\xac\x24\x5a\xb2\xab\x21\x8f\x5b\xba\x94\xd8\xc9\xdc\x5f\xf3\xc2\x91\x75\x63\x37\xc2\xee\x03\x61\x7e\x71\x95\x6f\x5e\xec\x2d\x2e\xf1\xbd\x00\x8d\x08\xa5\x66\xc5\xc4\x88\x4b\x92\xd6\x3f\x57\xea\xef\x73\x09\x56\xa6\xbb\x22\x67\x99\xf2\xe8\xc0\x15\x9c\x52\xf0\x16\x4b\xe4\x94\xfb\xff\x25\x30\xee\x73\xa6\x30\x61\x20\xd6\x54\x4b\xc7\xc4\x53\x81\x23\xe0\xaf\x18\xa4\xb2\xec\xcb\xcb\x95\xb3\x85\x3c\xe4\xa6\x75\x56\x13\x98\x0c\xb0\xc7\xc7\x8e\x5e\xd9\x2c\x9b\xb0\x4d\xab\xb3\xf7\x02\x40\x2b\x69\x49\x63\x25\xe2\x49\x00\xc4\x0f\xd4\xb3\xb6\xce\xdf\xe5\x18\x78\x03\x8d\xad\xd1\x9a\x71\x38\x2a\x3c\xe1\x28\x9c\x6d\x55\x60\x20\x13\x5a\x3a\x11\xa7\xc4\x9d\x5a\x36\xc3\x2a\x16\xd9\x67\x00\xb9\xcf\xde\x7c\x88\x00\xb0\x07\x42\x3a\x2e\x25\xee\x03\x4e\xd4\x6e\xac\x3b\x60\x1a\xe8\x48\xc0\xc2\x0d\x2a\xe0\x65\xc0\xc7\x4e\xaa\x46\x10\x55\x8c\x37\x9b\x80\x2f\x88\xe7\xe8\x55\x5b\x16\x1f\xa9\xac\x19\x46\x09\x80\xf9\x30\x25\x73\x89\xe4\xb4\x93\x61\x7a\x92\x7c\xd4\xfb\x15\x0a\x6e\xba\xb6\x95\xca\x2b\xb4\xa7\x0d\x9e\x30\x73\x77\x7b\xbe\x72\xd6\xc4\xf0\x60\xb2\x32\x6d\x12\x1d\x56\xce\x50\xe7\x02\xd4\x14\x1d\x14\x56\xb1\x1c\xce\x8f\xcf\xe4\xd3\xc7\xed\xae\xeb\x46\xc1\x35\xc4\xc2\x27\xcc\xa1\x30\xca\x38\x22\x5e\x0b\x13\x7a\x39\xae\x80\x93\x0a\x8b\x2a\x38\xd0\xec\x30\x46\x29\x1e\x59\xf6\x4f\x86\xa0\x21\x57\x8a\x87\x06\xb8\xd2\xe7\xae\x59\x60\x19\x61\x97\x30\xdf\xb2\x3b\x79\x94\x6c\x7a\xca\x85\xdd\x87\x59\x7a\x00\xcc\x33\x77\xc5\x5c\xb9\xeb\x71\xc5\x8f\x82\x78\xa9\x17\x02\x53\x62\x9a\x7c\x6c\xaa\x17\xae\xb4\x82\xac\x41\xcd\x9c\xd0\x14\x96\x78\x45\xe9\x64\xf3\x85\x57\x94\xc6\x09\x3e\x76\x9e\xc8\x5d\x1a\xe5\x72\x1a\x87\xac\x0c\xb0\x64\xc2\xf9\x29\xad\x09\x24\xe1\x4f\xff\x37\xab\xd2\xd7\x88\x6c\x31\xbf\x78\x71\x1f\x00\x11\xff\x36\x19\x2b\xdd\x5a\xb8\x01\xd6\xa9\x4f\xee\x3a\x96\x85\x9c\x60\x4a\xfc\xfc\xed\x13\x2a\xd8\x5e\x66\xaf\x4a\x99\xa2\xc2\x48\x39\x58\x29\xac\x53\x3b\xa3\x3c\x5e\xf1\xe8\x09\x98\x13\xc3\x51\xf6\xe2\x81\x36\x2d\x85\x7d\xd0\xcb\x47\x4a\x24\x17\x53\x2a\x89\xcf\xd2\x4b\xeb\x67\xad\xff\xb5\xec\x25\xd4\xae\x37\x2f\xa6\x50\xc5\x39\x55\x24\x72\x14\x4c\x54\x3e\x29\x51\x7f\x26\x93\xc9\xd0\x0b\x5d\x57\xf9\xc3\xeb\x81\x77\x46\xc1\x02\x26\x55\xa6\x60\x8e\x53\xe6\xc6\x3b\xb4\x7d\xda\x2a\x6d\xb2\x3b\x88\xb8\x50\x66\xcc\x7d\x09\x3c\x10\xf2\xa6\x12\x99\x0c\x1c\xe2\x89\x43\x81\xf9\xda\x8e\x17\x95\x17\x64\x23\xa4\x7e\x64\x75\x74\x29\x5f\xbb\x1b\xdc\x9b\x8c\x49\x58\x14\x2b\x27\x8a\x45\xc4\x25\x58\xb6\x47\x7c\xa2\xe4\xce\x7c\xd5\x48\xfc\x66\x84\x8b\xc4\x7d\x3e\xc7\xe1\x10\x44\x05\xbf\x5b\x00\x1e\x88\xdf\xa5\x12\x55\x76\x9f\xdc\x94\x30\x0b\x59\xca\x01\xfe\xde\x1e\xfa\xe7\xe6\xfa\xd9\xf6\x68\x02\xfd\xc3\x69\x95\x98\xb3\x84\x6a\x3d\xf1\x6f\xa7\xb3\x68\x83\xde\xf6\x95\xa0\x27\xe2\x1f\xc7\xb0\xb3\xdd\xb0\x83\x50\x17\xbb\x2e\x50\x48\x7a\x3c\xd0\x03\x4c\x7b\x96\xb0\x90\x4e\x56\x31\xed\x59\xbe\xc0\xc3\x93\xd4\xda\x28\xe4\x1e\x19\x11\x10\x9a\xb9\xbf\xfe\xe1\xf4\xbf\x7c\xbe\xbf\xfb\xf2\xc9\xb9\xbd\x1a\xfc\xb1\x5a\x25\x6d\x4a\x38\x2b\xdc\xa0\x6d\xd9\x9f\xab\xc4\xb2\x25\xd4\x4e\xfc\x59\x02\x93\x44\x91\xc7\x2a\xc7\x0f\x9b\x6c\xe4\xcb\x2c\x06\xc7\x10\xb8\x87\x10\xb8\x63\x97\x79\x59\x79\xca\xea\x98\x6d\x78\xe5\xe9\x76\x6a\x54\x76\x02\x4a\x49\x24\xc9\xff\xc0\xb2\x93\x3a\xff\xb1\x64\xd5\x18\x06\x86\x20\x25\xf6\x21\xbd\xd2\x3e\x1e\xa6\xb3\x3f\x50\x56\x14\x54\x84\x72\xab\x48\xff\xdc\x2d\x93\xfa\x3c\x1c\xf2\x6b\x3e\xb9\x4f\xe2\x87\x26\xd4\x10\x33\x2f\xb9\xda\x54\x36\xad\x0d\xc3\xd7\x98\x79\x69\x3a\xfe\xe3\xb0\x38\x1d\x5f\x49\x24\x87\xcf\x89\x64\xc4\xa3\x38\xda\x6f\x0e\x59\xb9\x7c\xb9\xe3\xe5\x68\x95\x44\x21\xf7\xa0\xdd\x24\xba\xe5\x1e\xcc\xf6\x74\xa1\x21\x89\xc2\x46\x91\xc8\x68\x11\xda\x23\x89\xae\x63\xa5\xe6\xfd\x2c\x02\x24\xa8\xd9\x85\x4d\xc2\xcb\x9d\x1e\xa7\xf1\xed\x64\x02\x5c\x20\x8f\x20\x1d\x0f\x46\x38\xa6\x6a\x0b\x65\x8a\x44\x11\xe8\xed\x0f\x54\xfd\xfc\x61\xd0\xff\xc1\x57\x3f\x17\x70\xba\x66\x02\x1a\xe5\x7c\x0d\x21\x20\xe5\x7e\x0d\xf4\xfb\xc4\xfd\x23\xf9\x12\x35\x20\x15\x00\xfa\x73\xf0\x65\x46\x41\xa6\x40\xb4\x84\x84\xfb\x2c\xaf\x7e\x73\x03\xcc\x7c\x30\xad\xc7\x2c\xc2\x0e\xa1\xc0\xfa\x61\x92\x0a\xb4\x9b\x5d\x67\xa9\xaa\xca\xb7\x7e\x50\xe0\xda\x35\x73\xd7\xa8\x28\xd4\x8c\x52\x62\x72\xc0\x55\x89\xbe\x2b\xc8\xd6\x33\x38\x3d\x1b\x3b\x92\x78\x9f\x0b\x70\xf3\x9b\x2d\x9f\xda\x26\x73\x0e\x0e\xea\x6f\x9b\x7c\x80\x29\x88\x63\xdb\xe4\x5e\xda\x26\x73\x37\x65\xeb\xda\x26\x4b\x03\xf5\xe3\x02\x1e\x72\x1f\x18\xf0\xb5\xef\x06\x55\x48\x91\x47\xe7\x35\x64\xc8\x37\x46\x25\x8e\x43\x4d\x90\x07\xc0\x3c\x74\x8b\x85\x2b\x38\x7a\x7b\x73\x6e\x58\x6e\xb8\x39\x7f\xae\x37\x24\x6f\xe6\x80\xb7\xf5\xa2\x42\x5b\x0a\xe5\x4b\x84\xed\xd4\x41\x58\xa3\x72\xca\xeb\x20\x6c\xc7\x94\xb0\x9d\x16\x11\x76\xc7\xfb\xbf\x25\xc2\x5e\xd4\x41\x58\xa3\x46\x84\xd7\x41\xd8\x0b\x53\xc2\x5e\xb4\x88\xb0\x3b\x6e\x43\x5e\x22\xec\xfb\x3a\x08\x6b\xd4\x06\xf0\x3a\x08\x7b\x69\x4a\xd8\xf7\x2d\x22\x6c\xc1\x02\xb5\xe5\x2a\xc5\xe8\xbc\x6c\x6d\x62\x1f\x85\x85\x92\x6f\x99\xb7\xe0\xcc\x31\xf7\xb0\xbb\xc0\xd5\xb7\x6d\xfa\xce\xd1\xf4\x7b\x7d\x5b\x66\x8f\xa6\xbf\x38\x9a\xbe\x36\xd3\x57\x69\x71\xd9\xa3\xe9\xdf\x1f\x4d\x5f\x9b\xe9\xcd\x1b\xd6\x77\x6e\xfa\xc5\xe4\x54\x2a\x1e\xd5\x91\x9e\x1a\x15\xb2\xd3\x33\x8d\xe7\x47\x37\xfd\x30\x63\x17\x99\xad\xe2\x51\x72\x1c\xdc\xff\x9a\x7e\x1d\x12\x29\x09\x67\x55\xd2\xdc\xcb\x16\xa5\xb9\x3b\xa6\x7e\xa3\x1b\x42\x83\x59\x0f\xa7\x71\x33\x71\x50\xea\x38\x26\xd3\x88\xb9\x1e\x3d\x8e\xc2\x9d\x2d\xb6\x95\x5b\xdc\x8f\xec\x59\x65\x4f\xc1\x5f\x25\xc9\x84\x55\xfd\xe3\x5b\xf9\xd4\x19\x44\x00\x9e\x51\x27\xfd\xf7\x58\x2a\x32\x9a\x5a\x76\x72\x56\xb5\x33\xde\x55\x6e\x02\xdc\x31\xef\x06\x11\x61\x4b\x41\x5b\x2b\xb8\xc9\x39\xdb\xda\x70\xd7\xff\x8a\x12\x96\xcc\x7a\x49\xa5\xd1\xfb\x81\x73\xa6\xbb\xc0\x94\xd9\x5b\x45\xd5\x17\xca\xc5\x77\x12\xcd\xb3\xfb\x79\xb7\x82\x11\x72\xe5\x9d\x44\x23\xb4\xf9\xfb\x83\x2f\x06\xc0\x0b\x3f\x11\xf5\xe2\x37\xa3\x8c\x72\xbd\x38\x04\x41\xdc\x12\x1c\x5a\x49\x64\xe4\x01\xbd\xbf\x78\x8c\x54\x35\xe4\x39\x4b\x6d\x10\x46\xde\xb0\x8d\x28\xa7\x33\xf6\xfb\x6f\xcd\xdf\x91\xd6\x42\xbd\x16\xb5\x0a\xe5\xc8\xbb\xbe\x55\x68\x59\xc6\x85\x2f\x9f\xbf\xe8\x9e\x2d\xfc\x2a\xdf\xff\x03\x00\x00\xff\xff\x28\x7c\x3b\x06\xee\x6f\x00\x00")

func contestGladeBytes() ([]byte, error) {
	return bindataRead(
		_contestGlade,
		"contest.glade",
	)
}

func contestGlade() (*asset, error) {
	bytes, err := contestGladeBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "contest.glade", size: 28654, mode: os.FileMode(0664), modTime: time.Unix(1598083296, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x93, 0x24, 0x34, 0x7b, 0x97, 0x73, 0x2f, 0x8e, 0x86, 0xa6, 0x5b, 0xcc, 0x31, 0x87, 0xfc, 0x1b, 0xa5, 0x21, 0x6d, 0x10, 0x54, 0x9a, 0xd5, 0x18, 0x7f, 0x48, 0x3c, 0x2f, 0xb3, 0x20, 0xf6, 0xdc}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetString returns the asset contents as a string (instead of a []byte).
func AssetString(name string) (string, error) {
	data, err := Asset(name)
	return string(data), err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// MustAssetString is like AssetString but panics when Asset would return an
// error. It simplifies safe initialization of global variables.
func MustAssetString(name string) string {
	return string(MustAsset(name))
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetDigest returns the digest of the file with the given name. It returns an
// error if the asset could not be found or the digest could not be loaded.
func AssetDigest(name string) ([sha256.Size]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s can't read by error: %v", name, err)
		}
		return a.digest, nil
	}
	return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s not found", name)
}

// Digests returns a map of all known files and their checksums.
func Digests() (map[string][sha256.Size]byte, error) {
	mp := make(map[string][sha256.Size]byte, len(_bindata))
	for name := range _bindata {
		a, err := _bindata[name]()
		if err != nil {
			return nil, err
		}
		mp[name] = a.digest
	}
	return mp, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"contest.glade": contestGlade,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"},
// AssetDir("data/img") would return []string{"a.png", "b.png"},
// AssetDir("foo.txt") and AssetDir("notexist") would return an error, and
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"contest.glade": &bintree{contestGlade, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}

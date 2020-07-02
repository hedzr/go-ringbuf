/*
 * Copyright © 2020 Hedzr Yeh.
 */

package untitled

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

// https://go101.org/article/unsafe.html
// https://blog.gopheracademy.com/advent-2017/unsafe-pointer-and-system-calls/
func Test_unsafe_pointer(t *testing.T) {
	// var st string = "hello"
	//
	// var s string
	// hdr := (*reflect.StringHeader)(unsafe.Pointer(&s)) // case 1
	// hdr.Data = uintptr(unsafe.Pointer(&st))              // case 6 (this case)
	// hdr.Len = len(st)
	//
	// // var str string
	// // str = *(*string)(unsafe.Pointer(&hdr))
	// t.Logf("str: %v", *(*string)(unsafe.Pointer(&hdr)))

	str := strings.Join([]string{"Go", "land"}, "")
	s := String2ByteSlice(str)
	t.Logf("%s\n", s) // Goland
	s[5] = 'g'
	t.Log(str) // Golang
}

func String2ByteSlice(str string) (bs []byte) {
	strHdr := (*reflect.StringHeader)(unsafe.Pointer(&str))
	sliceHdr := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	sliceHdr.Data = strHdr.Data
	sliceHdr.Len = strHdr.Len
	sliceHdr.Cap = strHdr.Len

	// The KeepAlive call is unessential now.
	// runtime.KeepAlive(&str)
	return
}

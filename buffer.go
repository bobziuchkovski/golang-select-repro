// Copyright (c) 2016 Bob Ziuchkovski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"errors"
	"unicode/utf8"
)

var errBufferWrite = errors.New("write attempted on a buffer that was previously read")

type Buffer interface {
	Bytes() []byte
	Len() int
	Write(value []byte)
	WriteByte(c byte)
	WriteRune(r rune)
	WriteString(s string)
}

type buffer struct {
	bytes   []byte
	runebuf [utf8.MaxRune]byte
	read    bool
}

func NewBuffer() Buffer {
	return &buffer{
		bytes: make([]byte, 64),
	}
}

func NewBufferFrom(bytes []byte) Buffer {
	return &buffer{
		bytes: bytes[:0],
	}
}

func (b *buffer) Bytes() []byte {
	b.read = true
	return b.bytes
}

func (b *buffer) Len() int {
	return len(b.bytes)
}

func (b *buffer) WriteByte(value byte) {
	if b.read {
		panic(errBufferWrite)
	}

	b.ensureCapacity(1)
	b.bytes = append(b.bytes, value)
}

func (b *buffer) WriteRune(value rune) {
	if b.read {
		panic(errBufferWrite)
	}

	if value < utf8.RuneSelf {
		b.WriteByte(byte(value))
		return
	}

	size := utf8.EncodeRune(b.runebuf[:], value)
	b.Write(b.runebuf[:size])
}

func (b *buffer) WriteString(value string) {
	if b.read {
		panic(errBufferWrite)
	}

	origlen := len(b.bytes)
	b.ensureCapacity(len(value))
	b.bytes = b.bytes[:origlen+len(value)]
	copy(b.bytes[origlen:], value)
}

func (b *buffer) Write(value []byte) {
	if b.read {
		panic(errBufferWrite)
	}

	origlen := len(b.bytes)
	b.ensureCapacity(len(value))
	b.bytes = b.bytes[:origlen+len(value)]
	copy(b.bytes[origlen:], value)
}

func (b *buffer) ensureCapacity(size int) {
	curlen := len(b.bytes)
	curcap := cap(b.bytes)
	if curlen+size > curcap {
		new := make([]byte, curlen, 2*curcap+size)
		copy(new, b.bytes)
		b.bytes = new
	}
}

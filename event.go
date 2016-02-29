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
	"time"
)

type Event struct {
	Time    time.Time
	Level   Level
	Context Context
	Frames  []uintptr
	Error   error
	Message string
}

func (e *Event) Clone() *Event {
	frames := make([]uintptr, len(e.Frames))
	copy(frames, e.Frames)
	return &Event{
		Time:    e.Time,
		Level:   e.Level,
		Context: e.Context,
		Frames:  frames,
		Error:   e.Error,
		Message: e.Message,
	}
}

func (e *Event) ErrorType() string {
	return ""
}

func (e *Event) Source() *Frame {
	return nil
}

func (e *Event) Stack() []*Frame {
	return nil
}

func getRecoveryFrames(skip int, depth int) []uintptr {
	return nil
}

func getFrames(skip int, depth int) []uintptr {
	return nil
}

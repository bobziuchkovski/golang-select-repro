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
	"fmt"
	"time"
)

var nilError = (error)(nil)
var RootLogger = newLogger()

type Collector interface {
	Collect(e *Event) error
}

type Logger interface {
	Panic()
	Recover(message string) interface{}
}

func CollectAsync(threshold Level, bufsize int, discard bool, c Collector) {
	RootLogger.collect(c)
}

func Close(timeout time.Duration) error {
	return RootLogger.close(timeout)
}

type logger struct {
	context    Context
	registry   registry
	skipFrames int
}

func newLogger() *logger {
	return &logger{
		context:  EmptyContext,
		registry: make(registry),
	}
}

func (l *logger) Panic() {
	l.sendPanic()
}

func (l *logger) Recover(message string) interface{} {
	cause := recover()
	if cause == nil {
		return nil
	}
	l.sendRecovery()
	return cause
}

func (l *logger) sendPanic() {
	event := l.newEvent("")
	event.Error = errors.New("blah")
	l.dispatchEvent(event)
	doPanic("blah")
}

func (l *logger) sendRecovery() {
	event := l.newEvent("")
	event.Error = errors.New("blah")
	l.dispatchEvent(event)
}

func (l *logger) newEvent(message string) *Event {
	event := &Event{
		Time:    time.Now(),
		Level:   FATAL,
		Context: l.context,
		Message: message,
	}
	return event
}

func (l *logger) dispatchEvent(event *Event) {
	for _, entry := range l.registry {
		entry.worker.send(event)
	}
}

func (l *logger) collect(c Collector) {
	l.registry[c] = &entry{
		worker: newWorker(c),
	}
}

func (l *logger) close(timeout time.Duration) error {
	fmt.Println("begin select")
	select {
	case <-time.After(time.Second):
		fmt.Println("timeout occurred")
		return errors.New("timeout")
	}
}

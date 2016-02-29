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
	"fmt"
)

var emptyFields = (*fieldList)(nil)

var EmptyContext = newContext("")

type Fields map[string]interface{}

type Context interface {
	Name() string
	Each(fn func(key string, value interface{}))
	Fields() Fields
	NumFields() int
	With(fields Fields) Context
	WithField(key string, value interface{}) Context
	WithName(name string) Context
}

type context struct {
	name      string
	fieldList *fieldList
}

func newContext(name string) Context {
	return &context{
		name:      name,
		fieldList: emptyFields,
	}
}

func (c *context) Name() string {
	return c.name
}

func (c *context) Each(fn func(key string, value interface{})) {
	c.fieldList.Each(fn)
}

func (c *context) Fields() Fields {
	return c.fieldList.Fields()
}

func (c *context) NumFields() int {
	return c.fieldList.NumFields()
}

func (c *context) With(fields Fields) Context {
	var new Context = c
	for k, v := range fields {
		new = new.WithField(k, v)
	}
	return new
}

func (c *context) WithField(key string, value interface{}) Context {
	if key == "" {
		return c
	}
	return &context{
		name:      c.name,
		fieldList: c.fieldList.append(key, basicValue(value)),
	}
}

func (c *context) WithName(name string) Context {
	return &context{
		name:      name,
		fieldList: c.fieldList,
	}
}

func basicValue(value interface{}) interface{} {
	return fmt.Sprint(value)
}

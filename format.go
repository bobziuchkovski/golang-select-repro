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
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	none   = 0
	red    = 31
	green  = 32
	yellow = 33
	blue   = 34
)

var (
	HumanSource        = FormatFormatter("%v:%v", FormatShortFile, FormatLine)
	HumanMessage       = Join(" ", FormatMessage, FormatHumanContext)
	HumanSourceMessage = FormatFormatter("%v: %v", HumanSource, HumanMessage)
	HumanReadable      = FormatFormatter("%v %v %v", TimeFormatter(time.Stamp), FormatLevel, HumanSourceMessage)

	JsonMessage       = Join(" ", FormatMessage, FormatJsonContext)
	JsonSourceMessage = FormatFormatter("%v: %v", HumanSource, JsonMessage)
)

type Formatter func(buffer Buffer, event *Event)

func Render(formatter Formatter, event *Event) []byte {
	buffer := NewBuffer()
	formatter(buffer, event)
	return buffer.Bytes()
}

func Join(sep string, formatters ...Formatter) Formatter {
	return func(buffer Buffer, event *Event) {
		for i, formatter := range formatters {
			origlen := buffer.Len()
			formatter(buffer, event)
			if buffer.Len() > origlen && i < len(formatters)-1 {
				buffer.WriteString(sep)
			}
		}
	}
}

func FormatFormatter(format string, formatters ...Formatter) Formatter {
	formatterIdx := 0
	segments := splitFormat(format)
	chain := make([]Formatter, len(segments))
	for i, seg := range segments {
		switch {
		case seg == "%v" && formatterIdx < len(formatters):
			chain[i] = formatters[formatterIdx]
			formatterIdx++
		case seg == "%v":
			chain[i] = Literal("%!v(MISSING)")
		default:
			chain[i] = Literal(seg)
		}
	}

	return func(buffer Buffer, event *Event) {
		for _, formatter := range chain {
			formatter(buffer, event)
		}
	}
}

func splitFormat(format string) []string {
	var (
		segments []string
		segstart int
		lastrune rune
	)

	runes := []rune(format)
	for i, r := range runes {
		switch {
		case lastrune == '%' && r == '%':
			lastrune = 0
		case lastrune == '%' && r == 'v':
			segend := i - 1
			if segstart != segend {
				segments = append(segments, string(runes[segstart:segend]))
			}
			segments = append(segments, "%v")
			segstart = i + 1
			lastrune = r
		default:
			lastrune = r
		}
	}

	if segstart < len(runes) {
		segments = append(segments, string(runes[segstart:]))
	}
	return segments
}

func Colorize(formatter Formatter) Formatter {
	return func(buffer Buffer, event *Event) {
		buffer.WriteString(fmt.Sprintf("\x1b[%dm", colorFor(event.Level)))
		formatter(buffer, event)
		buffer.WriteString("\x1b[0m")
	}
}

func colorFor(lvl Level) int {
	switch lvl {
	case DEBUG:
		return blue
	case INFO:
		return green
	case WARN:
		return yellow
	case ERROR, FATAL:
		return red
	default:
		return none
	}
}

func HostFormatter() Formatter {
	fqdn := false
	return hostFormatter(fqdn)
}

func FQDNFormatter() Formatter {
	fqdn := true
	return hostFormatter(fqdn)
}

func hostFormatter(fqdn bool) Formatter {
	h := host(fqdn)

	return func(buffer Buffer, event *Event) {
		if h == "" {
			h = host(fqdn)
		}
		if h == "" {
			buffer.WriteString("unknown")
		} else {
			buffer.WriteString(h)
		}
	}
}

func host(fqdn bool) string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	if fqdn {
		return name
	}
	idx := strings.Index(name, ".")
	if idx != -1 {
		name = name[:idx]
	}
	return name
}

func Literal(s string) Formatter {
	return func(buffer Buffer, event *Event) {
		buffer.WriteString(s)
	}
}

func TimeFormatter(timeFormat string) Formatter {
	return func(buffer Buffer, event *Event) {
		buffer.WriteString(event.Time.Format(timeFormat))
	}
}

func FormatLevel(buffer Buffer, event *Event) {
	buffer.WriteString(event.Level.String())
}

func FormatPackage(buffer Buffer, event *Event) {
	buffer.WriteString(event.Source().Package())
}

func FormatFile(buffer Buffer, event *Event) {
	buffer.WriteString(event.Source().File())
}

func FormatShortFile(buffer Buffer, event *Event) {
	short := event.Source().File()
	idx := strings.LastIndex(short, "/")
	if idx != -1 {
		short = short[idx+1:]
	}
	buffer.WriteString(short)
}

func FormatLine(buffer Buffer, event *Event) {
	buffer.WriteString(fmt.Sprintf("%d", event.Source().Line()))
}

func FormatRawMessage(buffer Buffer, event *Event) {
	buffer.WriteString(event.Message)
}

func FormatAsciiMessage(buffer Buffer, event *Event) {
	trimmed := strings.TrimSpace(event.Message)
	ascii := strconv.QuoteToASCII(trimmed)
	buffer.WriteString(ascii[1 : len(ascii)-1])
}

func FormatMessage(buffer Buffer, event *Event) {
	trimmed := strings.TrimSpace(event.Message)
	for _, r := range []rune(trimmed) {
		switch {
		case r == ' ':
			buffer.WriteRune(r)
		case unicode.IsControl(r), unicode.IsSpace(r):
			quoted := strconv.QuoteRune(r)
			buffer.WriteString(quoted[1 : len(quoted)-1])
		default:
			buffer.WriteRune(r)
		}
	}
}

func FormatHumanContext(buffer Buffer, event *Event) {
	fields := event.Context.Fields()

	// Sort field keys for predictable output ordering
	var sortedKeys []string
	for k := range fields {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for i, k := range sortedKeys {
		buffer.WriteString(k)
		buffer.WriteRune('=')
		buffer.Write(humanValue(fields[k]))
		if i < len(sortedKeys)-1 {
			buffer.WriteRune(' ')
		}
	}
}

func humanValue(v interface{}) []byte {
	s, ok := v.(string)
	if !ok {
		s = fmt.Sprint(v)
	}
	return humanString(s)
}

func humanString(s string) []byte {
	special := func(r rune) bool {
		switch {
		case r == '"', r == '\'', r == '\\', r == 0:
			return true
		case unicode.IsLetter(r), unicode.IsNumber(r), unicode.IsPunct(r), unicode.IsSymbol(r):
			return false
		default:
			return true
		}
	}
	if strings.IndexFunc(s, special) >= 0 {
		return []byte(strconv.Quote(s))
	}
	return []byte(s)
}

func FormatJsonContext(buffer Buffer, event *Event) {
	fields := event.Context.Fields()
	marshaled, _ := json.Marshal(fields)
	buffer.Write(marshaled)
}

func FormatStructuredContext(buffer Buffer, event *Event) {
	formatStructuredContext(buffer, event.Context)
}

func formatStructuredContext(buffer Buffer, context Context) {
	count := 0
	numfields := context.NumFields()

	// We iterate the key/val pairs directly so we see duplicate keys.
	// We're exploiting this for Loggly tags, but it's ugly...
	context.Each(func(name string, value interface{}) {
		if !validStructuredKey(name) {
			return
		}

		formatStructuredPair(buffer, name, value)
		if count < numfields-1 {
			buffer.WriteRune(' ')
		}
		count++
	})
}

// These restrictions are imposed by RFC5424
func validStructuredKey(name string) bool {
	if len(name) > 32 {
		return false
	}
	for _, r := range []rune(name) {
		switch {
		case r <= 32:
			return false
		case r >= 127:
			return false
		case r == '=', r == ']', r == '"':
			return false
		}
	}
	return true
}

func formatStructuredPair(buffer Buffer, name string, value interface{}) {
	buffer.WriteString(name)
	buffer.WriteRune('=')
	buffer.WriteRune('"')
	formatStructuredValue(buffer, value)
	buffer.WriteRune('"')
}

// See Section 6.3.3 of RFC 5424 for details on the character escapes
// XXX: Do we still need to send an escape if the value is already escaped?
func formatStructuredValue(buffer Buffer, v interface{}) {
	s, ok := v.(string)
	if !ok {
		s = fmt.Sprint(v)
	}

	for _, r := range []rune(s) {
		switch r {
		case '\'':
			buffer.WriteRune('\\')
			buffer.WriteRune('\'')
		case '\\':
			buffer.WriteRune('\\')
			buffer.WriteRune('\\')
		case ']':
			buffer.WriteRune('\\')
			buffer.WriteRune(']')
		default:
			buffer.WriteRune(r)
		}
	}
}

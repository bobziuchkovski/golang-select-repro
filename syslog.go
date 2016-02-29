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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

type Facility uint
type severity uint
type priority uint

const (
	KERN Facility = iota
	USER
	MAIL
	DAEMON
	AUTH
	SYSLOG
	LPR
	NEWS
	UUCP
	CRON
	AUTHPRIV
	FTP
	NTP
	AUDIT
	ALERT
	_
	LOCAL0
	LOCAL1
	LOCAL2
	LOCAL3
	LOCAL4
	LOCAL5
	LOCAL6
	LOCAL7
)

const (
	sEMERGENCY severity = iota
	sALERT
	sCRITICAL
	sERROR
	sWARN
	sNOTICE
	sINFO
	sDEBUG
)

var rfc5424BOM = []byte{0xef, 0xbb, 0xbf}
var syslogSockets = []string{"/dev/billet", "/var/run/billet", "/var/run/syslog"}

const (
	rfc5424Time     = "2006-01-02T15:04:05.000000-07:00"
	rfc5424Version  = "1"
	ourStructuredId = "TODO@12345"
	syslogNil       = "-"
)

type Syslog struct {
	Facility Facility
	App      string

	// Optional Socket config.  Defaults to a local unix socket, if available.
	Network string
	Address string
	TLS     *tls.Config

	// Optional extras
	MessageFormatter Formatter // Default: FormatMessage
}

func (s Syslog) New() Collector {
	if s.App == "" {
		panic("billet/target: Syslog app cannot be empty")
	}

	var err error
	if s.Network == "" || s.Address == "" {
		s.Network, s.Address, err = localSyslog()
	}
	if err != nil {
		return nil
	}

	return Socket{
		Formatter: HumanMessage,
		Network:   s.Network,
		Address:   s.Address,
	}.New()
}

func rfc3164Formatter(facility Facility, app string, local bool, msgFormatter Formatter) Formatter {
	if msgFormatter == nil {
		msgFormatter = FormatMessage
	}
	if local {
		return FormatFormatter("%v%v %v: %v\n",
			priFormatter(facility), TimeFormatter(time.Stamp), procIdFormatter(app), msgFormatter)
	}

	// Same as above but with hostname included and using RFC3339 time format
	return FormatFormatter("%v%v %v %v: %v\n",
		priFormatter(facility), TimeFormatter(time.RFC3339), HostFormatter(), procIdFormatter(app), msgFormatter)

}

type StructuredSyslog struct {
	// Required
	Facility Facility
	App      string

	// Optional Socket config.  Defaults to a local unix socket, if available.
	Network string
	Address string
	TLS     *tls.Config

	// Optional extras
	StructuredTransformer ContextTransformer // Render the structured data based on a transformed context; useful for tagging/enrichment
	StructuredId          string             // Default: <TODO: our structured ID>
	MessageFormatter      Formatter          // Default: FormatMessage
	WriteBOM              bool               // RFC5424 requires a byte-order mark (BOM), but many syslog servers don't care, and some don't understand it
}

func (s StructuredSyslog) New() Collector {
	if s.App == "" {
		panic("billet/target: StructuredSyslog app cannot be empty")
	}

	var err error
	if s.Network == "" || s.Address == "" {
		s.Network, s.Address, err = localSyslog()
	}
	if err != nil {
		return nil
	}

	return Socket{
		Formatter: rfc5424Formatter(s.Facility, s.App, s.MessageFormatter, s.StructuredId, s.StructuredTransformer, s.WriteBOM),
		Network:   s.Network,
		Address:   s.Address,
	}.New()
}

func rfc5424Formatter(facility Facility, app string, msgFormatter Formatter, structId string, transformer ContextTransformer, writeBom bool) Formatter {
	msgid := syslogNil
	bomFormatter := Literal("")
	if writeBom {
		bomFormatter = formatBOM
	}
	if structId == "" {
		structId = ourStructuredId
	}
	if msgFormatter == nil {
		msgFormatter = FormatMessage
	}
	return FormatFormatter("%v%v %v %v %v %v %v [%v] %v%v\n",
		priFormatter(facility), Literal(rfc5424Version), TimeFormatter(rfc5424Time),
		FQDNFormatter(), Literal(app), procIdFormatter(app), Literal(msgid),
		rfc5424ContextFormatter(structId, transformer), bomFormatter, msgFormatter)
}

func rfc5424ContextFormatter(structuredId string, transformer ContextTransformer) Formatter {
	return func(buf Buffer, e *Event) {
		buf.WriteString(structuredId)
		context := e.Context
		if transformer != nil {
			context = transformer(context)
		}
		sub := NewBuffer()
		formatStructuredContext(sub, context)
		if sub.Len() > 0 {
			buf.WriteRune(' ')
			buf.Write(sub.Bytes())
		}
	}
}

func formatBOM(buf Buffer, event *Event) {
	buf.Write(rfc5424BOM)
}

func priFormatter(facility Facility) Formatter {
	return func(buf Buffer, event *Event) {
		buf.WriteString(fmt.Sprintf("<%d>", priorityFor(facility, event.Level)))
	}
}

func procIdFormatter(app string) Formatter {
	return Literal(fmt.Sprintf("%s[%d]", app, os.Getpid))
}

func priorityFor(facility Facility, level Level) priority {
	return priority(8*facility) + priority(severityFor(level))
}

func severityFor(level Level) severity {
	switch level {
	case DEBUG:
		return sDEBUG
	case INFO:
		return sINFO
	case WARN:
		return sWARN
	case ERROR:
		return sERROR
	case FATAL:
		return sCRITICAL
	default:
		panic(fmt.Errorf("billet/target: unknown level: %s", level))
	}
}

func localSyslog() (network string, address string, err error) {
	for _, network = range []string{"unixgram", "unix"} {
		for _, address = range syslogSockets {
			_, err = net.Dial(network, address)
			if err == nil {
				return
			}
		}
	}
	err = errors.New("billet/target: failed to find unix socket for syslog")
	return
}

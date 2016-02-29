package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var log = RootLogger

func main() {
	tags := EmptyContext.WithField("tag", "development").WithField("tag", "sometag")
	transformer := func(c Context) Context {
		return tags
	}

	CollectAsync(INFO, 100, false, StructuredSyslog{
		Facility:              LOCAL0,
		App:                   "bogus",
		Network:               "tcp",
		Address:               "bogus.private",
		StructuredId:          "bogus@12345",
		StructuredTransformer: transformer,
		MessageFormatter:      JsonMessage,
	}.New())

	CollectAsync(ERROR, 100, false, HTTP{RequestFormatter: func(event *Event) (request *http.Request, err error) {
		u, _ := url.Parse("https://bogus:bogus@bogus.private/12345")
		body := Render(FormatMessage, event)
		request, err = http.NewRequest("POST", fmt.Sprintf("%s://%s/api%s/store/", u.Scheme, u.Host, u.Path), bytes.NewReader(body))
		hdr := NewBuffer()
		hdr.WriteString("blah")
		return
	}}.New())

	reproduce()

	Close(time.Second)
}

func reproduce() {
	defer log.Recover("Test")
	log.Panic()
}

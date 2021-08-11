package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"sort"
)

// Explore is the second step of the test helper algorithm. Its objective
// is to enumerate all the URLs we can discover by redirection from
// the original URL in the test list. Because the test list contains by
// definition noisy data, we need this preprocessing step to learn all
// the URLs that are actually implied by the original URL.
//
// Through the explore step, we also learn about the final page on which
// we land by following the given URL. This webpage is mainly useful to
// search for block pages using the Web Connectivity algorithm.

// RoundTrip describes a specific round trip.
type RoundTrip struct {
	// Request is the original HTTP request. The headers
	// also include cookies. The body has already been
	// consumed but we should not be using bodies anyway.
	Request *http.Request

	// Response is the HTTP response. The body has already
	// been consumed, so you should use Body instead.
	Response *http.Response

	// Body is the final response body. This field should only
	// be set for the final round trip and nil otherwise.
	Body []byte

	// sortIndex is an internal field using for sorting.
	sortIndex int
}

// Explore returns a list of round trips sorted so that the first
// round trip is the first element in the list, and so on.
func Explore(URL string) ([]*RoundTrip, error) {
	resp, body, err := get(URL)
	if err != nil {
		return nil, err
	}
	rts := rearrange(resp, body)
	return rts, nil
}

// rearrange takes in input the final response of an HTTP transaction
// and its body, and produces in output a list of round trips sorted
// such that the first round trip is the first element in the out array.
func rearrange(resp *http.Response, body []byte) (out []*RoundTrip) {
	index := 0
	for resp != nil && resp.Request != nil {
		out = append(out, &RoundTrip{
			sortIndex: index,
			Request:   resp.Request,
			Response:  resp,
			Body:      body,
		})
		body = nil // only store it for last response
		resp = resp.Request.Response
	}
	sh := &sortHelper{out}
	sort.Sort(sh)
	return
}

// sortHelper is the helper structure to sort round trips.
type sortHelper struct {
	v []*RoundTrip
}

// Len implements sort.Interface.Len.
func (sh *sortHelper) Len() int {
	return len(sh.v)
}

// Less implements sort.Interface.Less.
func (sh *sortHelper) Less(i, j int) bool {
	return sh.v[i].sortIndex >= sh.v[j].sortIndex
}

// Swap implements sort.Interface.Swap.
func (sh *sortHelper) Swap(i, j int) {
	sh.v[i], sh.v[j] = sh.v[j], sh.v[i]
}

// get gets the given URL and returns the final response after
// redirection, the final response body, and an error. If the
// error is nil, the final response and its body are both valid.
func get(URL string) (*http.Response, []byte, error) {
	jarjar, _ := cookiejar.New(nil)
	clnt := &http.Client{
		Transport: http.DefaultTransport,
		Jar:       jarjar,
	}
	resp, err := clnt.Get(URL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}

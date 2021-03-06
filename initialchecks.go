package main

import (
	"errors"
	"net"
	"net/url"
)

// InitialChecks is the first step of the test helper algorithm. We
// make sure we can parse the URL, we handle the scheme, and the domain
// name inside the URL's authority is valid.

// Errors returned by Preresolve.
var (
	// ErrInvalidURL indicates that the URL is invalid.
	ErrInvalidURL = errors.New("the URL is invalid")

	// ErrUnsupportedScheme indicates that we don't support the scheme.
	ErrUnsupportedScheme = errors.New("unsupported scheme")

	// ErrNoSuchHost indicates that the DNS resolution failed.
	ErrNoSuchHost = errors.New("no such host")
)

// InitialChecks checks whether the URL is valid and whether the
// domain inside the URL is an existing one. If these preliminary
// checks fail, there's no point in continuing.
func InitialChecks(URL string) error {
	parsed, err := url.Parse(URL)
	if err != nil {
		return ErrInvalidURL
	}
	switch parsed.Scheme {
	case "http", "https":
	default:
		return ErrUnsupportedScheme
	}
	// Assumptions:
	//
	// 1. the resolver will cache the resolution for later
	//
	// 2. an IP address does not cause an error because we are using
	// a resolve that behaves like getaddrinfo
	if _, err := net.LookupHost(parsed.Hostname()); err != nil {
		return ErrNoSuchHost
	}
	return nil
}

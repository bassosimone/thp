package main

import (
	"crypto/tls"
	"net"
)

// Generate is the third step of the algorithm. Given the
// observed round trips, we generate measurement targets and
// execute those measurements so the probe has a benchmark.

// URLMeasurement is a measurement of a given URL that
// includes connectivity measurement for each endpoint
// implied by the given URL.
type URLMeasurement struct {
	// URL is the URL we're using
	URL string

	// DNS contains the domain names resolved by the helper.
	DNS *DNSMeasurement

	// RoundTrip is the related round trip.
	RoundTrip *RoundTrip

	// Endpoints contains endpoint measurements.
	Endpoints []*HTTPEndpointMeasurement
}

// DNSMeasurement is a DNS measurement.
type DNSMeasurement struct {
	// Domain is the domain we wanted to resolve.
	Domain string

	// Addrs contains the resolved addresses.
	Addrs []string
}

// HTTPEndpointMeasurement is a measurement of a specific HTTP endpoint.
type HTTPEndpointMeasurement struct {
	// Endpoint is the endpoint we're measuring.
	Endpoint string

	// TCPConnectMeasurement is the related TCP connect measurement.
	TCPConnectMeasurement *TCPConnectMeasurement

	// TLSHandshakeMeasurement is the related TLS handshake measurement.
	TLSHandshakeMeasurement *TLSHandshakeMeasurement
}

// Implementation note: OONI uses nil to indicate no error but here
// it's more convenient to just use an empty string.

// TCPConnectMeasurement is a TCP connect measurement.
type TCPConnectMeasurement struct {
	// Failure is the error that occurred.
	Failure string
}

// TLSHandshakeMeasurement is a TLS handshake measurement.
type TLSHandshakeMeasurement struct {
	// Failure is the error that occurred.
	Failure string
}

// Generate takes in input a list of round trips and outputs
// a list of connectivity measurements for each of them.
func Generate(rts []*RoundTrip) ([]*URLMeasurement, error) {
	var out []*URLMeasurement
	for _, rt := range rts {
		addrs, err := net.LookupHost(rt.Request.URL.Hostname())
		if err != nil {
			return nil, err
		}
		currentURL := &URLMeasurement{
			DNS: &DNSMeasurement{
				Domain: rt.Request.URL.Hostname(),
				Addrs:  addrs,
			},
			RoundTrip: rt,
			URL:       rt.Request.URL.String(),
		}
		out = append(out, currentURL)
		for _, addr := range addrs {
			// simplified algorithm to choose the port.
			var endpoint string
			switch rt.Request.URL.Scheme {
			case "http":
				endpoint = net.JoinHostPort(addr, "80")
			case "https":
				endpoint = net.JoinHostPort(addr, "443")
			default:
				panic("should not happen")
			}
			currentEndpoint := &HTTPEndpointMeasurement{
				Endpoint: endpoint,
			}
			currentURL.Endpoints = append(currentURL.Endpoints, currentEndpoint)
			tcpConn, err := net.Dial("tcp", endpoint)
			if err != nil {
				s := err.Error()
				currentEndpoint.TCPConnectMeasurement = &TCPConnectMeasurement{
					Failure: s,
				}
				continue
			}
			defer tcpConn.Close() // suboptimal of course
			currentEndpoint.TCPConnectMeasurement = &TCPConnectMeasurement{}
			if rt.Request.URL.Scheme == "https" {
				tlsConn := tls.Client(tcpConn, &tls.Config{
					ServerName: rt.Request.URL.Hostname(),
				})
				err := tlsConn.Handshake()
				if err != nil {
					s := err.Error()
					currentEndpoint.TLSHandshakeMeasurement = &TLSHandshakeMeasurement{
						Failure: s,
					}
					continue
				}
				defer tlsConn.Close() // suboptimal of course
				currentEndpoint.TLSHandshakeMeasurement = &TLSHandshakeMeasurement{}
			}
		}
	}
	return out, nil
}

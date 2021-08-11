package main

import (
	"flag"
	"fmt"
	"log"
)

// notable URLs:
// - http://яндекс.рф/
// - ...

func main() {
	URL := flag.String("url", "", "URL to measure")
	flag.Parse()
	if *URL == "" {
		log.Fatal("usage: go run thp.go -url <URL>")
	}
	if err := InitialChecks(*URL); err != nil {
		log.Fatalf("initial checks failed: %s", err.Error())
	}
	rts, err := Explore(*URL)
	if err != nil {
		log.Fatalf("explore failed: %s", err.Error())
	}
	meas, err := Generate(rts)
	if err != nil {
		log.Fatalf("generate failed: %s", err.Error())
	}
	for _, m := range meas {
		fmt.Printf("# %s\n", m.URL)
		fmt.Printf("method: %s\n", m.RoundTrip.Request.Method)
		fmt.Printf("url: %s\n", m.RoundTrip.Request.URL.String())
		fmt.Printf("headers: %+v\n", m.RoundTrip.Request.Header)
		fmt.Printf("dns: %+v\n", m.DNS)
		for _, e := range m.Endpoints {
			fmt.Printf("## %s\n", e.Endpoint)
			fmt.Printf("tcp: %+v\n", e.TCPConnectMeasurement)
			fmt.Printf("tls: %+v\n", e.TLSHandshakeMeasurement)
		}
	}
}

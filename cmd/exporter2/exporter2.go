package main

import (
	"eth2-exporter/exporter2"
	"flag"
)

var clientURL string

func main() {
	flag.StringVar(&clientURL, "url", "", "url")
	flag.Parse()
	exporter2.Start(clientURL)
}

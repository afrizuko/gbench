package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"
)

var (
	c, n          int
	url, protocol string
	client        http.Client
)

func init() {
	fmt.Println("This is GoBench, Version 2.3")
	fmt.Println("Copyright 2021 Olupot Doug, Apothem Technology Ltd,")
	fmt.Println("Licensed to Apothem Group, http://www.apothemgroup.io/")
	fmt.Println("")

	flag.IntVar(&c, "c", 50, "specifies the number of concurrent connections to open")
	flag.IntVar(&n, "n", 50, "specifies the number of rqeuests to forward to the target")
	flag.StringVar(&url, "t", "", "specifies the target url to benchmark")
	flag.StringVar(&protocol, "p", "http", "specifies the protocol to benchmark")
}

func main() {

	flag.Parse()
	setupClient()

	ctl := make(chan result)
	fmt.Printf("Benchmarking %s (be patient).....\n", url)

	switch protocol {
	case "tcp":
		for i := 1; i <= n; i++ {
			go benchmarkTCP(url, ctl)
		}
	case "http":
		dumpHeaderInfo(url)
		for i := 1; i <= n; i++ {
			go benchmarkTarget(url, ctl)
		}
	}

	var successful int
	var avg float64
	for i := 1; i <= n; i++ {
		x := <-ctl
		successful += x.successful
		avg += x.timelapse.Seconds()
	}
	defer close(ctl)

	fmt.Println("\n\nTotal requests:", n)
	fmt.Println("Concurrency Level:", c)
	fmt.Println("Completed requests:", successful)
	fmt.Println("Incomplete requests:", (n - successful))
	fmt.Printf("Time per reqeuest: %.3fs\n", (avg / float64(n)))
}

type result struct {
	successful int
	timelapse  time.Duration
}

func dumpHeaderInfo(url string) {
	res, err := client.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	for k, v := range res.Header {
		fmt.Println(k+":", v)
	}
}

func benchmarkTarget(url string, ctl chan result) {
	now := time.Now()
	res, err := client.Get(url)
	if err != nil {
		ctl <- result{successful: 0, timelapse: time.Since(now)}
		return
	}
	res.Body.Close()
	ctl <- result{successful: 1, timelapse: time.Since(now)}
}

func setupClient() {
	client = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func benchmarkTCP(address string, ctl chan result) {
	now := time.Now()
	conn, err := net.Dial("tcp", address)
	if err != nil {
		ctl <- result{successful: 0, timelapse: time.Since(now)}
		return
	}
	defer conn.Close()
	ctl <- result{successful: 1, timelapse: time.Since(now)}
}

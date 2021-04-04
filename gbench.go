package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"time"
)

var client http.Client

func init() {

	fmt.Println("This is GoBench, Version 2.3")
	fmt.Println("Copyright 2021 Olupot Doug, Apothem Technology Ltd,")
	fmt.Println("Licensed to Apothem Group, http://www.apothemgroup.io/")
	fmt.Println("")
}

func main() {

	var c int
	flag.IntVar(&c, "c", 50, "specifies the number of concurrent connections to open")

	var n int
	flag.IntVar(&n, "n", 50, "specifies the number of rqeuests to forward to the target")

	var url string
	flag.StringVar(&url, "t", "", "specifies the target url to benchmark")

	flag.Parse()
	setupClient()

	fmt.Printf("Benchmarking %s (be patient).....\n", url)
	dumpHeaderInfo(url)

	ctl := make(chan result, c) // add the concurrency control

	for i := 1; i <= n; i++ {
		go benchmarkTarget(url, ctl)
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

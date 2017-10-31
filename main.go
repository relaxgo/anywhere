package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	currentDir, _ = os.Getwd()
	dir           = flag.String("dir", currentDir, "port for server file")
	port          = flag.String("port", "8000", "port for server file")
	host          = flag.String("host", "localhost", "host for bind")
	global        = flag.Bool("global", false, "default IP is 127.0.0.1, if set this flag, bind ip like 192.168.0.1 or other global ip")
)

func main() {
	flag.Parse()
	fileDir := os.ExpandEnv(*dir)
	fileDir, _ = filepath.Abs(fileDir)
	if *global {
		*host = GlobalIP().String()
	}
	url := *host + ":" + *port
	log.Printf("start at http://%s   dir: %s \n", url, fileDir)
	log.Fatal(http.ListenAndServe(url, NoCache(http.FileServer(http.Dir(fileDir)))))
}

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

func NoCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Delete any ETag headers that may have been set
		for _, v := range etagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}

		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func GlobalIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "get interface addrs faild, use 0.0.0.0 instend")
		return net.IPv4zero
	}

	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if v.IP.To4() != nil && v.IP.IsGlobalUnicast() {
				return v.IP
			}
		case *net.IPAddr:
			if v.IP.To4() != nil && v.IP.IsGlobalUnicast() {
				return v.IP
			}
		}
	}
	fmt.Fprintln(os.Stderr, "not find global ipv4, use 0.0.0.0 instend")
	return net.IPv4zero
}

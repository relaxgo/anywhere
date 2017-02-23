package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	currentDir, _ = os.Getwd()
	port          = flag.String("p", "8000", "port for server file")
	isHostPublic  = flag.Bool("a", false, "is hosting for public")
)

func main() {
	flag.Parse()
	fileDir := flag.Arg(0)
	fileDir = os.ExpandEnv(fileDir)
	fileDir, _ = filepath.Abs(fileDir)
	host := "localhost:" + *port
	if *isHostPublic {
		host = LocalIP().String() + ":" + *port
	}
	log.Printf("start at http://%s   dir: %s \n", host, fileDir)
	log.Fatal(http.ListenAndServe(host, NoCache(http.FileServer(http.Dir(fileDir)))))
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

func LocalIP() net.IP {
	ip := net.IPv4(0, 0, 0, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		return ip
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return ip
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil && !v.IP.IsLoopback() {
					ip = v.IP
					return ip
				}
			case *net.IPAddr:
				if v.IP.To4() != nil && !v.IP.IsLoopback() {
					ip = v.IP
					return ip
				}
			}
		}
	}
	return ip
}

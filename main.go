package main

import (
	"runtime"
	"flag"
	"log"
	"net/http"
	"net/url"

	"github.com/Evan2698/socks2http/core"
	"golang.org/x/net/proxy"
)

func main() {
	httpAddr := flag.String("http", "0.0.0.0:8000", "local http proxy address")
	socks5Addr := flag.String("socks5", "socks5://127.0.0.1:1080", "remote socks5 address")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	socksURL, err := url.Parse(*socks5Addr)
	if err != nil {
		log.Fatalln("proxy url parse error:", err)
	}
	socks5Dialer, err := proxy.FromURL(socksURL, proxy.Direct)
	if err != nil {
		log.Fatalln("can not make proxy dialer:", err)
	}
	if err := http.ListenAndServe(*httpAddr, &core.HttpProxyRoutineHandler{Dialer: socks5Dialer}); err != nil {
		log.Fatalln("can not start http server:", err)
	}

}

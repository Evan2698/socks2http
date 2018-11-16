package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"

	"github.com/Evan2698/socks2http/core"
	"golang.org/x/net/proxy"
)

func main() {
	httpAddr := flag.String("http", "127.0.0.1:8000", "local http proxy address")
	socks5Addr := flag.String("socks5", "socks5://127.0.0.1:1080", "remote socks5 address")
	flag.Parse()

	socksURL, err := url.Parse(*socks5Addr)
	if err != nil {
		log.Fatalln("proxy url parse error:", err)
	}
	socks5Dialer, err := proxy.FromURL(socksURL, proxy.Direct)
	if err != nil {
		log.Fatalln("can not make proxy dialer:", err)
	}
	if err := http.ListenAndServe(*httpAddr, &core.HttpProxyRotineHandler{Dialer: socks5Dialer}); err != nil {
		log.Fatalln("can not start http server:", err)
	}

}
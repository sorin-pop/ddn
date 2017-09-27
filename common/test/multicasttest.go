package main

import (
	"log"
	"net"

	"github.com/djavorszky/disco"
)

func main() {
	getIP()

	c, err := disco.Subscribe("224.0.0.1:9999")
	if err != nil {
		log.Fatalf("Couldn't subscribe: %v", err)
	}

	for {
		msg := <-c
		log.Printf("Sender: %s, Msg: %q", msg.Src, msg.Message)
	}
}

func getIP() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("interfaces: %v", err)
	}
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatalf("addrs: %v", err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			log.Println(ip)
		}
	}
}

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/tatsushid/go-fastping"
)

func genTargetIPAddrs(ipAddr1 string, ipAddr2 string) []string {
	parts1 := strings.Split(ipAddr1, ".")
	firstSuffix, _ := strconv.Atoi(parts1[len(parts1)-1])
	prefix := parts1[:len(parts1)-1]
	parts2 := strings.Split(ipAddr2, ".")
	lastSuffix, _ := strconv.Atoi(parts2[len(parts2)-1])

	ipAddrs := make([]string, 0)
	for suffix := firstSuffix; suffix <= lastSuffix; suffix++ {
		ipAddrs = append(ipAddrs, strings.Join(append(prefix, strconv.Itoa(suffix)), "."))
	}
	return ipAddrs
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Printf("Usage: %s ipAddr1 [ipAddr2]\n", os.Args[0])
		os.Exit(0)
	}

	p := fastping.NewPinger()
	p.Network("udp")

	ipAddr1 := os.Args[1]

	if len(os.Args) == 3 {
		ipAddr2 := os.Args[2]
		ipAddrs := genTargetIPAddrs(ipAddr1, ipAddr2)
		for _, ia := range ipAddrs {
			p.AddIP(ia)
		}
	} else {
		ra, err := net.ResolveIPAddr("ip4:icmp", ipAddr1)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		p.AddIPAddr(ra)
	}

	fmt.Print("\033[H\033[2J")
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("IP Addr: %s\n", addr.String())
	}

	p.OnIdle = func() {
		fmt.Print("\033[H\033[2J")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("%v\n", sig) // sig is a ^C, handle it
			p.Stop()
		}
	}()

	p.RunLoop()
	select {
	case <-p.Done():
		if err := p.Err(); err != nil {
			log.Fatalf("Ping failed: %v", err)
		}
	}
}

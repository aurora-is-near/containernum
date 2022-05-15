package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	var myIPs []net.IP
	resolver := &net.Resolver{PreferGo: true}
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s requires exactly two cmdline parameters: <name> and <interface>\n", os.Args[0])
		os.Exit(1)
	}
	inf, err := net.InterfaceByName(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s not an interface\n", os.Args[2])
		os.Exit(1)
	}
	myAddrs, err := inf.Addrs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s not configured\n", os.Args[2])
		os.Exit(1)
	}
	myIPs = make([]net.IP, 0, len(myAddrs))
	for _, a := range myAddrs {
		addr, _, err := net.ParseCIDR(a.String())
		if err != nil || addr == nil {
			continue
		}
		myIPs = append(myIPs, addr.To16())
	}
	hostname := os.Args[1]
	var wg sync.WaitGroup

	for i := 1; i < 255; i++ {
		j := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
			addrs, err := resolver.LookupIPAddr(ctx, fmt.Sprintf("%s%d", hostname, j))
			cancel()
			if err != nil || len(addrs) == 0 {
				return
			}
			for _, a := range addrs {
				for _, b := range myIPs {
					if a.IP.To16().Equal(b) {
						fmt.Fprintf(os.Stdout, "%d\n", j)
						os.Exit(0)
					}
				}
			}
		}()
	}
	wg.Wait()
	os.Exit(1)
}

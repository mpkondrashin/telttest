package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	maxGoRoutines = 20000
	timeout       = 100 * time.Millisecond
	startPort     = 1
	endPort       = 1024
)

func iterateInerfacesAddresses(n chan *net.IPNet) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	fmt.Printf("Found %d interfaces(s)\n", len(ifaces))
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			ipv4Addr, ipv4Net, err := net.ParseCIDR(addr.String())
			if err != nil {
				return err
			}
			if ipv4Addr.To4() == nil {
				continue
			}
			if ipv4Addr.IsLoopback() {
				continue
			}
			n <- ipv4Net
		}
	}
	close(n)
	return err
}

func iterateAddresses(ipChan chan net.IP) {
	ipNetChan := make(chan *net.IPNet)
	go func() {
		err := iterateInerfacesAddresses(ipNetChan)
		if err != nil {
			fmt.Println(err)
		}
	}()
	for n := range ipNetChan {
		fmt.Printf("Scan %v network\n", n)
		// https://stackoverflow.com/questions/60540465/how-to-list-all-ips-in-a-network
		mask := binary.BigEndian.Uint32(n.Mask)
		start := binary.BigEndian.Uint32(n.IP)
		finish := (start & mask) | (mask ^ 0xffffffff)
		for i := start + 1; i < finish; i++ {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, i)
			ipChan <- ip
		}
	}
	close(ipChan)
}

func iterateAddressWithPort(addr chan string) {
	ipChan := make(chan net.IP)
	go func() {
		iterateAddresses(ipChan)
	}()
	for ip := range ipChan {
		for p := startPort; p <= endPort; p++ {
			addr <- ip.String() + ":" + strconv.Itoa(p)
		}
	}
	close(addr)
}

func scan() {
	addrChan := make(chan string)
	go func() {
		iterateAddressWithPort(addrChan)
	}()
	limit := make(chan struct{}, maxGoRoutines)
	for addr := range addrChan {
		limit <- struct{}{}
		go func(addr string) {
			scanPort(addr)
			<-limit
		}(addr)
	}
}

func scanPort(addr string) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return
	}
	defer conn.Close()
	fmt.Println(addr)
}

func main() {
	fmt.Println("Scan local network")
	scan()
	fmt.Println("Done")
}

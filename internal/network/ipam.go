package network

import (
    "net"
    "sync"
)

type IPAM struct {
    mu    sync.Mutex
    used  map[string]bool
    subnet *net.IPNet
}

func NewIPAM() *IPAM {
    _, subnet, _ := net.ParseCIDR(Subnet)
    return &IPAM{
        used:  make(map[string]bool),
        subnet: subnet,
    }
}

func (ipam *IPAM) Allocate() (string, error) {
    ipam.mu.Lock()
    defer ipam.mu.Unlock()
    // Start from .2 (gateway is .1)
    for i := 2; i < 254; i++ {
        ip := net.IPv4(10, 0, 0, byte(i))
        if !ipam.used[ip.String()] {
            ipam.used[ip.String()] = true
            return ip.String(), nil
        }
    }
    return "", fmt.Errorf("no IP available")
}

func (ipam *IPAM) Release(ip string) {
    ipam.mu.Lock()
    defer ipam.mu.Unlock()
    delete(ipam.used, ip)
}
package network

import (
    "fmt"
    "net"
    "os/exec"
    "strconv"
    "strings"
    "syscall"
    "github.com/vishvananda/netlink"
)

const (
    BridgeName = "br0"
    Subnet     = "10.0.0.0/24"
    Gateway    = "10.0.0.1"
)

// SetupBridge ensures bridge exists and IP is assigned
func SetupBridge() error {
    _, err := netlink.LinkByName(BridgeName)
    if err != nil {
        // Create bridge
        la := netlink.NewLinkAttrs()
        la.Name = BridgeName
        br := &netlink.Bridge{LinkAttrs: la}
        if err := netlink.LinkAdd(br); err != nil {
            return err
        }
        // Set bridge up
        if err := netlink.LinkSetUp(br); err != nil {
            return err
        }
        // Add IP
        addr, _ := netlink.ParseAddr(Subnet)
        if err := netlink.AddrAdd(br, addr); err != nil {
            return err
        }
        // Enable IP forwarding
        exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()
        // NAT via iptables
        exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", Subnet, "-j", "MASQUERADE").Run()
    }
    return nil
}

// CreateVethPair creates veth pair, moves one end into container netns, assigns IP
func CreateVethPair(containerID string, containerPID int, ipAddr string) error {
    vethHost := fmt.Sprintf("veth-%s", containerID[:8])
    vethGuest := fmt.Sprintf("eth0-%s", containerID[:8])

    // Create veth pair
    veth := &netlink.Veth{
        LinkAttrs: netlink.LinkAttrs{Name: vethHost},
        PeerName:  vethGuest,
    }
    if err := netlink.LinkAdd(veth); err != nil {
        return err
    }
    // Move guest end to container netns
    netnsHandle, err := netlink.GetNetNsFromPid(containerPID)
    if err != nil {
        return err
    }
    guestLink, err := netlink.LinkByName(vethGuest)
    if err != nil {
        return err
    }
    if err := netlink.LinkSetNsFd(guestLink, int(netnsHandle)); err != nil {
        return err
    }
    // Set host end up
    hostLink, err := netlink.LinkByName(vethHost)
    if err != nil {
        return err
    }
    if err := netlink.LinkSetUp(hostLink); err != nil {
        return err
    }
    // Add bridge and attach host end
    bridge, err := netlink.LinkByName(BridgeName)
    if err != nil {
        return err
    }
    if err := netlink.LinkSetMaster(hostLink, bridge); err != nil {
        return err
    }
    // Configure guest inside netns
    if err := configureGuestNet(netnsHandle, vethGuest, ipAddr); err != nil {
        return err
    }
    return nil
}

func configureGuestNet(netnsHandle int, guestIface, ipAddr string) error {
    // Execute network config inside netns using nsenter
    cmd := exec.Command("nsenter", "-t", strconv.Itoa(netnsHandle), "-n", "ip", "addr", "add", ipAddr+"/24", "dev", guestIface)
    if err := cmd.Run(); err != nil {
        return err
    }
    cmd = exec.Command("nsenter", "-t", strconv.Itoa(netnsHandle), "-n", "ip", "link", "set", "lo", "up")
    if err := cmd.Run(); err != nil {
        return err
    }
    cmd = exec.Command("nsenter", "-t", strconv.Itoa(netnsHandle), "-n", "ip", "link", "set", guestIface, "up")
    if err := cmd.Run(); err != nil {
        return err
    }
    cmd = exec.Command("nsenter", "-t", strconv.Itoa(netnsHandle), "-n", "ip", "route", "add", "default", "via", Gateway)
    return cmd.Run()
}
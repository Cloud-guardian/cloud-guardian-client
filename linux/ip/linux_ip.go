package linux_ip

import (
	"encoding/binary"
	"fmt"
	"net"
	"encoding/hex"
	"bufio"
	"os"
	"strconv"
	"strings"
)

type routeEntry struct {
	Destination net.IP
	DestStr	    string // CIDR notation or "default"
	PrefixLength int
	Gateway     net.IP
	Iface       string
	Metric      int
	Proto       string
	Scope       string
	Src         net.IP
}

type Addr struct {
	IP           string
	PrefixLength int
	Scope        string
}

type Interface struct {
	Index        int    // positive integer that starts at one, zero is never used
	MTU          int    // maximum transmission unit
	Name         string // e.g., "en0", "lo0", "eth0.100"
	State        string // e.g., "UP", "DOWN"
	HardwareAddr string // IEEE MAC-48, EUI-48 and EUI-64 form
	IPAddresses  []Addr
}

func GetRoutes() ([]routeEntry, error) {
	var routes []routeEntry

	file, err := os.Open("/proc/net/route")
	if err != nil {
		return nil, fmt.Errorf("error opening /proc/net/route: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// skip header
	if !scanner.Scan() {
		return nil, fmt.Errorf("No data in /proc/net/route")
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 11 {
			continue
		}

		iface := fields[0]
		dest := parseHexIP(fields[1])
		gw := parseHexIP(fields[2])

		maskHex := fields[7]
		mask := parseHexIP(maskHex)
		ipMask := net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])

		metric, _ := strconv.Atoi(fields[6])

		entry := routeEntry{
			Destination: dest,
			// PrefixLength: mask.Mask.Size(),
			Gateway:     gw,
			Iface:       iface,
			Metric:      metric,
			Proto:       "kernel", // default assumption
			Scope:       "link",   // default assumption
		}

		// Determine if default route
		if dest.Equal(net.IPv4(0, 0, 0, 0)) && ipMask.String() == net.CIDRMask(0, 32).String() {
			entry.Proto = "dhcp" // heuristic
			entry.Scope = ""
		}

		// Try to guess src from iface
		ifi, err := net.InterfaceByName(iface)
		if err == nil {
			addrs, _ := ifi.Addrs()
			for _, a := range addrs {
				ip, _, _ := net.ParseCIDR(a.String())
				if ip.To4() != nil {
					entry.Src = ip
					break
				}
			}
		}
		dstStr := ""
		if entry.Destination.Equal(net.IPv4(0, 0, 0, 0)) && net.IP(ipMask).Equal(net.IPv4(0, 0, 0, 0)) {
			dstStr = "default"
		} else {
			dstStr = (&net.IPNet{IP: entry.Destination, Mask: net.CIDRMask(entry.PrefixLength, 32)}).String()
		}
		entry.DestStr = dstStr


		routes = append(routes, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/net/route: %w", err)
	}

	return routes, nil
}

func GetIPInterfaces() ([]Interface, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("error getting network interfaces: %w", err)
	}

	var result []Interface
	for _, ifi := range ifs {
		addrs, err := ifi.Addrs()
		if err != nil {
			return nil, fmt.Errorf("error getting addresses for interface %s: %w", ifi.Name, err)
		}

		state := "DOWN"
		if ifi.Flags&net.FlagUp != 0 {
			state = "UP"
		}

		var IPAddresses []Addr
		for _, addr := range addrs {
			ip, prefixLen, scope := parseIP(addr)
			IPAddresses = append(IPAddresses, Addr{
				IP:           ip,
				PrefixLength: prefixLen,
				Scope:        scope,
			})
		}

		result = append(result, Interface{
			Name:         ifi.Name,
			MTU:          ifi.MTU,
			State:        state,
			IPAddresses:  IPAddresses,
			HardwareAddr: ifi.HardwareAddr.String(),
		})
	}
	return result, nil
}

func parseIP(a net.Addr) (string, int, string) {
	// Returns the IP address, prefix length, and scope
	s := a.String()
	// Determine whether IPv4 or IPv6
	if ipnet, ok := a.(*net.IPNet); ok {
		ip := ipnet.IP
		if ip.To4() != nil {
			// IPv4
			return ip.String(), prefixLen(ipnet), scopeOf(ip)
		}
		// IPv6
		return ip.String(), prefixLen(ipnet), ipv6Scope(ip)
	}
	return s, 0, ""
}

func prefixLen(ipnet *net.IPNet) int {
	ones, _ := ipnet.Mask.Size()
	return ones
}

func scopeOf(ip net.IP) string {
	if ip.IsLoopback() {
		return "host"
	}
	if ip.IsLinkLocalUnicast() {
		return "link"
	}
	// private addresses
	if isPrivateIPv4(ip) {
		return "global"
	}
	return "global"
}

func ipv6Scope(ip net.IP) string {
	if ip.IsLoopback() {
		return "host"
	}
	if ip.IsLinkLocalUnicast() {
		return "link"
	}
	// Unique local addresses fc00::/7 are site/global - we call them global here
	return "global"
}

func isPrivateIPv4(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		i := binary.BigEndian.Uint32(ip4)
		// 10.0.0.0/8
		if i&0xff000000 == 0x0a000000 {
			return true
		}
		// 172.16.0.0/12
		if i&0xfff00000 == 0xac100000 {
			return true
		}
		// 192.168.0.0/16
		if i&0xffff0000 == 0xc0a80000 {
			return true
		}
	}
	return false
}

func parseHexIP(s string) net.IP {
	// /proc/net/route stores IP in little-endian hex
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 4 {
		return nil
	}
	return net.IPv4(b[3], b[2], b[1], b[0])
}


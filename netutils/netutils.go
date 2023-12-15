package netutils

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	IfaceIname int = iota
	IfaceMacddr
	IfaceCidr
	IfaceIp4
	IfaceIp6
	IfaceMask
)

type DiscoveryInfo struct {
	Name string
	Host string
	Ip4  string
	Port int
	Info []string
}

var dnslist = []string{
	"1.1.1.1", "1.0.0.1", //clouflare
	//	"208.67.222.222", "208.67.220.220", //opendns server
	"8.8.8.8", "8.8.4.4", //google
	//	"8.26.56.26", "8.20.247.20", //comodo
	//	"9.9.9.9", "149.112.112.112", //quad9
	//	"64.6.64.6", "64.6.65.6"
} // verisign

/* Get first finded IPv4 address of Linux network interface. */
func NetGetInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)
	if ief, err = net.InterfaceByName(interfaceName); err != nil { // get interface
		return
	}
	if addrs, err = ief.Addrs(); err != nil { // get addresses
		return
	}
	for _, addr := range addrs { // get ipv4 address
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}
	if ipv4Addr == nil {
		if len(addrs) != 0 {
			return "", fmt.Errorf(fmt.Sprintf("There isn't any ipv4 on interface %s\n", interfaceName))
		} else {
			return "", fmt.Errorf(fmt.Sprintf("There isn't any ip on interface %s\n", interfaceName))
		}
	}
	return ipv4Addr.String(), nil
}

/* Check if Network Interface has any IPv4 address. */
func NetIfaceHasIpv4(interfaceName string) bool {
	if _, err := NetGetInterfaceIpv4Addr(interfaceName); err == nil {
		return true
	}
	return false
}

/* Convert Domain to IP */
func ResolverDomain(domain string, debugflag ...bool) (addrs []string, err error) {
	if addr := net.ParseIP(domain); addr != nil {
		return []string{domain}, nil
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial:     nil,
	}

	for i := 0; i < len(dnslist); i++ {
		for _, pro := range []string{"udp", "tcp"} {
			/* setup a custom dial for Resolver, return connection to host (udp or tcp) */
			r.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(5000),
				}
				return d.DialContext(ctx, pro, dnslist[i]+":53")
			}

			/* start lookup using custom resolver, if success, return. */
			if addrs, err = r.LookupHost(context.Background(), domain); err == nil {
				return addrs, err
			}

			if len(debugflag) != 0 && debugflag[0] {
				log.Errorf("\nCan not used dns server %s for finding %s\n", dnslist[i], domain)
			}
		}
	}
	return net.LookupHost(domain) //system lockup if custome resolver fail.
}

func ResolverDomain2Ip4(domain string, debugflag ...bool) (addr string, err error) {
	if addrs, err := ResolverDomain(domain, debugflag...); err == nil {
		for _, v := range addrs {
			if strings.Contains(v, ".") {
				return v, nil
			}
		}
		return "", fmt.Errorf("there is not ipv4")
	} else {
		return "", err
	}
}

/*
	Check connection to http/https server

return nil if cant connect to server through interface
*/
func NetCheckConectionToServer(domain string, ifacenames ...string) error {
	tcpAddr := &net.TCPAddr{}

	if len(ifacenames) != 0 {
		ip4add, err := NetGetInterfaceIpv4Addr(ifacenames[0])
		if err != nil || len(ip4add) == 0 {
			return err
		} else {
			tcpAddr.IP = net.ParseIP(ip4add)
		}
	} else {
		tcpAddr = nil
	}

	d := net.Dialer{LocalAddr: tcpAddr, Timeout: time.Millisecond * 2000}

	if !strings.Contains(domain, "://") {
		domain = "http://" + domain
	}
	u, err := url.Parse(domain)
	if err != nil {
		return err
	}
	port := "80"
	if u.Scheme == "https" {
		port = "443"
	}

	host := u.Host
	if thost, tport, _ := net.SplitHostPort(u.Host); len(thost) != 0 {
		port = tport
		host = thost
	}
	ip4, err := ResolverDomain2Ip4(host)
	if err != nil {
		//		log.Error(err, host)

		return err
	}
	if conn, err := d.Dial("tcp", ip4+":"+port); err != nil {
		//		log.Error(err)
		return err
	} else {
		conn.Close()
		return nil
	}
}


/* Check if server is alive, timeout check is 666ms */
func ServerIsLive(domain string, ifacenames ...string) bool {
	tcpAddr := &net.TCPAddr{}

	if len(ifacenames) != 0 {
		ip4add, err := NetGetInterfaceIpv4Addr(ifacenames[0])
		if err != nil || len(ip4add) == 0 {
			return false
		} else {
			tcpAddr.IP = net.ParseIP(ip4add)
		}
	} else {
		tcpAddr = nil
	}

	d := net.Dialer{LocalAddr: tcpAddr, Timeout: time.Millisecond * 666}

	if !strings.Contains(domain, "://") {
		domain = "http://" + domain
	}
	u, err := url.Parse(domain)
	if err != nil {
		return false
	}
	port := "80"
	if u.Scheme == "https" {
		port = "443"
	}

	host := u.Host
	if thost, tport, _ := net.SplitHostPort(u.Host); len(thost) != 0 {
		port = tport
		host = thost
	}
	ip4, err := ResolverDomain2Ip4(host)
	if err != nil {
		//		log.Error(err, host)

		return false
	}
	if conn, err := d.Dial("tcp", ip4+":"+port); err != nil {
		//		log.Error(err)
		return false
	} else {
		conn.Close()
		return true
	}
}


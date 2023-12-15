package main

import (
	netutils "netutils/netutils"

	log "github.com/sirupsen/logrus"
)

func main() {
	iface := "enp1s0"
	ipv4, err := netutils.NetGetInterfaceIpv4Addr(iface)
	if err != nil {
		log.Error(err)
	}
	log.Info("IPv4 of enp1s0 is: ", ipv4)

	if netutils.NetIfaceHasIpv4("enp1s0") {
		log.Info("enp1s0 has IPv4")
	} else {
		log.Info("enp1s0 doesn't has IPv4")
	}
	if netutils.NetIfaceHasIpv4("vethc025c8d@if10") {
		log.Info("enp1s0 has IPv4")
	} else {
		log.Info("enp1s0 doesn't has IPv4")
	}

	ip, _ := netutils.ResolverDomain("stgapi.smartocr.vn", true)
	log.Info("IP address of stgapi.smartocr.vn: ", ip)

	svname := "https://stgapi.smartocr.vn"
	checksv := netutils.NetCheckConectionToServer(svname, "enp1s0")
	if checksv != nil {
		log.Error("Can't not connect to ", svname)
	} else {
		log.Info("Server ", svname)
	}

}

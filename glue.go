package main

import (
	"net"

	"github.com/miekg/dns"
)

type Glue struct {
	NS []NSData
}

func (g *Glue) CheckParent(domain string) (bool, []string, error) {
	parentGlue, err := getParentGlue(domain)
	if err != nil {
		return false, []string{}, err
	}
	ok, res := g.Compare(parentGlue)
	return ok, res, nil
}

func (g *Glue) CheckSelf(domain string) (bool, []string, error) {
	selfGlue, err := g.getSelfGlue(domain)
	if err != nil {
		return false, []string{}, err
	}
	ok, res := g.Compare(selfGlue)
	return ok, res, nil
}

func (g *Glue) Compare(parentGlue []net.IP) (bool, []string) {
	var NSips []net.IP
	for _, data := range g.NS {
		NSips = append(NSips, data.IP...)
	}

	m := make(map[string]bool)
	var ips []string
	for _, ip := range NSips {
		m[ip.String()] = false
	}
	for _, ip := range parentGlue {
		m[ip.String()] = true
	}
	for k, v := range m {
		if v == false {
			ips = append(ips, k)
		}
	}
	if len(ips) == 0 {
		return true, ips
	}
	return false, ips
}

func getParentGlue(domain string) ([]net.IP, error) {
	// TODO ask every parent
	log.Debugf("Finding NS of parent: %s", dns.Fqdn(getParentDomain(domain)))
	var ips []net.IP
	nsdata, err := findNS(getParentDomain(domain))
	if err != nil {
		return ips, err
	}
	// asking parent about NS
	log.Debugf("Asking parent %s (%s) NS of %s", nsdata[0].Info[0].IP.String(), getParentDomain(domain), domain)
	return getGlueIPs(domain, nsdata[0].Info[0].IP.String())
}

func (g *Glue) getSelfGlue(domain string) ([]net.IP, error) {
	// TODO all NS
	log.Debugf("Asking self %s (%s) NS of %s", g.NS[0].IP[0].String(), domain, domain)
	return getGlueIPs(domain, g.NS[0].IP[0].String())
}

func getGlueIPs(domain string, server string) ([]net.IP, error) {
	var ips []net.IP
	res, err := query(domain, dns.TypeNS, server, true)
	if err != nil {
		return ips, err
	}
	rrset := extractRR(res.Msg.Extra, dns.TypeA, dns.TypeAAAA)
	return extractIP(rrset), nil
}

package utils

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

type DomainProfile struct {
	Domain           string
	Dmarc            string
	Spf              string
	DmarcEnabled     bool
	SpfEnabled       bool
	SpoofingPossible bool
	Valid            bool
}

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)

func InterpretDomainInvestigation(dp DomainProfile) {
	fmt.Println()
	defer fmt.Println() // Leave space for other domain results

	if !dp.Valid {
		fmt.Println(NegativeBracket, dp.Domain, "is not valid!")
		return
	}

	if dp.DmarcEnabled {
		fmt.Println(InfoBracket, "DMARC Record:", dp.Dmarc)
		if dp.SpoofingPossible {
			fmt.Println(AlertBracket, "DMARC is unenforced!")
		}
	} else {
		fmt.Println(AlertBracket, "DMARC record does not exist!")
	}

	if dp.SpfEnabled {
		fmt.Println(InfoBracket, "SPF Record:", dp.Spf)
	} else {
		fmt.Println(AlertBracket, "SPF record does not exist!")
	}

	if dp.SpoofingPossible {
		fmt.Printf("%s Spoofing is possible for %s!\n", PositiveBracket, dp.Domain)
	} else {
		fmt.Println(NegativeBracket, "Spoofing is not possible for", dp.Domain)
	}
}

func DomainExists(domain string) bool {
	if !domainRegex.MatchString(domain) {
		return false
	}

	// If an IP lookup returns null; fall back to MX lookup
	ips, err := net.LookupHost(domain)
	if err != nil {
		if mxs, err := net.LookupMX(domain); err == nil && len(mxs) > 0 {
			return true
		}
		return false
	}
	return len(ips) > 0
}

func InvestigateDomain(domain string) DomainProfile {
	dp := DomainProfile{Domain: domain, Valid: true}

	if !DomainExists(domain) {
		dp.Valid = false
		return dp
	}

	dmarc, err := getDMARCRecord(domain)
	if err != nil {
		dp.Dmarc = err.Error()
		dp.DmarcEnabled = false
		dp.SpoofingPossible = true
	} else {
		dp.Dmarc = dmarc
		dp.DmarcEnabled = true
		if strings.Contains(strings.ToLower(dmarc), "p=none") {
			dp.SpoofingPossible = true
		}
	}

	spf, err := getSPFRecord(domain)
	if err != nil {
		dp.Spf = err.Error()
		dp.SpfEnabled = false
	} else {
		dp.Spf = spf
		dp.SpfEnabled = true
	}

	return dp
}

func getDMARCRecord(domain string) (string, error) {
	record, err := net.LookupTXT("_dmarc." + domain)
	if err != nil || len(record) == 0 {
		return "", fmt.Errorf("DMARC record not found")
	}
	return record[0], nil
}

func getSPFRecord(domain string) (string, error) {
	txts, err := net.LookupTXT(domain)
	if err != nil {
		return "", fmt.Errorf("cannot retrieve TXT records")
	}
	for _, txt := range txts {
		if strings.HasPrefix(strings.ToLower(txt), "v=spf1") {
			return txt, nil
		}
	}
	return "", fmt.Errorf("SPF record not found")
}

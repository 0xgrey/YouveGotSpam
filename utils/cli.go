package utils

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

var Banner = `
         __ __                 _____     _   _____
        |  |  |___ _ _ _ _ ___|   __|___| |_|   __|___ ___ _____
        |_   _| . | | | | | -_|  |  | . |  _|__   | . | .'|     |
          |_| |___|___|\_/|___|_____|___|_| |_____|  _|__,|_|_|_|
                                                  |_|`

var Usage = `
Usage:

Investigate for Spoofability:
	YouveGotSpam investigate <domain> [<domain>...]
	YouveGotSpam investigate </path/to/domains.txt>

Enumerate MDI:
	YouveGotSpam check_mdi <domain>
`

var (
	AlertBracket    = "\033[31m[!]\033[0m"
	InfoBracket     = "\033[34m[*]\033[0m"
	PositiveBracket = "\033[32m[+]\033[0m"
	NegativeBracket = "\033[31m[-]\033[0m"
)

func PrintUsage(clause ...error) {
	fmt.Println(Banner)
	for _, entry := range clause {
		fmt.Printf("Error: %s\n", entry)
	}
	fmt.Println(Usage)
}

func ParseOptFlags(flags []string) {
	fs := flag.NewFlagSet("optional", flag.ExitOnError)

	quiet := fs.Bool("quiet", false, "suppress banner")
	fs.BoolVar(quiet, "q", false, "suppress banner")

	fs.Parse(flags)

	if !*quiet {
		fmt.Println(Banner)
	}
}

// YouveGotSpam Actions

// TODO: add support for checking multiple domains
func ActionInvestigateDomain(domain string) (bool, error) {
	targetDomain := InvestigateDomain(domain)
	InterpretDomainInvestigation(targetDomain)
	return true, nil
}

func ActionCheckMDI(args []string) (bool, error) {
	domain := args[0]

	if !DomainExists(domain) {
		return false, fmt.Errorf("invalid domain: %s", domain)
	}

	// Discover Microsoft-managed domains
	mdiPayload := fmt.Sprintf(mdiRequestBody, domain)
	req, err := http.NewRequest("POST", autoDiscoverURL, bytes.NewBufferString(mdiPayload))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", xmlContentType)
	req.Header.Set("User-Agent", "AutodiscoverClient/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("reading response failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("parsing XML failed: %w", err)
	}

	var domains []string
	decoder := xml.NewDecoder(bytes.NewReader(body))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}

		if se, ok := token.(xml.StartElement); ok &&
			se.Name.Local == "Domain" &&
			se.Name.Space == namespaceAutodisc {
			var d string
			if err := decoder.DecodeElement(&d, &se); err != nil {
				// skip malformed entries
				continue
			}
			domains = append(domains, d)
		}
	}

	lenDomains := len(domains)
	if lenDomains != 0 {
		fmt.Println(PositiveBracket, lenDomains, "domains found!")
		for _, d := range domains {
			fmt.Println(d)
		}
	} else {
		fmt.Println(NegativeBracket, "No domains found!")
		return false, nil
	}
	fmt.Println()

	tenant := ""
	for _, d := range domains {
		if strings.Contains(d, "onmicrosoft.com") {
			parts := strings.Split(d, ".")
			if len(parts) > 0 {
				tenant = parts[0]
			}
		}
	}

	// Parse for Microsoft tenant
	if tenant != "" {
		fmt.Println(InfoBracket, "Tenant found:", tenant)
	} else {
		fmt.Println(InfoBracket, "No tenant found")
	}
	fmt.Println()

	// Check if MDI instance exists
	hostname := tenant + sensorAPIDomain
	if _, err := net.LookupHost(hostname); err == nil {
		fmt.Println(InfoBracket, "MDI instance found:", hostname)
	} else {
		fmt.Println(InfoBracket, "No MDI instance found:", hostname)
	}

	return true, nil
}

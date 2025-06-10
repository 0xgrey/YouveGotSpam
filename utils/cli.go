package utils

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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
	Flags:
		-s; -spoofcheck; run 'investigate' against collected domains


Global Flags (Optional):
	Suppress Banner: -q; -quiet
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

func ParseOptFlags(flags []string) map[string]bool {
	fs := flag.NewFlagSet("optional", flag.ExitOnError)

	quiet := fs.Bool("quiet", false, "suppress banner")
	fs.BoolVar(quiet, "q", false, "suppress banner")

	checkDmarc := fs.Bool("spoofcheck", false, "Check spoofing on domains")
	fs.BoolVar(checkDmarc, "s", false, "Check spoofing policies on domains")

	fs.Parse(flags)

	flagmap := make(map[string]bool)

	if !*quiet {
		fmt.Println(Banner)
	}

	if *checkDmarc {
		flagmap["spoofcheck"] = true
	}
	return flagmap
}

// func ParseMDIFlags(flags []string) bool {
// 	fs := flag.NewFlagSet("MDI DMARC", flag.ExitOnError)

// 	fs.Parse(flags)

// 	if *checkDmarc {
// 		return true
// 	}
// 	return false
// }

func SliceFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if lastIndex := len(lines) - 1; lines[lastIndex] == "" {
			lines = lines[:lastIndex]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

// YouveGotSpam Actions

// TODO: add support for checking multiple domains
func ActionInvestigateDomains(domains []string) (bool, error) {
	for _, domain := range domains {
		fmt.Printf("Checking %s ...", domain)
		InterpretDomainInvestigation(InvestigateDomain(domain))
	}
	return true, nil
}

func ActionCheckMDI(args []string) (bool, []string, error) {
	fmt.Println("Checking MDI...\n")
	domain := args[0]

	if !DomainExists(domain) {
		return false, nil, fmt.Errorf("invalid domain: %s", domain)
	}

	// Discover Microsoft-managed domains
	mdiPayload := fmt.Sprintf(mdiRequestBody, domain)
	req, err := http.NewRequest("POST", autoDiscoverURL, bytes.NewBufferString(mdiPayload))
	if err != nil {
		return false, nil, err
	}
	req.Header.Set("Content-Type", xmlContentType)
	req.Header.Set("User-Agent", "AutodiscoverClient/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("reading response failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil, fmt.Errorf("parsing XML failed: %w", err)
	}

	var domains []string
	decoder := xml.NewDecoder(bytes.NewReader(body))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, nil, err
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
		return false, nil, nil
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

	return true, domains, nil
}

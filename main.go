package main

import (
	"YouveGotSpam/utils"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		utils.PrintUsage()
		return
	}
	action := os.Args[1]

	switch action {
	case "investigate", "check":
		investigate()
	case "check_mdi":
		check_mdi()
	case "spoof":
		spoof()
	default:
		utils.PrintUsage()
		return
	}
}

func investigate() {
	if utils.FileExists(os.Args[2]) {
		domains, err := utils.SliceFile(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		flags := utils.ParseOptFlags(os.Args[3:])
		utils.ActionInvestigateDomains(domains, flags["table"])
		return
	}
	count, entryPoint := 0, 2
	// Parse for flags to allow multiple domain entries in command line arguments
	for _, item := range os.Args[entryPoint:] {
		// Do not parse domains as a flag that contain "-"
		if strings.Contains(item, "-") && !strings.Contains(item, ".") {
			break
		}
		count++
	}
	flags := utils.ParseOptFlags(os.Args[count+entryPoint:])
	utils.ActionInvestigateDomains(os.Args[entryPoint:count+entryPoint], flags["table"])
}

func check_mdi() {
	flags := utils.ParseOptFlags(os.Args[3:])
	_, domains, err := utils.ActionCheckMDI(os.Args[2:])
	if err != nil {
		fmt.Println(err)
	}
	if flags["spoofcheck"] {
		fmt.Printf("\nInvestigating spoofing capabilities ...\n\n")
		utils.ActionInvestigateDomains(domains, flags["table"])
	}
}

func spoof() {
	utils.ParseOptFlags(os.Args[3:])

	// Parse file
	configFile := os.Args[2]

	if !utils.FileExists(configFile) {
		log.Fatalf("Template does not exist")
	}

	emailConfig := utils.ParseSpoofEmail(configFile)
	fmt.Println(emailConfig)

	emailConfig.TargetDomain = strings.Split(emailConfig.To, "@")[1]

	// Confirm validity
	investigation := utils.InvestigateDomain(emailConfig.TargetDomain)
	if !investigation.Valid {
		fmt.Println(utils.NegativeBracket, emailConfig.TargetDomain, "is not valid!")
		return
	} else if !investigation.SpoofingPossible {
		fmt.Println(utils.NegativeBracket, "Spoofing is not possible on", emailConfig.TargetDomain)
		return
	}
	fmt.Println(utils.PositiveBracket, "Spoofing is possible on", emailConfig.TargetDomain)

	// Send spoofed email via direct send
	spoofResult, err := utils.SendSpoofedEmail(emailConfig)
	if err == nil && spoofResult {
		fmt.Println(utils.PositiveBracket, "Spoofed email sent (direct)!")
	} else {
		fmt.Println(utils.NegativeBracket, "Error when sending email (direct).")
	}

}

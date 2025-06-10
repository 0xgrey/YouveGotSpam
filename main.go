package main

import (
	"YouveGotSpam/utils"
	"fmt"
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
		utils.ParseOptFlags(os.Args[3:])
		utils.ActionInvestigateDomains(domains)
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
	utils.ParseOptFlags(os.Args[count+entryPoint:])
	utils.ActionInvestigateDomains(os.Args[entryPoint : count+entryPoint])
}

func check_mdi() {
	spoofcheck := utils.ParseOptFlags(os.Args[3:])["spoofcheck"]
	// checkDmarc := utils.ParseMDIFlags(os.Args[3:])
	_, domains, err := utils.ActionCheckMDI(os.Args[2:])
	if err != nil {
		fmt.Println(err)
	}
	if spoofcheck {
		fmt.Println("\nInvestigating spoofing capabilities ...\n")
		utils.ActionInvestigateDomains(domains)
	}
}

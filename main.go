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

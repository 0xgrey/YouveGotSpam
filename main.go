package main

import (
	"YouveGotSpam/utils"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		utils.PrintUsage()
		return
	}
	action := os.Args[1]

	switch action {
	case "investigate":
		utils.ParseOptFlags(os.Args[3:])
		_, err := utils.ActionInvestigateDomain(os.Args[2])
		if err != nil {
			fmt.Println(err)
		}
	case "check_mdi":
		utils.ParseOptFlags(os.Args[3:])
		_, err := utils.ActionCheckMDI(os.Args[2:])
		if err != nil {
			fmt.Println(err)
		}
	default:
		utils.PrintUsage()
		return
	}
}

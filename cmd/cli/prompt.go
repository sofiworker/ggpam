package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var stdinReader = bufio.NewReader(os.Stdin)

func promptYesNo(msg string) bool {
	fmt.Println()
	for {
		fmt.Printf("%s (y/n) ", msg)
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			fmt.Println()
			return false
		}
		ans := strings.ToLower(strings.TrimSpace(line))
		switch ans {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}

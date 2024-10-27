package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("TEST HOOK")
	fmt.Println("Args: ", os.Args)
	d, err := os.ReadFile("../COMMIT_EDITMSG")
	if err != nil {
		fmt.Println("NO COMMIT MSG FILE FOUND, ", err)
		os.Exit(1)
	}
	fmt.Println(string(d))

	ok := true
	s := strings.Split(string(d), "\n")
	fmt.Printf("Summary: %v\n", s[0])
	if len(s) > 1 {
		if len(s[1]) != 0 {
			fmt.Println("No empty line between summary and description")
			ok = false
		}
	}

	if !ok {
		fmt.Println("GOT AN ERROR")
		os.Exit(1)
	}
	fmt.Println("NO ERRORS")
	os.Exit(1)
}

package main

import (
	"fmt"
	"os"
)

func main() {
	d, err := os.ReadFile("./.git/COMMIT_EDITMSG")
	if err != nil {
		panic(err)
	}

	fmt.Print("\n\nABORTION\n\n")

	fmt.Println(string(d))
	// os.Exit(1)
}

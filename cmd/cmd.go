package cmd

import (
	"fmt"
	"os"
)

func dieOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unrecoverable error, panic")
		panic(err)
	}
}

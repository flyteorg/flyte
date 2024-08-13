package commandutils

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func AskForConfirmation(s string, reader io.Reader) bool {
	fmt.Printf("%s [y/n]: ", s)
	r := bufio.NewScanner(reader)
	for r.Scan() {
		response := strings.ToLower(strings.TrimSpace(r.Text()))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
	return false
}

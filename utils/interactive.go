package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ask .
func Ask(question string, possibleValue ...string) {
	if len(possibleValue) > 0 {
		fmt.Printf(">>> %s (%s): ", question, strings.Join(possibleValue, ", "))
	} else {
		fmt.Printf(">>> %s: ", question)
	}
}

// ExpectAnswer .
func ExpectAnswer(expected string) (bool, error) {
	buf := bufio.NewReader(os.Stdin)
	line, err := buf.ReadBytes('\n')
	if err != nil {
		return false, err
	}
	return strings.TrimRight(string(line), "\r\n") == expected, nil
}

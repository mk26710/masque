package helpers

import (
	"bufio"
	"os"
	"slices"
	"strings"
)

var yeses []string = []string{"y", "yes"}

func IsConfirm(s string) bool {
	return slices.Contains(yeses, strings.ToLower(s))
}

func ReadLine() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return scanner.Text(), nil
}

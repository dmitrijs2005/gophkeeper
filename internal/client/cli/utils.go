package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func GetSimpleText(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Println(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func GetPassword() ([]byte, error) {
	fmt.Println("-Enter password")
	return term.ReadPassword(int(os.Stdin.Fd()))
}

func GetMultiline(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Println(prompt)

	var lines []string

	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			break
		}

		lines = append(lines, line)
	}

	text := strings.Join(lines, "\n")

	return strings.TrimSpace(text), nil
}

func GetMetadata(reader *bufio.Reader) ([]string, error) {
	fmt.Println("Enter metadata, format name=value")

	var lines []string

	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			break
		}

		lines = append(lines, line)
	}

	return lines, nil
}

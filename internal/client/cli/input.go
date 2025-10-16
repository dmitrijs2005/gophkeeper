package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var readPassword = term.ReadPassword

func GetSimpleText(reader *bufio.Reader, prompt string, w io.Writer) (string, error) {
	if _, err := fmt.Fprint(w, prompt+"\n> "); err != nil {
		return "", err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) > 0 {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func GetPassword(w io.Writer) ([]byte, error) {
	if _, err := fmt.Fprint(w, "Enter password: "); err != nil {
		return nil, err
	}
	pw, err := readPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(w)
	if err != nil {
		return nil, err
	}
	return pw, nil
}

func GetMultiline(reader *bufio.Reader, prompt string, w io.Writer) (string, error) {
	if _, err := fmt.Fprint(w, prompt+"\n(двойной Enter — закончить)\n"); err != nil {
		return "", err
	}

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

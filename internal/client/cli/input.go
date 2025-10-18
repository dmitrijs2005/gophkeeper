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

// readPassword is a test seam for term.ReadPassword.
// In tests you can replace it with a stub to avoid touching the terminal.
var readPassword = term.ReadPassword

// GetSimpleText prints a prompt to w and reads a single line of input from reader.
// The trailing newline is trimmed. If EOF occurs after some input was read,
// the partial line is returned.
//
// Example prompt format:
//
//	Prompt text
//	> _
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

// GetPassword prints a password prompt to w and reads a password
// from the user's terminal without echo. A newline is printed after
// the read to keep the UI tidy.
//
// The returned byte slice should be wiped by the caller when no longer needed.
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

// GetMultiline prints a prompt to w and reads multiple lines until an empty
// line is entered (i.e., the user presses Enter twice). The trailing newline
// on each line is trimmed and the collected text is joined with '\n'.
//
// This helper is useful for note bodies or multi-line secrets.
func GetMultiline(reader *bufio.Reader, prompt string, w io.Writer) (string, error) {
	if _, err := fmt.Fprint(w, prompt+"\n(press Enter on an empty line to finish)\n"); err != nil {
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

	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

// GetMetadata prompts the user to enter metadata lines in "name=value" form,
// one per line, ending on an empty line. The raw lines are returned unchanged.
// Validation and parsing are left to the caller.
func GetMetadata(reader *bufio.Reader) ([]string, error) {
	fmt.Println("Enter metadata in the format name=value (empty line to finish)")

	lines := make([]string, 0)
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

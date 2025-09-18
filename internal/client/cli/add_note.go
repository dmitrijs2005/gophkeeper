package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dmitrijs2005/gophkeeper/internal/client/crypto"
	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

func (a *App) AddNote() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter title:")
	title, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	title = strings.TrimSpace(title)

	fmt.Println("Enter note text (double Enter to finish):")

	var lines []string

	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if line == "" { // пустая строка = конец ввода
			break
		}

		lines = append(lines, line)
	}

	note := strings.Join(lines, "\n")
	note = strings.TrimSpace(note)

	cypherText, nonce, err := crypto.EncryptEntry(models.Note{Text: note}, a.masterKey)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("cypherText", cypherText)

	err = a.clientService.AddEntry(context.Background(), models.EntryTypeNote, title, cypherText, nonce)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

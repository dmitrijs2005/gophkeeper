package cli

import (
	"context"
	"fmt"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/client/utils"
)

func (a *App) AddFile() {

	// test
	fn := "/home/dimak/diff.txt"
	title := "some_file"

	encryptedFile, err := utils.EncryptFile(fn)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	key, url, err := a.clientService.GetPresignedPutURL(context.Background())

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("key", key, "url", url)

	err = utils.UploadToS3PresignedURL(url, encryptedFile.Cyphertext)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("upload success")

	cypherText, nonce, err := utils.EncryptEntry(
		models.File{
			StorageKey: key,
			FileKey:    encryptedFile.Key,
			Nonce:      encryptedFile.Nonce,
		}, a.masterKey)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = a.clientService.AddEntry(context.Background(), models.EntryTypeFile, title, cypherText, nonce)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

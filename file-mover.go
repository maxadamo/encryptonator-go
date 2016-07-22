// move the file to queued directory and hands over to encryption task
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

// FileMover move files to directory encrypt
func FileMover(platform, rsyncPID, basePath string) error {
	pathQueued := path.Join(basePath, "queued")
	pathEncrypt := path.Join(basePath, "encrypt")

	log.Printf("searching for files inside %s", pathQueued)
	queuedFiles, err := ioutil.ReadDir(pathQueued)
	if err != nil {
		return err
	}

	for _, f := range queuedFiles {
		srcFile := path.Join(pathQueued, f.Name())
		dstFile := path.Join(pathEncrypt, f.Name())

		log.Printf("moving %s to %s", srcFile, dstFile)
		err := os.Rename(srcFile, dstFile)
		if err != nil {
			return err
		}

		keyFile := fmt.Sprintf("%s.aes", dstFile)
		if err := WriteAESKey(keyFile); err != nil {
			return err
		}

		// TODO(ma): encrypt dstFile
	}
	return nil
}

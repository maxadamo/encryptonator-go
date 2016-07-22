// create aes key, spit the file in chunks and ecnrypt each chunk
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
)

// WriteAESKey write random strings to a file
// TODO(ma): handle the errors and return them
func WriteAESKey(keyFile string) error {
	log.Printf("creating AES key %s", keyFile)
	key := make([]byte, 32)
	rand.Read(key) // never returns an error
	return ioutil.WriteFile(keyFile, key, 0600)
}

// Encrypt chunk
func Encrypt(keystring, plainstring string) string {
	// Byte array of the string
	plaintext := []byte(plainstring)

	// Key
	key := []byte(keystring)

	// Create the AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Empty array of 16 + plaintext length
	// Include the IV at the beginning
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	// Slice of first 16 bytes
	iv := ciphertext[:aes.BlockSize]

	// Write 16 rand bytes to fill iv
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return string(ciphertext)
}

// EncryptChunks encrypt single chunks
func EncryptChunks(aesKeyChunk, fileToChunk string) {
	log.Printf("splitting %s in to chunks and encrypt", fileToChunk)

	file, err := os.Open(fileToChunk)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()
	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	log.Printf("Splitting %s into %d pieces.\n", string(fileToChunk), totalPartsNum)

	// create and open file in appen mode
	fileName := string(fileToChunk) + ".enc"
	log.Printf("Creating %s", fileName)
	_, err = os.Create(fileName)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)

		file.Read(partBuffer)

		// crypt the chunk
		encryptedBuffer := Encrypt(aesKeyChunk, string(partBuffer))
		//fmt.Printf(aesKeyChunk)
		log.Printf("chunk content is: %s", encryptedBuffer)

		// write/save buffer to disk
		//ioutil.WriteFile(fileName, []byte(encryptedBuffer), os.ModeAppend)
		_, err := f.WriteString(string(encryptedBuffer))
		if err != nil {
			panic(err)
		}
		//log.Printf("Split to: %s", fileName)
	}
	f.Close()
}

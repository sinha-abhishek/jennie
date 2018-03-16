package cryptohelper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
)

var key []byte

type EncryptionConf struct {
	PvtKey string `json:"pvt_key"`
}

func InitializeAES() error {
	f, err := os.Open("./conf/conf.json")
	if err != nil {
		log.Println("Error in initializing crypto, file open ", err)
		return err
	}
	m := &EncryptionConf{}
	err = json.NewDecoder(f).Decode(m)
	if err != nil {
		log.Println("Error in initializing crypto json decode", err)
		return err
	}

	key = []byte(m.PvtKey)
	return nil
}

func Encrypt(plain []byte, salt string) ([]byte, error) {
	keyToUse := key //append(key, []byte(salt)...)
	block, err := aes.NewCipher(keyToUse)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plain)

	return ciphertext, err
}

func Decrypt(cipherData []byte, salt string) ([]byte, error) {
	// Create the AES cipher
	keyToUse := key // append(key, []byte(salt)...)
	block, err := aes.NewCipher(keyToUse)
	if err != nil {
		return nil, err
	}

	// Before even testing the decryption,
	// if the text is too small, then it is incorrect
	if len(cipherData) < aes.BlockSize {
		log.Println("Text is too short")
		return nil, errors.New("text too short")
	}

	// Get the 16 byte IV
	iv := cipherData[:aes.BlockSize]

	// Remove the IV from the ciphertext
	cipherData = cipherData[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(cipherData, cipherData)

	return cipherData, nil
}

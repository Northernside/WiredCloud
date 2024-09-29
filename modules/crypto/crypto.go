package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32) // 32 byte key
	_, err := rand.Read(key)
	return key, err
}

func EncryptFile(content []byte, key []byte, mimeType string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	// prepend the MIME type to the content (used to delimit the actual content)
	mimeTypeWithDelimiter := mimeType + "|"
	mimeTypeBytes := []byte(mimeTypeWithDelimiter)
	contentWithType := append(mimeTypeBytes, content...)

	// encrypting
	ciphertext := gcm.Seal(nonce, nonce, contentWithType, nil)
	return ciphertext, nil
}

func DecryptFile(encryptedContent []byte, key []byte) ([]byte, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedContent) < nonceSize {
		return nil, "", fmt.Errorf("malformed ciphertext")
	}

	nonce, ciphertext := encryptedContent[:nonceSize], encryptedContent[nonceSize:]

	// decrypting
	decryptedContent, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, "", err
	}

	// extract MIME type from content
	splitIndex := bytes.Index(decryptedContent, []byte("|"))
	if splitIndex == -1 {
		return nil, "", fmt.Errorf("invalid file format: missing MIME type")
	}

	mimeType := string(decryptedContent[:splitIndex]) // extract MIME type
	actualContent := decryptedContent[splitIndex+1:]  // skip the delimiter

	return actualContent, mimeType, nil
}

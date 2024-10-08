package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

type metadata struct {
	fileName string
	mimeType string
}

func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32) // 32 byte key
	_, err := rand.Read(key)
	return key, err
}

func EncryptFile(content []byte, key []byte, fileName, mimeType string) ([]byte, error) {
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

	var metaData metadata
	metaData.fileName = fileName
	metaData.mimeType = mimeType

	metaDataBytes := []byte(metaData.fileName + "#-#" + metaData.mimeType + "|")
	contentWithMetaData := append(metaDataBytes, content...)

	// encrypting
	ciphertext := gcm.Seal(nonce, nonce, contentWithMetaData, nil)
	return ciphertext, nil
}

func DecryptFile(encryptedContent []byte, key []byte) ([]byte, string, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedContent) < nonceSize {
		return nil, "", "", fmt.Errorf("invalid file format: missing nonce")
	}

	nonce, ciphertext := encryptedContent[:nonceSize], encryptedContent[nonceSize:]

	// decrypting
	decryptedContent, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, "", "", err
	}

	// extract metadata from content
	splitIndex := bytes.Index(decryptedContent, []byte("|"))
	if splitIndex == -1 {
		return nil, "", "", fmt.Errorf("invalid file format: missing delimiter")
	}

	metaData := string(decryptedContent[:splitIndex]) // extract metadata
	metaDataSplit := bytes.Split([]byte(metaData), []byte("#-#"))
	if len(metaDataSplit) != 2 {
		return nil, "", "", fmt.Errorf("invalid file format: missing metadata")
	}

	actualContent := decryptedContent[splitIndex+1:] // skip the delimiter
	fileName := string(metaDataSplit[0])             // extract file name
	mimeType := string(metaDataSplit[1])             // extract MIME type

	return actualContent, fileName, mimeType, nil
}

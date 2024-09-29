package routes

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"wiredcloud/modules/crypto"
	"wiredcloud/modules/env"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// multipart form (max. 15GB)
	err := r.ParseMultipartForm(15 << 30)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// reading
	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// encrypting
	encryptionKey, err := crypto.GenerateEncryptionKey()
	if err != nil {
		http.Error(w, "Failed to generate encryption key", http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(fileContent)
	encryptedContent, err := crypto.EncryptFile(fileContent, encryptionKey, mimeType)
	if err != nil {
		http.Error(w, "Failed to encrypt file", http.StatusInternalServerError)
		return
	}

	randomFilename, err := generateRandomFilename()
	if err != nil {
		http.Error(w, "Failed to generate random filename", http.StatusInternalServerError)
		return
	}

	// saving
	encryptedFileName := "uploads/" + randomFilename
	err = os.WriteFile(encryptedFileName, encryptedContent, 0644)
	if err != nil {
		http.Error(w, "Failed to save encrypted file", http.StatusInternalServerError)
		return
	}

	// readable encryption key
	keyHex := hex.EncodeToString(encryptionKey)

	shareableLink := fmt.Sprintf("%s/download?filename=%s&key=%s", env.GetEnv("SERVICE_URL"), randomFilename, keyHex)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"message": "File uploaded and encrypted successfully", "link": "%s"}`, shareableLink)
	w.Write([]byte(response))
}

func generateRandomFilename() (string, error) {
	bytes := make([]byte, 4) // 4 bytes -> 8 hex chars
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

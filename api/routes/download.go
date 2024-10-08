package routes

import (
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"wiredcloud/modules/crypto"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// filename and key query params
	filename := r.URL.Query().Get("filename")
	keyHex := r.URL.Query().Get("key")

	if filename == "" || keyHex == "" {
		http.Error(w, "Missing filename or key parameter", http.StatusBadRequest)
		return
	}

	// -> back to bytes
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		http.Error(w, "Invalid key", http.StatusBadRequest)
		return
	}

	// reading
	sanitizedFilename, err := sanitizeFileName("uploads/" + filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encryptedFileName := sanitizedFilename
	encryptedContent, err := os.ReadFile(encryptedFileName)
	if err != nil {
		http.Error(w, "Failed to read encrypted file", http.StatusNotFound)
		return
	}

	// decrypting
	decryptedContent, metaFileName, mimeType, err := crypto.DecryptFile(encryptedContent, key)
	if err != nil {
		log.Printf("Failed to decrypt file: %v", err)
		http.Error(w, "Failed to decrypt file", http.StatusInternalServerError)
		return
	}

	// display filename and filesize

	w.Header().Set("Content-Disposition", "attachment; filename="+metaFileName)
	w.Header().Set("Content-Length", strconv.Itoa(len(decryptedContent)))
	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(http.StatusOK)

	w.Write(decryptedContent)
}

func sanitizeFileName(filename string) (string, error) {
	cleanedFileName := filepath.Clean(filename)

	if strings.Contains(cleanedFileName, "..") {
		return "", errors.New("invalid file name: path traversal detected")
	}

	// log.Printf("cleanedFileName: %s", cleanedFileName)

	if !strings.HasPrefix(cleanedFileName, "uploads/") && !strings.HasPrefix(cleanedFileName, "uploads\\") {
		return "", errors.New("invalid file name: outside of allowed directory")
	}

	return cleanedFileName, nil
}

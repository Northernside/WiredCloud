package routes

import (
	"encoding/hex"
	"log"
	"net/http"
	"os"
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
	encryptedFileName := "uploads/" + filename
	encryptedContent, err := os.ReadFile(encryptedFileName)
	if err != nil {
		http.Error(w, "Failed to read encrypted file", http.StatusNotFound)
		return
	}

	// decrypting
	decryptedContent, mimeType, err := crypto.DecryptFile(encryptedContent, key)
	if err != nil {
		log.Printf("Failed to decrypt file: %v", err)
		http.Error(w, "Failed to decrypt file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(http.StatusOK)

	w.Write(decryptedContent)
}

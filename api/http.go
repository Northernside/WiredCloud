package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"wiredcloud/api/routes"
	"wiredcloud/modules/env"
)

func StartWebServer() {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err := os.Mkdir("uploads", 0755)
		if err != nil {
			log.Fatalf("Failed to create uploads directory: %v", err)
		}
	}

	http.HandleFunc("/", enableCORS(routes.Index))
	http.HandleFunc("/upload", enableCORS(routes.UploadFile))
	http.HandleFunc("/download", enableCORS(routes.DownloadFile))

	log.Printf("Starting REST API server on %s:%s", env.GetEnv("HOST"), env.GetEnv("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", env.GetEnv("HOST"), env.GetEnv("PORT")), nil))
}

func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		// preflight -> OK
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	}
}

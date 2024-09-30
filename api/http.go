package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"wiredcloud/api/routes"
	"wiredcloud/modules/env"
	"wiredcloud/modules/jwt"
)

func StartWebServer() {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err := os.Mkdir("uploads", 0755)
		if err != nil {
			log.Fatalf("Failed to create uploads directory: %v\n", err)
		}
	}

	userHandler("/", routes.Index, http.MethodGet)
	userHandler("/upload", routes.UploadFile, http.MethodPost)

	http.HandleFunc("/download", enableCORS(routes.DownloadFile))

	http.HandleFunc("/auth", enableCORS(routes.Auth))
	http.HandleFunc("/api/auth/discord", enableCORS(routes.AuthDiscord))
	http.HandleFunc("/api/auth/discord/callback", enableCORS(routes.AuthDiscordCallback))

	routes.InitWhitelist()

	log.Printf("Starting REST API server on %s:%s\n", env.GetEnv("HOST"), env.GetEnv("PORT"))
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

func userHandler(path string, handler http.HandlerFunc, method string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"message": "Method not allowed"}`))
			return
		}

		authorization := r.Header.Get("Cookie")
		if authorization == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Unauthorized"}`))
			return
		}

		token := authorization[6:] // @Northernside TODO: gotta add proper cookie parsing later
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Unauthorized"}`))
			return
		}

		// routes.WhitelistedIds
		for _, id := range routes.WhitelistedIds {
			if id == claims["discord_id"] {
				handler(w, r)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Unauthorized"}`))
	})
}

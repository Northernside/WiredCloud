package routes

import (
	"bufio"
	"log"
	"net/http"
	"os"

	"wiredcloud/modules/jwt"
	"wiredcloud/modules/sqlite"
)

var WhitelistedIds = []string{}

func InitWhitelist() {
	file, err := os.OpenFile("whitelist.txt", os.O_RDONLY, 0644)
	if err != nil {
		log.Println("Failed to open whitelist file")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		WhitelistedIds = append(WhitelistedIds, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println("Failed to read whitelist file")
	}

	log.Printf("Whitelist initialized with %d ids\n", len(WhitelistedIds))
}

func Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// token query param
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusBadRequest)
		return
	}

	// validate token
	claims, err := jwt.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// get user
	user, err := sqlite.GetUser("discord_id", claims["discord_id"].(string))
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// check if user is whitelisted
	for _, id := range WhitelistedIds {
		if id == user {
			log.Printf("User %s is whitelisted\n", user)
			w.Header().Add("Set-Cookie", "token="+token+"; Path=/; Max-Age=604800")
			http.Redirect(w, r, "/", http.StatusFound)

			return
		}

		log.Printf("User %s is not whitelisted\n", user)
	}
}

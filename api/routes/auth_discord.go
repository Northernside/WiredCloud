package routes

import (
	"fmt"
	"net/http"

	"wiredcloud/modules/env"
)

func AuthDiscord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=identify", env.GetEnv("DISCORD_CLIENT_ID"), env.GetEnv("DISCORD_REDIRECT_URI")), http.StatusFound)
}

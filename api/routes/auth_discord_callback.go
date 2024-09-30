package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"wiredcloud/modules/env"
	"wiredcloud/modules/jwt"
	"wiredcloud/modules/sqlite"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type discordUser struct {
	ID string `json:"id"`
}

func AuthDiscordCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code := r.URL.Query().Get("code")

	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "No code provided"}`))
		return
	}

	clientId := env.GetEnv("DISCORD_CLIENT_ID")
	clientSecret := env.GetEnv("DISCORD_CLIENT_SECRET")
	redirectUri := env.GetEnv("DISCORD_REDIRECT_URI")
	body := fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=authorization_code&code=%s&redirect_uri=%s", clientId, clientSecret, code, redirectUri)

	tokenReq, err := http.NewRequest("POST", "https://discord.com/api/v8/oauth2/token", bytes.NewBufferString(body))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to create token request", "error": "` + err.Error() + `"}`))
		return
	}

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenRes, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to exchange code for token", "error": "` + err.Error() + `"}`))
		return
	}
	defer tokenRes.Body.Close()

	if tokenRes.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to exchange code for token", "error": "` + tokenRes.Status + `"}`))
		return
	}

	// get user info
	userReq, err := http.NewRequest("GET", "https://discord.com/api/v8/users/@me", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to create user request", "error": "` + err.Error() + `"}`))
		return
	}

	token := tokenResponse{}
	err = json.NewDecoder(tokenRes.Body).Decode(&token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to decode token", "error": "` + err.Error() + `"}`))
		return
	}

	userReq.Header.Set("Authorization", token.TokenType+" "+token.AccessToken)
	userRes, err := http.DefaultClient.Do(userReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to get user info", "error": "` + err.Error() + `"}`))
		return
	}
	defer userRes.Body.Close()

	if userRes.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to get user info", "error": "` + userRes.Status + `"}`))
		return
	}

	dUser := discordUser{}
	err = json.NewDecoder(userRes.Body).Decode(&dUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to decode user info", "error": "` + err.Error() + `"}`))
		return
	}

	// log.Printf("New sign-in: %s", dUser.ID)

	_, err = sqlite.GetUser("discord_id", dUser.ID)
	if err != nil {
		err = sqlite.CreateUser(dUser.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Failed to create user", "error": "` + err.Error() + `"}`))
			return
		}
	}

	jwtToken, err := jwt.CreateToken(dUser.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Failed to create JWT token", "error": "` + err.Error() + `"}`))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/auth?token="+jwtToken, env.GetEnv("SERVICE_URL")), http.StatusFound)
}

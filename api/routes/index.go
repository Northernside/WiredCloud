package routes

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"wiredcloud/modules/env"
)

var (
	indexHTML     string
	EnvVarPattern = regexp.MustCompile(`{{\s?.Env\.([a-zA-Z0-9_]+)\s?}}`)
)

func Index(w http.ResponseWriter, r *http.Request) {
	if indexHTML == "" {
		file, err := os.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
			return
		}

		file = []byte(replaceEnvVars(string(file)))
		indexHTML = string(file)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(indexHTML))
}

func replaceEnvVars(content string) string {
	for _, match := range EnvVarPattern.FindAllStringSubmatch(content, -1) {
		content = strings.ReplaceAll(content, match[0], env.GetEnv(match[1]))
	}

	return content
}

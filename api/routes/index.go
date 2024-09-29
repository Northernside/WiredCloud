package routes

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"wiredcloud/modules/env"
)

var indexHTML string

func Index(w http.ResponseWriter, r *http.Request) {
	if indexHTML == "" {
		file, err := ioutil.ReadFile("index.html")
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
	var regex = regexp.MustCompile(`{{\s?.Env\.([a-zA-Z0-9_]+)\s?}}`)
	matches := regex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		content = strings.ReplaceAll(content, match[0], env.GetEnv(match[1]))
	}

	return content
}

package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func HandleTURNCredentials() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			slog.Error("Method not allowed", "method", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		turnKeyID := os.Getenv("TURN_KEY_ID")
		turnKeyAPIToken := os.Getenv("TURN_KEY_API_TOKEN")

		if turnKeyID == "" || turnKeyAPIToken == "" {
			slog.Error("TURN credentials not configured")
			http.Error(w, "TURN credentials not configured", http.StatusInternalServerError)
			return
		}

		apiURL := fmt.Sprintf("https://rtc.live.cloudflare.com/v1/turn/keys/%s/credentials/generate",
			url.PathEscape(turnKeyID))

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(`{"ttl": 86400}`))
		if err != nil {
			slog.Error("Failed to create request", "error", err)
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Authorization", "Bearer "+turnKeyAPIToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Failed to get TURN credentials", "error", err)
			http.Error(w, "Failed to get TURN credentials", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			return
		}

		slog.Debug("Forwarding response", "status", resp.StatusCode, "body", string(body))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}

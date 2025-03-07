package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Serve from a public directory with specific index
type spaHandler struct {
	publicDir string // The directory from which to serve
	indexFile string // The fallback/default file to serve
	basePath  string // The base path prefix to strip from requests
}

// isAssetFile checks if the given path is an asset file based on its extension
func isAssetFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	assetExts := []string{
		".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg",
		".ico", ".woff", ".woff2", ".ttf", ".eot", ".mp4", ".webp",
		".map", ".json", ".webm", ".mp3", ".wav", ".pdf", ".txt",
	}

	for _, assetExt := range assetExts {
		if ext == assetExt {
			return true
		}
	}
	return false
}

// Falls back to a supplied index (indexFile) when either:
// (1) Request path is not found and is not an asset file
// (2) Request path is a directory
// Otherwise serves the requested file.
func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Strip the base path prefix if present
	urlPath := r.URL.Path
	if h.basePath != "" {
		urlPath = strings.TrimPrefix(urlPath, h.basePath)
		if !strings.HasPrefix(urlPath, "/") {
			urlPath = "/" + urlPath
		}
	}

	// Join the cleaned URL path with the public directory
	p := filepath.Join(h.publicDir, filepath.Clean(urlPath))

	if info, err := os.Stat(p); err != nil {
		// File doesn't exist - serve index.html only if not requesting an asset
		if !isAssetFile(urlPath) {
			http.ServeFile(w, r, filepath.Join(h.publicDir, h.indexFile))
		} else {
			// If it's an asset that doesn't exist, return 404
			http.NotFound(w, r)
		}
		return
	} else if info.IsDir() {
		// It's a directory - serve the index file
		http.ServeFile(w, r, filepath.Join(h.publicDir, h.indexFile))
		return
	}

	// File exists, serve it
	http.ServeFile(w, r, p)
}

// Returns a request handler (http.Handler) that serves a single
// page application from a given public directory (publicDir).
func SpaHandler(publicDir string, indexFile string) http.Handler {
	return &spaHandler{publicDir, indexFile, ""}
}

// Returns a request handler (http.Handler) that serves a single
// page application from a given public directory (publicDir),
// stripping the specified base path from incoming requests.
func SpaHandlerWithBasePath(publicDir, indexFile, basePath string) http.Handler {
	// Ensure basePath starts with a slash but doesn't end with one
	if basePath != "" {
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		basePath = strings.TrimSuffix(basePath, "/")
	}
	return &spaHandler{publicDir, indexFile, basePath}
}

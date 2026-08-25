package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

//go:embed responses
var responseFS embed.FS

type mockResponse struct {
	status  int
	headers http.Header
	body    []byte
}

var responses = make(map[string]*mockResponse)

func main() {
	if err := loadResponses(); err != nil {
		log.Fatalf("Failed to load mock responses: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	addr := ":" + port

	http.HandleFunc("/", handleRequest)

	log.Printf("Mock origin server listening on %s (%d endpoints loaded)...", addr, len(responses))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func loadResponses() error {
	return fs.WalkDir(responseFS, "responses", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		// Check if this directory represents a response node (has body, headers, or status)
		bodyBytes, bodyErr := responseFS.ReadFile(path.Join(filePath, "body"))
		headersBytes, headersErr := responseFS.ReadFile(path.Join(filePath, "headers"))
		statusBytes, statusErr := responseFS.ReadFile(path.Join(filePath, "status"))

		if bodyErr != nil && headersErr != nil && statusErr != nil {
			return nil
		}

		relPath := strings.TrimPrefix(filePath, "responses")
		relPath = strings.Trim(relPath, "/")
		parts := strings.SplitN(relPath, "/", 2)
		if len(parts) < 1 || parts[0] == "" {
			return nil
		}

		method := strings.ToUpper(parts[0])
		reqPath := "/"
		if len(parts) > 1 {
			reqPath = "/" + parts[1]
		}

		// Parse status code
		statusCode := http.StatusOK
		if statusErr == nil {
			if s, err := strconv.Atoi(strings.TrimSpace(string(statusBytes))); err == nil {
				statusCode = s
			}
		}

		// Parse headers
		header := make(http.Header)
		if headersErr == nil {
			lines := strings.Split(string(headersBytes), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if k, v, ok := strings.Cut(line, ":"); ok {
					header.Add(strings.TrimSpace(k), strings.TrimSpace(v))
				}
			}
		}

		// Calculate Content-Length if missing
		if header.Get("Content-Length") == "" && len(bodyBytes) > 0 {
			header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))
		}

		key := fmt.Sprintf("%s %s", method, reqPath)
		responses[key] = &mockResponse{
			status:  statusCode,
			headers: header,
			body:    bodyBytes,
		}

		log.Printf("Registered mock endpoint: %s %s [status=%d, body_bytes=%d]", method, reqPath, statusCode, len(bodyBytes))
		return nil
	})
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	cleanPath := path.Clean(r.URL.Path)
	lookupMethod := r.Method
	if lookupMethod == "HEAD" {
		lookupMethod = "GET"
	}
	key := fmt.Sprintf("%s %s", lookupMethod, cleanPath)

	resp, ok := responses[key]
	if !ok {
		if strings.HasSuffix(cleanPath, "/") && len(cleanPath) > 1 {
			key = fmt.Sprintf("%s %s", lookupMethod, strings.TrimSuffix(cleanPath, "/"))
			resp, ok = responses[key]
		}
	}

	if !ok {
		http.NotFound(w, r)
		return
	}

	for k, vv := range resp.headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.status)
	if r.Method != "HEAD" {
		w.Write(resp.body)
	}
}

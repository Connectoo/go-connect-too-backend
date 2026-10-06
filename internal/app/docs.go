package app

import (
	_ "embed"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed spec/openapi.yaml
var openAPISpec []byte

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Go Connect API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/api/v1/docs/openapi.yaml',
      dom_id: '#swagger-ui',
      deepLinking: true
    });
  </script>
</body>
</html>`

// registerDocsRoutes mounts OpenAPI spec and Swagger UI (non-production only).
func registerDocsRoutes(r chi.Router) {
	r.Route("/docs", func(r chi.Router) {
		r.Get("/openapi.yaml", serveOpenAPISpec)
		r.Get("/", serveSwaggerUI)
	})
}

// serversBlockRe matches the top-level "servers:" block (the "servers:" line
// and the indented list items that follow it) up to the next top-level key.
var serversBlockRe = regexp.MustCompile(`(?m)^servers:\n(?:[ \t-].*\n?)*`)

// requestServerURL derives the API base URL (scheme://host/api/v1) from the
// incoming request, honoring proxy-set forwarding headers used by Render,
// Vercel, and other reverse proxies.
func requestServerURL(r *http.Request) string {
	scheme := "https"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = strings.TrimSpace(strings.Split(proto, ",")[0])
	} else if r.TLS == nil {
		scheme = "http"
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	host = strings.TrimSpace(strings.Split(host, ",")[0])

	return fmt.Sprintf("%s://%s/api/v1", scheme, host)
}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := openAPISpec
	if host := requestServerURL(r); host != "" {
		block := fmt.Sprintf("servers:\n- url: %s\n  description: Current host\n", host)
		spec = serversBlockRe.ReplaceAll(spec, []byte(block))
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec)
}

func serveSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

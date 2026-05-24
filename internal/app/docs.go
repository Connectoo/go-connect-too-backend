package app

import (
	_ "embed"
	"net/http"

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

func serveOpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

func serveSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

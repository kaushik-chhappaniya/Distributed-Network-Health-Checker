package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var OpenAPISpec []byte

// ServeOpenAPISpec serves the raw OpenAPI 3.0 JSON specification.
func ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(OpenAPISpec)
}

// ServeSwaggerUI serves the interactive Swagger UI HTML page with live "Try It Out" buttons.
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Network Health Checker - API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; padding: 0; background: #fafafa; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout",
        displayRequestDuration: true,
        docExpansion: "list"
      });
    };
  </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swaggerHTML))
}


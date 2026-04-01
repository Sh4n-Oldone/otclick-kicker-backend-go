package docs

import (
	"embed"
	"net/http"
)

//go:embed openapi.yaml swagger-ui entities paths components
var files embed.FS

func Handler() http.Handler {
	fs := http.FileServer(http.FS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/docs", "/docs/":
			http.Redirect(w, r, "/docs/swagger-ui/index.html", http.StatusFound)
			return
		}
		http.StripPrefix("/docs", fs).ServeHTTP(w, r)
	})
}

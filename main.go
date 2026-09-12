package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	handler "portofolio-go/api"
)

// Local development server. It mirrors Vercel: files that exist in public/
// are served as-is, every other request goes to the same handler that runs
// in production (api/index.go), so local and deployed pages never drift.
func main() {
	static := http.FileServer(http.Dir("public"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file := filepath.Join("public", filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			static.ServeHTTP(w, r)
			return
		}
		handler.Handler(w, r)
	})

	log.Println("Server Localhost berjalan di http://localhost:8080 🚀")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

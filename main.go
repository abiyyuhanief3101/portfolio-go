package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Handler untuk halaman utama
func handleHome(w http.ResponseWriter, r *http.Request) {
	// UPDATE 1: Lokasi HTML sekarang ada di folder 'api', bukan 'templates'
	// Dan namanya sekarang 'template.html' (sesuai perubahan terakhir kita)
	layout := filepath.Join("api", "template.html")

	tmpl, err := template.ParseFiles(layout)
	if err != nil {
		http.Error(w, "Error loading HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Data yang sama persis dengan yang di api/index.go
	data := map[string]interface{}{
		"Title":    "Abiyyu Hanief | Problem Solver",
		"Name":     "Abiyyu Hanief",
		"Role":     "Product Implementator & Fullstack Developer",
		"Headline": "Empowering Communities through Tech & Process.",
		"About":    "A problem-solver who uses technology and process to empower communities. My approach, refined through experiences in both program coordination and software development, is to own a solution from concept to completion.",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// UPDATE 2: Mapping Folder 'public' agar sesuai logika Vercel
	// Di HTML kita panggil "/css/style.css", jadi kita arahkan url "/css/" ke folder "public/css"

	// Melayani folder CSS
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))

	// Melayani folder JS
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("public/js"))))

	// Melayani folder IMG
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("public/img"))))

	// Rute Halaman Utama
	http.HandleFunc("/", handleHome)

	log.Println("Server Localhost berjalan di http://localhost:8080 🚀")
	log.Println("Struktur folder sudah disesuaikan dengan Vercel (api & public)")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

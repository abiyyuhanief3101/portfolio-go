package handler

import (
	_ "embed"
	"html/template"
	"net/http"
)

// UBAH DISINI: Sesuaikan dengan nama file baru
//
//go:embed template.html
var htmlContent string

// Di Vercel, nama fungsinya WAJIB 'Handler' (huruf besar H)
// Fungsi ini yang akan dipanggil Vercel setiap ada pengunjung
func Handler(w http.ResponseWriter, r *http.Request) {

	// Parsing HTML dari string yang sudah di-embed (bukan baca file lagi)
	tmpl, err := template.New("index").Parse(htmlContent)
	if err != nil {
		http.Error(w, "Maaf, ada kesalahan sistem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Data yang sama seperti sebelumnya
	data := map[string]interface{}{
		"Title":    "Abiyyu Hanief | Portfolio",
		"Name":     "Abiyyu Hanief",
		"Role":     "Product Implementator & Fullstack Developer",
		"Headline": "Empowering Communities through Tech & Process.",
		"About":    "A problem-solver who uses technology and process to empower communities. My approach, refined through experiences in both program coordination and software development, is to own a solution from concept to completion.",
	}

	// Tampilkan
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	layout := filepath.Join("templates", "index.html")

	tmpl, err := template.ParseFiles(layout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Data Bahasa Inggris & Profesional
	data := map[string]interface{}{
		"Title":    "Abiyyu Hanief | Problem Solver",
		"Name":     "Abiyyu Hanief",
		"Role":     "School System Implementor & Go Developer",
		"Headline": "Empowering Communities through Tech & Process.",
		"About":    "A problem-solver who uses technology and process to empower communities. My approach, refined through experiences in both program coordination and software development, is to own a solution from concept to completion. For me, building a full-stack application isn't just about writing code; it is about understanding a business's pain points and delivering a tool that saves time.",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleHome)

	log.Println("Server running at http://localhost:8080 🚀")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

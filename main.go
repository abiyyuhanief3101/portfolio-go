package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Kredensial Supabase (REST API)
const supabaseUrl = "https://kgscotrveqoixnufzxea.supabase.co"
const supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imtnc2NvdHJ2ZXFvaXhudWZ6eGVhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzUzMDA3MjIsImV4cCI6MjA5MDg3NjcyMn0.2cjGOOcuyxE1z-5yhQo1epzfFd92nGPBDgPshCTbBi8"

// Struktur data menyesuaikan kolom di Supabase
type BlogPost struct {
	Slug       string        `json:"slug"`
	Title      string        `json:"title"`
	Date       string        `json:"created_at"`
	Category   string        `json:"category"`
	RawContent string        `json:"content"`
	Summary    string        `json:"-"` // "-" artinya abaikan saat membaca JSON
	Content    template.HTML `json:"-"`
}

// Mesin Markdown Parser (Tetap sama, tidak ada yang diubah)
func mdToHTML(md string) template.HTML {
	html := template.HTMLEscapeString(md)

	re := regexp.MustCompile(`(?m)^#{3}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h3>$1</h3>")
	re = regexp.MustCompile(`(?m)^#{2}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h2>$1</h2>")
	re = regexp.MustCompile(`(?m)^#{1}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h1>$1</h1>")

	re = regexp.MustCompile(`\*\*(.*?)\*\*`)
	html = re.ReplaceAllString(html, "<strong>$1</strong>")

	re = regexp.MustCompile(`\*(.*?)\*`)
	html = re.ReplaceAllString(html, "<em>$1</em>")

	parts := strings.Split(html, "\n\n")
	var pParts []string
	for _, p := range parts {
		cleanP := strings.TrimSpace(p)
		if cleanP != "" {
			if strings.HasPrefix(cleanP, "<h") {
				pParts = append(pParts, cleanP)
			} else {
				pParts = append(pParts, "<p>"+strings.ReplaceAll(cleanP, "\n", "<br>")+"</p>")
			}
		}
	}
	return template.HTML(strings.Join(pParts, "\n"))
}

// Fungsi BARU: Menarik Data Posts dari Supabase via API
func fetchPostsFromSupabase() []BlogPost {
	req, _ := http.NewRequest("GET", supabaseUrl+"/rest/v1/posts?order=created_at.desc", nil)
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	var posts []BlogPost
	if err != nil || resp.StatusCode != 200 {
		log.Println("Error fetching posts:", err)
		return posts
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &posts)

	// Memproses Markdown dan memformat Tanggal
	for i := range posts {
		// Potong format waktu ISO 8601 (2026-04-05T...) menjadi tanggal saja (2026-04-05)
		if len(posts[i].Date) >= 10 {
			posts[i].Date = posts[i].Date[:10]
		}

		// Otomatis membuat Ringkasan (Summary) dari kalimat pertama
		lines := strings.Split(posts[i].RawContent, "\n")
		for _, line := range lines {
			cleanLine := strings.TrimSpace(line)
			if cleanLine != "" && !strings.HasPrefix(cleanLine, "#") {
				posts[i].Summary = cleanLine
				break
			}
		}

		// Terjemahkan Markdown murni dari DB menjadi HTML
		posts[i].Content = mdToHTML(posts[i].RawContent)
	}
	return posts
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("api/base.html", "api/template.html")
	if err != nil {
		http.Error(w, "Error loading HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":    "Abiyyu Hanief | Problem Solver",
		"Name":     "Abiyyu Hanief",
		"Role":     "Product Implementator & Fullstack Developer",
		"Headline": "Empowering Communities through Tech & Process.",
		"About":    "A problem-solver who uses technology and process to empower communities. My approach, refined through experiences in both program coordination and software development, is to own a solution from concept to completion.",
		"Posts":    fetchPostsFromSupabase(), // Data ditarik dari Supabase!
	}

	tmpl.ExecuteTemplate(w, "base", data)
}

func handleWins(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("api/base.html", "api/wins.html")
	if err != nil {
		http.Error(w, "Error loading HTML", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", nil)
}

func handleLayers(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("api/base.html", "api/game.html")
	if err != nil {
		http.Error(w, "Error loading HTML", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", nil)
}

// Handle Post direvisi agar melakukan kueri ke API Supabase berdasarkan Slug
func handlePost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/blog/")

	// Kondisi 1: Menampilkan halaman Index Blog
	if slug == "" {
		data := map[string]interface{}{
			"Title":   "Engineering Notes",
			"Posts":   fetchPostsFromSupabase(),
			"IsIndex": true,
		}
		tmpl, err := template.ParseFiles("api/base.html", "api/blog.html")
		if err == nil {
			tmpl.ExecuteTemplate(w, "base", data)
		}
		return
	}

	// Kondisi 2: Mencari artikel spesifik berdasarkan URL Slug
	req, _ := http.NewRequest("GET", supabaseUrl+"/rest/v1/posts?slug=eq."+slug, nil)
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		http.NotFound(w, r)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var posts []BlogPost
	json.Unmarshal(body, &posts)

	if len(posts) == 0 {
		http.NotFound(w, r) // Slug tidak ditemukan di database
		return
	}

	// Data ditemukan!
	post := posts[0]
	if len(post.Date) >= 10 {
		post.Date = post.Date[:10]
	}
	post.Content = mdToHTML(post.RawContent)

	data := map[string]interface{}{
		"Title":   post.Title,
		"Date":    post.Date,
		"Content": post.Content,
		"Posts":   fetchPostsFromSupabase(), // Untuk list artikel lain di sidebar
	}

	tmpl, err := template.ParseFiles("api/base.html", "api/blog.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func main() {
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("public/js"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("public/img"))))

	http.HandleFunc("/wins", handleWins)
	http.HandleFunc("/layers", handleLayers)
	http.HandleFunc("/blog/", handlePost)
	http.HandleFunc("/", handleHome)

	log.Println("Server Localhost berjalan di http://localhost:8080 🚀")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

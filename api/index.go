package handler

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Kredensial Supabase
const supabaseUrl = "https://kgscotrveqoixnufzxea.supabase.co"
const supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imtnc2NvdHJ2ZXFvaXhudWZ6eGVhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzUzMDA3MjIsImV4cCI6MjA5MDg3NjcyMn0.2cjGOOcuyxE1z-5yhQo1epzfFd92nGPBDgPshCTbBi8"

// --- EMBED HTML FILES ---
//
//go:embed base.html
var baseContent string

//go:embed template.html
var htmlContent string

//go:embed blog.html
var blogContent string

//go:embed wins.html
var winsContent string

//go:embed game.html
var gameContent string

//go:embed library.html
var libraryContent string

// --- STRUKTUR DATA ---
type BlogPost struct {
	Slug       string        `json:"slug"`
	Title      string        `json:"title"`
	Date       string        `json:"created_at"`
	Category   string        `json:"category"`
	RawContent string        `json:"content"`
	Summary    string        `json:"-"`
	Content    template.HTML `json:"-"`
}

type Book struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	CoverURL string `json:"cover_url"`
	Rating   int    `json:"rating"`
	Review   string `json:"review"`
}

// --- FUNGSI PARSER MARKDOWN ---
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

// --- FUNGSI FETCH SUPABASE ---
func fetchPostsFromSupabase() []BlogPost {
	req, _ := http.NewRequest("GET", supabaseUrl+"/rest/v1/posts?order=created_at.desc", nil)
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, _ := client.Do(req)
	var posts []BlogPost
	if resp == nil || resp.StatusCode != 200 {
		return posts
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &posts)

	for i := range posts {
		if len(posts[i].Date) >= 10 {
			posts[i].Date = posts[i].Date[:10]
		}
		lines := strings.Split(posts[i].RawContent, "\n")
		for _, line := range lines {
			cleanLine := strings.TrimSpace(line)
			if cleanLine != "" && !strings.HasPrefix(cleanLine, "#") {
				posts[i].Summary = cleanLine
				break
			}
		}
		posts[i].Content = mdToHTML(posts[i].RawContent)
	}
	return posts
}

func fetchBooksFromSupabase() []Book {
	req, _ := http.NewRequest("GET", supabaseUrl+"/rest/v1/books?order=created_at.desc", nil)
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, _ := client.Do(req)
	var books []Book
	if resp == nil || resp.StatusCode != 200 {
		return books
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &books)
	return books
}

// --- HANDLERS (LOGIKA TAMPILAN) ---
func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(htmlContent)

	data := map[string]interface{}{
		"Title":    "Abiyyu Hanief | Problem Solver",
		"Name":     "Abiyyu Hanief",
		"Role":     "Product Implementator & Fullstack Developer",
		"Headline": "Empowering Communities through Tech & Process.",
		"About":    "A problem-solver who uses technology and process to empower communities. My approach, refined through experiences in both program coordination and software development, is to own a solution from concept to completion.",
		"Posts":    fetchPostsFromSupabase(),
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func handleWins(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(winsContent)
	tmpl.ExecuteTemplate(w, "base", nil)
}

func handleLayers(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(gameContent)
	tmpl.ExecuteTemplate(w, "base", nil)
}

func handleLibrary(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(libraryContent)

	data := map[string]interface{}{
		"Title": "Digital Library | Abiyyu Hanief",
		"Books": fetchBooksFromSupabase(),
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/blog/")

	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(blogContent)

	// Halaman Daftar Artikel (Index)
	if slug == "" {
		data := map[string]interface{}{
			"Title":   "Engineering Notes",
			"Posts":   fetchPostsFromSupabase(),
			"IsIndex": true,
		}
		tmpl.ExecuteTemplate(w, "base", data)
		return
	}

	// Halaman Detail Artikel
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
		http.NotFound(w, r)
		return
	}

	post := posts[0]
	if len(post.Date) >= 10 {
		post.Date = post.Date[:10]
	}
	post.Content = mdToHTML(post.RawContent)

	data := map[string]interface{}{
		"Title":   post.Title,
		"Date":    post.Date,
		"Content": post.Content,
		"Posts":   fetchPostsFromSupabase(),
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// --- ENTRY POINT VERCEL ---
func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/wins" {
		handleWins(w, r)
		return
	}
	if path == "/layers" {
		handleLayers(w, r)
		return
	}
	if path == "/library" {
		handleLibrary(w, r)
		return
	}
	if strings.HasPrefix(path, "/blog") {
		handlePost(w, r)
		return
	}

	// Default rute diarahkan ke Home
	handleHome(w, r)
}

package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// BlogPost holds the data parsed from markdown files
type BlogPost struct {
	Slug    string
	Title   string
	Date    string
	Summary string // First few lines of the content
	Content template.HTML
}

func mdToHTML(md string) template.HTML {
	// 1. Escape HTML
	html := template.HTMLEscapeString(md)

	// 2. Headers
	re := regexp.MustCompile(`(?m)^#{3}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h3>$1</h3>")
	re = regexp.MustCompile(`(?m)^#{2}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h2>$1</h2>")
	re = regexp.MustCompile(`(?m)^#{1}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h1>$1</h1>")

	// 3. Bold
	re = regexp.MustCompile(`\*\*(.*?)\*\*`)
	html = re.ReplaceAllString(html, "<strong>$1</strong>")

	// 4. Italic
	re = regexp.MustCompile(`\*(.*?)\*`)
	html = re.ReplaceAllString(html, "<em>$1</em>")

	// 5. Paragraphs
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

func parseMarkdown(filename string, content string) BlogPost {
	post := BlogPost{
		Slug: strings.TrimSuffix(filename, ".md"),
	}

	// Split content by "---" separator
	parts := strings.SplitN(content, "---", 3)
	if len(parts) >= 3 {
		// Parse Frontmatter (parts[1])
		lines := strings.Split(parts[1], "\n")
		for _, line := range lines {
			kv := strings.SplitN(line, ":", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
				switch key {
				case "title":
					post.Title = val
				case "date":
					post.Date = val
				case "description":
					post.Summary = val
				}
			}
		}
		// Fallback: If description is empty, grab first line of body
		if post.Summary == "" {
			bodyLines := strings.Split(parts[2], "\n")
			for _, line := range bodyLines {
				cleanLine := strings.TrimSpace(line)
				if cleanLine != "" && !strings.HasPrefix(cleanLine, "#") {
					post.Summary = cleanLine
					break
				}
			}
		}

		// Render Content
		post.Content = mdToHTML(parts[2])
	}
	return post
}

func getPosts() []BlogPost {
	var posts []BlogPost
	files, err := os.ReadDir("posts")
	if err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				content, _ := os.ReadFile(filepath.Join("posts", f.Name()))
				posts = append(posts, parseMarkdown(f.Name(), string(content)))
			}
		}
	}
	return posts
}

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
		"Posts":    getPosts(),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/blog/")

	// Handle Index Page (/blog/)
	if slug == "" {
		data := map[string]interface{}{
			"Title":   "Engineering Notes",
			"Posts":   getPosts(),
			"IsIndex": true,
		}
		tmpl, err := template.ParseFiles(filepath.Join("api", "blog.html"))
		if err == nil {
			tmpl.Execute(w, data)
		}
		return
	}

	// Read file from local disk
	content, err := os.ReadFile(filepath.Join("posts", slug+".md"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	post := parseMarkdown(slug+".md", string(content))

	// Prepare data with both the single post and the list of all posts
	data := map[string]interface{}{
		"Title":   post.Title,
		"Date":    post.Date,
		"Content": post.Content,
		"Posts":   getPosts(),
	}

	// Parse the new blog template
	tmpl, err := template.ParseFiles(filepath.Join("api", "blog.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
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

	// Rute Blog Post
	http.HandleFunc("/blog/", handlePost)

	log.Println("Server Localhost berjalan di http://localhost:8080 🚀")
	log.Println("Struktur folder sudah disesuaikan dengan Vercel (api & public)")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

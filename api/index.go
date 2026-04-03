package handler

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
)

// UBAH DISINI: Sesuaikan dengan nama file baru
//
//go:embed template.html
var htmlContent string

//go:embed blog.html
var blogContent string

//go:embed posts/*.md
var postsFS embed.FS

//go:embed wins.html
var winsContent string

// BlogPost holds the data parsed from markdown files
type BlogPost struct {
	Slug    string
	Title   string
	Date    string
	Summary string // First few lines of the content
	Content template.HTML
}

func mdToHTML(md string) template.HTML {
	// 1. Escape HTML to prevent XSS (basic protection)
	html := template.HTMLEscapeString(md)

	// 2. Headers (# H1, ## H2, ### H3)
	re := regexp.MustCompile(`(?m)^#{3}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h3>$1</h3>")
	re = regexp.MustCompile(`(?m)^#{2}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h2>$1</h2>")
	re = regexp.MustCompile(`(?m)^#{1}\s+(.*)$`)
	html = re.ReplaceAllString(html, "<h1>$1</h1>")

	// 3. Bold (**text**)
	re = regexp.MustCompile(`\*\*(.*?)\*\*`)
	html = re.ReplaceAllString(html, "<strong>$1</strong>")

	// 4. Italic (*text*)
	re = regexp.MustCompile(`\*(.*?)\*`)
	html = re.ReplaceAllString(html, "<em>$1</em>")

	// 5. Paragraphs (Double newlines)
	// Replace double newlines with paragraph tags, but ignore lines that are already headers
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

		// Render full content to HTML
		post.Content = mdToHTML(parts[2])
	}
	return post
}

func getPosts() []BlogPost {
	var posts []BlogPost
	files, err := fs.ReadDir(postsFS, "posts")
	if err == nil {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				content, _ := fs.ReadFile(postsFS, "posts/"+f.Name())
				posts = append(posts, parseMarkdown(f.Name(), string(content)))
			}
		}
	}
	return posts
}

// Handler is the entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// TAMBAHKAN ROUTING WINS DISINI
	if r.URL.Path == "/wins" {
		handleWins(w, r)
		return
	}

	// Simple Routing bawaan kamu
	if strings.HasPrefix(r.URL.Path, "/blog") {
		handlePost(w, r)
	} else {
		handleHome(w, r)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	// Normalize path: /blog/slug -> slug, /blog -> ""
	path := strings.TrimPrefix(r.URL.Path, "/blog")
	slug := strings.TrimPrefix(path, "/")

	if slug == "" {
		// BLOG INDEX PAGE
		data := map[string]interface{}{
			"Title":   "Engineering Notes",
			"Posts":   getPosts(),
			"IsIndex": true,
		}
		tmpl, _ := template.New("blog").Parse(blogContent)
		tmpl.Execute(w, data)
		return
	}

	// Read specific file
	content, err := fs.ReadFile(postsFS, "posts/"+slug+".md")
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

	tmpl, _ := template.New("blog").Parse(blogContent)
	tmpl.Execute(w, data)
}

func handleHome(w http.ResponseWriter, r *http.Request) {

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
		"Posts":    getPosts(),
	}

	// Tampilkan
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TAMBAHKAN FUNGSI INI DI BAWAH
func handleWins(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("wins").Parse(winsContent)
	if err != nil {
		http.Error(w, "Maaf, ada kesalahan sistem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Tampilkan halaman tanpa data dinamis (karena statis)
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

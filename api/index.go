package handler

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Kredensial Supabase
const supabaseUrl = "https://kgscotrveqoixnufzxea.supabase.co"
const supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imtnc2NvdHJ2ZXFvaXhudWZ6eGVhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzUzMDA3MjIsImV4cCI6MjA5MDg3NjcyMn0.2cjGOOcuyxE1z-5yhQo1epzfFd92nGPBDgPshCTbBi8"

// Ganti dengan API Key Resend milikmu yang asli

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

//go:embed clean.html
var cleanContent string

//go:embed about.html
var aboutContent string

// --- STRUKTUR DATA ---
type BlogPost struct {
	Slug       string        `json:"slug"`
	Title      string        `json:"title"`
	Date       string        `json:"created_at"`
	Category   []string      `json:"category"` // 👈 DIUBAH DARI string MENJADI []string (Array)
	RawContent string        `json:"content"`
	Language   string        `json:"language"`
	Summary    string        `json:"-"`
	Content    template.HTML `json:"-"`
}

type Book struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Author   string  `json:"author"`
	CoverURL string  `json:"cover_url"`
	Rating   float64 `json:"rating"`
	Review   string  `json:"review"`
	PostSlug string  `json:"post_slug"`
	IsPinned bool    `json:"is_pinned"`
}

type EmailRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Desc1 string `json:"desc_1"`
	Desc2 string `json:"desc_2"`
	Desc3 string `json:"desc_3"`
}

type Metric struct {
	Label      string   `json:"label"`
	Value      string   `json:"value"`
	ValueColor string   `json:"value_color"` // Opsional: untuk menaruh class CSS seperti "text-terracotta"
	Desc       string   `json:"desc"`
	TechTags   []string `json:"tech_tags"` // Array untuk tag bahasa pemrograman
}

type SmallWin struct {
	Title          string   `json:"title"`
	TitleHighlight string   `json:"title_highlight"`
	Subtitle       string   `json:"subtitle"`
	Metrics        []Metric `json:"metrics"`
	DemoURL        string   `json:"demo_url"`
	GitHubURL      string   `json:"github_url"`
	EmbedDemo      bool     `json:"-"`
}

// --- FUNGSI PARSER MARKDOWN ---
var (
	reBold        = regexp.MustCompile(`\*\*(.*?)\*\*`)
	reItalic      = regexp.MustCompile(`\*(.*?)\*`)
	reCode        = regexp.MustCompile("`([^`]+)`")
	reLink        = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reBlockquote  = regexp.MustCompile(`^>+\s*`)
)

func applyInline(s string) string {
	s = reBold.ReplaceAllString(s, "<strong>$1</strong>")
	s = reItalic.ReplaceAllString(s, "<em>$1</em>")
	s = reCode.ReplaceAllString(s, "<code>$1</code>")
	s = reLink.ReplaceAllString(s, `<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>`)
	return s
}

func mdToHTML(md string) template.HTML {
	// Normalize pre-escaped HTML entities that may come from the DB or rich-text editors
	md = strings.ReplaceAll(md, "&gt;", ">")
	md = strings.ReplaceAll(md, "&lt;", "<")
	md = strings.ReplaceAll(md, "&amp;", "&")
	md = strings.ReplaceAll(md, "&quot;", "\"")
	lines := strings.Split(md, "\n")
	var result []string
	var listItems []string
	var bqItems []string
	var pending []string

	flushList := func() {
		if len(listItems) == 0 {
			return
		}
		result = append(result, "<ul><li>"+strings.Join(listItems, "</li><li>")+"</li></ul>")
		listItems = nil
	}

	flushBq := func() {
		if len(bqItems) == 0 {
			return
		}
		for _, item := range bqItems {
			result = append(result, "<blockquote>"+item+"</blockquote>")
		}
		bqItems = nil
	}

	flushPending := func() {
		if len(pending) == 0 {
			return
		}
		block := strings.TrimSpace(strings.Join(pending, "\n"))
		pending = nil
		if block == "" {
			return
		}
		if strings.HasPrefix(block, "### ") {
			result = append(result, "<h3>"+applyInline(template.HTMLEscapeString(block[4:]))+"</h3>")
		} else if strings.HasPrefix(block, "## ") {
			result = append(result, "<h2>"+applyInline(template.HTMLEscapeString(block[3:]))+"</h2>")
		} else if strings.HasPrefix(block, "# ") {
			result = append(result, "<h1>"+applyInline(template.HTMLEscapeString(block[2:]))+"</h1>")
		} else {
			// Single \n within a paragraph: HTML collapses to space naturally — no <br> needed
			processed := applyInline(template.HTMLEscapeString(block))
			result = append(result, "<p>"+processed+"</p>")
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			flushPending()
			flushBq()
			item := strings.TrimSpace(trimmed[2:])
			listItems = append(listItems, applyInline(template.HTMLEscapeString(item)))
		} else if reBlockquote.MatchString(trimmed) {
			flushPending()
			flushList()
			item := strings.TrimSpace(reBlockquote.ReplaceAllString(trimmed, ""))
			if item != "" {
				bqItems = append(bqItems, applyInline(template.HTMLEscapeString(item)))
			}
		} else if trimmed == "" {
			flushList()
			flushBq()
			flushPending()
		} else {
			if len(listItems) > 0 {
				flushList()
			}
			if len(bqItems) > 0 {
				flushBq()
			}
			pending = append(pending, line)
		}
	}
	flushList()
	flushBq()
	flushPending()

	return template.HTML(strings.Join(result, "\n"))
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

// Fungsi menarik data Small Wins dari Supabase
func fetchWinsFromSupabase() []SmallWin {
	req, _ := http.NewRequest("GET", supabaseUrl+"/rest/v1/small_wins?order=created_at.desc", nil)
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	var wins []SmallWin
	if err != nil || resp.StatusCode != 200 {
		log.Println("Error fetching small wins:", err)
		return wins
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Keajaiban JSONB: Supabase mengirim JSON, Golang otomatis membedahnya ke dalam Struct!
	json.Unmarshal(body, &wins)
	return wins
}

// --- HANDLERS (LOGIKA TAMPILAN) ---
func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(htmlContent)

	data := map[string]interface{}{
		"Title":    "Abiyyu Hanief | Home",
		"Name":     "Abiyyu Hanief",
		"Role":     "Product Implementator & Fullstack Developer",
		"Headline": "Empowering Communities through Tech & Process.",
		"About":    "A problem-solver who uses technology and process to empower communities. My approach, refined through experiences in both program coordination and software development, is to own a solution from concept to completion.",
		"Posts":    fetchPostsFromSupabase(),
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func handleWins(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"mod": func(i, j int) int { return i % j },
		"add": func(i, j int) int { return i + j },
	}

	tmpl := template.New("base").Funcs(funcMap)
	tmpl, _ = tmpl.Parse(baseContent)
	tmpl, err := tmpl.Parse(winsContent)
	if err != nil {
		http.Error(w, "Error loading HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wins := fetchWinsFromSupabase()
	for i := range wins {
		full := strings.TrimSpace(wins[i].Title + " " + wins[i].TitleHighlight)
		if strings.Contains(full, "Support Operations") {
			wins[i].DemoURL = "/demos/cs-dashboard.html"
			wins[i].EmbedDemo = true
		} else if strings.Contains(full, "Local-First Business") {
			wins[i].DemoURL = "/demos/mammos.html"
			wins[i].EmbedDemo = true
		} else if strings.Contains(full, "AI-Powered Customer") {
			wins[i].DemoURL = "/demos/cs-dashboard.html"
			wins[i].EmbedDemo = false
		}
	}

	data := map[string]interface{}{
		"Title":     "Projects | Abiyyu Hanief",
		"SmallWins": wins,
	}

	tmpl.ExecuteTemplate(w, "base", data)
}

func handleLayers(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(gameContent)
	tmpl.ExecuteTemplate(w, "base", nil)
}

func handleLibrary(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(libraryContent)

	allBooks := fetchBooksFromSupabase()

	var pinned []Book
	var others []Book

	// Pisahkan buku yang di-pin dan tidak
	for _, b := range allBooks {
		if b.IsPinned {
			pinned = append(pinned, b)
		} else {
			others = append(others, b)
		}
	}

	// Acak hanya buku yang TIDAK di-pin
	rgen := rand.New(rand.NewSource(time.Now().UnixNano()))
	rgen.Shuffle(len(others), func(i, j int) {
		others[i], others[j] = others[j], others[i]
	})

	// Gabungkan kembali: Pinned di atas, Others (yang sudah diacak) di bawah
	finalBooks := append(pinned, others...)

	data := map[string]interface{}{
		"Title": "Digital Library | Abiyyu Hanief",
		"Books": finalBooks,
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(aboutContent)
	data := map[string]interface{}{
		"Title": "About — Abiyyu Hanief",
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func handleClean(w http.ResponseWriter, r *http.Request) {
	// Karena ini halaman mandiri, kita tidak memakai base.html
	tmpl, _ := template.New("clean").Parse(cleanContent)
	tmpl.Execute(w, nil)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	// PERBAIKAN: Cara mengambil slug yang lebih kebal error
	path := strings.TrimPrefix(r.URL.Path, "/blog")
	slug := strings.TrimPrefix(path, "/")

	tmpl, _ := template.New("base").Parse(baseContent)
	tmpl, _ = tmpl.Parse(blogContent)

	// Halaman Daftar Artikel (Index)
	if slug == "" {
		data := map[string]interface{}{
			"Title":   "Notes",
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

func handleSendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqData EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Buat template HTML email yang rapi dan elegan
	htmlBody := fmt.Sprintf(`
	<div style="font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; color: #333; max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 20px; border: 1px solid #e0e0e0; border-radius: 10px;">
		<h2 style="color: #5A9A8F; text-align: center;">Your 3 Layers</h2>
		<p>Hello <strong>%s</strong>,</p>
		<p>Thank you for exploring your 3 Layers. Here is a copy of your psychological profile results:</p>
		
		<div style="background-color: #f9f9f9; padding: 15px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #C6743E;">
			<p style="margin-top: 0;"><strong>Layer 1 (The Persona):</strong><br>%s</p>
			<p><strong>Layer 2 (The Impression):</strong><br>%s</p>
			<p style="margin-bottom: 0;"><strong>Layer 3 (The Core Self):</strong><br>%s</p>
		</div>
		
		<br>
		<p style="border-top: 1px solid #eee; padding-top: 15px; font-size: 0.9em; color: #777;">
			Warm regards,<br>
			<strong>Abiyyu Hanief</strong><br>
			<a href="https://abiyyuhanief.id" style="color: #5A9A8F; text-decoration: none;">abiyyuhanief.id</a>
		</p>
	</div>`, reqData.Name, reqData.Desc1, reqData.Desc2, reqData.Desc3)

	// Rakit Payload (Data) untuk API Resend
	// Rakit Payload (Data) untuk API Resend tanpa Attachment
	resendPayload := map[string]interface{}{
		"from":    "hello@abiyyuhanief.id",
		"to":      []string{reqData.Email},
		"subject": "Your 3 Layers Psychological Profile",
		"html":    htmlBody,
	}

	resendApiKey := os.Getenv("RESEND_API_KEY") //

	payloadBytes, _ := json.Marshal(resendPayload)
	client := &http.Client{}

	// Tembak ke API Resend
	req, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+resendApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending email via Resend:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Kembalikan status sukses ke browser tanpa mengganggu UI
	w.WriteHeader(http.StatusOK)
}

// --- ENTRY POINT VERCEL ---
func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/about" {
		handleAbout(w, r)
		return
	}

	if path == "/clean" {
		handleClean(w, r)
		return
	}

	if path == "/wins" {
		http.Redirect(w, r, "/projects", http.StatusMovedPermanently)
		return
	}
	if path == "/projects" {
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

	if path == "/send-email" {
		handleSendEmail(w, r)
		return
	}

	// Default rute diarahkan ke Home
	handleHome(w, r)
}

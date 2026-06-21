package main

import (
	"bytes"
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

// Kredensial Supabase (REST API)
const supabaseUrl = "https://kgscotrveqoixnufzxea.supabase.co"
const supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imtnc2NvdHJ2ZXFvaXhudWZ6eGVhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzUzMDA3MjIsImV4cCI6MjA5MDg3NjcyMn0.2cjGOOcuyxE1z-5yhQo1epzfFd92nGPBDgPshCTbBi8"

// Ganti dengan API Key Resend milikmu yang asli

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

// Mesin Markdown Parser
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

// Fungsi Menarik Data Posts dari Supabase via API
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

// Fungsi menarik data Buku dari Supabase
func fetchBooksFromSupabase() []Book {
	req, _ := http.NewRequest("GET", supabaseUrl+"/rest/v1/books?order=created_at.desc", nil)
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	var books []Book
	if err != nil || resp.StatusCode != 200 {
		log.Println("Error fetching books:", err)
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

// --- HANDLERS ---

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("api/base.html", "api/template.html")
	if err != nil {
		http.Error(w, "Error loading HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

	tmpl := template.New("base.html").Funcs(funcMap)
	tmpl, err := tmpl.ParseFiles("api/base.html", "api/wins.html")
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
	tmpl, err := template.ParseFiles("api/base.html", "api/game.html")
	if err != nil {
		http.Error(w, "Error loading HTML", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", nil)
}

func handleLibrary(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("api/base.html", "api/library.html")
	if err != nil {
		http.Error(w, "Error loading HTML", http.StatusInternalServerError)
		return
	}

	allBooks := fetchBooksFromSupabase()

	var pinned []Book
	var others []Book
	for _, b := range allBooks {
		if b.IsPinned {
			pinned = append(pinned, b)
		} else {
			others = append(others, b)
		}
	}

	rgen := rand.New(rand.NewSource(time.Now().UnixNano()))
	rgen.Shuffle(len(others), func(i, j int) {
		others[i], others[j] = others[j], others[i]
	})

	data := map[string]interface{}{
		"Title": "Digital Library | Abiyyu Hanief",
		"Books": append(pinned, others...),
	}

	tmpl.ExecuteTemplate(w, "base", data)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/blog/")

	if slug == "" {
		data := map[string]interface{}{
			"Title":   "Notes",
			"Posts":   fetchPostsFromSupabase(),
			"IsIndex": true,
		}
		tmpl, err := template.ParseFiles("api/base.html", "api/blog.html")
		if err == nil {
			tmpl.ExecuteTemplate(w, "base", data)
		}
		return
	}

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

	tmpl, err := template.ParseFiles("api/base.html", "api/blog.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("api/base.html", "api/about.html")
	if err != nil {
		http.Error(w, "Error loading HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Title": "About — Abiyyu Hanief",
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// FUNGSI BARU: Menangani Halaman Cleaning Mode
func handleClean(w http.ResponseWriter, r *http.Request) {
	// Karena ini halaman mandiri, kita tidak menggunakan base.html
	tmpl, err := template.ParseFiles("api/clean.html")
	if err != nil {
		http.Error(w, "Error loading HTML", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
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
		
		<p><em>Note: We've also attached a snapshot of your result card to this email so you can easily share it to your socials!</em></p>
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

	resendApiKey := os.Getenv("RESEND_API_KEY") // Ambil API Key Resend dari environment variable

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

func main() {
	// Routing Aset Statis
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("public/css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("public/js"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("public/img"))))
	http.Handle("/demos/", http.StripPrefix("/demos/", http.FileServer(http.Dir("public/demos"))))

	// Routing Halaman Utama
	http.HandleFunc("/wins", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/projects", http.StatusMovedPermanently)
	})
	http.HandleFunc("/projects", handleWins)
	http.HandleFunc("/layers", handleLayers)
	http.HandleFunc("/blog/", handlePost)
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/library", handleLibrary)
	http.HandleFunc("/send-email", handleSendEmail)

	http.HandleFunc("/about", handleAbout)

	// ROUTE BARU: Cleaning Mode
	http.HandleFunc("/clean", handleClean)

	// Endpoint PWA
	http.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/manifest.json")
	})
	http.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "public/sw.js")
	})

	log.Println("Server Localhost berjalan di http://localhost:8080 🚀")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

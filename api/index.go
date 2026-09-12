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
	"sync"
	"time"
)

// Kredensial Supabase
const supabaseUrl = "https://kgscotrveqoixnufzxea.supabase.co"
const supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Imtnc2NvdHJ2ZXFvaXhudWZ6eGVhIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzUzMDA3MjIsImV4cCI6MjA5MDg3NjcyMn0.2cjGOOcuyxE1z-5yhQo1epzfFd92nGPBDgPshCTbBi8"

// Shared client so a slow Supabase response can't hang a page forever.
var httpClient = &http.Client{Timeout: 10 * time.Second}

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

//go:embed collaboration.html
var collaborationContent string

//go:embed notfound.html
var notFoundContent string

// --- STRUKTUR DATA ---
type BlogPost struct {
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Date        string        `json:"created_at"`
	Category    []string      `json:"category"` // 👈 DIUBAH DARI string MENJADI []string (Array)
	RawContent  string        `json:"content"`
	Language    string        `json:"language"`
	Summary     string        `json:"-"`
	ReadMinutes int           `json:"-"`
	Content     template.HTML `json:"-"`
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

type ContactRequest struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Message           string `json:"message"`
	CollaborationType string `json:"collaboration_type"`
	Website           string `json:"website"` // honeypot: must stay empty
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
}

// --- TEMPLATE RENDERING ---
var funcs = template.FuncMap{
	"add": func(i, j int) int { return i + j },
	// year returns the YYYY part of a YYYY-MM-DD date.
	"year": func(date string) string {
		if len(date) >= 4 {
			return date[:4]
		}
		return date
	},
	// techTags collects every tech tag across a project's metrics.
	"techTags": func(metrics []Metric) []string {
		var tags []string
		for _, m := range metrics {
			tags = append(tags, m.TechTags...)
		}
		return tags
	},
}

// page builds the data every page shares: SEO metadata and the active nav item.
func page(path, title, description, nav string) map[string]interface{} {
	return map[string]interface{}{
		"Title":       title,
		"Description": description,
		"Path":        path,
		"Nav":         nav,
		"OGType":      "website",
	}
}

func render(w http.ResponseWriter, content string, data map[string]interface{}) {
	tmpl, err := template.New("base").Funcs(funcs).Parse(baseContent)
	if err == nil {
		tmpl, err = tmpl.Parse(content)
	}
	if err != nil {
		http.Error(w, "Error loading HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Println("Error rendering template:", err)
	}
}

// --- FUNGSI PARSER MARKDOWN ---
var (
	reBold       = regexp.MustCompile(`\*\*(.*?)\*\*`)
	reItalic     = regexp.MustCompile(`\*(.*?)\*`)
	reCode       = regexp.MustCompile("`([^`]+)`")
	reLink       = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reBlockquote = regexp.MustCompile(`^>+\s*`)
	reRule       = regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`)
	reMarkup     = regexp.MustCompile("[*_`#>]+")
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
			result = append(result, "<h2>"+applyInline(template.HTMLEscapeString(block[2:]))+"</h2>")
		} else {
			// Single \n within a paragraph: HTML collapses to space naturally — no <br> needed
			processed := applyInline(template.HTMLEscapeString(block))
			result = append(result, "<p>"+processed+"</p>")
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if reRule.MatchString(trimmed) {
			flushPending()
			flushList()
			flushBq()
			result = append(result, "<hr>")
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
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

// summarize returns the first prose line of a post, stripped of markdown and
// trimmed at a word boundary.
func summarize(md string) string {
	for _, line := range strings.Split(md, "\n") {
		clean := strings.TrimSpace(line)
		if clean == "" || strings.HasPrefix(clean, "#") || reRule.MatchString(clean) {
			continue
		}
		clean = strings.TrimSpace(reMarkup.ReplaceAllString(clean, ""))
		runes := []rune(clean)
		if len(runes) <= 180 {
			return clean
		}
		cut := string(runes[:180])
		if i := strings.LastIndex(cut, " "); i > 120 {
			cut = cut[:i]
		}
		return strings.TrimRight(cut, ".,;:—- ") + "…"
	}
	return ""
}

// --- FUNGSI FETCH SUPABASE ---
func supabaseGet(path string, out interface{}) bool {
	req, _ := http.NewRequest("GET", supabaseUrl+"/rest/v1/"+path, nil)
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("Error fetching", path, err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("Supabase returned", resp.StatusCode, "for", path)
		return false
	}
	body, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(body, out) == nil
}

func fetchPostsFromSupabase() []BlogPost {
	var posts []BlogPost
	supabaseGet("posts?order=created_at.desc", &posts)

	for i := range posts {
		if len(posts[i].Date) >= 10 {
			posts[i].Date = posts[i].Date[:10]
		}
		posts[i].Summary = summarize(posts[i].RawContent)
		words := len(strings.Fields(posts[i].RawContent))
		posts[i].ReadMinutes = (words + 199) / 200
		if posts[i].ReadMinutes < 1 {
			posts[i].ReadMinutes = 1
		}
		posts[i].Content = mdToHTML(posts[i].RawContent)
	}
	return posts
}

func fetchBooksFromSupabase() []Book {
	var books []Book
	supabaseGet("books?order=created_at.desc", &books)
	return books
}

// Fungsi menarik data Small Wins dari Supabase
func fetchWinsFromSupabase() []SmallWin {
	var wins []SmallWin
	// Keajaiban JSONB: Supabase mengirim JSON, Golang otomatis membedahnya ke dalam Struct!
	supabaseGet("small_wins?order=created_at.desc", &wins)
	return wins
}

// --- HANDLERS (LOGIKA TAMPILAN) ---
func handleHome(w http.ResponseWriter, r *http.Request) {
	var posts []BlogPost
	var books []Book
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); posts = fetchPostsFromSupabase() }()
	go func() { defer wg.Done(); books = fetchBooksFromSupabase() }()
	wg.Wait()

	data := page("/", "Abiyyu Hanief — Product implementation & fullstack development",
		"Abiyyu Hanief builds systems that work for people — product implementation, fullstack development, and community programs across Indonesia.", "home")
	data["Projects"] = featuredProjects()
	if len(posts) > 3 {
		posts = posts[:3]
	}
	data["Posts"] = posts
	if len(books) > 0 {
		data["LatestBook"] = books[0]
	}
	data["IsHome"] = true
	render(w, htmlContent, data)
}

func handleWins(w http.ResponseWriter, r *http.Request) {
	wins := fetchWinsFromSupabase()
	for i := range wins {
		full := strings.TrimSpace(wins[i].Title + " " + wins[i].TitleHighlight)
		if strings.Contains(full, "Support Operations") || strings.Contains(full, "AI-Powered Customer") {
			wins[i].DemoURL = "/demos/cs-dashboard.html"
		} else if strings.Contains(full, "Local-First Business") {
			wins[i].DemoURL = "/demos/mammos.html"
		}
	}

	data := page("/projects", "Projects — Abiyyu Hanief",
		"Selected work by Abiyyu Hanief: websites, CMS platforms, POS and inventory systems, dashboards, and data work — each one shipped and measured.", "projects")
	data["Projects"] = projects
	data["SmallWins"] = wins
	render(w, winsContent, data)
}

func handleLayers(w http.ResponseWriter, r *http.Request) {
	data := page("/layers", "3 Layers — Abiyyu Hanief",
		"Discover your 3 Layers: a 60-second psychological icebreaker based on the Barnum Effect. Built for fun and introspection.", "")
	render(w, gameContent, data)
}

func handleLibrary(w http.ResponseWriter, r *http.Request) {
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

	data := page("/library", "Library — Abiyyu Hanief",
		"Books Abiyyu Hanief has read, with ratings and short reviews. My way of thought, maybe.", "library")
	data["Books"] = finalBooks
	render(w, libraryContent, data)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	data := page("/about", "About — Abiyyu Hanief",
		"Abiyyu (Abi) Hanief — Product Implementation Specialist at Nusatek and Information Systems graduate of Universitas Indonesia, working where technology meets people.", "about")
	render(w, aboutContent, data)
}

func handleClean(w http.ResponseWriter, r *http.Request) {
	// Karena ini halaman mandiri, kita tidak memakai base.html
	tmpl, _ := template.New("clean").Parse(cleanContent)
	tmpl.Execute(w, nil)
}

func handleCollaboration(w http.ResponseWriter, r *http.Request) {
	data := page("/collaboration", "Collaboration — Abiyyu Hanief",
		"Work with Abiyyu Hanief on business operations systems or community and branding websites. First discovery conversation is free.", "collaboration")
	data["Projects"] = projects
	render(w, collaborationContent, data)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := page(r.URL.Path, "Page not found — Abiyyu Hanief", "", "")
	data["NoIndex"] = true
	render(w, notFoundContent, data)
}

func handleContactSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bad request"})
		return
	}

	// Honeypot: bots fill hidden fields, humans never see this one
	if strings.TrimSpace(req.Website) != "" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Message = strings.TrimSpace(req.Message)

	emailRe := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if req.Name == "" || req.Email == "" || req.Message == "" || !emailRe.MatchString(req.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Please fill in your name, a valid email, and a message."})
		return
	}

	submission := map[string]interface{}{
		"name":               req.Name,
		"email":              req.Email,
		"message":            req.Message,
		"collaboration_type": req.CollaborationType,
	}
	payloadBytes, _ := json.Marshal(submission)

	insertReq, _ := http.NewRequest("POST", supabaseUrl+"/rest/v1/contact_submissions", bytes.NewBuffer(payloadBytes))
	insertReq.Header.Add("apikey", supabaseKey)
	insertReq.Header.Add("Authorization", "Bearer "+supabaseKey)
	insertReq.Header.Add("Content-Type", "application/json")
	insertReq.Header.Add("Prefer", "return=representation")

	insertResp, err := httpClient.Do(insertReq)
	if err != nil || insertResp.StatusCode >= 300 {
		log.Println("Error persisting contact submission:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Could not save your message. Please try again."})
		return
	}
	defer insertResp.Body.Close()

	var inserted []map[string]interface{}
	body, _ := io.ReadAll(insertResp.Body)
	json.Unmarshal(body, &inserted)

	var insertedID string
	if len(inserted) > 0 {
		if id, ok := inserted[0]["id"].(string); ok {
			insertedID = id
		}
	}

	go sendContactNotification(req, insertedID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func markContactEmailSent(id string) {
	if id == "" {
		return
	}
	payload, _ := json.Marshal(map[string]bool{"email_sent": true})
	req, _ := http.NewRequest("PATCH", supabaseUrl+"/rest/v1/contact_submissions?id=eq."+id, bytes.NewBuffer(payload))
	req.Header.Add("apikey", supabaseKey)
	req.Header.Add("Authorization", "Bearer "+supabaseKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Prefer", "return=minimal")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("Error marking contact submission as email_sent:", err)
		return
	}
	defer resp.Body.Close()
}

func sendContactNotification(req ContactRequest, submissionID string) {
	htmlBody := fmt.Sprintf(`
	<div style="font-family:'Helvetica Neue',Helvetica,Arial,sans-serif; color:#333; max-width:600px; margin:0 auto;">
		<h2 style="color:#3F756C;">New Collaboration Inquiry</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><strong>Type:</strong> %s</p>
		<p><strong>Message:</strong><br>%s</p>
	</div>`, req.Name, req.Email, req.CollaborationType, req.Message)

	resendPayload := map[string]interface{}{
		"from":    "hello@abiyyuhanief.id",
		"to":      []string{"nafidzabiyyu@gmail.com"},
		"subject": "New Collaboration Inquiry from " + req.Name,
		"html":    htmlBody,
	}

	resendApiKey := os.Getenv("RESEND_API_KEY")
	payloadBytes, _ := json.Marshal(resendPayload)

	httpReq, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(payloadBytes))
	httpReq.Header.Set("Authorization", "Bearer "+resendApiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Println("Error sending contact notification via Resend:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		markContactEmailSent(submissionID)
	} else {
		log.Println("Resend returned non-2xx status for contact notification:", resp.StatusCode)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	slug := strings.Trim(strings.TrimPrefix(r.URL.Path, "/blog"), "/")
	posts := fetchPostsFromSupabase()

	// Halaman Daftar Artikel (Index)
	if slug == "" {
		data := page("/blog", "Notes — Abiyyu Hanief",
			"Notes by Abiyyu Hanief: reflections, book reviews, and things learned while building — written in Indonesian and English.", "notes")
		data["Posts"] = posts
		render(w, blogContent, data)
		return
	}

	// Halaman Detail Artikel
	var post *BlogPost
	var more []BlogPost
	for i := range posts {
		if posts[i].Slug == slug {
			post = &posts[i]
		} else if len(more) < 3 {
			more = append(more, posts[i])
		}
	}
	if post == nil {
		handleNotFound(w, r)
		return
	}

	desc := post.Summary
	if desc == "" {
		desc = "A note by Abiyyu Hanief."
	}
	data := page("/blog/"+post.Slug, post.Title+" — Abiyyu Hanief", desc, "notes")
	data["OGType"] = "article"
	data["Post"] = post
	data["More"] = more
	render(w, blogContent, data)
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

	// Rakit Payload (Data) untuk API Resend tanpa Attachment
	resendPayload := map[string]interface{}{
		"from":    "hello@abiyyuhanief.id",
		"to":      []string{reqData.Email},
		"subject": "Your 3 Layers Psychological Profile",
		"html":    htmlBody,
	}

	resendApiKey := os.Getenv("RESEND_API_KEY") //

	payloadBytes, _ := json.Marshal(resendPayload)

	// Tembak ke API Resend
	req, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+resendApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println("Error sending email via Resend:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Kembalikan status sukses ke browser tanpa mengganggu UI
	w.WriteHeader(http.StatusOK)
}

// --- DATA PROYEK ---
// Vercel builds every file in api/ as its own function entrypoint, so this
// data lives here rather than in a second file.
// Lighthouse holds a full PageSpeed Insights report. All four categories are
// shown together — a single cherry-picked number reads as marketing, a full
// report reads as evidence.
type Lighthouse struct {
	Performance   int
	Accessibility int
	BestPractices int
	SEO           int
	Source        string
}

// Project is a shipped build shown on the home, projects and collaboration
// pages. It lives in code (not Supabase) because each entry carries layout
// decisions — preview image, evidence, links — that are edited alongside the
// templates.
type Project struct {
	Slug      string
	Title     string
	Client    string
	Category  string
	Summary   string
	Scope     []string
	Pages     []string
	URL       string
	URLLabel  string
	DemoURL   string
	Private   bool
	Self      bool
	Image     string
	ImageAlt  string
	Scores    *Lighthouse
	ScoreNote string
	Featured  bool
}

const lighthouseSource = "Lighthouse via PageSpeed Insights, Aug 2026 — reproducible on the live URL"

var projects = []Project{
	{
		Slug:     "sadewa",
		Title:    "Sadewa",
		Client:   "Sayap Dewantara Indonesia",
		Category: "Foundation site + CMS",
		Summary:  "A public site, self-serve CMS, and cookieless first-party analytics for an education foundation — replacing a Wix site and documenting 10 batches of a 15-year teaching program, at Rp 0/month infrastructure cost.",
		Scope:    []string{"Custom CMS", "Privacy-first analytics", "Design system & SEO"},
		Pages:    []string{"Home & About", "GUIM Story Archive", "Articles & Testimonials", "Admin CMS", "Analytics Dashboard"},
		URL:      "https://www.sadewaind.org",
		URLLabel: "sadewaind.org",
		Image:    "/img/work/sadewa.webp",
		ImageAlt: "Homepage of sadewaind.org",
		Scores:   &Lighthouse{99, 96, 100, 100, lighthouseSource},
		Featured: true,
	},
	{
		Slug:      "laksa-bogor",
		Title:     "LAKSA Bogor",
		Client:    "Dinas Pariwisata Kota Bogor",
		Category:  "GovTech & tourism",
		Summary:   "A mobile-first tourism directory that lifted 55 destinations out of an unindexable chatbot iframe into 76 real, searchable pages — plus a backoffice the tourism office runs without a developer.",
		Scope:     []string{"Directory platform", "Admin backoffice", "Visitor analytics"},
		Pages:     []string{"Destination Directory", "Urban Wellness", "Health Access Map", "Unified Search", "Admin CRUD & Analytics"},
		URL:       "https://www.laksabogor.info",
		URLLabel:  "laksabogor.info",
		Image:     "/img/work/laksa-bogor.webp",
		ImageAlt:  "Homepage of laksabogor.info",
		Scores:    &Lighthouse{75, 96, 100, 100, lighthouseSource},
		ScoreNote: "Performance is held down by destination imagery still served from a third-party source — a single known bottleneck, not a structural one. Everything I control scores 96+.",
		Featured:  true,
	},
	{
		Slug:     "mammos",
		Title:    "Mammo's Home Bakery",
		Category: "POS & finance system",
		Summary:  "An offline-first POS for a home bakery — order taking, stock, and cash flow running entirely on-device, no server cost.",
		Scope:    []string{"POS development", "Inventory sync", "Offline-first app"},
		Pages:    []string{"Cashier & Order Entry", "Stock Management", "Daily Reconciliation", "Sales Report"},
		DemoURL:  "/demos/mammos.html",
		Image:    "/img/work/mammos.webp",
		ImageAlt: "Opening frame of the Mammo's POS product demo",
		Featured: true,
	},
	{
		Slug:     "cs-dashboard",
		Title:    "CS Dashboard",
		Client:   "Siswamedia / Tooks",
		Category: "CS & finance dashboard",
		Summary:  "An internal dashboard unifying support tickets and financial reconciliation for a fast-moving team, cutting the gap between prototype and shipped feature.",
		Scope:    []string{"Dashboard development", "Finance reconciliation", "Rapid prototyping"},
		Pages:    []string{"Ticket Overview", "Reconciliation View", "Agent Performance", "Handoff Notes"},
		DemoURL:  "/demos/cs-dashboard.html",
		Image:    "/img/work/cs-dashboard.webp",
		ImageAlt: "Frame from the Siswamedia analytics product demo",
		Featured: true,
	},
	{
		Slug:     "gernas-tastaka",
		Title:    "Gernas Tastaka",
		Category: "Platform migration",
		Summary:  "A WordPress site rebuilt on Next.js and Payload CMS in 11 days — 123 media assets migrated, a fully Indonesian-language dashboard, and bilingual ID/EN pages built from 24 composable content blocks.",
		Scope:    []string{"WordPress migration", "Custom CMS & blocks", "Bilingual ID/EN"},
		Pages:    []string{"Home & Profile", "Programs & Training", "Articles & Gallery", "Partners & Contact", "Indonesian CMS Dashboard"},
		URL:      "https://www.gernastastaka.org",
		URLLabel: "gernastastaka.org",
		Image:    "/img/work/gernas-tastaka.webp",
		ImageAlt: "Homepage of gernastastaka.org",
		Scores:   &Lighthouse{100, 95, 100, 100, lighthouseSource},
	},
	{
		Slug:     "diversity-of-sumatra",
		Title:    "Diversity of Sumatra",
		Category: "Multi-location inventory",
		Summary:  "A stock management system across multiple production sites, replacing fully manual per-transaction logging with a live inventory ledger.",
		Scope:    []string{"Inventory system", "Multi-location sync", "Stock reporting"},
		Pages:    []string{"Stock In / Out", "Location Transfer", "Low-stock Alerts", "Reporting Dashboard"},
		Private:  true,
	},
	{
		Slug:     "abiyyuhanief-id",
		Title:    "abiyyuhanief.id",
		Category: "Portfolio & content platform",
		Summary:  "This site — a Go-powered portfolio and content platform running on the edge, built for near-zero latency and near-zero server cost.",
		Scope:    []string{"Web design", "Web development", "Content platform"},
		Pages:    []string{"Home", "Projects", "Notes", "Library", "About", "Collaboration"},
		URL:      "https://abiyyuhanief.id",
		URLLabel: "abiyyuhanief.id",
		Self:     true,
	},
}

func featuredProjects() []Project {
	var out []Project
	for _, p := range projects {
		if p.Featured {
			out = append(out, p)
		}
	}
	return out
}

// --- ENTRY POINT VERCEL ---
func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}

	switch {
	case path == "/":
		handleHome(w, r)
	case path == "/about":
		handleAbout(w, r)
	case path == "/clean":
		handleClean(w, r)
	case path == "/wins":
		http.Redirect(w, r, "/projects", http.StatusMovedPermanently)
	case path == "/projects":
		handleWins(w, r)
	case path == "/layers":
		handleLayers(w, r)
	case path == "/library":
		handleLibrary(w, r)
	case path == "/collaboration":
		handleCollaboration(w, r)
	case path == "/api/contact":
		handleContactSubmit(w, r)
	case path == "/blog" || strings.HasPrefix(path, "/blog/"):
		handlePost(w, r)
	case path == "/send-email":
		handleSendEmail(w, r)
	default:
		handleNotFound(w, r)
	}
}

package handler

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

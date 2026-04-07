package main

import (
	"archive/zip"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	htmltpl "html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

// --- Data Model ---

type GalleryConfig struct {
	CaptureOneURL string `json:"capture_one_url"`
	GalleryTitle  string `json:"gallery_title"`
	Subtitle      string `json:"subtitle"`
	Slug          string `json:"slug"`
	CoverIndex    int    `json:"cover_index"`
	IsPrivate     bool   `json:"is_private"`
	SecretToken   string `json:"secret_token"`
	// Appearance
	BackgroundColor string `json:"background_color"`
	CardColor       string `json:"card_color"`
	TextColor       string `json:"text_color"`
	AccentColor     string `json:"accent_color"`
	// Photo display
	FrameStyle   string `json:"frame_style"`
	BorderRadius string `json:"border_radius"`
	BorderWidth  string `json:"border_width"`
	BorderColor  string `json:"border_color"`
	Shadow       string `json:"shadow"`
	HoverEffect  string `json:"hover_effect"`
	// Layout
	Layout     string `json:"layout"`
	ColumnGap  string `json:"column_gap"`
	MaxColumns int    `json:"max_columns"`
	// Watermark/branding
	LogoURL       string `json:"logo_url"`
	FooterText    string `json:"footer_text"`
	ShowFilenames bool   `json:"show_filenames"`
	// Lightbox
	LightboxBg string `json:"lightbox_bg"`
	// Slideshow
	SlideshowTransition string `json:"slideshow_transition"`
}

type Photo struct {
	UUID        string `json:"uuid"`
	DisplayName string `json:"display_name"`
	SmallURL    string `json:"small_url"`
	MediumURL   string `json:"medium_url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Rating      int    `json:"rating"`
	CreatedAt   string `json:"created_at"`
}

type Gallery struct {
	ID      string        `json:"id"`
	Config  GalleryConfig `json:"config"`
	Photos  []Photo       `json:"photos"`
	SongID  string        `json:"song_id,omitempty"`
	SongIDs []string      `json:"song_ids,omitempty"`
}

type Song struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Duration float64 `json:"duration"`
	Filename string  `json:"filename"`
	Source   string  `json:"source"`
}

type SiteConfig struct {
	SiteTitle    string `json:"site_title"`
	SiteSubtitle string `json:"site_subtitle"`
	LogoURL      string `json:"logo_url"`
	AccentColor  string `json:"accent_color"`
	BgColor      string `json:"bg_color"`
	TextColor    string `json:"text_color"`
	CardColor    string `json:"card_color"`
}

type AdminAuth struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type AppData struct {
	Site      SiteConfig `json:"site"`
	Auth      AdminAuth  `json:"auth"`
	Galleries []Gallery  `json:"galleries"`
	Songs     []Song     `json:"songs,omitempty"`
}

var (
	appData  AppData
	dataFile string
	mu       sync.RWMutex
	sessions = make(map[string]time.Time) // session token -> expiry
	sessMu   sync.Mutex
)

// --- Helpers ---

func defaultSiteConfig() SiteConfig {
	return SiteConfig{
		SiteTitle:    "Photo Gallery",
		SiteSubtitle: "A collection of moments",
		AccentColor:  "#c8a97e",
		BgColor:      "#0a0a0a",
		TextColor:    "#f0f0f0",
		CardColor:    "#1a1a1a",
	}
}

func defaultGalleryConfig() GalleryConfig {
	return GalleryConfig{
		GalleryTitle:    "New Gallery",
		BackgroundColor: "#0a0a0a",
		CardColor:       "#1a1a1a",
		TextColor:       "#f0f0f0",
		AccentColor:     "#c8a97e",
		FrameStyle:      "none",
		BorderRadius:    "4px",
		BorderWidth:     "0px",
		BorderColor:     "#333333",
		Shadow:          "0 8px 32px rgba(0,0,0,0.4)",
		HoverEffect:     "lift",
		Layout:          "masonry",
		ColumnGap:       "16px",
		MaxColumns:      4,
		ShowFilenames:   false,
		LightboxBg:          "rgba(0,0,0,0.95)",
		SlideshowTransition: "fade",
	}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashPassword(pw string) string {
	h := sha256.Sum256([]byte(pw))
	return hex.EncodeToString(h[:])
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func loadData() {
	raw, err := os.ReadFile(dataFile)
	if err != nil {
		appData = AppData{Site: defaultSiteConfig()}
		return
	}
	if err := json.Unmarshal(raw, &appData); err != nil {
		appData = AppData{Site: defaultSiteConfig()}
		return
	}
	if appData.Site.AccentColor == "" {
		appData.Site = defaultSiteConfig()
	}
	// Set default auth if not configured
	if appData.Auth.Username == "" {
		appData.Auth.Username = "admin"
		appData.Auth.PasswordHash = hashPassword("admin")
		saveData()
	}
	// Ensure all galleries have secret tokens
	changed := false
	for i := range appData.Galleries {
		if appData.Galleries[i].Config.SecretToken == "" {
			appData.Galleries[i].Config.SecretToken = generateToken()
			changed = true
		}
	}
	if changed {
		saveData()
	}
	migrateOldData()
}

func migrateOldData() {
	if len(appData.Galleries) > 0 {
		return
	}
	type oldFormat struct {
		Config GalleryConfig `json:"config"`
		Photos []Photo       `json:"photos"`
	}
	raw, err := os.ReadFile(dataFile)
	if err != nil {
		return
	}
	var old oldFormat
	if err := json.Unmarshal(raw, &old); err != nil {
		return
	}
	if old.Config.CaptureOneURL != "" {
		slug := slugify(old.Config.GalleryTitle)
		if slug == "" {
			slug = "gallery"
		}
		old.Config.Slug = slug
		old.Config.SecretToken = generateToken()
		appData.Galleries = append(appData.Galleries, Gallery{
			ID:     generateID(),
			Config: old.Config,
			Photos: old.Photos,
		})
		saveData()
	}
}

func saveData() error {
	raw, err := json.MarshalIndent(appData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataFile, raw, 0644)
}

func findGalleryBySlug(slug string) *Gallery {
	for i := range appData.Galleries {
		if appData.Galleries[i].Config.Slug == slug {
			return &appData.Galleries[i]
		}
	}
	return nil
}

func findGalleryByID(id string) *Gallery {
	for i := range appData.Galleries {
		if appData.Galleries[i].ID == id {
			return &appData.Galleries[i]
		}
	}
	return nil
}

func findGalleryByToken(token string) *Gallery {
	for i := range appData.Galleries {
		if appData.Galleries[i].Config.SecretToken == token {
			return &appData.Galleries[i]
		}
	}
	return nil
}

func findSongByID(id string) *Song {
	for i := range appData.Songs {
		if appData.Songs[i].ID == id {
			return &appData.Songs[i]
		}
	}
	return nil
}

func mediaDir() string {
	dir := filepath.Join(filepath.Dir(dataFile), "media")
	os.MkdirAll(dir, 0755)
	return dir
}

// --- Auth ---

func createSession() string {
	token := generateToken()
	sessMu.Lock()
	sessions[token] = time.Now().Add(24 * time.Hour)
	sessMu.Unlock()
	return token
}

func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("gallery_session")
	if err != nil {
		return false
	}
	sessMu.Lock()
	expiry, ok := sessions[cookie.Value]
	sessMu.Unlock()
	if !ok || time.Now().After(expiry) {
		return false
	}
	return true
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		errMsg := ""
		if r.URL.Query().Get("error") == "1" {
			errMsg = "Invalid username or password"
		}
		loginTmpl.Execute(w, errMsg)
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	mu.RLock()
	validUser := appData.Auth.Username
	validHash := appData.Auth.PasswordHash
	mu.RUnlock()

	if username != validUser || hashPassword(password) != validHash {
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}

	token := createSession()
	http.SetCookie(w, &http.Cookie{
		Name:     "gallery_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("gallery_session")
	if err == nil {
		sessMu.Lock()
		delete(sessions, cookie.Value)
		sessMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "gallery_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()

	mu.Lock()
	newUser := r.FormValue("username")
	newPass := r.FormValue("new_password")
	if newUser != "" {
		appData.Auth.Username = newUser
	}
	if newPass != "" {
		appData.Auth.PasswordHash = hashPassword(newPass)
	}
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin?pw_changed=1", http.StatusSeeOther)
}

// --- Capture One API ---

type c1EstablishResponse struct {
	CloudSession struct {
		UUID        string `json:"uuid"`
		DisplayName string `json:"display_name"`
	} `json:"cloud_session"`
	AccessToken string `json:"access_token"`
}

type c1Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type c1Variant struct {
	UUID        string `json:"uuid"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
	Thumbnails  struct {
		Small  c1Thumbnail `json:"small"`
		Medium c1Thumbnail `json:"medium"`
	} `json:"thumbnails"`
	Rating int `json:"rating"`
}

type c1StateResponse struct {
	Variants      []c1Variant `json:"variants"`
	NextPageToken string      `json:"next_page_token"`
}

func extractGalleryID(url string) string {
	re := regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func syncFromCaptureOne(gallery *Gallery) error {
	galleryURL := gallery.Config.CaptureOneURL
	c1ID := extractGalleryID(galleryURL)
	if c1ID == "" {
		return fmt.Errorf("could not extract gallery ID from URL")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	establishBody := fmt.Sprintf(`{"cloud_session_uuid":"%s"}`, c1ID)
	establishURL := fmt.Sprintf("https://live.captureone.com/api/v1/session/establish/%s/", c1ID)

	resp, err := client.Post(establishURL, "application/json", strings.NewReader(establishBody))
	if err != nil {
		return fmt.Errorf("establish session: %w", err)
	}
	defer resp.Body.Close()

	var establish c1EstablishResponse
	if err := json.NewDecoder(resp.Body).Decode(&establish); err != nil {
		return fmt.Errorf("decode establish: %w", err)
	}

	if establish.CloudSession.DisplayName != "" && gallery.Config.GalleryTitle == "New Gallery" {
		gallery.Config.GalleryTitle = establish.CloudSession.DisplayName
		if gallery.Config.Slug == "" {
			gallery.Config.Slug = slugify(establish.CloudSession.DisplayName)
		}
	}

	var allPhotos []Photo
	nextToken := ""

	for {
		stateBody := fmt.Sprintf(`{"cloud_session_uuid":"%s","order_by":2`, c1ID)
		if nextToken != "" {
			stateBody += fmt.Sprintf(`,"next_page_token":"%s"`, nextToken)
		}
		stateBody += "}"

		req, _ := http.NewRequest("POST", "https://live.captureone.com/api/v1/session/state/", strings.NewReader(stateBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+establish.AccessToken)

		stateResp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("fetch state: %w", err)
		}

		var state c1StateResponse
		if err := json.NewDecoder(stateResp.Body).Decode(&state); err != nil {
			stateResp.Body.Close()
			return fmt.Errorf("decode state: %w", err)
		}
		stateResp.Body.Close()

		for _, v := range state.Variants {
			allPhotos = append(allPhotos, Photo{
				UUID:        v.UUID,
				DisplayName: v.DisplayName,
				SmallURL:    v.Thumbnails.Small.URL,
				MediumURL:   v.Thumbnails.Medium.URL,
				Width:       v.Thumbnails.Medium.Width,
				Height:      v.Thumbnails.Medium.Height,
				Rating:      v.Rating,
				CreatedAt:   v.CreatedAt,
			})
		}

		if state.NextPageToken == "" {
			break
		}
		nextToken = state.NextPageToken
	}

	gallery.Photos = allPhotos
	return nil
}

// --- Public Handlers ---

type gallerySongInfo struct {
	URL      string  `json:"url"`
	Duration float64 `json:"duration"`
	Title    string  `json:"title"`
}

type galleryPageData struct {
	Gallery
	HasSongs      bool
	SongsJSON     string
	TotalDuration float64
}

func getGallerySongIDs(g *Gallery) []string {
	if len(g.SongIDs) > 0 {
		return g.SongIDs
	}
	// Migrate old single SongID
	if g.SongID != "" {
		return []string{g.SongID}
	}
	return nil
}

func buildGalleryPage(g *Gallery) galleryPageData {
	d := galleryPageData{Gallery: *g}
	ids := getGallerySongIDs(g)
	var songs []gallerySongInfo
	var total float64
	for _, id := range ids {
		s := findSongByID(id)
		if s != nil {
			songs = append(songs, gallerySongInfo{
				URL:      "/media/" + s.Filename,
				Duration: s.Duration,
				Title:    s.Title,
			})
			total += s.Duration
		}
	}
	if len(songs) > 0 {
		d.HasSongs = true
		d.TotalDuration = total
		raw, _ := json.Marshal(songs)
		d.SongsJSON = string(raw)
	}
	return d
}

func handleOverview(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// Handle secret token access: /s/{token}
	if strings.HasPrefix(path, "s/") {
		token := strings.TrimPrefix(path, "s/")
		mu.RLock()
		g := findGalleryByToken(token)
		if g == nil {
			mu.RUnlock()
			http.NotFound(w, r)
			return
		}
		d := buildGalleryPage(g)
		mu.RUnlock()
		galleryTmpl.Execute(w, d)
		return
	}

	if path == "" {
		mu.RLock()
		d := AppData{Site: appData.Site}
		for _, g := range appData.Galleries {
			if !g.Config.IsPrivate {
				d.Galleries = append(d.Galleries, g)
			}
		}
		mu.RUnlock()
		overviewTmpl.Execute(w, d)
		return
	}

	mu.RLock()
	g := findGalleryBySlug(path)
	if g == nil {
		mu.RUnlock()
		http.NotFound(w, r)
		return
	}
	if g.Config.IsPrivate {
		mu.RUnlock()
		http.NotFound(w, r)
		return
	}
	d := buildGalleryPage(g)
	mu.RUnlock()
	galleryTmpl.Execute(w, d)
}

// --- Admin Handlers ---

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	d := struct {
		AppData
		Host string
	}{appData, r.Host}
	mu.RUnlock()
	adminOverviewTmpl.Execute(w, d)
}

func handleAdminNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	c1URL := r.FormValue("capture_one_url")
	if c1URL == "" {
		http.Error(w, "Capture One URL required", 400)
		return
	}

	g := Gallery{
		ID:     generateID(),
		Config: defaultGalleryConfig(),
	}
	g.Config.CaptureOneURL = c1URL
	g.Config.SecretToken = generateToken()
	g.Config.IsPrivate = r.FormValue("is_private") == "on"

	mu.Lock()
	appData.Galleries = append(appData.Galleries, g)
	gPtr := &appData.Galleries[len(appData.Galleries)-1]
	mu.Unlock()

	if err := syncFromCaptureOne(gPtr); err != nil {
		mu.Lock()
		appData.Galleries = appData.Galleries[:len(appData.Galleries)-1]
		mu.Unlock()
		http.Error(w, "Sync failed: "+err.Error(), 500)
		return
	}

	mu.Lock()
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin/gallery/"+g.ID, http.StatusSeeOther)
}

type adminEditData struct {
	Gallery
	Host          string
	Songs         []Song
	SelectedSongs []Song
	TotalDuration float64
}

func handleAdminGallery(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/gallery/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	mu.RLock()
	g := findGalleryByID(id)
	if g == nil {
		mu.RUnlock()
		http.NotFound(w, r)
		return
	}
	var selected []Song
	var totalDur float64
	for _, id := range getGallerySongIDs(g) {
		if s := findSongByID(id); s != nil {
			selected = append(selected, *s)
			totalDur += s.Duration
		}
	}
	d := adminEditData{*g, r.Host, appData.Songs, selected, totalDur}
	mu.RUnlock()
	adminEditTmpl.Execute(w, d)
}

func handleAdminGallerySave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	id := r.FormValue("gallery_id")

	mu.Lock()
	g := findGalleryByID(id)
	if g == nil {
		mu.Unlock()
		http.NotFound(w, r)
		return
	}

	g.Config.CaptureOneURL = r.FormValue("capture_one_url")
	g.Config.GalleryTitle = r.FormValue("gallery_title")
	g.Config.Subtitle = r.FormValue("subtitle")
	g.Config.Slug = slugify(r.FormValue("slug"))
	g.Config.IsPrivate = r.FormValue("is_private") == "on"
	g.Config.FrameStyle = r.FormValue("frame_style")
	g.Config.BackgroundColor = r.FormValue("background_color")
	g.Config.CardColor = r.FormValue("card_color")
	g.Config.TextColor = r.FormValue("text_color")
	g.Config.AccentColor = r.FormValue("accent_color")
	g.Config.BorderRadius = r.FormValue("border_radius")
	g.Config.BorderWidth = r.FormValue("border_width")
	g.Config.BorderColor = r.FormValue("border_color")
	g.Config.Shadow = r.FormValue("shadow")
	g.Config.HoverEffect = r.FormValue("hover_effect")
	g.Config.Layout = r.FormValue("layout")
	g.Config.ColumnGap = r.FormValue("column_gap")
	g.Config.FooterText = r.FormValue("footer_text")
	g.Config.ShowFilenames = r.FormValue("show_filenames") == "on"
	g.Config.LightboxBg = r.FormValue("lightbox_bg")
	g.Config.SlideshowTransition = r.FormValue("slideshow_transition")
	g.Config.LogoURL = r.FormValue("logo_url")

	maxCols := 4
	fmt.Sscanf(r.FormValue("max_columns"), "%d", &maxCols)
	g.Config.MaxColumns = maxCols

	coverIdx := 0
	fmt.Sscanf(r.FormValue("cover_index"), "%d", &coverIdx)
	g.Config.CoverIndex = coverIdx

	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin/gallery/"+id+"?saved=1", http.StatusSeeOther)
}

func handleAdminGalleryRegenToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	id := r.FormValue("gallery_id")

	mu.Lock()
	g := findGalleryByID(id)
	if g == nil {
		mu.Unlock()
		http.NotFound(w, r)
		return
	}
	g.Config.SecretToken = generateToken()
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin/gallery/"+id+"?token_regenerated=1", http.StatusSeeOther)
}

func handleAdminGallerySync(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	id := r.FormValue("gallery_id")

	mu.Lock()
	g := findGalleryByID(id)
	if g == nil {
		mu.Unlock()
		http.NotFound(w, r)
		return
	}

	if err := syncFromCaptureOne(g); err != nil {
		mu.Unlock()
		http.Error(w, "Sync failed: "+err.Error(), 500)
		return
	}

	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin/gallery/"+id+"?synced=1", http.StatusSeeOther)
}

func handleAdminGalleryDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	id := r.FormValue("gallery_id")

	mu.Lock()
	for i, g := range appData.Galleries {
		if g.ID == id {
			appData.Galleries = append(appData.Galleries[:i], appData.Galleries[i+1:]...)
			break
		}
	}
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func handleAdminSiteSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()

	mu.Lock()
	appData.Site.SiteTitle = r.FormValue("site_title")
	appData.Site.SiteSubtitle = r.FormValue("site_subtitle")
	appData.Site.LogoURL = r.FormValue("logo_url")
	appData.Site.AccentColor = r.FormValue("accent_color")
	appData.Site.BgColor = r.FormValue("bg_color")
	appData.Site.TextColor = r.FormValue("text_color")
	appData.Site.CardColor = r.FormValue("card_color")
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin?saved=1", http.StatusSeeOther)
}

func handleImageProxy(w http.ResponseWriter, r *http.Request) {
	imageURL := r.URL.Query().Get("url")
	if imageURL == "" || !strings.HasPrefix(imageURL, "https://live.captureone.com/") {
		http.Error(w, "invalid url", 400)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		http.Error(w, "fetch failed", 502)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	io.Copy(w, resp.Body)
}

// --- Download & Media ---

func handleDownloadGallery(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	token := r.URL.Query().Get("token")

	mu.RLock()
	var g *Gallery
	if token != "" {
		g = findGalleryByToken(token)
	} else {
		g = findGalleryBySlug(slug)
		if g != nil && g.Config.IsPrivate {
			g = nil
		}
	}
	if g == nil {
		mu.RUnlock()
		http.NotFound(w, r)
		return
	}
	photos := make([]Photo, len(g.Photos))
	copy(photos, g.Photos)
	title := g.Config.GalleryTitle
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, title))

	zw := zip.NewWriter(w)
	defer zw.Close()

	client := &http.Client{Timeout: 60 * time.Second}
	for _, p := range photos {
		resp, err := client.Get(p.MediumURL)
		if err != nil {
			continue
		}
		ext := ".jpg"
		fw, err := zw.Create(p.DisplayName + ext)
		if err != nil {
			resp.Body.Close()
			continue
		}
		io.Copy(fw, resp.Body)
		resp.Body.Close()
	}
}

func handleAdminMediaAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	youtubeURL := r.FormValue("youtube_url")
	if youtubeURL == "" {
		http.Error(w, "YouTube URL required", 400)
		return
	}

	songID := generateID()
	outFile := filepath.Join(mediaDir(), songID+".mp3")

	// Download audio with yt-dlp
	cmd := exec.Command("yt-dlp",
		"-x", "--audio-format", "mp3",
		"--audio-quality", "5",
		"--js-runtimes", "node,deno,bun",
		"-o", outFile,
		"--no-playlist",
		"--quiet",
		youtubeURL,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		http.Error(w, "Download failed: "+string(output), 500)
		return
	}

	// Get metadata with yt-dlp
	metaCmd := exec.Command("yt-dlp", "--print", "%(title)s\n%(uploader)s\n%(duration)s", "--js-runtimes", "node,deno,bun", "--no-playlist", youtubeURL)
	metaOut, _ := metaCmd.Output()
	lines := strings.Split(strings.TrimSpace(string(metaOut)), "\n")
	title := "Unknown"
	artist := ""
	var duration float64
	if len(lines) >= 1 {
		title = lines[0]
	}
	if len(lines) >= 2 {
		artist = lines[1]
	}
	if len(lines) >= 3 {
		duration, _ = strconv.ParseFloat(lines[2], 64)
	}

	// If duration is 0, get it from ffprobe
	if duration == 0 {
		probeCmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", outFile)
		probeOut, _ := probeCmd.Output()
		duration, _ = strconv.ParseFloat(strings.TrimSpace(string(probeOut)), 64)
	}

	song := Song{
		ID:       songID,
		Title:    title,
		Artist:   artist,
		Duration: duration,
		Filename: songID + ".mp3",
		Source:   youtubeURL,
	}

	mu.Lock()
	appData.Songs = append(appData.Songs, song)
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin?media_added=1", http.StatusSeeOther)
}

func handleAdminMediaDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	id := r.FormValue("song_id")

	mu.Lock()
	for i, s := range appData.Songs {
		if s.ID == id {
			os.Remove(filepath.Join(mediaDir(), s.Filename))
			appData.Songs = append(appData.Songs[:i], appData.Songs[i+1:]...)
			// Clear from any galleries using this song
			for j := range appData.Galleries {
				if appData.Galleries[j].SongID == id {
					appData.Galleries[j].SongID = ""
				}
			}
			break
		}
	}
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin?media_deleted=1", http.StatusSeeOther)
}

func handleAdminGallerySetSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	galleryID := r.FormValue("gallery_id")
	songIDs := r.Form["song_ids"]

	// Filter out empty values
	var filtered []string
	for _, id := range songIDs {
		if id != "" {
			filtered = append(filtered, id)
		}
	}

	mu.Lock()
	g := findGalleryByID(galleryID)
	if g != nil {
		g.SongIDs = filtered
		g.SongID = "" // clear legacy field
		saveData()
	}
	mu.Unlock()

	http.Redirect(w, r, "/admin/gallery/"+galleryID+"?song_set=1", http.StatusSeeOther)
}

func handleMediaServe(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	if !strings.HasSuffix(filename, ".mp3") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, filepath.Join(mediaDir(), filename))
}

// --- Templates & Main ---

var (
	overviewTmpl      *template.Template
	galleryTmpl       *template.Template
	adminOverviewTmpl *htmltpl.Template
	adminEditTmpl     *htmltpl.Template
	loginTmpl         *htmltpl.Template
)

func main() {
	dir, _ := os.Getwd()
	dataFile = filepath.Join(dir, "gallery_data.json")
	loadData()

	funcMap := template.FuncMap{
		"stars": func(n int) string {
			s := ""
			for i := 0; i < n; i++ {
				s += "★"
			}
			return s
		},
		"delay": func(i int) string {
			return fmt.Sprintf("%.2f", float64(i)*0.04)
		},
	}
	overviewTmpl = template.Must(template.New("overview").Funcs(funcMap).Parse(overviewHTML))
	galleryTmpl = template.Must(template.New("gallery").Funcs(funcMap).Parse(galleryHTML))
	htmlFuncMap := htmltpl.FuncMap{
		"divf": func(a float64, b int) float64 {
			if b == 0 {
				return 0
			}
			return a / float64(b)
		},
	}
	adminOverviewTmpl = htmltpl.Must(htmltpl.New("adminOverview").Parse(adminOverviewHTML))
	adminEditTmpl = htmltpl.Must(htmltpl.New("adminEdit").Funcs(htmlFuncMap).Parse(adminEditHTML))
	loginTmpl = htmltpl.Must(htmltpl.New("login").Parse(loginHTML))

	// Public routes
	http.HandleFunc("/proxy/image", handleImageProxy)
	http.HandleFunc("/download", handleDownloadGallery)
	http.HandleFunc("/media/", handleMediaServe)
	http.HandleFunc("/", handleOverview)

	// Auth routes
	http.HandleFunc("/admin/login", handleLogin)
	http.HandleFunc("/admin/logout", handleLogout)

	// Protected admin routes
	http.HandleFunc("/admin/gallery/save", requireAuth(handleAdminGallerySave))
	http.HandleFunc("/admin/gallery/sync", requireAuth(handleAdminGallerySync))
	http.HandleFunc("/admin/gallery/delete", requireAuth(handleAdminGalleryDelete))
	http.HandleFunc("/admin/gallery/regen-token", requireAuth(handleAdminGalleryRegenToken))
	http.HandleFunc("/admin/gallery/set-song", requireAuth(handleAdminGallerySetSong))
	http.HandleFunc("/admin/gallery/", requireAuth(handleAdminGallery))
	http.HandleFunc("/admin/new", requireAuth(handleAdminNew))
	http.HandleFunc("/admin/media/add", requireAuth(handleAdminMediaAdd))
	http.HandleFunc("/admin/media/delete", requireAuth(handleAdminMediaDelete))
	http.HandleFunc("/admin/site/save", requireAuth(handleAdminSiteSave))
	http.HandleFunc("/admin/password", requireAuth(handleChangePassword))
	http.HandleFunc("/admin", requireAuth(handleAdmin))

	log.Println("Gallery running on http://localhost:8082")
	log.Println("Admin panel: http://localhost:8082/admin")
	log.Println("Default login: admin / admin")
	log.Fatal(http.ListenAndServe("0.0.0.0:8082", nil))
}

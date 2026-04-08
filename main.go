package main

import (
	"archive/zip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	htmltpl "html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

// --- Data Model ---

type GalleryConfig struct {
	SourceType      string `json:"source_type"`
	CaptureOneURL   string `json:"capture_one_url"`
	NextcloudURL    string `json:"nextcloud_url"`
	NextcloudUser   string `json:"nextcloud_user"`
	NextcloudToken  string `json:"nextcloud_token"`
	NextcloudFolder string `json:"nextcloud_folder"`
	GalleryTitle    string `json:"gallery_title"`
	Subtitle        string `json:"subtitle"`
	Slug            string `json:"slug"`
	CoverIndex      int    `json:"cover_index"`
	IsPrivate       bool   `json:"is_private"`
	SecretToken     string `json:"secret_token"`
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
	// Nextcloud credentials (global)
	NextcloudURL   string `json:"nextcloud_url,omitempty"`
	NextcloudUser  string `json:"nextcloud_user,omitempty"`
	NextcloudToken string `json:"nextcloud_token,omitempty"`
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
	appData         AppData
	dataFile        string
	mu              sync.RWMutex
	passwordPepper  string
	loginCSRFSecret string
	sessions        = make(map[string]sessionState)
	sessMu          sync.Mutex
	loginStateMu    sync.Mutex
	loginAttempts   = make(map[string]loginAttempt)
)

type sessionState struct {
	Expiry    time.Time
	CSRFToken string
}

type loginAttempt struct {
	Failures    int
	LockedUntil time.Time
	LastAttempt time.Time
}

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
		SourceType:          "captureone",
		GalleryTitle:        "New Gallery",
		BackgroundColor:     "#0a0a0a",
		CardColor:           "#1a1a1a",
		TextColor:           "#f0f0f0",
		AccentColor:         "#c8a97e",
		FrameStyle:          "none",
		BorderRadius:        "4px",
		BorderWidth:         "0px",
		BorderColor:         "#333333",
		Shadow:              "0 8px 32px rgba(0,0,0,0.4)",
		HoverEffect:         "lift",
		Layout:              "masonry",
		ColumnGap:           "16px",
		MaxColumns:          4,
		ShowFilenames:       false,
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

func hashPasswordLegacy(pw string) string {
	h := sha256.Sum256([]byte(pw))
	return hex.EncodeToString(h[:])
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hashLen := sha256.Size
	blocks := (keyLen + hashLen - 1) / hashLen
	out := make([]byte, 0, blocks*hashLen)

	for block := 1; block <= blocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := slices.Clone(u)

		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}

	return out[:keyLen]
}

func passwordKeyMaterial(password string) []byte {
	return []byte(password + "\x00" + passwordPepper)
}

func hashPassword(pw string) string {
	const iterations = 120000
	salt := make([]byte, 16)
	rand.Read(salt)
	key := pbkdf2SHA256(passwordKeyMaterial(pw), salt, iterations, 32)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s",
		iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
}

func verifyPassword(password, stored string) (valid bool, needsUpgrade bool) {
	if strings.HasPrefix(stored, "pbkdf2_sha256$") {
		parts := strings.Split(stored, "$")
		if len(parts) != 4 {
			return false, false
		}
		iterations, err := strconv.Atoi(parts[1])
		if err != nil || iterations < 10000 {
			return false, false
		}
		salt, err := base64.RawStdEncoding.DecodeString(parts[2])
		if err != nil {
			return false, false
		}
		expected, err := base64.RawStdEncoding.DecodeString(parts[3])
		if err != nil {
			return false, false
		}
		actual := pbkdf2SHA256(passwordKeyMaterial(password), salt, iterations, len(expected))
		return subtle.ConstantTimeCompare(actual, expected) == 1, false
	}

	legacy := hashPasswordLegacy(password)
	return subtle.ConstantTimeCompare([]byte(legacy), []byte(stored)) == 1, true
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
		bootstrapAdminAuth()
		return
	}
	if err := json.Unmarshal(raw, &appData); err != nil {
		appData = AppData{Site: defaultSiteConfig()}
		bootstrapAdminAuth()
		return
	}
	applySiteDefaults(&appData.Site)
	// Set default auth if not configured
	if appData.Auth.Username == "" {
		bootstrapAdminAuth()
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

func bootstrapAdminAuth() {
	password := randomPassword()
	appData.Auth.Username = "admin"
	appData.Auth.PasswordHash = hashPassword(password)
	log.Printf("Initial admin credentials generated. Username: %s Password: %s", appData.Auth.Username, password)
}

func randomPassword() string {
	b := make([]byte, 18)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func applySiteDefaults(site *SiteConfig) {
	defaults := defaultSiteConfig()
	if site.SiteTitle == "" {
		site.SiteTitle = defaults.SiteTitle
	}
	if site.SiteSubtitle == "" {
		site.SiteSubtitle = defaults.SiteSubtitle
	}
	if site.AccentColor == "" {
		site.AccentColor = defaults.AccentColor
	}
	if site.BgColor == "" {
		site.BgColor = defaults.BgColor
	}
	if site.TextColor == "" {
		site.TextColor = defaults.TextColor
	}
	if site.CardColor == "" {
		site.CardColor = defaults.CardColor
	}
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

func stablePhotoID(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func gallerySlugFromForm(r *http.Request) string {
	slug := slugify(r.FormValue("slug"))
	if slug != "" {
		return slug
	}
	return slugify(r.FormValue("gallery_title"))
}

func downloadFilename(p Photo, index int) string {
	name := filepath.Base(strings.TrimSpace(p.DisplayName))
	if name == "" || name == "." || name == string(filepath.Separator) {
		name = filepath.Base(strings.TrimSpace(p.MediumURL))
	}
	if name == "" || name == "." || name == string(filepath.Separator) {
		name = fmt.Sprintf("photo-%d.jpg", index+1)
	}
	if filepath.Ext(name) == "" {
		name += filepath.Ext(strings.TrimSpace(p.MediumURL))
	}
	if filepath.Ext(name) == "" {
		name += ".jpg"
	}
	return name
}

func authorizedRemoteRequest(method, rawURL string) (*http.Request, error) {
	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("invalid url")
	}
	if strings.EqualFold(u.Host, "live.captureone.com") {
		return req, nil
	}

	mu.RLock()
	defer mu.RUnlock()
	for _, g := range appData.Galleries {
		if g.Config.SourceType != "nextcloud" || g.Config.NextcloudURL == "" || g.Config.NextcloudUser == "" || g.Config.NextcloudToken == "" {
			continue
		}

		base, err := url.Parse(g.Config.NextcloudURL)
		if err != nil {
			continue
		}
		if !strings.EqualFold(u.Scheme, base.Scheme) || !strings.EqualFold(u.Host, base.Host) {
			continue
		}

		basePath := strings.TrimSuffix(base.EscapedPath(), "/")
		expectedPrefix := path.Clean(basePath + "/remote.php/dav/files/" + g.Config.NextcloudUser + "/")
		requestPath := path.Clean(u.EscapedPath())
		if !strings.HasPrefix(requestPath+"/", expectedPrefix+"/") {
			continue
		}

		req.Header.Set("Authorization", "Basic "+base64Encode(g.Config.NextcloudUser+":"+g.Config.NextcloudToken))
		return req, nil
	}

	return nil, fmt.Errorf("invalid url")
}

func shouldUseSecureCookies(r *http.Request) bool {
	if os.Getenv("GALLERY_DEV_MODE") == "1" {
		return false
	}
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func getSession(r *http.Request) (string, sessionState, bool) {
	cookie, err := r.Cookie("gallery_session")
	if err != nil {
		return "", sessionState{}, false
	}
	sessMu.Lock()
	defer sessMu.Unlock()
	session, ok := sessions[cookie.Value]
	if !ok || time.Now().After(session.Expiry) {
		delete(sessions, cookie.Value)
		return "", sessionState{}, false
	}
	return cookie.Value, session, true
}

func issueCSRFCookie(w http.ResponseWriter, r *http.Request) string {
	token := generateToken()
	setCSRFCookie(w, r, token)
	return token
}

func setCSRFCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gallery_login_csrf",
		Value:    token,
		Path:     "/admin/login",
		HttpOnly: true,
		Secure:   shouldUseSecureCookies(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600,
	})
}

func loginCSRFToken(r *http.Request) string {
	cookie, err := r.Cookie("gallery_login_csrf")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func loginCSRFValue(token string) string {
	mac := hmac.New(sha256.New, []byte(loginCSRFSecret))
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func validLoginCSRF(r *http.Request) bool {
	cookieToken := loginCSRFToken(r)
	formToken := r.FormValue("csrf_token")
	if cookieToken == "" || formToken == "" {
		return false
	}
	expected := loginCSRFValue(cookieToken)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(formToken)) == 1
}

func loginThrottleKey(r *http.Request, username string) string {
	ip := r.RemoteAddr
	if host, _, ok := strings.Cut(r.RemoteAddr, ":"); ok {
		ip = host
	}
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		ip = forwarded
	}
	return strings.ToLower(strings.TrimSpace(username)) + "|" + ip
}

func canAttemptLogin(r *http.Request, username string) (bool, time.Duration) {
	loginStateMu.Lock()
	defer loginStateMu.Unlock()
	key := loginThrottleKey(r, username)
	attempt := loginAttempts[key]
	now := time.Now()
	if !attempt.LockedUntil.IsZero() && now.Before(attempt.LockedUntil) {
		return false, time.Until(attempt.LockedUntil)
	}
	if !attempt.LastAttempt.IsZero() && now.Sub(attempt.LastAttempt) > 30*time.Minute {
		delete(loginAttempts, key)
	}
	return true, 0
}

func recordLoginFailure(r *http.Request, username string) {
	loginStateMu.Lock()
	defer loginStateMu.Unlock()
	key := loginThrottleKey(r, username)
	attempt := loginAttempts[key]
	attempt.Failures++
	attempt.LastAttempt = time.Now()
	if attempt.Failures >= 5 {
		attempt.LockedUntil = attempt.LastAttempt.Add(15 * time.Minute)
		attempt.Failures = 0
	}
	loginAttempts[key] = attempt
}

func clearLoginFailures(r *http.Request, username string) {
	loginStateMu.Lock()
	delete(loginAttempts, loginThrottleKey(r, username))
	loginStateMu.Unlock()
}

func requestScheme(r *http.Request) string {
	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") || r.TLS != nil {
		return "https"
	}
	return "http"
}

// --- Auth ---

func createSession() string {
	token := generateToken()
	sessMu.Lock()
	sessions[token] = sessionState{
		Expiry:    time.Now().Add(24 * time.Hour),
		CSRFToken: generateToken(),
	}
	sessMu.Unlock()
	return token
}

func isAuthenticated(r *http.Request) bool {
	_, _, ok := getSession(r)
	return ok
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, session, ok := getSession(r)
		if !ok {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		if r.Method == http.MethodPost {
			if subtle.ConstantTimeCompare([]byte(r.FormValue("csrf_token")), []byte(session.CSRFToken)) != 1 {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}
		}
		next(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		errMsg := ""
		if code := r.URL.Query().Get("error"); code != "" {
			switch code {
			case "1":
				errMsg = "Invalid username or password"
			case "throttled":
				errMsg = "Too many failed attempts. Try again later."
			default:
				errMsg = "Unable to sign in"
			}
		}
		token := loginCSRFToken(r)
		if token == "" {
			token = issueCSRFCookie(w, r)
		} else {
			setCSRFCookie(w, r, token)
		}
		loginTmpl.Execute(w, struct {
			Error     string
			CSRFToken string
		}{
			Error:     errMsg,
			CSRFToken: loginCSRFValue(token),
		})
		return
	}

	r.ParseForm()
	if !validLoginCSRF(r) {
		http.Error(w, "invalid csrf token", http.StatusForbidden)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if ok, _ := canAttemptLogin(r, username); !ok {
		http.Redirect(w, r, "/admin/login?error=throttled", http.StatusSeeOther)
		return
	}

	mu.Lock()
	validUser := appData.Auth.Username
	validHash := appData.Auth.PasswordHash
	validPassword, needsUpgrade := verifyPassword(password, validHash)
	if username == validUser && validPassword && needsUpgrade {
		appData.Auth.PasswordHash = hashPassword(password)
		saveData()
		validHash = appData.Auth.PasswordHash
	}
	mu.Unlock()

	if username != validUser || !validPassword {
		recordLoginFailure(r, username)
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}
	clearLoginFailures(r, username)

	token := createSession()
	http.SetCookie(w, &http.Cookie{
		Name:     "gallery_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookies(r),
		MaxAge:   86400,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "gallery_login_csrf",
		Value:    "",
		Path:     "/admin/login",
		HttpOnly: true,
		Secure:   shouldUseSecureCookies(r),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("gallery_session")
	if err == nil {
		sessMu.Lock()
		delete(sessions, cookie.Value)
		sessMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "gallery_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   shouldUseSecureCookies(r),
		SameSite: http.SameSiteStrictMode,
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
	currentPass := r.FormValue("current_password")
	newUser := r.FormValue("username")
	newPass := r.FormValue("new_password")
	if ok, _ := verifyPassword(currentPass, appData.Auth.PasswordHash); !ok {
		mu.Unlock()
		http.Error(w, "current password is incorrect", http.StatusForbidden)
		return
	}
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

// --- Nextcloud API ---

// WebDAV Multistatus XML structures
type davMultistatus struct {
	XMLName   xml.Name      `xml:"multistatus"`
	Responses []davResponse `xml:"response"`
}

type davResponse struct {
	Href     string        `xml:"href"`
	Propstat []davPropstat `xml:"propstat"`
}

type davPropstat struct {
	Prop   davProp `xml:"prop"`
	Status string  `xml:"status"`
}

type davProp struct {
	ContentType   string `xml:"getcontenttype"`
	ContentLength int64  `xml:"getcontentlength"`
	ResourceType  struct {
		Collection *struct{} `xml:"collection"`
	} `xml:"resourcetype"`
	LastModified string `xml:"getlastmodified"`
}

func nextcloudDAVURL(baseURL, user, folder string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	davURL := fmt.Sprintf("%s/remote.php/dav/files/%s/", baseURL, user)
	if folder != "" {
		folder = strings.TrimPrefix(folder, "/")
		davURL += folder
		if !strings.HasSuffix(davURL, "/") {
			davURL += "/"
		}
	}
	return davURL
}

func nextcloudPROPFIND(ncURL, user, token, folder string) (*davMultistatus, error) {
	davURL := nextcloudDAVURL(ncURL, user, folder)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("PROPFIND", davURL, strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?>
<d:propfind xmlns:d="DAV:">
  <d:prop>
    <d:getcontenttype/>
    <d:getcontentlength/>
    <d:resourcetype/>
    <d:getlastmodified/>
  </d:prop>
</d:propfind>`))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Depth", "1")
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(user, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("webdav request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 207 {
		return nil, fmt.Errorf("webdav error: status %d", resp.StatusCode)
	}

	var ms davMultistatus
	if err := xml.NewDecoder(resp.Body).Decode(&ms); err != nil {
		return nil, fmt.Errorf("parse webdav response: %w", err)
	}
	return &ms, nil
}

func isImageContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.HasPrefix(ct, "image/jpeg") || strings.HasPrefix(ct, "image/png") ||
		strings.HasPrefix(ct, "image/webp") || strings.HasPrefix(ct, "image/tiff") ||
		strings.HasPrefix(ct, "image/heic")
}

func syncFromNextcloud(gallery *Gallery) error {
	ncURL := gallery.Config.NextcloudURL
	ncUser := gallery.Config.NextcloudUser
	ncToken := gallery.Config.NextcloudToken
	folder := gallery.Config.NextcloudFolder

	// Fall back to global credentials if per-gallery are empty
	if ncURL == "" {
		ncURL = appData.Site.NextcloudURL
	}
	if ncUser == "" {
		ncUser = appData.Site.NextcloudUser
	}
	if ncToken == "" {
		ncToken = appData.Site.NextcloudToken
	}

	if ncURL == "" || ncToken == "" || ncUser == "" {
		return fmt.Errorf("Nextcloud credentials not configured")
	}

	ms, err := nextcloudPROPFIND(ncURL, ncUser, ncToken, folder)
	if err != nil {
		return err
	}

	baseURL := strings.TrimSuffix(ncURL, "/")
	var photos []Photo
	for _, r := range ms.Responses {
		// Skip the folder itself (first response is always the folder)
		if strings.HasSuffix(r.Href, "/") {
			continue
		}

		// Check content type from propstat
		var ct string
		for _, ps := range r.Propstat {
			if ps.Prop.ContentType != "" {
				ct = ps.Prop.ContentType
			}
		}

		// Also check by file extension as fallback
		lowerHref := strings.ToLower(r.Href)
		isImage := isImageContentType(ct)
		if !isImage {
			for _, ext := range []string{".jpg", ".jpeg", ".png", ".webp", ".tiff", ".heic"} {
				if strings.HasSuffix(lowerHref, ext) {
					isImage = true
					break
				}
			}
		}
		if !isImage {
			continue
		}

		// Build WebDAV download URL for the file
		filePath := r.Href
		prefix := fmt.Sprintf("/remote.php/dav/files/%s/", ncUser)
		if idx := strings.Index(filePath, prefix); idx >= 0 {
			filePath = filePath[idx+len(prefix):]
		}

		fileURL := fmt.Sprintf("%s/remote.php/dav/files/%s/%s", baseURL, ncUser, filePath)

		displayName := filepath.Base(r.Href)

		photos = append(photos, Photo{
			UUID:        stablePhotoID(r.Href),
			DisplayName: displayName,
			SmallURL:    fileURL,
			MediumURL:   fileURL,
			Width:       0,
			Height:      0,
		})
	}

	gallery.Photos = photos
	// Store the credentials used for this gallery so the proxy can authenticate
	gallery.Config.NextcloudURL = ncURL
	gallery.Config.NextcloudUser = ncUser
	gallery.Config.NextcloudToken = ncToken

	return nil
}

// listNextcloudFolders lists subfolders at the given path
func listNextcloudFolders(ncURL, user, token, folder string) ([]string, error) {
	ms, err := nextcloudPROPFIND(ncURL, user, token, folder)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("/remote.php/dav/files/%s/", user)
	var folders []string
	for i, r := range ms.Responses {
		if i == 0 {
			continue // skip the folder itself
		}
		if !strings.HasSuffix(r.Href, "/") {
			continue
		}
		// Check if it's a collection
		isCollection := false
		for _, ps := range r.Propstat {
			if ps.Prop.ResourceType.Collection != nil {
				isCollection = true
			}
		}
		if !isCollection && !strings.HasSuffix(r.Href, "/") {
			continue
		}

		path := r.Href
		if idx := strings.Index(path, prefix); idx >= 0 {
			path = path[idx+len(prefix):]
		}
		path = strings.TrimSuffix(path, "/")
		if path != "" {
			folders = append(folders, path)
		}
	}
	return folders, nil
}

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
	_, session, _ := getSession(r)
	mu.RLock()
	d := struct {
		AppData
		Host      string
		Scheme    string
		CSRFToken string
	}{appData, r.Host, requestScheme(r), session.CSRFToken}
	mu.RUnlock()
	adminOverviewTmpl.Execute(w, d)
}

func handleAdminNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	r.ParseForm()
	sourceType := r.FormValue("source_type")
	if sourceType == "" {
		sourceType = "captureone"
	}

	g := Gallery{
		ID:     generateID(),
		Config: defaultGalleryConfig(),
	}
	g.Config.SourceType = sourceType
	g.Config.SecretToken = generateToken()
	g.Config.IsPrivate = r.FormValue("is_private") == "on"

	var syncErr error
	if sourceType == "nextcloud" {
		ncFolder := r.FormValue("nextcloud_folder")
		if ncFolder == "" {
			http.Error(w, "Nextcloud folder is required", 400)
			return
		}
		g.Config.NextcloudFolder = ncFolder
		// Use the folder name as gallery title
		g.Config.GalleryTitle = filepath.Base(ncFolder)
		g.Config.Slug = slugify(g.Config.GalleryTitle)

		mu.Lock()
		appData.Galleries = append(appData.Galleries, g)
		gPtr := &appData.Galleries[len(appData.Galleries)-1]
		mu.Unlock()

		syncErr = syncFromNextcloud(gPtr)
	} else {
		c1URL := r.FormValue("capture_one_url")
		if c1URL == "" {
			http.Error(w, "Capture One URL required", 400)
			return
		}
		g.Config.CaptureOneURL = c1URL

		mu.Lock()
		appData.Galleries = append(appData.Galleries, g)
		gPtr := &appData.Galleries[len(appData.Galleries)-1]
		mu.Unlock()

		syncErr = syncFromCaptureOne(gPtr)
	}

	if syncErr != nil {
		mu.Lock()
		// Remove the gallery we just added
		for i, gg := range appData.Galleries {
			if gg.ID == g.ID {
				appData.Galleries = append(appData.Galleries[:i], appData.Galleries[i+1:]...)
				break
			}
		}
		mu.Unlock()
		http.Error(w, "Sync failed: "+syncErr.Error(), 500)
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
	Scheme        string
	Songs         []Song
	SelectedSongs []Song
	TotalDuration float64
	CSRFToken     string
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
	_, session, _ := getSession(r)
	d := adminEditData{Gallery: *g, Host: r.Host, Scheme: requestScheme(r), Songs: appData.Songs, SelectedSongs: selected, TotalDuration: totalDur, CSRFToken: session.CSRFToken}
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

	g.Config.SourceType = r.FormValue("source_type")
	g.Config.CaptureOneURL = r.FormValue("capture_one_url")
	g.Config.NextcloudFolder = r.FormValue("nextcloud_folder")
	g.Config.GalleryTitle = r.FormValue("gallery_title")
	g.Config.Subtitle = r.FormValue("subtitle")
	g.Config.Slug = gallerySlugFromForm(r)
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

	var err error
	if g.Config.SourceType == "nextcloud" {
		err = syncFromNextcloud(g)
	} else {
		err = syncFromCaptureOne(g)
	}

	if err != nil {
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
	appData.Site.NextcloudURL = r.FormValue("nextcloud_url")
	appData.Site.NextcloudUser = r.FormValue("nextcloud_user")
	// Only update token if a new one was provided (don't overwrite with empty)
	if t := r.FormValue("nextcloud_token"); t != "" {
		appData.Site.NextcloudToken = t
	}
	saveData()
	mu.Unlock()

	http.Redirect(w, r, "/admin?saved=1", http.StatusSeeOther)
}

func handleImageProxy(w http.ResponseWriter, r *http.Request) {
	imageURL := r.URL.Query().Get("url")
	if imageURL == "" {
		http.Error(w, "invalid url", 400)
		return
	}

	req, err := authorizedRemoteRequest(http.MethodGet, imageURL)
	if err != nil {
		http.Error(w, "invalid url", 400)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "fetch failed", 502)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "fetch failed", 502)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	io.Copy(w, resp.Body)
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
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
	for i, p := range photos {
		req, err := authorizedRemoteRequest(http.MethodGet, p.MediumURL)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}
		fw, err := zw.Create(downloadFilename(p, i))
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
				var filtered []string
				for _, songID := range appData.Galleries[j].SongIDs {
					if songID != id {
						filtered = append(filtered, songID)
					}
				}
				appData.Galleries[j].SongIDs = filtered
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

func handleAdminNextcloudFolders(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("path")

	mu.RLock()
	ncURL := appData.Site.NextcloudURL
	ncUser := appData.Site.NextcloudUser
	ncToken := appData.Site.NextcloudToken
	mu.RUnlock()

	if ncURL == "" || ncUser == "" || ncToken == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "Nextcloud credentials not configured"})
		return
	}

	folders, err := listNextcloudFolders(ncURL, ncUser, ncToken, folder)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"folders": folders,
		"current": folder,
	})
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
	passwordPepper = envOrRandom("GALLERY_PASSWORD_PEPPER")
	loginCSRFSecret = envOrRandom("GALLERY_LOGIN_CSRF_SECRET")
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
		"urlencode": func(s string) string {
			return url.QueryEscape(s)
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
		"urlencode": func(s string) string {
			return url.QueryEscape(s)
		},
		"proxyURL": func(rawURL string) htmltpl.URL {
			return htmltpl.URL("/proxy/image?url=" + url.QueryEscape(rawURL))
		},
	}
	adminOverviewTmpl = htmltpl.Must(htmltpl.New("adminOverview").Funcs(htmlFuncMap).Parse(adminOverviewHTML))
	adminEditTmpl = htmltpl.Must(htmltpl.New("adminEdit").Funcs(htmlFuncMap).Parse(adminEditHTML))
	loginTmpl = htmltpl.Must(htmltpl.New("login").Parse(loginHTML))

	// Public routes
	http.HandleFunc("/proxy/image", handleImageProxy)
	http.HandleFunc("/download", handleDownloadGallery)
	http.HandleFunc("/media/", handleMediaServe)
	http.HandleFunc("/", handleOverview)

	// Auth routes
	http.HandleFunc("/admin/login", handleLogin)
	http.HandleFunc("/admin/logout", requireAuth(handleLogout))

	// Protected admin routes
	http.HandleFunc("/admin/gallery/save", requireAuth(handleAdminGallerySave))
	http.HandleFunc("/admin/gallery/sync", requireAuth(handleAdminGallerySync))
	http.HandleFunc("/admin/gallery/delete", requireAuth(handleAdminGalleryDelete))
	http.HandleFunc("/admin/gallery/regen-token", requireAuth(handleAdminGalleryRegenToken))
	http.HandleFunc("/admin/gallery/set-song", requireAuth(handleAdminGallerySetSong))
	http.HandleFunc("/admin/gallery/", requireAuth(handleAdminGallery))
	http.HandleFunc("/admin/new", requireAuth(handleAdminNew))
	http.HandleFunc("/admin/nextcloud/folders", requireAuth(handleAdminNextcloudFolders))
	http.HandleFunc("/admin/media/add", requireAuth(handleAdminMediaAdd))
	http.HandleFunc("/admin/media/delete", requireAuth(handleAdminMediaDelete))
	http.HandleFunc("/admin/site/save", requireAuth(handleAdminSiteSave))
	http.HandleFunc("/admin/password", requireAuth(handleChangePassword))
	http.HandleFunc("/admin", requireAuth(handleAdmin))

	log.Println("Gallery running on http://localhost:8082")
	log.Println("Admin panel: http://localhost:8082/admin")
	if os.Getenv("GALLERY_DEV_MODE") == "1" {
		log.Println("Dev mode enabled: secure cookies are relaxed for local HTTP")
	}
	log.Fatal(http.ListenAndServe("0.0.0.0:8082", nil))
}

func envOrRandom(name string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return generateToken()
}

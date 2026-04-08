package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplySiteDefaultsPreservesConfiguredFields(t *testing.T) {
	site := SiteConfig{
		SiteTitle:   "Custom Title",
		AccentColor: "",
		TextColor:   "#222222",
	}

	applySiteDefaults(&site)

	if site.SiteTitle != "Custom Title" {
		t.Fatalf("site title was overwritten: %q", site.SiteTitle)
	}
	if site.AccentColor == "" || site.BgColor == "" || site.CardColor == "" || site.SiteSubtitle == "" {
		t.Fatalf("expected missing defaults to be filled, got %+v", site)
	}
	if site.TextColor != "#222222" {
		t.Fatalf("expected existing text color to be preserved, got %q", site.TextColor)
	}
}

func TestGallerySlugFromFormFallsBackToTitle(t *testing.T) {
	form := url.Values{
		"slug":          {""},
		"gallery_title": {"Summer Wedding"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/gallery/save", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatalf("parse form: %v", err)
	}

	if got := gallerySlugFromForm(req); got != "summer-wedding" {
		t.Fatalf("unexpected slug: %q", got)
	}
}

func TestAuthorizedRemoteRequestValidatesExactHosts(t *testing.T) {
	appData = AppData{
		Galleries: []Gallery{
			{
				Config: GalleryConfig{
					SourceType:     "nextcloud",
					NextcloudURL:   "https://cloud.example.com",
					NextcloudUser:  "alice",
					NextcloudToken: "secret",
				},
			},
		},
	}

	req, err := authorizedRemoteRequest(http.MethodGet, "https://cloud.example.com/remote.php/dav/files/alice/folder/photo.jpg")
	if err != nil {
		t.Fatalf("expected nextcloud url to be allowed: %v", err)
	}
	if req.Header.Get("Authorization") == "" {
		t.Fatal("expected nextcloud request to include auth header")
	}

	if _, err := authorizedRemoteRequest(http.MethodGet, "https://cloud.example.com.evil/remote.php/dav/files/alice/folder/photo.jpg"); err == nil {
		t.Fatal("expected host-prefix attack URL to be rejected")
	}

	req, err = authorizedRemoteRequest(http.MethodGet, "https://live.captureone.com/some/path")
	if err != nil {
		t.Fatalf("expected capture one url to be allowed: %v", err)
	}
	if req.Header.Get("Authorization") != "" {
		t.Fatal("did not expect capture one request to include auth header")
	}
}

func TestDownloadFilenameKeepsExistingExtension(t *testing.T) {
	name := downloadFilename(Photo{
		DisplayName: "photo.png",
		MediumURL:   "https://example.com/photo.png",
	}, 0)

	if name != "photo.png" {
		t.Fatalf("unexpected filename: %q", name)
	}
}

func TestHandleAdminMediaDeleteClearsSongReferences(t *testing.T) {
	tmp := t.TempDir()
	dataFile = filepath.Join(tmp, "gallery_data.json")
	appData = AppData{
		Songs: []Song{
			{ID: "song-1", Filename: "song-1.mp3"},
		},
		Galleries: []Gallery{
			{ID: "g1", SongID: "song-1", SongIDs: []string{"song-1", "song-2"}},
			{ID: "g2", SongIDs: []string{"song-2", "song-1"}},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/media/delete", strings.NewReader("song_id=song-1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handleAdminMediaDelete(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if len(appData.Songs) != 0 {
		t.Fatalf("expected song to be deleted, got %d songs", len(appData.Songs))
	}
	if appData.Galleries[0].SongID != "" {
		t.Fatalf("expected legacy song id to be cleared, got %q", appData.Galleries[0].SongID)
	}
	if got := strings.Join(appData.Galleries[0].SongIDs, ","); got != "song-2" {
		t.Fatalf("unexpected gallery 1 songs: %q", got)
	}
	if got := strings.Join(appData.Galleries[1].SongIDs, ","); got != "song-2" {
		t.Fatalf("unexpected gallery 2 songs: %q", got)
	}
}

func TestVerifyPasswordSupportsLegacyAndModernHashes(t *testing.T) {
	passwordPepper = "pepper"

	modern := hashPassword("secret-pass")
	if ok, upgrade := verifyPassword("secret-pass", modern); !ok || upgrade {
		t.Fatalf("expected modern hash to verify without upgrade, got ok=%v upgrade=%v", ok, upgrade)
	}

	legacy := hashPasswordLegacy("legacy-pass")
	if ok, upgrade := verifyPassword("legacy-pass", legacy); !ok || !upgrade {
		t.Fatalf("expected legacy hash to verify with upgrade, got ok=%v upgrade=%v", ok, upgrade)
	}
}

func TestRequireAuthRejectsMissingCSRFFromPost(t *testing.T) {
	sessions = map[string]sessionState{
		"session-token": {
			Expiry:    time.Now().Add(time.Hour),
			CSRFToken: "csrf-token",
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/site/save", nil)
	req.AddCookie(&http.Cookie{Name: "gallery_session", Value: "session-token"})
	rr := httptest.NewRecorder()

	requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected csrf rejection, got %d", rr.Code)
	}
}

func TestShouldUseSecureCookiesDisabledInDevMode(t *testing.T) {
	t.Setenv("GALLERY_DEV_MODE", "1")
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8082/admin/login", nil)

	if shouldUseSecureCookies(req) {
		t.Fatal("expected secure cookies to be disabled in dev mode")
	}
}

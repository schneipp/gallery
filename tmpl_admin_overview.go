package main

const adminOverviewHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Gallery Admin</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    background: #111;
    color: #e0e0e0;
    min-height: 100vh;
  }
  .topbar {
    background: #1a1a1a;
    border-bottom: 1px solid #2a2a2a;
    padding: 16px 32px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .topbar h1 { font-size: 18px; font-weight: 600; color: #c8a97e; letter-spacing: 1px; }
  .topbar-right { display: flex; gap: 16px; align-items: center; }
  .topbar a { color: #888; text-decoration: none; font-size: 14px; transition: color 0.2s; }
  .topbar a:hover { color: #c8a97e; }
  .container { max-width: 1100px; margin: 0 auto; padding: 32px 24px; }
  .section {
    background: #1a1a1a;
    border: 1px solid #2a2a2a;
    border-radius: 12px;
    padding: 28px;
    margin-bottom: 24px;
  }
  .section-title {
    font-size: 15px; font-weight: 600; color: #c8a97e;
    text-transform: uppercase; letter-spacing: 1.5px;
    margin-bottom: 20px; padding-bottom: 12px;
    border-bottom: 1px solid #2a2a2a;
  }
  .field { margin-bottom: 18px; }
  .field label { display: block; font-size: 13px; color: #999; margin-bottom: 6px; font-weight: 500; }
  .field input[type="text"], .field select {
    width: 100%;
    padding: 10px 14px;
    background: #111;
    border: 1px solid #333;
    border-radius: 6px;
    color: #e0e0e0;
    font-size: 14px;
    font-family: inherit;
    transition: border-color 0.2s;
  }
  .field input:focus, .field select:focus { outline: none; border-color: #c8a97e; }
  .field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  .color-field { display: flex; align-items: center; gap: 10px; }
  .color-field input[type="color"] {
    width: 36px; height: 36px; border: 1px solid #333;
    border-radius: 6px; background: #111; cursor: pointer; padding: 2px;
  }
  .color-field input[type="text"] { flex: 1; }
  .btn {
    display: inline-flex; align-items: center; gap: 8px;
    padding: 10px 24px; border: none; border-radius: 8px;
    font-size: 14px; font-weight: 600; cursor: pointer;
    transition: all 0.2s; font-family: inherit; text-decoration: none;
  }
  .btn-primary { background: #c8a97e; color: #111; }
  .btn-primary:hover { background: #d4b88f; }
  .btn-sm { padding: 6px 14px; font-size: 13px; }

  /* Gallery Cards */
  .add-form {
    display: flex; gap: 12px; align-items: flex-end;
  }
  .add-form .field { flex: 1; margin-bottom: 0; }
  .gallery-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 20px;
    margin-top: 20px;
  }
  .gallery-card {
    background: #222;
    border: 1px solid #2a2a2a;
    border-radius: 10px;
    overflow: hidden;
    transition: border-color 0.2s, transform 0.2s;
    text-decoration: none;
    color: inherit;
    display: block;
  }
  .gallery-card:hover { border-color: #c8a97e; transform: translateY(-2px); }
  .gallery-card-cover {
    height: 200px;
    background-size: cover;
    background-position: center;
    background-color: #1a1a1a;
    position: relative;
  }
  .gallery-card-cover .photo-count {
    position: absolute;
    bottom: 10px;
    right: 10px;
    background: rgba(0,0,0,0.7);
    color: #ccc;
    font-size: 12px;
    padding: 4px 10px;
    border-radius: 12px;
    backdrop-filter: blur(4px);
  }
  .gallery-card-body { padding: 16px; }
  .gallery-card-body h3 {
    font-size: 16px; font-weight: 600; margin-bottom: 4px; color: #f0f0f0;
  }
  .gallery-card-body .slug {
    font-size: 12px; color: #666; font-family: monospace;
  }
  .empty-state {
    text-align: center;
    padding: 60px 20px;
    color: #666;
  }
  .empty-state p { font-size: 15px; margin-bottom: 8px; }
  .empty-state .hint { font-size: 13px; color: #555; }
  .badge {
    display: inline-block;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.5px;
    padding: 2px 8px;
    border-radius: 4px;
    text-transform: uppercase;
    margin-left: 6px;
    vertical-align: middle;
  }
  .badge-private { background: rgba(229,115,115,0.2); color: #e57373; }
  .badge-public { background: rgba(129,199,132,0.2); color: #81c784; }
  .checkbox-field {
    display: flex; align-items: center; gap: 10px; margin-top: 12px;
  }
  .checkbox-field input[type="checkbox"] { width: 18px; height: 18px; accent-color: #c8a97e; }
  .checkbox-field label { font-size: 14px; color: #ccc; }
</style>
</head>
<body>
<div class="topbar">
  <h1>Gallery Admin</h1>
  <div class="topbar-right">
    <a href="/" target="_blank">View Site &rarr;</a>
    <a href="/admin/logout">Logout</a>
  </div>
</div>
<div class="container">

  <!-- Site Settings -->
  <div class="section">
    <div class="section-title">Site Settings</div>
    <form method="POST" action="/admin/site/save">
      <div class="field-row">
        <div class="field">
          <label>Site Title</label>
          <input type="text" name="site_title" value="{{.Site.SiteTitle}}">
        </div>
        <div class="field">
          <label>Subtitle</label>
          <input type="text" name="site_subtitle" value="{{.Site.SiteSubtitle}}">
        </div>
      </div>
      <div class="field-row">
        <div class="field">
          <label>Logo URL</label>
          <input type="text" name="logo_url" value="{{.Site.LogoURL}}" placeholder="https://...">
        </div>
        <div class="field">
          <label>Accent Color</label>
          <div class="color-field">
            <input type="color" value="{{.Site.AccentColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="accent_color" value="{{.Site.AccentColor}}">
          </div>
        </div>
      </div>
      <div class="field-row">
        <div class="field">
          <label>Background</label>
          <div class="color-field">
            <input type="color" value="{{.Site.BgColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="bg_color" value="{{.Site.BgColor}}">
          </div>
        </div>
        <div class="field">
          <label>Text Color</label>
          <div class="color-field">
            <input type="color" value="{{.Site.TextColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="text_color" value="{{.Site.TextColor}}">
          </div>
        </div>
      </div>
      <div class="field" style="max-width:50%">
        <label>Card Color</label>
        <div class="color-field">
          <input type="color" value="{{.Site.CardColor}}" onchange="this.nextElementSibling.value=this.value">
          <input type="text" name="card_color" value="{{.Site.CardColor}}">
        </div>
      </div>
      <button type="submit" class="btn btn-primary btn-sm" style="margin-top:8px">Save Site Settings</button>
    </form>
  </div>

  <!-- Add Gallery -->
  <div class="section">
    <div class="section-title">Add Gallery</div>
    <form method="POST" action="/admin/new">
      <div class="add-form">
        <div class="field">
          <label>Capture One Live URL</label>
          <input type="text" name="capture_one_url" placeholder="https://live.captureone.com/your-gallery-id" required>
        </div>
        <button type="submit" class="btn btn-primary">Add &amp; Sync</button>
      </div>
      <div class="checkbox-field">
        <input type="checkbox" name="is_private" id="new_private">
        <label for="new_private">Private gallery (accessible only via secret link)</label>
      </div>
    </form>
  </div>

  <!-- Gallery List -->
  <div class="section">
    <div class="section-title">Galleries ({{len .Galleries}})</div>
    {{if .Galleries}}
    <div class="gallery-grid">
      {{range .Galleries}}
      <a class="gallery-card" href="/admin/gallery/{{.ID}}">
        <div class="gallery-card-cover" {{if .Photos}}style="background-image: url('/proxy/image?url={{(index .Photos 0).SmallURL}}')"{{end}}>
          <span class="photo-count">{{len .Photos}} photos</span>
        </div>
        <div class="gallery-card-body">
          <h3>{{.Config.GalleryTitle}}{{if .Config.IsPrivate}}<span class="badge badge-private">Private</span>{{else}}<span class="badge badge-public">Public</span>{{end}}</h3>
          <span class="slug">{{if .Config.IsPrivate}}/s/{{.Config.SecretToken}}{{else}}/{{.Config.Slug}}{{end}}</span>
        </div>
      </a>
      {{end}}
    </div>
    {{else}}
    <div class="empty-state">
      <p>No galleries yet</p>
      <span class="hint">Paste a Capture One Live URL above to create your first gallery</span>
    </div>
    {{end}}
  </div>

  <!-- Media Library -->
  <div class="section">
    <div class="section-title">Media Library</div>
    <form method="POST" action="/admin/media/add">
      <div class="add-form">
        <div class="field">
          <label>YouTube URL</label>
          <input type="text" name="youtube_url" placeholder="https://www.youtube.com/watch?v=..." required>
        </div>
        <button type="submit" class="btn btn-primary">Download Audio</button>
      </div>
    </form>
    {{if .Songs}}
    <div style="margin-top:16px">
      {{range .Songs}}
      <div style="display:flex;align-items:center;gap:12px;padding:10px 14px;background:#222;border-radius:8px;margin-bottom:8px">
        <div style="flex:1">
          <div style="font-size:14px;color:#f0f0f0">{{.Title}}</div>
          <div style="font-size:12px;color:#888">{{.Artist}} &middot; {{printf "%.0f" .Duration}}s</div>
        </div>
        <form method="POST" action="/admin/media/delete" style="margin:0" onsubmit="return confirm('Delete this song?')">
          <input type="hidden" name="song_id" value="{{.ID}}">
          <button type="submit" style="background:none;border:none;color:#e57373;cursor:pointer;font-size:13px">Delete</button>
        </form>
      </div>
      {{end}}
    </div>
    {{else}}
    <p style="font-size:13px;color:#666;margin-top:12px">No songs yet. Add a YouTube URL to download audio for slideshows.</p>
    {{end}}
  </div>

  <!-- Account -->
  <div class="section">
    <div class="section-title">Account</div>
    <form method="POST" action="/admin/password">
      <div class="field-row">
        <div class="field">
          <label>Username</label>
          <input type="text" name="username" value="{{.Auth.Username}}" autocomplete="username">
        </div>
        <div class="field">
          <label>New Password</label>
          <input type="password" name="new_password" placeholder="Leave empty to keep current" autocomplete="new-password">
        </div>
      </div>
      <button type="submit" class="btn btn-primary btn-sm" style="margin-top:8px">Update Credentials</button>
    </form>
  </div>
</div>
</body>
</html>`

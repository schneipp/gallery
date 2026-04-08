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
    <form method="POST" action="/admin/logout" style="margin:0">
      <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
      <button type="submit" style="background:none;border:none;color:#888;cursor:pointer;font-size:14px;font-family:inherit">Logout</button>
    </form>
  </div>
</div>
<div class="container">

  <!-- Site Settings -->
  <div class="section">
    <div class="section-title">Site Settings</div>
    <form method="POST" action="/admin/site/save">
      <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
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
      <div class="field-row">
        <div class="field">
          <label>Nextcloud URL</label>
          <input type="text" name="nextcloud_url" value="{{.Site.NextcloudURL}}" placeholder="https://cloud.example.com">
        </div>
        <div class="field">
          <label>Nextcloud Username</label>
          <input type="text" name="nextcloud_user" value="{{.Site.NextcloudUser}}" placeholder="username">
        </div>
      </div>
      <div class="field" style="max-width:50%">
        <label>Nextcloud App Token</label>
        <input type="password" name="nextcloud_token" placeholder="{{if .Site.NextcloudToken}}••••••••  (saved, leave empty to keep){{else}}App-specific password{{end}}">
      </div>
      <button type="submit" class="btn btn-primary btn-sm" style="margin-top:8px">Save Site Settings</button>
    </form>
  </div>

  <!-- Add Gallery -->
  <div class="section">
    <div class="section-title">Add Gallery</div>
    <form method="POST" action="/admin/new" id="newGalleryForm">
      <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
      <div class="field">
        <label>Source Type</label>
        <select name="source_type" id="newSourceType" onchange="toggleNewSourceFields()" style="width:auto;min-width:200px">
          <option value="captureone" selected>Capture One Live</option>
          <option value="nextcloud">Nextcloud</option>
        </select>
      </div>

      <!-- Capture One Fields -->
      <div id="new-captureone-fields">
        <div class="add-form">
          <div class="field">
            <label>Capture One Live URL</label>
            <input type="text" name="capture_one_url" placeholder="https://live.captureone.com/your-gallery-id">
          </div>
          <button type="submit" class="btn btn-primary">Add &amp; Sync</button>
        </div>
      </div>

      <!-- Nextcloud Fields -->
      <div id="new-nextcloud-fields" style="display:none">
        {{if .Site.NextcloudURL}}
        <input type="hidden" name="nextcloud_folder" id="nc_selected_folder" value="">
        <div class="field">
          <label>Select Folder from Nextcloud</label>
          <div id="nc-folder-browser" style="background:#111;border:1px solid #333;border-radius:8px;padding:12px;max-height:300px;overflow-y:auto">
            <div id="nc-breadcrumb" style="font-size:12px;color:#888;margin-bottom:8px;display:flex;gap:4px;flex-wrap:wrap"></div>
            <div id="nc-folder-list" style="color:#999;font-size:13px">Click "Load Folders" to browse...</div>
          </div>
          <button type="button" class="btn btn-primary btn-sm" style="margin-top:8px" onclick="loadNCFolders('')">Load Folders</button>
        </div>
        <div id="nc-selected-display" style="margin-top:8px;display:none">
          <span style="font-size:13px;color:#c8a97e">Selected: </span>
          <span id="nc-selected-name" style="font-size:13px;color:#fff;font-family:monospace"></span>
        </div>
        <button type="submit" class="btn btn-primary" style="margin-top:16px" id="nc-submit-btn" disabled>Add &amp; Sync</button>
        {{else}}
        <p style="color:#e57373;font-size:14px">Configure Nextcloud credentials in Site Settings above first.</p>
        {{end}}
      </div>

      <div class="checkbox-field">
        <input type="checkbox" name="is_private" id="new_private">
        <label for="new_private">Private gallery (accessible only via secret link)</label>
      </div>
    </form>
    <script>
      function toggleNewSourceFields() {
        const sourceType = document.getElementById('newSourceType').value;
        document.getElementById('new-captureone-fields').style.display = sourceType === 'captureone' ? 'block' : 'none';
        document.getElementById('new-nextcloud-fields').style.display = sourceType === 'nextcloud' ? 'block' : 'none';

        const c1Input = document.querySelector('[name="capture_one_url"]');
        if (sourceType === 'captureone') {
          c1Input.setAttribute('required', 'required');
        } else {
          c1Input.removeAttribute('required');
        }
      }
      toggleNewSourceFields();

      function loadNCFolders(path) {
        const list = document.getElementById('nc-folder-list');
        const breadcrumb = document.getElementById('nc-breadcrumb');
        list.innerHTML = '<span style="color:#888">Loading...</span>';

        fetch('/admin/nextcloud/folders?path=' + encodeURIComponent(path))
          .then(r => r.json())
          .then(data => {
            if (data.error) {
              list.innerHTML = '<span style="color:#e57373">' + data.error + '</span>';
              return;
            }

            // Build breadcrumb
            breadcrumb.innerHTML = '';
            const parts = path ? path.split('/') : [];
            const rootLink = document.createElement('a');
            rootLink.textContent = '🏠 Root';
            rootLink.style.cssText = 'color:#c8a97e;cursor:pointer;text-decoration:none';
            rootLink.onclick = () => loadNCFolders('');
            breadcrumb.appendChild(rootLink);
            let cumPath = '';
            parts.forEach((p, i) => {
              const sep = document.createElement('span');
              sep.textContent = ' / ';
              sep.style.color = '#555';
              breadcrumb.appendChild(sep);
              cumPath += (i > 0 ? '/' : '') + p;
              const link = document.createElement('a');
              link.textContent = p;
              link.style.cssText = 'color:#c8a97e;cursor:pointer;text-decoration:none';
              const linkPath = cumPath;
              link.onclick = () => loadNCFolders(linkPath);
              breadcrumb.appendChild(link);
            });

            list.innerHTML = '';
            if (!data.folders || data.folders.length === 0) {
              list.innerHTML = '<span style="color:#666">No subfolders</span>';
            }

            // "Select this folder" button if we're inside a folder
            if (path) {
              const selectBtn = document.createElement('div');
              selectBtn.style.cssText = 'padding:8px 12px;margin-bottom:4px;background:#1a3a1a;border:1px solid #2a5a2a;border-radius:6px;cursor:pointer;color:#81c784;font-size:13px';
              selectBtn.textContent = '✓ Use this folder: /' + path;
              selectBtn.onclick = () => selectNCFolder(path);
              list.prepend(selectBtn);
            }

            (data.folders || []).forEach(f => {
              const item = document.createElement('div');
              item.style.cssText = 'padding:8px 12px;margin-bottom:2px;border-radius:6px;cursor:pointer;transition:background 0.15s;font-size:13px;color:#e0e0e0';
              item.textContent = '📁 ' + f.split('/').pop();
              item.onmouseenter = () => item.style.background = '#222';
              item.onmouseleave = () => item.style.background = 'transparent';
              item.onclick = () => loadNCFolders(f);
              list.appendChild(item);
            });
          })
          .catch(err => {
            list.innerHTML = '<span style="color:#e57373">Error: ' + err.message + '</span>';
          });
      }

      function selectNCFolder(path) {
        document.getElementById('nc_selected_folder').value = path;
        document.getElementById('nc-selected-display').style.display = 'block';
        document.getElementById('nc-selected-name').textContent = '/' + path;
        document.getElementById('nc-submit-btn').disabled = false;
      }
    </script>
  </div>

  <!-- Gallery List -->
  <div class="section">
    <div class="section-title">Galleries ({{len .Galleries}})</div>
    {{if .Galleries}}
    <div class="gallery-grid">
      {{range .Galleries}}
      <a class="gallery-card" href="/admin/gallery/{{.ID}}">
        <div class="gallery-card-cover" {{if .Photos}}style="background-image: url('{{proxyURL (index .Photos 0).SmallURL}}')"{{end}}>
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
      <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
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
          <input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
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
      <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
      <div class="field-row">
        <div class="field">
          <label>Username</label>
          <input type="text" name="username" value="{{.Auth.Username}}" autocomplete="username">
        </div>
        <div class="field">
          <label>Current Password</label>
          <input type="password" name="current_password" placeholder="Required to confirm" autocomplete="current-password" required>
        </div>
      </div>
      <div class="field-row">
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

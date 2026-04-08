package main

const adminEditHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Edit: {{.Config.GalleryTitle}} - Admin</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    background: #111; color: #e0e0e0; min-height: 100vh;
  }
  .topbar {
    background: #1a1a1a; border-bottom: 1px solid #2a2a2a;
    padding: 16px 32px; display: flex; align-items: center; justify-content: space-between;
  }
  .topbar-left { display: flex; align-items: center; gap: 16px; }
  .topbar h1 { font-size: 18px; font-weight: 600; color: #c8a97e; letter-spacing: 1px; }
  .topbar a, .topbar-link {
    color: #888; text-decoration: none; font-size: 14px; transition: color 0.2s;
  }
  .topbar a:hover, .topbar-link:hover { color: #c8a97e; }
  .topbar-right { display: flex; gap: 16px; align-items: center; }
  .container { max-width: 900px; margin: 0 auto; padding: 32px 24px; }
  .section {
    background: #1a1a1a; border: 1px solid #2a2a2a;
    border-radius: 12px; padding: 28px; margin-bottom: 24px;
  }
  .section-title {
    font-size: 15px; font-weight: 600; color: #c8a97e;
    text-transform: uppercase; letter-spacing: 1.5px;
    margin-bottom: 20px; padding-bottom: 12px;
    border-bottom: 1px solid #2a2a2a;
  }
  .field { margin-bottom: 18px; }
  .field label { display: block; font-size: 13px; color: #999; margin-bottom: 6px; font-weight: 500; }
  .field input[type="text"], .field input[type="number"], .field select, .field textarea {
    width: 100%; padding: 10px 14px; background: #111; border: 1px solid #333;
    border-radius: 6px; color: #e0e0e0; font-size: 14px; font-family: inherit; transition: border-color 0.2s;
  }
  .field input:focus, .field select:focus, .field textarea:focus { outline: none; border-color: #c8a97e; }
  .field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  .field-row-3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 16px; }
  .color-field { display: flex; align-items: center; gap: 10px; }
  .color-field input[type="color"] {
    width: 36px; height: 36px; border: 1px solid #333;
    border-radius: 6px; background: #111; cursor: pointer; padding: 2px;
  }
  .color-field input[type="text"] { flex: 1; }
  .checkbox-field { display: flex; align-items: center; gap: 10px; }
  .checkbox-field input[type="checkbox"] { width: 18px; height: 18px; accent-color: #c8a97e; }
  .btn {
    display: inline-flex; align-items: center; gap: 8px;
    padding: 10px 24px; border: none; border-radius: 8px;
    font-size: 14px; font-weight: 600; cursor: pointer;
    transition: all 0.2s; font-family: inherit; text-decoration: none;
  }
  .btn-primary { background: #c8a97e; color: #111; }
  .btn-primary:hover { background: #d4b88f; }
  .btn-sync { background: #2a5a3a; color: #81c784; border: 1px solid #3a7a4a; }
  .btn-sync:hover { background: #3a7a4a; }
  .btn-danger { background: #5a2a2a; color: #e57373; border: 1px solid #7a3a3a; }
  .btn-danger:hover { background: #7a3a3a; }
  .btn-group { display: flex; gap: 12px; margin-top: 8px; }
  .sync-section { display: flex; gap: 12px; align-items: flex-end; }
  .sync-section .field { flex: 1; margin-bottom: 0; }
  .photo-count { font-size: 13px; color: #888; margin-top: 12px; }
  .photo-count strong { color: #c8a97e; }
  .preview-grid {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 8px; margin-top: 16px;
  }
  .preview-thumb {
    aspect-ratio: 1; border-radius: 6px; background-size: cover;
    background-position: center; border: 1px solid #2a2a2a;
  }
  .danger-zone {
    border-color: #4a2020;
  }
  .danger-zone .section-title { color: #e57373; }
  .badge {
    display: inline-block; font-size: 10px; font-weight: 600;
    letter-spacing: 0.5px; padding: 2px 8px; border-radius: 4px;
    text-transform: uppercase; margin-left: 8px; vertical-align: middle;
  }
  .badge-private { background: rgba(229,115,115,0.2); color: #e57373; }
  .badge-public { background: rgba(129,199,132,0.2); color: #81c784; }
  .checkbox-field { display: flex; align-items: center; gap: 10px; }
  .checkbox-field input[type="checkbox"] { width: 18px; height: 18px; accent-color: #c8a97e; }
  .secret-link-box {
    display: flex; align-items: center; gap: 8px; margin-top: 14px;
    padding: 12px 16px; background: #111; border: 1px solid #333;
    border-radius: 8px;
  }
  .secret-link-box input {
    flex: 1; background: none; border: none; color: #c8a97e;
    font-family: monospace; font-size: 13px; outline: none;
  }
  .secret-link-box .copy-btn {
    padding: 6px 14px; background: #2a2a2a; color: #ccc;
    border: 1px solid #3a3a3a; border-radius: 6px; cursor: pointer;
    font-size: 12px; font-family: inherit; transition: all 0.2s;
  }
  .secret-link-box .copy-btn:hover { background: #3a3a3a; color: #fff; }
  .regen-btn {
    padding: 6px 14px; background: #2a3a4a; color: #7ab8e0;
    border: 1px solid #3a5a7a; border-radius: 6px; cursor: pointer;
    font-size: 12px; font-family: inherit; transition: all 0.2s;
    margin-left: 4px;
  }
  .regen-btn:hover { background: #3a5a7a; }
</style>
</head>
<body>
<div class="topbar">
  <div class="topbar-left">
    <a href="/admin">&larr; All Galleries</a>
    <h1>{{.Config.GalleryTitle}}</h1>
  </div>
  <div class="topbar-right">
    <a href="/{{.Config.Slug}}" target="_blank">View Gallery &rarr;</a>
  </div>
</div>
<div class="container">

  <!-- Sync Section -->
  <div class="section">
    <div class="section-title">Source</div>
    <form method="POST" action="/admin/gallery/sync">
      <input type="hidden" name="gallery_id" value="{{.ID}}">
      <div class="field">
        <label>Source Type</label>
        <select name="source_type" id="sourceType" onchange="toggleSourceFields()">
          <option value="captureone" {{if eq .Config.SourceType "captureone"}}selected{{end}}>Capture One Live</option>
          <option value="nextcloud" {{if eq .Config.SourceType "nextcloud"}}selected{{end}}>Nextcloud</option>
        </select>
      </div>

      <!-- Capture One Fields -->
      <div id="captureone-fields" class="sync-section" style="display:flex;gap:12px;align-items:flex-end">
        <div class="field" style="flex:1;margin-bottom:0">
          <label>Capture One Live URL</label>
          <input type="text" name="capture_one_url" value="{{.Config.CaptureOneURL}}" placeholder="https://live.captureone.com/...">
        </div>
        <button type="submit" class="btn btn-sync">Re-sync Photos</button>
      </div>

      <!-- Nextcloud Fields -->
      <div id="nextcloud-fields" style="display:none">
        <div class="field">
          <label>Folder Path</label>
          <input type="text" name="nextcloud_folder" value="{{.Config.NextcloudFolder}}" placeholder="Photos/Gallery" style="margin-bottom:8px">
          <span style="font-size:12px;color:#666">Using credentials from Site Settings</span>
        </div>
        <button type="submit" class="btn btn-sync" style="margin-top:12px">Re-sync Photos</button>
      </div>
    </form>
    {{if .Photos}}
    <div class="photo-count"><strong>{{len .Photos}}</strong> photos synced</div>
    <div class="preview-grid">
      {{range .Photos}}
      <div class="preview-thumb" style="background-image: url('{{proxyURL .SmallURL}}')"></div>
      {{end}}
    </div>
    {{end}}
  </div>
  <script>
    function toggleSourceFields() {
      const sourceType = document.getElementById('sourceType').value;
      document.getElementById('captureone-fields').style.display = sourceType === 'captureone' ? 'flex' : 'none';
      document.getElementById('nextcloud-fields').style.display = sourceType === 'nextcloud' ? 'block' : 'none';
    }
    toggleSourceFields();
  </script>

  <!-- Slideshow -->
  <div class="section">
    <div class="section-title">Slideshow</div>
    <form method="POST" action="/admin/gallery/set-song" id="songForm">
      <input type="hidden" name="gallery_id" value="{{.ID}}">
      <div id="songList">
        {{if .SelectedSongs}}
        {{range $i, $s := .SelectedSongs}}
        <div class="song-row" style="display:flex;gap:10px;align-items:center;margin-bottom:8px">
          <span style="color:#666;font-size:12px;width:20px">{{$i | printf "%d"}}.</span>
          <select name="song_ids" style="flex:1;padding:8px 12px;background:#111;border:1px solid #333;border-radius:6px;color:#e0e0e0;font-size:13px">
            <option value="">— Remove —</option>
            {{range $.Songs}}
            <option value="{{.ID}}" {{if eq .ID $s.ID}}selected{{end}}>{{.Title}} — {{.Artist}} ({{printf "%.0f" .Duration}}s)</option>
            {{end}}
          </select>
          <button type="button" onclick="this.parentElement.remove()" style="background:none;border:none;color:#e57373;cursor:pointer;font-size:18px">&times;</button>
        </div>
        {{end}}
        {{else}}
        <p style="font-size:13px;color:#666;margin-bottom:12px" id="noSongsMsg">No songs selected. Add songs to enable slideshow.</p>
        {{end}}
      </div>
      <div style="display:flex;gap:12px;align-items:center;margin-top:12px">
        <button type="button" class="btn btn-sync btn-sm" onclick="addSongRow()">+ Add Song</button>
        <button type="submit" class="btn btn-primary btn-sm">Save Playlist</button>
      </div>
      {{if .SelectedSongs}}
      <p style="font-size:12px;color:#888;margin-top:12px">
        Total: {{printf "%.0f" .TotalDuration}}s of music ÷ {{len .Photos}} photos = {{printf "%.1f" (divf .TotalDuration (len .Photos))}}s per slide
      </p>
      {{end}}
    </form>
  </div>
  <script>
    function addSongRow() {
      const msg = document.getElementById('noSongsMsg');
      if (msg) msg.remove();
      const list = document.getElementById('songList');
      const idx = list.querySelectorAll('.song-row').length;
      const row = document.createElement('div');
      row.className = 'song-row';
      row.style = 'display:flex;gap:10px;align-items:center;margin-bottom:8px';
      row.innerHTML = '<span style="color:#666;font-size:12px;width:20px">' + idx + '.</span>' +
        '<select name="song_ids" style="flex:1;padding:8px 12px;background:#111;border:1px solid #333;border-radius:6px;color:#e0e0e0;font-size:13px">' +
        '<option value="">— Select song —</option>' +
        {{range .Songs}}'<option value="{{.ID}}">{{.Title}} — {{.Artist}} ({{printf "%.0f" .Duration}}s)</option>' +{{end}}
        '</select>' +
        '<button type="button" onclick="this.parentElement.remove()" style="background:none;border:none;color:#e57373;cursor:pointer;font-size:18px">&times;</button>';
      list.appendChild(row);
    }
  </script>

  <!-- Visibility -->
  <div class="section">
    <div class="section-title">Visibility &amp; Access</div>
    <div style="display:flex;align-items:center;gap:16px;margin-bottom:12px">
      <span style="font-size:14px">Status:</span>
      {{if .Config.IsPrivate}}<span class="badge badge-private">Private</span>{{else}}<span class="badge badge-public">Public</span>{{end}}
    </div>
    <p style="font-size:13px;color:#888;margin-bottom:12px">
      {{if .Config.IsPrivate}}This gallery is only accessible via the secret link below.{{else}}This gallery is visible on the homepage and accessible at <a href="/{{.Config.Slug}}" target="_blank" style="color:#c8a97e">/{{.Config.Slug}}</a>.{{end}}
    </p>
    <div style="font-size:13px;color:#999;margin-bottom:6px">Secret access link (always works, even for public galleries):</div>
    <div class="secret-link-box">
      <input type="text" id="secretLink" value="http://{{.Host}}/s/{{.Config.SecretToken}}" readonly>
      <button type="button" class="copy-btn" onclick="navigator.clipboard.writeText(document.getElementById('secretLink').value);this.textContent='Copied!';setTimeout(()=>this.textContent='Copy',1500)">Copy</button>
      <form method="POST" action="/admin/gallery/regen-token" style="display:inline">
        <input type="hidden" name="gallery_id" value="{{.ID}}">
        <button type="submit" class="regen-btn" onclick="return confirm('Regenerate token? The old link will stop working.')">Regenerate</button>
      </form>
    </div>
  </div>

  <!-- Config Form -->
  <form method="POST" action="/admin/gallery/save">
    <input type="hidden" name="gallery_id" value="{{.ID}}">
    <input type="hidden" name="capture_one_url" value="{{.Config.CaptureOneURL}}">

    <!-- Branding -->
    <div class="section">
      <div class="section-title">Branding</div>
      <div class="field-row-3">
        <div class="field">
          <label>Gallery Title</label>
          <input type="text" name="gallery_title" value="{{.Config.GalleryTitle}}">
        </div>
        <div class="field">
          <label>URL Slug</label>
          <input type="text" name="slug" value="{{.Config.Slug}}" placeholder="my-gallery">
        </div>
        <div class="field">
          <label>Cover Photo Index</label>
          <input type="number" name="cover_index" value="{{.Config.CoverIndex}}" min="0">
        </div>
      </div>
      <div class="field-row">
        <div class="field">
          <label>Subtitle</label>
          <input type="text" name="subtitle" value="{{.Config.Subtitle}}">
        </div>
        <div class="field">
          <label>Footer Text</label>
          <input type="text" name="footer_text" value="{{.Config.FooterText}}">
        </div>
      </div>
      <div class="field">
        <label>Logo URL (optional)</label>
        <input type="text" name="logo_url" value="{{.Config.LogoURL}}" placeholder="https://...">
      </div>
      <div class="field-row">
        <div class="field">
          <div class="checkbox-field">
            <input type="checkbox" name="show_filenames" id="show_filenames" {{if .Config.ShowFilenames}}checked{{end}}>
            <label for="show_filenames" style="margin-bottom:0">Show photo filenames</label>
          </div>
        </div>
        <div class="field">
          <div class="checkbox-field">
            <input type="checkbox" name="is_private" id="is_private" {{if .Config.IsPrivate}}checked{{end}}>
            <label for="is_private" style="margin-bottom:0">Private gallery</label>
          </div>
        </div>
      </div>
    </div>

    <!-- Colors -->
    <div class="section">
      <div class="section-title">Colors</div>
      <div class="field-row">
        <div class="field">
          <label>Background</label>
          <div class="color-field">
            <input type="color" value="{{.Config.BackgroundColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="background_color" value="{{.Config.BackgroundColor}}">
          </div>
        </div>
        <div class="field">
          <label>Card Background</label>
          <div class="color-field">
            <input type="color" value="{{.Config.CardColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="card_color" value="{{.Config.CardColor}}">
          </div>
        </div>
      </div>
      <div class="field-row">
        <div class="field">
          <label>Text Color</label>
          <div class="color-field">
            <input type="color" value="{{.Config.TextColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="text_color" value="{{.Config.TextColor}}">
          </div>
        </div>
        <div class="field">
          <label>Accent Color</label>
          <div class="color-field">
            <input type="color" value="{{.Config.AccentColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="accent_color" value="{{.Config.AccentColor}}">
          </div>
        </div>
      </div>
      <div class="field">
        <label>Lightbox Background</label>
        <input type="text" name="lightbox_bg" value="{{.Config.LightboxBg}}">
      </div>
    </div>

    <!-- Photo Style -->
    <div class="section">
      <div class="section-title">Photo Style</div>
      <div class="field-row">
        <div class="field">
          <label>Frame Style</label>
          <select name="frame_style">
            <option value="none" {{if eq .Config.FrameStyle "none"}}selected{{end}}>None (use borders below)</option>
            <option value="polaroid" {{if eq .Config.FrameStyle "polaroid"}}selected{{end}}>Polaroid</option>
            <option value="print" {{if eq .Config.FrameStyle "print"}}selected{{end}}>Classic Print</option>
            <option value="darkroom" {{if eq .Config.FrameStyle "darkroom"}}selected{{end}}>Dark Room</option>
            <option value="museum" {{if eq .Config.FrameStyle "museum"}}selected{{end}}>Museum Mat</option>
          </select>
        </div>
        <div class="field">
          <label>Hover Effect</label>
          <select name="hover_effect">
            <option value="lift" {{if eq .Config.HoverEffect "lift"}}selected{{end}}>Lift</option>
            <option value="zoom" {{if eq .Config.HoverEffect "zoom"}}selected{{end}}>Zoom</option>
            <option value="glow" {{if eq .Config.HoverEffect "glow"}}selected{{end}}>Glow</option>
            <option value="none" {{if eq .Config.HoverEffect "none"}}selected{{end}}>None</option>
          </select>
        </div>
      </div>
      <div class="field-row-3">
        <div class="field">
          <label>Border Radius</label>
          <input type="text" name="border_radius" value="{{.Config.BorderRadius}}">
        </div>
        <div class="field">
          <label>Border Width</label>
          <input type="text" name="border_width" value="{{.Config.BorderWidth}}">
        </div>
        <div class="field">
          <label>Border Color</label>
          <div class="color-field">
            <input type="color" value="{{.Config.BorderColor}}" onchange="this.nextElementSibling.value=this.value">
            <input type="text" name="border_color" value="{{.Config.BorderColor}}">
          </div>
        </div>
      </div>
      <div class="field">
        <label>Shadow</label>
        <input type="text" name="shadow" value="{{.Config.Shadow}}">
      </div>
    </div>

    <!-- Layout -->
    <div class="section">
      <div class="section-title">Layout</div>
      <div class="field-row-3">
        <div class="field">
          <label>Layout Style</label>
          <select name="layout">
            <option value="masonry" {{if eq .Config.Layout "masonry"}}selected{{end}}>Masonry</option>
            <option value="grid" {{if eq .Config.Layout "grid"}}selected{{end}}>Grid</option>
            <option value="justified" {{if eq .Config.Layout "justified"}}selected{{end}}>Justified</option>
          </select>
        </div>
        <div class="field">
          <label>Gap</label>
          <input type="text" name="column_gap" value="{{.Config.ColumnGap}}">
        </div>
        <div class="field">
          <label>Max Columns</label>
          <input type="number" name="max_columns" value="{{.Config.MaxColumns}}" min="1" max="8">
        </div>
      </div>
      <div class="field" style="margin-top:4px">
        <label>Slideshow Transition</label>
        <select name="slideshow_transition">
          <option value="fade" {{if eq .Config.SlideshowTransition "fade"}}selected{{end}}>Crossfade</option>
          <option value="slide" {{if eq .Config.SlideshowTransition "slide"}}selected{{end}}>Slide</option>
          <option value="flip" {{if eq .Config.SlideshowTransition "flip"}}selected{{end}}>Flip</option>
          <option value="zoom" {{if eq .Config.SlideshowTransition "zoom"}}selected{{end}}>Zoom</option>
          <option value="drop" {{if eq .Config.SlideshowTransition "drop"}}selected{{end}}>Drop</option>
          <option value="blur" {{if eq .Config.SlideshowTransition "blur"}}selected{{end}}>Blur</option>
          <option value="stack" {{if eq .Config.SlideshowTransition "stack"}}selected{{end}}>Card Stack 3D</option>
          <option value="filmstrip" {{if eq .Config.SlideshowTransition "filmstrip"}}selected{{end}}>Film Strip</option>
          <option value="random" {{if eq .Config.SlideshowTransition "random"}}selected{{end}}>Random</option>
        </select>
      </div>
    </div>

    <div class="btn-group">
      <button type="submit" class="btn btn-primary">Save Settings</button>
    </div>
  </form>

  <!-- Danger Zone -->
  <div class="section danger-zone" style="margin-top: 40px;">
    <div class="section-title">Danger Zone</div>
    <p style="font-size:14px;color:#999;margin-bottom:16px">Permanently delete this gallery and all its photos.</p>
    <form method="POST" action="/admin/gallery/delete" onsubmit="return confirm('Delete this gallery? This cannot be undone.')">
      <input type="hidden" name="gallery_id" value="{{.ID}}">
      <button type="submit" class="btn btn-danger">Delete Gallery</button>
    </form>
  </div>
</div>
<script>
  document.querySelectorAll('.color-field').forEach(cf => {
    const colorInput = cf.querySelector('input[type="color"]');
    const textInput = cf.querySelector('input[type="text"]');
    if (colorInput && textInput) {
      textInput.addEventListener('input', () => { colorInput.value = textInput.value; });
    }
  });
</script>
</body>
</html>`

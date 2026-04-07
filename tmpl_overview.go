package main

const overviewHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Site.SiteTitle}}</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Playfair+Display:wght@400;600;700&family=Inter:wght@300;400;500&display=swap" rel="stylesheet">
<style>
  :root {
    --accent: {{.Site.AccentColor}};
    --bg: {{.Site.BgColor}};
    --text: {{.Site.TextColor}};
    --card: {{.Site.CardColor}};
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  html { scroll-behavior: smooth; }
  body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
  }

  .site-header {
    text-align: center;
    padding: 100px 24px 60px;
    position: relative;
  }
  .site-header::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 60px;
    height: 2px;
    background: var(--accent);
  }
  .site-logo {
    max-height: 60px;
    margin-bottom: 24px;
    filter: brightness(0.9);
  }
  .site-title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: clamp(36px, 5vw, 64px);
    font-weight: 700;
    letter-spacing: 2px;
    margin-bottom: 12px;
  }
  .site-subtitle {
    font-size: 16px;
    font-weight: 300;
    color: color-mix(in srgb, var(--text) 50%, transparent);
    letter-spacing: 3px;
    text-transform: uppercase;
  }

  .galleries-wrap {
    max-width: 1400px;
    margin: 0 auto;
    padding: 60px 24px 80px;
  }

  .galleries-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
    gap: 32px;
  }
  @media (max-width: 500px) {
    .galleries-grid { grid-template-columns: 1fr; }
  }

  .gallery-card {
    position: relative;
    border-radius: 12px;
    overflow: hidden;
    background: var(--card);
    text-decoration: none;
    color: inherit;
    display: block;
    transition: transform 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94),
                box-shadow 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
    box-shadow: 0 8px 32px rgba(0,0,0,0.3);
    opacity: 0;
    animation: cardIn 0.6s ease forwards;
  }
  .gallery-card:hover {
    transform: translateY(-6px);
    box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  }
  @keyframes cardIn {
    from { opacity: 0; transform: translateY(24px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .gallery-card-image {
    aspect-ratio: 3/2;
    background-size: cover;
    background-position: center;
    background-color: #1a1a1a;
    position: relative;
  }
  .gallery-card-image::after {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(transparent 40%, rgba(0,0,0,0.6));
  }

  .gallery-card-info {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    padding: 24px;
    z-index: 1;
  }
  .gallery-card-title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 26px;
    font-weight: 600;
    margin-bottom: 6px;
    color: #fff;
  }
  .gallery-card-meta {
    font-size: 13px;
    color: rgba(255,255,255,0.6);
    letter-spacing: 1px;
  }
  .gallery-card-subtitle {
    font-size: 14px;
    color: rgba(255,255,255,0.5);
    margin-top: 4px;
    font-weight: 300;
  }

  .gallery-card-count {
    position: absolute;
    top: 16px;
    right: 16px;
    background: rgba(0,0,0,0.5);
    color: rgba(255,255,255,0.8);
    font-size: 12px;
    padding: 4px 12px;
    border-radius: 16px;
    backdrop-filter: blur(8px);
    z-index: 1;
    letter-spacing: 0.5px;
  }

  .empty-state {
    text-align: center;
    padding: 80px 20px;
    color: color-mix(in srgb, var(--text) 40%, transparent);
  }
  .empty-state p { font-size: 18px; font-family: 'Playfair Display', serif; }

  .site-footer {
    text-align: center;
    padding: 40px 24px;
    font-size: 13px;
    color: color-mix(in srgb, var(--text) 30%, transparent);
    letter-spacing: 1px;
  }
</style>
</head>
<body>
  <div class="site-header">
    {{if .Site.LogoURL}}<img class="site-logo" src="{{.Site.LogoURL}}" alt="Logo">{{end}}
    <h1 class="site-title">{{.Site.SiteTitle}}</h1>
    {{if .Site.SiteSubtitle}}<p class="site-subtitle">{{.Site.SiteSubtitle}}</p>{{end}}
  </div>

  <div class="galleries-wrap">
    {{if .Galleries}}
    <div class="galleries-grid">
      {{range $i, $g := .Galleries}}
      <a class="gallery-card" href="/{{$g.Config.Slug}}" style="animation-delay: {{delay $i}}s">
        <div class="gallery-card-image" {{if $g.Photos}}style="background-image: url('/proxy/image?url={{(index $g.Photos $g.Config.CoverIndex).MediumURL}}')"{{end}}>
          <span class="gallery-card-count">{{len $g.Photos}} photos</span>
          <div class="gallery-card-info">
            <div class="gallery-card-title">{{$g.Config.GalleryTitle}}</div>
            {{if $g.Config.Subtitle}}<div class="gallery-card-subtitle">{{$g.Config.Subtitle}}</div>{{end}}
          </div>
        </div>
      </a>
      {{end}}
    </div>
    {{else}}
    <div class="empty-state">
      <p>No galleries yet</p>
    </div>
    {{end}}
  </div>

  <div class="site-footer"></div>
</body>
</html>`

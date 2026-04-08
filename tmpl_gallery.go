package main

const galleryHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Config.GalleryTitle}}</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Caveat:wght@400;600&family=Playfair+Display:wght@400;600;700&family=Inter:wght@300;400;500&display=swap" rel="stylesheet">
<style>
  :root {
    --bg: {{.Config.BackgroundColor}};
    --card: {{.Config.CardColor}};
    --text: {{.Config.TextColor}};
    --accent: {{.Config.AccentColor}};
    --border-radius: {{.Config.BorderRadius}};
    --border-width: {{.Config.BorderWidth}};
    --border-color: {{.Config.BorderColor}};
    --shadow: {{.Config.Shadow}};
    --gap: {{.Config.ColumnGap}};
    --max-cols: {{.Config.MaxColumns}};
    --lightbox-bg: {{.Config.LightboxBg}};
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  html { scroll-behavior: smooth; }
  body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    overflow-x: hidden;
  }

  /* Back link */
  .back-link {
    position: absolute;
    top: 28px;
    left: 24px;
    z-index: 100;
    color: color-mix(in srgb, var(--text) 50%, transparent);
    text-decoration: none;
    font-size: 13px;
    letter-spacing: 1px;
    transition: color 0.2s;
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .back-link:hover { color: var(--accent); }
  .back-link svg { width: 16px; height: 16px; }

  /* Header */
  .gallery-header {
    text-align: center;
    padding: 80px 24px 48px;
    position: relative;
  }
  .gallery-header::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 60px;
    height: 2px;
    background: var(--accent);
  }
  .gallery-logo {
    max-height: 60px;
    margin-bottom: 24px;
    filter: brightness(0.9);
  }
  .gallery-title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: clamp(32px, 5vw, 56px);
    font-weight: 700;
    letter-spacing: 2px;
    color: var(--text);
    margin-bottom: 12px;
  }
  .gallery-subtitle {
    font-size: 16px;
    font-weight: 300;
    color: color-mix(in srgb, var(--text) 60%, transparent);
    letter-spacing: 3px;
    text-transform: uppercase;
  }
  .gallery-count {
    font-size: 13px;
    color: color-mix(in srgb, var(--text) 35%, transparent);
    margin-top: 16px;
    letter-spacing: 2px;
  }

  /* Gallery Container */
  .gallery-wrap {
    max-width: 1600px;
    margin: 0 auto;
    padding: 48px 24px;
  }

  /* Masonry Layout */
  .gallery-masonry {
    columns: var(--max-cols);
    column-gap: var(--gap);
  }
  @media (max-width: 1200px) { .gallery-masonry { columns: 3; } }
  @media (max-width: 800px) { .gallery-masonry { columns: 2; } }
  @media (max-width: 500px) { .gallery-masonry { columns: 1; } }

  /* Grid Layout */
  .gallery-grid {
    display: grid;
    grid-template-columns: repeat(var(--max-cols), 1fr);
    gap: var(--gap);
  }
  @media (max-width: 1200px) { .gallery-grid { grid-template-columns: repeat(3, 1fr); } }
  @media (max-width: 800px) { .gallery-grid { grid-template-columns: repeat(2, 1fr); } }
  @media (max-width: 500px) { .gallery-grid { grid-template-columns: 1fr; } }

  /* Justified Layout */
  .gallery-justified {
    display: flex;
    flex-wrap: wrap;
    gap: var(--gap);
  }
  .gallery-justified .photo-card {
    flex-grow: 1;
    height: 320px;
  }
  .gallery-justified .photo-card img {
    height: 100%;
    object-fit: cover;
    min-width: 100%;
  }

  /* Photo Card */
  .photo-card {
    break-inside: avoid;
    margin-bottom: var(--gap);
    position: relative;
    cursor: pointer;
    overflow: hidden;
    border-radius: var(--border-radius);
    border: var(--border-width) solid var(--border-color);
    background: var(--card);
    box-shadow: var(--shadow);
    transition: transform 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94),
                box-shadow 0.4s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  }

  /* Hover Effects */
  .hover-lift .photo-card:hover {
    transform: translateY(-8px);
    box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  }
  .hover-zoom .photo-card:hover img {
    transform: scale(1.05);
  }
  .hover-glow .photo-card:hover {
    box-shadow: 0 0 40px color-mix(in srgb, var(--accent) 30%, transparent);
  }

  /* Frame: Polaroid */
  .frame-polaroid .photo-card {
    background: #f5f2ed;
    border: none;
    border-radius: 2px;
    padding: 14px 14px 52px 14px;
    box-shadow: 0 4px 16px rgba(0,0,0,0.25), 0 1px 3px rgba(0,0,0,0.15);
    overflow: visible;
  }
  .frame-polaroid .photo-card img {
    border-radius: 0;
  }
  .frame-polaroid .photo-card .photo-overlay {
    bottom: 0;
    left: 14px;
    right: 14px;
    padding: 8px 4px 4px;
    background: none;
    opacity: 1;
  }
  .frame-polaroid .photo-card .photo-name {
    color: #555;
    font-family: 'Caveat', 'Segoe Print', cursive;
    font-size: 15px;
    letter-spacing: 0;
  }
  .frame-polaroid .photo-card .photo-rating {
    color: #b89a6a;
  }
  /* Slight random tilt for polaroids */
  .frame-polaroid .photo-card:nth-child(3n+1) { transform: rotate(-1.2deg); }
  .frame-polaroid .photo-card:nth-child(3n+2) { transform: rotate(0.8deg); }
  .frame-polaroid .photo-card:nth-child(3n)   { transform: rotate(-0.5deg); }
  .frame-polaroid .photo-card:nth-child(5n+1) { transform: rotate(1.5deg); }
  .frame-polaroid .photo-card:nth-child(7n)   { transform: rotate(-0.9deg); }
  .frame-polaroid.hover-lift .photo-card:hover {
    transform: rotate(0deg) translateY(-10px) scale(1.02);
    box-shadow: 0 24px 48px rgba(0,0,0,0.35);
    z-index: 10;
  }

  /* Frame: Classic Print (thin white border, like a photo print) */
  .frame-print .photo-card {
    background: #fff;
    border: none;
    border-radius: 1px;
    padding: 8px;
    box-shadow: 0 2px 12px rgba(0,0,0,0.2), 0 0 1px rgba(0,0,0,0.1);
  }
  .frame-print .photo-card img {
    border-radius: 0;
  }
  .frame-print .photo-card .photo-overlay {
    left: 8px;
    right: 8px;
    bottom: 8px;
    border-radius: 0;
  }

  /* Frame: Dark Room (black mat with subtle inner border) */
  .frame-darkroom .photo-card {
    background: #1a1a1a;
    border: 1px solid #333;
    border-radius: 0;
    padding: 16px;
    box-shadow: 0 8px 32px rgba(0,0,0,0.5);
  }
  .frame-darkroom .photo-card img {
    border: 1px solid #2a2a2a;
    border-radius: 0;
  }
  .frame-darkroom .photo-card .photo-overlay {
    left: 16px;
    right: 16px;
    bottom: 16px;
  }

  /* Frame: Museum (wide white mat, thin dark frame line) */
  .frame-museum .photo-card {
    background: #fafaf8;
    border: 2px solid #2a2a2a;
    border-radius: 0;
    padding: 28px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.15);
  }
  .frame-museum .photo-card img {
    border-radius: 0;
    box-shadow: inset 0 0 0 1px rgba(0,0,0,0.08);
  }
  .frame-museum .photo-card .photo-overlay {
    left: 28px;
    right: 28px;
    bottom: 28px;
  }
  .frame-museum .photo-card .photo-name {
    color: #444;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 2px;
  }

  .photo-card img {
    display: block;
    width: 100%;
    height: auto;
    transition: transform 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94);
    background: #1a1a1a;
  }

  .photo-card .photo-overlay {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    padding: 24px 16px 14px;
    background: linear-gradient(transparent, rgba(0,0,0,0.7));
    opacity: 0;
    transition: opacity 0.3s;
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
  }
  .photo-card:hover .photo-overlay { opacity: 1; }

  .photo-name {
    font-size: 12px;
    font-weight: 500;
    letter-spacing: 1px;
    color: rgba(255,255,255,0.85);
  }
  .photo-rating {
    font-size: 11px;
    color: var(--accent);
    letter-spacing: 2px;
  }

  /* Fade-in animation */
  .photo-card {
    opacity: 0;
    animation: fadeUp 0.6s ease forwards;
  }
  @keyframes fadeUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
  }

  /* Lightbox */
  .lightbox {
    position: fixed;
    inset: 0;
    background: var(--lightbox-bg);
    z-index: 1000;
    display: none;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(20px);
  }
  .lightbox.active { display: flex; }
  .lightbox-frame {
    position: relative;
    animation: lightboxIn 0.3s ease;
  }
  .lightbox-img {
    max-width: 90vw;
    max-height: 90vh;
    object-fit: contain;
    border-radius: 4px;
    box-shadow: 0 20px 80px rgba(0,0,0,0.6);
    display: block;
  }
  @keyframes lightboxIn {
    from { opacity: 0; transform: scale(0.95); }
    to { opacity: 1; transform: scale(1); }
  }

  /* Lightbox frame: Polaroid */
  .lightbox.lb-polaroid .lightbox-frame {
    background: #f5f2ed;
    padding: 18px 18px 64px 18px;
    border-radius: 2px;
    box-shadow: 0 20px 80px rgba(0,0,0,0.6), 0 2px 8px rgba(0,0,0,0.2);
  }
  .lightbox.lb-polaroid .lightbox-img {
    border-radius: 0;
    box-shadow: none;
    max-width: min(85vw, 1200px);
    max-height: 75vh;
  }
  .lightbox.lb-polaroid .lightbox-frame-name {
    position: absolute;
    bottom: 16px;
    left: 24px;
    font-family: 'Caveat', 'Segoe Print', cursive;
    font-size: 20px;
    color: #555;
  }
  .lightbox.lb-polaroid .lightbox-frame-rating {
    position: absolute;
    bottom: 18px;
    right: 24px;
    font-size: 14px;
    color: #b89a6a;
    letter-spacing: 2px;
  }

  /* Lightbox frame: Print */
  .lightbox.lb-print .lightbox-frame {
    background: #fff;
    padding: 10px;
    border-radius: 1px;
    box-shadow: 0 20px 80px rgba(0,0,0,0.6);
  }
  .lightbox.lb-print .lightbox-img {
    border-radius: 0;
    box-shadow: none;
    max-height: 82vh;
  }

  /* Lightbox frame: Dark Room */
  .lightbox.lb-darkroom .lightbox-frame {
    background: #1a1a1a;
    padding: 20px;
    border: 1px solid #333;
    box-shadow: 0 20px 80px rgba(0,0,0,0.7);
  }
  .lightbox.lb-darkroom .lightbox-img {
    border: 1px solid #2a2a2a;
    border-radius: 0;
    box-shadow: none;
    max-height: 80vh;
  }

  /* Lightbox frame: Museum */
  .lightbox.lb-museum .lightbox-frame {
    background: #fafaf8;
    padding: 36px;
    border: 2px solid #2a2a2a;
    box-shadow: 0 20px 80px rgba(0,0,0,0.5);
  }
  .lightbox.lb-museum .lightbox-img {
    border-radius: 0;
    box-shadow: none;
    max-height: 75vh;
  }
  .lightbox.lb-museum .lightbox-frame-name {
    position: absolute;
    bottom: 8px;
    left: 36px;
    font-size: 11px;
    color: #444;
    text-transform: uppercase;
    letter-spacing: 2px;
  }
  .lightbox-close {
    position: absolute;
    top: 24px;
    right: 24px;
    width: 44px;
    height: 44px;
    border: none;
    background: rgba(255,255,255,0.1);
    color: white;
    font-size: 24px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.2s;
  }
  .lightbox-close:hover { background: rgba(255,255,255,0.2); }
  .lightbox-nav {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    width: 52px;
    height: 52px;
    border: none;
    background: rgba(255,255,255,0.08);
    color: white;
    font-size: 28px;
    border-radius: 50%;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.2s;
  }
  .lightbox-nav { z-index: 10; }
  .lightbox-nav:hover { background: rgba(255,255,255,0.18); }
  .lightbox-prev { left: 24px; }
  .lightbox-next { right: 24px; }
  @media (max-width: 768px) {
    .lightbox-img { max-width: 96vw; max-height: 96vh; }
    .lightbox-nav { width: 40px; height: 40px; font-size: 22px; }
    .lightbox-prev { left: 8px; }
    .lightbox-next { right: 8px; }
    .lightbox-close { top: 12px; right: 12px; width: 36px; height: 36px; font-size: 20px; }
    .lightbox-frame { padding: 6px 6px 40px 6px !important; }
    .lightbox-frame-name { font-size: 16px !important; bottom: 10px !important; left: 12px !important; }
    .lightbox-frame-rating { font-size: 12px !important; bottom: 12px !important; right: 12px !important; }
    .slideshow-pause { top: 12px; right: 56px; width: 36px; height: 36px; }
    .slideshow-close { top: 12px; right: 12px; width: 36px; height: 36px; font-size: 20px; }
  }
  .lightbox-info {
    position: absolute;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    font-size: 13px;
    color: rgba(255,255,255,0.6);
    letter-spacing: 1px;
    text-align: center;
  }
  .lightbox-counter {
    font-size: 12px;
    color: rgba(255,255,255,0.4);
    margin-top: 4px;
  }

  /* Footer */
  .gallery-footer {
    text-align: center;
    padding: 48px 24px;
    font-size: 13px;
    color: color-mix(in srgb, var(--text) 40%, transparent);
    letter-spacing: 1px;
  }

  /* Loading skeleton */
  .photo-card img[data-src] {
    min-height: 200px;
    background: linear-gradient(90deg, #1a1a1a 25%, #222 50%, #1a1a1a 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }
  @keyframes shimmer {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }

  /* Gallery action buttons */
  .gallery-actions {
    display: flex;
    gap: 12px;
    justify-content: center;
    margin-top: 24px;
  }
  .gallery-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 22px;
    border: 1px solid color-mix(in srgb, var(--text) 20%, transparent);
    border-radius: 24px;
    background: transparent;
    color: color-mix(in srgb, var(--text) 70%, transparent);
    font-size: 13px;
    font-family: inherit;
    letter-spacing: 1px;
    cursor: pointer;
    text-decoration: none;
    transition: all 0.3s;
  }
  .gallery-btn:hover {
    border-color: var(--accent);
    color: var(--accent);
    background: color-mix(in srgb, var(--accent) 8%, transparent);
  }

  /* Slideshow */
  .slideshow {
    position: fixed;
    inset: 0;
    z-index: 2000;
    background: #000;
    display: none;
    flex-direction: column;
  }
  .slideshow.active { display: flex; }
  .slideshow-image-wrap {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    position: relative;
  }
  .slideshow-img {
    max-width: 100vw;
    max-height: 100vh;
    object-fit: contain;
    position: absolute;
  }
  /* Transition: Fade */
  .ss-trans-fade .slideshow-img {
    opacity: 0;
    transition: opacity 1.2s ease;
  }
  .ss-trans-fade .slideshow-img.active { opacity: 1; }
  .ss-trans-fade .slideshow-img.exit { opacity: 0; }

  /* Transition: Slide */
  .ss-trans-slide .slideshow-img {
    opacity: 0;
    transform: translateX(100%);
    transition: transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.5s ease;
  }
  .ss-trans-slide .slideshow-img.active { opacity: 1; transform: translateX(0); }
  .ss-trans-slide .slideshow-img.exit { opacity: 0; transform: translateX(-100%); }

  /* Transition: Flip */
  .ss-trans-flip .slideshow-image-wrap { perspective: 1200px; }
  .ss-trans-flip .slideshow-img {
    opacity: 0;
    transform: rotateY(90deg);
    transition: transform 0.8s cubic-bezier(0.4, 0, 0.2, 1), opacity 0.4s ease;
    backface-visibility: hidden;
  }
  .ss-trans-flip .slideshow-img.active { opacity: 1; transform: rotateY(0deg); }
  .ss-trans-flip .slideshow-img.exit { opacity: 0; transform: rotateY(-90deg); }

  /* Transition: Zoom */
  .ss-trans-zoom .slideshow-img {
    opacity: 0;
    transform: scale(0.3);
    transition: transform 0.7s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.5s ease;
  }
  .ss-trans-zoom .slideshow-img.active { opacity: 1; transform: scale(1); }
  .ss-trans-zoom .slideshow-img.exit { opacity: 0; transform: scale(1.5); }

  /* Transition: Drop */
  .ss-trans-drop .slideshow-img {
    opacity: 0;
    transform: translateY(-60px) scale(0.95);
    transition: transform 0.7s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.5s ease;
  }
  .ss-trans-drop .slideshow-img.active { opacity: 1; transform: translateY(0) scale(1); }
  .ss-trans-drop .slideshow-img.exit { opacity: 0; transform: translateY(60px) scale(0.95); }

  /* Transition: Blur */
  .ss-trans-blur .slideshow-img {
    opacity: 0;
    filter: blur(30px);
    transform: scale(1.1);
    transition: opacity 1s ease, filter 1s ease, transform 1s ease;
  }
  .ss-trans-blur .slideshow-img.active { opacity: 1; filter: blur(0); transform: scale(1); }
  .ss-trans-blur .slideshow-img.exit { opacity: 0; filter: blur(30px); transform: scale(0.9); }
  /* Polaroid frame in slideshow */
  .slideshow.ss-polaroid .slideshow-image-wrap {
    padding: 40px;
  }
  .slideshow.ss-polaroid .slideshow-frame {
    background: #f5f2ed;
    padding: 16px 16px 56px 16px;
    border-radius: 2px;
    box-shadow: 0 20px 60px rgba(0,0,0,0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
  }
  .slideshow.ss-polaroid .slideshow-img {
    max-width: calc(100vw - 120px);
    max-height: calc(100vh - 160px);
    position: relative;
    opacity: 1;
  }
  .slideshow.ss-polaroid .slideshow-frame-name {
    position: absolute;
    bottom: 14px;
    left: 22px;
    font-family: 'Caveat', cursive;
    font-size: 20px;
    color: #555;
  }

  /* Print frame in slideshow */
  .slideshow.ss-print .slideshow-frame {
    background: #fff;
    padding: 10px;
    box-shadow: 0 20px 60px rgba(0,0,0,0.5);
    display: flex;
  }
  .slideshow.ss-print .slideshow-img {
    max-width: calc(100vw - 60px);
    max-height: calc(100vh - 60px);
    position: relative;
    opacity: 1;
  }

  /* Darkroom frame in slideshow */
  .slideshow.ss-darkroom .slideshow-frame {
    background: #1a1a1a;
    padding: 18px;
    border: 1px solid #333;
    box-shadow: 0 20px 60px rgba(0,0,0,0.7);
    display: flex;
  }
  .slideshow.ss-darkroom .slideshow-img {
    border: 1px solid #2a2a2a;
    max-width: calc(100vw - 80px);
    max-height: calc(100vh - 80px);
    position: relative;
    opacity: 1;
  }

  /* Museum frame in slideshow */
  .slideshow.ss-museum .slideshow-frame {
    background: #fafaf8;
    padding: 36px;
    border: 2px solid #2a2a2a;
    box-shadow: 0 20px 60px rgba(0,0,0,0.5);
    display: flex;
  }
  .slideshow.ss-museum .slideshow-img {
    max-width: calc(100vw - 120px);
    max-height: calc(100vh - 120px);
    position: relative;
    opacity: 1;
  }

  /* Card Stack 3D mode */
  .slideshow-stack {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    perspective: 1200px;
    perspective-origin: 50% 50%;
  }
  .stack-scene {
    position: relative;
    width: 70vmin;
    height: 85vmin;
    transform-style: preserve-3d;
  }
  @media (max-width: 768px) {
    .stack-scene { width: 88vw; height: 70vh; }
  }
  .stack-card {
    position: absolute;
    inset: 0;
    background: #f5f2ed;
    border-radius: 4px;
    padding: 12px 12px 48px 12px;
    box-shadow: 0 2px 15px rgba(0,0,0,0.3);
    transform-origin: 50% 100%;
    transition: none;
    overflow: hidden;
    backface-visibility: hidden;
  }
  .stack-card img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    border-radius: 1px;
  }
  .stack-card-name {
    position: absolute;
    bottom: 12px;
    left: 16px;
    font-family: 'Caveat', cursive;
    font-size: 18px;
    color: #555;
  }
  /* Stack depth cards behind */
  .stack-card.depth-1 {
    transform: translateZ(-8px) translateY(-4px) rotate(-1.2deg);
    box-shadow: 0 2px 10px rgba(0,0,0,0.2);
  }
  .stack-card.depth-2 {
    transform: translateZ(-16px) translateY(-8px) rotate(0.8deg);
    box-shadow: 0 1px 8px rgba(0,0,0,0.15);
  }
  .stack-card.depth-3 {
    transform: translateZ(-24px) translateY(-12px) rotate(-0.5deg);
    box-shadow: 0 1px 6px rgba(0,0,0,0.1);
    opacity: 0.7;
  }
  .stack-card.depth-4 {
    transform: translateZ(-32px) translateY(-16px) rotate(1deg);
    opacity: 0.4;
  }
  /* Top card = no depth transform */
  .stack-card.top {
    transform: translateZ(0) rotate(0deg);
    box-shadow: 0 8px 40px rgba(0,0,0,0.35), 0 2px 10px rgba(0,0,0,0.2);
    z-index: 10;
  }
  /* Peel animation: card lifts, swings right in an arc, tucks under the stack */
  .stack-card.peeling {
    animation: cardPeel 1.1s cubic-bezier(0.4, 0, 0.2, 1) forwards;
    z-index: 20;
  }
  @keyframes cardPeel {
    0% {
      transform: translateZ(0) rotate(0deg);
      box-shadow: 0 8px 40px rgba(0,0,0,0.35);
    }
    /* Lift up and start swinging right */
    25% {
      transform: translateZ(60px) translateX(15%) translateY(-8%) rotateZ(6deg);
      box-shadow: 0 20px 50px rgba(0,0,0,0.3);
    }
    /* Out to the right, tilted */
    50% {
      transform: translateZ(30px) translateX(75%) translateY(-4%) rotateZ(14deg) rotateY(-10deg);
      box-shadow: -10px 15px 40px rgba(0,0,0,0.25);
    }
    /* Swinging back toward the stack, going behind */
    75% {
      transform: translateZ(-20px) translateX(30%) translateY(-10%) rotateZ(6deg) rotateY(-5deg);
      box-shadow: 0 4px 15px rgba(0,0,0,0.15);
    }
    /* Settle at the back of the stack */
    100% {
      transform: translateZ(-36px) translateY(-14px) rotate(0.8deg);
      box-shadow: 0 1px 6px rgba(0,0,0,0.1);
    }
  }

  /* Filmstrip mode */
  .slideshow-filmstrip {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    overflow: hidden;
  }
  .filmstrip-rail {
    display: flex;
    align-items: center;
    gap: 0;
    padding: 0;
    transition: transform 600ms cubic-bezier(0.4, 0, 0.2, 1), opacity 0.5s ease;
    position: relative;
    background: #0a0a0a;
  }
  .filmstrip-rail::before,
  .filmstrip-rail::after {
    content: '';
    position: absolute;
    top: -18px;
    bottom: -18px;
    left: -40px;
    right: -40px;
    border-top: 18px solid #1a1a1a;
    border-bottom: 18px solid #1a1a1a;
    pointer-events: none;
    z-index: 1;
    background: repeating-linear-gradient(
      90deg,
      transparent 0px,
      transparent 28px,
      rgba(255,255,255,0.06) 28px,
      rgba(255,255,255,0.06) 32px
    );
    background-position-y: -18px;
    background-size: 36px 18px;
    background-repeat: repeat-x;
    background-clip: border-box;
  }
  /* Sprocket holes on top and bottom borders */
  .filmstrip-rail::before {
    background: none;
    border-top: 18px solid #1a1a1a;
    border-bottom: none;
    bottom: auto;
    height: 18px;
    background-image: repeating-linear-gradient(
      90deg,
      transparent 0px, transparent 10px,
      #0a0a0a 10px, #0a0a0a 20px,
      transparent 20px, transparent 36px
    );
    background-size: 36px 10px;
    background-position: 8px 4px;
    background-repeat: repeat-x;
    border-radius: 0;
  }
  .filmstrip-rail::after {
    background: none;
    border-bottom: 18px solid #1a1a1a;
    border-top: none;
    top: auto;
    height: 18px;
    background-image: repeating-linear-gradient(
      90deg,
      transparent 0px, transparent 10px,
      #0a0a0a 10px, #0a0a0a 20px,
      transparent 20px, transparent 36px
    );
    background-size: 36px 10px;
    background-position: 8px 4px;
    background-repeat: repeat-x;
  }
  .filmstrip-cell {
    flex-shrink: 0;
    width: 22vh;
    height: 32vh;
    overflow: hidden;
    background: #000;
    border: 4px solid #000;
    outline: 1px solid #222;
    transition: opacity 0.6s ease;
  }
  .filmstrip-cell img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .filmstrip-cell.neighbors-hidden {
    opacity: 0;
  }
  /* The hero image that zooms out of the strip */
  .filmstrip-hero {
    position: fixed;
    z-index: 20;
    overflow: hidden;
    pointer-events: none;
    background: #000;
    transition: left 0.8s cubic-bezier(0.4, 0, 0.2, 1),
                top 0.8s cubic-bezier(0.4, 0, 0.2, 1),
                width 0.8s cubic-bezier(0.4, 0, 0.2, 1),
                height 0.8s cubic-bezier(0.4, 0, 0.2, 1),
                border 0.4s ease;
  }
  .filmstrip-hero img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .slideshow-progress {
    position: absolute;
    bottom: 0;
    left: 0;
    height: 3px;
    background: var(--accent);
    transition: width 0.3s linear;
    z-index: 10;
  }
  .slideshow-close {
    position: absolute;
    top: 20px;
    right: 20px;
    z-index: 10;
    width: 44px;
    height: 44px;
    border: none;
    background: rgba(255,255,255,0.1);
    color: white;
    font-size: 24px;
    border-radius: 50%;
    cursor: pointer;
    transition: background 0.2s;
  }
  .slideshow-close:hover { background: rgba(255,255,255,0.2); }
  .slideshow-pause {
    position: absolute;
    top: 20px;
    right: 76px;
    z-index: 10;
    width: 44px;
    height: 44px;
    border: none;
    background: rgba(255,255,255,0.1);
    color: white;
    font-size: 18px;
    border-radius: 50%;
    cursor: pointer;
    transition: background 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .slideshow-pause:hover { background: rgba(255,255,255,0.2); }
</style>
</head>
<body>
  <a class="back-link" href="/">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
    All Galleries
  </a>

  <div class="gallery-header">
    {{if .Config.LogoURL}}<img class="gallery-logo" src="{{.Config.LogoURL}}" alt="Logo">{{end}}
    <h1 class="gallery-title">{{.Config.GalleryTitle}}</h1>
    {{if .Config.Subtitle}}<p class="gallery-subtitle">{{.Config.Subtitle}}</p>{{end}}
    <p class="gallery-count">{{len .Photos}} PHOTOS</p>
    <div class="gallery-actions">
      <a class="gallery-btn" href="/download?slug={{.Config.Slug}}{{if .Config.IsPrivate}}&token={{.Config.SecretToken}}{{end}}">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
        Download
      </a>
      {{if .HasSongs}}
      <a class="gallery-btn" href="?slideshow=1" onclick="event.preventDefault(); history.replaceState(null,'','?slideshow=1'); startSlideshow();">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><polygon points="5 3 19 12 5 21 5 3"/></svg>
        Slideshow
      </a>
      {{end}}
    </div>
  </div>

  <div class="gallery-wrap">
    <div class="gallery-{{.Config.Layout}} hover-{{.Config.HoverEffect}} frame-{{.Config.FrameStyle}}" id="gallery">
      {{range $i, $p := .Photos}}
      <div class="photo-card" data-index="{{$i}}" style="animation-delay: {{delay $i}}s" onclick="openLightbox({{$i}})">
        <img data-src="/proxy/image?url={{urlencode $p.MediumURL}}" alt="{{$p.DisplayName}}" loading="lazy">
        <div class="photo-overlay">
          <span class="photo-name">{{if $.Config.ShowFilenames}}{{$p.DisplayName}}{{end}}</span>
          {{if $p.Rating}}<span class="photo-rating">{{stars $p.Rating}}</span>{{end}}
        </div>
      </div>
      {{end}}
    </div>
  </div>

  {{if .Config.FooterText}}
  <div class="gallery-footer">{{.Config.FooterText}}</div>
  {{end}}

  <!-- Lightbox -->
  <div class="lightbox lb-{{.Config.FrameStyle}}" id="lightbox" onclick="closeLightbox(event)">
    <button class="lightbox-close" onclick="closeLightbox(event)">&times;</button>
    <button class="lightbox-nav lightbox-prev" onclick="event.stopPropagation(); navigateLightbox(-1)">&#8249;</button>
    <div class="lightbox-frame" onclick="event.stopPropagation()">
      <img class="lightbox-img" id="lightbox-img" src="" alt="">
      <span class="lightbox-frame-name" id="lightbox-frame-name"></span>
      <span class="lightbox-frame-rating" id="lightbox-frame-rating"></span>
    </div>
    <button class="lightbox-nav lightbox-next" onclick="event.stopPropagation(); navigateLightbox(1)">&#8250;</button>
    <div class="lightbox-info">
      <div id="lightbox-name"></div>
      <div class="lightbox-counter" id="lightbox-counter"></div>
    </div>
  </div>

<script>
  // Lazy loading with IntersectionObserver
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target;
        img.src = img.dataset.src;
        img.removeAttribute('data-src');
        observer.unobserve(img);
      }
    });
  }, { rootMargin: '200px' });

  document.querySelectorAll('img[data-src]').forEach(img => observer.observe(img));

  // Lightbox
  const photos = [
    {{range .Photos}}
    {name: "{{.DisplayName}}", url: "/proxy/image?url={{urlencode .MediumURL}}", rating: {{.Rating}}},
    {{end}}
  ];
  let currentIndex = 0;

  function starsStr(n) { let s=''; for(let i=0;i<n;i++) s+='★'; return s; }

  function openLightbox(index) {
    currentIndex = index;
    const lb = document.getElementById('lightbox');
    const img = document.getElementById('lightbox-img');
    const name = document.getElementById('lightbox-name');
    const counter = document.getElementById('lightbox-counter');
    const frameName = document.getElementById('lightbox-frame-name');
    const frameRating = document.getElementById('lightbox-frame-rating');
    img.src = photos[index].url;
    name.textContent = '';
    counter.textContent = (index + 1) + ' / ' + photos.length;
    frameName.textContent = photos[index].name;
    frameRating.textContent = photos[index].rating ? starsStr(photos[index].rating) : '';
    lb.classList.add('active');
    document.body.style.overflow = 'hidden';
  }

  function closeLightbox(e) {
    if (e.target.classList.contains('lightbox') || e.target.classList.contains('lightbox-close')) {
      document.getElementById('lightbox').classList.remove('active');
      document.body.style.overflow = '';
    }
  }

  function navigateLightbox(dir) {
    currentIndex = (currentIndex + dir + photos.length) % photos.length;
    openLightbox(currentIndex);
  }

  document.addEventListener('keydown', (e) => {
    const lb = document.getElementById('lightbox');
    if (!lb.classList.contains('active')) return;
    if (e.key === 'Escape') { lb.classList.remove('active'); document.body.style.overflow = ''; }
    if (e.key === 'ArrowLeft') navigateLightbox(-1);
    if (e.key === 'ArrowRight') navigateLightbox(1);
  });

  // Touch swipe for lightbox
  (function() {
    const lb = document.getElementById('lightbox');
    let startX = 0, startY = 0, tracking = false;
    lb.addEventListener('touchstart', (e) => {
      startX = e.touches[0].clientX;
      startY = e.touches[0].clientY;
      tracking = true;
    }, {passive: true});
    lb.addEventListener('touchend', (e) => {
      if (!tracking) return;
      tracking = false;
      const dx = e.changedTouches[0].clientX - startX;
      const dy = e.changedTouches[0].clientY - startY;
      if (Math.abs(dx) < 50 || Math.abs(dy) > Math.abs(dx)) return; // too short or vertical
      if (dx < 0) navigateLightbox(1);  // swipe left = next
      else navigateLightbox(-1);        // swipe right = prev
    }, {passive: true});
  })();

  // --- Slideshow ---
  {{if .HasSongs}}
  const slideshowSongs = {{.SongsJSON}};
  const totalMusicDuration = {{.TotalDuration}};
  let slideshowTimer = null;
  let slideshowAudio = null;
  let slideshowIndex = 0;
  let slideshowSongIndex = 0;
  let slideshowRunning = false;

  function startSlideshow() {
    const ss = document.getElementById('slideshow');

    // Special slideshow modes
    if (ss.dataset.transition === 'filmstrip') { startFilmstripSlideshow(); return; }
    if (ss.dataset.transition === 'stack') { startStackSlideshow(); return; }

    const wrap = document.getElementById('ss-image-wrap');
    const progress = document.getElementById('ss-progress');
    const frameName = document.getElementById('ss-frame-name');

    const preloadCount = Math.min(3, photos.length);
    for (let i = 0; i < preloadCount; i++) {
      const img = new Image();
      img.src = photos[i].url;
    }

    slideshowIndex = 0;
    slideshowSongIndex = 0;
    slideshowRunning = true;
    ss.classList.add('active');
    document.body.style.overflow = 'hidden';

    const timePerSlide = (totalMusicDuration / photos.length) * 1000;

    showSlide(0, wrap, progress, frameName, timePerSlide);
    playSong(0);

    function advanceSlide() {
      if (!slideshowRunning) return;
      slideshowIndex++;
      if (slideshowIndex >= photos.length) {
        stopSlideshow();
        return;
      }
      if (slideshowIndex + 2 < photos.length) {
        const img = new Image();
        img.src = photos[slideshowIndex + 2].url;
      }
      showSlide(slideshowIndex, wrap, progress, frameName, timePerSlide);
      pausableTimeout(advanceSlide, timePerSlide);
    }
    pausableTimeout(advanceSlide, timePerSlide);
  }

  function playSong(index) {
    if (index >= slideshowSongs.length || !slideshowRunning) return;
    slideshowSongIndex = index;
    slideshowAudio = new Audio(slideshowSongs[index].url);
    slideshowAudio.volume = 1;
    slideshowAudio.play().catch(() => {});
    // When this song ends, play the next one
    slideshowAudio.addEventListener('ended', () => {
      if (slideshowRunning) playSong(index + 1);
    });
  }

  const allTransitions = ['fade','slide','flip','zoom','drop','blur']; // filmstrip excluded from random (it's a full-mode effect)

  function pickTransition() {
    const ss = document.getElementById('slideshow');
    const configured = ss.dataset.transition;
    if (configured !== 'random') return configured;
    return allTransitions[Math.floor(Math.random() * allTransitions.length)];
  }

  function setTransitionClass(cls) {
    const ss = document.getElementById('slideshow');
    allTransitions.forEach(t => ss.classList.remove('ss-trans-' + t));
    ss.classList.add('ss-trans-' + cls);
  }

  function showSlide(index, wrap, progress, frameName, duration) {
    const frameStyle = '{{.Config.FrameStyle}}';
    const hasFrame = frameStyle !== 'none';
    const trans = pickTransition();
    setTransitionClass(trans);

    // Transition durations vary
    const removeDuration = trans === 'fade' || trans === 'blur' ? 1500 : 1000;

    if (hasFrame) {
      let frame = wrap.querySelector('.slideshow-frame');
      if (!frame) {
        wrap.innerHTML = '';
        frame = document.createElement('div');
        frame.className = 'slideshow-frame';
        const img = document.createElement('img');
        img.className = 'slideshow-img active';
        frame.appendChild(img);
        if (frameStyle === 'polaroid') {
          const nameEl = document.createElement('span');
          nameEl.className = 'slideshow-frame-name';
          frame.appendChild(nameEl);
        }
        wrap.appendChild(frame);
      }
      const img = frame.querySelector('.slideshow-img');
      img.classList.remove('active');
      img.classList.add('exit');
      setTimeout(() => {
        img.classList.remove('exit');
        img.src = photos[index].url;
        img.onload = () => {
          requestAnimationFrame(() => img.classList.add('active'));
        };
        const nameInner = frame.querySelector('.slideshow-frame-name');
        if (nameInner) nameInner.textContent = photos[index].name;
      }, removeDuration / 2);
    } else {
      // Mark existing images for exit
      const existing = wrap.querySelectorAll('.slideshow-img');
      existing.forEach(img => {
        img.classList.remove('active');
        img.classList.add('exit');
        setTimeout(() => img.remove(), removeDuration);
      });
      const img = document.createElement('img');
      img.className = 'slideshow-img';
      img.src = photos[index].url;
      wrap.appendChild(img);
      requestAnimationFrame(() => {
        requestAnimationFrame(() => img.classList.add('active'));
      });
    }

    if (frameName) frameName.textContent = photos[index].name;

    const pct = ((index + 1) / photos.length) * 100;
    progress.style.width = pct + '%';
  }

  let slideshowPaused = false;
  let slideshowPendingTimeout = null;

  function stopSlideshow() {
    slideshowRunning = false;
    slideshowPaused = false;
    if (slideshowTimer) { clearInterval(slideshowTimer); slideshowTimer = null; }
    if (slideshowPendingTimeout) { clearTimeout(slideshowPendingTimeout); slideshowPendingTimeout = null; }
    if (slideshowAudio) { slideshowAudio.pause(); slideshowAudio = null; }
    const ss = document.getElementById('slideshow');
    ss.classList.remove('active');
    document.body.style.overflow = '';
    const wrap = document.getElementById('ss-image-wrap');
    if (wrap) wrap.innerHTML = '';
    const fs = document.getElementById('filmstrip-container');
    if (fs) fs.remove();
    const sc = document.getElementById('stack-container');
    if (sc) sc.remove();
    history.replaceState(null, '', window.location.pathname);
  }

  function togglePause() {
    const btn = document.getElementById('ss-pause');
    if (slideshowPaused) {
      // Resume
      slideshowPaused = false;
      btn.innerHTML = '&#9646;&#9646;';
      btn.title = 'Pause';
      if (slideshowAudio) slideshowAudio.play().catch(() => {});
    } else {
      // Pause
      slideshowPaused = true;
      btn.innerHTML = '&#9654;';
      btn.title = 'Resume';
      if (slideshowAudio) slideshowAudio.pause();
    }
  }

  // Wrap setTimeout to make it pausable
  function pausableTimeout(fn, delay) {
    let remaining = delay;
    let startTime = Date.now();
    let id = null;

    function tick() {
      if (slideshowPaused) {
        remaining -= (Date.now() - startTime);
        id = requestAnimationFrame(tick);
        return;
      }
      startTime = Date.now();
      if (remaining <= 0) {
        fn();
        return;
      }
      id = setTimeout(() => fn(), remaining);
      slideshowPendingTimeout = id;
    }

    // Simple approach: just check pause state periodically
    const checkPause = () => {
      if (!slideshowRunning) return;
      if (slideshowPaused) {
        setTimeout(checkPause, 100);
        return;
      }
      remaining -= (Date.now() - startTime);
      if (remaining <= 0) { fn(); return; }
      startTime = Date.now();
      slideshowPendingTimeout = setTimeout(fn, remaining);
    };
    startTime = Date.now();
    slideshowPendingTimeout = setTimeout(() => {
      fn();
    }, delay);

    // Override: monitor pause state
    const monitor = setInterval(() => {
      if (!slideshowRunning) { clearInterval(monitor); return; }
      if (slideshowPaused && slideshowPendingTimeout) {
        clearTimeout(slideshowPendingTimeout);
        slideshowPendingTimeout = null;
        remaining -= (Date.now() - startTime);
      } else if (!slideshowPaused && !slideshowPendingTimeout && remaining > 0) {
        startTime = Date.now();
        slideshowPendingTimeout = setTimeout(() => { clearInterval(monitor); fn(); }, remaining);
      }
    }, 100);
  }

  // --- Card Stack Slideshow ---
  function startStackSlideshow() {
    const ss = document.getElementById('slideshow');
    const progress = document.getElementById('ss-progress');
    slideshowIndex = 0;
    slideshowRunning = true;
    ss.classList.add('active');
    document.body.style.overflow = 'hidden';

    const timePerSlide = (totalMusicDuration / photos.length) * 1000;

    // Build stack DOM
    const container = document.createElement('div');
    container.className = 'slideshow-stack';
    container.id = 'stack-container';
    const scene = document.createElement('div');
    scene.className = 'stack-scene';
    scene.id = 'stack-scene';

    // Build the full stack (all photos, only top ones visible via CSS)
    const VISIBLE = 5;
    for (let i = 0; i < Math.min(VISIBLE + 2, photos.length); i++) {
      const card = createStackCard(i);
      if (i === 0) {
        card.classList.add('top');
      } else {
        card.classList.add('depth-' + i);
      }
      scene.appendChild(card);
    }

    container.appendChild(scene);
    ss.querySelector('.slideshow-image-wrap').appendChild(container);

    playSong(0);
    progress.style.width = ((1 / photos.length) * 100) + '%';

    // Start cycling
    pausableTimeout(() => stackAdvance(scene, progress, timePerSlide, 0), timePerSlide);
  }

  function createStackCard(photoIndex) {
    const card = document.createElement('div');
    card.className = 'stack-card';
    card.dataset.photoIndex = photoIndex;
    const img = document.createElement('img');
    img.src = photos[photoIndex].url;
    card.appendChild(img);
    const name = document.createElement('span');
    name.className = 'stack-card-name';
    name.textContent = photos[photoIndex].name;
    card.appendChild(name);
    return card;
  }

  function stackAdvance(scene, progress, timePerSlide, currentTop) {
    if (!slideshowRunning) return;
    const nextTop = currentTop + 1;
    if (nextTop >= photos.length) { stopSlideshow(); return; }

    slideshowIndex = nextTop;
    progress.style.width = (((nextTop + 1) / photos.length) * 100) + '%';

    // Peel the top card off in one arc
    const topCard = scene.querySelector('.stack-card.top');
    if (topCard) {
      topCard.classList.remove('top');
      topCard.classList.add('peeling');

      // After animation fully completes, move DOM and reassign
      topCard.addEventListener('animationend', function onEnd() {
        topCard.removeEventListener('animationend', onEnd);
        if (!slideshowRunning) return;
        // Remove animation class BEFORE moving DOM to prevent re-trigger
        topCard.classList.remove('peeling');
        // Set final resting position manually
        topCard.className = 'stack-card depth-4';
        // Now safe to move in DOM
        scene.appendChild(topCard);
        // Reassign all positions
        const allCards = scene.querySelectorAll('.stack-card');
        allCards.forEach((card, i) => {
          card.className = 'stack-card';
          if (i === 0) card.classList.add('top');
          else if (i <= 4) card.classList.add('depth-' + i);
          else card.classList.add('depth-4');
        });
      }, {once: true});
    }

    // Schedule next advance
    pausableTimeout(() => stackAdvance(scene, progress, timePerSlide, nextTop), timePerSlide);
  }

  // --- Filmstrip Slideshow ---
  function startFilmstripSlideshow() {
    const ss = document.getElementById('slideshow');
    const progress = document.getElementById('ss-progress');
    slideshowIndex = 0;
    slideshowRunning = true;
    ss.classList.add('active');
    document.body.style.overflow = 'hidden';

    const timePerSlide = (totalMusicDuration / photos.length) * 1000;

    // Build filmstrip DOM
    const container = document.createElement('div');
    container.className = 'slideshow-filmstrip';
    container.id = 'filmstrip-container';
    const rail = document.createElement('div');
    rail.className = 'filmstrip-rail';
    rail.id = 'filmstrip-rail';

    photos.forEach((p, i) => {
      const cell = document.createElement('div');
      cell.className = 'filmstrip-cell';
      cell.dataset.index = i;
      const img = document.createElement('img');
      img.src = p.url;
      cell.appendChild(img);
      rail.appendChild(cell);
    });

    container.appendChild(rail);
    ss.querySelector('.slideshow-image-wrap').appendChild(container);

    // Hero element for zoom
    const hero = document.createElement('div');
    hero.className = 'filmstrip-hero';
    hero.id = 'filmstrip-hero';
    hero.style.display = 'none';
    hero.innerHTML = '<img src="">';
    container.appendChild(hero);

    // Start music
    playSong(0);

    // Wait for layout to stabilize, then measure and start
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        const cellW = rail.children[0].offsetWidth;
        const screenW = window.innerWidth;

        // Start off-screen right
        rail.style.transition = 'none';
        rail.style.transform = 'translateX(' + screenW + 'px)';

        // Force reflow
        rail.offsetHeight;
        rail.style.transition = 'transform 600ms cubic-bezier(0.4, 0, 0.2, 1), opacity 0.5s ease';

        // Scroll to first photo
        setTimeout(() => {
          filmstripGoTo(0, cellW, rail, hero, progress, timePerSlide);
        }, 50);
      });
    });
  }

  function filmstripCenterOffset(index, cellW) {
    // Center cell[index] on screen. No CSS padding — pure math.
    const screenW = window.innerWidth;
    return (screenW / 2) - (index * cellW) - (cellW / 2);
  }

  // Animation durations (in ms) - kept short so it works even with fast slides
  const FS_SCROLL = 600;   // strip scroll time
  const FS_ZOOM = 500;     // zoom in/out time
  const FS_SHRINK = 500;   // shrink back time
  const FS_MIN_HOLD = 500; // minimum hold time at fullscreen
  const FS_OVERHEAD = FS_SCROLL + FS_ZOOM + FS_SHRINK + 200; // total animation overhead

  function filmstripGoTo(index, cellW, rail, hero, progress, timePerSlide) {
    if (!slideshowRunning || index >= photos.length) {
      stopSlideshow();
      return;
    }
    slideshowIndex = index;

    const cells = rail.querySelectorAll('.filmstrip-cell');
    const screenW = window.innerWidth;
    const screenH = window.innerHeight;

    // Phase 1: Show strip, scroll to center target cell
    rail.style.opacity = '1';
    cells.forEach(c => { c.style.opacity = '1'; });
    hero.style.display = 'none';

    rail.style.transition = 'transform ' + FS_SCROLL + 'ms cubic-bezier(0.4, 0, 0.2, 1), opacity 0.5s ease';
    const offset = filmstripCenterOffset(index, cellW);
    rail.style.transform = 'translateX(' + offset + 'px)';

    // Phase 2: After scroll completes, zoom the center cell
    setTimeout(() => {
      if (!slideshowRunning) return;
      const cell = cells[index];
      const cellRect = cell.getBoundingClientRect();

      // Place hero exactly over the cell
      hero.style.transition = 'none';
      hero.style.display = 'block';
      hero.style.left = cellRect.left + 'px';
      hero.style.top = cellRect.top + 'px';
      hero.style.width = cellRect.width + 'px';
      hero.style.height = cellRect.height + 'px';
      hero.style.border = '4px solid #000';
      hero.style.padding = '0px';
      hero.style.background = '#000';
      hero.style.boxShadow = 'none';
      hero.querySelector('img').src = photos[index].url;

      // Hero covers the cell, then we fade the strip and zoom hero
      hero.offsetHeight;
      hero.style.transition = 'all ' + FS_ZOOM + 'ms cubic-bezier(0.4,0,0.2,1)';

      const imgEl = cell.querySelector('img');
      const natW = imgEl.naturalWidth || cellRect.width;
      const natH = imgEl.naturalHeight || cellRect.height;
      const imgAspect = natW / natH;
      const screenAspect = screenW / screenH;
      const fillPct = screenW < 768 ? 0.96 : 0.88; // bigger on mobile
      let targetW, targetH;
      if (imgAspect > screenAspect) {
        targetW = screenW * fillPct;
        targetH = targetW / imgAspect;
      } else {
        targetH = screenH * fillPct;
        targetW = targetH * imgAspect;
      }

      // Add white frame padding to the hero for the zoomed state
      const framePad = screenW < 768 ? 6 : 14;
      hero.style.left = ((screenW - targetW) / 2 - framePad) + 'px';
      hero.style.top = ((screenH - targetH) / 2 - framePad) + 'px';
      hero.style.width = (targetW + framePad * 2) + 'px';
      hero.style.height = (targetH + framePad * 2) + 'px';
      hero.style.padding = framePad + 'px';
      hero.style.background = '#f5f2ed';
      hero.style.border = 'none';
      hero.style.boxShadow = '0 20px 60px rgba(0,0,0,0.5)';

      // Fade out the film strip smoothly as the hero zooms
      cell.style.opacity = '0';
      rail.style.opacity = '0.2';

      progress.style.width = (((index + 1) / photos.length) * 100) + '%';

      // Phase 3: Hold, then shrink back and advance (pausable)
      const holdTime = Math.max(timePerSlide - FS_OVERHEAD, FS_MIN_HOLD);
      pausableTimeout(() => {
        if (!slideshowRunning) return;
        const nextIndex = index + 1;
        if (nextIndex >= photos.length) { stopSlideshow(); return; }

        // Fade strip back in
        rail.style.opacity = '1';
        cell.style.opacity = '1';

        // Shrink hero back to cell, remove frame styling
        hero.style.transition = 'all ' + FS_SHRINK + 'ms cubic-bezier(0.4,0,0.2,1)';
        const currentRect = cell.getBoundingClientRect();
        hero.style.left = currentRect.left + 'px';
        hero.style.top = currentRect.top + 'px';
        hero.style.width = currentRect.width + 'px';
        hero.style.height = currentRect.height + 'px';
        hero.style.padding = '0px';
        hero.style.background = '#000';
        hero.style.border = '4px solid #000';
        hero.style.boxShadow = 'none';

        // After shrink, scroll to next
        setTimeout(() => {
          hero.style.display = 'none';
          filmstripGoTo(nextIndex, cellW, rail, hero, progress, timePerSlide);
        }, FS_SHRINK + 100);
      }, holdTime);
    }, FS_SCROLL + 100);
  }

  // Keyboard: space to pause slideshow, escape to stop
  document.addEventListener('keydown', (e) => {
    if (!slideshowRunning) return;
    if (e.key === ' ') { e.preventDefault(); togglePause(); }
    if (e.key === 'Escape') { e.preventDefault(); stopSlideshow(); }
  });

  // Auto-start slideshow if ?slideshow=1 in URL
  if (new URLSearchParams(window.location.search).get('slideshow') === '1') {
    window.addEventListener('load', () => setTimeout(startSlideshow, 500));
  }
  {{end}}
</script>

{{if .HasSongs}}
<!-- Slideshow overlay -->
<div class="slideshow ss-{{.Config.FrameStyle}} ss-trans-{{.Config.SlideshowTransition}}" id="slideshow" data-transition="{{.Config.SlideshowTransition}}">
  <button class="slideshow-pause" id="ss-pause" onclick="togglePause()" title="Pause/Resume">&#9646;&#9646;</button>
  <button class="slideshow-close" onclick="stopSlideshow()">&times;</button>
  <div class="slideshow-image-wrap" id="ss-image-wrap"></div>
  <div class="slideshow-progress" id="ss-progress"></div>
</div>
{{end}}

</body>
</html>`

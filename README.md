# 📸 Photo Gallery

A self-contained, elegant Go web application designed to showcase photo galleries with seamless integration for Capture One and YouTube audio backgrounds.

## ✨ Features

- **Capture One Integration**: Automatically sync galleries by connecting to the Capture One API, fetching photo variants and metadata.
- **Audio Integration**: Add atmospheric background music to your galleries by downloading audio from YouTube via `yt-dlp`.
- **Flexible Access**: 
  - **Public Galleries**: Accessible via a clean, slug-based URL.
  - **Private Galleries**: Secured with unique secret tokens for sharing with specific clients or friends.
- **Highly Customizable**: Fine-tune the appearance of every gallery, including:
  - Background and card colors, accent colors, and text styles.
  - Layout options (e.g., Masonry), column gaps, and max columns.
  - Frame styles, border radius, shadows, and hover effects.
- **Integrated Admin Panel**: A protected area to manage site configuration, synchronize galleries, and manage media.
- **Gallery Downloads**: Allow visitors to download entire galleries as ZIP files.
- **Performance**: Images are efficiently proxied and cached to ensure a smooth viewing experience.

## 🚀 Getting Started

### Prerequisites

- **Go**: Installed on your system.
- **yt-dlp**: Required for downloading audio from YouTube.
- **ffprobe**: Used as a fallback for determining audio duration.

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/schneipp/gallery.git
   cd gallery
   ```

2. Build the application:
   ```bash
   go build -o gallery
   ```

3. Run the application:
   ```bash
   ./gallery
   ```

The application will start on `http://localhost:8082`.

## 🛠 Administration

The admin panel is available at `/admin`.

- **Default Credentials**: 
  - Username: `admin`
  - Password: `admin`
- **Important**: Change your admin password immediately after the first login via the admin settings.

## 📂 Project Structure

- `main.go`: The core application logic, routing, and data management.
- `gallery_data.json`: Flat-file JSON database storing all site and gallery configurations.
- `media/`: Local directory where downloaded audio files are stored.
- `tmpl_*.go`: Embedded HTML templates for different views.

## 📜 License

This project is licensed under the terms specified in the `LICENSE` file.

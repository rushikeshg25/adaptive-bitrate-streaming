# Adaptive Bitrate Streaming (ABR) Platform

A high-performance video streaming platform built with **React (Vite)**, **Go (Fiber)**, and **FFmpeg**. This system implements a full end-to-end Adaptive Bitrate Streaming (ABR) pipeline, including chunked resumable uploads, background transcoding into HLS, and intelligent playback with `hls.js`.

## 🚀 Key Features

### 1. Robust Chunked Uploads
Large video files (tested up to 500MB+) are split into **5MB chunks** in the browser.
- **Sequential Uploads**: Chunks are sent one by one to ensure reliability.
- **Real-time Feedback**: A beautiful progress bar tracks the upload status.
- **Resilient**: Small chunks prevent server-side body limit timeouts and handle network jitters better.

### 2. Multi-Quality HLS Transcoding (ABR)
Once the upload is complete, the Go server merges the chunks and triggers a background **FFmpeg** pipeline to generate an HLS (HTTP Live Streaming) library:
- **720p**: High quality (2800kbps)
- **480p**: Standard quality (1400kbps)
- **360p**: Low quality (800kbps)
- **Master Playlist**: A `master.m3u8` entry point that allows the player to switch between qualities automatically.

### 3. Intelligent Video Player
Built with `hls.js`, the custom video player provides a premium experience:
- **Adaptive Bitrate Switching**: Automatically shifts to lower or higher quality based on the user's internet speed.
- **Manual Quality Selection**: A "Streaming Quality" settings menu to lock in a specific resolution.
- **Processing Awareness**: The UI polls the server for transcoding status and launches the player as soon as the video is ready.

## 🏗️ Technical Architecture

### Backend (Go + Fiber)
Located in `/server`, the backend is modularized for scalability:
- `main.go`: Entry point and route definitions.
- `handlers/upload.go`: Management of multipart chunks, reassembly, and status polling.
- `utils/video.go`: Integration with the system's FFmpeg binary for transcoding.
- `utils/metadata.go`: Persistent JSON-based tracking of video processing states.

### Frontend (React + Vite)
Located in `/web`, using modern web technologies:
- **Tailwind CSS**: For high-end, premium UI aesthetics.
- **Lucide React**: For sleek, consistent iconography.
- **Axios**: For granular upload progress tracking and API communication.
- **HLS.js**: For standard-compliant HLS playback.

## 🛠️ Getting Started

### Prerequisites
- **FFmpeg**: Must be installed and available in your system's PATH.
- **Go**: Version 1.20+ recommended.
- **Node.js**: Version 18+ recommended.

### Installation & Run
We've included a `Makefile` to simplify your workflow:

```bash
# Install all dependencies (Go & NPM)
make install

# Start both backend and frontend simultaneously
make dev
```

The app will be available at `http://localhost:5173` (Frontend) and `http://localhost:3000` (Backend).

## 📊 Transcoding Example
The system uses the following FFmpeg configuration to generate the ABR manifest:

```bash
ffmpeg -i input.mp4 \
  -map 0:v:0 -map 0:a:0 -map 0:v:0 -map 0:a:0 -map 0:v:0 -map 0:a:0 \
  -c:v libx264 -crf 22 -c:a aac -ar 48000 \
  -filter:v:0 scale=w=640:h=360 -maxrate:v:0 800k -bufsize:v:0 1200k \
  -filter:v:1 scale=w=854:h=480 -maxrate:v:1 1400k -bufsize:v:1 2100k \
  -filter:v:2 scale=w=1280:h=720 -maxrate:v:2 2800k -bufsize:v:2 4200k \
  -var_stream_map "v:0,a:0 v:1,a:1 v:2,a:2" \
  -master_pl_name master.m3u8 \
  -f hls -hls_time 10 -hls_playlist_type vod \
  -hls_segment_filename "v%v/segment%03d.ts" "v%v/index.m3u8"
```

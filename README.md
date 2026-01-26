# Adaptive Bitrate Streaming (ABR) Platform Technical Documentation

This document provides a technical overview and implementation guide for the Adaptive Bitrate Streaming platform. The system is designed to handle large video uploads, transcode them into multiple bitrate variants using HLS, and serve them via an adaptive web player.

## System Architecture

The platform is divided into a React frontend (Vite) and a Go backend (Fiber). It leverages FFmpeg for server-side video processing and HLS.js for client-side adaptive streaming.

### 1. File Upload Pipeline (Chunked Uploads)

To handle large video files (e.g., 30MB or larger) without hitting server timeouts or memory limits, the platform uses a chunked upload protocol.

- **Client-Side Slicing**: The browser uses the `File.slice()` API to split the video file into 5MB segments (Blobs).
- **Sequential Upload**: Segments are uploaded sequentially to the `/api/upload/chunk` endpoint. This ensures that the server can store them in order and simplifies error handling and retry logic.
- **Progress Tracking**: The frontend uses Axios's `onUploadProgress` hook to track the percentage of the current chunk's upload, calculating the overall percentage based on the number of completed chunks.
- **Body Limit**: The Go Fiber server is configured with a `BodyLimit` of 100MB to comfortably handle these 5MB chunks (the default limit is 4MB).

### 2. File Reassembly and Transcoding

Once all chunks are received, the client calls the `/api/upload/complete` endpoint.

- **Reassembly**: The server reads the chunks from the temporary storage directory and writes them into a single `.mp4` file.
- **Background Processing**: Transcoding is a CPU-intensive task. To avoid blocking the HTTP response, the server triggers the FFmpeg pipeline in a background goroutine.
- **Metadata Persistence**: A record is created in `server/videos/metadata.json` with the status set to `Processing`. This allows the frontend to poll for completion.

### 3. FFmpeg HLS Transcoding Engine

The server invokes FFmpeg with a multi-variant HLS configuration. The command generates three distinct quality levels:

- **720p (High)**: 1280x720, 2800kbps bitrate cap.
- **480p (Standard)**: 854x480, 1400kbps bitrate cap.
- **360p (Low)**: 640x360, 800kbps bitrate cap.

**FFmpeg Command Breakdown**:
- `-map 0:v:0 -map 0:a:0`: Maps the input video and audio streams for each output variant.
- `-c:v libx264`: Encodes video using the H.264 codec.
- `-var_stream_map`: Links specific video and audio streams to their respective HLS variant.
- `-master_pl_name master.m3u8`: Generates a master playlist that acts as the entry point for the player.
- `-f hls`: Specifies the HLS output format.

### 4. Adaptive Playback (Client Side)

The video player uses the `hls.js` library to perform bitrate switching.

- **Master Playlist Consumption**: The player loads `master.m3u8`, which contains metadata about the available bitrates and resolutions.
- **Automatic Bandwidth Detection**: HLS.js monitors the download speed of video segments. If the bandwidth drops, it automatically switches to a lower quality manifest (e.g., 360p) to prevent buffering.
- **Manual Override**: The UI includes a settings menu that allows users to manually set the quality level. Setting the `currentLevel` to `-1` in HLS.js reverts to automatic adaptive switching.
- **Status Polling**: Before playback begins, the frontend polls the `/api/videos` endpoint. It only initializes the player when the video status is marked as `Completed`.

## Development and Deployment

### Directory Structure
- `/server`: Go backend application.
- `/server/handlers`: Request handler logic.
- `/server/utils`: Core utilities for video processing and metadata.
- `/server/videos`: Final HLS output and metadata storage.
- `/server/temp_chunks`: Temporary directory for upload segments.
- `/web`: React frontend application.

### Makefile Commands
A Makefile is provided at the root for common tasks:
- `make install`: Installs Go and NPM dependencies.
- `make dev`: Launches both server and web applications in development mode.
- `make clean`: Removes all temporary chunks and uploaded video data.

## Server Configuration Notes
The server serves static files from the `/videos` directory but uses an `/api` prefix for all REST endpoints to prevent route collisions with the generated HLS directory structure.

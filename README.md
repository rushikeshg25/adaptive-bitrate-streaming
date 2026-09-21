# Adaptive bitrate streaming v1

A Go/Fiber upload server and React/HLS.js player. Requires FFmpeg and ffprobe with libx264/AAC support.

```sh
cd server && go run .
# In another terminal:
cd web && npm ci && npm run dev
```

The API listens on port 3000. The client uploads sequential 5 MiB chunks, requests completion, polls processing status and opens the adaptive master playlist. It supports automatic quality selection and explicit 360p/480p/720p selection.

## Upload protocol

POST `/api/upload/chunk` accepts multipart `uploadId`, canonical integer `index` (0..199), and `chunk` (1 byte..5 MiB). IDs contain 1..64 ASCII letters/digits/underscore/hyphen; the client uses random UUIDs. Chunk publication is atomic and an uncompleted chunk may be retried at its index.

POST `/api/upload/complete` takes `{ "uploadId": "...", "filename": "clip.mp4", "total": 3 }`. Total must be 1..200, all numbered chunks must exist, and the filename is a display label. Chunks are streamed into an isolated input file. One transcode runs at a time; busy completion returns 503 and the client retries within a bounded three-minute window. Successful completion requests leave a small receipt so retries return the same video ID. Uploaded chunks are then removed.

GET `/api/videos` returns metadata with Processing, Completed or Failed states. Metadata updates use file sync, atomic replacement and directory sync. Corrupt metadata is reported instead of overwritten. Processing jobs found at startup become Failed. SIGINT/SIGTERM shuts down admission and cancels workers. Chunk uploads and completion receipts are not a crash-safe transaction; after a process/disk failure the user may need to start a new upload.

## Output and limits

FFmpeg/ffprobe accept supported video containers with a two-minute processing deadline. Silent and audio-bearing videos are supported. Output contains three H.264 variants (640x360, 854x480, 1280x720), plus AAC when the input has audio. The server validates each finished variant and its segments before atomically publishing an explicit master playlist. This also handles FFmpeg versions that produce empty master playlists for short silent clips.

The client polls for at most three minutes, uses per-request timeouts and cancels requests on unmount. Only completed HLS playlists/segments are served; raw inputs and metadata files are excluded from the media route. Uploads are bounded to 200 chunks (1000 MiB). Original assembled inputs are deleted after processing. Abandoned chunks, completion receipts and finished videos require operator cleanup. V1 is a local, single-process demo without authentication, distributed workers or storage quotas; use trusted uploads. Scaling preserves the input display aspect ratio through sample aspect ratio metadata.

## Verification

```sh
cd server && go test -race ./...
cd ../web && npm run build && npm run lint
```

Tests cover unsafe IDs, missing chunks, idempotent completion, exact assembly, corrupt/interrupted metadata and real FFmpeg generation of all three variants. Browser playback across every platform has not been certified.

# Adaptive Bitrate Streaming Project Guide

> Generated: 2026-09-21 from commit `ed12de7`. Source inspection only; this documentation pass did not rerun builds or tests.

## What this is

This local demo turns uploaded videos into three HLS renditions for adaptive playback ([transcoder](../../server/utils/video.go#L18)). A React browser client uploads chunks, waits for processing, and offers automatic or explicit quality selection ([upload](../../web/src/components/FileUpload.tsx#L42), [player](../../web/src/components/VideoPlayer.tsx#L99)). A single Go/Fiber process handles HTTP, filesystem metadata and one background FFmpeg job at a time ([entry point](../../server/main.go#L17), [admission](../../server/handlers/upload.go#L102)).

## The five-file tour

| # | File | Why this one | Then look at |
| --- | --- | --- | --- |
| 1 | [server/main.go](../../server/main.go#L17) | Establishes storage, recovery, routes and shutdown. | [Startup](02-flow.md#startup) |
| 2 | [web/src/components/FileUpload.tsx](../../web/src/components/FileUpload.tsx#L42) | Defines the browser side of the upload protocol. | [Upload and completion](02-flow.md#upload-and-completion) |
| 3 | [server/handlers/upload.go](../../server/handlers/upload.go#L73) | Validates, assembles, admits and tracks work. | [Contracts](01-architecture.md#boundaries-and-contracts) |
| 4 | [server/utils/video.go](../../server/utils/video.go#L18) | Produces and validates the actual adaptive media. | [Background transcoding](02-flow.md#background-transcoding) |
| 5 | [web/src/components/VideoPlayer.tsx](../../web/src/components/VideoPlayer.tsx#L18) | Connects the finished master playlist to browser playback. | [Polling and playback](02-flow.md#polling-and-playback) |

## Run it

From the repository root, use separate terminals:

```sh
cd server
go mod download
go run .
```

```sh
cd web
npm ci
npm run dev
```

Prerequisites: Go matching the [1.24.6 module directive](../../server/go.mod#L3), Node/npm compatible with the [Vite toolchain](../../web/package.json#L35), and `ffmpeg`/`ffprobe` on PATH with H.264/libx264 and AAC support ([commands](../../server/utils/video.go#L24)). Node and FFmpeg versions are not pinned. No application environment variables are read by the inspected entry points: port 3000, working-directory-relative storage and the browser API origin are hardcoded ([server](../../server/main.go#L39), [client](../../web/src/components/FileUpload.tsx#L65)). Run the server from `server/` so it uses the intended data directories.

Verification commands, also recorded in the [root README](../../README.md#L27):

```sh
(cd server && go test -race ./...)
(cd web && npm run build && npm run lint)
```

[HISTORY.md](../../HISTORY.md#L60) records a previous passing delivery run. The real-media test [skips when FFmpeg is absent](../../server/utils/video_test.go#L13); a green test run alone therefore does not establish media generation on an unprovisioned machine.

## Reading order for this guide

1. [Architecture](01-architecture.md) — components, contracts and filesystem state.
2. [Flow](02-flow.md) — startup, upload, transcoding, playback and shutdown.
3. [Structure](03-structure.md) — source ownership and callers.
4. [Tech stack](04-tech-stack.md) — declared versions and tools.
5. [Decisions](05-decisions.md) — evidence, tradeoffs and newcomer traps.

The [v1 contract](../../V1.md#L3) defines the intended scope; [history](../../HISTORY.md#L10) records delivery chronology. This guide describes the inspected implementation rather than treating those documents as proof of runtime behavior.

## Open questions

- What production host, origin, storage quota and cleanup policy should replace the local demo assumptions? The [README](../../README.md#L25) leaves these outside v1, and [history](../../HISTORY.md#L68) records no production validation.
- Should native HLS playback expose manual quality controls? The [native fallback](../../web/src/components/VideoPlayer.tsx#L73) does not discover levels; the explicit controls operate through HLS.js.
- Should the checked-in [metadata fixture](../../server/videos/metadata.json) remain? It contains Completed records, but their media files are absent from this checkout; startup [only repairs Processing states](../../server/utils/metadata.go#L88).

# Tech Stack

## Languages and runtimes

| Language or runtime | Version | Pinned at |
| --- | --- | --- |
| Go | `1.24.6` module directive | [go.mod:3](../../server/go.mod#L3) |
| TypeScript | `~5.9.3` declared range | [package.json:33](../../web/package.json#L33) |
| Browser JavaScript | ES2022 compilation target and DOM libraries; browser versions unspecified | [tsconfig.app.json:4](../../web/tsconfig.app.json#L4) |
| Node.js / npm | Runtime/package-manager versions not pinned by package.json; Node type definitions are `^24.10.1`, which is not a Node runtime pin | [package.json:25](../../web/package.json#L25) |
| FFmpeg / ffprobe | External executables on PATH; version not pinned | [video.go:24](../../server/utils/video.go#L24), [video.go:41](../../server/utils/video.go#L41) |

## Frameworks and major libraries

Versions below are manifest declarations, not claims about a particular installed dependency tree.

| Library | Version | Used for | Manifest and usage |
| --- | --- | --- | --- |
| Fiber v2 | `v2.52.10` | HTTP routes, completed-media file responses, middleware and multipart storage. | [go.mod:6](../../server/go.mod#L6), [main.go:25](../../server/main.go#L25) |
| Google UUID | `v1.6.0` | Independent server-generated video IDs. | [go.mod:7](../../server/go.mod#L7), [upload.go:113](../../server/handlers/upload.go#L113) |
| React / React DOM | `^19.2.0` each | Hooks, component UI and StrictMode mounting. | [package.json:17](../../web/package.json#L17), [main.tsx:6](../../web/src/main.tsx#L6) |
| Axios | `^1.13.3` | Multipart/JSON HTTP requests, progress reporting and AbortSignal support. | [package.json:14](../../web/package.json#L14), [FileUpload.tsx:65](../../web/src/components/FileUpload.tsx#L65) |
| HLS.js | `^1.6.15` | HLS loading, adaptive levels and error recovery when supported. | [package.json:15](../../web/package.json#L15), [VideoPlayer.tsx:28](../../web/src/components/VideoPlayer.tsx#L28) |
| Tailwind CSS / Vite plugin | `^4.1.18` each | Utility classes and CSS build integration. | [package.json:13](../../web/package.json#L13), [package.json:19](../../web/package.json#L19), [vite.config.ts:7](../../web/vite.config.ts#L7) |
| Lucide React | `^0.563.0` | Upload, player and shell icons. | [package.json:16](../../web/package.json#L16), [App.tsx:4](../../web/src/App.tsx#L4) |
| Go standard library | Supplied with Go | File IO, JSON, contexts, subprocesses, mutexes and HTTP test fixtures. | [go.mod:3](../../server/go.mod#L3), [metadata.go:3](../../server/utils/metadata.go#L3), [video.go:3](../../server/utils/video.go#L3) |

## Data and infrastructure

| Service or facility | Role | Configured at |
| --- | --- | --- |
| Local filesystem | JSON metadata, chunk/receipt directories and HLS media; no database service. | [metadata.go:19](../../server/utils/metadata.go#L19), [upload.go:51](../../server/handlers/upload.go#L51) |
| FFmpeg encoders and HLS muxer | libx264/yuv420p H.264, optional AAC, MPEG-TS segments, VOD playlists. | [video.go:40](../../server/utils/video.go#L40) |
| Fiber HTTP listener | API and `/videos` on port 3000; wildcard CORS. | [main.go:33](../../server/main.go#L33), [main.go:53](../../server/main.go#L53) |
| Browser media APIs | Native HLS fallback, HTML video controls, UUID generation and request cancellation. | [VideoPlayer.tsx:73](../../web/src/components/VideoPlayer.tsx#L73), [FileUpload.tsx:45](../../web/src/components/FileUpload.tsx#L45) |

No container/deployment manifest or CI workflow is present in the inspected 31-file tracked tree. The checked-in execution topology is represented by the [Makefile](../../Makefile#L21); production validation is explicitly absent from [history](../../HISTORY.md#L68).

## Tooling

| Tool | Role | Configured at |
| --- | --- | --- |
| Make | Install, build and run both projects; `clean` also deletes video and chunk data. | [Makefile:10](../../Makefile#L10), [Makefile:33](../../Makefile#L33) |
| Vite `^7.2.4` | Dev server, production bundle and preview. | [package.json:7](../../web/package.json#L7), [package.json:35](../../web/package.json#L35) |
| React Vite plugin `^5.1.1` | React build/development transformation. | [package.json:28](../../web/package.json#L28), [vite.config.ts:7](../../web/vite.config.ts#L7) |
| TypeScript project build | Check browser and Vite config projects before bundling. | [package.json:8](../../web/package.json#L8), [tsconfig.json:3](../../web/tsconfig.json#L3) |
| ESLint `^9.39.1`, typescript-eslint `^8.46.4` | Static JS/TS checking with hooks `^7.0.1` and refresh `^0.4.24` plugins. | [package.json:29](../../web/package.json#L29), [package.json:34](../../web/package.json#L34), [eslint.config.js:12](../../web/eslint.config.js#L12) |
| Go testing / race detector | Behavioral handler/media/metadata checks; README requests `-race`. | [upload_test.go:19](../../server/handlers/upload_test.go#L19), [video_test.go:13](../../server/utils/video_test.go#L13), [README:30](../../README.md#L30) |

## Notes

- Node type definitions and `tsconfig.node.json`'s ES2023 target do not establish the supported Node runtime. Resolve that before changing deployment environments ([tsconfig.node.json:4](../../web/tsconfig.node.json#L4)).
- The real-media test checks FFmpeg availability but subsequently needs ffprobe as well, through the production transcoder ([video_test.go:14](../../server/utils/video_test.go#L14), [video.go:24](../../server/utils/video.go#L24)).
- Client quality selection is implemented through HLS.js, while native HLS support is a separate capability branch ([VideoPlayer.tsx:28](../../web/src/components/VideoPlayer.tsx#L28)). No browser compatibility matrix is certified ([README:34](../../README.md#L34)).

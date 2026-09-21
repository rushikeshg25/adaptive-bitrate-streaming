# Structure

## What lives where

The Go module is self-contained in `server/`, while `web/` owns a separately built React client. Most behavior resides in one HTTP handler file, two utility files and two browser components; the root Makefile only coordinates commands ([Makefile](../../Makefile#L10)). Storage follows the server's working directory ([main.go](../../server/main.go#L59)).

```text
Makefile                 install, build, run and destructive clean targets
README.md                public usage and limits
V1.md                    acceptance contract
HISTORY.md               delivery chronology and historical verification
server/
  main.go                process entry, HTTP routes, signals
  handlers/              upload lifecycle and handler tests
  utils/                 metadata, transcoding and utility tests
  videos/                metadata fixture plus runtime HLS output
  temp_chunks/           runtime uploads and receipts; created on startup
web/
  index.html             browser entry document
  src/                   React shell and styling
    components/          upload and HLS playback
  public/                static starter icon
  vite.config.ts         React and Tailwind build plugins
  package.json           scripts and declared dependencies
```

## Root

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [Makefile](../../Makefile#L1) | Installs dependencies, builds both halves, starts development processes and removes outputs/data. | `all`, `install`, `build`, `run-server`, `run-web`, `dev`, `clean` | Developer invoking make |
| [README.md](../../README.md#L1) | Public run instructions, API contract and operational limits. | Documentation | Developers/operators |
| [V1.md](../../V1.md#L1) | Scope and executable acceptance expectations. | Documentation | Delivery and review |
| [HISTORY.md](../../HISTORY.md#L1) | Evidence-linked chronology and previous verification. | Documentation | Maintainers |

## Server root

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [main.go](../../server/main.go#L17) | Ensure directories, recover metadata, configure Fiber/CORS/completed-media serving, wire routes and shutdown. | `main`, `setupDirectories` | Go process |
| [go.mod](../../server/go.mod#L1) | Module/runtime directive and server dependencies. | Module `server` | Go toolchain |

## Server handlers

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [upload.go](../../server/handlers/upload.go#L20) | Serialize upload mutation; validate/store chunks; assemble and admit work; persist receipt and worker status; serve only Completed HLS media. | `UploadChunk`, `CompleteUpload`, `ListVideos`, `HealthCheck`, `LegacyUpload`, `StopProcessing`, `ServeVideo` | Routes and signal handling in main; tests |
| [upload_test.go](../../server/handlers/upload_test.go#L19) | Reject unsafe/missing input; verify exact assembly, terminal metadata, repeat completion receipt and media access restrictions. | `TestRejectUnsafeAndMissingChunks`, `TestCompleteAssemblyAndReceipt`, `TestMediaRouteHidesInputsAndProcessingOutput` | `go test`; transcode function stub |

## Server utilities

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [metadata.go](../../server/utils/metadata.go#L11) | Metadata schema, mutex-protected JSON reads/upserts, synced atomic file replacement and interrupted-job recovery. | `VideoMetadata`, `SaveMetadata`, `GetAllMetadata`, `RecoverProcessing` | Handlers, main and tests |
| [video.go](../../server/utils/video.go#L13) | Deadline wrapper, ffprobe/FFmpeg execution, variant validation and explicit master publication. | `TranscodeToHLS`, `TranscodeToHLSContext` | Handler worker uses context variant; tests; standalone wrapper has no production caller in inspected source |
| [video_test.go](../../server/utils/video_test.go#L13) | Real silent-video HLS check plus corrupt metadata preservation and recovery check. | `TestTranscodeSilentVideo`, `TestMetadataRejectsCorruption` | `go test` |

## Server storage

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [videos/metadata.json](../../server/videos/metadata.json#L1) | Checked-in historical Completed records; runtime also mutates this path. Media for these records is absent from this checkout. | JSON metadata array | Metadata utility; not directly served over HTTP |

`temp_chunks/<uploadId>/` and `videos/<videoID>/` are runtime directories rather than separate source modules. Their layouts follow [upload.go:51](../../server/handlers/upload.go#L51) and [video.go:40](../../server/utils/video.go#L40); see [architecture state](01-architecture.md#state-and-persistence).

## Web source

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [index.html](../../web/index.html#L1) | Root DOM node and module entry script. | `#root` | Vite/browser |
| [src/main.tsx](../../web/src/main.tsx#L1) | Imports CSS and mounts App with React StrictMode. | Module side effect | HTML module script |
| [src/App.tsx](../../web/src/App.tsx#L6) | Composes player and upload panel around one current video URL. | Default `App` | `main.tsx` |
| [src/index.css](../../web/src/index.css#L1) | Tailwind import, dark color scheme, page styles and scrollbar styling. | Global CSS | `main.tsx` |

## Web components

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [FileUpload.tsx](../../web/src/components/FileUpload.tsx#L12) | Input/drop UI; 5 MiB slicing; sequential Axios upload; bounded completion retry; bounded polling; cancellation and progress/error state. | Default `FileUpload`; callback prop `onUploadSuccess` | App |
| [VideoPlayer.tsx](../../web/src/components/VideoPlayer.tsx#L10) | HLS.js instance lifecycle, native fallback, playback recovery, quality menu and video element. | Default `VideoPlayer`; `src`/`poster` props | App |

## Web tooling

| File | Responsibility | Key exports | Called by |
| --- | --- | --- | --- |
| [package.json](../../web/package.json#L6) | Development/build/lint/preview commands and dependency ranges. | npm scripts | npm and Makefile |
| [vite.config.ts](../../web/vite.config.ts#L6) | Register React and Tailwind Vite plugins; no API proxy configured. | Default Vite config | Vite |
| [eslint.config.js](../../web/eslint.config.js#L8) | Recommended JS/TS/hooks/refresh rules on TS/TSX; ignores dist. | Default flat config | ESLint |
| [tsconfig.json](../../web/tsconfig.json#L1) | Reference browser and tooling TypeScript projects. | Project references | `tsc -b` |
| [tsconfig.app.json](../../web/tsconfig.app.json#L1) | Strict browser TS/TSX checking, ES2022/DOM libs, bundler module resolution, no emit. | Browser compiler settings | TypeScript build |
| [tsconfig.node.json](../../web/tsconfig.node.json#L1) | Strict ES2023/Node checking for Vite config. | Tooling compiler settings | TypeScript build |

## Excluded

Generated binaries, runtime segments/playlists, build output and dependencies are excluded from source tables. Dependency lockfiles (`go.sum`, npm lock and Bun lock), ignore files and unused/starter SVG art do not explain runtime design; their omission is intentional. Tests change working-directory or package globals and do not use parallel execution ([handler tests](../../server/handlers/upload_test.go#L60), [utility tests](../../server/utils/video_test.go#L33)).

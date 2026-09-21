# Decisions

These choices are grounded in the inspected code. Rationale is labeled as inferred unless a comment or project document states it directly. The [v1 contract](../../V1.md#L3) prioritizes bounded input, safe assembly, cancellation, persistence and executable verification.

## Single process with one admitted transcode

- **What:** An in-process channel of capacity one controls FFmpeg admission; busy completion returns 503, without a queue. Upload writes and completion assembly also share a mutex.
- **Evidence:** [upload.go:20](../../server/handlers/upload.go#L20), [upload.go:102](../../server/handlers/upload.go#L102).
- **Why, apparently:** Inferred: bound CPU-heavy processing with minimal moving parts; the [README](../../README.md#L25) explicitly scopes deployment to a local single-process demo.
- **Tradeoff:** Predictable concurrency and a small runtime, but no distributed admission, durable queue or durable retry across page reloads; the browser retries busy completion within a bounded window.
- **Confidence:** Scope confirmed by documentation; specific implementation rationale inferred.

## Separate client upload IDs from media IDs

- **What:** The browser creates an upload UUID; completion generates a new video UUID. The submitted filename is a validated display label rather than the storage filename.
- **Evidence:** [FileUpload.tsx:51](../../web/src/components/FileUpload.tsx#L51), [upload.go:84](../../server/handlers/upload.go#L84), [upload.go:113](../../server/handlers/upload.go#L113).
- **Why, apparently:** Inferred: isolate temporary uploads and output paths from user-provided filenames, matching the isolation requirement in [V1.md:5](../../V1.md#L5).
- **Tradeoff:** Simple path ownership; callers must retain the mapping returned by completion to observe or play a video.
- **Confidence:** Mechanism confirmed by code; rationale inferred from the contract.

## Receipts make completion replayable

- **What:** A successful handoff leaves JSON in the upload directory. A later completion request replays the receipt before checking chunks or claiming processing capacity.
- **Evidence:** [upload.go:87](../../server/handlers/upload.go#L87), [upload.go:149](../../server/handlers/upload.go#L149), [receipt test](../../server/handlers/upload_test.go#L83).
- **Why, apparently:** Inferred: prevent duplicate jobs when the caller repeats an accepted completion request; the [README](../../README.md#L17) explicitly promises the same video ID on retry.
- **Tradeoff:** Straightforward retry semantics after the receipt exists, but no atomic transaction spans metadata, receipt, worker start and chunk deletion. The receipt says processing started, not that it completed.
- **Confidence:** Replay behavior documented and tested; crash limitations documented in [README:19](../../README.md#L19).

## Synced JSON replacement instead of a database

- **What:** Every metadata mutation upserts into a whole-file array, writes/syncs a temporary file, renames it and syncs the directory. Startup marks interrupted Processing records Failed.
- **Evidence:** [metadata.go:34](../../server/utils/metadata.go#L34), [metadata.go:88](../../server/utils/metadata.go#L88).
- **Why, apparently:** Inferred: keep local persistence simple while avoiding torn metadata replacement and falsely permanent Processing states.
- **Tradeoff:** No database dependency and corrupt data is preserved for diagnosis, but all operations scan the array, locks coordinate only one process, and recovery fails rather than resumes work.
- **Confidence:** Mechanism confirmed by code and [metadata tests](../../server/utils/video_test.go#L33); rationale inferred.

## Publish a validated explicit master playlist

- **What:** Encode three variants, check completed media playlists and referenced segment files, then atomically rename a handcrafted adaptive master into place.
- **Evidence:** [video.go:44](../../server/utils/video.go#L44); [HISTORY.md:62](../../HISTORY.md#L62) records the short-silent-clip failure observed with FFmpeg 8.0.1.
- **Why:** Some FFmpeg builds omitted master variants for short silent clips despite successful encoding; the source comment explicitly names this workaround.
- **Tradeoff:** Deterministic master generation independent of that FFmpeg behavior, but fixed resolutions/bandwidth declarations must stay synchronized with encoder arguments. The check verifies playlist/segment presence, not full decoding or client compatibility.
- **Confidence:** Confirmed by code comment and delivery history.

## Poll metadata and let HLS.js manage adaptation

- **What:** The browser polls the full metadata array at roughly two-second intervals within a three-minute window until it sees a terminal state, then loads the master. HLS.js provides Auto and manual levels; native HLS is a fallback.
- **Evidence:** [FileUpload.tsx:117](../../web/src/components/FileUpload.tsx#L117), [VideoPlayer.tsx:28](../../web/src/components/VideoPlayer.tsx#L28), [VideoPlayer.tsx:99](../../web/src/components/VideoPlayer.tsx#L99).
- **Why, apparently:** Inferred: use ordinary HTTP and existing browser/player capabilities instead of adding a push channel or implementing bitrate adaptation.
- **Tradeoff:** Small protocol surface, but polling transfers every video record and the UI cannot reopen older uploads. Native fallback does not get the same explicit quality controls.
- **Confidence:** Mechanism confirmed; rationale inferred.

## Gate media on terminal metadata

- **What:** The media handler only accepts known HLS filenames and serves a video after metadata says Completed.
- **Evidence:** [upload.go:182](../../server/handlers/upload.go#L182), [media access test](../../server/handlers/upload_test.go#L109).
- **Why, apparently:** Inferred: prevent raw input/metadata exposure and keep partially generated media unavailable.
- **Tradeoff:** A narrow publication boundary, but every segment request rereads metadata and a failed terminal status write hides even successfully generated output.
- **Confidence:** Behavior confirmed by code and test; rationale inferred.

## Gotchas

- **The body-limit comment is wrong:** Fiber is configured for 6 MiB, despite a `100MB` comment. Each chunk is limited to 5 MiB, leaving room for multipart overhead ([main.go:27](../../server/main.go#L27), [upload.go:48](../../server/handlers/upload.go#L48)).
- **The UI does not contain a video list:** The timeout error suggests checking one later, but App only renders the current preview and upload form ([FileUpload.tsx:126](../../web/src/components/FileUpload.tsx#L126), [App.tsx:19](../../web/src/App.tsx#L19)).
- **The committed metadata is not playable sample media:** Completed fixture records exist without corresponding media in this checkout; recovery only touches Processing records ([metadata.json](../../server/videos/metadata.json#L1), [metadata.go:93](../../server/utils/metadata.go#L93)).
- **Working directory changes storage:** Starting `server/bin/server` from the repository root stores data relative to the root. Follow the Makefile's `cd server` convention ([Makefile:22](../../Makefile#L22), [metadata.go:19](../../server/utils/metadata.go#L19)).
- **`make clean` deletes uploaded data:** It removes `server/videos/*` and `server/temp_chunks/*` as well as build output ([Makefile:33](../../Makefile#L33)).
- **Tests are not exhaustive playback certification:** Real encoding coverage uses one short silent clip; handler tests stub the transcode function. Audio, busy admission, shutdown and browser playback have no corresponding behavioral tests in the inspected test files ([video_test.go:13](../../server/utils/video_test.go#L13), [upload_test.go:60](../../server/handlers/upload_test.go#L60)).

## Conventions

- Handler failures use Fiber HTTP errors with short public messages; detailed transcoding/persistence failures go to server logs ([upload.go:95](../../server/handlers/upload.go#L95), [upload.go:164](../../server/handlers/upload.go#L164)).
- Go tests use temporary directories and restore overridden globals rather than running in parallel; maintain this isolation when extending them ([upload_test.go:60](../../server/handlers/upload_test.go#L60), [video_test.go:33](../../server/utils/video_test.go#L33)).
- React components own local state/effects; App connects them through a callback and URL prop rather than a shared store ([App.tsx:6](../../web/src/App.tsx#L6)).

# **ClipSense — Technical Overview & System Description**

## **1. Purpose of ClipSense**

ClipSense is an AI-powered video intelligence system built specifically for streamers, content creators, and editors who work with large sets of reaction videos, clip dumps, gameplay captures, or long-form session recordings.
In simple terms, ClipSense allows creators to submit video sources, then the system:

* **Understands** each clip using NLP and multimodal analysis
* **Classifies** clips by content, vibe, context, and role
* **Finds relationships** between clips
* **Arranges** them into coherent storylines
* **Helps creators export** sequences into editing tools or manual workflows

ClipSense uses a unified intake model. Public or user-provided video links, direct uploaded video files, and ZIP archives should all normalize into the same internal pipeline:

`source input -> stored source asset -> extracted video asset(s) -> metadata -> transcript -> analysis -> highlights/risk/context/storyline`

The current MVP implementation starts with ZIP upload, but ZIP is not the only intended product path.

ClipSense reduces hours of watching, sorting, guessing, and organizing down to minutes.

Its purpose is to become the intelligence layer that sits *before* editing — the creative pre-editor that handles sorting, labeling, and narrative structure.

---

## **2. Core Value**

ClipSense is built to answer a single question:

### **“What is the best order to watch or edit these clips so it tells a story?”**

Creators have tons of content but zero time to organize or structure it.

ClipSense solves that by combining:

* Audio transcription
* NLP semantic understanding
* Embeddings + similarity search
* Narrative modeling
* Human-adjustable storylines

It turns raw video chaos into structured narrative meaning.

---

# **3. High-Level Architecture**

ClipSense is built as a **distributed, multi-service system** designed for scalability, performance, and modular development.

The stack is split deliberately:

### **Frontend Web UI — Next.js (TypeScript)**

The creator interface where users:

* Upload or submit video sources: links, direct video files, or ZIP batches
* View clip library and metadata
* Explore AI analysis (roles, moods, summaries)
* Review and edit storylines
* Export sequences

### **Backend API — Go**

Handles all system operations:

* Authentication
* Batch lifecycle
* Clip metadata management
* Storyline management
* WebSocket status updates
* Job scheduling for AI worker

Go is chosen for:

* Speed
* Concurrency
* Stability
* Predictable memory usage

### **AI Processing Worker — Python**

The intelligence engine. It handles:

* Audio extraction
* Transcription
* Summarization
* Classification (role, vibe, topic)
* Embedding generation
* Cluster analysis
* Storyline generation

Python is the natural choice due to its massive ML/NLP ecosystem.

### **Rust Video Engine (Optional, Future Optimization)**

For extremely heavy video workloads, ClipSense introduces a Rust microservice for:

* High-speed video processing
* Scene boundary detection
* Frame sampling
* Audio segmentation

Rust is used only where raw performance matters.

---

# **4. Infrastructure Layer**

### **Data Storage**

1. **PostgreSQL**
   Stores structured data:

   * Batches
   * Clips
   * Analysis metadata
   * Storylines
   * Users
   * Jobs

2. **Vector DB (Qdrant or Pinecone)**
   Stores clip embeddings for:

   * Similarity search
   * Clustering
   * Storyline optimization
   * “Find similar clips” functions

3. **Redis**
   Used for:

   * Job queues
   * Caching
   * Rate limiting

4. **Object Storage (S3/GCS)**
   Stores:

   * Uploaded or linked source assets
   * Extracted audio
   * Video files
   * Exports

### **Compute Infrastructure**

* Docker containers for all services
* Optional Kubernetes deployments
* Terraform provisioning for cloud infrastructure
* NGINX gateway / load balancing
* Prometheus & Grafana for monitoring

---

# **5. The ClipSense AI Pipeline**

Each video goes through a multi-stage intelligent process.

## **Stage 1 — Ingestion**

User submits a source input → API validates → queue job → stored as a source asset.

## **Stage 2 — Video Preparation**

Python worker extracts:

* Metadata
* Audio stream
* Duration
* Thumbnails
* Optional: frame sampling via Rust

## **Stage 3 — Transcription (ASR)**

Using Whisper or ASR providers:

* Converts speech to text
* Splits into segments
* Handles accents, slang, overlapping speech

High transcription quality drives everything else.

## **Stage 4 — NLP Understanding**

The worker uses LLM/NLP models to infer:

* Summary
* Vibe (funny, intense, calm, hype)
* Topic (reaction, sports, gaming, ranting, emotional, etc.)
* Clip role (intro, setup, escalation, climax, resolution)
* Keywords

This is the ClipSense core IP — the semantic fingerprint of the clip.

## **Stage 5 — Embeddings**

Transforms each clip into a vector representing:

* Content
* Context
* Emotion
* Style

Stored in vector DB for downstream tasks.

## **Stage 6 — Cluster Analysis**

Clips are grouped by:

* Similar themes
* Similar tones
* Narrative compatibility
* Shared characters/topics

## **Stage 7 — Story Engine**

The system generates storylines using:

* Graph traversal
* Arc scoring
* Pacing analysis
* Coherence constraints
* Known storytelling patterns

This creates:

* “Best AI Sequence”
* “Chronological Sequence”
* “High-Energy First Sequence”
* “Balanced Narrative Sequence”

## **Stage 8 — Explanation**

LLM generates:

* Why this order works
* Story summary
* Mood arc
* “Director notes”

## **Stage 9 — Export**

Creator gets:

* JSON
* CSV
* Timeline previews
* EDL export (Adobe / DaVinci)

---

# **6. Usability — How Creators Use ClipSense**

### **1. Submit Source**

Submit a video link, upload a direct video file, or drag-and-drop a ZIP batch of videos.

### **2. ClipSense automatically processes**

Clips appear one by one with:

* Title
* Thumbnail
* Duration
* Summary
* Tags

The user starts seeing results instantly while the rest continues processing.

### **3. Explore Clips**

Filtering by:

* Topic
* Vibe
* Role
* Length
* Keyword

Makes large dumps manageable.

### **4. Storyline Suggestions**

The “Storyline View” shows:

* AI sequences
* Mood arcs
* Clip positions
* Transitions
* Gaps or missing links

The creator can:

* Accept AI order
* Manually adjust
* Clone & modify
* Request a new storyline
* Ask for a faster/slower pacing version

### **5. Export**

One click to export into editing tools.

---

# **7. Why ClipSense Works**

Most tools help with editing.
ClipSense helps with **thinking** before editing.

It removes:

* Hours of watching raw footage
* Guessing clip order
* Forgetting which clip contains what
* Manually building narratives
* Endless scrolling through folders

ClipSense becomes:

* The assistant
* The sorter
* The story architect
* The memory
* The pattern recognizer
* The context brain

Creators keep the creative control, but they stop doing low-level tasks.

---

# **8. Target Users**

* Streamers
* Reaction channels
* Clip channels
* Vloggers
* Gaming creators
* Podcast editors
* Agencies with large content workloads
* TikTok/Shorts editors
* Memers who need fast sorting

The more chaotic the video workflow…
The more ClipSense shines.

---

# **9. Why This Architecture is Chosen**

Each language has a role:

**TypeScript** → best for UI/UX
**Go** → best for stable, scalable API backend
**Python** → best for NLP, embeddings, ML workflows
**Rust** → best for future high-performance video transformations

This structure ensures:

* Speed
* Reliability
* Modularity
* Future growth
* Team scalability

---

# **10. ClipSense in One Sentence**

**ClipSense is an intelligent pre-editor that transforms raw piles of creator videos into organized, meaningful, story-driven sequences — automatically and beautifully.**

---

# **11. Implementation Snapshot**

The current stabilized MVP covers the web UI (Next.js/TypeScript), Go API, Python AI worker, Postgres, Redis, and Qdrant through Docker Compose. It is ready for local verification/testing, not for the next product roadmap phase.

* **Web** (`apps/web`): Next.js 14 app-router, Tailwind, current MVP drag-and-drop ZIP upload, dashboard list, batch detail with clips + basic AI storyline visibility.
* **API** (`apps/api`): Go (chi) REST service with Postgres persistence; endpoints for health, auth, batch list/create, batch detail, and batch export; uploaded source archives are stored under the shared upload volume.
* **AI Worker** (`apps/api/ai_worker`): Python worker consuming Redis batch jobs; safely extracts supported video files from ZIP archives, runs ffmpeg audio extraction, Whisper transcription, sentence-transformer embeddings, KMeans clustering, writes clips/storylines to Postgres, upserts clip vectors to Qdrant, and marks batches complete or failed.
* **Current MVP data flow**: ZIP upload (web) → `/api/batches` (Go saves source archive + inserts batch[pending]) → worker picks pending → extracts supported video assets → processes → inserts clips + storyline → marks batch complete → web dashboard reflects status via live fetch.
* **Current MVP intake limitation**: links and direct single-video uploads are product-intended future intake types, but the current implemented web/API/worker path is ZIP upload only.
* **Narrative intelligence status**: storyline generation is still basic clustering/order logic, not advanced narrative arc modeling or director-mode reasoning.
* **Active Docker-facing defaults**: `DATABASE_URL=postgres://clipsense:clipsense@postgres:5432/clipsense?sslmode=disable`, `REDIS_URL=redis://redis:6379`, `QDRANT_HOST=qdrant`, `UPLOAD_DIR=/data/uploads`, `PROCESS_DIR=/data/processing`, `API_ADDR=:8080`, `NEXT_PUBLIC_API_URL=http://localhost:8080`.
* **Prereqs**: ffmpeg available on PATH; Python deps from `apps/api/ai_worker/requirements.txt`; Go 1.21+; Node 18+ for web.

# **12. Orchestration**

Docker Compose brings up Postgres, Redis, Qdrant, API, worker, and web:

```
docker-compose up --build
```

Services:
* `postgres`: Postgres database for structured application data.
* `redis`: Redis queue for batch processing jobs.
* `qdrant`: Vector database for clip embeddings.
* `api`: Go server on `:8080`, exposed to the host as `http://localhost:8080`.
* `worker`: Python processor consuming Redis jobs and using the shared upload/processing volume.
* `web`: Next.js app on `:3000` using browser-facing `NEXT_PUBLIC_API_URL=http://localhost:8080`.

Shared volume `clipsense-data` lets the API and worker access uploaded source archives and processing files. Postgres and Qdrant each use their own named volumes.

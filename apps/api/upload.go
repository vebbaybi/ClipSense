package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const maxArchiveEntries = 200
const maxEntryBytes = 256 << 20
const maxExpandedBytes = 1 << 30

var uploadSlots = make(chan struct{}, 2)
var uploadStorageMu sync.Mutex
var videoExtensions = map[string]bool{".mp4": true, ".mov": true, ".mkv": true, ".webm": true, ".avi": true}

func uploadFailure(w http.ResponseWriter, err error) {
	var limit *http.MaxBytesError
	if errors.As(err, &limit) {
		publicError(w, 413, "upload_too_large")
		return
	}
	publicError(w, 400, "invalid_upload")
}

func removeUpload(file string) {
	if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
		log.Print("event=upload_cleanup_failed")
	}
}

func newUploadFile(root string) (*os.File, error) {
	uploadStorageMu.Lock()
	defer uploadStorageMu.Unlock()
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	// Includes retained successful/ambiguous handoffs: no unbounded disk growth.
	if len(entries) >= 32 {
		return nil, errors.New("upload storage full")
	}
	return os.CreateTemp(root, "upload-*.zip")
}

func createBatch(w http.ResponseWriter, r *http.Request) {
	select {
	case uploadSlots <- struct{}{}:
		defer func() { <-uploadSlots }()
	default:
		publicError(w, 429, "upload_busy")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, security.uploadBytes)
	if r.ContentLength > security.uploadBytes {
		publicError(w, 413, "upload_too_large")
		return
	}
	mr, err := r.MultipartReader()
	if err != nil {
		publicError(w, 400, "invalid_multipart")
		return
	}
	uploads := getenv("UPLOAD_DIR", "data/uploads")
	if err = os.MkdirAll(uploads, 0700); err != nil {
		publicError(w, 503, "upload_unavailable")
		return
	}
	// Only a server-created path is ever deleted. Multipart is streamed, never spooled.
	out, err := newUploadFile(uploads)
	if err != nil {
		publicError(w, 503, "upload_unavailable")
		return
	}
	dest := out.Name()
	transferred := false
	defer func() {
		out.Close()
		if !transferred {
			removeUpload(dest)
		}
	}()
	name := ""
	filename := ""
	files, fields := 0, 0
	for {
		part, readErr := mr.NextPart()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			uploadFailure(w, readErr)
			return
		}
		switch part.FormName() {
		case "file":
			files++
			filename = part.FileName()
			if files != 1 || filename == "" || len(filename) > 255 || strings.ContainsAny(filename, "\\\x00\r\n") || strings.ToLower(path.Ext(filename)) != ".zip" {
				publicError(w, 400, "invalid_upload_file")
				return
			}
			n, copyErr := io.Copy(out, part)
			if copyErr != nil {
				uploadFailure(w, copyErr)
				return
			}
			if n == 0 {
				publicError(w, 400, "empty_upload")
				return
			}
		case "name":
			fields++
			if fields > 1 || part.FileName() != "" {
				publicError(w, 400, "invalid_upload_name")
				return
			}
			data, readErr := io.ReadAll(io.LimitReader(part, 257))
			if readErr != nil {
				uploadFailure(w, readErr)
				return
			}
			if len(data) > 256 || bytes.ContainsAny(data, "\x00\r\n") {
				publicError(w, 400, "invalid_upload_name")
				return
			}
			name = string(data)
		default:
			publicError(w, 400, "unexpected_upload_field")
			return
		}
	}
	// Check trailing bytes too: the limit is for the whole request, not just the file.
	if _, err = io.Copy(io.Discard, r.Body); err != nil {
		uploadFailure(w, err)
		return
	}
	if files != 1 {
		publicError(w, 400, "missing file")
		return
	}
	if err = out.Close(); err != nil {
		publicError(w, 503, "upload_unavailable")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err = validateArchive(ctx, dest); err != nil {
		publicError(w, 400, err.Error())
		return
	}
	if name == "" {
		name = filename
	}
	id := uuid.NewString()
	now := time.Now().UTC()
	_, err = db.ExecContext(ctx, `INSERT INTO batches (id,user_id,name,status,zip_path,clip_count,duration_seconds,created_at,updated_at) VALUES ($1,$2,$3,'pending',$4,0,0,$5,$6)`, id, currentUserID(r), name, dest, now, now)
	if err != nil {
		publicError(w, 503, "upload_unavailable")
		return
	}
	payload, _ := json.Marshal(map[string]string{"id": id, "name": name, "zip_path": dest})
	qctx, qcancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer qcancel()
	if err = rdb.RPush(qctx, "jobs:batch", payload).Err(); err != nil {
		// A lost Redis response is ambiguous. Coordinate with the worker's pending
		// claim before deleting: a processing batch already owns the ZIP.
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		result, cleanupErr := db.ExecContext(cleanupCtx, `UPDATE batches SET status='failed',zip_path='',updated_at=CURRENT_TIMESTAMP WHERE id=$1 AND status='pending'`, id)
		if cleanupErr != nil {
			transferred = true
			log.Printf("event=upload_handoff_reconciliation_required batch_id=%s", id)
		} else if n, e := result.RowsAffected(); e != nil || n == 0 {
			transferred = true
		}
		publicError(w, 503, "upload_handoff_failed")
		return
	}
	transferred = true
	writeJSON(w, map[string]any{"batch": Batch{ID: id, UserID: currentUserID(r), Name: name, Status: "pending", CreatedAt: now, UpdatedAt: now}})
}

func archivePath(name string) (string, bool) {
	name = strings.ReplaceAll(name, "\\", "/")
	trimmed := strings.TrimSuffix(name, "/")
	parts := strings.Split(trimmed, "/")
	if len(name) > 512 || len(parts) > 8 || trimmed == "" {
		return "", false
	}
	for _, p := range parts {
		if p == "" || p == "." || p == ".." || strings.TrimRight(p, " .") != p {
			return "", false
		}
		for _, c := range p {
			if c < 32 || c > 126 || strings.ContainsRune(":*?\"<>|", c) {
				return "", false
			}
		}
		base := strings.ToUpper(strings.Split(p, ".")[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return "", false
		}
	}
	return strings.ToLower(trimmed), true
}

// Bound central-directory allocation before archive/zip parses any members.
// ZIP64 and multi-disk archives are deliberately outside this small-batch policy.
func boundedZipDirectory(f *os.File, size int64) bool {
	n := min(size, int64(65557))
	if n < 22 {
		return false
	}
	tail := make([]byte, n)
	if _, err := f.ReadAt(tail, size-n); err != nil {
		return false
	}
	i := bytes.LastIndex(tail, []byte{'P', 'K', 5, 6})
	if i < 0 || i+22 > len(tail) {
		return false
	}
	e := tail[i:]
	if len(e) != 22+int(binary.LittleEndian.Uint16(e[20:22])) {
		return false
	}
	entries := binary.LittleEndian.Uint16(e[10:12])
	dirSize := binary.LittleEndian.Uint32(e[12:16])
	offset := binary.LittleEndian.Uint32(e[16:20])
	return binary.LittleEndian.Uint32(e[4:8]) == 0 && entries > 0 && entries <= maxArchiveEntries && binary.LittleEndian.Uint16(e[8:10]) == entries && dirSize <= 128<<10 && int64(offset)+int64(dirSize) == size-int64(len(e))
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

func validateArchive(ctx context.Context, file string) error {
	invalid := errors.New("invalid_archive")
	f, err := os.Open(file)
	if err != nil {
		return invalid
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !boundedZipDirectory(f, st.Size()) {
		return invalid
	}
	z, err := zip.NewReader(f, st.Size())
	if err != nil || len(z.File) == 0 || len(z.File) > maxArchiveEntries {
		return invalid
	}
	seen := map[string]bool{}
	paths := map[string]bool{}
	var total uint64
	videos := 0
	for _, entry := range z.File {
		p, ok := archivePath(entry.Name)
		if !ok || seen[p] || entry.Mode()&os.ModeSymlink != 0 || (!entry.Mode().IsRegular() && !entry.FileInfo().IsDir()) || entry.Flags&1 != 0 {
			return errors.New("unsafe_archive")
		}
		seen[p] = true
		paths[p] = entry.FileInfo().IsDir()
		if entry.Method != zip.Store && entry.Method != zip.Deflate {
			return errors.New("unsupported_archive")
		}
		if entry.UncompressedSize64 > maxEntryBytes || total+entry.UncompressedSize64 > maxExpandedBytes || entry.UncompressedSize64 > 100*max(entry.CompressedSize64, 1) {
			return errors.New("archive_resource_limit")
		}
		total += entry.UncompressedSize64
		if entry.FileInfo().IsDir() {
			continue
		}
		if !videoExtensions[path.Ext(p)] {
			return errors.New("unsupported_archive_member")
		}
		videos++
		src, err := entry.Open()
		if err != nil {
			return invalid
		}
		n, err := io.Copy(io.Discard, io.LimitReader(contextReader{ctx, src}, int64(entry.UncompressedSize64)+1))
		src.Close()
		if err != nil || n != int64(entry.UncompressedSize64) || n == 0 {
			return invalid
		}
	}
	for p := range paths {
		for parent := path.Dir(p); parent != "."; parent = path.Dir(parent) {
			if dir, exists := paths[parent]; exists && !dir {
				return errors.New("unsafe_archive")
			}
		}
	}
	if videos == 0 {
		return errors.New("empty_archive")
	}
	return nil
}

package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func zipFixture(t *testing.T, names []string) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	for _, name := range names {
		w, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte("fixture-media-bytes"))
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func uploadRequest(t *testing.T, data []byte, files int) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	for i := 0; i < files; i++ {
		part, _ := m.CreateFormFile("file", "batch.zip")
		part.Write(data)
	}
	m.Close()
	req := httptest.NewRequest("POST", "/api/batches", &body)
	req.Header.Set("Content-Type", m.FormDataContentType())
	req = req.WithContext(context.WithValue(req.Context(), userKey, testUser))
	rec := httptest.NewRecorder()
	createBatch(rec, req)
	return rec
}

func TestUploadRejectsAndCleansBeforePersistence(t *testing.T) {
	securityTestSetup(t)
	root := t.TempDir()
	t.Setenv("UPLOAD_DIR", root)
	for _, tc := range []struct {
		name   string
		data   []byte
		files  int
		limit  int64
		status int
	}{
		{"missing", nil, 0, 1 << 20, 400}, {"zero", nil, 1, 1 << 20, 400}, {"corrupt", []byte("not-zip"), 1, 1 << 20, 400},
		{"empty", zipFixture(t, nil), 1, 1 << 20, 400}, {"multiple", zipFixture(t, []string{"clip.mp4"}), 2, 1 << 20, 400},
		{"cap", zipFixture(t, []string{"clip.mp4"}), 1, 50, 413},
		{"traversal", zipFixture(t, []string{"../clip.mp4"}), 1, 1 << 20, 400},
		{"unsupported", zipFixture(t, []string{"notes.txt"}), 1, 1 << 20, 400},
		{"duplicate", zipFixture(t, []string{"clip.mp4", "CLIP.mp4"}), 1, 1 << 20, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			security.uploadBytes = tc.limit
			rec := uploadRequest(t, tc.data, tc.files)
			if rec.Code != tc.status {
				t.Fatalf("got %d %s", rec.Code, rec.Body)
			}
			entries, _ := os.ReadDir(root)
			if len(entries) != 0 {
				t.Fatal("orphan upload")
			}
			if strings.Contains(rec.Body.String(), root) {
				t.Fatal("path leak")
			}
		})
	}
}

func TestArchivePathsAndResources(t *testing.T) {
	for _, name := range []string{"../x.mp4", "..\\x.mp4", "/x.mp4", "C:x.mp4", "a/./x.mp4", "a//x.mp4", "CON.mp4", strings.Repeat("a/", 9) + "x.mp4", "x.mp4/child.mp4"} {
		file := filepath.Join(t.TempDir(), "test.zip")
		names := []string{name}
		if name == "x.mp4/child.mp4" {
			names = append(names, "x.mp4")
		}
		os.WriteFile(file, zipFixture(t, names), 0600)
		if validateArchive(context.Background(), file) == nil {
			t.Fatalf("accepted %s", name)
		}
	}
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, _ := z.Create("bomb.mp4")
	w.Write(bytes.Repeat([]byte{0}, 100000))
	z.Close()
	file := filepath.Join(t.TempDir(), "bomb.zip")
	os.WriteFile(file, b.Bytes(), 0600)
	if err := validateArchive(context.Background(), file); err == nil {
		t.Fatal("ratio accepted")
	}
	b.Reset()
	z = zip.NewWriter(&b)
	h := &zip.FileHeader{Name: "link.mp4"}
	h.SetMode(os.ModeSymlink | 0777)
	w, _ = z.CreateHeader(h)
	w.Write([]byte("target"))
	z.Close()
	os.WriteFile(file, b.Bytes(), 0600)
	if err := validateArchive(context.Background(), file); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestUploadHandoffCleanup(t *testing.T) {
	for _, stage := range []string{"database-failure", "database-cleanup-failure", "queue-failure", "claimed", "success"} {
		t.Run(stage, func(t *testing.T) {
			mock, server := securityTestSetup(t)
			root := t.TempDir()
			t.Setenv("UPLOAD_DIR", root)
			if strings.HasPrefix(stage, "database-") {
				mock.ExpectExec("INSERT INTO batches").WillReturnError(errors.New("private SQL"))
				if stage == "database-cleanup-failure" {
					mock.ExpectExec("DELETE FROM batches").WillReturnError(errors.New("private cleanup failure"))
				} else {
					mock.ExpectExec("DELETE FROM batches").WillReturnResult(sqlmock.NewResult(0, 1))
				}
			} else {
				mock.ExpectExec("INSERT INTO batches").WillReturnResult(sqlmock.NewResult(0, 1))
			}
			if stage == "queue-failure" || stage == "claimed" {
				server.SetError("private Redis")
				rows := int64(1)
				if stage == "claimed" {
					rows = 0
				}
				mock.ExpectExec("UPDATE batches").WillReturnResult(sqlmock.NewResult(0, rows))
			}
			rec := uploadRequest(t, zipFixture(t, []string{"clip.mp4"}), 1)
			entries, _ := os.ReadDir(root)
			if stage == "success" || stage == "claimed" || stage == "database-cleanup-failure" {
				if len(entries) != 1 {
					t.Fatal("lost owned ZIP")
				}
			} else if len(entries) != 0 {
				t.Fatal("orphan ZIP")
			}
			if stage == "success" {
				if rec.Code != 200 {
					t.Fatal(rec.Body)
				}
			} else if rec.Code != 503 {
				t.Fatal(rec.Body)
			}
			if strings.Contains(rec.Body.String(), "private") || strings.Contains(rec.Body.String(), root) {
				t.Fatal("leaked internals")
			}
		})
	}
}

func TestChunkedOversizeAndMalformedMultipartCleanup(t *testing.T) {
	securityTestSetup(t)
	root := t.TempDir()
	t.Setenv("UPLOAD_DIR", root)
	security.uploadBytes = 1024
	for _, tc := range []struct {
		data   string
		status int
	}{
		{"--b\r\nContent-Disposition: form-data; name=\"file\"; filename=\"x.zip\"\r\n\r\n" + strings.Repeat("x", 2048), 413},
		{"--b\r\nContent-Disposition: form-data; name=\"file\"; filename=\"x.zip\"\r\n\r\ntruncated", 400},
	} {
		req := httptest.NewRequest("POST", "/", strings.NewReader(tc.data))
		req.ContentLength = -1
		req.Header.Set("Content-Type", "multipart/form-data; boundary=b")
		rec := httptest.NewRecorder()
		createBatch(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("got %d: %s", rec.Code, rec.Body)
		}
		entries, _ := os.ReadDir(root)
		if len(entries) != 0 {
			t.Fatal("partial file retained")
		}
	}
}

func TestUploadStorageQuota(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 32; i++ {
		f, err := newUploadFile(root)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	if f, err := newUploadFile(root); err == nil {
		f.Close()
		t.Fatal("storage quota bypassed")
	}
}

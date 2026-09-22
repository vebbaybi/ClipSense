package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// This test uses the real schema, queries, authentication middleware and HTTP
// routes. It must only run against the disposable Gate 4 Postgres database.
func TestGate4PostgresIsolation(t *testing.T) {
	dsn := os.Getenv("GATE4_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable Gate 4 Postgres; enforced by the CI isolation job")
	}
	u, err := url.Parse(dsn)
	if err != nil || u.Hostname() != "127.0.0.1" || u.Path != "/clipsense_gate4" {
		t.Fatal("Gate 4 fixture requires the isolated loopback clipsense_gate4 database")
	}
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal("cannot open fixture database")
	}
	t.Cleanup(func() { admin.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	schema := "gate4_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal("cannot create isolated fixture schema")
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error("cannot remove isolated fixture schema")
		}
	})
	// search_path applies to every pooled connection, including nested queries.
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	fixtureDB, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal("cannot open schema-scoped connection")
	}
	fixtureDB.SetMaxOpenConns(10)
	oldDB, oldSecret := db, jwtSecret
	db = fixtureDB
	jwtSecret = make([]byte, 32)
	if _, err := rand.Read(jwtSecret); err != nil {
		t.Fatal("cannot initialize disposable signing key")
	}
	t.Cleanup(func() {
		fixtureDB.Close()
		db, jwtSecret = oldDB, oldSecret
	})
	if err := migrate(); err != nil {
		t.Fatal("fixture migration failed")
	}
	exec := func(statement string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, statement, args...); err != nil {
			t.Fatal("fixture insertion failed")
		}
	}
	type owner struct{ user, batch, clip, story, marker, token string }
	owners := []owner{
		{user: uuid.NewString(), batch: uuid.NewString(), clip: uuid.NewString(), story: uuid.NewString(), marker: "synthetic-owner-A-content"},
		{user: uuid.NewString(), batch: uuid.NewString(), clip: uuid.NewString(), story: uuid.NewString(), marker: "synthetic-owner-B-content"},
	}
	now := time.Now().UTC()
	for i := range owners {
		o := &owners[i]
		exec(`INSERT INTO users VALUES ($1,$2,$3,$4)`, o.user, o.user+"@example.test", "unused-fixture-hash", now)
		exec(`INSERT INTO batches VALUES ($1,$2,$3,'complete',$4,1,1,$5,$5)`, o.batch, o.user, o.marker, "/synthetic/"+o.marker+".zip", now)
		exec(`INSERT INTO clips VALUES ($1,$2,$3,$3,$3,$3,'neutral','beat','general',1,$4)`, o.clip, o.batch, o.marker, now)
		exec(`INSERT INTO storylines VALUES ($1,$2,$3,'fixture',$4)`, o.story, o.batch, o.marker, now)
		exec(`INSERT INTO storyline_clips VALUES ($1,$2,0)`, o.story, o.clip)
		o.token, err = issueToken(o.user)
		if err != nil {
			t.Fatal("cannot issue fixture token")
		}
	}
	checks := dependencyChecks{database: db.PingContext, redis: func(context.Context) error { return nil }}
	server := httptest.NewServer(newRouter(checks))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 10 * time.Second
	get := func(t *testing.T, token, path string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
		if err != nil {
			t.Fatal("cannot construct fixture request")
		}
		req.Header.Set("Authorization", "Bearer "+token)
		response, err := client.Do(req)
		if err != nil {
			t.Fatal("fixture HTTP request failed")
		}
		defer response.Body.Close()
		body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		if err != nil {
			t.Fatal("cannot read fixture response")
		}
		return response.StatusCode, string(body)
	}
	for i, own := range owners {
		foreign := owners[1-i]
		for _, suffix := range []string{"", "/export?format=json", "/export?format=csv"} {
			if !t.Run(string(rune('A'+i))+"_own"+suffix, func(t *testing.T) {
				status, body := get(t, own.token, "/api/batches/"+own.batch+suffix)
				if status != 200 || !strings.Contains(body, own.marker) || strings.Contains(body, foreign.marker) {
					t.Fatalf("owned resource boundary failed: HTTP %d", status)
				}
			}) {
				t.FailNow()
			}
			if !t.Run(string(rune('A'+i))+"_foreign"+suffix, func(t *testing.T) {
				status, body := get(t, own.token, "/api/batches/"+foreign.batch+suffix)
				missingStatus, missingBody := get(t, own.token, "/api/batches/"+uuid.NewString()+suffix)
				if status != 404 || status != missingStatus || body != missingBody || strings.Contains(body, foreign.marker) {
					t.Fatalf("foreign resource was not safely hidden: HTTP %d", status)
				}
			}) {
				t.FailNow()
			}
		}
		if !t.Run(string(rune('A'+i))+"_list", func(t *testing.T) {
			status, body := get(t, own.token, "/api/batches")
			if status != 200 || !strings.Contains(body, own.batch) || strings.Contains(body, foreign.batch) || strings.Contains(body, foreign.marker) {
				t.Fatal("batch list violated owner boundary")
			}
		}) {
			t.FailNow()
		}
	}

	// Model a corrupt cross-batch membership permitted by the current schema.
	// This is not a claim that an unprivileged API can create that relationship.
	exec(`INSERT INTO storyline_clips VALUES ($1,$2,1)`, owners[0].story, owners[1].clip)
	t.Run("cross_batch_storyline_membership", func(t *testing.T) {
		status, body := get(t, owners[0].token, "/api/batches/"+owners[0].batch)
		if strings.Contains(body, owners[1].marker) || strings.Contains(body, owners[1].clip) {
			t.Fatalf("STOP: cross-user clip disclosed through owned batch detail (HTTP %d); payload withheld", status)
		}
		if status != 200 || !strings.Contains(body, owners[0].clip) {
			t.Fatalf("owned data unavailable after malformed relationship: HTTP %d", status)
		}
	})
}

package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServePlaylist(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "movie one.mp4"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Clip Dance Men Mostafa & Reyhaneh.mp4"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("nope"), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/playlist.m3u", nil)
	req.Host = "127.0.0.1:9090"
	rec := httptest.NewRecorder()

	servePlaylist(rec, req, dir, false)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "#EXTM3U\r\n") {
		t.Fatalf("missing #EXTM3U header: %q", body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "audio/x-mpegurl; charset=utf-8" {
		t.Fatalf("content-type = %q", ct)
	}
	if strings.Contains(body, "tvg-name") {
		t.Fatalf("playlist should use simple EXTINF lines: %q", body)
	}
	if !strings.Contains(body, "movie%20one.mp4") {
		t.Fatalf("missing encoded video URL: %q", body)
	}
	if !strings.Contains(body, "Mostafa%20%26%20Reyhaneh.mp4") {
		t.Fatalf("ampersand must be encoded in URL: %q", body)
	}
	if strings.Contains(body, "readme.txt") {
		t.Fatalf("non-video file included: %q", body)
	}
}

func TestListVideos(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.mkv", "b.MP4", "notes.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	videos, err := listVideos(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 2 {
		t.Fatalf("got %d videos, want 2: %v", len(videos), videos)
	}
}

func TestListVideosRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "movies")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "root.mp4"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "nested.mkv"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	flat, err := listVideos(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(flat) != 1 || flat[0] != "root.mp4" {
		t.Fatalf("non-recursive = %v, want [root.mp4]", flat)
	}

	all, err := listVideos(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("recursive = %v, want 2 videos", all)
	}
	if !contains(all, "root.mp4") || !contains(all, "movies/nested.mkv") {
		t.Fatalf("recursive = %v, want root.mp4 and movies/nested.mkv", all)
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func TestM3uTitle(t *testing.T) {
	title := m3uTitle("Film Men Mostafa & Reyhaneh.mp4")
	if title != "Film Men Mostafa & Reyhaneh" {
		t.Fatalf("unexpected title: %q", title)
	}
}

func TestAllowAllIPNets(t *testing.T) {
	nets := allowAllIPNets()
	if len(nets) != 2 {
		t.Fatalf("got %d nets, want 2", len(nets))
	}
	if !nets[0].Contains(net.ParseIP("192.168.1.10")) {
		t.Fatal("expected IPv4 net to allow LAN clients")
	}
}

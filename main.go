package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var videoExtensions = map[string]bool{
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".webm": true,
	".m4v":  true,
	".ts":   true,
	".flv":  true,
	".wmv":  true,
}

func main() {
	folder := flag.String("folder", ".", "path to the folder containing video files")
	port := flag.Int("port", 8080, "HTTP listening port")
	recursive := flag.Bool("r", false, "search for videos recursively in subdirectories")
	friendlyName := flag.String("name", "Video2IPTVnDLNA", "DLNA server name shown to TVs and media players")
	flag.Parse()

	info, err := os.Stat(*folder)
	if err != nil {
		log.Fatalf("folder %q: %v", *folder, err)
	}
	if !info.IsDir() {
		log.Fatalf("folder %q is not a directory", *folder)
	}

	absFolder, err := filepath.Abs(*folder)
	if err != nil {
		log.Fatalf("resolve folder path: %v", err)
	}

	if err := startServer(absFolder, *port, *recursive, *friendlyName); err != nil {
		log.Fatal(err)
	}
}

func servePlaylist(w http.ResponseWriter, r *http.Request, folder string, recursive bool) {
	videos, err := listVideos(folder, recursive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	baseURL := requestBaseURL(r)

	w.Header().Set("Content-Type", "audio/x-mpegurl; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	var b strings.Builder
	b.WriteString("#EXTM3U\r\n")
	for _, name := range videos {
		title := m3uTitle(name)
		streamURL := videoStreamURL(baseURL, name)
		fmt.Fprintf(&b, "#EXTINF:-1,%s\r\n", title)
		b.WriteString(streamURL)
		b.WriteString("\r\n")
	}

	_, _ = w.Write([]byte(b.String()))
}

func requestBaseURL(r *http.Request) string {
	host := r.Host
	if host == "" {
		host = "localhost"
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

func videoStreamURL(baseURL, filename string) string {
	return baseURL + "/videos/" + encodeFilenameForURL(filename)
}

func encodeFilenameForURL(relPath string) string {
	segments := strings.Split(filepath.ToSlash(relPath), "/")
	for i, seg := range segments {
		escaped := url.PathEscape(seg)
		escaped = strings.ReplaceAll(escaped, "&", "%26")
		segments[i] = escaped
	}
	return strings.Join(segments, "/")
}

func m3uTitle(relPath string) string {
	title := strings.TrimSuffix(relPath, filepath.Ext(relPath))
	title = filepath.ToSlash(title)
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.ReplaceAll(title, "\n", " ")
	return title
}

func listVideos(folder string, recursive bool) ([]string, error) {
	if !recursive {
		return listVideosInDir(folder)
	}

	var videos []string
	err := filepath.WalkDir(folder, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !isVideoFile(entry.Name()) {
			return nil
		}
		rel, err := filepath.Rel(folder, path)
		if err != nil {
			return err
		}
		videos = append(videos, filepath.ToSlash(rel))
		return nil
	})
	return videos, err
}

func listVideosInDir(folder string) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	var videos []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if isVideoFile(entry.Name()) {
			videos = append(videos, entry.Name())
		}
	}
	return videos, nil
}

func isVideoFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return videoExtensions[ext]
}

package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
        "io/fs"
)

//go:embed ui/*
var uiFS embed.FS

var (
	serverAddr     = getEnv("SERVER_ADDR", ":8443")
	uploadDir      = getEnv("UPLOAD_DIR", "./uploads")
	serverApiToken = getEnv("SERVER_API_TOKEN", "change-me-server-token") // 클라 → 서버 업로드 보호
	clientToken    = getEnv("CLIENT_TOKEN", "change-me-client-token")     // 브라우저 → 클라 호출 보호 (페이지에 주입)
)

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatalf("mkdir uploads: %v", err)
	}

        sub, err := fs.Sub(uiFS, "ui")
        if err != nil { log.Fatal(err) }

	mux := http.NewServeMux()

	// UI: index.html에 클라이언트 토큰을 주입해서 전달
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := uiFS.ReadFile("ui/index.html")
		if err != nil {
			http.Error(w, "index not found", 500)
			return
		}
		html := strings.ReplaceAll(string(b), "__CLIENT_TOKEN__", clientToken)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})

	// JS 서빙
	mux.Handle("/clientBridge.js", http.FileServer(http.FS(sub)))

	// 서버에 저장된 파일 리스트
	mux.HandleFunc("/api/files", func(w http.ResponseWriter, r *http.Request) {
		type fileInfo struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
			Path string `json:"path"`
			URL  string `json:"url"`
		}
		var list []fileInfo
		entries, _ := os.ReadDir(uploadDir)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			fi, _ := e.Info()
			name := e.Name()
			list = append(list, fileInfo{
				Name: name,
				Size: fi.Size(),
				Path: filepath.Join(uploadDir, name),
				URL:  "/api/download/" + name,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	})

	// 서버 → 브라우저/클라 다운로드용 엔드포인트
	mux.HandleFunc("/api/download/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/download/")
		fp := filepath.Join(uploadDir, filepath.Clean(name))
		f, err := os.Open(fp)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
                fi, err := f.Stat()
                if err != nil { 
                        http.Error(w, "stat error", 500); return
                }
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(fp)))
		http.ServeContent(w, r, name, fi.ModTime(), f)
	})

	// 클라이언트가 서버로 업로드하는 엔드포인트 (보호: X-Server-Token)
	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Server-Token") != serverApiToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err := r.ParseMultipartForm(1 << 30); err != nil { // ~1GB
			http.Error(w, "bad form", 400)
			return
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file missing", 400)
			return
		}
		defer file.Close()
		dst := filepath.Join(uploadDir, filepath.Base(hdr.Filename))
		out, err := os.Create(dst)
		if err != nil {
			http.Error(w, "write error", 500)
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "copy error", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	log.Printf("Server on %s (uploads: %s)", serverAddr, uploadDir)
	log.Fatal(http.ListenAndServe(serverAddr, withSecurityHeaders(mux)))
}

func statsModTime(f *os.File) (mt timeLike) {
	fi, _ := f.Stat()
	return timeLike{fi.ModTime()}
}

type timeLike struct{ t interface{ String() string } }

// 최소 보안 헤더
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}


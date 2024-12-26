package main

import (
	"archive/zip"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/* static/*
var contentFS embed.FS

// FileInfo stores information about files and directories
type FileInfo struct {
	Name          string
	Path          string
	IsDir         bool
	Size          int64
	FormattedSize string
	ModTime       string
	FileType      string
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3", ".wav", ".ogg", ".m4a", ".flac":
		return "audio"
	case ".mp4", ".avi", ".mov", ".wmv", ".mkv":
		return "video"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "zip"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg":
		return "image"
	case ".srt", ".ass", ".vtt", ".sub", ".ssa", ".txt":
		return "subtitles"
	default:
		return "other-media"
	}
}


// getDirSize calculates total size of a directory
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// getLocalIPs returns all local IP addresses
func getLocalIPs() ([]string, error) {
	var ips []string
	
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			if ipNet.IP.To4() == nil {
				continue
			}

			if ipNet.IP.IsLoopback() {
				continue
			}

			if ipNet.IP.IsLinkLocalUnicast() {
				continue
			}

			ips = append(ips, ipNet.IP.String())
		}
	}

	return ips, nil
}

func handleFileServer(w http.ResponseWriter, r *http.Request, root string) {
	urlPath := filepath.Clean(r.URL.Path)
	fullPath := filepath.Join(root, urlPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		files, err := os.ReadDir(fullPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fileInfos := make([]FileInfo, 0)
		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				continue
			}
			
			size := info.Size()
			formattedSize := formatSize(size)
			
			if file.IsDir() {
				dirSize, err := getDirSize(filepath.Join(fullPath, file.Name()))
				if err == nil {
					size = dirSize
					formattedSize = formatSize(dirSize)
				}
			}
			
			fileType := "folder"
			if !file.IsDir() {
				fileType = getFileType(file.Name())
			}
			
			fileInfos = append(fileInfos, FileInfo{
				Name:          file.Name(),
				Path:          filepath.Join(urlPath, file.Name()),
				IsDir:         file.IsDir(),
				Size:          size,
				FormattedSize: formattedSize,
				ModTime:       info.ModTime().Format("2006-01-02 15:04:05"),
				FileType:      fileType,
			})
		}

		renderTemplate(w, fileInfos, urlPath, fullPath)
		return
	}

	http.ServeFile(w, r, fullPath)
}

// Add this function to main.go
func loadSVGIcon(name string) (template.HTML, error) {
    content, err := contentFS.ReadFile(filepath.Join("static/icons", name+".svg"))
    if err != nil {
        return "", err
    }
    return template.HTML(content), nil
}

// Add this template function
var templateFuncs = template.FuncMap{
    "svgIcon": loadSVGIcon,
}

func renderTemplate(w http.ResponseWriter, files []FileInfo, urlPath string, absolutePath string) {
    tmpl, err := template.New("index.html").Funcs(templateFuncs).ParseFS(contentFS, "templates/index.html")
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        log.Printf("Template parsing error: %v", err)
        return
    }
    
    // Calculate total size from files
    var totalSize int64
    for _, file := range files {
        totalSize += file.Size
    }
    
    data := struct {
        Files        []FileInfo
        CurrentPath  string
        AbsolutePath string
        TotalSize    string
    }{
        Files:        files,
        CurrentPath:  urlPath,
        AbsolutePath: absolutePath,
        TotalSize:    formatSize(totalSize),
    }
    
    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        log.Printf("Template execution error: %v", err)
    }
}

func main() {
	var dirPath string
	flag.StringVar(&dirPath, "dir", "", "Directory to serve (default: ~/SharedFiles)")
	flag.Parse()

	if dirPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Error getting home directory:", err)
		}
		dirPath = filepath.Join(homeDir, "SharedFiles")
	} else {
		absPath, err := filepath.Abs(dirPath)
		if err != nil {
			log.Fatal("Error resolving path:", err)
		}
		dirPath = absPath
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Fatal("Error creating directory:", err)
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", dirPath)
	}

	// Set up handlers

	// Serve static files (CSS, JS, etc.)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		file, err := contentFS.ReadFile(r.URL.Path[1:]) // Remove leading slash
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Set content type based on file extension
		switch {
		case strings.HasSuffix(r.URL.Path, ".css"):
			w.Header().Set("Content-Type", "text/css")
		case strings.HasSuffix(r.URL.Path, ".svg"):
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		w.Write(file)
	})	

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handleFileServer(w, r, dirPath)
	})

	http.HandleFunc("/zip/", func(w http.ResponseWriter, r *http.Request) {
		zipPath := strings.TrimPrefix(r.URL.Path, "/zip")
		if zipPath == "" {
			zipPath = "/"
		}
		fullPath := filepath.Join(dirPath, filepath.Clean(zipPath))
		
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", filepath.Base(zipPath)))
		
		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()

		err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(fullPath, path)
			if err != nil {
				return err
			}
			if relPath == "." {
				return nil
			}
			header.Name = relPath

			if info.IsDir() {
				header.Name += "/"
				_, err = zipWriter.CreateHeader(header)
				return err
			}

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get local IP addresses
	ips, err := getLocalIPs()
	if err != nil {
		log.Printf("Warning: Could not determine local IP addresses: %v", err)
	}

	log.Printf("Serving files from: %s", dirPath)
	log.Printf("Local addresses:")
	log.Printf("  http://localhost:%s", port)
	for _, ip := range ips {
		log.Printf("  http://%s:%s", ip, port)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
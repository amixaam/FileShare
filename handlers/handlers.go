package handlers

import (
	"archive/zip"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fileshare/utils"
)

type Server struct {
    ContentFS    embed.FS
    RootDir      string
    ShowDotfiles bool
    Domain       string
}

func NewServer(contentFS embed.FS, rootDir string, showDotfiles bool, domain string) *Server {
    return &Server{
        ContentFS:    contentFS,
        RootDir:      rootDir,
        ShowDotfiles: showDotfiles,
        Domain:       domain,
    }
}

func (s *Server) HandleFileServer(w http.ResponseWriter, r *http.Request) {
    urlPath := filepath.Clean(r.URL.Path)
    fullPath := filepath.Join(s.RootDir, urlPath)

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
            if !s.ShowDotfiles && strings.HasPrefix(file.Name(), ".") {
                continue
            }

            info, err := file.Info()
            if err != nil {
                continue
            }
            
            size := info.Size()
            formattedSize := utils.FormatSize(size)
            
            if file.IsDir() {
                dirSize, err := utils.GetDirSize(filepath.Join(fullPath, file.Name()))
                if err == nil {
                    size = dirSize
                    formattedSize = utils.FormatSize(dirSize)
                }
            }
            
            fileType := "folder"
            if !file.IsDir() {
                fileType = utils.GetFileType(file.Name())
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

        s.renderTemplate(w, fileInfos, urlPath, fullPath)
        return
    }

    // If it's a hidden file and ShowDotfiles is false, return 404
    if !s.ShowDotfiles && strings.HasPrefix(filepath.Base(fullPath), ".") {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    http.ServeFile(w, r, fullPath)
}

func (s *Server) HandleStatic(w http.ResponseWriter, r *http.Request) {
    file, err := s.ContentFS.ReadFile(r.URL.Path[1:])
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
}

func (s *Server) HandleZip(w http.ResponseWriter, r *http.Request) {
    zipPath := strings.TrimPrefix(r.URL.Path, "/zip")
    if zipPath == "" {
        zipPath = "/"
    }
    fullPath := filepath.Join(s.RootDir, filepath.Clean(zipPath))
    
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
}

func (s *Server) loadSVGIcon(name string) (template.HTML, error) {
    content, err := s.ContentFS.ReadFile(filepath.Join("static/icons", name+".svg"))
    if err != nil {
        return "", err
    }
    return template.HTML(content), nil
}

func (s *Server) getTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "svgIcon": s.loadSVGIcon,
    }
}

func (s *Server) renderTemplate(w http.ResponseWriter, files []FileInfo, urlPath string, absolutePath string) {
    tmpl, err := template.New("index.html").Funcs(s.getTemplateFuncs()).ParseFS(s.ContentFS, "templates/index.html")
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        log.Printf("Template parsing error: %v", err)
        return
    }
    
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
        TotalSize:    utils.FormatSize(totalSize),
    }
    
    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        log.Printf("Template execution error: %v", err)
    }
}
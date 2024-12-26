package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func FormatSize(bytes int64) string {
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

func GetFileType(filename string) string {
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

func GetDirSize(path string) (int64, error) {
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

func GetLocalIPs() ([]string, error) {
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
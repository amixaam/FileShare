package handlers

type FileInfo struct {
    Name          string
    Path          string
    IsDir         bool
    Size          int64
    FormattedSize string
    ModTime       string
    FileType      string
}
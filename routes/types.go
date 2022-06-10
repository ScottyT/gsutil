package routes

import (
	"cloud.google.com/go/storage"
)

type FolderInfo struct {
	FolderPath  string   `json:"folderPath"`
	SourceFiles []string `json:"sourceFiles"`
	DestFolder  string   `json:"destFolder"`
}
type FilesList struct {
	Files []FilesList `json:"files"`
}
type ImageObjectsInfo struct {
	Name       string `json:"name"`
	ImageUrl   string `json:"imageUrl"`
	FolderName string `json:"folderName"`
}
type FolderObjectsInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
type FileObjectsInfo struct {
	Folders []*FolderObjectsInfo `json:"folders"`
	Images  []*ImageObjectsInfo  `json:"images"`
	Pdfs    []*ImageObjectsInfo  `json:"pdfs"`
}
type ResponseImgArr struct {
	downloadUrl string
	files       []ImageObjectsInfo
}

type command struct {
	name        string
	respMessage string
	args        []string
}
type ClientUploader struct {
	cl        *storage.Client
	directory *StorageUploader
}
type StorageUploader struct {
	bucketName string
	uploadPath string
}
type Message struct {
	message string
}

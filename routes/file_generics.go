package routes

import (
	"gsutil/config"
	"gsutil/generics"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

var uploader *ClientUploader
var su StorageUploader
var appconfig *config.EnvConfig
var respMessage *Message
var file *ImageObjectsInfo
var folder *FolderObjectsInfo

func GetFileObject[K comparable, V *ImageObjectsInfo | *FolderObjectsInfo](clu *ClientUploader, filename string, filetype V) {
	file_type := reflect.TypeOf(filetype).Elem()
	dir, _ := os.Getwd()
	storageKeyPath, err := filepath.Abs(dir + "/cr-storage-key.pem")
	if err != nil {
		panic("Unable to load service account file")
	}
	pkey, err := ioutil.ReadFile(storageKeyPath)
	if err != nil {
		resp = Response{Status: http.StatusNotFound, Error: "No private key found"}
	}
	opts := &storage.SignedURLOptions{
		GoogleAccessID: appconfig.SaEmail,
		PrivateKey:     pkey,
		Method:         "GET",
		Expires:        time.Now().Add(15 * time.Minute),
	}
	u, err := storage.SignedURL(clu.directory.bucketName, filename, opts)
	if err != nil {
		resp = Response{Status: http.StatusNotFound, Error: err.Error()}
	}
	sarr := generics.ReverseArrayGeneric(strings.Split(filename, "/"))
	switch file_type {
	case reflect.TypeOf(file).Elem():
		file = &ImageObjectsInfo{
			Name:       filename,
			FolderName: sarr[1],
			ImageUrl:   u,
		}
	case reflect.TypeOf(folder).Elem():
		folder = &FolderObjectsInfo{
			Name: sarr[1],
			Path: filename,
		}
	default:
		str := "The type needs to be either file or folder."
		respMessage.message = str
	}
}

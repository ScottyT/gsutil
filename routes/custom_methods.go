package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gsutil/config"
	"gsutil/generics"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

func InitStorageClient(bucket string, client *storage.Client) {
	appconfig = config.InitEnv()
	su = StorageUploader{
		bucketName: bucket,
	}
	uploader = &ClientUploader{
		cl:        client,
		directory: &su,
	}
}

func (c *command) download() string {
	//cmd := exec.CommandContext(r.Context(), "/bin/bash", "script.sh", folder, dir)
	cmd := exec.Command("gsutil", c.args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Command.Output: %v", err)
		return err.Error()
	}
	return bytes.NewBuffer(out).String()
}

func (clu *ClientUploader) List(prefix, delim string) ([]byte, Response) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	var folders []*FolderObjectsInfo
	var images []*ImageObjectsInfo
	var files *FileObjectsInfo
	var pdfs []*ImageObjectsInfo

	it := clu.cl.Bucket(clu.directory.bucketName).Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: delim,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, Response{Status: http.StatusBadRequest, Error: err.Error()}
		}
		if strings.Contains(attrs.ContentType, "image") {
			GetFileObject[*ImageObjectsInfo](clu, attrs.Name, &ImageObjectsInfo{})

			images = append(images, file)
		}
		if strings.Contains(attrs.ContentType, "pdf") {
			GetFileObject[*ImageObjectsInfo](clu, attrs.Name, &ImageObjectsInfo{})
			pdfs = append(pdfs, file)
		}

		if attrs.Prefix != "" {
			GetFileObject[*FolderObjectsInfo](clu, attrs.Prefix, &FolderObjectsInfo{})
			folders = append(folders, folder)
		}
		files = &FileObjectsInfo{Folders: folders, Images: images, Pdfs: pdfs}
	}
	e, err := json.Marshal(files)
	if err != nil {
		return nil, Response{Status: http.StatusNotFound, Error: err.Error()}
	}
	return e, Response{}
}

func (clu *ClientUploader) ReadImage(fileName string) (*ImageObjectsInfo, Response) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	rc, err := clu.cl.Bucket(clu.directory.bucketName).Object(fileName).Attrs(ctx)
	if err != nil {
		return nil, Response{Status: http.StatusNotFound, Error: err.Error()}
	}

	GetFileObject[*ImageObjectsInfo](clu, rc.Name, &ImageObjectsInfo{})
	if resp.Error != "" {
		return nil, resp
	}

	return file, Response{}
}

func (clu *ClientUploader) Upload(files []*multipart.FileHeader, path string) ([]ImageObjectsInfo, Response) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	var imageArr []ImageObjectsInfo
	var token string
	var ok bool
	var imageUrl string

	for _, file := range files {
		read, err := file.Open()
		if err != nil {
			return nil, Response{Status: http.StatusBadRequest, Error: err.Error()}
		}
		wc := clu.cl.Bucket(clu.directory.bucketName).Object(path + file.Filename).NewWriter(ctx)
		if _, err := io.Copy(wc, read); err != nil {
			return nil, Response{Status: http.StatusBadRequest, Error: "Error copying from file"}
		}
		uid := uuid.NewString()
		o := clu.cl.Bucket(clu.directory.bucketName).Object(path + file.Filename)
		//wc.Metadata["firebaseStorageDownloadTokens"] = uid
		objUpdate := storage.ObjectAttrsToUpdate{Metadata: map[string]string{
			"firebaseStorageDownloadTokens": uid,
		}}

		if err := wc.Close(); err != nil {
			return nil, Response{Status: http.StatusBadRequest, Error: err.Error()}
		}
		if _, err := o.Update(ctx, objUpdate); err != nil {
			return nil, Response{Status: http.StatusBadRequest, Error: err.Error()}
		}
		if token, ok = objUpdate.Metadata["firebaseStorageDownloadTokens"]; ok {
			imageUrl = "https://firebasestorage.googleapis.com/v0/b/" + clu.directory.bucketName + "/o/" + url.QueryEscape(wc.Attrs().Name) + "?alt=media&token=" + token
		}
		if strings.Contains(wc.Attrs().ContentType, "image") || strings.Contains(wc.Attrs().ContentType, "pdf") {
			imageArr = append(imageArr, ImageObjectsInfo{Name: wc.Attrs().Name, ImageUrl: imageUrl})
		}
	}
	return imageArr, Response{}
}

func (clu *ClientUploader) UploadImageInUser(file multipart.File, object string) (string, Response) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	var token string
	var ok bool
	var imageUrl string
	wc := clu.cl.Bucket(clu.directory.bucketName).Object(clu.directory.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return "", Response{Status: http.StatusBadRequest, Error: "io.Copy: " + err.Error()}
	}
	uid := uuid.NewString()
	o := clu.cl.Bucket(clu.directory.bucketName).Object(clu.directory.uploadPath + object)
	objUpdate := storage.ObjectAttrsToUpdate{Metadata: map[string]string{
		"firebaseStorageDownloadTokens": uid,
	}}
	if err := wc.Close(); err != nil {
		return "", Response{Status: http.StatusBadRequest, Error: "Writer.Close:" + err.Error()}
	}
	if _, err := o.Update(ctx, objUpdate); err != nil {
		return "", Response{Status: http.StatusBadRequest, Error: "Update error: " + err.Error()}
	}
	if token, ok = objUpdate.Metadata["firebaseStorageDownloadTokens"]; ok {
		imageUrl = "https://firebasestorage.googleapis.com/v0/b/" + clu.directory.bucketName + "/o/" + url.QueryEscape(wc.Attrs().Name) + "?alt=media&token=" + token
	}

	return imageUrl, Response{}
}

func (clu *ClientUploader) Moving(object, destDir string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	objName := generics.ReverseArrayGeneric(strings.Split(object, "/"))
	src := clu.cl.Bucket(clu.directory.bucketName).Object(object)
	dst := clu.cl.Bucket(clu.directory.bucketName).Object(destDir + objName[0])
	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("Object(%q).CopierFrom(%q).Run: %v", destDir+object, object, err)
	}
	if err := src.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", object, err)
	}
	return nil
	//fmt.Fprintf(w, "Blob %v moved to %v.\n", object, dstName)
}

func (clu *ClientUploader) CreateFolder(foldername string) (string, Response) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	b := []byte(foldername + "/")
	buf := bytes.NewBuffer(b)
	defer cancel()
	wc := clu.cl.Bucket(clu.directory.bucketName).Object(foldername + "/").NewWriter(ctx)
	wc.ContentType = "application/x-www-form-urlencoded;charset=UTF-8"
	if _, err := io.Copy(wc, buf); err != nil {
		return "", Response{Status: http.StatusBadRequest, Error: "io.Copy:" + err.Error()}
	}
	if err := wc.Close(); err != nil {
		return "", Response{Status: http.StatusBadRequest, Error: "Error closing the writer"}
	}
	return "Created" + foldername + "directory!", Response{}
}

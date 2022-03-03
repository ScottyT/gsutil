package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gsutil/config"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
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
	Name      string `json:"name"`
	ImageUrl  string `json:"imageUrl"`
	Subfolder string `json:"subFolder"`
}
type FolderObjectsInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
type FileObjectsInfo struct {
	Folders []FolderObjectsInfo `json:"folders"`
	Images  []ImageObjectsInfo  `json:"images"`
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

var uploader *ClientUploader
var su StorageUploader
var appconfig *config.EnvConfig
var fileslist FilesList

var cmd *exec.Cmd

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
func (c *command) move() string {
	cmd := exec.Command("gsutil", c.args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return err.Error()
	}
	return bytes.NewBuffer(out).String()
}

func reverseArray(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func (clu *ClientUploader) List(prefix, delim string) ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	var folders []FolderObjectsInfo
	var images []ImageObjectsInfo
	var files *FileObjectsInfo

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
			fmt.Printf("Bucket(%q).Objects(): %v", clu.directory.bucketName, err)
		}
		if strings.Contains(attrs.ContentType, "image") {
			token := attrs.Metadata["firebaseStorageDownloadTokens"]
			images = append(images, ImageObjectsInfo{
				Name:      attrs.Name,
				ImageUrl:  "https://firebasestorage.googleapis.com/v0/b/" + clu.directory.bucketName + "/o/" + url.QueryEscape(attrs.Name) + "?alt=media&token=" + token,
				Subfolder: "",
			})
		}
		if attrs.Prefix != "" {
			sarr := reverseArray(strings.Split(attrs.Prefix, "/"))
			folders = append(folders, FolderObjectsInfo{Name: sarr[1], Path: attrs.Prefix})
		}

		files = &FileObjectsInfo{Folders: folders, Images: images}
	}
	e, err := json.Marshal(files)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (clu *ClientUploader) ReadImage(fileName string) (*ImageObjectsInfo, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	var image *ImageObjectsInfo
	rc, err := clu.cl.Bucket(clu.directory.bucketName).Object(fileName).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	token := rc.Metadata["firebaseStorageDownloadTokens"]
	image = &ImageObjectsInfo{
		Name:      rc.Name,
		ImageUrl:  "https://firebasestorage.googleapis.com/v0/b/" + clu.directory.bucketName + "/o/" + url.QueryEscape(rc.Name) + "?alt=media&token=" + token,
		Subfolder: "",
	}

	return image, nil
}

func (clu *ClientUploader) Upload(files []*multipart.FileHeader, path string) ([]ImageObjectsInfo, error) {
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
			fmt.Println("Error reading file: ", err)
		}
		wc := clu.cl.Bucket(clu.directory.bucketName).Object(path + file.Filename).NewWriter(ctx)
		if _, err := io.Copy(wc, read); err != nil {
			return nil, fmt.Errorf("io.Copy: %v\n", err)
		}
		uid := uuid.NewString()
		o := clu.cl.Bucket(clu.directory.bucketName).Object(path + file.Filename)
		//wc.Metadata["firebaseStorageDownloadTokens"] = uid
		objUpdate := storage.ObjectAttrsToUpdate{Metadata: map[string]string{
			"firebaseStorageDownloadTokens": uid,
		}}

		if err := wc.Close(); err != nil {
			return nil, fmt.Errorf("Writer.Close: %v\n", err)
		}
		if _, err := o.Update(ctx, objUpdate); err != nil {
			fmt.Println(err)
		}
		if token, ok = objUpdate.Metadata["firebaseStorageDownloadTokens"]; ok {
			imageUrl = "https://firebasestorage.googleapis.com/v0/b/" + clu.directory.bucketName + "/o/" + url.QueryEscape(wc.Attrs().Name) + "?alt=media&token=" + token
		}
		if strings.Contains(wc.Attrs().ContentType, "image") {
			imageArr = append(imageArr, ImageObjectsInfo{Name: wc.Attrs().Name, ImageUrl: imageUrl})
		}
	}
	return imageArr, nil
}

func (clu *ClientUploader) UploadImageInUser(file multipart.File, object string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	var token string
	var ok bool
	var imageUrl string
	wc := clu.cl.Bucket(clu.directory.bucketName).Object(clu.directory.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	uid := uuid.NewString()
	o := clu.cl.Bucket(clu.directory.bucketName).Object(clu.directory.uploadPath + object)
	objUpdate := storage.ObjectAttrsToUpdate{Metadata: map[string]string{
		"firebaseStorageDownloadTokens": uid,
	}}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v\n", err)
	}
	if _, err := o.Update(ctx, objUpdate); err != nil {
		return "", fmt.Errorf("Update error: %v", err)
	}
	if token, ok = objUpdate.Metadata["firebaseStorageDownloadTokens"]; ok {
		imageUrl = "https://firebasestorage.googleapis.com/v0/b/" + clu.directory.bucketName + "/o/" + url.QueryEscape(wc.Attrs().Name) + "?alt=media&token=" + token
	}

	return imageUrl, nil
}

func (clu *ClientUploader) ComposeFile(object1, object2, toObject string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	fmt.Println(clu.cl)
	src1 := clu.cl.Bucket(clu.directory.bucketName).Object(object1)
	src2 := clu.cl.Bucket(clu.directory.bucketName).Object(object2)
	dst := clu.cl.Bucket(clu.directory.bucketName).Object(toObject)
	_, err := dst.ComposerFrom(src1, src2).Run(ctx)
	if err != nil {
		return "", fmt.Errorf("ComposerFrom: %v", err)
	}

	return "Created composite obj", nil
}

func (clu *ClientUploader) Moving(object, destDir string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	objName := reverseArray(strings.Split(object, "/"))
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

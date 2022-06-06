package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

var Bucket string

type FilesIterator struct {
	pageInfo *iterator.PageInfo
	nextFunc func() error
	items    []string
}

type Response struct {
	Status int
	Error  string
}

var resp Response

func SendResponseError(c *gin.Context, response Response) {
	if len(response.Error) > 0 {
		c.JSON(response.Status, response.Error)
	}
}

func ExecCommand(com *command) (string, string, *bytes.Buffer) {
	switch com.name {
	case "download":
		return com.download(), com.respMessage, nil
	}
	return "", "", nil
}
func Iterate(f func(n string), items []string) {
	for i := 0; i < len(items); i++ {
		f(items[i])
	}
}

// For goroutine use only
func IterateRoutine(f func(n string), items []string, done chan bool) {
	for i := 0; i < len(items); i++ {
		f(items[i])
	}
	done <- true
}
func MovingFiles(c *gin.Context) {
	var output FolderInfo
	if err := json.NewDecoder(c.Request.Body).Decode(&output); err != nil {
		fmt.Print(err)
	}
	su = StorageUploader{
		bucketName: appconfig.StorageBucket,
	}
	done := make(chan bool, 1)
	go IterateRoutine(func(n string) {
		err := uploader.Moving(n, output.DestFolder)
		if err != nil {
			fmt.Fprintln(c.Writer, err)
			return
		}
	}, output.SourceFiles, done)
	<-done
	fmt.Fprintln(c.Writer, "Files move operation done!")
}
func DeleteObjects(c *gin.Context) {
	ctx := context.Background()
	var output FolderInfo
	if err := json.NewDecoder(c.Request.Body).Decode(&output); err != nil {
		fmt.Print(err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	su = StorageUploader{
		bucketName: appconfig.StorageBucket,
	}
	bucket := uploader.cl.Bucket(su.bucketName)
	done := make(chan bool, 1)
	go IterateRoutine(func(n string) {
		bucket.Object(n).Delete(ctx)
		fmt.Fprintln(c.Writer, n+" was deleted!")
	}, output.SourceFiles, done)
	<-done
	fmt.Fprintln(c.Writer, "Files were deleted")
}
func GetObject(c *gin.Context) {
	var filename string

	if c.Query("bucket") == "employee" {
		su = StorageUploader{
			bucketName: appconfig.EmployeeBucket,
		}
	}
	if c.Query("bucket") == "default" {
		su = StorageUploader{
			bucketName: appconfig.StorageBucket,
		}
	}
	if c.Query("folder") == "" {
		filename = c.Param("path") + "/"
	} else {
		filename = c.Query("folder") + "/" + c.Param("path")
	}

	file, resp := uploader.ReadImage(filename)
	SendResponseError(c, resp)

	c.JSON(200, file)
}
func ListObjects(c *gin.Context) {
	var files FileObjectsInfo
	su = StorageUploader{
		bucketName: appconfig.StorageBucket,
	}
	items, resp := uploader.List("", c.Query("delimiter"))
	SendResponseError(c, resp)
	if err := json.Unmarshal(items, &files); err != nil {
		fmt.Fprintln(c.Writer, err)
	}
	c.JSON(200, gin.H{
		"folders": files.Folders,
	})
}
func ListObjectsInFolder(c *gin.Context) {
	var prefix string
	var files FileObjectsInfo
	if c.Query("bucket") == "employee" {
		su = StorageUploader{
			bucketName: appconfig.EmployeeBucket,
		}
	} else {
		su = StorageUploader{
			bucketName: appconfig.StorageBucket,
		}
	}
	if c.Query("folder") == "" {
		prefix = c.Param("jobid") + "/"
	} else {
		prefix = c.Query("subfolder")
	}
	e, resp := uploader.List(prefix, c.Query("delimiter"))
	SendResponseError(c, resp)
	if err := json.Unmarshal(e, &files); err != nil {
		SendResponseError(c, Response{Status: http.StatusBadRequest, Error: "There was an error unmarshaling the files."})
		return
	}
	if string(e) == "null" {
		SendResponseError(c, Response{Status: http.StatusNotFound, Error: "No files found."})
		return
	}
	c.JSON(200, gin.H{
		"folders": files.Folders,
		"images":  files.Images,
		"pdfs":    files.Pdfs,
	})
}

func UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		SendResponseError(c, Response{Status: http.StatusBadRequest, Error: err.Error()})
	}
	su = StorageUploader{
		bucketName: appconfig.StorageBucket,
	}
	var res Response
	files := form.File["multiFiles"]
	f, res := uploader.Upload(files, c.PostForm("path"))
	SendResponseError(c, res)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Upload was successful!",
		"files":   f,
	})
}

func UploadToUser(c *gin.Context) {
	f, err := c.FormFile("single")
	if err != nil {
		SendResponseError(c, Response{Status: http.StatusBadRequest, Error: err.Error()})
	}

	su = StorageUploader{
		bucketName: appconfig.EmployeeBucket,
		uploadPath: c.PostForm("user") + "/",
	}

	blobFile, err := f.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	image, resp := uploader.UploadImageInUser(blobFile, f.Filename)
	SendResponseError(c, resp)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "success",
		"imageUrl": image,
	})
}
func UploadCertBadge(c *gin.Context) {
	f, err := c.FormFile("single")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	su = StorageUploader{
		bucketName: appconfig.EmployeeBucket,
		uploadPath: c.PostForm("user") + "/cert-images/",
	}
	blobFile, err := f.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	image, resp := uploader.UploadImageInUser(blobFile, f.Filename)
	SendResponseError(c, resp)
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Certification badge uploaded!",
		"imageUrl":  image,
		"imageName": f.Filename,
	})
}

func FolderCreation(c *gin.Context) {
	su = StorageUploader{
		bucketName: appconfig.StorageBucket,
	}
	jobid := c.Query("jobid")
	folderPath := c.Param("folderPath")
	folderPathArr := strings.Split(folderPath, "/")
	var folderName string
	if folderName = jobid + "/" + strings.Split(folderPath, "/")[len(folderPathArr)-1]; folderPath != "" {
		folderName = jobid
	}
	f, resp := uploader.CreateFolder(folderName)
	SendResponseError(c, resp)
	c.JSON(200, f)
}

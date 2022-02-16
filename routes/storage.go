package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

var Bucket string

type FilesIterator struct {
	pageInfo *iterator.PageInfo
	nextFunc func() error
	items    []string
}

func ExecCommand(com *command) (string, string, *bytes.Buffer) {
	switch com.name {
	case "moving":
		return com.move(), com.respMessage, nil
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

func ListObjects(c *gin.Context) {
	var files FileObjectsInfo
	su = StorageUploader{
		bucketName: appconfig.StorageBucket,
	}
	items, err := uploader.List("", c.Query("delimiter"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	}
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
		prefix = c.Param("path") + "/"
	} else {
		prefix = c.Param("path") + "/" + c.Query("subfolder")
	}
	e, err := uploader.List(prefix, c.Query("delimiter"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	}
	if err := json.Unmarshal(e, &files); err != nil {
		fmt.Fprintln(c.Writer, err)
	}
	c.JSON(200, gin.H{
		"folders": files.Folders,
		"images":  files.Images,
	})
}

func UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	su = StorageUploader{
		bucketName: appconfig.StorageBucket,
	}
	files := form.File["multiFiles"]
	f, err := uploader.Upload(files, c.PostForm("path"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Upload was successful!",
		"files":   f,
	})
}

func UploadAvatar(c *gin.Context) {
	firebaseAuth := c.MustGet("firebaseAuth").(*auth.Client)
	f, err := c.FormFile("single")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
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
	image, err := uploader.UploadImageInUser(blobFile, f.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	user, err := firebaseAuth.GetUserByEmail(context.Background(), c.PostForm("user"))
	if err != nil {
		log.Fatalf("error getting user: %v\n", err)
	}
	params := (&auth.UserToUpdate{}).
		PhotoURL(image)
	_, err = firebaseAuth.UpdateUser(context.Background(), user.UID, params)
	if err != nil {
		log.Fatalf("error updating user: %v\n", err)
	}
	fmt.Printf("Successfully updated user: %v\n", user)
	c.JSON(200, gin.H{
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
	image, err := uploader.UploadImageInUser(blobFile, f.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message":   "Certification badge uploaded!",
		"imageUrl":  image,
		"imageName": f.Filename,
	})
}

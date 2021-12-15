package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gsutil/config"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
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
	case "listing":
		return com.respMessage, "", com.listing()
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

func MovingObjects(w http.ResponseWriter, r *http.Request) {
	env, err := config.LoadConfig("./")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}
	var output FolderInfo
	if err := json.NewDecoder(r.Body).Decode(&output); err != nil {
		log.Fatal(err)
	}
	c := &command{}
	var message string
	for _, item := range output.SourceFiles {
		//item is the prefix from client-side
		c = &command{
			name:        "moving",
			args:        []string{"mv", "gs://" + env.StorageBucket + "/" + item + "*", "gs://" + env.StorageBucket + "/" + output.DestFolder},
			respMessage: "Files successfully moved!",
		}
		_, message, _ = ExecCommand(c)
	}
	fmt.Fprint(w, message)
}
func DeleteObjects(c *gin.Context) {
	ctx := context.Background()
	env, err := config.LoadConfig("./")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	firebaseStorage := c.MustGet("firebaseStorage").(*storage.Client)

	var output FolderInfo
	if err := json.NewDecoder(c.Request.Body).Decode(&output); err != nil {
		fmt.Print(err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	bucket := firebaseStorage.Bucket(env.StorageBucket)
	Iterate(func(n string) {
		bucket.Object(n).Delete(ctx)
		fmt.Fprintln(c.Writer, n+" was deleted!")
	}, output.SourceFiles)
}

func ListObjects(c *gin.Context) {
	var prefix string
	/* s := strings.Split(c.Param("path"), "__")
	jobid, path := s[0], s[1] */
	if c.Query("folder") == "" {
		prefix = c.Param("path") + "/"
	} else {
		prefix = c.Param("path") + "/" + c.Query("subfolder")
	}
	e, err := uploader.List(prefix, c.Query("delimiter"))
	if err != nil {
		fmt.Fprintln(c.Writer, err)
	}

	c.Data(http.StatusOK, gin.MIMEJSON, e)
}

func UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
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

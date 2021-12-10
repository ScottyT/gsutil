package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gsutil/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
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
	ctx := context.Background()
	env, err := config.LoadConfig("./")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	dir, _ := os.Getwd()
	serviceAccountKeyFilePath, err := filepath.Abs(dir + "/" + env.CredentialFile)
	if err != nil {
		panic("Unable to load service account file")
	}
	opt := option.WithCredentialsFile(serviceAccountKeyFilePath)

	cred, err := credentials.NewIamCredentialsClient(ctx, opt)
	if err != nil {
		panic(err)
	}

	var resp FileResponse
	if err := json.NewDecoder(c.Request.Body).Decode(&resp); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()
	//firebaseStorage := c.MustGet("firebaseStorage").(*storage.Client)

	//ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	//defer cancel()
	it := client.Bucket(env.StorageBucket).Objects(ctx, &storage.Query{
		Prefix:    resp.Prefix,
		Delimiter: resp.Delim,
	})
	var folders []FolderObjectsInfo
	var images []ImageObjectsInfo
	var files *FileObjectsInfo
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Bucket(%q).Objects(): %v", env.StorageBucket, err)
		}
		opts := &storage.SignedURLOptions{
			GoogleAccessID: env.SaEmail,
			SignBytes: func(b []byte) ([]byte, error) {
				req := &credentialspb.SignBlobRequest{
					Payload: b,
					Name:    env.SaEmail,
				}
				response, err := cred.SignBlob(ctx, req)
				if err != nil {
					fmt.Fprintf(c.Writer, "There was an error with creds: %v\n", err)
				}
				return response.SignedBlob, err
			},
			Scheme:  storage.SigningSchemeV4,
			Method:  "GET",
			Expires: time.Now().Add(15 * time.Minute),
		}
		url, err := storage.SignedURL(env.StorageBucket, attrs.Name, opts)
		if err != nil {
			fmt.Printf("Bucket(%q).SignedURL: %v", env.StorageBucket, err)
		}
		if strings.Contains(attrs.ContentType, "image") {
			images = append(images, ImageObjectsInfo{
				name:      attrs.Name,
				imageUrl:  url,
				subfolder: c.Param("subfolder"),
			})
		} else {
			queryArray := reverseArray(strings.Split(c.Param("prefix"), "/"))
			folders = append(folders, FolderObjectsInfo{name: queryArray[0], path: c.Param("prefix") + c.Param("delim")})
		}

		files = &FileObjectsInfo{folders: folders, images: images}
	}
	fmt.Fprintln(c.Writer, files)
}

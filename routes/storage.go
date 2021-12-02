package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gsutil/config"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var Bucket string

type response struct {
	data    *bytes.Buffer
	message string
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
func MovingObjects(w http.ResponseWriter, r *http.Request) {
	env, err := config.LoadConfig("./")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}
	var output FolderInfo
	if err := json.NewDecoder(r.Body).Decode(&output); err != nil {
		log.Fatal(err)
	}
	/* if os.Getenv("ENV") == "production" {
		Bucket = os.Getenv("STORAGE_BUCKET")
	} else {
		Bucket = config.ViperEnvKey("STORAGE_BUCKET")
	} */
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

func ListDir(w http.ResponseWriter, r *http.Request) {
	dir, _ := os.Getwd()
	err := filepath.Walk(dir+"/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Fprintf(w, "dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	//list := []FileInfo{}
	/* files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range files {
		f := FileInfo{
			Name:    entry.Name(),
			Size:    entry.Size(),
			Mode:    entry.Mode(),
			ModTime: entry.ModTime(),
			IsDir:   entry.IsDir(),
		}
		list = append(list, f)

	} */
	/* output, err := json.Marshal(list)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(output)) */
}

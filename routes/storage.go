package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gsutil/config"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	var output FolderInfo
	if err := json.NewDecoder(r.Body).Decode(&output); err != nil {
		log.Fatal(err)
	}
	if config.ViperEnvKey("ENV") == "development" {
		Bucket = config.ViperEnvKey("STORAGE_BUCKET")
	} else {
		Bucket = os.Getenv("STORAGE_BUCKET")
	}
	for _, item := range output.SourceFiles {
		fmt.Println("item: ", item)
		//item is the prefix from client-side
		c := &command{
			name:        "moving",
			args:        []string{"-m", "cp", "-r", "gs://" + Bucket + "/" + item, "gs://" + Bucket + "/" + output.DestFolder},
			respMessage: "Files successfully moved!",
		}
		val, message, _ := ExecCommand(c)
		fmt.Fprint(w, val, message)
	}
}

func ListDir(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["folder"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "Folder needs to be specified.", 404)
		return
	}
	folder := keys[0]
	dir, _ := os.Getwd()
	fmt.Print(dir)
	list := []FileInfo{}
	files, err := ioutil.ReadDir("/app/" + folder)
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
		fmt.Println(entry)
		log.Writer()
	}
	output, err := json.Marshal(list)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(output))
}

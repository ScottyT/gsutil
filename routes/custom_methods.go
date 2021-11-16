package routes

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type FolderInfo struct {
	FolderPath  string   `json:"folderPath"`
	SourceFiles []string `json:"sourceFiles"`
	DestFolder  string   `json:"destFolder"`
}
type FilesList struct {
	Files []FilesList `json:"files"`
}
type ImageInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

type command struct {
	name        string
	respMessage string
	args        []string
}
type action struct {
	command
	commandName string
}

func (c *command) download() string {
	//dir, _ := os.Getwd()
	//cmd := exec.CommandContext(r.Context(), "/bin/bash", "script.sh", folder, dir)
	fmt.Println(c.name)
	//cmd := exec.Command("gsutil cp -r gs://" + config.ViperEnvKey("STORAGE_BUCKET") + "/" + string(folder) + " .") //USED FOR LOCAL TESTING
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
	//cmd := exec.CommandContext(r.Context(), "/bin/bash", "move.sh", sourcefolder, destfolder)
	//cmd := exec.Command("gsutil", "-m", "mv", "-p", "gs://code-red-app-313517.appspot.com/"+sourcefolder, "gs://code-red-app-313517.appspot.com/"+destfolder)
	/* cmd := exec.Command("gsutil", "cp -m -p mv gs://code-red-app-313517.appspot.com/"+string(sourcefolder)+
	" gs://code-red-app-313517.appspot.com/"+string(destfolder)) //USED FOR LOCAL TESTING */
	cmd := exec.Command("gsutil", c.args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	/* if err := cmd.Run(); err != nil {
		log.Printf("Command.Output: %v", err)
	} */
	if err != nil {
		return err.Error()
	}
	return bytes.NewBuffer(out).String()
	//fmt.Println("Moving: ", gsutilCmd)
}

func (c *command) listing() *bytes.Buffer {
	fmt.Println(c.args)
	cmd := exec.Command("gsutil", c.args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("Command.Output: " + err.Error())

		//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	list := bytes.NewBuffer(out)
	return list
	/* if err := cmd.Run(); err != nil {
		log.Printf("Command.Output: %v", err)
	}

	/* ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err), ""
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	it := client.Bucket(bucket).Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: delim,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Bucket(%q).Objects(): %v", bucket, err), ""
		}
		fmt.Println(attrs.Name)
		return nil, attrs.Name
	}
	return nil, "" */
}

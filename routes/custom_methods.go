package routes

import (
	"bytes"
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

var cmd *exec.Cmd

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
	/* if err := cmd.Run(); err != nil {
		log.Printf("Command.Output: %v", err)
	} */
	if err != nil {
		return err.Error()
	}
	return bytes.NewBuffer(out).String()
}

func (c *command) listing() *bytes.Buffer {
	cmd := exec.Command("gsutil", c.args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("Command.Output: " + err.Error())

		//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	list := bytes.NewBuffer(out)
	return list
}

package routes

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

type FolderInfo struct {
	FolderPath  string   `json:"folderPath"`
	SourceFiles []string `json:"sourceFiles"`
	DestFolder  string   `json:"destFolder"`
}
type FileResponse struct {
	Prefix string `json:"prefix"`
	Delim  string `json:"delim"`
}
type FilesList struct {
	Files []FilesList `json:"files"`
}
type ImageObjectsInfo struct {
	name      string
	imageUrl  string
	subfolder string
}
type FolderObjectsInfo struct {
	name string
	path string
}
type FileObjectsInfo struct {
	folders []FolderObjectsInfo
	images  []ImageObjectsInfo
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

func reverseArray(arr []string) []string {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

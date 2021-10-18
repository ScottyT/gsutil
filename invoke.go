package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"gsutil/config"
	"gsutil/middleware"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

func main() {
	r := gin.Default()
	firebaseAuth := config.SetupFirebase()
	r.Use(func(c *gin.Context) {
		// set firebase auth
		c.Set("firebaseAuth", firebaseAuth)
	})
	r.Use(middleware.AuthMiddleware)
	r.POST("/zip", gin.WrapF(scriptHandler))
	r.GET("/list", gin.WrapF(listDir))
	/* http.HandleFunc("/zip", scriptHandler)
	http.HandleFunc("/list", listDir) */
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" //Change this when testing locally
		fmt.Printf("defaulting to port %s", port)
	}
	r.Run(":" + port)
	//log.Fatal(http.ListenAndServe(port, r))
}

func scriptHandler(w http.ResponseWriter, r *http.Request) {
	output := "output"
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", output))
	keys, ok := r.URL.Query()["folder"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "Folder needs to be specified.", 404)
	}
	folder := keys[0]
	dir, _ := os.Getwd()
	cmd := exec.CommandContext(r.Context(), "/bin/bash", "script.sh", folder, dir)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Command.Output: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	fmt.Println(out)

	files, err := ioutil.ReadDir("/app/" + folder)
	if err != nil {
		log.Fatal(err)
	}
	if err := ZipFiles(output, folder, files); err != nil {
		log.Fatal(err)
	}
	http.ServeFile(w, r, output)
	fmt.Fprintf(w, "Done!")
}
func listFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("Crawling: %#v\n", path)

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil

}

func zipMe(filepaths []string, target string) error {

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	file, err := os.OpenFile(target, flags, 0644)

	if err != nil {
		return fmt.Errorf("Failed to open zip for writing: %s", err)
	}
	defer file.Close()

	zipw := zip.NewWriter(file)
	defer zipw.Close()

	for _, filename := range filepaths {
		if err := addFileToZip(filename, zipw); err != nil {
			return fmt.Errorf("Failed to add file %s to zip: %s", filename, err)
		}
	}
	return nil

}

func addFileToZip(filename string, zipw *zip.Writer) error {
	file, err := os.Open(filename)

	if err != nil {
		return fmt.Errorf("Error opening file %s: %s", filename, err)
	}
	defer file.Close()

	wr, err := zipw.Create(filename)
	if err != nil {

		return fmt.Errorf("Error adding file; '%s' to zip : %s", filename, err)
	}

	if _, err := io.Copy(wr, file); err != nil {
		return fmt.Errorf("Error writing %s to zip: %s", filename, err)
	}

	return nil
}
func ZipFiles(filename string, foldername string, files []fs.FileInfo) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	if err != nil {
		return err
	}
	if err = zipSource(zipWriter, "/app/"+foldername, "output"); err != nil {
		return err
	}
	return nil
}

func zipSource(w *zip.Writer, source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("Crawling: %#v\n", path)
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}
func addFilesToDir(w *zip.Writer, directoryPath, zipName string) error {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		path := directoryPath + file.Name()
		fmt.Print("path: ", path)

		if err != nil {
			log.Fatal(err)
		}
		if !file.IsDir() {
			fmt.Print("Name is file")
			dat, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v. ", err)
				os.Exit(1)
			}
			fw, err := w.Create(zipName + file.Name())
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("file: ", file.Name())
			if _, err := fw.Write(dat); err != nil {
				fmt.Println(err)
			}
		} else if file.IsDir() {
			newBase := directoryPath + file.Name() + "/"
			fmt.Println("Recursing and Adding SubDir: " + file.Name())
			fmt.Println("Recursing and Adding SubDir: " + newBase)
			addFilesToDir(w, newBase, zipName+file.Name()+"/")
		}
	}
	return nil
}
func listDir(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

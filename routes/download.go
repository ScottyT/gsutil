package routes

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	var output FolderInfo
	if err := json.NewDecoder(r.Body).Decode(&output); err != nil {
		log.Fatal(err)
	}
	zipFolderName := "job_" + output.FolderPath + "_files.zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", string(zipFolderName)))

	folder := output.FolderPath
	dir, _ := os.Getwd()

	c := &command{
		name:        "download",
		args:        []string{"-m", "cp", "-r", "gs://" + su.bucketName + "/" + folder, dir},
		respMessage: "Files downloaded!",
	}
	val, message, _ := ExecCommand(c)

	files, err := ioutil.ReadDir(dir + "/" + folder)
	if err != nil {
		log.Fatal(err)
	}
	if err := ZipFiles(zipFolderName, folder, files); err != nil {
		log.Fatal(err)
	}
	http.ServeFile(w, r, zipFolderName)
	fmt.Fprint(w, val, message)
}
func ZipFiles(filename string, foldername string, files []fs.FileInfo) error {
	newZipFile, err := os.Create(filename)
	dir, _ := os.Getwd()
	if err != nil {
		return err
	}
	defer newZipFile.Close()
	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	if err != nil {
		return err
	}
	if err = zipSource(zipWriter, dir+"/"+foldername, filename); err != nil {
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
		//fmt.Printf("Crawling: %#v\n", path)
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

package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const MaxUploadSize = 20 * 1024 * 1024 // 20 mb
const MaxAge = 5 * time.Minute
const UploadPath = "./tmp"

var cache = NewCache(time.Second)

func main() {
	if err := os.MkdirAll(UploadPath, os.ModePerm); err != nil {
		log.Printf("Failed to mkdir: %s", err)
		return
	}

	cache.BeforeDeleteCallback = func(file *entry) {
		err := os.Remove(file.filepath)
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Printf("Deleted file: %s (%s)", file.filepath, file.filename)
	}
	port := getEnv("PORT", "3006")

	http.HandleFunc("/upload", uploadFileHandler())
	http.HandleFunc("/download", downloadFileHandler())

	//fs := http.FileServer(http.Dir(UploadPath))
	//http.Handle("/files/", http.StripPrefix("/files", fs))

	log.Printf("Server started on localhost:%s, use /upload for uploading files and /download?file={fileName} for downloading", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func downloadFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileId := r.URL.Query().Get("file")
		if fileId == "" {
			renderError(w, "file query param is empty", http.StatusBadRequest)
			return
		}
		f := cache.Get(fileId)
		if f == nil {
			renderError(w, "FILE_NOT_FOUND", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+f.filename)
		http.ServeFile(w, r, f.filepath)
	}
}

func uploadFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
		if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
			log.Printf("FILE_TOO_BIG: %s", err)
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		file, header, err := r.FormFile("uploadFile")
		if err != nil {
			log.Printf("INVALID_FILE: %s", err)
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("READ_FAILED: %s", err)
			renderError(w, "READ_FAILED", http.StatusBadRequest)
			return
		}

		fileId := randToken(12)
		fileExtension := filepath.Ext(header.Filename)
		newFilename := fileId + fileExtension
		newPath := filepath.Join(UploadPath, newFilename)
		fmt.Printf("FileType: %s, File: %s\n", fileExtension, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			log.Printf("CANT_WRITE_FILE: %s", err)
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			log.Printf("CANT_WRITE_FILE: %s", err)
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		cache.Add(newFilename, newPath, header.Filename, MaxAge)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"file": "%s"}`, newFilename)
	}
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "%s"}`, message)
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func getEnv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

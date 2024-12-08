package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Video to Audio</title>
		</head>
		<body>
			<h1>Convert video to audio</h1>
			<form enctype="multipart/form-data" action="/upload" method="post">
				<input type="file" name="video" accept="video/*" required>
				<button type="submit">Upload</button>
			</form>
		</body>
		</html>`
		w.Write([]byte(html))
	})

	r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20) // limit file to 10MB
		file, handler, err := r.FormFile("video")
		if err != nil {
			http.Error(w, "Error to get file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Save video in disk temporarily
		tempDir := os.TempDir()
		videoPath := filepath.Join(tempDir, handler.Filename)
		videoFile, err := os.Create(videoPath)
		if err != nil {
			http.Error(w, "Error to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(videoPath) // Clean after save
		defer videoFile.Close()

		if _, err := io.Copy(videoFile, file); err != nil {
			http.Error(w, "Error to copy file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Extract the audio wih the FFmpeg
		audioPath := filepath.Join(tempDir, "output_audio.mp3")
		cmd := exec.Command("ffmpeg", "-i", videoPath, "-q:a", "0", "-map", "a", audioPath)
		if err := cmd.Run(); err != nil {
			http.Error(w, "Erro ao processar o vídeo com FFmpeg: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(audioPath) // Clean after use

		// Return audio to user
		w.Header().Set("Content-Disposition", "attachment; filename=output_audio.mp3")
		w.Header().Set("Content-Type", "audio/mpeg")
		audioFile, err := os.Open(audioPath)
		if err != nil {
			http.Error(w, "Error to open processed audio: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer audioFile.Close()
		io.Copy(w, audioFile)
	})

	fmt.Println("Runinng in http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

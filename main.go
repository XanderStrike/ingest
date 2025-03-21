package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	uploadPath    = "./uploads"
	templatesPath = "./templates"
	staticPath    = "./templates"  // Using templates dir for static files too
	defaultMaxSize = 0             // 0 means unlimited
)

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Create templates directory if it doesn't exist
	if err := os.MkdirAll(templatesPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create templates directory: %v", err)
	}

	// Define routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadPath))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Get max file size from environment variable
	maxSizeStr := getEnv("MAX_FILE_SIZE", "0")
	maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
	if err != nil {
		log.Printf("Invalid MAX_FILE_SIZE value: %s, using unlimited", maxSizeStr)
		maxSize = defaultMaxSize
	}
	
	if maxSize > 0 {
		log.Printf("Maximum file size set to %s", formatBytes(maxSize))
	} else {
		log.Printf("Maximum file size: unlimited")
	}

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Ensure templates directory exists
	if err := os.MkdirAll(templatesPath, os.ModePerm); err != nil {
		log.Printf("Failed to create templates directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Read the template file
	templatePath := filepath.Join(templatesPath, "index.html")
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Printf("Failed to read template file: %v", err)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	// Serve the template
	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get max file size from environment variable
	maxSizeStr := getEnv("MAX_FILE_SIZE", "0")
	maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
	if err != nil {
		log.Printf("Invalid MAX_FILE_SIZE value: %s, using unlimited", maxSizeStr)
		maxSize = defaultMaxSize
	}

	// If maxSize is set, limit the request body size
	if maxSize > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	}

	// Get file from request
	file, handler, err := r.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			log.Printf("File too large")
			http.Error(w, fmt.Sprintf("File too large. Maximum size is %s", formatBytes(maxSize)), http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create file on server
	dst, err := os.Create(filepath.Join(uploadPath, handler.Filename))
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, file); err != nil {
		if err.Error() == "http: request body too large" {
			os.Remove(filepath.Join(uploadPath, handler.Filename))
			http.Error(w, fmt.Sprintf("File too large. Maximum size is %s", formatBytes(maxSize)), http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Return success status (for XHR requests)
	w.WriteHeader(http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal attacks by cleaning the filename
	// and ensuring it doesn't contain path separators
	cleanFilename := filepath.Base(filename)

	// Create the full path to the file using absolute paths
	absUploadPath, err := filepath.Abs(uploadPath)
	if err != nil {
		log.Printf("Error getting absolute path: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(absUploadPath, cleanFilename)

	// Log the file path for debugging
	log.Printf("Attempting to delete file: %s", filePath)

	// Check if file exists and is accessible
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("File not found: %s", filePath)
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			log.Printf("Error accessing file: %v", err)
			http.Error(w, "Error accessing file", http.StatusInternalServerError)
		}
		return
	}

	// Verify it's a regular file, not a directory
	if fileInfo.IsDir() {
		log.Printf("Not a file: %s", filePath)
		http.Error(w, "Not a file", http.StatusBadRequest)
		return
	}

	// Try to open the file first to verify permissions
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		log.Printf("Cannot open file (permission issue?): %v", err)
		http.Error(w, "Permission denied", http.StatusInternalServerError)
		return
	}
	file.Close()

	// Delete the file
	err = os.Remove(filePath)
	if err != nil {
		log.Printf("Error deleting file: %v", err)
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted file: %s", cleanFilename)
	// Return success status
	w.WriteHeader(http.StatusOK)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// formatBytes converts bytes to a human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

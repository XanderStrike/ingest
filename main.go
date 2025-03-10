package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	uploadPath    = "./uploads"
	maxUploadSize = 10 << 20 // 10 MB
)

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Define routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadPath))))

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

	html := `
<!DOCTYPE html>
<html>
<head>
    <title>File Upload Service</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        h1 {
            color: #333;
        }
        form {
            margin: 20px 0;
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 5px;
        }
        input[type="file"] {
            margin: 10px 0;
        }
        input[type="submit"] {
            background-color: #4CAF50;
            color: white;
            padding: 10px 15px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        input[type="submit"]:hover {
            background-color: #45a049;
        }
        .files {
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <h1>File Upload Service</h1>
    <form action="/upload" method="post" enctype="multipart/form-data">
        <h3>Select a file to upload:</h3>
        <input type="file" name="file" required>
        <br>
        <input type="submit" value="Upload">
    </form>
    <div class="files">
        <h3>Uploaded Files:</h3>
        <ul id="fileList">
            <!-- Files will be listed here -->
        </ul>
    </div>

    <script>
        // Fetch and display uploaded files
        fetch('/uploads/')
            .then(response => response.text())
            .then(html => {
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const links = doc.querySelectorAll('a');
                const fileList = document.getElementById('fileList');
                
                links.forEach(link => {
                    if (link.href.indexOf('?') === -1) { // Skip parent directory link
                        const li = document.createElement('li');
                        const a = document.createElement('a');
                        a.href = link.href;
                        a.textContent = link.textContent;
                        li.appendChild(a);
                        fileList.appendChild(li);
                    }
                });
            })
            .catch(error => console.error('Error fetching files:', error));
    </script>
</body>
</html>
`
	fmt.Fprint(w, html)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get file from request
	file, handler, err := r.FormFile("file")
	if err != nil {
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
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Redirect back to index
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

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
    <title>Ingest</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .drop-area {
            margin: 20px 0;
            padding: 40px;
            border: 3px dashed #ccc;
            border-radius: 10px;
            text-align: center;
            transition: all 0.3s ease;
            background-color: #f9f9f9;
            cursor: pointer;
        }
        .drop-area.highlight {
            border-color: #4CAF50;
            background-color: #e8f5e9;
        }
        .drop-area p {
            font-size: 18px;
            margin-bottom: 10px;
        }
        .drop-area .icon {
            font-size: 48px;
            color: #999;
            margin-bottom: 10px;
        }
        .upload-list {
            margin-top: 30px;
        }
        .upload-item {
            margin-bottom: 15px;
            padding: 15px;
            border: 1px solid #ddd;
            border-radius: 5px;
            background-color: #f5f5f5;
        }
        .progress-bar {
            height: 10px;
            background-color: #e0e0e0;
            border-radius: 5px;
            margin: 10px 0;
            overflow: hidden;
        }
        .progress-bar-fill {
            height: 100%;
            background-color: #4CAF50;
            width: 0%;
            transition: width 0.3s ease;
        }
        .upload-details {
            display: flex;
            justify-content: space-between;
            font-size: 14px;
            color: #666;
        }
        .files {
            margin-top: 30px;
        }
        .files h3 {
            border-bottom: 1px solid #eee;
            padding-bottom: 10px;
        }
        .files ul {
            list-style-type: none;
            padding: 0;
        }
        .files li {
            padding: 8px 0;
            border-bottom: 1px solid #f5f5f5;
        }
        .files a {
            color: #2196F3;
            text-decoration: none;
        }
        .files a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <h1>Ingest</h1>
    
    <div class="drop-area" id="dropArea">
        <div class="icon">📁</div>
        <p>Drag & drop files here</p>
        <p>or click to select files</p>
    </div>
    
    <div class="upload-list" id="uploadList">
        <!-- Upload progress items will appear here -->
    </div>
    
    <div class="files">
        <h3>Uploaded Files</h3>
        <ul id="fileList">
            <!-- Files will be listed here -->
        </ul>
    </div>

    <script>
        // DOM elements
        const dropArea = document.getElementById('dropArea');
        const uploadList = document.getElementById('uploadList');
        const fileList = document.getElementById('fileList');
        
        // Prevent default drag behaviors
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            dropArea.addEventListener(eventName, preventDefaults, false);
            document.body.addEventListener(eventName, preventDefaults, false);
        });
        
        // Highlight drop area when item is dragged over it
        ['dragenter', 'dragover'].forEach(eventName => {
            dropArea.addEventListener(eventName, highlight, false);
        });
        
        ['dragleave', 'drop'].forEach(eventName => {
            dropArea.addEventListener(eventName, unhighlight, false);
        });
        
        // Handle dropped files
        dropArea.addEventListener('drop', handleDrop, false);
        
        // Handle click to select files
        dropArea.addEventListener('click', () => {
            const input = document.createElement('input');
            input.type = 'file';
            input.multiple = true;
            input.onchange = e => {
                const files = e.target.files;
                handleFiles(files);
            };
            input.click();
        });
        
        function preventDefaults(e) {
            e.preventDefault();
            e.stopPropagation();
        }
        
        function highlight() {
            dropArea.classList.add('highlight');
        }
        
        function unhighlight() {
            dropArea.classList.remove('highlight');
        }
        
        function handleDrop(e) {
            const dt = e.dataTransfer;
            const files = dt.files;
            handleFiles(files);
        }
        
        function handleFiles(files) {
            [...files].forEach(uploadFile);
        }
        
        function uploadFile(file) {
            // Create upload item element
            const uploadItem = document.createElement('div');
            uploadItem.className = 'upload-item';
            uploadItem.innerHTML = `
                <div class="filename">${file.name} (${formatFileSize(file.size)})</div>
                <div class="progress-bar">
                    <div class="progress-bar-fill"></div>
                </div>
                <div class="upload-details">
                    <span class="status">Preparing...</span>
                    <span class="speed">0 KB/s</span>
                </div>
            `;
            uploadList.appendChild(uploadItem);
            
            const progressBar = uploadItem.querySelector('.progress-bar-fill');
            const status = uploadItem.querySelector('.status');
            const speedElement = uploadItem.querySelector('.speed');
            
            // Create FormData
            const formData = new FormData();
            formData.append('file', file);
            
            // Upload variables for speed calculation
            let startTime = Date.now();
            let lastLoaded = 0;
            let lastTime = startTime;
            
            // Create and send XHR request
            const xhr = new XMLHttpRequest();
            xhr.open('POST', '/upload');
            
            xhr.upload.addEventListener('progress', e => {
                if (e.lengthComputable) {
                    // Calculate progress percentage
                    const percent = (e.loaded / e.total) * 100;
                    progressBar.style.width = percent + '%';
                    
                    // Calculate upload speed
                    const currentTime = Date.now();
                    const elapsedTime = (currentTime - lastTime) / 1000; // seconds
                    
                    if (elapsedTime > 0.2) { // Update speed every 200ms
                        const loadDifference = e.loaded - lastLoaded;
                        const speed = loadDifference / elapsedTime; // bytes per second
                        
                        speedElement.textContent = formatFileSize(speed) + '/s';
                        
                        lastLoaded = e.loaded;
                        lastTime = currentTime;
                    }
                    
                    // Update status text
                    status.textContent = \`\${Math.round(percent)}% uploaded\`;
                }
            });
            
            xhr.addEventListener('load', () => {
                if (xhr.status === 200 || xhr.status === 303) {
                    status.textContent = 'Upload complete';
                    speedElement.textContent = '';
                    
                    // Refresh file list
                    refreshFileList();
                    
                    // Remove upload item after a delay
                    setTimeout(() => {
                        uploadItem.style.opacity = '0';
                        setTimeout(() => {
                            uploadList.removeChild(uploadItem);
                        }, 300);
                    }, 3000);
                } else {
                    status.textContent = 'Upload failed';
                    console.error('Upload failed:', xhr.statusText);
                }
            });
            
            xhr.addEventListener('error', () => {
                status.textContent = 'Upload failed';
                console.error('Upload failed');
            });
            
            xhr.send(formData);
        }
        
        function formatFileSize(bytes) {
            if (bytes === 0) return '0 Bytes';
            
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }
        
        function refreshFileList() {
            // Clear current list
            fileList.innerHTML = '';
            
            // Fetch and display uploaded files
            fetch('/uploads/')
                .then(response => response.text())
                .then(html => {
                    const parser = new DOMParser();
                    const doc = parser.parseFromString(html, 'text/html');
                    const links = doc.querySelectorAll('a');
                    
                    links.forEach(link => {
                        if (link.href.indexOf('?') === -1 && !link.textContent.includes('Parent Directory')) {
                            const li = document.createElement('li');
                            const a = document.createElement('a');
                            a.href = link.href;
                            a.textContent = link.textContent;
                            li.appendChild(a);
                            fileList.appendChild(li);
                        }
                    });
                    
                    if (fileList.children.length === 0) {
                        const li = document.createElement('li');
                        li.textContent = 'No files uploaded yet';
                        fileList.appendChild(li);
                    }
                })
                .catch(error => console.error('Error fetching files:', error));
        }
        
        // Initial file list load
        refreshFileList();
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

	// Return success status (for XHR requests)
	w.WriteHeader(http.StatusOK)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

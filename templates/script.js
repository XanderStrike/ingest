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
            status.textContent = `${Math.round(percent)}% uploaded`;
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
        } else if (xhr.status === 413) {
            status.textContent = 'File too large';
            speedElement.textContent = '';
            console.error('Upload failed: File too large');
            
            // Show error message
            try {
                const response = xhr.responseText;
                alert(response || 'File too large');
            } catch (e) {
                alert('File too large');
            }
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

function addFileToList(fileName, fileUrl, fileSize) {
    const li = document.createElement('li');
    
    // File info (name and link)
    const fileInfo = document.createElement('div');
    fileInfo.className = 'file-info';
    
    const a = document.createElement('a');
    a.href = fileUrl;
    a.textContent = fileName;
    fileInfo.appendChild(a);
    
    // Add file size
    const sizeSpan = document.createElement('span');
    sizeSpan.className = 'file-size';
    sizeSpan.textContent = typeof fileSize === 'number' ? formatFileSize(fileSize) : fileSize;
    fileInfo.appendChild(sizeSpan);
    
    li.appendChild(fileInfo);
    
    // Delete button
    const deleteBtn = document.createElement('button');
    deleteBtn.className = 'delete-btn';
    deleteBtn.textContent = 'Delete';
    deleteBtn.onclick = () => deleteFile(fileName, li);
    li.appendChild(deleteBtn);
    
    fileList.appendChild(li);
}

function deleteFile(fileName, listItem) {
    if (!confirm(`Are you sure you want to delete ${fileName}?`)) {
        return;
    }
    
    console.log(`Attempting to delete file: ${fileName}`);
    
    const formData = new FormData();
    formData.append('filename', fileName);
    
    fetch('/delete', {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (response.ok) {
            console.log(`File deleted successfully: ${fileName}`);
            // Remove the item from the list
            listItem.remove();
            
            // If no files left, show message
            if (fileList.children.length === 0) {
                const li = document.createElement('li');
                li.textContent = 'No files uploaded yet';
                fileList.appendChild(li);
            }
        } else {
            response.text().then(text => {
                console.error(`Delete failed: ${text}`);
                alert(`Failed to delete file: ${text}`);
            }).catch(err => {
                console.error('Error parsing response:', err);
                alert('Failed to delete file');
            });
        }
    })
    .catch(error => {
        console.error('Error deleting file:', error);
        alert('Error deleting file');
    });
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
            
            // Process links from HTML
            let hasFiles = false;
            links.forEach(link => {
                if (link.href.indexOf('?') === -1 && !link.textContent.includes('Parent Directory')) {
                    hasFiles = true;
                    const fileName = link.textContent.trim();
                    
                    // Get file size
                    fetch(`/uploads/${fileName}`, { method: 'HEAD' })
                        .then(response => {
                            const contentLength = response.headers.get('content-length');
                            const size = contentLength ? parseInt(contentLength, 10) : 'Unknown';
                            addFileToList(fileName, link.href, size);
                        })
                        .catch(() => {
                            addFileToList(fileName, link.href, 'Unknown');
                        });
                }
            });
            
            if (!hasFiles) {
                const li = document.createElement('li');
                li.textContent = 'No files uploaded yet';
                fileList.appendChild(li);
            }
        })
        .catch(error => console.error('Error fetching files:', error));
}

// Initial file list load
document.addEventListener('DOMContentLoaded', () => {
    refreshFileList();
});

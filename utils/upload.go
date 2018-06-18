package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	// TokenLength is the length of an upload token in bytes
	TokenLength = 8
	// MaxFileSize is the maximum allowed upload file size
	MaxFileSize = 5242880
	// TempDirectory is the name of the directory for temporary files
	TempDirectory = "tmp"
)

// RequestUtil wraps request so that multipart form data
// can be read with automatic progress updates
type RequestUtil struct {
	R *http.Request
}

// ImageProcessRequest contains fields necessary for image processing
type ImageProcessRequest struct {
	File          *os.File
	OriginalName  string
	RequestedName string
	MimeType      string
	Success       func(*ImageProcessRequest)
	Error         func(*ImageProcessRequest)
}

// UploadFile contains information about a single file being uploaded
type UploadFile struct {
	BytesRead int    `json:"bytes_read"`
	Status    string `json:"status"`
}

// UploadProgress keeps track of a set of files being uploaded
type UploadProgress struct {
	Files    map[string]UploadFile `json:"files"`
	EchoSelf string                `json:"echo_self"`
}

var activeUploads map[string]UploadProgress
var activeUploadMutex sync.Mutex

func init() {
	activeUploads = make(map[string]UploadProgress)
	removeTemporaryFiles()
}

func removeTemporaryFiles() {
	files, err := ioutil.ReadDir(TempDirectory)
	if err != nil {
		return
	}
	for _, file := range files {
		os.Remove(TempDirectory + string(os.PathSeparator) + file.Name())
	}
}

// GetUploadProgress attempts to get the upload progress for an
// upload
func GetUploadProgress(token string) (*UploadProgress, error) {
	activeUploadMutex.Lock()
	defer activeUploadMutex.Unlock()

	uploadProgress, ok := activeUploads[token]
	if !ok {
		return nil, errors.New("Token is invalid")
	}
	return &uploadProgress, nil
}

// AttemptSkipMultipart reads up to 50mb of a multipart request
// in an attempt to get to the end of a chunked request
func (r *RequestUtil) AttemptSkipMultipart() {
	numBytes := 0
	maxBytes := 52428800 // 50MB Maximum
	mr, err := r.R.MultipartReader()
	if err != nil {
		return
	}
	buff := make([]byte, 4096)
	for {
		part, err := mr.NextPart()
		if err != nil {
			return
		}
		for {
			n, err := part.Read(buff)
			numBytes += n
			if numBytes > maxBytes {
				return
			}
			if err == io.EOF {
				break
			}
		}
	}
}

func attemptFinishRead(mr *multipart.Reader, part *multipart.Part) {
	numBytes := 0
	maxBytes := 52428800 // 50MB Maximum
	buff := make([]byte, 4096)
	for {
		for {
			n, err := part.Read(buff)
			numBytes += n
			if numBytes > maxBytes {
				return
			}
			if err == io.EOF {
				break
			}
		}

		var err error
		part, err = mr.NextPart()
		if err != nil {
			return
		}
	}
}

// MultipartProgressReader pulls image files from a multipart request, places
// them into a temporary file, and passes them along to the image processor
func (r *RequestUtil) MultipartProgressReader(token string, echoSelf string,
	ch chan *ImageProcessRequest, maxFileCount int) error {

	mr, err := r.R.MultipartReader()
	if err != nil {
		return err
	}

	activeUploadMutex.Lock()
	if val, ok := activeUploads[token]; ok {
		if strings.Compare(echoSelf, val.EchoSelf) == 0 {
			return errors.New("Token is taken")
		}
		token = ""
	} else {
		activeUploads[token] = UploadProgress{
			Files:    make(map[string]UploadFile),
			EchoSelf: echoSelf,
		}
	}
	activeUploadMutex.Unlock()

	numFiles := 0

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}

		if part.FileName() == "" {
			continue // This is probably not a file
		}

		numFiles++
		if numFiles > maxFileCount {
			attemptFinishRead(mr, part)
			break
		}

		err = processPart(token, part, ch)
		if err != nil {
			fmt.Println("Error processing part: " + err.Error())
		}
	}

	close(ch)

	activeUploadMutex.Lock()
	delete(activeUploads, token)
	activeUploadMutex.Unlock()
	return nil
}

func processPart(token string, part *multipart.Part,
	ch chan *ImageProcessRequest) error {

	totalCount := 0

	mimeType := part.Header.Get("Content-Type")

	tempFile, err := ioutil.TempFile(TempDirectory, "upload")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if _, ok := supportedMimeTypes[mimeType]; !ok {
		return errors.New("Unsupported mime type for image uploads: " + mimeType)
	}

	// This loop handles the upload, and tracks progress
	buffer := make([]byte, 4096)
	isEof := false
	for !isEof {
		numRead, err := part.Read(buffer)

		if err != nil {
			if err == io.EOF {
				if numRead == 0 {
					break
				}
				isEof = true
			} else {
				tempFile.Close()
				os.Remove(tempFile.Name())
				return errors.New("Failed to read file: " + err.Error())
			}
		}
		totalCount += numRead

		if totalCount > MaxFileSize {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return errors.New("File too large")
		}

		_, err = tempFile.Write(buffer[:numRead])
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return err
		}

		activeUploadMutex.Lock()
		uploadProgress, ok := activeUploads[token]
		if ok {
			uploadProgress.Files[part.FileName()] = UploadFile{
				BytesRead: totalCount,
				Status:    "Uploading",
			}
		}
		activeUploadMutex.Unlock()
	}

	activeUploadMutex.Lock()
	uploadProgress, ok := activeUploads[token]
	if ok {
		uploadProgress.Files[part.FileName()] = UploadFile{
			BytesRead: totalCount,
			Status:    "Processing",
		}
	}
	activeUploadMutex.Unlock()

	ch <- &ImageProcessRequest{
		File:         tempFile,
		OriginalName: part.FileName(),
		MimeType:     mimeType,
	}
	return nil
}

package controllers

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/anishmgoyal/calagora/constants"
	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/utils"
	"github.com/anishmgoyal/calagora/wsock"
)

type imageProcessSuccessNotification struct {
	OriginalName string        `json:"name"`
	ID           int           `json:"id"`
	Image        *models.Image `json:"image"`
}

type imageProcessErrorNotification struct {
	OriginalName string `json:"name"`
	Media        string `json:"media"`
	MediaID      int    `json:"media_id"`
}

type uploadToken struct {
	Token string `json:"token"`
}

// Upload handles the route '/upload'
func Upload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		postUpload(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

type uploadPair struct {
	Image        models.Image `json:"image"`
	ID           int          `json:"id"`
	OriginalName string       `json:"name"`
}

type postUploadResponse struct {
	Successful   bool         `json:"successful"`
	Error        string       `json:"error,omitempty"`
	Images       []uploadPair `json:"images"`
	FailedImages []string     `json:"failed_images"`
}

// Errors will not properly be rendered with most requests.
// They could be by simply waiting till the end of the request... but
// that opens up the possibility of a DDOS. SO, ignore all requests that
// are invalid. Clients will freak out but validation is done on the client
// side, so if an error pops up here, the client is doing something bad.
func postUpload(w http.ResponseWriter, r *http.Request) {
	var token string
	var echoSelf string

	viewData := BaseViewData(w, r)
	ru := utils.RequestUtil{R: r}

	response := postUploadResponse{
		Successful:   false,
		Images:       make([]uploadPair, 0, 8),
		FailedImages: make([]string, 0, 8),
	}
	if viewData.Session == nil {
		response.Error = "Can't Upload Without Logging In"
		ru.AttemptSkipMultipart()
		RenderTextJSON(w, response)
		return
	}

	args := URIArgs(r)
	if len(args) > 2 {
		token = args[2]
		if len(args) > 3 {
			echoSelf = args[3]
		}
	} else {
		response.Error = "Invalid Request Arguments"
		ru.AttemptSkipMultipart()
		RenderTextJSON(w, response)
		return
	}

	csrfToken := args[1]
	compareToken := strings.Replace(viewData.Session.CsrfToken, "/", "_", -1)
	if strings.Compare(csrfToken, compareToken) != 0 {
		response.Error = "CSRF Token is Invalid"
		ru.AttemptSkipMultipart()
		RenderTextJSON(w, response)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error = err.Error()
		ru.AttemptSkipMultipart()
		RenderTextJSON(w, response)
		return
	}

	listing, err := models.GetListingByID(Base.Db, id)
	if err != nil || listing == nil {
		response.Error = "Couldn't Find Listing"
		ru.AttemptSkipMultipart()
		RenderTextJSON(w, response)
		return
	}

	imageCount, err := listing.GetImageCount(Base.Db)
	if err != nil {
		response.Error = "Couldn't Get Image Count For Listing"
		ru.AttemptSkipMultipart()
		RenderTextJSON(w, response)
		return
	}

	remainingImageCount := models.MaxListingImages - imageCount
	if remainingImageCount <= 0 {
		response.Error = "You can't upload more than 8 images per listing"
		ru.AttemptSkipMultipart()
		RenderTextJSON(w, response)
		return
	}

	ch := make(chan *utils.ImageProcessRequest, 8)
	ru.MultipartProgressReader(token, echoSelf, ch, remainingImageCount)
	for ipr, more := <-ch; more; ipr, more = <-ch {
		image := models.Image{
			Media:   models.MediaListing,
			MediaID: id,
			Ordinal: 0,
			User:    viewData.Session.User,
		}
		if ok, _ := image.Create(Base.Db); !ok {
			response.FailedImages = append(response.FailedImages, ipr.OriginalName)
			ipr.File.Close()
			os.Remove(ipr.File.Name())
		} else {
			response.Successful = true
			response.Images = append(response.Images, uploadPair{
				Image:        image,
				ID:           image.ID,
				OriginalName: ipr.OriginalName,
			})

			ipr.RequestedName = strconv.Itoa(image.ID) + "_" + image.Media + "_" +
				strconv.Itoa(image.MediaID)

			// This is called if an image is successfully uploaded and saved
			ipr.Success = func(ipr *utils.ImageProcessRequest) {
				if constants.DoUploadAWS {
					image.URL = constants.S3URLStub + ipr.RequestedName
				} else {
					// For testing, create a relative URL instead of an absolute url
					image.URL = "/local/" + ipr.RequestedName
				}
				ok, _ := image.Save(Base.Db)

				if !ok {
					Base.WebsockChannel <- wsock.UserJSONNotification(
						&viewData.Session.User,
						"IM_PROCESS_FAILED", imageProcessErrorNotification{
							OriginalName: ipr.OriginalName,
							Media:        image.Media,
							MediaID:      image.MediaID,
						}, false)

					return
				}

				// Let the client know we're done
				Base.WebsockChannel <- wsock.UserJSONNotification(
					&viewData.Session.User,
					"IM_PROCESS_DONE", imageProcessSuccessNotification{
						OriginalName: ipr.OriginalName,
						Image:        &image,
						ID:           image.ID,
					}, false)
			}

			// This is called if an image cannot be successfully uploaded and saved
			ipr.Error = func(ipr *utils.ImageProcessRequest) {
				image.Delete(Base.Db)

				// Let the client know we failed
				Base.WebsockChannel <- wsock.UserJSONNotification(
					&viewData.Session.User,
					"IM_PROCESS_FAILED", imageProcessErrorNotification{
						OriginalName: ipr.OriginalName,
						Media:        image.Media,
						MediaID:      image.MediaID,
					}, false)
			}

			Base.ImageChannel <- ipr
		}
	}
	RenderTextJSON(w, response)
}

type webAPIUploadProgressResponse struct {
	Successful     bool                  `json:"successful"`
	Error          string                `json:"error"`
	UploadProgress *utils.UploadProgress `json:"upload_progress"`
}

// WebAPIUploadProgress handles the route '/webapi/upload/progress/'
func WebAPIUploadProgress(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)

	response := webAPIUploadProgressResponse{Successful: false}
	if viewData.Session == nil {
		response.Error = "Requires Login"
		RenderJSON(w, response)
		return
	}

	args := URIArgs(r)
	if len(args) != 1 {
		response.Error = "Invalid Arguments"
		RenderJSON(w, response)
		return
	}

	token := args[0]
	uploadProgress, err := utils.GetUploadProgress(token)
	if err != nil {
		response.Error = err.Error()
		RenderJSON(w, response)
		return
	}

	response.Successful = true
	response.UploadProgress = uploadProgress
	RenderJSON(w, response)
}

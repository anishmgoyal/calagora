package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/anishmgoyal/calagora/models"
	"github.com/anishmgoyal/calagora/wsock"
)

type imageDelete struct {
	Media   string `json:"media"`
	MediaID int    `json:"media_id"`
	ID      int    `json:"id"`
}

type webAPIImageDeleteResponse struct {
	Successful bool `json:"successful"`
}

// WebAPIImageDelete handles the route '/webapi/image/delete/'
func WebAPIImageDelete(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIImageDeleteResponse{Successful: false}
	if viewData.Session == nil {
		RenderJSON(w, response)
		return
	}
	args := URIArgs(r)
	if len(args) != 1 {
		RenderJSON(w, response)
		return
	}

	idStr := args[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RenderJSON(w, response)
		return
	}

	image, err := models.GetImageByID(Base.Db, id)
	if err != nil || image.User.ID != viewData.Session.User.ID {
		RenderJSON(w, response)
		return
	}

	ok, err := image.Delete(Base.Db)
	if !ok && err != nil {
		log.Println("Failed to delete file: ", err)
		RenderJSON(w, response)
		return
	}

	Base.WebsockChannel <- wsock.UserJSONNotification(&viewData.Session.User,
		"IM_DELETE", imageDelete{
			Media:   image.Media,
			MediaID: image.MediaID,
			ID:      image.ID,
		}, false)

	response.Successful = true
	RenderJSON(w, response)
}

type imageMetaData struct {
	ID         string        `json:"id"`
	Successful bool          `json:"successful"`
	Image      *models.Image `json:"image,omitempty"`
}

type webAPIImageMetaResponse struct {
	Images []imageMetaData `json:"images"`
}

// WebAPIImageMeta handles the route '/webapi/image/meta/'
func WebAPIImageMeta(w http.ResponseWriter, r *http.Request) {
	viewData := BaseViewData(w, r)
	response := webAPIImageMetaResponse{
		Images: make([]imageMetaData, 0, 8),
	}
	if viewData.Session == nil {
		RenderJSON(w, response)
		return
	}

	imageList := r.FormValue("images")
	images := strings.Split(imageList, ",")
	if len(images) == 0 {
		RenderJSON(w, response)
		return
	}

	for _, idStr := range images {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			meta := imageMetaData{
				ID:         idStr,
				Successful: false,
			}
			response.Images = append(response.Images, meta)
			continue
		}
		image, err := models.GetImageByID(Base.Db, id)
		if err != nil || image.User.ID != viewData.Session.User.ID {
			meta := imageMetaData{
				ID:         idStr,
				Successful: false,
			}
			response.Images = append(response.Images, meta)
			continue
		}
		meta := imageMetaData{
			ID:         idStr,
			Successful: true,
			Image:      image,
		}
		response.Images = append(response.Images, meta)
	}
	RenderJSON(w, response)
}

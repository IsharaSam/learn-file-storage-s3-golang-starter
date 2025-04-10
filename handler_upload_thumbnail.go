package main

import (
	"fmt"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
	"io"
	"net/http"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
	}
	defer file.Close()
	mediaType := header.Header.Get("Content-Type")

	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}

	imageData, err := io.ReadAll(file) //

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read file", err)
	}

	video, err := cfg.db.GetVideo(videoID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to find video", err)
	}

	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "You are not authorized to upload this thumbnail", nil)
	}

	thumbnail := thumbnail{
		mediaType: mediaType,
		data:      imageData,
	}

	videoThumbnails[videoID] = thumbnail

	ThumbnailURL := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v\n\n", cfg.port, videoID)

	fmt.Println(ThumbnailURL)

	video.ThumbnailURL = &ThumbnailURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		delete(videoThumbnails, videoID)

		respondWithError(w, http.StatusInternalServerError, "Unable to save video", err)
	}

	respondWithJSON(w, http.StatusOK, video)
}

package upload

import (
	"encoding/json"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/delivery/interfaces"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"
)

type Request struct {
	MaxDownloads int16  `json:"max_downloads,omitempty" validate:"min=1" example:"5"`
	TTL          string `json:"ttl,omitempty" example:"2h30m"`
	Password     string `json:"password,omitempty" example:"1234"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

var validate *validator.Validate

// New @Summary Upload file
// @Description Uploads file on server
// @Tags file
// @Accept multipart/json
// @Produce json
// @Param request body Request true "File data"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 422 {object} Response
// @Failure 500 {object} Response
// @Router /upload [post]
func New(fileService interfaces.FileService, log *slog.Logger, cfg config.Config) http.HandlerFunc {
	validate = validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.upload.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		err := r.ParseMultipartForm(cfg.MaxFileSizeInBytes)
		if err != nil {
			log.Error("failed to parse form", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"failed to parse multipart/form")
			return
		}

		request := Request{
			MaxDownloads: cfg.MaxDownloads,
			TTL:          cfg.DefaultTtl.String(),
			Password:     "",
		}

		jsonData := r.FormValue("json")
		if err := json.Unmarshal([]byte(jsonData), &request); err != nil {
			log.Error("failed to parse json", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"failed to parse json")
			return
		}

		parsedTtl, err := time.ParseDuration(request.TTL)
		if err != nil {
			log.Error("failed to parse ttl", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"failed to parse ttl. it must be like '1h20m30s'")
			return
		}

		if err := validate.Struct(request); err != nil {
			var validationErrors validator.ValidationErrors
			errors.As(err, &validationErrors)
			log.Error("invalid request", sl.Error(err))
			response.RenderValidationError(w, r, validationErrors)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Error("file is required", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"file is required")
			return
		}

		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Error("failed to close file", sl.Error(err))
			}
		}(file)

		ctx := r.Context()
		alias, err := fileService.UploadFile(ctx, dto.UploadFileCommand{
			File:         file,
			FileSize:     header.Size,
			Filename:     header.Filename,
			Password:     request.Password,
			MaxDownloads: request.MaxDownloads,
			TTL:          parsedTtl,
		})

		if err != nil {
			if response.RenderFileServiceError(w, r, err) {
				log.Error("failed to upload file", sl.Error(err))
				return
			}

			log.Error("failed to upload file", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to upload file")
			return
		}

		log.Info("file was successfully uploaded", slog.String("alias", alias))
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Response: response.Response{},
			Alias:    alias,
		})
	}
}

package upload

import (
	"encoding/json"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/delivery/interfaces"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services"
	"expire-share/internal/services/dto"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"
)

type Request struct {
	MaxDownloads int16  `json:"max_downloads,omitempty" validate:"min=1"`
	TTL          string `json:"ttl,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

func New(fileService interfaces.FileService, log *slog.Logger, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.upload.New"
		log = log.With(
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

		if err := validator.New().Struct(request); err != nil {
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
			MaxDownloads: request.MaxDownloads,
			TTL:          parsedTtl,
		})

		if err != nil {
			if errors.Is(err, services.ErrFileSizeTooBig) {
				log.Error("file is too big", sl.Error(err))
				response.RenderError(w, r,
					http.StatusUnprocessableEntity,
					fmt.Sprintf("file size is very big. it must be less than %s", cfg.MaxFileSize))
				return
			}

			log.Error("failed to upload file", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to upload file")
			return
		}

		log.Info("file was uploaded", slog.String("alias", alias))
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Response: response.Response{},
			Alias:    alias,
		})
	}
}

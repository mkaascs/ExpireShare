package upload

import (
	"context"
	"expire-share/internal/config"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/delivery/response"
	"expire-share/internal/delivery/util"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/lib/log/sl"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
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

func (r *Request) SetDefault(cfg config.Service) {
	if r.MaxDownloads == 0 {
		r.MaxDownloads = cfg.MaxDownloads
	}

	if r.TTL == "" {
		r.TTL = cfg.DefaultTtl.String()
	}
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

type FileUploader interface {
	UploadFile(ctx context.Context, command commands.UploadFile) (string, error)
}

// New @Summary Upload file
//
//	@Description	Uploads file on server
//	@Tags			file
//	@Accept			multipart/json
//	@Produce		json
//	@Param			request	body		Request	true	"File data"
//	@Success		201		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		422		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/upload [post]
func New(uploader FileUploader, log *slog.Logger, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.upload.api.New"
		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		request, ok := middlewares.GetParsedBodyRequest[Request](r)
		if !ok {
			log.Error("failed to parse request")
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		claims, err := middlewares.GetUserClaims(r)
		if err != nil {
			log.Error("failed to get user claims", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		parsedTtl, err := time.ParseDuration(request.TTL)
		if err != nil {
			log.Info("failed to parse ttl", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"failed to parse ttl. it must be like '1h20m30s'")
			return
		}

		err = r.ParseMultipartForm(cfg.MaxFileSizeInBytes)
		if err != nil {
			log.Info("failed to parse form", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"failed to parse multipart/form")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Info("file is required", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"file is required")
			return
		}

		defer func(file multipart.File) {
			if err := file.Close(); err != nil {
				log.Error("failed to close file", sl.Error(err))
			}
		}(file)

		alias, err := uploader.UploadFile(r.Context(), commands.UploadFile{
			File:         file,
			FileSize:     header.Size,
			Filename:     header.Filename,
			Password:     request.Password,
			MaxDownloads: request.MaxDownloads,
			TTL:          parsedTtl,
			RequestingUserInfo: commands.RequestingUserInfo{
				UserID: claims.UserID,
				Roles:  claims.Roles,
			},
		})

		if err != nil {
			if response.RenderFileServiceError(w, r, err) || util.IsCtxError(err) {
				log.Info("failed to upload file", sl.Error(err))
				return
			}

			log.Error("failed to upload file", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		log.Info("file was successfully uploaded", slog.String("alias", alias))
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Alias: alias,
		})
	}
}

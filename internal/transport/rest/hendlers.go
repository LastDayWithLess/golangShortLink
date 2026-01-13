package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"short_link/internal/logger"
	"short_link/internal/model"
	"short_link/internal/service"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type HTTPHandler struct {
	linksServ *service.LinkService
}

func NewHTTPHanler(linksServ *service.LinkService) *HTTPHandler {
	return &HTTPHandler{
		linksServ: linksServ,
	}
}

func (h *HTTPHandler) SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {

	h.linksServ.Logger.Error(message)

	errDTO := model.ErrorDTO{
		Message: message,
		Time:    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errDTO); err != nil {
		h.linksServ.Logger.Error(err.Error())
	}
}

func (h *HTTPHandler) HandleCreateShortLink(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	if !strings.Contains(contentType, "application/json") {
		h.SendErrorResponse(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var linkDTO model.LinkDTO

	if err := json.NewDecoder(r.Body).Decode(&linkDTO); err != nil {
		h.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	link, err := h.linksServ.Create(ctx, linkDTO.URL)

	if err != nil {
		h.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	fullURL := &url.URL{
		Scheme: "http",
		Host:   r.Host,
		Path:   r.URL.Path + "/" + link.ShortURL,
	}

	link.ShortURL = fullURL.String()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(link); err != nil {
		h.linksServ.Logger.Error(err.Error(), logger.String("original_url", link.URL), logger.String("short_url", link.ShortURL))
	}

}

func (h *HTTPHandler) HandleGetAllShortLink(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	links, err := h.linksServ.GetAllShortLink(ctx)

	if err != nil {
		h.linksServ.Logger.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(links); err != nil {
		h.linksServ.Logger.Error(err.Error())
	}
}

func (h *HTTPHandler) HandleRedirection(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shortLink := mux.Vars(r)["shortLink"]

	if shortLink == "" {
		h.SendErrorResponse(w, http.StatusBadRequest, "Short link is required")
		return
	}

	link, err := h.linksServ.GetOriginalLink(ctx, shortLink)

	if err != nil {
		if errors.Is(err, service.ErrLinkBadRequest) {
			h.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			h.SendErrorResponse(w, http.StatusInternalServerError, "Failed to get original link")
		}
		return
	}

	http.Redirect(w, r, *link, http.StatusFound)
}

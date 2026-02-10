package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"short_link/internal/model"
	"short_link/internal/service"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	linksServ *service.LinkService
}

func NewHTTPHanler(linksServ *service.LinkService) *HTTPHandler {
	return &HTTPHandler{
		linksServ: linksServ,
	}
}

func (h *HTTPHandler) HandleCreateShortLink(c *gin.Context) {
	contentType := c.Request.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"success":  false,
			"error":    "Unsupported Media Type",
			"message":  "Content-Type must be application/json",
			"received": contentType,
			"expected": "application/json",
		})
		return
	}

	var req model.LinkDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	link, err := h.linksServ.Create(ctx, req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	fullURL := &url.URL{
		Scheme: "http",
		Host:   c.Request.Host,
		Path:   c.Request.URL.Path + "/" + link.ShortURL,
	}

	link.ShortURL = fullURL.String()

	c.JSON(http.StatusCreated, link)
}

func (h *HTTPHandler) HandleGetAllShortLink(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	links, err := h.linksServ.GetAllShortLink(ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "link not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, links)
}

func (h *HTTPHandler) HandleRedirection(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shortLink := c.Param("shortLink")
	if shortLink == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Short link is required",
		})
		return
	}

	link, err := h.linksServ.GetOriginalLink(ctx, shortLink)
	if err != nil {
		if errors.Is(err, service.ErrLinkBadRequest) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad request",
				"details": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get original link",
				"details": err.Error(),
			})
		}
		return
	}

	fmt.Println(link)

	c.Redirect(http.StatusMovedPermanently, *link)
}

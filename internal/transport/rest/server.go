package rest

import (
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	httpHandler *HTTPHandler
}

func NewServer(httpHandler *HTTPHandler) *HTTPServer {
	return &HTTPServer{
		httpHandler: httpHandler,
	}
}

func (s *HTTPServer) StartServer() error {

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.POST("/oneLink", s.httpHandler.HandleCreateShortLink)
	router.GET("/oneLink", s.httpHandler.HandleGetAllShortLink)
	router.GET("/oneLink/:shortLink", s.httpHandler.HandleRedirection)

	router.Run(":8080")

	return nil
}

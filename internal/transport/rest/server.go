package rest

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
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
	router := mux.NewRouter()

	router.Path("/oneLink").Methods("POST").HandlerFunc(s.httpHandler.HandleCreateShortLink)
	router.Path("/oneLink").Methods("GET").HandlerFunc(s.httpHandler.HandleGetAllShortLink)
	router.Path("/oneLink/{shortLink}").Methods("GET").HandlerFunc(s.httpHandler.HandleRedirection)

	if err := http.ListenAndServe(":8080", router); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}

	return nil
}

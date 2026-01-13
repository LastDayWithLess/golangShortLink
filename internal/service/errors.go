package service

import "errors"

var ErrLinkNotFound = errors.New("link not found")
var ErrLinkBadRequest = errors.New("uncorrected link or request body")
var ErrTooManyAttempts = errors.New("too many generation attempts")

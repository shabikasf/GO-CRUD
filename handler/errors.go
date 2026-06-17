package handler

import (
	"errors"
)

var ErrEncodingResponse = errors.New("failed to encode response")
var ErrDecodingResponse = errors.New("failed to decode request body")
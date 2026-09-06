package library

import "errors"

var ErrUnauthenticated = errors.New("not authenticated")

var ErrNotFound = errors.New("not found")

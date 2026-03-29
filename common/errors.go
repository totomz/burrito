package common

import "errors"

// ErrNotFound the requested resource was not found
var ErrNotFound = errors.New("not found")

// ErrNoSeatAvailable means that the rea
var ErrNoSeatAvailable = errors.New("no free seats available")

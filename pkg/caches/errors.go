package caches

import "github.com/goodblaster/errors"

var ErrInvalidKey = errors.New("invalid key: %v")
var ErrSinglePathKeyRequired = errors.New("single path key required: %v")
var ErrKeyAlreadyExists = errors.New("key already exists: %s")
var ErrKeyNotFound = errors.New("key not found: %s")

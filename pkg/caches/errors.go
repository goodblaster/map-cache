package caches

import "github.com/goodblaster/errors"

var ErrInvalidKey = errors.New("invalid key")
var ErrSinglePathKeyRequired = errors.New("single path key required")
var ErrKeyAlreadyExists = errors.New("key already exists")
var ErrKeyNotFound = errors.New("key not found")

package caches

import "github.com/goodblaster/errors"

// Key and value errors
var ErrInvalidKey = errors.New("invalid key: %v")
var ErrSinglePathKeyRequired = errors.New("single path key required: %v")
var ErrKeyAlreadyExists = errors.New("key already exists: %s")
var ErrKeyNotFound = errors.New("key not found: %s")
var ErrNotAnArray = errors.New("not an array: %s")
var ErrNotANumber = errors.New("not a number")
var ErrIncrementValueNotNumber = errors.New("increment value must be a number")

// Cache management errors
var ErrCacheAlreadyExists = errors.New("cache already exists")
var ErrCacheNotFound = errors.New("cache not found")

// Trigger errors
var ErrTriggerNotFound = errors.New("trigger not found")
var ErrTriggerRecursionLimit = errors.New("trigger recursion depth limit exceeded (max: %d) - possible infinite loop detected")
var ErrWildcardEmptySegment = errors.New("wildcard at index %d matched empty segment")
var ErrSegmentMismatch = errors.New("segment mismatch at index %d: %s != %s")
var ErrMismatchedPathLengths = errors.New("mismatched path lengths: %v vs %v")

// Command and expression errors
var ErrUnknownCommandType = errors.New("unknown command type: %s")
var ErrInvalidExpression = errors.New("invalid expression: %w")
var ErrEvaluationError = errors.New("evaluation error: %w")
var ErrExpressionNotBoolean = errors.New("expression did not return a boolean")
var ErrInvalidForExpression = errors.New("invalid FOR expression: %s")
var ErrForExpressionNeedsWildcard = errors.New("FOR expression must include a wildcard: %s")

// Interpolation errors
var ErrWildcardInterpolation = errors.New("wildcard interpolation error for key %q: %w")
var ErrInterpolation = errors.New("interpolation error for key %q: %w")
var ErrWildcardInTemplate = errors.New("wildcards not allowed in templated string: %q")
var ErrWildcardWithFallback = errors.New("wildcards not allowed with fallback operator: %q")
var ErrInvalidFallbackExpression = errors.New("fallback expression must have exactly 2 parts (key || default), got %d parts in: %q")

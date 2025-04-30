package v1errors

import (
	"testing"

	"github.com/goodblaster/errors"
	"github.com/stretchr/testify/assert"
)

func TestWebError(t *testing.T) {
	var err error = &Error{Msg: "web error"}
	err = errors.Wrap(err, "wrapped error")
	err = errors.Wrap(err, "another wrapped error")

	assert.EqualValues(t, "web error", WebError(err).Error())
}

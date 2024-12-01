package server

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

type ServerTest interface {
	Run() error
}

func TestNew(t *testing.T) {
	s := New()
	assert.IsType(t, s, &Server{})
	assert.Implements(t, (*ServerTest)(nil), s)
}

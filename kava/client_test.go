package kava

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	_, err := NewClient()
	assert.NoError(t, err)
}

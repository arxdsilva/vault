//go:build darwin

package main

import (
	"github.com/asilva/vault/internal/core"
	"github.com/asilva/vault/internal/platform/darwin"
)

func newTrasher() core.Trasher {
	return darwin.NewTrash()
}

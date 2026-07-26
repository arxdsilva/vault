//go:build windows

package main

import (
	"github.com/asilva/vault/internal/core"
	"github.com/asilva/vault/internal/platform/windows"
)

func newTrasher() core.Trasher {
	return windows.NewTrash()
}

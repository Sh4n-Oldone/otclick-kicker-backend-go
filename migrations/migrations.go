package migrations

import (
	"embed"
)

//go:embed database/*
var Database embed.FS

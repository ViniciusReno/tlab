// Package tlab embeds the resources required by the offline application.
package tlab

import "embed"

// Files contains local assets, versioned migrations, and attributed demo data.
//
//go:embed migrations/*.sql data/demo/* web/templates/*.html web/static/*
var Files embed.FS

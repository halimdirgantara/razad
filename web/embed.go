package web

import "embed"

//go:embed build/*
var UI embed.FS

// BuildID forces a Go recompile when the frontend build changes.
// Update this when rebuilding the frontend.
const BuildID = "2026-06-09-v2"

package ubi8javaextension

import libcnb "github.com/buildpacks/libcnb"

// A libcnb style represenation of the Extension.toml metadata
// to allow parsing of the file to Go structs.

// extension is the contents of the extension.toml file.
type Extension struct {
	// API is the api version expected by the extension.
	API string `toml:"api"`

	// Info is information about the extension.
	// The format is the same as the BuildpackInfo today
	// it's beneficial to leave the type the same for now, to allow
	// passing it to places that expect a BuildpackInfo
	Info libcnb.BuildpackInfo `toml:"extension"`

	// Path is the path to the extension.
	Path string `toml:"-"`

	// Metadata is arbitrary metadata attached to the extension.
	Metadata map[string]interface{} `toml:"metadata"`
}

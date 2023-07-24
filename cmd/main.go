package main

import (
	"os"

	"github.com/paketo-buildpacks/libjvm/v2"
	"github.com/paketo-buildpacks/libjvm/v2/helper"
	"github.com/paketo-buildpacks/libpak/v2/log"
	ubi8javaextension "github.com/paketo-community/ubi-java-extension/v1"

	"github.com/paketo-buildpacks/libpak/v2"
)

func main() {
	logger := log.NewPaketoLogger(os.Stdout)

	// not used directly, but this forces the helper module to be included in the module
	// we need the helper module because of the way that `scripts/build/sh` builds the helper cmd
	_ = helper.ActiveProcessorCount{Logger: logger}

	libpak.ExtensionMain(
		ubi8javaextension.Detect,
		libjvm.NewGenerate(logger, ubi8javaextension.Generate()).Generate,
	)
}

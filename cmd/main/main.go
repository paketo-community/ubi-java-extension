package main

import (
	"os"

	"github.com/paketo-buildpacks/libjvm/v2"
	"github.com/paketo-buildpacks/libpak/v2/log"
	ubi8javaextension "github.com/paketo-community/ubi-java-extension/v1"

	"github.com/paketo-buildpacks/libpak/v2"
)

func main() {

	libpak.ExtensionMain(
		ubi8javaextension.Detect,
		libjvm.NewGenerate(log.NewPaketoLogger(os.Stdout), ubi8javaextension.Generate()).Generate,
	)
}

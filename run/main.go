package main

import (
	ubi8javaextension "github.com/BarDweller/ubi-java-extension"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

func main() {
	dependencyManager := postal.NewService(cargo.NewTransport())

	packit.RunExtension(
		ubi8javaextension.Detect(),
		ubi8javaextension.Generate(dependencyManager),
	)
}

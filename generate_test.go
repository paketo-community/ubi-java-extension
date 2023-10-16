package ubi8javaextension_test

import (
	_ "embed"
	"io"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm/v2"
	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/log"
	ubijavaextension "github.com/paketo-community/ubi-java-extension/v1"
	"github.com/sclevine/spec"
)

type RunDockerfileProps struct {
	Source string
}

//go:embed templates/run.Dockerfile
var runDockerfileTemplate string

type BuildDockerfileProps struct {
	JAVA_EXTENSION_HELPERS string
	JAVA_VERSION           string
	CNB_USER_ID            int
	CNB_GROUP_ID           int
	CNB_STACK_ID           string
	PACKAGES               string
}

//go:embed templates/build.Dockerfile
var buildDockerfileTemplate string

func testFillPropsToTemplate(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
	)

	context("Adding props on templates with FillPropsToTemplate", func() {

		it("Should fill with properties the template/build.Dockerfile", func() {

			buildDockerfileProps := BuildDockerfileProps{
				JAVA_VERSION:           "17",
				CNB_USER_ID:            1000,
				CNB_GROUP_ID:           1000,
				CNB_STACK_ID:           "ubi8-paketo",
				PACKAGES:               "openssl-devel java-17-openjdk-devel nss_wrapper which",
				JAVA_EXTENSION_HELPERS: "helper1, helper2",
			}

			output, err := ubijavaextension.FillPropsToTemplate(buildDockerfileProps, buildDockerfileTemplate)

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(`ARG base_image
FROM ${base_image}

USER root

ARG build_id=0
RUN echo ${build_id}

RUN microdnf --setopt=install_weak_deps=0 --setopt=tsflags=nodocs install -y openssl-devel java-17-openjdk-devel nss_wrapper which && microdnf clean all

RUN echo "17" > /bpi.paketo.ubi.java.version
RUN echo "helper1, helper2" > /bpi.paketo.ubi.java.helpers

USER 1000:1000`))

		})

		it("Should fill with properties the template/run.Dockerfile", func() {

			RunDockerfileProps := RunDockerfileProps{
				Source: "paketo-buildpacks/ubi8-paketo-run-java-17",
			}

			output, err := ubijavaextension.FillPropsToTemplate(RunDockerfileProps, runDockerfileTemplate)

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(`FROM paketo-buildpacks/ubi8-paketo-run-java-17`))

		})
	})
}

func testGenerate(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect     = NewWithT(t).Expect
		workingDir string
		err        error

		generateResult libjvm.GenerateContentResult
	)

	context("GenerateContentBuilder invoked", func() {
		it.Before(func() {

			workingDir = t.TempDir()
			Expect(err).NotTo(HaveOccurred())

			err = os.Chdir(workingDir)
			Expect(err).NotTo(HaveOccurred())
		})

		it("Java version 17 recognised", func() {
			generateResult, err = ubijavaextension.Generate()(libjvm.GenerateContentContext{
				Logger: log.NewDiscardLogger(),
				ConfigurationResolver: libpak.ConfigurationResolver{Configurations: []libpak.BuildModuleConfiguration{
					{Name: "BP_JVM_VERSION", Default: "17"},
				}},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateResult.BuildDockerfile).NotTo(BeNil())
			Expect(generateResult.RunDockerfile).NotTo(BeNil())

			buf := new(strings.Builder)
			_, _ = io.Copy(buf, generateResult.RunDockerfile)
			Expect(buf.String()).To(ContainSubstring("paketocommunity/ubi8-paketo-run-java-17"))

			buf.Reset()
			_, _ = io.Copy(buf, generateResult.BuildDockerfile)
			Expect(buf.String()).To(ContainSubstring("java-17-openjdk-devel"))
		})

		it("Java version 1.8 recognised", func() {
			generateResult, err = ubijavaextension.Generate()(libjvm.GenerateContentContext{
				Logger: log.NewDiscardLogger(),
				ConfigurationResolver: libpak.ConfigurationResolver{Configurations: []libpak.BuildModuleConfiguration{
					{Name: "BP_JVM_VERSION", Default: "1.8"},
				}},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateResult.BuildDockerfile).NotTo(BeNil())
			Expect(generateResult.RunDockerfile).NotTo(BeNil())

			buf := new(strings.Builder)
			_, _ = io.Copy(buf, generateResult.RunDockerfile)
			Expect(buf.String()).To(ContainSubstring("paketocommunity/ubi8-paketo-run-java-8"))

			buf.Reset()
			_, _ = io.Copy(buf, generateResult.BuildDockerfile)
			Expect(buf.String()).To(ContainSubstring("java-1.8.0-openjdk-devel"))
		})

		it("Java version 11 recognised", func() {
			generateResult, err = ubijavaextension.Generate()(libjvm.GenerateContentContext{
				Logger: log.NewDiscardLogger(),
				ConfigurationResolver: libpak.ConfigurationResolver{Configurations: []libpak.BuildModuleConfiguration{
					{Name: "BP_JVM_VERSION", Default: "11"},
				}},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateResult.BuildDockerfile).NotTo(BeNil())
			Expect(generateResult.RunDockerfile).NotTo(BeNil())

			buf := new(strings.Builder)
			_, _ = io.Copy(buf, generateResult.RunDockerfile)
			Expect(buf.String()).To(ContainSubstring("paketocommunity/ubi8-paketo-run-java-11"))

			buf.Reset()
			_, _ = io.Copy(buf, generateResult.BuildDockerfile)
			Expect(buf.String()).To(ContainSubstring("java-11-openjdk-devel"))
		})

	}, spec.Sequential())

}

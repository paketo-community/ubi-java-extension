package ubi8javaextension_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ubijavaextension "github.com/BarDweller/ubi-java-extension"
	"github.com/BarDweller/ubi-java-extension/fakes"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/packit/v2/cargo"

	"github.com/BurntSushi/toml"
	postal "github.com/paketo-buildpacks/packit/v2/postal"
)

type RunDockerfileProps struct {
	Source string
}

//go:embed templates/run.Dockerfile
var runDockerfileTemplate string

type BuildDockerfileProps struct {
	JAVA_VERSION string
	CNB_USER_ID  int
	CNB_GROUP_ID int
	CNB_STACK_ID string
	PACKAGES     string
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
				JAVA_VERSION: "17",
				CNB_USER_ID:  1000,
				CNB_GROUP_ID: 1000,
				CNB_STACK_ID: "ubi8-paketo",
				PACKAGES:     "openssl-devel java-17-openjdk-devel nss_wrapper which",
			}

			output, err := ubijavaextension.FillPropsToTemplate(buildDockerfileProps, buildDockerfileTemplate)

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(`ARG base_image
FROM ${base_image}

USER root

ARG build_id=0
RUN echo ${build_id}

RUN microdnf --setopt=install_weak_deps=0 --setopt=tsflags=nodocs install -y openssl-devel java-17-openjdk-devel nss_wrapper which && microdnf clean all

RUN echo uid:gid "1000:1000"
USER 1000:1000

RUN echo "CNB_STACK_ID: ubi8-paketo"`))

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
		Expect               = NewWithT(t).Expect
		workingDir           string
		planPath             string
		testBuildPlan        packit.BuildpackPlan
		buf                  = new(bytes.Buffer)
		generateResult       packit.GenerateResult
		err                  error
		cnbDir               string
		dependencyManager    *fakes.DependencyManager
		BuildDockerfileProps = ubijavaextension.BuildDockerfileProps{
			CNB_USER_ID:  ubijavaextension.CNB_USER_ID,
			CNB_GROUP_ID: ubijavaextension.CNB_GROUP_ID,
			CNB_STACK_ID: "ubi8-paketo",
			PACKAGES:     "openssl-devel java-17-openjdk-devel nss_wrapper which",
		}
	)

	context("Generate called with NO jdk/jre in buildplan", func() {
		it.Before(func() {

			workingDir = t.TempDir()
			Expect(err).NotTo(HaveOccurred())

			err = toml.NewEncoder(buf).Encode(testBuildPlan)
			Expect(err).NotTo(HaveOccurred())

			Expect(os.WriteFile(filepath.Join(workingDir, "plan"), buf.Bytes(), 0600)).To(Succeed())

			err = os.Chdir(workingDir)
			Expect(err).NotTo(HaveOccurred())
		})

		it("Java no longer requested in buildplan", func() {
			dependencyManager = &fakes.DependencyManager{}
			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{Name: "Java", ID: "jdk", Version: "17"}

			generateResult, err = ubijavaextension.Generate(dependencyManager)(packit.GenerateContext{
				WorkingDir: workingDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(generateResult.BuildDockerfile).To(BeNil())
		})
	}, spec.Sequential())

	context("Generate called with jdk/jre in the buildplan", func() {
		it.Before(func() {
			workingDir = t.TempDir()
			cnbDir, err = os.MkdirTemp("", "cnb")

			err = toml.NewEncoder(buf).Encode(testBuildPlan)
			Expect(err).NotTo(HaveOccurred())

			planPath = filepath.Join(workingDir, "plan")
			Expect(os.WriteFile(planPath, buf.Bytes(), 0600)).To(Succeed())

			err = os.Chdir(workingDir)
			Expect(err).NotTo(HaveOccurred())

			t.Setenv("CNB_BP_PLAN_PATH", planPath)
			t.Setenv("CNB_EXTENSION_DIR", cnbDir)
			t.Setenv("CNB_OUTPUT_DIR", workingDir)
			t.Setenv("CNB_PLATFORM_DIR", workingDir)
			t.Setenv("CNB_STACK_ID", "ubi8-paketo")
		})

		it("Specific version of java requested", func() {
			dependencyManager := postal.NewService(cargo.NewTransport())

			versionTests := []struct {
				Name                               string
				Metadata                           map[string]interface{}
				RunDockerfileProps                 ubijavaextension.RunDockerfileProps
				BuildDockerfileProps               ubijavaextension.BuildDockerfileProps
				buildDockerfileExpectedJavaVersion string
			}{
				{
					Name: "jdk",
					Metadata: map[string]interface{}{
						"version":        "17",
						"version-source": "BP_JVM_VERSION",
					},
					RunDockerfileProps: ubijavaextension.RunDockerfileProps{
						Source: "paketo-buildpacks/ubi8-paketo-run-java-17",
					},
					BuildDockerfileProps: ubijavaextension.BuildDockerfileProps{
						CNB_USER_ID:  BuildDockerfileProps.CNB_USER_ID,
						CNB_GROUP_ID: BuildDockerfileProps.CNB_GROUP_ID,
						CNB_STACK_ID: BuildDockerfileProps.CNB_STACK_ID,
						JAVA_VERSION: "17",
						PACKAGES:     "openssl-devel java-17-openjdk-devel nss_wrapper which",
					},
					buildDockerfileExpectedJavaVersion: "17",
				},
				{
					Name: "jdk",
					Metadata: map[string]interface{}{
						"version":        "11",
						"version-source": "BP_JVM_VERSION",
					},
					RunDockerfileProps: ubijavaextension.RunDockerfileProps{
						Source: "paketo-buildpacks/ubi8-paketo-run-java-11",
					},
					BuildDockerfileProps: ubijavaextension.BuildDockerfileProps{
						CNB_USER_ID:  BuildDockerfileProps.CNB_USER_ID,
						CNB_GROUP_ID: BuildDockerfileProps.CNB_GROUP_ID,
						CNB_STACK_ID: BuildDockerfileProps.CNB_STACK_ID,
						JAVA_VERSION: "11",
						PACKAGES:     "openssl-devel java-11-openjdk-devel nss_wrapper which",
					},
					buildDockerfileExpectedJavaVersion: "11",
				},
				{
					Name: "jdk",
					Metadata: map[string]interface{}{
						"version":        "8",
						"version-source": "BP_JVM_VERSION",
					},
					RunDockerfileProps: ubijavaextension.RunDockerfileProps{
						Source: "paketo-buildpacks/ubi8-paketo-run-java-8",
					},
					BuildDockerfileProps: ubijavaextension.BuildDockerfileProps{
						CNB_USER_ID:  BuildDockerfileProps.CNB_USER_ID,
						CNB_GROUP_ID: BuildDockerfileProps.CNB_GROUP_ID,
						CNB_STACK_ID: BuildDockerfileProps.CNB_STACK_ID,
						JAVA_VERSION: "8",
						PACKAGES:     "openssl-devel java-1.8.0-openjdk-devel nss_wrapper which",
					},
					buildDockerfileExpectedJavaVersion: "8",
				},
			}

			for _, tt := range versionTests {

				extensionToml, _ := readExtensionTomlTemplateFile(tt.buildDockerfileExpectedJavaVersion)
				Expect(err).NotTo(HaveOccurred())
				Expect(os.WriteFile(cnbDir+"/extension.toml", []byte(extensionToml), 0600)).To(Succeed())

				generateResult, err = ubijavaextension.Generate(dependencyManager)(packit.GenerateContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name:     tt.Name,
								Metadata: tt.Metadata,
							},
						},
					},
					Stack: "ubi8-paketo",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(generateResult).NotTo(Equal(nil))

				runDockerfileContent, _ := ubijavaextension.FillPropsToTemplate(tt.RunDockerfileProps, runDockerfileTemplate)
				tt.BuildDockerfileProps.JAVA_VERSION = tt.buildDockerfileExpectedJavaVersion
				buildDockerfileContent, _ := ubijavaextension.FillPropsToTemplate(tt.BuildDockerfileProps, buildDockerfileTemplate)

				buf := new(strings.Builder)
				_, _ = io.Copy(buf, generateResult.RunDockerfile)
				Expect(buf.String()).To(Equal(runDockerfileContent))
				buf.Reset()
				_, _ = io.Copy(buf, generateResult.BuildDockerfile)
				Expect(buf.String()).To(Equal(buildDockerfileContent))
			}

		})

		it("should return the default when java version has NOT been requested", func() {
			dependencyManager := postal.NewService(cargo.NewTransport())

			versionTests := []struct {
				Name                               string
				Metadata                           map[string]interface{}
				RunDockerfileProps                 ubijavaextension.RunDockerfileProps
				BuildDockerfileProps               ubijavaextension.BuildDockerfileProps
				buildDockerfileExpectedJavaVersion string
			}{
				{
					Name:     "jdk",
					Metadata: map[string]interface{}{},
					RunDockerfileProps: ubijavaextension.RunDockerfileProps{
						Source: "paketo-buildpacks/ubi8-paketo-run-java-17",
					},
					BuildDockerfileProps: ubijavaextension.BuildDockerfileProps{
						CNB_USER_ID:  BuildDockerfileProps.CNB_USER_ID,
						CNB_GROUP_ID: BuildDockerfileProps.CNB_GROUP_ID,
						CNB_STACK_ID: BuildDockerfileProps.CNB_STACK_ID,
						JAVA_VERSION: "17",
						PACKAGES:     "openssl-devel java-17-openjdk-devel nss_wrapper which",
					},
					buildDockerfileExpectedJavaVersion: "17",
				},
				{
					Name: "jdk",
					Metadata: map[string]interface{}{
						"version":        "",
						"version-source": "BP_JVM_VERSION",
					},
					RunDockerfileProps: ubijavaextension.RunDockerfileProps{
						Source: "paketo-buildpacks/ubi8-paketo-run-java-17",
					},
					BuildDockerfileProps: ubijavaextension.BuildDockerfileProps{
						CNB_USER_ID:  BuildDockerfileProps.CNB_USER_ID,
						CNB_GROUP_ID: BuildDockerfileProps.CNB_GROUP_ID,
						CNB_STACK_ID: BuildDockerfileProps.CNB_STACK_ID,
						JAVA_VERSION: "17",
						PACKAGES:     "openssl-devel java-17-openjdk-devel nss_wrapper which",
					},
					buildDockerfileExpectedJavaVersion: "17",
				},
				{
					Name: "jdk",
					Metadata: map[string]interface{}{
						"version":        "x",
						"version-source": "BP_JVM_VERSION",
					},
					RunDockerfileProps: ubijavaextension.RunDockerfileProps{
						Source: "paketo-buildpacks/ubi8-paketo-run-java-17",
					},
					BuildDockerfileProps: ubijavaextension.BuildDockerfileProps{
						CNB_USER_ID:  BuildDockerfileProps.CNB_USER_ID,
						CNB_GROUP_ID: BuildDockerfileProps.CNB_GROUP_ID,
						CNB_STACK_ID: BuildDockerfileProps.CNB_STACK_ID,
						JAVA_VERSION: "17",
						PACKAGES:     "openssl-devel java-17-openjdk-devel nss_wrapper which",
					},
					buildDockerfileExpectedJavaVersion: "17",
				},
			}

			for _, tt := range versionTests {

				extensionToml, _ := readExtensionTomlTemplateFile(tt.buildDockerfileExpectedJavaVersion)
				Expect(err).NotTo(HaveOccurred())
				Expect(os.WriteFile(cnbDir+"/extension.toml", []byte(extensionToml), 0600)).To(Succeed())

				generateResult, err = ubijavaextension.Generate(dependencyManager)(packit.GenerateContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name:     tt.Name,
								Metadata: tt.Metadata,
							},
						},
					},
					Stack: "ubi8-paketo",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(generateResult).NotTo(Equal(nil))

				runDockerfileContent, _ := ubijavaextension.FillPropsToTemplate(tt.RunDockerfileProps, runDockerfileTemplate)
				tt.BuildDockerfileProps.JAVA_VERSION = tt.buildDockerfileExpectedJavaVersion
				buildDockerfileContent, _ := ubijavaextension.FillPropsToTemplate(tt.BuildDockerfileProps, buildDockerfileTemplate)

				buf := new(strings.Builder)
				_, _ = io.Copy(buf, generateResult.RunDockerfile)
				Expect(buf.String()).To(Equal(runDockerfileContent))
				buf.Reset()
				_, _ = io.Copy(buf, generateResult.BuildDockerfile)
				Expect(buf.String()).To(Equal(buildDockerfileContent))
			}

		})

		it("Should error on below cases of requested java", func() {
			dependencyManager := postal.NewService(cargo.NewTransport())

			versionTests := []struct {
				Name           string
				Metadata       map[string]interface{}
				BP_JVM_VERSION string
			}{
				{
					Name:           "jdk",
					Metadata:       map[string]interface{}{},
					BP_JVM_VERSION: "16",
				},
				{
					Name:           "jdk",
					Metadata:       map[string]interface{}{},
					BP_JVM_VERSION: "1.9.10",
				},
			}

			for _, tt := range versionTests {

				t.Setenv("BP_JVM_VERSION", tt.BP_JVM_VERSION)

				fmt.Printf("testing with dir %s, %+v\n", cnbDir, tt)

				generateResult, err = ubijavaextension.Generate(dependencyManager)(packit.GenerateContext{
					WorkingDir: workingDir,
					CNBPath:    cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{
								Name:     tt.Name,
								Metadata: tt.Metadata,
							},
						},
					},
					Stack: "ubi8-paketo",
				})

				Expect(err).To(HaveOccurred())
			}
		})

	}, spec.Sequential())

}

func readExtensionTomlTemplateFile(defaultJavaVersion ...string) (string, error) {
	var version string
	if len(defaultJavaVersion) == 0 {
		version = "17"
	} else {
		version = defaultJavaVersion[0]
	}

	template := `
api = "0.7"

[extension]
id = "redhat-runtimes/java"
name = "RedHat Runtimes Java Dependency Extension"
version = "0.0.1"
description = "This extension installs the appropriate java runtime via dnf"

[metadata]

  [[metadata.configurations]]
    build = true
    default = "%s"
    description = "The Default Java version (testcase)"
    name = "BP_JVM_VERSION"	
`
	return fmt.Sprintf(template, version), nil
}

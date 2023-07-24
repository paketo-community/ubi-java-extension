package ubi8javaextension

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/buildpacks/libcnb/v2"
	"github.com/fatih/color"

	"github.com/paketo-buildpacks/libjvm/v2"
)

// Should be externalized
var CNB_USER_ID = 1000
var CNB_GROUP_ID = 1000

//go:embed templates/build.Dockerfile
var buildDockerfileTemplate string

type BuildDockerfileProps struct {
	JAVA_VERSION           string
	JAVA_EXTENSION_HELPERS string
	CNB_USER_ID            int
	CNB_GROUP_ID           int
	CNB_STACK_ID           string
	PACKAGES               string
}

//go:embed templates/run.Dockerfile
var runDockerfileTemplate string

type RunDockerfileProps struct {
	Source string
}

func Generate() libjvm.GenerateContentBuilder {
	return func(ctx libjvm.GenerateContentContext) (libjvm.GenerateContentResult, error) {

		//obtain jvm version from plan using config resolver via libjvm
		jvmVersion := libjvm.NewJVMVersion(ctx.Logger)
		JAVA_VERSION, err := jvmVersion.GetJVMVersion(ctx.Context.ApplicationPath, ctx.ConfigurationResolver)
		if err != nil {
			return libjvm.GenerateContentResult{}, fmt.Errorf("unable to resolve JVM version\n%w", err)
		}

		//map java version to appropriate UBI packages/images
		buildver, runver, err := mapRequestedVersionToPackageAndRunImage(JAVA_VERSION)
		if err != nil {
			return libjvm.GenerateContentResult{}, fmt.Errorf("unable to obtain ubi package/runtime\n%w", err)
		}

		//log choice of package being installed.
		f := color.New(color.Faint)
		ctx.Logger.Body(f.Sprintf("Using UBI Java package %s", buildver))

		// Create build.Dockerfile content
		buildDockerfileProps := BuildDockerfileProps{
			JAVA_VERSION:           JAVA_VERSION,
			JAVA_EXTENSION_HELPERS: strings.Join(contributeHelpers(ctx.Context, JAVA_VERSION), ","),
			CNB_USER_ID:            CNB_USER_ID,
			CNB_GROUP_ID:           CNB_GROUP_ID,
			CNB_STACK_ID:           ctx.Context.StackID,
			PACKAGES:               "openssl-devel " + buildver + " nss_wrapper which",
		}
		buildDockerfileContent, err := FillPropsToTemplate(buildDockerfileProps, buildDockerfileTemplate)
		if err != nil {
			return libjvm.GenerateContentResult{}, err
		}

		// Create run.Dockerfile content
		runDockerfileProps := RunDockerfileProps{
			Source: runver,
		}
		runDockerfileContent, err := FillPropsToTemplate(runDockerfileProps, runDockerfileTemplate)
		if err != nil {
			return libjvm.GenerateContentResult{}, err
		}

		return libjvm.GenerateContentResult{
			ExtendConfig:    libjvm.ExtendConfig{Build: libjvm.ExtendImageConfig{Args: []libjvm.ExtendImageConfigArg{}}},
			BuildDockerfile: strings.NewReader(buildDockerfileContent),
			RunDockerfile:   strings.NewReader(runDockerfileContent),
			GenerateResult:  libcnb.NewGenerateResult(),
		}, nil
	}
}

// Based on the function from libjvm build.go, this will determine the libjvm helpers required for this build.
// these will be converted to layers via the companion buildpack.
func contributeHelpers(context libcnb.GenerateContext, version string) []string {
	helpers := []string{"active-processor-count", "java-opts", "jvm-heap", "link-local-dns", "memory-calculator",
		"security-providers-configurer", "jmx", "jfr"}

	if libjvm.IsBeforeJava9(version) {
		helpers = append(helpers, "security-providers-classpath-8")
		helpers = append(helpers, "debug-8")
	} else {
		helpers = append(helpers, "security-providers-classpath-9")
		helpers = append(helpers, "debug-9")
		helpers = append(helpers, "nmt")
	}
	// Java 18 bug - cacerts keystore type not readable
	if libjvm.IsBeforeJava18(version) {
		helpers = append(helpers, "openssl-certificate-loader")
	}
	return helpers
}

func FillPropsToTemplate(properties interface{}, templateString string) (result string, Error error) {

	templ, err := template.New("template").Parse(templateString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = templ.Execute(&buf, properties)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

func mapRequestedVersionToPackageAndRunImage(requestedVersion string) (packages string, runImage string, err error) {
	//TODO: extenalise mappings via extension.toml ?
	//TODO: consider sloppy version matching, with next ver up migration.
	var buildver, runver string
	switch requestedVersion {
	case "8", "1.8", "1.8.0":
		buildver = "java-1.8.0-openjdk-devel"
		runver = "paketocommunity/ubi8-paketo-run-java-8"
	case "11":
		buildver = "java-11-openjdk-devel"
		runver = "paketocommunity/ubi8-paketo-run-java-11"
	case "17":
		buildver = "java-17-openjdk-devel"
		runver = "paketocommunity/ubi8-paketo-run-java-17"
	default:
		buildver = ""
		runver = ""
		err = fmt.Errorf("unable to map requested Java version of %s to a UBI supported runtime", requestedVersion)
	}
	return buildver, runver, err
}

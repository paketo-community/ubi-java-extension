package ubi8javaextension

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/packit/v2"

	postal "github.com/paketo-buildpacks/packit/v2/postal"

	libcnb "github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libjvm"
	"github.com/paketo-buildpacks/libpak"
)

// Should be externalized
var CNB_USER_ID = 1000
var CNB_GROUP_ID = 1000

//go:embed templates/build.Dockerfile
var buildDockerfileTemplate string

type BuildDockerfileProps struct {
	JAVA_VERSION string
	CNB_USER_ID  int
	CNB_GROUP_ID int
	CNB_STACK_ID string
	PACKAGES     string
}

//go:embed templates/run.Dockerfile
var runDockerfileTemplate string

type RunDockerfileProps struct {
	Source string
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

func Generate(dependencyManager DependencyManager) packit.GenerateFunc {
	return func(packitContext packit.GenerateContext) (packit.GenerateResult, error) {

		var logger bard.Logger
		var buildpack libcnb.Buildpack
		var extension Extension
		var err error
		logger = bard.NewLogger(os.Stdout)

		//check if jdk/jre are requested in the buildplan.
		jdkRequired, jreRequired, err := getJDKJRERequired(packitContext.Plan)
		if err != nil {
			return packit.GenerateResult{}, packit.Fail.WithMessage("unable to resolve jdk/jre required\n%w", err)
		}

		//if not, we have no work left to do.
		if !jdkRequired && !jreRequired {
			//neither jdk, nor jre was still in the plan when generate was invoked.
			//it's possible a previous extension satisified the requirement.
			//exit without error.
			logger.Debug("Java extension not required, jdk/jre no longer in plan")
			return packit.GenerateResult{}, packit.Fail.WithMessage("jdk/jre no longer requested by build plan")
		}

		//parse the extension toml, and convert to a 'Buildpack' structure we can use with Config Resolver
		// logger.Title expects Buildpack, but only reads the Info fields.
		// libpak.ConfigurationResolver expects to read bp.Metadata

		extension, err = getExtension(packitContext)
		if err != nil {
			return packit.GenerateResult{}, packit.Fail.WithMessage("unable to obtain Extension\n%w", err)
		}

		//TODO: revisit libjvm once libcnb supports extensions, try to clean up this requirement.
		logger.Debugf("Extension: %+v", extension)
		buildpack.Info = extension.Info
		buildpack.Metadata = extension.Metadata

		//log out the info for this extension as a pretty header.
		logger.Title(buildpack)

		//use the buildpack.Metadata to determine requested Java Version.
		cr, err := libpak.NewConfigurationResolver(buildpack, &logger)
		if err != nil {
			return packit.GenerateResult{}, packit.Fail.WithMessage("unable to create configuration resolver\n%w", err)
		}
		jvmVersion := libjvm.NewJVMVersion(logger)
		JAVA_VERSION, err := jvmVersion.GetJVMVersion(packitContext.WorkingDir, cr)
		if err != nil {
			return packit.GenerateResult{}, packit.Fail.WithMessage("unable to determine jvm version\n%w", err)
		}

		//map java version to appropriate UBI packages/images
		buildver, runver, err := mapRequestedVersionToPackageAndRunImage(JAVA_VERSION)
		if err != nil {
			return packit.GenerateResult{}, packit.Fail.WithMessage("unable to obtain ubi package/runtime\n%w\n", err)
		}

		//log choice of package being installed.
		f := color.New(color.Faint)
		logger.Body(f.Sprintf("Using UBI Java package %s", buildver))

		// Create build.Dockerfile
		buildDockerfileProps := BuildDockerfileProps{
			JAVA_VERSION: JAVA_VERSION,
			CNB_USER_ID:  CNB_USER_ID,
			CNB_GROUP_ID: CNB_GROUP_ID,
			CNB_STACK_ID: packitContext.Stack,
			PACKAGES:     "openssl-devel " + buildver + " nss_wrapper which",
		}
		buildDockerfileContent, err := FillPropsToTemplate(buildDockerfileProps, buildDockerfileTemplate)
		if err != nil {
			return packit.GenerateResult{}, err
		}

		// Create run.Dockerfile
		RunDockerfileProps := RunDockerfileProps{
			Source: runver,
		}
		runDockerfileContent, err := FillPropsToTemplate(RunDockerfileProps, runDockerfileTemplate)
		if err != nil {
			return packit.GenerateResult{}, err
		}

		//exit with success.
		return packit.GenerateResult{
			ExtendConfig:    packit.ExtendConfig{Build: packit.ExtendImageConfig{Args: []packit.ExtendImageConfigArg{}}},
			BuildDockerfile: strings.NewReader(buildDockerfileContent),
			RunDockerfile:   strings.NewReader(runDockerfileContent),
		}, nil
	}
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

func getExtension(ctx packit.GenerateContext) (extension Extension, err error) {

	extensionFile := filepath.Join(ctx.CNBPath, "extension.toml")
	if _, err = toml.DecodeFile(extensionFile, &extension); err != nil && !os.IsNotExist(err) {
		return extension, fmt.Errorf("unable to decode extension content %s\n%w", extensionFile, err)
	}

	return extension, nil
}

func getJDKJRERequired(plan packit.BuildpackPlan) (jdkRequired bool, jreRequired bool, err error) {

	var cnbplan libcnb.BuildpackPlan

	//convert packit plan into libcnb plan
	//thankfully simple as metadta shares same type =)
	for _, bpe := range plan.Entries {
		cnbplan.Entries = append(cnbplan.Entries, libcnb.BuildpackPlanEntry{Name: bpe.Name, Metadata: bpe.Metadata})
	}

	//use libpak plan resolver to test for jdk/jre requests in the plan.
	pr := libpak.PlanEntryResolver{Plan: cnbplan}
	_, jdkRequired, err = pr.Resolve("jdk")
	if err != nil {
		return false, false, fmt.Errorf("error during resolution of jdk plan entry\n%w", err)
	}
	_, jreRequired, err = pr.Resolve("jre")
	if err != nil {
		return false, false, fmt.Errorf("error during resolution of jre plan entry\n%w", err)
	}

	return jdkRequired, jreRequired, nil
}

func mapRequestedVersionToPackageAndRunImage(requestedVersion string) (packages string, runImage string, err error) {
	//TODO: extenalise mappings via extension.toml ?
	//TODO: consider sloppy version matching, with next ver up migration.
	var buildver, runver string
	switch requestedVersion {
	case "8", "1.8", "1.8.0":
		buildver = "java-1.8.0-openjdk-devel"
		runver = "paketo-buildpacks/ubi8-paketo-run-java-8"
	case "11":
		buildver = "java-11-openjdk-devel"
		runver = "paketo-buildpacks/ubi8-paketo-run-java-11"
	case "17":
		buildver = "java-17-openjdk-devel"
		runver = "paketo-buildpacks/ubi8-paketo-run-java-17"
	default:
		buildver = ""
		runver = ""
		err = fmt.Errorf("Unable to map requested Java version of %s to a UBI supported runtime", requestedVersion)
	}
	return buildver, runver, err
}

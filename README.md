# Paketo Java Extension for UBI

The Java Extension for
[UBI](https://www.redhat.com/en/blog/introducing-red-hat-universal-base-image)
allows builders to be created that can build Java applications on top of
Red Hat's Java UBI containers. For example
[ubi8/openjdk-21-runtime](https://catalog.redhat.com/software/containers/ubi8/openjdk-21-runtime/653fd184292263c0a2f14d69).

## Integration

The UBI Java extension provides a jdk/jre as dependencies.
Downstream buildpacks, like Maven or Gradle, can
require the jdk/jre dependency by generating a Build Plan TOML
file that requires `jre` and or `jdk`.

The extension integrates with the existing Paketo buildpacks
so that building your application will have the same experience
as building with non ubi stacks. The main difference is that
the Java runtime will be provided by the extension instead of the
appropriate build pack.

## Behavior

This extension will participate if any of the following conditions are met

* Another buildpack requires `jdk`
* Another buildpack requires `jre`

The extension will do the following if a JDK or JRE is requested:

* Installs a JDK to the Builder image using `dnf` with all commands available via `$PATH`
* Installs a JRE to the Run image by switching the run image to one with it pre-installed.
    * `$JAVA_HOME` will be configured in the run image.

## Configuration

| Environment Variable          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| ----------------------------- |-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `$BP_JVM_VERSION`             | Configure the JVM version (e.g. `8`, `11`, `17`, `21`).  The extension will install rpms that provide a level of Java that is compatible with this version of the JVM specification.  UBI only provides a single version of each supported line, patch releases etc can change the exact version of the JDK or JRE. Builds will be performed with whatever the current version for the select specification is within the UBI release stream. |


## Limitations.

* Due to CNCF builder image extension limitations, this extension does not install any helpers, this currently means no support for most of the buildpack java configuration env vars (all beyond `BP_JVM_VERSION`).

* No support for JLink. As the Runtime image is based from a UBI Java Runtime Image, that contains the appropriate Java Runtime pre-installed, this extension will not support JLink, as it is not possible to modify the JVM within the Runtime image. 


## Future goals.

* Companion buildpack, this buildpack would `provide` the dependency `java-postconfig-required` and a companion buildpack would `require` the same. (The odd inverse dependency is because extensions are not allowed to use `require`). The companion buildpack would then be selected into builds where the extension is present, and would be able to perform post configuration tasks via layers (extensions also cannot create layers). 

* Support for private rpm repository configuration. Currently UBI Builders will require access to the internet to pull their rpms from the standard repositories. Future support may be added to allow configuration of a locally hosted repository that the Builders can use instead. Because the installing dependencies via dnf can require additional transitive dependencies, it's not planned to allow passing rpms directly to the extension, but instead to prefer the route of having a local repository. 


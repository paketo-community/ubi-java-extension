package ubi8javaextension_test

import (
	"testing"

	ubi8javaextension "github.com/BarDweller/ubi-java-extension"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
		dc     packit.DetectContext
		result packit.DetectResult
		err    error
	)

	context("Detect Result Check", func() {
		it.Before(func() {
			dc = packit.DetectContext{}
		})
		it("includes build plan options", func() {
			result, err = ubi8javaextension.Detect()(dc)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(
				packit.DetectResult{
					Plan: packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: "jdk"},
							{Name: "jre"},
						},
						Or: []packit.BuildPlan{
							{
								Provides: []packit.BuildPlanProvision{
									{Name: "jdk"},
								},
							},
							{
								Provides: []packit.BuildPlanProvision{
									{Name: "jre"},
								},
							},
						},
					},
				}))

		})
	})
}

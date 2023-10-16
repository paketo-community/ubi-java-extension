package ubi8javaextension_test

import (
	"testing"

	libcnb "github.com/buildpacks/libcnb/v2"
	ubi8javaextension "github.com/paketo-community/ubi-java-extension/v1"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/libjvm/v2"
	"github.com/sclevine/spec"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect
		dc     libcnb.DetectContext
		result libcnb.DetectResult
		err    error
	)

	context("Detect Result Check", func() {
		it.Before(func() {
			dc = libcnb.DetectContext{}
		})
		it("includes build plan options", func() {
			result, err = ubi8javaextension.Detect(dc)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(
				libcnb.DetectResult{
					Pass: true,
					Plans: []libcnb.BuildPlan{
						{
							Provides: []libcnb.BuildPlanProvide{
								{Name: libjvm.PlanEntryJDK},
								{Name: libjvm.PlanEntryJRE},
							},
						},
						{
							Provides: []libcnb.BuildPlanProvide{
								{Name: libjvm.PlanEntryJDK},
							},
						},
						{
							Provides: []libcnb.BuildPlanProvide{
								{Name: libjvm.PlanEntryJRE},
							},
						},
					},
				}))
		})
	})
}

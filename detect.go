package ubi8javaextension

import (
	"github.com/paketo-buildpacks/packit/v2"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		return packit.DetectResult{
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
		}, nil
	}
}

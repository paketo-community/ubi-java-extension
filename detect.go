package ubi8javaextension

import (
	libcnb "github.com/buildpacks/libcnb/v2"
	"github.com/paketo-buildpacks/libjvm/v2"
)

func Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	return libcnb.DetectResult{
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
	}, nil
}

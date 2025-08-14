package imageoptimizerdefaults_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fastly/go-fastly/v11/fastly"

	root "github.com/fastly/cli/pkg/commands/imageoptimizerdefaults"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestImageOptimizerDefaultsUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--version 1",
			WantError: "error parsing arguments: required flag --service-id not provided",
		},
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:                             testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn:      updateImageOptimizerDefaultsOK,
			},
			WantOutput: "Updated Image Optimizer default settings for service 123 (version 1)\n\nAllow Video: false\nJPEG Quality: 85\nJPEG Type: auto\nResize Filter: lanczos3\nUpscale: false\nWebP: false\nWebP Quality: 85\n",
		},
		{
			Args: "--service-id 123 --version 1 --webp true --upscale false --allow-video true",
			API: mock.API{
				ListVersionsFn:                             testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn:      updateImageOptimizerDefaultsWithBoolsOK,
			},
			WantOutput: "Updated Image Optimizer default settings for service 123 (version 1)\n\nAllow Video: true\nJPEG Quality: 85\nJPEG Type: auto\nResize Filter: lanczos3\nUpscale: false\nWebP: true\nWebP Quality: 85\n",
		},
		{
			Args: "--service-id 123 --version 1 --resize-filter bicubic --webp-quality 90 --jpeg-quality 80",
			API: mock.API{
				ListVersionsFn:                             testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn:      updateImageOptimizerDefaultsWithOptionsOK,
			},
			WantOutput: "Updated Image Optimizer default settings for service 123 (version 1)\n\nAllow Video: false\nJPEG Quality: 80\nJPEG Type: auto\nResize Filter: bicubic\nUpscale: false\nWebP: false\nWebP Quality: 90\n",
		},
		{
			Args: "--service-id 123 --version 1 --webp invalid",
			API: mock.API{
				ListVersionsFn:                        testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn: updateImageOptimizerDefaultsOK,
			},
			WantError: "'webp' flag must be one of the following [true, false]",
		},
		{
			Args: "--service-id 123 --version 1 --upscale invalid",
			API: mock.API{
				ListVersionsFn:                        testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn: updateImageOptimizerDefaultsOK,
			},
			WantError: "'upscale' flag must be one of the following [true, false]",
		},
		{
			Args: "--service-id 123 --version 1 --allow-video invalid",
			API: mock.API{
				ListVersionsFn:                        testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn: updateImageOptimizerDefaultsOK,
			},
			WantError: "'allow-video' flag must be one of the following [true, false]",
		},
		{
			Args: "--service-id 123 --version 1 --resize-filter invalid",
			API: mock.API{
				ListVersionsFn:                        testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn: updateImageOptimizerDefaultsOK,
			},
			WantError: "invalid resize filter: invalid. Valid options: lanczos3, lanczos2, bicubic, bilinear, nearest",
		},
		{
			Args: "--service-id 123 --version 1 --jpeg-type invalid",
			API: mock.API{
				ListVersionsFn:                        testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn: updateImageOptimizerDefaultsOK,
			},
			WantError: "invalid jpeg type: invalid. Valid options: auto, baseline, progressive",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:                             testutil.ListVersions,
				UpdateImageOptimizerDefaultSettingsFn:      updateImageOptimizerDefaultsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func updateImageOptimizerDefaultsOK(_ context.Context, i *fastly.UpdateImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error) {
	return &fastly.ImageOptimizerDefaultSettings{
		ResizeFilter: "lanczos3",
		Webp:         false,
		WebpQuality:  85,
		JpegType:     "auto",
		JpegQuality:  85,
		Upscale:      false,
		AllowVideo:   false,
	}, nil
}

func updateImageOptimizerDefaultsWithBoolsOK(_ context.Context, i *fastly.UpdateImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error) {
	return &fastly.ImageOptimizerDefaultSettings{
		ResizeFilter: "lanczos3",
		Webp:         fastly.ToValue(i.Webp),
		WebpQuality:  85,
		JpegType:     "auto",
		JpegQuality:  85,
		Upscale:      fastly.ToValue(i.Upscale),
		AllowVideo:   fastly.ToValue(i.AllowVideo),
	}, nil
}

func updateImageOptimizerDefaultsWithOptionsOK(_ context.Context, i *fastly.UpdateImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error) {
	resizeFilter := "bicubic"
	if i.ResizeFilter != nil {
		switch *i.ResizeFilter {
		case fastly.ImageOptimizerLanczos3:
			resizeFilter = "lanczos3"
		case fastly.ImageOptimizerLanczos2:
			resizeFilter = "lanczos2"
		case fastly.ImageOptimizerBicubic:
			resizeFilter = "bicubic"
		case fastly.ImageOptimizerBilinear:
			resizeFilter = "bilinear"
		case fastly.ImageOptimizerNearest:
			resizeFilter = "nearest"
		}
	}
	jpegType := "auto"
	if i.JpegType != nil {
		switch *i.JpegType {
		case fastly.ImageOptimizerAuto:
			jpegType = "auto"
		case fastly.ImageOptimizerBaseline:
			jpegType = "baseline"
		case fastly.ImageOptimizerProgressive:
			jpegType = "progressive"
		}
	}
	return &fastly.ImageOptimizerDefaultSettings{
		ResizeFilter: resizeFilter,
		Webp:         false,
		WebpQuality:  fastly.ToValue(i.WebpQuality),
		JpegType:     jpegType,
		JpegQuality:  fastly.ToValue(i.JpegQuality),
		Upscale:      false,
		AllowVideo:   false,
	}, nil
}

func updateImageOptimizerDefaultsError(_ context.Context, _ *fastly.UpdateImageOptimizerDefaultSettingsInput) (*fastly.ImageOptimizerDefaultSettings, error) {
	return nil, errTest
}

var errTest = errors.New("an expected error occurred")
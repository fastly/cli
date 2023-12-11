package s3_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/s3"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateS3Input(t *testing.T) {
	red := fastly.S3RedundancyStandard
	sse := fastly.S3ServerSideEncryptionAES
	for _, testcase := range []struct {
		name      string
		cmd       *s3.CreateCommand
		want      *fastly.CreateS3Input
		wantError string
	}{
		{
			name: "required values set flag serviceID using access credentials",
			cmd:  createCommandRequired(),
			want: &fastly.CreateS3Input{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				BucketName:     fastly.ToPointer("bucket"),
				AccessKey:      fastly.ToPointer("access"),
				SecretKey:      fastly.ToPointer("secret"),
			},
		},
		{
			name: "required values set flag serviceID using IAM role",
			cmd:  createCommandRequiredIAMRole(),
			want: &fastly.CreateS3Input{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				BucketName:     fastly.ToPointer("bucket"),
				IAMRole:        fastly.ToPointer("arn:aws:iam::123456789012:role/S3Access"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateS3Input{
				ServiceID:                    "123",
				ServiceVersion:               4,
				Name:                         fastly.ToPointer("logs"),
				BucketName:                   fastly.ToPointer("bucket"),
				Domain:                       fastly.ToPointer("domain"),
				AccessKey:                    fastly.ToPointer("access"),
				SecretKey:                    fastly.ToPointer("secret"),
				Path:                         fastly.ToPointer("path"),
				Period:                       fastly.ToPointer(3600),
				Format:                       fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
				MessageType:                  fastly.ToPointer("classic"),
				FormatVersion:                fastly.ToPointer(2),
				ResponseCondition:            fastly.ToPointer("Prevent default logging"),
				TimestampFormat:              fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
				Redundancy:                   &red,
				Placement:                    fastly.ToPointer("none"),
				PublicKey:                    fastly.ToPointer(pgpPublicKey()),
				ServerSideEncryptionKMSKeyID: fastly.ToPointer("kmskey"),
				ServerSideEncryption:         &sse,
				CompressionCodec:             fastly.ToPointer("zstd"),
			},
		},
		{
			name:      "error missing serviceID",
			cmd:       createCommandMissingServiceID(),
			want:      nil,
			wantError: errors.ErrNoServiceID.Error(),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.cmd.Globals.APIClient,
				Manifest:           testcase.cmd.Manifest,
				Out:                out,
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        verboseMode,
			})

			switch {
			case err != nil && testcase.wantError == "":
				t.Fatalf("unexpected error getting service details: %v", err)
				return
			case err != nil && testcase.wantError != "":
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			case err == nil && testcase.wantError != "":
				t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
			case err == nil && testcase.wantError == "":
				have, err := testcase.cmd.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
				testutil.AssertErrorContains(t, err, testcase.wantError)
				testutil.AssertEqual(t, testcase.want, have)
			}
		})
	}
}

func TestUpdateS3Input(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *s3.UpdateCommand
		api       mock.API
		want      *fastly.UpdateS3Input
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetS3Fn:        getS3OK,
			},
			want: &fastly.UpdateS3Input{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetS3Fn:        getS3OK,
			},
			want: &fastly.UpdateS3Input{
				ServiceID:                    "123",
				ServiceVersion:               4,
				Name:                         "log",
				NewName:                      fastly.ToPointer("new1"),
				BucketName:                   fastly.ToPointer("new2"),
				AccessKey:                    fastly.ToPointer("new3"),
				SecretKey:                    fastly.ToPointer("new4"),
				IAMRole:                      fastly.ToPointer(""),
				Domain:                       fastly.ToPointer("new5"),
				Path:                         fastly.ToPointer("new6"),
				Period:                       fastly.ToPointer(3601),
				GzipLevel:                    fastly.ToPointer(0),
				Format:                       fastly.ToPointer("new7"),
				FormatVersion:                fastly.ToPointer(3),
				MessageType:                  fastly.ToPointer("new8"),
				ResponseCondition:            fastly.ToPointer("new9"),
				TimestampFormat:              fastly.ToPointer("new10"),
				Placement:                    fastly.ToPointer("new11"),
				Redundancy:                   fastly.ToPointer(fastly.S3RedundancyReduced),
				ServerSideEncryption:         fastly.ToPointer(fastly.S3ServerSideEncryptionKMS),
				ServerSideEncryptionKMSKeyID: fastly.ToPointer("new12"),
				PublicKey:                    fastly.ToPointer("new13"),
				CompressionCodec:             fastly.ToPointer("new14"),
			},
		},
		{
			name:      "error missing serviceID",
			cmd:       updateCommandMissingServiceID(),
			want:      nil,
			wantError: errors.ErrNoServiceID.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.name, func(t *testing.T) {
			testcase.cmd.Globals.APIClient = testcase.api

			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.api,
				Manifest:           testcase.cmd.Manifest,
				Out:                out,
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        verboseMode,
			})

			switch {
			case err != nil && testcase.wantError == "":
				t.Fatalf("unexpected error getting service details: %v", err)
				return
			case err != nil && testcase.wantError != "":
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			case err == nil && testcase.wantError != "":
				t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
			case err == nil && testcase.wantError == "":
				have, err := testcase.cmd.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
				testutil.AssertErrorContains(t, err, testcase.wantError)
				testutil.AssertEqual(t, testcase.want, have)
			}
		})
	}
}

func createCommandRequired() *s3.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &s3.CreateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		BucketName:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bucket"},
		AccessKey:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "access"},
		SecretKey:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "secret"},
	}
}

func createCommandRequiredIAMRole() *s3.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &s3.CreateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		BucketName:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bucket"},
		IAMRole:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "arn:aws:iam::123456789012:role/S3Access"},
	}
}

func createCommandAll() *s3.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &s3.CreateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName:                 argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		BucketName:                   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bucket"},
		AccessKey:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "access"},
		SecretKey:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "secret"},
		Domain:                       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "domain"},
		Path:                         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "path"},
		Period:                       argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3600},
		Format:                       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:                argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		MessageType:                  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		PublicKey:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: pgpPublicKey()},
		Redundancy:                   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: string(fastly.S3RedundancyStandard)},
		ServerSideEncryption:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: string(fastly.S3ServerSideEncryptionAES)},
		ServerSideEncryptionKMSKeyID: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "kmskey"},
		CompressionCodec:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "zstd"},
	}
}

func createCommandMissingServiceID() *s3.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *s3.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &s3.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
	}
}

func updateCommandAll() *s3.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &s3.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		NewName:                      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new1"},
		BucketName:                   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		AccessKey:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		SecretKey:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		IAMRole:                      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: ""},
		Domain:                       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		Path:                         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		Period:                       argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3601},
		GzipLevel:                    argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 0},
		Format:                       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:                argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		MessageType:                  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		ResponseCondition:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		TimestampFormat:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		Placement:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		Redundancy:                   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: string(fastly.S3RedundancyReduced)},
		ServerSideEncryption:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: string(fastly.S3ServerSideEncryptionKMS)},
		ServerSideEncryptionKMSKeyID: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new12"},
		PublicKey:                    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new13"},
		CompressionCodec:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new14"},
	}
}

func updateCommandMissingServiceID() *s3.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func TestValidateRedundancy(t *testing.T) {
	for _, testcase := range []struct {
		value     string
		want      fastly.S3Redundancy
		wantError string
	}{
		{value: "standard", want: fastly.S3RedundancyStandard},
		{value: "standard_ia", want: fastly.S3RedundancyStandardIA},
		{value: "onezone_ia", want: fastly.S3RedundancyOneZoneIA},
		{value: "glacier", want: fastly.S3RedundancyGlacierFlexibleRetrieval},
		{value: "glacier_ir", want: fastly.S3RedundancyGlacierInstantRetrieval},
		{value: "deep_archive", want: fastly.S3RedundancyGlacierDeepArchive},
		{value: "reduced_redundancy", want: fastly.S3RedundancyReduced},
		{value: "bad_value", wantError: "unknown redundancy"},
	} {
		t.Run(testcase.value, func(t *testing.T) {
			have, err := s3.ValidateRedundancy(testcase.value)

			switch {
			case err != nil && testcase.wantError == "":
				t.Fatalf("unexpected error ValidateRedundancy: %v", err)
				return
			case err != nil && testcase.wantError != "":
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			case err == nil && testcase.wantError != "":
				t.Fatalf("expected error, have nil (redundancy: %s)", testcase.value)
			case err == nil && testcase.wantError == "":
				testutil.AssertEqual(t, testcase.want, have)
			}
		})
	}
}

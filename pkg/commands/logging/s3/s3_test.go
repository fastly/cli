package s3_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/s3"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v7/fastly"
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
				Name:           fastly.String("log"),
				BucketName:     fastly.String("bucket"),
				AccessKey:      fastly.String("access"),
				SecretKey:      fastly.String("secret"),
			},
		},
		{
			name: "required values set flag serviceID using IAM role",
			cmd:  createCommandRequiredIAMRole(),
			want: &fastly.CreateS3Input{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.String("log"),
				BucketName:     fastly.String("bucket"),
				IAMRole:        fastly.String("arn:aws:iam::123456789012:role/S3Access"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateS3Input{
				ServiceID:                    "123",
				ServiceVersion:               4,
				Name:                         fastly.String("logs"),
				BucketName:                   fastly.String("bucket"),
				Domain:                       fastly.String("domain"),
				AccessKey:                    fastly.String("access"),
				SecretKey:                    fastly.String("secret"),
				Path:                         fastly.String("path"),
				Period:                       fastly.Int(3600),
				Format:                       fastly.String(`%h %l %u %t "%r" %>s %b`),
				MessageType:                  fastly.String("classic"),
				FormatVersion:                fastly.Int(2),
				ResponseCondition:            fastly.String("Prevent default logging"),
				TimestampFormat:              fastly.String("%Y-%m-%dT%H:%M:%S.000"),
				Redundancy:                   &red,
				Placement:                    fastly.String("none"),
				PublicKey:                    fastly.String(pgpPublicKey()),
				ServerSideEncryptionKMSKeyID: fastly.String("kmskey"),
				ServerSideEncryption:         &sse,
				CompressionCodec:             fastly.String("zstd"),
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

			serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
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
				have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
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
				NewName:                      fastly.String("new1"),
				BucketName:                   fastly.String("new2"),
				AccessKey:                    fastly.String("new3"),
				SecretKey:                    fastly.String("new4"),
				IAMRole:                      fastly.String(""),
				Domain:                       fastly.String("new5"),
				Path:                         fastly.String("new6"),
				Period:                       fastly.Int(3601),
				GzipLevel:                    fastly.Int(0),
				Format:                       fastly.String("new7"),
				FormatVersion:                fastly.Int(3),
				MessageType:                  fastly.String("new8"),
				ResponseCondition:            fastly.String("new9"),
				TimestampFormat:              fastly.String("new10"),
				Placement:                    fastly.String("new11"),
				Redundancy:                   fastly.S3RedundancyPtr(fastly.S3RedundancyReduced),
				ServerSideEncryption:         fastly.S3ServerSideEncryptionPtr(fastly.S3ServerSideEncryptionKMS),
				ServerSideEncryptionKMSKeyID: fastly.String("new12"),
				PublicKey:                    fastly.String("new13"),
				CompressionCodec:             fastly.String("new14"),
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

			serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
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
				have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
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
	})("token", "endpoint")

	return &s3.CreateCommand{
		Base: cmd.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "log"},
		BucketName:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "bucket"},
		AccessKey:    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "access"},
		SecretKey:    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "secret"},
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
	})("token", "endpoint")

	return &s3.CreateCommand{
		Base: cmd.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "log"},
		BucketName:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "bucket"},
		IAMRole:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "arn:aws:iam::123456789012:role/S3Access"},
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
	})("token", "endpoint")

	return &s3.CreateCommand{
		Base: cmd.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName:                 cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "logs"},
		BucketName:                   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "bucket"},
		AccessKey:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "access"},
		SecretKey:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "secret"},
		Domain:                       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "domain"},
		Path:                         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "path"},
		Period:                       cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3600},
		Format:                       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:                cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 2},
		MessageType:                  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		PublicKey:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: pgpPublicKey()},
		Redundancy:                   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: string(fastly.S3RedundancyStandard)},
		ServerSideEncryption:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: string(fastly.S3ServerSideEncryptionAES)},
		ServerSideEncryptionKMSKeyID: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "kmskey"},
		CompressionCodec:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "zstd"},
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
		Base: cmd.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
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
		Base: cmd.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		NewName:                      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new1"},
		BucketName:                   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		AccessKey:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		SecretKey:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		IAMRole:                      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: ""},
		Domain:                       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Path:                         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		Period:                       cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3601},
		GzipLevel:                    cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 0},
		Format:                       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:                cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3},
		MessageType:                  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		ResponseCondition:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		TimestampFormat:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		Placement:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
		Redundancy:                   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: string(fastly.S3RedundancyReduced)},
		ServerSideEncryption:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: string(fastly.S3ServerSideEncryptionKMS)},
		ServerSideEncryptionKMSKeyID: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new12"},
		PublicKey:                    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new13"},
		CompressionCodec:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new14"},
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

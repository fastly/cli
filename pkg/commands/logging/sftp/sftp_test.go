package sftp_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/sftp"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateSFTPInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *sftp.CreateCommand
		want      *fastly.CreateSFTPInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateSFTPInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				Address:        fastly.ToPointer("127.0.0.1"),
				User:           fastly.ToPointer("user"),
				SSHKnownHosts:  fastly.ToPointer(knownHosts()),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateSFTPInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.ToPointer("log"),
				Address:           fastly.ToPointer("127.0.0.1"),
				Port:              fastly.ToPointer(80),
				User:              fastly.ToPointer("user"),
				Password:          fastly.ToPointer("password"),
				PublicKey:         fastly.ToPointer(pgpPublicKey()),
				SecretKey:         fastly.ToPointer(sshPrivateKey()),
				SSHKnownHosts:     fastly.ToPointer(knownHosts()),
				Path:              fastly.ToPointer("/log"),
				Period:            fastly.ToPointer(3600),
				FormatVersion:     fastly.ToPointer(2),
				Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
				ResponseCondition: fastly.ToPointer("Prevent default logging"),
				MessageType:       fastly.ToPointer("classic"),
				TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
				CompressionCodec:  fastly.ToPointer("zstd"),
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

func TestUpdateSFTPInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *sftp.UpdateCommand
		api       mock.API
		want      *fastly.UpdateSFTPInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetSFTPFn:      getSFTPOK,
			},
			want: &fastly.UpdateSFTPInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.ToPointer("new1"),
				Address:           fastly.ToPointer("new2"),
				Port:              fastly.ToPointer(81),
				User:              fastly.ToPointer("new3"),
				SSHKnownHosts:     fastly.ToPointer("new4"),
				Password:          fastly.ToPointer("new5"),
				PublicKey:         fastly.ToPointer("new6"),
				SecretKey:         fastly.ToPointer("new7"),
				Path:              fastly.ToPointer("new8"),
				Period:            fastly.ToPointer(3601),
				FormatVersion:     fastly.ToPointer(3),
				GzipLevel:         fastly.ToPointer(0),
				Format:            fastly.ToPointer("new9"),
				ResponseCondition: fastly.ToPointer("new10"),
				TimestampFormat:   fastly.ToPointer("new11"),
				MessageType:       fastly.ToPointer("new13"),
				CompressionCodec:  fastly.ToPointer("new14"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetSFTPFn:      getSFTPOK,
			},
			want: &fastly.UpdateSFTPInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
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

func createCommandRequired() *sftp.CreateCommand {
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

	return &sftp.CreateCommand{
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
		EndpointName:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		Address:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "127.0.0.1"},
		User:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "user"},
		SSHKnownHosts: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: knownHosts()},
	}
}

func createCommandAll() *sftp.CreateCommand {
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

	return &sftp.CreateCommand{
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
		EndpointName:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		Address:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "127.0.0.1"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "user"},
		SSHKnownHosts:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: knownHosts()},
		Port:              argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 80},
		Password:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "password"},
		PublicKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: pgpPublicKey()},
		SecretKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: sshPrivateKey()},
		Path:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "/log"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3600},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "zstd"},
	}
}

func createCommandMissingServiceID() *sftp.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *sftp.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &sftp.UpdateCommand{
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

func updateCommandAll() *sftp.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &sftp.UpdateCommand{
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
		NewName:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new1"},
		Address:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		SSHKnownHosts:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		Port:              argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 81},
		Password:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		PublicKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		SecretKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		Path:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3601},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		GzipLevel:         argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 0},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new13"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new14"},
	}
}

func updateCommandMissingServiceID() *sftp.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

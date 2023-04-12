package sftp_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/sftp"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v8/fastly"
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
				Name:           fastly.String("log"),
				Address:        fastly.String("127.0.0.1"),
				User:           fastly.String("user"),
				SSHKnownHosts:  fastly.String(knownHosts()),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateSFTPInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.String("log"),
				Address:           fastly.String("127.0.0.1"),
				Port:              fastly.Int(80),
				User:              fastly.String("user"),
				Password:          fastly.String("password"),
				PublicKey:         fastly.String(pgpPublicKey()),
				SecretKey:         fastly.String(sshPrivateKey()),
				SSHKnownHosts:     fastly.String(knownHosts()),
				Path:              fastly.String("/log"),
				Period:            fastly.Int(3600),
				FormatVersion:     fastly.Int(2),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				ResponseCondition: fastly.String("Prevent default logging"),
				MessageType:       fastly.String("classic"),
				TimestampFormat:   fastly.String("%Y-%m-%dT%H:%M:%S.000"),
				Placement:         fastly.String("none"),
				CompressionCodec:  fastly.String("zstd"),
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
				NewName:           fastly.String("new1"),
				Address:           fastly.String("new2"),
				Port:              fastly.Int(81),
				User:              fastly.String("new3"),
				SSHKnownHosts:     fastly.String("new4"),
				Password:          fastly.String("new5"),
				PublicKey:         fastly.String("new6"),
				SecretKey:         fastly.String("new7"),
				Path:              fastly.String("new8"),
				Period:            fastly.Int(3601),
				FormatVersion:     fastly.Int(3),
				GzipLevel:         fastly.Int(0),
				Format:            fastly.String("new9"),
				ResponseCondition: fastly.String("new10"),
				TimestampFormat:   fastly.String("new11"),
				Placement:         fastly.String("new12"),
				MessageType:       fastly.String("new13"),
				CompressionCodec:  fastly.String("new14"),
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
	})("token", "endpoint")

	return &sftp.CreateCommand{
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
		EndpointName:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "log"},
		Address:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "127.0.0.1"},
		User:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "user"},
		SSHKnownHosts: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: knownHosts()},
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
	})("token", "endpoint")

	return &sftp.CreateCommand{
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
		EndpointName:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "log"},
		Address:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "127.0.0.1"},
		User:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "user"},
		SSHKnownHosts:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: knownHosts()},
		Port:              cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 80},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "password"},
		PublicKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: pgpPublicKey()},
		SecretKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: sshPrivateKey()},
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "/log"},
		Period:            cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3600},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 2},
		MessageType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "zstd"},
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

func updateCommandAll() *sftp.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &sftp.UpdateCommand{
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
		NewName:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new1"},
		Address:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		User:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		SSHKnownHosts:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		Port:              cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 81},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		PublicKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		SecretKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		Period:            cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3601},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3},
		GzipLevel:         cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 0},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new12"},
		MessageType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new13"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new14"},
	}
}

func updateCommandMissingServiceID() *sftp.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

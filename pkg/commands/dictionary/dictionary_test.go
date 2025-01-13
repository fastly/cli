package dictionary_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/dictionary"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestDictionaryDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--version 1 --service-id 123",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--version 1 --service-id 123 --name dict-1",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetDictionaryFn: describeDictionaryOK,
			},
			WantOutput: describeDictionaryOutput,
		},
		{
			Args: "--version 1 --service-id 123 --name dict-1",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetDictionaryFn: describeDictionaryOKDeleted,
			},
			WantOutput: describeDictionaryOutputDeleted,
		},
		{
			Args: "--version 1 --service-id 123 --name dict-1 --verbose",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				GetDictionaryFn:       describeDictionaryOK,
				GetDictionaryInfoFn:   getDictionaryInfoOK,
				ListDictionaryItemsFn: listDictionaryItemsOK,
			},
			WantOutput: describeDictionaryOutputVerbose,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestDictionaryCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--version 1",
			WantError: "error reading service: no service ID found",
		},
		{
			Args: "--version 1 --service-id 123 --name denylist --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateDictionaryFn: createDictionaryOK,
			},
			WantOutput: createDictionaryOutput,
		},
		{
			Args: "--version 1 --service-id 123 --name denylist --write-only --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateDictionaryFn: createDictionaryOK,
			},
			WantOutput: createDictionaryOutputWriteOnly,
		},
		{
			Args: "--version 1 --service-id 123 --name denylist --write-only fish --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: unexpected 'fish'",
		},
		{
			Args: "--version 1 --service-id 123 --name denylist --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateDictionaryFn: createDictionaryDuplicate,
			},
			WantError: "Duplicate record",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDeleteDictionary(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name allowlist --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteDictionaryFn: deleteDictionaryOK,
			},
			WantOutput: deleteDictionaryOutput,
		},
		{
			Args: "--service-id 123 --version 1 --name allowlist --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteDictionaryFn: deleteDictionaryError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestListDictionary(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--version 1",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListDictionariesFn: listDictionariesOk,
			},
			WantError: "error reading service: no service ID found",
		},
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--version 1 --service-id 123",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListDictionariesFn: listDictionariesOk,
			},
			WantOutput: listDictionariesOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestUpdateDictionary(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--version 1 --name oldname --new-name newname",
			WantError: "error reading service: no service ID found",
		},
		{
			Args:      "--service-id 123 --name oldname --new-name newname",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args:      "--service-id 123 --version 1 --new-name newname",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name oldname --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: required flag --new-name or --write-only not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name oldname --new-name dict-1 --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateDictionaryFn: updateDictionaryNameOK,
			},
			WantOutput: updateDictionaryNameOutput,
		},
		{
			Args: "--service-id 123 --version 1 --name oldname --new-name dict-1 --write-only true --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateDictionaryFn: updateDictionaryNameOK,
			},
			WantOutput: updateDictionaryNameOutput,
		},
		{
			Args: "--service-id 123 --version 1 --name oldname --write-only true --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateDictionaryFn: updateDictionaryWriteOnlyOK,
			},
			WantOutput: updateDictionaryOutput,
		},
		{
			Args: "-v --service-id 123 --version 1 --name oldname --new-name dict-1 --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateDictionaryFn: updateDictionaryNameOK,
			},
			WantOutput: updateDictionaryOutputVerbose,
		},
		{
			Args: "--service-id 123 --version 1 --name oldname --new-name dict-1 --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateDictionaryFn: updateDictionaryError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func describeDictionaryOK(i *fastly.GetDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           fastly.ToPointer(i.Name),
		CreatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly:      fastly.ToPointer(false),
		DictionaryID:   fastly.ToPointer("456"),
		UpdatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func describeDictionaryOKDeleted(i *fastly.GetDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           fastly.ToPointer(i.Name),
		CreatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly:      fastly.ToPointer(false),
		DictionaryID:   fastly.ToPointer("456"),
		UpdatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		DeletedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:08Z"),
	}, nil
}

func createDictionaryOK(i *fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	if i.WriteOnly == nil {
		i.WriteOnly = fastly.ToPointer(fastly.Compatibool(false))
	}
	return &fastly.Dictionary{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
		CreatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly:      fastly.ToPointer(bool(fastly.ToValue(i.WriteOnly))),
		DictionaryID:   fastly.ToPointer("456"),
		UpdatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

// getDictionaryInfoOK mocks the response from fastly.GetDictionaryInfo, which
// is not otherwise used in the fastly-cli and will need to be updated here if
// that call changes. This function requires i.ID to equal "456" to enforce the
// input to this call matches the response to GetDictionaryInfo in
// describeDictionaryOK.
func getDictionaryInfoOK(i *fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error) {
	if i.DictionaryID == "456" {
		return &fastly.DictionaryInfo{
			ItemCount:   fastly.ToPointer(2),
			LastUpdated: testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
			Digest:      fastly.ToPointer("digest_hash"),
		}, nil
	}
	return nil, errFail
}

// listDictionaryItemsOK mocks the response from fastly.ListDictionaryItems
// which is primarily used in the fastly-cli.dictionaryitem package and will
// need to be updated here if that call changes.
func listDictionaryItemsOK(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
	return []*fastly.DictionaryItem{
		{
			ServiceID:    fastly.ToPointer(i.ServiceID),
			DictionaryID: fastly.ToPointer(i.DictionaryID),
			ItemKey:      fastly.ToPointer("foo"),
			ItemValue:    fastly.ToPointer("bar"),
			CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		},
		{
			ServiceID:    fastly.ToPointer(i.ServiceID),
			DictionaryID: fastly.ToPointer(i.DictionaryID),
			ItemKey:      fastly.ToPointer("baz"),
			ItemValue:    fastly.ToPointer("bear"),
			CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
			DeletedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:06:08Z"),
		},
	}, nil
}

func createDictionaryDuplicate(*fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	return nil, errors.New("Duplicate record")
}

func deleteDictionaryOK(*fastly.DeleteDictionaryInput) error {
	return nil
}

func deleteDictionaryError(*fastly.DeleteDictionaryInput) error {
	return errTest
}

func listDictionariesOk(i *fastly.ListDictionariesInput) ([]*fastly.Dictionary, error) {
	return []*fastly.Dictionary{
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("dict-1"),
			CreatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			WriteOnly:      fastly.ToPointer(false),
			DictionaryID:   fastly.ToPointer("456"),
			UpdatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		},
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("dict-2"),
			CreatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			WriteOnly:      fastly.ToPointer(false),
			DictionaryID:   fastly.ToPointer("456"),
			UpdatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		},
	}, nil
}

func updateDictionaryNameOK(i *fastly.UpdateDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.NewName,
		CreatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly:      fastly.ToPointer(bool(fastly.ToValue(i.WriteOnly))),
		DictionaryID:   fastly.ToPointer("456"),
		UpdatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func updateDictionaryWriteOnlyOK(i *fastly.UpdateDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           fastly.ToPointer(i.Name),
		CreatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly:      fastly.ToPointer(bool(fastly.ToValue(i.WriteOnly))),
		DictionaryID:   fastly.ToPointer("456"),
		UpdatedAt:      testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func updateDictionaryError(_ *fastly.UpdateDictionaryInput) (*fastly.Dictionary, error) {
	return nil, errTest
}

var (
	errTest = errors.New("an expected error occurred")
	errFail = errors.New("this error should not be returned and indicates a failure in the code")
)

var (
	createDictionaryOutput          = "SUCCESS: Created dictionary denylist (id 456, service 123, version 4)\n"
	createDictionaryOutputWriteOnly = "SUCCESS: Created dictionary denylist as write-only (id 456, service 123, version 4)\n"
	deleteDictionaryOutput          = "SUCCESS: Deleted dictionary allowlist (service 123 version 4)\n"
	updateDictionaryOutput          = "SUCCESS: Updated dictionary oldname (service 123 version 4)\n"
	updateDictionaryNameOutput      = "SUCCESS: Updated dictionary dict-1 (service 123 version 4)\n"
)

var updateDictionaryOutputVerbose = strings.Join(
	[]string{
		"Fastly API endpoint: https://api.fastly.com",
		"Fastly API token provided via config file (profile: user)",
		"",
		"Service ID (via --service-id): 123",
		"",
		"INFO: Service version 1 is not editable, so it was automatically cloned because --autoclone is enabled. Now operating on",
		"version 4.",
		"",
		strings.TrimSpace(updateDictionaryNameOutput),
		"",
		updateDictionaryOutputVersionCloned,
	},
	"\n")

var updateDictionaryOutputVersionCloned = strings.TrimSpace(`
Version: 4
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`) + "\n"

var describeDictionaryOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`) + "\n"

var describeDictionaryOutputDeleted = strings.TrimSpace(`
Service ID: 123
Version: 1
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
Deleted (UTC): 2001-02-03 04:05
`) + "\n"

var describeDictionaryOutputVerbose = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
Digest: digest_hash
Item Count: 2
Item 1/2:
	Item Key: foo
	Item Value: bar
Item 2/2:
	Item Key: baz
	Item Value: bear
`) + "\n"

var listDictionariesOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Version: 1
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
ID: 456
Name: dict-2
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`) + "\n"

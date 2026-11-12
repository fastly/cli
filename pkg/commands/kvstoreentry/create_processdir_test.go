package kvstoreentry

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
)

func TestProcessDirClosesUploadedFiles(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "foo.txt")
	if err := os.WriteFile(filePath, []byte("FOO\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var capturedBody fastly.LengthReader
	cmd := &CreateCommand{
		Base: argparser.Base{
			Globals: &global.Data{
				APIClient: &mock.API{
					InsertKVStoreKeyFn: func(_ context.Context, in *fastly.InsertKVStoreKeyInput) error {
						if in.Key == "foo.txt" && capturedBody == nil {
							if lr, ok := in.Body.(fastly.LengthReader); ok {
								capturedBody = lr
							}
						}
						return nil
					},
				},
				Flags: global.Flags{AutoYes: true, NonInteractive: true},
			},
		},
		dirPath:        dir,
		dirConcurrency: 50,
	}

	if err := cmd.ProcessDir(nil, io.Discard); err != nil {
		t.Fatalf("ProcessDir: %v", err)
	}
	if capturedBody == nil {
		t.Fatal("expected InsertKVStoreKey LengthReader body to be captured")
	}
	if _, err := capturedBody.Read(make([]byte, 1)); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("expected upload file to be closed after ProcessDir, got read err: %v", err)
	}
}

func TestProcessDirRetriesAfterConnectionLossWithRewoundBody(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "foo.txt")
	if err := os.WriteFile(filePath, []byte("FOO\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	calls := 0
	cmd := &CreateCommand{
		Base: argparser.Base{
			Globals: &global.Data{
				APIClient: &mock.API{
					InsertKVStoreKeyFn: func(_ context.Context, in *fastly.InsertKVStoreKeyInput) error {
						calls++
						if calls == 1 {
							_, _ = io.ReadAll(in.Body)
							return errors.New("net/http: cannot rewind body after connection loss")
						}
						data, err := io.ReadAll(in.Body)
						if err != nil {
							return err
						}
						if string(data) != "FOO\n" {
							return errors.New("retry body mismatch")
						}
						return nil
					},
				},
				Flags: global.Flags{AutoYes: true, NonInteractive: true},
			},
		},
		dirPath:        dir,
		dirConcurrency: 50,
	}

	if err := cmd.ProcessDir(nil, io.Discard); err != nil {
		t.Fatalf("ProcessDir: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 insert attempts, got %d", calls)
	}
}

func TestProcessDirRetryStillFailsReportsConnectionLoss(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "foo.txt")
	if err := os.WriteFile(filePath, []byte("FOO\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	rewindErr := errors.New("net/http: cannot rewind body after connection loss")
	cmd := &CreateCommand{
		Base: argparser.Base{
			Globals: &global.Data{
				APIClient: &mock.API{
					InsertKVStoreKeyFn: func(_ context.Context, _ *fastly.InsertKVStoreKeyInput) error {
						return rewindErr
					},
				},
				Flags: global.Flags{AutoYes: true, NonInteractive: true},
			},
		},
		dirPath:        dir,
		dirConcurrency: 50,
	}

	err := cmd.ProcessDir(nil, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to process all the provided files") {
		t.Fatalf("unexpected error: %v", err)
	}
}

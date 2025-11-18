package ngwaflist

import (
	"context"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/lists"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

type ListCreateInput struct {
	CommandScope scope.Type
	Description  argparser.OptionalString
	Entries      string
	Name         string
	Type         string
	WorkspaceID  *argparser.OptionalWorkspaceID
	FC           *fastly.Client
	Out          io.Writer
}

func ListCreate(argsInput ListCreateInput) (*lists.List, error) {
	input := lists.CreateInput{
		Entries: fastly.ToPointer(strings.Split(argparser.Content(argsInput.Entries), ",")),
		Name:    &argsInput.Name,
		Type:    &argsInput.Type,
	}
	if argsInput.Description.WasSet {
		input.Description = &argsInput.Description.Value
	}
	inputWorkspaceID := ""
	if argsInput.CommandScope == scope.ScopeTypeWorkspace {
		if err := argsInput.WorkspaceID.Parse(); err != nil {
			return nil, err
		}
		inputWorkspaceID = argsInput.WorkspaceID.Value
	}

	var err error
	input.Scope, err = generateScope(argsInput.CommandScope, inputWorkspaceID)
	if err != nil {
		return nil, err
	}

	return lists.Create(context.TODO(), argsInput.FC, &input)
}

type ListDeleteInput struct {
	CommandScope scope.Type
	ListID       string
	WorkspaceID  *argparser.OptionalWorkspaceID
	FC           *fastly.Client
	Out          io.Writer
}

func ListDelete(argsInput ListDeleteInput) error {
	input := lists.DeleteInput{
		ListID: &argsInput.ListID,
	}
	inputWorkspaceID := ""
	if argsInput.CommandScope == scope.ScopeTypeWorkspace {
		if err := argsInput.WorkspaceID.Parse(); err != nil {
			return err
		}
		inputWorkspaceID = argsInput.WorkspaceID.Value
	}
	var err error
	input.Scope, err = generateScope(argsInput.CommandScope, inputWorkspaceID)
	if err != nil {
		return err
	}

	return lists.Delete(context.TODO(), argsInput.FC, &input)
}

type ListGetInput struct {
	CommandScope scope.Type
	ListID       string
	WorkspaceID  *argparser.OptionalWorkspaceID
	FC           *fastly.Client
	Out          io.Writer
}

func ListGet(argsInput ListGetInput) (*lists.List, error) {
	input := lists.GetInput{
		ListID: &argsInput.ListID,
	}
	inputWorkspaceID := ""
	if argsInput.CommandScope == scope.ScopeTypeWorkspace {
		if err := argsInput.WorkspaceID.Parse(); err != nil {
			return nil, err
		}
		inputWorkspaceID = argsInput.WorkspaceID.Value
	}
	var err error
	input.Scope, err = generateScope(argsInput.CommandScope, inputWorkspaceID)
	if err != nil {
		return nil, err
	}

	return lists.Get(context.TODO(), argsInput.FC, &input)
}

type ListListInput struct {
	CommandScope scope.Type
	ListID       string
	Type         string
	WorkspaceID  *argparser.OptionalWorkspaceID
	FC           *fastly.Client
	Out          io.Writer
}

func ListList(argsInput ListListInput) (*lists.Lists, error) {
	input := lists.ListInput{}
	inputWorkspaceID := ""
	if argsInput.CommandScope == scope.ScopeTypeWorkspace {
		if err := argsInput.WorkspaceID.Parse(); err != nil {
			return nil, err
		}
		inputWorkspaceID = argsInput.WorkspaceID.Value
	}
	var err error
	input.Scope, err = generateScope(argsInput.CommandScope, inputWorkspaceID)
	if err != nil {
		return nil, err
	}

	data, err := lists.ListLists(context.TODO(), argsInput.FC, &input)
	if err != nil {
		return nil, err
	}

	listFilteredByType := []lists.List{}

	for _, list := range data.Data {
		if list.Type == argsInput.Type {
			listFilteredByType = append(listFilteredByType, list)
		}
	}
	data.Data = listFilteredByType

	return data, nil
}

type ListUpdateInput struct {
	CommandScope scope.Type
	Description  argparser.OptionalString
	Entries      argparser.OptionalString
	ListID       string
	WorkspaceID  *argparser.OptionalWorkspaceID
	FC           *fastly.Client
	Out          io.Writer
}

func ListUpdate(argsInput ListUpdateInput) (*lists.List, error) {
	input := lists.UpdateInput{
		ListID: &argsInput.ListID,
	}
	if argsInput.Description.WasSet {
		input.Description = &argsInput.Description.Value
	}
	if argsInput.Entries.WasSet {
		input.Entries = fastly.ToPointer(strings.Split(argparser.Content(argsInput.Entries.Value), ","))
	}
	inputWorkspaceID := ""
	if argsInput.CommandScope == scope.ScopeTypeWorkspace {
		if err := argsInput.WorkspaceID.Parse(); err != nil {
			return nil, err
		}
		inputWorkspaceID = argsInput.WorkspaceID.Value
	}
	var err error
	input.Scope, err = generateScope(argsInput.CommandScope, inputWorkspaceID)
	if err != nil {
		return nil, err
	}

	return lists.Update(context.TODO(), argsInput.FC, &input)
}

func generateScope(inputScope scope.Type, workspaceID string) (*scope.Scope, error) {
	if inputScope == scope.ScopeTypeAccount {
		return &scope.Scope{
			Type:      scope.ScopeTypeAccount,
			AppliesTo: ngwaf.DefaultAccountScope,
		}, nil
	}
	if inputScope == scope.ScopeTypeWorkspace {
		return &scope.Scope{
			Type:      scope.ScopeTypeWorkspace,
			AppliesTo: []string{workspaceID},
		}, nil
	}
	return nil, fsterr.ErrInvalidNGWAFScopeType
}

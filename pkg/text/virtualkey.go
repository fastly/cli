package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"
)

// PrintVirtualKeyWithToken displays a virtual key including its access token.
// This is only returned by the create and rotate operations.
func PrintVirtualKeyWithToken(out io.Writer, k *key.VirtualKeyWithToken) {
	fmt.Fprintf(out, "ID: %s\n", k.ID)
	fmt.Fprintf(out, "Name: %s\n", k.Name)
	fmt.Fprintf(out, "Model: %s\n", k.Model)
	fmt.Fprintf(out, "Provider: %s\n", k.Provider)
	fmt.Fprintf(out, "User ID: %s\n", k.UserID)
	if k.UserName != "" {
		fmt.Fprintf(out, "User Name: %s\n", k.UserName)
	}
	fmt.Fprintf(out, "Customer ID: %s\n", k.CustomerID)
	if k.ExpiresAt != nil {
		fmt.Fprintf(out, "Expires At: %s\n", *k.ExpiresAt)
	}
	fmt.Fprintf(out, "Created At: %s\n", k.CreatedAt)
	fmt.Fprintf(out, "Updated At: %s\n", k.UpdatedAt)
	fmt.Fprintf(out, "Access Token: %s\n", k.AccessToken)
}

// PrintVirtualKey displays a virtual key.
func PrintVirtualKey(out io.Writer, k *key.VirtualKey) {
	fmt.Fprintf(out, "ID: %s\n", k.ID)
	fmt.Fprintf(out, "Name: %s\n", k.Name)
	fmt.Fprintf(out, "Model: %s\n", k.Model)
	fmt.Fprintf(out, "Provider: %s\n", k.Provider)
	fmt.Fprintf(out, "User ID: %s\n", k.UserID)
	fmt.Fprintf(out, "Customer ID: %s\n", k.CustomerID)
	if k.ExpiresAt != nil {
		fmt.Fprintf(out, "Expires At: %s\n", *k.ExpiresAt)
	}
	fmt.Fprintf(out, "Created At: %s\n", k.CreatedAt)
	fmt.Fprintf(out, "Updated At: %s\n", k.UpdatedAt)
}

// PrintVirtualKeyListItem displays a virtual key list item.
func PrintVirtualKeyListItem(out io.Writer, k *key.VirtualKeyListItem) {
	fmt.Fprintf(out, "ID: %s\n", k.ID)
	fmt.Fprintf(out, "Name: %s\n", k.Name)
	fmt.Fprintf(out, "Model: %s\n", k.Model)
	fmt.Fprintf(out, "Provider: %s\n", k.Provider)
	if k.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", k.UserID)
	}
	if k.CreatedBy != "" {
		fmt.Fprintf(out, "Created By: %s\n", k.CreatedBy)
	}
	if k.ExpiresAt != nil {
		fmt.Fprintf(out, "Expires At: %s\n", *k.ExpiresAt)
	}
	if k.LastUsedAt != nil {
		fmt.Fprintf(out, "Last Used At: %s\n", *k.LastUsedAt)
	}
	if k.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted At: %s\n", *k.DeletedAt)
	}
	fmt.Fprintf(out, "Created At: %s\n", k.CreatedAt)
	fmt.Fprintf(out, "Updated At: %s\n", k.UpdatedAt)
}

// PrintVirtualKeysTbl displays virtual keys in a table format.
func PrintVirtualKeysTbl(out io.Writer, keysToPrint []key.VirtualKeyListItem) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Model", "Provider", "Created By", "Created At")

	if keysToPrint == nil {
		tbl.Print()
		return
	}

	for _, k := range keysToPrint {
		tbl.AddLine(k.ID, k.Name, k.Model, k.Provider, k.CreatedBy, k.CreatedAt)
	}
	tbl.Print()
}

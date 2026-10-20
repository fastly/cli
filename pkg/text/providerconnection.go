package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"
)

// PrintProviderConnection displays a provider connection.
func PrintProviderConnection(out io.Writer, pc *providerconnection.ProviderConnection) {
	fmt.Fprintf(out, "ID: %s\n", pc.ID)
	fmt.Fprintf(out, "Name: %s\n", pc.Name)
	fmt.Fprintf(out, "Models: %s\n", strings.Join(pc.Models, ", "))
	fmt.Fprintf(out, "Auth Type: %s\n", pc.AuthType)
	if pc.BaseURL != "" {
		fmt.Fprintf(out, "Base URL: %s\n", pc.BaseURL)
	}
	if pc.Region != "" {
		fmt.Fprintf(out, "Region: %s\n", pc.Region)
	}
	if len(pc.VirtualKeyIDs) > 0 {
		fmt.Fprintf(out, "Virtual Key IDs: %s\n", strings.Join(pc.VirtualKeyIDs, ", "))
	}
	fmt.Fprintf(out, "Created At: %s\n", pc.CreatedAt)
	fmt.Fprintf(out, "Updated At: %s\n", pc.UpdatedAt)
	if pc.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted At: %s\n", *pc.DeletedAt)
	}
}

// PrintProviderConnectionsTbl displays provider connections in a table format.
func PrintProviderConnectionsTbl(out io.Writer, connectionsToPrint []providerconnection.ProviderConnection) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Models", "Auth Type", "Base URL", "Region", "Created At")

	if connectionsToPrint == nil {
		tbl.Print()
		return
	}

	for _, pc := range connectionsToPrint {
		tbl.AddLine(pc.ID, pc.Name, strings.Join(pc.Models, ", "), pc.AuthType, pc.BaseURL, pc.Region, pc.CreatedAt)
	}
	tbl.Print()
}

// PrintBedrockSetup displays the setup information for a Bedrock provider
// connection.
func PrintBedrockSetup(out io.Writer, s *providerconnection.BedrockSetup) {
	fmt.Fprintf(out, "ARC Role ARN: %s\n", s.ArcRoleARN)
	fmt.Fprintf(out, "External ID: %s\n", s.ExternalID)
}

// PrintVertexSetup displays the setup information for a Vertex AI provider
// connection.
func PrintVertexSetup(out io.Writer, s *providerconnection.VertexSetup) {
	fmt.Fprintf(out, "OIDC Issuer URL: %s\n", s.OIDCIssuerURL)
	fmt.Fprintf(out, "Audience: %s\n", s.Audience)
	fmt.Fprintf(out, "Customer ID: %s\n", s.CustomerID)
}

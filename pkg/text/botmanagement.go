package text

import (
	"fmt"
	"io"
	"strings"

	bmworkspaces "github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces"
	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces/policy"
)

// PrintBotManagementWorkspace displays a Bot Management workspace.
func PrintBotManagementWorkspace(out io.Writer, w *bmworkspaces.Workspace) {
	fmt.Fprintf(out, "ID: %s\n", w.ID)
	fmt.Fprintf(out, "Name: %s\n", w.Name)
	fmt.Fprintf(out, "Description: %s\n", w.Description)
	fmt.Fprintf(out, "Protection Mode: %s\n", w.ProtectionMode)
	fmt.Fprintf(out, "Services: %s\n", strings.Join(w.Services, ", "))
	fmt.Fprintf(out, "Updated At: %s\n", w.UpdatedAt)
}

// PrintBotManagementWorkspaceTbl displays Bot Management workspaces in a table
// format.
func PrintBotManagementWorkspaceTbl(out io.Writer, ws []bmworkspaces.Workspace) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Protection Mode", "Services", "Updated At")
	for _, w := range ws {
		tbl.AddLine(w.ID, w.Name, w.ProtectionMode, strings.Join(w.Services, ", "), w.UpdatedAt)
	}
	tbl.Print()
}

// PrintBotManagementCategory displays a Bot Management category.
func PrintBotManagementCategory(out io.Writer, c *policy.Category) {
	fmt.Fprintf(out, "ID: %s\n", c.CategoryID)
	fmt.Fprintf(out, "Name: %s\n", c.Name)
	fmt.Fprintf(out, "Description: %s\n", c.Description)
	fmt.Fprintf(out, "Action: %s\n", c.Action.Type)
}

// PrintBotManagementCategoryTbl displays Bot Management categories in a table
// format.
func PrintBotManagementCategoryTbl(out io.Writer, cs []policy.Category) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Action")
	for _, c := range cs {
		tbl.AddLine(c.CategoryID, c.Name, c.Action.Type)
	}
	tbl.Print()
}

// PrintBotManagementBot displays a Bot Management bot.
func PrintBotManagementBot(out io.Writer, b *policy.Bot) {
	fmt.Fprintf(out, "ID: %s\n", b.BotID)
	fmt.Fprintf(out, "Name: %s\n", b.Name)
	if b.CategoryID != "" {
		fmt.Fprintf(out, "Category ID: %s\n", b.CategoryID)
	}
	fmt.Fprintf(out, "Action: %s\n", b.Action.Type)
}

// PrintBotManagementBotTbl displays Bot Management bots in a table format.
func PrintBotManagementBotTbl(out io.Writer, bs []policy.Bot) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Category ID", "Action")
	for _, b := range bs {
		tbl.AddLine(b.BotID, b.Name, b.CategoryID, b.Action.Type)
	}
	tbl.Print()
}

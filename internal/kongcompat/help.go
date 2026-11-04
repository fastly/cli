package kongcompat

// HelpCommand is kingpin's implicit help command, "help [<command>...]".
// Add it as the first command of the grammar to keep it:
//
//	Help kongcompat.HelpCommand `cmd:"" help:"Show help."`
type HelpCommand struct {
	Command []string `arg:"" optional:"" sep:"none" help:"Show help on command."`
}

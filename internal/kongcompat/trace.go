package kongcompat

import "github.com/alecthomas/kong"

// Trace is kong.Trace without its side effect on the parser.
// kong.Trace marks the values it decoded as set, and a later Parse trusts
// those marks: it skips the existingfile and existingdir checks and treats
// required flags as given.
// Use Trace instead of kong.Trace when the same parser parses again
// afterwards, as kingpin programs did with ParseContext followed by Parse.
func Trace(k *kong.Kong, args []string) (*kong.Context, error) {
	set := map[*kong.Value]bool{}
	values(k.Model.Node, func(v *kong.Value) { set[v] = v.Set })
	defer values(k.Model.Node, func(v *kong.Value) { v.Set = set[v] })
	return kong.Trace(k, args)
}

// values calls fn on every flag and argument of the grammar under n.
func values(n *kong.Node, fn func(*kong.Value)) {
	nodes(n, func(n *kong.Node) {
		for _, v := range ownValues(n) {
			fn(v)
		}
	})
}

// ownValues returns n's own flags and arguments, not its subcommands'.
func ownValues(n *kong.Node) []*kong.Value {
	var out []*kong.Value
	for _, f := range n.Flags {
		out = append(out, f.Value)
	}
	return append(out, n.Positional...)
}

// nodes calls fn on n and every node under it, parents first.
func nodes(n *kong.Node, fn func(*kong.Node)) {
	fn(n)
	for _, c := range n.Children {
		nodes(c, fn)
	}
}

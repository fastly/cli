package kongcompat

import (
	"strings"

	"github.com/alecthomas/kong"
)

// Vars holds defaults and help texts that are only known at run time.
// A flag, argument or command refers to one by name in its tag, as in
// default:"${limit_default}".
//
// Unlike with kong.Vars, values are used as is, so a $ stays a $.
// Only a default or help that is exactly ${name} gets replaced.
type Vars map[string]string

// Apply implements kong.Option.
// kong would interpolate its own variables again in every value, so these
// are written straight into the grammar instead, with $ escaped for the one
// interpolation defaults and help still go through.
func (v Vars) Apply(k *kong.Kong) error {
	return kong.PostBuild(func(k *kong.Kong) error {
		v.fill(k.Model.Node)
		return nil
	}).Apply(k)
}

func (v Vars) fill(n *kong.Node) {
	nodes(n, func(n *kong.Node) {
		v.replace(&n.Help)
		for _, val := range ownValues(n) {
			v.replace(&val.Default)
			v.replace(&val.Help)
		}
	})
}

// replace substitutes the variable when *ref is exactly ${name}.
// Its $ signs are doubled, since kong still interpolates the result once.
func (v Vars) replace(ref *string) {
	name, ok := strings.CutPrefix(*ref, "${")
	if !ok || !strings.HasSuffix(name, "}") {
		return
	}
	if s, ok := v[strings.TrimSuffix(name, "}")]; ok {
		*ref = strings.ReplaceAll(s, "$", "$$")
	}
}

// CommandVars applies the Vars returned by every command struct that has a
// Vars method.
func CommandVars() kong.Option {
	return kong.PostBuild(func(k *kong.Kong) error {
		vars := Vars{}
		collectVars(k.Model.Node, vars)
		vars.fill(k.Model.Node)
		return nil
	})
}

func collectVars(root *kong.Node, vars Vars) {
	nodes(root, func(n *kong.Node) {
		if n.Target.IsValid() && n.Target.CanAddr() {
			if v, ok := n.Target.Addr().Interface().(interface{ Vars() Vars }); ok {
				for name, value := range v.Vars() {
					vars[name] = value
				}
			}
		}
	})
}

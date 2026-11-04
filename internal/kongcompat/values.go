package kongcompat

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
)

// checkValue parses value the way kingpin did, to report a bad one in
// kingpin's words before kong decodes it.
// Other types are left to kong, whose errors valueError cleans up.
func checkValue(v *kong.Value, value string) error {
	if !v.Target.IsValid() || v.IsCounter() {
		return nil
	}
	if isPath(v) {
		return checkPath(value, v.Tag.Type == "existingdir")
	}
	t := v.Target.Type()
	if t.Kind() == reflect.Slice && t.PkgPath() == "" {
		t = t.Elem()
		if t.Kind() == reflect.Bool {
			_, err := strconv.ParseBool(value)
			return err
		}
	}
	switch {
	case t == reflect.TypeOf(net.IP{}):
		if net.ParseIP(value) == nil {
			return fmt.Errorf("'%s' is not an IP address", value)
		}
		return nil
	case t == reflect.TypeOf(&url.URL{}) && version == 2:
		if _, err := url.Parse(value); err != nil {
			return fmt.Errorf("invalid URL: %s", err)
		}
		return nil
	case t == reflect.TypeOf(time.Duration(0)) && !v2From23():
		_, err := time.ParseDuration(value)
		return err
	case t == reflect.TypeOf(map[string]string{}):
		if !strings.ContainsAny(value, ":=") {
			return fmt.Errorf("expected KEY=VALUE got '%s'", value)
		}
		return nil
	case t.PkgPath() != "":
		// Other named types have mappers of their own.
		return nil
	}
	var err error
	switch t.Kind() {
	case reflect.Int:
		// kingpin parsed an int as a float and truncated it.
		_, err = strconv.ParseFloat(value, 64)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err = strconv.ParseInt(value, 0, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, err = strconv.ParseUint(value, 0, t.Bits())
	case reflect.Float32, reflect.Float64:
		_, err = strconv.ParseFloat(value, t.Bits())
	}
	return err
}

// Values decodes values the way kingpin did where kong differs:
//
//   - an int is parsed as a float and truncated, so 1.5 means 1;
//   - a []bool takes a value each time, true or false;
//   - a map[string]string entry is KEY=VALUE or KEY:VALUE, split at the
//     first separator;
//   - an existingfile or existingdir value is kept as given, not made
//     absolute;
//   - a noflag value is a boolean flag named no-x, which kingpin v3 set to
//     false when it was given;
//   - a boolarg value is a positional boolean, given as true or false;
//   - a counter's environment variable or default counts as one, whatever
//     it says;
//   - a file is opened by the name given, and "-" is just a file name;
//   - a regular expression is compiled, in lists too;
//   - a map[string]string flag left out holds an empty map, not nil;
//   - a list argument's default or environment variable holds several
//     values, split at the separator, as for a flag.
func Values() kong.Option {
	return kong.OptionFunc(func(k *kong.Kong) error {
		for _, opt := range []kong.Option{
			kong.TypeMapper(reflect.TypeOf(0), kong.MapperFunc(floatInt)),
			kong.TypeMapper(reflect.TypeOf([]bool{}), kong.MapperFunc(boolList)),
			kong.TypeMapper(reflect.TypeOf(map[string]string{}), kong.MapperFunc(stringMap)),
			kong.NamedMapper("existingfile", existingPath(false)),
			kong.NamedMapper("existingdir", existingPath(true)),
			kong.NamedMapper("noflag", noFlag{}),
			kong.NamedMapper("boolarg", kong.MapperFunc(boolArg)),
			kong.NamedMapper("counter", kong.MapperFunc(counter)),
			kong.TypeMapper(reflect.TypeOf((*os.File)(nil)), kong.MapperFunc(openFile)),
			kong.TypeMapper(reflect.TypeOf((*regexp.Regexp)(nil)), kong.MapperFunc(compileRegexp)),
			kong.WithBeforeApply(func(ctx *kong.Context) error {
				fillMaps(ctx.Kong)
				return nil
			}),
			kong.PostBuild(func(k *kong.Kong) error {
				values(k.Model.Node, func(v *kong.Value) {
					if v.Flag == nil && v.Target.Kind() == reflect.Slice && !reflect.PointerTo(v.Target.Type()).Implements(mapperValue) {
						v.Mapper = argList{v.Mapper}
					}
				})
				return nil
			}),
		} {
			if err := opt.Apply(k); err != nil {
				return err
			}
		}
		return nil
	})
}

var mapperValue = reflect.TypeOf((*kong.MapperValue)(nil)).Elem()

// argList splits the default or environment variable of a list argument,
// which kong only splits for flags.
// Values from the command line come as tokens of their own.
type argList struct{ kong.Mapper }

func (m argList) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	t := ctx.Scan.Peek()
	s, ok := t.Value.(string)
	if t.Type != kong.FlagValueToken || !ok || ctx.Value.Tag.Sep == -1 {
		return m.Mapper.Decode(ctx, target)
	}
	ctx.Scan.Pop()
	var tokens []kong.Token
	for _, part := range kong.SplitEscaped(s, ctx.Value.Tag.Sep) {
		tokens = append(tokens, kong.Token{Type: kong.FlagValueToken, Value: part})
	}
	return m.Mapper.Decode(ctx.WithScanner(kong.ScanFromTokens(tokens...)), target)
}

func floatInt(ctx *kong.DecodeContext, target reflect.Value) error {
	var s string
	if err := ctx.Scan.PopValueInto("int", &s); err != nil {
		return err
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	target.SetInt(int64(f))
	return nil
}

// valueTexts pops the raw text of a value: a flag's value split at sep, or
// for a positional one, the next argument, or all the remaining ones when
// many is set.
func valueTexts(ctx *kong.DecodeContext, sep rune, many bool) ([]string, error) {
	if ctx.Value.Flag == nil && !many {
		t := ctx.Scan.Pop()
		if !t.IsValue() {
			return nil, fmt.Errorf("missing value")
		}
		return []string{fmt.Sprint(t.Value)}, nil
	}
	if ctx.Value.Flag == nil {
		var out []string
		for _, t := range ctx.Scan.PopWhile(func(t kong.Token) bool { return t.IsValue() }) {
			out = append(out, fmt.Sprint(t.Value))
		}
		return out, nil
	}
	t := ctx.Scan.Pop()
	if t.IsEOL() {
		return nil, fmt.Errorf("missing value")
	}
	return kong.SplitEscaped(fmt.Sprint(t.Value), sep), nil
}

// parseDuration parses a duration as the program's kingpin did.
// durations_v2.go swaps in str2duration for kingpin v2.3.0 and later.
var parseDuration = time.ParseDuration

// ParseDuration parses a duration as the program's kingpin did, for the
// generated code that decodes values itself.
func ParseDuration(s string) (time.Duration, error) {
	return parseDuration(s)
}

// counter adds one for every value, as kingpin's counter did.
func counter(ctx *kong.DecodeContext, target reflect.Value) error {
	if ctx.Scan.Peek().Type == kong.FlagValueToken {
		ctx.Scan.Pop()
	}
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetInt(target.Int() + 1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		target.SetUint(target.Uint() + 1)
	default:
		return fmt.Errorf("a counter needs an integer, not %s", target.Type())
	}
	return nil
}

// compileRegexp compiles a regular expression, which kong's own decoding
// can't do into a list.
func compileRegexp(ctx *kong.DecodeContext, target reflect.Value) error {
	texts, err := valueTexts(ctx, -1, false)
	if err != nil {
		return err
	}
	re, err := regexp.Compile(texts[0])
	if err != nil {
		return err
	}
	target.Set(reflect.ValueOf(re))
	return nil
}

// openFile opens a file the way kingpin's File did.
func openFile(ctx *kong.DecodeContext, target reflect.Value) error {
	texts, err := valueTexts(ctx, -1, false)
	if err != nil {
		return err
	}
	f, err := os.Open(texts[0])
	if err != nil {
		return err
	}
	target.Set(reflect.ValueOf(f))
	return nil
}

// boolArg decodes a positional boolean, which takes a value, unlike a
// boolean flag.
func boolArg(ctx *kong.DecodeContext, target reflect.Value) error {
	texts, err := valueTexts(ctx, -1, false)
	if err != nil {
		return err
	}
	b, err := strconv.ParseBool(texts[0])
	if err != nil {
		return err
	}
	target.SetBool(b)
	return nil
}

func boolList(ctx *kong.DecodeContext, target reflect.Value) error {
	texts, err := valueTexts(ctx, ctx.Value.Tag.Sep, true)
	if err != nil {
		return err
	}
	for _, s := range texts {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		target.Set(reflect.Append(target, reflect.ValueOf(b)))
	}
	return nil
}

func stringMap(ctx *kong.DecodeContext, target reflect.Value) error {
	texts, err := valueTexts(ctx, ctx.Value.Tag.MapSep, true)
	if err != nil {
		return err
	}
	if target.IsNil() {
		target.Set(reflect.MakeMap(target.Type()))
	}
	for _, s := range texts {
		i := strings.IndexAny(s, ":=")
		if i < 0 {
			return fmt.Errorf("expected KEY=VALUE got '%s'", s)
		}
		target.SetMapIndex(reflect.ValueOf(s[:i]), reflect.ValueOf(s[i+1:]))
	}
	return nil
}

func existingPath(dir bool) kong.MapperFunc {
	return func(ctx *kong.DecodeContext, target reflect.Value) error {
		many := target.Kind() == reflect.Slice
		sep := ctx.Value.Tag.Sep
		if !many {
			sep = -1
		}
		texts, err := valueTexts(ctx, sep, many)
		if err != nil {
			return err
		}
		for _, s := range texts {
			// kong also decodes the defaults of commands that were not
			// selected, which kingpin never checked.
			// Parse checks the defaults that are used itself.
			if ctx.Value.Active && (many || !ctx.Value.HasDefault || s != ctx.Value.Default) {
				if err := checkPath(s, dir); err != nil {
					return err
				}
			}
			if target.Kind() == reflect.Slice {
				target.Set(reflect.Append(target, reflect.ValueOf(s)))
			} else {
				target.SetString(s)
			}
		}
		return nil
	}
}

// checkPath checks an existing file or directory as kingpin did.
func checkPath(path string, dir bool) error {
	st, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("path '%s' does not exist", path)
	case err != nil:
		return err
	case !dir && st.IsDir():
		return fmt.Errorf("'%s' is a directory", st.Name())
	case dir && !st.IsDir():
		return fmt.Errorf("'%s' is a file", st.Name())
	}
	return nil
}

// noFlag decodes a kingpin v3 boolean flag named no-x, which is false when
// given without a value.
type noFlag struct{}

func (noFlag) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	if ctx.Scan.Peek().Type != kong.FlagValueToken {
		target.SetBool(false)
		return nil
	}
	switch v := ctx.Scan.Pop().Value.(type) {
	case bool:
		target.SetBool(v)
	default:
		b, err := strconv.ParseBool(fmt.Sprint(v))
		if err != nil {
			return fmt.Errorf("expected a boolean but got %q", v)
		}
		target.SetBool(b)
	}
	return nil
}

func (noFlag) IsBool() bool { return true }

// valueError strips the flag or argument name kong puts in front of a
// value's own parse error, since kingpin reported that error as is.
func valueError(k *kong.Kong, err error) error {
	names := func(prefix string) bool {
		found := false
		values(k.Model.Node, func(v *kong.Value) { found = found || v.ShortSummary() == prefix })
		return found
	}
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			return err
		}
		// Only drop the prefix that names a flag or argument.
		// Errors from strconv, net/url or the program wrap their own cause
		// and quote what failed, and kingpin showed them whole.
		msg := err.Error()
		prefix, prefixed := strings.CutSuffix(msg, ": "+inner.Error())
		if msg != inner.Error() && !(prefixed && names(prefix)) {
			return err
		}
		err = inner
	}
}

// newOptions are extra options New gives every parser.
// Files copied only for some kingpin versions, such as durations_v2.go, add
// to it.
var newOptions []kong.Option

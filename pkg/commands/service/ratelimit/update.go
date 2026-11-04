package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply checks the values kong cannot check by itself.
func (c *UpdateCommand) AfterApply(ctx *kong.Context) error {
	if (c.Action != "" || kongcompat.Given(ctx, "action")) && !slices.Contains(rateLimitActionFlagOpts, c.Action) {
		return fmt.Errorf("enum value must be one of %s, got '%s'", strings.Join(rateLimitActionFlagOpts, ","), c.Action)
	}
	if (c.LoggerType != "" || kongcompat.Given(ctx, "logger-type")) && !slices.Contains(rateLimitLoggerFlagOpts, c.LoggerType) {
		return fmt.Errorf("enum value must be one of %s, got '%s'", strings.Join(rateLimitLoggerFlagOpts, ","), c.LoggerType)
	}
	if (c.WindowSize != "" || kongcompat.Given(ctx, "window-size")) && !slices.Contains(rateLimitWindowSizeFlagOpts, c.WindowSize) {
		return fmt.Errorf("enum value must be one of %s, got '%s'", strings.Join(rateLimitWindowSizeFlagOpts, ","), c.WindowSize)
	}
	return nil
}

// UpdateCommand calls the Fastly API to create an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ID string `required:"" help:"Alphanumeric string identifying the rate limiter"`

	// Optional.
	Action       string `help:"The action to take when a rate limiter violation is detected"`
	ClientKeys   string `name:"client-key" help:"Comma-separated list of VCL variable used to generate a counter key to identify a client"`
	FeatRevision int    `name:"feature-revision" help:"Revision number of the rate limiting feature implementation"`
	HTTPMethods  string `help:"Comma-separated list of HTTP methods to apply rate limiting to"`
	argparser.JSONOutput
	LoggerType          string "help:\"Name of the type of logging endpoint to be used when action is `log_only`\""
	NameFlag            string `name:"name" help:"A human readable name for the rate limiting rule"`
	PenaltyDuration     int    `name:"penalty-box-dur" help:"Length of time in minutes that the rate limiter is in effect after the initial violation is detected"`
	ResponseContent     string `help:"HTTP response body data"`
	ResponseContentType string `help:"HTTP Content-Type (e.g. application/json)"`
	ResponseObjectName  string `help:"Name of existing response object. Required if action is response_object"`
	ResponseStatus      int    `help:"HTTP response status code (e.g. 429)"`
	RpsLimit            int    `help:"Upper limit of requests per second allowed by the rate limiter"`
	URIDictName         string `help:"The name of an Edge Dictionary containing URIs as keys"`
	WindowSize          string `help:"Number of seconds during which the RPS limit must be exceeded in order to trigger a violation"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if err := c.responseFlagValidator(); err != nil {
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: "When updating a response, all response flags (--response-content, --response-content-type, --response-status) should be set",
		}
	}

	input := c.constructInput()
	o, err := c.Globals.APIClient.UpdateERL(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Updated rate limiter '%s' (%s)", fastly.ToValue(o.Name), fastly.ToValue(o.RateLimiterID))
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateERLInput {
	var input fastly.UpdateERLInput
	input.ERLID = c.ID

	// NOTE: rateLimitActions is defined in ./create.go
	if c.Action != "" {
		for _, a := range fastly.ERLActions {
			if c.Action == string(a) {
				input.Action = new(a)
				break
			}
		}
	}

	if c.ClientKeys != "" {
		clientKeys := strings.Split(strings.ReplaceAll(c.ClientKeys, " ", ""), ",")
		input.ClientKey = &clientKeys
	}

	if c.FeatRevision > 0 {
		input.FeatureRevision = new(c.FeatRevision)
	}

	if c.HTTPMethods != "" {
		httpMethods := strings.Split(strings.ReplaceAll(c.HTTPMethods, " ", ""), ",")
		input.HTTPMethods = &httpMethods
	}

	// NOTE: rateLimitLoggers is defined in ./create.go
	if c.LoggerType != "" {
		for _, l := range fastly.ERLLoggers {
			if c.LoggerType == string(l) {
				input.LoggerType = new(l)
				break
			}
		}
	}

	if c.NameFlag != "" {
		input.Name = new(c.NameFlag)
	}

	if c.PenaltyDuration > 0 {
		input.PenaltyBoxDuration = new(c.PenaltyDuration)
	}

	if c.ResponseContent != "" && c.ResponseContentType != "" && c.ResponseStatus > 0 {
		input.Response = &fastly.ERLResponseType{
			ERLContent:     new(c.ResponseContent),
			ERLContentType: new(c.ResponseContentType),
			ERLStatus:      new(c.ResponseStatus),
		}
	}

	if c.ResponseObjectName != "" {
		input.ResponseObjectName = new(c.ResponseObjectName)
	}

	if c.RpsLimit > 0 {
		input.RpsLimit = new(c.RpsLimit)
	}

	if c.URIDictName != "" {
		input.URIDictionaryName = new(c.URIDictName)
	}

	// NOTE: rateLimitWindowSizes is defined in ./create.go
	if c.WindowSize != "" {
		for _, w := range fastly.ERLWindowSizes {
			if c.WindowSize == fmt.Sprint(w) {
				input.WindowSize = new(w)
				break
			}
		}
	}

	return &input
}

// responseFlagValidator ensures if a user specifies one of the response flags,
// that they must specify ALL of the response flags.
func (c *UpdateCommand) responseFlagValidator() error {
	var state int
	if c.ResponseContent != "" {
		state++
	}
	if c.ResponseContentType != "" {
		state++
	}
	if c.ResponseStatus > 0 {
		state++
	}
	if state > 0 && state < 3 {
		return errors.New("invalid flag use")
	}
	return nil
}

// Hints returns the completion hints of the flags whose values are only
// known at run time, as kingpin's HintOptions gave them.
func (c *UpdateCommand) Hints() map[string][]string {
	return map[string][]string{
		"action":      rateLimitActionFlagOpts,
		"logger-type": rateLimitLoggerFlagOpts,
		"window-size": rateLimitWindowSizeFlagOpts,
	}
}

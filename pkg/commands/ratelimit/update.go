package ratelimit

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("update", "Update a rate limiter by its ID")

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying the rate limiter").Required().StringVar(&c.id)

	// Optional.
	c.CmdClause.Flag("action", "The action to take when a rate limiter violation is detected").HintOptions(rateLimitActionFlagOpts...).EnumVar(&c.action, rateLimitActionFlagOpts...)
	c.CmdClause.Flag("client-key", "Comma-separated list of VCL variable used to generate a counter key to identify a client").StringVar(&c.clientKeys)
	c.CmdClause.Flag("feature-revision", "Revision number of the rate limiting feature implementation").IntVar(&c.featRevision)
	c.CmdClause.Flag("http-methods", "Comma-separated list of HTTP methods to apply rate limiting to").StringVar(&c.httpMethods)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("logger-type", "Name of the type of logging endpoint to be used when action is `log_only`").HintOptions(rateLimitLoggerFlagOpts...).EnumVar(&c.loggerType, rateLimitLoggerFlagOpts...)
	c.CmdClause.Flag("name", "A human readable name for the rate limiting rule").StringVar(&c.name)
	c.CmdClause.Flag("penalty-box-dur", "Length of time in minutes that the rate limiter is in effect after the initial violation is detected").IntVar(&c.penaltyDuration)
	c.CmdClause.Flag("response-content", "HTTP response body data").StringVar(&c.responseContent)
	c.CmdClause.Flag("response-content-type", "HTTP Content-Type (e.g. application/json)").StringVar(&c.responseContentType)
	c.CmdClause.Flag("response-object-name", "Name of existing response object. Required if action is response_object").StringVar(&c.responseObjectName)
	c.CmdClause.Flag("response-status", "HTTP response status code (e.g. 429)").IntVar(&c.responseStatus)
	c.CmdClause.Flag("rps-limit", "Upper limit of requests per second allowed by the rate limiter").IntVar(&c.rpsLimit)
	c.CmdClause.Flag("uri-dict-name", "The name of an Edge Dictionary containing URIs as keys").StringVar(&c.uriDictName)
	c.CmdClause.Flag("window-size", "Number of seconds during which the RPS limit must be exceeded in order to trigger a violation").HintOptions(rateLimitWindowSizeFlagOpts...).EnumVar(&c.windowSize, rateLimitWindowSizeFlagOpts...)

	return &c
}

// UpdateCommand calls the Fastly API to create an appropriate resource.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	action              string
	clientKeys          string
	featRevision        int
	httpMethods         string
	id                  string
	loggerType          string
	name                string
	penaltyDuration     int
	responseContent     string
	responseContentType string
	responseObjectName  string
	responseStatus      int
	rpsLimit            int
	uriDictName         string
	windowSize          string
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
	o, err := c.Globals.APIClient.UpdateERL(input)
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
	input.ERLID = c.id

	// NOTE: rateLimitActions is defined in ./create.go
	if c.action != "" {
		for _, a := range fastly.ERLActions {
			if c.action == string(a) {
				input.Action = fastly.ToPointer(a)
				break
			}
		}
	}

	if c.clientKeys != "" {
		clientKeys := strings.Split(strings.ReplaceAll(c.clientKeys, " ", ""), ",")
		input.ClientKey = &clientKeys
	}

	if c.featRevision > 0 {
		input.FeatureRevision = fastly.ToPointer(c.featRevision)
	}

	if c.httpMethods != "" {
		httpMethods := strings.Split(strings.ReplaceAll(c.httpMethods, " ", ""), ",")
		input.HTTPMethods = &httpMethods
	}

	// NOTE: rateLimitLoggers is defined in ./create.go
	if c.loggerType != "" {
		for _, l := range fastly.ERLLoggers {
			if c.loggerType == string(l) {
				input.LoggerType = fastly.ToPointer(l)
				break
			}
		}
	}

	if c.name != "" {
		input.Name = fastly.ToPointer(c.name)
	}

	if c.penaltyDuration > 0 {
		input.PenaltyBoxDuration = fastly.ToPointer(c.penaltyDuration)
	}

	if c.responseContent != "" && c.responseContentType != "" && c.responseStatus > 0 {
		input.Response = &fastly.ERLResponseType{
			ERLContent:     fastly.ToPointer(c.responseContent),
			ERLContentType: fastly.ToPointer(c.responseContentType),
			ERLStatus:      fastly.ToPointer(c.responseStatus),
		}
	}

	if c.responseObjectName != "" {
		input.ResponseObjectName = fastly.ToPointer(c.responseObjectName)
	}

	if c.rpsLimit > 0 {
		input.RpsLimit = fastly.ToPointer(c.rpsLimit)
	}

	if c.uriDictName != "" {
		input.URIDictionaryName = fastly.ToPointer(c.uriDictName)
	}

	// NOTE: rateLimitWindowSizes is defined in ./create.go
	if c.windowSize != "" {
		for _, w := range fastly.ERLWindowSizes {
			if c.windowSize == fmt.Sprint(w) {
				input.WindowSize = fastly.ToPointer(w)
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
	if c.responseContent != "" {
		state++
	}
	if c.responseContentType != "" {
		state++
	}
	if c.responseStatus > 0 {
		state++
	}
	if state > 0 && state < 3 {
		return errors.New("invalid flag use")
	}
	return nil
}

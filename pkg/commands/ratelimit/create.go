package ratelimit

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// rateLimitActions is a list of supported actions.
// It is used to construct the API input.
// It is also used to construct input for the --action enum flag.
// We build the flag input dynamically so we can avoid hardcoding the values.
// This is in case the underlying values in go-fastly change between releases.
var rateLimitActions = []fastly.ERLAction{
	fastly.ERLActionLogOnly,
	fastly.ERLActionResponse,
	fastly.ERLActionResponseObject,
}

// rateLimitActionFlagOpts is a string representation of rateLimitActions
// suitable for use within the enum flag definition below.
var rateLimitActionFlagOpts = func() (actions []string) {
	for _, a := range rateLimitActions {
		actions = append(actions, string(a))
	}
	return actions
}()

// rateLimitLoggers is a list of supported logger types.
// It is used to construct the API input.
// It is also used to construct input for the --logger-type enum flag.
// We build the flag input dynamically so we can avoid hardcoding the values.
// This is in case the underlying values in go-fastly change between releases.
var rateLimitLoggers = []fastly.ERLLogger{
	fastly.ERLLogAzureBlob,
	fastly.ERLLogBigQuery,
	fastly.ERLLogCloudFiles,
	fastly.ERLLogDataDog,
	fastly.ERLLogDigitalOcean,
	fastly.ERLLogElasticSearch,
	fastly.ERLLogFtp,
	fastly.ERLLogGcs,
	fastly.ERLLogGoogleAnalytics,
	fastly.ERLLogHeroku,
	fastly.ERLLogHoneycomb,
	fastly.ERLLogHTTP,
	fastly.ERLLogHTTPS,
	fastly.ERLLogKafta,
	fastly.ERLLogKinesis,
	fastly.ERLLogLogEntries,
	fastly.ERLLogLoggly,
	fastly.ERLLogLogShuttle,
	fastly.ERLLogNewRelic,
	fastly.ERLLogOpenStack,
	fastly.ERLLogPaperTrail,
	fastly.ERLLogPubSub,
	fastly.ERLLogS3,
	fastly.ERLLogScalyr,
	fastly.ERLLogSftp,
	fastly.ERLLogSplunk,
	fastly.ERLLogStackDriver,
	fastly.ERLLogSumoLogic,
	fastly.ERLLogSysLog,
}

// rateLimitLoggerFlagOpts is a string representation of rateLimitLoggers
// suitable for use within the enum flag definition below.
var rateLimitLoggerFlagOpts = func() (loggers []string) {
	for _, l := range rateLimitLoggers {
		loggers = append(loggers, string(l))
	}
	return loggers
}()

// rateLimitWindowSizes is a list of supported time window sizes.
// It is used to construct the API input.
// It is also used to construct input for the --window-size enum flag.
// We build the flag input dynamically so we can avoid hardcoding the values.
// This is in case the underlying values in go-fastly change between releases.
var rateLimitWindowSizes = []fastly.ERLWindowSize{
	fastly.ERLSize1,
	fastly.ERLSize10,
	fastly.ERLSize60,
}

// rateLimitWindowSizeFlagOpts is a string representation of rateLimitWindowSizes
// suitable for use within the enum flag definition below.
var rateLimitWindowSizeFlagOpts = func() (windowSizes []string) {
	for _, w := range rateLimitWindowSizes {
		windowSizes = append(windowSizes, fmt.Sprint(w))
	}
	return windowSizes
}()

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("create", "Create a rate limiter for a particular service and version").Alias("add")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.CmdClause.Flag("action", "The action to take when a rate limiter violation is detected").HintOptions(rateLimitActionFlagOpts...).EnumVar(&c.action, rateLimitActionFlagOpts...)
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("client-key", "Comma-separated list of VCL variable used to generate a counter key to identify a client").StringVar(&c.clientKeys)
	c.CmdClause.Flag("http-methods", "Comma-separated list of HTTP methods to apply rate limiting to").StringVar(&c.httpMethods)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("logger-type", "Name of the type of logging endpoint to be used when action is `log_only`").HintOptions(rateLimitLoggerFlagOpts...).EnumVar(&c.loggerType, rateLimitLoggerFlagOpts...)
	c.CmdClause.Flag("name", "A human readable name for the rate limiting rule").StringVar(&c.name)
	c.CmdClause.Flag("penalty-box-dur", "Length of time in minutes that the rate limiter is in effect after the initial violation is detected").IntVar(&c.penaltyDuration)
	c.CmdClause.Flag("response-content", "HTTP response body data").StringVar(&c.responseContent)
	c.CmdClause.Flag("response-content-type", "HTTP Content-Type (e.g. application/json)").StringVar(&c.responseContentType)
	c.CmdClause.Flag("response-status", "HTTP response status code (e.g. 429)").IntVar(&c.responseStatus)
	c.CmdClause.Flag("rps-limit", "Upper limit of requests per second allowed by the rate limiter").IntVar(&c.rpsLimit)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("window-size", "Number of seconds during which the RPS limit must be exceeded in order to trigger a violation").HintOptions(rateLimitWindowSizeFlagOpts...).EnumVar(&c.windowSize, rateLimitWindowSizeFlagOpts...)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base
	cmd.JSONOutput

	action              string
	autoClone           cmd.OptionalAutoClone
	clientKeys          string
	httpMethods         string
	loggerType          string
	manifest            manifest.Data
	name                string
	penaltyDuration     int
	responseContent     string
	responseContentType string
	responseStatus      int
	rpsLimit            int
	serviceName         cmd.OptionalServiceNameID
	serviceVersion      cmd.OptionalServiceVersion
	windowSize          string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	_, s := c.Globals.Token()
	if s == lookup.SourceUndefined {
		return fsterr.ErrNoToken
	}

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if err := c.responseFlagValidator(); err != nil {
		return fsterr.RemediationError{
			Inner:       err,
			Remediation: "When defining a response, all response flags (--response-content, --response-content-type, --response-status) should be set",
		}
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput()
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion.Number

	o, err := c.Globals.APIClient.CreateERL(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Created rate limiter '%s' (%s)", o.Name, o.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateERLInput {
	var input fastly.CreateERLInput

	if c.action != "" {
		for _, a := range rateLimitActions {
			if c.action == string(a) {
				input.Action = fastly.ERLActionPtr(a)
				break
			}
		}
	}

	if c.clientKeys != "" {
		clientKeys := strings.Split(strings.ReplaceAll(c.clientKeys, " ", ""), ",")
		input.ClientKey = &clientKeys
	}

	if c.httpMethods != "" {
		httpMethods := strings.Split(strings.ReplaceAll(c.httpMethods, " ", ""), ",")
		input.HTTPMethods = &httpMethods
	}

	if c.loggerType != "" {
		for _, l := range rateLimitLoggers {
			if c.loggerType == string(l) {
				input.LoggerType = fastly.ERLLoggerPtr(l)
				break
			}
		}
	}

	if c.name != "" {
		input.Name = fastly.String(c.name)
	}

	if c.penaltyDuration > 0 {
		input.PenaltyBoxDuration = fastly.Int(c.penaltyDuration)
	}

	if c.responseContent != "" && c.responseContentType != "" && c.responseStatus > 0 {
		input.Response = &fastly.ERLResponseType{
			ERLContent:     c.responseContent,
			ERLContentType: c.responseContentType,
			ERLStatus:      c.responseStatus,
		}
	}

	if c.rpsLimit > 0 {
		input.RpsLimit = fastly.Int(c.rpsLimit)
	}

	if c.windowSize != "" {
		for _, w := range rateLimitWindowSizes {
			if c.windowSize == fmt.Sprint(w) {
				input.WindowSize = fastly.ERLWindowSizePtr(w)
				break
			}
		}
	}

	return &input
}

// responseFlagValidator ensures if a user specifies one of the response flags,
// that they must specify ALL of the response flags.
func (c *CreateCommand) responseFlagValidator() error {
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

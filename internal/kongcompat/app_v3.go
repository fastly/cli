package kongcompat

// UsageForContextWithTemplate prints the usage of the command ctx reached
// as uc describes it (kingpin v3).
func (a *Application) UsageForContextWithTemplate(uc *UsageContext, ctx *ParseContext) error {
	return a.usageFor(ctx, uc)
}

// Model returns the command's data model (kingpin v3).
// parent is ignored, since the model finds its parent by itself.
func (c *CmdClause) Model(parent *CmdModel) *CmdModel {
	return c.model()
}

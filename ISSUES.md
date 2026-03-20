<div align="center">
  <h3 align="center">CLI Issues</h3>
  <p align="center">Best practices for submitting an issue to the Fastly CLI repository.</p>
</div>

## Issue Type: Bug

Issues related to the CLI behavior not working as intended. 

- The CLI crashes or exits with an unexpected error
- A command produces incorrect output or wrong results
- Commands or flags don't work as documented

**Example:** "When I run `fastly service list --json`, malformed JSON is produced."

## Issue Type: Feature Request

Issues related to suggesting improvements to the CLI:

- New commands or subcommands based on existing Fastly APIs
- Improved error messages or user experience
- Adding support for a third party integration

**Example:** "Add a `fastly service version validate` command, which already exists in the Fastly API."

## Fastly Support

CLI behavior specific to your environment or service / account should be routed to the Fastly support team @ support.fastly.com or support@fastly.com. 

- A feature is missing from your account / service
- Partial content is returned that you may not have access to with your current Fastly account role
- My site is not loading after a configuration change

**Example:** When running `fastly service vcl snippet create`, an error is thrown that the provided VCL is not valid

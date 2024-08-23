#!/usr/bin/env node
import { execFileSync } from "node:child_process";

import { pkgForCurrentPlatform } from "./package-helpers.js";

const pkg = pkgForCurrentPlatform();

let location;
try {
  // Check for the binary package from our "optionalDependencies". This
  // package should have been installed alongside this package at install time.
  location = (await import(pkg)).default;
} catch (e) {
  throw new Error(`The package "${pkg}" could not be found, and is needed by @fastly/cli.
    Either the package is missing or the platform/architecture you are using is not supported.
    If you are installing @fastly/cli with npm, make sure that you don't specify the
    "--no-optional" flag. The "optionalDependencies" package.json feature is used
    by @fastly/cli to install the correct binary executable for your current platform.
    If your platform is not supported, you can open an issue at https://github.com/fastly/cli/issues`);
}

try {
  execFileSync(location, process.argv.slice(2), { stdio: "inherit" });
} catch(err) {
  if (err.code) {
    // Spawning child process failed
    throw err;
  }
  if (err.status != null) {
    process.exitCode = err.status;
  }
}

import { pkgForCurrentPlatform } from "./package-helpers.js";

const pkg = pkgForCurrentPlatform();

let location;
try {
  // Check for the binary package from our "optionalDependencies". This
  // package should have been installed alongside this package at install time.
  location = (await import(pkg)).default;
} catch (e) {
  throw new Error(`The package "${pkg}" could not be found, and is needed by @fastly/cli.
    If you are installing @fastly/cli with npm, make sure that you don't specify the
    "--no-optional" flag. The "optionalDependencies" package.json feature is used
    by @fastly/cli to install the correct binary executable for your current platform.`);
}

export default location;

import { platform, arch } from "node:process";

const knownPackages = {
  "darwin arm64": "@fastly/cli-darwin-arm64",
  "darwin x64": "@fastly/cli-darwin-x64",
  "linux arm64": "@fastly/cli-linux-arm64",
  "linux x64": "@fastly/cli-linux-x64",
  "linux x64": "@fastly/cli-linux-386",
  "win32 arm64": "@fastly/cli-win32-arm64",
  "win32 x64": "@fastly/cli-win32-x64",
  "win32 x32": "@fastly/cli-win32-386",
};

export function pkgForCurrentPlatform() {
  let platformKey = `${platform} ${arch}`;
  if (platformKey in knownPackages) {
    return knownPackages[platformKey];
  }
  throw new Error(
    `Unsupported platform: "${platformKey}". "@fastly/cli does not have a precompiled binary for the platform/architecture you are using. You can open an issue on https://github.com/fastly/cli/issues to request for your platform/architecture to be included."`
  );
}

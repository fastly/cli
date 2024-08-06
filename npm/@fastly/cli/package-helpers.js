import { platform, arch } from "node:process";

export function pkgForCurrentPlatform() {
  return `@fastly/cli-${platform}-${arch}`;
}

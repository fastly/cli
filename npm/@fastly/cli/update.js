#!/usr/bin/env node

import { fileURLToPath } from "node:url";
import { dirname, join, parse } from "node:path";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import decompress from "decompress";
import decompressTargz from "decompress-targz";

const __dirname = dirname(fileURLToPath(import.meta.url));
const input = process.argv.slice(2).at(0);
const tag = input ? `v${input}` : "dev";

let packages = [
  {
    releaseAsset: `fastly_${tag}_darwin-arm64.tar.gz`,
    binaryAsset: "fastly",
    description: "The macOS (M-series) binary for the Fastly CLI",
    os: "darwin",
    cpu: "arm64",
  },
  {
    releaseAsset: `fastly_${tag}_darwin-amd64.tar.gz`,
    binaryAsset: "fastly",
    description: "The macOS (Intel) binary for the Fastly CLI",
    os: "darwin",
    cpu: "x64",
  },
  {
    releaseAsset: `fastly_${tag}_linux-arm64.tar.gz`,
    binaryAsset: "fastly",
    description: "The Linux (arm64) binary for the Fastly CLI",
    os: "linux",
    cpu: "arm64",
  },
  {
    releaseAsset: `fastly_${tag}_linux-amd64.tar.gz`,
    binaryAsset: "fastly",
    description: "The Linux (64-bit) binary for the Fastly CLI",
    os: "linux",
    cpu: "x64",
  },
  {
    releaseAsset: `fastly_${tag}_linux-386.tar.gz`,
    binaryAsset: "fastly",
    description: "The Linux (32-bit) binary for the Fastly CLI",
    os: "linux",
    cpu: "x32",
  },
  {
    releaseAsset: `fastly_${tag}_windows-arm64.tar.gz`,
    binaryAsset: "fastly.exe",
    description: "The Windows (arm64) binary for the Fastly CLI",
    os: "win32",
    cpu: "arm64",
  },
  {
    releaseAsset: `fastly_${tag}_windows-amd64.tar.gz`,
    binaryAsset: "fastly.exe",
    description: "The Windows (64-bit) binary for the Fastly CLI",
    os: "win32",
    cpu: "x64",
  },
  {
    releaseAsset: `fastly_${tag}_windows-386.tar.gz`,
    binaryAsset: "fastly.exe",
    description: "The Windows (32-bit) binary for the Fastly CLI",
    os: "win32",
    cpu: "x32",
  },
];

let response = await fetch(
  `https://api.github.com/repos/fastly/cli/releases/tags/${tag}`
);
if (!response.ok) {
  console.error(
    `Response from https://api.github.com/repos/fastly/cli/releases/tags/${tag} was not ok`,
    response
  );
  console.error(await response.text());
  process.exit(1);
}
response = await response.json();
const id = response.id;
let assets = await fetch(
  `https://api.github.com/repos/fastly/cli/releases/${id}/assets`
);
if (!assets.ok) {
  console.error(
    `Response from https://api.github.com/repos/fastly/cli/releases/${id}/assets was not ok`,
    assets
  );
  console.error(await response.text());
  process.exit(1);
}
assets = await assets.json();

let generatedPackages = [];

for (const info of packages) {
  const packageName = `cli-${info.os}-${info.cpu}`;
  const asset = assets.find((asset) => asset.name === info.releaseAsset);
  if (!asset) {
    console.error(
      `Can't find an asset named ${info.releaseAsset} for the release https://github.com/fastly/cli/releases/tag/${tag}`
    );
    process.exit(1);
  }
  const packageDirectory = join(__dirname, "../", packageName.split("/").pop());
  await mkdir(packageDirectory, { recursive: true });
  await writeFile(
    join(packageDirectory, "package.json"),
    packageJson(packageName, tag, info.description, info.os, info.cpu)
  );
  await writeFile(
    join(packageDirectory, "index.js"),
    indexJs(info.binaryAsset)
  );
  generatedPackages.push(packageName);
  const browser_download_url = asset.browser_download_url;
  const archive = await fetch(browser_download_url);
  if (!archive.ok) {
    console.error(`Response from ${browser_download_url} was not ok`, archive);
    console.error(await response.text());
    process.exit(1);
  }
  let buf = await archive.arrayBuffer();

  await decompress(Buffer.from(buf), packageDirectory, {
    // Remove the leading directory from the extracted file.
    strip: 1,
    plugins: [decompressTargz()],
    // Only extract the binary file and nothing else
    filter: (file) => parse(file.path).base === info.binaryAsset,
  });
}

// Generate `optionalDependencies` section in the root package.json
const rootPackageJsonPath = join(__dirname, "./package.json");
let rootPackageJson = await readFile(rootPackageJsonPath, "utf8");
rootPackageJson = JSON.parse(rootPackageJson);
rootPackageJson["optionalDependencies"] = generatedPackages.reduce(
  (acc, packageName) => {
    acc[`@fastly/${packageName}`] = `=${tag.substring(1)}`;
    return acc;
  },
  {}
);
await writeFile(rootPackageJsonPath, JSON.stringify(rootPackageJson, null, 4));

function indexJs(binaryAsset) {
  return `
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
const __dirname = dirname(fileURLToPath(import.meta.url))
let location = join(__dirname, '${binaryAsset}')
export default location
`;
}
function packageJson(name, version, description, os, cpu) {
  version = version.startsWith("v") ? version.replace("v", "") : version;
  return JSON.stringify(
    {
      name: `@fastly/${name}`,
      bin: {
        [name]: "fastly",
      },
      type: "module",
      version,
      main: "index.js",
      description,
      license: "Apache-2.0",
      preferUnplugged: false,
      os: [os],
      cpu: [cpu],
    },
    null,
    4
  );
}

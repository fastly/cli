#!/usr/bin/env bash
# Fetch the pinned Viceroy release assets and prepare them for embedding
# into the fastly CLI binary.
#
# Usage:
#   scripts/fetch-viceroy.sh                  # fetch for host platform only
#   scripts/fetch-viceroy.sh --host           # same as default
#   scripts/fetch-viceroy.sh --all            # fetch every supported platform
#   scripts/fetch-viceroy.sh --refresh-checksums [--host|--all]
#
# The script reads the desired version from pkg/embedded/viceroy/VICEROY_VERSION,
# downloads the matching upstream asset, verifies its SHA-256 against
# pkg/embedded/viceroy/checksums.txt (unless --refresh-checksums is set, in
# which case the file is rewritten), compresses the executable with zstd,
# and writes it to pkg/embedded/viceroy/assets/viceroy_<os>_<arch>.zst.

set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
pkg_dir="${repo_root}/pkg/embedded/viceroy"
assets_dir="${pkg_dir}/assets"
version_file="${pkg_dir}/VICEROY_VERSION"
checksums_file="${pkg_dir}/checksums.txt"

if [ ! -f "${version_file}" ]; then
  echo "fetch-viceroy: missing ${version_file}" >&2
  exit 1
fi

viceroy_version="$(tr -d '[:space:]' < "${version_file}")"
if [ -z "${viceroy_version}" ]; then
  echo "fetch-viceroy: VICEROY_VERSION is empty" >&2
  exit 1
fi

mode="host"
refresh_checksums=0
for arg in "$@"; do
  case "${arg}" in
    --host) mode="host" ;;
    --all) mode="all" ;;
    --refresh-checksums) refresh_checksums=1 ;;
    -h|--help)
      sed -n '2,16p' "$0"
      exit 0 ;;
    *)
      echo "fetch-viceroy: unknown argument: ${arg}" >&2
      exit 1 ;;
  esac
done

# Supported (os, arch) pairs. Keep in sync with pkg/embedded/viceroy/embed_*.go.
platforms="darwin amd64
darwin arm64
linux amd64
linux arm64
windows amd64"

# Pick the platforms to process based on --host / --all.
if [ "${mode}" = "all" ]; then
  selected_platforms="${platforms}"
else
  host_os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "${host_os}" in
    darwin) host_os="darwin" ;;
    linux) host_os="linux" ;;
    mingw*|msys*|cygwin*|windows*) host_os="windows" ;;
    *)
      echo "fetch-viceroy: unsupported host OS: ${host_os}" >&2
      exit 1 ;;
  esac

  host_arch="$(uname -m)"
  case "${host_arch}" in
    x86_64|amd64) host_arch="amd64" ;;
    arm64|aarch64) host_arch="arm64" ;;
    *)
      echo "fetch-viceroy: unsupported host arch: ${host_arch}" >&2
      exit 1 ;;
  esac

  selected_platforms="$(echo "${platforms}" | grep -E "^${host_os} ${host_arch}$" || true)"
  if [ -z "${selected_platforms}" ]; then
    echo "fetch-viceroy: no embedded Viceroy ships for ${host_os}/${host_arch}" >&2
    exit 1
  fi
fi

mkdir -p "${assets_dir}"

for tool in curl tar shasum zstd awk; do
  if ! command -v "${tool}" >/dev/null 2>&1; then
    echo "fetch-viceroy: missing required tool: ${tool}" >&2
    exit 1
  fi
done

# Discover the upstream download URL via the same metadata endpoint
# the runtime uses (pkg/github/github.go:23). The endpoint returns the
# URL of the latest release; we substitute the version string to pin
# to the version requested in VICEROY_VERSION, mirroring
# Asset.DownloadVersion at pkg/github/github.go:128. This keeps the
# build-time fetch in lockstep with the runtime download path if the
# upstream URL template ever changes.
meta_url_for() {
  local os="$1" arch="$2"
  local endpoint="https://developer.fastly.com/api/internal/releases/meta/viceroy/${os}/${arch}"
  curl -fsSL "${endpoint}"
}

sha256() {
  shasum -a 256 "$1" | awk '{print $1}'
}

lookup_checksum() {
  local label="$1"
  [ -f "${checksums_file}" ] || return 0
  awk -v label="${label}" '
    {
      sub(/#.*/, "")
      if (NF < 2) next
      if ($2 == label) { print $1; exit }
    }
  ' "${checksums_file}"
}

workdir="$(mktemp -d)"
trap 'rm -rf "${workdir}"' EXIT

# Accumulate (label, digest) pairs to write when --refresh-checksums is set.
new_sums_file="${workdir}/new_sums.txt"
: > "${new_sums_file}"

while IFS=' ' read -r os arch; do
  [ -z "${os}" ] && continue

  meta="$(meta_url_for "${os}" "${arch}" || true)"
  if [ -z "${meta}" ]; then
    echo "fetch-viceroy: failed to fetch release metadata for ${os}/${arch}" >&2
    exit 1
  fi
  latest_version="$(echo "${meta}" | awk -F'"' '/"version":/ { for (i = 1; i <= NF; i++) if ($i == "version") { print $(i+2); exit } }')"
  latest_url="$(echo "${meta}" | awk -F'"' '/"url":/ { for (i = 1; i <= NF; i++) if ($i == "url") { print $(i+2); exit } }')"
  if [ -z "${latest_version}" ] || [ -z "${latest_url}" ]; then
    echo "fetch-viceroy: malformed metadata response for ${os}/${arch}" >&2
    exit 1
  fi

  # Pin to the version requested in VICEROY_VERSION by substituting the
  # latest version string in the URL. Matches the runtime's
  # Asset.DownloadVersion behaviour.
  url="$(echo "${latest_url}" | awk -v from="${latest_version}" -v to="${viceroy_version}" '{gsub(from, to); print}')"
  archive_name="$(basename "${url}")"

  echo "fetch-viceroy: downloading ${os}/${arch} v${viceroy_version}"
  archive_path="${workdir}/${archive_name}"
  if ! curl -fsSL "${url}" -o "${archive_path}"; then
    echo "fetch-viceroy: download failed: ${url}" >&2
    exit 1
  fi

  extract_dir="${workdir}/${os}_${arch}"
  mkdir -p "${extract_dir}"
  tar -xzf "${archive_path}" -C "${extract_dir}"

  bin_name="viceroy"
  [ "${os}" = "windows" ] && bin_name="viceroy.exe"
  bin_path="${extract_dir}/${bin_name}"
  if [ ! -f "${bin_path}" ]; then
    bin_path="$(find "${extract_dir}" -type f -name "${bin_name}" -print -quit)"
  fi
  if [ -z "${bin_path}" ] || [ ! -f "${bin_path}" ]; then
    echo "fetch-viceroy: viceroy binary not found in ${archive_name}" >&2
    exit 1
  fi

  asset_label="viceroy_${os}_${arch}"
  digest="$(sha256 "${bin_path}")"

  if [ "${refresh_checksums}" -eq 1 ]; then
    printf '%s  %s\n' "${digest}" "${asset_label}" >> "${new_sums_file}"
  else
    expected="$(lookup_checksum "${asset_label}")"
    if [ -z "${expected}" ]; then
      echo "fetch-viceroy: no checksum on file for ${asset_label}. Re-run with --refresh-checksums and review the diff." >&2
      exit 1
    fi
    if [ "${digest}" != "${expected}" ]; then
      echo "fetch-viceroy: checksum mismatch for ${asset_label}" >&2
      echo "  expected: ${expected}" >&2
      echo "  got:      ${digest}" >&2
      exit 1
    fi
  fi

  zstd -q -19 --rm -f "${bin_path}" -o "${assets_dir}/${asset_label}.zst"
  echo "fetch-viceroy: wrote ${assets_dir}/${asset_label}.zst"
done <<< "${selected_platforms}"

# Rewrite checksums.txt when requested. Preserve any platform entries that
# weren't part of this run (so --host --refresh-checksums updates one line
# without losing the others).
if [ "${refresh_checksums}" -eq 1 ]; then
  merged_file="${workdir}/merged_sums.txt"
  : > "${merged_file}"

  if [ -f "${checksums_file}" ]; then
    awk '
      {
        line = $0
        sub(/#.*/, "")
        if (NF >= 2) {
          print $1 " " $2
        }
      }
    ' "${checksums_file}" >> "${merged_file}"
  fi

  while IFS= read -r line; do
    digest="$(echo "${line}" | awk '{print $1}')"
    label="$(echo "${line}" | awk '{print $2}')"
    grep -v " ${label}\$" "${merged_file}" > "${merged_file}.tmp" || true
    mv "${merged_file}.tmp" "${merged_file}"
    printf '%s %s\n' "${digest}" "${label}" >> "${merged_file}"
  done < "${new_sums_file}"

  {
    echo "# SHA-256 checksums of the raw Viceroy executables, pre-compression."
    echo "# Format matches \`sha256sum\`. Refresh via:"
    echo "#   scripts/fetch-viceroy.sh --all --refresh-checksums"
    echo "# Building with -tags viceroy_embed verifies the downloaded asset matches."
    echo "#"
    echo "# Pinned Viceroy version: see VICEROY_VERSION."
    echo
    sort -k2 "${merged_file}" | awk '{printf "%s  %s\n", $1, $2}'
  } > "${checksums_file}"
  echo "fetch-viceroy: refreshed ${checksums_file}"
fi

#!/usr/bin/env bash
set -euo pipefail

repository="Onicc/frp-panel"
version="edge"
github_proxy=""
agent_args=()

while (($#)); do
  case "$1" in
    --version)
      [[ $# -ge 2 ]] || { echo "--version requires a value" >&2; exit 2; }
      version="$2"
      shift 2
      ;;
    --github-proxy)
      [[ $# -ge 2 ]] || { echo "--github-proxy requires a value" >&2; exit 2; }
      github_proxy="$2"
      shift 2
      ;;
    *)
      agent_args+=("$1")
      shift
      ;;
  esac
done

case "$(uname -s)" in
  Linux) asset_os="linux" ;;
  Darwin) asset_os="darwin" ;;
  *) echo "Unsupported operating system: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) asset_arch="amd64" ;;
  arm64|aarch64) asset_arch="arm64" ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

asset="frp-panel-agent-${asset_os}-${asset_arch}"
if [[ "$version" == "latest" ]]; then
  release_url="https://github.com/${repository}/releases/latest/download"
else
  release_url="https://github.com/${repository}/releases/download/${version}"
fi
if [[ -n "$github_proxy" ]]; then
  release_url="${github_proxy%/}/${release_url}"
fi

temp_dir="$(mktemp -d)"
trap 'rm -rf -- "$temp_dir"' EXIT
binary="$temp_dir/$asset"
checksum="$temp_dir/checksums.txt"

curl --fail --location --proto '=https' --tlsv1.2 --retry 3 \
  --output "$binary" "$release_url/$asset"
curl --fail --location --proto '=https' --tlsv1.2 --retry 3 \
  --output "$checksum" "$release_url/checksums.txt"

expected="$(awk -v asset="$asset" '$2 == asset {print $1}' "$checksum")"
[[ -n "$expected" ]] || { echo "Asset is missing from checksums.txt" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$binary" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "$binary" | awk '{print $1}')"
fi
[[ "$actual" == "$expected" ]] || { echo "Checksum verification failed" >&2; exit 1; }
chmod 0755 "$binary"

runner=()
if [[ "$(id -u)" -ne 0 ]]; then
  runner=(sudo)
fi
"${runner[@]}" "$binary" service install "${agent_args[@]}"

echo "frp-panel-agent installed successfully; no files were written to $(pwd)."

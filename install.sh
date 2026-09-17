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

download_release_asset() {
  local url="$1"
  local output="$2"
  local attempts=1

  # The rolling edge release is replaced by CI. During that short window the
  # old release may already be gone while the new assets are not available
  # yet, so retry the complete asset download instead of failing on a 404.
  if [[ "$version" == "edge" ]]; then
    attempts=6
  fi

  while (( attempts > 0 )); do
    if curl --fail --location --proto '=https' --tlsv1.2 --retry 3 \
      --output "$output" "$url"; then
      return 0
    fi

    attempts=$((attempts - 1))
    if (( attempts > 0 )); then
      echo "Release asset is not available yet; retrying in 10 seconds ($attempts attempts left)" >&2
      sleep 10
    fi
  done

  echo "Unable to download release asset: $url" >&2
  echo "The $version release may still be publishing, or may not contain an asset for $(uname -s)/$(uname -m)." >&2
  return 1
}

download_release_asset "$release_url/$asset" "$binary"
download_release_asset "$release_url/checksums.txt" "$checksum"

expected="$(awk -v asset="$asset" '$2 == asset {print $1}' "$checksum")"
[[ -n "$expected" ]] || { echo "Asset is missing from checksums.txt" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  actual="$(sha256sum "$binary" | awk '{print $1}')"
else
  actual="$(shasum -a 256 "$binary" | awk '{print $1}')"
fi
[[ "$actual" == "$expected" ]] || { echo "Checksum verification failed" >&2; exit 1; }
chmod 0755 "$binary"

if [[ "$(id -u)" -eq 0 ]]; then
  "$binary" service install "${agent_args[@]}"
else
  sudo "$binary" service install "${agent_args[@]}"
fi

echo "frp-panel-agent installed successfully; no files were written to $(pwd)."

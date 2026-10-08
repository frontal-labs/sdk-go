#!/usr/bin/env bash
set -euo pipefail

apidiff() {
	go run golang.org/x/exp/cmd/apidiff@v0.0.0-20240823005443-9b4947da3948 "$@"
}

if [[ $# -ne 1 ]]; then
	echo "usage: scripts/check_apidiff.sh <previous-tag>" >&2
	exit 2
fi

previous_tag="$1"
current_root="$(pwd)"
temporary_root="$(mktemp -d)"
old_root="$temporary_root/old"
old_exports="$temporary_root/old-exports"
new_exports="$temporary_root/new-exports"
mkdir -p "$old_root" "$old_exports" "$new_exports"
trap 'rm -rf -- "$temporary_root"' EXIT

git archive "$previous_tag" | tar -x -C "$old_root"

export_module_packages() {
	module_root="$1"
	export_root="$2"
	packages_file="$3"
	(
		cd "$module_root"
		go list ./... > "$packages_file"
		while IFS= read -r package; do
			file_name="$(printf '%s' "$package" | tr '/.' '__').export"
			apidiff -w "$export_root/$file_name" "$package"
		done < "$packages_file"
	)
}

export_module_packages "$old_root" "$old_exports" "$temporary_root/old-packages"
export_module_packages "$current_root" "$new_exports" "$temporary_root/new-packages"

while IFS= read -r package; do
	if ! grep -Fxq "$package" "$temporary_root/new-packages"; then
		echo "incompatible API change: removed package $package" >&2
		exit 1
	fi
	file_name="$(printf '%s' "$package" | tr '/.' '__').export"
	changes="$(apidiff -incompatible "$old_exports/$file_name" "$new_exports/$file_name")"
	if [[ -n "$changes" ]]; then
		echo "$changes"
		exit 1
	fi
done < "$temporary_root/old-packages"

echo "No incompatible exported API changes since $previous_tag."

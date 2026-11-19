#!/usr/bin/env bash
# changelog-test.sh
#
# Checks changelog.sh and next-version.sh on a repository of their own, made in a temporary folder: a release whose Unreleased section lists changes ships those, under the version's heading and below a new empty Unreleased section; one with nothing written gets a section from the commits since the previous version tag (scopes that are not the app left out, a fix after a feature still a minor bump); a section already there is kept; notes prints a section without its heading; and set-version.sh writes the version into nsgo.yaml.
set -euo pipefail

scripts=$(cd "$(dirname "$0")" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$work"

git init -q
git config user.name test
git config user.email test@example.com
git config commit.gpgsign false
git config tag.gpgsign false

# commit <subject> makes an empty commit dated 2026-01-02, the date the sections take from the commit they are written at.
commit() {
  GIT_AUTHOR_DATE=2026-01-02T00:00:00Z GIT_COMMITTER_DATE=2026-01-02T00:00:00Z git commit -q --allow-empty -m "$1"
}

# expect <name> <want> <got>
expect() {
  if [ "$2" != "$3" ]; then
    printf 'changelog-test: %s\n--- want\n%s\n--- got\n%s\n' "$1" "$2" "$3" >&2
    exit 1
  fi
}

mkdir -p apps/mirra
printf 'name: Mirra\nversion: 0.3.0\n' >apps/mirra/nsgo.yaml
printf '# Changelog\n\n## Unreleased\n\n- A web version.\n\n## 0.3.0\n\n- Dark amber look.\n' >CHANGELOG.md
git add -A
commit "chore: start"
git tag v0.3.0

# Nothing worth a release yet: docs and website changes call for no version.
commit "docs: tidy the README"
commit "fix(website): link"
expect "no release" "" "$("$scripts/next-version.sh" HEAD)"

commit "feat: run in the browser"
commit "fix: keep the key"
expect "minor bump" "v0.4.0" "$("$scripts/next-version.sh" HEAD)"

"$scripts/set-version.sh" v0.4.0
expect "set-version" "version: 0.4.0" "$(grep '^version:' apps/mirra/nsgo.yaml)"

"$scripts/changelog.sh" add v0.4.0
expect "Unreleased promoted" "# Changelog

## Unreleased

## 0.4.0 - 2026-01-02

- A web version.

## 0.3.0

- Dark amber look." "$(cat CHANGELOG.md)"
expect "notes" "- A web version." "$("$scripts/changelog.sh" notes v0.4.0)"

# Adding the same version again keeps its section.
"$scripts/changelog.sh" add v0.4.0 2>/dev/null
expect "kept" "- A web version." "$("$scripts/changelog.sh" notes v0.4.0)"

git add -A
commit "chore(release): v0.4.0"
git tag v0.4.0

# Nothing written under Unreleased: the section comes from the commits, without the website's.
commit "fix: saving a work twice"
commit "fix(website): typo"
commit "perf: faster thumbnails"
expect "patch bump" "v0.4.1" "$("$scripts/next-version.sh" HEAD)"
"$scripts/changelog.sh" add v0.4.1
expect "from commits" "### Fixes

- Saving a work twice

### Performance

- Faster thumbnails" "$("$scripts/changelog.sh" notes v0.4.1)"
expect "order" "## Unreleased
## 0.4.1 - 2026-01-02
## 0.4.0 - 2026-01-02
## 0.3.0" "$(grep '^## ' CHANGELOG.md)"

# A tagged commit gets no version again.
git tag v0.4.1
expect "tagged" "" "$("$scripts/next-version.sh" HEAD)"

echo "changelog-test: ok"

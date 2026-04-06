#!/bin/bash
# Run this after initialising the wiki via GitHub web UI
# (create any page, then this script replaces it with full content)
set -e

rm -rf /tmp/pkgbadge-wiki-push
git clone "https://$(gh auth token)@github.com/Will-Luck/pkgbadge.wiki.git" /tmp/pkgbadge-wiki-push
cd /tmp/pkgbadge-wiki-push

cp "$(dirname "$0")"/*.md .
rm -f push-wiki.sh

git add -A
git commit -m "docs: full wiki documentation"
git push origin master
echo "Wiki published."

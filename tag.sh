#!/usr/bin/env bash

APP=$1

LAST_TAG=$(git tag --list "$APP-v*" --sort=-v:refname | head -n 1)

if [ -z "$LAST_TAG" ]; then
  VERSION="0.1.0"
else
  VERSION=${LAST_TAG#$APP-v}
fi

IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

COMMITS=$(git log "$LAST_TAG"..HEAD --pretty=format:%s)

if echo "$COMMITS" | grep -q "^BREAKING"; then
  ((MAJOR++)); MINOR=0; PATCH=0
elif echo "$COMMITS" | grep -q "^feat"; then
  ((MINOR++)); PATCH=0
elif echo "$COMMITS" | grep -q "^fix"; then
  ((PATCH++))
else
  echo "No version bump needed."
  exit 0
fi

NEW_VERSION="$MAJOR.$MINOR.$PATCH"

TAG="$APP-v$NEW_VERSION"

git tag "$TAG"

echo "Created tag $TAG"
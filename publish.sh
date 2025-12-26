NEW_VERSION=$(grep -m 1 '"version"' package.json | awk -F '"' '{print $4}')
echo "New version is v$NEW_VERSION"
git tag -a v$NEW_VERSION -m "Release v$NEW_VERSION" && echo "Created tag v$NEW_VERSION"
BUILD_TIME=`date +%FT%T%z`
GIT_REVISION=`git rev-parse --short HEAD`
GIT_BRANCH=`git rev-parse --symbolic-full-name --abbrev-ref HEAD`
GIT_DIRTY=`git diff-index --quiet HEAD -- || echo "✗-"`
GIT_TAG_DIRTY=`git describe --exact-match --tags > /dev/null 2> /dev/null || echo "+"`
GIT_TAG=`git describe --abbrev=0 --tags`

GITINFO="package main

var buildTime string = \"${BUILD_TIME}\"
var gitRevision string = \"${GIT_DIRTY}${GIT_REVISION}${GIT_TAG_DIRTY}\"
var gitBranch string = \"${GIT_BRANCH}\"
var gitTag string = \"${GIT_TAG}\""

echo "$GITINFO" > gitinfo.go


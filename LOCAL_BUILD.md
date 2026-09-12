# Local Build

Use this set of commands to perform a local build for tesing.

```bash
SEMVER=v0.0.999; echo ${SEMVER}
BUILD_DATE=$(gdate --utc +%FT%T.%3NZ); echo ${BUILD_DATE}
GIT_COMMIT=$(git rev-parse HEAD); echo ${GIT_COMMIT}

go build -ldflags "-X planeshift/cmd.semVer=${SEMVER} -X planeshift/cmd.buildDate=${BUILD_DATE} -X planeshift/cmd.gitCommit=${GIT_COMMIT} -X planeshift/cmd.gitRef=/refs/tags/${SEMVER}" && \
./planeshift version | jq .
```


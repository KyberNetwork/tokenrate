#!/bin/bash

set -euo pipefail

readonly DOCKER_PASSWORD=${DOCKER_PASSWORD:-}

if [[ -z "$DOCKER_PASSWORD" ]]; then
    echo 'DOCKER_PASSWORD is not available, aborting.'
    exit 1
fi

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

docker tag kybernetwork/tokenrate:"$TRAVIS_COMMIT" kybernetwork/tokenrate:"$TRAVIS_BRANCH"
if [[ -n "$TRAVIS_TAG" ]]; then
    docker tag kybernetwork/tokenrate:"$TRAVIS_COMMIT" kybernetwork/tokenrate:"$TRAVIS_TAG"
fi

docker push kybernetwork/tokenrate

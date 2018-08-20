#!/bin/bash

set -e

docker login -u="$QUAY_USERNAME" -p="$QUAY_PASSWORD" quay.io

tag=${TRAVIS_TAG#"v"}

docker build -t heetch/regula  .
docker tag heetch/regula quay.io/heetch/regula:latest
docker tag heetch/regula quay.io/heetch/regula:$tag
docker push quay.io/heetch/regula:$tag
docker push quay.io/heetch/regula:latest

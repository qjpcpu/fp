#!/bin/bash
COVERAGE=`go test -coverprofile c.out |grep coverage |grep -oE '[0-9]+[^%]*'`
echo "Coverage: ${COVERAGE}%"
curl -s "https://img.shields.io/badge/coverage-$COVERAGE-green" > codcov.svg

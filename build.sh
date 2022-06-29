#!/bin/bash
curl "https://img.shields.io/badge/coavrege-`go test -coverprofile c.out |grep coverage |grep -oE '[0-9]+[^%]*'`-green" > codcov.svg

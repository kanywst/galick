#!/bin/bash

echo "Delegating build to Makefile..."
cd "$(dirname "$0")/.." && make build

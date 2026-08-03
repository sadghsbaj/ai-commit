#!/bin/bash
set -x
tempDir=$(mktemp -d)
cd $tempDir
git log
git init
git config user.email "test@example.com"
git config user.name "Test User"
git -c commit.gpgsign=false commit --allow-empty -m "Initial empty commit"

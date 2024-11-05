#!/bin/bash

LOG_FILE="/tmp/llrss-test-$(date +%Y%m%d).log"

go test -count=1 -race -buildvcs -v ./... >$LOG_FILE
RES=$?
cat $LOG_FILE | sed ''/PASS/s//$(printf "\033[32mPASS\033[0m")/'' | sed ''/FAIL/s//$(printf "\033[31mFAIL\033[0m")/''
exit $RES

#!/bin/bash
source .travis/utils.sh
STATUS_CODE=0 # reset

####################
fold_start 'Check:go-vet'
    bash .travis/checks/check-go-vet.sh
    check_result $? "Check 'go vet': "
fold_end
####################

####################
fold_start 'Check:gofmt'
    bash .travis/checks/check-code-format.sh
    check_result $? "Check gofmt: "
fold_end
####################

####################
fold_start 'Check:signoff'
    git log  --pretty=oneline | head -n 10 # Very useful for debug we should keep this
    echo "----"
    python3 .travis/checks/check-signoff.py
    check_result $? "Check signoff: "
fold_end
####################

exit $STATUS_CODE
#!/bin/bash

# A script to validate the integrity of all generated test cases.
#
# For each test case, it performs two checks:
# 1. Terraform Validation: Runs `terraform plan` to ensure the .tfvar.json is valid for the .tf file.
# 2. JSON Schema Validation: Runs `jsonschema` to ensure the .tfvar.json is valid against the .schema.json.


OVERALL_SUCCESS=true
FAILED_TESTS=()

# Discover all test case directories that match the two-digit prefix (e.g. 01-foo-bar)
# This makes the script future-proof when new tests are added.

# shellcheck disable=SC2207 # we want word splitting here to create an array
TEST_DIRS=($(find tests -type d -regex ".*/[0-9][0-9]-[^/]*$" | sort))

TOTAL_CASES=${#TEST_DIRS[@]}

echo "Starting validation for ${TOTAL_CASES} test case(s)..."

# Extract just the two-digit index (directory name prefix) for logging consistency
for DIR_PATH in "${TEST_DIRS[@]}"; do
    DIR_NAME=$(basename "$DIR_PATH")
    i=${DIR_NAME:0:2}

    if [ -z "$DIR_PATH" ]; then
        echo "WARNING: Could not find directory for index $i. Skipping."
        continue
    fi

    echo "========================================"
    echo "Validating test case #$i: $DIR_NAME"
    echo "========================================"

    TEST_CASE_SUCCESS=true

    # --- Terraform Validation ---
    echo "--- Terraform Validation ---"
    (
        cd "$DIR_PATH"
        terraform init -upgrade >/dev/null 2>&1
        terraform plan -var-file="test.tfvar.json" -no-color >/dev/null
    )

    TF_EXIT_CODE=$?

    if [ $TF_EXIT_CODE -eq 0 ]; then
        echo "PASS:  Terraform plan PASSED"
    else
        echo "FAIL:  Terraform plan FAILED"
        TEST_CASE_SUCCESS=false
        # On failure, show the error
        (cd "$DIR_PATH" && terraform plan -var-file="test.tfvar.json")
    fi

    # --- JSON Schema Validation ---
    echo "--- JSON Schema Validation ---"
    jsonschema validate "$DIR_PATH/test.schema.json" "$DIR_PATH/test.tfvar.json" >/dev/null 2>&1
    JSON_EXIT_CODE=$?

    if [ $JSON_EXIT_CODE -eq 0 ]; then
        echo "PASS:  JSON Schema validation PASSED"
    else
        echo "FAIL:  JSON Schema validation FAILED"
        TEST_CASE_SUCCESS=false
        # On failure, show the error
        jsonschema validate "$DIR_PATH/test.schema.json" "$DIR_PATH/test.tfvar.json"
    fi

    if ! $TEST_CASE_SUCCESS; then
        OVERALL_SUCCESS=false
        FAILED_TESTS+=($DIR_NAME)
    fi

    echo "---"
    echo ""
done

# Final summary
echo "========================================"
if $OVERALL_SUCCESS; then
    echo "PASS:  All ${TOTAL_CASES} test cases passed validation!"
else
    echo "FAIL:  One or more test cases failed validation. Failed tests:"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    echo "Please review the errors above."
    exit 1
fi
echo "========================================"

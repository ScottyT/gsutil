# Usage:
# /bin/bash local-runner.sh script.sh
##
#!/usr/bin/env bash
set -eo pipefail

requireEnv() {
  test "${!1}" || (echo "local-runner: '$1' not found" >&2 && exit 1)
}

requireEnv GOOGLE_CLOUD_PROJECT
requireEnv CREDENTIAL_FILE

gcloud auth activate-service-account --key-file "${CREDENTIAL_FILE}"

export CLOUDSDK_CORE_PROJECT="${GOOGLE_CLOUD_PROJECT}"

"$@"
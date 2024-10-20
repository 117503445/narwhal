#!/bin/bash

# Capture stack trace
export RUST_BACKTRACE=1

if [ -z "$VALIDATOR_ID" -a "$KUBERNETES_PORT" ]; then
    export VALIDATOR_ID=${HOSTNAME##*-}
    # assuming that WORKER_ID isn't also set.
    # currently they match the validator it is assigned to.
    export WORKER_ID=${WORKER_ID:=$VALIDATOR_ID}
fi


# Environment variables to use on the script
NODE_BIN="./bin/node"
PRIMARY_KEYS_PATH=${KEYS_PATH:="/validators/validator-$VALIDATOR_ID/primary-key.json"}
WORKER_KEYS_PATH=${KEYS_PATH:="/validators/validator-$VALIDATOR_ID/worker-key.json"}
COMMITTEE_PATH=${COMMITTEE_PATH:="/validators/committee.json"}
WORKERS_PATH=${WORKERS_PATH:="/validators/workers.json"}
PARAMETERS_PATH=${PARAMETERS_PATH:="/validators/parameters.json"}
DATA_PATH=${DATA_PATH:="/validators"}

if [[ "$CLEANUP_DISABLED" = "true" ]]; then
  echo "Will not clean up existing directories..."
else
  if [[ "$NODE_TYPE" = "primary" ]]; then
    # Clean up only the primary node's data
    rm -r "${DATA_PATH}/validator-$VALIDATOR_ID/db-primary"
  elif [[ "$NODE_TYPE" = "worker" ]]; then
    # Clean up only the specific worker's node data
    rm -r "${DATA_PATH}/validator-$VALIDATOR_ID/db-worker-${WORKER_ID}"
  fi
fi

echo "NODE_BIN=$NODE_BIN, PRIMARY_KEYS_PATH=$PRIMARY_KEYS_PATH, WORKER_KEYS_PATH=$WORKER_KEYS_PATH, COMMITTEE_PATH=$COMMITTEE_PATH, WORKERS_PATH=$WORKERS_PATH, PARAMETERS_PATH=$PARAMETERS_PATH, DATA_PATH=$DATA_PATH"

# If this is a primary node, then run as primary
if [[ "$NODE_TYPE" = "primary" ]]; then
  echo "Bootstrapping primary node"

  LOG_PATH="/logs/validator-$VALIDATOR_ID-primary.log"
  echo "" > $LOG_PATH

  $NODE_BIN $LOG_LEVEL run \
  --primary-keys $PRIMARY_KEYS_PATH \
  --worker-keys $WORKER_KEYS_PATH \
  --committee $COMMITTEE_PATH \
  --workers $WORKERS_PATH \
  --store "${DATA_PATH}/validator-$VALIDATOR_ID/db-primary" \
  --parameters $PARAMETERS_PATH \
  primary > $LOG_PATH 2>&1
elif [[ "$NODE_TYPE" = "worker" ]]; then
  echo "Bootstrapping new worker node with id $WORKER_ID"

  LOG_PATH="/logs/validator-$VALIDATOR_ID-worker-$WORKER_ID.log"
  echo "" > $LOG_PATH

  $NODE_BIN $LOG_LEVEL run \
  --primary-keys $PRIMARY_KEYS_PATH \
  --worker-keys $WORKER_KEYS_PATH \
  --committee $COMMITTEE_PATH \
  --workers $WORKERS_PATH \
  --store "${DATA_PATH}/validator-$VALIDATOR_ID/db-worker-$WORKER_ID" \
  --parameters $PARAMETERS_PATH \
  worker --id $WORKER_ID > $LOG_PATH 2>&1
elif [[ "$NODE_TYPE" = "qexecutor" ]]; then
  echo "Bootstrapping new qexecutor node with id $WORKER_ID"

  LOG_PATH="/logs/validator-$EXECUTOR_ID-qexecutor.log"
  echo "" > $LOG_PATH

  ./bin/q executor > $LOG_PATH 2>&1
elif [[ "$NODE_TYPE" = "q-worker-master" ]]; then
  echo "Bootstrapping new worker-master node with id $WORKER_MASTER_ID"

  LOG_PATH="/logs/validator-$WORKER_MASTER_ID-q-worker-master.log"
  echo "" > $LOG_PATH

  ./bin/q worker-master > $LOG_PATH 2>&1
else
  echo "Unknown provided value for parameter: NODE_TYPE=$NODE_TYPE"
  exit 1
fi


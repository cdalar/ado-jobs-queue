#!/bin/bash
make
# set -a; source .env; set +a
source .env
./ado_jobs_queue --url "$URL" --token "$TOKEN" > status.json
cat status.json | jq
waiting=$(cat status.json | jq '.count.waiting')
running=$(cat status.json | jq '.count.running')
total=$(cat status.json | jq '.count.total')

if [ $waiting -gt 0 ] || [ $total -eq 0 ]; then
  echo "Jobs in queue, Spin up a new runner"
  ONCTL_CLOUD=hetzner
  cd agents; onctl up -a azure/agent-pool.sh --dot-env .env.test; cd ..
fi
if [ $waiting -eq 0 ] && [ $running -eq 0 ] && [ $total -ne 0 ]; then
  echo "No jobs in queue, deleting vms"
  ONCTL_CLOUD=hetzner
  cd agents; onctl delete all -f; cd ..
  ./ado_jobs_queue --url "$URL" --token "$TOKEN"
fi
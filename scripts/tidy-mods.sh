#!/bin/sh
set -e
export GOWORK=off
export GOPROXY=https://goproxy.io,https://proxy.golang.org,direct
export GOSUMDB=off
tidy_one() {
  d="$1"
  n=1
  while [ "$n" -le 5 ]; do
    echo "=== tidy $d (attempt $n) ==="
    if (cd "/build/$d" && go mod tidy); then
      echo "OK $d"
      return 0
    fi
    echo "retry $d after failure"
    n=$((n + 1))
    sleep 2
  done
  echo "FAILED $d"
  return 1
}
for d in packages/go-shared services/gateway services/user-service services/catalog-service services/order-service services/payment-service services/delivery-service services/notification-service; do
  tidy_one "$d"
done
echo DONE

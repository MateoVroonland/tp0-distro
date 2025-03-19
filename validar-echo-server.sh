#!/bin/bash

TEST_MSG="Mensaje de test"
SERVER_PORT=12345
NETWORK_NAME="tp0_testing_net"

response=$(docker run --rm --network $NETWORK_NAME busybox \
   sh -c "echo '$TEST_MSG' | nc -w 5 server $SERVER_PORT" 2>/dev/null)

if [ "$response" = "$TEST_MSG" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
#!/bin/bash

# Set the target hostname
TARGET="speech.googleapis.com"

# Get the IP addresses of the target
get_ip_addresses() {
    # Use dig to get the A records for the target
    dig +short "$TARGET"
}

# Monitor the number of connections every 5 seconds
while true; do
    # Resolve the target hostname to IP addresses
    IP_ADDRESSES=$(get_ip_addresses)

    # Initialize a counter for connections
    CONNECTIONS=0

    # Count the number of connections to each resolved IP address
    for IP in $IP_ADDRESSES; do
        COUNT=$(netstat -an | grep ":443" | grep "$IP" | wc -l)
        CONNECTIONS=$((CONNECTIONS + COUNT))
    done

    # Output the total number of connections
    echo "Number of connections to $TARGET: $CONNECTIONS"

    # Wait for 5 seconds before the next check
    sleep 1
done


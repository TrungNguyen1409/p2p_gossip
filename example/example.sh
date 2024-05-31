#!/bin/bash

API_ADDRESS="127.0.0.1"
API_PORT="9001"

string_to_bytes() {
    local string="$1"
    echo -n "$string" | xxd -p | tr -d '\n'
}

create_announce_message() {
    TTL=1
    RESERVED=1
    DATATYPE=1
    MESSAGE_DATA="Calling announce"

    (
        printf "%02x" $TTL  # TTL
        printf "%02x" $RESERVED  # RESERVED
        printf "%04x" $DATATYPE  # DATATYPE
        string_to_bytes "$MESSAGE_DATA"  # MessageData
    )
}

create_notify_message() {
    RESERVED=1
    DATATYPE=2

    (
        printf "%04x" $RESERVED  # RESERVED
        printf "%04x" $DATATYPE  # DATATYPE
    )
}

create_notification_message() {
    MESSAGE_ID=100
    DATATYPE=3
    MESSAGE_DATA="Notification data"

    (
        printf "%04x" $MESSAGE_ID  # MESSAGE_ID
        printf "%04x" $DATATYPE  # DATATYPE
        string_to_bytes "$MESSAGE_DATA"  # MessageData
    )
}

create_validation_message() {
    MESSAGE_ID=101
    RESERVED=0

    (
        printf "%04x" $MESSAGE_ID  # MESSAGE_ID
        printf "%04x" $RESERVED  # RESERVED
    )
}

send_message() {
    local message_type=$1
    local message=$2

    local length=$((4 + ${#message} / 2))

    (
        printf "%04x" $length  # Length of the message including header
        printf "%04x" $message_type  # MessageType
        echo -n "$message"
    ) | xxd -p -r | nc -v $API_ADDRESS $API_PORT

    echo ""
}

echo "Sending AnnounceMsg"
send_message 500 "$(create_announce_message)"

echo "Sending NotifyMsg"
send_message 501 "$(create_notify_message)"

echo "Sending NotificationMsg"
send_message 502 "$(create_notification_message)"

echo "Sending ValidationMsg"
send_message 503 "$(create_validation_message)"

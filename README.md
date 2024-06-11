# Gossip

- Thua Duc Nguyen
- Duc Trung Nguyen

### Prerequisites
- Taskfile: https://taskfile.dev/

### Compile server / client
```bash
task clean
task build_client 
task build_server 
```

### Usage of client 

```
Usage: ./gossip_client [options]
Options:
  -a    Send a GOSSIP_ANNOUNCE message
  -d string
        GOSSIP host module IP
  -m string
        GOSSIP host module port
  -n    Send a GOSSIP_NOTIFY message
  -p int
        GOSSIP host module port

Examples:
  Send a GOSSIP_ANNOUNCE message:
    ./gossip_client -a -d 127.0.0.1 -p 9001 -m announce_message

  Send a GOSSIP_NOTIFY message:
    ./gossip_client -n -d 127.0.0.1 -p 9001 -m notify_message
```
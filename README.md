# Gossip

- Thua Duc Nguyen
- Duc Trung Nguyen

### Prerequisites
- Taskfile: https://taskfile.dev/

### Compile server / client
```bash
task build
```

### Usage of client 

```
Usage: ./client [options]
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
    ./client -a -d 127.0.0.1 -p 9001 -m announce_message

  Send a GOSSIP_NOTIFY message:
    ./client -n -d 127.0.0.1 -p 9001 -m notify_message
```


### Test NotifyMsg
```

1. task build
2. ./bootstrapper : to start bootstrapping server
3. ./server run first server
4. ./server run second test server 

// imitate another module send notify message to gossip API
5. ./client -n -d  127.0.0.1 -p 9001 -t 1  

// trigger one gossip node spread notify message to its neighbour
6. ./client -a -d 127.0.0.1 -p 51142 -m randomm

all receiving gossip node will handle the notify message if there is matching in notify Channel
```


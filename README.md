# Tunnlr

Tnnlr is a simple utility to managing ssh tunnels.  It is currently a very ugly work in progress.

## Usage

```bash
# Get the tool
go get github.com/turtlemonvh/tnnlr

# Install binary
go install $GOPATH/src/github.com/turtlemonvh/tnnlr/tnnlr

# Create a config file with tunnels for this project
cat > $GOPATH/src/github.com/turtlemonvh/tnnlr/.tunnlr << EOF
[
    {
        "defaultUrl": "/", 
        "host": "52.33.93.76", 
        "localPort": 15673, 
        "name": "rabbitmq_dashboard", 
        "remotePort": 15672
    }, 
    {
        "defaultUrl": "/ui/", 
        "host": "53.34.92.76", 
        "localPort": 8503, 
        "name": "consul_dashboard", 
        "remotePort": 8500
    }
]
EOF

# Start up the web ui on localhost:8080
tnnlr

# Go to localhost:8080 and click the "Reload Tunnels from File" button.

```

## TODO

- Updates and cleanup to configuration
- Option to load whole sets of tunnels at a time easily, via file select in browser
- Less ugly code
- Less ugly UI
- Command line interface
- Poll in the background and keep tunnels open
- Kill runnels on shutdown

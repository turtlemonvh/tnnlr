# Tunnlr

Tnnlr is a simple utility to managing ssh tunnels.  It is currently a very ugly work in progress, but it does work 

* keep tunnels open
* compose new tunnels using a helpful ui
* reload a whole set of tunnels quickly

## Usage

```bash
# Get the tool
$ go get github.com/turtlemonvh/tnnlr

# Install binary
$ ( cd $GOPATH/src/github.com/turtlemonvh/tnnlr && go install ./tnnlr )

# Create a config file with tunnels for this project
# You can also just start the server and create it through the UI
$ cat > $GOPATH/src/github.com/turtlemonvh/tnnlr/.tunnlr << EOF
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
$ tnnlr

# Go to localhost:8080 and click the "Reload Tunnels from File" button.

# Check help docs for more information
$ tnnlr -h
APP:
cli

COMMAND:
tnnlr


AVAILABLE SUBCOMMANDS:
ls : List running tunnels
help : Print this help message

PARSING ORDER: (set values will override in this order)
CLI Flag > Environment

VARIABLES:
+-------------+---------+----------+-----------+----------------------------------------+
|    FLAG     | DEFAULT | REQUIRED | ENV NAME  |              DESCRIPTION               |
+-------------+---------+----------+-----------+----------------------------------------+
| --log-level | info    | No       | LOG_LEVEL | Logging levels. Options are:           |
|             |         |          |           | [panic,fatal,error,warning,info,debug] |
| --tunnels   | .tnnlr  | No       | TUNNELS   | Configuration file listing             |
|             |         |          |           | tunnels to load.                       |
| --ssh-exec  | ssh     | No       | SSH_EXEC  | The executable process to use          |
|             |         |          |           | for ssh. Can be a full path or         |
|             |         |          |           | just a cmd name.                       |
| --port      |    8080 | No       | PORT      | The port to run the webserver          |
|             |         |          |           | on.                                    |
+-------------+---------+----------+-----------+----------------------------------------+


Authors:
Timothy Van Heest (timothy@ionic.com)

```

## TODO

- Serve logfiles through UI
- Options to let tunnels continue running on shutdown
- Option for https default url
- Option to load whole sets of tunnels at a time easily, via file select in browser
- Less ugly code
- Less ugly UI
- Command line interface
- Make `ls` command do something useful


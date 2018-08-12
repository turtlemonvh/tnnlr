# Tunnlr

Tnnlr is a simple utility to managing ssh tunnels.  It is currently a very ugly work in progress, but it does help to do the following

* keep ssh tunnels open
* compose new tunnels using a helpful ui
* reload a whole set of tunnels quickly

Tnnlr shells out to a local version of ssh instead of bundling in its own ssh utilities.  This was done so that complex local ssh configurations are respected by default.

So if you find yourself commonly launching a set of the same tunnels to access admin apis and other utilies when working on a project, tnnlr may save you some time.

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
$ go run tnnlr -h
APP:
cli

COMMAND:
tnnlr


AVAILABLE SUBCOMMANDS:
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
|             |         |          |           | tunnels. This can be read from         |
|             |         |          |           | and written to via the web UI.         |
| --ssh-exec  | ssh     | No       | SSH_EXEC  | The executable to use                  |
|             |         |          |           | for ssh. Can be a full path or         |
|             |         |          |           | just a command name that works         |
|             |         |          |           | in your shell.                         |
| --port      |    8080 | No       | PORT      | The port to run the server on          |
|             |         |          |           | for the web UI.                        |
+-------------+---------+----------+-----------+----------------------------------------+


Authors:
Timothy Van Heest (timothy@ionic.com)
```

## Web UI

It's not pretty but it works.

![Alt text](webui.png?raw=true "Web UI")

## TODO

- Options to let tunnels continue running on shutdown
- Option for https default url
- Option to load whole sets of tunnels at a time easily, via file select in browser
- Less ugly code
- Less ugly UI
- Command line interface
- Make `ls` command do something useful


# Tnnlr [![Build Status](https://travis-ci.org/turtlemonvh/tnnlr.png?branch=master)](https://travis-ci.org/turtlemonvh/tnnlr)

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
$ cat > $GOPATH/src/github.com/turtlemonvh/tnnlr/.tnnlr << EOF
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

## Tips

### SSH config

For your ssh config (`~/.ssh/config`), something like this works well.  Note that `StrictHostKeyChecking` means you won't get any warning about conneting to new hosts. 

```
Host *
  User myusername
  StrictHostKeyChecking no
  ServerAliveInterval 60
```

Depending on your network configuration, you may want to use [the `TCPKeepAlive` setting](https://unix.stackexchange.com/questions/34004/how-does-tcp-keepalive-work-in-ssh) as well.

### Alternatives

There are many other ssh tunnel management tools that are more mature than this.  I created my own for a mixture of these reasons

* I wanted to be able to configure it programmatically from the output of other tools (e.g. inventory scripts). This is why the config is in json.
* I wanted the tunnels to be project specific. Different projects have different configurations, and I may only be able to create the tunnel when I'm VPN'd to a specific network.
* I wanted something that would repect my ssh config, which is pretty complex.
* I mainly wanted this for connecting to dashboards for monitoring products (Kibana, Grafana, CheckMK), so I wanted to be able to include a default url that I could just click on to get to the dashboard I wanted to view.
* I kept forgetting the sytax for creating an ssh tunnel, and I wanted to make sure the tool would have an option show that so I could run the tunnel directly (via ssh on the command line).
* I wanted something that would work on multiple OSes (Mac, Windows, Linux).
* I wanted a web UI. In addition to being easy to work with, this means I can use this same program to manage tunnels in a headless VM.

Some things that seem to be present in other tools that are missing here

* SOCKS proxy
* UN/PW authentication for a tunnel
* A less ugly UI
* Tunnels that are created automatically on startup

Sample of example tools (there are [many more](https://github.com/search?q=ssh+tunnel)):

* https://www.tynsoe.org/v2/stm/ and https://www.opoet.com/pyro/
    * MacOS
* https://github.com/tinned-software/ssh-tunnel-manager
    * A collection of bash scripts for managing persistent tunnels
* https://github.com/pahaz/sshtunnel
    * Python library, so not a tool
    * I used this in v1, but had issues with tunnels dying
* https://github.com/agebrock/tunnel-ssh
    * NodeJS library
* https://github.com/jfifield/sshtunnel
    * Java tool
    * Limited conifguration

## TODO

- Options to let tunnels continue running on shutdown
- Option for https default url
- Option to load whole sets of tunnels at a time easily, via file select in browser
- Option for un/pw auth for ssh connections (pub/priv key only right now)
- Option for multiple named dashboard views for a single tunnel
- Less ugly code
- Less ugly UI
- Command line interface
- Make `ls` command do something useful
- Move from dep to go modules

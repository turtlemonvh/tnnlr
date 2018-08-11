package tnnlr

var homePage string = `
<!doctype html>
<head>
    <title>Tnnlr</title>
    <style>
        #tips {
            float: left;
            clear: left;
        }
        #tips > h2 {
            float: none;
        }
        table {
            float: left;
            clear: left;
        }
        table, tr,td {
            border: 1px solid black;
            padding: 2px;
        }
        .msg {
            font-weight: bold;
            border: 1px solid red;
            padding: 10px;
            float: left;
            clear: both;
        }
        form {
            float: left;
            clear: both;
            margin: 10px 0px 10px 0px;
        }
        h2 {
            float: left;
            clear: left;
            margin: 20px 0px 20px 0px;
        }
        .submit input {
            float: right;
        }
    </style>
</head>
<body>
    {{ if $.HasMessages }}
    <h2>Messages</h2>
        {{range $nmsg, $msg := $.Messages }}
            <p class="msg">{{ $msg.String }}</p>
        {{ end }}
    {{ end }}

    <h2>Existing tunnels</h2>
    <table>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Host</th>
            <th>Local Port</th>
            <th>Remote Port</th>
            <th>Default URL</th>
            <th>Bash Command</th>
            <th>Is Alive?</th>
            <th>Close</th>
            <th>Reload</th>
        </tr>
    {{range $tunnelId, $tunnel := $.Tunnels }}
        <tr>
            <td>{{ $tunnelId }}</td>
            <td>{{ $tunnel.Name }}</td>
            <td>{{ $tunnel.Host }}</td>
            <td>{{ $tunnel.LocalPort }}</td>
            <td>{{ $tunnel.RemotePort }}</td>
            <td><a href="http://localhost:{{ $tunnel.LocalPort }}{{ $tunnel.DefaultUrl }}" target="_blank">http://localhost:{{ $tunnel.LocalPort }}{{ $tunnel.DefaultUrl }}</a></td>
            <td><a href="bash_command/{{ $tunnelId }}/" target="_blank">Show command</a></td>
            <td>{{ $tunnel.IsAlive }}</td>
            <td><a href="remove/{{ $tunnelId }}/">Remove</a></td>
            <td><a href="reload/{{ $tunnelId }}/">Reload</a></td>
        </tr>
    {{ end }}
    </table>

    <form action="/save/" method="post">
        <input type="submit" value="Save Tunnels to File">
    </form>
    <form action="/reload/" method="post">
        <input type="submit" value="Reload Tunnels from File">
    </form>

    <h2>Add new tunnel</h2>
    <form class="new_tunnel" action="/add/" method="post">
        <table>
        <tr>
            <td>Tunnel Name</td>
            <td><input type="text" name="name"></td>
        </tr>
        <tr>
            <td>Host</td>
            <td><input type="text" name="host"></td>
        </tr>
        <tr>
            <td>SSH Username</td>
            <td><input type="text" name="username"></td>
        </tr>
        <tr>
            <td>Local Port</td>
            <td><input type="text" name="localPort"></td>
        </tr>
        <tr>
            <td>Remote Port</td>
            <td><input type="text" name="remotePort"></td>
        </tr>
        <tr>
            <td>Default URL</td>
            <td><input type="text" name="defaultUrl"></td>
        </tr>
        <tr>
            <td colspan="2" class="submit"><input type="submit" value="Submit"></td>
        </tr>
        </table>
        
    </form>

    <div id="tips">
        <hr>
        <h2>Tips</h2>
        <ul>
            <li>
            The process may be marked "not alive" because of a network timeout.  Try reloading this page to check status again.
            </li>
            <li>
            Running "reload" both re-loads the definition of a process disk and restarts that process.  Be sure to save any edited process state to disk before reloading.
            </li>
        </ul>
    </div>

</body>
`

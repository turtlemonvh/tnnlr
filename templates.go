package main

var homePage string = `
<!doctype html>
<head>
    <title>Port Fwd</title>
    <style>
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
        input[type=submit] {
            font-size: 2em;
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
            <p class="msg">{{ $msg }}</p>
        {{ end }}
    {{ end }}

    <h2>Existing tunnels</h2>
    <form action="/save/" method="post">
        <input type="submit" value="Save Tunnels">
    </form>
    <form action="/reload/" method="post">
        <input type="submit" value="Reload Tunnels from File">
    </form>
    <table>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Host</th>
            <th>Local Port</th>
            <th>Remote Port</th>
            <th>Default URL</th>
            <th>Bash Command</th>
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
            <td><a href="remove/{{ $tunnelId }}/">Remove</a></td>
            <td><a href="reload_one/{{ $tunnelId }}/">Reload</a></td>
        </tr>
    {{ end }}
    </table>

    <h2>Add new tunnel</h2>
    <form class="new_tunnel" action="/add/" method="post">
        <table>
        <tr>
            <td>Name</td>
            <td><input type="text" name="name"></td>
        </tr>
        <tr>
            <td>Host</td>
            <td><input type="text" name="host"></td>
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
</body>
`

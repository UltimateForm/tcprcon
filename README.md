# tcprcon

- [tcprcon](#tcprcon)
  - [Features](#features)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Interactive Mode](#interactive-mode)
    - [Single Command Mode](#single-command-mode)
    - [Using Environment Variable for Password](#using-environment-variable-for-password)
  - [CLI Flags](#cli-flags)
  - [Using as a Library](#using-as-a-library)
    - [Streaming Responses](#streaming-responses)
  - [License](#license)


A fully native RCON client implementation, zero third parties*

<sub>*except for other golang maintained packages about terminal emulators, until i fully master tty :(</sub>

![tcprcon demo](.meta/demo.png)

## Features

- **Interactive Terminal UI**: full-screen exclusive TUI (like vim or nano)
- **Single Command Mode**: execute a single RCON command and exit
- **Multiple Authentication Methods**: supports password via CLI flag, environment variable (`rcon_password`), or secure prompt
- **Configurable Logging**: syslog-style severity levels for debugging
- **Installable as library**: use the RCON client in your own Go projects, ([see examples](#using-as-a-library))

## Installation

```bash
go install github.com/UltimateForm/tcprcon@latest
```

Or build from source:

<sub>note: requires golang 1.22+</sub>

```bash
git clone https://github.com/UltimateForm/tcprcon.git
cd tcprcon
go build -o tcprcon .
```

## Usage

### Interactive Mode

```bash
tcprcon --address=192.168.1.100 --port=7778
```

### Single Command Mode

```bash
tcprcon --address=192.168.1.100 --cmd="playerlist"
```

### Using Environment Variable for Password

```bash
export rcon_password="your_password"
tcprcon --address=192.168.1.100
```

## CLI Flags

```
  -address string
    	RCON address, excluding port (default "localhost")
  -cmd string
    	command to execute, if provided will not enter into interactive mode
  -log uint
    	sets log level (syslog severity tiers) for execution (default 4)
  -port uint
    	RCON port (default 7778)
  -pw string
    	RCON password, if not provided will attempt to load from env variables, if unavailable will prompt
```

## Using as a Library

The RCON client can be used as a library in your own Go projects:

```go
import (
    "github.com/UltimateForm/tcprcon/pkg/rcon"
    "github.com/UltimateForm/tcprcon/pkg/common_rcon"
    "github.com/UltimateForm/tcprcon/pkg/packet"
)

func main() {
    client, err := rcon.New("192.168.1.100:7778")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Authenticate
    success, err := common_rcon.Authenticate(client, "your_password")
    if err != nil || !success {
        panic("auth failed")
    }

    // Send command
    execPacket := packet.New(client.Id(), packet.SERVERDATA_EXECCOMMAND, []byte("playerlist"))
    client.Write(execPacket.Serialize())

    // Read response
    response, err := packet.Read(client)
    if err != nil {
        panic(err)
    }
    fmt.Println(response.BodyStr())
}
```

### Streaming Responses

For continuous listening (e.g., server broadcasts or multiple responses), use `CreateResponseChannel`:

<sub>usually you will want a more ellegant way of handling the concurrent nature of this, this example is just for illustration</sub>

```go
import (
    "context"
    "fmt"
    "io"

    "github.com/UltimateForm/tcprcon/pkg/rcon"
    "github.com/UltimateForm/tcprcon/pkg/common_rcon"
    "github.com/UltimateForm/tcprcon/pkg/packet"
)

func main() {
    client, _ := rcon.New("192.168.1.100:7778")
    defer client.Close()

    common_rcon.Authenticate(client, "your_password")

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create a channel that streams incoming packets
    packetChan := packet.CreateResponseChannel(client, ctx)

    // Send a command
    execPacket := packet.New(client.Id(), packet.SERVERDATA_EXECCOMMAND, []byte("listen event"))
    client.Write(execPacket.Serialize())

    // Listen for responses
    for pkt := range packetChan {
        if pkt.Error != nil {
            if pkt.Error == io.EOF {
                fmt.Println("Connection closed")
                break
            }
            continue // Timeout or other non-fatal error
        }
        fmt.Printf("Received: %s\n", pkt.BodyStr())
    }
}
```

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See [LICENSE](LICENSE) for details.


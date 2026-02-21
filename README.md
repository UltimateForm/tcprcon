# tcprcon

- [tcprcon](#tcprcon)
  - [Using as a Library](#using-as-a-library)
    - [Streaming Responses](#streaming-responses)
  - [tcprcon-cli](#tcprcon-cli)
  - [Caveats](#caveats)
    - [Handling Server Broadcasts](#handling-server-broadcasts)
    - [Server Protocol Compliance](#server-protocol-compliance)
  - [License](#license)



A fully native RCON client implementation, zero deps

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

## tcprcon-cli

https://github.com/UltimateForm/tcprcon-cli


## Caveats

### Handling Server Broadcasts

Servers can (and will) often broadcast events over the TCP connection in an asynchronous manner. These are typically game events like killfeed messages, player logins, chat, etc. Some servers operate on an opt-in basis, requiring the RCON client to signal its interest in receiving these broadcasts, while others broadcast them by default.

What this means in practice:

Let's say you send a command packet (e.g., "status" with ID 54) and then immediately try to read its response. It's possible you might first receive a broadcast packet with a body like "Login: player B just joined game" instead of your expected status response. This highlights the importance of checking the `ID` field of incoming packets.

Generally, the best practice is to decouple your command writes from your response reads. The example under [Using as a Library](#using-as-a-library) demonstrates a synchronous request-response pattern for a `playerlist` command, which can be unoptimal in such scenarios. For a more robust approach, you should handle your writes (commands) and reads (responses and broadcasts) in parallel, as shown in the [Streaming Responses](#streaming-responses) section.

### Server Protocol Compliance

Ideally, all RCON servers would consistently follow the Valve protocol defined at https://developer.valvesoftware.com/wiki/Source_RCON_Protocol, eliminating surprises. However, in reality, some server implementations—such as that of Rust—exhibit unorthodox behavior.

The Rust game server commits the following notable violations of the RCON protocol:

- **Initial Logging Packet (ID 0, Type 4):** After a client sends a `SERVERDATA_EXECCOMMAND` (e.g., `info`), the server typically responds with an immediate `SERVERDATA_RESPONSE_VALUE` packet that has an `ID` of `0` and often a `Type` of `4` (which is not a standard RCON packet type). The `Body` of this packet usually contains a server-side log message echoing the received command (e.g., `[RCON][<client_ip>:<client_port>] <command>`). The `ID 0` is non-compliant, as the server should echo the client's original `ID`.
- **Repeated Command Output:** The actual command output (e.g., `hostname: LinuxGSM...`) is often sent twice: once with the correct echoed client `ID`, and again with an `ID` of `0`. This is redundant and non-compliant.
- **Misuse of `ID -1`:** The server uses `ID -1` (`0xFF FF FF FF`) as a general "end of response stream" or broadcast indicator following command output. According to the Source RCON Protocol, `ID -1` is specifically reserved to indicate an **authentication failure** within a `SERVERDATA_AUTH_RESPONSE` packet. Its use in the context of command responses is a significant deviation.

These are the most prominent violations; other quirks might exist with greater room for nuanced interpretation, which are not listed here.

The concluding point is that you should anticipate such cases. In general, this library will function—even with servers like Rust—because it provides the fundamental tools for writing and reading data according to Valve's protocol over a TCP socket. However, depending on these aforementioned server-specific behaviors, you might need to adapt how and when you send commands and process responses in your application.

Specifically for Rust servers, you might implement simple checks to filter out extraneous packets. For example, you could ignore all `SERVERDATA_RESPONSE_VALUE` packets with `ID -1` (after successful authentication) or `ID 0`, or filter out any packet with a `Type` value greater than `3` (as types `0-3` cover standard RCON messages). This allows your application to focus on the actual command responses while gracefully discarding server-initiated noise.

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See [LICENSE](LICENSE) for details.

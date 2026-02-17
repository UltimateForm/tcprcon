# tcprcon

- [tcprcon](#tcprcon)
  - [Using as a Library](#using-as-a-library)
    - [Streaming Responses](#streaming-responses)
  - [tcprcon-cli](#tcprcon-cli)
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

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See [LICENSE](LICENSE) for details.

package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/UltimateForm/tcprcon/internal/packet"
)

func main() {
	address := os.Getenv("rcon_address")
	port := os.Getenv("rcon_port")
	password := os.Getenv("rcon_password")
	log.Printf("Dialing %v at port %v\n", address, port)
	writerCon, err := net.Dial("tcp", address+":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer writerCon.Close()
	counter := int32(0)
	log.Println("Building auth packet")
	authPacket := packet.BuildPacket(counter, packet.SERVERDATA_AUTH, []byte(password))
	writer := bufio.NewWriter(writerCon)
	reader := bufio.NewReader(writerCon)
	log.Println("Sending auth packet")
	written, err := writer.Write(authPacket)
	log.Printf("Written %v bytes of auth packet to connection", written)
	for {
		log.Println("Flushing writer...")
		writer.Flush()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Reading from server...")
		dword := make([]byte, 4)
		readLengthDword, err := reader.Read(dword)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Read %v bytes from server response head", readLengthDword)
		packetSize := binary.LittleEndian.Uint32(dword)
		log.Printf("Response packet size %v", packetSize)
		packetBytes := make([]byte, packetSize)
		readLength, err := reader.Read(packetBytes)
		if err != nil {
			log.Fatal(err)
		}
		counter++
		log.Printf("Read %v bytes from server response", readLength)
		id := int32(binary.LittleEndian.Uint32(packetBytes[0:4]))
		packetType := int32(binary.LittleEndian.Uint32(packetBytes[4:8]))
		body := string(packetBytes[8:])
		log.Printf("Server response: Id: %v, type: %v, body: %v\n", id, packetType, body)
		log.Println("-----STARTING NEW EXCHANGE-----")
		stdinread := bufio.NewReader(os.Stdin)
		fmt.Print(">")
		cmd, _, err := stdinread.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		execPacket := packet.BuildPacket(counter, packet.SERVERDATA_EXECCOMMAND, cmd)
		writer.Write(execPacket)
	}

}

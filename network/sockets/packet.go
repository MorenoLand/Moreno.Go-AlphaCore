package sockets

import (
	"io"
	"net"

	"Moreno.AlphaCore/network/packet"
)

func ReadPacket(connection net.Conn) (packet.Packet, error) {
	header := make([]byte, packet.HeaderSize)
	if _, err := io.ReadFull(connection, header); err != nil {
		return packet.Packet{}, err
	}
	size, opcode, err := packet.ParseHeader(header)
	if err != nil || size > packet.MaxPacketSize {
		return packet.Packet{}, packet.ErrInvalidPacketSize
	}
	data := make([]byte, int(size))
	if _, err := io.ReadFull(connection, data); err != nil {
		return packet.Packet{}, err
	}
	return packet.Packet{Size: size, Opcode: opcode, Data: data}, nil
}

func WriteAll(connection net.Conn, data []byte) error {
	for len(data) > 0 {
		written, err := connection.Write(data)
		if err != nil {
			return err
		}
		data = data[written:]
	}
	return nil
}

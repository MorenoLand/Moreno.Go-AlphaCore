package packet

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	HeaderSize    = 6
	MaxPacketSize = 0x8000
)

var (
	ErrShortPacket       = errors.New("packet is shorter than its header")
	ErrInvalidPacketSize = errors.New("packet size is invalid")
	ErrPacketTooLarge    = errors.New("packet is too large")
)

type Packet struct {
	Size   uint16
	Opcode Opcode
	Data   []byte
}

func ParseHeader(header []byte) (uint16, Opcode, error) {
	if len(header) != HeaderSize {
		return 0, 0, ErrShortPacket
	}
	declared := binary.BigEndian.Uint16(header[:2])
	if declared < 4 {
		return 0, 0, ErrInvalidPacketSize
	}
	return declared - 4, Opcode(binary.LittleEndian.Uint32(header[2:])), nil
}

func Parse(data []byte) (Packet, error) {
	if len(data) < HeaderSize {
		return Packet{}, ErrShortPacket
	}
	size, opcode, err := ParseHeader(data[:HeaderSize])
	if err != nil {
		return Packet{}, err
	}
	if int(size) != len(data)-HeaderSize {
		return Packet{}, ErrInvalidPacketSize
	}
	return Packet{Size: size, Opcode: opcode, Data: append([]byte(nil), data[HeaderSize:]...)}, nil
}

func Encode(opcode Opcode, data []byte) ([]byte, error) {
	if len(data)+4 > 0xffff {
		return nil, ErrPacketTooLarge
	}
	body := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint32(body, uint32(opcode))
	copy(body[4:], data)
	packet := make([]byte, HeaderSize+len(data))
	binary.BigEndian.PutUint16(packet, uint16(len(body)))
	copy(packet[2:], body)
	if opcode != SMSGUpdateObject || len(packet) <= 100 {
		return packet, nil
	}
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("compress update packet: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("finish update packet compression: %w", err)
	}
	compressedData := make([]byte, 4+compressed.Len())
	binary.LittleEndian.PutUint32(compressedData, uint32(len(data)))
	copy(compressedData[4:], compressed.Bytes())
	return Encode(SMSGCompressedUpdateObject, compressedData)
}

func EncodeSRP6(data []byte) ([]byte, error) {
	if len(data) > 0xffff {
		return nil, ErrPacketTooLarge
	}
	packet := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(packet, uint16(len(data)))
	copy(packet[2:], data)
	return packet, nil
}

func StringBytes(value string) ([]byte, error) {
	result := make([]byte, 0, len(value)+1)
	for _, character := range value {
		if character > 0xff {
			return []byte{0}, fmt.Errorf("string contains a non-Latin-1 character")
		}
		result = append(result, byte(character))
	}
	return append(result, 0), nil
}

func ReadString(data []byte, start int, terminator byte) (string, error) {
	if start < 0 || start > len(data) {
		return "", io.ErrUnexpectedEOF
	}
	index := bytes.IndexByte(data[start:], terminator)
	if index < 0 {
		return "", io.ErrUnexpectedEOF
	}
	return string(data[start : start+index]), nil
}

func ReadStringFrom(reader io.ByteReader, terminator byte) (string, error) {
	var result bytes.Buffer
	for {
		value, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		if value == terminator {
			return result.String(), nil
		}
		result.WriteByte(value)
	}
}

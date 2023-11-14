package quics

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/quic-go/quic-go"
	"google.golang.org/protobuf/proto"
)

func RecvMsg(stream quic.Stream, msg proto.Message) (err error) {
	msgLen := make([]byte, 4)
	n, err := stream.Read(msgLen)
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("recv invalid msgLen data")
	}
	ilen := int(binary.LittleEndian.Uint32(msgLen))
	data := make([]byte, ilen)
	n, err = stream.Read(data)
	if err != nil {
		return err
	}
	if n != ilen {
		return fmt.Errorf("recv invalid msg length")
	}
	err = proto.Unmarshal(data, msg)
	if err != nil {
		return fmt.Errorf("unmarshal msg err: %v", err)
	}
	return err
}

func SendMsg(stream quic.Stream, msg proto.Message) (err error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("mashal msg err: %v", err)
	}
	msgLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(msgLen, uint32(len(data)))
	if _, err = stream.Write(msgLen); err != nil {
		return err
	}
	_, err = stream.Write(data)
	return err
}

type StreamHeader [1]byte

const (
	StreamHeaderFlag_StreamType byte = 0b11
)

type StreamType byte

const (
	StreamType_HTTP StreamType = 0b01
	StreamType_TCP  StreamType = 0b10
)

func (h StreamHeader) StreamType() StreamType {
	return StreamType(h[0] & StreamHeaderFlag_StreamType)
}

func (h *StreamHeader) SetStreamType(t StreamType) {
	(*h)[0] = (*h)[0] & ^StreamHeaderFlag_StreamType
	(*h)[0] |= byte(t)
}

func WriteHeader(stream quic.SendStream, h StreamHeader) (err error) {
	// fmt.Printf("header: %x\n", h)
	n, err := stream.Write(h[:])
	if err != nil {
		return
	}
	if n != len(h) {
		return errors.New("write header: invalid length")
	}
	return
}

func ReadHeader(stream quic.ReceiveStream) (h StreamHeader, err error) {
	n, err := stream.Read(h[:])
	if err != nil {
		return
	}
	if n != len(h) {
		return h, errors.New("read header: invalid length")
	}
	return
}

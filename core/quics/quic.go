package quics

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/mengseeker/nlink/core/log"
	"github.com/quic-go/quic-go"
	"google.golang.org/protobuf/proto"
)

var (
	l = log.With("Unit", "quics")
)

func SendMsg(stream quic.Stream, msg proto.Message) (err error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("mashal msg err: %v", err)
	}
	l.Debugf("SendMsg, len: %d", len(data))
	return SendFrame(stream, data)
}

func RecvMsg(stream quic.Stream, msg proto.Message) (err error) {
	frame, err := RecvFrame(stream)
	if err != nil {
		return
	}
	err = proto.Unmarshal(frame, msg)
	if err != nil {
		return fmt.Errorf("unmarshal msg err: %v", err)
	}
	l.Debugf("RecvMsg, len: %d", len(frame))
	return err
}

func SendFrame(stream quic.Stream, frame []byte) (err error) {
	msgLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(msgLen, uint32(len(frame)))
	if _, err = stream.Write(msgLen); err != nil {
		return err
	}
	n, err := stream.Write(frame)
	if n != len(frame) {
		panic("sendFram invalid length")
	}
	return err
}

func RecvFrame(stream quic.Stream) (frame []byte, err error) {
	msgLen := make([]byte, 4)
	n, err := stream.Read(msgLen)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, fmt.Errorf("recv invalid frameLen data")
	}
	ilen := int(binary.LittleEndian.Uint32(msgLen))
	frame = make([]byte, ilen)

	for start := 0; start < ilen; {
		n, err = stream.Read(frame[start:])
		start += n
		if start == ilen {
			return frame, nil
		}
		if err != nil {
			return
		}
	}
	return
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
	l.Debugf("WriteHeader: %x", h)
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
	l.Debugf("ReadHeader: %x", h)
	return
}

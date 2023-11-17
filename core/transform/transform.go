package transform

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

func SendMsg(w io.Writer, msg proto.Message) (err error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("mashal msg err: %v", err)
	}
	return SendFrame(w, data)
}

func RecvMsg(r io.Reader, msg proto.Message) (err error) {
	frame, err := RecvFrame(r)
	if err != nil {
		return
	}
	err = proto.Unmarshal(frame, msg)
	if err != nil {
		return fmt.Errorf("unmarshal msg err: %v", err)
	}
	return err
}

func SendFrame(w io.Writer, frame []byte) (err error) {
	msgLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(msgLen, uint32(len(frame)))
	if _, err = w.Write(msgLen); err != nil {
		return err
	}
	n, err := w.Write(frame)
	if n != len(frame) {
		panic("sendFram invalid length")
	}
	return err
}

func RecvFrame(r io.Reader) (frame []byte, err error) {
	msgLen := make([]byte, 4)
	n, err := io.ReadFull(r, msgLen)
	if err != nil {
		return nil, err
	}
	if n != 4 {
		return nil, fmt.Errorf("recv invalid frameLen")
	}

	ilen := int(binary.LittleEndian.Uint32(msgLen))
	frame = make([]byte, ilen)
	n, err = io.ReadFull(r, frame)
	if n != ilen {
		return nil, fmt.Errorf("recv invalid frame data")
	}
	return
}

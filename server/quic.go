package server

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
	"github.com/quic-go/quic-go"
	"google.golang.org/protobuf/proto"
)

type QuicServer struct {
	Handler Handler
	Config  *ServerConfig
	Log     *log.Logger
}

func NewQuicServer(cfg ServerConfig, handler Handler, log *log.Logger) (*QuicServer, error) {
	if cfg.Addr == "" {
		cfg.Addr = "0.0.0.0:8899"
	}
	if cfg.ReadBufferSize <= 0 {
		cfg.ReadBufferSize = 4 << 10
	}

	s := QuicServer{
		Config:  &cfg,
		Handler: handler,
		Log:     log.With("Unit", "QuicServer"),
	}
	return &s, nil
}

func (s *QuicServer) Start(c context.Context) (err error) {
	tls, err := NewServerTlsConfig(s.Config.TLS_Cert, s.Config.TLS_Key, s.Config.TLS_CA)
	if err != nil {
		return
	}
	lis, err := quic.ListenAddr(s.Config.Addr, tls, &quic.Config{})
	if err != nil {
		return
	}

	// handle
	for {
		conn, err := lis.Accept(c)
		if err != nil {
			if errors.Is(err, quic.ErrServerClosed) {
				return nil
			}
			return err
		}
		go s.handleClient(c, conn)
	}
}

func (s *QuicServer) handleClient(c context.Context, conn quic.Connection) {
	for {
		stream, err := conn.AcceptStream(c)
		if err != nil {
			s.Log.Errorf("AcceptStream err: %v", err)
			return
		}
		go s.handleStream(c, stream)
	}
}

const (
	QuicHeaderFlag_StreamType     = 0b00000001
	QuicHeaderFlag_StreamType_Tcp = 0b00000001
)

func (s *QuicServer) handleStream(c context.Context, stream quic.Stream) {
	defer stream.Close()
	header := make([]byte, 1)
	_, err := stream.Read(header)
	if err != nil {
		s.Log.Errorf("read stream header err: %v", err)
		return
	}
	if header[0]&QuicHeaderFlag_StreamType == QuicHeaderFlag_StreamType_Tcp {
		if err = s.Handler.TCPCall(&QuicTCPCallStream{stream: stream}); err != nil {
			s.Log.Errorf("handle http stream err: %v", err)
		}
	} else {
		if err = s.Handler.HTTPCall(&QuicHTTPCallStream{stream: stream}); err != nil {
			s.Log.Errorf("handle http stream err: %v", err)
		}
	}
}

type QuicHTTPCallStream struct {
	stream quic.Stream
}

func (c *QuicHTTPCallStream) Context() context.Context {
	return c.stream.Context()
}

func (c *QuicHTTPCallStream) Send(data *api.HTTPResponse) error {
	return QuicSendMsg(c.stream, data)
}

func (c *QuicHTTPCallStream) Recv() (*api.HTTPRequest, error) {
	var req api.HTTPRequest
	return &req, QuicRecvMsg(c.stream, &req)
}

type QuicTCPCallStream struct {
	stream quic.Stream
}

func (c *QuicTCPCallStream) Context() context.Context {
	return c.stream.Context()
}

func (c *QuicTCPCallStream) Send(data *api.SockData) error {
	return QuicSendMsg(c.stream, data)
}

func (c *QuicTCPCallStream) Recv() (*api.SockRequest, error) {
	var req api.SockRequest
	return &req, QuicRecvMsg(c.stream, &req)
}

func QuicRecvMsg(stream quic.Stream, msg proto.Message) (err error) {
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

func QuicSendMsg(stream quic.Stream, msg proto.Message) (err error) {
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

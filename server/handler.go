package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
	"github.com/mengseeker/nlink/core/log"
)

type Proxy_HTTPCallServer interface {
	Context() context.Context
	Send(*api.HTTPResponse) error
	Recv() (*api.HTTPRequest, error)
}

type Proxy_TCPCallServer interface {
	Context() context.Context
	Send(*api.SockData) error
	Recv() (*api.SockRequest, error)
}

type Handler struct {
	ReadBufferSize int
	Log            *log.Logger
	HTTPClient     *http.Client
}

func (s *Handler) HTTPCall(stream Proxy_HTTPCallServer) (err error) {
	r, err := stream.Recv()
	if err != nil {
		return err
	}

	req := r.GetRequest()
	proxyReq, err := http.NewRequestWithContext(stream.Context(), req.Method, req.Url, nil)
	for i := range req.Headers {
		proxyReq.Header.Add(req.Headers[i].Key, req.Headers[i].Value)
	}

	if r.Request.HasBody {
		ir, iw := io.Pipe()
		proxyReq.Body = ir
		go s.handleHTTPRequestBody(stream, r.Body, iw)
	}

	resp, err := s.HTTPClient.Do(proxyReq)
	if err != nil {
		s.Log.Error("http call err", "error", err)
		return err
	}
	defer resp.Body.Close()

	proxyResp := api.HTTPResponse_Response{
		Code:          int32(resp.StatusCode),
		ContentLength: resp.ContentLength,
	}

	for k, v := range resp.Header {
		for i := range v {
			proxyResp.Headers = append(proxyResp.Headers, &api.Header{Key: k, Value: v[i]})
		}
	}

	bf := make([]byte, s.ReadBufferSize)
	n, err := resp.Body.Read(bf)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			s.Log.Error("read response body err", "error", err)
			return
		}
		// err = nil
	}
	err = stream.Send(&api.HTTPResponse{
		Response: &proxyResp,
		Body:     bf[:n],
	})
	if err != nil {
		s.Log.Error("write response err", "error", err)
		return
	}

	for {
		n, err = resp.Body.Read(bf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			} else {
				s.Log.Error("read response body err", "error", err)
				return
			}
		}
		err = stream.Send(&api.HTTPResponse{
			Body: bf[:n],
		})
		if err != nil {
			s.Log.Error("write response err", "error", err)
			return
		}
	}
	resp.Body.Close()
	return nil
}

func (s *Handler) handleHTTPRequestBody(stream Proxy_HTTPCallServer, d1 []byte, w io.WriteCloser) {
	defer w.Close()
	_, err := w.Write(d1)
	if err != nil {
		s.Log.Error("write request data err", "error", err)
		return
	}

	for {
		br, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			s.Log.Error("recv request stream data err", "error", err)
			return
		}
		_, err = w.Write(br.GetBody())
		if err != nil {
			s.Log.Error("write request stream data err", "error", err)
			return
		}
	}
}

func (s *Handler) TCPCall(stream Proxy_TCPCallServer) (err error) {
	r, err := stream.Recv()
	if err != nil {
		return err
	}
	remoteConn, err := net.Dial("tcp", r.Req.Host)
	if err != nil {
		return fmt.Errorf("dial remote err %v", err)
	}
	defer remoteConn.Close()
	// try notify client
	err = stream.Send(&api.SockData{})
	if err != nil {
		return fmt.Errorf("notify client err %v", err)
	}

	// read remote data and write back
	done := make(chan int)
	go func() {
		defer close(done)
		remoteReadBuff := make([]byte, s.ReadBufferSize)
		for {
			n, err := remoteConn.Read(remoteReadBuff)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				s.Log.Error("read remote err", "error", err)
				return
			}
			back := api.SockData{
				Data: remoteReadBuff[:n],
			}
			err = stream.Send(&back)
			if err != nil {
				s.Log.Error("write back err", "error", err)
				return
			}
		}
	}()

	// read data and write to remote
	if r.Data != nil {
		_, err = remoteConn.Write(r.Data.Data)
		if err != nil {
			s.Log.Error("write remote err", "error", err)
			return
		}
	}
	for {
		r, err = stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			}
			return err
		}
		if err != nil {
			s.Log.Error("read req err", "error", err)
			return err
		}
		if r.Data != nil {
			_, err = remoteConn.Write(r.Data.Data)
			if err != nil {
				s.Log.Error("write remote err", "error", err)
				return err
			}
		}
	}

	<-done
	return
}

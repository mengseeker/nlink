package server

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
)

func (s *Server) HTTPCall(stream api.Proxy_HTTPCallServer) (err error) {
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

	resp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		slog.Error("http call err", "error", err)
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

	bf := make([]byte, s.Config.ReadBufferSize)
	n, err := resp.Body.Read(bf)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			slog.Error("read response body err", "error", err)
			return
		}
		// err = nil
	}
	err = stream.Send(&api.HTTPResponse{
		Response: &proxyResp,
		Body:     bf[:n],
	})
	if err != nil {
		slog.Error("write response err", "error", err)
		return
	}

	for {
		n, err = resp.Body.Read(bf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			} else {
				slog.Error("read response body err", "error", err)
				return
			}
		}
		err = stream.Send(&api.HTTPResponse{
			Body: bf[:n],
		})
		if err != nil {
			slog.Error("write response err", "error", err)
			return
		}
	}
	resp.Body.Close()
	return nil
}

func (s *Server) handleHTTPRequestBody(stream api.Proxy_HTTPCallServer, d1 []byte, w io.WriteCloser) {
	defer w.Close()
	_, err := w.Write(d1)
	if err != nil {
		slog.Error("write request data err", "error", err)
		return
	}

	for {
		br, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			slog.Error("recv request stream data err", "error", err)
			return
		}
		_, err = w.Write(br.GetBody())
		if err != nil {
			slog.Error("write request stream data err", "error", err)
			return
		}
	}
}

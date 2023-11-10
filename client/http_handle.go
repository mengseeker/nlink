package client

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
	"gopkg.in/elazarl/goproxy.v1"
)

func DirectReqHandle(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	deleteRequestHeaders(req)
	resp, err := ctx.RoundTrip(req)
	if err != nil {
		if resp == nil {
			ctx.Logf("error read response %v %v:", req.URL.Host, err.Error())
			resp = goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusBadGateway,
				err.Error())
		}
	}
	return req, resp
}

func DirectHandleConnect(req string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return goproxy.OkConnect, ctx.Req.URL.Host
}

func RejectReq(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	return req, goproxy.NewResponse(req,
		goproxy.ContentTypeText, http.StatusForbidden, http.StatusText(http.StatusForbidden))
}

func newForwardHTTPHandle(pv *FunctionProvider, name string) (handle goproxy.FuncReqHandler) {
	handle = func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		cli, err := pv.DialHTTP(req.Context(), name)
		if err != nil {
			slog.Error("dial forward err", "name", name, "error", err)
			return req, goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusBadGateway, "proxy dial forward err")
		}
		// defer cli.CloseSend()
		deleteRequestHeaders(req)
		resp := &http.Response{}
		resp.Request = req
		resp.TransferEncoding = req.TransferEncoding
		resp.Header = make(http.Header)

		remoteReq := api.HTTPRequest{
			Request: &api.HTTPRequest_Request{
				Method:  req.Method,
				Url:     req.URL.String(),
				Headers: make([]*api.Header, 0, len(req.Header)),
			},
		}
		for k, hs := range req.Header {
			for i := range hs {
				remoteReq.Request.Headers = append(remoteReq.Request.Headers, &api.Header{Key: k, Value: hs[i]})
			}
		}
		readBuffer := make([]byte, pv.ReadBufferSize)
		hasBody := true
		n, err := req.Body.Read(readBuffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				hasBody = false
			} else {
				cli.CloseSend()
				return req, goproxy.NewResponse(req,
					goproxy.ContentTypeText, http.StatusBadGateway, fmt.Sprintf("proxy read request err: %v", err))
			}
		}
		remoteReq.Body = readBuffer[:n]

		// dial
		err = cli.Send(&remoteReq)
		if err != nil {
			cli.CloseSend()
			return req, goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusBadGateway, fmt.Sprintf("proxy dial err: %v", err))
		}

		// read response
		remoteResp, err := cli.Recv()
		if err != nil {
			cli.CloseSend()
			return req, goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusBadGateway, fmt.Sprintf("proxy recv err: %v", err))
		}

		// reply, shoule send header
		resp.StatusCode = int(remoteResp.Response.Code)
		resp.ContentLength = remoteResp.Response.ContentLength
		for _, h := range remoteResp.Response.Headers {
			resp.Header.Add(h.Key, h.Value)
		}
		br, bw := io.Pipe()
		resp.Body = br

		// handle body, copy request body to remote
		if hasBody {
			go func() {
				defer cli.CloseSend()
				defer req.Body.Close()
				handleReadBuff := make([]byte, pv.ReadBufferSize)
				for {
					n, err := req.Body.Read(handleReadBuff)
					if err != nil {
						if errors.Is(err, io.EOF) {
							return
						} else {
							slog.Error("handle read body data err", "error", err)
							return
						}
					}
					data := api.HTTPRequest{
						Body: readBuffer[:n],
					}
					err = cli.Send(&data)
					if err != nil {
						slog.Error("handle send data err", "error", err)
						return
					}
				}
			}()
		} else {
			cli.CloseSend()
			slog.Debug("client CloseSend")
			req.Body.Close()
		}
		// copy remote data back
		go func() {
			defer bw.Close()
			for {
				remoteResp, err := cli.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					} else {
						slog.Error("handle read remote err", "error", err)
						return
					}
				}
				if remoteResp.Body != nil {
					_, err = bw.Write(remoteResp.Body)
					if err != nil {
						slog.Error("handle write back err", "error", err)
						return
					}
				}
			}
		}()
		return req, resp
	}
	return
}

func newForwardHTTPSHandle(pv *FunctionProvider, name string) (handle goproxy.FuncHttpsHandler) {
	handle = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		cli, err := pv.DialTCP(ctx.Req.Context(), name)
		if err != nil {
			slog.Error("dial forward tcp err", "name", name, "error", err)
			return &goproxy.ConnectAction{
				Action: goproxy.ConnectReject,
			}, host
		}
		// dial
		_, err = cli.Recv() // read notify
		if err != nil {
			slog.Error("read remote err", "name", name, "error", err)
			return &goproxy.ConnectAction{
				Action: goproxy.ConnectReject,
			}, host
		}

		return &goproxy.ConnectAction{
			Action: goproxy.ConnectHijack,
			Hijack: func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
				defer client.Close()
				defer req.Body.Close()
				defer cli.CloseSend()
				// handle body, copy request body to remote
				go func() {
					readBuffer := make([]byte, pv.ReadBufferSize)
					for {
						n, err := client.Read(readBuffer)
						if err != nil {
							if errors.Is(err, io.EOF) {
								return
							} else {
								slog.Error("handle read body data err", "error", err)
								return
							}
						}
						data := api.SockRequest{
							Data: &api.SockData{
								Data: readBuffer[:n],
							},
						}
						err = cli.Send(&data)
						if err != nil {
							slog.Error("handle send data err", "error", err)
							return
						}
					}
				}()
				// copy remote data back
				for {
					resp, err := cli.Recv()
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						} else {
							slog.Error("handle read remote err", "error", err)
							return
						}
					}
					_, err = client.Write(resp.Data)
					if err != nil {
						slog.Error("handle write back err", "error", err)
						return
					}
				}
			},
		}, host
	}
	return
}

func deleteRequestHeaders(req *http.Request) {
	req.RequestURI = "" // this must be reset when serving a request with the client
	// req.Header.Del("Accept-Encoding")
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Connection")
}

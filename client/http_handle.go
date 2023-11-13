package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
	"gopkg.in/elazarl/goproxy.v1"
)

var (
	ProxyHeaders = map[string]bool{
		"Proxy-Connection":    true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Connection":          true,
	}
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

func RejectReq(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	return req, goproxy.NewResponse(req,
		goproxy.ContentTypeText, http.StatusForbidden, http.StatusText(http.StatusForbidden))
}

func NewForwardReqFromHTTP(req *http.Request) (remote *api.HTTPRequest_Request) {
	remoteReq := api.HTTPRequest_Request{
		Method:  req.Method,
		Url:     req.URL.String(),
		Headers: make([]*api.Header, 0, len(req.Header)),
	}
	for k, hs := range req.Header {
		for i := range hs {
			if !ProxyHeaders[k] {
				remoteReq.Headers = append(remoteReq.Headers, &api.Header{Key: k, Value: hs[i]})
			}
		}
	}
	return &remoteReq
}

func NewHTTPResponse(req *http.Request) (resp *http.Response) {
	resp = &http.Response{}
	resp.Request = req
	resp.TransferEncoding = req.TransferEncoding
	resp.Header = make(http.Header)
	return
}

func deleteRequestHeaders(req *http.Request) {
	req.RequestURI = "" // this must be reset when serving a request with the client
	// req.Header.Del("Accept-Encoding")
	for k := range ProxyHeaders {
		req.Header.Del(k)
	}
}

func newForwardHTTPHandle(pv *FunctionProvider, name string) (handle goproxy.FuncReqHandler) {
	handle = func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		cli, err := pv.DialHTTP(req.Context(), name)
		if err != nil {
			ctx.Warnf("[%s] dial forward err: %v", name, err)
			return req, goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusBadGateway, "proxy dial forward err")
		}

		// copy request body
		go handleForwardHTTPRequest(pv, ctx, req, cli)

		// wait remote response
		remoteResp, err := cli.Recv()
		if err != nil {
			cli.CloseSend()
			return req, goproxy.NewResponse(req,
				goproxy.ContentTypeText, http.StatusBadGateway, fmt.Sprintf("proxy recv err: %v", err))
		}

		// copy response
		respr, respw := io.Pipe()
		resp := NewHTTPResponse(req)
		// reply, shoule send header
		resp.StatusCode = int(remoteResp.Response.Code)
		if remoteResp.Response.ContentLength > 0 {
			resp.ContentLength = remoteResp.Response.ContentLength
		}
		for _, h := range remoteResp.Response.Headers {
			resp.Header.Add(h.Key, h.Value)
		}
		resp.Body = respr

		// copy remote data back
		go func() {
			defer respw.Close()
			_, err = respw.Write(remoteResp.Body)
			if err != nil {
				ctx.Warnf("handle write back err: %v", err)
				return
			}

			for {
				remoteResp, err := cli.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					} else {
						ctx.Warnf("handle read remote err: %v", err)
						return
					}
				}
				if remoteResp.Body != nil {
					_, err = respw.Write(remoteResp.Body)
					if err != nil {
						ctx.Warnf("handle write back err: %v", err)
						return
					}
				}
			}
		}()
		return req, resp
	}

	return
}

func handleForwardHTTPRequest(pv *FunctionProvider, ctx *goproxy.ProxyCtx, req *http.Request, cli Proxy_HTTPCallClient) {
	defer cli.CloseSend()
	defer req.Body.Close()
	readBuffer := make([]byte, pv.ReadBufferSize)

	n, err := req.Body.Read(readBuffer)
	if err != nil && !errors.Is(err, io.EOF) {
		ctx.Warnf("handle read body data err: %v", err)
		return
	}

	data := api.HTTPRequest{
		Request: NewForwardReqFromHTTP(req),
		Body:    readBuffer[:n],
	}
	if !errors.Is(err, io.EOF) {
		data.Request.HasBody = true
	}

	err = cli.Send(&data)
	if err != nil {
		ctx.Warnf("handle send data err: %v", err)
		return
	}

	for {
		n, err = req.Body.Read(readBuffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			} else {
				ctx.Warnf("handle read body data err: %v", err)
				return
			}
		}
		data := api.HTTPRequest{
			Body: readBuffer[:n],
		}
		err = cli.Send(&data)
		if err != nil {
			ctx.Warnf("handle send data err: %v", err)
			return
		}
	}
}

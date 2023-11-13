package client

import (
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/mengseeker/nlink/core/api"
	"gopkg.in/elazarl/goproxy.v1"
)

func DirectHandleConnect(req string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return goproxy.OkConnect, ctx.Req.URL.Host
}

func newForwardHTTPSHandle(pv *FunctionProvider, name string) (handle goproxy.FuncHttpsHandler) {
	handle = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		cli, err := pv.DialTCP(ctx.Req.Context(), name)
		if err != nil {
			ctx.Warnf("dial forward %s tcp err: %v", name, err)
			return &goproxy.ConnectAction{
				Action: goproxy.ConnectReject,
			}, host
		}
		// dial
		_, err = cli.Recv() // read notify
		if err != nil {
			ctx.Warnf("[%v] read remote err: %v", name, err)
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
								ctx.Warnf("handle read body data err: %v", err)
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
							ctx.Warnf("handle send data err: %v", err)
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
							ctx.Warnf("handle read remote err: %v", err)
							return
						}
					}
					_, err = client.Write(resp.Data)
					if err != nil {
						ctx.Warnf("handle write back err: %v", err)
						return
					}
				}
			},
		}, host
	}
	return
}

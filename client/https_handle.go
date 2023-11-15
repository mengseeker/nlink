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
	l.Info("direct connect", "host", req)
	return goproxy.OkConnect, ctx.Req.URL.Host
}

func RejectHandleConnect(req string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	l.Info("reject connect", "host", req)
	return goproxy.RejectConnect, ctx.Req.URL.Host
}

func newForwardHTTPSHandle(pv *FunctionProvider, name string) (handle goproxy.FuncHttpsHandler) {
	handle = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		hlog := l.With("forward", name)
		hlog.Info("forward connect", "host", host)
		cli, err := pv.DialTCP(ctx.Req.Context(), name)
		if err != nil {
			hlog.Errorf("dial forward tcp err: %v", err)
			return &goproxy.ConnectAction{
				Action: goproxy.ConnectReject,
			}, host
		}
		// dial
		dial := api.SockRequest{
			Req: &api.SockRequest_Sock{
				Host: host,
			},
		}
		err = cli.Send(&dial)
		if err != nil {
			hlog.Errorf("dial remote err: %v", err)
			return &goproxy.ConnectAction{
				Action: goproxy.ConnectReject,
			}, host
		}

		return &goproxy.ConnectAction{
			Action: goproxy.ConnectHijack,
			Hijack: func(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
				// fmt.Printf("now2: %d\n", time.Now().UnixMilli())
				defer client.Close()
				defer req.Body.Close()

				// handle body, copy request body to remote
				go func() {
					defer cli.CloseSend()
					readBuffer := make([]byte, pv.ReadBufferSize)
					// defer func() { fmt.Printf("now3: %d\n", time.Now().UnixMilli()) }()
					for {
						n, err := client.Read(readBuffer)
						if err != nil {
							if errors.Is(err, io.EOF) {
								return
							} else {
								hlog.Errorf("handle read body data err: %v", err)
								return
							}
						}
						if n == 0 {
							continue
						}
						data := api.SockRequest{
							Data: &api.SockData{
								Data: readBuffer[:n],
							},
						}
						err = cli.Send(&data)
						if err != nil {
							hlog.Errorf("handle send data err: %v", err)
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
							hlog.Errorf("handle read remote err: %v", err)
							return
						}
					}
					_, err = client.Write(resp.Data)
					if err != nil {
						hlog.Errorf("handle write back err: %v", err)
						return
					}
				}
				// fmt.Printf("now4: %d\n", time.Now().UnixMilli())
			},
		}, host
	}
	return
}

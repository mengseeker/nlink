package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mengseeker/nlink/core/log"
)

type WailsApp struct {
	ctx        context.Context
	pxy        *Proxy
	logIO      io.ReadCloser
	logScanner *bufio.Scanner
	logs       chan string
}

func (a *WailsApp) Startup(ctx context.Context) {
	a.ctx = ctx
	// handle logs
	ir, iw := io.Pipe()
	a.logIO = ir
	log.Out = iw
	a.logScanner = bufio.NewScanner(a.logIO)
	a.logs = make(chan string, 1000)
	go a.handleLogs()
}

func (a *WailsApp) Restart(configJson string) string {
	if a.pxy != nil {
		a.pxy.Stop()
	}
	var cfg ProxyConfig
	if err := json.Unmarshal([]byte(configJson), &cfg); err != nil {
		return fmt.Sprintf("invalid config, err: %s", err.Error())
	}
	pxy, err := NewProxy(a.ctx, cfg)
	if err != nil {
		return fmt.Sprintf("new proxy err: %s", err.Error())
	}
	go func() {
		if err = pxy.Start(); err != nil {
			a.logs <- err.Error()
			return
		}
	}()
	a.pxy = pxy
	return ""
}

func (a *WailsApp) Stop() {
	a.pxy.Stop()
}

func (a *WailsApp) Logs() []string {
	logs := make([]string, 0, 100)
	var log string
LOOP:
	for i := 0; i < 1000; i++ {
		select {
		case log = <-a.logs:
			logs = append(logs, log)
		default:
			break LOOP
		}
	}
	return logs
}

func (a *WailsApp) handleLogs() {
	go func() {
		<-a.ctx.Done()
		a.logIO.Close()
	}()
	for a.logScanner.Scan() {
		select {
		case a.logs <- a.logScanner.Text():
		default:
			fmt.Println("ignore log: ", a.logScanner.Text())
		}
	}
}

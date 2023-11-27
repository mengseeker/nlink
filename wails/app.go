package wails

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mengseeker/nlink/client"
	"github.com/mengseeker/nlink/core/log"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type WailsApp struct {
	ctx        context.Context
	pxy        *client.Proxy
	log        *log.Logger
	logIO      io.ReadCloser
	logScanner *bufio.Scanner
	logs       chan string
}

type WailsResult struct {
	Success bool
	Msg     string
	Result  any
}

func (a *WailsApp) Startup(ctx context.Context) {
	a.ctx = ctx
	// handle logs
	ir, iw := io.Pipe()
	a.log = log.NewLogger()
	a.logIO = ir
	log.Out = iw
	a.logScanner = bufio.NewScanner(a.logIO)
	a.logs = make(chan string, 1000)
	go a.handleLogs()
}

func (a *WailsApp) Restart(configJson string) WailsResult {
	if err := a.restart(configJson); err != nil {
		return WailsResult{
			Success: false,
			Msg:     err.Error(),
		}
	}
	return WailsResult{Success: true}
}

func (a *WailsApp) Stop() WailsResult {
	a.stop()
	return WailsResult{Success: true}
}

func (a *WailsApp) Logs() WailsResult {
	return WailsResult{Success: true, Result: a.getLogs()}
}

func (a *WailsApp) OpenDirectoryDialog(title string) WailsResult {
	res, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		ShowHiddenFiles: true,
		Title:           title,
	})
	if err != nil {
		return WailsResult{Success: false, Msg: err.Error()}
	}
	return WailsResult{Success: true, Result: res}
}

func (a *WailsApp) OpenFileDialog(title string) WailsResult {
	res, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		ShowHiddenFiles: true,
		Title:           title,
	})
	if err != nil {
		return WailsResult{Success: false, Msg: err.Error()}
	}
	return WailsResult{Success: true, Result: res}
}

func (a *WailsApp) restart(configJson string) error {
	if a.pxy != nil {
		a.pxy.Stop()
	}
	var cfg client.ProxyConfig
	if err := json.Unmarshal([]byte(configJson), &cfg); err != nil {
		return fmt.Errorf("invalid config, err: %s", err.Error())
	}
	pxy, err := client.NewProxy(a.ctx, cfg)
	if err != nil {
		return fmt.Errorf("new proxy err: %s", err.Error())
	}
	go func() {
		if err = pxy.Start(); err != nil {
			a.log.Errorf("start proxy err: %s", err.Error())
			return
		}
	}()
	a.pxy = pxy
	return nil
}

func (a *WailsApp) stop() {
	a.pxy.Stop()
}

func (a *WailsApp) getLogs() []string {
	logs := make([]string, 0, 100)
	var log string
	logs = append(logs, <-a.logs)
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
		}
		fmt.Println(a.logScanner.Text())
	}
}
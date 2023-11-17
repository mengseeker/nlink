package client

import "context"

type WailsApp struct {
	ctx context.Context
}

func (a *WailsApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *WailsApp) Restart(configJson string) (err string) {
	return "not impl"
}

func (a *WailsApp) Stop() {
}

func (a *WailsApp) Logs() []string {
	return nil
}

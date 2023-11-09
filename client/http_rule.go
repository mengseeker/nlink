package client

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/elazarl/goproxy.v1"
)

type RuleCond string

const (
	RuleCond_HostMatch RuleCond = "HOST"
	RuleCond_GEOIP     RuleCond = "GEOIP"
	RuleCond_IPCR      RuleCond = "IPCR"
	RuleCond_MATCH     RuleCond = "MATCH"
)

type RuleAction string

const (
	RuleAction_Reject  RuleAction = "REJECT"
	RuleAction_Direct  RuleAction = "DIRECT"
	RuleAction_Forward RuleAction = "FORWARD"
)

type ProxyRule struct {
	Cond        RuleCond
	CondParam   string
	Action      RuleAction
	ActionParam string
}

var (
	ErrInvalidSyntax = errors.New("invalid syntax")
)

func UnmashalProxyRule(raw string) (r ProxyRule, err error) {
	rs := strings.Split(raw, ",")
	if len(rs) != 2 {
		err = ErrInvalidSyntax
		return
	}
	condStr := strings.SplitN(rs[0], ":", 1)
	r.Cond = RuleCond(strings.TrimSpace(condStr[0]))
	if len(condStr) > 1 {
		r.CondParam = strings.TrimSpace(condStr[1])
	}

	actionStr := strings.SplitN(rs[1], ":", 1)
	r.Action = RuleAction(strings.TrimSpace(actionStr[0]))
	if len(actionStr) > 1 {
		r.ActionParam = strings.TrimSpace(actionStr[1])
	}
	return
}

func (r ProxyRule) BuildProxyConds(pv *FunctionProvider) (conds []goproxy.ReqCondition, err error) {
	switch r.Cond {
	case RuleCond_MATCH:
	case RuleCond_HostMatch:
		conds = append(conds, goproxy.ReqHostMatches(regexp.MustCompile(r.CondParam)))
	case RuleCond_GEOIP:
		conds = append(conds, newGEOIPCond(pv, r.CondParam))
	case RuleCond_IPCR:
		conds = append(conds, newIPCRCond(pv, r.CondParam))
	default:
		err = fmt.Errorf("unsuport rule cond: %q", r.Cond)
	}
	return
}

func newGEOIPCond(pv *FunctionProvider, country string) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		ip := pv.Resolver(req.URL.Host)
		return pv.GEOIP(ip) == Country(country)
	}
}

func newIPCRCond(pv *FunctionProvider, ipcr string) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		ip := pv.Resolver(req.URL.Host)
		return ipcrMatch(ipcr, ip)
	}
}

func (r ProxyRule) BuildProxyAction(pv *FunctionProvider) (reqHandle goproxy.FuncReqHandler, httpsHandle goproxy.FuncHttpsHandler, err error) {
	switch r.Action {
	case RuleAction_Direct:
		reqHandle = DirectReqHandle
		httpsHandle = DirectHandleConnect
	case RuleAction_Reject:
		reqHandle = RejectReq
		httpsHandle = goproxy.AlwaysReject
	case RuleAction_Forward:
		reqHandle, httpsHandle, err = newForwardHandle(pv, r.ActionParam)
	default:
		err = fmt.Errorf("unsuport rule action: %q", r.Action)
	}
	return
}

func ipcrMatch(ipcr, ip string) bool {
	// TODO
	return false
}

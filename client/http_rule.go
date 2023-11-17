package client

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/elazarl/goproxy.v1"
)

type RuleCond string

const (
	RuleCond_HostMatch  RuleCond = "host-match"
	RuleCond_HostPrefix RuleCond = "host-prefix"
	RuleCond_HostSuffix RuleCond = "host-suffix"
	RuleCond_HostRegexp RuleCond = "host-regexp"
	RuleCond_GEOIP      RuleCond = "geoip"
	RuleCond_IPCIDR     RuleCond = "ip-cidr"
	RuleCond_MatchAll   RuleCond = "match-all"
)

type RuleAction string

const (
	RuleAction_Reject  RuleAction = "reject"
	RuleAction_Direct  RuleAction = "direct"
	RuleAction_Forward RuleAction = "forward"
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
	condStr := strings.SplitN(rs[0], ":", 2)
	// fmt.Sprintln("condStr", condStr)
	r.Cond = RuleCond(strings.TrimSpace(strings.ToLower(condStr[0])))
	if len(condStr) > 1 {
		r.CondParam = strings.TrimSpace(condStr[1])
	}

	actionStr := strings.SplitN(rs[1], ":", 2)
	r.Action = RuleAction(strings.TrimSpace(strings.ToLower(actionStr[0])))
	if len(actionStr) > 1 {
		r.ActionParam = strings.TrimSpace(actionStr[1])
	}
	return
}

func (r ProxyRule) BuildProxyConds(pv *FunctionProvider) (conds []goproxy.ReqCondition, err error) {
	switch r.Cond {
	case RuleCond_MatchAll:
		conds = append(conds, goproxy.ReqConditionFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) bool { return true }))
	case RuleCond_HostPrefix:
		conds = append(conds, goproxy.ReqConditionFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
			return strings.HasPrefix(req.URL.Host, r.CondParam)
		}))
	case RuleCond_HostSuffix:
		conds = append(conds, goproxy.ReqConditionFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
			return strings.HasSuffix(strings.SplitN(req.URL.Host, ":", 2)[0], r.CondParam)
		}))
	case RuleCond_HostMatch:
		conds = append(conds, goproxy.ReqConditionFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
			return strings.Contains(req.URL.Host, r.CondParam)
		}))
	case RuleCond_HostRegexp:
		regexp, err := regexp.Compile(r.CondParam)
		if err != nil {
			return nil, err
		}
		conds = append(conds, goproxy.ReqHostMatches(regexp))
	case RuleCond_GEOIP:
		conds = append(conds, newGEOIPCond(pv, r.CondParam))
	case RuleCond_IPCIDR:
		conds = append(conds, newIPCIDRRCond(pv, r.CondParam))
	default:
		err = fmt.Errorf("unsuport rule cond: %q", r.Cond)
	}
	return
}

func newGEOIPCond(pv *FunctionProvider, country string) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		ip := pv.Resolv(ctx.Req.Context(), req.URL.Host)
		return pv.GEOIP(ip) == country
	}
}

func newIPCIDRRCond(pv *FunctionProvider, cidr string) goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		ip := pv.Resolv(ctx.Req.Context(), req.URL.Host)
		return ipcidrMatch(cidr, ip)
	}
}

func (r ProxyRule) BuildProxyAction(pv *FunctionProvider) (reqHandle goproxy.FuncReqHandler, httpsHandle goproxy.FuncHttpsHandler, err error) {
	switch r.Action {
	case RuleAction_Direct:
		reqHandle = DirectReqHandle
		httpsHandle = DirectHandleConnect
	case RuleAction_Reject:
		reqHandle = RejectReq
		httpsHandle = RejectHandleConnect
	case RuleAction_Forward:
		reqHandle, httpsHandle = newForwardHTTPHandle(pv, r.ActionParam), newForwardHTTPSHandle(pv, r.ActionParam)
	default:
		err = fmt.Errorf("unsuport rule action: %q", r.Action)
	}
	return
}

func ipcidrMatch(cidr string, ip net.IP) bool {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ipnet.Contains(ip)
}

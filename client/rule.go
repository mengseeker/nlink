package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/socks"
)

type RuleCondType string

const (
	RuleCondType_HostMatch  RuleCondType = "host-match"
	RuleCondType_HostPrefix RuleCondType = "host-prefix"
	RuleCondType_HostSuffix RuleCondType = "host-suffix"
	RuleCondType_HostRegexp RuleCondType = "host-regexp"
	RuleCondType_GEOIP      RuleCondType = "geoip"
	RuleCondType_IPCIDR     RuleCondType = "ip-cidr"
	RuleCondType_HasServer  RuleCondType = "has-server"
	RuleCondType_MatchAll   RuleCondType = "match-all"
)

type RuleActionType string

const (
	RuleActionType_Reject  RuleActionType = "reject"
	RuleActionType_Direct  RuleActionType = "direct"
	RuleActionType_Forward RuleActionType = "forward"
)

type RuleCond struct {
	Cond      RuleCondType
	CondParam string
}

type RuleAction struct {
	Action      RuleActionType
	ActionParam string
}

type Rule struct {
	Conds []RuleCond
	RuleAction
}

var (
	ErrInvalidSyntax = errors.New("invalid syntax")
)

func UnmashalProxyRule(raw string) (r Rule, err error) {
	rs := strings.Split(raw, ",")
	if len(rs) != 2 {
		err = ErrInvalidSyntax
		return
	}
	conds := strings.Split(rs[0], "&&")
	for _, v := range conds {
		cond := RuleCond{}
		condStr := strings.SplitN(v, ":", 2)
		cond.Cond = RuleCondType(strings.TrimSpace(strings.ToLower(condStr[0])))
		if len(condStr) > 1 {
			cond.CondParam = strings.TrimSpace(condStr[1])
		}
		r.Conds = append(r.Conds, cond)
	}

	actionStr := strings.SplitN(rs[1], ":", 2)
	r.Action = RuleActionType(strings.TrimSpace(strings.ToLower(actionStr[0])))
	if len(actionStr) > 1 {
		r.ActionParam = strings.TrimSpace(actionStr[1])
	}
	return r, nil
}

func (r Rule) Check(forwards map[string]Forward) error {
	for _, c := range r.Conds {
		switch c.Cond {
		case RuleCondType_HostMatch, RuleCondType_HostPrefix, RuleCondType_HostSuffix, RuleCondType_GEOIP, RuleCondType_HasServer:
			if c.CondParam == "" {
				return fmt.Errorf("rule condition %q params must not empty", c.Cond)
			}
		case RuleCondType_HostRegexp:
			_, err := regexp.Compile(c.CondParam)
			if err != nil {
				return fmt.Errorf("rule condition %q params must a valid regexp, compile err: %v", c.Cond, err)
			}
		case RuleCondType_IPCIDR:
			_, _, err := net.ParseCIDR(c.CondParam)
			if err != nil {
				return fmt.Errorf("rule condition %q params must a valid IPCIDR, parse err: %v", c.Cond, err)
			}
		case RuleCondType_MatchAll:
		default:
			return fmt.Errorf("unsupport rule condition %q", c.Cond)
		}
	}

	switch r.Action {
	case RuleActionType_Direct:
	case RuleActionType_Reject:
	case RuleActionType_Forward:
		if r.ActionParam == "" || forwards[r.ActionParam] == nil {
			return fmt.Errorf("rule action %s params must in forward servers", r.Action)
		}
	default:
		return fmt.Errorf("unsupport rule action %q", r.Action)
	}
	return nil
}

type MatchMeta struct {
	Schema string
	Host   string
	Port   string
}

func NewMatchMetaFromHTTPRequest(req *http.Request) MatchMeta {
	domain, port := ParseHost(req.Host)
	if port == "" {
		port = "80"
	}
	return MatchMeta{
		Schema: "http",
		Host:   domain,
		Port:   port,
	}
}

func NewMatchMetaFromHTTPSHost(host string) MatchMeta {
	domain, port := ParseHost(host)
	if port == "" {
		port = "443"
	}
	return MatchMeta{
		Schema: "https",
		Host:   domain,
		Port:   port,
	}
}

func NewMatchMetaFromSocksMeta(meta *socks.Metadata) MatchMeta {
	return MatchMeta{
		Schema: "tcp",
		Host:   meta.String(),
		Port:   meta.DstPort,
	}
}

func (rm *RuleMapper) Match(meta MatchMeta) RuleHandler {
	rm.lock.RLock()
	if h, exist := rm.cache[meta]; exist {
		rm.lock.RUnlock()
		return h
	}
	rm.lock.RUnlock()
	for i := range rm.matchs {
		if rm.matchs[i](meta) {
			h := rm.actions[i]
			rm.lock.Lock()
			rm.cache[meta] = h
			rm.lock.Unlock()
			return h
		}
	}
	return &DirectRuleHandler{log.NewLogger()}
}

func (r RuleCond) NewMatchFunc(pv *FuncProvider) func(mm MatchMeta) bool {
	switch r.Cond {
	case RuleCondType_HostMatch:
		return func(mm MatchMeta) bool {
			return strings.Contains(mm.Host, r.CondParam)
		}
	case RuleCondType_HostPrefix:
		return func(mm MatchMeta) bool {
			return strings.HasPrefix(mm.Host, r.CondParam)
		}
	case RuleCondType_HostSuffix:
		return func(mm MatchMeta) bool {
			return strings.HasSuffix(mm.Host, r.CondParam)
		}
	case RuleCondType_HostRegexp:
		reg := regexp.MustCompile(r.CondParam)
		return func(mm MatchMeta) bool {
			return reg.Match([]byte(mm.Host))
		}
	case RuleCondType_GEOIP:
		return func(mm MatchMeta) bool {
			return pv.GEOIP(pv.Resolv(mm.Host)) == string(r.CondParam)
		}
	case RuleCondType_IPCIDR:
		_, ipnet, _ := net.ParseCIDR(r.CondParam)
		return func(mm MatchMeta) bool {
			return ipnet.Contains(pv.Resolv(mm.Host))
		}
	case RuleCondType_HasServer:
		ok := pv.HasServer(r.CondParam)
		return func(mm MatchMeta) bool {
			return ok
		}
	case RuleCondType_MatchAll:
		return func(mm MatchMeta) bool { return true }
	default:
		return func(mm MatchMeta) bool { return false }
	}
}

func (r Rule) NewMatchFunc(pv *FuncProvider) func(mm MatchMeta) bool {
	funcs := []func(mm MatchMeta) bool{}
	for _, c := range r.Conds {
		funcs = append(funcs, c.NewMatchFunc(pv))
	}
	return func(mm MatchMeta) bool {
		for _, c := range funcs {
			if !c(mm) {
				return false
			}
		}
		return true
	}
}

func (r Rule) NewRuleHandler(pv *FuncProvider, forwards map[string]Forward) RuleHandler {
	switch r.Action {
	case RuleActionType_Reject:
		return &RejectRuleHandler{log: pv.log}
	case RuleActionType_Direct:
		return &DirectRuleHandler{log: pv.log}
	case RuleActionType_Forward:
		return forwards[r.ActionParam]
	default:
		return &RejectRuleHandler{log: pv.log}
	}
}

type RuleMapper struct {
	matchs  []func(MatchMeta) bool
	actions []RuleHandler
	cache   map[MatchMeta]RuleHandler
	lock    sync.RWMutex
}

func NewRuleMapper(
	ctx context.Context,
	rules []string,
	pv *FuncProvider,
	forwards map[string]Forward,
) (*RuleMapper, error) {

	mp := RuleMapper{
		cache: make(map[MatchMeta]RuleHandler),
		lock:  sync.RWMutex{},
	}
	var rls []Rule
	for _, rs := range rules {
		r, err := UnmashalProxyRule(rs)
		if err != nil {
			return nil, err
		}
		if err = r.Check(forwards); err != nil {
			return nil, err
		}
		rls = append(rls, r)
	}
	if len(rls) == 0 {
		rls = append(rls, Rule{
			Conds:      []RuleCond{{Cond: RuleCondType_MatchAll}},
			RuleAction: RuleAction{Action: RuleActionType_Direct},
		})
	}
	for _, r := range rls {
		mp.matchs = append(mp.matchs, r.NewMatchFunc(pv))
		mp.actions = append(mp.actions, r.NewRuleHandler(pv, forwards))
	}

	go func() {
		tk := time.NewTicker(time.Minute)
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				mp.lock.Lock()
				mp.cache = make(map[MatchMeta]RuleHandler, len(mp.cache)/2)
				mp.lock.Unlock()
			}
		}
	}()

	return &mp, nil
}

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/mengseeker/nlink/core/log"
	"github.com/mengseeker/nlink/core/socks"

	"github.com/mengseeker/nlink/core/api"
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

type Rule struct {
	Cond        RuleCond
	CondParam   string
	Action      RuleAction
	ActionParam string
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
	return r, nil
}

func (r Rule) Check(forwards map[string]*ForwardClient) error {
	switch r.Cond {
	case RuleCond_HostMatch, RuleCond_HostPrefix, RuleCond_HostSuffix, RuleCond_GEOIP:
		if r.CondParam == "" {
			return fmt.Errorf("rule condition %q params must not empty", r.Cond)
		}
	case RuleCond_HostRegexp:
		_, err := regexp.Compile(r.CondParam)
		if err != nil {
			return fmt.Errorf("rule condition %q params must a valid regexp, compile err: %v", r.Cond, err)
		}
	case RuleCond_IPCIDR:
		_, _, err := net.ParseCIDR(r.CondParam)
		if err != nil {
			return fmt.Errorf("rule condition %q params must a valid IPCIDR, parse err: %v", r.Cond, err)
		}
	case RuleCond_MatchAll:
	default:
		return fmt.Errorf("unsupport rule condition %q", r.Cond)
	}

	switch r.Action {
	case RuleAction_Direct:
	case RuleAction_Reject:
	case RuleAction_Forward:
		if r.ActionParam == "" || forwards[r.ActionParam] == nil {
			return fmt.Errorf("rule action %s params must in forward servers", r.Action)
		}
	default:
		return fmt.Errorf("unsupport rule action %q", r.Cond)
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

func (r Rule) NewMatchFunc(pv *FuncProvider) func(mm MatchMeta) bool {
	switch r.Cond {
	case RuleCond_HostMatch:
		return func(mm MatchMeta) bool {
			return strings.Contains(mm.Host, r.CondParam)
		}
	case RuleCond_HostPrefix:
		return func(mm MatchMeta) bool {
			return strings.HasPrefix(mm.Host, r.CondParam)
		}
	case RuleCond_HostSuffix:
		return func(mm MatchMeta) bool {
			return strings.HasSuffix(mm.Host, r.CondParam)
		}
	case RuleCond_HostRegexp:
		reg := regexp.MustCompile(r.CondParam)
		return func(mm MatchMeta) bool {
			return reg.Match([]byte(mm.Host))
		}
	case RuleCond_GEOIP:
		return func(mm MatchMeta) bool {
			return pv.GEOIP(pv.Resolv(mm.Host)) == string(r.CondParam)
		}
	case RuleCond_IPCIDR:
		_, ipnet, _ := net.ParseCIDR(r.CondParam)
		return func(mm MatchMeta) bool {
			return ipnet.Contains(pv.Resolv(mm.Host))
		}
	case RuleCond_MatchAll:
		return func(mm MatchMeta) bool { return true }
	default:
		return func(mm MatchMeta) bool { return false }
	}
}

func (r Rule) NewRuleHandler(pv *FuncProvider, forwards map[string]*ForwardClient) RuleHandler {
	switch r.Action {
	case RuleAction_Reject:
		return &RejectRuleHandler{log: pv.log}
	case RuleAction_Direct:
		return &DirectRuleHandler{log: pv.log}
	case RuleAction_Forward:
		return forwards[r.ActionParam]
	default:
		return &RejectRuleHandler{log: pv.log}
	}
}

type RuleHandler interface {
	HTTPRequest(req *http.Request) (resp *http.Response)
	Conn(conn net.Conn, remote *api.ForwardMeta)
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
	forwards map[string]*ForwardClient,
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
		rls = append(rls, Rule{Cond: RuleCond_MatchAll, Action: RuleAction_Direct})
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

type RejectRuleHandler struct {
	log *log.Logger
}

func (h *RejectRuleHandler) HTTPRequest(req *http.Request) (resp *http.Response) {
	h.log.Info("reject request", "url", req.URL)
	return goproxy.NewResponse(req,
		goproxy.ContentTypeText, http.StatusForbidden, http.StatusText(http.StatusForbidden))
}

func (h *RejectRuleHandler) Conn(conn net.Conn, remote *api.ForwardMeta) {
	h.log.Info("reject connect", "address", remote.Address)
	conn.Close()
}

type DirectRuleHandler struct {
	log *log.Logger
}

func (h *DirectRuleHandler) HTTPRequest(req *http.Request) (resp *http.Response) {
	h.log.Info("direct request", "url", req.URL)
	// if return nil, goproxy will direct request
	deleteRequestHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.log.With("url", req.URL.String()).Errorf("http call err: %v", err)
		return NewErrHTTPResponse(req, err.Error())
	}
	return resp
}

func (h *DirectRuleHandler) Conn(conn net.Conn, remote *api.ForwardMeta) {
	l := h.log.With("network", remote.Network, "address", remote.Address)
	l.Info("direct connect")
	defer conn.Close()
	remoteConn, err := net.Dial(remote.Network, remote.Address)
	if err != nil {
		l.Info("dial remote err: %v", err)
		return
	}
	defer remoteConn.Close()
	wg := sync.WaitGroup{}
	copy := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		_, err := io.Copy(w, r)
		if err != nil {
			l.Errorf("copy data err: %v", err)
		}
	}
	wg.Add(2)
	go copy(remoteConn, conn)
	go copy(conn, remoteConn)
	wg.Wait()
}

package http_server

import (
	"TDKCache/peers"
	"TDKCache/peers/protobuf/pb"
	"TDKCache/service/consistenthash"
	"TDKCache/service/log"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	defaultReplicas = 50
	defaultBasePath = "/TDKCache"
)

type HTTPPool struct {
	self        string
	addr        string
	mu          sync.Mutex
	peersMap    *consistenthash.HashRing
	httpGetters map[string]*httpGetter
	router      *httprouter.Router
}

var hsLogger *log.TubeEntry

func newLogger(addr string) *log.TubeEntry {
	return log.Mylog.WithFields(logrus.Fields{
		"component": "TDKCache",
		"category":  fmt.Sprintf("HTTP Server <%s>", addr),
	})
}

func NewHTTPPool(addr string) *HTTPPool {
	hsLogger = newLogger(addr)
	p := &HTTPPool{
		self:   addr,
		addr:   addr,
		router: registerHandlers(),
	}
	return p
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peersMap = consistenthash.NewHashRing(nil, defaultReplicas)
	p.peersMap.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter)

	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + defaultBasePath}
	}

}

func (p *HTTPPool) Add(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peersMap.Add(peers...)
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + defaultBasePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (peers.PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peersMap.Get(key); peer != "" && peer != p.self {
		hsLogger.Info("pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

func (p *HTTPPool) Start(addrs []string, g peers.GroupCache) {
	p.Set(addrs...)
	g.RegisterPeers(p)
	p.ListenAndServe()
}

// HTTP客户端的实现
// baseURL的格式: http://10.0.0.2:8080/TDKCache/
// TODO: 改进为RPC模式
type httpGetter struct {
	baseURL string
}

// 使用HTTP利用protobuf传输

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"http://%v/PBGet?group=%v&key=%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	hsLogger.Debug("send get request: %v", u)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		hsLogger.Error("server return: %v", res.Status)
		return nil, fmt.Errorf("server return: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		hsLogger.Error("reading response body: %v", err)
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	out := &pb.Response{}
	// 解码protobuf响应
	if err = proto.Unmarshal(bytes, out); err != nil {
		hsLogger.Error("decoding response body: %v", err)
		return nil, fmt.Errorf("decoding response body: %v", err)
	}

	return out.Value, nil
}

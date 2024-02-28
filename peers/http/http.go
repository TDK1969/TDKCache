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

// HTTP客户端的实现
// baseURL的格式: http://10.0.0.2:8080/TDKCache/
// TODO: 改进为RPC模式
type httpGetter struct {
	baseURL string
}

// 目前实现的不是gRPC
// HTTP
/*
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	// 构造http请求
	u := fmt.Sprintf(
		"http://%v/Get?group=%v&key=%v",
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
		hsLogger.Error("reading response bod: %v", err)
		return nil, fmt.Errorf("reading response bod: %v", err)
	}
	hsLogger.Debug("successfully get key")
	return bytes, nil
}
*/

// protobuf
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"http://%v/PBGet?group=%v&key=%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	hsLogger.Debug("send get request: %v", u)
	res, err := http.Get(u)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		hsLogger.Error("server return: %v", res.Status)
		return fmt.Errorf("server return: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		hsLogger.Error("reading response body: %v", err)
		return fmt.Errorf("reading response body: %v", err)
	}

	// 解码protobuf响应
	if err = proto.Unmarshal(bytes, out); err != nil {
		hsLogger.Error("decoding response body: %v", err)
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

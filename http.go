package cache

import (
	"catch/hash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

//type GinServer struct {
//	host     string
//	engine   *gin.Engine
//	basePath string
//}
//
//func NewHttpServer(addr string, basePath string) *GinServer {
//	return &GinServer{
//		host:     addr,
//		engine:   gin.Default(),
//		basePath: basePath,
//	}
//}
//
//func (s *GinServer) Init() {
//	s.RegisterRoute()
//	s.Run()
//}
//
//func (s *GinServer) RegisterRoute() {
//	ginGroup := s.engine.Group(s.basePath)
//	{
//		ginGroup.GET("/", func(ctx *gin.Context) {
//			group := ctx.Query("group")
//			key := ctx.Query("key")
//			// 根据key获取相应的数据
//		})
//	}
//}

//func (s *GinServer) Run() {
//	err := s.engine.Run(s.host)
//	if err != nil {
//		log.Println("http listen err: ", err)
//		return
//	}
//}

type httpClient struct {
	bathPath string
	host     string
}

func (h *httpClient) generateURL(group string, key string) string {
	Url := url.URL{
		Path: h.bathPath,
		Host: h.host,
	}
	values := url.Values{} //拼接query参数
	values.Add("group", group)
	values.Add("key", key)
	Url.RawQuery = values.Encode()
	return Url.String()
}

func (h *httpClient) Get(group string, key string) ([]byte, error) {
	res, err := http.Get(h.generateURL(group, key))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ NodeGetter = (*httpClient)(nil)

type ClientPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self          string
	basePath      string
	mu            sync.RWMutex // guards peers and remoteGetters
	nodes         *hash.Map
	remoteGetters map[string]NodeGetter // keyed by e.g. "http://10.0.0.2:8008"
}

func NewClientPool(addr string) IRemoteGetter {
	return &ClientPool{
		self:          addr,
		nodes:         hash.New(hash.DefaultReplicas, nil),
		remoteGetters: make(map[string]NodeGetter),
	}
}

func (h *ClientPool) Register(addrs ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nodes.Add(addrs...)
	for _, node := range addrs {
		h.remoteGetters[node] = &httpClient{
			host:     node,
			bathPath: h.basePath,
		}
	}
}

func (h *ClientPool) PickNode(key string) (NodeGetter, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if node := h.nodes.Get(key); node != "" && node != h.self {
		log.Printf("Pick peer %s", node)
		return h.remoteGetters[node], true
	}
	return nil, false
}

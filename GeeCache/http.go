package GeeCache

import (
	"GeeCache/GeeCache/consistenthash"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPServer struct {
	self           string                 // 记录自己的地址 包括主机名/IP和端口
	basePath       string                 // 节点间通讯地址的前缀
	mu             sync.Mutex             // 锁
	consistHashMap *consistenthash.Map    //一致性哈希的map，用key选择节点
	httpClient     map[string]*httpClient // 映射远程节点的url与对应的httpGetter
}

func NewHTTPServer(self string) *HTTPServer {
	return &HTTPServer{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPServer) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v))
}

func (p *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) { // 当url不匹配时，panic报错
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// 一个完整的URL为/<basepath>/<groupname>/<key>，首先先将<groupname>/<key>提取出来
	// 再用SplitN进行划分即可，分成groupname和key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupname := parts[0]
	key := parts[1]
	group := GetGroup(groupname) // 找对应groupname的组
	if group == nil {
		http.Error(w, "no such group: "+groupname, http.StatusNotFound)
		return
	}
	view, err := group.Get(key) // 当当前组的key
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice()) //w.Write() 将缓存值作为 httpResponse 的 body
}

func (p *HTTPServer) Set(clients ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consistHashMap = consistenthash.New(defaultReplicas, nil)
	p.consistHashMap.Add(clients...) // 添加节点
	p.httpClient = make(map[string]*httpClient, len(clients))
	for _, name := range clients { // 为传入节点创建了一个httpGetter
		p.httpClient[name] = &httpClient{baseURL: name + p.basePath}
	}
}

func (p *HTTPServer) PickPeer(key string) (PeerGetter, bool) { // 根据一致性哈希算法，使用key选择节点，将节点返回
	p.mu.Lock()
	defer p.mu.Unlock()
	if clientName := p.consistHashMap.Get(key); clientName != "" && clientName != p.self {
		p.Log("Pick peer %s", clientName)
		return p.httpClient[clientName], true
	}
	return nil, false
}

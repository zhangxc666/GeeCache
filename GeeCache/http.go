package GeeCache

import (
	"GeeCache/GeeCache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string                 // 记录自己的地址 包括主机名/IP和端口
	basePath    string                 // 节点间通讯地址的前缀
	mu          sync.Mutex             // 锁
	peers       *consistenthash.Map    //一致性哈希的map，用key选择节点
	httpGetters map[string]*httpGetter // 映射远程节点的url与对应的httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...) // 添加节点
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers { // 为传入节点创建了一个httpGetter
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) { // 根据key选择节点，将节点返回客户端
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)

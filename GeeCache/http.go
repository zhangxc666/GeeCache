package GeeCache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/"

type HTTPPool struct {
	self     string // 记录自己的地址 包括主机名/IP和端口
	basePath string // 节点间通讯地址的前缀
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

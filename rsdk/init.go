package rsdk

import (
	"time"

	"github.com/ifaceless/rpcx/rproxy"
)

var (
	proxy *rproxy.Proxy
)

func Init(widowSize int, timeout time.Duration) {
	proxy = rproxy.NewProxy(rproxy.WindowSize(widowSize), rproxy.WindowTimeout(timeout))
	go proxy.Serve()
}

func Release() {
	proxy.Shutdown()
}

package broadcast

import (
	"fmt"
	"net"
)

type BroadcastResult struct {
	Target []byte
	IP     net.IP
	Port   int
}

func (l BroadcastResult) String() string {
	return fmt.Sprintf("BroadcastResult{ Target:%x IP:%v Port:%v}", l.Target, l.IP, l.Port)
}

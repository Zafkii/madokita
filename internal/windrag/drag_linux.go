//go:build linux

package windrag

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func ScreenCursorPos() (x, y int) {
	conn, err := xgb.NewConn()
	if err != nil {
		return 0, 0
	}
	defer conn.Close()

	root := xproto.Setup(conn).DefaultScreen(conn).Root
	reply, err := xproto.QueryPointer(conn, root).Reply()
	if err != nil {
		return 0, 0
	}
	return int(reply.RootX), int(reply.RootY)
}

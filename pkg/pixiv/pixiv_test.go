package pixiv

import (
	"os"
	"testing"
)

func Test(t *testing.T) {
	os.Setenv("PIXIV_PROXY", "socks5://192.168.31.169:20170")
	r, e := RandomImgsWithRetry()
	print(r, e)
}

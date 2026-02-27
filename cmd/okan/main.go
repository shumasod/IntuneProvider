// Command okan はシステムを監視して関西弁で説教してくれるコマンドです。
// おかん (okan) は関西弁で「お母さん」のことです。
package main

import (
	"github.com/shumasod/IntuneProvider/internal/okan"
)

// version はビルド時に -ldflags で注入される
var version = "dev"

func main() {
	okan.SetVersion(version)
	okan.Execute()
}

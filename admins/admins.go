package admins

import (
	"github.com/go-ini/ini"
	"strconv"
	"strings"
)

const (
	DefaultCfgPath = "./config/config.ini"
)

var Admins []int64
var SuperAdmin int64

func init() {
	conf, err := ini.Load(DefaultCfgPath)
	if err != nil {
		panic(err)
	}
	k, err := conf.Section("admin").GetKey("Admins")
	sli := strings.Split(k.String(), ",")
	for _, s := range sli {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic("Admins parse error" + err.Error())
		}
		Admins = append(Admins, id)
	}
	k, err = conf.Section("admin").GetKey("SuperAdmin")
	id, err := strconv.ParseInt(k.String(), 10, 64)
	if err != nil {
		panic("Admins parse error" + err.Error())
	}
	SuperAdmin = id
}

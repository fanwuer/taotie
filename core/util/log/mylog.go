package log

import "taotie/core/util"

func New(filename string) {
	logConf, err := util.ReadfromFile(filename)
	if err != nil {
		panic(err)
	}
	err = Init(string(logConf))
	if err != nil {
		panic(err)
	}
}

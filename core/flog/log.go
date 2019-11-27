package flog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"taotie/core/util/log"
)

var jsconf = `
{
  "UseShortFile": true,
  "Appenders": {
    "console": {
      "Type": "console"
    },
    "base": {
      "Type": "dailyfile",
      "Target": "%s"
    }
  },
  "Loggers": {
    "baseLogger": {
      "Appenders": [
        "base"
      ],
      "Level": "NOTICE"
    },
    "otherLogger": {
      "Appenders": [
        "console"
      ],
      "Level": "NOTICE"
    }
  },
  "Root": {
    "Level": "debug",
    "Appenders": [
      "console"
    ]
  }
}
 `

var (
	Log = log.CurLoggerMananger().Logger("Root")
)

func InitLog(logFile string) {
	os.MkdirAll(filepath.Dir(logFile), 0777)
	m, err := log.NewLoggerManagerWithJsconf(fmt.Sprintf(jsconf, logFile))
	if err != nil {
		panic("log error:" + err.Error())
	}

	Log = m.Logger("baseLogger")
}

func SetLogLevel(level string) {
	if num, ok := log.LogLevelMap[strings.ToUpper(level)]; ok {
		Log.SetLevel(num)
	} else {
		panic("no this level")
	}
}

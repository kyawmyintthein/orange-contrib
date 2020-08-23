package main

import (
	"context"
	"fmt"

	"github.com/kyawmyintthein/orange-contrib/logx"
)

func main() {
	// logx.Init(&logx.LogCfg{
	// 	LogFilePath: "./main.log",
	// 	LogLevel:    "info",
	// 	LogFormat:   "json",
	// 	LogRotation: false,
	// })

	logx.Errorf(context.Background(), fmt.Errorf("not found error"), "Error! log error")
}

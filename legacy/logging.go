package main

import (
	"github.com/maja42/logicat/utils"
	"github.com/sirupsen/logrus"
)

// SetupLogger creates a new root logger
func SetupLogger(consoleLevel logrus.Level) utils.Logger {
	l := utils.NewStdLogger("")
	l.SetLevel(consoleLevel)

	// template := "%[shortLevelName]s[%04[relativeCreated]d] %-150[message]s%[fields]s\n"
	// l.SetFormatter(lcf.NewFormatter(template, nil))
	//
	// if err := lcf.WindowsEnableNativeANSI(true); err != nil {
	// 	l.Warnf("Failed to enable native ANSI logger: %s", err)
	// }
	return l
}

package rendering

import (
	"github.com/maja42/logicat/utils"
	"github.com/sirupsen/logrus"
)

var logger utils.Logger = utils.NewStdLogger("")

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

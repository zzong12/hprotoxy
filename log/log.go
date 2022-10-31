package log

import "github.com/sirupsen/logrus"

var Log = logrus.New()

func init() {
	Log.Formatter = &logrus.TextFormatter{
		DisableQuote: true,
	}
	Log.SetReportCaller(true)
}

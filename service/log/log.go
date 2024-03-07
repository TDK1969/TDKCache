package log

import (
	"TDKCache/service/conf"
	"fmt"
	"io"
	"os"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type TubeLogger struct {
	logger *logrus.Logger
}

type LogEntry struct {
	entry *logrus.Entry
}

func NewLogger(component string, category string) *LogEntry {
	return &LogEntry{
		entry: Mylog.logger.WithFields(
			logrus.Fields{
				"component": component,
				"category":  category,
			},
		),
	}
}

func (logger *TubeLogger) WithFields(fields logrus.Fields) *LogEntry {
	return &LogEntry{
		entry: logger.logger.WithFields(fields),
	}
}

func (e *LogEntry) Trace(format string, v ...any) {
	e.entry.Tracef(fmt.Sprintf(format, v...))
}

func (e *LogEntry) Debug(format string, v ...any) {
	e.entry.Debug(fmt.Sprintf(format, v...))
}

func (e *LogEntry) Info(format string, v ...any) {
	e.entry.Info(fmt.Sprintf(format, v...))
}

func (e *LogEntry) Warn(format string, v ...any) {
	e.entry.Warn(fmt.Sprintf(format, v...))
}

func (e *LogEntry) Error(format string, v ...any) {
	e.entry.Error(fmt.Sprintf(format, v...))
}

func (e *LogEntry) Fatal(format string, v ...any) {
	e.entry.Fatal(fmt.Sprintf(format, v...))
}

func (e *LogEntry) Panic(format string, v ...any) {
	e.entry.Panic(fmt.Sprintf(format, v...))
}

var Mylog = &TubeLogger{logger: logrus.New()}

func init() {

	// 设置输出文件
	filePath := conf.Conf.GetString("log.logDir")
	linkName := filePath + "latest_log.log"
	//        打开指定处的文件，并指定权限为：可读可写，可创建
	src, err := os.OpenFile(linkName, os.O_RDWR|os.O_CREATE, 0755) //0755-> rwx r-x r-x linux知识
	if err != nil {
		fmt.Println("err:", err)
	}
	//log.Out = src
	Mylog.logger.SetOutput(io.MultiWriter(os.Stdout, src))
	Mylog.logger.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:        true,
		NoColors:        false,
		NoFieldsColors:  false,
		FieldsOrder:     []string{"component", "category", "req"},
	})

	// 设置日志级别。低于 Debug 级别的 Trace 将不会被打印
	Mylog.logger.SetLevel(logrus.AllLevels[conf.Conf.GetUint32("log.logLevel")])

	// 设置日志切割 rotatelogs
	writer, _ := rotatelogs.New(
		filePath+"%Y-%m-%d.log",
		//在项目根目录下生成软链文件 latest_log.log 指向最新的日志文件。注意！！！必须在管理员权限下开终端启动。
		rotatelogs.WithLinkName(linkName),
		//日志最大保存时间
		rotatelogs.WithMaxAge(7*24*time.Hour),
		////设置日志切割时间间隔(1天)(隔多久分割一次)
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	// lfshook 决定哪些日志级别可用日志分割
	writeMap := lfshook.WriterMap{
		logrus.PanicLevel: writer,
		logrus.FatalLevel: writer,
		logrus.ErrorLevel: writer,
		logrus.WarnLevel:  writer,
		logrus.InfoLevel:  writer,
		logrus.DebugLevel: writer,
	}

	// 配置 lfshook
	hook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
		// 设置日期格式
		TimestampFormat: "2006.01.02 - 15:04:05",
	})

	//为 logrus 实例添加自定义 hook
	Mylog.logger.AddHook(hook)
}

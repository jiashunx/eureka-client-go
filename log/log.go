package log

import (
    "fmt"
    stdLog "log"
    "os"
)

// Level 日志级别
type Level int8

const (
    DebugLevel Level = iota - 1
    InfoLevel
    WarnLevel
    ErrorLevel
)

// LevelNames 日志级别名称
var LevelNames = map[Level]string{
    DebugLevel: "DEBUG",
    InfoLevel:  "INFO",
    WarnLevel:  "WARN",
    ErrorLevel: "ERROR",
}

// Logger 日志接口
type Logger interface {
    SetLevel(level Level)
    Debug(a ...any)
    Debugf(format string, a ...any)
    Info(a ...any)
    Infof(format string, a ...any)
    Warn(a ...any)
    Warnf(format string, a ...any)
    Error(a ...any)
    Errorf(format string, a ...any)
}

// loggerImpl 默认日志实现(输出控制台)
type loggerImpl struct {
    level      Level
    defaultLog *stdLog.Logger
}

// SetLevel 设置日志级别
func (logger *loggerImpl) SetLevel(level Level) {
    i := int8(level)
    if i >= int8(DebugLevel) && i <= int8(ErrorLevel) {
        logger.level = Level(i)
    }
}

// Debug 输出debug日志
func (logger *loggerImpl) Debug(a ...any) {
    if logger.level > DebugLevel {
        return
    }
    logger.print(DebugLevel, a...)
}

// Debugf 输出debug日志(参数格式化处理)
func (logger *loggerImpl) Debugf(format string, a ...any) {
    if logger.level > DebugLevel {
        return
    }
    logger.printf(DebugLevel, format, a...)
}

// Info 输出info日志
func (logger *loggerImpl) Info(a ...any) {
    if logger.level > InfoLevel {
        return
    }
    logger.print(InfoLevel, a...)
}

// Infof 输出info日志(参数格式化处理)
func (logger *loggerImpl) Infof(format string, a ...any) {
    if logger.level > InfoLevel {
        return
    }
    logger.printf(InfoLevel, format, a...)
}

// Warn 输出warn日志
func (logger *loggerImpl) Warn(a ...any) {
    if logger.level > WarnLevel {
        return
    }
    logger.print(WarnLevel, a...)
}

// Warnf 输出warn日志(参数格式化处理)
func (logger *loggerImpl) Warnf(format string, a ...any) {
    if logger.level > WarnLevel {
        return
    }
    logger.printf(WarnLevel, format, a...)
}

// Error 输出error日志
func (logger *loggerImpl) Error(a ...any) {
    if logger.level > ErrorLevel {
        return
    }
    logger.print(ErrorLevel, a...)
}

// Errorf 输出error日志(参数格式化处理)
func (logger *loggerImpl) Errorf(format string, a ...any) {
    if logger.level > ErrorLevel {
        return
    }
    logger.printf(ErrorLevel, format, a...)
}

// print 输出日志
func (logger *loggerImpl) print(level Level, a ...any) {
    na := make([]any, 0)
    na = append(na, fmt.Sprintf("%5s", LevelNames[level]))
    na = append(na, a...)
    logger.defaultLog.Println(na...)
}

// printf 输出日志(参数格式化处理)
func (logger *loggerImpl) printf(level Level, format string, a ...any) {
    logger.defaultLog.Printf(fmt.Sprintf("%5s", LevelNames[level])+" "+format+" \n", a...)
}

// DefaultLogger 获取默认日志对象
var DefaultLogger = func() Logger {
    return &loggerImpl{
        level:      InfoLevel,
        defaultLog: stdLog.New(os.Stdout, "", stdLog.Ldate|stdLog.Ltime|stdLog.Lmicroseconds|stdLog.Llongfile),
    }
}

package main

import (
    "strings"

    "github.com/natefinch/lumberjack"
    "github.com/sirupsen/logrus"
)

// Custom hook to log messages containing specific keywords to a separate file
type KeywordHook struct {
    writer   *lumberjack.Logger
    keywords []string
}

// Levels specifies which log levels the hook applies to
func (hook *KeywordHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

// Fire writes the log entry to the keyword-specific log file if it contains any keyword
func (hook *KeywordHook) Fire(entry *logrus.Entry) error {
    for _, keyword := range hook.keywords {
        if strings.Contains(entry.Message, keyword) {
            line, err := entry.String()
            if err != nil {
                return err
            }
            _, err = hook.writer.Write([]byte(line))
            return err
        }
    }
    return nil
}

func main() {
    // Configure the base log file (logs all messages)
    baseLogger := &lumberjack.Logger{
        Filename:   "all_logs.log",
        MaxSize:    5,    // Max megabytes before rotating
        MaxBackups: 3,    // Max number of backups
        MaxAge:     28,   // Max days to keep old backups
        Compress:   true, // Compress old backups
    }

    // Configure the error log file (logs error-level messages)
    errorLogger := &lumberjack.Logger{
        Filename:   "error_logs.log",
        MaxSize:    5,
        MaxBackups: 3,
        MaxAge:     28,
        Compress:   true,
    }

    // Configure the keyword-specific log file (logs messages containing specific keywords)
    keywordLogger := &lumberjack.Logger{
        Filename:   "keyword_logs.log",
        MaxSize:    5,
        MaxBackups: 3,
        MaxAge:     28,
        Compress:   true,
    }

    // Set the base log output for all logs
    logrus.SetOutput(baseLogger)
    logrus.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
    })

    // Add a hook to log errors to the error log file
    logrus.AddHook(&ErrorHook{
        writer: errorLogger,
    })

    // Add a hook to log messages containing specific keywords to the keyword log file
    logrus.AddHook(&KeywordHook{
        writer:   keywordLogger,
        keywords: []string{"important", "critical", "alert"},
    })

    // Example log messages
    logrus.Info("This is a normal info message")
    logrus.Warn("This is a warning message")
    logrus.Error("This is an error message")
    logrus.Info("This is an important message that should go to the keyword log")
    logrus.Info("This is a critical alert")
}

// ErrorHook logs error-level messages to a separate file
type ErrorHook struct {
    writer *lumberjack.Logger
}

// Levels specifies that this hook applies to error, fatal, and panic levels
func (hook *ErrorHook) Levels() []logrus.Level {
    return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

// Fire writes the log entry to the error log file
func (hook *ErrorHook) Fire(entry *logrus.Entry) error {
    line, err := entry.String()
    if err != nil {
        return err
    }
    _, err = hook.writer.Write([]byte(line))
    return err
}





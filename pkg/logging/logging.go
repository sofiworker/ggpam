package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	DefaultLevel       = "info"
	DefaultHomeLogging = "false"
	DefaultFilename    = "ggpam.log"
)

const (
	LogLevel = "GGPAM_LOG_LEVEL"
	LogPath  = "GGPAM_LOG_FILE"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Config struct {
	Level           Level
	FilePath        string
	AlsoStderr      bool
	derivedFromHome bool
}

type logger struct {
	mu              sync.Mutex
	level           Level
	logger          *log.Logger
	file            *os.File
	derivedFromHome bool
	alsoStderr      bool
}

var defaultLogger = newLogger()

func newLogger() *logger {
	l := log.New(os.Stderr, "", log.LstdFlags|log.LUTC)
	return &logger{
		level:      LevelInfo,
		logger:     l,
		alsoStderr: true,
	}
}

// ConfigureDefault 基于环境变量与编译时默认值配置全局日志。
// home 为可选的 Home 目录，用于派生默认日志路径。
func ConfigureDefault(home string) error {
	cfg := buildDefaultConfig(home)
	return defaultLogger.configure(cfg)
}

// UpdateHome 在默认路径来自 Home 时，切换到新的 Home 目录。
func UpdateHome(home string) error {
	if home == "" {
		return nil
	}
	return defaultLogger.updateHome(home)
}

func buildDefaultConfig(home string) Config {
	levelStr := firstNonEmpty(os.Getenv(LogLevel), DefaultLevel)
	level := parseLevel(levelStr)

	filePath := os.Getenv(LogPath)
	fromHome := false
	if filePath == "" && isTruthy(DefaultHomeLogging) {
		targetHome := home
		if targetHome == "" {
			if h, err := os.UserHomeDir(); err == nil {
				targetHome = h
			}
		}
		if targetHome != "" {
			filePath = filepath.Join(targetHome, DefaultFilename)
			fromHome = true
		}
	}
	return Config{
		Level:           level,
		FilePath:        filePath,
		AlsoStderr:      true,
		derivedFromHome: fromHome,
	}
}

func (l *logger) configure(cfg Config) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
		l.file = nil
	}

	var writers []io.Writer
	if cfg.FilePath != "" {
		f, err := openFile(cfg.FilePath)
		if err != nil {
			return err
		}
		l.file = f
		writers = append(writers, f)
	}
	if cfg.AlsoStderr || len(writers) == 0 {
		writers = append(writers, os.Stderr)
	}

	l.logger.SetOutput(io.MultiWriter(writers...))
	l.level = cfg.Level
	l.derivedFromHome = cfg.derivedFromHome
	l.alsoStderr = cfg.AlsoStderr
	return nil
}

func (l *logger) updateHome(home string) error {
	l.mu.Lock()
	fromHome := l.derivedFromHome
	level := l.level
	alsoStderr := l.alsoStderr
	l.mu.Unlock()
	if !fromHome {
		return nil
	}
	cfg := Config{
		Level:           level,
		FilePath:        filepath.Join(home, DefaultFilename),
		AlsoStderr:      alsoStderr,
		derivedFromHome: true,
	}
	return l.configure(cfg)
}

func openFile(path string) (*os.File, error) {
	if path == "" {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
}

func parseLevel(value string) Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *logger) logf(level Level, format string, args ...any) {
	l.mu.Lock()
	enabled := level >= l.level
	logger := l.logger
	l.mu.Unlock()
	if !enabled || logger == nil {
		return
	}
	logger.Printf("[%s] %s", level.String(), fmt.Sprintf(format, args...))
}

// Debugf 输出调试日志。
func Debugf(format string, args ...any) {
	defaultLogger.logf(LevelDebug, format, args...)
}

// Infof 输出信息级别日志。
func Infof(format string, args ...any) {
	defaultLogger.logf(LevelInfo, format, args...)
}

// Warnf 输出警告级别日志。
func Warnf(format string, args ...any) {
	defaultLogger.logf(LevelWarn, format, args...)
}

// Errorf 输出错误级别日志。
func Errorf(format string, args ...any) {
	defaultLogger.logf(LevelError, format, args...)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func isTruthy(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

package logging

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

// Logger wraps zerolog.Logger and provides additional methods.
type Logger struct {
	logger zerolog.Logger
	client string
	host   string
}

// NewCustomLogger creates a new CustomLogger with default settings.
func NewCustomLogger() *Logger {
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	output.FormatLevel = func(i interface{}) string {
		level := strings.ToUpper(fmt.Sprintf("%-6s", i))
		switch level {
		case "DEBUG ":
			return fmt.Sprintf("\033[36m| %-6s|\033[0m", level) // Cyan
		case "INFO  ":
			return fmt.Sprintf("\033[32m| %-6s|\033[0m", level) // Green
		case "WARN  ":
			return fmt.Sprintf("\033[33m| %-6s|\033[0m", level) // Yellow
		case "ERROR ":
			return fmt.Sprintf("\033[31m| %-6s|\033[0m", level) // Red
		case "FATAL ":
			return fmt.Sprintf("\033[35m| %-6s|\033[0m", level) // Magenta
		default:
			return fmt.Sprintf("| %-6s|", level) // Default color
		}
	}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return &Logger{logger: logger}
}

// Level sets the log level for the logger.
func (c *Logger) Level(level zerolog.Level) *Logger {
	c.logger = c.logger.Level(level)

	return c
}

// Build finalizes and returns the CustomLogger.
func (c *Logger) Build() *Logger {
	return &Logger{logger: c.logger}
}

// Info logs a message at Info level.
func (c *Logger) Info(msg string) {
	c.logger.Info().Msg(c.formatWithNetwork(msg))
}

// InfoF logs a formatted message at Info level.
func (c *Logger) InfoF(format string, v ...interface{}) {
	c.logger.Info().Msgf(c.formatWithNetworkF(format, v...))
}

// Error logs a message at Error level.
func (c *Logger) Error(msg string) {
	c.logger.Error().Msg(c.formatWithNetwork(msg))
}

// ErrorF logs a message at Error level.
func (c *Logger) ErrorF(format string, v ...interface{}) {
	c.logger.Error().Msgf(c.formatWithNetworkF(format, v...))
}

// Fatal logs a message at Fatal level and exits the application.
func (c *Logger) Fatal(msg string) {
	c.logger.Fatal().Msg(c.formatWithNetwork(msg))
}

// FatalF logs a formatted message at Fatal level and exits the application.
func (c *Logger) FatalF(format string, v ...interface{}) {
	c.logger.Fatal().Msgf(c.formatWithNetworkF(format, v...))
}

// Debug logs a message at Debug level.
func (c *Logger) Debug(msg string) {
	c.logger.Debug().Msg(c.formatWithNetwork(msg))
}

// DebugF logs a formatted message at Debug level.
func (c *Logger) DebugF(format string, v ...interface{}) {
	c.logger.Debug().Msgf(c.formatWithNetworkF(format, v...))
}

func (c *Logger) Client(client string) {
	c.client = client
}

func (c *Logger) Host(host string) {
	c.host = host
}

func (c *Logger) formatWithNetwork(s string) string {
	return fmt.Sprintf("\033[1;34mHost:\033[0m%s \033[1;34mClient:\033[0m%s \033[1;37m%s\033[0m", c.host, c.client, s)
}

func (c *Logger) formatWithNetworkF(format string, v ...interface{}) string {
	return fmt.Sprintf("\033[1;34mHost:\033[0m%s \033[1;34mClient:\033[0m%s \033[1;37m%s\033[0m", c.host, c.client, fmt.Sprintf(format, v...))
}

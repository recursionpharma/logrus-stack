package logrus_stack

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stack"
)

// NewHook is the initializer for LogrusStackHook{} (implementing logrus.Hook).
// Set levels to callerLevels for which "caller" value may be set, providing a
// single frame of stack. Set levels to stackLevels for which "stack" value may
// be set, providing the full stack (minus logrus).
func NewHook(callerLevels []logrus.Level, stackLevels []logrus.Level) LogrusStackHook {
	return LogrusStackHook{
		CallerLevels: callerLevels,
		StackLevels:  stackLevels,
	}
}

// StandardHook is a convenience initializer for LogrusStackHook{} with
// default args.
func StandardHook() LogrusStackHook {
	return NewHook(logrus.AllLevels, []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel})
}

// LogrusStackHook is an implementation of logrus.Hook interface.
type LogrusStackHook struct {
	// Set levels to CallerLevels for which "caller" value may be set,
	// providing a single frame of stack.
	CallerLevels []logrus.Level

	// Set levels to StackLevels for which "stack" value may be set,
	// providing the full stack (minus logrus).
	StackLevels []logrus.Level
}

// Levels provides the levels to filter.
func (hook LogrusStackHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called by logrus when something is logged.
func (hook LogrusStackHook) Fire(entry *logrus.Entry) error {
	var frames stack.Stack

	// Callers(0) is this function, and Callers(1) is the function that invokes
	// this function.
	_frames := stack.Callers(2)

	// Remove logrus's own frames that seem to appear after the code is through
	// certain hoops. e.g. http handler in a separate package.
	// This is a workaround.
	for idx, frame := range _frames {
		// Skip the initial logrus frames -- original code analyzes the entire stack
		if !strings.Contains(strings.ToLower(frame.File), "github.com/sirupsen/logrus") {
			frames = append(frames, _frames[idx:]...)
			break
		}
	}

	if len(frames) > 0 {
		// If we have a frame, we set it to "caller" field for assigned levels.
		for _, level := range hook.CallerLevels {
			if entry.Level == level {
				entry.Data["caller"] = frames[0]
				break
			}
		}

		// Set the available frames to "stack" field.
		for _, level := range hook.StackLevels {
			if entry.Level == level {
				entry.Data["stack"] = frames
				break
			}
		}
	}

	return nil
}

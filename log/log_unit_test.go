package log

import (
	"io"
	stdLog "log"
	"testing"
)

// discardLogger builds an internal logger that emits to io.Discard so tests
// do not pollute stdout.
func discardLogger(level Level) *loggerImpl {
	return &loggerImpl{
		level:      level,
		defaultLog: stdLog.New(io.Discard, "", 0),
	}
}

// TestUnitLog_SetLevel exercises both branches of the range check in SetLevel:
// in-range values are applied, out-of-range values are ignored.
func TestUnitLog_SetLevel(t *testing.T) {
	log := &loggerImpl{}

	// In-range values (true branch): every valid level plus the boundaries.
	for _, lvl := range []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel} {
		log.SetLevel(lvl)
		if log.level != lvl {
			t.Fatalf("SetLevel(%d) => level %d, want %d", lvl, log.level, lvl)
		}
	}

	// Out-of-range values (false branch): level must remain unchanged.
	for _, lvl := range []Level{Level(100), Level(-100), Level(ErrorLevel + 1), Level(TraceLevel - 1)} {
		log.SetLevel(ErrorLevel)
		log.SetLevel(lvl)
		if log.level != ErrorLevel {
			t.Fatalf("SetLevel(out-of-range %d) changed level to %d, want %d", lvl, log.level, ErrorLevel)
		}
	}
}

// TestUnitLog_LevelNames verifies every LevelNames entry, which is looked up
// by print/printf.
func TestUnitLog_LevelNames(t *testing.T) {
	want := map[Level]string{
		TraceLevel: "TRACE",
		DebugLevel: "DEBUG",
		InfoLevel:  "INFO",
		WarnLevel:  "WARN",
		ErrorLevel: "ERROR",
	}
	for lvl, name := range want {
		if got := LevelNames[lvl]; got != name {
			t.Fatalf("LevelNames[%d] = %q, want %q", lvl, got, name)
		}
	}
}

// TestUnitLog_MethodsEmit calls every method pair at TraceLevel so that all of
// them emit (false branch of each guard) and reach print/printf.
func TestUnitLog_MethodsEmit(t *testing.T) {
	log := discardLogger(TraceLevel)

	log.Trace("trace", "emit")
	log.Tracef("tracef %s", "emit")
	log.Debug("debug", "emit")
	log.Debugf("debugf %s", "emit")
	log.Info("info", "emit")
	log.Infof("infof %s", "emit")
	log.Warn("warn", "emit")
	log.Warnf("warnf %s", "emit")
	log.Error("error", "emit")
	log.Errorf("errorf %s", "emit")
}

// TestUnitLog_MethodsSuppress exercises the early-return (true branch) guards
// for Trace/Debug/Info/Warn by using ErrorLevel as the current level, while
// Error/Errorf still emit.
func TestUnitLog_MethodsSuppress(t *testing.T) {
	log := discardLogger(ErrorLevel)

	log.Trace("trace", "suppressed")
	log.Tracef("tracef %s", "suppressed")
	log.Debug("debug", "suppressed")
	log.Debugf("debugf %s", "suppressed")
	log.Info("info", "suppressed")
	log.Infof("infof %s", "suppressed")
	log.Warn("warn", "suppressed")
	log.Warnf("warnf %s", "suppressed")

	// Error/Errorf emit at ErrorLevel.
	log.Error("error", "emit")
	log.Errorf("errorf %s", "emit")
}

// TestUnitLog_ErrorSuppress exercises the true branch of the Error/Errorf guard
// using an out-of-range current level (only constructible via the internal
// struct), since SetLevel clamps to ErrorLevel.
func TestUnitLog_ErrorSuppress(t *testing.T) {
	log := discardLogger(Level(100))

	log.Error("error", "suppressed")
	log.Errorf("errorf %s", "suppressed")
}

// TestUnitLog_PrintRecover triggers the deferred recover in print by using a
// nil defaultLog, which panics on the method call.
func TestUnitLog_PrintRecover(t *testing.T) {
	log := &loggerImpl{level: TraceLevel, defaultLog: nil}
	// Reaches print via the public method.
	log.Trace("recover-print")
	// Also call print directly to be explicit.
	log.print(TraceLevel, "recover-print-direct")
}

// TestUnitLog_PrintfRecover triggers the deferred recover in printf by using a
// nil defaultLog, which panics on the method call.
func TestUnitLog_PrintfRecover(t *testing.T) {
	log := &loggerImpl{level: TraceLevel, defaultLog: nil}
	// Reaches printf via the public method.
	log.Tracef("recover-printf %s", "value")
	// Also call printf directly to be explicit.
	log.printf(TraceLevel, "recover-printf-direct %s", "value")
}

// TestUnitLog_DefaultLogger covers DefaultLogger().
func TestUnitLog_DefaultLogger(t *testing.T) {
	logger := DefaultLogger()
	if logger == nil {
		t.Fatal("DefaultLogger() returned nil")
	}
	impl, ok := logger.(*loggerImpl)
	if !ok {
		t.Fatalf("DefaultLogger() returned %T, want *loggerImpl", logger)
	}
	if impl.level != InfoLevel {
		t.Fatalf("DefaultLogger() level = %d, want %d", impl.level, InfoLevel)
	}
	if impl.defaultLog == nil {
		t.Fatal("DefaultLogger() defaultLog is nil")
	}
}

// TestUnitLog_DefaultLoggerImpl covers the package-level DefaultLoggerImpl var.
func TestUnitLog_DefaultLoggerImpl(t *testing.T) {
	if DefaultLoggerImpl == nil {
		t.Fatal("DefaultLoggerImpl is nil")
	}
	impl, ok := DefaultLoggerImpl.(*loggerImpl)
	if !ok {
		t.Fatalf("DefaultLoggerImpl is %T, want *loggerImpl", DefaultLoggerImpl)
	}
	if impl.level != InfoLevel {
		t.Fatalf("DefaultLoggerImpl level = %d, want %d", impl.level, InfoLevel)
	}

	// Exercise the shared instance.
	DefaultLoggerImpl.SetLevel(ErrorLevel)
	DefaultLoggerImpl.Errorf("default logger impl: %s", "unit")
	DefaultLoggerImpl.SetLevel(InfoLevel)
}

package log

import "testing"

func TestDefaultLogger(t *testing.T) {
    logger := DefaultLogger()
    logger.SetLevel(DebugLevel)
    logger.Debug("hello", "world")
    logger.Debugf("hello: %s", "world")
    logger.SetLevel(InfoLevel)
    logger.Debug("hello", "world")
    logger.Debugf("hello: %s", "world")
    logger.Info("hello", "world")
    logger.Infof("hello: %s", "world")
    logger.Warn("hello", "world")
    logger.Warnf("hello: %s", "world")
    logger.Error("hello", "world")
    logger.Errorf("hello: %s", "world")
}

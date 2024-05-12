package log

import "testing"

func TestDefaultLogger(t *testing.T) {
    logger := DefaultLogger()
    logger.SetLevel(DebugLevel)
    logger.Trace("hello1", "world")
    logger.Tracef("hello1: %s", "world")
    logger.Debug("hello2", "world")
    logger.Debugf("hello2: %s", "world")
    logger.SetLevel(InfoLevel)
    logger.Debug("hello3", "world")
    logger.Debugf("hello3: %s", "world")
    logger.Info("hello4", "world")
    logger.Infof("hello4: %s", "world")
    logger.Warn("hello5", "world")
    logger.Warnf("hello5: %s", "world")
    logger.Error("hello6", "world")
    logger.Errorf("hello6: %s", "world")
}

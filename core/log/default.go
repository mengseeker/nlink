package log

var (
	DefaultLogger = NewLogger()
)

var Debug = DefaultLogger.Debug
var DebugCtx = DefaultLogger.DebugCtx
var Enabled = DefaultLogger.Enabled
var Error = DefaultLogger.Error
var ErrorCtx = DefaultLogger.ErrorCtx
var Handler = DefaultLogger.Handler
var Info = DefaultLogger.Info
var InfoCtx = DefaultLogger.InfoCtx
var Log = DefaultLogger.Log
var LogAttrs = DefaultLogger.LogAttrs
var WarnCtx = DefaultLogger.WarnCtx
var With = DefaultLogger.With
var WithGroup = DefaultLogger.WithGroup
var Debugf = DefaultLogger.Debugf
var Infof = DefaultLogger.Infof
var Warnf = DefaultLogger.Warnf
var Errorf = DefaultLogger.Errorf

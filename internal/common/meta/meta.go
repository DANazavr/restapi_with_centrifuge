package meta

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	RequestIDKey contextKey = "requestID"
	RootCtxKey   contextKey = "rootCtx"
	LogKey       contextKey = "logKey"
	ConfigKey    contextKey = "configKey"
	ServerKey    contextKey = "serverKey"
)

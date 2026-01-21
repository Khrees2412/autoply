package app

import "context"

// contextKey is used to store App in context
type contextKey struct{}

var appContextKey = contextKey{}

// GetAppFromContext retrieves the App from context
func GetAppFromContext(ctx context.Context) *App {
	app, ok := ctx.Value(appContextKey).(*App)
	if !ok {
		return nil
	}
	return app
}

// SetAppInContext stores the App in context
func SetAppInContext(ctx context.Context, app *App) context.Context {
	return context.WithValue(ctx, appContextKey, app)
}

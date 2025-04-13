# What

Slog driver for getting logs to show up properly in Google Cloud logs and easy to read local logging. In
particular for Google Cloud Run services/jobs and Google App Engine.

Google cloud supports structured logging, but the fields need to be a certain way for it to show up
properly and to report logs to Error Logging and so forth.

For local logging it's hard to read structured logs, so there only part of the logs is structured, i.e.
the log attributes show up in a nice pretty printed JSON list and the message is printed as regular text.

# How

Add this to your application and it will automatically setup the correct setup
for getting the correct logs for Google Cloud, and pretty and easy to read logs
when you're developing locally.

Using the meta library (recommended)
```go
func init() {
	if slogdriver.OnGCP() {
		projectID, err := metadata.ProjectID()
		if err != nil {
			panic(err)
		}
		slog.SetDefault(slog.New(slogdriver.NewCloudHandler(projectID, slog.DebugLevel)))
	} else {
		slog.SetDefault(slog.New(slogdriver.NewLocalHandler(slog.DebugLevel)))
	}
}
```

Using environment variable (you need to set this yourself)
```go
func init() {
	if slogdriver.OnGCP() {
		slog.SetDefault(slog.New(slogdriver.NewCloudHandler(os.Getenv("PROJECT_ID"), slog.DebugLevel))
	} else {
		slog.SetDefault(slog.New(slogdriver.NewLocalHandler(slog.DebugLevel)))
	}
}
```

Logging without trace capability
```go
slog.Info("some info", slog.String("username", username))
```

Logging with trace capability (you need to use the `WithTraceContext` middleware for this to work)
```go
slog.ErrorContext("some error text", slog.Any("error", err))
slog.InfoContext("some info", slog.String("username", username))
```

Setting up the trace middleware using [chi](https://github.com/go-chi/chi) (recommended)
```go
r := chi.NewRouter()
r.Use(slogdriver.WithTraceContext)
r.Get("/v1/hello", HelloHandler)
http.ListenAndServe(addr, r)
```

Using regular net/http
```go
mux := http.NewServeMux()
mux.HandleFunc("/v1/hello", HelloHandler)
wrappedMux := slogdriver.WithTraceContext(mux)
http.ListenAndServe(addr, wrappedMux)
```

##### Based on
- [betterstack.com/guides/logging-in-go](https://betterstack.com/community/guides/logging/logging-in-go/#customizing-slog-handlers)
- [github.com/remko/cloudrun-slog](https://github.com/remko/cloudrun-slog/blob/main/main.go)

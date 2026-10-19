package webtest

import "time"

// WebClientOptions customizes the behavior of a [WebClient]. The zero value is
// valid: no file logging, the default 30 second request timeout, and no
// destination restrictions.
type WebClientOptions struct {
	// WriteToFile enables appending request logs to a file. Default false.
	WriteToFile bool `json:"writeToFile"`

	// FilePath is the log file that requests are appended to when WriteToFile
	// is true. When empty, a file named ./YYYYMMDDThhmmssZ.log (a Windows-safe
	// UTC timestamp) is created and reused for the client's configuration.
	//
	// FilePath is trusted local filesystem configuration. Do not populate it
	// from untrusted input. Paths containing a ".." segment are rejected, but
	// absolute paths and symlinks are not confined to a directory.
	FilePath string `json:"filePath"`

	// WriteHeaders appends the request headers to the log file as a JSON line.
	WriteHeaders bool `json:"writeHeaders"`

	// WriteRequest appends the request body to the log file as a JSON line.
	// The full body is buffered in memory to do so.
	WriteRequest bool `json:"writeRequest"`

	// WriteResponse appends the response body to the log file as a JSON line.
	// The full body is buffered in memory and restored for the caller.
	WriteResponse bool `json:"writeResponse"`

	// LogMetadata appends sent and received payload sizes to the call summary
	// line. Bodies are buffered in memory to measure them.
	LogMetadata bool `json:"logMetadata"`

	// RequestTimeout is the end-to-end timeout for each request. Zero or a
	// negative value uses the package default of 30 seconds; it never
	// disables the timeout.
	RequestTimeout time.Duration `json:"requestTimeout"`

	// AllowedOrigins optionally restricts requests and redirects to exact
	// HTTP(S) origins such as "https://api.example.com" or
	// "http://localhost:8080". Entries must not contain credentials, paths,
	// queries, or fragments, and hostnames are matched exactly. An empty list
	// allows any destination.
	AllowedOrigins []string `json:"allowedOrigins"`
}

// Package webtest is a lightweight HTTP client for API testing and
// benchmark-style command-line workflows.
//
// A [WebClient] wraps the common request methods (GET, POST, PATCH, PUT,
// DELETE), prints a colored per-request summary with timing information, and
// can optionally append request logs to a file. Helpers such as
// [WebClient.AddQueryParams] and [ReadCsv] support data-driven request loops.
//
//	client := webtest.InitWebClient().
//		Headers(map[string]string{"Content-Type": "application/json"}).
//		Options(webtest.WebClientOptions{RequestTimeout: 10 * time.Second})
//
//	res, err := client.Get("http://localhost:8080/health")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.CloseResponse(res)
//
// # Console output
//
// Every request prints a summary line colored by status class (2xx green,
// 3xx yellow, everything else red):
//
//	GET http://localhost:8080/resource 200 OK total=0.043s connect=0.007s
//	GET http://localhost:8080/resource 200 OK total=0.012s conn=reused
//
// # Security
//
// When request URLs can be influenced by untrusted input, set
// [WebClientOptions.AllowedOrigins] to restrict the client to explicit
// origins. The allowlist is enforced before the initial request and on every
// redirect.
//
// [WebClientOptions.FilePath] and the path passed to [ReadCsv] are trusted
// local filesystem configuration and must not be derived from untrusted input.
package webtest

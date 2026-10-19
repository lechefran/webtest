package webtest_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lechefran/webtest"
)

func ExampleInitWebClient() {
	client := webtest.InitWebClient().
		Headers(map[string]string{"Content-Type": "application/json"}).
		Options(webtest.WebClientOptions{RequestTimeout: 10 * time.Second})

	res, err := client.Get("http://localhost:8080/health")
	if err != nil {
		log.Fatal(err)
	}
	defer client.CloseResponse(res)

	fmt.Println(webtest.Is2xxSuccessful(res))
}

func ExampleWebClient_Post() {
	client := webtest.InitWebClient().
		Headers(map[string]string{"Content-Type": "application/json"})

	res, err := client.Post("http://localhost:8080/resource", []byte(`{"name":"widget"}`))
	if err != nil {
		log.Fatal(err)
	}
	defer client.CloseResponse(res)

	fmt.Println(res.StatusCode)
}

func ExampleWebClient_AddQueryParams() {
	client := webtest.InitWebClient()

	url, err := client.AddQueryParams("https://api.example.com/search?page=1", map[string]string{
		"q":    "go http",
		"sort": "desc",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(url)
	// Output: https://api.example.com/search?page=1&q=go+http&sort=desc
}

func ExampleWebClientOptions_allowedOrigins() {
	client := webtest.InitWebClient().Options(webtest.WebClientOptions{
		AllowedOrigins: []string{"https://api.example.com"},
	})

	_, err := client.Get("http://169.254.169.254/latest/meta-data")
	fmt.Println(err)
	// Output: request destination "http://169.254.169.254/latest/meta-data" is not in the allowed origins
}

func ExampleWebClientOptions_fileLogging() {
	client := webtest.InitWebClient().Options(webtest.WebClientOptions{
		WriteToFile:   true,
		FilePath:      "./requests.log",
		WriteHeaders:  true,
		WriteRequest:  true,
		WriteResponse: true,
		LogMetadata:   true,
	})

	res, err := client.Get("http://localhost:8080/resource")
	if err != nil {
		log.Fatal(err)
	}
	defer client.CloseResponse(res)
}

func ExampleReadCsv() {
	dir, err := os.MkdirTemp("", "webtest-example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "input.csv")
	if err := os.WriteFile(path, []byte("id,name\n1,alpha\n2,beta\n"), 0o600); err != nil {
		log.Fatal(err)
	}

	rows, err := webtest.ReadCsv(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range rows[1:] {
		fmt.Println(row[0], row[1])
	}
	// Output:
	// 1 alpha
	// 2 beta
}

func ExampleIs2xxSuccessful() {
	fmt.Println(webtest.Is2xxSuccessful(&http.Response{StatusCode: http.StatusNoContent}))
	fmt.Println(webtest.Is2xxSuccessful(&http.Response{StatusCode: http.StatusNotFound}))
	fmt.Println(webtest.Is2xxSuccessful(nil))
	// Output:
	// true
	// false
	// false
}

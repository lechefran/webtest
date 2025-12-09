# webtest

_webtest_ is a testing package that can be used to test APIs and web services. It simplifies the testing setup with 
flexibility and safely.

## Install
```text
go get github.com/lechefran/webtest
```


## Example
```go
PingUrl := os.Getenv("PING_URL")
InstallUrl := os.Getenv("INSTALL_URL")
ApiUrl := os.Getenv("API_URL")

var wg sync.WaitGroup

// headers
headers := map[string]string{}
headers["Content-Type"] = "application/json"

// options
opts := webtest.WebClientOptions{
    WriteToFile: false,
}

// initialize client and call ping endpoint
client := webtest.InitWebClient().Headers(&headers).Options(&opts)

if res, err := client.Get(PingUrl); err != nil {
    log.Println("Cannot connect to benchmarking service...")
    log.Println(err)
} else {
    log.Println("Successfully connected to benchmarking service...")
    _ = res.Body.Close()
}

body := struct {
    Install       string `json:"install"`
    LoadIndexes   bool   `json:"loadIndexes"`
    DocumentCount int    `json:"documentCount"`
}{
    Install:       "full",
    LoadIndexes:   true,
    DocumentCount: 2000000,
}

// run full installation for documents using the same client
if bytes, err := json.Marshal(body); err != nil {
    log.Fatal(err)
} else {
    if res, err := client.Post(InstallUrl, bytes); err != nil {
        log.Fatal(err)
    } else {
        _ = res.Body.Close()
    }
}

wg.Add(1)
go func() {
    defer wg.Done()

    // read csv file
    ids := webtest.ReadCsv("./ids.csv") 
    
    // initialize web client and add standard headers and client options
    client := webtest.InitWebClient().Headers(&headers).Options(&webtest.WebClientOptions{
        WriteToFile: true,
        FilePath:    "./log/col-scan-restaurant-id-results.txt",
    })
    
    for _, id := range ids {
        res, err := client.Get(addUrlQueryParams(ApiUrl, map[string]string{"id": id[0]}))
        if err != nil {
            log.Fatal(err)
        }
        err = res.Body.Close()
        if err != nil {
            log.Println(err)
        }
    }
}()
wg.Wait()
```
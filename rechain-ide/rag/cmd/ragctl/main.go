package main

import (
  "bytes"
  "encoding/json"
  "flag"
  "fmt"
  "io"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "time"
)

type IndexRequest struct {
  SchemaVersion string   `json:"schema_version"`
  Repo          string   `json:"repo"`
  Files         []string `json:"files"`
}

func main() {
  ragURL := flag.String("rag", "http://localhost:8083", "rag base URL")
  repo := flag.String("repo", "", "repo name")
  root := flag.String("root", ".", "root folder to index")
  interval := flag.Int("interval", 10, "seconds between re-index")
  flag.Parse()

  for {
    files, err := collectFiles(*root)
    if err != nil {
      fmt.Println("collect error:", err)
      return
    }

    req := IndexRequest{
      SchemaVersion: "0.1.0",
      Repo:          *repo,
      Files:         files,
    }

    if err := postIndex(*ragURL, req); err != nil {
      fmt.Println("index error:", err)
      return
    }

    fmt.Printf("indexed %d files\n", len(files))
    time.Sleep(time.Duration(*interval) * time.Second)
  }
}

func collectFiles(root string) ([]string, error) {
  out := []string{}
  err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if info.IsDir() {
      name := info.Name()
      if name == ".git" || name == "node_modules" || name == "vendor" {
        return filepath.SkipDir
      }
      return nil
    }
    if strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), ".md") {
      out = append(out, path)
    }
    return nil
  })
  return out, err
}

func postIndex(ragURL string, req IndexRequest) error {
  body, _ := json.Marshal(req)
  resp, err := httpPost(ragURL+"/index", body)
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  data, _ := io.ReadAll(resp.Body)
  if resp.StatusCode != 200 {
    return fmt.Errorf("bad status: %s", string(data))
  }
  return nil
}

func httpPost(url string, body []byte) (*http.Response, error) {
  return http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
}

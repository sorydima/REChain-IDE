package main

import (
  "bytes"
  "encoding/json"
  "flag"
  "fmt"
  "io"
  "net/http"
)

type IndexRequest struct {
  SchemaVersion string   `json:"schema_version"`
  Repo          string   `json:"repo"`
  Files         []string `json:"files"`
}

func main() {
  ragURL := flag.String("rag", "http://localhost:8083", "rag base URL")
  repo := flag.String("repo", "", "repo name")
  files := flag.String("files", "", "comma-separated file paths")
  flag.Parse()

  if *files == "" {
    fmt.Println("missing -files")
    return
  }

  req := IndexRequest{
    SchemaVersion: "0.1.0",
    Repo:          *repo,
    Files:         splitFiles(*files),
  }

  body, _ := json.Marshal(req)
  resp, err := http.Post(*ragURL+"/index", "application/json", bytes.NewReader(body))
  if err != nil {
    fmt.Println("index error:", err)
    return
  }
  defer resp.Body.Close()

  data, _ := io.ReadAll(resp.Body)
  fmt.Println(string(data))
}

func splitFiles(s string) []string {
  out := []string{}
  current := ""
  for _, ch := range s {
    if ch == ',' {
      if current != "" {
        out = append(out, current)
      }
      current = ""
      continue
    }
    current += string(ch)
  }
  if current != "" {
    out = append(out, current)
  }
  return out
}

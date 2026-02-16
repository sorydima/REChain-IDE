package main

import (
  "bytes"
  "encoding/json"
  "flag"
  "fmt"
  "io"
  "net/http"
  "os"
  "time"
)

func main() {
  server := flag.String("server", "http://localhost:8081", "orchestrator base url")
  cmd := flag.String("cmd", "health", "health|submit|status|result|metrics")
  input := flag.String("input", "", "task input")
  task := flag.String("task", "", "task id")
  flag.Parse()

  switch *cmd {
  case "health":
    get(*server + "/health")
  case "metrics":
    get(*server + "/metrics")
  case "status":
    if *task == "" {
      fatal("missing -task")
    }
    get(*server + "/tasks/" + *task)
  case "result":
    if *task == "" {
      fatal("missing -task")
    }
    get(*server + "/tasks/" + *task + "/result")
  case "submit":
    if *input == "" {
      fatal("missing -input")
    }
    submit(*server, *input)
  default:
    fatal("unknown cmd")
  }
}

func submit(server string, input string) {
  payload := map[string]interface{}{
    "schema_version": "0.1.0",
    "type":           "patch",
    "input":          input,
    "context":        []interface{}{},
    "constraints":    []interface{}{},
    "metadata": map[string]interface{}{
      "requester": "cli",
      "priority":  "normal",
    },
  }
  body, _ := json.Marshal(payload)
  req, err := http.NewRequest(http.MethodPost, server+"/tasks", bytes.NewReader(body))
  if err != nil {
    fatal(err.Error())
  }
  req.Header.Set("Content-Type", "application/json")
  client := &http.Client{Timeout: 5 * time.Second}
  resp, err := client.Do(req)
  if err != nil {
    fatal(err.Error())
  }
  defer resp.Body.Close()
  data, _ := io.ReadAll(resp.Body)
  fmt.Println(string(data))
}

func get(url string) {
  client := &http.Client{Timeout: 5 * time.Second}
  resp, err := client.Get(url)
  if err != nil {
    fatal(err.Error())
  }
  defer resp.Body.Close()
  data, _ := io.ReadAll(resp.Body)
  fmt.Println(string(data))
}

func fatal(msg string) {
  fmt.Fprintln(os.Stderr, msg)
  os.Exit(1)
}

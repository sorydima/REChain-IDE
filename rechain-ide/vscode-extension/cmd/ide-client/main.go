package main

import (
  "bytes"
  "encoding/json"
  "flag"
  "fmt"
  "io"
  "net/http"
  "time"
)

const schemaVersion = "0.1.0"

type TaskSpec struct {
  SchemaVersion string       `json:"schema_version"`
  ID            string       `json:"id"`
  Type          string       `json:"type"`
  Input         string       `json:"input"`
  Context       []ContextRef `json:"context"`
  Constraints   []Constraint `json:"constraints"`
  Metadata      Metadata     `json:"metadata"`
}

type ContextRef struct {
  Type string `json:"type"`
  Path string `json:"path"`
  Rev  string `json:"rev"`
}

type Constraint struct {
  Key   string      `json:"key"`
  Value interface{} `json:"value"`
}

type Metadata struct {
  Requester string `json:"requester"`
  Priority  string `json:"priority"`
}

type TaskStatus struct {
  SchemaVersion string  `json:"schema_version"`
  ID            string  `json:"id"`
  State         string  `json:"state"`
  Progress      float64 `json:"progress"`
  StartedAt     string  `json:"started_at"`
  UpdatedAt     string  `json:"updated_at"`
}

func main() {
  server := flag.String("server", "http://localhost:8081", "orchestrator base URL")
  input := flag.String("input", "", "task input prompt")
  taskType := flag.String("type", "patch", "task type")
  flag.Parse()

  if *input == "" {
    fmt.Println("missing -input")
    return
  }

  spec := TaskSpec{
    SchemaVersion: schemaVersion,
    Type:          *taskType,
    Input:         *input,
    Context:       []ContextRef{},
    Constraints:   []Constraint{},
    Metadata:      Metadata{Requester: "vscode", Priority: "normal"},
  }

  body, _ := json.Marshal(spec)
  resp, err := http.Post(*server+"/tasks", "application/json", bytes.NewReader(body))
  if err != nil {
    fmt.Println("submit error:", err)
    return
  }
  defer resp.Body.Close()

  var status TaskStatus
  if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
    fmt.Println("decode error:", err)
    return
  }

  for {
    time.Sleep(200 * time.Millisecond)
    r, err := http.Get(*server + "/tasks/" + status.ID)
    if err != nil {
      fmt.Println("status error:", err)
      return
    }
    data, _ := io.ReadAll(r.Body)
    r.Body.Close()

    var s TaskStatus
    if err := json.Unmarshal(data, &s); err != nil {
      fmt.Println("status decode error:", err)
      return
    }

    if s.State == "completed" || s.State == "canceled" {
      fmt.Println(string(data))
      return
    }
  }
}

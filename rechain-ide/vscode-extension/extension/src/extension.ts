import * as vscode from "vscode";
import * as fs from "fs";
import * as os from "os";
import * as path from "path";
import { exec } from "child_process";
import { randomUUID } from "crypto";

const defaultServer = "http://localhost:8081";

export function activate(context: vscode.ExtensionContext) {
  const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
  statusBar.text = "REChain: idle";
  statusBar.tooltip = "REChain queue depth";
  statusBar.show();
  context.subscriptions.push(statusBar);

  const disposable = vscode.commands.registerCommand("rechain.submitTask", async () => {
    const input = await vscode.window.showInputBox({
      title: "Submit REChain task",
      prompt: "Describe the change you want",
    });

    if (!input) {
      return;
    }

    const server = await vscode.window.showInputBox({
      title: "Orchestrator URL",
      value: defaultServer,
    });

    if (!server) {
      return;
    }

    const requestId = randomUUID();

    const models = await vscode.window.showInputBox({
      title: "Models (optional)",
      prompt: "Comma-separated driver IDs, e.g. model_a,hf_gigachat3_702b_preview",
      value: "",
    });

    const routing = await vscode.window.showInputBox({
      title: "Routing policy (optional)",
      prompt: "latency, cost, quality, weighted, or weighted_quality",
      value: "latency",
    });

    const priority = await vscode.window.showQuickPick(["high", "normal", "low"], {
      title: "Priority",
      placeHolder: "Select task priority",
    });

    const retries = await vscode.window.showInputBox({
      title: "Retries (optional)",
      prompt: "Number of retries per driver, e.g. 1",
      value: "",
    });

    const retryBackoff = await vscode.window.showInputBox({
      title: "Retry backoff ms (optional)",
      prompt: "Backoff between retries, e.g. 250",
      value: "",
    });

    const budgetUsd = await vscode.window.showInputBox({
      title: "Budget USD (optional)",
      prompt: "Total cost cap, e.g. 0.05",
      value: "",
    });

    let weightCost = "";
    let weightLatency = "";
    if (routing === "weighted" || routing === "weighted_quality") {
      weightCost = (await vscode.window.showInputBox({
        title: "Weight cost",
        prompt: "e.g. 0.4",
        value: "0.4",
      })) || "";
      weightLatency = (await vscode.window.showInputBox({
        title: "Weight latency",
        prompt: "e.g. 0.6",
        value: "0.6",
      })) || "";
    }

    let weightQuality = "";
    if (routing === "weighted_quality") {
      weightQuality = (await vscode.window.showInputBox({
        title: "Weight quality",
        prompt: "e.g. 0.2",
        value: "0.2",
      })) || "";
    }

    const includeContext = await vscode.window.showQuickPick(["yes", "no"], {
      title: "Include open file context",
      placeHolder: "Attach open file paths as context references",
    });

    const contextRefs = includeContext === "yes" ? collectOpenFiles() : [];

    const task = {
      schema_version: "0.1.0",
      type: "patch",
      input,
      context: contextRefs,
      constraints: [
        models ? { key: "models", value: models } : null,
        routing ? { key: "routing", value: routing } : null,
        retries ? { key: "retries", value: Number(retries) } : null,
        retryBackoff ? { key: "retry_backoff_ms", value: Number(retryBackoff) } : null,
        budgetUsd ? { key: "budget_usd", value: Number(budgetUsd) } : null,
        weightCost ? { key: "weight_cost", value: Number(weightCost) } : null,
        weightLatency ? { key: "weight_latency", value: Number(weightLatency) } : null,
        weightQuality ? { key: "weight_quality", value: Number(weightQuality) } : null,
      ].filter(Boolean),
      metadata: { requester: "vscode", priority: priority || "normal" },
    };

    const status = await submitTask(server, task, requestId);
    if (!status) {
      vscode.window.showErrorMessage("Failed to submit task");
      return;
    }

    const finalStatus = await pollStatus(server, status.id, requestId);
    if (!finalStatus) {
      vscode.window.showErrorMessage("Failed to fetch status");
      return;
    }

    if (finalStatus.state !== "completed") {
      vscode.window.showWarningMessage(`Task ${finalStatus.id} finished with state: ${finalStatus.state}`);
      return;
    }

    const result = await fetchResult(server, finalStatus.id, requestId);
    if (!result) {
      vscode.window.showErrorMessage("Failed to fetch merge result");
      return;
    }

    const action = await vscode.window.showInformationMessage(
      `Diff ready for task ${finalStatus.id}`,
      "Preview",
      "Apply"
    );

    const applied = await applyDiff(result.diff, action === "Apply");
    if (applied) {
      vscode.window.showInformationMessage(`Applied diff for task ${finalStatus.id}`);
    } else {
      vscode.window.showWarningMessage(`Diff previewed but not applied for task ${finalStatus.id}`);
    }
  });

  const metricsCmd = vscode.commands.registerCommand("rechain.showMetrics", async () => {
    const server = await vscode.window.showInputBox({
      title: "Orchestrator URL",
      value: defaultServer,
    });
    if (!server) {
      return;
    }
    const text = await fetchMetrics(server);
    if (!text) {
      vscode.window.showErrorMessage("Failed to fetch metrics");
      return;
    }
    const doc = await vscode.workspace.openTextDocument({ content: text, language: "text" });
    await vscode.window.showTextDocument(doc, { preview: false });
  });

  const interval = setInterval(async () => {
    const depth = await fetchQueueDepth(defaultServer);
    if (depth === null) {
      statusBar.text = "REChain: n/a";
      return;
    }
    statusBar.text = `REChain: queue ${depth}`;
  }, 3000);

  context.subscriptions.push(disposable, metricsCmd, { dispose: () => clearInterval(interval) });
}

export function deactivate() {}

async function submitTask(server: string, task: any, rid: string): Promise<any | null> {
  const res = await fetch(`${server}/tasks`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-Request-Id": rid },
    body: JSON.stringify(task),
  });
  if (!res.ok) return null;
  return res.json();
}

async function pollStatus(server: string, id: string, rid: string): Promise<any | null> {
  for (let i = 0; i < 40; i++) {
    await new Promise((r) => setTimeout(r, 250));
    const res = await fetch(`${server}/tasks/${id}`, { headers: { "X-Request-Id": rid } });
    if (!res.ok) return null;
    const s = await res.json();
    if (s.state === "completed" || s.state === "canceled" || s.state === "failed") {
      return s;
    }
  }
  return null;
}

async function fetchResult(server: string, id: string, rid: string): Promise<any | null> {
  const res = await fetch(`${server}/tasks/${id}/result`, { headers: { "X-Request-Id": rid } });
  if (!res.ok) return null;
  return res.json();
}

async function applyDiff(diff: string, shouldApply: boolean): Promise<boolean> {
  const folder = vscode.workspace.workspaceFolders?.[0];
  if (!folder) {
    const doc = await vscode.workspace.openTextDocument({ content: diff, language: "diff" });
    await vscode.window.showTextDocument(doc, { preview: false });
    return false;
  }

  const tmpFile = path.join(os.tmpdir(), `rechain_patch_${Date.now()}.diff`);
  fs.writeFileSync(tmpFile, diff, "utf8");

  const doc = await vscode.workspace.openTextDocument({ content: diff, language: "diff" });
  await vscode.window.showTextDocument(doc, { preview: false });

  if (!shouldApply) {
    return false;
  }

  const applied = await runGitApply(folder.uri.fsPath, tmpFile);
  return applied;
}

function runGitApply(cwd: string, patchFile: string): Promise<boolean> {
  return new Promise((resolve) => {
    exec(`git apply "${patchFile}"`, { cwd }, (err) => {
      resolve(!err);
    });
  });
}

async function fetchMetrics(server: string): Promise<string | null> {
  const res = await fetch(`${server}/metrics`);
  if (!res.ok) return null;
  return res.text();
}

async function fetchQueueDepth(server: string): Promise<number | null> {
  try {
    const res = await fetch(`${server}/queue-depth`);
    if (!res.ok) return null;
    const data = await res.json();
    return typeof data.queue_depth === "number" ? data.queue_depth : null;
  } catch {
    return null;
  }
}

function collectOpenFiles(): { type: string; path: string; rev: string }[] {
  const openDocs = vscode.workspace.textDocuments;
  const items: { type: string; path: string; rev: string }[] = [];
  for (const doc of openDocs) {
    if (doc.isUntitled) continue;
    if (doc.uri.scheme !== "file") continue;
    items.push({ type: "file", path: doc.uri.fsPath, rev: "" });
  }
  return items;
}

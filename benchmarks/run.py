import argparse
import hashlib
import json
import re
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=["seed", "verify", "load"])
    parser.add_argument("--kubeconfig", required=True)
    parser.add_argument("--dataset", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--rate", type=int, default=50)
    parser.add_argument("--duration", default="2m")
    parser.add_argument("--vus", type=int)
    parser.add_argument("--max-vus", type=int)
    parser.add_argument("--products", type=int, default=1000)
    parser.add_argument("--workload", choices=["storefront", "inventory"], default="storefront")
    args = parser.parse_args()
    if args.workload == "inventory" and args.mode != "load":
        parser.error("inventory supports load mode; seed with the storefront workload first")
    args.output.mkdir(parents=True, exist_ok=True)
    kubectl = ["kubectl", "--kubeconfig", args.kubeconfig, "-n", "ecommerce"]

    def run(*command, **kwargs):
        return subprocess.run([*kubectl, *command], check=True, text=True, **kwargs)

    name = f"{args.workload}-{args.mode}-{int(time.time())}"
    script = Path(__file__).with_name(f"{args.workload}.js")
    data = {script.name: script.read_text()}
    if args.workload == "inventory":
        data["inventory.proto"] = (Path(__file__).parent.parent / "shared/proto/inventory/v1/inventory.proto").read_text()
    if args.mode != "seed":
        data["dataset.json"] = args.dataset.read_text()
    config = {"apiVersion": "v1", "kind": "ConfigMap",
              "metadata": {"name": name}, "data": data}
    run("apply", "-f", "-", input=json.dumps(config), capture_output=True)
    env = {"MODE": args.mode, "RATE": str(args.rate), "DURATION": args.duration,
           "PRODUCTS": str(args.products)}
    if args.vus is not None:
        env["VUS"] = str(args.vus)
    if args.max_vus is not None:
        env["MAX_VUS"] = str(args.max_vus)
    job = {
        "apiVersion": "batch/v1", "kind": "Job",
        "metadata": {"name": name, "labels": {"app": "storefront-benchmark"}},
        "spec": {
            "backoffLimit": 0, "activeDeadlineSeconds": 1800,
            "ttlSecondsAfterFinished": 86400,
            "template": {
                "metadata": {"labels": {"app": "storefront-benchmark"}},
                "spec": {
                    "restartPolicy": "Never", "automountServiceAccountToken": False,
                    "securityContext": {"runAsNonRoot": True, "runAsUser": 12345},
                    "containers": [{
                        "name": "k6", "image": "grafana/k6:1.6.1",
                        "args": ["run", "--log-format", "raw", f"/scripts/{script.name}"],
                        "env": [{"name": k, "value": v} for k, v in env.items()],
                        "resources": {"requests": {"cpu": "250m", "memory": "256Mi"},
                                      "limits": {"cpu": "2", "memory": "1Gi"}},
                        "volumeMounts": [{"name": "scripts", "mountPath": "/scripts", "readOnly": True}],
                    }],
                    "volumes": [{"name": "scripts", "configMap": {"name": name}}],
                },
            },
        },
    }
    run("apply", "-f", "-", input=json.dumps(job), capture_output=True)
    metadata = {"job": name, "startedAt": datetime.now(timezone.utc).isoformat(),
                "mode": args.mode, "workload": args.workload, "parameters": env,
                "scriptSha256": hashlib.sha256(script.read_bytes()).hexdigest(),
                "datasetSha256": hashlib.sha256(args.dataset.read_bytes()).hexdigest() if args.mode == "load" else None,
                "target": "Istio Gateway → web → domain gRPC services" if args.workload == "storefront" else "Inventory gRPC through Istio waypoint → PostgreSQL",
                "mix": {"browse": 0.45, "search": 0.35, "detail": 0.20} if args.workload == "storefront" else {"GetStock": 1}}
    (args.output / "run.json").write_text(json.dumps(metadata, indent=2) + "\n")
    print(f"Running {name}", flush=True)
    deadline = time.monotonic() + 1860
    while time.monotonic() < deadline:
        result = run("get", "job", name, "-o", "json", capture_output=True)
        status = json.loads(result.stdout).get("status", {})
        if status.get("succeeded") or status.get("failed"):
            break
        time.sleep(5)
    else:
        run("delete", "job", name, capture_output=True)
        raise SystemExit("Benchmark exceeded its deadline")
    logs = run("logs", f"job/{name}", capture_output=True).stdout
    (args.output / "k6.log").write_text(logs)
    result = re.search(r"BENCHMARK_SUMMARY=(\{[^\n]+\})", logs)
    if result:
        summary = json.loads(result[1])
        (args.output / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
    if args.mode == "seed":
        match = re.search(r"DATASET=(\{[^\n]+\})", logs)
        if match:
            args.dataset.parent.mkdir(parents=True, exist_ok=True)
            args.dataset.write_text(json.dumps(json.loads(match[1]), indent=2) + "\n")
        elif status.get("succeeded"):
            raise SystemExit(f"Seed produced no dataset; inspect {args.output / 'k6.log'}")
    metadata["finishedAt"] = datetime.now(timezone.utc).isoformat()
    metadata["succeeded"] = bool(status.get("succeeded"))
    (args.output / "run.json").write_text(json.dumps(metadata, indent=2) + "\n")
    pods = json.loads(run("get", "pods", "-o", "json", capture_output=True).stdout)
    environment = [{
        "name": pod["metadata"]["name"], "node": pod["spec"].get("nodeName"),
        "containers": [{"name": c["name"], "image": c["image"], "resources": c.get("resources", {})}
                       for c in pod["spec"]["containers"]],
        "status": [{"name": c["name"], "imageID": c.get("imageID"), "restarts": c["restartCount"]}
                   for c in pod.get("status", {}).get("containerStatuses", [])],
    } for pod in pods["items"]]
    (args.output / "environment.json").write_text(json.dumps(environment, indent=2) + "\n")
    run("delete", "configmap", name, capture_output=True)
    if not metadata["succeeded"]:
        raise SystemExit(f"Benchmark failed; inspect {args.output / 'k6.log'}")
    print(f"Results: {args.output}", flush=True)


if __name__ == "__main__":
    main()

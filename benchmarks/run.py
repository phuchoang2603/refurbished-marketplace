import argparse
import json
import re
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=["seed", "load"])
    parser.add_argument("--kubeconfig", required=True)
    parser.add_argument("--dataset", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--rate", type=int, default=50)
    parser.add_argument("--duration", default="2m")
    parser.add_argument("--products", type=int, default=1000)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    kubectl = ["kubectl", "--kubeconfig", args.kubeconfig, "-n", "ecommerce"]

    def run(*command, **kwargs):
        return subprocess.run([*kubectl, *command], check=True, text=True, **kwargs)

    name = f"storefront-{args.mode}-{int(time.time())}"
    script = Path(__file__).with_name("storefront.js")
    data = {"storefront.js": script.read_text()}
    if args.mode == "load":
        data["dataset.json"] = args.dataset.read_text()
    config = {"apiVersion": "v1", "kind": "ConfigMap",
              "metadata": {"name": name}, "data": data}
    run("apply", "-f", "-", input=json.dumps(config), capture_output=True)
    env = {"MODE": args.mode, "RATE": str(args.rate), "DURATION": args.duration,
           "PRODUCTS": str(args.products)}
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
                        "args": ["run", "--log-format", "raw", "/scripts/storefront.js"],
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
                "mode": args.mode, "parameters": env,
                "target": "Istio Gateway → web → domain gRPC services",
                "mix": {"browse": 0.45, "search": 0.35, "detail": 0.20}}
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
        else:
            raise SystemExit(f"Seed failed; inspect {args.output / 'k6.log'}")
    metadata["finishedAt"] = datetime.now(timezone.utc).isoformat()
    metadata["succeeded"] = bool(status.get("succeeded"))
    (args.output / "run.json").write_text(json.dumps(metadata, indent=2) + "\n")
    run("delete", "configmap", name, capture_output=True)
    if not metadata["succeeded"]:
        raise SystemExit(f"Benchmark failed; inspect {args.output / 'k6.log'}")
    print(f"Results: {args.output}", flush=True)


if __name__ == "__main__":
    main()

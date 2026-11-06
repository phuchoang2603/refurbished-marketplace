import argparse
import hashlib
import json
import re
import subprocess
import sys
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
    parser.add_argument("--write-rate", type=int, default=0)
    parser.add_argument("--seed-vus", type=int, default=20)
    parser.add_argument("--deadline-seconds", type=int, default=1800)
    parser.add_argument("--diverse", action="store_true")
    parser.add_argument("--duration", default="2m")
    parser.add_argument("--vus", type=int)
    parser.add_argument("--max-vus", type=int)
    parser.add_argument("--page-size", type=int, default=20,
                        help="Expected catalog cards; 0 skips the count check for mixed rollout traffic")
    parser.add_argument("--products", type=int, default=1000)
    parser.add_argument("--workload", choices=["storefront", "inventory", "mixed"], default="storefront")
    args = parser.parse_args()
    if args.workload == "inventory" and args.mode != "load":
        parser.error("inventory supports load mode; seed with the storefront workload first")
    if args.workload == "mixed" and args.mode == "verify":
        parser.error("verify with the storefront workload")
    if args.write_rate and (args.workload != "mixed" or args.mode != "load"):
        parser.error("write-rate requires mixed load mode")
    if args.rate < 1 or args.write_rate < 0 or args.products < 1 or args.seed_vus < 1:
        parser.error("rates and product count must be positive; write-rate may be zero")
    if args.deadline_seconds <= 120:
        parser.error("deadline must exceed 120 seconds")
    args.output.mkdir(parents=True, exist_ok=True)
    kubectl = ["kubectl", "--kubeconfig", args.kubeconfig, "-n", "ecommerce"]

    def run(*command, **kwargs):
        try:
            return subprocess.run([*kubectl, *command], check=True, text=True, **kwargs)
        except subprocess.CalledProcessError as error:
            if error.stderr:
                print(error.stderr, file=sys.stderr)
            raise

    name = f"{args.workload}-{args.mode}-{int(time.time())}"
    script = Path(__file__).with_name(f"{args.workload}.js")
    script_paths = [script, script.with_name("dataset.js")]
    if args.workload == "mixed":
        script_paths.append(script.with_name("storefront.js"))
    if args.workload == "inventory":
        script_paths.append(Path(__file__).parent.parent / "shared/proto/inventory/v1/inventory.proto")
    data = {path.name: path.read_text() for path in script_paths}
    configs = []
    if args.mode != "seed":
        fixture = json.loads(args.dataset.read_text())
        parts = []
        shard = []
        shard_bytes = 2
        for product in fixture["products"]:
            encoded = json.dumps(product, separators=(",", ":"))
            product_bytes = len(encoded.encode()) + 1
            if product_bytes > 512 * 1024:
                parser.error("a dataset product exceeds the ConfigMap shard size")
            if shard and shard_bytes + product_bytes > 512 * 1024:
                filename = f"products-{len(parts)}.json"
                configs.append((f"{name}-part-{len(parts)}", {filename: "[" + ",".join(shard) + "]"}))
                parts.append(filename)
                shard = []
                shard_bytes = 2
            shard.append(encoded)
            shard_bytes += product_bytes
        if shard:
            filename = f"products-{len(parts)}.json"
            configs.append((f"{name}-part-{len(parts)}", {filename: "[" + ",".join(shard) + "]"}))
            parts.append(filename)
        if not parts:
            parser.error("dataset has no products")
        data["dataset.json"] = json.dumps({"prefix": fixture["prefix"], "parts": parts})
    configs.append((name, data))
    for config_name, config_data in configs:
        config = {"apiVersion": "v1", "kind": "ConfigMap",
                  "metadata": {"name": config_name}, "data": config_data}
        run("create", "-f", "-", input=json.dumps(config), capture_output=True)
    env = {"MODE": args.mode, "RATE": str(args.rate), "DURATION": args.duration,
           "PRODUCTS": str(args.products), "PAGE_SIZE": str(args.page_size),
           "WRITE_RATE": str(args.write_rate), "SEED_VUS": str(args.seed_vus),
           "DIVERSE": str(args.diverse).lower(), "DEADLINE_SECONDS": str(args.deadline_seconds)}
    if args.vus is not None:
        env["VUS"] = str(args.vus)
    if args.max_vus is not None:
        env["MAX_VUS"] = str(args.max_vus)
    job = {
        "apiVersion": "batch/v1", "kind": "Job",
        "metadata": {"name": name, "labels": {"app": "storefront-benchmark"}},
        "spec": {
            "backoffLimit": 0, "activeDeadlineSeconds": args.deadline_seconds,
            "ttlSecondsAfterFinished": 86400,
            "template": {
                "metadata": {"labels": {"app": "storefront-benchmark"}},
                "spec": {
                    "restartPolicy": "Never", "automountServiceAccountToken": False,
                    "securityContext": {"runAsNonRoot": True, "runAsUser": 12345,
                                        "seccompProfile": {"type": "RuntimeDefault"}},
                    "containers": [{
                        "securityContext": {"allowPrivilegeEscalation": False, "capabilities": {"drop": ["ALL"]}},
                        "name": "k6", "image": "grafana/k6:1.6.1",
                        "args": ["run", "--log-format", "raw", f"/scripts/{script.name}"],
                        "env": [{"name": k, "value": v} for k, v in env.items()],
                        "resources": {"requests": {"cpu": "250m", "memory": "256Mi"},
                                      "limits": {"cpu": "2", "memory": "1Gi"}},
                        "volumeMounts": [{"name": "scripts", "mountPath": "/scripts", "readOnly": True}],
                    }],
                    "volumes": [{"name": "scripts", "projected": {"sources": [{"configMap": {"name": n}} for n, _ in configs]}}],
                },
            },
        },
    }
    run("create", "-f", "-", input=json.dumps(job), capture_output=True)
    metadata = {"job": name, "startedAt": datetime.now(timezone.utc).isoformat(),
                "mode": args.mode, "workload": args.workload, "parameters": env,
                "scriptSha256": hashlib.sha256(script.read_bytes()).hexdigest(),
                "scriptsSha256": {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in script_paths},
                "datasetSha256": hashlib.sha256(args.dataset.read_bytes()).hexdigest() if args.mode != "seed" else None,
                "target": "Istio Gateway → web → domain gRPC services" if args.workload != "inventory" else "Inventory gRPC through Istio waypoint → PostgreSQL",
                "mix": {"browse": 0.45, "search": 0.35, "detail": 0.20} if args.workload != "inventory" else {"GetStock": 1}}
    (args.output / "run.json").write_text(json.dumps(metadata, indent=2) + "\n")
    print(f"Running {name}", flush=True)
    deadline = time.monotonic() + args.deadline_seconds + 60
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
    summary = None
    result = re.search(r"BENCHMARK_SUMMARY=(\{[^\n]+\})", logs)
    if result:
        summary = json.loads(result[1])
        (args.output / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
    if args.mode == "seed":
        match = re.search(r"DATASET=(\{[^\n]+\})", logs)
        if match:
            args.dataset.parent.mkdir(parents=True, exist_ok=True)
            args.dataset.write_text(json.dumps(json.loads(match[1]), indent=2) + "\n")
        elif args.workload == "mixed":
            meta = re.search(r"DATASET_META=(\{[^\n]+\})", logs)
            products = [json.loads(line) for line in re.findall(r"DATASET_PRODUCT=(\{[^\n]+\})", logs)]
            if meta and products:
                fixture = {**json.loads(meta[1]), "products": products}
                args.dataset.parent.mkdir(parents=True, exist_ok=True)
                args.dataset.write_text(json.dumps(fixture, indent=2) + "\n")
                metadata["seededProducts"] = len(products)
                if len(products) != args.products or len({p["id"] for p in products}) != args.products:
                    status["succeeded"] = 0
            else:
                status["succeeded"] = 0
        elif status.get("succeeded"):
            raise SystemExit(f"Seed produced no dataset; inspect {args.output / 'k6.log'}")
    metadata["finishedAt"] = datetime.now(timezone.utc).isoformat()
    metadata["succeeded"] = bool(status.get("succeeded"))
    if args.workload == "mixed" and args.mode == "load" and args.write_rate:
        metrics = summary.get("metrics", {}) if summary else {}
        accepted = metrics.get("writes_succeeded", {}).get("values", {}).get("count", 0)
        visibility = metrics.get("projection_visible", {}).get("values", {})
        observed = visibility.get("passes", 0) + visibility.get("fails", 0)
        metadata["projectionCoverage"] = {"acceptedWrites": accepted, "observations": observed}
        if not accepted or accepted != observed:
            metadata["succeeded"] = False
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
    run("delete", "configmap", *[n for n, _ in configs], capture_output=True)
    if not metadata["succeeded"]:
        raise SystemExit(f"Benchmark failed; inspect {args.output / 'k6.log'}")
    print(f"Results: {args.output}", flush=True)


if __name__ == "__main__":
    main()

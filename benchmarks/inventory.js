import grpc from "k6/net/grpc";
import execution from "k6/execution";
import { check } from "k6";

const client = new grpc.Client();
client.load(["/scripts"], "inventory.proto");
const dataset = JSON.parse(open("/scripts/dataset.json"));
let connected = false;

export const options = {
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
  scenarios: {
    inventory: {
      executor: "constant-arrival-rate",
      rate: Number(__ENV.RATE || 100),
      timeUnit: "1s",
      duration: __ENV.DURATION || "2m",
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    checks: ["rate==1"],
    grpc_req_duration: ["p(95)<100"],
    dropped_iterations: ["count==0"],
  },
};

export default function () {
  if (!connected) {
    client.connect("inventory.ecommerce.svc.cluster.local:9097", { plaintext: true });
    connected = true;
  }
  const product = dataset.products[execution.scenario.iterationInTest % dataset.products.length];
  const response = client.invoke("inventory.v1.InventoryService/GetStock", {
    product_id: product.id,
  });
  check(response, {
    "gRPC OK": (r) => r.status === grpc.StatusOK,
    "correct stock returned": (r) =>
      r.message && r.message.productId === product.id && r.message.availableQty === 100,
  });
}

export function handleSummary(data) {
  return { stdout: `BENCHMARK_SUMMARY=${JSON.stringify(data)}\n` };
}

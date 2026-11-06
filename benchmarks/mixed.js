import http from "k6/http";
import { fail, sleep } from "k6";
import execution from "k6/execution";
import { Counter, Rate, Trend } from "k6/metrics";
import readStorefront from "./storefront.js";
import { dataset } from "./dataset.js";

const seed = __ENV.MODE === "seed";
const writeRate = Number(__ENV.WRITE_RATE || 0);
const base = __ENV.BASE_URL || "http://ecommerce-ingress-istio.ecommerce.svc.cluster.local";
const headers = { Host: __ENV.HOST || "shop-dev.phuchoang.sbs" };
const writeLatency = new Trend("write_latency", true);
const writeErrors = new Rate("write_errors");
const writes = new Counter("writes_succeeded");
const lag = new Trend("projection_lag", true);
const lowerLag = new Trend("projection_lag_lower_bound", true);
const window = new Trend("projection_poll_window", true);
const visible = new Rate("projection_visible");
const timeouts = new Counter("projection_timeouts");
const pollErrors = new Rate("projection_poll_errors");
const readCount = new Counter("storefront_reads");
const scenarios = seed
  ? {
      seed: {
        executor: "shared-iterations",
        exec: "write",
        vus: Number(__ENV.SEED_VUS || 20),
        iterations: Number(__ENV.PRODUCTS || 1000),
        maxDuration: `${Number(__ENV.DEADLINE_SECONDS || 1800) - 120}s`,
      },
    }
  : {
      reads: {
        executor: "constant-arrival-rate",
        exec: "read",
        rate: Number(__ENV.RATE || 50),
        timeUnit: "1s",
        duration: __ENV.DURATION || "10m",
        preAllocatedVUs: Number(__ENV.VUS || 200),
        maxVUs: Number(__ENV.MAX_VUS || 200),
      },
    };
if (!seed && writeRate > 0) {
  scenarios.writes = {
    executor: "constant-arrival-rate",
    exec: "write",
    rate: writeRate,
    timeUnit: "1s",
    duration: __ENV.DURATION || "10m",
    preAllocatedVUs: 40,
    maxVUs: 40,
    gracefulStop: "45s",
  };
}
const thresholds = { dropped_iterations: ["count==0"] };
for (const scenario of Object.keys(scenarios)) {
  thresholds[`dropped_iterations{scenario:${scenario}}`] = ["count==0"];
}
if (!seed) {
  Object.assign(thresholds, {
    storefront_errors: ["rate<0.01"],
    storefront_latency: ["p(95)<500"],
    checks: ["rate==1"],
    "storefront_latency{operation:browse}": ["p(95)<500"],
    "storefront_latency{operation:search}": ["p(95)<500"],
    "storefront_latency{operation:detail}": ["p(95)<500"],
  });
}
if (seed || writeRate > 0) {
  Object.assign(thresholds, {
    write_errors: ["rate==0"],
    projection_visible: ["rate==1"],
    projection_poll_errors: ["rate==0"],
  });
  if (!seed) thresholds.projection_lag = ["p(95)<5000"];
}
export const options = {
  scenarios,
  thresholds,
  setupTimeout: "30s",
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
};

export function setup() {
  if (!seed && writeRate === 0) return {};
  const prefix = seed ? `Benchmark-${Date.now()}` : `${dataset.prefix}-Write-${Date.now()}`;
  const response = http.post(
    `${base}/auth/register`,
    {
      email: `${prefix.toLowerCase()}@example.invalid`,
      password: `Bench-${prefix}-only!`,
    },
    { headers, redirects: 0, tags: { name: "register" } },
  );
  const access = response.cookies["__Host-web_access_token"];
  if (response.status !== 303 || !access?.length) fail(`registration status ${response.status}`);
  if (seed) console.log(`DATASET_META=${JSON.stringify({ prefix })}`);
  return { prefix, cookie: `__Host-web_access_token=${access[0].value}` };
}

export function read() {
  readStorefront();
  readCount.add(1);
}

export function write(auth) {
  const iteration = execution.scenario.iterationInTest;
  const type = ["Laptop", "Phone", "Tablet", "Monitor", "Camera"][iteration % 5];
  const name = `${auth.prefix} ${type} ${iteration}`;
  const response = http.post(
    `${base}/seller/products`,
    {
      name,
      description: "Synthetic refurbished marketplace benchmark product",
      category: ["laptop", "phone", "tablet", "accessory", "camera"][iteration % 5],
      condition: "good",
      price: "199.99",
      initial_stock: "100",
    },
    {
      headers: { ...headers, Cookie: auth.cookie },
      redirects: 0,
      timeout: "10s",
      tags: { name: "create", operation: "write" },
    },
  );
  const acknowledgedAt = Date.now();
  const match = (response.headers.Location || "").match(/\/products\/([^?]+)/);
  const valid = response.status === 303 && !!match;
  writeLatency.add(response.timings.duration);
  writeErrors.add(!valid);
  if (!valid) return;
  writes.add(1);
  const id = match[1];
  if (seed) console.log(`DATASET_PRODUCT=${JSON.stringify({ id, name })}`);
  let lastMiss = 0;
  let found = false;
  let upper = 0;
  while (Date.now() - acknowledgedAt < 30000) {
    const started = Date.now() - acknowledgedAt;
    const remaining = 30000 - started;
    if (remaining <= 0) break;
    const result = http.get(`${base}/products?q=${encodeURIComponent(name)}`, {
      headers,
      timeout: `${Math.min(2000, remaining)}ms`,
      tags: { name: "projection_poll", operation: "projection_poll" },
    });
    pollErrors.add(result.status !== 200);
    upper = Date.now() - acknowledgedAt;
    const containsProduct = result.status === 200 && result.body.includes(`id="product-${id}"`);
    if (containsProduct && upper <= 30000) {
      found = true;
      lag.add(upper);
      lowerLag.add(lastMiss);
      window.add(upper - lastMiss);
      break;
    }
    if (result.status === 200 && !containsProduct) lastMiss = started;
    sleep(0.5);
  }
  visible.add(found);
  timeouts.add(found ? 0 : 1);
  if (!seed)
    console.log(
      `PROJECTION_OBSERVATION=${JSON.stringify({
        id,
        acknowledgedAt,
        visible: found,
        lowerMs: lastMiss,
        upperMs: upper,
        censoredAtMs: found ? null : Date.now() - acknowledgedAt,
      })}`,
    );
}

export function handleSummary(data) {
  return { stdout: `BENCHMARK_SUMMARY=${JSON.stringify(data)}\n` };
}

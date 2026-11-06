import http from "k6/http";
import { check, fail, sleep } from "k6";
import { Trend, Rate } from "k6/metrics";

const seed = __ENV.MODE === "seed";
const base = __ENV.BASE_URL || "http://ecommerce-ingress-istio.ecommerce.svc.cluster.local";
const headers = { Host: __ENV.HOST || "shop-dev.phuchoang.sbs" };
const dataset = seed ? null : JSON.parse(open("/scripts/dataset.json"));
const latency = new Trend("storefront_latency", true);
const errors = new Rate("storefront_errors");

export const options = seed
  ? {
      vus: 1,
      iterations: 1,
      setupTimeout: "15m",
    }
  : {
      scenarios: {
        storefront: {
          executor: "constant-arrival-rate",
          rate: Number(__ENV.RATE || 50),
          timeUnit: "1s",
          duration: __ENV.DURATION || "2m",
          preAllocatedVUs: Number(__ENV.VUS || 50),
          maxVUs: Number(__ENV.MAX_VUS || 200),
        },
      },
      thresholds: {
        storefront_errors: ["rate<0.01"],
        storefront_latency: ["p(95)<500"],
        dropped_iterations: ["count==0"],
        checks: ["rate==1"],
        "storefront_latency{operation:browse}": ["p(95)<500"],
        "storefront_latency{operation:search}": ["p(95)<500"],
        "storefront_latency{operation:detail}": ["p(95)<500"],
      },
    };

export function setup() {
  if (!seed) return;
  const prefix = `Benchmark-${Date.now()}`;
  const registration = http.post(
    `${base}/auth/register`,
    {
      email: `${prefix.toLowerCase()}@example.invalid`,
      password: `Bench-${prefix}-only!`,
    },
    { headers, redirects: 0 },
  );
  if (registration.status !== 303) fail(`registration status ${registration.status}`);
  const access = registration.cookies["__Host-web_access_token"];
  if (!access || !access.length) fail("registration did not issue an access cookie");
  const authenticated = { ...headers, Cookie: `__Host-web_access_token=${access[0].value}` };
  const products = [];
  const count = Number(__ENV.PRODUCTS || 1000);
  for (let i = 0; i < count; i++) {
    const name = `${prefix} ${["Laptop", "Phone", "Tablet", "Monitor", "Camera"][i % 5]} ${i}`;
    const response = http.post(
      `${base}/seller/products`,
      {
        name,
        description: "Synthetic refurbished marketplace benchmark product",
        category: ["laptop", "phone", "tablet", "accessory", "camera"][i % 5],
        condition: "good",
        price: "199.99",
        initial_stock: "100",
      },
      { headers: authenticated, redirects: 0 },
    );
    if (response.status !== 303) fail(`product ${i}: status ${response.status}`);
    const location = response.headers.Location || "";
    const match = location.match(/\/products\/([^?]+)/);
    if (!match) fail(`product ${i}: missing product location`);
    products.push({ id: match[1], name });
  }
  for (const product of [products[0], products[products.length - 1]]) {
    let projected = false;
    for (let attempt = 0; attempt < 120; attempt++) {
      const response = http.get(`${base}/products?q=${encodeURIComponent(product.name)}`, {
        headers,
      });
      if (response.status === 200 && response.body.includes(product.id)) {
        projected = true;
        break;
      }
      sleep(2);
    }
    if (!projected) fail(`catalog projection did not catch up for ${product.id}`);
  }
  console.log(`DATASET=${JSON.stringify({ prefix, products })}`);
}

export default function () {
  if (seed) return;
  const slot = __ITER % 20;
  const product = dataset.products[(__VU * 997 + __ITER) % dataset.products.length];
  let operation, path, expected;
  if (slot < 9) {
    operation = "browse";
    path = "/products";
    expected = dataset.prefix;
  } else if (slot < 16) {
    operation = "search";
    path = `/products?q=${encodeURIComponent(`${dataset.prefix} ${["Laptop", "Phone", "Tablet", "Monitor", "Camera"][__ITER % 5]}`)}`;
    expected = dataset.prefix;
  } else {
    operation = "detail";
    path = `/products/${product.id}`;
    expected = product.name;
  }
  const response = http.get(`${base}${path}`, {
    headers,
    tags: { name: operation, operation },
    timeout: "10s",
  });
  const valid = check(
    response,
    {
      "HTTP 200": (r) => r.status === 200,
      "seeded product rendered": (r) => r.body.includes(expected),
    },
    { operation },
  );
  latency.add(response.timings.duration, { operation });
  errors.add(!valid, { operation });
}

export function handleSummary(data) {
  return { stdout: `BENCHMARK_SUMMARY=${JSON.stringify(data)}\n` };
}

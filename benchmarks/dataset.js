import { SharedArray } from "k6/data";

const metadata = __ENV.MODE === "seed" ? null : JSON.parse(open("/scripts/dataset.json"));
export const dataset = metadata && {
  prefix: metadata.prefix,
  products: new SharedArray("catalog products", () =>
    metadata.parts.flatMap((part) => JSON.parse(open(`/scripts/${part}`))),
  ),
};

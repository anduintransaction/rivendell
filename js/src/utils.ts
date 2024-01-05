import * as yaml from "yaml";

export function prefixMultiline(s: string, pad: string): string {
  return s
    .split("\n")
    .map((i) => `${pad}${i}`)
    .join("\n");
}

export function toK8sYaml(obj: any): string {
  return yaml.stringify(obj, undefined, {
    defaultStringType: "QUOTE_DOUBLE",
  });
}

export function prefixMultiline(s: string, pad: string): string {
  return s
    .split("\n")
    .map((i) => `${pad}${i}`)
    .join("\n");
}

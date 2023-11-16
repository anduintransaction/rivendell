import { SpawnOptions, Subprocess } from "bun";

export function prefixMultiline(s: string, pad: string): string {
  return s
    .split("\n")
    .map((i) => `${pad}${i}`)
    .join("\n");
}

export const KUBECTL_BIN = "kubectl";

export function kubectlRun(
  args: string[],
  opts?: SpawnOptions.OptionsObject<"pipe", "inherit", "inherit">,
): Subprocess<"pipe", "inherit", "inherit"> {
  const child = Bun.spawn([KUBECTL_BIN, ...args], {
    ...opts,
    stdin: "pipe",
    stdout: "inherit",
    stderr: "inherit",
  });
  return child;
}

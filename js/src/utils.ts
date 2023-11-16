import { spawn, SpawnOptionsWithoutStdio } from "child_process";

export function prefixMultiline(s: string, pad: string): string {
  return s
    .split("\n")
    .map((i) => `${pad}${i}`)
    .join("\n");
}

export interface RunOpts extends SpawnOptionsWithoutStdio {
  args: string[];
}

export const KUBECTL_BIN = "kubectl";

export function kubectlRun(opts: RunOpts) {
  const { args, ...rest } = opts;
  const child = spawn(KUBECTL_BIN, args, {
    ...rest,
    stdio: "pipe",
  });
  child.stdout.pipe(process.stdout);
  child.stderr.pipe(process.stderr);
  return child;
}

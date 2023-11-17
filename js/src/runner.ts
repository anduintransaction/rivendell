import { Subprocess } from "bun";
import yaml from "yaml";
import chalk from "chalk";
import { DeployStep, Plan, WaitStep } from "./common";

export interface WaitRunner {
  wait(step: WaitStep): Promise<void>;
}

export interface DeployRunner {
  deploy(step: DeployStep): Promise<void>;
}

export abstract class Runner implements WaitRunner, DeployRunner {
  abstract wait(_: WaitStep): Promise<void>;
  abstract deploy(_: DeployStep): Promise<void>;

  async run(plan: Plan) {
    console.log("# Execution");
    for (const step of plan) {
      switch (step.type) {
        case "wait": {
          const msg =
            `Started to wait for "${step.wait.kind}/${step.wait.name}" in module "${step.module}"`;
          console.log(chalk.yellow(msg));
          await this.wait(step);
          console.log(chalk.green("====> Success"));
          break;
        }

        case "deploy": {
          const msg =
            `Started to deploy "${step.object.kind}/${step.object.metadata?.name}" in module "${step.module}"`;
          console.log(chalk.blue(msg));
          await this.deploy(step);
          console.log(chalk.green("====> Success"));
          break;
        }
      }
    }
  }
}

export class NoopRunner extends Runner {
  wait(_: WaitStep): Promise<void> {
    return Promise.resolve();
  }
  deploy(_: DeployStep): Promise<void> {
    return Promise.resolve();
  }
}

export class KubeRunner extends Runner {
  kubeCtx: string;
  dryRun: boolean;
  namespace: string;

  constructor(
    kubeCtx: string = "",
    namespace: string = "default",
    dryRun: boolean = false,
  ) {
    super();
    this.kubeCtx = kubeCtx;
    this.namespace = namespace;
    this.dryRun = dryRun;
  }

  commonArgs() {
    const args = ["kubectl"];
    if (this.kubeCtx !== "") args.push(`--context=${this.kubeCtx}`);
    if (this.namespace !== "") args.push(`--namespace=${this.namespace}`);
    return args;
  }

  async waitForJob(name: string, timeout: number = 300) {
    const args = [...this.commonArgs(), "wait"];
    args.push(`--timeout=${timeout}s`);

    // try to wait for both condition
    const children: Subprocess[] = [];
    return new Promise<void>((resolve, reject) => {
      children.push(
        Bun.spawn([...args, "--for=condition=complete", `job/${name}`], {
          stdout: "inherit",
          stderr: "inherit",
          onExit(_, exitCode, __, ___) {
            if (exitCode === 0) {
              resolve();
            } else {
              reject(`job "${name}" failed`);
            }
          },
        }),
      );

      children.push(
        Bun.spawn([...args, "--for=condition=failed", `job/${name}`], {
          stdout: "inherit",
          stderr: "inherit",
          onExit(_, exitCode, __, ___) {
            if (exitCode === 0) {
              reject(`job "${name}" failed`);
            } else {
              resolve();
            }
          },
        }),
      );
    }).finally(() => {
      children.forEach((p) => {
        if (p.exitCode === null) p.kill();
      });
    });
  }

  async waitForRollout(name: string, kind: string, timeout: number = 300) {
    const args = [...this.commonArgs(), "rollout", "status"];
    args.push(`--timeout=${timeout}s`);
    args.push(`${kind}/${name}`);
    const child = Bun.spawn(args, { stdout: "inherit" });
    const code = await child.exited;
    if (code !== 0) {
      throw new Error(`exited with code ${code}`);
    }
  }

  async wait(w: WaitStep) {
    if (this.dryRun) return;

    switch (w.wait.kind.toLowerCase()) {
      case "job": {
        await this.waitForJob(w.wait.name, w.wait.timeout);
        break;
      }

      case "deployment":
      case "statefulset": {
        await this.waitForRollout(w.wait.name, w.wait.kind, w.wait.timeout);
        break;
      }

      default: {
        const msg =
          `Dont know how to wait on object kind "${w.wait.kind}. Skipping"`;
        console.log(chalk.magenta(msg));
        break;
      }
    }
  }

  async deleteJob(name: string) {
    const args = this.commonArgs();
    args.push("delete", "jobs", name, "--ignore-not-found");
    await Bun.spawn(args, { stdin: "ignore", stdout: "inherit" }).exited;
  }

  async deploy(step: DeployStep) {
    const args = this.commonArgs();
    if (step.object.kind.toLowerCase() === "job") {
      await this.deleteJob(step.object.metadata?.name!);
    }

    args.push("apply", "-f", "-");
    if (this.dryRun) {
      args.push("--dry-run=server");
    }

    const child = Bun.spawn(args, {
      stdin: "pipe",
      stdout: "inherit",
      stderr: "inherit",
    });
    const manifest = yaml.stringify(step.object);
    const enc = new TextEncoder();
    child.stdin.write(enc.encode(manifest));
    child.stdin.end();
    const code = await child.exited;
    if (code !== 0) {
      throw new Error(`exited with code ${code}`);
    }
  }
}

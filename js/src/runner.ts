import yaml from "yaml";
import chalk from "chalk";
import { DeployStep, Plan, WaitStep } from "./common";
import { ChildProcess, execFileSync, spawn } from "child_process";

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
  namespace: string;

  static KUBECTL_BIN = "kubectl";

  constructor(
    kubeCtx: string = "",
    namespace: string = "default",
  ) {
    super();
    this.kubeCtx = kubeCtx;
    this.namespace = namespace;
  }

  commonArgs() {
    const args = [];
    if (this.kubeCtx !== "") args.push(`--context=${this.kubeCtx}`);
    if (this.namespace !== "") args.push(`--namespace=${this.namespace}`);
    return args;
  }

  async waitForJob(name: string, timeout: number = 300) {
    const args = [...this.commonArgs(), "wait"];
    args.push(`--timeout=${timeout}s`);

    // try to wait for both condition
    const children: ChildProcess[] = [];
    return new Promise<void>((resolve, reject) => {
      const completeArgs = [...args, "--for=condition=complete", `job/${name}`];
      const c1 = spawn(KubeRunner.KUBECTL_BIN, completeArgs, {
        stdio: ["ignore", "inherit", "inherit"],
      });
      c1.on("exit", (exitCode) => {
        if (exitCode === 0) {
          resolve();
        } else {
          reject(`job "${name}" failed`);
        }
      });
      children.push(c1);

      const failedArgs = [...args, "--for=condition=failed", `job/${name}`];
      const c2 = spawn(KubeRunner.KUBECTL_BIN, failedArgs, {
        stdio: ["ignore", "inherit", "inherit"],
      });
      c2.on("exit", (exitCode) => {
        if (exitCode === 0) {
          reject(`job "${name}" failed`);
        } else {
          resolve();
        }
      });
      children.push(c2);
    }).finally(() => {
      children.forEach((p) => {
        if (p.exitCode === null) p.kill();
      });
    });
  }

  waitForRollout(name: string, kind: string, timeout: number = 300) {
    const args = [...this.commonArgs(), "rollout", "status"];
    args.push(`--timeout=${timeout}s`);
    args.push(`${kind}/${name}`);
    execFileSync(KubeRunner.KUBECTL_BIN, args, {
      stdio: ["ignore", "inherit", "inherit"],
    });
  }

  async wait(w: WaitStep) {
    switch (w.wait.kind.toLowerCase()) {
      case "job": {
        await this.waitForJob(w.wait.name, w.wait.timeout);
        break;
      }

      case "deployment":
      case "statefulset": {
        this.waitForRollout(w.wait.name, w.wait.kind, w.wait.timeout);
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

  deleteJob(name: string) {
    const args = this.commonArgs();
    args.push("delete", "jobs", name, "--ignore-not-found");
    execFileSync(KubeRunner.KUBECTL_BIN, args, {
      stdio: ["ignore", "inherit", "inherit"],
    });
  }

  async deploy(step: DeployStep) {
    if (step.object.kind.toLowerCase() === "job") {
      this.deleteJob(step.object.metadata?.name!);
    }

    const args = this.commonArgs();
    args.push("apply", "-f", "-");
    const manifest = yaml.stringify(step.object);
    execFileSync(KubeRunner.KUBECTL_BIN, args, {
      input: manifest,
      stdio: ["pipe", "inherit", "inherit"],
    });
  }
}

export class KubeDryrunRunner extends KubeRunner {
  override async wait(_: WaitStep): Promise<void> {
    return Promise.resolve();
  }

  override async deploy(step: DeployStep) {
    if (step.object.kind.toLowerCase() === "job") {
      // skip
      return;
    }

    const args = this.commonArgs();
    args.push("apply", "--dry-run=server", "-f", "-");
    const manifest = yaml.stringify(step.object);
    execFileSync(KubeRunner.KUBECTL_BIN, args, {
      input: manifest,
      stdio: ["pipe", "inherit", "inherit"],
    });
  }
}

export class KubeDiffRunner extends KubeRunner {
  override async wait(_: WaitStep): Promise<void> {
    return Promise.resolve();
  }

  override async deploy(step: DeployStep) {
    if (step.object.kind.toLowerCase() === "job") {
      // skip
      return;
    }

    const args = this.commonArgs();
    args.push("diff", "-f", "-");
    const manifest = yaml.stringify(step.object);
    execFileSync(KubeRunner.KUBECTL_BIN, args, {
      input: manifest,
      stdio: ["pipe", "inherit", "inherit"],
    });
  }
}

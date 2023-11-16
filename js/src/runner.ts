import { DeployStep, Plan, WaitStep } from "./common";
import { kubectlRun } from "./utils";
import yaml from "yaml";
import chalk from "chalk";

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
            `Started to wait for "${step.wait.kind}/${step.wait.name}" in module {${step.module}}`;
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

export class DryRunnner extends Runner {
  wait(_: WaitStep): Promise<void> {
    return Promise.resolve();
  }

  deploy(step: DeployStep): Promise<void> {
    return new Promise((resolve, reject) => {
      const child = kubectlRun({
        args: ["apply", "-f", "-", "--dry-run=server"],
      });
      const manifest = yaml.stringify(step.object);
      const buf = Buffer.from(manifest);
      child.stdin.write(buf);
      child.stdin.end();
      child.on("exit", (code) => {
        if (code === 0) {
          resolve();
        } else {
          reject(`exit with status ${code}`);
        }
      });
    });
  }
}

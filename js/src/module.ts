import { DeployStep, K8sObject, Plan, Wait, WaitStep } from "./common";
import { Context } from "./context";

export interface SourceGenerator {
  (ctx: Context): K8sObject[];
}

export interface ModuleOpts {
  deps?: string[];
  generator?: SourceGenerator;
  waits?: Wait[];
}

export class Module {
  name: string;
  deps: string[];
  generator: SourceGenerator;
  waits: Wait[];

  constructor(
    name: string,
    opts?: ModuleOpts,
  ) {
    this.name = name;
    this.deps = [...(opts?.deps || [])].sort();
    this.generator = opts?.generator || ((_) => []);
    this.waits = opts?.waits || [];
  }

  toPlan(ctx: Context): Plan {
    const objs = this.generator(ctx);
    const deploys: DeployStep[] = objs.map((obj) => ({
      type: "deploy",
      module: this.name,
      object: obj,
    }));
    const waits: WaitStep[] = this.waits.map((wait) => ({
      type: "wait",
      module: this.name,
      wait: wait,
    }));
    return [...waits, ...deploys];
  }
}

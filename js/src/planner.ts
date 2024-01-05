import chalk from "chalk";
import { DeployStep, Plan, WaitStep } from "./common";
import { Context } from "./context";
import { Module } from "./module";
import { ModuleGraph, Walker } from "./graph";
import { prefixMultiline, toK8sYaml } from "./utils";

export class Planner {
  ctx: Context;

  constructor(ctx: Context) {
    this.ctx = ctx;
  }

  async toPlan(m: Module): Promise<Plan> {
    const objs = await m.generator(this.ctx);
    const deploys: DeployStep[] = objs.map((obj) => ({
      type: "deploy",
      module: m.name,
      object: obj,
    }));
    const waits: WaitStep[] = m.waits.map((wait) => ({
      type: "wait",
      module: m.name,
      wait: wait,
    }));
    return [...waits, ...deploys];
  }

  planFromModules(modules: Module[]) {
    const graph = ModuleGraph.resolve(...modules);
    return this.planFromGraph(graph);
  }

  async planFromGraph(graph: ModuleGraph) {
    const plan: Plan = [];
    for (const item of Walker.bfs(graph)) {
      const subPlan = await this.toPlan(item.m);
      plan.push(...subPlan);
    }
    return plan;
  }

  static showManifests(plan: Plan) {
    const manifests = plan
      .filter((p) => p.type === "deploy")
      .map((p) => toK8sYaml((p as DeployStep).object).trim());
    console.log(manifests.join("\n---\n"));
  }

  static show(plan: Plan, verbose: boolean = false) {
    console.log("# Execution plans:");
    for (const step of plan) {
      switch (step.type) {
        case "deploy": {
          const action = chalk.blue("[DEPLOY]");
          const msg =
            `${step.module} | ${step.object.kind} / ${step.object.metadata?.name}`;
          console.log(`${action} ${msg}`);
          if (verbose) {
            let manifest = prefixMultiline(toK8sYaml(step.object), "  ");
            console.log(chalk.grey(manifest));
          }
          break;
        }
        case "wait": {
          const action = chalk.yellow("[ WAIT ]");
          const timeout = step.wait.timeout || 0;
          const msg =
            `${step.module} | ${step.wait.kind} / ${step.wait.name} (timeout: ${timeout}s)`;
          console.log(`${action} ${msg}`);
          break;
        }
      }
    }
  }
}

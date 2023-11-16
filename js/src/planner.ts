import * as yaml from "yaml";
import chalk, {} from "chalk";
import { Plan } from "./common";
import { Context } from "./context";
import { Module } from "./module";
import { ModuleGraph, Walker } from "./graph";
import { prefixMultiline } from "./utils";

export class Planner {
  ctx: Context;

  constructor(ctx: Context) {
    this.ctx = ctx;
  }

  planFromModules(modules: Module[]): Plan {
    const graph = ModuleGraph.resolve(...modules);
    return this.planFromGraph(graph);
  }

  planFromGraph(graph: ModuleGraph): Plan {
    const plan: Plan = [];
    Walker.bfs(graph, (m: Module) => {
      const subPlan = m.toPlan(this.ctx);
      plan.push(...subPlan);
    });
    return plan;
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
            let manifest = prefixMultiline(yaml.stringify(step.object), "  ");
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

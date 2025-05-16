import { Context, KubeRunner, Module, ModuleGraph, Planner } from "./src";

const sleep = (t: number): Promise<void> =>
  new Promise((resolve) => setTimeout(resolve, t));

const modules: Module[] = [
  new Module("redis", {
    generator: async (_: Context) => {
      console.log("generating redis");
      await sleep(1000);
      return [
        {
          apiVersion: "apps/v1",
          kind: "Deployment",
          metadata: {
            name: "redis",
          },
        },
        {
          apiVersion: "v1",
          kind: "Service",
          metadata: {
            name: "redis",
          },
        },
      ];
    },
  }),

  new Module("foundationdb", {
    generator: async (_: Context) => {
      console.log("generating fdb");
      await sleep(1000);
      return [
        {
          apiVersion: "apps/v1",
          kind: "Deployment",
          metadata: {
            name: "foundationdb",
          },
        },
        {
          apiVersion: "v1",
          kind: "Service",
          metadata: {
            name: "foundationdb",
          },
        },
      ];
    },
  }),

  new Module("wait-stargazer", {
    deps: ["redis", "foundationdb"],
    generator: async (_: Context) => {
      return [
        {
          apiVersion: "batch/v1",
          kind: "Job",
          metadata: {
            name: "wait-stargazer",
          },
        },
      ];
    },
  }),

  new Module("gondor", {
    deps: ["wait-stargazer"],
    waits: [
      {
        kind: "Job",
        name: "wait-stargazer",
      },
    ],
    generator: async (_: Context) => {
      return [
        {
          apiVersion: "apps/v1",
          kind: "Deployment",
          metadata: {
            name: "gondor",
          },
        },
        {
          apiVersion: "v1",
          kind: "Service",
          metadata: {
            name: "gondor",
          },
        },
      ];
    },
  }),

  new Module("gondor-portal", {
    deps: ["wait-stargazer"],
    waits: [
      {
        kind: "Job",
        name: "wait-stargazer",
      },
    ],
    generator: async (_: Context) => {
      return [
        {
          apiVersion: "apps/v1",
          kind: "Deployment",
          metadata: {
            name: "gondor-portal",
          },
        },
        {
          apiVersion: "v1",
          kind: "Service",
          metadata: {
            name: "gondor-portal",
          },
        },
      ];
    },
  }),
];

async function main() {
  const graph = ModuleGraph.resolve(...modules);
  graph.printGraphViz();
  console.log("");

  const planner = new Planner(new Context("local"));
  const plan = await planner.planFromGraph(graph);
  Planner.show(plan);
  console.log("");

  const runner = new KubeRunner();
  runner.run(plan);
}

main();

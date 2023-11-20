import { Context, KubeRunner, Module, ModuleGraph, Planner } from "./src";

const modules: Module[] = [
  new Module("redis", {
    generator: (_: Context) => {
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
    generator: (_: Context) => {
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
    generator: (_: Context) => {
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
    generator: (_: Context) => {
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
    generator: (_: Context) => {
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

const graph = ModuleGraph.resolve(...modules);
graph.show();
console.log("");

const planner = new Planner(new Context());
const plan = planner.planFromGraph(graph);
Planner.show(plan);
console.log("");

const runner = new KubeRunner();
runner.run(plan);

import { Module } from "./module";

export class ModuleGraph {
  modules: Record<string, Module>;
  children: Record<string, string[]>;
  roots: string[];
  leafs: string[];

  constructor() {
    this.modules = {};
    this.children = {};
    this.roots = [];
    this.leafs = [];
  }

  addModule(m: Module) {
    if (m.name in this.modules) {
      throw new Error(`Already have a module defined with name "${m.name}"`);
    }
    this.modules[m.name] = m;
    this.children[m.name] = [];
    if (m.deps.length === 0) {
      this.roots.push(m.name);
    }
  }

  reverseDepsResolve() {
    for (const [_, m] of Object.entries(this.modules)) {
      for (const parent of m.deps) {
        if (!(parent in this.modules)) {
          throw new Error(`Unknown dep "${parent}" of module "${m.name}"`);
        }
        this.children[parent].push(m.name);
      }
    }

    for (const [m, children] of Object.entries(this.children)) {
      if (children.length === 0) {
        this.leafs.push(m);
      }
    }
  }

  cyclicCheck() {
    if (Object.keys(this.modules).length > 0 && this.roots.length === 0) {
      throw new Error("Cyclic detected in graph. No root node exists");
    }

    for (const candidate of this.roots) {
      for (const item of Walker.bfs(this, candidate)) {
        const children = this.children[item.m.name];
        for (const child of children) {
          if (item.visited[child]) {
            throw new Error(
              `cyclic detected in graph. Path started at ${candidate}`,
            );
          }
        }
      }
    }
  }

  show() {
    console.log("# Module is printed in reverse order");
    const printIt = (it: ItWalker) => {
      const pad = " ".repeat(it.depth);
      console.log(`${pad}- ${it.name}`);
    };

    const stack: ItWalker[] = this.leafs.map((m) => ({ name: m, depth: 0 }));
    while (stack.length > 0) {
      const it = stack.pop()!;
      printIt(it);
      const m = this.modules[it.name];
      const parents: ItWalker[] = m.deps.map((d) => ({
        name: d,
        depth: it.depth + 1,
      }));
      stack.push(...parents);
    }
  }

  static resolve(...modules: Module[]): ModuleGraph {
    modules.sort((a, b) => a.name.localeCompare(b.name));
    const graph = new ModuleGraph();
    modules.forEach((m) => graph.addModule(m));
    graph.reverseDepsResolve();
    graph.cyclicCheck();
    return graph;
  }
}

interface ItWalker {
  name: string;
  depth: number;
}

interface WalkerItem {
  m: Module;
  depth: number;
  visited: Record<string, boolean>;
}

export const Walker = {
  *bfs(graph: ModuleGraph, startNode?: string) {
    const queue: ItWalker[] = !!startNode
      ? [{ name: startNode!, depth: 0 }]
      : graph.roots.map((m) => ({ name: m, depth: 0 }));
    const visited: Record<string, boolean> = {};
    for (const m of Object.keys(graph.modules)) {
      visited[m] = false;
    }

    while (queue.length > 0) {
      const it = queue.shift()!;
      if (visited[it.name]) continue;

      visited[it.name] = true;
      const item: WalkerItem = {
        m: graph.modules[it.name],
        depth: it.depth,
        visited: Object.assign({}, visited),
      };
      yield item;

      const nonVisitedChildrens: ItWalker[] = graph
        .children[it.name]
        .filter((m) => !visited[m])
        .map((m) => ({ name: m, depth: it.depth + 1 }));
      queue.push(...nonVisitedChildrens);
    }
  },

  *dfs(
    graph: ModuleGraph,
    startNode?: string,
  ) {
    const stack: ItWalker[] = !!startNode
      ? [{ name: startNode!, depth: 0 }]
      : graph.roots.map((m) => ({ name: m, depth: 0 }));
    const visited: Record<string, boolean> = {};
    for (const m of Object.keys(graph.modules)) {
      visited[m] = false;
    }

    while (stack.length > 0) {
      const it = stack.pop()!;
      if (visited[it.name]) continue;

      visited[it.name] = true;
      const item: WalkerItem = {
        m: graph.modules[it.name],
        depth: it.depth,
        visited: Object.assign({}, visited),
      };
      yield item;

      const nonVisitedChildrens: ItWalker[] = graph
        .children[it.name]
        .filter((m) => !visited[m])
        .map((m) => ({ name: m, depth: it.depth + 1 }));
      stack.push(...nonVisitedChildrens);
    }
  },
};

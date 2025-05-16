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

    // Use a proper cycle detection algorithm with path tracking
    const visited: Record<string, boolean> = {};
    const recStack: Record<string, boolean> = {};

    const checkCycle = (node: string): boolean => {
      // If not visited, mark node as visited
      if (!visited[node]) {
        visited[node] = true;
        recStack[node] = true;

        // Check all children
        for (const child of this.children[node]) {
          // If child is not visited and checking the child results in a cycle
          if (!visited[child] && checkCycle(child)) {
            return true;
          } // If child is in recursion stack, we found a cycle
          else if (recStack[child]) {
            return true;
          }
        }
      }

      // Remove from recursion stack
      recStack[node] = false;
      return false;
    };

    // Check from all roots
    for (const root of this.roots) {
      if (checkCycle(root)) {
        throw new Error(`Cyclic detected in graph. Path started at ${root}`);
      }
    }

    // If there are any unvisited nodes, check them too
    for (const node in this.modules) {
      if (!visited[node] && checkCycle(node)) {
        throw new Error(`Cyclic detected in graph. Path started at ${node}`);
      }
    }
  }

  printGraphViz() {
    const prefix = "\t";
    const logWithPrefix = (s: string) => console.log(`${prefix}${s}`);
    const logEdge = (from: string, to: string) =>
      logWithPrefix(`"${from}" -> "${to}";`);

    console.log("digraph DependencyGraph {");
    logWithPrefix("rankdir=LR;");
    for (const name in this.modules) {
      const m = this.modules[name];
      m.deps.forEach((d) => logEdge(d, name));
    }
    console.log("}");
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
}

interface WalkerItem {
  m: Module;
  visited: Record<string, boolean>;
}

export const Walker = {
  *bfs(graph: ModuleGraph, startNode?: string) {
    // Topological sort implementation using BFS (Kahn's algorithm)
    // 1. Calculate in-degree for each node
    const inDegree: Record<string, number> = {};
    for (const m in graph.modules) {
      inDegree[m] = 0;
    }

    for (const [_, m] of Object.entries(graph.modules)) {
      inDegree[m.name] += m.deps.length;
    }

    // 2. Enqueue nodes with in-degree of 0 (no dependencies)
    const queue: ItWalker[] = !!startNode
      ? [{ name: startNode }]
      : graph.roots.map((m) => ({ name: m }));
    const visited: Record<string, boolean> = {};

    // Initialize visited status for all nodes
    for (const m of Object.keys(graph.modules)) {
      visited[m] = false;
    }

    // 3. Process queue
    while (queue.length > 0) {
      const it = queue.shift()!;
      if (visited[it.name]) continue;

      visited[it.name] = true;
      const item: WalkerItem = {
        m: graph.modules[it.name],
        visited: Object.assign({}, visited),
      };
      yield item;

      // For each child, decrease its in-degree and add to queue if in-degree becomes 0
      for (const child of graph.children[it.name]) {
        inDegree[child]--;
        if (inDegree[child] === 0) {
          queue.push({ name: child });
        }
      }
    }
  },

  *dfs(
    graph: ModuleGraph,
    startNode?: string,
  ) {
    const stack: ItWalker[] = !!startNode
      ? [{ name: startNode }]
      : graph.roots.map((m) => ({ name: m }));
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
        visited: Object.assign({}, visited),
      };
      yield item;

      const nonVisitedChildrens: ItWalker[] = graph
        .children[it.name]
        .filter((m) => !visited[m])
        .map((m) => ({ name: m }));
      stack.push(...nonVisitedChildrens);
    }
  },
};

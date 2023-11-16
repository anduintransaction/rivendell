import { Module } from "./module";
import { ModuleGraph } from "./graph";

describe("graph test", () => {
  test("able to resolve graph", () => {
    const modules: Module[] = [
      new Module("parent"),
      new Module("children", { deps: ["parent"] }),
      new Module("grandchildren", { deps: ["children"] }),
    ];

    const graph = ModuleGraph.resolve(...modules);

    expect(graph.modules).toHaveProperty("parent");
    expect(graph.modules).toHaveProperty("children");
    expect(graph.modules).toHaveProperty("grandchildren");

    expect(graph.children["parent"]).toHaveLength(1);
    expect(graph.children["children"]).toHaveLength(1);
    expect(graph.children["grandchildren"]).toHaveLength(0);

    expect(graph.roots).toHaveLength(1);
    expect(graph.leafs).toHaveLength(1);
  });

  test("should throw existed module", () => {
    const modules: Module[] = [
      new Module("foo"),
      new Module("foo"),
    ];
    expect(() => {
      ModuleGraph.resolve(...modules);
    }).toThrow();
  });

  test("should throw unknown dep", () => {
    const modules: Module[] = [
      new Module("foo"),
      new Module("bar", { deps: ["zoo"] }),
    ];
    expect(() => {
      ModuleGraph.resolve(...modules);
    }).toThrow();
  });

  test("should throw at cyclic [1]", () => {
    const modules: Module[] = [
      new Module("a", { deps: ["c"] }),
      new Module("b", { deps: ["a"] }),
      new Module("c", { deps: ["b"] }),
    ];
    expect(() => {
      ModuleGraph.resolve(...modules);
    }).toThrow();
  });

  test("should throw at cyclic [2]", () => {
    const modules: Module[] = [
      new Module("a"),
      new Module("b", { deps: ["a", "d"] }),
      new Module("c", { deps: ["b"] }),
      new Module("d", { deps: ["c"] }),
    ];
    expect(() => {
      ModuleGraph.resolve(...modules);
    }).toThrow();
  });
});

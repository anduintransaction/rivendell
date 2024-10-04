import * as yaml from "yaml";

export function prefixMultiline(s: string, pad: string): string {
  return s
    .split("\n")
    .map((i) => `${pad}${i}`)
    .join("\n");
}

export function toK8sYaml(obj: any): string {
  return yaml.stringify(obj, undefined, {
    defaultKeyType: "PLAIN",
    defaultStringType: "QUOTE_DOUBLE",
  });
}

function makeOrderedMap<T extends number | string | symbol>(
  data: T[],
): Record<T, number> {
  const results = {} as Record<T, number>;
  data.forEach((v, i) => {
    results[v] = i;
  });
  return results;
}

// Ref: https://github.com/helm/helm/blob/main/pkg/releaseutil/kind_sorter.go#L31
export const K8sKindInstallOrder = makeOrderedMap<string>([
  "priorityclass",
  "namespace",
  "networkpolicy",
  "resourcequota",
  "limitrange",
  "podsecuritypolicy",
  "poddisruptionbudget",
  "serviceaccount",
  "secret",
  "secretlist",
  "configmap",
  "storageclass",
  "persistentvolume",
  "persistentvolumeclaim",
  "customresourcedefinition",
  "clusterrole",
  "clusterrolelist",
  "clusterrolebinding",
  "clusterrolebindinglist",
  "role",
  "rolelist",
  "rolebinding",
  "rolebindinglist",
  "service",
  "daemonset",
  "pod",
  "replicationcontroller",
  "replicaset",
  "deployment",
  "horizontalpodautoscaler",
  "statefulset",
  "job",
  "cronjob",
  "ingressclass",
  "ingress",
  "apiservice",
]);

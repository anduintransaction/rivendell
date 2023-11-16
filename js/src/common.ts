import { TypeMeta } from "@kubernetes-models/base";
import { IObjectMeta } from "@kubernetes-models/apimachinery/apis/meta/v1";

export interface K8sObject extends TypeMeta {
  "metadata"?: IObjectMeta;
}

export interface Wait {
  kind: string;
  name: string;
  timeout?: number;
}

export type DeployStep = {
  type: "deploy";
  module: string;
  object: K8sObject;
};

export type WaitStep = {
  type: "wait";
  module: string;
  wait: Wait;
};

export type Step = DeployStep | WaitStep;
export type Plan = Step[];

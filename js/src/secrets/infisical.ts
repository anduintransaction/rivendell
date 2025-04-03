import { InfisicalSDK } from "@infisical/sdk";
import { SecretProvider } from "../context";

export interface InfisicalOpt {
  siteUrl: string;
  clientId: string;
  clientSecret: string;
  projectId: string;
  debug?: boolean;
  forceEnv?: string;
}

export class InfisicalSecretProvider implements SecretProvider {
  client: InfisicalSDK;
  opts: InfisicalOpt;

  constructor(opts: InfisicalOpt) {
    this.opts = opts;
    this.client = new InfisicalSDK({
      siteUrl: opts.siteUrl,
    });
  }

  async doAuth(): Promise<void> {
    await this.client.auth().universalAuth.login({
      clientId: this.opts.clientId,
      clientSecret: this.opts.clientSecret,
    });
  }

  getTargetEnv(env: string): string {
    return this.opts.forceEnv || env;
  }

  async get(env: string, name: string) {
    const parts = name.split("/");
    const [secretName, ...paths] = parts.reverse();
    const secret = await this.client.secrets().getSecret({
      projectId: this.opts.projectId,
      secretName: secretName,
      environment: this.getTargetEnv(env),
      secretPath: `/${paths.reverse().join("/")}`,
      type: "shared",
      expandSecretReferences: true,
      viewSecretValue: true,
    });
    return {
      name: name,
      value: secret.secretValue,
    };
  }

  async getPrefix(env: string, prefix: string) {
    const secrets = await this.client.secrets().listSecrets({
      projectId: this.opts.projectId,
      environment: this.getTargetEnv(env),
      secretPath: `/${prefix}`,
      expandSecretReferences: true,
      viewSecretValue: true,
      includeImports: false,
    });
    return secrets.secrets.map((s) => ({
      name: s.secretKey,
      value: s.secretValue,
    }));
  }
}

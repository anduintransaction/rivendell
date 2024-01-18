import { ClientSettings, InfisicalClient, LogLevel } from "@infisical/sdk";
import { SecretProvider } from "../context";

export interface InfisicalOpt extends ClientSettings {
  projectId: string;
  debug?: boolean;
  forceEnv?: string;
}

export class InfisicalSecretProvider implements SecretProvider {
  client: InfisicalClient;
  forceEnv: string | undefined;
  projectId: string;

  constructor(opts: InfisicalOpt) {
    const { projectId, debug, forceEnv, ...rest } = opts;
    this.forceEnv = forceEnv;
    this.projectId = projectId;
    this.client = new InfisicalClient({
      ...rest,
      logLevel: debug ? LogLevel.Debug : LogLevel.Warn,
    });
  }

  getTargetEnv(env: string): string {
    return this.forceEnv || env;
  }

  async get(env: string, name: string) {
    const parts = name.split("/");
    const [secretName, ...paths] = parts.reverse();
    const secret = await this.client.getSecret({
      projectId: this.projectId,
      secretName: secretName,
      environment: this.getTargetEnv(env),
      path: `/${paths.reverse().join("/")}`,
      type: "shared",
    });
    return {
      name: name,
      value: secret.secretValue,
    };
  }

  async getPrefix(env: string, prefix: string) {
    const secrets = await this.client.listSecrets({
      projectId: this.projectId,
      environment: this.getTargetEnv(env),
      path: `/${prefix}`,
      attachToProcessEnv: false,
      includeImports: false,
    });
    return secrets.map((s) => ({
      name: s.secretKey,
      value: s.secretValue,
    }));
  }
}

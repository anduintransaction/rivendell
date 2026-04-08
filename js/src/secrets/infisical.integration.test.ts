/**
 * Integration test for InfisicalSecretProvider.
 *
 * Requires the following environment variables to be set:
 *   ANDUIN_INFISICAL_CLIENT_ID
 *   ANDUIN_INFISICAL_CLIENT_SECRET
 *
 * Run with:
 *   bun test src/secrets/infisical.integration.test.ts
 */
import { beforeAll, describe, expect, test } from "bun:test";
import { InfisicalSecretProvider } from "./infisical";

const INFISICAL_SITE_URL = "https://infisical.anduin.center";
const INFISICAL_PROJECT_ID = "655c57eb5a031f5c8aa11a95";
const INFISICAL_ENV = "gondor-minas-tirith";
const INFISICAL_SECRET_NAME = "doris/operatorUser/username";
const INFISICAL_SECRET_PREFIX = "doris/operatorUser";

const {
  ANDUIN_INFISICAL_CLIENT_ID,
  ANDUIN_INFISICAL_CLIENT_SECRET,
} = process.env;

const skip = !ANDUIN_INFISICAL_CLIENT_ID ||
  !ANDUIN_INFISICAL_CLIENT_SECRET;

const describeOrSkip = skip ? describe.skip : describe;

describeOrSkip("InfisicalSecretProvider (integration)", () => {
  let provider: InfisicalSecretProvider;

  beforeAll(async () => {
    provider = new InfisicalSecretProvider({
      siteUrl: INFISICAL_SITE_URL,
      clientId: ANDUIN_INFISICAL_CLIENT_ID!,
      clientSecret: ANDUIN_INFISICAL_CLIENT_SECRET!,
      projectId: INFISICAL_PROJECT_ID,
    });
    await provider.doAuth();
  });

  test("doAuth succeeds and client is authenticated", async () => {
    // If doAuth threw, beforeAll would have failed. A second call should also succeed.
    await expect(provider.doAuth()).resolves.toBeUndefined();
  });

  test("getTargetEnv returns env when forceEnv is not set", () => {
    expect(provider.getTargetEnv("staging")).toBe("staging");
  });

  test("getTargetEnv returns forceEnv when set", () => {
    const forced = new InfisicalSecretProvider({
      siteUrl: INFISICAL_SITE_URL,
      clientId: ANDUIN_INFISICAL_CLIENT_ID!,
      clientSecret: ANDUIN_INFISICAL_CLIENT_SECRET!,
      projectId: INFISICAL_PROJECT_ID,
      forceEnv: "prod",
    });
    expect(forced.getTargetEnv("staging")).toBe("prod");
  });

  test("get returns a secret with the correct name", async () => {
    const result = await provider.get(INFISICAL_ENV, INFISICAL_SECRET_NAME);
    expect(result.name).toBe(INFISICAL_SECRET_NAME);
    expect(typeof result.value).toBe("string");
  });

  test("getPrefix returns an array of secrets", async () => {
    const results = await provider.getPrefix(
      INFISICAL_ENV,
      INFISICAL_SECRET_PREFIX,
    );
    expect(Array.isArray(results)).toBe(true);
    for (const s of results) {
      expect(typeof s.name).toBe("string");
      expect(s.name.length).toBeGreaterThan(0);
    }
  });
});

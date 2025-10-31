// In-memory store for API keys (camelCase to match your UI columns)
import type { APIKey } from "../";

// Deterministic RNG
function mulberry32(seed: number) {
  return function () {
    let t = (seed += 0x6d2b79f5);
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}
const rand = mulberry32(42);

function randInt(min: number, max: number) {
  return Math.floor(rand() * (max - min + 1)) + min;
}
function chance(p: number) {
  return rand() < p;
}
function sample<T>(arr: T[]) {
  return arr[randInt(0, arr.length - 1)];
}

const adjectives = [
  "alpha","bravo","charlie","delta","echo","foxtrot","gamma","omega","nova","prime",
  "rapid","silent","stellar","terra","ultra","vector","zen","spark","quantum","lunar",
];
const nouns = [
  "key","access","token","gateway","switch","bridge","anchor","beacon","signal","pilot",
  "runner","mirror","rocket","engine","circuit","tensor","matrix","nexus","orbit","pulse",
];

function randomName(): string {
  const base = `${sample(adjectives)}-${sample(nouns)}`;
  const suffix = chance(0.6) ? `_${randInt(1, 999)}` : chance(0.5) ? `.v${randInt(1, 9)}` : "";
  const name = base + suffix;
  return name.slice(0, Math.max(3, Math.min(35, name.length)));
}

function randomRawKey(): string {
  const bytes = new Uint8Array(24);
  crypto.getRandomValues(bytes);
  return Array.from(bytes).map((b) => b.toString(16).padStart(2, "0")).join("");
}

export function shortKeyFrom(raw: string) {
  if (raw.length <= 6) return raw;
  return `${raw.slice(0, 3)}...${raw.slice(-3)}`;
}

function dateDaysAgo(days: number) {
  const d = new Date();
  d.setDate(d.getDate() - days);
  return d;
}
function dateBetweenDaysAgo(minDays: number, maxDays: number) {
  return dateDaysAgo(randInt(minDays, maxDays));
}
function iso(d: Date | null) {
  return d ? d.toISOString() : null;
}

export function generateAPIKeys(count: number): APIKey[] {
  const set = new Set<string>();
  const out: APIKey[] = [];

  while (out.length < count) {
    let name = randomName();
    while (set.has(name)) name = `${name}-${randInt(1, 9999)}`;
    set.add(name);

    const createdAt = dateBetweenDaysAgo(0, 540);
    const neverUsed = chance(0.2);
    const recentlyUsed = !neverUsed && chance(0.3);

    const maxAgeDays = Math.max(0, Math.floor((Date.now() - createdAt.getTime()) / 86400000));
    const last_used_at = neverUsed
      ? null
      : recentlyUsed
      ? dateBetweenDaysAgo(0, maxAgeDays)
      : dateBetweenDaysAgo(Math.floor(maxAgeDays / 2), maxAgeDays);

    const disabled = chance(0.15);
    const disabled_at = disabled
      ? dateBetweenDaysAgo(0, Math.max(0, Math.floor(((last_used_at ?? createdAt).getTime() - createdAt.getTime()) / 86400000)))
      : null;

    const raw = randomRawKey();
    out.push({
      name,
      short_key: shortKeyFrom(raw),
      created_at: createdAt.toISOString(),
      last_used_at: iso(last_used_at),
      disabled_at: iso(disabled_at),
    });
  }

  // Edge cases
  if (out[0]) out[0] = { ...out[0], name: "a-b", last_used_at: null };
  if (out[1]) out[1] = { ...out[1], name: "prod.key_v1", disabled_at: out[1].disabled_at ?? new Date().toISOString() };
  if (out[2]) out[2] = { ...out[2], name: "this-is-a-very-long-key-name-version-1", short_key: "aaa...bbb" };

  return out;
}

// Public state that matches your client list shape
export const apiKeys: { api_keys: APIKey[] } = {
  api_keys: generateAPIKeys(150),
};

export function findIndexByName(name: string) {
  return apiKeys.api_keys.findIndex((k) => k.name === name);
}
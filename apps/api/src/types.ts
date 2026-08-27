export interface NameValue {
  name: string;
  value: string;
}
export interface RequestSpec {
  schemaVersion: '1';
  method: string;
  url: { scheme: 'http' | 'https'; host: string; port?: number; path: string; query: NameValue[] };
  headers: NameValue[];
  body: { type: string; content?: unknown };
  timeoutMs: number;
  provenance?: Record<string, string>;
}
export interface Evidence {
  source: string;
  field: string;
  value: string;
}
export interface Finding {
  id: string;
  ruleId: string;
  severity: string;
  category: string;
  title: string;
  summary: string;
  evidence: Evidence[];
  confidence: string;
  nextSteps: string[];
}
export interface Analysis {
  sourceSha256: string;
  requests: RequestSpec[];
  sensitiveValues: unknown[];
  findings: Finding[];
  generatedAtUtc: string;
}
export interface Session {
  id: string;
  createdAt: string;
  analysis: Analysis;
}

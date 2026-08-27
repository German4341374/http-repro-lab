import { afterEach, describe, expect, it } from 'vitest';
import type { FastifyInstance } from 'fastify';
import { buildApp } from '../src/app.js';

const analysis = {
  sourceSha256: 'a'.repeat(64),
  requests: [
    {
      schemaVersion: '1',
      method: 'GET',
      url: { scheme: 'https', host: 'example.invalid', path: '/ok', query: [] },
      headers: [],
      body: { type: 'none' },
      timeoutMs: 1000,
    },
  ],
  sensitiveValues: [],
  findings: [
    {
      id: 'one',
      ruleId: 'HTTP_AUTH_401',
      severity: 'high',
      category: 'authentication',
      title: 'Observed 401',
      summary: 'Synthetic',
      evidence: [],
      confidence: 'exact',
      nextSteps: [],
    },
  ],
  generatedAtUtc: '2026-01-01T00:00:00Z',
};
let app: FastifyInstance | undefined;
afterEach(async () => {
  if (app !== undefined) await app.close();
});

describe('local control plane', () => {
  it('reports health and restrictive security headers', async () => {
    app = await buildApp();
    const response = await app.inject({ method: 'GET', url: '/health' });
    expect(response.statusCode).toBe(200);
    expect(response.json()).toEqual({ status: 'ok' });
    expect(response.headers['content-security-policy']).toContain("default-src 'self'");
    const ready = await app.inject({ method: 'GET', url: '/ready' });
    expect(ready.json()).toEqual({ status: 'ready', storage: 'memory' });
  });
  it('creates and reads an analysis session', async () => {
    app = await buildApp();
    const created = await app.inject({
      method: 'POST',
      url: '/api/v1/analyses',
      payload: analysis,
    });
    expect(created.statusCode).toBe(201);
    const response = await app.inject({
      method: 'GET',
      url: `/api/v1/analyses/${created.json<{ id: string }>().id}`,
    });
    expect(response.statusCode).toBe(200);
    expect(response.json<{ analysis: { requests: unknown[] } }>().analysis.requests).toHaveLength(
      1,
    );
  });
  it('paginates and filters requests', async () => {
    app = await buildApp();
    const created = await app.inject({
      method: 'POST',
      url: '/api/v1/analyses',
      payload: analysis,
    });
    const response = await app.inject({
      method: 'GET',
      url: `/api/v1/analyses/${created.json<{ id: string }>().id}/requests?page=1&pageSize=1&q=example`,
    });
    expect(response.json()).toMatchObject({ total: 1, page: 1, pageSize: 1 });
    const fallback = await app.inject({
      method: 'GET',
      url: `/api/v1/analyses/${created.json<{ id: string }>().id}/requests?page=nope&pageSize=999`,
    });
    expect(fallback.json()).toMatchObject({ page: 1, pageSize: 100 });
  });
  it('filters findings by severity', async () => {
    app = await buildApp();
    const created = await app.inject({
      method: 'POST',
      url: '/api/v1/analyses',
      payload: analysis,
    });
    const response = await app.inject({
      method: 'GET',
      url: `/api/v1/analyses/${created.json<{ id: string }>().id}/findings?severity=high`,
    });
    expect(response.json<{ items: unknown[] }>().items).toHaveLength(1);
  });
  it('returns a uniform error for missing sessions', async () => {
    app = await buildApp();
    const response = await app.inject({ method: 'GET', url: '/api/v1/analyses/missing' });
    expect(response.statusCode).toBe(404);
    expect(response.json()).toMatchObject({ error: { code: 'NOT_FOUND' } });
    const requests = await app.inject({ method: 'GET', url: '/api/v1/analyses/missing/requests' });
    const findings = await app.inject({ method: 'GET', url: '/api/v1/analyses/missing/findings' });
    const events = await app.inject({ method: 'GET', url: '/api/v1/analyses/missing/events' });
    expect([requests.statusCode, findings.statusCode, events.statusCode]).toEqual([404, 404, 404]);
  });
  it('rejects malformed analyses', async () => {
    app = await buildApp();
    const response = await app.inject({
      method: 'POST',
      url: '/api/v1/analyses',
      payload: { requests: [] },
    });
    expect(response.statusCode).toBe(400);
    expect(response.json()).toMatchObject({ error: { code: 'INPUT_INVALID' } });
  });
  it('exports Prometheus text without credentials', async () => {
    app = await buildApp();
    const response = await app.inject({ method: 'GET', url: '/metrics' });
    expect(response.body).toContain('http_repro_sessions 0');
    expect(response.body.toLowerCase()).not.toContain('authorization');
  });

  it('emits a terminal server-sent event for a completed analysis', async () => {
    app = await buildApp();
    const created = await app.inject({
      method: 'POST',
      url: '/api/v1/analyses',
      payload: analysis,
    });
    const response = await app.inject({
      method: 'GET',
      url: `/api/v1/analyses/${created.json<{ id: string }>().id}/events`,
    });
    expect(response.statusCode).toBe(200);
    expect(response.headers['content-type']).toContain('text/event-stream');
    expect(response.body).toContain('analysis.ready');
  });
});

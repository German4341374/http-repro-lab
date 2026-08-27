import cors from '@fastify/cors';
import helmet from '@fastify/helmet';
import Fastify, { type FastifyInstance } from 'fastify';
import { SessionStore } from './store.js';
import type { Analysis } from './types.js';

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null;

function isAnalysis(value: unknown): value is Analysis {
  return (
    isRecord(value) &&
    typeof value.sourceSha256 === 'string' &&
    Array.isArray(value.requests) &&
    Array.isArray(value.findings) &&
    Array.isArray(value.sensitiveValues) &&
    typeof value.generatedAtUtc === 'string'
  );
}

function page(value: unknown, fallback: number, maximum?: number): number {
  const parsed = Number(value ?? fallback);
  if (!Number.isSafeInteger(parsed) || parsed < 1) return fallback;
  return maximum === undefined ? parsed : Math.min(parsed, maximum);
}

export interface AppOptions {
  store?: SessionStore;
  logger?: boolean;
}

export async function buildApp(options: AppOptions = {}): Promise<FastifyInstance> {
  const app = Fastify({
    logger: options.logger ?? false,
    bodyLimit: 10 * 1024 * 1024,
    requestIdHeader: 'x-request-id',
  });
  const store = options.store ?? new SessionStore();
  await app.register(helmet, {
    contentSecurityPolicy: {
      directives: {
        defaultSrc: ["'self'"],
        scriptSrc: ["'self'"],
        styleSrc: ["'self'"],
        objectSrc: ["'none'"],
        frameAncestors: ["'none'"],
      },
    },
  });
  await app.register(cors, { origin: false });

  app.setErrorHandler((error, request, reply) => {
    const normalized = error instanceof Error ? error : new Error('Unknown error');
    const metadata = isRecord(error) ? error : {};
    const errorCode = typeof metadata.code === 'string' ? metadata.code : 'INTERNAL_ERROR';
    const candidateStatus = typeof metadata.statusCode === 'number' ? metadata.statusCode : 500;
    request.log.error({ event: 'request.failed', errorCode }, normalized.message);
    const status = candidateStatus >= 400 ? candidateStatus : 500;
    void reply.status(status).send({
      error: {
        code: status === 500 ? 'INTERNAL_ERROR' : 'REQUEST_INVALID',
        message: status === 500 ? 'Internal server error' : normalized.message,
        requestId: request.id,
      },
    });
  });

  app.get('/health', () => ({ status: 'ok' }));
  app.get('/ready', () => ({ status: 'ready', storage: 'memory' }));
  app.get('/metrics', async (_request, reply) =>
    reply
      .type('text/plain; version=0.0.4')
      .send(
        `# HELP http_repro_sessions Current in-memory analysis sessions\n# TYPE http_repro_sessions gauge\nhttp_repro_sessions ${String(store.size())}\n`,
      ),
  );

  app.post('/api/v1/analyses', async (request, reply) => {
    if (!isAnalysis(request.body))
      return reply.status(400).send({
        error: {
          code: 'INPUT_INVALID',
          message: 'Body must be a valid analysis result',
          requestId: request.id,
        },
      });
    const session = store.create(request.body);
    return reply.status(201).send({ id: session.id, createdAt: session.createdAt });
  });
  app.get<{ Params: { id: string } }>('/api/v1/analyses/:id', async (request, reply) => {
    const session = store.get(request.params.id);
    if (session === undefined)
      return reply.status(404).send({
        error: { code: 'NOT_FOUND', message: 'Analysis not found', requestId: request.id },
      });
    return session;
  });
  app.get<{
    Params: { id: string };
    Querystring: { page?: string; pageSize?: string; q?: string };
  }>('/api/v1/analyses/:id/requests', async (request, reply) => {
    const session = store.get(request.params.id);
    if (session === undefined)
      return reply.status(404).send({
        error: { code: 'NOT_FOUND', message: 'Analysis not found', requestId: request.id },
      });
    const current = page(request.query.page, 1);
    const pageSize = page(request.query.pageSize, 25, 100);
    const query = request.query.q?.toLowerCase() ?? '';
    const filtered = session.analysis.requests.filter((item) =>
      JSON.stringify(item).toLowerCase().includes(query),
    );
    const start = (current - 1) * pageSize;
    return {
      items: filtered.slice(start, start + pageSize),
      page: current,
      pageSize,
      total: filtered.length,
    };
  });
  app.get<{ Params: { id: string }; Querystring: { severity?: string } }>(
    '/api/v1/analyses/:id/findings',
    async (request, reply) => {
      const session = store.get(request.params.id);
      if (session === undefined)
        return reply.status(404).send({
          error: { code: 'NOT_FOUND', message: 'Analysis not found', requestId: request.id },
        });
      const severity = request.query.severity?.toLowerCase();
      return {
        items: session.analysis.findings.filter(
          (item) => severity === undefined || item.severity.toLowerCase() === severity,
        ),
      };
    },
  );
  app.get<{ Params: { id: string } }>('/api/v1/analyses/:id/events', async (request, reply) => {
    if (store.get(request.params.id) === undefined)
      return reply.status(404).send({
        error: { code: 'NOT_FOUND', message: 'Analysis not found', requestId: request.id },
      });
    reply.raw.writeHead(200, {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    });
    reply.raw.write('event: analysis.ready\ndata: {"status":"complete"}\n\n');
    reply.raw.end();
    return reply;
  });
  return app;
}

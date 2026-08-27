import { buildApp } from './app.js';

const host = process.env.HTTP_REPRO_BIND ?? '127.0.0.1';
const port = Number(process.env.HTTP_REPRO_PORT ?? '8080');
if (!Number.isSafeInteger(port) || port < 1 || port > 65535)
  throw new Error('HTTP_REPRO_PORT must be a valid TCP port');
const app = await buildApp({ logger: true });
const shutdown = async (signal: string) => {
  app.log.info({ event: 'server.shutdown', signal }, 'graceful shutdown');
  await app.close();
  process.exit(0);
};
process.on('SIGTERM', () => {
  void shutdown('SIGTERM');
});
process.on('SIGINT', () => {
  void shutdown('SIGINT');
});
await app.listen({ host, port });

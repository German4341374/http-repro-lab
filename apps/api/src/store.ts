import { randomUUID } from 'node:crypto';
import type { Analysis, Session } from './types.js';

export class SessionStore {
  readonly #sessions = new Map<string, Session>();
  create(analysis: Analysis): Session {
    const session = { id: randomUUID(), createdAt: new Date().toISOString(), analysis };
    this.#sessions.set(session.id, session);
    return session;
  }
  get(id: string): Session | undefined {
    return this.#sessions.get(id);
  }
  size(): number {
    return this.#sessions.size;
  }
}

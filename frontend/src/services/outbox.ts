export interface OutboxTransaction {
  id: string;
  timestamp: string;
  topic: string;
  payload: any;
  status: 'PENDING' | 'SYNCING' | 'SYNCED' | 'FAILED';
  error?: string;
  size: number;
}

type Subscriber = (transactions: OutboxTransaction[]) => void;

class TransactionalOutboxManager {
  private queue: OutboxTransaction[] = [];
  private subscribers: Set<Subscriber> = new Set();
  private isProcessing = false;
  private syncHandler: ((tx: OutboxTransaction) => Promise<void>) | null = null;

  constructor() {
    this.loadFromStorage();
  }

  private loadFromStorage() {
    try {
      const stored = localStorage.getItem('erp_transactional_outbox');
      if (stored) {
        this.queue = JSON.parse(stored);
      }
    } catch (e) {
      console.error('Failed to load outbox from localStorage', e);
      this.queue = [];
    }
  }

  private saveToStorage() {
    try {
      localStorage.setItem('erp_transactional_outbox', JSON.stringify(this.queue));
    } catch (e) {
      console.error('Failed to save outbox to localStorage', e);
    }
    this.notify();
  }

  public subscribe(sub: Subscriber): () => void {
    this.subscribers.add(sub);
    sub([...this.queue]);
    return () => {
      this.subscribers.delete(sub);
    };
  }

  private notify() {
    const stateCopy = [...this.queue];
    this.subscribers.forEach(sub => sub(stateCopy));
  }

  public registerSyncHandler(handler: (tx: OutboxTransaction) => Promise<void>) {
    this.syncHandler = handler;
    this.triggerSync();
  }

  public async push(topic: string, payload: any) {
    const id = 'tx_' + Math.random().toString(36).substring(2, 11) + '_' + Date.now();
    const serialized = JSON.stringify(payload);
    
    const transaction: OutboxTransaction = {
      id,
      timestamp: new Date().toISOString(),
      topic,
      payload,
      status: 'PENDING',
      size: serialized.length,
    };

    this.queue.unshift(transaction); // Add to the front of visual log, but we'll process chronologically (FIFO or LIFO, let's process pending ones)
    this.saveToStorage();
    
    // Non-blocking trigger of outbox processing
    this.triggerSync();
    
    return id;
  }

  public async triggerSync() {
    if (this.isProcessing || !this.syncHandler) return;
    this.isProcessing = true;

    try {
      // Find oldest pending transaction (FIFO order)
      while (true) {
        const pending = [...this.queue]
          .reverse()
          .find(tx => tx.status === 'PENDING' || tx.status === 'FAILED');

        if (!pending) break;

        // Mark as SYNCING
        this.updateStatus(pending.id, 'SYNCING');
        
        try {
          await this.syncHandler(pending);
          // Mark as SYNCED
          this.updateStatus(pending.id, 'SYNCED');
        } catch (err: any) {
          console.error(`Outbox sync failed for ${pending.id}:`, err);
          this.updateStatus(pending.id, 'FAILED', err.message || 'Unknown network synchronization error');
          // Break loop on failure to prevent pounding backend during downtime
          break;
        }
      }
    } finally {
      this.isProcessing = false;
    }
  }

  private updateStatus(id: string, status: OutboxTransaction['status'], error?: string) {
    this.queue = this.queue.map(tx => {
      if (tx.id === id) {
        return { ...tx, status, error };
      }
      return tx;
    });
    this.saveToStorage();
  }

  public clearSynced() {
    this.queue = this.queue.filter(tx => tx.status !== 'SYNCED');
    this.saveToStorage();
  }

  public clearAll() {
    this.queue = [];
    this.saveToStorage();
  }

  public getQueue() {
    return [...this.queue];
  }
}

export const outboxManager = new TransactionalOutboxManager();

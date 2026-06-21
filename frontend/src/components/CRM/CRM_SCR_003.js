// File: /src/components/CRM/CRM_SCR_003.js
import { outboxManager } from '../../services/outbox.js';

export default class SalesOrderPipelineController {
  constructor(container, context) {
    this.container = container;
    this.context = context; // Expected to contain: active_selected_legal_entity_id, api, token
    
    // Local state cache buffer
    this.state = {
      active_selected_legal_entity_id: context.active_selected_legal_entity_id || 'default_entity_id',
      sales_orders: [],
      outbox_logs: [],
      selected_order: null,
      is_issuance_modal_open: false,
    };
    
    this.draggedCardId = null;
    this.outboxUnsubscribe = null;
  }

  init() {
    this.setupEventListeners();
    this.loadOutboxTelemetry();
    this.loadPipelineData();
  }

  destroy() {
    if (this.outboxUnsubscribe) {
      this.outboxUnsubscribe();
    }
  }

  // Reactive state mutation triggers UI update in <10ms
  setState(updater) {
    const start = performance.now();
    
    if (typeof updater === 'function') {
      this.state = { ...this.state, ...updater(this.state) };
    } else {
      this.state = { ...this.state, ...updater };
    }
    
    this.render();
    
    const duration = performance.now() - start;
    if (duration > 10) {
      console.warn(`[Performance Alert] UI state update took ${duration.toFixed(2)}ms (KPI ceiling is 10ms)`);
    }
  }

  setupEventListeners() {
    // 1. Native Drag and Drop Handlers for Columns
    const lanes = this.container.querySelectorAll('[data-status]');
    lanes.forEach(lane => {
      lane.addEventListener('dragover', (e) => this.onDragOver(e));
      lane.addEventListener('drop', (e) => this.onDrop(e));
      lane.addEventListener('dragleave', (e) => this.onDragLeave(e));
    });

    // 2. Modal Confirmation Handlers
    const btnCancel = this.container.querySelector('#btn-modal-cancel');
    const btnSubmit = this.container.querySelector('#btn-modal-submit');
    
    btnCancel.addEventListener('click', () => this.toggleModal(false));
    btnSubmit.addEventListener('click', () => this.confirmOrderIssuance());

    // 3. Clear Synced Outbox events
    const btnClearSynced = this.container.querySelector('#btn-clear-synced');
    btnClearSynced.addEventListener('click', () => {
      outboxManager.clearSynced();
    });

    // 4. Force Sync trigger
    const btnForceSync = this.container.querySelector('#btn-force-sync');
    btnForceSync.addEventListener('click', () => {
      outboxManager.triggerSync();
    });
  }

  loadOutboxTelemetry() {
    // Subscribe to outbox changes
    this.outboxUnsubscribe = outboxManager.subscribe((queue) => {
      this.setState({ outbox_logs: queue });
    });

    // Register async background worker sync handler (Decoupled Messaging Core)
    outboxManager.registerSyncHandler(async (tx) => {
      this.updateWorkerStatus('RUNNING');
      
      try {
        if (tx.topic === 'crm.sales.order.confirmed') {
          const orderID = tx.payload.order_id;
          
          // Call API Gateway to confirm sales order
          const url = `${this.context.gatewayUrl || 'http://localhost:8080'}/api/v1/crm/sales-orders/${orderID}`;
          const response = await fetch(url, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${this.context.token}`,
            },
            body: JSON.stringify({ status: 'CONFIRMED' }),
          });

          if (!response.ok) {
            const errBody = await response.text();
            throw new Error(`Server returned status ${response.status}: ${errBody}`);
          }
          
          this.showToast(`Order ${tx.payload.order_number} confirmed successfully!`, 'success');
        }
      } catch (err) {
        this.showToast(`Failed to sync outbox transaction: ${err.message}`, 'error');
        this.updateWorkerStatus('FAILED');
        throw err;
      }
      
      this.updateWorkerStatus('IDLE');
    });
  }

  updateWorkerStatus(status) {
    const statusEl = this.container.querySelector('#worker-status-indicator');
    if (statusEl) {
      statusEl.textContent = status;
      statusEl.className = status === 'RUNNING' 
        ? 'text-yellow-400 animate-pulse' 
        : status === 'FAILED' 
          ? 'text-red-500 font-bold' 
          : 'text-emerald-400';
    }
  }

  async loadPipelineData() {
    // Fallback Mock Data as specified in the blueprint
    const mockOrders = [
      { 
        id: "so_crm_2026_84920", 
        order_number: "SO-2026-009941", 
        customer_id: "cust_amer_902", 
        customer_corporate_name: "Acme North American Distribution Corp", 
        gross_valuation_amount: 425000.00, 
        credit_check_status: "PASSED", 
        priority_level_pill: "HIGH_TRACK",
        status: "STAGED_DRAFT",
        legal_entity_id: "default_entity_id" // Valid tenant
      },
      { 
        id: "so_crm_2026_84945", 
        order_number: "SO-2026-009942", 
        customer_id: "cust_apac_410", 
        customer_corporate_name: "Nippon Advanced Heavy Machinery Fabrication Ltd", 
        gross_valuation_amount: 2500000.00, 
        credit_check_status: "CREDIT_LIMIT_EXCEEDED", 
        priority_level_pill: "CRITICAL_EXEC",
        status: "CREDIT_HOLD_REVIEW",
        legal_entity_id: "default_entity_id" // Valid tenant
      },
      {
        id: "so_mismatched_tenant_test",
        order_number: "SO-MISMATCH-999",
        customer_id: "mismatch_cust",
        customer_corporate_name: "Mismatched Tenant LLC",
        gross_valuation_amount: 100.00,
        credit_check_status: "PASSED",
        priority_level_pill: "LOW_TRACK",
        status: "STAGED_DRAFT",
        legal_entity_id: "wrong_entity_id" // Mismatched tenant row (Will be evicted!)
      }
    ];

    try {
      // 1. Attempt to fetch real orders from CRM service (via gateway)
      const url = `${this.context.gatewayUrl || 'http://localhost:8080'}/api/v1/crm/sales-orders`;
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${this.context.token}`,
        }
      });

      if (response.ok) {
        const data = await response.json();
        if (Array.isArray(data)) {
          // Map backend states to board lanes
          const apiOrders = data.map(o => ({
            id: o.id,
            order_number: o.order_number,
            customer_id: o.customer_id,
            customer_corporate_name: o.customer_id.substring(0, 12) + "...", // placeholder since domain is minimal
            gross_valuation_amount: parseFloat(o.total_gross_value) || 0.0,
            credit_check_status: "PASSED",
            priority_level_pill: "HIGH_TRACK",
            status: o.status === 'CONFIRMED' || o.status === 'SHIPPED' ? 'CONFIRMED_FULFILLING' : o.status === 'CREDIT_HOLD' ? 'CREDIT_HOLD_REVIEW' : 'STAGED_DRAFT',
            legal_entity_id: o.legal_entity_id
          }));
          
          this.setState({ sales_orders: [...apiOrders, ...mockOrders] });
          return;
        }
      }
    } catch (e) {
      console.warn("Failed to load real CRM data, falling back to mock blueprint data:", e);
    }

    // Default fallback
    this.setState({ sales_orders: mockOrders });
  }

  // --- HTML5 Native Drag & Drop Handlers ---
  onDragStart(e, id) {
    this.draggedCardId = id;
    e.dataTransfer.setData('text/plain', id);
    e.currentTarget.classList.add('opacity-50', 'rotate-1');
  }

  onDragEnd(e) {
    e.currentTarget.classList.remove('opacity-50', 'rotate-1');
  }

  onDragOver(e) {
    e.preventDefault();
    const lane = e.currentTarget;
    lane.classList.add('bg-slate-800/40', 'border-blue-500/50');
  }

  onDragLeave(e) {
    const lane = e.currentTarget;
    lane.classList.remove('bg-slate-800/40', 'border-blue-500/50');
  }

  onDrop(e) {
    e.preventDefault();
    const lane = e.currentTarget;
    lane.classList.remove('bg-slate-800/40', 'border-blue-500/50');
    
    const targetStatus = lane.dataset.status;
    const orderId = e.dataTransfer.getData('text/plain') || this.draggedCardId;

    if (!orderId) return;

    const order = this.state.sales_orders.find(o => o.id === orderId);
    if (!order) return;

    // Transition Rule Invariants: Block confirm/moving if credit limit is exceeded
    if (targetStatus === 'CONFIRMED_FULFILLING' && order.credit_check_status === 'CREDIT_LIMIT_EXCEEDED') {
      this.showToast('Transition Blocked: Credit Check limit exceeded. Request audit approval.', 'error');
      return;
    }

    // Trigger Outbox dispatch when dropping in "CONFIRMED_FULFILLING"
    if (targetStatus === 'CONFIRMED_FULFILLING' && order.status !== 'CONFIRMED_FULFILLING') {
      this.setState({ selected_order: order });
      this.toggleModal(true);
      return;
    }

    // Mutate state locally instantly (<10ms)
    this.updateOrderLocally(orderId, targetStatus);
  }

  updateOrderLocally(orderId, status) {
    this.setState(prev => ({
      sales_orders: prev.sales_orders.map(o => {
        if (o.id === orderId) {
          return { ...o, status };
        }
        return o;
      })
    }));
  }

  toggleModal(isOpen) {
    const modal = this.container.querySelector('#modal-confirm-order');
    const orderNumberEl = this.container.querySelector('#modal-order-number');

    if (isOpen && this.state.selected_order) {
      orderNumberEl.textContent = this.state.selected_order.order_number;
      modal.classList.remove('hidden');
      this.state.is_issuance_modal_open = true;
    } else {
      modal.classList.add('hidden');
      this.state.is_issuance_modal_open = false;
    }
  }

  // --- Asynchronous Event Outbox emission (Non-blocking) ---
  async confirmOrderIssuance() {
    const order = this.state.selected_order;
    if (!order) return;

    this.toggleModal(false);

    // 1. Serialize payload & add straight to outbox buffer (non-blocking)
    const payload = {
      order_id: order.id,
      order_number: order.order_number,
      customer_id: order.customer_id,
      total_amount: order.gross_valuation_amount,
    };

    // Push is persistent and triggers background worker automatically
    await outboxManager.push('crm.sales.order.confirmed', payload);
    
    // 2. Immediately update local client state viewport cache (Unblocked UI thread)
    this.updateOrderLocally(order.id, 'CONFIRMED_FULFILLING');
    this.showToast(`Event crm.sales.order.confirmed added to outbox for SO ${order.order_number}`, 'info');
  }

  showToast(message, type = 'info') {
    const toastContainer = this.container.querySelector('#toast-container');
    if (!toastContainer) return;

    const toast = document.createElement('div');
    toast.className = `p-4 rounded-md text-xs font-bold shadow-lg border transition-all-normal ${
      type === 'success' 
        ? 'bg-emerald-950/80 border-emerald-500 text-emerald-400' 
        : type === 'error'
          ? 'bg-red-950/80 border-red-500 text-red-400 animate-bounce'
          : 'bg-slate-800 border-slate-700 text-blue-400'
    }`;
    toast.textContent = message;

    toastContainer.appendChild(toast);
    setTimeout(() => {
      toast.classList.add('opacity-0');
      setTimeout(() => toast.remove(), 300);
    }, 4000);
  }

  // --- Rendering Shell Engine (Strict claim shield isolation & stats update) ---
  render() {
    const tenantID = this.state.active_selected_legal_entity_id;
    this.container.querySelector('#active-tenant-id').textContent = tenantID;

    // Filter pipeline by Tenant Isolation Shield
    const filteredOrders = this.state.sales_orders.filter(order => {
      // RULE 3 (The Claim Shield): Verify active selected legal entity. Evict row immediately from layout if mismatched.
      return order.legal_entity_id === tenantID;
    });

    // Update KPI metrics
    const backlogCount = filteredOrders.filter(o => o.status !== 'CONFIRMED_FULFILLING').length;
    const valuation = filteredOrders.reduce((sum, o) => sum + o.gross_valuation_amount, 0);
    const pendingOutbox = this.state.outbox_logs.filter(t => t.status === 'PENDING' || t.status === 'FAILED').length;

    this.container.querySelector('#metric-backlog-count').textContent = backlogCount;
    this.container.querySelector('#metric-pipeline-valuation').textContent = `$${valuation.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
    this.container.querySelector('#metric-outbox-pending').textContent = pendingOutbox;

    // Categorize lanes
    const stagedLane = this.container.querySelector('#lane-staged-draft');
    const creditLane = this.container.querySelector('#lane-credit-hold');
    const confirmedLane = this.container.querySelector('#lane-confirmed-fulfilling');

    stagedLane.innerHTML = '';
    creditLane.innerHTML = '';
    confirmedLane.innerHTML = '';

    let stagedCount = 0;
    let creditCount = 0;
    let confirmedCount = 0;

    filteredOrders.forEach(order => {
      const card = this.createCardDOM(order);

      if (order.status === 'STAGED_DRAFT') {
        stagedLane.appendChild(card);
        stagedCount++;
      } else if (order.status === 'CREDIT_HOLD_REVIEW') {
        creditLane.appendChild(card);
        creditCount++;
      } else if (order.status === 'CONFIRMED_FULFILLING') {
        confirmedLane.appendChild(card);
        confirmedCount++;
      }
    });

    this.container.querySelector('#count-staged-draft').textContent = stagedCount;
    this.container.querySelector('#count-credit-hold').textContent = creditCount;
    this.container.querySelector('#count-confirmed-fulfilling').textContent = confirmedCount;

    // Render Outbox Telemetry Feed
    const feed = this.container.querySelector('#outbox-logs-feed');
    feed.innerHTML = '';

    if (this.state.outbox_logs.length === 0) {
      feed.innerHTML = `<div class="text-slate-500 italic text-center py-4">No outbox events queued</div>`;
    } else {
      this.state.outbox_logs.forEach(tx => {
        const item = document.createElement('div');
        item.className = `p-2 rounded border text-[10px] flex flex-col gap-1 ${
          tx.status === 'SYNCED' 
            ? 'bg-slate-900/40 border-emerald-500/30 text-slate-400' 
            : tx.status === 'SYNCING'
              ? 'bg-blue-950/20 border-blue-500/40 text-blue-300'
              : tx.status === 'FAILED'
                ? 'bg-red-950/40 border-red-500/50 text-red-300'
                : 'bg-slate-900/80 border-slate-700 text-slate-300'
        }`;

        item.innerHTML = `
          <div class="flex items-center justify-between font-bold">
            <span class="truncate pr-2">${tx.topic}</span>
            <span class="px-1.5 py-0.5 rounded text-[8px] uppercase ${
              tx.status === 'SYNCED' ? 'bg-emerald-950 text-emerald-400' :
              tx.status === 'SYNCING' ? 'bg-blue-950 text-blue-400 animate-pulse' :
              tx.status === 'FAILED' ? 'bg-red-950 text-red-400' : 'bg-slate-800 text-slate-400'
            }">${tx.status}</span>
          </div>
          <div class="flex items-center justify-between text-[9px] text-slate-500 mt-1">
            <span>ID: ${tx.id.substring(3, 10)}...</span>
            <span>Size: ${tx.size} bytes</span>
          </div>
          ${tx.error ? `<div class="text-[9px] text-red-400 mt-1 border-t border-red-950/40 pt-1">Error: ${tx.error}</div>` : ''}
        `;
        feed.appendChild(item);
      });
    }
  }

  createCardDOM(order) {
    const card = document.createElement('div');
    card.id = order.id;
    card.className = `p-4 rounded bg-slate-elevated border transition-all-fast flex flex-col gap-3 relative overflow-hidden select-none cursor-grab ${
      order.credit_check_status === 'CREDIT_LIMIT_EXCEEDED' 
        ? 'border-red-500/30 hover:border-red-500 shadow-[0_0_10px_rgba(239,68,68,0.05)]' 
        : 'border-slate-700 hover:border-blue-500'
    }`;
    
    // Make card draggable
    card.setAttribute('draggable', 'true');
    card.addEventListener('dragstart', (e) => this.onDragStart(e, order.id));
    card.addEventListener('dragend', (e) => this.onDragEnd(e));
    
    const pillColor = order.priority_level_pill === 'CRITICAL_EXEC' 
      ? 'bg-red-950/40 text-red-400 border border-red-900/50' 
      : 'bg-emerald-950/40 text-emerald-400 border border-emerald-900/50';

    card.innerHTML = `
      <div class="flex items-center justify-between">
        <span class="text-xs font-mono font-bold text-slate-400">${order.order_number}</span>
        <span class="px-2 py-0.5 text-[9px] font-bold tracking-wider rounded uppercase ${pillColor}">${order.priority_level_pill}</span>
      </div>
      
      <div>
        <h4 class="text-xs font-bold text-slate-200 line-clamp-1">${order.customer_corporate_name}</h4>
        <p class="text-[10px] text-slate-400 mt-1">Valuation: <span class="font-mono font-bold text-glow-blue text-blue-400">$${order.gross_valuation_amount.toLocaleString(undefined, { minimumFractionDigits: 2 })}</span></p>
      </div>

      <div class="flex items-center justify-between pt-2 border-t border-slate-800">
        <span class="text-[9px] font-mono text-slate-500">Credit check:</span>
        <span class="text-[10px] font-bold uppercase ${
          order.credit_check_status === 'PASSED' ? 'text-emerald-400' : 'text-red-400 text-glow-crimson pulse-glow-crimson border border-red-900/50 px-1.5 py-0.5 rounded bg-red-950/30'
        }">${order.credit_check_status === 'PASSED' ? 'PASSED' : 'HOLD'}</span>
      </div>

      <!-- Action buttons -->
      ${order.status === 'STAGED_DRAFT' && order.credit_check_status === 'PASSED' ? `
        <button class="mt-2 w-full py-1.5 bg-blue-600/10 hover:bg-blue-600 text-blue-400 hover:text-white rounded text-[10px] uppercase font-bold tracking-wider transition-all-fast flex items-center justify-center gap-1" data-action="confirm">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
          </svg>
          Confirm Order
        </button>
      ` : ''}
    `;

    // Hook inline button
    const confirmBtn = card.querySelector('[data-action="confirm"]');
    if (confirmBtn) {
      confirmBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.setState({ selected_order: order });
        this.toggleModal(true);
      });
    }

    return card;
  }
}

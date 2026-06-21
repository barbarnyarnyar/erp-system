import React, { useState } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Login } from './components/Login';
import { SliceWrapper } from './components/SliceWrapper';

// @ts-ignore
import crmHtml from './components/CRM/CRM_SCR_003.html?raw';
// @ts-ignore
import CRMController from './components/CRM/CRM_SCR_003.js';

import { 
  Building2, 
  ChevronRight, 
  FolderGit2, 
  KeyRound, 
  Layers, 
  LogOut, 
  Users 
} from 'lucide-react';

const AppContent: React.FC = () => {
  const { token, activeSelectedLegalEntityId, logout, setLegalEntityId, isAuthenticated } = useAuth();
  const [activeScreen, setActiveScreen] = useState('CRM_SCR_003');

  if (!isAuthenticated || !token) {
    return <Login />;
  }

  // Sidebar navigation elements
  const navigation = [
    {
      group: 'Revenue Engine',
      items: [
        { id: 'CRM_SCR_003', name: 'Sales Order Pipeline', code: 'CRM_SCR_003', icon: FolderGit2, disabled: false },
      ]
    },
    {
      group: 'System Backbone',
      items: [
        { id: 'AUTH_SCR_001', name: 'Access Governance', code: 'AUTH_SCR_001', icon: KeyRound, disabled: true },
        { id: 'FM_SCR_001', name: 'Universal Ledger', code: 'FM_SCR_001', icon: Building2, disabled: true },
      ]
    }
  ];

  return (
    <div className="flex h-screen bg-[#080c14] text-slate-100 overflow-hidden font-sans">
      
      {/* Left Sidebar Panel: Navigation & Tenant switcher */}
      <aside className="w-64 bg-[#0f172a] border-r border-slate-800 flex flex-col justify-between shrink-0 glassmorphism">
        
        <div className="flex flex-col overflow-y-auto">
          {/* Logo & Header */}
          <div className="p-6 border-b border-slate-800 flex items-center gap-3">
            <div className="p-2 rounded bg-blue-600/10 border border-blue-500/20 text-blue-400">
              <Layers className="w-5 h-5" />
            </div>
            <div>
              <span className="font-bold text-sm uppercase tracking-widest text-glow-blue text-blue-400">ERP System Shell</span>
              <p className="text-[10px] text-slate-400">Multi-Tenant Platform v1.0</p>
            </div>
          </div>

          {/* Tenant Switcher Hook */}
          <div className="p-4 border-b border-slate-800 bg-slate-900/40">
            <label className="text-[10px] font-bold uppercase tracking-wider text-slate-400 block mb-2">Selected Legal Entity</label>
            <div className="relative">
              <select
                value={activeSelectedLegalEntityId}
                onChange={(e) => setLegalEntityId(e.target.value)}
                className="w-full bg-[#080c14] border border-slate-700 rounded px-3 py-2 text-xs font-mono text-blue-400 focus:outline-none focus:border-blue-500 cursor-pointer appearance-none"
              >
                <option value="default_entity_id">default_entity_id (Active Seeding)</option>
                <option value="wrong_entity_id">wrong_entity_id (Eviction Trigger)</option>
              </select>
              <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-slate-400">
                <ChevronRight className="w-3.5 h-3.5 rotate-90" />
              </div>
            </div>
          </div>

          {/* Navigation Links */}
          <nav className="p-4 flex flex-col gap-6">
            {navigation.map(group => (
              <div key={group.group} className="flex flex-col gap-1.5">
                <span className="text-[9px] font-bold uppercase tracking-widest text-slate-500 px-3">{group.group}</span>
                {group.items.map(item => {
                  const Icon = item.icon;
                  return (
                    <button
                      key={item.id}
                      disabled={item.disabled}
                      onClick={() => setActiveScreen(item.id)}
                      className={`w-full text-left px-3 py-2.5 rounded text-xs font-bold transition-all duration-150 flex items-center gap-3 ${
                        activeScreen === item.id 
                          ? 'bg-blue-600/10 border border-blue-500/20 text-blue-400' 
                          : item.disabled
                            ? 'text-slate-600 cursor-not-allowed opacity-50'
                            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/30 border border-transparent'
                      }`}
                    >
                      <Icon className="w-4 h-4 shrink-0" />
                      <span className="flex-1">{item.name}</span>
                      <span className="text-[9px] font-mono opacity-50">{item.code}</span>
                    </button>
                  );
                })}
              </div>
            ))}
          </nav>
        </div>

        {/* Sidebar Footer Actions */}
        <div className="p-4 border-t border-slate-800 flex flex-col gap-2 bg-slate-900/20">
          <div className="flex items-center gap-3 px-3 py-2">
            <Users className="w-4 h-4 text-slate-500" />
            <div className="truncate">
              <span className="text-xs font-bold block text-slate-300">System Admin</span>
              <span className="text-[10px] text-slate-500 font-mono">admin@erp-shell.io</span>
            </div>
          </div>
          
          <button
            onClick={logout}
            className="w-full py-2.5 px-3 rounded hover:bg-red-950/20 text-red-400 border border-transparent hover:border-red-900/30 text-xs font-bold transition-all duration-150 flex items-center gap-3"
          >
            <LogOut className="w-4 h-4" />
            <span>Sign Out</span>
          </button>
        </div>

      </aside>

      {/* Center Workspace Viewport */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {activeScreen === 'CRM_SCR_003' && (
          <SliceWrapper
            html={crmHtml}
            ControllerClass={CRMController}
            context={{
              active_selected_legal_entity_id: activeSelectedLegalEntityId,
              gatewayUrl: 'http://localhost:8080',
              token: token
            }}
          />
        )}
      </div>

    </div>
  );
};

export default function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { KeyRound, ShieldAlert } from 'lucide-react';

export const Login: React.FC = () => {
  const { login } = useAuth();
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('admin123');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const response = await fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.error || 'Failed to authenticate');
      }

      const data = await response.json();
      if (data.access_token) {
        login(data.access_token);
      } else {
        throw new Error('Access token not found in response');
      }
    } catch (err: any) {
      setError(err.message || 'Connection error. Check gateway status.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-[#080c14] text-slate-100 p-4">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-md p-8 rounded-md bg-[#0f172a] border border-slate-700 glassmorphism shadow-2xl flex flex-col gap-6"
      >
        <div className="flex flex-col items-center gap-2">
          <div className="p-3 rounded-full bg-blue-600/10 text-blue-400 border border-blue-500/20">
            <KeyRound className="w-8 h-8" />
          </div>
          <h2 className="text-xl font-bold tracking-tight uppercase mt-2">Enterprise Access Gateway</h2>
          <p className="text-xs text-slate-400 text-center">Multi-Tenant Identity & Access Governance System</p>
        </div>

        {error && (
          <div className="p-4 rounded bg-red-950/40 border border-red-500/40 text-red-400 text-xs flex items-center gap-3">
            <ShieldAlert className="w-5 h-5 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        <div className="flex flex-col gap-2">
          <label className="text-xs text-slate-400 font-bold uppercase tracking-wider">Username</label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="px-4 py-3 bg-[#080c14] border border-slate-700 rounded-md focus:border-blue-500 focus:outline-none text-sm font-mono"
            required
          />
        </div>

        <div className="flex flex-col gap-2">
          <label className="text-xs text-slate-400 font-bold uppercase tracking-wider">Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="px-4 py-3 bg-[#080c14] border border-slate-700 rounded-md focus:border-blue-500 focus:outline-none text-sm font-mono"
            required
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full py-3 bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 text-white font-bold rounded-md text-xs uppercase tracking-wider transition-all duration-200 shadow-lg"
        >
          {loading ? 'Authenticating...' : 'Sign In'}
        </button>
      </form>
    </div>
  );
};

"use client";

import React, { useState } from "react";
import { Key, Plus } from "lucide-react";

interface KeyItem {
  id: string;
  name: string;
  prefix: string;
  tenant: string;
  rateLimit: number;
  created: string;
  status: string;
}

const initialKeys: KeyItem[] = [
  {
    id: "key-1",
    name: "Production Partner API Key",
    prefix: "ec_live_9f82...",
    tenant: "Enterprise Corp",
    rateLimit: 1000,
    created: "2026-07-28",
    status: "active",
  },
  {
    id: "key-2",
    name: "Mobile App Client Key",
    prefix: "ec_live_3a11...",
    tenant: "Acme Inc",
    rateLimit: 500,
    created: "2026-07-29",
    status: "active",
  },
];

export default function ApiKeysPage() {
  const [keys] = useState<KeyItem[]>(initialKeys);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white">API Keys & Multi-Tenancy</h1>
          <p className="text-sm text-gray-400 mt-1">Generate API access tokens, set rate limit quotas, and manage tenant boundaries.</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2.5 bg-gradient-to-r from-purple-600 to-indigo-600 text-white text-sm font-semibold rounded-xl shadow-lg glow-purple">
          <Plus className="w-4 h-4" />
          Generate New API Key
        </button>
      </div>

      {/* Keys Table */}
      <div className="glass-card rounded-2xl border border-gray-800 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-gray-300">
            <thead className="text-xs uppercase bg-gray-900/80 text-gray-400 border-b border-gray-800">
              <tr>
                <th className="px-6 py-4">Key Identifier</th>
                <th className="px-6 py-4">Key Prefix</th>
                <th className="px-6 py-4">Tenant</th>
                <th className="px-6 py-4">Rate Quota</th>
                <th className="px-6 py-4">Created Date</th>
                <th className="px-6 py-4">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800/60">
              {keys.map((k) => (
                <tr key={k.id} className="hover:bg-gray-800/30 transition-colors">
                  <td className="px-6 py-4 font-semibold text-white">
                    <div className="flex items-center gap-2">
                      <Key className="w-4 h-4 text-purple-400" />
                      {k.name}
                    </div>
                  </td>
                  <td className="px-6 py-4 font-mono text-purple-300 text-xs">{k.prefix}</td>
                  <td className="px-6 py-4 font-medium text-gray-300">{k.tenant}</td>
                  <td className="px-6 py-4 font-mono">{k.rateLimit} req / min</td>
                  <td className="px-6 py-4 text-gray-400 text-xs">{k.created}</td>
                  <td className="px-6 py-4">
                    <span className="px-2.5 py-1 text-xs font-semibold rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      Active
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

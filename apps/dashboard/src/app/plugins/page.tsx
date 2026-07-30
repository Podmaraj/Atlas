"use client";

import React, { useState } from "react";
import { Puzzle, ShieldCheck, Key, Gauge, Database, ArrowRightLeft, Check, Power, LucideIcon } from "lucide-react";

interface PluginItem {
  id: string;
  name: string;
  category: string;
  description: string;
  enabled: boolean;
  scope: string;
  icon: LucideIcon;
}

const initialPlugins: PluginItem[] = [
  {
    id: "plg-1",
    name: "jwt",
    category: "Authentication",
    description: "Validates incoming Bearer JWT tokens, signature verification (HS256/RS256), and extracts tenant claims.",
    enabled: true,
    scope: "Global",
    icon: ShieldCheck,
  },
  {
    id: "plg-2",
    name: "rate-limit",
    category: "Traffic Control",
    description: "Distributed Redis sliding window rate limiting per IP, API key, or tenant quota.",
    enabled: true,
    scope: "Route Level",
    icon: Gauge,
  },
  {
    id: "plg-3",
    name: "api-key",
    category: "Authentication",
    description: "Enforces API Key verification via X-API-Key headers with Redis key hash lookup.",
    enabled: true,
    scope: "Global",
    icon: Key,
  },
  {
    id: "plg-4",
    name: "response-cache",
    category: "Optimization",
    description: "Caches GET request responses in Redis with automatic TTL and Cache-Control parsing.",
    enabled: true,
    scope: "Route Level",
    icon: Database,
  },
  {
    id: "plg-5",
    name: "cors-security",
    category: "Security",
    description: "Applies CORS headers, HSTS, frame options, XSS blocking, and preflight OPTIONS handling.",
    enabled: true,
    scope: "Global",
    icon: ShieldCheck,
  },
  {
    id: "plg-6",
    name: "request-transformer",
    category: "Transformation",
    description: "Dynamic header insertion, header stripping, query param injection, and path rewriter.",
    enabled: false,
    scope: "Service Level",
    icon: ArrowRightLeft,
  },
];

export default function PluginsPage() {
  const [plugins, setPlugins] = useState<PluginItem[]>(initialPlugins);

  const togglePlugin = (id: string) => {
    setPlugins(plugins.map(p => p.id === id ? { ...p, enabled: !p.enabled } : p));
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white">Plugin Catalog</h1>
        <p className="text-sm text-gray-400 mt-1">Enable, disable, and configure Gateway Data Plane plugins with zero downtime.</p>
      </div>

      {/* Plugin Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {plugins.map((plugin) => {
          const Icon = plugin.icon;
          return (
            <div
              key={plugin.id}
              className={`glass-card p-6 rounded-2xl border transition-all ${
                plugin.enabled ? "border-purple-500/40 glow-purple" : "border-gray-800 opacity-60"
              }`}
            >
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className={`p-2.5 rounded-xl ${plugin.enabled ? "bg-purple-500/20 text-purple-400" : "bg-gray-800 text-gray-500"}`}>
                    <Icon className="w-6 h-6" />
                  </div>
                  <div>
                    <h3 className="font-bold text-white text-base">{plugin.name}</h3>
                    <span className="text-xs text-gray-400">{plugin.category}</span>
                  </div>
                </div>

                <button
                  onClick={() => togglePlugin(plugin.id)}
                  className={`p-2 rounded-xl border transition-all ${
                    plugin.enabled
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                      : "bg-gray-800 text-gray-400 border-gray-700"
                  }`}
                  title="Toggle Plugin"
                >
                  <Power className="w-4 h-4" />
                </button>
              </div>

              <p className="text-xs text-gray-300 min-h-[48px] leading-relaxed">{plugin.description}</p>

              <div className="flex items-center justify-between pt-4 mt-4 border-t border-gray-800/80 text-xs">
                <span className="px-2.5 py-1 bg-gray-900 text-gray-400 font-mono rounded-lg border border-gray-800">
                  Scope: {plugin.scope}
                </span>
                <span className={`font-semibold ${plugin.enabled ? "text-emerald-400" : "text-gray-500"}`}>
                  {plugin.enabled ? "Active" : "Disabled"}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

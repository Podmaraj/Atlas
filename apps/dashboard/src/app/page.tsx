"use client";

import React from "react";
import { 
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer 
} from "recharts";
import { 
  Clock, 
  Server, 
  ShieldCheck, 
  Zap, 
  TrendingUp, 
  AlertTriangle 
} from "lucide-react";

const trafficData = [
  { time: "00:00", rps: 12400, latency: 1.2 },
  { time: "04:00", rps: 8900, latency: 1.1 },
  { time: "08:00", rps: 24500, latency: 1.6 },
  { time: "12:00", rps: 41200, latency: 2.1 },
  { time: "16:00", rps: 38900, latency: 1.9 },
  { time: "20:00", rps: 29100, latency: 1.4 },
  { time: "24:00", rps: 18700, latency: 1.3 },
];

const nodes = [
  { id: "node-us-east-1a", ip: "10.0.1.12", status: "online", conns: 14200, cpu: 18.4, mem: 42.1 },
  { id: "node-us-east-1b", ip: "10.0.1.15", status: "online", conns: 15100, cpu: 21.0, mem: 44.8 },
  { id: "node-eu-west-1a", ip: "10.0.2.88", status: "online", conns: 9800, cpu: 14.2, mem: 38.5 },
  { id: "node-ap-south-1a", ip: "10.0.3.41", status: "online", conns: 11300, cpu: 16.9, mem: 40.0 },
];

export default function OverviewPage() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white">System Overview</h1>
        <p className="text-sm text-gray-400 mt-1">Real-time Gateway cluster telemetry and performance analytics.</p>
      </div>

      {/* Metric KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="glass-card p-6 rounded-2xl border border-gray-800 relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-400 font-medium">Requests / Sec</span>
            <div className="p-2 bg-purple-500/10 text-purple-400 rounded-xl">
              <Zap className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-bold text-white font-mono">41,200</div>
            <div className="flex items-center gap-1 text-xs text-emerald-400 mt-2 font-medium">
              <TrendingUp className="w-4 h-4" />
              <span>+18.4% from last hour</span>
            </div>
          </div>
        </div>

        <div className="glass-card p-6 rounded-2xl border border-gray-800">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-400 font-medium">P99 Latency</span>
            <div className="p-2 bg-indigo-500/10 text-indigo-400 rounded-xl">
              <Clock className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-bold text-white font-mono">1.4 ms</div>
            <div className="flex items-center gap-1 text-xs text-emerald-400 mt-2 font-medium">
              <ShieldCheck className="w-4 h-4" />
              <span>Ultra-low overhead</span>
            </div>
          </div>
        </div>

        <div className="glass-card p-6 rounded-2xl border border-gray-800">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-400 font-medium">Active Gateway Nodes</span>
            <div className="p-2 bg-emerald-500/10 text-emerald-400 rounded-xl">
              <Server className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-bold text-white font-mono">4 / 4</div>
            <div className="flex items-center gap-1 text-xs text-emerald-400 mt-2 font-medium">
              <span>100% Healthy</span>
            </div>
          </div>
        </div>

        <div className="glass-card p-6 rounded-2xl border border-gray-800">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-400 font-medium">Global Error Rate</span>
            <div className="p-2 bg-rose-500/10 text-rose-400 rounded-xl">
              <AlertTriangle className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-3xl font-bold text-white font-mono">0.002 %</div>
            <div className="flex items-center gap-1 text-xs text-gray-400 mt-2 font-medium">
              <span>Within SLA limits</span>
            </div>
          </div>
        </div>
      </div>

      {/* Traffic Graph */}
      <div className="glass-card p-6 rounded-2xl border border-gray-800">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-lg font-bold text-white">Live Ingress Traffic (RPS)</h2>
            <p className="text-xs text-gray-400">Total requests routed across all active data plane proxy nodes.</p>
          </div>
        </div>
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={trafficData}>
              <defs>
                <linearGradient id="colorRps" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.4}/>
                  <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0}/>
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" opacity={0.5} />
              <XAxis dataKey="time" stroke="#9ca3af" fontSize={12} />
              <YAxis stroke="#9ca3af" fontSize={12} />
              <Tooltip 
                contentStyle={{ backgroundColor: "#111827", borderColor: "#374151", borderRadius: "12px" }}
                itemStyle={{ color: "#c084fc" }}
              />
              <Area type="monotone" dataKey="rps" stroke="#a855f7" strokeWidth={3} fillOpacity={1} fill="url(#colorRps)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Cluster Nodes List */}
      <div className="glass-card p-6 rounded-2xl border border-gray-800">
        <h2 className="text-lg font-bold text-white mb-4">Data Plane Gateway Cluster Nodes</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-gray-300">
            <thead className="text-xs uppercase bg-gray-900/60 text-gray-400 border-b border-gray-800">
              <tr>
                <th className="px-4 py-3">Node ID</th>
                <th className="px-4 py-3">IP Address</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Active Conns</th>
                <th className="px-4 py-3">CPU Usage</th>
                <th className="px-4 py-3">Memory</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800/60">
              {nodes.map((node) => (
                <tr key={node.id} className="hover:bg-gray-800/30 transition-colors">
                  <td className="px-4 py-3.5 font-mono text-purple-300 font-semibold">{node.id}</td>
                  <td className="px-4 py-3.5 font-mono text-gray-400">{node.ip}</td>
                  <td className="px-4 py-3.5">
                    <span className="px-2.5 py-1 text-xs font-semibold rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      Online
                    </span>
                  </td>
                  <td className="px-4 py-3.5 font-mono">{node.conns.toLocaleString()}</td>
                  <td className="px-4 py-3.5 font-mono">{node.cpu}%</td>
                  <td className="px-4 py-3.5 font-mono">{node.mem}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

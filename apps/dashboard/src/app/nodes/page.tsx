"use client";

import React from "react";
import { Server, Cpu, HardDrive, Wifi, Radio } from "lucide-react";

const nodes = [
  { id: "node-us-east-1a", hostname: "gw-east-1.edgecore.io", ip: "10.0.1.12", version: "v1.25.0", conns: 14200, cpu: 18.4, mem: 42.1, ping: "2s ago" },
  { id: "node-us-east-1b", hostname: "gw-east-2.edgecore.io", ip: "10.0.1.15", version: "v1.25.0", conns: 15100, cpu: 21.0, mem: 44.8, ping: "1s ago" },
  { id: "node-eu-west-1a", hostname: "gw-west-1.edgecore.io", ip: "10.0.2.88", version: "v1.25.0", conns: 9800, cpu: 14.2, mem: 38.5, ping: "3s ago" },
  { id: "node-ap-south-1a", hostname: "gw-south-1.edgecore.io", ip: "10.0.3.41", version: "v1.25.0", conns: 11300, cpu: 16.9, mem: 40.0, ping: "2s ago" },
];

export default function ClusterNodesPage() {
  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-white">Cluster Nodes</h1>
        <p className="text-sm text-gray-400 mt-1">Live status, connections, and system load for Data Plane gateway instances.</p>
      </div>

      {/* Nodes Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {nodes.map((node) => (
          <div key={node.id} className="glass-card p-6 rounded-2xl border border-gray-800 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-emerald-500/10 text-emerald-400 rounded-xl">
                  <Server className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="font-bold text-white text-lg font-mono">{node.id}</h3>
                  <span className="text-xs text-gray-400 font-mono">{node.hostname} ({node.ip})</span>
                </div>
              </div>
              <span className="px-2.5 py-1 text-xs font-semibold rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                Online
              </span>
            </div>

            <div className="grid grid-cols-3 gap-3 pt-2">
              <div className="glass-panel p-3 rounded-xl border border-gray-800 text-xs">
                <span className="text-gray-400 flex items-center gap-1 mb-1">
                  <Wifi className="w-3.5 h-3.5 text-purple-400" />
                  Active Conns
                </span>
                <span className="font-bold font-mono text-white text-sm">{node.conns.toLocaleString()}</span>
              </div>
              <div className="glass-panel p-3 rounded-xl border border-gray-800 text-xs">
                <span className="text-gray-400 flex items-center gap-1 mb-1">
                  <Cpu className="w-3.5 h-3.5 text-indigo-400" />
                  CPU Load
                </span>
                <span className="font-bold font-mono text-white text-sm">{node.cpu}%</span>
              </div>
              <div className="glass-panel p-3 rounded-xl border border-gray-800 text-xs">
                <span className="text-gray-400 flex items-center gap-1 mb-1">
                  <HardDrive className="w-3.5 h-3.5 text-cyan-400" />
                  Memory
                </span>
                <span className="font-bold font-mono text-white text-sm">{node.mem}%</span>
              </div>
            </div>

            <div className="flex items-center justify-between text-xs text-gray-400 pt-2 border-t border-gray-800/80">
              <span>Engine Version: {node.version}</span>
              <span className="flex items-center gap-1">
                <Radio className="w-3.5 h-3.5 text-emerald-400 animate-pulse" />
                Heartbeat: {node.ping}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

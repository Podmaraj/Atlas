"use client";

import React, { useState } from "react";
import { Server, Plus, Heart } from "lucide-react";

interface ServiceItem {
  id: string;
  name: string;
  protocol: string;
  host: string;
  port: number;
  instancesCount: number;
  lbStrategy: string;
  healthCheckPath: string;
  status: string;
}

const initialServices: ServiceItem[] = [
  {
    id: "svc-1",
    name: "user-service",
    protocol: "http",
    host: "user-svc.internal",
    port: 8080,
    instancesCount: 3,
    lbStrategy: "round_robin",
    healthCheckPath: "/health",
    status: "healthy",
  },
  {
    id: "svc-2",
    name: "order-service",
    protocol: "http",
    host: "order-svc.internal",
    port: 8081,
    instancesCount: 2,
    lbStrategy: "least_connections",
    healthCheckPath: "/actuator/health",
    status: "healthy",
  },
  {
    id: "svc-3",
    name: "payment-service",
    protocol: "https",
    host: "payment-svc.internal",
    port: 8443,
    instancesCount: 4,
    lbStrategy: "consistent_hashing",
    healthCheckPath: "/ping",
    status: "healthy",
  },
];

export default function ServicesPage() {
  const [services] = useState<ServiceItem[]>(initialServices);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white">Service Registry</h1>
          <p className="text-sm text-gray-400 mt-1">Manage upstream microservices targets, load balancing algorithms, and health pings.</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2.5 bg-gradient-to-r from-purple-600 to-indigo-600 text-white text-sm font-semibold rounded-xl shadow-lg glow-purple">
          <Plus className="w-4 h-4" />
          Register Service
        </button>
      </div>

      {/* Services Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {services.map((svc) => (
          <div key={svc.id} className="glass-card p-6 rounded-2xl border border-gray-800 space-y-4 hover:border-purple-500/40 transition-all">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2.5 bg-purple-500/10 text-purple-400 rounded-xl">
                  <Server className="w-6 h-6" />
                </div>
                <div>
                  <h3 className="font-bold text-white text-lg">{svc.name}</h3>
                  <span className="text-xs font-mono text-gray-400">{svc.protocol}://{svc.host}:{svc.port}</span>
                </div>
              </div>
              <span className="px-2.5 py-1 text-xs font-semibold rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                Healthy
              </span>
            </div>

            <div className="grid grid-cols-2 gap-3 pt-4 border-t border-gray-800/80 text-xs">
              <div className="glass-panel p-3 rounded-xl border border-gray-800">
                <span className="text-gray-400 block mb-1">Replicas / Instances</span>
                <span className="font-bold font-mono text-white text-base">{svc.instancesCount} Replicas</span>
              </div>
              <div className="glass-panel p-3 rounded-xl border border-gray-800">
                <span className="text-gray-400 block mb-1">LB Algorithm</span>
                <span className="font-mono text-purple-300 font-medium">{svc.lbStrategy}</span>
              </div>
            </div>

            <div className="flex items-center justify-between text-xs text-gray-400 pt-2">
              <span className="flex items-center gap-1">
                <Heart className="w-3.5 h-3.5 text-rose-400" />
                Check: {svc.healthCheckPath}
              </span>
              <span className="text-gray-500 font-mono">Interval: 10s</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

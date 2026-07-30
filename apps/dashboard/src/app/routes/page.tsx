"use client";

import React, { useState } from "react";
import { Plus, Route as RouteIcon, Trash2, ArrowRight } from "lucide-react";

interface RouteItem {
  id: string;
  name: string;
  paths: string[];
  methods: string[];
  service: string;
  priority: number;
  stripPath: boolean;
  status: string;
}

const initialRoutes: RouteItem[] = [
  {
    id: "route-1",
    name: "User Profile API",
    paths: ["/api/v1/user/profile"],
    methods: ["GET", "PUT"],
    service: "user-service",
    priority: 10,
    stripPath: true,
    status: "active",
  },
  {
    id: "route-2",
    name: "Orders Ingress Regex",
    paths: ["^/orders/\\d+$"],
    methods: ["GET", "POST"],
    service: "order-service",
    priority: 20,
    stripPath: false,
    status: "active",
  },
  {
    id: "route-3",
    name: "Payments Webhook Prefix",
    paths: ["/webhooks/payments/*"],
    methods: ["POST"],
    service: "payment-service",
    priority: 15,
    stripPath: true,
    status: "active",
  },
];

export default function RoutesPage() {
  const [routes, setRoutes] = useState<RouteItem[]>(initialRoutes);
  const [showModal, setShowModal] = useState(false);
  const [name, setName] = useState("");
  const [path, setPath] = useState("");
  const [methods, setMethods] = useState("GET, POST");
  const [service, setService] = useState("user-service");
  const [priority, setPriority] = useState(10);

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !path) return;

    const newRt: RouteItem = {
      id: `route-${Date.now()}`,
      name,
      paths: [path],
      methods: methods.split(",").map(m => m.trim().toUpperCase()),
      service,
      priority: Number(priority),
      stripPath: true,
      status: "active",
    };

    setRoutes([newRt, ...routes]);
    setShowModal(false);
    setName("");
    setPath("");
  };

  const handleDelete = (id: string) => {
    setRoutes(routes.filter(r => r.id !== id));
  };

  return (
    <div className="space-y-6">
      {/* Top Action Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white">Route Management</h1>
          <p className="text-sm text-gray-400 mt-1">Configure Ingress matching rules and path transformations for your services.</p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="flex items-center gap-2 px-4 py-2.5 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white text-sm font-semibold rounded-xl transition-all shadow-lg glow-purple"
        >
          <Plus className="w-4 h-4" />
          Create New Route
        </button>
      </div>

      {/* Routes List Table */}
      <div className="glass-card rounded-2xl border border-gray-800 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-gray-300">
            <thead className="text-xs uppercase bg-gray-900/80 text-gray-400 border-b border-gray-800">
              <tr>
                <th className="px-6 py-4">Route Name</th>
                <th className="px-6 py-4">Paths / Expressions</th>
                <th className="px-6 py-4">Methods</th>
                <th className="px-6 py-4">Target Service</th>
                <th className="px-6 py-4">Priority</th>
                <th className="px-6 py-4">Status</th>
                <th className="px-6 py-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-800/60">
              {routes.map((rt) => (
                <tr key={rt.id} className="hover:bg-gray-800/30 transition-colors">
                  <td className="px-6 py-4 font-semibold text-white">
                    <div className="flex items-center gap-2">
                      <RouteIcon className="w-4 h-4 text-purple-400" />
                      {rt.name}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-1">
                      {rt.paths.map((p, idx) => (
                        <span key={idx} className="px-2.5 py-1 text-xs font-mono bg-gray-900 border border-gray-700/60 rounded-md text-purple-300">
                          {p}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-1">
                      {rt.methods.map((m, idx) => (
                        <span key={idx} className="px-2 py-0.5 text-[11px] font-mono font-bold bg-purple-950/80 text-purple-300 border border-purple-800/60 rounded">
                          {m}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="flex items-center gap-1.5 font-mono text-indigo-300 text-xs">
                      <ArrowRight className="w-3.5 h-3.5 text-gray-500" />
                      {rt.service}
                    </span>
                  </td>
                  <td className="px-6 py-4 font-mono font-bold text-gray-200">{rt.priority}</td>
                  <td className="px-6 py-4">
                    <span className="px-2.5 py-1 text-xs font-semibold rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      Active
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => handleDelete(rt.id)}
                      className="p-2 text-gray-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-colors"
                      title="Delete Route"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Modal for Creating Route */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="glass-panel w-full max-w-lg p-6 rounded-2xl border border-gray-800 shadow-2xl space-y-6">
            <h2 className="text-xl font-bold text-white">Create Route Rule</h2>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1">Route Name</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Catalog Ingress"
                  className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-xl text-white text-sm focus:outline-none focus:border-purple-500"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1">Path Pattern (Exact, Prefix, or Regex)</label>
                <input
                  type="text"
                  value={path}
                  onChange={(e) => setPath(e.target.value)}
                  placeholder="e.g. /api/v1/products/*"
                  className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-xl text-white font-mono text-sm focus:outline-none focus:border-purple-500"
                  required
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium text-gray-400 mb-1">HTTP Methods</label>
                  <input
                    type="text"
                    value={methods}
                    onChange={(e) => setMethods(e.target.value)}
                    placeholder="GET, POST"
                    className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-xl text-white text-sm focus:outline-none focus:border-purple-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-400 mb-1">Priority</label>
                  <input
                    type="number"
                    value={priority}
                    onChange={(e) => setPriority(Number(e.target.value))}
                    className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-xl text-white text-sm focus:outline-none focus:border-purple-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-400 mb-1">Target Upstream Service</label>
                <select
                  value={service}
                  onChange={(e) => setService(e.target.value)}
                  className="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded-xl text-white text-sm focus:outline-none focus:border-purple-500"
                >
                  <option value="user-service">user-service (Port 8001)</option>
                  <option value="order-service">order-service (Port 8002)</option>
                  <option value="payment-service">payment-service (Port 8003)</option>
                </select>
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-gray-800">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 text-sm text-gray-400 hover:text-white rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white text-sm font-semibold rounded-xl shadow-lg glow-purple"
                >
                  Save Route
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

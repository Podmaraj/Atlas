"use client";

import "./globals.css";
import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { 
  LayoutDashboard, 
  Route as RouteIcon, 
  Server, 
  Puzzle, 
  Key, 
  Activity, 
  Zap,
  Globe
} from "lucide-react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const queryClient = new QueryClient();

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  const navItems = [
    { name: "Overview", href: "/", icon: LayoutDashboard },
    { name: "Routes", href: "/routes", icon: RouteIcon },
    { name: "Services", href: "/services", icon: Server },
    { name: "Plugins", href: "/plugins", icon: Puzzle },
    { name: "API Keys", href: "/api-keys", icon: Key },
    { name: "Cluster Nodes", href: "/nodes", icon: Activity },
  ];

  return (
    <html lang="en" className="dark">
      <head>
        <title>EdgeCore API Gateway Dashboard</title>
        <meta name="description" content="Enterprise API Gateway Control Plane Dashboard" />
      </head>
      <body className="flex h-screen bg-gray-950 text-gray-100 overflow-hidden">
        <QueryClientProvider client={queryClient}>
          {/* Sidebar */}
          <aside className="w-64 glass-panel border-r border-gray-800 flex flex-col justify-between p-4 z-20">
            <div>
              {/* Brand Header */}
              <div className="flex items-center gap-3 px-3 py-4 mb-6 border-b border-gray-800/60">
                <div className="p-2 bg-gradient-to-tr from-purple-600 to-indigo-500 rounded-xl shadow-lg glow-purple">
                  <Zap className="w-6 h-6 text-white" />
                </div>
                <div>
                  <h1 className="font-bold text-lg tracking-wider text-white">EdgeCore</h1>
                  <p className="text-xs text-purple-400 font-mono font-medium">Enterprise Gateway</p>
                </div>
              </div>

              {/* Navigation List */}
              <nav className="space-y-1">
                {navItems.map((item) => {
                  const Icon = item.icon;
                  const isActive = pathname === item.href;
                  return (
                    <Link
                      key={item.name}
                      href={item.href}
                      className={`flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200 ${
                        isActive
                          ? "bg-purple-600/20 text-purple-300 border border-purple-500/30 glow-purple"
                          : "text-gray-400 hover:text-gray-100 hover:bg-gray-800/50"
                      }`}
                    >
                      <Icon className={`w-5 h-5 ${isActive ? "text-purple-400" : "text-gray-400"}`} />
                      {item.name}
                    </Link>
                  );
                })}
              </nav>
            </div>

            {/* Status Footer */}
            <div className="p-3 glass-card rounded-xl border border-gray-800">
              <div className="flex items-center justify-between text-xs mb-2">
                <span className="text-gray-400">Cluster Status</span>
                <span className="flex items-center gap-1.5 text-emerald-400 font-medium">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  Operational
                </span>
              </div>
              <p className="text-[11px] text-gray-500 font-mono">Control Plane: v1.25.0</p>
            </div>
          </aside>

          {/* Main Content View */}
          <main className="flex-1 flex flex-col min-w-0 overflow-y-auto bg-gradient-to-br from-gray-950 via-gray-900 to-gray-950">
            {/* Top Navigation Bar */}
            <header className="h-16 glass-panel border-b border-gray-800/80 flex items-center justify-between px-8 sticky top-0 z-10">
              <div className="flex items-center gap-4">
                <Globe className="w-5 h-5 text-gray-400" />
                <span className="text-sm font-medium text-gray-300">Environment:</span>
                <span className="px-2.5 py-1 bg-purple-950/80 text-purple-300 text-xs font-semibold rounded-full border border-purple-700/50">
                  Production Cluster
                </span>
              </div>
              <div className="flex items-center gap-3">
                <div className="px-3 py-1.5 glass-card rounded-lg border border-gray-800 flex items-center gap-2 text-xs">
                  <Activity className="w-4 h-4 text-emerald-400" />
                  <span className="text-gray-300">Data Plane Sync: Active</span>
                </div>
              </div>
            </header>

            {/* Page Body */}
            <div className="p-8">{children}</div>
          </main>
        </QueryClientProvider>
      </body>
    </html>
  );
}

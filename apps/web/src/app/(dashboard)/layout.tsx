import Link from 'next/link';
import { ReactNode } from 'react';

// DashboardLayout provides the authenticated shell (sidebar + main content).
export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen grid md:grid-cols-[240px_1fr]">
      <aside className="border-r border-slate-800 bg-slate-950/80 p-4 space-y-4">
        <div>
          <p className="text-indigo-300 font-semibold">ClipSense</p>
          <p className="text-slate-400 text-sm">Pre-editor console</p>
        </div>
        <nav className="space-y-2">
          <Link className="block px-3 py-2 rounded-lg hover:bg-slate-800" href="/dashboard">Dashboard</Link>
          <Link className="block px-3 py-2 rounded-lg hover:bg-slate-800" href="/dashboard/batches/new">New Batch</Link>
          <Link className="block px-3 py-2 rounded-lg hover:bg-slate-800" href="/dashboard/settings">Settings</Link>
        </nav>
      </aside>
      <main className="p-6 space-y-6">{children}</main>
    </div>
  );
}

import './globals.css';
import type { ReactNode } from 'react';

export const metadata = {
  title: 'ClipSense',
  description: 'AI-powered pre-editor for video storytellers'
};

export default function RootLayout({ children }: { children?: ReactNode }) {
  return (
    <html lang="en" className="bg-slate-950 text-slate-100">
      <body className="min-h-screen antialiased font-[Inter]">
        {children}
      </body>
    </html>
  );
}

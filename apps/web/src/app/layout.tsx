import './globals.css';
import { ReactNode } from 'react';

export const metadata = {
  title: 'ClipSense',
  description: 'AI-powered pre-editor for video storytellers'
};

// RootLayout wraps the entire app with html/body shells and global styles.
export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className="bg-slate-950 text-slate-100">
      <body className="min-h-screen antialiased font-[Inter]">
        {children}
      </body>
    </html>
  );
}

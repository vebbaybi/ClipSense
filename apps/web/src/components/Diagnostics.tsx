'use client';
import {useEffect} from 'react';
import {telemetry} from '@/lib/telemetry.mjs';

export function Diagnostics() {
  useEffect(() => {
    const report = () => telemetry('application.error');
    window.addEventListener('error', report);
    window.addEventListener('unhandledrejection', report);
    return () => {
      window.removeEventListener('error', report);
      window.removeEventListener('unhandledrejection', report);
    };
  }, []);
  return null;
}

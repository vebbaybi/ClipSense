'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api/client'
import { useAuthToken } from '@/hooks/useAuthToken'

type Batch = {
  id: string
  name: string
  status: string
  created_at: string
  clip_count: number
  duration_seconds: number
}

export default function DashboardPage() {
  const router = useRouter()
  const { token } = useAuthToken()
  const [batches, setBatches] = useState<Batch[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    if (!token) return
    api<{ batches: Batch[] }>('/api/batches', { method: 'GET' }, token)
      .then(res => setBatches(res.batches))
      .catch(e => setError(e.message || 'Failed to load'))
  }, [token])

  if (!token) {
    router.replace('/?login=1')
    return null
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-slate-400">Overview</p>
          <h1 className="text-2xl font-semibold">Your batches</h1>
        </div>
        <Link className="btn btn-primary" href="/dashboard/batches/new">New batch</Link>
      </div>
      {error && <div className="card text-red-400 text-sm">{error}</div>}
      <div className="grid gap-3">
        {batches.length === 0 && (
          <div className="card text-slate-400">No batches yet. Create one to start.</div>
        )}
        {batches.map(batch => (
          <Link key={batch.id} href={`/dashboard/batches/${batch.id}`} className="card hover:border-indigo-500 transition-colors">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-400">{new Date(batch.created_at).toLocaleString()}</p>
                <p className="text-lg font-semibold">{batch.name}</p>
                <p className="text-slate-400 text-sm">{batch.status}</p>
              </div>
              <div className="text-right text-sm text-slate-300">
                <p>{batch.clip_count || 0} clips</p>
                <p>{batch.duration_seconds ? `${Math.round(batch.duration_seconds/60)} min` : '-'}</p>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}


'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { api } from '@/lib/api/client'
import { useAuthToken } from '@/hooks/useAuthToken'

type Clip = { id: string; filename: string; title: string; summary: string; mood: string; role: string; duration_seconds: number }
type Storyline = { id: string; title: string; type: string; clips: Clip[] }
type Batch = { id: string; name: string; status: string }

export default function BatchPage() {
  const { token } = useAuthToken()
  const router = useRouter()
  const params = useParams()
  const id = params?.id as string
  const [batch, setBatch] = useState<Batch | null>(null)
  const [clips, setClips] = useState<Clip[]>([])
  const [storylines, setStorylines] = useState<Storyline[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    if (!token) return
    api<{ batch: Batch; clips: Clip[]; storylines: Storyline[] }>(`/api/batches/${id}`, { method: 'GET' }, token)
      .then(res => { setBatch(res.batch); setClips(res.clips); setStorylines(res.storylines) })
      .catch(e => setError(e.message || 'Failed to load'))
  }, [token, id])

  if (!token) { router.replace('/?login=1'); return null }
  if (!batch) return <div className="card">Loading… {error}</div>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-slate-400">Batch</p>
          <h1 className="text-2xl font-semibold">{batch.name}</h1>
          <p className="text-slate-400 text-sm">{batch.status} · {clips.length} clips</p>
        </div>
        <a className="btn btn-secondary" href={`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/batches/${batch.id}/export?format=csv`} target="_blank">Export CSV</a>
      </div>

      <section className="grid md:grid-cols-3 gap-4">
        {clips.map((clip) => (
          <div key={clip.id} className="card space-y-2">
            <p className="text-sm text-slate-400">{Math.round(clip.duration_seconds)}s</p>
            <h3 className="text-lg font-semibold">{clip.title || clip.filename}</h3>
            <p className="text-slate-300 text-sm line-clamp-3">{clip.summary}</p>
            <p className="text-xs text-slate-400">Mood: {clip.mood} · Role: {clip.role}</p>
          </div>
        ))}
      </section>

      <section className="space-y-2">
        <h2 className="text-xl font-semibold">Storylines</h2>
        {storylines.length === 0 && <div className="card text-slate-400">No storylines yet</div>}
        {storylines.map((s) => (
          <div key={s.id} className="card space-y-2">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-400">{s.type}</p>
                <h3 className="text-lg font-semibold">{s.title}</h3>
              </div>
              <p className="text-sm text-slate-300">{s.clips.length} beats</p>
            </div>
            <div className="flex flex-wrap gap-2 text-xs text-slate-300">
              {s.clips.map((c, idx) => (
                <span key={c.id} className="px-2 py-1 rounded bg-slate-800/80">{idx + 1}. {c.title || c.filename}</span>
              ))}
            </div>
          </div>
        ))}
      </section>
    </div>
  )
}

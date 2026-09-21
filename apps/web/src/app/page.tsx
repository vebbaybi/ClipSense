'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api/client'
import { useAuthToken } from '@/hooks/useAuthToken'

export default function LandingPage() {
  const router = useRouter()
  const { token, ready, setToken } = useAuthToken()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const submit = async () => {
    setError('')
    try {
      const path = mode === 'login' ? '/api/auth/login' : '/api/auth/register'
      const res = await api<{ token: string }>(path, {
        method: 'POST',
        body: JSON.stringify({ email, password }),
        headers: { 'Content-Type': 'application/json' }
      })
      setToken(res.token)
      router.push('/dashboard')
    } catch (e: any) {
      setError(e.message || 'Failed')
    }
  }

  useEffect(() => {
    if (ready && token) router.replace('/dashboard')
  }, [ready, router, token])

  return (
    <main className="max-w-5xl mx-auto px-6 py-16 space-y-16">
      <header className="space-y-4 text-center">
        <p className="text-indigo-300 font-semibold">ClipSense</p>
        <h1 className="text-4xl md:text-5xl font-bold leading-tight">Turn raw clip chaos into narrative-ready storylines.</h1>
        <p className="text-slate-300 text-lg max-w-3xl mx-auto">Upload a zip of videos, get transcripts, vibes, roles, and AI-generated story arcs in minutes.</p>
      </header>

      <div className="grid md:grid-cols-2 gap-6 items-start">
        <div className="card space-y-3">
          <div className="flex gap-2 text-sm">
            <button className={`btn ${mode === 'login' ? 'btn-primary' : 'btn-secondary'}`} onClick={() => setMode('login')}>Login</button>
            <button className={`btn ${mode === 'register' ? 'btn-primary' : 'btn-secondary'}`} onClick={() => setMode('register')}>Register</button>
          </div>
          <input placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} />
          <input placeholder="Password" type="password" value={password} onChange={e => setPassword(e.target.value)} />
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <button className="btn btn-primary" onClick={submit}>{mode === 'login' ? 'Sign in' : 'Create account'}</button>
        </div>
        <div className="card space-y-3">
          <p className="text-slate-300">Already set up?</p>
          <div className="flex gap-3">
            <Link className="btn btn-secondary" href="/dashboard">Go to dashboard</Link>
          </div>
          <p className="text-xs text-slate-500">API: {process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}</p>
        </div>
      </div>
    </main>
  )
}

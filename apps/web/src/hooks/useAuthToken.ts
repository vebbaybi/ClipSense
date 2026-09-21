'use client'

import { useEffect, useState } from 'react'

const TOKEN_KEY = 'clipsense_token'

export function useAuthToken() {
  const [token, setToken] = useState<string | null>(null)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    const t = localStorage.getItem(TOKEN_KEY)
    if (t) setToken(t)
    setReady(true)
  }, [])

  const save = (t: string) => {
    localStorage.setItem(TOKEN_KEY, t)
    setToken(t)
    setReady(true)
  }

  const clear = () => {
    localStorage.removeItem(TOKEN_KEY)
    setToken(null)
    setReady(true)
  }

  return { token, ready, setToken: save, clear }
}

'use client'

import { useEffect, useState } from 'react'

const TOKEN_KEY = 'clipsense_token'

export function useAuthToken() {
  const [token, setToken] = useState<string | null>(null)

  useEffect(() => {
    const t = localStorage.getItem(TOKEN_KEY)
    if (t) setToken(t)
  }, [])

  const save = (t: string) => {
    localStorage.setItem(TOKEN_KEY, t)
    setToken(t)
  }

  const clear = () => {
    localStorage.removeItem(TOKEN_KEY)
    setToken(null)
  }

  return { token, setToken: save, clear }
}

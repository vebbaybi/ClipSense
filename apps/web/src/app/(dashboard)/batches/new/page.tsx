'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useDropzone } from 'react-dropzone'
import toast, { Toaster } from 'react-hot-toast'
import { useAuthToken } from '@/hooks/useAuthToken'

export default function NewBatchPage() {
  const router = useRouter()
  const { token } = useAuthToken()
  const [uploading, setUploading] = useState(false)

  const onDrop = async (acceptedFiles: File[]) => {
    if (!token) { router.replace('/?login=1'); return }
    if (!acceptedFiles.length) return
    setUploading(true)
    const file = acceptedFiles[0]
    const form = new FormData()
    form.append('file', file)
    form.append('name', file.name.replace(/\.zip$/i, ''))

    const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    const res = await fetch(`${base}/api/batches`, { method: 'POST', body: form, headers: { Authorization: `Bearer ${token}` } })
    if (res.ok) {
      const data = await res.json()
      toast.success('Batch uploaded')
      router.push(`/dashboard/batches/${data.batch.id}`)
    } else {
      toast.error(await res.text())
    }
    setUploading(false)
  }

  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop, accept: { 'application/zip': ['.zip'] } })

  return (
    <div className="space-y-6">
      <Toaster position="bottom-right" />
      <h1 className="text-2xl font-semibold">New batch</h1>
      <div
        {...getRootProps()}
        className={`card border-dashed border-2 ${isDragActive ? 'border-indigo-500 bg-indigo-500/10' : 'border-slate-700'}`}
      >
        <input {...getInputProps()} />
        <p className="text-lg font-semibold">Drop a .zip of clips here</p>
        <p className="text-slate-400 text-sm">We unpack, analyze, and generate storylines automatically.</p>
        <button disabled={uploading} className="btn btn-primary mt-4">{uploading ? 'Uploading…' : 'Select file'}</button>
      </div>
    </div>
  )
}

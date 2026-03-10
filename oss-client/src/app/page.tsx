'use client'
import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export const dynamic = 'force-dynamic'

export default function Home() {
  const router = useRouter()
  useEffect(() => {
    router.push('/projects')
  }, [router])
  return null
}

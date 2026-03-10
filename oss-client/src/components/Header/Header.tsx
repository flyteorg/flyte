'use client'

import React from 'react'

interface HeaderProps {
  logoComponent?: React.ReactNode
  showSearch?: boolean
}

export function Header({ logoComponent }: HeaderProps) {
  return (
    <div
      className={`flex h-[56px] w-full items-center justify-between bg-(--system-gray-1) px-5 py-3`}
    >
      {logoComponent ? logoComponent : <div />}
    </div>
  )
}

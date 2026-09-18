import type { ReactNode } from 'react'

export default function Layout({
  children,
  narrow = false,
}: {
  children: ReactNode
  narrow?: boolean
}) {
  return (
    <div className={`page${narrow ? ' page-narrow' : ''}`}>{children}</div>
  )
}
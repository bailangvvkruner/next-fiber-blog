'use client'

import React from 'react'
import { AppProgressProvider as ProgressProvider } from '@bprogress/next'

interface GlobalProgressProviderProps {
  children: React.ReactNode
}

const GlobalProgressProvider: React.FC<GlobalProgressProviderProps> = (
  {
    children
  }
) => {
  return (
    <ProgressProvider
      height="3px"
      color="#46B952"
      options={{ showSpinner: false }}
      shallowRouting
    >
      {children}
    </ProgressProvider>
  )
}

export default GlobalProgressProvider;
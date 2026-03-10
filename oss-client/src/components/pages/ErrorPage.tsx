import React from 'react'

export const ErrorPage = () => (
  <div className="flex min-h-full w-full items-center justify-center">
    <div className="flex flex-col items-center text-center">
      <h1 className="mb-2 text-2xl font-semibold text-white">Error</h1>
      <p className="text-base font-normal text-white">
        We&apos;re having trouble loading this page
      </p>
    </div>
  </div>
)

export default ErrorPage

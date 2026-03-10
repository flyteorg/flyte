import React, { ChangeEvent, useCallback } from 'react'

const KeyNameInput: React.FC<{
  label: string
  type: string | string[] | undefined
  required?: boolean
  onKeyChange?: (key: string) => void
}> = ({ label, onKeyChange }) => {
  const handleChangeInput = useCallback((ev: ChangeEvent<HTMLInputElement>) => {
    if (onKeyChange && ev.target.value) {
      onKeyChange(ev.target.value)
    }
  }, [])

  return (
    <>
      <input type="text" defaultValue={label} onBlur={handleChangeInput} />
    </>
  )
}

export default KeyNameInput

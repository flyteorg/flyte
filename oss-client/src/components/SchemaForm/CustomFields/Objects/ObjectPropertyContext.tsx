import { createContext, useContext } from 'react'

type ObjectPropertyCtx = {
  parentName: string
  index: number
  depth: number
  objectDescriptionId?: string
}

const ObjectPropertyContext = createContext<ObjectPropertyCtx | null>(null)

export const ObjectPropertyContextProvider: React.FC<
  ObjectPropertyCtx & { children: React.ReactNode }
> = ({ children, index, parentName, depth, objectDescriptionId }) => {
  return (
    <ObjectPropertyContext.Provider
      value={{ index, parentName, depth, objectDescriptionId }}
    >
      {children}
    </ObjectPropertyContext.Provider>
  )
}

export const useObjectPropertyContext = () => {
  const ctx = useContext(ObjectPropertyContext)
  return (
    ctx || {
      parentName: '',
      index: 0,
      depth: 0,
      objectDescriptionId: undefined,
    }
  )
}

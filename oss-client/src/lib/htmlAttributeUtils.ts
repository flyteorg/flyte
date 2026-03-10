
export function skipFalseAttributes(attributes: Record<string,boolean|undefined|null>):Record<string,true> {
  return Object.fromEntries(Object.entries(attributes).filter(([,val])=>!!val).map(([k])=>([k,true])))
}
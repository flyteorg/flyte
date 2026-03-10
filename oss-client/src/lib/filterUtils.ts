import { Filter_Function, FilterSchema } from '@/gen/flyteidl2/common/list_pb'
import { create } from '@bufbuild/protobuf'

export const getFilter = (args: {
  function: Filter_Function
  field: string
  values: string[]
}) => create(FilterSchema, args)

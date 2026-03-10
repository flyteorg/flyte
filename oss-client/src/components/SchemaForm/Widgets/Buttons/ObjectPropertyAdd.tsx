import { IconButtonProps } from '@rjsf/utils'
import { Button } from '@/components/Button'
import { PlusIcon } from '@heroicons/react/20/solid'

type ObjectPropertyAddExtraProps = {
  propertyName?: string
  propertyCount?: number
}

const ObjectPropertyAdd: React.FC<
  IconButtonProps & ObjectPropertyAddExtraProps
> = ({ onClick, propertyCount, propertyName, ...otherProps }) => (
  <div className="row flex items-end justify-end">
    <Button onClick={onClick} {...otherProps} outline color="white">
      <PlusIcon />
      Add {propertyName ?? 'item'}
      {propertyCount !== undefined ? `[${propertyCount}]` : ''}
    </Button>
  </div>
)

export default ObjectPropertyAdd

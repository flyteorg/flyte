import { IconButtonProps } from '@rjsf/utils'
import { Button } from '@/components/Button'
import { ChevronDownIcon } from '@heroicons/react/20/solid'


const CollectionMoveDown:React.FC<IconButtonProps> = ({onClick,...otherProps})=>(
  <div className="row flex items-end justify-end">
    <Button onClick={onClick} {...otherProps}  color="union"><ChevronDownIcon /></Button>
  </div>
)

export default CollectionMoveDown

import { IconButtonProps } from '@rjsf/utils'
import { Button } from '@/components/Button'
import { ChevronUpIcon } from '@heroicons/react/20/solid'


const CollectionMoveUp:React.FC<IconButtonProps> = ({onClick,...otherProps})=>(
  <div className="row flex items-end justify-end">
    <Button onClick={onClick} {...otherProps}  color="union"><ChevronUpIcon /></Button>
  </div>
)

export default CollectionMoveUp

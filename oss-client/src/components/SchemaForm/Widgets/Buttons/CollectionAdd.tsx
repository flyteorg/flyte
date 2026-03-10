import { IconButtonProps } from '@rjsf/utils'
import { Button } from '@/components/Button'
import { PlusIcon } from '@heroicons/react/20/solid'


const CollectionAdd:React.FC<IconButtonProps> = ({onClick,...otherProps})=>(
  <div className="row flex items-end justify-end">
    <Button onClick={onClick} {...otherProps}  color="union"><PlusIcon />Add item</Button>
  </div>
)

export default CollectionAdd

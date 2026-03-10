import { IconButtonProps } from '@rjsf/utils'
import { Button } from '@/components/Button'
import { MinusIcon } from '@heroicons/react/20/solid'

const CollectionRemove:React.FC<IconButtonProps> = ({onClick,...otherProps})=>(
  <Button onClick={onClick} {...otherProps}  color="union"><MinusIcon /></Button>
)

export default CollectionRemove
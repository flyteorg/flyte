import { Button } from '@/components/Button'
import { PlusIcon } from '@heroicons/react/20/solid'
export const CreateProjectButton: React.FC<{ onClick: VoidFunction }> = ({
  onClick,
}) => {
  return (
    <Button onClick={onClick} color="union">
      <PlusIcon className="h-4 w-4" />
      New project
    </Button>
  )
}

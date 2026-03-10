import { Input, InputGroup, InputProps } from '@/components/Input'
import { MagnifyingGlassIcon } from '@heroicons/react/16/solid'
import clsx from 'clsx'

type SearchBarProps = {
  placeholder?: string
  className?: string
} & InputProps

export const SearchBar = ({
  placeholder = 'Search',
  className,
  ...inputProps
}: SearchBarProps) => {
  return (
    <InputGroup
      size="sm"
      className={clsx(
        'w-[300px] rounded-lg border-none bg-(--system-gray-2) shadow-none outline-none',
        className,
      )}
    >
      <MagnifyingGlassIcon
        className="shrink-0 text-(--system-gray-5)"
        data-slot="icon"
      />
      <Input
        size="sm"
        type="search"
        placeholder={placeholder}
        hideBorder={true}
        noOutline={true}
        typographyClassName="text-xs/6 font-medium text-zinc-950 placeholder:font-medium placeholder:text-(--system-gray-5) dark:text-white"
        {...inputProps}
      />
    </InputGroup>
  )
}

import { ComponentType, SVGProps } from 'react'
import { Meta, StoryObj } from '@storybook/nextjs'

// require.context is a Webpack API that discovers modules at build time.
// This means the gallery automatically includes any new icon added to the directory.
// eslint-disable-next-line
const iconContext = (require as any).context(
  '../components/icons',
  false,
  /\.tsx$/,
)

type IconEntry = {
  exportName: string
  fileName: string
  Component: ComponentType<SVGProps<SVGSVGElement>>
}

const icons: IconEntry[] = iconContext
  .keys()
  .flatMap((key: string) => {
    const mod = iconContext(key) as Record<string, unknown>
    const fileName = key.replace('./', '').replace('.tsx', '')
    return Object.entries(mod)
      .filter(([, value]) => typeof value === 'function')
      .map(([exportName, Component]) => ({
        exportName,
        fileName,
        Component: Component as ComponentType<SVGProps<SVGSVGElement>>,
      }))
  })
  .sort((a: IconEntry, b: IconEntry) =>
    a.exportName.localeCompare(b.exportName),
  )

const meta: Meta = {
  title: 'primitives/Icons',
  parameters: {
    layout: 'padded',
  },
}

export default meta

export const Gallery: StoryObj = {
  render: () => (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-4">
      {icons.map(({ exportName, Component }) => (
        <div
          key={exportName}
          className="flex flex-col items-center gap-2 rounded-lg border border-gray-200 p-4 dark:border-gray-700"
          title={exportName}
        >
          <Component className="size-6" />
          <span className="w-full text-center text-xs break-words text-gray-500 dark:text-gray-400">
            {exportName}
          </span>
        </div>
      ))}
    </div>
  ),
}

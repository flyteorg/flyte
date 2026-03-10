import { CopyButton } from '@/components/CopyButton'
import { CrossfadeWithHeight } from '@/components/CrossfadeWithHeight'
import React, { useState } from 'react'
import { ToggleButtonGroup } from './ToggleButtonGroup'

type BaseTabSectionProps = {
  heading: string | React.ReactNode
  copyButtonContent?: string
  showRawJsonToggle?: boolean
}

type SimpleTabSection = BaseTabSectionProps & {
  children: React.ReactNode
}

type ToggleRawTabSection = BaseTabSectionProps & {
  sectionContent: (opts: { isRawView: boolean }) => React.ReactNode
  showRawJsonToggle: true
  defaultView?: 'raw' | 'pretty'
}

type TabSectionProps = SimpleTabSection | ToggleRawTabSection

function isToggleTab(props: TabSectionProps): props is ToggleRawTabSection {
  return props.showRawJsonToggle === true
}

const TabSectionLayout = ({
  heading,
  copyButtonContent,
  rightControls,
  children,
}: {
  heading: string | React.ReactNode | null
  copyButtonContent?: string
  rightControls?: React.ReactNode
  children: React.ReactNode
}) => (
  <div className="flex w-full flex-col" data-testid={`tabsection-${heading}`}>
    <div
      className={`sticky top-0 z-10 flex items-center justify-between ${rightControls || copyButtonContent ? 'py-2' : ''} bg-(--system-gray-2)`}
    >
      {heading ? <h3 className="py-1 text-sm font-bold">{heading}</h3> : null}
      <div className="flex gap-3">
        {rightControls}
        {copyButtonContent && (
          <CopyButton
            size="sm"
            className="!px-[9px] !py-[3px] *:data-[slot=icon]:!size-4"
            value={copyButtonContent}
          />
        )}
      </div>
    </div>
    <div className="relative overflow-hidden rounded-2xl border border-(--system-gray-2) bg-(--system-black)">
      {children}
    </div>
  </div>
)

export const TabSection = (props: TabSectionProps) => {
  const [isRawView, setIsRawView] = useState(
    isToggleTab(props) ? props.defaultView !== 'pretty' : true,
  )

  if (isToggleTab(props)) {
    const { heading, copyButtonContent, sectionContent } = props
    const toggleControls = (
      <ToggleButtonGroup
        isRawView={isRawView}
        onDisableRaw={() => setIsRawView(false)}
        onEnableRaw={() => setIsRawView(true)}
      />
    )

    return (
      <TabSectionLayout
        heading={heading}
        copyButtonContent={copyButtonContent}
        rightControls={toggleControls}
      >
        <CrossfadeWithHeight mode={isRawView ? 'raw' : 'formatted'}>
          <CrossfadeWithHeight.View name="raw">
            {sectionContent({ isRawView: true })}
          </CrossfadeWithHeight.View>
          <CrossfadeWithHeight.View name="formatted">
            {sectionContent({ isRawView: false })}
          </CrossfadeWithHeight.View>
        </CrossfadeWithHeight>
      </TabSectionLayout>
    )
  }

  const { heading, copyButtonContent, children } = props
  return (
    <TabSectionLayout heading={heading} copyButtonContent={copyButtonContent}>
      {children}
    </TabSectionLayout>
  )
}

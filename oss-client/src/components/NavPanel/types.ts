type BaseNavItem = {
  className?: string
  displayText: string // used as unique key
  shouldHideIconOnCollapse?: boolean
}

type MakeHrefProps = {
  project: string
  domain: string
  pathname?: string
}

export type NavLink = BaseNavItem & {
  displayComponent?: React.ReactNode // if this is passed, render this instead of displayText
  makeHref: (props: MakeHrefProps) => string
  icon?: React.ReactNode
  onClick?: () => void
  target?: string
  /** If false, Next.js will not prefetch this link (e.g. for settings/profile to avoid loading it on every page) */
  prefetch?: boolean
  type: 'link'
}

export type NavSectionHeading = BaseNavItem & {
  icon?: React.ReactNode
  type: 'heading'
}

export type NavWidget = BaseNavItem & {
  type: 'widget'
  widget: (size: NavPanelWidth) => React.ReactNode
}

export type NavItemType = NavLink | NavSectionHeading | NavWidget

export type NavPanelWidth = 'thin' | 'wide'

export type NavPanelMode = 'overlay' | 'embedded'

export type NavPanelType = 'default' | 'settings'

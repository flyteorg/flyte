import React, { ReactElement, ReactNode, useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import useMeasure from 'react-use-measure'

type Mode = string

type CrossfadeWithHeightProps = {
  mode: Mode
  children: ReactElement<CrossfadeWithHeightViewProps>[]
  className?: string
}

type CrossfadeWithHeightViewProps = {
  name: Mode
  children: ReactNode
}

export const CrossfadeWithHeight = ({
  mode,
  children,
  className = '',
}: CrossfadeWithHeightProps) => {
  const selectedChild = React.Children.toArray(children).find(
    (child: ReactNode) =>
      React.isValidElement(child) &&
      (child as ReactElement<CrossfadeWithHeightViewProps>).props.name === mode,
  ) as ReactElement<CrossfadeWithHeightViewProps>

  const [ref, bounds] = useMeasure()
  const [height, setHeight] = useState<number | 'auto'>('auto')

  useEffect(() => {
    if (bounds.height > 0) {
      setHeight(bounds.height)
    }
  }, [bounds.height, selectedChild.props.name])

  return (
    <motion.div
      animate={{ height }}
      transition={{ duration: 0.3, ease: 'easeInOut' }}
      className={`relative overflow-hidden ${className}`}
      style={{ height }}
    >
      <AnimatePresence mode="wait">
        <motion.div
          key={selectedChild.props.name}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          ref={ref}
        >
          {selectedChild.props.children}
        </motion.div>
      </AnimatePresence>
    </motion.div>
  )
}

CrossfadeWithHeight.View = function View({
  children,
}: CrossfadeWithHeightViewProps) {
  return <>{children}</>
}

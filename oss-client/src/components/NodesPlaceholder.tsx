import { TraceNode } from '@/components/TraceNode'

const sampleTrace = {
  name: 'Sample Agent Trace',
  icon: '🗂️',
  duration: 7.31,
  children: [
    {
      name: 'ChatOpenAI',
      icon: '📄',
      duration: 0.95,
    },
    {
      name: 'search_latest_knowledge',
      icon: '🗂️',
      duration: 0.76,
      children: [
        {
          name: 'VectorStoreRetriever',
          icon: '📄',
          duration: 0.52,
          children: [
            {
              name: 'something',
              icon: '🗂️',
              duration: 1.76,
              children: [
                {
                  name: 'ChatOpenAI',
                  icon: '📄',
                  duration: 0.95,
                  children: [
                    {
                      name: 'Hello_World',
                      icon: '📄',
                      duration: 0.95,
                    },
                  ],
                },
              ],
            },
          ],
        },
      ],
    },
    {
      name: 'ChatOpenAI',
      icon: '📄',
      duration: 3.44,
      children: [
        {
          name: 'something',
          icon: '🗂️',
          duration: 1.76,
        },
      ],
    },
    {
      name: 'ChatOpenAI',
      icon: '📄',
      duration: 3.44,
      children: [
        {
          name: 'something',
          icon: '🗂️',
          duration: 1.76,
          children: [
            {
              name: 'ChatOpenAI',
              icon: '📄',
              duration: 0.95,
              children: [
                {
                  name: 'Hello_World',
                  icon: '📄',
                  duration: 0.95,
                },
              ],
            },
          ],
        },
      ],
    },
    {
      name: 'ChatOpenAI',
      icon: '📄',
      duration: 3.44,
    },
    {
      name: 'ChatOpenAI',
      icon: '📄',
      duration: 3.44,
    },
    {
      name: 'ChatOpenAI',
      icon: '📄',
      duration: 3.44,
    },
  ],
}

export function NodesPlaceholder() {
  return <TraceNode node={sampleTrace} />
}

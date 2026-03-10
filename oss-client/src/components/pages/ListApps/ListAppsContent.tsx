import { CopyButton } from '@/components/CopyButton'
import { ArrowTopRightIcon } from '@/components/icons/ArrowTopRightIcon'
import { TableState } from '@/components/Tables'
import { useListApps } from '@/hooks/useApps'
import { DOCS_BYOC_USER_GUIDE_URL } from '@/lib/constants'
import { getLocation } from '@/lib/windowUtils'
import { ProjectDomainPageParams } from '@/types/pageParams'
import { python } from '@codemirror/lang-python'
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import { useTheme } from 'next-themes'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { ListAppsTable } from './table/ListAppsTable'

type ListAppsContentProps = {
  listAppsQuery: ReturnType<typeof useListApps>
  searchQuery: string
}

const useCodeSnippets = () => {
  const params = useParams<ProjectDomainPageParams>()
  const { hostname } = getLocation()
  return [
    {
      id: '1',
      label:
        'If you don’t have a local flyte configuration, first create it with:',
      code: `flyte create config \\
    --endpoint ${hostname} \\
    --project ${params.project} \\
    --domain ${params.domain} \\
    --builder remote`,
    },
    {
      id: '2',
      label: 'Then create an app script called app.py:',
      code: `import flyte
import flyte.app

image = flyte.Image.from_debian_base(python_version=(3, 12)).with_pip_packages("streamlit==1.41.1")

# The 'App' declaration.
# Uses the 'ImageSpec' declared above.
# In this case we do not need to supply any app code
# as we are using the built-in Streamlit 'hello' app.
app_env = flyte.app.AppEnvironment(
    name="streamlit-hello",
    image=image,
    command="streamlit hello --server.port 8080",
    resources=flyte.Resources(cpu="1", memory="1Gi"),
)`,
    },
    {
      id: '3',
      label: 'Then serve the app:',
      code: `flyte serve app.py app_env`,
    },
  ]
}

export const ListAppsContent = ({
  listAppsQuery,
  searchQuery,
}: ListAppsContentProps) => {
  const { resolvedTheme } = useTheme()
  const codeSnippets = useCodeSnippets()

  return (
    <TableState
      data={listAppsQuery.data?.apps}
      dataLabel="apps"
      isError={listAppsQuery.isError}
      isLoading={listAppsQuery.isLoading}
      searchQuery={searchQuery}
      subtitle="Apps allow you to build and serve your own web apps, enabling you to build model endpoints, AI inference-time components, interactive dashboards, connectors, and more."
      content={
        <div>
          <Link
            target="_blank"
            href={`${DOCS_BYOC_USER_GUIDE_URL}/core-concepts/introducing-apps/`}
            className="flex items-center justify-center gap-2 p-3 text-sm"
          >
            <span>How to create an app</span>
            <ArrowTopRightIcon className="h-2.125 w-2" />
          </Link>
          <div className="max-w-[600px] min-w-[550px]">
            {codeSnippets.map(({ code, id, label }) => (
              <div key={id} className="mt-6 w-full">
                <p className="text-sm font-bold">{label}</p>
                <div className="relative mt-2 w-full text-[11px] [&_.cm-editor]:!bg-transparent [&_.cm-focused]:!outline-none [&_.cm-gutters]:!bg-transparent [&_.cm-scroller]:!rounded-2xl [&_.cm-scroller]:!border [&_.cm-scroller]:!border-(--system-white)/14 [&_.cm-scroller>:where(.cm-content)]:!p-5">
                  <div className="pointer-events-auto absolute top-3 right-3 z-20">
                    <CopyButton value={code} />
                  </div>
                  <CodeMirror
                    readOnly
                    editable={false}
                    theme={resolvedTheme === 'dark' ? vscodeDark : vscodeLight}
                    extensions={[python(), EditorView.lineWrapping]}
                    basicSetup={{
                      lineNumbers: false,
                      foldGutter: false,
                      highlightActiveLine: false,
                      highlightActiveLineGutter: false,
                    }}
                    value={code}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      }
    >
      {(data) => <ListAppsTable data={data} />}
    </TableState>
  )
}

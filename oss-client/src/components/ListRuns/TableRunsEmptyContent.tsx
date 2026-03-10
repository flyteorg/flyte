import { getLocation } from '@/lib/windowUtils'
import { ProjectDomainPageParams } from '@/types/pageParams'
import { python } from '@codemirror/lang-python'
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import { useTheme } from 'next-themes'
import { useParams } from 'next/navigation'
import { CopyButton } from '../CopyButton'

export const TableRunsEmptyContent = () => {
  const { resolvedTheme } = useTheme()
  const params = useParams<ProjectDomainPageParams>()
  const { hostname } = getLocation()
  const codeSnippets = [
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
      label: 'Then create a simple script called hello_world.py:',
      code: `import flyte

env = flyte.TaskEnvironment(name="hello_world")

@env.task
def fn(x: int) -> int:
    slope, intercept = 2, 5
    return slope * x + intercept

@env.task
def main(n: int) -> float:
    y_list = list(flyte.map(fn, range(n)))
    return sum(y_list) / len(y_list)`,
    },
    {
      id: '3',
      label: 'Then run the task:',
      code: `flyte run hello_world.py main --n 10`,
    },
  ]

  return (
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
  )
}

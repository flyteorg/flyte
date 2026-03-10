import {
  FLYTE_LIGHT_THEME,
  registerFlyteLightTheme,
} from '@/utils/monacoThemes'
import MonacoEditor from '@monaco-editor/react'
import { useFormContext } from 'react-hook-form'
import stringify from 'safe-stable-stringify'
import { TabLayout } from './LaunchFormTabs'
import { LaunchFormState } from './types'

export const LaunchFormInputs = () => {
  const { setValue, setError, clearErrors, formState, watch } =
    useFormContext<LaunchFormState>()
  const formDataValues = watch('formData')

  return (
    <TabLayout>
      <div
        style={{
          height: 'calc(100vh - 295px)',
          width: '100%',
          border: formState.errors.inputs ? '1px solid #ef4444' : 'none',
          borderRadius: formState.errors.inputs ? '4px' : '0',
        }}
        className="monaco-editor-transparent"
      >
        <MonacoEditor
          beforeMount={registerFlyteLightTheme}
          height="100%"
          width="100%"
          defaultLanguage="json"
          onChange={(value) => {
            try {
              const parsedProperties = JSON.parse(value!)
              setValue('formData', parsedProperties)
              clearErrors('inputs')
            } catch (error) {
              console.error(error)
              setError('inputs', {
                message: 'Invalid JSON',
              })
            }
          }}
          value={stringify(formDataValues, null, 2)}
          theme={FLYTE_LIGHT_THEME}
          options={{
            lineNumbers: 'off',
            minimap: { enabled: false },
            readOnly: false,
            scrollBeyondLastLine: false,
            wordWrap: 'on',
            renderLineHighlight: 'none',
            overviewRulerLanes: 0,
            hideCursorInOverviewRuler: true,
            scrollbar: { vertical: 'auto', horizontal: 'auto' },
          }}
        />
      </div>
    </TabLayout>
  )
}

import { ChartIcon } from '@/components/icons/ChartIcon'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import { getFormDataFromSchemaDefaults } from '@/components/SchemaForm/utils'
import type { JSONSchema7 } from 'json-schema'
import React from 'react'
import { LoadingSpinner } from '../LoadingSpinner'
import ThemedJSONTree from '../ThemedJSONTree'
import { IORendererProps } from './types'

export const IORenderer: React.FC<IORendererProps> = ({
  isLoading,
  isRawView = true,
  jsonSchema,
  formData,
  formDataProvided = false,
  noDataMessage = 'No data available',
  expandLevel,
}) => {
  const hasData =
    !!jsonSchema?.properties && Object.keys(jsonSchema.properties).length > 0

  if (!isRawView) {
    return (
      <div className="w-full rounded-lg py-4 shadow-sm">
        <LicensedEditionPlaceholder title="Formatted view" fullWidth hideBorder />
      </div>
    )
  }

  return (
    <div className="w-full rounded-lg py-4 shadow-sm">
      {isLoading ? (
        <div className="flex h-16 items-center justify-center px-4">
          <LoadingSpinner delay={500} />
        </div>
      ) : !hasData ? (
        <div className="flex h-16 items-center justify-center gap-2 px-4 text-sm text-(--system-gray-5)">
          <ChartIcon className="size-5" />
          <span>{noDataMessage}</span>
        </div>
      ) : (
        <div className="px-4">
          <div
            className="-mr-4 overflow-auto pr-4 [scrollbar-gutter:stable] [&::-webkit-scrollbar]:size-2 [&::-webkit-scrollbar-corner]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-400 dark:[&::-webkit-scrollbar-thumb]:bg-zinc-600 [&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 dark:[&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 [&::-webkit-scrollbar-track]:bg-transparent"
            style={{
              maxHeight: '606px',
              scrollbarWidth: 'thin',
              scrollbarColor: 'rgb(161 161 170) transparent',
            }}
          >
            <ThemedJSONTree
              data={
                formDataProvided
                  ? (formData ?? {})
                  : formData !== undefined
                    ? formData
                    : jsonSchema
                      ? (getFormDataFromSchemaDefaults(
                          jsonSchema as JSONSchema7,
                        ) as Record<string, unknown>)
                      : undefined
              }
              expandLevel={expandLevel}
            />
          </div>
        </div>
      )}
    </div>
  )
}

import { Secret } from '@/gen/flyteidl2/secret/definition_pb'

export type SecretsTableRow = {
  name: string | undefined
  scope: { title: string; subtitle: string }
  createdAt: string | undefined
  original: Secret
  actions: Secret
}

import type { Dataset, Language } from '../types'
import type { T } from '../i18n'

export type Run = (action: () => Promise<void>, message?: string) => Promise<void>
export interface ViewProps { dataset: Dataset; language: Language; t: T; busy: boolean; run: Run; onUpdate: (data: Dataset) => void }

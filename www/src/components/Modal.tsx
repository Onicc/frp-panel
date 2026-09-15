import { type FormEvent, type PropsWithChildren, useEffect, useRef, useState } from 'react'
import { useI18n } from '../i18n'

type ModalProps = PropsWithChildren<{
  open: boolean
  title: string
  submitLabel?: string
  onOpenChange: (open: boolean) => void
  onSubmit: () => Promise<void>
}>

// All create/edit dialogs use this contract: success closes and resets,
// failure remains open with an actionable error.
export function Modal({ open, title, submitLabel, onOpenChange, onSubmit, children }: ModalProps) {
  const { t } = useI18n()
  const form = useRef<HTMLFormElement>(null)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!open) setError('')
  }, [open])
  if (!open) return null

  async function submit(event: FormEvent) {
    event.preventDefault()
    setPending(true)
    setError('')
    try {
      await onSubmit()
      form.current?.reset()
      onOpenChange(false)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause))
    } finally {
      setPending(false)
    }
  }

  return <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onOpenChange(false)}>
    <section className="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div className="modal-header"><h2 id="modal-title">{title}</h2><button className="icon-button" aria-label="Close" onClick={() => onOpenChange(false)}>×</button></div>
      <form ref={form} onSubmit={submit}>
        <div className="modal-body">{children}{error && <p className="error" role="alert">{error}</p>}</div>
        <div className="modal-actions"><button type="button" className="button secondary" onClick={() => onOpenChange(false)}>{t('cancel')}</button><button className="button primary" disabled={pending}>{pending ? t('saving') : (submitLabel ?? t('confirm'))}</button></div>
      </form>
    </section>
  </div>
}

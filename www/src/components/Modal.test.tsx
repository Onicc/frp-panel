import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { I18nProvider } from '../i18n'
import { Modal } from './Modal'

describe('Modal success contract', () => {
  it('closes after a successful submission', async () => {
    const onOpenChange = vi.fn()
    render(<I18nProvider><Modal open title="Create" onOpenChange={onOpenChange} onSubmit={async () => undefined}><input aria-label="name" /></Modal></I18nProvider>)
    fireEvent.click(screen.getByRole('button', { name: '确认' }))
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false))
  })

  it('stays open and shows an error when submission fails', async () => {
    const onOpenChange = vi.fn()
    render(<I18nProvider><Modal open title="Create" onOpenChange={onOpenChange} onSubmit={async () => { throw new Error('conflict') }} /></I18nProvider>)
    fireEvent.click(screen.getByRole('button', { name: '确认' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('conflict')
    expect(onOpenChange).not.toHaveBeenCalled()
  })
})

import { describe, expect, it } from 'vitest'
import { displayVersion, displayVersionDetails } from './displayVersion'

describe('displayVersion', () => {
  it('normalizes stable releases without exposing build identifiers', () => {
    expect(displayVersion('v1.2.3')).toBe('v1.2.3')
    expect(displayVersion('1.2.3')).toBe('v1.2.3')
  })

  it('marks old rolling builds without misrepresenting them as stable', () => {
    expect(displayVersion('main')).toBe('Legacy build')
    expect(displayVersion('edge')).toBe('Legacy build')
    expect(displayVersion('edge-SNAPSHOT-7cd4e60')).toBe('Legacy build')
  })

  it('does not invent a version for missing or unrecognized builds', () => {
    expect(displayVersion()).toBe('—')
    expect(displayVersion('dev-build')).toBe('Development build')
  })

  it('keeps the exact build available as a short hover detail', () => {
    expect(displayVersionDetails('edge-SNAPSHOT-7cd4e60', '7cd4e601ce20')).toBe('Legacy build · 7cd4e60')
  })
})

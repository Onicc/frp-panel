import { config } from '@vue/test-utils'
import { afterEach } from 'vitest'

config.global.stubs = { teleport: true, transition: false, 'transition-group': false }
afterEach(() => document.body.innerHTML = '')

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import Input from './Input.vue'

describe('Input', () => {
  it('forwards required semantics and supports numeric v-model modifiers', async () => {
    const wrapper = mount(Input, {
      props: { modelValue: 7000, required: true, modelModifiers: { number: true } },
    })
    const input = wrapper.get('input')
    expect(input.element.required).toBe(true)
    await input.setValue('7443')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([7443])
  })
})

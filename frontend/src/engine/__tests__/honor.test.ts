import { describe, expect, it } from 'vitest'
import { clampHonor, reactivityForHonor } from '../honor'

describe('reactivityForHonor', () => {
  it('classifies the good band', () => {
    expect(reactivityForHonor(75).alignment).toBe('good')
    expect(reactivityForHonor(60).alignment).toBe('good')
    expect(reactivityForHonor(100).alignment).toBe('good')
  })

  it('classifies the neutral band', () => {
    expect(reactivityForHonor(0).alignment).toBe('neutral')
    expect(reactivityForHonor(59).alignment).toBe('neutral')
    expect(reactivityForHonor(-59).alignment).toBe('neutral')
  })

  it('classifies the evil band', () => {
    expect(reactivityForHonor(-60).alignment).toBe('evil')
    expect(reactivityForHonor(-100).alignment).toBe('evil')
  })

  it('applies the documented shop price modifiers', () => {
    expect(reactivityForHonor(80).shopPriceModifier).toBe(-0.1)
    expect(reactivityForHonor(0).shopPriceModifier).toBe(0)
    expect(reactivityForHonor(-80).shopPriceModifier).toBe(0.2)
  })
})

describe('clampHonor', () => {
  it('clamps to the -100..100 range', () => {
    expect(clampHonor(150)).toBe(100)
    expect(clampHonor(-150)).toBe(-100)
    expect(clampHonor(42)).toBe(42)
  })
})

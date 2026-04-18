import { describe, it, expect } from 'vitest';
import { cn, shortName, sessionName, initials } from './utils.js';

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar');
  });

  it('handles conditional classes', () => {
    expect(cn('foo', false && 'bar', 'baz')).toBe('foo baz');
  });

  it('deduplicates conflicting tailwind classes', () => {
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500');
  });
});

describe('shortName', () => {
  it('returns single word unchanged', () => {
    expect(shortName('Alice')).toBe('Alice');
  });

  it('abbreviates last name', () => {
    expect(shortName('Fabian Thorsen')).toBe('Fabian T.');
  });

  it('abbreviates last name for three words', () => {
    expect(shortName('Mary Jane Watson')).toBe('Mary W.');
  });

  it('trims surrounding whitespace', () => {
    expect(shortName('  Alice  ')).toBe('Alice');
  });

  it('handles empty string', () => {
    expect(shortName('')).toBe('');
  });
});

describe('sessionName', () => {
  it('returns session name when set', () => {
    expect(sessionName({ name: 'My Game', game_mode: 'americano' })).toBe('My Game');
  });

  it('returns tennis default for tennis mode', () => {
    expect(sessionName({ game_mode: 'tennis' })).toBe('OpenPadel 2v2');
  });

  it('returns americano default for americano mode', () => {
    expect(sessionName({ game_mode: 'americano' })).toBe('OpenPadel Americano');
  });

  it('returns americano default for mexicano mode', () => {
    expect(sessionName({ game_mode: 'mexicano' })).toBe('OpenPadel Americano');
  });

  it('returns americano default when no name and no mode', () => {
    expect(sessionName({})).toBe('OpenPadel Americano');
  });
});

describe('initials', () => {
  it('returns ? for empty string', () => {
    expect(initials('')).toBe('?');
  });

  it('returns single uppercase initial for one word', () => {
    expect(initials('alice')).toBe('A');
  });

  it('returns two initials for two words', () => {
    expect(initials('Fabian Thorsen')).toBe('FT');
  });

  it('returns first and last initial for three words', () => {
    expect(initials('Mary Jane Watson')).toBe('MW');
  });

  it('is always uppercase', () => {
    expect(initials('alice bob')).toBe('AB');
  });

  it('handles extra whitespace', () => {
    expect(initials('  Alice   Bob  ')).toBe('AB');
  });
});

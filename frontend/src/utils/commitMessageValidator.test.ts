// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from 'vitest';
import { validateCommitMessage } from './commitMessageValidator';

describe('validateCommitMessage', () => {
  it('returns false for empty input', () => {
    expect(validateCommitMessage('').isValid).toBe(false);
    expect(validateCommitMessage('   ').isValid).toBe(false);
  });

  it('allows special Git commits like Merge and Revert', () => {
    expect(validateCommitMessage('Merge branch dev into main').isValid).toBe(true);
    expect(validateCommitMessage('Revert "feat: something"').isValid).toBe(true);
    expect(validateCommitMessage('fixup! feat: something').isValid).toBe(true);
  });

  it('validates standard conventional commit', () => {
    const res = validateCommitMessage('feat: add user login');
    expect(res.isValid).toBe(true);
    expect(res.type).toBe('feat');
    expect(res.description).toBe('add user login');
    expect(res.warning).toBeUndefined();
  });

  it('validates scoped conventional commit with breaking change', () => {
    const res = validateCommitMessage('fix(auth)!: prevent token leak');
    expect(res.isValid).toBe(true);
    expect(res.type).toBe('fix');
    expect(res.scope).toBe('auth');
    expect(res.isBreaking).toBe(true);
  });

  it('detects missing space after colon and provides autoFix', () => {
    const res = validateCommitMessage('feat:add user login');
    expect(res.isValid).toBe(false);
    expect(res.canAutoFix).toBe(true);
    expect(res.fixedMessage).toBe('feat: add user login');
  });

  it('detects ending period and provides autoFix', () => {
    const res = validateCommitMessage('feat(ui): add new button.');
    expect(res.isValid).toBe(false);
    expect(res.canAutoFix).toBe(true);
    expect(res.fixedMessage).toBe('feat(ui): add new button');
  });

  it('detects missing empty line in multiline commit and provides autoFix', () => {
    const res = validateCommitMessage('feat: add feature\nThis is detailed body');
    expect(res.isValid).toBe(false);
    expect(res.canAutoFix).toBe(true);
    expect(res.fixedMessage).toBe('feat: add feature\n\nThis is detailed body');
  });

  it('warns about non-conventional format', () => {
    const res = validateCommitMessage('just doing some stuff');
    expect(res.isValid).toBe(false);
    expect(res.warning).toBeDefined();
  });

  it('rejects header exceeding 120 characters limit', () => {
    const longHeader = 'feat: ' + 'a'.repeat(125);
    const res = validateCommitMessage(longHeader);
    expect(res.isValid).toBe(false);
    expect(res.warning).toContain('120');
  });

  it('gives soft advisory for header between 73 and 120 characters', () => {
    const mediumHeader = 'feat: ' + 'a'.repeat(80);
    const res = validateCommitMessage(mediumHeader);
    expect(res.isValid).toBe(true);
    expect(res.warning).toContain('72');
  });
});

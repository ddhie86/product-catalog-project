// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { formatRupiah, isSafeRedirectUrl } from './format';

describe('formatRupiah', () => {
  it('memformat angka ke Rupiah tanpa desimal', () => {
    expect(formatRupiah(25000)).toMatch(/Rp\s?25\.000/);
  });
  it('menganggap nilai tidak valid sebagai 0', () => {
    expect(formatRupiah(undefined)).toMatch(/Rp\s?0/);
  });
});

describe('isSafeRedirectUrl', () => {
  it('menerima https', () => expect(isSafeRedirectUrl('https://pay.example/inv')).toBe(true));
  it('menolak javascript: dan data:', () => {
    expect(isSafeRedirectUrl('javascript:alert(1)')).toBe(false);
    expect(isSafeRedirectUrl('data:text/html,hi')).toBe(false);
  });
  it('menolak http non-localhost dan teks bebas', () => {
    expect(isSafeRedirectUrl('http://evil.example')).toBe(false);
    expect(isSafeRedirectUrl('bukan url')).toBe(false);
  });
});

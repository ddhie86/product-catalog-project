// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { fireEvent, render, screen } from '@testing-library/react';
import { expect, it } from 'vitest';
import ProductImage from './ProductImage.jsx';

it('shows the fallback when the image source is empty', () => {
  render(<ProductImage src="" alt="Produk contoh" />);

  expect(screen.getByRole('img', { name: 'Produk contoh' })).toHaveTextContent('Gambar produk tidak tersedia');
  expect(screen.queryByRole('img', { name: 'Produk contoh' }).querySelector('img')).toBeNull();
});

it('replaces a broken image with the fallback without retrying the broken source', () => {
  render(<ProductImage src="https://example.com/broken.webp" alt="Produk contoh" />);

  fireEvent.error(screen.getByRole('img', { name: 'Produk contoh' }));

  const fallback = screen.getByRole('img', { name: 'Produk contoh' });
  expect(fallback).toHaveTextContent('Gambar produk tidak tersedia');
  expect(fallback.querySelector('img')).toBeNull();
});

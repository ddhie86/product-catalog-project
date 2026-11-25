// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useEffect, useState } from 'react';

const FALLBACK_LABEL = 'Gambar produk tidak tersedia';

/**
 * Displays a product image and a consistent inline placeholder when the URL is
 * empty or the image cannot be loaded.
 */
export default function ProductImage({ src, alt = 'Gambar produk', className = '', loading }) {
  const [hasError, setHasError] = useState(!src);

  useEffect(() => {
    setHasError(!src);
  }, [src]);

  if (!src || hasError) {
    return (
      <div
        role="img"
        aria-label={alt || FALLBACK_LABEL}
        className={`flex h-full w-full flex-col items-center justify-center gap-2 bg-paper p-4 text-ink-soft ${className}`}
      >
        <svg
          aria-hidden="true"
          viewBox="0 0 64 64"
          className="h-12 w-12 opacity-60"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <rect x="8" y="11" width="48" height="42" rx="5" stroke="currentColor" strokeWidth="2.5" />
          <circle cx="23" cy="25" r="4" fill="currentColor" />
          <path d="m12 45 13-13 9 9 7-7 11 11" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
        <span className="text-center text-xs font-medium">{FALLBACK_LABEL}</span>
      </div>
    );
  }

  return (
    <img
      src={src}
      alt={alt}
      loading={loading}
      onError={() => setHasError(true)}
      className={className}
    />
  );
}

// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { useId } from 'react';

const base =
  'block w-full rounded-md border bg-white px-3 py-2 text-sm text-ink placeholder:text-ink-soft disabled:bg-paper disabled:text-ink-soft';

function Field({ label, error, hint, children, id }) {
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">
        {label}
      </label>
      {children}
      {hint && !error && <p className="mt-1 text-xs text-ink-soft">{hint}</p>}
      {error && (
        <p id={`${id}-error`} className="mt-1 text-xs text-danger">
          {error}
        </p>
      )}
    </div>
  );
}

const borderFor = (error) => (error ? 'border-danger' : 'border-line');

export function TextField({ label, error, hint, className = '', ...props }) {
  const id = useId();
  return (
    <Field id={id} label={label} error={error} hint={hint}>
      <input id={id} aria-invalid={Boolean(error)} aria-describedby={error ? `${id}-error` : undefined} className={`${base} ${borderFor(error)} ${className}`} {...props} />
    </Field>
  );
}

export function TextAreaField({ label, error, hint, rows = 4, ...props }) {
  const id = useId();
  return (
    <Field id={id} label={label} error={error} hint={hint}>
      <textarea id={id} rows={rows} aria-invalid={Boolean(error)} aria-describedby={error ? `${id}-error` : undefined} className={`${base} ${borderFor(error)}`} {...props} />
    </Field>
  );
}

export function SelectField({ label, error, hint, options = [], placeholder, ...props }) {
  const id = useId();
  return (
    <Field id={id} label={label} error={error} hint={hint}>
      <select id={id} aria-invalid={Boolean(error)} aria-describedby={error ? `${id}-error` : undefined} className={`${base} ${borderFor(error)}`} {...props}>
        {placeholder !== undefined && <option value="">{placeholder}</option>}
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </Field>
  );
}

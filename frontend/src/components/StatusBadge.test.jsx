import { render, screen } from '@testing-library/react';
import { StatusBadge } from './StatusBadge.jsx';

it('menampilkan label Indonesia untuk status', () => {
  render(<StatusBadge status="PAID" />);
  expect(screen.getByText('Lunas')).toBeInTheDocument();
});

it('menampilkan status mentah bila tidak dikenal', () => {
  render(<StatusBadge status="REFUNDED" />);
  expect(screen.getByText('REFUNDED')).toBeInTheDocument();
});

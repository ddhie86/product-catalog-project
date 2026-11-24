// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import { fireEvent, render, screen } from '@testing-library/react';
import { Modal } from './Modal.jsx';

it('keeps focus in the form when the close callback changes during typing', () => {
  const renderModal = () => (
    <Modal open title="Balas pesan" onClose={() => {}}>
      <textarea aria-label="Isi balasan" />
    </Modal>
  );
  const { rerender } = render(renderModal());
  const textbox = screen.getByRole('textbox');
  textbox.focus();
  fireEvent.change(textbox, { target: { value: 'Balasan sedang diketik' } });

  rerender(renderModal());

  expect(document.activeElement).toBe(textbox);
  expect(textbox).toHaveValue('Balasan sedang diketik');
});

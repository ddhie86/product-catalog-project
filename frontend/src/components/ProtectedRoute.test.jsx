import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AuthProvider } from '@/modules/auth/AuthContext.jsx';
import ProtectedRoute from './ProtectedRoute.jsx';

function renderAt(path) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AuthProvider>
        <Routes>
          <Route path="/admin/login" element={<p>Halaman login</p>} />
          <Route path="/admin" element={<ProtectedRoute />}>
            <Route index element={<p>Isi admin</p>} />
          </Route>
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  );
}

afterEach(() => localStorage.clear());

it('mengalihkan ke login bila belum masuk', () => {
  renderAt('/admin');
  expect(screen.getByText('Halaman login')).toBeInTheDocument();
});

it('menampilkan halaman admin bila token ada', () => {
  localStorage.setItem('pc_admin_token', 'token-uji');
  renderAt('/admin');
  expect(screen.getByText('Isi admin')).toBeInTheDocument();
});

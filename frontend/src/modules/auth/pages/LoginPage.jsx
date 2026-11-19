import { useState } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../AuthContext.jsx';
import { Button } from '@/components/Button.jsx';
import { TextField } from '@/components/FormFields.jsx';

export default function LoginPage() {
  const { isAuthenticated, login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = location.state?.from?.pathname || '/admin';

  const [form, setForm] = useState({ email: '', password: '' });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  if (isAuthenticated) return <Navigate to={from} replace />;

  const onChange = (e) => setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

  async function onSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(form);
      navigate(from, { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-sm flex-col justify-center px-4">
      <h1 className="text-2xl font-semibold">Masuk ke panel admin</h1>
      <form onSubmit={onSubmit} className="mt-6 space-y-4 rounded-lg border border-line bg-white p-6" noValidate>
        <TextField label="Email" name="email" type="email" autoComplete="username" required value={form.email} onChange={onChange} />
        <TextField label="Kata sandi" name="password" type="password" autoComplete="current-password" required value={form.password} onChange={onChange} />
        {error && (
          <p role="alert" className="text-sm text-danger">
            {error}
          </p>
        )}
        <Button type="submit" loading={loading} className="w-full">
          Masuk
        </Button>
      </form>
    </main>
  );
}

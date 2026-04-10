import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api';

export function Login() {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res = await api.login(username, password);
      localStorage.setItem('token', res.token);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="min-h-screen flex items-center justify-center px-6"
      style={{ background: '#F9F8F6' }}
    >
      <div className="w-full max-w-sm editorial-fade-in">
        <div className="text-center mb-12">
          <h1
            className="text-3xl md:text-4xl tracking-tight mb-3"
            style={{ fontFamily: 'ui-serif, Georgia, serif' }}
          >
            参考答案阅览室
          </h1>
          <p className="text-xs tracking-widest uppercase opacity-40">
            Reference Answer RSS
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <div
              className="text-sm py-3 px-4 border"
              style={{ borderColor: 'rgba(28,28,28,0.2)', color: '#1C1C1C' }}
            >
              {error}
            </div>
          )}

          <div>
            <label
              className="block text-xs tracking-widest uppercase opacity-50 mb-2"
              htmlFor="username"
            >
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full border px-4 py-3 text-sm bg-transparent transition-colors focus:outline-none"
              style={{
                borderColor: 'rgba(28,28,28,0.15)',
              }}
              onFocus={(e) => (e.target.style.borderColor = '#1C1C1C')}
              onBlur={(e) => (e.target.style.borderColor = 'rgba(28,28,28,0.15)')}
              placeholder="Enter username"
              required
              autoFocus
            />
          </div>

          <div>
            <label
              className="block text-xs tracking-widest uppercase opacity-50 mb-2"
              htmlFor="password"
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full border px-4 py-3 text-sm bg-transparent transition-colors focus:outline-none"
              style={{
                borderColor: 'rgba(28,28,28,0.15)',
              }}
              onFocus={(e) => (e.target.style.borderColor = '#1C1C1C')}
              onBlur={(e) => (e.target.style.borderColor = 'rgba(28,28,28,0.15)')}
              placeholder="Enter password"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 text-sm tracking-widest uppercase transition-colors"
            style={{
              background: '#1C1C1C',
              color: '#F9F8F6',
              opacity: loading ? 0.6 : 1,
            }}
          >
            {loading ? 'Signing in...' : 'Sign In'}
          </button>
        </form>

        <p className="text-center mt-8 text-xs opacity-30">
          Powered by Xinzhi &middot; Editorial Style
        </p>
      </div>
    </div>
  );
}

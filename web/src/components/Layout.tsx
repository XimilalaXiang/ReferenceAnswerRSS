import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';

export function Layout() {
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="min-h-screen" style={{ background: '#F9F8F6' }}>
      <nav className="border-b" style={{ borderColor: 'rgba(28,28,28,0.1)' }}>
        <div className="max-w-5xl mx-auto px-6 md:px-12 py-5 flex items-center justify-between">
          <Link to="/" className="hover-underline">
            <h1 className="text-xl md:text-2xl tracking-tight" style={{ fontFamily: 'ui-serif, Georgia, serif' }}>
              参考答案阅览室
            </h1>
          </Link>
          <div className="flex items-center gap-6 text-xs tracking-widest uppercase" style={{ fontFamily: 'ui-sans-serif, system-ui, sans-serif' }}>
            <Link
              to="/"
              className={`hover-underline pb-0.5 transition-colors ${isActive('/') ? 'opacity-100' : 'opacity-50 hover:opacity-100'}`}
            >
              Articles
            </Link>
            <Link
              to="/settings"
              className={`hover-underline pb-0.5 transition-colors ${isActive('/settings') ? 'opacity-100' : 'opacity-50 hover:opacity-100'}`}
            >
              Settings
            </Link>
            <button
              onClick={handleLogout}
              className="opacity-50 hover:opacity-100 transition-colors hover-underline pb-0.5"
            >
              Logout
            </button>
          </div>
        </div>
      </nav>

      <main className="max-w-5xl mx-auto px-6 md:px-12 py-10 md:py-16">
        <Outlet />
      </main>

      <footer className="border-t py-8 text-center" style={{ borderColor: 'rgba(28,28,28,0.1)' }}>
        <p className="text-xs tracking-widest uppercase opacity-40">
          ReferenceAnswerRSS &middot; Powered by Xinzhi
        </p>
      </footer>
    </div>
  );
}

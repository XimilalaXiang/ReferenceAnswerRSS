import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api, type Article } from '../api';

export function Articles() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [loading, setLoading] = useState(true);
  const pageSize = 20;

  useEffect(() => {
    loadArticles();
  }, [page, keyword]);

  const loadArticles = async () => {
    setLoading(true);
    try {
      const res = await api.listArticles(page, pageSize, keyword);
      setArticles(res.articles || []);
      setTotal(res.total);
    } catch (err) {
      console.error('Failed to load articles:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
    setKeyword(searchInput);
  };

  const totalPages = Math.ceil(total / pageSize);

  const formatDate = (ts: number) => {
    const d = new Date(ts);
    return d.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  return (
    <div className="editorial-fade-in">
      <div className="flex flex-col md:flex-row md:items-end justify-between mb-10 gap-4">
        <div>
          <p className="text-xs tracking-widest uppercase opacity-40 mb-1">
            {total} Articles
          </p>
          <h2
            className="text-2xl md:text-3xl tracking-tight"
            style={{ fontFamily: 'ui-serif, Georgia, serif' }}
          >
            Latest Readings
          </h2>
        </div>
        <form onSubmit={handleSearch} className="flex gap-0">
          <input
            type="text"
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            placeholder="Search articles..."
            className="border px-4 py-2 text-sm bg-transparent transition-colors focus:outline-none w-48 md:w-64"
            style={{ borderColor: 'rgba(28,28,28,0.15)' }}
            onFocus={(e) => (e.target.style.borderColor = '#1C1C1C')}
            onBlur={(e) => (e.target.style.borderColor = 'rgba(28,28,28,0.15)')}
          />
          <button
            type="submit"
            className="px-4 py-2 text-xs tracking-widest uppercase transition-colors border"
            style={{
              background: '#1C1C1C',
              color: '#F9F8F6',
              borderColor: '#1C1C1C',
            }}
          >
            Search
          </button>
        </form>
      </div>

      {loading ? (
        <div className="py-20 text-center opacity-40 text-sm">Loading...</div>
      ) : articles.length === 0 ? (
        <div className="py-20 text-center">
          <p className="text-sm opacity-40">
            {keyword ? 'No articles found' : 'No articles yet. Wait for first sync.'}
          </p>
        </div>
      ) : (
        <div className="divide-y" style={{ borderColor: 'rgba(28,28,28,0.08)' }}>
          {articles.map((article, i) => (
            <Link
              key={article.id}
              to={`/articles/${article.id}`}
              className="block py-6 md:py-8 group transition-colors"
              style={{
                borderColor: 'rgba(28,28,28,0.08)',
                animationDelay: `${i * 50}ms`,
              }}
            >
              <div className="flex items-start gap-4 md:gap-6">
                <span
                  className="text-xs tabular-nums opacity-20 mt-1.5 hidden md:block w-8"
                  style={{ fontFamily: 'ui-monospace, monospace' }}
                >
                  {String((page - 1) * pageSize + i + 1).padStart(2, '0')}
                </span>
                <div className="flex-1 min-w-0">
                  <h3
                    className="text-lg md:text-xl tracking-tight mb-1.5 group-hover:opacity-60 transition-opacity"
                    style={{ fontFamily: 'ui-serif, Georgia, serif' }}
                  >
                    {article.title}
                  </h3>
                  {article.description && (
                    <p className="text-sm opacity-50 line-clamp-2 leading-relaxed">
                      {article.description}
                    </p>
                  )}
                  <p
                    className="text-xs opacity-30 mt-2"
                    style={{ fontFamily: 'ui-monospace, monospace' }}
                  >
                    {formatDate(article.createdAt)}
                  </p>
                </div>
                <span className="text-xs opacity-0 group-hover:opacity-40 transition-opacity mt-2">
                  &rarr;
                </span>
              </div>
            </Link>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4 mt-12 pt-8 border-t" style={{ borderColor: 'rgba(28,28,28,0.08)' }}>
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="text-xs tracking-widest uppercase opacity-40 hover:opacity-100 transition-opacity disabled:opacity-15"
          >
            &larr; Prev
          </button>
          <span className="text-xs tabular-nums opacity-40" style={{ fontFamily: 'ui-monospace, monospace' }}>
            {page} / {totalPages}
          </span>
          <button
            onClick={() => setPage(Math.min(totalPages, page + 1))}
            disabled={page === totalPages}
            className="text-xs tracking-widest uppercase opacity-40 hover:opacity-100 transition-opacity disabled:opacity-15"
          >
            Next &rarr;
          </button>
        </div>
      )}
    </div>
  );
}

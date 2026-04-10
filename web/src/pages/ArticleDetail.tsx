import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import Markdown from 'react-markdown';
import { api, type Article } from '../api';

export function ArticleDetail() {
  const { id } = useParams<{ id: string }>();
  const [article, setArticle] = useState<Article | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    api.getArticle(id)
      .then(setArticle)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  const formatDate = (ts: number) => {
    const d = new Date(ts);
    return d.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  if (loading) {
    return <div className="py-20 text-center opacity-40 text-sm">Loading...</div>;
  }

  if (error || !article) {
    return (
      <div className="py-20 text-center">
        <p className="text-sm opacity-50">{error || 'Article not found'}</p>
        <Link to="/" className="text-xs hover-underline opacity-40 mt-4 inline-block">
          &larr; Back to articles
        </Link>
      </div>
    );
  }

  return (
    <article className="editorial-fade-in">
      <Link
        to="/"
        className="text-xs tracking-widest uppercase opacity-40 hover:opacity-100 transition-opacity hover-underline inline-block mb-8"
      >
        &larr; Back to articles
      </Link>

      <header className="mb-10 pb-8 border-b" style={{ borderColor: 'rgba(28,28,28,0.1)' }}>
        <p className="text-xs tracking-widest uppercase opacity-30 mb-3">
          {article.authorName || 'Unknown Author'}
        </p>
        <h1
          className="text-3xl md:text-4xl lg:text-5xl tracking-tight leading-tight mb-4"
          style={{ fontFamily: 'ui-serif, Georgia, serif' }}
        >
          {article.title}
        </h1>
        {article.description && (
          <p className="text-base opacity-50 leading-relaxed max-w-2xl">
            {article.description}
          </p>
        )}
        <div className="flex items-center gap-4 mt-4 text-xs opacity-30" style={{ fontFamily: 'ui-monospace, monospace' }}>
          <span>{formatDate(article.createdAt)}</span>
          {article.link && (
            <>
              <span>&middot;</span>
              <a
                href={article.link}
                target="_blank"
                rel="noopener noreferrer"
                className="hover-underline hover:opacity-100 transition-opacity"
              >
                View Original
              </a>
            </>
          )}
        </div>
      </header>

      <div className="prose-editorial max-w-none text-sm md:text-base" style={{ color: 'rgba(28,28,28,0.8)' }}>
        {article.markdown ? (
          <Markdown>{article.markdown}</Markdown>
        ) : (
          <p className="opacity-40 italic">No content available.</p>
        )}
      </div>

      <footer className="mt-12 pt-8 border-t" style={{ borderColor: 'rgba(28,28,28,0.1)' }}>
        <Link
          to="/"
          className="text-xs tracking-widest uppercase opacity-40 hover:opacity-100 transition-opacity hover-underline"
        >
          &larr; Back to articles
        </Link>
      </footer>
    </article>
  );
}

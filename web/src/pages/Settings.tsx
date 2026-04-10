import { useEffect, useState } from 'react';
import { api, type SettingsResponse } from '../api';

export function Settings() {
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [copied, setCopied] = useState('');

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const res = await api.getSettings();
      setSettings(res);
    } catch (err) {
      console.error('Failed to load settings:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSync = async () => {
    setSyncing(true);
    try {
      await api.triggerSync();
      setTimeout(loadSettings, 2000);
    } catch (err) {
      console.error('Sync failed:', err);
    } finally {
      setSyncing(false);
    }
  };

  const handleRegenerateToken = async () => {
    if (!confirm('Regenerate feed token? The old RSS URL will stop working.')) return;
    try {
      const res = await api.regenerateFeedToken();
      setSettings((prev) => prev ? { ...prev, ...res } : null);
    } catch (err) {
      console.error('Failed to regenerate token:', err);
    }
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    setCopied(label);
    setTimeout(() => setCopied(''), 2000);
  };

  const formatTime = (ts: number) => {
    if (!ts) return 'Never';
    return new Date(ts).toLocaleString('zh-CN');
  };

  if (loading) {
    return <div className="py-20 text-center opacity-40 text-sm">Loading...</div>;
  }

  if (!settings) {
    return <div className="py-20 text-center opacity-40 text-sm">Failed to load settings</div>;
  }

  return (
    <div className="editorial-fade-in max-w-2xl">
      <p className="text-xs tracking-widest uppercase opacity-40 mb-1">Configuration</p>
      <h2
        className="text-2xl md:text-3xl tracking-tight mb-10"
        style={{ fontFamily: 'ui-serif, Georgia, serif' }}
      >
        Settings
      </h2>

      <section className="mb-12">
        <h3
          className="text-lg tracking-tight mb-4 pb-2 border-b"
          style={{ fontFamily: 'ui-serif, Georgia, serif', borderColor: 'rgba(28,28,28,0.1)' }}
        >
          RSS Feed
        </h3>
        <p className="text-sm opacity-50 mb-4 leading-relaxed">
          Use these URLs to subscribe in FreshRSS or any other RSS reader.
        </p>

        <div className="space-y-4">
          <FeedUrlRow
            label="RSS 2.0"
            url={settings.feedURL}
            copied={copied === 'rss'}
            onCopy={() => copyToClipboard(settings.feedURL, 'rss')}
          />
          <FeedUrlRow
            label="Atom"
            url={settings.atomURL}
            copied={copied === 'atom'}
            onCopy={() => copyToClipboard(settings.atomURL, 'atom')}
          />
        </div>

        <button
          onClick={handleRegenerateToken}
          className="mt-4 text-xs tracking-widest uppercase opacity-40 hover:opacity-100 transition-opacity hover-underline"
        >
          Regenerate Token
        </button>
      </section>

      <section className="mb-12">
        <h3
          className="text-lg tracking-tight mb-4 pb-2 border-b"
          style={{ fontFamily: 'ui-serif, Georgia, serif', borderColor: 'rgba(28,28,28,0.1)' }}
        >
          Sync Status
        </h3>

        <div className="space-y-3 text-sm">
          <div className="flex justify-between">
            <span className="opacity-50">Status</span>
            <span className={settings.syncStatus.isRunning ? 'opacity-100' : 'opacity-70'}>
              {settings.syncStatus.isRunning ? 'Syncing...' : 'Idle'}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="opacity-50">Articles</span>
            <span style={{ fontFamily: 'ui-monospace, monospace' }}>
              {settings.syncStatus.articleCount}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="opacity-50">Last Sync</span>
            <span style={{ fontFamily: 'ui-monospace, monospace', fontSize: '0.8rem' }}>
              {formatTime(settings.syncStatus.lastSyncAt)}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="opacity-50">Next Sync</span>
            <span style={{ fontFamily: 'ui-monospace, monospace', fontSize: '0.8rem' }}>
              {formatTime(settings.syncStatus.nextSyncAt)}
            </span>
          </div>
          {settings.syncStatus.lastError && (
            <div className="flex justify-between">
              <span className="opacity-50">Last Error</span>
              <span className="text-xs opacity-60 max-w-xs text-right">
                {settings.syncStatus.lastError}
              </span>
            </div>
          )}
        </div>

        <button
          onClick={handleSync}
          disabled={syncing || settings.syncStatus.isRunning}
          className="mt-6 px-6 py-3 text-xs tracking-widest uppercase transition-colors"
          style={{
            background: '#1C1C1C',
            color: '#F9F8F6',
            opacity: syncing || settings.syncStatus.isRunning ? 0.4 : 1,
          }}
        >
          {syncing ? 'Triggering...' : 'Sync Now'}
        </button>
      </section>
    </div>
  );
}

function FeedUrlRow({
  label,
  url,
  copied,
  onCopy,
}: {
  label: string;
  url: string;
  copied: boolean;
  onCopy: () => void;
}) {
  return (
    <div>
      <p className="text-xs tracking-widest uppercase opacity-40 mb-1">{label}</p>
      <div className="flex items-center gap-2">
        <input
          readOnly
          value={url}
          className="flex-1 border px-3 py-2 text-xs bg-transparent transition-colors focus:outline-none"
          style={{
            borderColor: 'rgba(28,28,28,0.15)',
            fontFamily: 'ui-monospace, monospace',
          }}
          onClick={(e) => (e.target as HTMLInputElement).select()}
        />
        <button
          onClick={onCopy}
          className="px-3 py-2 text-xs border transition-colors"
          style={{ borderColor: 'rgba(28,28,28,0.15)' }}
        >
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
    </div>
  );
}

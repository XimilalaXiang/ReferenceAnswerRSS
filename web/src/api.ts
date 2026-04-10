const API_BASE = '/api';

function getToken(): string | null {
  return localStorage.getItem('token');
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401) {
    localStorage.removeItem('token');
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export interface Article {
  id: string;
  xinzhiId: string;
  title: string;
  link: string;
  description: string;
  markdown: string;
  authorId: string;
  authorName: string;
  createdAt: number;
  syncedAt: number;
}

export interface ArticleListResponse {
  articles: Article[];
  total: number;
  page: number;
  pageSize: number;
}

export interface LoginResponse {
  token: string;
  username: string;
  feedToken: string;
}

export interface SettingsResponse {
  feedToken: string;
  feedURL: string;
  atomURL: string;
  syncStatus: SyncStatus;
}

export interface SyncStatus {
  lastSyncAt: number;
  nextSyncAt: number;
  articleCount: number;
  lastError: string;
  isRunning: boolean;
}

export const api = {
  login: (username: string, password: string) =>
    request<LoginResponse>('/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  listArticles: (page = 1, pageSize = 20, keyword = '') =>
    request<ArticleListResponse>(
      `/articles?page=${page}&pageSize=${pageSize}&keyword=${encodeURIComponent(keyword)}`
    ),

  getArticle: (id: string) =>
    request<Article>(`/articles/${id}`),

  getSettings: () =>
    request<SettingsResponse>('/settings'),

  triggerSync: () =>
    request<{ message: string }>('/sync', { method: 'POST' }),

  getSyncStatus: () =>
    request<SyncStatus>('/sync/status'),

  regenerateFeedToken: () =>
    request<{ feedToken: string; feedURL: string; atomURL: string }>(
      '/feed-token/regenerate',
      { method: 'POST' }
    ),
};

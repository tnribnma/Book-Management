const state = {
  apiBase: localStorage.getItem('bm_api_base') || 'http://localhost:8080',
  token: localStorage.getItem('bm_token') || '',
  user: JSON.parse(localStorage.getItem('bm_user') || 'null'),
};

function $(sel, root = document) { return root.querySelector(sel); }
function $all(sel, root = document) { return [...root.querySelectorAll(sel)]; }

function toast(msg, isError = false) {
  const el = $('#toast');
  el.textContent = msg;
  el.classList.toggle('error', isError);
  el.classList.add('show');
  clearTimeout(toast._t);
  toast._t = setTimeout(() => el.classList.remove('show'), 3200);
}

function logLine(line) {
  const el = $('#requestLog');
  const ts = new Date().toLocaleTimeString();
  el.textContent = `[${ts}] ${line}\n` + el.textContent;
}

function fmt(v) {
  if (v === null || v === undefined || v === '') return '—';
  return v;
}

function pick(obj, keys, fallback = null) {
  for (const k of keys) {
    if (obj && obj[k] !== undefined && obj[k] !== null) return obj[k];
  }
  return fallback;
}

function unwrap(json, ...keys) {
  if (json == null) return json;
  for (const k of keys) {
    if (json[k] !== undefined) return json[k];
  }
  return json;
}

async function api(method, path, body, { auth = true } = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (auth && state.token) headers['Authorization'] = `Bearer ${state.token}`;

  let res, data, text;
  try {
    res = await fetch(state.apiBase + path, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch (err) {
    logLine(`${method} ${path} -> NETWORK ERROR: ${err.message}`);
    toast(`Network error: cannot reach ${state.apiBase}`, true);
    throw err;
  }

  text = await res.text();
  try { data = text ? JSON.parse(text) : {}; } catch { data = { raw: text }; }

  logLine(`${method} ${path} -> ${res.status}`);

  if (!res.ok) {
    const msg = pick(data, ['error', 'message'], `HTTP ${res.status}`);
    toast(msg, true);
    const err = new Error(msg);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

function setSession(token, user) {
  state.token = token || '';
  state.user = user || null;
  if (token) localStorage.setItem('bm_token', token); else localStorage.removeItem('bm_token');
  if (user) localStorage.setItem('bm_user', JSON.stringify(user)); else localStorage.removeItem('bm_user');
  renderSession();
}

function renderSession() {
  const label = $('#userLabel');
  const logoutBtn = $('#logoutBtn');
  if (state.token) {
    const name = pick(state.user || {}, ['username', 'email'], 'signed in');
    const role = pick(state.user || {}, ['role'], null);
    label.textContent = role ? `${name} (${role})` : name;
    logoutBtn.classList.remove('hidden');
  } else {
    label.textContent = 'not signed in';
    logoutBtn.classList.add('hidden');
  }
}

function decodeJwtPayload(token) {
  try {
    const part = token.split('.')[1];
    const json = atob(part.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(json);
  } catch { return null; }
}

function initTabs() {
  $all('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      $all('.tab-btn').forEach(b => b.classList.remove('active'));
      $all('.panel').forEach(p => p.classList.remove('active'));
      btn.classList.add('active');
      $('#tab-' + btn.dataset.tab).classList.add('active');
    });
  });
}

function initAuth() {
  $('#apiBaseInput').value = state.apiBase;
  $('#saveBaseBtn').addEventListener('click', () => {
    state.apiBase = $('#apiBaseInput').value.trim().replace(/\/$/, '');
    localStorage.setItem('bm_api_base', state.apiBase);
    toast('API base saved: ' + state.apiBase);
  });

  $('#registerForm').addEventListener('submit', async e => {
    e.preventDefault();
    const f = new FormData(e.target);
    const body = { username: f.get('username'), email: f.get('email'), password: f.get('password') };
    try {
      await api('POST', '/users/register', body, { auth: false });
      toast('Registered. Now log in.');
      e.target.reset();
    } catch {  }
  });

  $('#loginForm').addEventListener('submit', async e => {
    e.preventDefault();
    const f = new FormData(e.target);
    const body = { email: f.get('email'), password: f.get('password') };
    try {
      const data = await api('POST', '/auth/login', body, { auth: false });
      const token = pick(data, ['token', 'access_token', 'jwt']);
      let user = pick(data, ['user']);
      if (!user && token) user = decodeJwtPayload(token);
      if (!token) throw new Error('Login response had no token field');
      setSession(token, user);
      toast('Logged in.');
      e.target.reset();
      loadProfile();
    } catch {  }
  });

  $('#logoutBtn').addEventListener('click', () => {
    setSession(null, null);
    toast('Logged out.');
  });
}

function bookRow(b) {
  const id = pick(b, ['id', 'book_id']);
  const title = pick(b, ['title']);
  const author = pick(b, ['author']);
  const isbn = pick(b, ['isbn']);
  const copies = pick(b, ['total_copies', 'copies']);
  const available = pick(b, ['available_copies']);
  const copiesStr = available !== null ? `${fmt(available)} / ${fmt(copies)}` : fmt(copies);

  const tr = document.createElement('tr');
  tr.innerHTML = `
    <td>${fmt(id)}</td>
    <td>${fmt(title)}</td>
    <td>${fmt(author)}</td>
    <td>${fmt(isbn)}</td>
    <td>${copiesStr}</td>
    <td class="row-actions"></td>
  `;
  const actions = $('.row-actions', tr);

  const editBtn = document.createElement('button');
  editBtn.className = 'btn small';
  editBtn.textContent = 'edit copies';
  editBtn.addEventListener('click', async () => {
    const val = prompt('New total_copies value:', copies ?? '1');
    if (val === null) return;
    try {
      await api('PUT', `/books/${id}`, { total_copies: Number(val) });
      toast('Book updated.');
      loadBooks();
    } catch {}
  });

  const delBtn = document.createElement('button');
  delBtn.className = 'btn small danger';
  delBtn.textContent = 'delete';
  delBtn.addEventListener('click', async () => {
    if (!confirm(`Delete book #${id}?`)) return;
    try {
      await api('DELETE', `/books/${id}`);
      toast('Book deleted.');
      loadBooks();
    } catch {}
  });

  actions.append(editBtn, delBtn);
  return tr;
}

async function loadBooks() {
  const tbody = $('#booksTable tbody');
  try {
    const data = await api('GET', '/books');
    const list = unwrap(data, 'books', 'data');
    const filter = $('#bookSearch').value.trim().toLowerCase();
    tbody.innerHTML = '';
    const rows = (Array.isArray(list) ? list : []).filter(b => {
      if (!filter) return true;
      return String(pick(b, ['id'])).includes(filter) ||
             String(pick(b, ['title'], '')).toLowerCase().includes(filter);
    });
    if (!rows.length) {
      tbody.innerHTML = '<tr><td class="empty" colspan="6">no books</td></tr>';
      return;
    }
    rows.forEach(b => tbody.appendChild(bookRow(b)));
  } catch {
    tbody.innerHTML = '<tr><td class="empty" colspan="6">failed to load</td></tr>';
  }
}

function initBooks() {
  $('#createBookForm').addEventListener('submit', async e => {
    e.preventDefault();
    const f = new FormData(e.target);
    const body = {
      title: f.get('title'),
      author: f.get('author'),
      isbn: f.get('isbn') || undefined,
      category_id: f.get('category_id') ? Number(f.get('category_id')) : undefined,
      total_copies: f.get('total_copies') ? Number(f.get('total_copies')) : 1,
    };
    try {
      await api('POST', '/books', body);
      toast('Book created.');
      e.target.reset();
      loadBooks();
    } catch {}
  });

  $('#refreshBooksBtn').addEventListener('click', loadBooks);
  $('#bookSearch').addEventListener('input', loadBooks);
}

async function loadCategories() {
  const tbody = $('#categoriesTable tbody');
  try {
    const data = await api('GET', '/categories');
    const list = unwrap(data, 'categories', 'data');
    tbody.innerHTML = '';
    const rows = Array.isArray(list) ? list : [];
    if (!rows.length) {
      tbody.innerHTML = '<tr><td class="empty" colspan="3">no categories</td></tr>';
      return;
    }
    rows.forEach(c => {
      const tr = document.createElement('tr');
      tr.innerHTML = `<td>${fmt(pick(c, ['id']))}</td><td>${fmt(pick(c, ['name']))}</td><td>${fmt(pick(c, ['description']))}</td>`;
      tbody.appendChild(tr);
    });
  } catch {
    tbody.innerHTML = '<tr><td class="empty" colspan="3">failed to load</td></tr>';
  }
}

function initCategories() {
  $('#createCategoryForm').addEventListener('submit', async e => {
    e.preventDefault();
    const f = new FormData(e.target);
    const body = { name: f.get('name'), description: f.get('description') || undefined };
    try {
      await api('POST', '/categories', body);
      toast('Category created.');
      e.target.reset();
      loadCategories();
    } catch {}
  });
  $('#refreshCategoriesBtn').addEventListener('click', loadCategories);
}

async function loadBorrowings() {
  const tbody = $('#borrowingsTable tbody');
  try {
    const data = await api('GET', '/my-borrowings');
    const list = unwrap(data, 'borrowings', 'data');
    tbody.innerHTML = '';
    const rows = Array.isArray(list) ? list : [];
    if (!rows.length) {
      tbody.innerHTML = '<tr><td class="empty" colspan="6">no borrowings</td></tr>';
      return;
    }
    rows.forEach(r => {
      const tr = document.createElement('tr');
      const book = pick(r, ['book_title', 'title']) ?? pick(r, ['book_id']);
      tr.innerHTML = `
        <td>${fmt(pick(r, ['id']))}</td>
        <td>${fmt(book)}</td>
        <td>${fmt(pick(r, ['borrowed_at', 'borrow_date']))}</td>
        <td>${fmt(pick(r, ['due_date']))}</td>
        <td>${fmt(pick(r, ['returned_at', 'return_date']))}</td>
        <td>${fmt(pick(r, ['fine_amount', 'fine']))}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch {
    tbody.innerHTML = '<tr><td class="empty" colspan="6">failed to load</td></tr>';
  }
}

function initBorrowing() {
  $('#borrowForm').addEventListener('submit', async e => {
    e.preventDefault();
    const f = new FormData(e.target);
    try {
      await api('POST', '/borrow', { book_id: Number(f.get('book_id')) });
      toast('Book borrowed.');
      e.target.reset();
      loadBorrowings();
    } catch {}
  });

  $('#returnForm').addEventListener('submit', async e => {
    e.preventDefault();
    const f = new FormData(e.target);
    try {
      await api('POST', '/return', { borrow_id: Number(f.get('borrow_id')) });
      toast('Book returned.');
      e.target.reset();
      loadBorrowings();
    } catch {}
  });

  $('#reserveForm').addEventListener('submit', async e => {
    e.preventDefault();
    const f = new FormData(e.target);
    try {
      await api('POST', '/reserve', { book_id: Number(f.get('book_id')) });
      toast('Book reserved.');
      e.target.reset();
    } catch {}
  });

  $('#refreshBorrowingsBtn').addEventListener('click', loadBorrowings);
}

async function loadProfile() {
  const out = $('#profileOutput');
  try {
    const data = await api('GET', '/profile');
    out.textContent = JSON.stringify(data, null, 2);
    const user = unwrap(data, 'user', 'data');
    if (user && typeof user === 'object') {
      state.user = user;
      localStorage.setItem('bm_user', JSON.stringify(user));
      renderSession();
    }
  } catch {
    out.textContent = 'failed to load';
  }
}

function initProfile() {
  $('#refreshProfileBtn').addEventListener('click', loadProfile);
}

async function loadUsers() {
  const tbody = $('#usersTable tbody');
  try {
    const data = await api('GET', '/admin/users');
    const list = unwrap(data, 'users', 'data');
    tbody.innerHTML = '';
    const rows = Array.isArray(list) ? list : [];
    if (!rows.length) {
      tbody.innerHTML = '<tr><td class="empty" colspan="4">no users</td></tr>';
      return;
    }
    rows.forEach(u => {
      const tr = document.createElement('tr');
      tr.innerHTML = `
        <td>${fmt(pick(u, ['id']))}</td>
        <td>${fmt(pick(u, ['username']))}</td>
        <td>${fmt(pick(u, ['email']))}</td>
        <td>${fmt(pick(u, ['role']))}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch {
    tbody.innerHTML = '<tr><td class="empty" colspan="4">failed to load (admin only)</td></tr>';
  }
}

function initAdmin() {
  $('#refreshUsersBtn').addEventListener('click', loadUsers);
}

async function loadDashboard() {
  const out = $('#dashboardOutput');
  try {
    const data = await api('GET', '/reports/dashboard');
    out.textContent = JSON.stringify(data, null, 2);
  } catch {
    out.textContent = 'failed to load (admin/librarian only)';
  }
}

function initReports() {
  $('#refreshDashboardBtn').addEventListener('click', loadDashboard);
}

function initLog() {
  $('#clearLogBtn').addEventListener('click', () => { $('#requestLog').textContent = ''; });
}

function init() {
  initTabs();
  initAuth();
  initBooks();
  initCategories();
  initBorrowing();
  initProfile();
  initAdmin();
  initReports();
  initLog();
  renderSession();
  if (state.token) {
    loadProfile();
  }
  loadBooks();
  loadCategories();
}

document.addEventListener('DOMContentLoaded', init);

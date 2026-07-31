CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user' CHECK (role IN ('admin', 'librarian', 'user')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    isbn VARCHAR(20) UNIQUE,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    publisher VARCHAR(255),
    edition VARCHAR(50),
    published_year INTEGER,
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity >= 0),
    available_copies INTEGER NOT NULL DEFAULT 1 CHECK (available_copies >= 0),
    shelf VARCHAR(50),
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'borrowed', 'lost', 'reserved')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS borrowings (
    id SERIAL PRIMARY KEY,
    book_id INTEGER REFERENCES books(id) ON DELETE CASCADE NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    issue_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMP NOT NULL,
    return_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'borrowed' CHECK (status IN ('borrowed', 'returned', 'overdue', 'lost')),
    fine_amount DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reservations (
    id SERIAL PRIMARY KEY,
    book_id INTEGER REFERENCES books(id) ON DELETE CASCADE NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    reservation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'fulfilled', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE books ADD COLUMN IF NOT EXISTS book_url VARCHAR(500);
ALTER TABLE books ADD COLUMN IF NOT EXISTS book_type VARCHAR(10) DEFAULT 'link';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'book_type_check'
  ) THEN
    ALTER TABLE books ADD CONSTRAINT book_type_check
      CHECK (book_type IN ('link', 'pdf', 'epub'));
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_books_book_url ON books(book_url) WHERE book_url IS NOT NULL;

CREATE INDEX idx_books_book_url ON books(book_url) WHERE book_url IS NOT NULL;

CREATE INDEX idx_books_title ON books(title);
CREATE INDEX idx_books_author ON books(author);
CREATE INDEX idx_books_isbn ON books(isbn);
CREATE INDEX idx_books_status ON books(status);
CREATE INDEX idx_borrowings_user ON borrowings(user_id);
CREATE INDEX idx_borrowings_book ON borrowings(book_id);
CREATE INDEX idx_borrowings_status ON borrowings(status);
CREATE INDEX idx_reservations_user ON reservations(user_id);

INSERT INTO categories (name, description) VALUES 
('Fiction', 'Fiction and Literature'),
('Non-Fiction', 'Non-fiction books'),
('Science', 'Science & Technology'),
('History', 'History & Biography'),
('Programming', 'Computer Science & Programming')
ON CONFLICT (name) DO NOTHING;

INSERT INTO books (title, author, isbn, category_id, quantity, available_copies, shelf, status, publisher, published_year)
VALUES
  ('Clean Code', 'Robert C. Martin', '9780132350884', 5, 3, 3, 'A1', 'available', 'Prentice Hall', 2008),
  ('The Pragmatic Programmer', 'Andrew Hunt', '9780135957059', 5, 2, 2, 'A2', 'available', 'Addison-Wesley', 2019),
  ('Sapiens', 'Yuval Noah Harari', '9780062316097', 4, 2, 2, 'B1', 'available', 'Harper', 2015),
  ('A Brief History of Time', 'Stephen Hawking', '9780553380163', 3, 1, 1, 'B2', 'available', 'Bantam', 1998),
  ('1984', 'George Orwell', '9780451524935', 1, 4, 4, 'C1', 'available', 'Signet', 1949)
ON CONFLICT (isbn) DO NOTHING;

INSERT INTO users (email, password_hash, full_name, role)
VALUES (
  'admin@library.local',
  '$2a$10$0NXPFZ3QYuo8rAgNaNd9H.byGM3nd6i1wg7wlnE2gjSWqOyGKFpqm',
  'System Admin',
  'admin'
)
ON CONFLICT (email) DO UPDATE
SET role = 'admin',
    full_name = EXCLUDED.full_name;
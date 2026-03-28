-- ========================================================================
-- SETUP ENUMERATION TYPES
-- ========================================================================
CREATE TYPE asset_base_type AS ENUM ('liquid', 'investment', 'physical');
CREATE TYPE liability_base_type AS ENUM ('short_term', 'long_term');
CREATE TYPE transaction_type AS ENUM ('income', 'expense');

-- ========================================================================
-- TABEL USERS (Autentikasi & Otorisasi)
-- ========================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ========================================================================
-- TABEL ASSET CATEGORIES (Master Data per User)
-- ========================================================================
CREATE TABLE asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    base_type asset_base_type NOT NULL, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name, base_type)
);

-- ========================================================================
-- TABEL ASSETS (Harta/Aset)
-- ========================================================================
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES asset_categories(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    current_value DECIMAL(15, 2) DEFAULT 0.00,
    details JSONB, 
    is_active BOOLEAN DEFAULT TRUE, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ========================================================================
-- TABEL LIABILITY CATEGORIES (Master Data Utang per User)
-- ========================================================================
CREATE TABLE liability_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    base_type liability_base_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name, base_type)
);

-- ========================================================================
-- TABEL LIABILITIES (Kewajiban/Utang)
-- ========================================================================
CREATE TABLE liabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES liability_categories(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    principal_amount DECIMAL(15, 2) NOT NULL, -- Total pinjaman awal
    remaining_balance DECIMAL(15, 2) NOT NULL, -- Sisa utang saat ini (pengurang Net Worth)
    details JSONB, -- Simpan metadata utang (bunga, tenor, tanggal jatuh tempo)
    is_active BOOLEAN DEFAULT TRUE, -- Kalau utang lunas, set false
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ========================================================================
-- TABEL TRANSACTION CATEGORIES (Master Data Pemasukan/Pengeluaran)
-- ========================================================================
CREATE TABLE transaction_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    base_type transaction_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name, base_type)
);

-- ========================================================================
-- TABEL TRANSACTIONS (Cashflow Harian yang Udah Disempurnakan)
-- ========================================================================
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    asset_id UUID REFERENCES assets(id) ON DELETE SET NULL,
    liability_id UUID REFERENCES liabilities(id) ON DELETE SET NULL,
    category_id UUID NOT NULL REFERENCES transaction_categories(id) ON DELETE RESTRICT,
    amount DECIMAL(15, 2) NOT NULL,
    transaction_date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- ========================================================================
-- TABEL NET WORTH HISTORIES (Snapshot buat Chart Line)
-- ========================================================================
CREATE TABLE net_worth_histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    recorded_date DATE NOT NULL,
    total_assets DECIMAL(15, 2) NOT NULL,
    total_liabilities DECIMAL(15, 2) NOT NULL,
    -- Fitur PostgreSQL yang otomatis ngurangin aset - liabilitas
    net_worth DECIMAL(15, 2) GENERATED ALWAYS AS (total_assets - total_liabilities) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, recorded_date)
);

-- ========================================================================
-- TABEL REFRESH TOKENS (Manajemen Sesi / Keamanan)
-- ========================================================================
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ========================================================================
-- INDEXING
-- ========================================================================
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_net_worth_date ON net_worth_histories(recorded_date);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
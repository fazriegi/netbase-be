-- ========================================================================
-- DROP TABLES IN REVERSE ORDER OF DEPENDENCY
-- ========================================================================
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS net_worth_histories CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS transaction_categories CASCADE;
DROP TABLE IF EXISTS liabilities CASCADE;
DROP TABLE IF EXISTS liability_categories CASCADE;
DROP TABLE IF EXISTS assets CASCADE;
DROP TABLE IF EXISTS asset_categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- ========================================================================
-- DROP CUSTOM ENUMERATION TYPES
-- ========================================================================
DROP TYPE IF EXISTS transaction_type CASCADE;
DROP TYPE IF EXISTS liability_base_type CASCADE;
DROP TYPE IF EXISTS asset_base_type CASCADE;

-- rollback assets
ALTER TABLE assets ALTER COLUMN current_value TYPE DECIMAL(15, 2);

-- rollback liabilities
ALTER TABLE liabilities ALTER COLUMN principal_amount TYPE DECIMAL(15, 2);
ALTER TABLE liabilities ALTER COLUMN remaining_balance TYPE DECIMAL(15, 2);

-- rollback transactions
ALTER TABLE transactions ALTER COLUMN amount TYPE DECIMAL(15, 2);

-- rollback net_worth_histories
ALTER TABLE net_worth_histories DROP COLUMN IF EXISTS net_worth;
ALTER TABLE net_worth_histories ALTER COLUMN total_assets TYPE DECIMAL(15, 2);
ALTER TABLE net_worth_histories ALTER COLUMN total_liabilities TYPE DECIMAL(15, 2);
ALTER TABLE net_worth_histories ADD COLUMN net_worth DECIMAL(15, 2) GENERATED ALWAYS AS (total_assets - total_liabilities) STORED;

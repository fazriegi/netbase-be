-- assets
ALTER TABLE assets ALTER COLUMN current_value TYPE NUMERIC(20, 2);

-- liabilities
ALTER TABLE liabilities ALTER COLUMN principal_amount TYPE NUMERIC(20, 2);
ALTER TABLE liabilities ALTER COLUMN remaining_balance TYPE NUMERIC(20, 2);

-- transactions
ALTER TABLE transactions ALTER COLUMN amount TYPE NUMERIC(20, 2);

-- net_worth_histories
ALTER TABLE net_worth_histories DROP COLUMN IF EXISTS net_worth;
ALTER TABLE net_worth_histories ALTER COLUMN total_assets TYPE NUMERIC(20, 2);
ALTER TABLE net_worth_histories ALTER COLUMN total_liabilities TYPE NUMERIC(20, 2);
ALTER TABLE net_worth_histories ADD COLUMN net_worth NUMERIC(20, 2) GENERATED ALWAYS AS (total_assets - total_liabilities) STORED;

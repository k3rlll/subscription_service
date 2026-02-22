-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- create table for subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL, -- price bigint for future, just in case...
    start_date DATE NOT NULL, 
    end_date DATE,-- will be null for active subscriptions, otherwise it will be the date when the subscription was cancelled or expired
    
    -- optional parameters for filtering and sorting/monitoring
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- index of method GetListSubs
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id 
ON subscriptions(user_id);

-- index of method CalculateTotalCost 
-- sql query loook like this: WHERE user_id = ? AND start_date <= ? AND end_date >= ?
-- will be used for method CalculateTotalCost, so we can quickly find all active subscriptions for a user in a given date range
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_dates 
ON subscriptions(user_id, start_date, end_date);

-- Optional parameter service_name is used for filtering subscriptions by service name
-- if user have a lot of subscriptions, this index will help to quickly find a specific service
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_service 
ON subscriptions(user_id, service_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS subscriptions;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP INDEX IF EXISTS idx_subscriptions_user_dates;
DROP INDEX IF EXISTS idx_subscriptions_user_service;
-- +goose StatementEnd

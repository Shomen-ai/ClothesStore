-- Richer order details: delivery method/cost, recipient and payment info, so the
-- order card can show the full breakdown (see the account "История заказов" page).
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_method VARCHAR(50)   DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_cost   NUMERIC(10,2) DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS recipient_name  VARCHAR(255)  DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method  VARCHAR(30)   DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_status  VARCHAR(20)   DEFAULT 'unpaid';

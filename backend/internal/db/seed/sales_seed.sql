-- ============================================================================
-- sales_seed.sql — dense order generation 2026-05-19 → 2026-06-30
-- ============================================================================
-- Adds 5–8 orders/day to keep dashboards and reports populated through the
-- ВКР defense window. Past dates get the existing weighted status mix;
-- future dates only get pending/confirmed (can't ship/deliver tomorrow yet).
--
-- Side effects, matching what the app does on a real checkout:
--   • product_sizes.stock_qty is decremented (except for cancelled orders);
--   • promo_codes.activations_count is incremented when a promo is applied.
--
-- IDEMPOTENT: the whole block is gated by MAX(created_at) — re-running is a
-- no-op once the first run advanced the watermark past 2026-05-19.
--
-- Apply:
--   scp backend/internal/db/seed/sales_seed.sql root@31.130.151.134:/tmp/
--   ssh root@31.130.151.134 \
--     'docker cp /tmp/sales_seed.sql app-db-1:/tmp/ \
--      && docker exec app-db-1 psql -U store -d clothesstore -f /tmp/sales_seed.sql'
-- ============================================================================

DO $$
DECLARE
    -- Config
    start_date    DATE := '2026-05-19';
    end_date      DATE := '2026-06-30';
    today         DATE := CURRENT_DATE;
    promo_chance  NUMERIC := 0.20;    -- fraction of orders that try a promo

    -- State
    last_existing TIMESTAMP;
    cur_day       DATE;
    n_orders      INT;
    orders_made   INT := 0;
    items_made    INT := 0;

    -- Per-order locals
    i             INT;
    j             INT;
    v_user_id     BIGINT;
    v_address_id  BIGINT;
    v_status      TEXT;
    v_n_items     INT;
    v_order_id    BIGINT;
    v_ts          TIMESTAMP;

    v_size_id     BIGINT;
    v_product_id  BIGINT;
    v_price       NUMERIC(10,2);
    v_avail       INT;
    v_qty         INT;

    v_items_total NUMERIC(12,2);

    v_promo_id    BIGINT;
    v_promo_type  TEXT;
    v_promo_value NUMERIC(10,2);
    v_discount    NUMERIC(10,2);

    r             NUMERIC;
BEGIN
    -- ── Idempotency gate ───────────────────────────────────────────────
    SELECT COALESCE(MAX(created_at), '1970-01-01'::timestamp)
      INTO last_existing FROM orders;

    IF last_existing >= start_date::timestamp THEN
        RAISE NOTICE 'sales_seed: skipping — already seeded (max order = %)', last_existing;
        RETURN;
    END IF;

    RAISE NOTICE 'sales_seed: starting % → % (today=%)', start_date, end_date, today;

    -- ── Day loop ───────────────────────────────────────────────────────
    cur_day := start_date;
    WHILE cur_day <= end_date LOOP
        n_orders := 5 + floor(random() * 4)::int;   -- 5..8

        FOR i IN 1..n_orders LOOP
            -- pick a customer
            SELECT id INTO v_user_id
              FROM users WHERE role = 'customer'
              ORDER BY random() LIMIT 1;

            IF v_user_id IS NULL THEN CONTINUE; END IF;

            -- their address (default preferred)
            SELECT id INTO v_address_id
              FROM addresses WHERE user_id = v_user_id
              ORDER BY is_default DESC NULLS LAST, random() LIMIT 1;

            IF v_address_id IS NULL THEN CONTINUE; END IF;

            -- status, weighted by date bucket
            r := random();
            IF cur_day <= today THEN
                v_status := CASE
                    WHEN r < 0.45 THEN 'delivered'
                    WHEN r < 0.66 THEN 'shipped'
                    WHEN r < 0.81 THEN 'confirmed'
                    WHEN r < 0.91 THEN 'cancelled'
                    ELSE                  'pending'
                END;
            ELSE
                v_status := CASE WHEN r < 0.60 THEN 'pending' ELSE 'confirmed' END;
            END IF;

            -- random timestamp within the day
            v_ts := cur_day::timestamp + (random() * INTERVAL '23 hours 59 minutes');

            -- create the order with zero totals; we'll fix after items
            INSERT INTO orders(user_id, address_id, status, total_price, discount_amount, created_at, promo_code_id)
                VALUES (v_user_id, v_address_id, v_status, 0, 0, v_ts, NULL)
                RETURNING id INTO v_order_id;

            -- ── items ──────────────────────────────────────────────────
            v_n_items     := 1 + floor(random() * 4)::int;    -- 1..4
            v_items_total := 0;

            FOR j IN 1..v_n_items LOOP
                SELECT ps.id, ps.product_id, p.price, ps.stock_qty
                  INTO v_size_id, v_product_id, v_price, v_avail
                  FROM product_sizes ps
                  JOIN products p ON p.id = ps.product_id
                  WHERE ps.stock_qty > 0 AND p.is_active = TRUE
                  ORDER BY random() LIMIT 1;

                IF v_size_id IS NULL THEN CONTINUE; END IF;

                v_qty := LEAST(1 + floor(random() * 3)::int, v_avail);   -- 1..3, capped
                IF v_qty <= 0 THEN CONTINUE; END IF;

                INSERT INTO order_items(order_id, product_id, product_size_id, quantity, price_at_order)
                    VALUES (v_order_id, v_product_id, v_size_id, v_qty, v_price);

                v_items_total := v_items_total + (v_price * v_qty);
                items_made    := items_made + 1;

                -- decrement stock unless cancelled (matches app behavior)
                IF v_status <> 'cancelled' THEN
                    UPDATE product_sizes SET stock_qty = stock_qty - v_qty WHERE id = v_size_id;
                END IF;
            END LOOP;

            -- if every item was rejected (rare), drop the empty order
            IF v_items_total = 0 THEN
                DELETE FROM orders WHERE id = v_order_id;
                CONTINUE;
            END IF;

            -- ── promo (chance, only for non-cancelled to avoid weird ROI) ─
            v_promo_id    := NULL;
            v_promo_type  := NULL;
            v_promo_value := 0;
            v_discount    := 0;

            IF random() < promo_chance AND v_status <> 'cancelled' THEN
                SELECT pc.id, pc.discount_type, pc.discount_value
                  INTO v_promo_id, v_promo_type, v_promo_value
                  FROM promo_codes pc
                  WHERE pc.is_active = TRUE
                    AND (pc.expires_at IS NULL OR pc.expires_at > v_ts)
                    AND (pc.max_activations IS NULL OR pc.activations_count < pc.max_activations)
                  ORDER BY random() LIMIT 1;

                IF v_promo_id IS NOT NULL THEN
                    IF v_promo_type = 'percent' THEN
                        v_discount := round(v_items_total * v_promo_value / 100, 2);
                    ELSE  -- 'fixed'
                        v_discount := LEAST(v_promo_value, v_items_total);
                    END IF;

                    UPDATE promo_codes
                       SET activations_count = activations_count + 1
                     WHERE id = v_promo_id;
                END IF;
            END IF;

            -- finalize totals on the order
            UPDATE orders
               SET total_price     = v_items_total - v_discount,
                   discount_amount = v_discount,
                   promo_code_id   = v_promo_id
             WHERE id = v_order_id;

            orders_made := orders_made + 1;
        END LOOP;

        cur_day := cur_day + 1;
    END LOOP;

    RAISE NOTICE 'sales_seed: done — % orders, % items', orders_made, items_made;
END $$;

-- ============================================================================
-- reviews_seed.sql — demo product reviews for the ВКР defense
-- ============================================================================
-- Picks ~100 (user, product) pairs where the user has a delivered order
-- containing the product, then writes a review with a positively-skewed
-- star distribution (50/25/15/7/3) and ~70% chance of attached text drawn
-- from a small RU pool sized to the rating bucket.
--
-- IDEMPOTENT: gated by "any product_reviews exist already" — re-running is a
-- no-op once the table has anything in it.
--
-- Apply:
--   scp backend/internal/db/seed/reviews_seed.sql root@31.130.151.134:/tmp/
--   ssh root@31.130.151.134 \
--     'docker cp /tmp/reviews_seed.sql app-db-1:/tmp/ \
--      && docker exec app-db-1 psql -U store -d clothesstore -f /tmp/reviews_seed.sql'
-- ============================================================================

DO $$
DECLARE
    target_count   INT     := 100;
    text_prob      NUMERIC := 0.70;
    inserted_count INT     := 0;
    rc             INT;

    pos_templates  TEXT[]  := ARRAY[
      'Ткань приятная, размер сел идеально.',
      'Качество за свою цену — топ.',
      'Цвет на фото точнее, чем казалось.',
      'Доставка быстрая, упаковка крутая.',
      'Сидит как влитой.',
      'Купил повторно — настолько понравилось.',
      'Стильно, лаконично, без претензий.',
      'Точно по размерной сетке, бери смело.',
      'Подошло идеально, рекомендую.',
      'Соответствует описанию, никаких сюрпризов.',
      'Хороший пошив, нитки нигде не торчат.',
      'Носится отлично, моё.',
      'Качество выше ожиданий.',
      'Отличный материал, не электризуется.',
      'Бренд держит уровень.'
    ];

    mid_templates  TEXT[]  := ARRAY[
      'Нормально, но ожидал большего.',
      'За свою цену — ок.',
      'Носится, но не вау.',
      'После стирки слегка сел, имейте в виду.',
      'Маломерит, берите на размер больше.',
      'Ткань нормальная, ничего особенного.',
      'Среднее впечатление, но не возвращаю.',
      'Сел, но не идеально — придётся подшить.'
    ];

    neg_templates  TEXT[]  := ARRAY[
      'Швы кривые, разочарование.',
      'Цвет не совпал с фото.',
      'Качество не для такой цены.',
      'Ткань тонкая и просвечивает.',
      'Сел после первой стирки.',
      'Жмёт там, где не должно.',
      'Не то, что ожидал — буду возвращать.'
    ];

    -- Eligible pairs: user must have a delivered order containing the product.
    -- ORDER BY random() lives outside the DISTINCT (Postgres won't let random()
    -- sit in an ORDER BY of a DISTINCT select unless it's in the select list,
    -- and including it there breaks DISTINCT's grouping).
    eligible CURSOR FOR
      SELECT user_id, product_id FROM (
        SELECT DISTINCT o.user_id, oi.product_id
        FROM order_items oi
        JOIN orders o ON o.id = oi.order_id
        JOIN users  u ON u.id = o.user_id
        WHERE o.status = 'delivered' AND u.role = 'customer'
      ) e
      ORDER BY random();

    pair      RECORD;
    v_rating  INT;
    v_text    TEXT;
    v_created TIMESTAMP;
    r         NUMERIC;
BEGIN
    -- Idempotency gate.
    IF EXISTS (SELECT 1 FROM product_reviews LIMIT 1) THEN
        RAISE NOTICE 'reviews_seed: already populated — skipping';
        RETURN;
    END IF;

    OPEN eligible;
    LOOP
        FETCH eligible INTO pair;
        EXIT WHEN NOT FOUND OR inserted_count >= target_count;

        -- 5★ 50% / 4★ 25% / 3★ 15% / 2★ 7% / 1★ 3%
        r := random();
        IF    r < 0.50 THEN v_rating := 5;
        ELSIF r < 0.75 THEN v_rating := 4;
        ELSIF r < 0.90 THEN v_rating := 3;
        ELSIF r < 0.97 THEN v_rating := 2;
        ELSE                v_rating := 1;
        END IF;

        -- ~70% have text, drawn from the bucket that matches the rating tone.
        IF random() < text_prob THEN
            IF v_rating >= 4 THEN
                v_text := pos_templates[1 + floor(random() * array_length(pos_templates, 1))::int];
            ELSIF v_rating = 3 THEN
                v_text := mid_templates[1 + floor(random() * array_length(mid_templates, 1))::int];
            ELSE
                v_text := neg_templates[1 + floor(random() * array_length(neg_templates, 1))::int];
            END IF;
        ELSE
            v_text := '';
        END IF;

        -- Random timestamp within the past 30 days; updated_at matches.
        v_created := NOW() - (random() * INTERVAL '30 days');

        INSERT INTO product_reviews(user_id, product_id, rating, text, created_at, updated_at)
        VALUES (pair.user_id, pair.product_id, v_rating, v_text, v_created, v_created)
        ON CONFLICT (user_id, product_id) DO NOTHING;

        GET DIAGNOSTICS rc = ROW_COUNT;
        inserted_count := inserted_count + rc;
    END LOOP;
    CLOSE eligible;

    RAISE NOTICE 'reviews_seed: inserted % review(s)', inserted_count;
END $$;

-- Adds a singular "type" to each category so the storefront can display product
-- names as "<тип> <название>" (e.g. "Футболка Valor") without renaming products.
ALTER TABLE categories ADD COLUMN IF NOT EXISTS type_name VARCHAR(50) DEFAULT '';

UPDATE categories SET type_name = CASE name
  WHEN 'Аксессуары'           THEN 'Аксессуар'
  WHEN 'Джерси'               THEN 'Джерси'
  WHEN 'Жилеты'               THEN 'Жилет'
  WHEN 'Куртки и Ветровки'    THEN 'Куртка'
  WHEN 'Лонгсливы и Свитшоты' THEN 'Лонгслив'
  WHEN 'Носки'                THEN 'Носки'
  WHEN 'Рубашки'              THEN 'Рубашка'
  WHEN 'Свитеры и Джемперы'   THEN 'Свитер'
  WHEN 'Сумки'                THEN 'Сумка'
  WHEN 'Толстовки'            THEN 'Толстовка'
  WHEN 'Футболки'             THEN 'Футболка'
  WHEN 'Шарфы'                THEN 'Шарф'
  WHEN 'Шорты'                THEN 'Шорты'
  WHEN 'Штаны'                THEN 'Штаны'
  ELSE type_name END;

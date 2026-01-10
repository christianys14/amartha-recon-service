-- migrate:up
SET @i := 0;
CREATE PROCEDURE seed_transactions(IN total INT)
BEGIN
  DECLARE i INT DEFAULT 1;
  DECLARE ts_start DATETIME DEFAULT '2026-01-01 00:00:00';
  WHILE i <= total DO
    INSERT INTO transactions (
      transaction_id,
      terminal_rrn,
      amount,
      transaction_type,
      bank_code,
      transaction_time,
      updated_at
    ) VALUES (
      LPAD(ABS(CRC32(CONCAT(UUID(), i))), 12, '10'),
      LPAD(ABS(CRC32(CONCAT(UUID(), i))), 12, '10'), -- 12-digit numeric string unik
      ROUND(RAND() * 10000, 2),
      IF(RAND() < 0.5, 'DEBIT', 'CREDIT'),
      CASE WHEN RAND() < 1/3 THEN '014' WHEN RAND() < 2/3 THEN '008' ELSE '002' END,
      TIMESTAMPADD(HOUR, FLOOR(RAND() * 9 * 24) + FLOOR(RAND() * 24), ts_start),
      TIMESTAMPADD(HOUR, FLOOR(RAND() * 9 * 24) + FLOOR(RAND() * 24), ts_start)
    );
    SET i = i + 1;
END WHILE;
END;
CALL seed_transactions(1000);
-- migrate:down


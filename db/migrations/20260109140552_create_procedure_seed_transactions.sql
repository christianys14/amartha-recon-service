-- migrate:up
CREATE PROCEDURE seed_transactions(IN total INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    WHILE i <= total
        DO
            INSERT INTO transactions (transaction_id,
                                      terminal_rrn,
                                      amount,
                                      transaction_type,
                                      bank_code,
                                      transaction_time,
                                      updated_at)
            VALUES (LPAD(i, 12, '0'), -- 12-digit numeric string
                    LOWER(CONCAT(HEX(UUID_TO_BIN(UUID())), '-', LPAD(i, 4, '0'))), -- contoh pembentuk UUID-ish
                    ROUND((RAND() * 10000), 2),
                    IF(RAND() < 0.5, 'DEBIT', 'CREDIT'),
                    CASE WHEN RAND() < 1 / 3 THEN '014' WHEN RAND() < 2 / 3 THEN '008' ELSE '002' END,
                    NOW() - INTERVAL FLOOR(RAND() * 365) DAY,
                    NOW() - INTERVAL FLOOR(RAND() * 365) DAY);
            SET i = i + 1;
END WHILE;
END;
DELIMITER ;
CALL seed_transactions(1000);
-- migrate:down


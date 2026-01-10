-- migrate:up
select
    transactions.transaction_id,
    transactions.terminal_rrn,
    transactions.amount,
    transactions.transaction_type,
    transactions.bank_code,
    transactions.transaction_time
from transactions;

-- migrate:down


-- migrate:up
create table transactions
(
    id               bigint primary key auto_increment,
    transaction_id   varchar(255)   not null,
    terminal_rrn     varchar(255)   not null,
    amount           decimal(19, 2) not null,
    transaction_type enum ('DEBIT','CREDIT'),
    bank_code        char(3)        not null,
    transaction_time timestamp default current_timestamp,
    updated_at       timestamp default current_timestamp on update current_timestamp
);

create index idx_transaction_id_rrn on transactions (transaction_id, terminal_rrn);
create index idx_timestamp on transactions (transaction_time, updated_at);
create index idx_bank_code on transactions (bank_code);
-- migrate:down
drop transactions;
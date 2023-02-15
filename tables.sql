create extension pg_trgm; /* trigram extension for faster text-search */

/*
This table is used to store the current state (latest exported epoch) of all validators
It also acts as a lookup-table to store the index-pubkey association
In order to save db space we only use the unique validator index in all other tables
In the future it is better to replace this table with an in memory cache (redis)
*/
drop table if exists validators;
create table validators
(
    validatorindex             int         not null,
    pubkey                     bytea       not null,
    pubkeyhex                  text        not null default '',
    withdrawableepoch          bigint      not null,
    withdrawalcredentials      bytea       not null,
    balance                    bigint      not null,
    balance1d                  bigint,
    balance7d                  bigint,
    balance31d                 bigint,
    balanceactivation          bigint,
    effectivebalance           bigint      not null,
    slashed                    bool        not null,
    activationeligibilityepoch bigint      not null,
    activationepoch            bigint      not null,
    exitepoch                  bigint      not null,
    lastattestationslot        bigint,
    status                     varchar(20) not null default '',
    primary key (validatorindex)
);
create index idx_validators_pubkey on validators (pubkey);
create index idx_validators_pubkeyhex on validators (pubkeyhex);
create index idx_validators_pubkeyhex_pattern_pos on validators (pubkeyhex varchar_pattern_ops);
create index idx_validators_status on validators (status);
create index idx_validators_balanceactivation on validators (balanceactivation);
create index idx_validators_activationepoch on validators (activationepoch);
CREATE INDEX validators_is_offline_vali_idx ON validators (validatorindex, lastattestationslot, pubkey);

drop table if exists validator_pool;
create table validator_pool
(
    publickey bytea not null,
    pool      varchar(40),
    primary key (publickey)
);

drop table if exists validator_names;
create table validator_names
(
    publickey bytea not null,
    name      varchar(40),
    primary key (publickey)
);
create index idx_validator_names_publickey on validator_names (publickey);
create index idx_validator_names_name on validator_names(name);

drop table if exists validator_set;
create table validator_set
(
    epoch                      int    not null,
    validatorindex             int    not null,
    withdrawableepoch          bigint not null,
    withdrawalcredentials      bytea  not null,
    effectivebalance           bigint not null,
    slashed                    bool   not null,
    activationeligibilityepoch bigint not null,
    activationepoch            bigint not null,
    exitepoch                  bigint not null,
    primary key (validatorindex, epoch)
);

drop table if exists validator_performance;
create table validator_performance
(
    validatorindex  int    not null,
    balance         bigint not null,
    performance1d   bigint not null,
    performance7d   bigint not null,
    performance31d  bigint not null,
    performance365d bigint not null,
    rank7d          int    not null,
    primary key (validatorindex)
);
create index idx_validator_performance_balance on validator_performance (balance);
create index idx_validator_performance_performance1d on validator_performance (performance1d);
create index idx_validator_performance_performance7d on validator_performance (performance7d);
create index idx_validator_performance_performance31d on validator_performance (performance31d);
create index idx_validator_performance_performance365d on validator_performance (performance365d);
create index idx_validator_performance_rank7d on validator_performance (rank7d);

drop table if exists proposal_assignments;
create table proposal_assignments
(
    epoch          int not null,
    validatorindex int not null,
    proposerslot   int not null,
    status         int not null, /* Can be 0 = scheduled, 1 executed, 2 missed */
    primary key (epoch, validatorindex, proposerslot)
);
create index idx_proposal_assignments_epoch on proposal_assignments (epoch);

drop table if exists sync_committees;
create table sync_committees
(
    period         int not null,
    validatorindex int not null,
    committeeindex int not null,
    primary key (period, validatorindex, committeeindex)
);

drop table if exists sync_committees_count_per_validator;
create table sync_committees_count_per_validator (
	period int not null unique,
	count_so_far float8 not null,
    primary key (period)
);

drop table if exists validator_balances_recent;
create table validator_balances_recent
(
    epoch          int    not null,
    validatorindex int    not null,
    balance        bigint not null,
    primary key (epoch, validatorindex)
);
create index idx_validator_balances_recent_epoch on validator_balances_recent (epoch);
create index idx_validator_balances_recent_validatorindex on validator_balances_recent (validatorindex);
create index idx_validator_balances_recent_balance on validator_balances_recent (balance);

drop table if exists validator_stats;
create table validator_stats
(
    validatorindex          int not null,
    day                     int not null,
    start_balance           bigint,
    end_balance             bigint,
    min_balance             bigint,
    max_balance             bigint,
    start_effective_balance bigint,
    end_effective_balance   bigint,
    min_effective_balance   bigint,
    max_effective_balance   bigint,
    missed_attestations     int,
    orphaned_attestations   int,
    participated_sync       int,
    missed_sync             int,
    orphaned_sync           int,
    proposed_blocks         int,
    missed_blocks           int,
    orphaned_blocks         int,
    attester_slashings      int,
    proposer_slashings      int,
    deposits                int,
    deposits_amount         bigint,
    withdrawals             int,
    withdrawals_amount      bigint,
    primary key (validatorindex, day)
);
create index idx_validator_stats_day on validator_stats (day);

drop table if exists validator_stats_status;
create table validator_stats_status
(
    day    int     not null,
    status boolean not null,
    primary key (day)
);

drop table if exists validator_attestation_streaks;
create table validator_attestation_streaks
(
    validatorindex int     not null,
    status         int     not null,
    start          int     not null,
    length         int     not null,
    longest        boolean not null,
    current        boolean not null,
    primary key (validatorindex, status, start)
);
create index idx_validator_attestation_streaks_validatorindex on validator_attestation_streaks (validatorindex);
create index idx_validator_attestation_streaks_status on validator_attestation_streaks (status);
create index idx_validator_attestation_streaks_length on validator_attestation_streaks (length);
create index idx_validator_attestation_streaks_start on validator_attestation_streaks (start);

drop table if exists queue;
create table queue
(
    ts                        timestamp without time zone,
    entering_validators_count int not null,
    exiting_validators_count  int not null,
    primary key (ts)
);

drop table if exists validatorqueue_activation;
create table validatorqueue_activation
(
    index     int   not null,
    publickey bytea not null,
    primary key (index, publickey)
);

drop table if exists validatorqueue_exit;
create table validatorqueue_exit
(
    index     int   not null,
    publickey bytea not null,
    primary key (index, publickey)
);

create table epochs_notified (epoch int not null primary key, sentOn timestamp not null);

drop table if exists epochs;
create table epochs
(
    epoch                   int    not null,
    blockscount             int    not null default 0,
    proposerslashingscount  int    not null,
    attesterslashingscount  int    not null,
    attestationscount       int    not null,
    depositscount           int    not null,
    withdrawalcount         int    not null default 0,
    voluntaryexitscount     int    not null,
    validatorscount         int    not null,
    averagevalidatorbalance bigint not null,
    totalvalidatorbalance   bigint not null,
    finalized               bool,
    eligibleether           bigint,
    globalparticipationrate float,
    votedether              bigint,
    primary key (epoch)
);

drop table if exists blocks;
create table blocks
(
    epoch                       int     not null,
    slot                        int     not null,
    blockroot                   bytea   not null,
    parentroot                  bytea   not null,
    stateroot                   bytea   not null,
    signature                   bytea   not null,
    randaoreveal                bytea,
    graffiti                    bytea,
    graffiti_text               text    null,
    eth1data_depositroot        bytea,
    eth1data_depositcount       int     not null,
    eth1data_blockhash          bytea,
    syncaggregate_bits          bytea,
    syncaggregate_signature     bytea,
    syncaggregate_participation float   not null default 0,
    proposerslashingscount      int     not null,
    attesterslashingscount      int     not null,
    attestationscount           int     not null,
    depositscount               int     not null,
    withdrawalcount             int     not null default 0,
    voluntaryexitscount         int     not null,
    proposer                    int     not null,
    status                      text    not null, /* Can be 0 = scheduled, 1 proposed, 2 missed, 3 orphaned */

    -- https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2
    -- https://github.com/ethereum/consensus-specs/blob/v1.1.9/specs/bellatrix/beacon-chain.md#executionpayload
    exec_parent_hash            bytea, 
    exec_fee_recipient          bytea, 
    exec_state_root             bytea, 
    exec_receipts_root          bytea, 
    exec_logs_bloom             bytea, 
    exec_random                 bytea, 
    exec_block_number           int,   
    exec_gas_limit              int,   
    exec_gas_used               int,   
    exec_timestamp              int,   
    exec_extra_data             bytea, 
    exec_base_fee_per_gas       bigint,
    exec_block_hash             bytea, 
    exec_transactions_count     int     not null default 0,

    primary key (slot, blockroot)
);
create index idx_blocks_proposer on blocks (proposer);
create index idx_blocks_epoch on blocks (epoch);
create index idx_blocks_graffiti_text on blocks using gin (graffiti_text gin_trgm_ops);
create index idx_blocks_blockrootstatus on blocks (blockroot, status);
create index idx_blocks_exec_block_number on blocks (exec_block_number);

drop table if exists blocks_withdrawals;
create table blocks_withdrawals
(
    block_slot         int not null,
    block_root         bytea not null,
    withdrawalindex    int not null,
    validatorindex     int not null,
    address            bytea not null,
    amount             bigint not null, -- in GWei
    primary key (block_slot, block_root, withdrawalindex)
);

create index idx_blocks_withdrawals_recipient on blocks_withdrawals (address);
create index idx_blocks_withdrawals_validatorindex on blocks_withdrawals (validatorindex);

drop table if exists blocks_bls_change;
create table blocks_bls_change
(
    block_slot           int     not null,
    block_root           bytea   not null,
    validatorindex       int     not null,
    signature            bytea   not null,
    pubkey               bytea   not null,
    address              bytea   not null,
    primary key (block_slot, block_root, validatorindex)
);
create index idx_blocks_bls_change_pubkey on blocks_bls_change (pubkey);
create index idx_blocks_bls_change_address on blocks_bls_change (address);

drop table if exists blocks_transactions;
create table blocks_transactions
(
    block_slot         int    not null,
    block_index        int    not null,
    block_root         bytea  not null default '',
    raw                bytea  not null,
    txhash             bytea  not null,
    nonce              int    not null,
    gas_price          bytea  not null,
    gas_limit          bigint not null,
    sender             bytea  not null,
    recipient          bytea  not null,
    amount             bytea  not null,
    payload            bytea  not null,
    max_priority_fee_per_gas  bigint,
    max_fee_per_gas           bigint,
    primary key (block_slot, block_index)
);

drop table if exists blocks_proposerslashings;
create table blocks_proposerslashings
(
    block_slot         int    not null,
    block_index        int    not null,
    block_root         bytea  not null default '',
    proposerindex      int    not null,
    header1_slot       bigint not null,
    header1_parentroot bytea  not null,
    header1_stateroot  bytea  not null,
    header1_bodyroot   bytea  not null,
    header1_signature  bytea  not null,
    header2_slot       bigint not null,
    header2_parentroot bytea  not null,
    header2_stateroot  bytea  not null,
    header2_bodyroot   bytea  not null,
    header2_signature  bytea  not null,
    primary key (block_slot, block_index)
);

drop table if exists blocks_attesterslashings;
create table blocks_attesterslashings
(
    block_slot                   int       not null,
    block_index                  int       not null,
    block_root                   bytea     not null default '',
    attestation1_indices         integer[] not null,
    attestation1_signature       bytea     not null,
    attestation1_slot            bigint    not null,
    attestation1_index           int       not null,
    attestation1_beaconblockroot bytea     not null,
    attestation1_source_epoch    int       not null,
    attestation1_source_root     bytea     not null,
    attestation1_target_epoch    int       not null,
    attestation1_target_root     bytea     not null,
    attestation2_indices         integer[] not null,
    attestation2_signature       bytea     not null,
    attestation2_slot            bigint    not null,
    attestation2_index           int       not null,
    attestation2_beaconblockroot bytea     not null,
    attestation2_source_epoch    int       not null,
    attestation2_source_root     bytea     not null,
    attestation2_target_epoch    int       not null,
    attestation2_target_root     bytea     not null,
    primary key (block_slot, block_index)
);

drop table if exists blocks_attestations;
create table blocks_attestations
(
    block_slot      int   not null,
    block_index     int   not null,
    block_root      bytea not null default '',
    aggregationbits bytea not null,
    validators      int[] not null,
    signature       bytea not null,
    slot            int   not null,
    committeeindex  int   not null,
    beaconblockroot bytea not null,
    source_epoch    int   not null,
    source_root     bytea not null,
    target_epoch    int   not null,
    target_root     bytea not null,
    primary key (block_slot, block_index)
);
create index idx_blocks_attestations_beaconblockroot on blocks_attestations (beaconblockroot);
create index idx_blocks_attestations_source_root on blocks_attestations (source_root);
create index idx_blocks_attestations_target_root on blocks_attestations (target_root);

drop table if exists blocks_deposits;
create table blocks_deposits
(
    block_slot            int    not null,
    block_index           int    not null,
    block_root            bytea  not null default '',
    proof                 bytea[],
    publickey             bytea  not null,
    withdrawalcredentials bytea  not null,
    amount                bigint not null,
    signature             bytea  not null,
    primary key (block_slot, block_index)
);

drop table if exists blocks_voluntaryexits;
create table blocks_voluntaryexits
(
    block_slot     int   not null,
    block_index    int   not null,
    block_root     bytea not null default '',
    epoch          int   not null,
    validatorindex int   not null,
    signature      bytea not null,
    primary key (block_slot, block_index)
);

drop table if exists network_liveness;
create table network_liveness
(
    ts                     timestamp without time zone,
    headepoch              int not null,
    finalizedepoch         int not null,
    justifiedepoch         int not null,
    previousjustifiedepoch int not null,
    primary key (ts)
);

drop table if exists graffitiwall;
create table graffitiwall
(
    x         int  not null,
    y         int  not null,
    color     text not null,
    slot      int  not null,
    validator int  not null,
    primary key (x, y)
);

drop table if exists eth1_deposits;
create table eth1_deposits
(
    tx_hash                bytea                       not null,
    tx_input               bytea                       not null,
    tx_index               int                         not null,
    block_number           int                         not null,
    block_ts               timestamp without time zone not null,
    from_address           bytea                       not null,
    publickey              bytea                       not null,
    withdrawal_credentials bytea                       not null,
    amount                 bigint                      not null,
    signature              bytea                       not null,
    merkletree_index       bytea                       not null,
    removed                bool                        not null,
    valid_signature        bool                        not null,
    primary key (tx_hash, merkletree_index)
);
create index idx_eth1_deposits on eth1_deposits (publickey);
create index idx_eth1_deposits_from_address on eth1_deposits (from_address);

drop table if exists eth1_deposits_aggregated;
create table eth1_deposits_aggregated
(
    from_address         bytea  not null,
    amount               bigint not null,
    validcount           int    not null,
    invalidcount         int    not null,
    slashedcount         int    not null,
    totalcount           int    not null,
    activecount          int    not null,
    pendingcount         int    not null,
    voluntary_exit_count int    not null,
    primary key (from_address)
);

drop table if exists users;
create table users
(
    id                      serial                 not null unique,
    password                character varying(256) not null,
    email                   character varying(100) not null unique,
    email_confirmed         bool                   not null default 'f',
    email_confirmation_hash character varying(40) unique,
    email_confirmation_ts   timestamp without time zone,
    password_reset_hash     character varying(40),
    password_reset_ts       timestamp without time zone,
    register_ts             timestamp without time zone,
    api_key                 character varying(256) unique,
    stripe_customer_id      character varying(256) unique,
    user_group              varchar(10),
    primary key (id, email)
);

drop table if exists users_datatable;
create table users_datatable
(
    user_id        int                         not null,
    key            character varying(256)      not null,
    state          jsonb                       not null,
    updated_at     timestamp without time zone not null default 'now()',
    primary key (user_id, key) 
);

drop table if exists users_stripe_subscriptions;
create table users_stripe_subscriptions
(
    subscription_id character varying(256) unique not null,
    customer_id     character varying(256)        not null,
    price_id        character varying(256)        not null,
    active          bool not null default 'f',
    payload         json                          not null,
    purchase_group    character varying(30)         not null default 'api',
    primary key (customer_id, subscription_id, price_id)
);

drop table if exists users_app_subscriptions;
create table users_app_subscriptions
(
    id              serial                        not null,
    user_id         int                           not null,
    product_id      character varying(256)        not null,
    price_micros    bigint                        not null,
    currency        character varying(10)         not null,
    created_at      timestamp without time zone   not null,
    updated_at      timestamp without time zone   not null,
    validate_remotely boolean not null default 't',
    active          bool not null default 'f',
    store           character varying(50)         not null,
    expires_at      timestamp without time zone   not null,
    reject_reason   character varying(50),
    receipt         character varying(99999)       not null,
    receipt_hash    character varying(1024)        not null unique,
    subscription_id character varying(256)         default ''
);
create index idx_user_app_subscriptions on users_app_subscriptions (user_id);

drop table if exists oauth_apps;
create table oauth_apps
(
    id           serial                      not null,
    owner_id     int                         not null,
    redirect_uri character varying(100)      not null unique,
    app_name     character varying(35)       not null,
    active       bool                        not null default 't',
    created_ts   timestamp without time zone not null,
    primary key (id, redirect_uri)
);

drop table if exists oauth_codes;
create table oauth_codes
(
    id         serial                      not null,
    user_id    int                         not null,
    code       character varying(64)       not null,
    client_id  character varying(128)      not null,
    consumed   bool                        not null default 'f',
    app_id     int                         not null,
    created_ts timestamp without time zone not null,
    primary key (user_id, app_id, client_id)
);

drop table if exists users_devices;
create table users_devices
(
    id                 serial                      not null,
    user_id            int                         not null,
    refresh_token      character varying(64)       not null,
    device_name        character varying(20)       not null,
    notification_token character varying(500),
    notify_enabled     bool                        not null default 't',
    active             bool                        not null default 't',
    app_id             int                         not null,
    created_ts         timestamp without time zone not null,
    primary key (user_id, refresh_token)
);

drop table if exists users_clients;
create table users_clients
(
    id             serial                      not null,
    user_id        int                         not null,
    client         character varying(12)       not null,
    client_version int                         not null,
    notify_enabled bool                        not null default 't',
    created_ts     timestamp without time zone not null,
    primary key (user_id, client)
);


drop table if exists users_subscriptions;
create table users_subscriptions
(
    id                serial                      not null,
    user_id           int                         not null,
    event_name        character varying(100)      not null,
    event_filter      text                        not null default '',
    event_threshold   real                        default 0,
    last_sent_ts      timestamp without time zone,
    last_sent_epoch   int,
    created_ts        timestamp without time zone not null,
    created_epoch     int                         not null,
    unsubscribe_hash  bytea,
    internal_state    varchar,
    primary key (user_id, event_name, event_filter)
);
create index idx_users_subscriptions_unsubscribe_hash on users_subscriptions (unsubscribe_hash);

CREATE TYPE notification_channels as ENUM ('webhook_discord', 'webhook', 'email', 'push');

drop table if exists users_notification_channels;
create table users_notification_channels
(
    user_id int                   not null,
    channel notification_channels not null,
    active  boolean default 't'   not null,
    primary key (user_id, channel)
);

drop table if exists notification_queue;
create table notification_queue(
    id                  serial not null,
    created             timestamp without time zone not null,
    sent                timestamp without time zone, -- record when the transaction was dispatched
    -- delivered           timestamp without time zone,  --record when the transaction arrived
    channel             notification_channels not null,
    content             jsonb not null
);

-- deprecated
-- drop table if exists users_notifications;
-- create table users_notifications
-- (
--     id              serial                      not null,
--     user_id         int                         not null,
--     event_name      character varying(100)      not null,
--     event_filter    text                        not null default '',
--     sent_ts         timestamp without time zone,
--     epoch           int                         not null,
--     primary key(user_id, event_name, event_filter, sent_ts)
-- );

drop table if exists users_validators_tags;
create table users_validators_tags
(
    user_id             int                    not null,
    validator_publickey bytea                  not null,
    tag                 character varying(100) not null,
    primary key (user_id, validator_publickey, tag)
);

drop table if exists validator_tags;
create table validator_tags
(
    publickey bytea                  not null,
    tag       character varying(100) not null,
    primary key (publickey, tag)
);

drop table if exists users_webhooks;
create table users_webhooks
(   
    id                serial                  not null,
    user_id           int                     not null,
    -- label             varchar(200)            not null,
    url               character varying(1024) not null,
    retries           int                     not null default 0, -- a backoff parameter that indicates if the requests was successful and when to retry it again
    request           jsonb,
    response          jsonb,
    last_sent         timestamp without time zone,
    event_names       text[]                  not null,
    destination       character varying(200), -- discord for example could be a destination and the request would be adapted
    primary key (user_id, id)
);

drop table if exists mails_sent;
create table mails_sent
(
    email character varying(100)      not null,
    ts    timestamp without time zone not null,
    cnt   int                         not null,
    primary key (email, ts)
);

drop table if exists chart_images;
create table chart_images
(
    name  varchar(100) not null primary key,
    image bytea        not null
);

drop table if exists api_statistics;
create table api_statistics
(
    ts     timestamp without time zone not null,
    apikey varchar(64)                 not null,
    call   varchar(64)                 not null,
    count  int                         not null default 0,
    primary key (ts, apikey, call)
);

drop table if exists stake_pools_stats;
create table stake_pools_stats
(
    id serial not null,
    address text not null,
    deposit int,
    name text not null,
    category text,
    PRIMARY KEY(id, address, deposit, name)
);

drop table if exists price;
create table price
(
    ts     timestamp without time zone not null,
    eur numeric(20,10)                not null,
    usd numeric(20,10)                not null,
    rub numeric(20,10)                not null,
    cny numeric(20,10)                not null,
    cad numeric(20,10)                not null,
    jpy numeric(20,10)                not null,
    gbp numeric(20,10)                not null,
    aud numeric(20,10)                not null,
    primary key (ts)
);

drop table if exists staking_pools_chart;
create table staking_pools_chart
(
    epoch                      int  not null,
    name                       text not null,
    income                     bigint not null,
    balance                    bigint not null,
    PRIMARY KEY(epoch, name)
);

drop table if exists stats_sharing;
CREATE TABLE stats_sharing (
                               id 				bigserial 			primary key,
                               ts 				timestamp  			not null,
                               share           bool             not null,
                               user_id 		 	bigint	 	 		not null,
                               foreign key(user_id) references users(id)
);

drop table if exists finality_checkpoints;
create table finality_checkpoints (
                                      head_epoch               int   not null,
                                      head_root                bytea not null,
                                      current_justified_epoch  int   not null,
                                      current_justified_root   bytea not null,
                                      previous_justified_epoch int   not null,
                                      previous_justified_root  bytea not null,
                                      finalized_epoch          int   not null,
                                      finalized_root           bytea not null,
                                      primary key (head_epoch, head_root)
);

drop table if exists rocketpool_export_status;
create table rocketpool_export_status
(
    rocketpool_storage_address bytea not null,
    eth1_block int not null,
    primary key (rocketpool_storage_address)
);

drop table if exists rocketpool_minipools;
create table rocketpool_minipools
(
    rocketpool_storage_address bytea not null,

    address bytea not null,
    pubkey bytea not null,
    node_address bytea not null,
    node_fee float not null,
    deposit_type varchar(20) not null, -- none (invalid), full, half, empty .. see: https://github.com/rocket-pool/rocketpool/blob/683addf4ac/contracts/types/MinipoolDeposit.sol
    status text not null, -- Initialized, Prelaunch, Staking, Withdrawable, Dissolved .. see: https://github.com/rocket-pool/rocketpool/blob/683addf4ac/contracts/types/MinipoolStatus.sol
    status_time timestamp without time zone,
    penalty_count numeric not null default 0,
    primary key(rocketpool_storage_address, address)
);

drop table if exists rocketpool_nodes;
create table rocketpool_nodes
(
    rocketpool_storage_address bytea not null,

    address bytea not null,
    timezone_location varchar(200) not null,
    rpl_stake numeric not null,
    min_rpl_stake numeric not null,
    max_rpl_stake numeric not null,
    rpl_cumulative_rewards numeric not null default 0,
    smoothing_pool_opted_in boolean not null default false,
    claimed_smoothing_pool  numeric not null,
    unclaimed_smoothing_pool  numeric not null,
    unclaimed_rpl_rewards numeric not null,

    primary key(rocketpool_storage_address, address)
);

drop table if exists rocketpool_dao_proposals;
create table rocketpool_dao_proposals
(
    rocketpool_storage_address bytea not null,

    id int not null,
    dao text not null,
    proposer_address bytea not null,
    message text not null,
    created_time timestamp without time zone,
    start_time timestamp without time zone,
    end_time timestamp without time zone,
    expiry_time timestamp without time zone,
    votes_required float not null,
    votes_for float not null,
    votes_against float not null,
    member_voted boolean not null,
    member_supported boolean not null,
    is_cancelled boolean not null,
    is_executed boolean not null,
    payload bytea not null,
    state text not null,

    primary key(rocketpool_storage_address, id)
);

drop table if exists rocketpool_dao_proposals_member_votes;
create table rocketpool_dao_proposals_member_votes
(
    rocketpool_storage_address bytea not null,

    id int not null,
    member_address bytea not null,
    voted       boolean not null,
    supported   boolean not null,

    primary key(rocketpool_storage_address, id, member_address)
);

drop table if exists rocketpool_dao_members;
create table rocketpool_dao_members
(
    rocketpool_storage_address bytea not null,

    address bytea not null,
    id varchar(200) not null,
    url varchar(200) not null,
    joined_time timestamp without time zone,
    last_proposal_time timestamp without time zone,
    rpl_bond_amount numeric not null,
    unbonded_validator_count int not null,

    primary key(rocketpool_storage_address, address)
);

drop table if exists rocketpool_network_stats;
create table rocketpool_network_stats
(
    id 				    bigserial,
    ts timestamp without time zone not null,
    rpl_price  numeric not null,
    claim_interval_time interval not null,
    claim_interval_time_start timestamp without time zone not null,
    current_node_fee float not null,
    current_node_demand numeric not null,
    reth_supply numeric not null,
    effective_rpl_staked numeric not null,
    node_operator_rewards numeric not null,
    reth_exchange_rate float not null,
    node_count numeric not null,
    minipool_count numeric not null,
    odao_member_count numeric not null,
    total_eth_staking numeric not null,
    total_eth_balance numeric not null,

    primary key(id)
);

drop table if exists rocketpool_reward_tree;
create table rocketpool_reward_tree
(
    id 				    bigserial,
    data                jsonb not null,

    primary key(id)
);

drop table if exists eth_store_stats;
create table eth_store_stats
(
    day			                int not null,
    validator			        int not null,
    effective_balances_sum_wei numeric not null,
    start_balances_sum_wei numeric not null,
    end_balances_sum_wei numeric not null,
    deposits_sum_wei numeric not null,
    tx_fees_sum_wei numeric not null,
    consensus_rewards_sum_wei numeric not null,
    total_rewards_wei numeric not null,
    apr float   not null,
    
    primary key(day, validator)
);
create index idx_eth_store_validator on eth_store_stats (validator, day desc);

drop table if exists historical_pool_performance;
create table historical_pool_performance
(
    day                int not null,
    pool        varchar(40) not null,
    validators int not null,
    effective_balances_sum_wei numeric not null,
    start_balances_sum_wei numeric not null,
    end_balances_sum_wei numeric not null,
    deposits_sum_wei numeric not null,
    tx_fees_sum_wei numeric not null,
    consensus_rewards_sum_wei numeric not null,
    total_rewards_wei numeric not null,
    apr float   not null,
    
    primary key(day, pool)
);
create index idx_historical_pool_performance_pool on historical_pool_performance (pool, day desc);

--- need to drop all three tabls in the correct order to correctly resolve foreign key constrains
DROP TABLE IF EXISTS relays;
DROP TABLE IF EXISTS blocks_tags;
DROP TABLE IF EXISTS tags;

CREATE TABLE tags (
	id varchar NOT NULL,
	metadata jsonb NOT NULL,
	PRIMARY KEY (id)
);

CREATE TABLE blocks_tags (
	slot int4 NOT NULL,
	blockroot bytea NOT NULL,
	tag_id varchar NOT NULL,
	PRIMARY KEY (slot, blockroot, tag_id),
	FOREIGN KEY (slot, blockroot) REFERENCES blocks(slot, blockroot),
	FOREIGN KEY (tag_id) REFERENCES tags(id)
);
CREATE INDEX idx_blocks_tags_slot ON blocks_tags (slot);
CREATE INDEX idx_blocks_tags_tag_id ON blocks_tags (tag_id);

CREATE TABLE relays (
	tag_id varchar NOT NULL,
	endpoint varchar NOT NULL,
	PRIMARY KEY (tag_id, endpoint),
	FOREIGN KEY (tag_id) REFERENCES tags(id)
);

DROP TABLE IF EXISTS relays;

CREATE TABLE relays (
	tag_id varchar NOT NULL,
	endpoint varchar NOT NULL,
	public_link varchar NULL,
	is_censoring bool NULL,
	is_ethical bool NULL,
	PRIMARY KEY (tag_id, endpoint),
	FOREIGN KEY (tag_id) REFERENCES tags(id)
);


DROP TABLE IF EXISTS relays_blocks;

CREATE TABLE relays_blocks (
	tag_id varchar NOT NULL,
	block_slot int4 NOT NULL,
	block_root bytea NOT NULL,
	exec_block_hash bytea NOT NULL,
	builder_pubkey bytea NOT NULL,
	proposer_pubkey bytea NOT NULL,
	proposer_fee_recipient bytea NOT NULL,
	value numeric NOT NULL,
	PRIMARY KEY (block_slot, block_root, tag_id)
);
CREATE INDEX relays_blocks_block_root_idx ON public.relays_blocks (block_root);
CREATE INDEX relays_blocks_builder_pubkey_idx ON public.relays_blocks (builder_pubkey);
CREATE INDEX relays_blocks_exec_block_hash_idx ON public.relays_blocks (exec_block_hash);
CREATE INDEX relays_blocks_value_idx ON public.relays_blocks (value);


DROP TABLE IF EXISTS validator_queue_deposits;
CREATE TABLE validator_queue_deposits (
	validatorindex int4 NOT NULL,
	block_slot int4 NULL,
	block_index int4 NULL,
	CONSTRAINT validator_queue_deposits_fk FOREIGN KEY (block_slot,block_index) REFERENCES blocks_deposits(block_slot,block_index),
	CONSTRAINT validator_queue_deposits_fk_validators FOREIGN KEY (validatorindex) REFERENCES validators(validatorindex)
);
CREATE INDEX idx_validator_queue_deposits_block_slot ON validator_queue_deposits USING btree (block_slot);
CREATE UNIQUE INDEX idx_validator_queue_deposits_validatorindex ON validator_queue_deposits USING btree (validatorindex);

drop table if exists service_status;
create table service_status (name text not null, executable_name text not null, version text not null, pid int not null, status text not null, metadata jsonb, last_update timestamp not null, primary key (name, executable_name, version, pid));

DROP TABLE IF EXISTS chart_series;
CREATE TABLE chart_series (
    "time" timestamp without time zone NOT NULL,
    indicator character varying(50) NOT NULL,
    value numeric NOT NULL,
    primary key ("time", indicator)
);

drop table if exists chart_series_status;
create table chart_series_status
(
    day    int     not null,
    status boolean not null,
    primary key (day)
);


drop table if exists global_notifications;
create table global_notifications
(
    target varchar(20) not null primary key, 
    content text not null,
    enabled bool not null
);
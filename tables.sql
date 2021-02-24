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
create index idx_validators_status on validators (status);
create index idx_validators_balanceactivation on validators (balanceactivation);
create index idx_validators_activationepoch on validators (activationepoch);

drop table if exists validator_names;
create table validator_names
(
    publickey bytea not null,
    name      varchar(40),
    primary key (publickey)
);
create index idx_validator_names_publickey on validator_names (publickey);

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

drop table if exists attestation_assignments_p;
create table attestation_assignments_p
(
    epoch          int not null,
    validatorindex int not null,
    attesterslot   int not null,
    committeeindex int not null,
    status         int not null, /* Can be 0 = scheduled, 1 executed, 2 missed */
    inclusionslot  int not null default 0, /* Slot this attestation was included for the first time */
    week           int not null,
    primary key (validatorindex, week, epoch)
) PARTITION BY LIST (week);

CREATE TABLE attestation_assignments_0 PARTITION OF attestation_assignments_p FOR VALUES IN (0);
CREATE TABLE attestation_assignments_1 PARTITION OF attestation_assignments_p FOR VALUES IN (1);
CREATE TABLE attestation_assignments_2 PARTITION OF attestation_assignments_p FOR VALUES IN (2);
CREATE TABLE attestation_assignments_3 PARTITION OF attestation_assignments_p FOR VALUES IN (3);
CREATE TABLE attestation_assignments_4 PARTITION OF attestation_assignments_p FOR VALUES IN (4);
CREATE TABLE attestation_assignments_5 PARTITION OF attestation_assignments_p FOR VALUES IN (5);
CREATE TABLE attestation_assignments_6 PARTITION OF attestation_assignments_p FOR VALUES IN (6);
CREATE TABLE attestation_assignments_7 PARTITION OF attestation_assignments_p FOR VALUES IN (7);
CREATE TABLE attestation_assignments_8 PARTITION OF attestation_assignments_p FOR VALUES IN (8);
CREATE TABLE attestation_assignments_9 PARTITION OF attestation_assignments_p FOR VALUES IN (9);

drop table if exists validator_balances_p;
create table validator_balances_p
(
    epoch            int    not null,
    validatorindex   int    not null,
    balance          bigint not null,
    effectivebalance bigint not null,
    week             int    not null,
    primary key (validatorindex, week, epoch)
) PARTITION BY LIST (week);

CREATE TABLE validator_balances_0 PARTITION OF validator_balances_p FOR VALUES IN (0);
CREATE TABLE validator_balances_1 PARTITION OF validator_balances_p FOR VALUES IN (1);
CREATE TABLE validator_balances_2 PARTITION OF validator_balances_p FOR VALUES IN (2);
CREATE TABLE validator_balances_3 PARTITION OF validator_balances_p FOR VALUES IN (3);
CREATE TABLE validator_balances_4 PARTITION OF validator_balances_p FOR VALUES IN (4);
CREATE TABLE validator_balances_5 PARTITION OF validator_balances_p FOR VALUES IN (5);
CREATE TABLE validator_balances_6 PARTITION OF validator_balances_p FOR VALUES IN (6);
CREATE TABLE validator_balances_7 PARTITION OF validator_balances_p FOR VALUES IN (7);
CREATE TABLE validator_balances_8 PARTITION OF validator_balances_p FOR VALUES IN (8);
CREATE TABLE validator_balances_9 PARTITION OF validator_balances_p FOR VALUES IN (9);

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
    proposed_blocks         int,
    missed_blocks           int,
    orphaned_blocks         int,
    attester_slashings      int,
    proposer_slashings      int,
    deposits                int,
    deposits_amount         bigint,
    primary key (validatorindex, day)
);

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
    validatorindex int not null,
    status         int not null,
    start          int not null,
    length         int not null,
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

drop table if exists epochs;
create table epochs
(
    epoch                   int    not null,
    blockscount             int    not null default 0,
    proposerslashingscount  int    not null,
    attesterslashingscount  int    not null,
    attestationscount       int    not null,
    depositscount           int    not null,
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
    epoch                  int   not null,
    slot                   int   not null,
    blockroot              bytea not null,
    parentroot             bytea not null,
    stateroot              bytea not null,
    signature              bytea not null,
    randaoreveal           bytea,
    graffiti               bytea,
    graffiti_text          text  null,
    eth1data_depositroot   bytea,
    eth1data_depositcount  int   not null,
    eth1data_blockhash     bytea,
    proposerslashingscount int   not null,
    attesterslashingscount int   not null,
    attestationscount      int   not null,
    depositscount          int   not null,
    voluntaryexitscount    int   not null,
    proposer               int   not null,
    status                 text  not null, /* Can be 0 = scheduled, 1 proposed, 2 missed, 3 orphaned */
    primary key (slot, blockroot)
);
create index idx_blocks_proposer on blocks (proposer);

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
    primary key (id, email)
);

drop table if exists users_stripe_subscriptions;
create table users_stripe_subscriptions
(
    subscription_id character varying(256) unique not null,
    customer_id     character varying(256)        not null,
    price_id        character varying(256)        not null,
    active          bool not null default 'f',
    payload         json                          not null,
    primary key (customer_id, subscription_id, price_id)
);

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
    consumed   bool                        not null default 'f',
    app_id     int                         not null,
    created_ts timestamp without time zone not null,
    primary key (user_id, code)
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
    id              serial                      not null,
    user_id         int                         not null,
    event_name      character varying(100)      not null,
    event_filter    text                        not null default '',
    last_sent_ts    timestamp without time zone,
    last_sent_epoch int,
    created_ts      timestamp without time zone not null,
    created_epoch   int                         not null,
    primary key (user_id, event_name, event_filter)
);

drop table if exists users_notifications;
create table users_notifications
(
    id              serial                      not null,
    user_id         int                         not null,
    event_name      character varying(100)      not null,
    event_filter    text                        not null default '',
    sent_ts         timestamp without time zone,
    epoch           int                         not null,
    primary key(user_id, event_name, event_filter, sent_ts)
);

drop table if exists users_validators_tags;
create table users_validators_tags
(
    user_id             int                    not null,
    validator_publickey bytea                  not null,
    tag                 character varying(100) not null,
    primary key (user_id, validator_publickey, tag)
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
    primary key (ts)
);

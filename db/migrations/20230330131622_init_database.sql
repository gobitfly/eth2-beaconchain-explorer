-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

/* trigram extension for faster text-search */
/*
This TABLE IS used to store the current state (latest exported epoch) of ALL validators
It also acts AS a lookup-table to store the index-pubkey association
IN ORDER to save db space we only use the UNIQUE validator INDEX IN ALL other tables
IN the future it IS better to REPLACE this TABLE WITH an IN memory cache (redis)
 */
CREATE TABLE IF NOT EXISTS
    validators (
        validatorindex INT NOT NULL,
        pubkey bytea NOT NULL,
        pubkeyhex TEXT NOT NULL DEFAULT '',
        withdrawableepoch BIGINT NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        balance BIGINT NOT NULL,
        balanceactivation BIGINT,
        effectivebalance BIGINT NOT NULL,
        slashed bool NOT NULL,
        activationeligibilityepoch BIGINT NOT NULL,
        activationepoch BIGINT NOT NULL,
        exitepoch BIGINT NOT NULL,
        lastattestationslot BIGINT,
        status VARCHAR(20) NOT NULL DEFAULT '',
        PRIMARY KEY (validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validators_pubkey ON validators (pubkey);

CREATE INDEX IF NOT EXISTS idx_validators_pubkeyhex ON validators (pubkeyhex);

CREATE INDEX IF NOT EXISTS idx_validators_pubkeyhex_pattern_pos ON validators (pubkeyhex varchar_pattern_ops);

CREATE INDEX IF NOT EXISTS idx_validators_status ON validators (status);

CREATE INDEX IF NOT EXISTS idx_validators_balanceactivation ON validators (balanceactivation);

CREATE INDEX IF NOT EXISTS idx_validators_activationepoch ON validators (activationepoch);

CREATE INDEX IF NOT EXISTS validators_is_offline_vali_idx ON validators (validatorindex, lastattestationslot, pubkey);

CREATE INDEX IF NOT EXISTS idx_validators_withdrawalcredentials ON validators (withdrawalcredentials, validatorindex);

CREATE TABLE IF NOT EXISTS
    validator_pool (
        publickey bytea NOT NULL,
        pool VARCHAR(40),
        PRIMARY KEY (publickey)
    );

CREATE TABLE IF NOT EXISTS
    validator_names (
        publickey bytea NOT NULL,
        NAME VARCHAR(40),
        PRIMARY KEY (publickey)
    );

CREATE INDEX IF NOT EXISTS idx_validator_names_publickey ON validator_names (publickey);

CREATE INDEX IF NOT EXISTS idx_validator_names_name ON validator_names (NAME);

CREATE TABLE IF NOT EXISTS
    validator_set (
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        withdrawableepoch BIGINT NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        effectivebalance BIGINT NOT NULL,
        slashed bool NOT NULL,
        activationeligibilityepoch BIGINT NOT NULL,
        activationepoch BIGINT NOT NULL,
        exitepoch BIGINT NOT NULL,
        PRIMARY KEY (validatorindex, epoch)
    );

CREATE TABLE IF NOT EXISTS
    validator_performance (
        validatorindex INT NOT NULL,
        balance BIGINT NOT NULL,
        cl_performance_1d BIGINT NOT NULL,
        cl_performance_7d BIGINT NOT NULL,
        cl_performance_31d BIGINT NOT NULL,
        cl_performance_365d BIGINT NOT NULL,
        cl_performance_total BIGINT NOT NULL,
        el_performance_1d BIGINT NOT NULL,
        el_performance_7d BIGINT NOT NULL,
        el_performance_31d BIGINT NOT NULL,
        el_performance_365d BIGINT NOT NULL,
        el_performance_total BIGINT NOT NULL,
        mev_performance_1d BIGINT NOT NULL,
        mev_performance_7d BIGINT NOT NULL,
        mev_performance_31d BIGINT NOT NULL,
        mev_performance_365d BIGINT NOT NULL,
        mev_performance_total BIGINT NOT NULL,
        rank7d INT NOT NULL,
        PRIMARY KEY (validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validator_performance_balance ON validator_performance (balance);

CREATE INDEX IF NOT EXISTS idx_validator_performance_rank7d ON validator_performance (rank7d);

CREATE TABLE IF NOT EXISTS
    proposal_assignments (
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        proposerslot INT NOT NULL,
        status INT NOT NULL,
        /* Can be 0 = scheduled, 1 executed, 2 missed */
        PRIMARY KEY (epoch, validatorindex, proposerslot)
    );

CREATE INDEX IF NOT EXISTS idx_proposal_assignments_epoch ON proposal_assignments (epoch);

CREATE TABLE IF NOT EXISTS
    sync_committees (
        period INT NOT NULL,
        validatorindex INT NOT NULL,
        committeeindex INT NOT NULL,
        PRIMARY KEY (period, validatorindex, committeeindex)
    );

CREATE TABLE IF NOT EXISTS
    sync_committees_count_per_validator (
        period INT NOT NULL UNIQUE,
        count_so_far float8 NOT NULL,
        PRIMARY KEY (period)
    );

CREATE TABLE IF NOT EXISTS
    validator_balances_recent (
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        balance BIGINT NOT NULL,
        PRIMARY KEY (epoch, validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validator_balances_recent_epoch ON validator_balances_recent (epoch);

CREATE INDEX IF NOT EXISTS idx_validator_balances_recent_validatorindex ON validator_balances_recent (validatorindex);

CREATE INDEX IF NOT EXISTS idx_validator_balances_recent_balance ON validator_balances_recent (balance);

CREATE TABLE IF NOT EXISTS
    validator_stats (
        validatorindex INT NOT NULL,
        DAY INT NOT NULL,
        start_balance BIGINT,
        end_balance BIGINT,
        min_balance BIGINT,
        max_balance BIGINT,
        start_effective_balance BIGINT,
        end_effective_balance BIGINT,
        min_effective_balance BIGINT,
        max_effective_balance BIGINT,
        missed_attestations INT,
        orphaned_attestations INT,
        participated_sync INT,
        missed_sync INT,
        orphaned_sync INT,
        proposed_blocks INT,
        missed_blocks INT,
        orphaned_blocks INT,
        attester_slashings INT,
        proposer_slashings INT,
        deposits INT,
        deposits_amount BIGINT,
        withdrawals INT,
        withdrawals_amount BIGINT,
        cl_rewards_gwei BIGINT,
        cl_rewards_gwei_total BIGINT,
        el_rewards_wei DECIMAL,
        el_rewards_wei_total DECIMAL,
        mev_rewards_wei DECIMAL,
        mev_rewards_wei_total DECIMAL,
        PRIMARY KEY (validatorindex, DAY)
    );

CREATE INDEX IF NOT EXISTS idx_validator_stats_day ON validator_stats (DAY);

CREATE TABLE IF NOT EXISTS
    validator_stats_status (
        DAY INT NOT NULL,
        status BOOLEAN NOT NULL,
        income_exported BOOLEAN NOT NULL DEFAULT FALSE,
        PRIMARY KEY (DAY)
    );

CREATE TABLE IF NOT EXISTS
    validator_attestation_streaks (
        validatorindex INT NOT NULL,
        status INT NOT NULL,
        START INT NOT NULL,
        LENGTH INT NOT NULL,
        longest BOOLEAN NOT NULL,
        CURRENT BOOLEAN NOT NULL,
        PRIMARY KEY (validatorindex, status, START)
    );

CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_validatorindex ON validator_attestation_streaks (validatorindex);

CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_status ON validator_attestation_streaks (status);

CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_length ON validator_attestation_streaks (LENGTH);

CREATE INDEX IF NOT EXISTS idx_validator_attestation_streaks_start ON validator_attestation_streaks (START);

CREATE TABLE IF NOT EXISTS
    queue (
        ts TIMESTAMP WITHOUT TIME ZONE,
        entering_validators_count INT NOT NULL,
        exiting_validators_count INT NOT NULL,
        PRIMARY KEY (ts)
    );

CREATE TABLE IF NOT EXISTS
    validatorqueue_activation (
        INDEX INT NOT NULL,
        publickey bytea NOT NULL,
        PRIMARY KEY (INDEX, publickey)
    );

CREATE TABLE IF NOT EXISTS
    validatorqueue_exit (
        INDEX INT NOT NULL,
        publickey bytea NOT NULL,
        PRIMARY KEY (INDEX, publickey)
    );

CREATE TABLE IF NOT EXISTS
    epochs_notified (
        epoch INT NOT NULL PRIMARY KEY,
        sentOn TIMESTAMP NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    epochs (
        epoch INT NOT NULL,
        blockscount INT NOT NULL DEFAULT 0,
        proposerslashingscount INT NOT NULL,
        attesterslashingscount INT NOT NULL,
        attestationscount INT NOT NULL,
        depositscount INT NOT NULL,
        withdrawalcount INT NOT NULL DEFAULT 0,
        voluntaryexitscount INT NOT NULL,
        validatorscount INT NOT NULL,
        averagevalidatorbalance BIGINT NOT NULL,
        totalvalidatorbalance BIGINT NOT NULL,
        finalized bool,
        eligibleether BIGINT,
        globalparticipationrate FLOAT,
        votedether BIGINT,
        rewards_exported bool NOT NULL DEFAULT FALSE,
        PRIMARY KEY (epoch)
    );

CREATE TABLE IF NOT EXISTS
    blocks (
        epoch INT NOT NULL,
        slot INT NOT NULL,
        blockroot bytea NOT NULL,
        parentroot bytea NOT NULL,
        stateroot bytea NOT NULL,
        signature bytea NOT NULL,
        randaoreveal bytea,
        graffiti bytea,
        graffiti_text TEXT NULL,
        eth1data_depositroot bytea,
        eth1data_depositcount INT NOT NULL,
        eth1data_blockhash bytea,
        syncaggregate_bits bytea,
        syncaggregate_signature bytea,
        syncaggregate_participation FLOAT NOT NULL DEFAULT 0,
        proposerslashingscount INT NOT NULL,
        attesterslashingscount INT NOT NULL,
        attestationscount INT NOT NULL,
        depositscount INT NOT NULL,
        withdrawalcount INT NOT NULL DEFAULT 0,
        voluntaryexitscount INT NOT NULL,
        proposer INT NOT NULL,
        status TEXT NOT NULL,
        /* Can be 0 = scheduled, 1 proposed, 2 missed, 3 orphaned */
        -- https://ethereum.github.io/beacon-APIs/#/Beacon/getBlockV2
        -- https://github.com/ethereum/consensus-specs/blob/v1.1.9/specs/bellatrix/beacon-chain.md#executionpayload
        exec_parent_hash bytea,
        exec_fee_recipient bytea,
        exec_state_root bytea,
        exec_receipts_root bytea,
        exec_logs_bloom bytea,
        exec_random bytea,
        exec_block_number INT,
        exec_gas_limit INT,
        exec_gas_used INT,
        exec_timestamp INT,
        exec_extra_data bytea,
        exec_base_fee_per_gas BIGINT,
        exec_block_hash bytea,
        exec_transactions_count INT NOT NULL DEFAULT 0,
        PRIMARY KEY (slot, blockroot)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_proposer ON blocks (proposer);

CREATE INDEX IF NOT EXISTS idx_blocks_epoch ON blocks (epoch);

CREATE INDEX IF NOT EXISTS idx_blocks_graffiti_text ON blocks USING gin (graffiti_text gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_blocks_blockrootstatus ON blocks (blockroot, status);

CREATE INDEX IF NOT EXISTS idx_blocks_exec_block_number ON blocks (exec_block_number);

CREATE TABLE IF NOT EXISTS
    blocks_withdrawals (
        block_slot INT NOT NULL,
        block_root bytea NOT NULL,
        withdrawalindex INT NOT NULL,
        validatorindex INT NOT NULL,
        address bytea NOT NULL,
        amount BIGINT NOT NULL,
        -- in GWei
        PRIMARY KEY (block_slot, block_root, withdrawalindex)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_recipient ON blocks_withdrawals (address);

CREATE INDEX IF NOT EXISTS idx_blocks_withdrawals_validatorindex ON blocks_withdrawals (validatorindex);

CREATE TABLE IF NOT EXISTS
    blocks_bls_change (
        block_slot INT NOT NULL,
        block_root bytea NOT NULL,
        validatorindex INT NOT NULL,
        signature bytea NOT NULL,
        pubkey bytea NOT NULL,
        address bytea NOT NULL,
        PRIMARY KEY (block_slot, block_root, validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_pubkey ON blocks_bls_change (pubkey);

CREATE INDEX IF NOT EXISTS idx_blocks_bls_change_address ON blocks_bls_change (address);

CREATE TABLE IF NOT EXISTS
    blocks_transactions (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        raw bytea NOT NULL,
        txhash bytea NOT NULL,
        nonce INT NOT NULL,
        gas_price bytea NOT NULL,
        gas_limit BIGINT NOT NULL,
        sender bytea NOT NULL,
        recipient bytea NOT NULL,
        amount bytea NOT NULL,
        payload bytea NOT NULL,
        max_priority_fee_per_gas BIGINT,
        max_fee_per_gas BIGINT,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    blocks_proposerslashings (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        proposerindex INT NOT NULL,
        header1_slot BIGINT NOT NULL,
        header1_parentroot bytea NOT NULL,
        header1_stateroot bytea NOT NULL,
        header1_bodyroot bytea NOT NULL,
        header1_signature bytea NOT NULL,
        header2_slot BIGINT NOT NULL,
        header2_parentroot bytea NOT NULL,
        header2_stateroot bytea NOT NULL,
        header2_bodyroot bytea NOT NULL,
        header2_signature bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    blocks_attesterslashings (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        attestation1_indices INTEGER[] NOT NULL,
        attestation1_signature bytea NOT NULL,
        attestation1_slot BIGINT NOT NULL,
        attestation1_index INT NOT NULL,
        attestation1_beaconblockroot bytea NOT NULL,
        attestation1_source_epoch INT NOT NULL,
        attestation1_source_root bytea NOT NULL,
        attestation1_target_epoch INT NOT NULL,
        attestation1_target_root bytea NOT NULL,
        attestation2_indices INTEGER[] NOT NULL,
        attestation2_signature bytea NOT NULL,
        attestation2_slot BIGINT NOT NULL,
        attestation2_index INT NOT NULL,
        attestation2_beaconblockroot bytea NOT NULL,
        attestation2_source_epoch INT NOT NULL,
        attestation2_source_root bytea NOT NULL,
        attestation2_target_epoch INT NOT NULL,
        attestation2_target_root bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    blocks_attestations (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        aggregationbits bytea NOT NULL,
        validators INT[] NOT NULL,
        signature bytea NOT NULL,
        slot INT NOT NULL,
        committeeindex INT NOT NULL,
        beaconblockroot bytea NOT NULL,
        source_epoch INT NOT NULL,
        source_root bytea NOT NULL,
        target_epoch INT NOT NULL,
        target_root bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_attestations_beaconblockroot ON blocks_attestations (beaconblockroot);

CREATE INDEX IF NOT EXISTS idx_blocks_attestations_source_root ON blocks_attestations (source_root);

CREATE INDEX IF NOT EXISTS idx_blocks_attestations_target_root ON blocks_attestations (target_root);

CREATE TABLE IF NOT EXISTS
    blocks_deposits (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        proof bytea[],
        publickey bytea NOT NULL,
        withdrawalcredentials bytea NOT NULL,
        amount BIGINT NOT NULL,
        signature bytea NOT NULL,
        valid_signature bool NOT NULL DEFAULT TRUE,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_deposits_publickey ON blocks_deposits (publickey);

CREATE TABLE IF NOT EXISTS
    blocks_voluntaryexits (
        block_slot INT NOT NULL,
        block_index INT NOT NULL,
        block_root bytea NOT NULL DEFAULT '',
        epoch INT NOT NULL,
        validatorindex INT NOT NULL,
        signature bytea NOT NULL,
        PRIMARY KEY (block_slot, block_index)
    );

CREATE TABLE IF NOT EXISTS
    network_liveness (
        ts TIMESTAMP WITHOUT TIME ZONE,
        headepoch INT NOT NULL,
        finalizedepoch INT NOT NULL,
        justifiedepoch INT NOT NULL,
        previousjustifiedepoch INT NOT NULL,
        PRIMARY KEY (ts)
    );

CREATE TABLE IF NOT EXISTS
    graffitiwall (
        x INT NOT NULL,
        y INT NOT NULL,
        color TEXT NOT NULL,
        slot INT NOT NULL,
        VALIDATOR INT NOT NULL,
        PRIMARY KEY (x, y)
    );

CREATE TABLE IF NOT EXISTS
    eth1_deposits (
        tx_hash bytea NOT NULL,
        tx_input bytea NOT NULL,
        tx_index INT NOT NULL,
        block_number INT NOT NULL,
        block_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        from_address bytea NOT NULL,
        publickey bytea NOT NULL,
        withdrawal_credentials bytea NOT NULL,
        amount BIGINT NOT NULL,
        signature bytea NOT NULL,
        merkletree_index bytea NOT NULL,
        removed bool NOT NULL,
        valid_signature bool NOT NULL,
        PRIMARY KEY (tx_hash, merkletree_index)
    );

CREATE INDEX IF NOT EXISTS idx_eth1_deposits ON eth1_deposits (publickey);

CREATE INDEX IF NOT EXISTS idx_eth1_deposits_from_address ON eth1_deposits (from_address);

CREATE TABLE IF NOT EXISTS
    eth1_deposits_aggregated (
        from_address bytea NOT NULL,
        amount BIGINT NOT NULL,
        validcount INT NOT NULL,
        invalidcount INT NOT NULL,
        slashedcount INT NOT NULL,
        totalcount INT NOT NULL,
        activecount INT NOT NULL,
        pendingcount INT NOT NULL,
        voluntary_exit_count INT NOT NULL,
        PRIMARY KEY (from_address)
    );

CREATE TABLE IF NOT EXISTS
    users (
        id serial NOT NULL UNIQUE,
        PASSWORD CHARACTER VARYING(256) NOT NULL,
        email CHARACTER VARYING(100) NOT NULL UNIQUE,
        email_confirmed bool NOT NULL DEFAULT 'f',
        email_confirmation_hash CHARACTER VARYING(40) UNIQUE,
        email_confirmation_ts TIMESTAMP WITHOUT TIME ZONE,
        email_change_to_value CHARACTER VARYING(100),
        password_reset_hash CHARACTER VARYING(40),
        password_reset_ts TIMESTAMP WITHOUT TIME ZONE,
        register_ts TIMESTAMP WITHOUT TIME ZONE,
        api_key CHARACTER VARYING(256) UNIQUE,
        stripe_customer_id CHARACTER VARYING(256) UNIQUE,
        user_group VARCHAR(10),
        PRIMARY KEY (id, email)
    );

CREATE TABLE IF NOT EXISTS
    users_datatable (
        user_id INT NOT NULL,
        KEY CHARACTER VARYING(256) NOT NULL,
        state jsonb NOT NULL,
        updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT 'now()',
        PRIMARY KEY (user_id, KEY)
    );

CREATE TABLE IF NOT EXISTS
    users_stripe_subscriptions (
        subscription_id CHARACTER VARYING(256) UNIQUE NOT NULL,
        customer_id CHARACTER VARYING(256) NOT NULL,
        price_id CHARACTER VARYING(256) NOT NULL,
        active bool NOT NULL DEFAULT 'f',
        payload json NOT NULL,
        purchase_group CHARACTER VARYING(30) NOT NULL DEFAULT 'api',
        PRIMARY KEY (customer_id, subscription_id, price_id)
    );

CREATE TABLE IF NOT EXISTS
    users_app_subscriptions (
        id serial NOT NULL,
        user_id INT NOT NULL,
        product_id CHARACTER VARYING(256) NOT NULL,
        price_micros BIGINT NOT NULL,
        currency CHARACTER VARYING(10) NOT NULL,
        created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        validate_remotely BOOLEAN NOT NULL DEFAULT 't',
        active bool NOT NULL DEFAULT 'f',
        store CHARACTER VARYING(50) NOT NULL,
        expires_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        reject_reason CHARACTER VARYING(50),
        receipt CHARACTER VARYING(99999) NOT NULL,
        receipt_hash CHARACTER VARYING(1024) NOT NULL UNIQUE,
        subscription_id CHARACTER VARYING(256) DEFAULT ''
    );

CREATE INDEX IF NOT EXISTS idx_user_app_subscriptions ON users_app_subscriptions (user_id);

CREATE TABLE IF NOT EXISTS
    oauth_apps (
        id serial NOT NULL,
        owner_id INT NOT NULL,
        redirect_uri CHARACTER VARYING(100) NOT NULL UNIQUE,
        app_name CHARACTER VARYING(35) NOT NULL,
        active bool NOT NULL DEFAULT 't',
        created_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        PRIMARY KEY (id, redirect_uri)
    );

CREATE TABLE IF NOT EXISTS
    oauth_codes (
        id serial NOT NULL,
        user_id INT NOT NULL,
        code CHARACTER VARYING(64) NOT NULL,
        client_id CHARACTER VARYING(128) NOT NULL,
        consumed bool NOT NULL DEFAULT 'f',
        app_id INT NOT NULL,
        created_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        PRIMARY KEY (user_id, app_id, client_id)
    );

CREATE TABLE IF NOT EXISTS
    users_devices (
        id serial NOT NULL,
        user_id INT NOT NULL,
        refresh_token CHARACTER VARYING(64) NOT NULL,
        device_name CHARACTER VARYING(20) NOT NULL,
        notification_token CHARACTER VARYING(500),
        notify_enabled bool NOT NULL DEFAULT 't',
        active bool NOT NULL DEFAULT 't',
        app_id INT NOT NULL,
        created_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        PRIMARY KEY (user_id, refresh_token)
    );

CREATE TABLE IF NOT EXISTS
    users_clients (
        id serial NOT NULL,
        user_id INT NOT NULL,
        client CHARACTER VARYING(12) NOT NULL,
        client_version INT NOT NULL,
        notify_enabled bool NOT NULL DEFAULT 't',
        created_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        PRIMARY KEY (user_id, client)
    );

CREATE TABLE IF NOT EXISTS
    users_subscriptions (
        id serial NOT NULL,
        user_id INT NOT NULL,
        event_name CHARACTER VARYING(100) NOT NULL,
        event_filter TEXT NOT NULL DEFAULT '',
        event_threshold REAL DEFAULT 0,
        last_sent_ts TIMESTAMP WITHOUT TIME ZONE,
        last_sent_epoch INT,
        created_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        created_epoch INT NOT NULL,
        unsubscribe_hash bytea,
        internal_state VARCHAR,
        PRIMARY KEY (user_id, event_name, event_filter)
    );

CREATE INDEX IF NOT EXISTS idx_users_subscriptions_unsubscribe_hash ON users_subscriptions (unsubscribe_hash);

DO $$
BEGIN
IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'notification_channels') THEN
        CREATE TYPE notification_channels AS ENUM('webhook_discord', 'webhook', 'email', 'push');
END IF;
END$$;

CREATE TABLE IF NOT EXISTS
    users_notification_channels (
        user_id INT NOT NULL,
        channel notification_channels NOT NULL,
        active BOOLEAN DEFAULT 't' NOT NULL,
        PRIMARY KEY (user_id, channel)
    );

CREATE TABLE IF NOT EXISTS
    notification_queue (
        id serial NOT NULL,
        created TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        sent TIMESTAMP WITHOUT TIME ZONE,
        -- record when the transaction was dispatched
        -- delivered           timestamp without time zone,  --record when the transaction arrived
        channel notification_channels NOT NULL,
        CONTENT jsonb NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    users_validators_tags (
        user_id INT NOT NULL,
        validator_publickey bytea NOT NULL,
        tag CHARACTER VARYING(100) NOT NULL,
        PRIMARY KEY (user_id, validator_publickey, tag)
    );

CREATE TABLE IF NOT EXISTS
    validator_tags (
        publickey bytea NOT NULL,
        tag CHARACTER VARYING(100) NOT NULL,
        PRIMARY KEY (publickey, tag)
    );

CREATE TABLE IF NOT EXISTS
    users_webhooks (
        id serial NOT NULL,
        user_id INT NOT NULL,
        -- label             varchar(200)            not null,
        url CHARACTER VARYING(1024) NOT NULL,
        retries INT NOT NULL DEFAULT 0,
        -- a backoff parameter that indicates if the requests was successful and when to retry it again
        request jsonb,
        response jsonb,
        last_sent TIMESTAMP WITHOUT TIME ZONE,
        event_names TEXT[] NOT NULL,
        destination CHARACTER VARYING(200),
        -- discord for example could be a destination and the request would be adapted
        PRIMARY KEY (user_id, id)
    );

CREATE TABLE IF NOT EXISTS
    mails_sent (
        email CHARACTER VARYING(100) NOT NULL,
        ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        cnt INT NOT NULL,
        PRIMARY KEY (email, ts)
    );

CREATE TABLE IF NOT EXISTS
    chart_images (
        NAME VARCHAR(100) NOT NULL PRIMARY KEY,
        image bytea NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    api_statistics (
        ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        apikey VARCHAR(64) NOT NULL,
        CALL VARCHAR(64) NOT NULL,
        COUNT INT NOT NULL DEFAULT 0,
        PRIMARY KEY (
            ts,
            apikey,
            CALL
        )
    );

CREATE TABLE IF NOT EXISTS
    stake_pools_stats (
        id serial NOT NULL,
        address TEXT NOT NULL,
        deposit INT,
        NAME TEXT NOT NULL,
        category TEXT,
        PRIMARY KEY (id, address, deposit, NAME)
    );

CREATE TABLE IF NOT EXISTS
    price (
        ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        eur NUMERIC(20, 10) NOT NULL,
        usd NUMERIC(20, 10) NOT NULL,
        rub NUMERIC(20, 10) NOT NULL,
        cny NUMERIC(20, 10) NOT NULL,
        cad NUMERIC(20, 10) NOT NULL,
        jpy NUMERIC(20, 10) NOT NULL,
        gbp NUMERIC(20, 10) NOT NULL,
        aud NUMERIC(20, 10) NOT NULL,
        PRIMARY KEY (ts)
    );

CREATE TABLE IF NOT EXISTS
    stats_sharing (
        id bigserial PRIMARY KEY,
        ts TIMESTAMP NOT NULL,
        SHARE bool NOT NULL,
        user_id BIGINT NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users (id)
    );

CREATE TABLE IF NOT EXISTS
    finality_checkpoints (
        head_epoch INT NOT NULL,
        head_root bytea NOT NULL,
        current_justified_epoch INT NOT NULL,
        current_justified_root bytea NOT NULL,
        previous_justified_epoch INT NOT NULL,
        previous_justified_root bytea NOT NULL,
        finalized_epoch INT NOT NULL,
        finalized_root bytea NOT NULL,
        PRIMARY KEY (head_epoch, head_root)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_export_status (
        rocketpool_storage_address bytea NOT NULL,
        eth1_block INT NOT NULL,
        PRIMARY KEY (rocketpool_storage_address)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_minipools (
        rocketpool_storage_address bytea NOT NULL,
        address bytea NOT NULL,
        pubkey bytea NOT NULL,
        node_address bytea NOT NULL,
        node_fee FLOAT NOT NULL,
        deposit_type VARCHAR(20) NOT NULL,
        -- none (invalid), full, half, empty .. see: https://github.com/rocket-pool/rocketpool/blob/683addf4ac/contracts/types/MinipoolDeposit.sol
        status TEXT NOT NULL,
        -- Initialized, Prelaunch, Staking, Withdrawable, Dissolved .. see: https://github.com/rocket-pool/rocketpool/blob/683addf4ac/contracts/types/MinipoolStatus.sol
        status_time TIMESTAMP WITHOUT TIME ZONE,
        penalty_count NUMERIC NOT NULL DEFAULT 0,
        node_deposit_balance NUMERIC,
        node_refund_balance NUMERIC,
        user_deposit_balance NUMERIC,
        is_vacant BOOLEAN DEFAULT FALSE,
        VERSION INT DEFAULT 0,
        PRIMARY KEY (rocketpool_storage_address, address)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_nodes (
        rocketpool_storage_address bytea NOT NULL,
        address bytea NOT NULL,
        timezone_location VARCHAR(200) NOT NULL,
        rpl_stake NUMERIC NOT NULL,
        min_rpl_stake NUMERIC NOT NULL,
        max_rpl_stake NUMERIC NOT NULL,
        rpl_cumulative_rewards NUMERIC NOT NULL DEFAULT 0,
        smoothing_pool_opted_in BOOLEAN NOT NULL DEFAULT FALSE,
        claimed_smoothing_pool NUMERIC NOT NULL,
        unclaimed_smoothing_pool NUMERIC NOT NULL,
        unclaimed_rpl_rewards NUMERIC NOT NULL,
        effective_rpl_stake NUMERIC NOT NULL,
        deposit_credit NUMERIC,
        PRIMARY KEY (rocketpool_storage_address, address)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_dao_proposals (
        rocketpool_storage_address bytea NOT NULL,
        id INT NOT NULL,
        dao TEXT NOT NULL,
        proposer_address bytea NOT NULL,
        message TEXT NOT NULL,
        created_time TIMESTAMP WITHOUT TIME ZONE,
        start_time TIMESTAMP WITHOUT TIME ZONE,
        end_time TIMESTAMP WITHOUT TIME ZONE,
        expiry_time TIMESTAMP WITHOUT TIME ZONE,
        votes_required FLOAT NOT NULL,
        votes_for FLOAT NOT NULL,
        votes_against FLOAT NOT NULL,
        member_voted BOOLEAN NOT NULL,
        member_supported BOOLEAN NOT NULL,
        is_cancelled BOOLEAN NOT NULL,
        is_executed BOOLEAN NOT NULL,
        payload bytea NOT NULL,
        state TEXT NOT NULL,
        PRIMARY KEY (rocketpool_storage_address, id)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_dao_proposals_member_votes (
        rocketpool_storage_address bytea NOT NULL,
        id INT NOT NULL,
        member_address bytea NOT NULL,
        voted BOOLEAN NOT NULL,
        supported BOOLEAN NOT NULL,
        PRIMARY KEY (rocketpool_storage_address, id, member_address)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_dao_members (
        rocketpool_storage_address bytea NOT NULL,
        address bytea NOT NULL,
        id VARCHAR(200) NOT NULL,
        url VARCHAR(200) NOT NULL,
        joined_time TIMESTAMP WITHOUT TIME ZONE,
        last_proposal_time TIMESTAMP WITHOUT TIME ZONE,
        rpl_bond_amount NUMERIC NOT NULL,
        unbonded_validator_count INT NOT NULL,
        PRIMARY KEY (rocketpool_storage_address, address)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_network_stats (
        id bigserial,
        ts TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        rpl_price NUMERIC NOT NULL,
        claim_interval_time INTERVAL NOT NULL,
        claim_interval_time_start TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        current_node_fee FLOAT NOT NULL,
        current_node_demand NUMERIC NOT NULL,
        reth_supply NUMERIC NOT NULL,
        effective_rpl_staked NUMERIC NOT NULL,
        node_operator_rewards NUMERIC NOT NULL,
        reth_exchange_rate FLOAT NOT NULL,
        node_count NUMERIC NOT NULL,
        minipool_count NUMERIC NOT NULL,
        odao_member_count NUMERIC NOT NULL,
        total_eth_staking NUMERIC NOT NULL,
        total_eth_balance NUMERIC NOT NULL,
        PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS
    rocketpool_reward_tree (
        id bigserial,
        DATA jsonb NOT NULL,
        PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS
    eth_store_stats (
        DAY INT NOT NULL,
        VALIDATOR INT NOT NULL,
        effective_balances_sum_wei NUMERIC NOT NULL,
        start_balances_sum_wei NUMERIC NOT NULL,
        end_balances_sum_wei NUMERIC NOT NULL,
        deposits_sum_wei NUMERIC NOT NULL,
        tx_fees_sum_wei NUMERIC NOT NULL,
        consensus_rewards_sum_wei NUMERIC NOT NULL,
        total_rewards_wei NUMERIC NOT NULL,
        apr FLOAT NOT NULL,
        PRIMARY KEY (DAY, VALIDATOR)
    );

CREATE INDEX IF NOT EXISTS idx_eth_store_validator ON eth_store_stats (VALIDATOR, DAY DESC);

CREATE TABLE IF NOT EXISTS
    historical_pool_performance (
        DAY INT NOT NULL,
        pool VARCHAR(40) NOT NULL,
        validators INT NOT NULL,
        effective_balances_sum_wei NUMERIC NOT NULL,
        start_balances_sum_wei NUMERIC NOT NULL,
        end_balances_sum_wei NUMERIC NOT NULL,
        deposits_sum_wei NUMERIC NOT NULL,
        tx_fees_sum_wei NUMERIC NOT NULL,
        consensus_rewards_sum_wei NUMERIC NOT NULL,
        total_rewards_wei NUMERIC NOT NULL,
        apr FLOAT NOT NULL,
        PRIMARY KEY (DAY, pool)
    );

CREATE INDEX IF NOT EXISTS idx_historical_pool_performance_pool ON historical_pool_performance (pool, DAY DESC);

CREATE TABLE IF NOT EXISTS
    tags (
        id VARCHAR NOT NULL,
        metadata jsonb NOT NULL,
        PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS
    blocks_tags (
        slot int4 NOT NULL,
        blockroot bytea NOT NULL,
        tag_id VARCHAR NOT NULL,
        PRIMARY KEY (slot, blockroot, tag_id),
        FOREIGN KEY (slot, blockroot) REFERENCES blocks (slot, blockroot),
        FOREIGN KEY (tag_id) REFERENCES tags (id)
    );

CREATE INDEX IF NOT EXISTS idx_blocks_tags_slot ON blocks_tags (slot);

CREATE INDEX IF NOT EXISTS idx_blocks_tags_tag_id ON blocks_tags (tag_id);

CREATE TABLE IF NOT EXISTS
    relays (
        tag_id VARCHAR NOT NULL,
        endpoint VARCHAR NOT NULL,
        public_link VARCHAR NULL,
        is_censoring bool NULL,
        is_ethical bool NULL,
        export_failure_count INT NOT NULL DEFAULT 0,
        last_export_try_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
        last_export_success_ts TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
        PRIMARY KEY (tag_id, endpoint),
        FOREIGN KEY (tag_id) REFERENCES tags (id)
    );

CREATE TABLE IF NOT EXISTS
    relays_blocks (
        tag_id VARCHAR NOT NULL,
        block_slot int4 NOT NULL,
        block_root bytea NOT NULL,
        exec_block_hash bytea NOT NULL,
        builder_pubkey bytea NOT NULL,
        proposer_pubkey bytea NOT NULL,
        proposer_fee_recipient bytea NOT NULL,
        VALUE NUMERIC NOT NULL,
        PRIMARY KEY (block_slot, block_root, tag_id)
    );

CREATE INDEX IF NOT EXISTS relays_blocks_block_root_idx ON public.relays_blocks (block_root);

CREATE INDEX IF NOT EXISTS relays_blocks_builder_pubkey_idx ON public.relays_blocks (builder_pubkey);

CREATE INDEX IF NOT EXISTS relays_blocks_exec_block_hash_idx ON public.relays_blocks (exec_block_hash);

CREATE INDEX IF NOT EXISTS relays_blocks_value_idx ON public.relays_blocks (VALUE);

CREATE TABLE IF NOT EXISTS
    validator_queue_deposits (
        validatorindex int4 NOT NULL,
        block_slot int4 NULL,
        block_index int4 NULL,
        CONSTRAINT validator_queue_deposits_fk FOREIGN KEY (block_slot, block_index) REFERENCES blocks_deposits (block_slot, block_index),
        CONSTRAINT validator_queue_deposits_fk_validators FOREIGN KEY (validatorindex) REFERENCES validators (validatorindex)
    );

CREATE INDEX IF NOT EXISTS idx_validator_queue_deposits_block_slot ON validator_queue_deposits USING btree (block_slot);

CREATE UNIQUE INDEX IF NOT EXISTS idx_validator_queue_deposits_validatorindex ON validator_queue_deposits USING btree (validatorindex);

CREATE TABLE IF NOT EXISTS
    service_status (
        NAME TEXT NOT NULL,
        executable_name TEXT NOT NULL,
        VERSION TEXT NOT NULL,
        pid INT NOT NULL,
        status TEXT NOT NULL,
        metadata jsonb,
        last_update TIMESTAMP NOT NULL,
        PRIMARY KEY (NAME, executable_name, VERSION, pid)
    );

CREATE TABLE IF NOT EXISTS
    chart_series (
        "time" TIMESTAMP WITHOUT TIME ZONE NOT NULL,
        indicator CHARACTER VARYING(50) NOT NULL,
        VALUE NUMERIC NOT NULL,
        PRIMARY KEY ("time", indicator)
    );

CREATE TABLE IF NOT EXISTS
    chart_series_status (
        DAY INT NOT NULL,
        status BOOLEAN NOT NULL,
        PRIMARY KEY (DAY)
    );

CREATE TABLE IF NOT EXISTS
    global_notifications (
        target VARCHAR(20) NOT NULL PRIMARY KEY,
        CONTENT TEXT NOT NULL,
        enabled bool NOT NULL
    );

CREATE TABLE IF NOT EXISTS
    node_jobs (
        id VARCHAR(40),
        TYPE VARCHAR(40) NOT NULL,
        -- can be one of: BLS_TO_EXECUTION_CHANGES, VOLUNTARY_EXITS
        status VARCHAR(40) NOT NULL,
        -- can be one of: PENDING, SUBMITTED_TO_NODE, COMPLETED
        created_time TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
        submitted_to_node_time TIMESTAMP WITHOUT TIME ZONE,
        completed_time TIMESTAMP WITHOUT TIME ZONE,
        DATA jsonb NOT NULL,
        PRIMARY KEY (id)
    );

CREATE TABLE IF NOT EXISTS
    ad_configurations (
        id VARCHAR(40),
        --uuid
        template_id VARCHAR(100) NOT NULL,
        --relative path to the main html file of the page
        jquery_selector VARCHAR(40) NOT NULL,
        --selector with the html
        insert_mode VARCHAR(10) NOT NULL,
        -- can be before, after, replace or insert
        refresh_interval INT NOT NULL,
        -- defines how often the ad is refreshed in seconds, 0 = don't refresh
        enabled bool NOT NULL,
        -- defines if the ad is active
        for_all_users bool NOT NULL,
        -- if set the ad will be shown to all users even if they have NoAds
        banner_id INT,
        -- an ad must either have a banner_id OR an html_content
        html_content TEXT,
        PRIMARY KEY (id)
    );

CREATE INDEX IF NOT EXISTS idx_ad_configuration_for_template ON ad_configurations (template_id, enabled);

CREATE TABLE IF NOT EXISTS
    explorer_configurations (
        category VARCHAR(40),
        -- the category is used to group settings together
        KEY VARCHAR(40),
        -- identifies the config
        VALUE TEXT,
        -- holds the value of the config
        data_type VARCHAR(40),
        -- defines to which data type the value is mapped
        PRIMARY KEY (category, KEY)
    );

CREATE INDEX IF NOT EXISTS idx_explorer_configurations ON explorer_configurations (category, KEY);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'NOT SUPPORTED';
-- +goose StatementEnd

/* 
 
 # validators-states
 
 see: https://hackmd.io/ofFJ5gOmQpu1jjHilHbdQQ 
 
 Possible statuses:
 
 pending_initialized - When the first deposit is processed, but not enough funds are available (or not yet the end of the first epoch) to get validator into the activation queue.
 pending_queued      - When validator is waiting to get activated, and have enough funds etc. while in the queue, validator activation epoch keeps changing until it gets to the front and make it through (finalization is a requirement here too).
 active_ongoing      - When validator must be attesting, and have not initiated any exit.
 active_exiting      - When validator is still active, but filed a voluntary request to exit.
 active_slashed      - When validator is still active, but have a slashed status and is scheduled to exit.
 exited_unslashed    - When validator has reached reguler exit epoch, not being slashed, and doesn't have to attest any more, but cannot withdraw yet.
 exited_slashed      - When validator has reached reguler exit epoch, but was slashed, have to wait for a longer withdrawal period.
 withdrawal_possible - After validator has exited, a while later is permitted to move funds, and is truly out of the system.
 withdrawal_done     - (not possible in phase0, except slashing full balance) - actually having moved funds away
 
 we add:
 
 deposited_invalid
 deposited_valid
 
 */
/*
 https://ethereum.github.io/eth2.0-APIs/#/Beacon/getStateValidators
 - replacing slashed (bool) with slashed_epoch (uint64)
 - adding last_attestation_slot
 - adding last_proposal_slot
 */
drop table if exists validators;
create table validators (
    validator_index int not null,
    status text not null,
    pubkey bytea not null,
    withdrawal_credentials bytea not null,
    balance bigint not null,
    effective_balance bigint not null,
    withdrawable_epoch bigint not null,
    slashed_epoch bigint not null,
    exit_epoch bigint not null,
    activation_eligibility_epoch bigint not null,
    activation_epoch bigint not null,
    last_attestation_slot bigint,
    last_proposal_slot bigint,
    primary key (validator_index)
);
create index idx_validators_pubkey on validators (pubkey);

/*
 validator-balances per epoch
 */
drop table if exists validator_balances;
create table validator_balances (
    epoch int not null,
    validator_index int not null,
    balance bigint not null,
    effective_balance bigint not null,
    primary key (epoch, validator_index)
);
create index idx_validator_balances_epoch on validator_balances (epoch);

/*
 stats per epoch
 */
drop table if exists epochs;
create table epochs (
    epoch int not null,
    finalized bool not null,
    blocks_count int not null,
    proposer_slashings_count int not null,
    attester_slashings_count int not null,
    attestations_executed_count int not null,
    attestations_missed_count int not null,
    deposits_count int not null,
    voluntary_exits_count int not null,
    validators_count int not null,
    validators_total_balance bigint not null,
    validators_total_effective_balance bigint not null,
    validators_pending_initialized_count int not null,
    validators_pending_initialized_total_balance bigint not null,
    validators_pending_initialized_total_effective_balance bigint not null,
    validators_pending_queued_count int not null,
    validators_pending_queued_total_balance bigint not null,
    validators_pending_queued_total_effective_balance bigint not null,
    validators_active_ongoing_count int not null,
    validators_active_ongoing_total_balance bigint not null,
    validators_active_ongoing_total_effective_balance bigint not null,
    validators_active_exiting_count int not null,
    validators_active_exiting_total_balance bigint not null,
    validators_active_exiting_total_effective_balance bigint not null,
    validators_active_slashed_count int not null,
    validators_active_slashed_total_balance bigint not null,
    validators_active_slashed_total_effective_balance bigint not null,
    validators_exited_unslashed_count int not null,
    validators_exited_unslashed_total_balance bigint not null,
    validators_exited_unslashed_total_effective_balance bigint not null,
    validators_exited_slashed_count int not null,
    validators_exited_slashed_total_balance bigint not null,
    validators_exited_slashed_total_effective_balance bigint not null,
    attesting_gwei bigint not null,
    target_attesting_gwei bigint not null,
    source_attesting_gwei bigint not null,
    head_attesting_gwei bigint not null,
    previous_justified_epoch int not null,
    previous_justified_root bytea not null,
    current_justified_epoch int not null,
    current_justified_root bytea not null,
    finalized_epoch int not null,
    finalized_root bytea not null,
    primary key (epoch)
);

/*
 https://ethereum.github.io/eth2.0-APIs/#/Beacon/getBlock
 https://github.com/ethereum/eth2.0-specs/blob/v0.12.2/specs/phase0/beacon-chain.md#signedbeaconblock
 - without containers (stored in separate tables)
 - adding counts of items in containers
 - adding status
 */
drop table if exists blocks;
create table blocks (
    epoch int not null,
    slot int not null,
    root bytea not null,
    parent_root bytea not null,
    state_root bytea not null,
    signature bytea not null,
    randao_reveal bytea not null,
    graffiti bytea not null,
    proposer_index int not null,
    status text not null,
    /* Can be 0 = scheduled, 1 proposed, 2 missed, 3 orphaned */
    eth1_data_deposit_root bytea not null,
    eth1_data_deposit_count int not null,
    eth1_data_block_hash bytea not null,
    proposer_slashings_count int not null,
    attester_slashings_count int not null,
    attestations_count int not null,
    deposits_count int not null,
    voluntary_exits_count int not null,
    primary key (slot, blockroot)
);

drop table if exists blocks_attestations;
create table blocks_attestations (
    block_root bytea not null,
    block_slot int not null,
    index int not null,
    aggregation_bits bytea not null,
    signature bytea not null,
    beacon_block_root bytea not null,
    source_epoch int not null,
    source_root bytea not null,
    target_epoch int not null,
    target_root bytea not null,
    primary key (block_root),
    constraint fk_block_root
);

drop table if exists blocks_deposits;
create table blocks_deposits (
    block_root             bytea   not null,
    block_slot             int     not null,
    proof                  bytea[] not null,
    pubkey                 bytea   not null,
    withdrawal_credentials bytea   not null,
    amount                 bigint  not null,
    signature              bytea   not null,
    primary key (block_root)
);

drop table if exists blocks_voluntary_exits;
create table blocks_voluntaryexits (
    block_root      int not null,
    block_slot      int not null,
    epoch           int not null,
    slot            int not null,
    validator_index int not null,
    signature      bytea not null,
    primary key (block_root)
);

drop table if exists blocks_proposer_slashings;
create table blocks_proposer_slashings (
    block_root bytea not null,
    block_slot int not null,
    proposer_index int not null,
    header1_slot int not null,
    header1_parent_root bytea not null,
    header1_state_root bytea not null,
    header1_body_root bytea not null,
    header1_signature bytea not null,
    header2_slot int not null,
    header2_parent_root bytea not null,
    header2_state_root bytea not null,
    header2_body_root bytea not null,
    header2_signature bytea not null,
    primary key (block_root)
);

drop table if exists blocks_attester_slashings;
create table blocks_attester_slashings (
    block_root bytea not null,
    block_slot int not null,
    attestation1_indices int [] not null,
    attestation1_signature bytea not null,
    attestation1_slot bigint not null,
    attestation1_index int not null,
    attestation1_beacon_block_root bytea not null,
    attestation1_source_epoch int not null,
    attestation1_source_root bytea not null,
    attestation1_target_epoch int not null,
    attestation1_target_root bytea not null,
    attestation2_indices int [] not null,
    attestation2_signature bytea not null,
    attestation2_slot bigint not null,
    attestation2_index int not null,
    attestation2_beacon_bloc_kroot bytea not null,
    attestation2_source_epoch int not null,
    attestation2_source_root bytea not null,
    attestation2_target_epoch int not null,
    attestation2_target_root bytea not null,
    primary key (block_root)
);

/* 
 https://ethereum.github.io/eth2.0-APIs/#/Validator/getProposerDuties 
 without pubkey
 */
drop table if exists attester_duties;
create table attester_duties (
    epoch int not null,
    slot int not null,
    validator_index int not null,
    validator_committee_index int not null,
    committee_index int not null,
    committee_length int not null,
    committees_at_slot int not null,
    primary key (epoch, validator_index)
);

/* 
 https://ethereum.github.io/eth2.0-APIs/#/Validator/getProposerDuties 
 without pubkey
 */
drop table if exists proposer_duties;
create table attester_duties (
    epoch int not null,
    validator_index int not null,
    primary key (epoch, validator_index)
);

drop table if exists validator_performance;
create table validator_performance (
    validatorindex int not null,
    balance bigint not null,
    performance1d bigint not null,
    performance7d bigint not null,
    performance31d bigint not null,
    performance365d bigint not null,
    primary key (validatorindex)
);
create index idx_validator_performance_balance on validator_performance (balance);
create index idx_validator_performance_performance1d on validator_performance (performance1d);
create index idx_validator_performance_performance7d on validator_performance (performance7d);
create index idx_validator_performance_performance31d on validator_performance (performance31d);
create index idx_validator_performance_performance365d on validator_performance (performance365d);

drop table if exists validator_balances;
create table validator_balances (
    epoch int not null,
    validator_index int not null,
    balance bigint not null,
    effectivebalance bigint not null,
    primary key (validator_index, epoch)
);
create index idx_validator_balances_epoch on validator_balances (epoch);

drop table if exists blocks;
create table blocks (
    epoch int not null,
    slot int not null,
    blockroot bytea not null,
    parentroot bytea not null,
    stateroot bytea not null,
    signature bytea not null,
    randaoreveal bytea not null,
    graffiti bytea not null,
    eth1data_depositroot bytea not null,
    eth1data_depositcount int not null,
    eth1data_blockhash bytea not null,
    proposerslashingscount int not null,
    attesterslashingscount int not null,
    attestationscount int not null,
    depositscount int not null,
    voluntaryexitscount int not null,
    proposer int not null,
    status text not null,
    primary key (slot, blockroot)
);
create index idx_blocks_proposer on blocks (proposer);

drop table if exists blocks_proposerslashings;
create table blocks_proposerslashings (
    block_slot int not null,
    block_index int not null,
    block_root bytea not null,
    proposer_index int not null,
    header1_slot int not null,
    header1_parentroot bytea not null,
    header1_stateroot bytea not null,
    header1_bodyroot bytea not null,
    header1_signature bytea not null,
    header2_slot int not null,
    header2_parentroot bytea not null,
    header2_stateroot bytea not null,
    header2_bodyroot bytea not null,
    header2_signature bytea not null,
    primary key (block_slot, block_index)
);

drop table if exists blocks_attesterslashings;
create table blocks_attesterslashings (
    block_slot int not null,
    block_index int not null,
    attestation1_signature bytea not null,
    attestation1_slot int not null,
    attestation1_index int not null,
    attestation1_beaconblockroot bytea not null,
    attestation1_source_epoch int not null,
    attestation1_source_root bytea not null,
    attestation1_target_epoch int not null,
    attestation1_target_root bytea not null,
    attestation2_signature bytea not null,
    attestation2_slot int not null,
    attestation2_index int not null,
    attestation2_beaconblockroot bytea not null,
    attestation2_source_epoch int not null,
    attestation2_source_root bytea not null,
    attestation2_target_epoch int not null,
    attestation2_target_root bytea not null,
    primary key (block_slot, block_index)
);

drop table if exists blocks_attestations;
create table blocks_attestations (
    block_slot int not null,
    block_index int not null,
    aggregationbits bytea not null,
    validators int [] not null,
    signature bytea not null,
    slot int not null,
    committeeindex int not null,
    beaconblockroot bytea not null,
    source_epoch int not null,
    source_root bytea not null,
    target_epoch int not null,
    target_root bytea not null,
    primary key (block_slot, block_index)
);
create index idx_blocks_attestations_beaconblockroot on blocks_attestations (beaconblockroot);
create index idx_blocks_attestations_source_root on blocks_attestations (source_root);
create index idx_blocks_attestations_target_root on blocks_attestations (target_root);

drop table if exists blocks_deposits;
create table blocks_deposits (
    block_slot int not null,
    block_index int not null,
    proof bytea [],
    publickey bytea not null,
    withdrawalcredentials bytea not null,
    amount bigint not null,
    signature bytea not null,
    primary key (block_slot, block_index)
);

drop table if exists blocks_voluntaryexits;
create table blocks_voluntaryexits (
    block_slot int not null,
    block_index int not null,
    epoch int not null,
    validatorindex int not null,
    signature bytea not null,
    primary key (block_slot, block_index)
);

drop table if exists users;
create table users (
    id serial not null,
    username character varying(32) not null unique,
    password character varying(256) not null,
    email character varying(100),
    primary key (id)
);

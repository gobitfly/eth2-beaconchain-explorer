--
-- PostgreSQL database dump
--

-- Dumped from database version 12.4
-- Dumped by pg_dump version 12.4 (Ubuntu 12.4-0ubuntu0.20.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_statistics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_statistics (
    ts timestamp without time zone NOT NULL,
    apikey character varying(64) NOT NULL,
    call character varying(64) NOT NULL,
    count integer DEFAULT 0 NOT NULL
);


--
-- Name: attestation_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attestation_assignments (
    epoch integer NOT NULL,
    validatorindex integer NOT NULL,
    attesterslot integer NOT NULL,
    committeeindex integer NOT NULL,
    status integer NOT NULL,
    inclusionslot integer DEFAULT 0 NOT NULL
);


--
-- Name: blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks (
    epoch integer NOT NULL,
    slot integer NOT NULL,
    blockroot bytea NOT NULL,
    parentroot bytea NOT NULL,
    stateroot bytea NOT NULL,
    signature bytea NOT NULL,
    randaoreveal bytea,
    graffiti bytea,
    graffiti_text text,
    eth1data_depositroot bytea,
    eth1data_depositcount integer NOT NULL,
    eth1data_blockhash bytea,
    proposerslashingscount integer NOT NULL,
    attesterslashingscount integer NOT NULL,
    attestationscount integer NOT NULL,
    depositscount integer NOT NULL,
    voluntaryexitscount integer NOT NULL,
    proposer integer NOT NULL,
    status text NOT NULL
);


--
-- Name: blocks_attestations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks_attestations (
    block_slot integer NOT NULL,
    block_index integer NOT NULL,
    aggregationbits bytea NOT NULL,
    validators integer[] NOT NULL,
    signature bytea NOT NULL,
    slot integer NOT NULL,
    committeeindex integer NOT NULL,
    beaconblockroot bytea NOT NULL,
    source_epoch integer NOT NULL,
    source_root bytea NOT NULL,
    target_epoch integer NOT NULL,
    target_root bytea NOT NULL
);


--
-- Name: blocks_attesterslashings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks_attesterslashings (
    block_slot integer NOT NULL,
    block_index integer NOT NULL,
    attestation1_indices integer[] NOT NULL,
    attestation1_signature bytea NOT NULL,
    attestation1_slot bigint NOT NULL,
    attestation1_index integer NOT NULL,
    attestation1_beaconblockroot bytea NOT NULL,
    attestation1_source_epoch integer NOT NULL,
    attestation1_source_root bytea NOT NULL,
    attestation1_target_epoch integer NOT NULL,
    attestation1_target_root bytea NOT NULL,
    attestation2_indices integer[] NOT NULL,
    attestation2_signature bytea NOT NULL,
    attestation2_slot bigint NOT NULL,
    attestation2_index integer NOT NULL,
    attestation2_beaconblockroot bytea NOT NULL,
    attestation2_source_epoch integer NOT NULL,
    attestation2_source_root bytea NOT NULL,
    attestation2_target_epoch integer NOT NULL,
    attestation2_target_root bytea NOT NULL
);


--
-- Name: blocks_deposits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks_deposits (
    block_slot integer NOT NULL,
    block_index integer NOT NULL,
    proof bytea[],
    publickey bytea NOT NULL,
    withdrawalcredentials bytea NOT NULL,
    amount bigint NOT NULL,
    signature bytea NOT NULL
);


--
-- Name: blocks_proposerslashings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks_proposerslashings (
    block_slot integer NOT NULL,
    block_index integer NOT NULL,
    proposerindex integer NOT NULL,
    header1_slot bigint NOT NULL,
    header1_parentroot bytea NOT NULL,
    header1_stateroot bytea NOT NULL,
    header1_bodyroot bytea NOT NULL,
    header1_signature bytea NOT NULL,
    header2_slot bigint NOT NULL,
    header2_parentroot bytea NOT NULL,
    header2_stateroot bytea NOT NULL,
    header2_bodyroot bytea NOT NULL,
    header2_signature bytea NOT NULL
);


--
-- Name: blocks_voluntaryexits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocks_voluntaryexits (
    block_slot integer NOT NULL,
    block_index integer NOT NULL,
    epoch integer NOT NULL,
    validatorindex integer NOT NULL,
    signature bytea NOT NULL
);


--
-- Name: chart_images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chart_images (
    name character varying(100) NOT NULL,
    image bytea NOT NULL
);


--
-- Name: epochs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.epochs (
    epoch integer NOT NULL,
    blockscount integer DEFAULT 0 NOT NULL,
    proposerslashingscount integer NOT NULL,
    attesterslashingscount integer NOT NULL,
    attestationscount integer NOT NULL,
    depositscount integer NOT NULL,
    voluntaryexitscount integer NOT NULL,
    validatorscount integer NOT NULL,
    averagevalidatorbalance bigint NOT NULL,
    totalvalidatorbalance bigint NOT NULL,
    finalized boolean,
    eligibleether bigint,
    globalparticipationrate double precision,
    votedether bigint
);


--
-- Name: eth1_deposits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.eth1_deposits (
    tx_hash bytea NOT NULL,
    tx_input bytea NOT NULL,
    tx_index integer NOT NULL,
    block_number integer NOT NULL,
    block_ts timestamp without time zone NOT NULL,
    from_address bytea NOT NULL,
    publickey bytea NOT NULL,
    withdrawal_credentials bytea NOT NULL,
    amount bigint NOT NULL,
    signature bytea NOT NULL,
    merkletree_index bytea NOT NULL,
    removed boolean NOT NULL,
    valid_signature boolean NOT NULL
);


--
-- Name: graffitiwall; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.graffitiwall (
    x integer NOT NULL,
    y integer NOT NULL,
    color text NOT NULL,
    slot integer NOT NULL,
    validator integer NOT NULL
);


--
-- Name: mails_sent; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.mails_sent (
    email character varying(100) NOT NULL,
    ts timestamp without time zone NOT NULL,
    cnt integer NOT NULL
);


--
-- Name: network_liveness; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.network_liveness (
    ts timestamp without time zone NOT NULL,
    headepoch integer NOT NULL,
    finalizedepoch integer NOT NULL,
    justifiedepoch integer NOT NULL,
    previousjustifiedepoch integer NOT NULL
);


--
-- Name: oauth_apps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_apps (
    id integer NOT NULL,
    owner_id integer NOT NULL,
    redirect_uri character varying(100) NOT NULL,
    app_name character varying(35) NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_ts timestamp without time zone NOT NULL
);


--
-- Name: oauth_apps_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_apps_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_apps_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_apps_id_seq OWNED BY public.oauth_apps.id;


--
-- Name: oauth_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_codes (
    id integer NOT NULL,
    user_id integer NOT NULL,
    code character varying(64) NOT NULL,
    consumed boolean DEFAULT false NOT NULL,
    app_id integer NOT NULL,
    created_ts timestamp without time zone NOT NULL
);


--
-- Name: oauth_codes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_codes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_codes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_codes_id_seq OWNED BY public.oauth_codes.id;


--
-- Name: proposal_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.proposal_assignments (
    epoch integer NOT NULL,
    validatorindex integer NOT NULL,
    proposerslot integer NOT NULL,
    status integer NOT NULL
);


--
-- Name: queue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.queue (
    ts timestamp without time zone NOT NULL,
    entering_validators_count integer NOT NULL,
    exiting_validators_count integer NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    password character varying(256) NOT NULL,
    email character varying(100) NOT NULL,
    email_confirmed boolean DEFAULT false NOT NULL,
    email_confirmation_hash character varying(40),
    email_confirmation_ts timestamp without time zone,
    password_reset_hash character varying(40),
    password_reset_ts timestamp without time zone,
    register_ts timestamp without time zone
);


--
-- Name: users_devices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users_devices (
    id integer NOT NULL,
    user_id integer NOT NULL,
    refresh_token character varying(64) NOT NULL,
    device_name character varying(20) NOT NULL,
    notification_token character varying(500),
    notify_enabled boolean DEFAULT false NOT NULL,
    active boolean DEFAULT true NOT NULL,
    app_id integer NOT NULL,
    created_ts timestamp without time zone NOT NULL
);


--
-- Name: users_devices_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_devices_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_devices_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_devices_id_seq OWNED BY public.users_devices.id;


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: users_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users_subscriptions (
    id integer NOT NULL,
    user_id integer NOT NULL,
    event_name character varying(100) NOT NULL,
    event_filter text DEFAULT ''::text NOT NULL,
    last_sent_ts timestamp without time zone,
    last_sent_epoch integer,
    created_ts timestamp without time zone NOT NULL,
    created_epoch integer NOT NULL
);


--
-- Name: users_subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_subscriptions_id_seq OWNED BY public.users_subscriptions.id;


--
-- Name: users_validators_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users_validators_tags (
    user_id integer NOT NULL,
    validator_publickey bytea NOT NULL,
    tag character varying(100) NOT NULL
);


--
-- Name: validator_balances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validator_balances (
    epoch integer NOT NULL,
    validatorindex integer NOT NULL,
    balance bigint NOT NULL,
    effectivebalance bigint NOT NULL
);


--
-- Name: validator_names; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validator_names (
    publickey bytea NOT NULL,
    name character varying(40)
);


--
-- Name: validator_performance; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validator_performance (
    validatorindex integer NOT NULL,
    balance bigint NOT NULL,
    performance1d bigint NOT NULL,
    performance7d bigint NOT NULL,
    performance31d bigint NOT NULL,
    performance365d bigint NOT NULL
);


--
-- Name: validator_set; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validator_set (
    epoch integer NOT NULL,
    validatorindex integer NOT NULL,
    withdrawableepoch bigint NOT NULL,
    withdrawalcredentials bytea NOT NULL,
    effectivebalance bigint NOT NULL,
    slashed boolean NOT NULL,
    activationeligibilityepoch bigint NOT NULL,
    activationepoch bigint NOT NULL,
    exitepoch bigint NOT NULL
);


--
-- Name: validator_status_stats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validator_status_stats (
    epoch integer NOT NULL,
    status character varying(40) NOT NULL,
    validators_count integer NOT NULL,
    total_balance bigint NOT NULL,
    total_effective_balance bigint NOT NULL
);


--
-- Name: validatorqueue_activation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validatorqueue_activation (
    index integer NOT NULL,
    publickey bytea NOT NULL
);


--
-- Name: validatorqueue_exit; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validatorqueue_exit (
    index integer NOT NULL,
    publickey bytea NOT NULL
);


--
-- Name: validators; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.validators (
    validatorindex integer NOT NULL,
    pubkey bytea NOT NULL,
    withdrawableepoch bigint NOT NULL,
    withdrawalcredentials bytea NOT NULL,
    balance bigint NOT NULL,
    effectivebalance bigint NOT NULL,
    slashed boolean NOT NULL,
    activationeligibilityepoch bigint NOT NULL,
    activationepoch bigint NOT NULL,
    exitepoch bigint NOT NULL,
    lastattestationslot bigint,
    status varchar(20) not null default '',
    name character varying(40)
);


--
-- Name: oauth_apps id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_apps ALTER COLUMN id SET DEFAULT nextval('public.oauth_apps_id_seq'::regclass);


--
-- Name: oauth_codes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_codes ALTER COLUMN id SET DEFAULT nextval('public.oauth_codes_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: users_devices id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_devices ALTER COLUMN id SET DEFAULT nextval('public.users_devices_id_seq'::regclass);


--
-- Name: users_subscriptions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_subscriptions ALTER COLUMN id SET DEFAULT nextval('public.users_subscriptions_id_seq'::regclass);


--
-- Name: api_statistics api_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_statistics
    ADD CONSTRAINT api_statistics_pkey PRIMARY KEY (ts, apikey, call);


--
-- Name: attestation_assignments attestation_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attestation_assignments
    ADD CONSTRAINT attestation_assignments_pkey PRIMARY KEY (epoch, validatorindex, attesterslot, committeeindex);


--
-- Name: blocks_attestations blocks_attestations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks_attestations
    ADD CONSTRAINT blocks_attestations_pkey PRIMARY KEY (block_slot, block_index);


--
-- Name: blocks_attesterslashings blocks_attesterslashings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks_attesterslashings
    ADD CONSTRAINT blocks_attesterslashings_pkey PRIMARY KEY (block_slot, block_index);


--
-- Name: blocks_deposits blocks_deposits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks_deposits
    ADD CONSTRAINT blocks_deposits_pkey PRIMARY KEY (block_slot, block_index);


--
-- Name: blocks blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (slot, blockroot);


--
-- Name: blocks_proposerslashings blocks_proposerslashings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks_proposerslashings
    ADD CONSTRAINT blocks_proposerslashings_pkey PRIMARY KEY (block_slot, block_index);


--
-- Name: blocks_voluntaryexits blocks_voluntaryexits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocks_voluntaryexits
    ADD CONSTRAINT blocks_voluntaryexits_pkey PRIMARY KEY (block_slot, block_index);


--
-- Name: chart_images chart_images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chart_images
    ADD CONSTRAINT chart_images_pkey PRIMARY KEY (name);


--
-- Name: epochs epochs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.epochs
    ADD CONSTRAINT epochs_pkey PRIMARY KEY (epoch);


--
-- Name: validator_status_stats epochs_status_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validator_status_stats
    ADD CONSTRAINT epochs_status_stats_pkey PRIMARY KEY (epoch, status);


--
-- Name: eth1_deposits eth1_deposits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.eth1_deposits
    ADD CONSTRAINT eth1_deposits_pkey PRIMARY KEY (tx_hash, merkletree_index);


--
-- Name: graffitiwall graffitiwall_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graffitiwall
    ADD CONSTRAINT graffitiwall_pkey PRIMARY KEY (x, y);


--
-- Name: mails_sent mails_sent_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.mails_sent
    ADD CONSTRAINT mails_sent_pkey PRIMARY KEY (email, ts);


--
-- Name: network_liveness network_liveness_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.network_liveness
    ADD CONSTRAINT network_liveness_pkey PRIMARY KEY (ts);


--
-- Name: oauth_apps oauth_apps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_apps
    ADD CONSTRAINT oauth_apps_pkey PRIMARY KEY (id, redirect_uri);


--
-- Name: oauth_apps oauth_apps_redirect_uri_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_apps
    ADD CONSTRAINT oauth_apps_redirect_uri_key UNIQUE (redirect_uri);


--
-- Name: oauth_codes oauth_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_codes
    ADD CONSTRAINT oauth_codes_pkey PRIMARY KEY (user_id, code);


--
-- Name: proposal_assignments proposal_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.proposal_assignments
    ADD CONSTRAINT proposal_assignments_pkey PRIMARY KEY (epoch, validatorindex, proposerslot);


--
-- Name: queue queue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.queue
    ADD CONSTRAINT queue_pkey PRIMARY KEY (ts);


--
-- Name: users_devices users_devices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_devices
    ADD CONSTRAINT users_devices_pkey PRIMARY KEY (user_id, refresh_token);


--
-- Name: users users_email_confirmation_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_confirmation_hash_key UNIQUE (email_confirmation_hash);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_id_key UNIQUE (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id, email);


--
-- Name: users_subscriptions users_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_subscriptions
    ADD CONSTRAINT users_subscriptions_pkey PRIMARY KEY (user_id, event_name, event_filter);


--
-- Name: users_validators_tags users_validators_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users_validators_tags
    ADD CONSTRAINT users_validators_tags_pkey PRIMARY KEY (user_id, validator_publickey, tag);


--
-- Name: validator_balances validator_balances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validator_balances
    ADD CONSTRAINT validator_balances_pkey PRIMARY KEY (validatorindex, epoch);


--
-- Name: validator_names validator_names_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validator_names
    ADD CONSTRAINT validator_names_pkey PRIMARY KEY (publickey);


--
-- Name: validator_performance validator_performance_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validator_performance
    ADD CONSTRAINT validator_performance_pkey PRIMARY KEY (validatorindex);


--
-- Name: validator_set validator_set_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validator_set
    ADD CONSTRAINT validator_set_pkey PRIMARY KEY (validatorindex, epoch);


--
-- Name: validatorqueue_activation validatorqueue_activation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validatorqueue_activation
    ADD CONSTRAINT validatorqueue_activation_pkey PRIMARY KEY (index, publickey);


--
-- Name: validatorqueue_exit validatorqueue_exit_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validatorqueue_exit
    ADD CONSTRAINT validatorqueue_exit_pkey PRIMARY KEY (index, publickey);


--
-- Name: validators validators_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.validators
    ADD CONSTRAINT validators_pkey PRIMARY KEY (validatorindex);


--
-- Name: idx_attestation_assignments_epoch; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attestation_assignments_epoch ON public.attestation_assignments USING btree (epoch);


--
-- Name: idx_attestation_assignments_validatorindex; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_attestation_assignments_validatorindex ON public.attestation_assignments USING btree (validatorindex);


--
-- Name: idx_blocks_attestations_beaconblockroot; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blocks_attestations_beaconblockroot ON public.blocks_attestations USING btree (beaconblockroot);


--
-- Name: idx_blocks_attestations_source_root; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blocks_attestations_source_root ON public.blocks_attestations USING btree (source_root);


--
-- Name: idx_blocks_attestations_target_root; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blocks_attestations_target_root ON public.blocks_attestations USING btree (target_root);


--
-- Name: idx_blocks_proposer; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_blocks_proposer ON public.blocks USING btree (proposer);


--
-- Name: idx_proposal_assignments_epoch; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_proposal_assignments_epoch ON public.proposal_assignments USING btree (epoch);


--
-- Name: idx_validator_balances_epoch; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validator_balances_epoch ON public.validator_balances USING btree (epoch);


--
-- Name: idx_validator_performance_balance; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validator_performance_balance ON public.validator_performance USING btree (balance);


--
-- Name: idx_validator_performance_performance1d; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validator_performance_performance1d ON public.validator_performance USING btree (performance1d);


--
-- Name: idx_validator_performance_performance31d; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validator_performance_performance31d ON public.validator_performance USING btree (performance31d);


--
-- Name: idx_validator_performance_performance365d; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validator_performance_performance365d ON public.validator_performance USING btree (performance365d);


--
-- Name: idx_validator_performance_performance7d; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validator_performance_performance7d ON public.validator_performance USING btree (performance7d);


--
-- Name: idx_validators_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validators_name ON public.validators USING btree (name);


--
-- Name: idx_validators_pubkey; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_validators_pubkey ON public.validators USING btree (pubkey);


--
-- PostgreSQL database dump complete
--

COPY public.attestation_assignments FROM stdin;
0	16767	11	0	1	0
0	13200	25	3	0	0
0	3496	6	2	1	0
0	1151	6	2	1	0
0	19976	27	3	0	0
0	4481	6	2	1	0
0	13731	5	3	0	0
0	4222	24	1	0	0
0	9403	8	0	1	0
0	920	21	3	0	0
0	15538	0	0	0	0
0	19826	7	2	0	0
0	13547	26	1	0	0
0	5506	11	3	1	0
0	14550	6	2	1	0
0	17701	6	2	1	0
0	8864	17	0	0	0
0	5803	13	1	1	0
0	3332	6	2	1	0
0	842	6	3	0	0
0	5533	5	3	1	0
0	10286	18	1	1	0
0	14530	9	0	0	0
0	5595	18	2	0	0
0	2699	6	2	1	0
0	14741	6	2	1	0
0	9794	8	1	0	0
0	19594	17	1	0	0
0	18001	13	0	0	0
0	7401	19	0	0	0
0	811	26	2	0	0
0	16563	10	2	0	0
0	6717	25	1	0	0
0	8216	29	3	0	0
0	12263	6	2	1	0
0	8355	6	2	1	0
0	10005	5	1	0	0
0	17093	6	2	1	0
0	15641	14	2	0	0
0	741	30	0	0	0
0	12100	6	2	1	0
0	497	5	2	0	0
0	8060	8	1	0	0
0	13391	21	2	0	0
0	652	24	0	0	0
0	15550	23	1	0	0
0	11134	17	1	1	0
0	10808	13	1	1	0
0	287	27	0	0	0
0	13607	23	3	0	0
0	19146	6	2	1	0
0	14006	22	0	0	0
0	17039	6	2	1	0
0	493	30	1	0	0
0	16260	11	2	0	0
0	1061	6	2	1	0
0	181	28	0	0	0
0	9029	24	3	0	0
0	16993	6	2	1	0
0	8486	6	2	1	0
0	9251	11	0	1	0
0	14887	24	2	0	0
0	8182	23	0	0	0
0	5261	11	0	1	0
0	6661	0	3	0	0
0	7328	26	0	0	0
0	16856	26	0	0	0
0	1280	11	0	1	0
0	8805	24	3	1	0
0	629	14	3	0	0
0	6996	14	2	0	0
0	4069	11	0	1	0
0	11175	31	1	1	0
0	9198	6	2	1	0
0	10743	6	2	1	0
0	10691	3	2	1	0
0	5192	11	0	1	0
0	16564	19	2	0	0
0	1012	14	1	0	0
0	10544	29	0	1	0
0	15372	7	1	0	0
0	9270	31	0	0	0
0	17013	26	3	0	0
0	18130	11	0	1	0
0	19018	29	0	0	0
0	16105	0	1	0	0
0	12600	31	3	0	0
0	13533	22	3	0	0
0	18895	11	0	1	0
0	12339	13	1	1	0
0	6316	27	2	0	0
0	15744	23	1	0	0
0	6166	3	2	1	0
0	8898	16	2	0	0
0	4337	28	3	0	0
0	9844	21	0	1	0
0	8379	0	2	0	0
0	13440	23	1	0	0
0	10403	11	0	1	0
0	6719	2	2	0	0
\.
COPY public.blocks FROM stdin;
0	2	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	8597	2
1	32	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	13693	2
1	36	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	12140	2
0	13	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	12540	2
0	18	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	13487	2
0	20	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	12570	2
0	21	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	18403	2
0	24	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	86	2
0	28	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	19455	2
0	30	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	8465	2
1	39	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	15258	2
1	40	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	11858	2
1	41	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	16738	2
1	44	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	7843	2
1	45	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	11879	2
1	47	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	8954	2
1	50	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	1010	2
1	51	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	13138	2
1	54	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	12084	2
0	15	\\x05deee7df488fa4c2d596f0d457dbb1351936c20a9aecd85795ce57e097e5172	\\xeca78a15b029ff130bda662241364b0b1042e60dd483d14db9ded7fdbc821a22	\\x1c81956ec8e0d6fe773fb0b0daae2ac8a4ae189d4f0611c09023437ea4c7edd6	\\xb9e36574dae6a8bb4bc383ded99958433071ec9e8a3f5245ba996368a024c72454c8b0c46aa86468e98d3c7dd3a1e4f016779fc4cf1dd38652ce75cef030f82dacd82d04d10a65cdc4a2bcdb91e8a9bb04a9b9abc530c3d53478bae545342494	\\xad2da6041dc49e39cee77fbbb2edddfbc886d46e26166c80f7580ec3037812f27c209411af4e1181c13cdce29da7a50807d835917658c60691d754bbd3e1f6ea350364181b21e6fa75c38ca2efd971caa0a6e3b8bd061fba5e1b954a385993e3	\\x5032502e4f5247202d205032502056616c696461746f72000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	67	0	0	19806	1
0	8	\\x76262e91970d375a19bfe8a867288d7b9cde43c8635f598d93d39d041706fc76	\\x3f1b68a7a135d12dc613eae6cf894ef2ded091116e550d95ab1d6a5a4f992ce1	\\xed16f3cc818b49da3b9ef2e906ebee4596f00629f417fac6f52f7081a7275a37	\\x8a5482064e482eb9acec7cb23d516713dbbbf49f278abc028e4d199fc00a851ed993e4160b09a323dde7b21ad91d71c518da1434d297a10d3e432fa5602a3a95b8df64d2485d4568a89119fcafb333dbaf9701a9520390a9346603ae5cf84701	\\xb74e300725e5ec4120c5b1cdbfd8707969b5885cb62fb8c045ba7fdeb350af08e8d92bca755f7c993bf510c7fe46d8190c2b9575e1e1b7d3ed8836e635861ad080dc718d31e13a81d82a36b51af24149bdbe13ff03690c4f910b8cd26af4a9a3	\\x706f61703977676769467875534d474f596363736537413442466232366b7742	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	5	0	0	5193	1
1	35	\\x6c36de27a2179afa4523b79655b1fbac93c5b412ef3f5b5b069b1c6ee5398690	\\x0c236734d76c2e641a87bb512f92005783f244c0b01041906081795a50325917	\\x31fbe059fdb9ee4b14544090468f234c2df069370e9400dd5b55c449f99cf40b	\\xae5aba0b595a0452128ce29cae5add1774b50db9e919cfffec192a7b08974fc076ee717b660aa99691177bd7f8d6728a0bcabee3cc6f6fada0b30092537e1c9576f0a1ad70082db83b48244971775f53bc5fed37eb224ed66c9865a0261db781	\\x8348be733fd2e28d8f8aa29f4f5e561eece176feae581ac2b3d5d197767b7f096ed040345c205cfb8cc60178b0085dcc028ec05c77e93366e2b9b5a819575df019f7cbf694950a5642ff6f61ca458e28b83c253a8499a5fdb5e20287bda4fb10	\\x76616c696461746f722d3535353834393835642d347236647a00000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	81	0	0	3775	1
0	23	\\x7f8ce6d918d6a40acfacdbaa2120f2697b5248e5bc4a24e3fafb7d3ba559040b	\\xf2f0c67839f85a6dfe6b764daf49611623c19eae69b7862148ada86bd8f671d5	\\xfac664aa4a372a273aeed7c0d6d613d027af95f97e117daafd9d52f1ffda08d5	\\x815e55680e25a02211fbc01e3667d31c874bb7f4378c25cc419a351efb68fba4cce7f08cca848f45d07fc6c6886094730c76fca649b1e87072f77581e5cd988cba2e46b23d89a93d5b806d87884f7451f769446ec331c870cdc3d6218d4b7ca1	\\xac6751687f837d13ac5f59ae0057344eba01fe923ae99702f373a341f6583576d01b677f81517e076a49afec7c4d4be80e492a679b30df2fa0f70cf570edfafb28a42e4355331cf10a2cdbf4badcef53d267c376aa19c91e687ca9eb6e71197b	\\x53746566616e2333393137000000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	128	0	0	2895	1
0	14	\\xeca78a15b029ff130bda662241364b0b1042e60dd483d14db9ded7fdbc821a22	\\x8f181722abe3a0383815ec5b71acef8e93b3002508e5d298686719f3c461371a	\\x48695e3a2e4c23758c7b99e9e236d394222864f7ea6ddbf2ff8e67aec0c329cd	\\xb3ac4fc44a6f0ebd67ec1e2d8bd859729fd19e0cf3eeaef83f3af3cbd67099ade7872ca2512851ef9d18bf28ff7a5a9f11089b90f362e960c6534354fc4ff9c28f9b8950161132d3f25740f87c5c2772d20112933dab15c42de1d92200fb0c6c	\\xa1e002fdeea4b9bd68ab5d25fe7896960eeeafb6f81b5208fed1dad875e9c5007a19b0e45f74a9d4797c07d57a3f8d9a0a409a33db110de7f1cdd32c7993aa8a14ca2185afac65b852e48b9645b01c112d597dc16a4c452922087eab2d009053	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	44	0	0	6859	1
1	37	\\x43df93c3e43815efadee5a331028236db1bc8f8d3300034f0889e9071118c591	\\x6c36de27a2179afa4523b79655b1fbac93c5b412ef3f5b5b069b1c6ee5398690	\\x971aaaee95995e6c34863cb402f3119379918e84a5784d04e73676315c7ea1ca	\\xa1d51fe949bd6c81e45ad7d103daadb0d5ab7a282dbc8e9991d4b0a2522688e4b102690270611ae3bb93cae4d3490afb08fcbb965cafa57773f91f42f00ecbd4cc8e60d1584098cf2b55165fc2fe0b17025d2d22d1d744a140bf39fd74aea8c7	\\x917c9dbc96742303290886ce456987270c7d30c484473575356111065b0a522d9a63a2e3e6d5d32167945ad02d82a38d03dee1031dc76c101af65c700544d08e551efce91123e244e4c56b047b0fc63f83e51e1be832e318613c16699c593027	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	61	0	0	11460	1
0	7	\\x3f1b68a7a135d12dc613eae6cf894ef2ded091116e550d95ab1d6a5a4f992ce1	\\xd346e6e92200d0ac2f9d5d946a79791b0c3621baaf6e5b20e2772bc7bdb410cb	\\x8ed92658ecef75538fdd098024634ca2e21357db5eb29698e03519523733e9e4	\\x8fad30a1391a4f71e5555272da3a3c3ab7c08f6fed1a1b570a4155f464d002085c239f67b4a40fb314a471f5813b2b9d13d6ab6fed0c7681a510ce5899c6ba80e4fce57105c83e34c59496c3905354878c6774e6448f777c47659650df6c757b	\\xb257116c70bf7aea18197bb3f650474c3a668e66c44d9cec7bacf02735b94bbb708ed16cdabf42f71215d85b7f2d751f1514304fc1b490e36729df02fb5ec0c6333ff6569c66a0b70037541f79aa41c426da812b74389d5a21f6f85909151151	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	15	0	0	11087	1
1	55	\\x06f9c4b8e316eed1a5c7968850ab82f0977774eb7616d8ff0a8394b0e1f0944f	\\x07d6f3ebae0bacd67a58939fb15b1fcb014b2239f76f78a608ba46faee7e2aa9	\\x68ad6078c37aecc493a31d6a80cc2e5effeb65ae9b1bfb44da79bc3a45673518	\\xa4727473dc61dc0a8205a0fd123110ff4e9fdcc6f3eade537546557222e85673354a3bf73e3db04faeca851a3c84485512db5996e985afaeb479049d4b4679baf849f01cf201cb75a65091f226f102b2a5275b387072e94c6a998bafbf108397	\\x98a1b5aa05db83250fa1917f7871529e56f538186cc7aeed337fc8313328a5be5c11f79a065b2e60aa9c653cdf9d8a121806bc0e41deeb1d192d5352a00a56e0c5961a8eac38fcccef35032eb0d3010f3bcbbd58f5f74a9c947d9d0d05be88e5	\\x706f61706734654d372f637752692f5a68615a67307a7036623941364a6c6341	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	113	0	0	4549	1
0	31	\\xe21c3247b62fb3585889e054c16b622bb997f27ac070929924f0176c110bbbe2	\\x7af6146507bdb26b828c94ec233be37aa7549ff23a684e53543a4a062d5a7dce	\\x2747aa33a6fcf2f542b9d446efcc5afdebc199229428a83a772061191ac1de1a	\\xb1148b64502e7bdb8b4447a3e0ed28dc3f4f6d688686314d1f9e788adc049cb902388c8e13cd98f8f241d4973870efa501de8fbe0c9297d895874ecdabf310104031f4710786e43fea3bf53289094be2e6e3e521ee195e0ccf36e60e411da4bf	\\xa123840d7f0825e115f557ba6ae6b589060c0cda2c65511afc390338d300f565a53cbcfdbd44600c761b0e0fe637e8911935c685027f1a0c964ca4b4e0aa73b42914ee5a97528c40b9edc202366d1b34b4de3c44c1cc0239461d31b64995fad8	\\x706f61707368777a33682b72503646556d6359725766344d77795541494e4541	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	128	0	0	17282	1
1	43	\\x23b11fbb75fe343946e69b3d1de30c3a3b8e297806c5583a7e389de94532f43c	\\x699f8f631f40d3d11b33a8fb331304f613fcb270fcdbe41d1a51d9dca3ee5749	\\x40a1c700697ae8707de2d63f37e25c02b21bf6685ed5419eb93014a67835e38a	\\xb339b8ec618f3c6bd69d880a031b0e31b1b5164645850900270826ea70d9c537f3d5d47b3a918cc6d14a852fc5c60c7e16c928a3b96c8b8f21cebb368b9b5bc221585661a265ff2690630462b31fa395b3a588882f463113a123f18092d856b0	\\xa6e1d02534af62c0be68627caf0ffff5f7e9eaf79e39f5ddebccb493d804fad478022993f3e65ea39baef719adc14fd10afc5c9eb47c4f36ef204b9a72cddcd3db09458e70f0f1f25997fceededf9243063ec468b767ab5283d65be3dd26ab97	\\x706f61706734654d372f637752692f5a68615a67307a7036623941364a6c6341	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	90	0	0	14101	1
0	29	\\x7af6146507bdb26b828c94ec233be37aa7549ff23a684e53543a4a062d5a7dce	\\x5ad74d9ddf0ab96ad9eb69a5e09e70ee15ea488ab220d98f2ccbac615d0de75c	\\x91e11375dd3168c0a84ae9bf3c12d22e337b01e1b55fef5d3190c8fc8d97b933	\\x847b83b50a3a22372415cc3c8aeed55b0f536158bcac1fcabb28995c31e0aa8bebe17b2a9f663b9cdaf81eb2180dfbb70adcdf20bed5b90b578c0a685a0411e5af579fdbedcdeea0b82461a5eeec3c6c1873223591b3f51dfd9df12282b8afa6	\\x89b7b95ee56b2ee43759af2d8a240e93fb4ee8f0d5ae077a56abd5b586357b54d2e56e28bd25a2b6d822db6f49642a3b0035aa72d426050773483965f4b02f9d42c721091991713337455e13a12043dc19a078c54a22d60dad7a70708518dcc4	\\x74656b752f76302e31322e332d6465762d353033363532633900000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	38	0	0	6427	1
0	6	\\xd346e6e92200d0ac2f9d5d946a79791b0c3621baaf6e5b20e2772bc7bdb410cb	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	\\x7c1057dd809a1172282ce9b1396687e6e5bab901ddeb1a21463cb50aad126569	\\xb560a6e3c334a216d480ce0d52395f37071c882c896723238c8d4a6925474aaeff4068a8084977bbb64b757be305b936024d9243bba9f8bd0ca1ae909c0389e58dec8047bed502aa641288b74f28b2f3bc97a15f7af6955a143d31f946fc599c	\\x8460c3cc8faf6971215ca09918d70cc43ad31b9e59e58074df5ab941abd6f935a2f9b9b1a367f96628960d7b119ade500693deab5dce76f9f110c7663069ebb5179bfde4e46396b85732e4316454e454d8d5e5ee5d611e7efab4d21b1dcd25d1	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	18	0	0	10383	1
0	27	\\x5ad74d9ddf0ab96ad9eb69a5e09e70ee15ea488ab220d98f2ccbac615d0de75c	\\x1ddf7c337072e2d8686da753c427cb6bc211887c29cdc0d4a15f6bdb4ff3c633	\\x6ef015c791e51f2b87268154f54a45099ac2645b40b6f8f8569365c547cc755c	\\x807337f6bcf6f02f48c29706a128de8a3951ac8f96dac7f02ed2199c201a7e78b1a56f67e554a128b99285e701b7ee790230e207cf787fd832e122e822565d18faed35d671f3bd439835b437327356939d59eefdb23f9923ba4db06190a81113	\\xb607da27ad82ca9e20164c26aade425d2d687c038e7b164ebd6d86b11aa2807d6b7be4983407edacab59eefd4a0f42b80fad64ddb959c2449b0ebe65da35727ee77591b19e7fbd273e6a785e2902239d16bdad5cabda18b251ac462b027f83ea	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	16	0	0	11630	1
1	59	\\x56881cf025a4702f22c213c9353aea9529b181240bd36908f421a3f74910a5f8	\\xee6f8a25204af5db2f051ea50d1332a91bda7dc7a72c1b38336846b7a14418fc	\\xb75ea62add2359fd5a2eda55b19a91cde9f05d3fb9719222fbbd98c7d79aaa8e	\\x949b8c0f855bbc13b92061b52c06001a12abc99902c04bfcc416a1ca333dc93313902d9091cda83a0868734e292c7dae07779dd032f9571c0f5d952e753694e9d595fb9471da191835b65219f4a64e7890f8c064018923f13b720b43b33ebb28	\\xb28a3f73a7fb2e24047758618828a99c7514f74bb9d238568331b55bafb23ad9fa02270d96d155e45a4fcb095665949a02d85e9bbbf70af9334f4bedf05d17768c116c20f2e055a983c5ef99c884900754e5b95798bb15027fceb51117a69433	\\x53746566616e2333393137000000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	69	0	0	2995	1
1	61	\\x7678a673904492262b8a74f7b5cb32e0b692853f0178097885cc8f3a4503c12b	\\x7413cf9dd7b9fc319f1d25e97554ed9aec9b5953410b77236c43db11476d86cf	\\x7077dd3e53889fb8504cc1acf9dc98039eea744eae3d9a2b53a36e042eb077f8	\\x8dda0adb604ec43ae3d9b301df47f4548768fbd94546fa274f066edbfe966293908405d853bdda430a9da6f0573dfe9015cf1e6686f05941ea69d5b0a8b862bdf65c32a3ea3476699ec71f99a93e0459027575e8480a6b80a1d89cf518de5409	\\xb8ff025715b2ce5e45491385331f8159ffcf9f54fb6c0bd6985de1119ed2d42ef58a7f4a7ee6f4b26bee98dd7e873b780ea33ca3cc207dcd5934dede685aba2bdd2070bbedb27a9fe03600d484b7f644f172325fd0f6821f8806101b0f31756a	\\x4c69676874686f7573652f76302e322e302f6632366164633061000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	8	0	0	1985	1
1	34	\\x0c236734d76c2e641a87bb512f92005783f244c0b01041906081795a50325917	\\xab5f7d1ed8a7864b0152d6676225b4fb5eab0e10c0ed9f6cd13a20933242ed4e	\\xe0d78124e0d0db9fe1c31c52c30ac9e4bd5faa643ccf5975ba4565fbdd2a600c	\\xa9f89155482ea5de7ac3091b01f9890f7ae2d0c6c5c6f8b4bcab4f8c4e0b88d5dec65f1737b4720506d0952d2d0cbe581085ff44e477b9f89effdd0b58d42592185a9f28108bca1183f1b93b54927d782cbc2a6afd86e0716ee6142f5fd9c48f	\\x8f03b66349a08e9134e67465ba6b122cebf66503223099a1fa1243f9b7bd6661452a4f0f9e3e3dba6582d2d66f90c79512783181dbab4ecd552c79507e8b0ff62c19035e8891083bb002389745a8b901a1c95dd12627dfcb2401dd277de8b99d	\\x76616c696461746f722d3535353834393835642d376876346a00000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	104	0	0	4013	1
0	17	\\x7431436f5226a17a528b2aa72b5787a62845a58ebd2b34441f9f556c64aa8f9a	\\xead8ca2574c42afcb767d858f4b7f2ffbefcca09ee44619f42c60c779c0430e2	\\xc7c0e3e67017844e26af53bbbdb282972c36efb5151f7cbd47a530a8c1b11e41	\\xa84e30f452069baf7d2855162e0eadf92dcce0dd5e2b90aeafef8102686c8d9f2ff6187f0bd02d61a5dd01f469b70b9f1149f4a0962c54ea84432c54edebd28df533e847ea5ae05db49c47f0a2f9f161022f3873dd9b7c1d983f93ee8ca1ceb8	\\x8d45346bc80f56fce763a3295713a6259c0caa395ce8daeb465d73edc83ea8edb922d94385559f2743a1eec60cc8c31c09a5b7bbceda1b906905c063ba180cbc3bbe8ac25e3cec5441cec5bd326f1552075e6aebd23de2d326d59d4615c7ec7e	\\x4d6574616c20416c62657274202d2068747470733a2f2f657468322e6e657773	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	18	0	0	5184	1
0	3	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	\\x817c1d7bf2f42faba7a38eee0045791b4ea11052bfbf1ff325fc06333e0bab80	\\xabbd14e47ce75be671bfeb1158da75ffe8563cdad57453e7647a7efde21725d9ad7bf5fc878036aaea659648c90969f205c73574a0123204c297ed2ee02f8d13a722458f29e7942cd6a51928bde801857681f01d4cf514bd87a236dae1b0e63d	\\xb3e4fbe3b3252fb56cb69903d71a3b7e923b55da4e9b2ee556a1ed969bce5bd8c502d32785e6ddf94c7556b9f84002e708e54082c85012564c4a6e5277c5b4d6370f958bf2789f87acbb5f3313877ef74835d6d67f92b781b8e5109e372a064f	\\x706f61706734654d372f637752692f5a68615a67307a7036623941364a6c6341	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	26	0	0	4488	1
1	56	\\xda108765c0db70b5909ce3c6a3c49458439540f988920b293508ea732c96d740	\\x06f9c4b8e316eed1a5c7968850ab82f0977774eb7616d8ff0a8394b0e1f0944f	\\xbb76bf809c7909a074e506753356a5aa0c3b9574f3f975c0e5bb00d338fc1c97	\\xaa7a15f596b8b3a5968bcf132060f67b906c495fcbdb59926233d12ed37a23cd025f6a19d12959102143339f1d1d16e41712b1c24b70370f66cde80c2afd0861bc476fbc6ff19fdd26bfcf1a539968c9f939b0c80e9be94ae06cc5c894f85057	\\x805fc151263106e292ad64737f44ac470837e4571adc57a351ee82b5bfd00ea479c3d4c755e7d2d8ffc9ccec2a55e79105e88515492d7a8d3ebb7eaddacfd64a6b41e0495cc9b77b8134205707f7892632b8350cf51275a26b108a3a17f71e01	\\x706f617034597845546f4c647245424770564e5632584a4237466e4847457742	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	6	0	0	5094	1
1	42	\\x699f8f631f40d3d11b33a8fb331304f613fcb270fcdbe41d1a51d9dca3ee5749	\\x523069b2e16e78b78d4725b49cad590d63546777ce5d80d646c42a7b5c560cfd	\\xa3b95bcef68de635a0d9e95548f9f1443d1a4c55d37919e029c497ac2ae8f59c	\\x84cf97256fb355665864efd9d571eb6bf90fc2828dabe75c4cdbb006d61409c69631f3694defc40df92b4d17b09a6e0404edafd3e5cff7cd6d55cbfef3666a2739632c98ce6a98d8e73fb7e7480773487743e7e742fb4fc1249bba8706c7fbf9	\\x8972642f666fbad9dd72bd5aecc6fe6eaf889e4af825be94ef55fc536a653df87651763d01aeb165ce684b915cc6050b196741bb8678374fc41ce3eee803d4dc1c56a9ed59ed51574c2b14b951a151321f1ac32fb4f1c0b0752d198575b6a93b	\\x5032502e4f5247202d205032502056616c696461746f72000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	128	0	0	19848	1
1	48	\\xe5ca07938d0b2f4a1ef5f49fe0ce4b6a4f2449f3ee9fe17e98bce6d4b523cd74	\\x902bff5de4975ce1066208a57a1f6576a26d2c97f37732e90f16ec18d34114bf	\\x0c2771d0bc9b7f90a42e79a35ad65de86896160c9a4a51ee4eb8588ab6d24192	\\xb97de0e21a86eb2935bc52cada81ca4f38ad82c0cb7cf2493d712e6fd332a86cc26f37fa6dfa968c368c890472d26a2806b7e0e20d72933e6ce6e076cf61541c9e6400002d23cbb9a898bcbcba1916514f3129f8850497422c8b710887ce9b0f	\\xb8a340441c11728a8d23e2e24e4daf58f68b050e6c4664bea661c857cb2bf9cb6a29508cf8fc0c822088a892571f448f0a3b0ce6ac314683c20e1af727ff61bf6a42e0e7c5ed71760786950facde68da6c6489d21f5b96b5f3954c745c6632e1	\\x706f617036442f37715267766e4e63434638394e5471453865672f4b766b3041	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	78	0	0	4378	1
1	49	\\x175ace11058ca5d1e9ddca252f9505cf706211db0936e5dd65ac65dccdabb041	\\xe5ca07938d0b2f4a1ef5f49fe0ce4b6a4f2449f3ee9fe17e98bce6d4b523cd74	\\xbbcd1e85a228cadb77785ddf009ca18fe882dddc94e631f66213482fa018c8f7	\\x9856d143e48e76d5d112c649e0c4fae2c53adb29d8338fa0927747dc3b11a57aabf1f197700e9f9e6847c90f04c5dbae062ef6d1c4948e365f6cc5bfa778d60c7965f9500aa3ad93f7b35c750e852d81fda1b2d9b2649f551c34ab4763d2bde5	\\x902419ddfd1634811c7a1bc44be009583953bde38886daf245cf8a5325a94cb2ebf10084432f65e1a23c1bcbc85346c6186594e67eedab15c31d74ecf79b1a320de63a8f52564036fd59c875c853a8ac5c5873bd058ed5055fb6f550f02218d2	\\x74656b752f76302e31322e332d6465762d353033363532633900000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	27	0	0	6123	1
1	53	\\x07d6f3ebae0bacd67a58939fb15b1fcb014b2239f76f78a608ba46faee7e2aa9	\\x06f23ccf329c74310e48d05583da06f41832b15432e721dead5505a1f9e99812	\\xd9bb620d1b5fffe0d67c4f1b3d0f096ffee3e6abdecdadff9f866a0c87deac01	\\xa01adfe8fedef1e3c96e23dc8a65a7db340eacbc3d69d2474c45e16edbe9d2e68463b62de6510e0372e5f1ee54663c050f3914b450eeb39cbd5df13dcc9f092b806c6cbcac4b3869dcf6f41209d345d1d66c0eda60e1e4bfcf2fb513cb6a1e0c	\\x8bf8962ac3b7d480911f6daa34a095672d0137f6b862f2cfd19e14d2911eed7d9f5cb4f4dcce8ad5072e24e52bdac9c201cc04653cae5c5eda1fca8ab4bcf169a59c1a13a28de699fa7200bf69b9bf96e323b0b5f509b0ae0b1c58709049d9f8	\\x677261666669746977616c6c3a35303a3235303a236666303030300000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	80	0	0	15298	1
1	57	\\xaa46d5da76fd8f00e993d81d7d64a5ddc48247930708f76e07e51905cd712833	\\xda108765c0db70b5909ce3c6a3c49458439540f988920b293508ea732c96d740	\\xdb8d4c4b215be04f420d0413e3d7461becfb385dec1c14c10d490a0188808fbd	\\xa1ade33de070158413562ee2a486f55114eafe1dc549ab7f4b310d39ea112952e28aa80efd19ad613a4906d36d1355c519b27faebd4f59ab4bcb63610dec02e52fa8f6fee0e97252a59e6bed91bc5ad936bc43a1d4c9aa7f8f1fc4c1d76d5153	\\x9716b353288b64864428ec41c1fb1b1aafcaefac2fd7e80533f79c6b07c861a68d3d2752c57fa9137bceb5d386b3e1c5017fc10807b0b42e64ff458f2d45509b781a7a61d0047107de3124fdc703bb46dc1ecbe863638175cdb3345ffd7a7bf1	\\x706f617079794d417a326c2b516c704a59766d56754c704242697a30796a7342	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	6	0	0	13599	1
0	22	\\xf2f0c67839f85a6dfe6b764daf49611623c19eae69b7862148ada86bd8f671d5	\\xe6156915ea96adc1fc1d1248bfb4ecb6cb4b3ee06b3898e16add50e950f9b1f7	\\xfaa17675d24c6bb61ba430001c88696085e128d609df9d93b073a899bc357d38	\\xa1d406f74d3379f0a123a29f1ab377d454d85ff45eb76f4ee1f1fd13889efdb37c6fe2b513670854e6bfedc32004a9010ce8db03025c412f63cfab7779ab1b4eb8b08546ff42fa3e9ed70e85a29fb8b76a7ab3fa6ed374bedca5dfad288240e9	\\xa40784f5e4f46b73d177e6428f8772d07712eeedcdc1933b51bb064fd992af1537398f6deb4d76c2af2ce9d350befb5804b1bb89c1e6635dd2cdff7b2be5b3708c01c2c8cc996596fc9b1611f56ac02deb390eda85a85e08e83f5f25e118ee6c	\\x4070726f746f6c616d6264610000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	18	0	0	18618	1
0	1	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	\\x88caa8d6e2cd6b56d688e1cd5739c86559409a77557a9cda117c500ebe7ede0d	\\x8e8ab1febbb4a6732b10fa7eadb2347b6c88ffdac557a1b71c755719000dbe346e8bee3ea37b013c89e134beeb927ae90db22e66995a281380b5c7bc1b4b6fa5cf7a97a5d7a234083fb697a2946ff569c65775c6677da09a79757c738edb7928	\\xae31d21e3bdc77712576ddabae60f51e3592de21678e473d488e42a2f6b55c7a9b1e4fed509e88440dd07dba57a8d0f400b2234dc5f764abbd47001291304fa7c8f2d166e42d0a299e35e2ca8f4d39b615fdf7c40f6b11b772f8984b07a4d18a	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	10	0	0	9804	1
1	62	\\xf24823400b02125072f27a2c01f796fed1c6de3069175069c27053f202213e0c	\\x7678a673904492262b8a74f7b5cb32e0b692853f0178097885cc8f3a4503c12b	\\x14bb5631d9e974691870fd1a6877afb109f69d5a3bb5335c3f5b9b813deffbc8	\\xa507027a1d403a0af11c022b3da428f6462a63931e9b3f864a8925be001f735bffb723dcba3545b1b31fcfb4b1a71cf7084db2c49fa6718fd163ebba62e465afce098e4a34996100f5dec2d786c5801e611e8dceec820d771221af041c53f706	\\x81374e6fa0d0035a3c7564c906728ac760e21f0afae24f6194e83c2e9cd1cc30e76058daee1138d0ed3c21be561a2f7a020e5f95ea6a151ff96c701c586652521aa9110c2f6fb663100134423adaf73d01a1956c690329e41a2e96af7e298747	\\x706f617046575664766d6168646a2b564b6e334c713567704f30644644516f41	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	86	0	0	12433	1
1	63	\\x5c4d344e5a679f8ddda0f9b9ccf0cc5d467aefa58ddf695278c543d965a2bde1	\\xf24823400b02125072f27a2c01f796fed1c6de3069175069c27053f202213e0c	\\xd4773be54cf4a34bb05bcb9965b208abe5ab0a91e6ea9a220e15e33989df1f68	\\xb25dcaa1ee4a3448e2f3d85d9eb3258d86f51abf7cb9b5362ddc16e4863c941a8c02f43ec56241f46c4c39c16b9600760b558f6107c514830573e9223ed5ac7a9e136035b306f3b0a11bc270a9023b1a7e32373e6482447c03dff171ef68094e	\\x82eb2ae20f897f6283156ee59ac03417b92cb2490ee7f76ffde40be340cf18c73096b08ce15457028b331817689aab340bb474ce03c6e509f5ed9933ec39605fd96b66320755b2e353f4f4b7d0f49b21ba3390634fc206b1311e1ba2ba618bb2	\\x74656b752f76302e31322e332d6465762d353033363532633900000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	23	0	0	5563	1
0	10	\\x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf	\\x6f88add8cb13638d6e2017143c7999c4efdf68008a039e6d3c6b20245190ebbd	\\x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9	\\x8ddf5b2ce68b8e3522678a7068a204a66412b096feaed4379d142b17cb3fe1109af8364641e4b66aad5ea567c891271c04a23dc2397194a01b1d4b5ddcbbe8a12b5094d7a3502f64e23fe9788f533c0fba791009644c35848916445c6bce8753	\\x95583ecd0ef2e1d047f13a098e392ab2bf3983e9bda5ba1802998ebf24f72cedc8399985d986265d21ef6830af9168ab0b452339e199572fb9ec748506f7add4278167aa3d096ca6830e8e58c7ee91b5d7afdb916c9dbe8bf029a0ec00328081	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	16	0	0	8309	1
1	33	\\xab5f7d1ed8a7864b0152d6676225b4fb5eab0e10c0ed9f6cd13a20933242ed4e	\\xe21c3247b62fb3585889e054c16b622bb997f27ac070929924f0176c110bbbe2	\\xa38509e9ea59d3030d92767a1e26103d35a14a6c4dbe58dd988ca645009925bc	\\x8999b7f7da235ae078841586b981c5f85ba9423c1ded720944e27615c37f947c024764e5822f7349c0e24c7f9e3943e503be47a482f41e1822e80a238cd276ba896b2ca47047892a8fc25cb0654692ee6e87f6e65bae08e5146584060bdea546	\\xa237741cd0a8eb0112e9a59ee82099e1990c1f5ab2e5c000550476969591298db54c5191753fa15eced1770507e800d9033a35595260e64584c73b9af941f7247015d73ee06305b57c8d7098c2c17233a3849f561300f0885b4f16cf2bcf8ac8	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	67	0	0	10313	1
0	25	\\x787c7110d1e7db50ce2d09bf6b150211d1240e4d0d74695835b1c548d37011c6	\\x7f8ce6d918d6a40acfacdbaa2120f2697b5248e5bc4a24e3fafb7d3ba559040b	\\x2f5411d1de9511f3488fa7a9ee4d80f6e816b28454abec2b439b6bb5854245e1	\\x935a09cba0fd83488ba0c656c2ed6c4bb972c9c0bef0e25d677955a240fa95e701f96d874105846e089eb95cde65825b026bfb065cbe40db45d501f0f8cf4d90c8ede49b26e013c1b79b18f4e8ab7b8511e14d0be8af6d972b4c2dbc438d87b5	\\xb8b3a0c31cf4e914f9603e3bfd76318dfdc8a05c34ac06a843fddc4fae0a8f6b849900da63b6edb7e0dd4080936229aa0f309144d355da365f5b67570a145a9973035dced4932869d691bd7e6ec2306ad440021f9d646d80e7426489e582ee1d	\\x74656b752f76302e31322e332d6465762d353033363532633900000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	28	0	0	5809	1
0	4	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	\\x9859516a62d97be70647491ae1a7d856d1f31227d4c797776b0fe3d79001d3a8	\\xb841963efa222b953e3d0db8ec73a5e4f48f76faebbf0a70e5795a7f7cb144d7b1c015dcd1de344b73a0208f5c90df620f728c273d9a52232056b078f9d67ea22c7fa28570ca677ef764fb404afa1954ff67eac0f429f99f7debd909846907e4	\\x8505abaefc568507f1dfeebc6a496ba68da32a0ad40d5e19e6387e8f9ec80b77772f73f5f064faea7a533b1714975ec70e84e428cfdccbac8eb1c54328998f55f9469bccec11e059470380d9ea69d3cf93ede277b64c800c772b9afa6ca6d407	\\x4c69676874686f7573652f76302e322e302f6632366164633061000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	8	0	0	1072	1
0	11	\\x97b93933a0b644c477b9e2fc52c0fdd41d14d74d2efb73aa2e9df3d1ffaef54c	\\x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf	\\xf6cf568ea809cd3baef6cf1443d7d9dfced4318abc24bc0f47085ddae78819e8	\\xa8bd637ffb1db0a562217c710c00b2fee8152a825b236f8f6b8bd9ccac9c0abc8f460412afcd4f3b0a1c25a2b7a3a27f03985c68eadfe0aeec0eaaf0d1ba7d49a5f018bc1cfc772ffaa3332fd819a04257ba2addf19b22f0494624065197b902	\\x85270761bffb2b5b6f4b1921c5d643c708511682c720a2eecdd98c5374a98f2c01685950d64e529abeb3afe69000fab301e6bfb8a3f1d73cc51994a0af4bc51320e83814211a249e7a8b2b5ea9faa38114e02b9756e5e4a2264ccaa735dfbd0d	\\x5032502e4f5247202d205032502056616c696461746f72000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	64	0	0	17836	1
0	19	\\xe6156915ea96adc1fc1d1248bfb4ecb6cb4b3ee06b3898e16add50e950f9b1f7	\\x7431436f5226a17a528b2aa72b5787a62845a58ebd2b34441f9f556c64aa8f9a	\\xf326ba42da7b958af2d49d47eb8c9560605be59d99f76eeccde2ba0ef208531e	\\xb82007b86e68546ba637905e6cefb9592bf46ef2e1d2c9813038225f2d9775455f15e4322f852a73291d61e242c583c916d8e9e6d67224245d180cf5bf3c779b3d25c542c3123ac392e93be79d6fd8415ae47fb1581bc9e17a87c8d7bce1bbf0	\\xaed8e9159d7b4009350a989091f14e1cbc69a9d12d287fa5550fa38d3a200998da6b13691026bbfa8adf4693c6abf0790132eeebb5fdc5400e189ca55aa882460c79a0baa46c0a1d0f4576bdc8829744d4813741cb43bf3dbc753d5c9041859f	\\x706f61704d6c77744746547337734537324b622b757557454277317658686342	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	17	0	0	14521	1
1	52	\\x06f23ccf329c74310e48d05583da06f41832b15432e721dead5505a1f9e99812	\\x175ace11058ca5d1e9ddca252f9505cf706211db0936e5dd65ac65dccdabb041	\\xdbcb2b8d8dc88d4ed48d5ba587398ab76ec4b43c2749ff02f4f89c4ad10df04f	\\x8b56792a5f1fa3a675080483057aeb68015db16464e405e2e964f23364a1e33f90f562b427b1f6eb87952f2076de160c03303bc9cc848a3694bcf2129bc8deea2ce197fdabbd8b6cbc0d4621b1105b12bd8d46d909f1d296778e9101beabe2ac	\\x964df2f96cd78bb58cd423111784adf680f7d6ae145e398d762d69fa92299d06230d33224de3b34bb63e31f4461057af01ab5089f662990cce2ce4aad0660417bee2edc33a3a43a1f55dd6dc18854a2124c6a1f9566fcbd82edba0f64ad18e80	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	110	0	0	6905	1
1	38	\\x523069b2e16e78b78d4725b49cad590d63546777ce5d80d646c42a7b5c560cfd	\\x43df93c3e43815efadee5a331028236db1bc8f8d3300034f0889e9071118c591	\\x99c359ae73dd315ea40e59041418f1a0c10a31c9dcfbab432dd64e2ae3a0a0b2	\\x875d9e9d8fa7af5cbcd3364a33b0f68020ff134a94dce309eaf3fc117afa9d6f3ab1bf2968d77890a6077daa073100471734eaf6e9bdde08a673abc1ea999ceba310446dfedb0295b5e4f17857914ad089d68e4e8cfbeac884f928ef75dc44c2	\\xb7d470199785f0d9197d1d0ef1f678267506ea862e8012249d1faa441d89e805952b19c8753dedd4519bb7082985e463044f554b2f5082edfdff8c5059bd44273bc5a6e1f4ba1b45bc48175c9932fb272b2f34b59e0c9647ce39ab83a0e1ed48	\\x74656b752f76302e31322e332d6465762d353033363532633900000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	26	0	0	5791	1
1	58	\\xee6f8a25204af5db2f051ea50d1332a91bda7dc7a72c1b38336846b7a14418fc	\\xaa46d5da76fd8f00e993d81d7d64a5ddc48247930708f76e07e51905cd712833	\\xad5bceaf19c9deebf839eab4f2f3958577d4540d366fdc537dfb29b915064ab3	\\xb1bc4d4be3fa6a21cf2e4335a25d88bc33e93c403d79647ebf7b20c4d8329bb1f62e215ea6661f7029918179a9c12a0b02b341d908e83db995177ed13c2d63034e7cad5361ef525e984b0cb7a7475962afba7b1a3aef8f7a85bdfa06e5b9f5ce	\\x9603ad48344901ad04534f7eeeb64465cfec128b92a0e3a8bafb29f005b4feea6f820579e0ad0d7260fbf2f5511d67250a374b310f3114472d13c376d20ff711fab6d4555d52ef06ee5a2167e0023947e107a162dc90395ceb3f7eda642e7b4d	\\xf09f8d9c00000000000000000000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	93	0	0	4778	1
1	46	\\x902bff5de4975ce1066208a57a1f6576a26d2c97f37732e90f16ec18d34114bf	\\x23b11fbb75fe343946e69b3d1de30c3a3b8e297806c5583a7e389de94532f43c	\\xd5c023d21316f030278c9caf428d97f348ab494f786a56341eb0cb507be66b51	\\xb0d375c8caacd0fdf1dd1b239b99da07880db8a32ad35a5f2a1c13a73700647ee2d89bb978ba529131eb703cb763eb6e0091d59b98e3fa5c6dbcbd589d0c4464422a0137a1d3b393d3bbe693c0074c7b30f1eea8e3c7cf88aa1a71883d0dc964	\\xa8acc6cc409a49d8a8978ef0db573af0c01b3107afd1f17e6a9d03c3b71d9e3018a497bb3667a075f77257492e6c485d091475cc00f95f456da88c0742a4c5a35000aba82e29692669cda481f19b5273f7f434c98c31d4b8b3d6e5db183f7dba	\\xf09f8d9c00000000000000000000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	119	0	0	12292	1
0	12	\\x8f181722abe3a0383815ec5b71acef8e93b3002508e5d298686719f3c461371a	\\x97b93933a0b644c477b9e2fc52c0fdd41d14d74d2efb73aa2e9df3d1ffaef54c	\\x584c118474f40426649f70be4571901f09d0cdd5f2188e9393e6589ebac0fe5a	\\xa569db5e438d92c9d84a9a86e723860f42f737753d372a6bd83feb14641f2a21e29f075e29b7d6bebbfeeb80286050590ab566ca3557f9f8ba9d272d827e51df1f8a228f55a959efa4b4480708a79f37169646c48f5cc2a592db6a272af80aef	\\xb83340f0cd3d2af8a0fd870ceea21884769d944244a543a489298c18332229844c2ee4b44a07cd48ee46d4d3d795230601f4408bdf3a4b1c52068011efa8e9fe056ffc858f3d608b74df8b4a25f44ddb2e6234e706f1eb5cd7ed07adb63b697c	\\x53746566616e2333393137000000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	62	0	0	2585	1
0	5	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	\\x05c8203ac417c19cf8a8d307ce1e2624a42019c3f9fab66f51aca765b28fe4f7	\\xaa0b49e3ef4872d7a2d28a09ec55be7d06dd9c75dfffa225257e32d07f5a86e94de8a0d4a571aecfcad4588b31f22919100804a9a839f31b2b6ce14ed2de16b585f75677d87235894ec0467d0cc2e75079ba63d768d933f93d2960b6ddb93059	\\xb02dd7fced1a851c11136ac64cc1b5afcfda24a736a671c1076601a617c6dd7ed0efb4bf2b159123e7694a2e37552b4e0dbaba921e580acadd0fd230859ef570ed3f6746d78f7373ffb992051ef43063148fe2fc44a5cb9c45b8387e22dfb03e	\\x706f617041364c426b46494d546c343376526a4d5a37563938713633366b3041	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	30	0	0	15248	1
0	26	\\x1ddf7c337072e2d8686da753c427cb6bc211887c29cdc0d4a15f6bdb4ff3c633	\\x787c7110d1e7db50ce2d09bf6b150211d1240e4d0d74695835b1c548d37011c6	\\x65047bd0dbfff78d3c6389137ec159d2508bc314dedabb3086835bbca48dccb9	\\x8e54d8478ead05a0e3fc4a314212445f42f53ccdd88b1c4f9864870b9f80625fb3429b911749ddd464326765ad85a38609fb4d02333624ebcb95acc83bbd59b244bd364c2abbe97a5f56dba725e1b6d2e818678c43483bf280325d2a89850235	\\x80153317a2c1c14673aa8d4f3188d55978e5556c2b24da234c924aa92992aef077ddbc978a4beea81b1580abe7c1854e0ef7e92839f8f44821e360306b49637dfa3b7e4b19ddf82388fbb2b62e23d8c093756a07c6e9b5d904c5dc2be893f5bd	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	65	0	0	6862	1
0	16	\\xead8ca2574c42afcb767d858f4b7f2ffbefcca09ee44619f42c60c779c0430e2	\\x05deee7df488fa4c2d596f0d457dbb1351936c20a9aecd85795ce57e097e5172	\\x66f3115487c2e0e48866f567132d86beaefc2dc9c1b0634851f7abe9cc33427a	\\x8e645853b42d3a11fc123f652a48a3a47b357158b1e109be6dc911933ee1ceeee4b3c5c262f8387bb14f78445161c55304e8ebc87bf3a88cccb670ea25f64138d3d52e41469407fe727f0baaff526b776dd02765866f8d94133b68749cb42f15	\\xad31d4e9834d187ca0ac3e57e7c8bacd74bde279e43276987431f5c9c9f61c773dc0228232976e5a0ade47bb9423191415212d1d6f6de3c7c62340dac67b8807d601c7a07c9b6fac035ada19bbd8bb12abecdd49bd131107c24eb781287bac00	\\x2020ce9e204e4c5020ce9e202000000000000000000000000000000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\x560f09b440e530d7fee7ed01c5449fe318de36621234e4fc8dcc49a48332af19	0	0	24	0	0	11621	1
0	9	\\x6f88add8cb13638d6e2017143c7999c4efdf68008a039e6d3c6b20245190ebbd	\\x76262e91970d375a19bfe8a867288d7b9cde43c8635f598d93d39d041706fc76	\\x6a86d90f5524d1f40fcc2d14e341363a5c7e1faf67497fde79049521fe164fb4	\\xb54fb2450d81bc04e5ec410c9501c44efcaceba42726d0731fa6d7d458b2a75bac0c152747a7da4ec7466187e4eb3459174106413cf38a5cbc1970a08c0ce4d54ca1876bab7ddf065366bfe7c99637683a3ed858ab4c8979bb2bb8cf7112211a	\\x9360c81057175badf8a1875eaa61ab0bbd66957aecfe7b34fd3aeb0ed2b30325c8407e2fbe44bcb5090e7dbf8b058eea0318b9cb6c3be549daebfd1db0a4187731e33c689ab0f7f9cb0da8819fefc1d932f4773c392713d39d878f55664863d2	\\x76616c696461746f722d3535353834393835642d376e77686a00000000000000	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	63	0	0	4067	1
1	60	\\x7413cf9dd7b9fc319f1d25e97554ed9aec9b5953410b77236c43db11476d86cf	\\x56881cf025a4702f22c213c9353aea9529b181240bd36908f421a3f74910a5f8	\\x922d9ff0e816d51e550bdaffc3bcddd00bfab2ff1a7cb5d118eb49ecd79aacf3	\\x888791373333a40bc1e2638ef39233dd8494a45c1292034ce39406850f3d516559c7ca3bed14381a06e59a4e7ae90f2c185ea85d62ca06f8e594608c32636756df1d1eca43bea8276c436b7184cefa6536d0fc5529a5f146a2c194e9ad39ff6f	\\xa1619dcce9cca74cc01ba8dc4d7598e9b4f87fce71cd9db49674ee38efce050aaf648bd2d79a2efa0f196d52f96a52820cfe7c0eb6547f8f4c239f23961b1a6ad243ca02c9812309617802f47d1a4375b14937479ff6fcea7cdeacc622cd341c	\\x706f61707368777a33682b72503646556d6359725766344d77795541494e4543	\N	\\x53d90f778f975dcca3f30e072b5c1a85cfd7a1b977b78620d94f143d06432f9b	22637	\\xe0c057333355956e8fb8d88382f5676bbe083fbf8b978f0db719b4d02ae70777	0	0	18	0	0	17237	1
0	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	\\x0000000000000000000000000000000000000000000000000000000000000000	\\xaf6aafa94dcc22a5cf0253c4f5dd886397900034bf7c149c54635b537f34f64e	\\x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000	\\x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000	\\x0000000000000000000000000000000000000000000000000000000000000000	\N	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	0	0	20084	0	0	1
\.
COPY public.blocks_attestations FROM stdin;
1	0	\\x0a43311a854130be3346946611e04ae9cdf98914	{2763,2256,11588,3679,7119,9906,6905,10792,4575,14524,9855,3487,13302,4684,6901,14864,12088,9945,6712,11023,17887,4884,18562,5446,6943,4018,19477,15242,12318,19280,9306,14835,19882,14199,10170,10261,3773,10035,4566,12217,7014,10047,17208,17941,14074,8346,2263,17948,10251,2506,2907,11222,6828,3910,9911,17593,15754,10306,10075,14096,17064,2223,5256,7080,9280,13296,2293}	\\x884da60b9ad4e1f9df55383959649b2823d36fa0662f5112942b310e7c1812c9194a065c8ac36df25bf86b3a087ae316011a8ca2eff8dd21b47cec14cbf2f293465fd2ba915d7d45e661fa0838a763af6133830d2a04c244a3c159225f9ee0bd	0	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	1	\\x31411ce602c00ccf51c0e1a0160b20ae41c793	{2070,10370,11439,19300,15116,11514,19895,3443,17950,16525,19832,11258,7124,2875,3045,17830,11335,2520,2731,17795,9907,3303,2117,10605,14689,2952,13968,17695,6866,17241,9228,4733,19211,2353,14822,19181,17118,11579,10845,3731,17944,11609,4556,11309,2822,4713,4097,2360,10456,5354,3895,3962,13825,14893,4053,15196,10793,17753}	\\x8f3e1da737b3202b0f8168b4ba7565445a66085d0cbcbf6a5282e0cd2bc89c6f30b1cf8f13dc3f217a55fba88671fb7d0e684becdd5c03ae1112de99a59d7a5c555382aed191327772a03a8b21478b97d6406a1ec044de02f2de4cbe3377adab	0	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	2	\\xc40542b91e2436bee14807610810a22c0b572515	{19012,2994,3875,9963,8696,7033,11148,7006,13970,12052,2895,18565,9389,10645,4735,2764,16725,16482,2928,3212,4712,7084,13944,12499,14108,15919,12162,9893,9993,4042,19287,10511,2669,13892,8328,6842,2328,10822,2113,9724,14231,11593,3005,17639,14828,12559,19835,3990,2342,9599,12233,10184,19149,8670,18571,2579,19723,2437,5295,9872,5450,10824}	\\xb50bbd326de7486d96cba35005f97e646b1e2bc7b2e526e0c63424de3fff39716df3b52587ad5e17f8dd728048cadd0c08e89b947b16f5dffc978dfb75f065c93aa4b592c5773c784e1bba75bbabb13c725ccd1c7a59d1578586f6002dd489d8	0	2	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	3	\\x4f8e0043081102fba666501b248075acc38c2a19	{9172,11101,14268,14207,15941,3072,9873,12303,4737,3486,12051,17912,4487,11154,3855,3149,17803,15052,16851,5296,10465,6652,19715,17932,10106,16006,14243,14109,9841,10439,10552,6871,8336,3379,9988,2978,3083,7050,5146,3345,11558,15356,2505,17895,14811,2467,10027,4023,3153,14846,14332,17282,2934,5209,17866,17140,14061,9718,3669,6841,6825,14170}	\\xb3f009b05b7b4a094efc485bf8e26604eac04d69beacb0778fd5e1cf295a21cb3b5faa77d2e429b138fb4461f94b1be00bba29d717e2ec0264411404240d385c61737b604f4b752a925b807ea376d06e74a8412d9a13e8c25ad55e4af9480b06	0	3	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	4	\\x0000000040	{18136}	\\xb8d0b9ceb3ec1963cdc79ac49ed3ee324fa7828bef6518ac0eea47fd7eca3eacde2cd2459d64ea8bc66560a92fc901ac02ec8177997e9b3cd173d3aa7395982ec96dd6e46883c0aa3374e64ead1aeea94a8d3674c7b5e20a63fbf76962f26440	0	2	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	5	\\x0000000000000000000002	{9066}	\\x99d0267617e51b2fe39312da927f64a3036952fec1f8780591b0789be4f0b35f4b52d8223318b4c73409e4d37ad31c4e0180cefba9fae6b09bb875c87ff6d301a631dcbd720662da5af6fe0c3e3e629ee0c186ae2d418791f35c9f6093c30a1d	0	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	6	\\x000000000000000000000000000004	{9448}	\\xa06ab232f87c54bff74474a9e562e0fee3863f57a88feabc4cefb8cad4e618e65b245f431ef8224eabcc907ef35b50a9043d2c3c0c78f9b1f8599bebaf7faf3a912ac79b562dd9b4128dce100f07d5333098400a341f580b95b067f6bb8fe804	0	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	7	\\x00000000000000000008	{9007}	\\x841909177379458cb85022fada1e605778be5a11c26df871a4eb2c73ced0b717e43350b0f700776672da1b9dcf2e7f6e0294c46afd956e1f74b3da365fc2abcff212ae432932ef5376c86bf53ffc79d2fd4f65480a33b2a6288df5e48d9c112d	0	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	8	\\x000000000000000000000004	{19639}	\\xa07f6ea07a94d8dbb51559ed66782b7cfbf54905fe3064fe5354b595c54e3e68d0f9855db570ec515e42b3b6b52b667f00f4933656751074c1f5f7127e958d8547c44acc2b35747bd8c8c3010ce029a9fc5ce4db016b56830df2432129d8e733	0	3	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
1	9	\\x0000000000000000000000000000000080	{8683}	\\x8966cf4ba93daa4a82482a3b7bca75be27a129c1e405c6864206c61009c5291cae0c5ea491b4f2508648c858a31d70db01288280250e599d9593f94c34a22248ff190c09ddab71eb08de7d481dd46ec60208cd88e209240f2c436d8cbd78ab6c	0	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	0	\\x12002040011200114080201002086009000282	{5524,5841,8631,9129,6512,9249,8621,8399,15919,19287,6535,5741,8907,5856,6433,17639,5669,8912,19835,19149,5812,5172}	\\xb4852e651ac865f791557a759079e2ed6c9b5c4ff7e0c06332c37ef5224cc11b7386cd96036b8ff3a233b0716734d60300a8bafc12c2d5b19ca1cfdb37c54b3eb2d9227761e615c0a0980b628d465aa2be9bd1643325378221eaac2888db7b10	0	2	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	1	\\x0000210400c0240040802140120002410020	{5695,9184,5840,5843,8449,9183,9424,6060,5648,17121,5624,9296,9123,8842,8874,5182,8217,5967}	\\x890c2e924aa5a9e3a2d6bcc06fc040ce31f032b6bb66adc362ea7be2aaf0805950c6c0d5e5d4db33a15922b29e48ed9f152c8bff9caad05f6388af7284146f08a2e21cb7f3a9b0db2275d0b0610cf6358aa3e26d6a735d033f9b66ad1f4cbbfe	0	3	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	2	\\x31d71ce62ae01ccfd1d0e9efd60b20be71e793	{2070,10370,11439,19300,9014,9058,5951,15116,9236,11514,19895,3443,17950,16525,19832,11258,7124,2875,5721,8365,17257,3045,17830,11335,2520,5807,2731,17795,9907,3303,2117,10605,14689,2952,13968,9376,17243,17695,6866,17241,8404,9228,4733,19211,9489,5549,5126,17026,2353,5902,14822,19181,17118,11579,9377,17117,10845,3731,17944,11609,4556,11309,2822,6529,4713,4097,2360,8910,6174,10456,5354,3895,3962,5630,13825,14893,4053,15196,10793,17753}	\\xaac077d346be5759d21117f4563d0a3f7b6903255df2f6af6442a27ed5c85e4a340427c1bedcfd3411621aa7755389ef171479e121877d4830647f2fa56ccfac64a5d4e632dee8664781fae456dd83939e15ccd4c40364b67f2e2132a87afe0e	0	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	3	\\x0000040009	{6499,3487,6562}	\\x84ece1a52f32085d0b0d6fcb997d29d574a3a0cd7e9fbf2bf127c245402a19edcca667692bfecc8042b8398fc50fa57b10277382f81a7e91eab44ae6bdafa8cedb95d364746a47caf4dd47890c6419e6665f1cb3d332c283e927d5d149dd6292	0	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	4	\\x5080040058a082001014510008000004000010	{5476,8669,8419,6499,6562,17226,9002,5144,8455,17098,9051,19477,19280,5609,6053,19882,6085,17239,5895,5749}	\\xa6609a167213c17065e96cd4d5f661bf31a6d614fdc23dcb44083e8b0a14a7b2d0bb9eb6f0d951137dab41d617e7201502e34d3f036c06dc6501b68f9e7e0727607dcd50a614eb2ff5e64602b90f3ba0e685f878477d1aad274b0c3ffb0e8c76	0	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	5	\\x509fa33f6dadbc7b42a63f09a2cbc3644cfbb716	{10219,2200,4429,8921,19194,16970,8728,1641,18702,2308,16975,15952,4781,4092,5347,8586,10318,10410,10260,16487,8813,17870,16853,17357,15911,12262,18676,11509,14082,13876,3974,8952,10549,12300,1415,8195,17201,18554,16987,1793,3555,12034,9340,10821,14648,17266,11285,17066,8538,13962,17474,14171,9080,19054,11269,8852,12297,5033,17283,2179,3717,5225,11105,10803,10011,1070,17880,4805,16771,10505,7044,19010,19065,4008,18885,9612,16990,12557,16535,3969,3403,11584,19719,3997,4695,8975,1058}	\\x819ba766e42cfb976072a0e0b723a9d1cc03162fb46d9057187707dd2d825a06dcaa1f7b2efb8d73f24586ea3c5a620d0b118bdf37d6ea0d3581f28ce10537855d5918bde1b03e752e24da24b4d842c2acfc5fd976a92a3e73440551c61fe6dd	1	2	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	6	\\x480000450042c100000000900802000400210102	{7025,10088,10650,11646,7113,10504,10536,9990,8785,11414,11627,8329,6876,11057,10386,17587,11440,16004,8787}	\\x94e2acf8beee506c861c1ee96942e62a2cbe9d1d885a16cc267668a0813f8bf2e71bff15b159600cc0d9447976f043fa147fb247cf894c0b0c8cb92472d3e66adf46db72e5fefb81c311df14c21411e7c4363b6d7796a2c7bf8f50733111792e	2	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	7	\\x14c72c094c854467888d132a21d1573af1844505	{13916,17046,17443,4516,14901,14067,14370,14844,4430,2197,10650,15172,17162,10364,8444,12286,18715,5076,18600,8785,4510,8353,2371,1336,3829,2832,2168,3483,14192,15440,15821,3948,4809,9357,16496,4583,3369,1312,3920,3351,2659,12229,12469,17823,3528,3509,1647,2387,19158,14338,18956,4387,9019,16804,14053,8390,3852,14718,12089,16004,14900,9660,19755,15017}	\\x85481216f0a7706ac22402ac1fb46ea7e5929368ef7c46e933ec4cd5ba8390e15a85d597689eda8c4fa17075b91257af0655bdd115f13c05e38c94a8bf6003488d0d164dad30d91174341ec0d779429f8f37835f1dbcda1bacd03c0764c3ab42	2	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	8	\\x0c000000004200001000000000000c00800040	{17220,5789,14760,13009,6536,5966,5859,5771,6311}	\\xb1310ae314804953ee323345577901828a9d68bd0df248b5a69331e1af08bcd669000e30a9b64b9eb3c4528c63d1b01b130321357ffc0094a144969013b94c778d70f377539d255b0ea9c4ee2c27a91054a7cce1904f57e1020f899df7b5bc76	1	2	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	9	\\x000820000050080c0002880018280000000080	{6344,5852,17114,5854,6098,13095,5514,6492,6555,6146,6001,6583,5483,9617,5082}	\\x841fb6f71e0c11885cb5aa7238468e37b968a51041eb1c5fd61b3ef7732194bbdaa8faada6d77df9012f76c5a6e0bd6500a5738c9fd9215bcef294e87440604a60de1d73f3b075b2a199a86b7e8bcc2fd644c863b5b1f7962a2304c6a7fb7c31	1	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	10	\\x011c4440800c49831c29229be4041322b080400b	{9947,19364,3713,3635,10298,3690,16499,9842,9132,1101,3185,8941,10073,2025,10652,9994,18526,14535,10114,4380,14465,5240,13908,14705,15298,10564,13684,8403,14056,1716,2298,10065,10807,10279,1234,10840,15876,18864,9954,1039,3693,10236,10584,9119,1564,5291,3973}	\\x820beb67b07bfd355b791ca1a0889cb75e3d64d5e3306ec20c8767069ccf0e1abf8ff1ee867eadf9c2a45a656624e13e05d4b285f31acf1c7af25423cb12d9ea77777c5b2f4fe4480189f22545fd694ee713df8d509d24fab22c68f5d5347fc8	2	1	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	11	\\x01df4142044cd9250c3df799ec841d433040d81b	{9947,18965,15814,19364,3713,3635,19179,4675,9231,3690,18370,16499,18209,9132,1101,3905,3185,8941,15806,10073,19094,2025,3865,3341,18526,14535,4380,19731,14465,10719,5240,17865,13908,15509,4492,14705,3756,14476,15298,13684,8403,14056,1716,13864,2298,10065,10807,10279,3349,1234,4037,2402,15876,16981,18864,15056,1039,3693,8729,18979,14310,9119,4716,1564,5291,3973,10661}	\\x882cb544afa0c59149f0c90f4997be991f8a409c54525c208bcaff613b2a7f10061e5fc64dd255c8be8bc7456b1e27b70c16b47578c136de38c58d2fda46a0718d683404da255fc565975766df779704f47868ab7be321acc1b6ec0074ac469d	2	1	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	12	\\x6f76b48680f570098877f1b2bc825e10ff6aa30d	{8915,8462,1952,17570,3730,11529,14562,16939,8431,3763,14582,2001,10847,1334,12437,17750,4095,2340,8788,4017,1660,19349,15060,3876,1419,9092,19246,3843,14256,18719,12118,2345,18110,2093,18989,4559,14316,9644,10698,2449,8504,17970,2867,17769,9658,18950,1779,18783,4701,11549,14246,9445,4088,14255,18547,12062,14993,9358,1309,14002,9927,17819,2540,3174,9732,14847,9031,15828,18897,17766,8393,8699,8846,17668,9858,10034,18896,4545,10676}	\\x98a1efa4f2a2253cf056de39a6c7d2696bd9a956771cc562303096100ee14fd7fa526bd3cdb217238be1bbd4a774d5b9012a860c31f8b3824d8d7ac164e03740a98144e761956d828679497a4ff43a4dc7fa9ffcc5190d47f8fa0dbcf992fdac	1	1	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	13	\\x10b00d1a4f3fce342807d98ee1989e036953cd02	{17606,2914,8730,5332,4418,11733,9604,3817,9475,18619,15180,15502,18669,17452,17315,19021,13910,2030,14128,12259,3808,3741,14210,2271,11461,2382,19810,4665,2502,1856,19500,16815,2478,14371,1924,9936,17828,12511,9330,13956,19055,14324,10458,4864,8721,13661,17035,18557,3268,1195,5461,8576,15013,18210,18789,2828,1383,14260,15086,10208,17978,9686,9600,3380,2403,12048,14078,2743,2314,14266,2839}	\\x8a678d20fcd29afeb99dbc6b3edfe34c37170860d3cc408fc2c40fe9169b52a8f41e21328fe2622125a2684bec58094d16dc0f05866affcb483a4a8986a01e04e76d4d1183b2738cad739334d3196e97129eeff39ebd89fa3013e0a82e98c521	1	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	14	\\x0000000000804000824088800000006020000008	{6941,11461,11391,10543,9369,9936,9330,10458,6896,10181,10208,11438}	\\x902b4a10e524da70c2890da49d182fc28eff9db0c53e5ee507d633c32d7a7c1464975db9c03ba3643333c011e4c357df1772b825e5973c9585a9b6ce47dd8baf0850dfa47eb7078d18476440008d0947552076ec90755ea3ca7904e41ec34ac3	1	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	15	\\x0000002040080400000000080121000a0004	{5642,13059,6410,6217,17238,5921,12962,6080,17151,5662,6043}	\\xa9f5b8cccef3fb2bc364fa0bf6c58f79b813017957ce5d1d08653a9b050e115373042e8d63b73f9b7486e5a5eecf33ea0255182c18b496aaa58311c61918fe79ed07b6133e40c33e70707dfdc6f493fe257dd40ded69da1dc1295ee7948c7bde	1	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	16	\\x024400808000100000b0020010464004800012	{5999,12967,17147,17132,6315,5823,17267,6206,5833,5857,17185,6321,5830,5697,5990,5904,6430,5832,6342}	\\x8f16c3a62155fd77b43b33ba7ba3dcfeb37f03b31a0d4c225411ee7c946567ef0fb2035e6a253ebb391cf89b29bc9d5a0f8466050a51571f104d63e8278629ec1cde2cf5eb327138ed90dae7dd4424728628c811bf45766face6bf65af4a88cc	1	3	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	17	\\xcaa5cff2aca691230d5d26b226d7db6bb95a3f0b	{19284,10442,10092,1045,18642,4395,11296,15019,1167,10672,19025,15548,10389,12175,12721,2820,4104,18680,4527,2582,16949,19908,12168,15084,2944,17736,19092,4868,18766,19008,15016,4736,10257,15068,1610,13680,18865,9680,10725,14998,2791,3912,9493,14515,3775,6957,17595,17831,8985,1749,9497,16802,17193,2882,8647,4531,3629,1196,14057,2861,19353,5250,3148,9392,2439,14590,16524,10245,18859,1818,5062,2802,1823,1973,1525,10538,2877,2962,13928,1922,1036,2813,3918,14348,3552,3868}	\\x9028b01385660d5a1abe5cc8eb32e384ea72d26551cb9500815beec5e5a86ba80d4d64322281e74929d6b44b9b56b307140f0e0723648bd89e6cffc5c18a4a030e382f1ffe07a64d273c40545381fc676479295253746565391dd0e077b31207	1	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	18	\\xca30c314860681204c05441420d4fb4108b82c0b	{19284,10442,10092,1045,11404,11296,1167,10672,10389,12175,8248,2820,9959,2582,12168,15084,2944,4868,19008,10257,1610,13680,10220,18865,9680,9493,9807,10265,6957,9497,2882,8647,4531,3629,1196,14057,2861,19353,10638,5250,3148,9392,10245,1818,1525,10538,11201,9329,1922,1036,3918,14348,3552,3868}	\\x974e3a4d4155f4f60501899e2df53c42af328702422d5dd77ccbaba6744522d76b4c79de930b93da78ec77963968e8b407e0c99295fcfd5b3ec69819ee68f9b39748cadb1c5b3a1c44a834d01181c58f87d5a52fd829f30ca38f1cec0863a115	1	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	19	\\xa51f8c01791ddbd4537c1a72fb418d211ad4e70e	{1059,1168,14827,2067,1948,2662,1815,8902,14264,19991,9048,11590,5042,14129,4400,17813,17923,3326,3495,4664,19854,12132,12296,9012,8520,3066,15209,3467,1857,14735,14372,3907,11064,1498,8654,14589,9313,9125,10759,1803,16718,9499,15914,19157,17585,5196,13460,2736,15930,2468,18541,17867,3851,9422,9559,2385,3017,13679,1782,4536,15168,19640,3787,9478,17466,2203,8656,18170,2306,8844,9017,1046,18576,3750,17236,10012,8722,5357,14134}	\\xaaf39996a4217986cb40dfbe6f6f6f03518f0736c19de197e2d447a4488d390c158cc46f8c95711e839994abf251a2ca0ae298f98ba4374b9c1f04bbc8f73304a6099d650b552ae998d71c2cd3eae0c252e69269b9fa38a4581db760a06832c4	2	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	20	\\x0000810400802000250260000082404400080001	{6963,11590,11642,11150,7106,11064,10166,7101,11115,11036,11022,11013,9165,9307,8255,11152,9387,10781}	\\x8cf32f7fb15b052bf565a2e068db8de235e279252e252197ba11b5fd8e9922b8076f912aa77f23238eb9d14c44657da80ffde49f1f18d653746e814671e4a274f393badb0485842a7dc1673b3660757ebeff0252105a3cad12010f69fb49abb2	2	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	21	\\x71572fcec7c2fa8031dde88df6f20beb34ee760e	{12121,3521,3441,1934,2992,3821,2788,17091,3722,10487,3450,3132,3430,8275,8307,2926,1682,16518,2284,12722,3953,1917,18934,6386,4824,4100,3885,2710,1423,4044,19906,14843,19150,3469,2543,1437,18981,18741,17909,18272,13471,2800,4894,18836,3608,8220,8761,5063,11472,2607,4522,15292,17600,10192,2056,15339,1753,8370,9703,1095,15198,15901,2098,9341,12516,15940,2100,17845,2024,4513,11219,16592,3683,2033,8501,14367,6902,10052,4683,18966,10817,12203,17033,1548,6667,4411,17946,9271}	\\xb45ca7e430810b8c0f8f84c53f0f3a20693d11b71f780953b1f6d4a53c3cf38e8fea3ab360752dd53de8fb71e1b204500c9c6455b483b7583078c3edd133e6e2b71ff3e93baa73f19a864b3631964262b0934d0125803fd49d323f05df1154c0	2	2	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	22	\\x00200000000000000000000080000000000008	{11891,12327,12332}	\\xa290556874145fbe4f83b35ea5ee06f30dddee733d9b6fe299d97f3986695cd65497283ab068aad73f52ba45979eb3481093da23983b03347cc6f93b7a7f60c24b2e46a21eef01dd3b02f85936085bbf9e7b6e8f7430295c00a51a3a43f1ce8e	2	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	23	\\x000000000000000000000008	{5507}	\\x8cc261ada8d7c50411af357f4e08729eac9ddb71db1e7eafbde2550ded422ef4168b98a23a5ab6d1e0bd2f114957524b0193549aa7fadda9fcefe697bd70f9a7ca71962d6563bda75375dc30a3f74ae8c8c46896d9c62e5116f8bfbd753d82a0	2	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	24	\\x000020	{11906}	\\xa0abb831a03b501404381c72a61e7de88d988de159738dad8fc70d2d341df3e2cb304002dc782b3789a9b4d8a9d3a0b50360084ecc05d619b29e535da7defea243afe7ea28e8720e21dac6a44f1a0fbcfd2855108f6cae90df90aa6d56bf6e42	2	3	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
3	25	\\x02	{5578}	\\x8f3affc1632d9c37c893b2278f46eda0c23a804d17b50caa16d82b1e7442d49aef49f45e8e0c1d46e115c657c65e2da70396325fdb352c26747dada3ba1a9eec3c91bae76a430e9e4da74d9fd04b5f4eb2654e8688d110e9b653b7ce528f66c3	2	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	0	\\x10e79666eef284b6bd81bd5b27f3933743fbb716	{14073,12287,14148,1533,19637,17002,3677,2266,20017,7112,9018,2169,2986,13626,2522,9716,3238,3822,10585,2957,17293,19716,2258,10649,1138,6940,6954,9545,1397,6126,6859,2873,1557,12107,18774,4688,11568,4694,8596,3461,18351,9116,2893,11238,9194,4738,3909,17539,9328,18920,15346,18884,11594,18678,1515,3493,17858,1567,8402,16768,8791,1132,18890,10240,11576,11117,14559,11467,7061,9515,8169,19929,8869,3075,4038,6868,9719,17953,4391,8954,1412,14937,10689,15804,17686,11557,1873,2012,11228,2746}	\\xa5632a7c90f4d4c8c6b8cb64f26bcc42914ba16c86d037bf8399939939f513ebf0648fc6e1f08aa828c91c816db5cdd20af6613b24f69a67cff872c72ef617d3e43a34f8b506bdbc3e5c427f61e235345b96e4f24762f5af078c32df1f218c6f	3	1	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	1	\\x7f57a37a2cbe8dc139335ed3534910ccb77ddc04	{11224,2904,17463,4875,8473,14470,15790,15880,1220,17713,17609,1680,1902,13979,12166,7214,14596,14084,3770,1379,18298,17174,11413,11533,1650,4473,14548,6966,8660,9362,1871,9998,3385,3590,9775,9011,3033,4620,17723,17709,4584,17976,8425,19294,3563,20055,11486,19735,19617,18682,4674,18985,18717,18517,10391,8243,4755,16538,3020,14219,15532,2556,13926,17851,11453,19943,18590,14328,8552,2464,12164,18728,4747,14020,14637,18894,18270,14534,9070,10241,2286,10700,1992,2356,19229}	\\xa9a220b35edd5d28c14ac003c3b0a3dcab04c4d88f0baa1863efc620f27e96ce7c7c393c56b0b38c1bd357a5dab421090d435d46f8a85d6e443b768340d7c768f3534a8e900144652a770646eb863bbef7c4f7a7d0c5513101ad4440eff2be5b	3	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	2	\\x82a476323941b775c5bc17d9c43a004470f9d115	{12283,3965,2413,3856,3155,8986,5328,16148,18737,3263,5352,14169,2490,1942,3597,1744,2899,15829,10691,10387,14168,14579,3181,18958,14326,9715,3761,13905,2654,11643,20054,15045,8684,17261,3706,4714,3785,10880,14069,4086,9717,2341,17662,16142,4500,3926,14677,9310,14385,18734,4644,17901,15624,14047,3137,12072,2757,2880,12487,1625,17408,4377,14463,1522,19084,2405,9412,8881,18688,5329,5160,11333,14792}	\\x92d256aea2e875b08b8748f45132b964e74e1f7b8b811c75febf1f077460092e82470fbebb64e10864bfa7a2fdf437d200bb78caf23a0f66b63e4cb036a5507c9d6febda10c6141ef1cc4ebfc31458f1d18fc1802d9a5346f00e19f77bf3f68f	3	2	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	3	\\x751c5a9c38861b9b2817d61620271f88d2c06a10	{14992,9414,12261,4845,18524,16492,12135,9450,2932,9252,4790,2943,8984,18671,14334,14369,15999,2844,4505,8226,17666,3800,9477,3289,9052,19791,14804,5330,1313,5150,3100,19035,3302,9077,17671,12151,2693,8479,1357,3699,9605,2811,3758,2997,9501,19241,10690,9065,18686,8290,8888,4605,19998,17166,12236,2431,8429,19306,14064,1181,15916,2140,3135,4585,8215,1838,18826,8225}	\\xa3cce9fa1c062cf08afcf6875872af1849fd61f09c4bb4ff0c5f15dd1fea9ed2cd5ece3e968c982ddc3ede731d0bbf4605e8429c7c97516cca9afc1c2df1da5fe612df61c4b510973810bb5bcec2c76b25a248e8f666f408a6588650c7ded4f5	3	3	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	4	\\x000309190080894009030001100081800000a6	{11433,6803,1116,11424,9404,6952,14169,10205,10387,8342,14326,11643,20054,6912,10542,1350,16142,9938,7051,6798,11496,9802,10525,6970,5329}	\\xacfbab11995bb972326f18c773ad3ceb502243f7a7871334a21d9709e382d54f8cbbc648fe14e8a7df3e70e8ddd7a3d9027fcff008cd7e8f6026369a925106782a9ab7b12457d357bc4f225c84910a3814f2ff612a9be10d1e360d33574bd9b4	3	2	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	5	\\x02800500400100000400000013200001000210	{10039,9885,11608,11274,10607,10250,9972,8250,11444,6960,8290,10178,11365,9877}	\\x988de60c7164a10e2460e66924b62ec96511067984fe78b1d95a787e08ddeff86b3139b97816c0d6b359f3611c3a991617122d4f64a9774c3aaec2697bdffb3ef315abe0b2c703f72a2c5848c126730dfd13b57906cf0b2e12a7967b68c40c7e	3	3	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	6	\\x0400002940014000000010040120041000004004	{13916,10650,15172,6587,8444,12286,8785,9357,6408,1312,6552,3509,18956,9660,15017}	\\xaf4d9d19e403566203f2bc61831e57dafa4a48efb39fb5ef83c24434a314f0f01cc349ef0ca0a5594f220497fb4706c70f0e9a00676f1706b3d47106c2625e88a5042f0f308cf5bf7f5e94b352414b4cfa99fe396907aca93f262a9af093dc33	2	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
4	7	\\x25400c00810040001011c0865420008230000008	{12121,6370,3441,3722,3132,3430,12722,6386,14843,1437,18741,13471,8220,8761,6400,11472,4522,17600,10192,15339,1095,2100,11219,3683,2033,9271}	\\x8aa753204188f176241fcb065154b1324b85c9cb4ea1a03559ee528e8976763f788df5a5b671a011e6361ead2c65b6ab049d95ab274a5eb939a172ee024fc0d819a48735799bf9e15b9e47b25f096710aa3636fab5b763a87a3785dcd70d64f3	2	2	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	0	\\x0100000000000000000000000000000000000002	{5785,6021}	\\xa9c7865a50e976eddd1cfbf002e82d13c3b6f8b80ded39ce40b9e9c717439ce29760103fb2bf99535d2ff08469ec52ad03517487cea2a2a6255e9c394c78449744d667adecfa00c8cb23aace0e366b065b1fe4b9a73ead8e5fb290d6c063943e	3	2	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	1	\\x71572fcec7c2fa8031dde88df6f20beb34ee760e	{12121,3521,3441,1934,2992,3821,2788,17091,3722,10487,3450,3132,3430,8275,8307,2926,1682,16518,2284,12722,3953,1917,18934,6386,4824,4100,3885,2710,1423,4044,19906,14843,19150,3469,2543,1437,18981,18741,17909,18272,13471,2800,4894,18836,3608,8220,8761,5063,11472,2607,4522,15292,17600,10192,2056,15339,1753,8370,9703,1095,15198,15901,2098,9341,12516,15940,2100,17845,2024,4513,11219,16592,3683,2033,8501,14367,6902,10052,4683,18966,10817,12203,17033,1548,6667,4411,17946,9271}	\\xb45ca7e430810b8c0f8f84c53f0f3a20693d11b71f780953b1f6d4a53c3cf38e8fea3ab360752dd53de8fb71e1b204500c9c6455b483b7583078c3edd133e6e2b71ff3e93baa73f19a864b3631964262b0934d0125803fd49d323f05df1154c0	2	2	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	2	\\xc40542b91e2436aea1480761081082240b552515	{19012,2994,3875,9963,8696,7033,11148,7006,13970,12052,2895,18565,9389,10645,4735,2764,16725,16482,2928,3212,4712,7084,13944,12499,14108,12162,9893,9993,4042,10511,2669,13892,8328,6842,2328,10822,2113,9724,14231,11593,3005,14828,12559,3990,2342,9599,12233,10184,8670,18571,2579,19723,2437,5295,9872,5450,10824}	\\x98408ea1946a2dc37739727c1f3efb3915cefbc93cfd8e3fe00eaa7f429d9f07091595668d7b40480c6ae284290278fa0f0597d1371a66878db8cea6454da764b287af1e0375715b8416afa187d8e9fb972577f267a7078afd6d74e095bb587d	0	2	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	3	\\x7f57a37a2cbe8dc139335ed3534910ccb77ddc04	{11224,2904,17463,4875,8473,14470,15790,15880,1220,17713,17609,1680,1902,13979,12166,7214,14596,14084,3770,1379,18298,17174,11413,11533,1650,4473,14548,6966,8660,9362,1871,9998,3385,3590,9775,9011,3033,4620,17723,17709,4584,17976,8425,19294,3563,20055,11486,19735,19617,18682,4674,18985,18717,18517,10391,8243,4755,16538,3020,14219,15532,2556,13926,17851,11453,19943,18590,14328,8552,2464,12164,18728,4747,14020,14637,18894,18270,14534,9070,10241,2286,10700,1992,2356,19229}	\\xa9a220b35edd5d28c14ac003c3b0a3dcab04c4d88f0baa1863efc620f27e96ce7c7c393c56b0b38c1bd357a5dab421090d435d46f8a85d6e443b768340d7c768f3534a8e900144652a770646eb863bbef7c4f7a7d0c5513101ad4440eff2be5b	3	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	4	\\xdfce7c9bd2abbfdca90e8970cf93f02f105bab05	{7075,17689,8282,9495,3716,13930,14891,11345,1738,8318,3314,8854,3298,6810,9746,2184,19311,4656,16777,11662,10256,10518,6854,17718,3638,2302,5287,13895,19291,10151,7046,9734,14141,11092,2702,9213,14297,11310,16806,6732,12124,2749,10312,1531,9498,3639,17807,2631,9142,1532,1544,2420,11492,2419,10561,8733,7094,18923,17734,2409,1656,14068,4799,4710,4445,2316,2530,9524,10688,4002,18693,3136,8286,19807,9049,2243,2292,16458,14337,15861,8623,8357,15192,10464,15078,5454,19373,8285}	\\x889e3406159f0827b8a683b141c7fe7eff1cb5fef963580790540197ab3a9ebcdcf16ae1f30d47a75d249cd3608077290746fc2980e670280ef32d46cec413dc695c82bad6ec99db9c9550eeb0281bd59b3022a10c3eb99051a5ac9b49f3cbf9	4	1	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	5	\\xa51f8c01791ddbd4537c1a72fb418d211ad4e70e	{1059,1168,14827,2067,1948,2662,1815,8902,14264,19991,9048,11590,5042,14129,4400,17813,17923,3326,3495,4664,19854,12132,12296,9012,8520,3066,15209,3467,1857,14735,14372,3907,11064,1498,8654,14589,9313,9125,10759,1803,16718,9499,15914,19157,17585,5196,13460,2736,15930,2468,18541,17867,3851,9422,9559,2385,3017,13679,1782,4536,15168,19640,3787,9478,17466,2203,8656,18170,2306,8844,9017,1046,18576,3750,17236,10012,8722,5357,14134}	\\xaaf39996a4217986cb40dfbe6f6f6f03518f0736c19de197e2d447a4488d390c158cc46f8c95711e839994abf251a2ca0ae298f98ba4374b9c1f04bbc8f73304a6099d650b552ae998d71c2cd3eae0c252e69269b9fa38a4581db760a06832c4	2	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	6	\\x08020000002004000000002004800002002404	{6051,5784,11162,5604,5115,3058,6544,5815,5465,6120,5911}	\\xaa551c9cc730c2f831bf7090b7d4ca3739d7e6bb293e6c3a8b6fa229342efbaebe40cdbcc3475a103ac0b42218a31b0f08c9c3aa39f89f2f6dcf773ab6c4f0ce5d080eda80b924dfea80656786892c495713ccd510122531878efa5caaf35b37	3	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	7	\\x10e78666ce5280a6ad81b55b26b3013543f99712	{14073,12287,14148,1533,19637,17002,3677,2266,20017,9018,2169,2986,13626,2522,9716,3238,3822,2957,17293,19716,2258,1138,9545,1397,6126,2873,1557,12107,18774,4688,4694,8596,3461,18351,9116,2893,9194,4738,3909,17539,9328,18920,15346,18884,18678,1515,3493,17858,1567,8402,16768,1132,18890,14559,7061,9515,8169,19929,8869,3075,4038,9719,17953,4391,8954,1412,14937,10689,15804,17686,1873,2012,2746}	\\x81532b0b01f89eb302dcb4604ecb6d51aa7878d26a66d0af05a4602109667d216cc2cbb9b71280eba0082b408ae17a3215953d7e902ca44adb57ab84883e5696675741a6acbbd14d12f3a2756555c799d469aa13f272d9506901969e33429882	3	1	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	8	\\x0486840e8d303cad443551f02987ad0d883e7d03	{16982,13947,2497,5113,18801,14336,12265,3870,2689,3617,18982,9487,3200,2055,12156,9120,3600,1108,14696,19348,9010,19845,17760,4468,14549,2480,15840,8387,4007,8644,9416,17308,10760,9266,12144,5040,3675,19623,4105,8542,19668,11764,19350,15760,7009,2231,19848,4829,16973,13973,4526,14645,8652,2018,2121,18927,16965,8498,18891,2368,18519,2665,17705,19745,3475,16494,19101}	\\xb2c460fcf97077d646dd1e75659b30df4e8ecfd359fb4518acd85c4a910b12642aedc4a567c508c9c5f77d939aba9a1508d7e62d8fbc13335546ab3b39868d13af78567438a7e0efef3236ebbca4881d84f756ecd087cfe223a4f80e47216eee	4	3	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	9	\\x8018400010000040100000000042020001018010	{11614,10341,9226,9887,10002,10526,10336,11764,6967,10188,6889,10612,7069,6931}	\\x96091ae2ecf71ae31d5816bb67afc3aa9b742edc0acc7bb68a2b9c8a157811d8b80011df1637f8d7ba1f7d23cd490c5b1837d4db94e2e1d682d083bd2dfabca4b143cffdfe16847896c10b269ef40c4cdfbd26ec5cc397dea2d520af12158c2e	4	3	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	10	\\x83df414204ccf925ac3df399ee845d473842da1b	{9947,6105,6102,18965,15814,19364,3713,3635,19179,4675,9231,3690,18370,16499,18209,9132,1101,3905,6165,3185,8941,15806,5920,10073,19094,2025,3865,3341,18526,14535,6513,5521,4380,19731,14465,10719,5240,17865,13908,4492,14705,3756,14476,15298,13684,8403,14056,5715,1716,13864,2298,10065,10807,10279,3349,1234,4037,2402,15876,5559,16981,18864,14758,15056,6121,1039,3693,6349,8729,5986,18979,14310,9119,4716,1564,5291,3973,10661}	\\x837f3f4fb28f826584067e865284ac0eaa75d8c59b2853e6b4ab8d8adebdbdb8a5c23ff10d37bc9f994c45579f196a0205cc200de0f3f150936b4eee3512e479cb47dd7602b4f862a3b66dc2bc3e7d370b53d4d0a29b7163e10ff02c839daecc	2	1	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	11	\\x00000000000000008200000000000420	{5961,5821,5829,5849}	\\xae31a0fd7130464904f8cac8e00488847f17ff3675449b0a0979ec2704a5131ffaf76e050c30801661599a2f13fe411e0be39cbdf54c2423f7bc3ce39a3197af791cd069ef398aac4916a9c256dddd9e3abd17d932a0b804d6731f4a3c81b76b	3	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	12	\\x82a476323941b775c5bc17d9c43a004470f9d115	{12283,3965,2413,3856,3155,8986,5328,16148,18737,3263,5352,14169,2490,1942,3597,1744,2899,15829,10691,10387,14168,14579,3181,18958,14326,9715,3761,13905,2654,11643,20054,15045,8684,17261,3706,4714,3785,10880,14069,4086,9717,2341,17662,16142,4500,3926,14677,9310,14385,18734,4644,17901,15624,14047,3137,12072,2757,2880,12487,1625,17408,4377,14463,1522,19084,2405,9412,8881,18688,5329,5160,11333,14792}	\\x92d256aea2e875b08b8748f45132b964e74e1f7b8b811c75febf1f077460092e82470fbebb64e10864bfa7a2fdf437d200bb78caf23a0f66b63e4cb036a5507c9d6febda10c6141ef1cc4ebfc31458f1d18fc1802d9a5346f00e19f77bf3f68f	3	2	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	13	\\xebdaf69717d01ab36f3dfe9ecd525f060b365503	{10355,14574,12035,11034,11166,18208,9175,18220,3270,1164,9374,10218,9721,18773,4577,11225,8904,9187,9407,18907,19113,18832,3370,6834,2082,5061,17717,14520,18834,2275,2773,4452,18930,17891,4748,18707,3344,3038,11432,10384,3806,10202,8672,8496,1211,9492,8784,15228,10437,9166,2407,17733,6976,12183,16677,3825,8837,17816,2370,2657,14394,9698,6778,2539,11008,4435,15944,6224,14547,2548,11476,18121,2429,18689,8581,10156,14731,4786,14151,8480,5166,3602,1937,4046,16769,9520,11559,4402,8214}	\\xa5f242cb9ae41439ffbbd1c02fda735c1a95b08793165ad7651c266d3e9811c95a32b3f6cb949a90b5ffc4cb704873dd13608a1edcafb837db4612ed67deb2fbf5588dad29ca0b288db15313369f6864ec1142b6e92a1d5189ed3c5aea709ff8	4	0	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	14	\\x00010040004000010029800020000000200214	{1204,1692,14520,18930,8496,9492,15228,16677,1368,1121,8480,16769,9520}	\\x969e6c03940c5129f52e0cd9e86b0982fa86950efed7484b08e4cb24c027a61f167422a3c27c42762621491bad80a04214fbc3a7a4877859729cdb0b60a45c84469a7f00b92486fa0d02c707a67fffcf14fa83f66631883b2902d5386fc928a1	4	0	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	15	\\xcde693905268d50eb2c60791c846ff11815d20	{4515,8662,9852,16585,3531,15964,2390,14571,9182,4885,3209,15945,14706,2075,1652,8433,3893,8822,13664,18365,2622,1635,3700,1701,3463,8262,9533,14038,6387,4482,1236,3009,5221,8475,8603,8771,17673,12067,13557,17654,1034,17648,13620,19890,2394,16508,18579,15146,2616,18624,1906,1566,9352,2924,10116,8781,18915,15290,17945,4099,4918,8704,17003,2110,1136,14713,3516,19178}	\\xae02fe6d1130e57c26e3ab66c4019cae7b6def273436662cce506fac59ef98040106f2a10b0331629fff1d55e77f17d6029f8d2365a9cf9345fc1da6df90050a484f972b6d48e15b625a10af1432fcd38aa28a47bd936455613a67387451cc32	4	2	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	16	\\x08000405000002010412002400801020200010	{9852,7027,10737,6934,9863,11434,10190,8603,9909,6884,10510,9834,10116,10488,11393,11408}	\\xa2a9b866d18dbb123c610d18b8256f8064314e22ea3d77a749719ca0407c10f686f2c928906705798d37efe62c8cd29118b43a1b584905072cc3c9c81cfd3560ee082915426f050b2b0d085bf7bcb8d8b40def3fa04a812f1ef5c9098429e20d	4	2	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	17	\\x4f8e0043081102fba666501b248075acc38c2a19	{9172,11101,14268,14207,15941,3072,9873,12303,4737,3486,12051,17912,4487,11154,3855,3149,17803,15052,16851,5296,10465,6652,19715,17932,10106,16006,14243,14109,9841,10439,10552,6871,8336,3379,9988,2978,3083,7050,5146,3345,11558,15356,2505,17895,14811,2467,10027,4023,3153,14846,14332,17282,2934,5209,17866,17140,14061,9718,3669,6841,6825,14170}	\\xb3f009b05b7b4a094efc485bf8e26604eac04d69beacb0778fd5e1cf295a21cb3b5faa77d2e429b138fb4461f94b1be00bba29d717e2ec0264411404240d385c61737b604f4b752a925b807ea376d06e74a8412d9a13e8c25ad55e4af9480b06	0	3	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	18	\\x000000000000000000000000000000000001	{17587}	\\xb00bf38596cca13e71ba232a5a182a5036a1f4c6781baf4a36310b6f1e5bee96cb5338fe75c082bd0175bccd2332f41f0376d2f21525744929f373e44e4acb95f56cc99085bcc74db27afe37412f08821582ea26cbf39e101c8e453b4dc0fdae	2	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	19	\\x04	{3060}	\\x87ef5e8d407f84081ccc10cf1d6fc1d3a3f0d2ace28b8827569dc9c8e4fe93b7328e76faa7a0c633fd19787de924d331040ececcb76b886ea23c0f534866dd51fc0201704345846141d91b2ff0823d0d4a6038a36f6348e9d596e12bef681acc	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	20	\\x000000000008	{5597}	\\x823565d8d52e5d5863da08d17d95e3a57f831a1d30daea12876d6ccd1d95976cee1f35194a7d1ec7749829a03d856e5218c27a45f5381e7ba1df42d1a904e9feee0bd9f2a36d336f31e751ebd11491bb322946f53949d9818a95a4f0340be887	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	21	\\x0000000008	{5953}	\\x8a129300c7c0a99ce05f696d2c4c652e0d972eeacfa30062166203c25de85359e57642774c8af2e1fd12b32cabab21b312913c5fd0cd6c58f3633bf94e48fa8fa71cfc1fe9e094754e502d0ed923a4a9b35ea27a1d1ff1b1049a3b9de893e975	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	22	\\x000000000000000000000000000000000040	{17262}	\\xa6a63e289dcf3fa0a852557593b0d253136e72a3dd434785cb5b976a2f28dd1e2f42f1e94cf1bd65366696f086f8974602024fbf12c6fb1d4a1d014f17b108ce7ed2c71a80c93836f63ed156142aacb305e4325ce1dd5bc0e0f5b4f3dedceac4	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	23	\\x000000000000000000000000000080	{12013}	\\xa6c2b29bb954f646f4b67dfd674cd2325686c8003adc216a57c3e4d79731e909becfaaf237d7b7b536ed3906b510e848016f1cfa0df1c0f4f5a9e064327861d0d615f1edad7984a22abff07b2e036fcea01bcd4a5356175d96559fd9b776f1a7	4	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	24	\\x00000000000001	{5491}	\\xa11fdafce82e6842c6432ee2e0dca0b34c8a19de57715fefe530a33e6c53bc03c408628a04b7b39e17a16be33286471307da7d7cf59cdc40a0ef23f38223d313d974ecdf34210968549460e890bf1bf3a951e19380d9bedd96e69fbfeaade325	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	25	\\x00000000000004	{6570}	\\x81062d621e2cfefccfab0aba7a6b325b5bdb4524c9b6dcc095f2b2a1f7203a69c90b25bcbb3dce6adf231b4c1960833f08250b4e8a26fc15a5d6c75c91bd899691401b9d487eb8432dbb7058dd57544e9f73869242017baa82b8b7106a8cb0da	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	26	\\x00000020	{6448}	\\xb792a59901be106bba5612184fe234cea995ea3f7381c6e4136af16d1f5bdeec6a4e65057690e6158c593ebebbf5cb4c17fcdcc2c587827165a5b8b423ef21979dd880b198de55ad96d4fc1e3576cffa2e9cb13832749a8824801ba845024a53	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	27	\\x0000000000000000000000000000000000000004	{6012}	\\xb006975a3e2d3aaa1a991bd3dab2a690f29bf030ef692c77d9eb4ed18fc5f36c83dfdef4bac2479d68e23de81d98ac7302c247b95a00496462c12ef1e8db6c3884a39d064897c119df4dae03b7f31ff9d6d5b6f7509ed74a2141a10f4f19ded0	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	28	\\x000000000000000020	{15242}	\\xa5019b8b95ee2b33589dfe1c5d787926eef49fc9d7c79e8d9b476cc0732c9d5dd987a3b5df37bb270e3157de39ab61b21900aa01433008241a29757951b99d956b6a49ae233ce2b43d9ce28f4a54f87ac3387d15f8a7a979fd23bd018d1c531c	0	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
5	29	\\x0000000020	{6175}	\\xaa798a9d8bdb5d6e1adef4a4e82dcd23f1d6bd943bb223fe384e32bab67ccb6d303a59b229d0915be7df98a17f9af11913daf7e90b041c4137c13d441aa904f5005aa2bff7ca9462e3682f878fac4bfd977cc14a9e73a245e8019e7ff06ee5e1	4	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	0	\\x00804000000050088200a00000004420	{6109,5185,6118,5666,5528,5961,5821,5472,6250,5829,5680,5849}	\\x8c1365d89b8756908c5d7064ea7d0be14979d2f55168e140421d9739cb86ad5e472ff658e14a94b72c359ba27a254cc109b66e143c388eca3645bf3677f3ca7b997cba3c90046038c4870c3e8a6bbea7843039ec9c6d829cda0d302a4f91f813	3	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	1	\\x54066a87a2135ea941dcd4edc067e3e2f4fdf61e	{8858,6875,18643,2973,4586,3311,9847,18362,16519,10055,17846,14607,10985,6722,15073,9957,10833,3325,4852,17352,19341,4426,10098,14174,15782,1377,18277,3223,13666,12214,6734,9032,5107,11641,2938,19743,2797,17637,9711,4373,3519,6770,1877,12134,14189,19138,5445,4534,9627,14469,14163,14301,8306,1118,3170,5057,1137,8327,10642,18710,14222,3333,8458,6964,15092,1646,1847,10815,18947,17572,19753,17952,3757,11007,3613,14355,18638,17982,3096,5337,18602,3078,14723}	\\x8d489e78956c337e86af4a8b9ebc4d658ad15a9c56a4aab19eb30f96c61d43e2ccd3a5470af47b8c2e924b70fbb9b0be0128f54a7a8ba39d9628bf683410c99542c7c55ca30c7faaf49b77b1756b2afe6ad7c627c6ca97e519deda82fb0c500a	5	2	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	2	\\x200000000000000000200000000000008080	{12116,11972,11908,11896}	\\x9816c1dda93594e1cd48c11158adb9633bebb5c3fe21ddced09371ae2be5d8a4e132a233302a94283925f2dd89b8bb650c9b5592c2569cbcc5cf5937d47b7682c4d76e6c0711a2054a035b637e0001a066610252fd35eba6f0b1c50550162d0e	4	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	3	\\x0000210500c026004080214016800241002420	{5695,9184,3486,5840,5843,8449,3149,9183,9424,6060,5648,17121,5624,9296,9123,7050,8842,3345,8874,5182,8217,5209,5967,3669}	\\x8bf41a32b36604707b33028c181cb66cb42f1be6c4e5b862a4a5a47276729009e39e5be8c0527a5bf132be6c8e8351fa1734d94de46b667e2fb9f0c8bf09009e14522a9edad632d1a422157ecbdfa824029b80a964e091bafa31ba914ad9901d	0	3	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	4	\\x820001020000000ac0000104100000800838	{7076,10786,10487,8307,10499,11422,6903,11615,8337,11472,10192,11219,9948,6902,9848,10052}	\\x83eb08e727e09514e0934a8613db54db534b34bf68b680054c550a8afb1b70120848e4400a97bf0c51adb9894f5e08d500cadc7280f345626ea83dd0599ada911e8818e392a9d27ecf54f0067f5b9978130839ce7ea8ff2624e6082e742f2bcc	2	2	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	5	\\x00858040001408040000081041000008002208	{10321,8604,11561,9832,10651,11401,11553,8325,8709,8258,10153,11382,7029,10104,10432,11577,10433}	\\x942d1be5c0a722045a7ef3849408fb2dff086b4b2f8cf4f69d68aecce325d64dcfe979ad175bf69e98113b6f8079637507f1db438665cde558f383eff395bd962ed7f8faaa58d79d2cdaef28482fecef3568f0612cbbc0a56cdc99c222c16a67	5	0	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	6	\\x770822bfce6bb4f2495572eb0c0e1b200ddf4709	{9118,4410,1242,4580,5047,12079,5444,18831,3952,11058,17103,8528,17224,8751,3899,17949,15950,2703,18340,3718,17640,17322,3168,5277,12074,15094,17168,17877,14347,4835,4621,3916,5258,2078,3342,18943,1477,3930,2037,5226,18782,4385,4893,12564,1962,17125,3500,12081,14138,3027,1483,13940,4444,18793,2355,5441,19218,19841,1828,14137,3215,9190,10668,12309,8533,14257,10432,4635,4662,4609,2379,18747,4101,3898,19089,3749,3684,9062}	\\xadea8d72267f8afdafd4c6e26cd09615320a57ce5209d824740d497d8074629647bb5450865d5c86ae4126bdd461b4271001574344d5ead415010347f8164ba746c258f4dd213e70a131fe0f95a88341c9b0417ae07e4d682551bd92df220fd5	5	0	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	7	\\x0486840e8d303cad443551f02987ad0d883e7d03	{16982,13947,2497,5113,18801,14336,12265,3870,2689,3617,18982,9487,3200,2055,12156,9120,3600,1108,14696,19348,9010,19845,17760,4468,14549,2480,15840,8387,4007,8644,9416,17308,10760,9266,12144,5040,3675,19623,4105,8542,19668,11764,19350,15760,7009,2231,19848,4829,16973,13973,4526,14645,8652,2018,2121,18927,16965,8498,18891,2368,18519,2665,17705,19745,3475,16494,19101}	\\xb2c460fcf97077d646dd1e75659b30df4e8ecfd359fb4518acd85c4a910b12642aedc4a567c508c9c5f77d939aba9a1508d7e62d8fbc13335546ab3b39868d13af78567438a7e0efef3236ebbca4881d84f756ecd087cfe223a4f80e47216eee	4	3	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	8	\\x00000000000000000400000c00000010	{6087,5629,14775,6298}	\\xb18ee5bb172f717afd51540f331628b46d1b9dccd1656201184e3fa09c8592f9b49bcd078d0eee7c56e60e66ff65fcd1134e6a448c6ecc5de640a9518025c20d7074e54cae9ddca71ed966065ab83b83cdc4185b5c5a4a7c765fcb876cd2156e	4	1	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	9	\\x10b00d1a4f3fce342807d98ee1989e036953cd02	{17606,2914,8730,5332,4418,11733,9604,3817,9475,18619,15180,15502,18669,17452,17315,19021,13910,2030,14128,12259,3808,3741,14210,2271,11461,2382,19810,4665,2502,1856,19500,16815,2478,14371,1924,9936,17828,12511,9330,13956,19055,14324,10458,4864,8721,13661,17035,18557,3268,1195,5461,8576,15013,18210,18789,2828,1383,14260,15086,10208,17978,9686,9600,3380,2403,12048,14078,2743,2314,14266,2839}	\\x8a678d20fcd29afeb99dbc6b3edfe34c37170860d3cc408fc2c40fe9169b52a8f41e21328fe2622125a2684bec58094d16dc0f05866affcb483a4a8986a01e04e76d4d1183b2738cad739334d3196e97129eeff39ebd89fa3013e0a82e98c521	1	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	10	\\xcaa5cff2aca691230d5d26b226d7db6bb95a3f0b	{19284,10442,10092,1045,18642,4395,11296,15019,1167,10672,19025,15548,10389,12175,12721,2820,4104,18680,4527,2582,16949,19908,12168,15084,2944,17736,19092,4868,18766,19008,15016,4736,10257,15068,1610,13680,18865,9680,10725,14998,2791,3912,9493,14515,3775,6957,17595,17831,8985,1749,9497,16802,17193,2882,8647,4531,3629,1196,14057,2861,19353,5250,3148,9392,2439,14590,16524,10245,18859,1818,5062,2802,1823,1973,1525,10538,2877,2962,13928,1922,1036,2813,3918,14348,3552,3868}	\\x9028b01385660d5a1abe5cc8eb32e384ea72d26551cb9500815beec5e5a86ba80d4d64322281e74929d6b44b9b56b307140f0e0723648bd89e6cffc5c18a4a030e382f1ffe07a64d273c40545381fc676479295253746565391dd0e077b31207	1	0	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	11	\\x55d29393a33fc72157b6d8d5d725abbf15c6020d	{13959,3801,12126,1758,1069,14226,14035,12308,1276,2715,1627,14680,8929,5255,14286,4818,18242,2224,10819,14211,3968,8324,9743,18874,1337,10654,5244,1271,14800,12115,3201,4021,14996,13870,8693,1363,19030,9519,8549,2508,16005,3368,2741,20034,2289,2028,18917,14536,17714,3539,9122,19235,10756,8803,9367,3445,12113,19700,17878,3346,17836,10398,8280,1762,17271,17811,10722,14719,11506,15879,3243,8376,19271,10228,3560,2771,10040,18612,8886,13865,18580,18797,1740,2676}	\\xb44d50bef0f1ca19f00b164155ff8eab5b1ceeebb9da7ddbb82ba5e08f86f6bccf55c03a935789e3ad25d62aea23eda50e85e2c068e06fda38d94013f5a2f99dc399a6f3dafbdd866e877f7229d0df885e48f963a3b08628534dd9dcbe0e78ad	5	1	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	12	\\x7f57a37a2cbe8dc139335ed3534910ccb77ddc04	{11224,2904,17463,4875,8473,14470,15790,15880,1220,17713,17609,1680,1902,13979,12166,7214,14596,14084,3770,1379,18298,17174,11413,11533,1650,4473,14548,6966,8660,9362,1871,9998,3385,3590,9775,9011,3033,4620,17723,17709,4584,17976,8425,19294,3563,20055,11486,19735,19617,18682,4674,18985,18717,18517,10391,8243,4755,16538,3020,14219,15532,2556,13926,17851,11453,19943,18590,14328,8552,2464,12164,18728,4747,14020,14637,18894,18270,14534,9070,10241,2286,10700,1992,2356,19229}	\\xa9a220b35edd5d28c14ac003c3b0a3dcab04c4d88f0baa1863efc620f27e96ce7c7c393c56b0b38c1bd357a5dab421090d435d46f8a85d6e443b768340d7c768f3534a8e900144652a770646eb863bbef7c4f7a7d0c5513101ad4440eff2be5b	3	0	\\x21dd54cc88833f37666ed4fa3649b31a7d55a2a18dddca9e1a8bffaf05d8ddab	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	13	\\xbe2885f9d693b0b2f947d890013ed7b252d90714	{3634,9000,17513,14560,16507,8991,14586,14636,8311,14144,11457,9884,2152,19253,1644,19666,9135,3615,9706,16137,8375,1741,2740,3054,15308,14099,17250,5343,18670,2440,8848,12133,17724,14569,3231,9730,19236,11075,15083,10647,14360,3214,3832,5249,4846,1056,1996,2317,18124,12101,8434,3788,9224,17453,15179,14271,12273,9171,3217,15189,18685,2282,16936,3296,1180,18582,18626,10350,13827,8200,3543,3902,5273,3241,1999,2304,9470,10204}	\\x9931cbad631e9f0c572c6332ebc8d5ca550ef266c71cba3d5a1c8f528044d43634824aa0238bee6eaefd9c14f2e1508401f764e5197b84932ab1329ba8c674037fdd254609e53499e1ea87631c5686aed4b123b37078ef4c72e5ef0508620cae	5	3	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	14	\\x000000000000000000000020	{5115}	\\xb9a7c483dce8c293e41a80f6ad206cfdfe5efbfcccf7c04c9dd886b24ab78beba8de030b610edb298774e3a8d1c8b24514140ce400f656944e07558b404c41d1481fe3db4cea254433f51ffd138a962001322cadd94d5e6806fdefefd77bcf83	3	3	\\x4f042e61cbf74a5064584f761951679c92844d8b314c97a37a516fba651cda16	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	15	\\x00000020	{12959}	\\x8848f88598b9d4237f78c20ecb5a0fdf4d67d5e1324ca8a4eeb13b850478378672a7f07558d3bb5cef4723f42ea2d0480aa3e505731d8281354862030883d5d7a9985b8be8e738a09fd60d8f09e0dd8d337cb65b9ae9c18241195295ec093fe7	2	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	16	\\x0000000000000020	{9202}	\\xb8a16f8a655f01a4e1e516b8b0011e0eb57ee5466efce35aee2ce95908ce5e5ee162be471fa24a9684b527fef925b02509a7f5ce840cff75fde89b5d0ebf0f73dcb1ebecd2149ee1b86d4903ce9df0d82dfa64e46295fa3e9b382dc975748e2e	4	1	\\x35b622b5f42154f9294fd88e95cf5916391d530658efc6534cdb64c5a96a1cad	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
6	17	\\x000000000000000080	{15585}	\\xb36e95ed2558cca3868a2da5a59ea45059657dfac710b29c3f170a609dc1823ae474396e1cae9905e979ddb9d34316cb168d16cc17adbfc0322c77b966dac67df9b91b2c7c3dcae961b8be8543aeb6b267907cb0eec3bc70899fb5621c39389b	0	1	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	0	\\x100001010000000040200100000a000842002418	{10789,8800,11647,7091,11220,8294,8334,8715,11112,17134,10791,1760,10738,11294,8278}	\\x83463d23bd43e3945f4b388075687b4447f24af761683b6eafd619b94a192c3561fbb46b467c1024218be62e9c9fcb7e18284b01d2fbc34804e9e9ae498429b0c81eec8f93d075dc90f409b2f1bc13a3bf030cd3c33176946ec3c0fec9c269a0	6	1	\\xd346e6e92200d0ac2f9d5d946a79791b0c3621baaf6e5b20e2772bc7bdb410cb	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	1	\\x2f837ef2e4a813b9b39a32f3bcfc24f007585d03	{14501,16604,9311,8653,7045,9321,4612,2174,19185,3323,1788,12102,8229,2605,2838,9466,12276,3604,3698,14527,3113,18252,17342,3411,17225,4690,14119,18730,4837,14288,4593,3652,11047,18568,5272,2585,1582,3908,3933,2744,2803,13942,17937,1102,18788,9449,12224,3526,3244,4462,16009,19846,4777,2090,2043,4581,3388,1733,8715,17857,8507,17122,13670,2446,2149,2655,17854,8435,14480,17383,17134,12504,2753,3490,1729,19346,1760,3981,2551,20030,5111,15080}	\\x8bfaf0b7dbd51f9b59eabfad6c11b5c9b015c4a7a770867e50862d79a60a2e05d272816fe2f0a43489f46cc4e636742b1666912baf96bd34f25d88a701f914eec1e209aeb4466a4f35cc8fff1deb50d8b158fe6e56e2c861f0acc8ecb782f312	6	1	\\xd346e6e92200d0ac2f9d5d946a79791b0c3621baaf6e5b20e2772bc7bdb410cb	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	2	\\x770822bfce6bb4f2495572eb0c0e1b200ddf4709	{9118,4410,1242,4580,5047,12079,5444,18831,3952,11058,17103,8528,17224,8751,3899,17949,15950,2703,18340,3718,17640,17322,3168,5277,12074,15094,17168,17877,14347,4835,4621,3916,5258,2078,3342,18943,1477,3930,2037,5226,18782,4385,4893,12564,1962,17125,3500,12081,14138,3027,1483,13940,4444,18793,2355,5441,19218,19841,1828,14137,3215,9190,10668,12309,8533,14257,10432,4635,4662,4609,2379,18747,4101,3898,19089,3749,3684,9062}	\\xadea8d72267f8afdafd4c6e26cd09615320a57ce5209d824740d497d8074629647bb5450865d5c86ae4126bdd461b4271001574344d5ead415010347f8164ba746c258f4dd213e70a131fe0f95a88341c9b0417ae07e4d682551bd92df220fd5	5	0	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	3	\\x800400000002000000100a428a00000000000510	{6956,10332,10056,10349,11295,9824,11598,7077,11644,11477,10292,10197,8807,10990}	\\xaaea9ffd0ba4b8789d24f23912bbc85c333c36e30ed82f9dea047e609b7d21dcef4dd5542a26eca8bf488c92811ca12a01a5872d49ac079fd882cf5b8a252024c6273b7b972644c5d68ae3bb6bc1912542a49662a83089be268e5207391e4667	6	0	\\xd346e6e92200d0ac2f9d5d946a79791b0c3621baaf6e5b20e2772bc7bdb410cb	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	4	\\xb94e5edca2232300d3ace5d91aac849b6536861d	{6742,3712,19176,3619,6956,18902,10332,1431,18566,14540,15236,8371,15924,9697,1444,3866,4436,19120,13455,17094,2452,2133,17602,10056,16956,4576,3478,10850,17697,17391,11313,1870,9845,19712,16002,13972,16541,1232,2680,3468,17304,1931,1304,2716,3275,7077,17371,11644,11477,19689,8460,4083,18665,14483,19620,14048,9638,3647,1260,5188,17194,1822,5451,3156,18699,1559,14125,9441,8849,14466,8807,19733,3510,17179,9274,10990}	\\x8cd056bd57ef740aada34dcacdee505500bcd4fb5f81e1ed203473f7ff44936d1e9f9771cabe570579cb07897a3659f00a772e3797013603e933b19411e8ce9da8ffb8e59f9b26c0545831304239a5e96115505eaf84628ac280b187587c51af	6	0	\\xd346e6e92200d0ac2f9d5d946a79791b0c3621baaf6e5b20e2772bc7bdb410cb	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	5	\\x27bf489d18a3664d81166adbc994d7ba67b56513	{19850,9939,14317,3365,8348,1139,17523,11600,8659,5631,19915,15336,2542,14629,12254,2176,9481,10395,14720,8452,11206,10814,9110,13875,19043,4827,14997,1538,15007,2129,13594,11541,7298,8738,11510,15204,3678,6873,10310,2344,2029,8712,18597,2584,4039,6784,18521,19632,8482,3118,2309,2742,8632,1048,5460,2291,8359,10550,9130,2373,19754,3295,18858,18476,14606,15250,14618,10393,2678,12282,11143,1651,1291,10732,14610,18701,9714,1978,9163,2841,18273,20022}	\\xb1112be308bb24748c1d8cc0dcfd3c5ad487d4dfc2e08c9198c6c4173400a63ee32377071d62b18db661f205597024450ae0ae584bab46c78dfc558cf6bb6bf8d3b1ea373e4b67270a598d41a7276a4ff38d19812275c4cae494c5f4df349764	6	3	\\xd346e6e92200d0ac2f9d5d946a79791b0c3621baaf6e5b20e2772bc7bdb410cb	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	6	\\xbe2804f8d693b0b2b946d890013ed3b212d90704	{3634,9000,17513,14560,16507,8991,14586,14636,14144,2152,19253,1644,19666,9135,3615,9706,16137,8375,1741,2740,3054,15308,14099,17250,5343,18670,2440,8848,12133,17724,14569,3231,9730,19236,15083,14360,3214,3832,5249,4846,1056,1996,2317,18124,12101,8434,3788,9224,17453,15179,14271,12273,3217,15189,18685,2282,16936,3296,1180,18582,18626,13827,8200,3543,3902,5273,3241,1999,2304,9470}	\\x91b86561b18d9922263659cff0f384dcb61a020d95206b4db027a1f60420a391881652c2793026b007d2888f120c64b207ac543a092fdd72d8f3f4c15f3e6dcf9be5b2fb85cd859fd141155146967a2a528351d58f033be363e76a3fac1d8473	5	3	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
7	7	\\x55d29393a33fc72157b6d8d5d725abbf15c6020d	{13959,3801,12126,1758,1069,14226,14035,12308,1276,2715,1627,14680,8929,5255,14286,4818,18242,2224,10819,14211,3968,8324,9743,18874,1337,10654,5244,1271,14800,12115,3201,4021,14996,13870,8693,1363,19030,9519,8549,2508,16005,3368,2741,20034,2289,2028,18917,14536,17714,3539,9122,19235,10756,8803,9367,3445,12113,19700,17878,3346,17836,10398,8280,1762,17271,17811,10722,14719,11506,15879,3243,8376,19271,10228,3560,2771,10040,18612,8886,13865,18580,18797,1740,2676}	\\xb44d50bef0f1ca19f00b164155ff8eab5b1ceeebb9da7ddbb82ba5e08f86f6bccf55c03a935789e3ad25d62aea23eda50e85e2c068e06fda38d94013f5a2f99dc399a6f3dafbdd866e877f7229d0df885e48f963a3b08628534dd9dcbe0e78ad	5	1	\\xa24292649aa34a863c5c16a40e92b6e508facb751486ccb9794fc9459420d7f7	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xebe49843e96e538592c9e1976e155e9057c251ed9312eba517afce3ca1a0660c
\.
COPY public.blocks_attesterslashings FROM stdin;
\.
COPY public.blocks_deposits FROM stdin;
0	0	\N	\\x8fcf28896a85e5e76ee9e508438e23e7253da1a23a6501e3a7d56182520dbcf4cdb44af3267318188f1f4168342146da	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x9800d7c29908ddd1754490577753e65bd27484ca456b33b914e937416d58a6dd5b4eb420c22c144ae46f203fcc869e5c18f4e6daf99122294dc0ea691242625e7b75dfedf5d799899f44cf648245ccfc9652a957ae9216b022567c4cc4a72066
0	1	\N	\\x873e73ee8b3e4fcf1d2fb0f1036ba996ac9910b5b348f6438b5f8ef50857d4da9075d0218a9d1b99a9eae235a39703e1	\\x00b8cdcf79ba7e74300a07e9d8f8121dd0d8dd11dcfd6d3f2807c45b426ac968	64000000000	\\xa6aef7016df5e73c2c142e3f7a0a0a4c85283a6e8fa3c2f52093aa029d04560ebdfebd4101d3a1bab0ff2812bbe1e51c13621b9366c83c63a081cbcaa3d500da9af1cf05b9e2c0bd7feffab31cbe6a158655f23b124f4f2fea746e8f93d5a1f1
0	2	\N	\\x8c2f535d3bec65f95cb4ba455566e4ec3de8da5c13a681699e0f80d7942d6fdcbcef18c8cf18f9da14aa379bdd6d29c5	\\x006490500934b8b1876401dc09b7904d04c0897a9a28ecde4ddb1a60fd5c4c60	32000000000	\\x8ac5120c07c2202fbc31ae2e5dc9e737cfa4056344f908ab66223bae755227f642efeb9d7974eacb2ecb1dfef4021b79068aaf9ce1d0b42ddaa136fdfcf8b29d6a2fbb06259d0f542bb35a4658908c14e893baee1170aa22bb98a78f98983264
0	3	\N	\\xa8d9b5b62cc31149ad58a281a2293cd3f4dca11855c98983e76ffb60479d8e98e5592e5415f0400a8a23efdd842b3605	\\x002f5bc32089b840a516c800c3c597bff24536c6c7c9f1113c2a0ac847608f91	32000000000	\\x84c256474e1a1f1b902b256485ef5881432847a3e14bfc6f5a48e250404e9f85737cecf1bb1cdf3b73cb0a1f38117dbf01d045896c2e8f918e4f5e9c985b0b60708c459a6c7618200f68c3a65e6b86b9d82dcf66cfce6419ae775f3c56a9277b
0	4	\N	\\xadf943279435f1c194add1cdfe99e3fde5284d0451a63822b03ec301bb1cab4399d016f812461f0b71a8206b96ca3378	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8acb8c35711572f0729e71d79a2b24598017a20ced75680c6688d253fc5a7dc4e3c7b2c5002e247c810eb77f47c75bfe0fa2dbbd6b0ad0078274773d8196902db0529ba23fed23b61315c0fd09aaa64fcf335c8c7c6d802dcb08b4e4a2a7372c
0	5	\N	\\x85f2c3045f02ac7b8b235b9aa855be7749d52de8b8e9855db5a9e97cf101185c05ebfa5d3d3dc2783e009294525ced20	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa555d0136508b3ee62c822a4e6f0f8d87679e74e2bc06a57fb279a4d5a8ccf683e3692a480cea1812862752929c1e708181dac71164f0a1cf09bf7a2d979de2519d93ffd6413d53ea0f335418f156f89fad725b3003809f33722bcd98d9199a8
0	6	\N	\\x9274b8d712b53c71dfb2d8762334d63ec9fb18fa6278aaf47b2782b14447539b6e5f56a9836e6ba51e4a206f1a8dbe02	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x888123b3de4fa234d9b529628ff6f097a58c45d065fb6d15358c0d47fca500bdd2fd1823be549378c40dd7ad947b49a415ec15f3cf996f0f1f5532903d11b8e5bc85607e76c9dc1c23cb37a427eafef7c4ce2594061c5f6cba6561d7367f27af
0	7	\N	\\xb4fc18feaa072d37538e1fb8157a365bcffd086de9d335dca2eaab1508fb3a48504b8b982d05919f7989d818e88dd07b	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa00beb954f714a5792f7b3df495965378d497c36b27c09aa8faac0e113e36c70591a5ca2ef8a6ab3662ab1d930ed50fc011f49ee23aba8c354112151cddd652ab68a27f355d51b1e1438729eb0d0a12dc5e9adbb91a580560917377d72d3155d
0	8	\N	\\x9939f64a7b916476b076abd67cad897991917d4b7da38487597696844a890cb5156c2928ebf97963e77d5d2099ca881d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xad14254a0ff7f77bea8f71a259d8e56635898aae0594c460bc7b2cc06f8957ec64f02450382197becfdef2ed5e91278c126adb1f3f332b11b4829c989bf825aa8ec1b7f7bfcbf7785fce494b5e9586ec40f1841886c53340e51f30331c97f125
0	9	\N	\\x8e368a304c4c564617c3fa265361fe679df8160e6af06c5f5dba57d8ad134e7073a7ad631459116aef53a788e3c256b3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x86f0d7a21e189bea04aa447eb9237bb5a246cabc31fdfb2d9aa8a9b63f2f331014007008b12c5167ce399b3f9949f7051931ee1beb287e88987beaaccc6ea7aa3b77c7dc749a47f1ef596b52f50c28c8ec7fc2c322ebc23944e10d236ca054fa
0	10	\N	\\x81903c25cba2b6f37f23449d731dc6978222a9df18ea08cba0a58c7d43d4b117e2ada338385fb1514b1692a5ca4881fd	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x826eceb6ad2c191a8bfbdef905b7f62cab64d180ab85c8f84495b594a5a6cd7aa4afe6c658ab55f3d42482f91309afa80231afe401c7c770245bafe432d9bbf87639827aec7a5581cb0c5bc39561179ac63641928db30daa06bdff556c74aad5
0	11	\N	\\x8687275aab69c59d56efa991c5db0b56600912ea03b6b4ff3e297f0a62434791c479d415aa494bd0b41a1604b3de0c07	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb0bdd47ccdff4b716e9482d652e9d52e286ccce7859e590aba7c79f954817ac80a34494632debf0532fd63afcfce5fb100fa1e83d2a4c1853b557412a7370a709239f3a2339c65fd7ffdbbf8018fafc1f5c958eea292233992f380b58c797528
0	12	\N	\\x8725e7a24df271ffe645fe2a4fa46bae935666d4dc654dbf77e4853ce70229828ea1b60c49ef4644975cf2b466ed8674	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x80490aafac86089d554d2bbba453d2284257dd59b7a2a36cca84092ea981e443200f1c36fd32a3b78a5eeddf8305c21f10e2b9301b54d88cddaf6151a30afa39ba57968b77f335ee0dbe1d9215f858eb178e83cd210c5f6b5a5e9309556940ed
0	13	\N	\\x9414c11a0e3b3712d0f62b7900dded64dc9c1c556c40e4dc40c0a7f5ceaec02135d45858aff36ef50b433e6fe24b6e39	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaed65f2f7a36851e0ccfecc6e0cb865bc22cb3bd9fee5f6da5d467a91583efb9b205f0fcc115e120bcd789ce74ddd55706de0174c60ea6fdeda77b69d325d5dd0bc0bc0f22baf49762c60a985a7b13b5bc3e3e61ccbf7f82d56199da5bb617a6
0	14	\N	\\x90d42bdd0b5a7b62f665f75790739233b3a700f5f8ced95233fadcc72ca16977c8d6fb00b63d08ef750a88fc0ffadd6d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb378ac392f3916296d53318981035f600e329fdc2c550e0d2b12730f5a66e79306cfd162522666042e14675a72080df20a43b44d983a1deb3d1b1d711996a1a5dd03b3c8397df0bf1335ccd1580c9d62b73269045ca3fb3886817aa7c7e7ac9d
0	15	\N	\\x98573c25eb271690db29113fc890175213b1044bcd68e82f413bc694743ea265ca5eb8c50ebe2ae898426c0c9de650b3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa3e7fe6eb0f300703b4bb5c2282bba09e84c82b709edc01a77d782da69eba36ecf8c20063361f02f484167b218a532fd02c9af6fbb6b362534d945370278769f5fc622105099ba168cb0586b42743f1a5a2caf2ab9cb8b7b11dc4da3c638497b
0	16	\N	\\x88048df033f7eb4df024ce81b29459adfa92f8c447e732ff6b77f51c5e721b365dad73eb25c1617c11d456ed916cb935	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb6a5975d4200f2232f78c1f68fd0e313d9c00395ea64fe937a7f7af86fecae955dc8e8cde5ccd81df689faaa2cce8201033df473f43d952cf989a50aefc13a0d056ae8374a91ea7f9b155f0f0fb0375c68338d3e24e876b6c95e35900f53a2f3
0	17	\N	\\x98b0d39db01e3bbd6b750c43c26f7078581a326a9778760db3b72b22b74cf1b9a2025f4347f04bbd36494ec5bfd8547b	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x91e2324ce96b619ad46ee00d4ecbbef08f2700da307e0b84da2a3247850a4bff533cdc568d3b7a3b57163abd5690aa0b15b854814e34088be81b2d94a5028704cb755042e540d5f81488d10fee0ea57f8d3d2b4f6d2c9f6339fb7ef0d78d178e
0	18	\N	\\x8a290fd6ea767704eecc68079e64cb9b43915652ebf9ba619b811149aefa069718dd97f8775004845ea9197f8c1cbee9	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x9785ef99e316c6c877fc2fdc6042533d230b292aea91e359c3b48e05a7677cbf8e0e47c0f1a8bc7a5a56880d821f2539173bb095a687edee5fd4030715642fc8d6ef16278a407cc410c24dc2be665dfbbe999aef6dcc8af9b64e00cef66b1d19
0	19	\N	\\x94ce62f60d60f9d52f2d9456b9f22edbe7846fff4b98f4316466910bae8065c86eb0756869d5cc27427f098493bc7677	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb8b780c266f01f5efb556808c6f8fc862763ab82e8ab8ab9ab98ab06755d281f5a2e4f3151c726b2e5f8b17e9b79457f1655d836ba3c1aed0fd0ceead38709d377dbdaf1362bd33d287189a765636e0f5d7c593af4580219d2299c6798267dbf
0	20	\N	\\xab730f62ce3c14945e245b08b029dbf579d376cacaf82e5ae9f2b28fa72eec68e33103c693c502209a532edca594e0a3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x80c26cea1953be6a8879dac3c6c0d4798a72b5b27ac9eaa83e55fb66cf6ae5456f786f2925afbd35930e1f5fcc06074f08485fab890d03feb89cf30ada6c26a1cfb2c361f52f287c77adfc997a7975c80570c836164e7f78665882c8df01c8ff
0	21	\N	\\xa898422079b7db772520e7b43ae7c36bdba9f1594fb6b38f69a5d46c1c4f9fbd1ee738abe7badc786a68a9d6dfd37122	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x91774a402546f905387b12249640e0099904a16cedbc37f48b1eae2143a23db93118d55895cce87e0ead86e15bc12b560b612de5081aedbb3a183d92f85280e6dacbf8e626158aa5b6dee3f9979f0b859767936c787555332f33d2dcbf6808d3
0	22	\N	\\xad547ad6f3fb353aa0c664e4c770560b1809cd934aa356d27c0ceb41622363d9e607a90c3f206dd954f18040a336c20f	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb7537ef0b2ce874972f78176a009d2a201b999d6405ef68c45e53f50cd29cf1f2b0b83b7b1bd699ff01303de663cc965167aa9099373083fec5380dccbd8942e73ae3e4e9181d5babe61b01209c0702538219b29fd2d9cdb3e9a56c2229cd4c5
0	23	\N	\\xb05fbe5c06ff586d649d8b21e9f75b37bf430fe04eb712850a8af2fc3455b5bd5051a258cde20aa2d59da5a0f0564ac1	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8eb7a4cca039ae6acfa8a87e38e991180f912497da94b831d41278e3b15d73b1528d5bb8595f0508541c9777c5f9d8f514e8b0362ae6673673dade2ebffb3dcf9a50cc8d46f9df4efc05ae955c9509a2665e8b9b7df50d83f0bea460426bbeb5
0	24	\N	\\x8c18668cf8856f3bac3912161df97a6bb9b98141f899be6f3762935542d7857e160c506f31757b454fb3113315e7b9be	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb45f39882e66f58f6dad694960ec2fc66118aec1f92deb0ff58614f1be1560e9769382bc1fb212f30b8eb75797de4f5a0f1777b1588998636493bf72f844ca82bb6b09105b11b1be3401dae55f541a30a17341f3247608d5e38152d0f17ace2c
0	25	\N	\\xac605f3f4df79840e1e948df5d67e8484f4c3cddc0d6e561bd7bb6e48a8a1c9baf154c316b1b3d7f4ee23ca7a5cbf7f7	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x9776aa6b6e3ffa6822b881f55bd2f2a02201246e36cdae5c5aa2dc2ad8204300f3ff04e9809febe723a28ce2ee4325fd0425275a0c446830a0052bd3f0934f5fa1a77e32bd40505dabf092c92b467b5e57cde38701dc5dac8ee6506871cc6e2c
0	26	\N	\\xaf6612098003d793949804ced355362d9da566cf2fd00059de7321241754d8aadb0b77c94587ed7695dc3b1b51d3dbd7	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x881de0a9619d5750ce6fb7f9e85726c41424d4af46a99f87c5d5d0bfeefa71fd207f85defc052d23c903fb56fe785f6d1330f97d390688a92f44fa6b3cc7c4037a99a84f571fff0e4d9e108e34842b1b85ae46c9f7fc91b8089f77f33b0654ad
0	27	\N	\\x8618f2f494a8abaa9915c87b35f3c3b2899972f6ec1e3696dd54aefebd3fbb48c3b4a02bdf40e17a437c3adc9f0804aa	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb034e08f52ddae2b7c6ebcfba64d72d4a38a8b2e987bc54027b598e5b8760027b4a9100ac64fac8dfa0b73a488e2f69c0e4a32002eb172354d900795aaf68ddfd2d339e53adf9e5b2a5335a8c9aede40dae05562b5ffa0c10b61a3cbe6523390
0	28	\N	\\x83251cd9095f5fae3b2dfb006b16545d34bdaf1894687f2740300e9f8f6229eda1db3767c80a38fb3dbd8118ab415c77	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x84ff95d9dfc223e2d52885a0738c1913bab11c634c640c3dd86cec6ad1895601a78e4df23d6927e3d7d62ff0e46d5a7401416982e07f38f78451034cad4bbe3afc8f8878f19a79aa5992b284f7cffed8585749b72be196d21eeff635717dbba6
0	29	\N	\\x87d8e3f68aa7cf2433219e3cdbe6232aab611358f6f4235794ae4225d24e22578a71b52847d459531ec081b4d148afe0	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x85b146ffa63b150eb5a40904c4d48fce777710984aca9755cb5044b4bfce12dc170dd509aa3e979d735b7cf53fef2c2b0ff0a5833597289aa590ba3d3481a2d99532c4bdfb16bb3148f04b06bbf9622a3252ac457c628b904fe07c83c11792aa
0	30	\N	\\x92705f79d60b3ae862ece24de3bfd234aa24ebf2be5f4235ffa72dcd4b882fefe1753d82ad355188bf22fa251e05e82d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa7bf258efd8d9addb26aca8ae9b6672fbcfad280bd1404ea836426b5725bc64c337fd7836c39acb7180e4cd165685e8a033e12f84ff3a2bb8104ed86717a726fff789581a0d0da38a74cdb7ed92653a8f403e6b69d648e0061c86d896b8d8f34
0	31	\N	\\x99ef4f94bc863fcf2c393e03cbad1bfeaea19eece05a3084897bf6a39705eda7a40434bd76a01e6a170bcd319a445b9c	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa8779d81c50ed108ec91621dd4994de2ba2ef1d75beceb0b8ae9e4d755f86563d500844ac6be04bccd6ea83ffb22556016772b23f39fbd46fb737578e2e4c358e5733af4e32d09919b0d80285ab7ae7cd0afac008dd180ea1c43965ee19d42fa
0	32	\N	\\xa6fd2c4289250d399876550d0b9fcda435f41a22c24e4c8b0e25c0cee8042a0c9d3f7c02928403b36e22a4a3d7ea8bd0	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb9c783c375e8419cfc6c2663fbd2e9a97eabc87923b818e3ee022dba2a82040860eb2817bbc25380f86490c4f4a701f70e2d75a357926115fc149a008e59d6b937e42f060bd48c078d93f4f95d1836bbfef723ec555d89c6dac15298d7214cf4
0	33	\N	\\x92c83a2b73b866ae670c423ea00194d856bdf552cf1c9f8aa6ea92929f73d52dae530e65f61ee0f558df022afa5f12f8	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8286606a92686ee51e6b45d66d0760100415ff264949033d07d9a270cece926c90398313741a9354ec81361a3196ad5f0990123f0f3e528ba935dbc36bfd4a287be409ddcdabed24c210191aba1630f27856ebcc1a04336649c13c362fc398d5
0	34	\N	\\x98250d9ee1c76975aaf3a1d68dab0f23a203eca5d6878c1f37c9c1ed2b97d62b1eccb9979097089dcb817ad0a84ccfae	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb4e670f85ec59c8dfca2e3e75ae1c22d2dba0eb5ee34ea4ff42ce5638f4c5f8eac7a54034c4122dfe8133c0e98e991900d8524cc7da0f74f75768a2ea620c114964933f285c8a48098fa8470561d565df754e3ad413f9796194a756e9ce9c493
0	35	\N	\\xa27ede36730c6f4d5e089629b4ea7ee1d17b037c25892e41c49175f831206d14c8e633f397fd7d6d39114cb5c6e51821	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaed3cc411402651dbdf71b389210426a401620e1db2743532bfee8f12d2ae6485dcb93a18e57e43265b8defd03447d020380f532a99103edfce88775c80352c02c2a9f41b2b9487d56d8774c15808324f5861184cc91d92fa46a674deae9fe17
0	36	\N	\\x92cfbbf8217e197d73492f838cc381b58c39eb68becc1e9cacc76d5392c94148ebb27d38a2efb1d9cca8673a9cfc1a10	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x832edfba013abd5fbaeabc482d1b4a80a7b2d37d779bc8ca8bec909e8545358e160b28824d1411ab728719e9ae6d35d411f381c293ee7a00f75da69913401ff0c6b8bd1b7f3591594c600f02cf1588b26feb157c6666fafd41c7f563aa4356eb
0	37	\N	\\x80b81131547bc13b96f1a30ca884c297050acec603c201ac364fa7d7ce995828e3a2e26eb35e0cee569ea6399b4bd055	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xadc992e8d9f933b450d85ceea8603b229d2f88ca9ae73aa111f2cc8cbf3ff05307c0e518543a849b21e9721d03b3d53410ee03553d05fcbdb2c42b8702750910ad9b5b79874f99f94564f02c7be43ac1a31855193ba77b992c548a05889f384b
0	38	\N	\\xb30dfc42bdd14ffbf8d4315c666fdbf926409b4f316373a318bed26a956299a4372154e0d7a9113c1273c04a8777cf5d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xad47fab3a30a64428eee0a8535b57a85b331cb9a4ecf8164c26f143f354d1aa79332e6f22505865c0d0bac53b9733e820b4be83295a0c78e74ddb287d9b621355245d159fdd486f00fafe77b16d4852b3383086d4737061a6adc609d39087f76
0	39	\N	\\xa713c85d5be62a7ea34e67054e261dfb700794a830e2380c5e5f71dcf29275345926192e0bbc6877dc1ea03f8a9192e3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8885e8eda194488f2a432b4b9fc1a481dd58c22095d85f721d79cc743802c5f3287aed3301463aea8dfb215fba3ceb1503a1ee6c2236ff3c0cea8f90a537777aa9217526c2bd272ed79390e36f74859229f529aa3a9fb33f0fedeb46db2ce43f
0	40	\N	\\x95d513534de5fc40c03e1e657f1632e17f2ae65e708aa6a25c13ebbebbda3dbcbdd762bddf1c5fa5c6fda289d9ab6406	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8395836ec9b8c84d390ea38051a55c0c383b56d34f1f883d462ba4aafca5bc60cdd142fd737b4cbc73d01210099e4fee1655f790fcf33cbfa101484725b156e454f8b00369a0ecd57a1d5f441c599ff65b9aba2fb220acb27975ac4470ac70b0
0	41	\N	\\x8b37ebd467bd81b891774c695721e6c78d99111be938ff272c72cfc864683c3e1ff5139c4824199fe29d0520cbd8c874	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaf6cc9695c4c19c8345430ae3f66cfb007cb4b11ec59e021a059e755d26b2d17faea338c4a7c63d89debfdebc930b92d100fd83c8eccc0de10d2c8676d9a449050e38c09f9a7632e419f3e80dd797619d85351e1b8cd5e1d0df94019deea0474
0	42	\N	\\x8371247ae81e360ea915f7408a75fb732eee95885dfa7eb946f790293b4038ca0ad6be6cc26e31247f1cfd98a060c9f7	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8b2391daa92bfb1440fe60a3b32e55225f5cddd01f9b34867c4bffde0b75d5831d301a1c6eb84a797006c084f8c8469107c582fddb0dfcd0516036f0505953621c33ca2a56d3985bfd99d57cd0e4cbb1ca74ff35167a45aa75926e377ffe09b9
0	43	\N	\\x92bb2a1e7c4e984d7896b8e6468748fc654f045ffed539d983d6eab198caaafbfd7970c8d7e1cb768d7180091d30dc0d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8112f576f552fa5205de485cf877bd0e19266ebfc3d2b35012de257fa57b700e2ec5cc1b03273a2d27fb4947ad3a523213d97fe40224d287f01ba230ab9f98c7b50f8178e06b738066f0026454f2f38e8bc013d334a8a7942b2534670d65e1fc
0	44	\N	\\x8023022f1f335337748eb141477e789d683514ae739b9dac7b72bb3326201065abbf6f3301a5185bb8751776e5bf45cf	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb7a08b5a72761f23dec2605fdf980d2b9375b67a2cf74c3ce953d89c94cb8a797533de9d79b280b349ba070ba82a4308149b623e14428fe55b0bf9dea4f68dc887475d7c2c094e14a29368c07bcddd93ec50373a6e0d1ac3790ce56a66e19b25
0	45	\N	\\xb1e8517c990087ec06de1c80910b5f5a4dca67f73451b61063d575e567da0cbb7ae65d2a4246ed1d176b21e896a4a4f8	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb1d765e6726a92c4502678e5cdd2d51a0f5033d1a944bb03a2a118fa16e93e4675e8b61ddf66eae47c19b2d53cdea47f0b9e5403f996eb3e3c47c9335d04bdbb1cbf6548e2ca70dafeb49201e3e25ffdd492a82ee48655746845edbe8bb11903
0	46	\N	\\xb179dada48cc4f390cde0048047f5055b4cbb49437c4ed31d6f66e5db5079715cdb9d199ff13ac0521c907644939db19	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x86a557d4f19e5c2351609e25c5f7b7b253251b1c91dd192a41eada12065b8df7d57f8d74c2f298126ae1f76beed07bed02fe063b4ece95580346c35f391d2d84093777b967b11c184cacae1d2535817135fcf2c5b2748672ba6348161fb7768f
0	47	\N	\\x9264398974c228e6449b253f97af9b48e463615165b7f2a3272cd60ecb111f89aff99c99fe1aa6f14a999003420ca88f	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaa4eaca896fde56694f7efc830353a9c9a3020386d0caa1c75427aa08792126b212cad6a70db9fb98e330cc6442e0fee131d891d0e6737f74c05e0451e04b751911ef0c13c7be884c665d6aeb7051009bdec8fc9608d38248b05218398ba48c7
0	48	\N	\\xa80c96e1e93c8cc968dc639c25a0c7ae3bdde02f96b61ef2b7b3b0b1ce81c6bca2de60594beb0bde3b1189af504468de	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x81a41cc98377ff812c70fefd368f632f86b531d57d5c4992696f969968d469ea1773a1855dd43564377eae3fc1b9a4240ca56efd7631adc3875d134c7f276262f452ccd5207cd53176f11e9b91a72b9fd2fc85049c2e6378e863c2393decb678
0	49	\N	\\xacb7d7184b3c9a6a07e1364968b847262a758d08264746fc21881259b06f8bda1b5956a05274f07eb4a650bd123b7e60	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x89361aa59baf22879710deaef415a9c5fca0c9da133915400d7b11365a364ef5abca961c7e2abb51820ed9cb931937e70b2285a167827152fd1e48c1aa338777fabc72a79df511b57299dc5ce0a5b5aa0b5cae8e116e3a8f5a2fc86676a5ebcf
0	50	\N	\\x8f1bd3e33f74091106eea31c36dc916879ac5f1354c58787bd0188208fb8383dbbe12e6c0c273d28899d50f96c5f4f2f	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x9787f00d38f5a23199915460697cce972d72e0b883fcabd0c330d1d5f4cb31c9055903ca5792bd514da6b3f65d3af28e0d418fa52a786c33d01872ef2907953cc0ce4768e182884b9fb6a8396bc5e0901754696719f9752af40b6e888b395ea9
0	51	\N	\\xb585a904d9258b05ac280c1f92ec59f3b6e0e2fd456556ed6bcecb09fee79efef6c277e589156ee27f5d19c6af0d0cda	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb6a7ca860bd211eb6569ff1a9ec9c062318c8aab980eb6c30aa1844a82afe3c3507e10ea9dd86bb288e5110e65fab9e209db7ea6e5e3bf82fa5e38b1f252d1f37b9614749153731dbf70bb4dfca53caebffa5665bb62415987c67be9b4f1c2ba
0	52	\N	\\x9070ba2c2247de494ec37e99ebff0a634dac8b689aaa63c1901af20e18244a79d5db1b67e9ec1b350bc14800f671c601	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb5f757c0e8496dd2959b76e7e51913ec95c950161840488e0a16268870c6c2430efb34bc698998dea0831e14a42525c508d7a736676964fb241fcdb0f3d105460440087b030604afa3ebb437231dbe66e6e7d769db12e59529bc38ea24fab0ef
0	53	\N	\\xa1a87efcdd48f52a92b53a6b0488fcd274a389265026f0fe1707564f64ca0cd808beb12d246548b3412ba0e36ecbe551	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8bb0538a82eaa3b2ac701278b2275c953d1650184b2a6a10f3fe77ee066d2692f4ce962319dbfdc9d4480a8d5f27933e12bf008a4ac9dc4fcbd71b9026d1bbd373c67ae92a2959538e8178e0963be386ffd29aa52f74130f70467d61aa031999
0	54	\N	\\x80201089ba0e64c9c1f31b54b6d18179b9044ffaff7aae4fd1deeab691eb3d6ae3aad0a6087cdeda9eef8d54a8339116	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x976c56a4f76dd5dc9aa8523c42c5aa74d8ab67ca76e404630a88544b76bdc523ca30f127d7627a679f6ed465d2c34424004c1e4f006185e42e7724da24145f3b3778079372f643f8d85a81747c26b8cf065f0bd11734203d00de25c3d10335d4
0	55	\N	\\x82b0104dd9dccb64bf6ac783e33894530b4135b98d9aa150053d8df1e0091a9646a133c0392480c60bd623c42af41703	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa7a1cc84f617a516ee8f58a3298caec783732134d2023163dbd6184776bb8825a45af01a1e62fcd935d9358620bd51fd184b14fe2648500db425a9cf36785adc41162d3ba5bd2832627978c2156fa9ccafbfb0f7619a775d33e8f4a3bf0f2428
0	56	\N	\\xa6a4eda250cf96eb2a9a708aaf0b079a93564c72b40e6fa5914494f2f021b914e7aaef94f6931d116ac561d3ff6a796c	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb7b60d6e142ec10c4b7317f09cc1d597eeda2b82895ae3b6764734d81515380a8f04797920e355329e9e55b0b0b1df9409efaca44aa6022f83cd7ee7cd1f73d51414e98ae7864d812b0e17c476e76bc0ad334fdb1bf0bafba8dc2887561e81fd
0	57	\N	\\xa2c4aa707db0f09f650aacd3f45ab4f4a946e69d24fb12ec14e387439e799bbbc7c3043cad010ce3bc3c65334519003f	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x839612f7f7e0f63fd26fdf83770f019836b10d67e93dfb3aa8409b8507eb0e867b39dbf921bb5c3c95acdb2f1dce51a600c0bdb9c63f5d5a55623a2090bc61b979693c44886512cd84dc522da47a6862d3b2dadb04e2bf23e320b727455db4f5
0	58	\N	\\xb868263a518c6746890eda7fd062d66d444a9896160a59839e681360d7631cf19ffbc6e0f7304a87b3f4b9f8122b6f60	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa19731a8e71e0c8a4c077c18b99101c6428d6e4c3914cf69084ee5068dd07b59fc84847a8f3beffb54b38ad1dc6f710701227e7f3f783712da391dc466ba5070c3b9556824aae7fa533475e92bb0bcf0933052a4dab668bc5ab7a05f82ac4aad
0	59	\N	\\xb7306789767ec9e8cfef3aced2cca73986af30e30c4a6260df4ffe8fdcb0be469db93dd76c980cd348cf58a71a686aa0	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb229060aa1f11cf58f35c9c258145ead1e39f1832f6e967720e5b197cb6ae96d9a048a01d48fa587ce4c6316d1afd873040b7efc2d54610c5dff71f6e48fc965ed5ffb377bc4357f02469bc446e4b2e356bf27c5ffe291a24519a9d3c232c5d6
0	60	\N	\\x8d0c273e62eee67e9f7c84134c2784935c286ca2468bced2135c62e9d47a5e6d45d9a0d50dbbda662d336d97ee13bbfb	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x996645e1a685415aa9294cf9065b33c10674aab15f2b96b922fad99a59ecdb616c9dd16d690d8e572be645972535772e18547d20ac7cdbd412294a24c283b08693f55b43decd36cbdaf474166db5cc361a7c1999e3975232607654f78deea238
0	61	\N	\\x8996948fa486e76917c9d982cc1cc76b709dacb491a491fdf53531ad175c8c09a9b3349303681c1162feaf3d1c520c7b	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x95bcfad90503bee2620013d83c9d959469bea67fb5867b181e2b6eda1f123a53084d1868a72b667cc8f38d2940da7d1c16f8bb8e84cabf181a79a002912e519e8430daa835ed7aaf67a3430228fd98731136c0c5bef1da4bc06896f13d2f8511
0	62	\N	\\xb27c5dfbf2c744f37662473d2f5080bc699a42d36fb980cf0c9d015487ccc650db2df1de200cee5720768ee7b0114827	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb98bd4a9a6f9e7830ff246fa74ba15e84c924a6e1a0619abb7ec309b38d935dafa6282aa8227ac4ae750639b651a93d912b90b011f5276a2b8b721439b00edcc9f70438f5735ef06c752557ebb916773f96b1227db3946d32628ead41dfb2031
0	63	\N	\\x89b71fa06df462e7d2afa3297c178db3d91c14193bffc5ac69354f3f81b23e0e9c5ddcec0247cf8e059dedbcb59b3d18	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa1c4e41d3bda7f3a0cb1584399d2fe5b1540c976158322010f68099b57cc3eb7b66b5dab523d42c4d5cc3a9b77ca3cde14bd7abb946117562826e48240f57a5e7cddd60b50284a40f1a3da4a922f57c713a71da11cc8890c8ce42641758a1db2
0	64	\N	\\x8d67428727ce758f8110b07cb341acee119681b156b6a2c3081a82adea7ef14b59eba851d971144551354fbdaf7c766c	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x962d0f1a80ae6163ff889c7029a1a5434c569e5354f57cb7bf88bf7984f688b03b2ecbecaa0efe95bcfb7c2a1bd6b42e0838daf49347ce70347c8d953c1b3fe9749b07a38faea3b5a184a8fb9c17da31b2411c96802bd2c86c58b85d194473bc
0	65	\N	\\x94c85dc9e1d948c2ce4a8ad8abec088942a50236b0464ce38217b2b2da3ed946a642a2c3b40b346a5cea3ed162e2357a	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x914e5b4dae897ef6be65d79fe6f50f4981bf22be7cff9ec00a432fe9e862bd1637b92022895fb6db1975ad1e600f4b550fac4b7aaa1b0515308db838aa9321ee350204f30a096479f54b53e6cb58918c62778af47420ffd9566b9f6f765592d9
0	66	\N	\\xb69f673d93a36e95c642ea6df46ab71ca80f45e8b5b6a0a559a841b7fa9d5cec8a7ca84e8a29a18df8d3c7f4cfd164aa	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaabe3d6865ec5039421f6aabb44c9d1d71558f8541055a1f1c50f2aa9c59bed23fdf28ff733d4813c69b869d6ee82abe0d9e53348d50c8fe48d4cfb793e5e90f2facae9a8a659dd3ceb840f59d9426fff6c3ac61d88fd1d8645738f128402b94
0	67	\N	\\x94286f9563e5d95814379fb1d7cea0db5eacaf8c0600203ba03119d02487d50a51b2a7682497603822d84c5e209a2f2b	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb0d75e4d4f6500f67312852ef0a23a579472a2cd3b85903dec07927cbf3cc3aaa870efcd6e0ee46f3743187bbb5edffb025ce05f1a736eee79483aaaffbb2a5236f7cdd6b3bfb84169c7602deaf6ba3ae4a95a38b1449cffc8d60b10f9c7e9d7
0	68	\N	\\xa40fd1f45c59a7c225566d0745362beb1ae3cbcf9dea291df9f4786bdc8bfe612aaa555c8a0e209f75d9013e8c5f5883	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa459ee4b0a8a5a80290599fd35a34b04bcac27aa7731619ed541ebd8c5b73d1f2be02d7868c51e39f84d8ac0710b5b710cf3f2e6c8cbbb24dc2378f9b6660fda9d5e901d108e64ea23be1324d735fba93b0194272c314e677e5c04df748da9da
0	69	\N	\\x89f637597fc6b946d5e67167f1483c91494ca5c79f430ab8d1a6b974110adba0a9a3a9fca3f6c19330b7136f04e4c7ea	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x96b7806528ff6d2363df8f3fd13e12927171e0e5c5ebefe05d487fb6e96c56735b5250119954161bac1415434fbef390001420bb7c7da8aff2dfe571a6202679e4df18add62e80251b36ca191b2925fbe2357d1f138691f12ec73adc80790968
0	70	\N	\\xb7d983835bf3a212e6e5b18e2f983ca92ad464901b884124c5701ef1819d25346d393dd081e0429266a43903e2ffabca	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa7d231386f74f6f87a89aeac9182d7d514a61c224bdd4c0fa1f2b1c2af42f53631206423c9f34c8a2e74c755ecfd6e74059e56f07d09474c488d54268677730200e215f4296ae0f0d1843e297203907176d9713b7620717f5de4b2ec874da0ac
0	71	\N	\\x8215888ee750664dcf04c57cc3ad66f4357ea7237f2853adcc18b137b3747ebd95c96d0f357fe49d2ff549ca8d67bafc	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb85f9038d8157b61bc187912c973aae769b8eb3858ecf884f7846f98da613d2a605e2dc5da8dca8192733832c4383b9f16ef183c41a3fe7c5e22629d5cc8b56798e9e3ed799ebe49abeb6ea74aa751c045538ae743f9c782a315d7d09d7af2a6
0	72	\N	\\xa5f8c73eef1e0e2fb66cb0b920027573ebcd97c2c19c20a2b5eb2fd864904669723735997a829c7ec02ab836702a2e3c	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaab721d5fc1ce1a346911509773cb2558a0e4317f8198f777be81e07d243fc5a469dbc2b1099fdbbdb717d5128390e3f107927a8e91027eaa1c879b05d886b5b82cfc870a98cfcc780816d04db9dba66cadd5b5277eb5176812949f9d8413b21
0	73	\N	\\x82aaf18c09941a1a1200aafbf77d65ffa95379a53d54b39579e2ad74bdfd866aa894735d97b4a5feac1ab8ec4caf8745	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8a88b2a8b85ecafa9e8accf597833ea8813476f0b9b8255999c40ca975f4298bdba293017bba25ad5c70d2666baa984f0a6746a20c3f577aa5a828059ea3b65a8d0f9feeeba8a2fe377f35529ae8f77d3c8b46619fbc7f13d8c4d65269f784a8
0	74	\N	\\xa1c39c1f71e9a22c7fe0a06ea4e03383b9c359f1f503c2755cb549bdd1939c3870a27de26e640805d21bfdd0ddb40260	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8ff26d67d70997d59430cf0bde8e77242f3c3efe025f0542da58f9c614ce2241494a1dc6407fd84ba25fb1b1086b5d8513bf97d40be466302666c97ce769a56acd6b9fcb09105cd5464adb9db6f759fa451fe39dcbe235584cb50e8cb9a181cd
0	75	\N	\\x99471ccca8ddf24a565e42d994e1e55539ae07d6b305d6d8776155541547cea10093099d434753f0e02a680fd9f1e589	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8d75474a9af57469de4dc88206c7c65fe751c8e638dc00fab393e110382cf1464239c42401e0105d5af036bc9996d76808029ecdd89f2839c58f5100cf28380f28acf2f7950e5e66d2d39a8aff2c297397d96771a060e2a37e4c0d4185edb28f
0	76	\N	\\xb054848f937b2fc921e29ca8a549310689af24c95f0f44f8bd376073e8cd7ae44d599ab4fc4ff0def597453bc38f4b0e	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x984fafdaa449c66e092478439d2b7648009232a8a61c5c07a45033ca465dd5f0742d54d321abb700e04dafa18617fe0913f05dc5082f084d08f062ab81dbdd59dd8bcf5caff37b2702c2d9b71e7bed9bac2a95237ef5be446cae438eeb9af306
0	77	\N	\\xb3343682614d1444c616b6c2002d57d5ecf9727ab817685abf231dee19b86f26700da2b4d1a40246b506809668d53092	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8798be9121daff341dc1103a98522cb7e76bb775a9d5e542e8c7c37948646b877dfde8787f3628ab418b156be3433bc3035e9333155087426ef24be08e87becb019c9a7bc1567d24e0041c85674db2f2530a1906e53ea4da34334be4373a2b83
0	78	\N	\\xb9ede095f216e4e31ffdd1e0f0146a68cfe7b84503930333aed1a64f08f3a8ea201a3800cbfe8dbf13d33071fab9ccf2	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8850f7114c415c134e15fdc8657c48d83dfe18424f8cdfb0bcda045dbdd66d4e200671fc86fecb1ffd7d2a07cddbc2be03cd3686efe298c1bb0295fc2ce4dba57c890fa88e8fdd88ee4b5bf86d8bb6afdb804b66c37ce59bb8b23704761c8571
0	79	\N	\\x868261009e0d7371c82e9b4b0ba42c7da45c62f11cf069b986d97eb6671af29a170e287084dbffc136f3d7a0956fc1d9	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8fb8a3d59a1679dd2bfb7bfcf5107e5f239cbb1b56a1828d8d94946e8a9320b895e3cbfbd8d5570c92fd74b3dace78a718c1acd563f7f31d201634ad20d69eb5c2c4c3c8f7125e34782d9e26138abcb06efee70936215a585863a9d6ed0db838
0	80	\N	\\xa4b5ea77092b7e41d10c66f0ff0a0c6faa6e1a01326ce3af7da5fa47beca9b0da2521eb3ff2e14706d5109009b2326af	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8973530ad20336853522d07829706f767ed9b4a032ca8ac004d36f6ef0b441b594571a93d7875c2d7af3889721179e1d0387d585d4fa2aae054a4cdf79ce7c3ade50a3363239b56e3d7b8094fbefd9675e1a5fc768a4640e37de9a28f39f7634
0	81	\N	\\x88f63706c021c9ddd421ae1989a69edd113e711f3ea4de4cd22dc61d41fcc268b7609e5a51101c0eddb4815eb8fc3424	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb6f8e82ba77c7ef524ace062298a2adff0204e61328c87a4a293d7f00a824d4cd975f8cd1b6e3fd0b5939d8026c816b917eaab7a45855c0f398aaaf89d2792b1538cc16974af1f8f9da89cfe91c556d3f36ab44e37de83323b9f2b7e2ad087b5
0	82	\N	\\x93f2e8e009823611c369989eb6a29fd2932ff665ce6319e6e2f575b20adb1682b43fb742c95e5fed75a68c24beaa6e5c	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb58e5a67881a9ed6ca5c1d345f1396946b8a763de9edfc47d7dfe9d3d69169262e0ddd678dd6535cd1e359bf1a2715371787424fb8c15e1ebe13ab1223e61715174354751e2c67dfc15e2331588fe5961896867f021e527fdd53725b093ad202
0	83	\N	\\xb879e1944ef8b9303ab3cfa7a4f97028e4063c70f029c1ddf25c31d5499f2df78c34a1e40253b95340ce4a5a8b97bab6	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xac76f8df667076b123b6acbc38c0c44050f91be6b87f8f24835bceba0396f1a6504e9da1ea0bd84c86c5fe17e7da7804137eeb577b41f4464ffb4055e6d0c0476e62b0ef6bca8e7ca928d7d912bcf1fa689d6af4bb1693a5ac363de6e4a39ae3
0	84	\N	\\xad65e788945d6b989ca83c4dae3e626bc49bf1ae1a6499edd9be700256eae91738d89a94e7584ce597e5c5a240644ab7	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x83b654b4957a89ae20ee8f3cf396d3d8fd447023c896aeaefe99f7d272e4d8c0b1901b597f26b22b8a6579866ba99d6414ae55c4397f661fbf8ccdb8384b444d087535c73217a784ab5504b275fd52462be27c6cb8b9f916ee78ee62560c1b18
0	85	\N	\\xb222a1943f7b04aa52cc85d315604c7bc8d8039d8ec9edb881fde1a1e5e1515f0b7d2cde70ad2b052b0b85249a870860	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8eb528bfdd55445828cd0138be06a566087ec1cc4f44e82691d615415ac23bd0ea0ff5030271620199584b9a14b4ee7b0278eb4755533770b69b4f0a5e616d590eb8683ab384dd2ad3b0222ad56f5465c3d3db74fac3a1594292a037b2300fc4
0	86	\N	\\xa7e5be5aea8a2b20ac8183b73a7c4dbdc234543c990e6f0c7e6cb8704814794a3133a407993ab31b35f17e001f1381bf	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8653f36e7bd4be4e788b61d7592773dcd3ef2148d1e9766d131d6d00cb9e5a61379c4c258b3ec49f1ed6926ccc3bd56a02b3d60d4dcd2fb8b23860ff56865f19f9d8fc7ea4ccb3bb40152b4d3c241a7a9cf38c84ead7d2ea47980d32bfaa66b2
0	87	\N	\\x97703c4b915b4537a05ee7531c5a0ffe564641cd24ee9358b8d6b85f55c3d68c7f134bff63644dbfef9ce1ffcf2c81ef	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa470674e02ebf2e4eda0ee4f54124e3ecbeda630f23787a5171807988a0725298a3813df04b518666d3bda774ddd39ae13f28ab6360ccc7f9c1f4a934c301a51c8a075b7a03f77a24a0ef58eb09197797200f422e85114d0fe6f62c9e842a476
0	88	\N	\\xb17d56fde958711d33f93081f997c146ed0987a3ebd3619e743732fc8ad0561fc1d59cd2d183e67beda306c44fb287cf	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb44029b8c2182c9c63466ec37850549db484d99c322f79810bd4f8153071520a3c09581a47baf6f5af93a32a9d07ed40029edaf8f26acef481c9ad024cd6542d3791da03f7b454f66c8d2ff04a389116c87ad7524281afc42997af0f44cac7f3
0	89	\N	\\x8cc4217b53eff3c0db05cd09782f03ccd2a7438954fd441e48639a0af1f1903320f1d5081b9f9d17c836fe3840dfc6e5	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x95fc7c69763b011a668fb2713910029df4ee64c38edba5df8d480074f048a6c91f0de0e927e3a1674e61b4cfb94fe9c60038f7a066a924c965d78785d84f2b79bf6459acaac83c8e0c2fbbe2524a2512daef311aa93e848055a8172fc8099815
0	90	\N	\\xae2052814ce1c7b48beef13acc80280dd844beed7773e2fc8ddc0c5cb0e4a76f576d85006dad8cd324a793e874403410	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa30070943efe7f64557ba5d6bb7270609fb973afaf0711214a3acad4521960171b984ba8eb320063e0f9ced80f3f46db1648c1963ae9b8b7e4e6e33933f79beb3bbf32cea2c02caeb7a751acce7913f784973f7e7b39bb9e73e76cd7d6e28756
0	91	\N	\\x807ef674ff12b472daf2d029702ffe5bdcf161ffa2f0226ffc85fe7ef5bcb66b17557b2eab9a59edc73d83c9ab92fbb9	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x89112da90d5286cfd726259f43c74b1c9c3020b230f6bd1beab61a3941c7aa511bd40bc07bb6787f44cdf2622cfc9d6c0cb2a135932cc1719754d4a268fd4ddd71661797b2ff5cc63fe8e14e5b131a8b58d80c7475fba56dcd1461c1b1e74e7f
0	92	\N	\\xa72f914efbcfd6b2501c2f733129fb84dc602885fcea70c2ec55782e241d2c333a3882970769dde34dc259da5b52fcd6	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb535ccf150cb8cbc0377c9fce7a2229291889a810ef189f66894878e4be8c131a4dd73125a5a1339ee42e3e6b57b833e190d59c996dc0d7e09d069c312d21d9b98fe4989048586ec9326f7297727d6d123bb3a940843dbb36f7b44a58c1dcac8
0	93	\N	\\x880a0b92dc9140cbe6e1bef997192c3bdab8994b79d7b62895aebfd67982e64447422f79e568765128c3189fb4575a6b	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x89da76cc415daf5804019a1c2195607b06d59e656fc54ee99c2ad99d85f5298022e9233018d1e45c37ed6e75c945df1816f7b5b600507894b3335a12876ffc9b2b8684c76fa99300faa047d5df8cf27367dbcf191665953b272f97b2f811c308
0	94	\N	\\x923be44bb715dfaec171e326c57f84b416276a50dc6ba194c0eb0b6780328fa1d5de6643685af06cca0c607691cff36d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x84c8d1e6fbc2cc05b81a100e218e854e55229a6f86c0db1db5dfb48fc117795583b4b50b07f32a078693d02bbaa76f2c126d7474350f4e8e167d67a588ec558ce60ac13a0cb2c147d4f338f42125b503a535ba7fdfd287f94f5a96aa00f05897
0	95	\N	\\xb7cba745fd5aa261107954094ce4e454bd56fdf72e956e7388ffe6af861fc0d46c8f075c8176e3e2b198445d92340318	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb9661958653c7e8311345b7b981aa8d055868f3899c24dfc8a3205ffa2ab0723a1a8f181bca236d9edff7efa112d43af106a291d5485d30207fb1d5826dd8235db707189954e519b38f1b17283d66ff49eca388dc256ef4ea6b78105b7bd4d6b
0	96	\N	\\x93d0a4555de1f63b62e61db97dc2b9deaa85ec397a813975eec44fd010e291dae6b4a67d3c2cab2940374301fd7bddda	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8d670b7c95806062c36e416c32ea3bb4282c5aa6e980c8f15290efb751346359f6a363cc495d91da0c047e4b0591440218d05416c52a59e43e86b8a4389631e4f00e54a802f980a521d77c025712608b33156b0fd40b907f1a75f8dc4082ee79
0	97	\N	\\x98a03ced5d966d1c72a91471933ce3a324fa0382764ab5ef93534655f9a1768a36fac6278605447887a45159797a5e7e	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xadf10a35618a2bcd2a908686ea837bb2c5df87a97600476cfc5e379e5e460177bf41725aafa248279234939d86107bdf1897291fa463bca2bff47bb6f3367e12840215adfc774d79c5822cc2520d575d7c0ba863a89d21f6d60bcc005b403e48
0	98	\N	\\x8e743af5edf5e2130eecbbb629a60c17e2ce46d2bdefa4f55e5e94fa9868224970ee5ccf826cf29a8e404c437697ccb8	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa9372887f704a701f7a73cf3d1a186dbac8ebf80506ce4d1550977e01611874f77758784239ef6c1f4b4eca0e7cb347509c38926fa311e810ca7bf786cb4c8fee422609b76421eda698f79c3b54f26a9dcab4d1cbf4c53c431d95f9d9994db9b
0	99	\N	\\xb7daf482af8eaca97800fff286f9e52d25945e595ec9d1cb7f375c49b79586d695484ddd3485537b3bf4e12d68330230	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa656450e5d3cb7e32b2fbd0785ca3a2c5583f39090bfa21cc92c65f7cc63fe8d044b83dbe505aa2b31ea6223c213e0360f1e9ee2d164fff329c972dabecd796fab2ac8b9e4545d453fcb39718dd3eecec7a1b8b8d04522fbe39b8ac8c8214629
\.
COPY public.blocks_proposerslashings FROM stdin;
\.
COPY public.blocks_voluntaryexits FROM stdin;
\.
COPY public.epochs FROM stdin;
0	32	0	0	908	0	0	20084	32361680940	649952000000000	t	642688000000000	0.597689688205719	384128000000000
1	32	0	0	1393	0	0	20084	32361680940	649952000000000	t	642688000000000	0.5767278075218201	370656000000000
\.
COPY public.eth1_deposits FROM stdin;
\\x342d3551439a13555c62f95d27b2fbabc816e4c23a6e58c28e69af6fae6d0159	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001208976a7deec59f3ebcdcbd67f512fdd07a9a7cab72b63e85bc7a22bb689c2a40c0000000000000000000000000000000000000000000000000000000000000030ae9e6a550ac71490cdf134533b1688fcbdb16f113d7190eacf4f2e9ca6e013d5bd08c37cb2bde9bbdec8ffb8edbd495b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200062a90ebe71c4c01c4e057d7d13b944d9705f524ebfa24290c22477ab0517e40000000000000000000000000000000000000000000000000000000000000060a87a4874d276982c471e981a113f8af74a31ffa7d18898a02df2419de2a7f02084065784aa2f743d9ddf80952986ea0b012190cd866f1f2d9c633a7a33c2725d0b181906d413c82e2c18323154a2f7c7ae6f72686782ed9e423070daa00db05b	0	3086571	2020-07-21 18:08:26	\\x35d803f11e900fb6300946b525f0d08d1ffd4bed	\\xae9e6a550ac71490cdf134533b1688fcbdb16f113d7190eacf4f2e9ca6e013d5bd08c37cb2bde9bbdec8ffb8edbd495b	\\x0062a90ebe71c4c01c4e057d7d13b944d9705f524ebfa24290c22477ab0517e4	32000000000	\\xa87a4874d276982c471e981a113f8af74a31ffa7d18898a02df2419de2a7f02084065784aa2f743d9ddf80952986ea0b012190cd866f1f2d9c633a7a33c2725d0b181906d413c82e2c18323154a2f7c7ae6f72686782ed9e423070daa00db05b	\\x0000000000000000	f	f
\\x6bab2263e1801ae3ffd14a31c08602c17f0e105e8ab849855adbd661d8b87bfd	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012061cef7d8a3f7c590a2dc066ae1c95def5ce769b3e9471fdb34f36f7a7246965e0000000000000000000000000000000000000000000000000000000000000030b1d0ec8f907e023ea7b8cb1236be8a74d02ba3f13aba162da4a68e9ffa2e395134658d150ef884bcfaeecdf35c28649600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000a6aa2a632a6c4847cf87ef96d789058eb65bfaa4cc4e0ebc39237421c22e5400000000000000000000000000000000000000000000000000000000000000608d0f8ec11935010202d6dde9ab437f8d835b9cfd5052c001be5af9304f650ada90c5363022e1f9ef2392dd222cfe55b40dfd52578468d2b2092588d4ad3745775ea4d8199216f3f90e57c9435c501946c030f7bfc8dbd715a55effa6674fd5a4	0	3086579	2020-07-21 18:10:26	\\x35d803f11e900fb6300946b525f0d08d1ffd4bed	\\xb1d0ec8f907e023ea7b8cb1236be8a74d02ba3f13aba162da4a68e9ffa2e395134658d150ef884bcfaeecdf35c286496	\\x00a6aa2a632a6c4847cf87ef96d789058eb65bfaa4cc4e0ebc39237421c22e54	32000000000	\\x8d0f8ec11935010202d6dde9ab437f8d835b9cfd5052c001be5af9304f650ada90c5363022e1f9ef2392dd222cfe55b40dfd52578468d2b2092588d4ad3745775ea4d8199216f3f90e57c9435c501946c030f7bfc8dbd715a55effa6674fd5a4	\\x0100000000000000	f	f
\\xdaf29ed67bf1d7ccc92e1a4f8fdfe5779c84a7d2612f8b1b3cc31e4fc2703db5	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120349e8bd4873a30ae48088ca3b0cdd250e0ee4b066472e2d7ca61e5663d712ed800000000000000000000000000000000000000000000000000000000000000308fcf28896a85e5e76ee9e508438e23e7253da1a23a6501e3a7d56182520dbcf4cdb44af3267318188f1f4168342146da0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000609800d7c29908ddd1754490577753e65bd27484ca456b33b914e937416d58a6dd5b4eb420c22c144ae46f203fcc869e5c18f4e6daf99122294dc0ea691242625e7b75dfedf5d799899f44cf648245ccfc9652a957ae9216b022567c4cc4a72066	1	3092261	2020-07-22 17:51:01	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8fcf28896a85e5e76ee9e508438e23e7253da1a23a6501e3a7d56182520dbcf4cdb44af3267318188f1f4168342146da	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x9800d7c29908ddd1754490577753e65bd27484ca456b33b914e937416d58a6dd5b4eb420c22c144ae46f203fcc869e5c18f4e6daf99122294dc0ea691242625e7b75dfedf5d799899f44cf648245ccfc9652a957ae9216b022567c4cc4a72066	\\x0200000000000000	f	t
\\x094cdf68d68e887cceb19dfabe79756ab12b46b2b599b18c31d5bbb330e513cf	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120e052639983e0db170f1223a940920853aac77627ffea602df2973916e717cb0b000000000000000000000000000000000000000000000000000000000000003088569c0665532f1cf039a4887b9a87a873ce8d656b2987d0d211e16cd61bcb44ca9a4447f45ccf4436ad8c0799dd528e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000f69a845444c94c419d6b1e74633163946bb9c7817d5d5705b30937d686915800000000000000000000000000000000000000000000000000000000000000608d41380ac30df33ef951e7014833f0cafac5f884406fc77da818824c89cd5585cba6bd4dc83acb9ff327314d3b28f6f204d69c52e02a00f400fd599505d0223d9c9a835221d5387d3ddfe319f483ccf7f0ffe65547066ca6a98b171358dd86ba	0	3094731	2020-07-23 04:08:42	\\x11c100b0173d8cd7e73f6e1808e5a811fc5a93e7	\\x88569c0665532f1cf039a4887b9a87a873ce8d656b2987d0d211e16cd61bcb44ca9a4447f45ccf4436ad8c0799dd528e	\\x00f69a845444c94c419d6b1e74633163946bb9c7817d5d5705b30937d6869158	32000000000	\\x8d41380ac30df33ef951e7014833f0cafac5f884406fc77da818824c89cd5585cba6bd4dc83acb9ff327314d3b28f6f204d69c52e02a00f400fd599505d0223d9c9a835221d5387d3ddfe319f483ccf7f0ffe65547066ca6a98b171358dd86ba	\\x0300000000000000	f	f
\\xf5110f464bd54b5b37969b295131970115563cc06bc0b5573d8bb24754d1820e	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001201b42f6caf3f79e0f060720730133ebdd129615503ee6f4a29f7b5b98c19a94d10000000000000000000000000000000000000000000000000000000000000030873e73ee8b3e4fcf1d2fb0f1036ba996ac9910b5b348f6438b5f8ef50857d4da9075d0218a9d1b99a9eae235a39703e100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000b8cdcf79ba7e74300a07e9d8f8121dd0d8dd11dcfd6d3f2807c45b426ac9680000000000000000000000000000000000000000000000000000000000000060a6aef7016df5e73c2c142e3f7a0a0a4c85283a6e8fa3c2f52093aa029d04560ebdfebd4101d3a1bab0ff2812bbe1e51c13621b9366c83c63a081cbcaa3d500da9af1cf05b9e2c0bd7feffab31cbe6a158655f23b124f4f2fea746e8f93d5a1f1	0	3096914	2020-07-23 13:14:44	\\x11c100b0173d8cd7e73f6e1808e5a811fc5a93e7	\\x873e73ee8b3e4fcf1d2fb0f1036ba996ac9910b5b348f6438b5f8ef50857d4da9075d0218a9d1b99a9eae235a39703e1	\\x00b8cdcf79ba7e74300a07e9d8f8121dd0d8dd11dcfd6d3f2807c45b426ac968	32000000000	\\xa6aef7016df5e73c2c142e3f7a0a0a4c85283a6e8fa3c2f52093aa029d04560ebdfebd4101d3a1bab0ff2812bbe1e51c13621b9366c83c63a081cbcaa3d500da9af1cf05b9e2c0bd7feffab31cbe6a158655f23b124f4f2fea746e8f93d5a1f1	\\x0400000000000000	f	t
\\x0c025ec458bbe33e625e33b45f9608d66712201750131fc1471f9bc7de51278b	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120f5da58762a52499649f84f3f2e213a0c6ff0039b4392910b6b72f2488e4705e400000000000000000000000000000000000000000000000000000000000000308c2f535d3bec65f95cb4ba455566e4ec3de8da5c13a681699e0f80d7942d6fdcbcef18c8cf18f9da14aa379bdd6d29c5000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020006490500934b8b1876401dc09b7904d04c0897a9a28ecde4ddb1a60fd5c4c6000000000000000000000000000000000000000000000000000000000000000608ac5120c07c2202fbc31ae2e5dc9e737cfa4056344f908ab66223bae755227f642efeb9d7974eacb2ecb1dfef4021b79068aaf9ce1d0b42ddaa136fdfcf8b29d6a2fbb06259d0f542bb35a4658908c14e893baee1170aa22bb98a78f98983264	0	3097139	2020-07-23 14:10:59	\\x162fa0abecf596e07fb8ef128c15cd589f7b473a	\\x8c2f535d3bec65f95cb4ba455566e4ec3de8da5c13a681699e0f80d7942d6fdcbcef18c8cf18f9da14aa379bdd6d29c5	\\x006490500934b8b1876401dc09b7904d04c0897a9a28ecde4ddb1a60fd5c4c60	32000000000	\\x8ac5120c07c2202fbc31ae2e5dc9e737cfa4056344f908ab66223bae755227f642efeb9d7974eacb2ecb1dfef4021b79068aaf9ce1d0b42ddaa136fdfcf8b29d6a2fbb06259d0f542bb35a4658908c14e893baee1170aa22bb98a78f98983264	\\x0500000000000000	f	t
\\x20d80c684652b51913cf4538e25b5bbeaa83cbf9ade0adbb5805666d3d195f0e	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001201b42f6caf3f79e0f060720730133ebdd129615503ee6f4a29f7b5b98c19a94d10000000000000000000000000000000000000000000000000000000000000030873e73ee8b3e4fcf1d2fb0f1036ba996ac9910b5b348f6438b5f8ef50857d4da9075d0218a9d1b99a9eae235a39703e100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000b8cdcf79ba7e74300a07e9d8f8121dd0d8dd11dcfd6d3f2807c45b426ac9680000000000000000000000000000000000000000000000000000000000000060a6aef7016df5e73c2c142e3f7a0a0a4c85283a6e8fa3c2f52093aa029d04560ebdfebd4101d3a1bab0ff2812bbe1e51c13621b9366c83c63a081cbcaa3d500da9af1cf05b9e2c0bd7feffab31cbe6a158655f23b124f4f2fea746e8f93d5a1f1	0	3097251	2020-07-23 14:38:59	\\x11c100b0173d8cd7e73f6e1808e5a811fc5a93e7	\\x873e73ee8b3e4fcf1d2fb0f1036ba996ac9910b5b348f6438b5f8ef50857d4da9075d0218a9d1b99a9eae235a39703e1	\\x00b8cdcf79ba7e74300a07e9d8f8121dd0d8dd11dcfd6d3f2807c45b426ac968	32000000000	\\xa6aef7016df5e73c2c142e3f7a0a0a4c85283a6e8fa3c2f52093aa029d04560ebdfebd4101d3a1bab0ff2812bbe1e51c13621b9366c83c63a081cbcaa3d500da9af1cf05b9e2c0bd7feffab31cbe6a158655f23b124f4f2fea746e8f93d5a1f1	\\x0600000000000000	f	t
\\xd956d74847717b68c6f9ca1657d6e197933f831472a10d986aab9185a5309a42	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001204d7f336b66eee3aff400810d6ac0ad6aa498a38ba236d59a37e32e8d206ac5660000000000000000000000000000000000000000000000000000000000000030a8d9b5b62cc31149ad58a281a2293cd3f4dca11855c98983e76ffb60479d8e98e5592e5415f0400a8a23efdd842b3605000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020002f5bc32089b840a516c800c3c597bff24536c6c7c9f1113c2a0ac847608f91000000000000000000000000000000000000000000000000000000000000006084c256474e1a1f1b902b256485ef5881432847a3e14bfc6f5a48e250404e9f85737cecf1bb1cdf3b73cb0a1f38117dbf01d045896c2e8f918e4f5e9c985b0b60708c459a6c7618200f68c3a65e6b86b9d82dcf66cfce6419ae775f3c56a9277b	0	3097254	2020-07-23 14:39:44	\\xf9a0a0706997fcec8389f152c15d540a8b7a8507	\\xa8d9b5b62cc31149ad58a281a2293cd3f4dca11855c98983e76ffb60479d8e98e5592e5415f0400a8a23efdd842b3605	\\x002f5bc32089b840a516c800c3c597bff24536c6c7c9f1113c2a0ac847608f91	32000000000	\\x84c256474e1a1f1b902b256485ef5881432847a3e14bfc6f5a48e250404e9f85737cecf1bb1cdf3b73cb0a1f38117dbf01d045896c2e8f918e4f5e9c985b0b60708c459a6c7618200f68c3a65e6b86b9d82dcf66cfce6419ae775f3c56a9277b	\\x0700000000000000	f	t
\\x89c33ef96934bcc86cbf16fa6c72ddcae1b40787b54ff93793cf5f7a547ade5f	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001202684a3616b3a3ce59940e32d1628d6d56b4830728b5f996439c07c71c3b25fb80000000000000000000000000000000000000000000000000000000000000030adf943279435f1c194add1cdfe99e3fde5284d0451a63822b03ec301bb1cab4399d016f812461f0b71a8206b96ca33780000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000608acb8c35711572f0729e71d79a2b24598017a20ced75680c6688d253fc5a7dc4e3c7b2c5002e247c810eb77f47c75bfe0fa2dbbd6b0ad0078274773d8196902db0529ba23fed23b61315c0fd09aaa64fcf335c8c7c6d802dcb08b4e4a2a7372c	0	3098329	2020-07-23 19:08:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xadf943279435f1c194add1cdfe99e3fde5284d0451a63822b03ec301bb1cab4399d016f812461f0b71a8206b96ca3378	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8acb8c35711572f0729e71d79a2b24598017a20ced75680c6688d253fc5a7dc4e3c7b2c5002e247c810eb77f47c75bfe0fa2dbbd6b0ad0078274773d8196902db0529ba23fed23b61315c0fd09aaa64fcf335c8c7c6d802dcb08b4e4a2a7372c	\\x0800000000000000	f	t
\\x9e2d22c7199b57a2e4d3d720bb06a7b4bfa9a3d856808f6b47993182eca6d92b	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120aa3caa35512bb293813dd19fdc967eb615ad82fedd87ee17ed3c4aeabf65993500000000000000000000000000000000000000000000000000000000000000309274b8d712b53c71dfb2d8762334d63ec9fb18fa6278aaf47b2782b14447539b6e5f56a9836e6ba51e4a206f1a8dbe020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060888123b3de4fa234d9b529628ff6f097a58c45d065fb6d15358c0d47fca500bdd2fd1823be549378c40dd7ad947b49a415ec15f3cf996f0f1f5532903d11b8e5bc85607e76c9dc1c23cb37a427eafef7c4ce2594061c5f6cba6561d7367f27af	1	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x9274b8d712b53c71dfb2d8762334d63ec9fb18fa6278aaf47b2782b14447539b6e5f56a9836e6ba51e4a206f1a8dbe02	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x888123b3de4fa234d9b529628ff6f097a58c45d065fb6d15358c0d47fca500bdd2fd1823be549378c40dd7ad947b49a415ec15f3cf996f0f1f5532903d11b8e5bc85607e76c9dc1c23cb37a427eafef7c4ce2594061c5f6cba6561d7367f27af	\\x0a00000000000000	f	t
\\xcb656d2b606433fb255644b0440d3cfba72226fdd9a4e9456bd2897ad8318654	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001202a430a197eaa8b63068fb47d74b214d62cd3900aeb20b72e4a2f1929267d353e000000000000000000000000000000000000000000000000000000000000003090d42bdd0b5a7b62f665f75790739233b3a700f5f8ced95233fadcc72ca16977c8d6fb00b63d08ef750a88fc0ffadd6d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b378ac392f3916296d53318981035f600e329fdc2c550e0d2b12730f5a66e79306cfd162522666042e14675a72080df20a43b44d983a1deb3d1b1d711996a1a5dd03b3c8397df0bf1335ccd1580c9d62b73269045ca3fb3886817aa7c7e7ac9d	9	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x90d42bdd0b5a7b62f665f75790739233b3a700f5f8ced95233fadcc72ca16977c8d6fb00b63d08ef750a88fc0ffadd6d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb378ac392f3916296d53318981035f600e329fdc2c550e0d2b12730f5a66e79306cfd162522666042e14675a72080df20a43b44d983a1deb3d1b1d711996a1a5dd03b3c8397df0bf1335ccd1580c9d62b73269045ca3fb3886817aa7c7e7ac9d	\\x1200000000000000	f	t
\\x1b0822aef9e4c5fe0229f35597f4277ac9bc167d5293c164a5dd5b5468911f76	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120aef9d5778350689cc623189f9c0f3abb5214fcc722d0a010d0996d63da9fbe95000000000000000000000000000000000000000000000000000000000000003098b0d39db01e3bbd6b750c43c26f7078581a326a9778760db3b72b22b74cf1b9a2025f4347f04bbd36494ec5bfd8547b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091000000000000000000000000000000000000000000000000000000000000006091e2324ce96b619ad46ee00d4ecbbef08f2700da307e0b84da2a3247850a4bff533cdc568d3b7a3b57163abd5690aa0b15b854814e34088be81b2d94a5028704cb755042e540d5f81488d10fee0ea57f8d3d2b4f6d2c9f6339fb7ef0d78d178e	12	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x98b0d39db01e3bbd6b750c43c26f7078581a326a9778760db3b72b22b74cf1b9a2025f4347f04bbd36494ec5bfd8547b	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x91e2324ce96b619ad46ee00d4ecbbef08f2700da307e0b84da2a3247850a4bff533cdc568d3b7a3b57163abd5690aa0b15b854814e34088be81b2d94a5028704cb755042e540d5f81488d10fee0ea57f8d3d2b4f6d2c9f6339fb7ef0d78d178e	\\x1500000000000000	f	t
\\xbfae8a5049ef5ae4f76c65fd5783256cb42c4d59859a36196211738d158da68c	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120362fd852a8cf75a41fcf4852f8b54797ec6cb7c3f362251085da559416d6811200000000000000000000000000000000000000000000000000000000000000308e368a304c4c564617c3fa265361fe679df8160e6af06c5f5dba57d8ad134e7073a7ad631459116aef53a788e3c256b30000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091000000000000000000000000000000000000000000000000000000000000006086f0d7a21e189bea04aa447eb9237bb5a246cabc31fdfb2d9aa8a9b63f2f331014007008b12c5167ce399b3f9949f7051931ee1beb287e88987beaaccc6ea7aa3b77c7dc749a47f1ef596b52f50c28c8ec7fc2c322ebc23944e10d236ca054fa	4	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8e368a304c4c564617c3fa265361fe679df8160e6af06c5f5dba57d8ad134e7073a7ad631459116aef53a788e3c256b3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x86f0d7a21e189bea04aa447eb9237bb5a246cabc31fdfb2d9aa8a9b63f2f331014007008b12c5167ce399b3f9949f7051931ee1beb287e88987beaaccc6ea7aa3b77c7dc749a47f1ef596b52f50c28c8ec7fc2c322ebc23944e10d236ca054fa	\\x0d00000000000000	f	t
\\xc7996f8b21d774c1886d69c9d821c53e5123fdcfd5a672daf85fce6dec474a59	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120a9cbd970ad7a28df151a78b286002267c9e6e2c046509d0a46ad63cb493891860000000000000000000000000000000000000000000000000000000000000030b4fc18feaa072d37538e1fb8157a365bcffd086de9d335dca2eaab1508fb3a48504b8b982d05919f7989d818e88dd07b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060a00beb954f714a5792f7b3df495965378d497c36b27c09aa8faac0e113e36c70591a5ca2ef8a6ab3662ab1d930ed50fc011f49ee23aba8c354112151cddd652ab68a27f355d51b1e1438729eb0d0a12dc5e9adbb91a580560917377d72d3155d	2	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xb4fc18feaa072d37538e1fb8157a365bcffd086de9d335dca2eaab1508fb3a48504b8b982d05919f7989d818e88dd07b	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa00beb954f714a5792f7b3df495965378d497c36b27c09aa8faac0e113e36c70591a5ca2ef8a6ab3662ab1d930ed50fc011f49ee23aba8c354112151cddd652ab68a27f355d51b1e1438729eb0d0a12dc5e9adbb91a580560917377d72d3155d	\\x0b00000000000000	f	t
\\xfa2c7e061f27412df113c1e5adcbc5c3327b64191469e1ada32cff924ce0f6e6	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120b95c96dca4c100beb9a87cb582765a2962096b5d1e380981702ee1e4fbf65f5f000000000000000000000000000000000000000000000000000000000000003098573c25eb271690db29113fc890175213b1044bcd68e82f413bc694743ea265ca5eb8c50ebe2ae898426c0c9de650b30000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060a3e7fe6eb0f300703b4bb5c2282bba09e84c82b709edc01a77d782da69eba36ecf8c20063361f02f484167b218a532fd02c9af6fbb6b362534d945370278769f5fc622105099ba168cb0586b42743f1a5a2caf2ab9cb8b7b11dc4da3c638497b	10	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x98573c25eb271690db29113fc890175213b1044bcd68e82f413bc694743ea265ca5eb8c50ebe2ae898426c0c9de650b3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa3e7fe6eb0f300703b4bb5c2282bba09e84c82b709edc01a77d782da69eba36ecf8c20063361f02f484167b218a532fd02c9af6fbb6b362534d945370278769f5fc622105099ba168cb0586b42743f1a5a2caf2ab9cb8b7b11dc4da3c638497b	\\x1300000000000000	f	t
\\xeb377eb6aced2319c04d992698f4125cdcbdf41b2120bedb031906be0eb67ccd	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001202c86169403e0fd84eb4e2ec2be54b0550eb1810388e93f82458b9672c869e144000000000000000000000000000000000000000000000000000000000000003085f2c3045f02ac7b8b235b9aa855be7749d52de8b8e9855db5a9e97cf101185c05ebfa5d3d3dc2783e009294525ced200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060a555d0136508b3ee62c822a4e6f0f8d87679e74e2bc06a57fb279a4d5a8ccf683e3692a480cea1812862752929c1e708181dac71164f0a1cf09bf7a2d979de2519d93ffd6413d53ea0f335418f156f89fad725b3003809f33722bcd98d9199a8	0	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x85f2c3045f02ac7b8b235b9aa855be7749d52de8b8e9855db5a9e97cf101185c05ebfa5d3d3dc2783e009294525ced20	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa555d0136508b3ee62c822a4e6f0f8d87679e74e2bc06a57fb279a4d5a8ccf683e3692a480cea1812862752929c1e708181dac71164f0a1cf09bf7a2d979de2519d93ffd6413d53ea0f335418f156f89fad725b3003809f33722bcd98d9199a8	\\x0900000000000000	f	t
\\x4ae79042a480dae94edcaa3c6d4061ea948226c79de1250106c12e0465be0646	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120fddb6ae335ce3c2b39d1f78695996acd99ed2470ecc3b33ef4291e0793274ca900000000000000000000000000000000000000000000000000000000000000309939f64a7b916476b076abd67cad897991917d4b7da38487597696844a890cb5156c2928ebf97963e77d5d2099ca881d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060ad14254a0ff7f77bea8f71a259d8e56635898aae0594c460bc7b2cc06f8957ec64f02450382197becfdef2ed5e91278c126adb1f3f332b11b4829c989bf825aa8ec1b7f7bfcbf7785fce494b5e9586ec40f1841886c53340e51f30331c97f125	3	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x9939f64a7b916476b076abd67cad897991917d4b7da38487597696844a890cb5156c2928ebf97963e77d5d2099ca881d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xad14254a0ff7f77bea8f71a259d8e56635898aae0594c460bc7b2cc06f8957ec64f02450382197becfdef2ed5e91278c126adb1f3f332b11b4829c989bf825aa8ec1b7f7bfcbf7785fce494b5e9586ec40f1841886c53340e51f30331c97f125	\\x0c00000000000000	f	t
\\x7cc09ec0f00d490e6038681b6e83a1ca7f2b7cc5ad9b34bd19848849c697e7c5	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012096ce56a18bc2b3f60b67454c620462eec4b6a8414a3428ea57ea219ff4468a9200000000000000000000000000000000000000000000000000000000000000308725e7a24df271ffe645fe2a4fa46bae935666d4dc654dbf77e4853ce70229828ea1b60c49ef4644975cf2b466ed86740000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091000000000000000000000000000000000000000000000000000000000000006080490aafac86089d554d2bbba453d2284257dd59b7a2a36cca84092ea981e443200f1c36fd32a3b78a5eeddf8305c21f10e2b9301b54d88cddaf6151a30afa39ba57968b77f335ee0dbe1d9215f858eb178e83cd210c5f6b5a5e9309556940ed	7	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8725e7a24df271ffe645fe2a4fa46bae935666d4dc654dbf77e4853ce70229828ea1b60c49ef4644975cf2b466ed8674	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x80490aafac86089d554d2bbba453d2284257dd59b7a2a36cca84092ea981e443200f1c36fd32a3b78a5eeddf8305c21f10e2b9301b54d88cddaf6151a30afa39ba57968b77f335ee0dbe1d9215f858eb178e83cd210c5f6b5a5e9309556940ed	\\x1000000000000000	f	t
\\xb92c27f4e2230c3c61358161b2fd475a44fe40b8a46f607d188aee13246e44b1	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012002192054daee783c6eea1019ec5cf1f1642893d01a3ebeff9b37a62e968379c80000000000000000000000000000000000000000000000000000000000000030a898422079b7db772520e7b43ae7c36bdba9f1594fb6b38f69a5d46c1c4f9fbd1ee738abe7badc786a68a9d6dfd371220000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091000000000000000000000000000000000000000000000000000000000000006091774a402546f905387b12249640e0099904a16cedbc37f48b1eae2143a23db93118d55895cce87e0ead86e15bc12b560b612de5081aedbb3a183d92f85280e6dacbf8e626158aa5b6dee3f9979f0b859767936c787555332f33d2dcbf6808d3	16	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xa898422079b7db772520e7b43ae7c36bdba9f1594fb6b38f69a5d46c1c4f9fbd1ee738abe7badc786a68a9d6dfd37122	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x91774a402546f905387b12249640e0099904a16cedbc37f48b1eae2143a23db93118d55895cce87e0ead86e15bc12b560b612de5081aedbb3a183d92f85280e6dacbf8e626158aa5b6dee3f9979f0b859767936c787555332f33d2dcbf6808d3	\\x1900000000000000	f	t
\\x70a824dd722743d07c41e11e2186497156267c6fa107203e135980fc6d51408f	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120f2888c15039fd043217d32b6c403528a4a3b9335a722890aadf93c7ba5f3b0d0000000000000000000000000000000000000000000000000000000000000003094ce62f60d60f9d52f2d9456b9f22edbe7846fff4b98f4316466910bae8065c86eb0756869d5cc27427f098493bc76770000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b8b780c266f01f5efb556808c6f8fc862763ab82e8ab8ab9ab98ab06755d281f5a2e4f3151c726b2e5f8b17e9b79457f1655d836ba3c1aed0fd0ceead38709d377dbdaf1362bd33d287189a765636e0f5d7c593af4580219d2299c6798267dbf	14	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x94ce62f60d60f9d52f2d9456b9f22edbe7846fff4b98f4316466910bae8065c86eb0756869d5cc27427f098493bc7677	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb8b780c266f01f5efb556808c6f8fc862763ab82e8ab8ab9ab98ab06755d281f5a2e4f3151c726b2e5f8b17e9b79457f1655d836ba3c1aed0fd0ceead38709d377dbdaf1362bd33d287189a765636e0f5d7c593af4580219d2299c6798267dbf	\\x1700000000000000	f	t
\\x6caf2b943dd0ee8452b4ddd95f2431a47e55d1319da39d3706425bdf1a5bb110	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001208f17b31fd0274ab5689570586a8b04b2ce02d0869ed61158a1448ffae483c69800000000000000000000000000000000000000000000000000000000000000308687275aab69c59d56efa991c5db0b56600912ea03b6b4ff3e297f0a62434791c479d415aa494bd0b41a1604b3de0c070000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b0bdd47ccdff4b716e9482d652e9d52e286ccce7859e590aba7c79f954817ac80a34494632debf0532fd63afcfce5fb100fa1e83d2a4c1853b557412a7370a709239f3a2339c65fd7ffdbbf8018fafc1f5c958eea292233992f380b58c797528	6	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8687275aab69c59d56efa991c5db0b56600912ea03b6b4ff3e297f0a62434791c479d415aa494bd0b41a1604b3de0c07	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb0bdd47ccdff4b716e9482d652e9d52e286ccce7859e590aba7c79f954817ac80a34494632debf0532fd63afcfce5fb100fa1e83d2a4c1853b557412a7370a709239f3a2339c65fd7ffdbbf8018fafc1f5c958eea292233992f380b58c797528	\\x0f00000000000000	f	t
\\x31abe136ff882f7b89dee1ef5a2c6c57c2c321defc9e7b939d83ac346f13de09	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120f154df679bad2671c33723c012a32b3d855d8a8db378a4b661c142ea72c19be900000000000000000000000000000000000000000000000000000000000000309414c11a0e3b3712d0f62b7900dded64dc9c1c556c40e4dc40c0a7f5ceaec02135d45858aff36ef50b433e6fe24b6e390000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060aed65f2f7a36851e0ccfecc6e0cb865bc22cb3bd9fee5f6da5d467a91583efb9b205f0fcc115e120bcd789ce74ddd55706de0174c60ea6fdeda77b69d325d5dd0bc0bc0f22baf49762c60a985a7b13b5bc3e3e61ccbf7f82d56199da5bb617a6	8	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x9414c11a0e3b3712d0f62b7900dded64dc9c1c556c40e4dc40c0a7f5ceaec02135d45858aff36ef50b433e6fe24b6e39	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaed65f2f7a36851e0ccfecc6e0cb865bc22cb3bd9fee5f6da5d467a91583efb9b205f0fcc115e120bcd789ce74ddd55706de0174c60ea6fdeda77b69d325d5dd0bc0bc0f22baf49762c60a985a7b13b5bc3e3e61ccbf7f82d56199da5bb617a6	\\x1100000000000000	f	t
\\xc73215cc6a63c17b27e71ab4a262de1e253fbf1a40d3957575bd652a3c985152	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001201dc91bf69a129190f5eaca3c77d0bc7656ab3ddecae19dcf10d74bfe0e80a5bb00000000000000000000000000000000000000000000000000000000000000308a290fd6ea767704eecc68079e64cb9b43915652ebf9ba619b811149aefa069718dd97f8775004845ea9197f8c1cbee90000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000609785ef99e316c6c877fc2fdc6042533d230b292aea91e359c3b48e05a7677cbf8e0e47c0f1a8bc7a5a56880d821f2539173bb095a687edee5fd4030715642fc8d6ef16278a407cc410c24dc2be665dfbbe999aef6dcc8af9b64e00cef66b1d19	13	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8a290fd6ea767704eecc68079e64cb9b43915652ebf9ba619b811149aefa069718dd97f8775004845ea9197f8c1cbee9	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x9785ef99e316c6c877fc2fdc6042533d230b292aea91e359c3b48e05a7677cbf8e0e47c0f1a8bc7a5a56880d821f2539173bb095a687edee5fd4030715642fc8d6ef16278a407cc410c24dc2be665dfbbe999aef6dcc8af9b64e00cef66b1d19	\\x1600000000000000	f	t
\\xac02085ed8939ea3e288109815ad12bb05243ae060c3c83b21809185a0fcfc7f	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001203c039c524f7cc5dedd7e0775a9be5174d3b6178842a8e0cbb6415d6cbc8bede30000000000000000000000000000000000000000000000000000000000000030ab730f62ce3c14945e245b08b029dbf579d376cacaf82e5ae9f2b28fa72eec68e33103c693c502209a532edca594e0a30000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091000000000000000000000000000000000000000000000000000000000000006080c26cea1953be6a8879dac3c6c0d4798a72b5b27ac9eaa83e55fb66cf6ae5456f786f2925afbd35930e1f5fcc06074f08485fab890d03feb89cf30ada6c26a1cfb2c361f52f287c77adfc997a7975c80570c836164e7f78665882c8df01c8ff	15	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xab730f62ce3c14945e245b08b029dbf579d376cacaf82e5ae9f2b28fa72eec68e33103c693c502209a532edca594e0a3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x80c26cea1953be6a8879dac3c6c0d4798a72b5b27ac9eaa83e55fb66cf6ae5456f786f2925afbd35930e1f5fcc06074f08485fab890d03feb89cf30ada6c26a1cfb2c361f52f287c77adfc997a7975c80570c836164e7f78665882c8df01c8ff	\\x1800000000000000	f	t
\\xd4e8d8b944fd8f2e68b35113ed0e08545d8e62a03469ae15febb194ab77e63e7	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001207499664aaaf9385f58aaf00d4cd94eeabf19528e7d505e6e551738c8a11ab4ac000000000000000000000000000000000000000000000000000000000000003081903c25cba2b6f37f23449d731dc6978222a9df18ea08cba0a58c7d43d4b117e2ada338385fb1514b1692a5ca4881fd0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060826eceb6ad2c191a8bfbdef905b7f62cab64d180ab85c8f84495b594a5a6cd7aa4afe6c658ab55f3d42482f91309afa80231afe401c7c770245bafe432d9bbf87639827aec7a5581cb0c5bc39561179ac63641928db30daa06bdff556c74aad5	5	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x81903c25cba2b6f37f23449d731dc6978222a9df18ea08cba0a58c7d43d4b117e2ada338385fb1514b1692a5ca4881fd	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x826eceb6ad2c191a8bfbdef905b7f62cab64d180ab85c8f84495b594a5a6cd7aa4afe6c658ab55f3d42482f91309afa80231afe401c7c770245bafe432d9bbf87639827aec7a5581cb0c5bc39561179ac63641928db30daa06bdff556c74aad5	\\x0e00000000000000	f	t
\\x1899fbc5dea12f4eda381e8594beabfa437d835295224a3d586d12b3fb146a3f	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120793fb36f47b5de5a709f044fbc134804e717d765f351851a7433aedb93aac4d4000000000000000000000000000000000000000000000000000000000000003088048df033f7eb4df024ce81b29459adfa92f8c447e732ff6b77f51c5e721b365dad73eb25c1617c11d456ed916cb9350000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b6a5975d4200f2232f78c1f68fd0e313d9c00395ea64fe937a7f7af86fecae955dc8e8cde5ccd81df689faaa2cce8201033df473f43d952cf989a50aefc13a0d056ae8374a91ea7f9b155f0f0fb0375c68338d3e24e876b6c95e35900f53a2f3	11	3098330	2020-07-23 19:08:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x88048df033f7eb4df024ce81b29459adfa92f8c447e732ff6b77f51c5e721b365dad73eb25c1617c11d456ed916cb935	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb6a5975d4200f2232f78c1f68fd0e313d9c00395ea64fe937a7f7af86fecae955dc8e8cde5ccd81df689faaa2cce8201033df473f43d952cf989a50aefc13a0d056ae8374a91ea7f9b155f0f0fb0375c68338d3e24e876b6c95e35900f53a2f3	\\x1400000000000000	f	t
\\x7c601380b0092114d957defba102cf52bf17c8d78ab3d7a1561b13744e20e787	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120e6a64a2e2b7247c48a777f9c12eb73ef12878197fba2bdc65352cb87c548f8490000000000000000000000000000000000000000000000000000000000000030ac605f3f4df79840e1e948df5d67e8484f4c3cddc0d6e561bd7bb6e48a8a1c9baf154c316b1b3d7f4ee23ca7a5cbf7f70000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000609776aa6b6e3ffa6822b881f55bd2f2a02201246e36cdae5c5aa2dc2ad8204300f3ff04e9809febe723a28ce2ee4325fd0425275a0c446830a0052bd3f0934f5fa1a77e32bd40505dabf092c92b467b5e57cde38701dc5dac8ee6506871cc6e2c	3	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xac605f3f4df79840e1e948df5d67e8484f4c3cddc0d6e561bd7bb6e48a8a1c9baf154c316b1b3d7f4ee23ca7a5cbf7f7	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x9776aa6b6e3ffa6822b881f55bd2f2a02201246e36cdae5c5aa2dc2ad8204300f3ff04e9809febe723a28ce2ee4325fd0425275a0c446830a0052bd3f0934f5fa1a77e32bd40505dabf092c92b467b5e57cde38701dc5dac8ee6506871cc6e2c	\\x1d00000000000000	f	t
\\xb13ff66e2daa3ebca152a9b628faea319feba52bffa8ea32e77553f5e170e12c	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120d5fe4bc4001cc45392af85f3f9b24637b5ad0e3fb8e22864c50f9b79e3968b5c0000000000000000000000000000000000000000000000000000000000000030ad547ad6f3fb353aa0c664e4c770560b1809cd934aa356d27c0ceb41622363d9e607a90c3f206dd954f18040a336c20f0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b7537ef0b2ce874972f78176a009d2a201b999d6405ef68c45e53f50cd29cf1f2b0b83b7b1bd699ff01303de663cc965167aa9099373083fec5380dccbd8942e73ae3e4e9181d5babe61b01209c0702538219b29fd2d9cdb3e9a56c2229cd4c5	0	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xad547ad6f3fb353aa0c664e4c770560b1809cd934aa356d27c0ceb41622363d9e607a90c3f206dd954f18040a336c20f	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb7537ef0b2ce874972f78176a009d2a201b999d6405ef68c45e53f50cd29cf1f2b0b83b7b1bd699ff01303de663cc965167aa9099373083fec5380dccbd8942e73ae3e4e9181d5babe61b01209c0702538219b29fd2d9cdb3e9a56c2229cd4c5	\\x1a00000000000000	f	t
\\xfbf8fec0e71be049f6c072575654090d452f27d0f9000eb10c36bd31b5146ed0	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120ac63dfc3ef944d0354ab6c600fa2288b48114e37e8b6460571f301fb6a03b4400000000000000000000000000000000000000000000000000000000000000030b05fbe5c06ff586d649d8b21e9f75b37bf430fe04eb712850a8af2fc3455b5bd5051a258cde20aa2d59da5a0f0564ac10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000608eb7a4cca039ae6acfa8a87e38e991180f912497da94b831d41278e3b15d73b1528d5bb8595f0508541c9777c5f9d8f514e8b0362ae6673673dade2ebffb3dcf9a50cc8d46f9df4efc05ae955c9509a2665e8b9b7df50d83f0bea460426bbeb5	1	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xb05fbe5c06ff586d649d8b21e9f75b37bf430fe04eb712850a8af2fc3455b5bd5051a258cde20aa2d59da5a0f0564ac1	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8eb7a4cca039ae6acfa8a87e38e991180f912497da94b831d41278e3b15d73b1528d5bb8595f0508541c9777c5f9d8f514e8b0362ae6673673dade2ebffb3dcf9a50cc8d46f9df4efc05ae955c9509a2665e8b9b7df50d83f0bea460426bbeb5	\\x1b00000000000000	f	t
\\x1253d9aeec5b5ba7370c30644d3eeec750a4637d2328d8a1776528a41f8ab7dd	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120310e137d06a81a33d62bc30861e3122f2f2a9541bef0e9cf12decf94190f3ae3000000000000000000000000000000000000000000000000000000000000003083251cd9095f5fae3b2dfb006b16545d34bdaf1894687f2740300e9f8f6229eda1db3767c80a38fb3dbd8118ab415c770000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091000000000000000000000000000000000000000000000000000000000000006084ff95d9dfc223e2d52885a0738c1913bab11c634c640c3dd86cec6ad1895601a78e4df23d6927e3d7d62ff0e46d5a7401416982e07f38f78451034cad4bbe3afc8f8878f19a79aa5992b284f7cffed8585749b72be196d21eeff635717dbba6	6	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x83251cd9095f5fae3b2dfb006b16545d34bdaf1894687f2740300e9f8f6229eda1db3767c80a38fb3dbd8118ab415c77	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x84ff95d9dfc223e2d52885a0738c1913bab11c634c640c3dd86cec6ad1895601a78e4df23d6927e3d7d62ff0e46d5a7401416982e07f38f78451034cad4bbe3afc8f8878f19a79aa5992b284f7cffed8585749b72be196d21eeff635717dbba6	\\x2000000000000000	f	t
\\x9b726ca0da4092514d5110ec77cf1fc4b5a14fe5b94c647753aafdf5168ddb88	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001200eeca9e47b05d4fe1b9ea5377f0054ddcb89cf3f034399836cba9f3b83d0e45000000000000000000000000000000000000000000000000000000000000000308618f2f494a8abaa9915c87b35f3c3b2899972f6ec1e3696dd54aefebd3fbb48c3b4a02bdf40e17a437c3adc9f0804aa0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b034e08f52ddae2b7c6ebcfba64d72d4a38a8b2e987bc54027b598e5b8760027b4a9100ac64fac8dfa0b73a488e2f69c0e4a32002eb172354d900795aaf68ddfd2d339e53adf9e5b2a5335a8c9aede40dae05562b5ffa0c10b61a3cbe6523390	5	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8618f2f494a8abaa9915c87b35f3c3b2899972f6ec1e3696dd54aefebd3fbb48c3b4a02bdf40e17a437c3adc9f0804aa	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb034e08f52ddae2b7c6ebcfba64d72d4a38a8b2e987bc54027b598e5b8760027b4a9100ac64fac8dfa0b73a488e2f69c0e4a32002eb172354d900795aaf68ddfd2d339e53adf9e5b2a5335a8c9aede40dae05562b5ffa0c10b61a3cbe6523390	\\x1f00000000000000	f	t
\\x77b356ee30611d87222eb117eefbc5869f282e4ac032454fb7531229025b5993	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001202319938dbfe3ab6ade4f9b33d78bbd2a22312334dbab732e1e8894931196fc03000000000000000000000000000000000000000000000000000000000000003087d8e3f68aa7cf2433219e3cdbe6232aab611358f6f4235794ae4225d24e22578a71b52847d459531ec081b4d148afe00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091000000000000000000000000000000000000000000000000000000000000006085b146ffa63b150eb5a40904c4d48fce777710984aca9755cb5044b4bfce12dc170dd509aa3e979d735b7cf53fef2c2b0ff0a5833597289aa590ba3d3481a2d99532c4bdfb16bb3148f04b06bbf9622a3252ac457c628b904fe07c83c11792aa	7	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x87d8e3f68aa7cf2433219e3cdbe6232aab611358f6f4235794ae4225d24e22578a71b52847d459531ec081b4d148afe0	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x85b146ffa63b150eb5a40904c4d48fce777710984aca9755cb5044b4bfce12dc170dd509aa3e979d735b7cf53fef2c2b0ff0a5833597289aa590ba3d3481a2d99532c4bdfb16bb3148f04b06bbf9622a3252ac457c628b904fe07c83c11792aa	\\x2100000000000000	f	t
\\x90c394a2d4555f8b37a25f6986ff85b0d96928b7025b7d3659707134b5f10f3e	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120d5aa33b19cc1f49bee040850a61f9433e9c59ee07b23e94b04261801956a83f500000000000000000000000000000000000000000000000000000000000000308c18668cf8856f3bac3912161df97a6bb9b98141f899be6f3762935542d7857e160c506f31757b454fb3113315e7b9be0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b45f39882e66f58f6dad694960ec2fc66118aec1f92deb0ff58614f1be1560e9769382bc1fb212f30b8eb75797de4f5a0f1777b1588998636493bf72f844ca82bb6b09105b11b1be3401dae55f541a30a17341f3247608d5e38152d0f17ace2c	2	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8c18668cf8856f3bac3912161df97a6bb9b98141f899be6f3762935542d7857e160c506f31757b454fb3113315e7b9be	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb45f39882e66f58f6dad694960ec2fc66118aec1f92deb0ff58614f1be1560e9769382bc1fb212f30b8eb75797de4f5a0f1777b1588998636493bf72f844ca82bb6b09105b11b1be3401dae55f541a30a17341f3247608d5e38152d0f17ace2c	\\x1c00000000000000	f	t
\\x720f02805cf6059cbb077825ddd4f5d3257deacd6998ccd4bc3e106d60d5f1c2	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012086d45c3e393cbb56a72825b6dde015e89872511eb97d1d0869114c2aa83847060000000000000000000000000000000000000000000000000000000000000030af6612098003d793949804ced355362d9da566cf2fd00059de7321241754d8aadb0b77c94587ed7695dc3b1b51d3dbd70000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060881de0a9619d5750ce6fb7f9e85726c41424d4af46a99f87c5d5d0bfeefa71fd207f85defc052d23c903fb56fe785f6d1330f97d390688a92f44fa6b3cc7c4037a99a84f571fff0e4d9e108e34842b1b85ae46c9f7fc91b8089f77f33b0654ad	4	3098331	2020-07-23 19:08:59	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xaf6612098003d793949804ced355362d9da566cf2fd00059de7321241754d8aadb0b77c94587ed7695dc3b1b51d3dbd7	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x881de0a9619d5750ce6fb7f9e85726c41424d4af46a99f87c5d5d0bfeefa71fd207f85defc052d23c903fb56fe785f6d1330f97d390688a92f44fa6b3cc7c4037a99a84f571fff0e4d9e108e34842b1b85ae46c9f7fc91b8089f77f33b0654ad	\\x1e00000000000000	f	t
\\x01309adb6064b32e9fef742bc3d79c8ae4a9c31bd151c5ae80a12979a6da4e06	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001203801934798a5ee37bfa14b3c3c9a15dcfbd57d313f4eef188e07db789d8d5327000000000000000000000000000000000000000000000000000000000000003099ef4f94bc863fcf2c393e03cbad1bfeaea19eece05a3084897bf6a39705eda7a40434bd76a01e6a170bcd319a445b9c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060a8779d81c50ed108ec91621dd4994de2ba2ef1d75beceb0b8ae9e4d755f86563d500844ac6be04bccd6ea83ffb22556016772b23f39fbd46fb737578e2e4c358e5733af4e32d09919b0d80285ab7ae7cd0afac008dd180ea1c43965ee19d42fa	1	3098332	2020-07-23 19:09:14	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x99ef4f94bc863fcf2c393e03cbad1bfeaea19eece05a3084897bf6a39705eda7a40434bd76a01e6a170bcd319a445b9c	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa8779d81c50ed108ec91621dd4994de2ba2ef1d75beceb0b8ae9e4d755f86563d500844ac6be04bccd6ea83ffb22556016772b23f39fbd46fb737578e2e4c358e5733af4e32d09919b0d80285ab7ae7cd0afac008dd180ea1c43965ee19d42fa	\\x2300000000000000	f	t
\\xcc1ebf28bca2491ead6d5d81b8c676787c398502322dd8ad18bdaa39976e3673	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001208aac200919af9eeaa0a16c7777aa5c75f60e1515201bda9d08338dc8597883cf000000000000000000000000000000000000000000000000000000000000003092705f79d60b3ae862ece24de3bfd234aa24ebf2be5f4235ffa72dcd4b882fefe1753d82ad355188bf22fa251e05e82d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060a7bf258efd8d9addb26aca8ae9b6672fbcfad280bd1404ea836426b5725bc64c337fd7836c39acb7180e4cd165685e8a033e12f84ff3a2bb8104ed86717a726fff789581a0d0da38a74cdb7ed92653a8f403e6b69d648e0061c86d896b8d8f34	0	3098332	2020-07-23 19:09:14	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x92705f79d60b3ae862ece24de3bfd234aa24ebf2be5f4235ffa72dcd4b882fefe1753d82ad355188bf22fa251e05e82d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xa7bf258efd8d9addb26aca8ae9b6672fbcfad280bd1404ea836426b5725bc64c337fd7836c39acb7180e4cd165685e8a033e12f84ff3a2bb8104ed86717a726fff789581a0d0da38a74cdb7ed92653a8f403e6b69d648e0061c86d896b8d8f34	\\x2200000000000000	f	t
\\x9597a3d8a37343524a947277fcdd2c03ee68455df3de8794541b13603baddc89	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012007f8141c3d89fdbe251988afcd8dbfdb601d92911a581b345fb0d24698a6388c0000000000000000000000000000000000000000000000000000000000000030a27ede36730c6f4d5e089629b4ea7ee1d17b037c25892e41c49175f831206d14c8e633f397fd7d6d39114cb5c6e518210000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060aed3cc411402651dbdf71b389210426a401620e1db2743532bfee8f12d2ae6485dcb93a18e57e43265b8defd03447d020380f532a99103edfce88775c80352c02c2a9f41b2b9487d56d8774c15808324f5861184cc91d92fa46a674deae9fe17	3	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xa27ede36730c6f4d5e089629b4ea7ee1d17b037c25892e41c49175f831206d14c8e633f397fd7d6d39114cb5c6e51821	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaed3cc411402651dbdf71b389210426a401620e1db2743532bfee8f12d2ae6485dcb93a18e57e43265b8defd03447d020380f532a99103edfce88775c80352c02c2a9f41b2b9487d56d8774c15808324f5861184cc91d92fa46a674deae9fe17	\\x2700000000000000	f	t
\\x403e7916692508aa289dd1a5fe6d236f33b6dcf65d7c8a92c36363bdd80c10c4	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001205f393c0c226094994c629f01b6a2ca4f963712be7134c53feda59bb989a4a132000000000000000000000000000000000000000000000000000000000000003092cfbbf8217e197d73492f838cc381b58c39eb68becc1e9cacc76d5392c94148ebb27d38a2efb1d9cca8673a9cfc1a100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060832edfba013abd5fbaeabc482d1b4a80a7b2d37d779bc8ca8bec909e8545358e160b28824d1411ab728719e9ae6d35d411f381c293ee7a00f75da69913401ff0c6b8bd1b7f3591594c600f02cf1588b26feb157c6666fafd41c7f563aa4356eb	4	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x92cfbbf8217e197d73492f838cc381b58c39eb68becc1e9cacc76d5392c94148ebb27d38a2efb1d9cca8673a9cfc1a10	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x832edfba013abd5fbaeabc482d1b4a80a7b2d37d779bc8ca8bec909e8545358e160b28824d1411ab728719e9ae6d35d411f381c293ee7a00f75da69913401ff0c6b8bd1b7f3591594c600f02cf1588b26feb157c6666fafd41c7f563aa4356eb	\\x2800000000000000	f	t
\\x31401749240844cf9d25687dff984f68a05fe63fe03404700d76fe6ea74512e3	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120ce89a4fb1ccb0db287e25e9815b877930a5f6052ab170cba95c8e3d78fb17179000000000000000000000000000000000000000000000000000000000000003080b81131547bc13b96f1a30ca884c297050acec603c201ac364fa7d7ce995828e3a2e26eb35e0cee569ea6399b4bd0550000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060adc992e8d9f933b450d85ceea8603b229d2f88ca9ae73aa111f2cc8cbf3ff05307c0e518543a849b21e9721d03b3d53410ee03553d05fcbdb2c42b8702750910ad9b5b79874f99f94564f02c7be43ac1a31855193ba77b992c548a05889f384b	5	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x80b81131547bc13b96f1a30ca884c297050acec603c201ac364fa7d7ce995828e3a2e26eb35e0cee569ea6399b4bd055	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xadc992e8d9f933b450d85ceea8603b229d2f88ca9ae73aa111f2cc8cbf3ff05307c0e518543a849b21e9721d03b3d53410ee03553d05fcbdb2c42b8702750910ad9b5b79874f99f94564f02c7be43ac1a31855193ba77b992c548a05889f384b	\\x2900000000000000	f	t
\\x55b1fe94dae5549fa3c04cb95508310732af53032933a9c417edb040601ccfb1	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120918abd24d99061a5f64e672692cfc9e0ccab2ab0db6c85b8f4d34fae141f357a0000000000000000000000000000000000000000000000000000000000000030b30dfc42bdd14ffbf8d4315c666fdbf926409b4f316373a318bed26a956299a4372154e0d7a9113c1273c04a8777cf5d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060ad47fab3a30a64428eee0a8535b57a85b331cb9a4ecf8164c26f143f354d1aa79332e6f22505865c0d0bac53b9733e820b4be83295a0c78e74ddb287d9b621355245d159fdd486f00fafe77b16d4852b3383086d4737061a6adc609d39087f76	6	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xb30dfc42bdd14ffbf8d4315c666fdbf926409b4f316373a318bed26a956299a4372154e0d7a9113c1273c04a8777cf5d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xad47fab3a30a64428eee0a8535b57a85b331cb9a4ecf8164c26f143f354d1aa79332e6f22505865c0d0bac53b9733e820b4be83295a0c78e74ddb287d9b621355245d159fdd486f00fafe77b16d4852b3383086d4737061a6adc609d39087f76	\\x2a00000000000000	f	t
\\xb188ff09e65fe2ddaa4b36315989ebabbfd4c6c7e7e27997c4c912dd5d7c88e0	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120780215657b29bf5d2b192d5997bf67fe07fc50aa235371c3d1da43f7e6f2fcd10000000000000000000000000000000000000000000000000000000000000030a713c85d5be62a7ea34e67054e261dfb700794a830e2380c5e5f71dcf29275345926192e0bbc6877dc1ea03f8a9192e30000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000608885e8eda194488f2a432b4b9fc1a481dd58c22095d85f721d79cc743802c5f3287aed3301463aea8dfb215fba3ceb1503a1ee6c2236ff3c0cea8f90a537777aa9217526c2bd272ed79390e36f74859229f529aa3a9fb33f0fedeb46db2ce43f	7	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xa713c85d5be62a7ea34e67054e261dfb700794a830e2380c5e5f71dcf29275345926192e0bbc6877dc1ea03f8a9192e3	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8885e8eda194488f2a432b4b9fc1a481dd58c22095d85f721d79cc743802c5f3287aed3301463aea8dfb215fba3ceb1503a1ee6c2236ff3c0cea8f90a537777aa9217526c2bd272ed79390e36f74859229f529aa3a9fb33f0fedeb46db2ce43f	\\x2b00000000000000	f	t
\\x05fa1efe99b27e34d43db4b602a083a617d5f554f277c5975782adfa6897c346	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012063298a9db8bd550788472e1a8774e62ada3a4a519f1b64108ded9ed34acc040e000000000000000000000000000000000000000000000000000000000000003095d513534de5fc40c03e1e657f1632e17f2ae65e708aa6a25c13ebbebbda3dbcbdd762bddf1c5fa5c6fda289d9ab64060000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000608395836ec9b8c84d390ea38051a55c0c383b56d34f1f883d462ba4aafca5bc60cdd142fd737b4cbc73d01210099e4fee1655f790fcf33cbfa101484725b156e454f8b00369a0ecd57a1d5f441c599ff65b9aba2fb220acb27975ac4470ac70b0	8	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x95d513534de5fc40c03e1e657f1632e17f2ae65e708aa6a25c13ebbebbda3dbcbdd762bddf1c5fa5c6fda289d9ab6406	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8395836ec9b8c84d390ea38051a55c0c383b56d34f1f883d462ba4aafca5bc60cdd142fd737b4cbc73d01210099e4fee1655f790fcf33cbfa101484725b156e454f8b00369a0ecd57a1d5f441c599ff65b9aba2fb220acb27975ac4470ac70b0	\\x2c00000000000000	f	t
\\xd056b1f41089211a43c17e5f4a22c58339e949260de226c4acb15449a6cd25df	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120b9d30396bf16d773083eb51d47cfbecb6e96c407f0b2e38600512e96c4f9b1540000000000000000000000000000000000000000000000000000000000000030a6fd2c4289250d399876550d0b9fcda435f41a22c24e4c8b0e25c0cee8042a0c9d3f7c02928403b36e22a4a3d7ea8bd00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b9c783c375e8419cfc6c2663fbd2e9a97eabc87923b818e3ee022dba2a82040860eb2817bbc25380f86490c4f4a701f70e2d75a357926115fc149a008e59d6b937e42f060bd48c078d93f4f95d1836bbfef723ec555d89c6dac15298d7214cf4	0	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xa6fd2c4289250d399876550d0b9fcda435f41a22c24e4c8b0e25c0cee8042a0c9d3f7c02928403b36e22a4a3d7ea8bd0	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb9c783c375e8419cfc6c2663fbd2e9a97eabc87923b818e3ee022dba2a82040860eb2817bbc25380f86490c4f4a701f70e2d75a357926115fc149a008e59d6b937e42f060bd48c078d93f4f95d1836bbfef723ec555d89c6dac15298d7214cf4	\\x2400000000000000	f	t
\\xe5aef47e1714b6ae0e4ace465bd1d1d1e529192dd882b55e5c047a72deef255e	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120a8551dcfb9702345a56945a9271c498203112aaa7c4db9cd73726d35dee4e1c0000000000000000000000000000000000000000000000000000000000000003092c83a2b73b866ae670c423ea00194d856bdf552cf1c9f8aa6ea92929f73d52dae530e65f61ee0f558df022afa5f12f80000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000608286606a92686ee51e6b45d66d0760100415ff264949033d07d9a270cece926c90398313741a9354ec81361a3196ad5f0990123f0f3e528ba935dbc36bfd4a287be409ddcdabed24c210191aba1630f27856ebcc1a04336649c13c362fc398d5	1	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x92c83a2b73b866ae670c423ea00194d856bdf552cf1c9f8aa6ea92929f73d52dae530e65f61ee0f558df022afa5f12f8	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8286606a92686ee51e6b45d66d0760100415ff264949033d07d9a270cece926c90398313741a9354ec81361a3196ad5f0990123f0f3e528ba935dbc36bfd4a287be409ddcdabed24c210191aba1630f27856ebcc1a04336649c13c362fc398d5	\\x2500000000000000	f	t
\\x352fd76b4b5912a16125923efdaf2a5dfc2110c12e4843f5b08580350abef8dc	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001204537a1403957c982d1478b7c4652b5f9831501c3d74b9af15c1fb67444448a7a000000000000000000000000000000000000000000000000000000000000003098250d9ee1c76975aaf3a1d68dab0f23a203eca5d6878c1f37c9c1ed2b97d62b1eccb9979097089dcb817ad0a84ccfae0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b4e670f85ec59c8dfca2e3e75ae1c22d2dba0eb5ee34ea4ff42ce5638f4c5f8eac7a54034c4122dfe8133c0e98e991900d8524cc7da0f74f75768a2ea620c114964933f285c8a48098fa8470561d565df754e3ad413f9796194a756e9ce9c493	2	3098333	2020-07-23 19:09:29	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x98250d9ee1c76975aaf3a1d68dab0f23a203eca5d6878c1f37c9c1ed2b97d62b1eccb9979097089dcb817ad0a84ccfae	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb4e670f85ec59c8dfca2e3e75ae1c22d2dba0eb5ee34ea4ff42ce5638f4c5f8eac7a54034c4122dfe8133c0e98e991900d8524cc7da0f74f75768a2ea620c114964933f285c8a48098fa8470561d565df754e3ad413f9796194a756e9ce9c493	\\x2600000000000000	f	t
\\x327215af90b759285a5ef85520472d6aabf4fde61bc05dc55cf41559e7f96f05	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012065333c7b72f2512383d678d5126d92987d44eb0d255a0287023986f163b0353b00000000000000000000000000000000000000000000000000000000000000308b37ebd467bd81b891774c695721e6c78d99111be938ff272c72cfc864683c3e1ff5139c4824199fe29d0520cbd8c8740000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060af6cc9695c4c19c8345430ae3f66cfb007cb4b11ec59e021a059e755d26b2d17faea338c4a7c63d89debfdebc930b92d100fd83c8eccc0de10d2c8676d9a449050e38c09f9a7632e419f3e80dd797619d85351e1b8cd5e1d0df94019deea0474	0	3098334	2020-07-23 19:09:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8b37ebd467bd81b891774c695721e6c78d99111be938ff272c72cfc864683c3e1ff5139c4824199fe29d0520cbd8c874	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xaf6cc9695c4c19c8345430ae3f66cfb007cb4b11ec59e021a059e755d26b2d17faea338c4a7c63d89debfdebc930b92d100fd83c8eccc0de10d2c8676d9a449050e38c09f9a7632e419f3e80dd797619d85351e1b8cd5e1d0df94019deea0474	\\x2d00000000000000	f	t
\\x87f7672b8be3265789896dc17c078a6b5a614d3a014994e0d3b096a189adbecf	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001204735961e51aacf62462b2209cf75636692e2e921d09020d4fc32a35e19a0ad5e00000000000000000000000000000000000000000000000000000000000000308371247ae81e360ea915f7408a75fb732eee95885dfa7eb946f790293b4038ca0ad6be6cc26e31247f1cfd98a060c9f70000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000608b2391daa92bfb1440fe60a3b32e55225f5cddd01f9b34867c4bffde0b75d5831d301a1c6eb84a797006c084f8c8469107c582fddb0dfcd0516036f0505953621c33ca2a56d3985bfd99d57cd0e4cbb1ca74ff35167a45aa75926e377ffe09b9	1	3098334	2020-07-23 19:09:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8371247ae81e360ea915f7408a75fb732eee95885dfa7eb946f790293b4038ca0ad6be6cc26e31247f1cfd98a060c9f7	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8b2391daa92bfb1440fe60a3b32e55225f5cddd01f9b34867c4bffde0b75d5831d301a1c6eb84a797006c084f8c8469107c582fddb0dfcd0516036f0505953621c33ca2a56d3985bfd99d57cd0e4cbb1ca74ff35167a45aa75926e377ffe09b9	\\x2e00000000000000	f	t
\\x1bc89ea1d442900e8c9f52954e2835a49ad8ea4b892c2db3c4f8347f5633f1d8	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001207dfa9350494c9ef808870ad00faab235c0278185d8cd8b0f65b99536e7583270000000000000000000000000000000000000000000000000000000000000003092bb2a1e7c4e984d7896b8e6468748fc654f045ffed539d983d6eab198caaafbfd7970c8d7e1cb768d7180091d30dc0d0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde9637009100000000000000000000000000000000000000000000000000000000000000608112f576f552fa5205de485cf877bd0e19266ebfc3d2b35012de257fa57b700e2ec5cc1b03273a2d27fb4947ad3a523213d97fe40224d287f01ba230ab9f98c7b50f8178e06b738066f0026454f2f38e8bc013d334a8a7942b2534670d65e1fc	2	3098334	2020-07-23 19:09:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x92bb2a1e7c4e984d7896b8e6468748fc654f045ffed539d983d6eab198caaafbfd7970c8d7e1cb768d7180091d30dc0d	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\x8112f576f552fa5205de485cf877bd0e19266ebfc3d2b35012de257fa57b700e2ec5cc1b03273a2d27fb4947ad3a523213d97fe40224d287f01ba230ab9f98c7b50f8178e06b738066f0026454f2f38e8bc013d334a8a7942b2534670d65e1fc	\\x2f00000000000000	f	t
\\x5a9b6452d1bde02ed89f4f9724184145aa0bf3a50272414521ebcd80f941248f	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001200c897918f70f68f4b7b624409b8ba02ad78e8e95208bb1ef571b36f15d818ec000000000000000000000000000000000000000000000000000000000000000308023022f1f335337748eb141477e789d683514ae739b9dac7b72bb3326201065abbf6f3301a5185bb8751776e5bf45cf0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b7a08b5a72761f23dec2605fdf980d2b9375b67a2cf74c3ce953d89c94cb8a797533de9d79b280b349ba070ba82a4308149b623e14428fe55b0bf9dea4f68dc887475d7c2c094e14a29368c07bcddd93ec50373a6e0d1ac3790ce56a66e19b25	3	3098334	2020-07-23 19:09:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\x8023022f1f335337748eb141477e789d683514ae739b9dac7b72bb3326201065abbf6f3301a5185bb8751776e5bf45cf	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb7a08b5a72761f23dec2605fdf980d2b9375b67a2cf74c3ce953d89c94cb8a797533de9d79b280b349ba070ba82a4308149b623e14428fe55b0bf9dea4f68dc887475d7c2c094e14a29368c07bcddd93ec50373a6e0d1ac3790ce56a66e19b25	\\x3000000000000000	f	t
\\x75d8333548df3ca590214056cac6a3a134bf1f1b8a0f7e79ec2300060b2f113f	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120885ab289b851a7c6c2343eb7b255d8ffb183916bd74b36edc65dbcf4c1b4bd760000000000000000000000000000000000000000000000000000000000000030b1e8517c990087ec06de1c80910b5f5a4dca67f73451b61063d575e567da0cbb7ae65d2a4246ed1d176b21e896a4a4f80000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde963700910000000000000000000000000000000000000000000000000000000000000060b1d765e6726a92c4502678e5cdd2d51a0f5033d1a944bb03a2a118fa16e93e4675e8b61ddf66eae47c19b2d53cdea47f0b9e5403f996eb3e3c47c9335d04bdbb1cbf6548e2ca70dafeb49201e3e25ffdd492a82ee48655746845edbe8bb11903	4	3098334	2020-07-23 19:09:44	\\x388ea662ef2c223ec0b047d41bf3c0f362142ad5	\\xb1e8517c990087ec06de1c80910b5f5a4dca67f73451b61063d575e567da0cbb7ae65d2a4246ed1d176b21e896a4a4f8	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	32000000000	\\xb1d765e6726a92c4502678e5cdd2d51a0f5033d1a944bb03a2a118fa16e93e4675e8b61ddf66eae47c19b2d53cdea47f0b9e5403f996eb3e3c47c9335d04bdbb1cbf6548e2ca70dafeb49201e3e25ffdd492a82ee48655746845edbe8bb11903	\\x3100000000000000	f	t
\.
COPY public.proposal_assignments FROM stdin;
0	0	0	1
0	9804	1	1
0	8597	2	2
0	4488	3	1
0	1072	4	1
0	15248	5	1
0	10383	6	1
0	11087	7	1
0	5193	8	1
0	4067	9	1
0	8309	10	1
0	17836	11	1
0	2585	12	1
0	12540	13	2
0	6859	14	1
0	19806	15	1
0	11621	16	1
0	5184	17	1
0	13487	18	2
0	14521	19	1
0	12570	20	2
0	18403	21	2
0	18618	22	1
0	2895	23	1
0	86	24	2
0	5809	25	1
0	6862	26	1
0	11630	27	1
0	19455	28	2
0	6427	29	1
0	8465	30	2
0	17282	31	1
1	13693	32	2
1	10313	33	1
1	4013	34	1
1	3775	35	1
1	12140	36	2
1	11460	37	1
1	5791	38	1
1	15258	39	2
1	11858	40	2
1	16738	41	2
1	19848	42	1
1	14101	43	1
1	7843	44	2
1	11879	45	2
1	12292	46	1
1	8954	47	2
1	4378	48	1
1	6123	49	1
1	1010	50	2
1	13138	51	2
1	6905	52	1
1	15298	53	1
1	12084	54	2
1	4549	55	1
1	5094	56	1
1	13599	57	1
1	4778	58	1
1	2995	59	1
1	17237	60	1
1	1985	61	1
1	12433	62	1
1	5563	63	1
\.
COPY public.queue FROM stdin;
2020-11-16 20:00:00	17230	19997
2020-11-16 19:00:00	17262	20009
2020-11-16 18:00:00	17302	19853
2020-11-16 17:00:00	17342	19665
2020-11-16 16:00:00	17374	19625
2020-11-16 15:00:00	17418	19649
2020-11-16 14:00:00	17454	19685
2020-11-16 13:00:00	17490	19681
2020-11-16 12:00:00	17530	19665
2020-11-16 11:00:00	17566	19633
2020-11-16 10:00:00	17309	19669
2020-11-16 09:00:00	17353	19709
2020-11-16 08:00:00	17389	19741
2020-11-16 07:00:00	17433	19781
2020-11-16 06:00:00	17461	19817
2020-11-16 05:00:00	17501	19853
2020-11-16 04:00:00	17537	19893
2020-11-16 03:00:00	17573	19929
2020-11-16 02:00:00	17617	19963
2020-11-16 01:00:00	17649	20003
2020-11-16 00:00:00	17761	19972
2020-11-15 23:00:00	17761	20012
2020-11-15 22:00:00	17761	20048
2020-11-15 21:00:00	17797	20084
2020-11-15 20:00:00	17837	20124
2020-11-15 19:00:00	17873	20152
2020-11-15 18:00:00	17913	20188
2020-11-15 17:00:00	17949	20228
2020-11-15 16:00:00	17985	20248
2020-11-15 15:00:00	18021	20253
2020-11-15 14:00:00	18061	20293
2020-11-15 13:00:00	18097	20318
2020-11-15 12:00:00	18137	20350
2020-11-15 11:00:00	18173	20386
2020-11-15 10:00:00	18209	20422
2020-11-15 09:00:00	18249	20462
2020-11-15 08:00:00	18293	20495
2020-11-15 07:00:00	18325	20535
2020-11-15 06:00:00	18365	20571
2020-11-15 05:00:00	18397	20607
2020-11-15 04:00:00	18437	20644
2020-11-15 03:00:00	18473	20680
2020-11-15 02:00:00	18509	20712
2020-11-15 01:00:00	18549	20748
2020-11-15 00:00:00	18474	20784
2020-11-14 23:00:00	18514	20824
2020-11-14 22:00:00	18550	20831
2020-11-14 21:00:00	18586	20867
2020-11-14 20:00:00	18630	20907
2020-11-14 19:00:00	18662	20941
2020-11-14 18:00:00	18698	20977
2020-11-14 17:00:00	18738	21016
2020-11-14 16:00:00	18774	21038
2020-11-14 15:00:00	18810	21074
2020-11-14 14:00:00	18850	21099
2020-11-14 13:00:00	18886	21135
2020-11-14 12:00:00	18926	21175
2020-11-14 11:00:00	18962	21203
2020-11-14 10:00:00	18998	21239
2020-11-14 09:00:00	19042	21279
2020-11-14 08:00:00	19074	21315
2020-11-14 07:00:00	19075	21354
2020-11-14 06:00:00	19111	21390
2020-11-14 05:00:00	19147	21426
2020-11-14 04:00:00	19191	21466
2020-11-14 03:00:00	19223	21502
2020-11-14 02:00:00	19259	21538
2020-11-14 01:00:00	19299	21578
2020-11-14 00:00:00	19313	21611
2020-11-13 23:00:00	19345	21647
2020-11-13 22:00:00	19385	21685
2020-11-13 21:00:00	19372	21721
2020-11-13 20:00:00	19412	21761
2020-11-13 19:00:00	19448	21797
2020-11-13 18:00:00	18851	21829
2020-11-13 17:00:00	17756	21865
2020-11-13 16:00:00	17792	21901
2020-11-13 15:00:00	17828	21913
2020-11-13 14:00:00	17872	21953
2020-11-13 13:00:00	17908	21973
2020-11-13 12:00:00	17944	22011
2020-11-13 11:00:00	17984	22047
2020-11-13 10:00:00	18016	22083
2020-11-13 09:00:00	18056	22123
2020-11-13 08:00:00	18092	22157
2020-11-13 07:00:00	18128	22193
2020-11-13 06:00:00	18168	22233
2020-11-13 05:00:00	18208	22252
2020-11-13 04:00:00	18244	22292
2020-11-13 03:00:00	18280	22328
2020-11-13 02:00:00	18316	22364
2020-11-13 01:00:00	18356	22404
2020-11-13 00:00:00	18396	22440
2020-11-12 23:00:00	18432	22480
2020-11-12 22:00:00	18468	22512
2020-11-12 21:00:00	18504	22532
2020-11-12 20:00:00	18544	22572
2020-11-12 19:00:00	18580	22608
2020-11-12 18:00:00	18620	22644
2020-11-12 17:00:00	18656	22680
\.
COPY public.validator_balances FROM stdin;
0	0	32000000000	32000000000
0	1	64000000000	32000000000
0	2	32000000000	32000000000
0	3	32000000000	32000000000
0	4	32000000000	32000000000
0	5	32000000000	32000000000
0	6	32000000000	32000000000
0	7	32000000000	32000000000
0	8	32000000000	32000000000
0	9	32000000000	32000000000
0	10	32000000000	32000000000
0	11	32000000000	32000000000
0	12	32000000000	32000000000
0	13	32000000000	32000000000
0	14	32000000000	32000000000
0	15	32000000000	32000000000
0	16	32000000000	32000000000
0	17	32000000000	32000000000
0	18	32000000000	32000000000
0	19	32000000000	32000000000
0	20	32000000000	32000000000
0	21	32000000000	32000000000
0	22	32000000000	32000000000
0	23	32000000000	32000000000
0	24	32000000000	32000000000
0	25	32000000000	32000000000
0	26	32000000000	32000000000
0	27	32000000000	32000000000
0	28	32000000000	32000000000
0	29	32000000000	32000000000
0	30	32000000000	32000000000
0	31	32000000000	32000000000
0	32	32000000000	32000000000
0	33	32000000000	32000000000
0	34	32000000000	32000000000
0	35	32000000000	32000000000
0	36	32000000000	32000000000
0	37	32000000000	32000000000
0	38	32000000000	32000000000
0	39	32000000000	32000000000
0	40	32000000000	32000000000
0	41	32000000000	32000000000
0	42	32000000000	32000000000
0	43	32000000000	32000000000
0	44	32000000000	32000000000
0	45	32000000000	32000000000
0	46	32000000000	32000000000
0	47	32000000000	32000000000
0	48	32000000000	32000000000
0	49	32000000000	32000000000
0	50	32000000000	32000000000
0	51	32000000000	32000000000
0	52	32000000000	32000000000
0	53	32000000000	32000000000
0	54	32000000000	32000000000
0	55	32000000000	32000000000
0	56	32000000000	32000000000
0	57	32000000000	32000000000
0	58	32000000000	32000000000
0	59	32000000000	32000000000
0	60	32000000000	32000000000
0	61	32000000000	32000000000
0	62	32000000000	32000000000
0	63	32000000000	32000000000
0	64	32000000000	32000000000
0	65	32000000000	32000000000
0	66	32000000000	32000000000
0	67	32000000000	32000000000
0	68	32000000000	32000000000
0	69	32000000000	32000000000
0	70	32000000000	32000000000
0	71	32000000000	32000000000
0	72	32000000000	32000000000
0	73	32000000000	32000000000
0	74	32000000000	32000000000
0	75	32000000000	32000000000
0	76	32000000000	32000000000
0	77	32000000000	32000000000
0	78	32000000000	32000000000
0	79	32000000000	32000000000
0	80	32000000000	32000000000
0	81	32000000000	32000000000
0	82	32000000000	32000000000
0	83	32000000000	32000000000
0	84	32000000000	32000000000
0	85	32000000000	32000000000
0	86	32000000000	32000000000
0	87	32000000000	32000000000
0	88	32000000000	32000000000
0	89	32000000000	32000000000
0	90	32000000000	32000000000
0	91	32000000000	32000000000
0	92	32000000000	32000000000
0	93	32000000000	32000000000
0	94	32000000000	32000000000
0	95	32000000000	32000000000
0	96	32000000000	32000000000
0	97	32000000000	32000000000
0	98	32000000000	32000000000
0	99	32000000000	32000000000
\.
COPY public.validator_names FROM stdin;
\\x98b5c7619415d97972d56e9cb9b7ae529898931aaa2880a66cd9a2a3a95e6de7b97a56d196843fc60667bc83a0739f3e	foo3
\\x94c9185719b303d15a48137ee22c69256340eec5881447aa1d6f50d12df4994dd76b769f1fdfc1be4c3631df50239dc4	foo3
\\x82cb7ceecc0e4f4787d590685f0ff55895da87ff4138eb39f95f2b20f71781f68ad430a15b4fe3e19db1b00b7756db2b	foo3
\\x8cbf21efca571e50129b1b213732cf1dea094ff8e46be7b3458ca12fcbae57e3b58a84d01f2acaf635edb2b628431663	foo3
\\xaf6defcb01966eb048bbd066a1c41b4275f62dc3971919577d19d418f1b0641433af2a012c0f0e89d92558d8f6fc85d1	foo3
\\xa2bd5ece3adffbb05bf1ec8dc4b8833ccde4b1ef1fc7146503e52f65a1cdc3f212a7f56c6e27e7c95b32a3503d7b14d5	foo3
\\xb101f8b9ffccac9f91bcc73ca9f28dcc59e1da1f7c0b4524efa3280c9d285cf76d0f5b6651a405b90d8502e1e824809a	foo3
\\x98994d60220e14097d022578ced1d8fae1578feedfb2f1040ac1faa00ad6d949e2a3dcf7329be690bd41116b6c3d631b	foo3
\\x8d6ef5f4e8ba50292d3b329b4872be1423e7d3944fdb478f54ee077549dd40973fc75e58107baa4fb627b679431c95a7	foo3
\\x9854ed9b8ab11203731b28eed13129684190ac7fb7b28382a5cb865b947df2c22874cc69020e4bb3c25fcb791025e93e	foo3
\.
COPY public.validator_performance FROM stdin;
7266	28638060852	-7196073	28891764	-3110838109	-3361939148
16965	40370778623	-9822567	-59957045	-23031923577	-23629221377
37336	27335825992	5526220	45873957	-4962907920	-4664174008
65194	28780498080	5631925	42660800	-2912940325	-3219501920
42581	19242272876	-5831992	-31244182	-12797548490	-12757727124
51380	31707306474	6439930	52516684	-528286768	-292693526
68087	18138127119	2944233	30342824	-13168478601	-13861872881
7442	28554199651	-7236872	28842175	-3167298586	-3445800349
42582	19289680074	-5831992	-31229098	-12753194850	-12710319926
91752	32007716116	6522970	7716116	7716116	7716116
12077	16626360240	-4911093	-29977317	-13338153119	-15373639760
40985	28918757041	2914567	42346507	-3027032715	-3081242959
47870	22546408958	4337834	5813671	-9120522063	-9453591042
65017	29643151457	7694890	44918680	-1919005688	-2356848543
80902	29198839283	5156872	37568797	-2760126894	-2801160717
42119	15623617700	-4604125	-28103560	-16202705390	-16376382300
8507	15755328719	2758595	19273137	-14007936682	-16244671281
73460	23740710036	4738342	11361390	-7580347792	-8259289964
78708	18145287423	2932859	25874153	-13096369774	-13854712577
79359	15544138687	-3825364	-27324799	-16113822778	-16455861313
82612	32065167548	4514227	47732105	65167548	65167548
31834	63060282030	0	0	0	-939717970
33009	29003953834	3452479	42740224	-3026924534	-2996046166
42159	28938644666	5689116	44857026	-2986793326	-3061355334
80165	51985435290	-7080478	-6051627	-11414213167	-12014564710
6664	14899376496	0	0	-15359426641	-17100623504
24419	28455156791	-7299980	28644146	-3205266413	-3544843209
27936	22091045415	4461760	5931153	-8826287009	-9908954585
48355	22482740479	4252344	5580512	-9177409459	-9517259521
61446	15799874265	-4911093	-29977317	-16279699008	-16200125735
5371	14955856705	0	0	-15419583892	-17044143295
52253	27803220661	5420708	42779356	-4134411890	-4196779339
29648	27018583523	5138234	44770443	-3988169747	-4981416477
24279	15498906089	0	0	-13888289289	-16501093911
74256	23697108233	4627918	11748993	-7634523824	-8302891767
33430	29058587741	3219493	50560119	-2967311785	-2941412259
40381	29929125209	5835189	44299009	-1871556305	-2070874791
62605	15793543807	-4911093	-29977317	-16266337209	-16206456193
80156	52045672557	-7201154	-13191789	-11347084470	-11954327443
91188	32012372057	6417550	12372057	12372057	12372057
32934	15378580808	0	-18931767	-15920298114	-16621419192
75164	27333891624	5419329	47853641	-4653783112	-4666108376
34638	30548421346	6359043	48936263	-1344039056	-1451578654
45786	27753800940	-8594691	-52462097	-4421374368	-4246199060
63319	15343585495	0	-13636290	-15871791632	-16656414505
75249	20867906944	-6445930	-39345953	-10423227351	-11132093056
85307	32055233738	6511787	47039355	55233738	55233738
32108	15397719226	0	-20382978	-15922247362	-16602280774
36842	27700494275	5200442	46349393	-4147202280	-4299505725
41684	29897603803	6048513	49409329	-1906533581	-2102396197
16577	20140111543	0	0	0	-11859888457
40805	29862481241	5756811	46924416	-1932264732	-2137518759
69608	28843581510	3421662	42051692	-2844016637	-3156418490
9799	27761484032	-8594691	-52462097	-2517047443	-4238515968
14467	19849806569	0	0	0	-12150193431
3144	18575546063	0	0	0	-13424453937
11904	17909966918	3029954	24469325	-13121054332	-14090033082
50143	18239535017	4167523	-5487369	-13718522951	-13760464983
62960	15790495975	-4911093	-29977317	-16270252024	-16209504025
66581	23544541712	-4052161	25509334	-7996966317	-8455458288
25609	15396416336	0	0	-14007160876	-16603583664
28394	23357224523	4639027	37017039	-7934957195	-8642775477
90737	31979557511	-9822567	-20442489	-20442489	-20442489
4924	30849220093	6197133	51716792	-457328267	-1150779907
13839	17940816973	2786146	24637138	-13206937654	-14059183027
20478	21015803097	2720362	-16728160	-10658614656	-10984196903
37708	22371139803	4577281	41165541	-9848986550	-9628860197
85604	32055405619	4644146	49546723	55405619	55405619
27197	22069049570	8202054	9273324	-8887512893	-9930950430
35707	22866966652	-7059864	-43093443	-9370961648	-9133033348
236	28591580221	-7268076	29205023	-3143074365	-3408419779
51254	31695232590	6392173	59451089	-512546716	-304767410
52620	15767535600	-4911093	-29977317	-16363522512	-16232464400
71298	28799278678	3220951	42216238	-2877124041	-3200721322
88632	64001932162	2029412	1932162	1932162	1932162
21661	18993598643	3543194	27596960	-10521981971	-13006401357
92255	32004578896	4578896	4578896	4578896	4578896
73166	18050486478	2789012	26157919	-13228998438	-13949513522
87399	32044833303	6533594	44833303	44833303	44833303
12196	16613618454	-4911093	-29977317	-13339401860	-15386381546
16457	20141117262	0	0	0	-11858882738
51335	15236910845	0	0	-15728913470	-16763089155
79037	15364632953	0	-18128325	-15902944885	-16635367047
91962	32009801895	9801895	9801895	9801895	9801895
8609	16153661591	-4911093	-29977317	-13677747745	-15846338409
12939	14613041287	0	0	-15036130034	-17386958713
48946	28958693930	5547106	46655419	-2885860742	-3041306070
61651	15799175942	-4911093	-29977317	-16276916142	-16200824058
68080	15298314849	0	-4843998	-15805008307	-16701685151
7449	28639789370	-7328746	33931591	-3098071225	-3360210630
17649	16188661207	-4911093	-29977317	-15277969663	-15811338793
56667	16251231650	2385960	19150857	-15484632072	-15748768350
61275	15805274324	-4911093	-29977317	-16273115039	-16194725676
12395	14630280841	0	0	-15059468652	-17369719159
47543	29089148156	6091860	43769057	-2754213780	-2910851844
64048	32131648084	0	0	0	131648084
73866	23701818700	4630893	11547643	-7617012143	-8298181300
81406	32092592676	6494222	58814033	92592676	92592676
957	28601387470	-7255034	28736904	-3124729750	-3398612530
10837	20187130322	0	0	0	-11812869678
\.
COPY public.validators FROM stdin;
5186	\\xaa9c3cd4769e95918d1dfdf52e5275e5a72160c15c60eec3d5ef45976230f4c0b8672737442a4e16091100348cf7a754	12403	\\x00467cb216d1f5d3ad850786e84cb5b3c35a7afd0cc166ab431a65716499a6d0	20675166866	20000000000	t	0	0	4216	\N	\N
7237	\\x92fd108abb19937f9810e494d89cd2ae5cfbf1f6e301bab575d54a47a5568fc47b1d06362a01a8b993872a05d2ee8350	11442	\\x00ac53ae5e6333ef99698f62672cd44a1de5c4a79166bec2469b6cf5697a6247	20091973763	20000000000	t	0	0	3433	\N	\N
10432	\\xacb4204d4833a78caf2f05e5555587249605749efca042e74b202caecea0d8601fa462f249ce8929ea1d1a5bc9c21b7d	11474	\\x0049d19c29c714c4eb995c805b8310d8f3d2101825d77fc43f39749cb39772ae	20130133869	20000000000	t	0	0	3673	\N	\N
13647	\\x94514e77e49a20e6af43ebb89fd0723318e9593cdd8b25ba6fdc82c335e416315b1697999b7eaa94bc0105d9e3d32a6b	11563	\\x008c835ba9e30f3cbb70f8dd65de63c40d065aa16cafec5d2288c97d5269b7e1	51031123631	32000000000	t	0	0	4127	\N	\N
18172	\\x80efdb9609537433b09c27c6a9bfe415aa3f3d3b1cc04ff5209b0021e1636506995dc685210af7c1623ae37be2782390	11529	\\x0065a2dfb0594bb633f82fe279da1faa7c0c9f04a2f25fae285fde919475d8ab	19934386383	20000000000	t	0	0	3946	\N	\N
20032	\\x84da8c61b1ac76e6c364a002eafce8bc90155d42553a17832bbd23c29de76362b231814f23a287fd1588e43d9bda6a3c	2331	\\x00deccddfbe432e3d2c63fd9464ba113e97add92ba804ede904dabb8ab20f1aa	31884319786	32000000000	f	0	0	2075	\N	\N
21212	\\x8c1610d0e9d70777206f483c2eeba97c0daf2659e83527f71097b6b36e9e5a7190dd95ffc84b8522d422cff2577ee83b	15220	\\x000724923d28c7cc963f49c404104fd70bf125c5d56af87aab65e5187b0f106c	24038682799	24000000000	t	35	320	7033	\N	\N
22949	\\xb4a68cd4b91854743ce7182190fc635610b1e42410f5813d2190f5cb2c9fb872049427a66bbaab1d27e274914a978cae	11542	\\x003bfadaad3754dbde537e7552229fb7003f946892ca04ef81aa588a19a57024	19505057582	19000000000	t	119	754	4035	\N	\N
38496	\\xaaac37afa9ad7bda561348a5e2b3ebcd0e3b37f16cc6f4d4974297f7270532d73d41c55733cc692e36448e2752dc33ab	9491	\\x00e56703cb7463fb590d18d3fdcd65ff29fee0764b4892fe44774ea1d6b4601d	63811605937	32000000000	f	3832	4834	9235	\N	\N
38620	\\xb7b0c73e5c2c9b0aa4962387f97705d86a870709463326f78bd350bcb80ce2990d34a6aa65d2b83d498f293ff28f2bcc	9492	\\x0082ac27dd1a41938e34a4b77468ca0699fc51b8cc17ce17848423e9766b8be2	63813050416	32000000000	f	3832	4865	9236	\N	\N
40475	\\xb17001418c1b9749bda4458908931b28bcb7929d631b2e7660f957256716c5e0bfa0da92d7ec7cb09e5c376f0a339be8	13627	\\x00ae31b41f709f058754669384fbc1da057dfe28a8ee8416d3f68149f60a3757	22736553869	22000000000	t	3929	5329	5440	\N	\N
47699	\\xa280f37b589282eb5132abfa3832d6eaae39567aa7b581eb61d68124d5afac74d48900cc74813c01fd60132e3edfd0fc	9537	\\x00c4c15782e4aa72a447868d1f873fe73d0690e02c51043d5ff1e6ae8d0bf101	63913352215	32000000000	f	4378	7135	9281	\N	\N
82412	\\x8d600108f8c128751eb816dc165d72de5312433389d365d26c3cd5d502a054a6ca9f92201fc56b31a6c9915f22449778	9223372036854775807	\\x00a1dc26bb64d4dc3b544418c94f074f722518022239deed004413e32df60d08	31897608584	32000000000	f	16097	20873	9223372036854775807	\N	\N
85873	\\x905fa4c6aa9985972208821913c4b0be03d10bdb33bd5c253f46e1e0e50698baab2db2d0f7c058bbeb661e7e903619fa	9223372036854775807	\\x00d77be6277f1cdcfce33fdcb127b95fe91e09eec04aecc521dc94866f0055f0	31930194050	32000000000	f	18791	21738	9223372036854775807	\N	\N
86363	\\x9059fd4595cef2cb206208429844f393848ff540d531ea46bd42b30a41244226bf52345f3acca22468eff9cc37afde4c	9223372036854775807	\\x00d77be6277f1cdcfce33fdcb127b95fe91e09eec04aecc521dc94866f0055f0	31938711420	32000000000	f	18797	21861	9223372036854775807	\N	\N
87067	\\x8d88eba6e237a5a36e8ee10f8aaaaa64f611fedac35086c313903b14be288661bdd4a7e4ff579ce3b2fe27e90a5c6939	9223372036854775807	\\x00d77be6277f1cdcfce33fdcb127b95fe91e09eec04aecc521dc94866f0055f0	31945787396	32000000000	f	18806	22037	9223372036854775807	\N	\N
87561	\\xac5c9def872d2ba15b2aae96c79ca8971507f5e617e9242cd2a2a962494e7b5ec4743664f9c3af999be27a7d29dae510	9223372036854775807	\\x00ba9c94adfdf48914fa65960637f439c866378c95857c06db458b33a6de2843	31950347720	32000000000	f	18811	22160	9223372036854775807	\N	\N
88856	\\xad03abce26509fdf3fac169b4ff6c453a541a71a3aab392a1f7d515e42f03c6b24702abce86aad3149ded2460ce012ca	9223372036854775807	\\x00d77be6277f1cdcfce33fdcb127b95fe91e09eec04aecc521dc94866f0055f0	31962050105	32000000000	f	18820	22477	9223372036854775807	\N	\N
89228	\\x8645464b42a50b2c258690566d0f7a5f732cfd531cec1f536d93d91fc317ad4d53b606665d885fee01988fd963514076	9223372036854775807	\\x005eb76d2a0cbd191477747d8392695e4df4267b445a0ebc3f068438b0445feb	31965469589	32000000000	f	18822	22570	9223372036854775807	\N	\N
90137	\\x8b90a9a5bf2230021f96db0dba6c814ad1645928802e6935690e0d3e10d4461fcd7335ca8312f6fb4e22b8e9dbfa4bc6	9223372036854775807	\\x002e7e46732816b5ed78db014598571d95baecba305379077fb86ede695373df	31973790308	32000000000	f	18831	22797	9223372036854775807	\N	\N
91022	\\x815e4f41754d71826fed57a45507a244463275d5533cb830d18bc10b31c0faa46b0e4718d7618cccd9bc9349ae75f050	9223372036854775807	\\x0057512eca92473c452f50a92b3c9f961cb67d78e97266f746ee46011a207d2b	31981856366	32000000000	f	20577	23018	9223372036854775807	\N	\N
92206	\\xa66d3d457b0791462d64373dd35351df693b62aca86cf47646bcca69fad23514be2ce763fdfabbc3bbe26bc1ecb38591	9223372036854775807	\\x003042555070410a4c2aee1267f972d07e7e5a60a5fdc31e5d2ba87e746ca5fe	31994287154	32000000000	f	20581	23314	9223372036854775807	\N	\N
92552	\\x9984dc76ab1da628bd5da316c9283c0916a346880e6727e1142105b2f432ee522b45bda03e50e01b4119329648732564	9223372036854775807	\\x00ca1ca6ae1d9868aa0cd1117639b9e01f7b5454425b83b225f1fadbcbaf6291	31997435387	32000000000	f	20583	23401	9223372036854775807	\N	\N
92824	\\xb0b31cf09aff9e3e04004c45d512720f2d58baa0861486754fdfb44802a9fe114134cf5236cd1f43a1e1aae78bae976c	9223372036854775807	\\x00e28e28148fe169cb04e494a302945e7a5f63efd7e1c3eaebbd7656682fef06	31999891703	32000000000	f	20700	23469	9223372036854775807	\N	\N
65006	\\xb7ce49d50119b62444bb90fd5529621db9a241804eb19cbc376f3011cfbaa71ee60a509c2fd0ccc87b39cb5627fdb2ee	9223372036854775807	\\x00b44fb5342b7ee9e6e42d4c3e0cd755415270c01cf7bdfc01f3c68689d03c27	29662759235	29000000000	f	8824	11461	9223372036854775807	751124	\N
65007	\\xa10cf87024b6d864969f790a1de8d700a06fe49dea1dae9d062f1b2ad253a030783404d53eb68080f4601b2db07b21bc	9223372036854775807	\\x00708dd98687dd0e2a7b4371a00b12cb63bffaecfa2dd6d9b81bad4d660476cf	29621590926	29000000000	f	8824	11461	9223372036854775807	751108	\N
65008	\\x9367f0cad11da297aa5cca1117b91bb86d1ea4ceb829bd90de6ce8b4890c96fefdf4fe4833b9bfb935801093972936c3	9223372036854775807	\\x004244e960ed243d64e0d932aec5dac0106d6bfe4ef3def1abc240ac11552d6b	29703229512	29000000000	f	8824	11461	9223372036854775807	751106	\N
65009	\\xb889cc5a594c215068bce7e87d854940790edbb5a4ebc9b8985dda403dd390727aa8b07a46caaa7258ef29d83061dc2d	9223372036854775807	\\x00c34dffa53ba65f8788ceb1d3918b8aa2af4c2261ed2b220e4ca482c11839a8	29670298411	29000000000	f	8824	11462	9223372036854775807	751114	\N
65010	\\xa3d539ca3ae31efe0c6e09de7a454940289badd2c94211fdeaef07acd35dd68f8a63884731d6faebef6c1e037eca0e2a	9223372036854775807	\\x009e15c1aa6ac5adc6d1ab4c22e784a0745e00d603ed66191878965d02ba3e76	29732931574	29000000000	f	8824	11462	9223372036854775807	751123	\N
65011	\\xa3cf19e54e63e902cbc2add3b42caed7a3ce8955e5486ac4281d222304f3d199d05702788cd2142df2224729ed701c9a	9223372036854775807	\\x00df262ea139918274138d013887808c8c650005c36999f0a21b014df3088cfd	29638537338	29000000000	f	8824	11462	9223372036854775807	751115	\N
65012	\\x956a1b150103a6549ec6ee402044ee8e71a12e841d8fa2b10094d389498aa31d86e060add6d6906e875c2632c4fb5f48	9223372036854775807	\\x0010cec72007a5bce35f056d708d97025de547efc43eb2725f6e671b8d7993b6	29658220609	29000000000	f	8824	11462	9223372036854775807	751113	\N
65013	\\xadf1c8db38bd86caf8e5d01b8015d7f38150ce3f07ee18e3e66700941b22faf564eb50e98a5b30a1364aa9c928c5eb77	9223372036854775807	\\x00847c44978057837a982262f582e9b422de687a8fc92b29e7e9ced2ce29bcd6	29624594556	29000000000	f	8824	11463	9223372036854775807	751137	\N
65014	\\xa9e7a1bfee122b0c0199973688a83dd72abdc9509b246730bd3ca162caf7e4a8eb645ce2dab81196b19cf6540b4b0e4e	9223372036854775807	\\x00aa61f21f11043cde0c25b06e660313dd87e2d7e5813abd868d8d8115384235	29699646413	29000000000	f	8824	11463	9223372036854775807	751113	\N
65015	\\xb2d747869b4413c3edfad7f8711a947e8669788842519be3d6a379638bf5536a5ee82f5dce77b4941def41ce876d82fc	9223372036854775807	\\x000d8f3fa1856554fec79fa297988319b5f30467b1dd82125befa0bd7a1f0b00	29632952384	29000000000	f	8824	11463	9223372036854775807	751107	\N
65016	\\x866264154e919f3eecdf862ace2170a5c033b40b06eef94ef4f0c77f147216ac1ff2a06fdc9d4b2279f3af3e2fb06f95	9223372036854775807	\\x00c69148f39837b903a93f632acb670ec630989fd07696d7ddf8248db0e8d80c	29703665519	29000000000	f	8824	11463	9223372036854775807	751118	\N
65017	\\x831c6c2e774ecc91406d6ff1ee9472a9f2e6259b194db98da471d673a12286bece803c7d0faab25d1ed6047e324534f5	9223372036854775807	\\x005dba1339ec4e96d81e65b93333b83f12e1e21f69c3a639dbd551e0de857f31	29643402949	29000000000	f	8824	11464	9223372036854775807	751118	\N
65018	\\xb552b23550f11cb16d490a75cec3d86495dcca49c6e094dfc2bb4adcaf8b7a2c25105d43864302077d0de2ae0b95072e	9223372036854775807	\\x0047ca79ebf2dc5cfcee7aa16b4d7cdc28408ffb45b367656070f269142b400a	29732360090	29000000000	f	8824	11464	9223372036854775807	751131	\N
65019	\\x83fb85e155dea0caf5ec0ac3a69a25c15510b3eb64f39565d1e86f3b36f9bafed3dd5be0a326a279ed4cd4f03c5e187b	9223372036854775807	\\x0085d6f2ce4b546bf9d291f94adbe69075bb0cb39c7e3669b81e230f915139b0	29714819854	29000000000	f	8824	11464	9223372036854775807	751118	\N
65020	\\x90b84d09f4685dee3b67deec797eb48b7cbf26bf912dc7cafdf2d00e165b91e7451f26454b8b017e59cf0248dfe142fa	9223372036854775807	\\x005fa251e6c8874084994af15580d0677dce2b77e615df4b4f5092a97a084e25	29625333632	29000000000	f	8824	11464	9223372036854775807	751132	\N
65021	\\xa3981a7b5e899b71511e4506be3526111450301124f18a55495399e1036534f57f954c82ff640a9155442ef7fc72e625	9223372036854775807	\\x00ddeb4dd3af5eb5c5beac1472156a3953137ba028a9f4dce848033401a16127	29626691777	29000000000	f	8824	11465	9223372036854775807	751125	\N
65022	\\x89ed85c33885167a8fc4214e6e4e3680883a7931da8eff696406d0b419651bdd7bd6b0f830c729d632a418a6ed3fb949	9223372036854775807	\\x00f8c0f6d5cc2408de7cf494bb2c39fc51369397afb17167ba8433215dbf9001	29712559971	29000000000	f	8824	11465	9223372036854775807	751109	\N
65023	\\x84203ea7f3e0758155f6cb3949f19b2af97d23ac23a665c9b737615ebe9c01dc140e8cc4006a57be5f1e0c4ff305323c	9223372036854775807	\\x00a73bea6224453c187efb6fb0e0b67fa87db4dd4731aa1e5b2848cac7344ac5	29616790689	29000000000	f	8824	11465	9223372036854775807	751124	\N
65024	\\x9767427fedd24c79306deb07017333386620e8fae1f97ce971d2f9d15b190d6fffb1cca0e57fb2146f7c13e825525f4d	9223372036854775807	\\x0068f2e3c3bd203cfe8ad2479d084fa60c9f2ef234cd69859778e04993388642	29634077723	29000000000	f	8824	11465	9223372036854775807	751119	\N
65025	\\xa0e254695f47f7b1902cb8eccf8f4fd07550e54bad49c7dec22f491780f004f8fa29883745e5ddd9c596d4145c7e21a1	9223372036854775807	\\x000e0101eba9d8d79731fac3667bbc35c9bb683b7d06a82278e848640ddd261e	29683543572	29000000000	f	8824	11466	9223372036854775807	751113	\N
65026	\\x93a3695f4b9d6d6b8cc993241499027c24941b0f30c319337217f08442469d1e6695352ef392126f37ba4ecff9f3d1b1	9223372036854775807	\\x00d7575ec616df15ed4e5458c55fe2bc45cac346f9527c8666b44fdcb32ec834	29702349989	29000000000	f	8824	11466	9223372036854775807	751132	\N
0	\\x8fcf28896a85e5e76ee9e508438e23e7253da1a23a6501e3a7d56182520dbcf4cdb44af3267318188f1f4168342146da	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31222312792	31000000000	f	0	0	9223372036854775807	744312	\N
1	\\x873e73ee8b3e4fcf1d2fb0f1036ba996ac9910b5b348f6438b5f8ef50857d4da9075d0218a9d1b99a9eae235a39703e1	9223372036854775807	\\x00b8cdcf79ba7e74300a07e9d8f8121dd0d8dd11dcfd6d3f2807c45b426ac968	70155341822	32000000000	f	0	0	9223372036854775807	657038	\N
2	\\x8c2f535d3bec65f95cb4ba455566e4ec3de8da5c13a681699e0f80d7942d6fdcbcef18c8cf18f9da14aa379bdd6d29c5	20291	\\x006490500934b8b1876401dc09b7904d04c0897a9a28ecde4ddb1a60fd5c4c60	16720844489	16000000000	f	0	0	20035	636652	\N
3	\\xa8d9b5b62cc31149ad58a281a2293cd3f4dca11855c98983e76ffb60479d8e98e5592e5415f0400a8a23efdd842b3605	22937	\\x002f5bc32089b840a516c800c3c597bff24536c6c7c9f1113c2a0ac847608f91	15347768243	15000000000	f	0	0	22681	459696	carlbeek | Lighthouse 🦏
4	\\xadf943279435f1c194add1cdfe99e3fde5284d0451a63822b03ec301bb1cab4399d016f812461f0b71a8206b96ca3378	5479	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31403584679	31000000000	f	0	0	5223	\N	\N
5	\\x85f2c3045f02ac7b8b235b9aa855be7749d52de8b8e9855db5a9e97cf101185c05ebfa5d3d3dc2783e009294525ced20	5479	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31400723583	31000000000	f	0	0	5223	\N	\N
6	\\x9274b8d712b53c71dfb2d8762334d63ec9fb18fa6278aaf47b2782b14447539b6e5f56a9836e6ba51e4a206f1a8dbe02	5479	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31401991670	31000000000	f	0	0	5223	\N	\N
7	\\xb4fc18feaa072d37538e1fb8157a365bcffd086de9d335dca2eaab1508fb3a48504b8b982d05919f7989d818e88dd07b	5480	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31399247475	31000000000	f	0	0	5224	\N	\N
8	\\x9939f64a7b916476b076abd67cad897991917d4b7da38487597696844a890cb5156c2928ebf97963e77d5d2099ca881d	5480	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31408442944	31000000000	f	0	0	5224	\N	\N
9	\\x8e368a304c4c564617c3fa265361fe679df8160e6af06c5f5dba57d8ad134e7073a7ad631459116aef53a788e3c256b3	5480	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31400389830	31000000000	f	0	0	5224	\N	\N
10	\\x81903c25cba2b6f37f23449d731dc6978222a9df18ea08cba0a58c7d43d4b117e2ada338385fb1514b1692a5ca4881fd	5480	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31408703275	31000000000	f	0	0	5224	\N	\N
11	\\x8687275aab69c59d56efa991c5db0b56600912ea03b6b4ff3e297f0a62434791c479d415aa494bd0b41a1604b3de0c07	5481	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31399622684	31000000000	f	0	0	5225	\N	\N
12	\\x8725e7a24df271ffe645fe2a4fa46bae935666d4dc654dbf77e4853ce70229828ea1b60c49ef4644975cf2b466ed8674	5481	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31527128444	31000000000	f	0	0	5225	\N	\N
13	\\x9414c11a0e3b3712d0f62b7900dded64dc9c1c556c40e4dc40c0a7f5ceaec02135d45858aff36ef50b433e6fe24b6e39	5485	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31401199004	31000000000	f	0	0	5229	\N	\N
14	\\x90d42bdd0b5a7b62f665f75790739233b3a700f5f8ced95233fadcc72ca16977c8d6fb00b63d08ef750a88fc0ffadd6d	5483	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31408052750	31000000000	f	0	0	5227	\N	\N
15	\\x98573c25eb271690db29113fc890175213b1044bcd68e82f413bc694743ea265ca5eb8c50ebe2ae898426c0c9de650b3	5484	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31400909501	31000000000	f	0	0	5228	\N	\N
16	\\x88048df033f7eb4df024ce81b29459adfa92f8c447e732ff6b77f51c5e721b365dad73eb25c1617c11d456ed916cb935	5485	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31406629962	31000000000	f	0	0	5229	\N	\N
17	\\x98b0d39db01e3bbd6b750c43c26f7078581a326a9778760db3b72b22b74cf1b9a2025f4347f04bbd36494ec5bfd8547b	5483	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31407409808	31000000000	f	0	0	5227	\N	\N
18	\\x8a290fd6ea767704eecc68079e64cb9b43915652ebf9ba619b811149aefa069718dd97f8775004845ea9197f8c1cbee9	5482	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31412755695	31000000000	f	0	0	5226	\N	\N
19	\\x94ce62f60d60f9d52f2d9456b9f22edbe7846fff4b98f4316466910bae8065c86eb0756869d5cc27427f098493bc7677	5483	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31403037766	31000000000	f	0	0	5227	\N	\N
20	\\xab730f62ce3c14945e245b08b029dbf579d376cacaf82e5ae9f2b28fa72eec68e33103c693c502209a532edca594e0a3	5482	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31411752236	31000000000	f	0	0	5226	\N	\N
21	\\xa898422079b7db772520e7b43ae7c36bdba9f1594fb6b38f69a5d46c1c4f9fbd1ee738abe7badc786a68a9d6dfd37122	5482	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31409789347	31000000000	f	0	0	5226	\N	\N
22	\\xad547ad6f3fb353aa0c664e4c770560b1809cd934aa356d27c0ceb41622363d9e607a90c3f206dd954f18040a336c20f	5484	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31395370029	31000000000	f	0	0	5228	\N	\N
23	\\xb05fbe5c06ff586d649d8b21e9f75b37bf430fe04eb712850a8af2fc3455b5bd5051a258cde20aa2d59da5a0f0564ac1	5481	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31401382397	31000000000	f	0	0	5225	\N	\N
24	\\x8c18668cf8856f3bac3912161df97a6bb9b98141f899be6f3762935542d7857e160c506f31757b454fb3113315e7b9be	5481	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31400976831	31000000000	f	0	0	5225	\N	\N
25	\\xac605f3f4df79840e1e948df5d67e8484f4c3cddc0d6e561bd7bb6e48a8a1c9baf154c316b1b3d7f4ee23ca7a5cbf7f7	5484	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31401128704	31000000000	f	0	0	5228	\N	\N
26	\\xaf6612098003d793949804ced355362d9da566cf2fd00059de7321241754d8aadb0b77c94587ed7695dc3b1b51d3dbd7	5483	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31526596207	31000000000	f	0	0	5227	\N	\N
27	\\x8618f2f494a8abaa9915c87b35f3c3b2899972f6ec1e3696dd54aefebd3fbb48c3b4a02bdf40e17a437c3adc9f0804aa	5484	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31404965632	31000000000	f	0	0	5228	\N	\N
28	\\x83251cd9095f5fae3b2dfb006b16545d34bdaf1894687f2740300e9f8f6229eda1db3767c80a38fb3dbd8118ab415c77	5482	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31405982687	31000000000	f	0	0	5226	\N	\N
29	\\x87d8e3f68aa7cf2433219e3cdbe6232aab611358f6f4235794ae4225d24e22578a71b52847d459531ec081b4d148afe0	5485	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31416205374	31000000000	f	0	0	5229	\N	\N
30	\\x92705f79d60b3ae862ece24de3bfd234aa24ebf2be5f4235ffa72dcd4b882fefe1753d82ad355188bf22fa251e05e82d	5485	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31397407304	31000000000	f	0	0	5229	\N	\N
31	\\x99ef4f94bc863fcf2c393e03cbad1bfeaea19eece05a3084897bf6a39705eda7a40434bd76a01e6a170bcd319a445b9c	5486	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31405951880	31000000000	f	0	0	5230	\N	\N
32	\\xa6fd2c4289250d399876550d0b9fcda435f41a22c24e4c8b0e25c0cee8042a0c9d3f7c02928403b36e22a4a3d7ea8bd0	5486	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31404977031	31000000000	f	0	0	5230	\N	\N
33	\\x92c83a2b73b866ae670c423ea00194d856bdf552cf1c9f8aa6ea92929f73d52dae530e65f61ee0f558df022afa5f12f8	5486	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31407756519	31000000000	f	0	0	5230	\N	\N
34	\\x98250d9ee1c76975aaf3a1d68dab0f23a203eca5d6878c1f37c9c1ed2b97d62b1eccb9979097089dcb817ad0a84ccfae	5486	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	31404395876	31000000000	f	0	0	5230	\N	\N
35	\\xa27ede36730c6f4d5e089629b4ea7ee1d17b037c25892e41c49175f831206d14c8e633f397fd7d6d39114cb5c6e51821	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28492502378	28000000000	f	0	0	9223372036854775807	744322	\N
36	\\x92cfbbf8217e197d73492f838cc381b58c39eb68becc1e9cacc76d5392c94148ebb27d38a2efb1d9cca8673a9cfc1a10	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28616676210	28000000000	f	0	0	9223372036854775807	744334	\N
37	\\x80b81131547bc13b96f1a30ca884c297050acec603c201ac364fa7d7ce995828e3a2e26eb35e0cee569ea6399b4bd055	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28506925195	28000000000	f	0	0	9223372036854775807	744294	\N
38	\\xb30dfc42bdd14ffbf8d4315c666fdbf926409b4f316373a318bed26a956299a4372154e0d7a9113c1273c04a8777cf5d	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28628625331	28000000000	f	0	0	9223372036854775807	744304	\N
39	\\xa713c85d5be62a7ea34e67054e261dfb700794a830e2380c5e5f71dcf29275345926192e0bbc6877dc1ea03f8a9192e3	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28695755427	28000000000	f	0	0	9223372036854775807	744337	\N
40	\\x95d513534de5fc40c03e1e657f1632e17f2ae65e708aa6a25c13ebbebbda3dbcbdd762bddf1c5fa5c6fda289d9ab6406	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28740843676	28000000000	f	0	0	9223372036854775807	744320	\N
41	\\x8b37ebd467bd81b891774c695721e6c78d99111be938ff272c72cfc864683c3e1ff5139c4824199fe29d0520cbd8c874	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28614160222	28000000000	f	0	0	9223372036854775807	744320	\N
42	\\x8371247ae81e360ea915f7408a75fb732eee95885dfa7eb946f790293b4038ca0ad6be6cc26e31247f1cfd98a060c9f7	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28624565239	28000000000	f	0	0	9223372036854775807	744338	\N
43	\\x92bb2a1e7c4e984d7896b8e6468748fc654f045ffed539d983d6eab198caaafbfd7970c8d7e1cb768d7180091d30dc0d	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28597774465	28000000000	f	0	0	9223372036854775807	744336	\N
44	\\x8023022f1f335337748eb141477e789d683514ae739b9dac7b72bb3326201065abbf6f3301a5185bb8751776e5bf45cf	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28606039767	28000000000	f	0	0	9223372036854775807	744300	\N
45	\\xb1e8517c990087ec06de1c80910b5f5a4dca67f73451b61063d575e567da0cbb7ae65d2a4246ed1d176b21e896a4a4f8	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28563139005	28000000000	f	0	0	9223372036854775807	744315	\N
46	\\xb179dada48cc4f390cde0048047f5055b4cbb49437c4ed31d6f66e5db5079715cdb9d199ff13ac0521c907644939db19	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28679546876	28000000000	f	0	0	9223372036854775807	744293	\N
47	\\x9264398974c228e6449b253f97af9b48e463615165b7f2a3272cd60ecb111f89aff99c99fe1aa6f14a999003420ca88f	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28547864728	28000000000	f	0	0	9223372036854775807	744321	\N
48	\\xa80c96e1e93c8cc968dc639c25a0c7ae3bdde02f96b61ef2b7b3b0b1ce81c6bca2de60594beb0bde3b1189af504468de	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28659548076	28000000000	f	0	0	9223372036854775807	744311	\N
49	\\xacb7d7184b3c9a6a07e1364968b847262a758d08264746fc21881259b06f8bda1b5956a05274f07eb4a650bd123b7e60	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28584093206	28000000000	f	0	0	9223372036854775807	744292	\N
50	\\x8f1bd3e33f74091106eea31c36dc916879ac5f1354c58787bd0188208fb8383dbbe12e6c0c273d28899d50f96c5f4f2f	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28879084519	29000000000	f	0	0	9223372036854775807	744315	\N
51	\\xb585a904d9258b05ac280c1f92ec59f3b6e0e2fd456556ed6bcecb09fee79efef6c277e589156ee27f5d19c6af0d0cda	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28519546632	28000000000	f	0	0	9223372036854775807	744291	\N
52	\\x9070ba2c2247de494ec37e99ebff0a634dac8b689aaa63c1901af20e18244a79d5db1b67e9ec1b350bc14800f671c601	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28551987490	28000000000	f	0	0	9223372036854775807	744331	\N
53	\\xa1a87efcdd48f52a92b53a6b0488fcd274a389265026f0fe1707564f64ca0cd808beb12d246548b3412ba0e36ecbe551	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28518393815	28000000000	f	0	0	9223372036854775807	744305	\N
54	\\x80201089ba0e64c9c1f31b54b6d18179b9044ffaff7aae4fd1deeab691eb3d6ae3aad0a6087cdeda9eef8d54a8339116	9223372036854775807	\\x0010361af430aa7ab4a9567eaaca50ec5e02315ca1513d9ee8d73bde96370091	28588276704	28000000000	f	0	0	9223372036854775807	744326	\N
\.

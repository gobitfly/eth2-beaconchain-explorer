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
-- Name: idx_eth1_deposits; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_eth1_deposits ON public.eth1_deposits USING btree (publickey);


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
0	4302	26	0	0	0
0	2054	5	3	0	0
0	15924	17	1	0	0
0	4600	12	0	0	0
0	14729	17	0	0	0
0	15059	23	1	0	0
0	1916	15	2	1	16
0	9720	7	0	1	8
0	5864	4	0	0	0
0	16050	11	2	0	0
0	10480	16	1	1	17
0	12074	19	0	1	20
0	3998	10	2	0	0
0	3272	21	1	0	0
0	16107	27	1	0	0
0	8452	9	1	1	10
0	3679	22	0	0	0
0	7740	30	2	0	0
0	13807	4	1	1	5
0	903	4	2	1	5
0	14421	11	3	0	0
0	14892	14	2	0	0
0	6319	8	2	0	0
0	12123	30	3	1	31
0	5767	5	0	0	0
0	14356	10	3	0	0
0	2803	26	1	0	0
0	10415	10	3	1	11
0	9971	8	0	1	9
0	12248	31	1	1	32
0	3676	28	1	0	0
0	4809	10	0	0	0
0	754	12	2	1	14
0	16272	18	2	0	0
0	2702	7	1	0	0
0	2756	0	1	0	0
0	2652	12	0	0	0
0	15520	15	2	0	0
0	6552	29	0	0	0
0	12058	17	1	1	18
0	15406	24	2	0	0
0	2507	9	1	0	0
0	14563	28	0	0	0
0	935	8	0	1	9
0	15946	13	2	0	0
0	14809	15	2	0	0
0	6563	19	3	0	0
0	7311	18	2	0	0
0	15212	12	3	0	0
0	15265	11	2	0	0
0	2296	1	0	0	0
0	4593	15	1	0	0
0	1932	5	1	1	6
0	12076	16	0	1	17
0	8065	4	2	0	0
0	8295	15	1	1	16
0	8216	11	3	1	14
0	14827	10	0	0	0
0	1149	14	2	1	15
0	2693	14	2	0	0
0	15043	4	0	1	5
0	11072	30	1	1	31
0	6110	10	2	0	0
0	4099	16	3	0	0
0	14851	28	1	0	0
0	15302	27	3	0	0
0	1292	11	0	1	14
0	15631	7	3	1	8
0	11981	7	0	1	8
0	6042	27	0	0	0
0	1469	29	0	1	31
0	5030	8	3	0	0
0	1314	15	3	1	16
0	3734	3	3	0	0
0	2077	30	3	0	0
0	13861	30	3	1	31
0	9094	16	0	1	17
0	711	10	0	1	11
0	10090	17	2	1	18
0	1587	15	0	1	16
0	5165	12	0	0	0
0	15863	22	3	0	0
0	6031	13	0	0	0
0	2613	24	2	0	0
0	8026	19	2	0	0
0	3493	11	2	0	0
0	15725	12	1	0	0
0	6657	21	2	0	0
0	10434	6	0	1	7
0	10890	17	1	1	18
0	7677	2	2	0	0
0	12681	13	1	1	14
0	14929	6	0	1	7
0	4031	16	2	0	0
0	14441	30	0	0	0
0	3997	0	2	0	0
0	14883	30	1	0	0
0	15523	18	1	0	0
0	8186	11	0	0	0
0	11548	8	1	1	9
\.
COPY public.blocks FROM stdin;
0	1	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	\\x36074a4a661024dfe1026a733a07490108d2547706899d2d0feaaca12b430ce6	\\xa34416b701c34b62aae1dfce76e9950d4cbef15389867616458dab845943e96237075b4791161655f88ddaffad32ca6c03a95e9e2631ce7bd82ec469bcacfd9bf2e2d8904c29361368a5b1d2c8f5e81bf91b068b61078d8a7885a1df9c89b88b	\\x8836ad7c6b828cfe9826f6482e9c43d3f3acd699e5843f711f9f3754ed9885d1415eed813284a4535001b0e28da3d3e11608a9277c78eeefbe698b09ed667f9bb4ea1e917ba9ccad8428d2013ee0ce899688698ee540e1366725048b9525cee3	\\x0000000000000000000000000000000000000000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xec2c3b648d3da79afc1ea2fde04550081ac2d725deeb1ff710b734917ce2b390	0	0	4	0	0	9220	1
0	3	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	\\x3716b993a3b90fd0ef07e0ab583e3390ec20c7777624d31ede90a8717ee0a046	\\xb7bc328cd9803a7e033ed330b1c2fb3d06084cf708d117eaf2ecc359f9c5878b1c387f606242697cbc79b2be96491e8217cdee97e86884ff068a421c158d90c663a2871258558ff639226a436abd047d5c50ae0bca111216e67703f5ae069ab9	\\x8603a06ccf039371172b91f71f3a70c5cf68f35c1a0b1928c8ff1268fa5abde65c9c0c68463fe11dc0fc4590699397411568bd13bc29eafce218f9e42a1978d2a47e3a6a2cf9489fd5d76f95de2a37613454aac36472dca840f924b501a80a6f	\\x4e696d6275732f76302e362e302d66356261313337372d686f70650000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xb607308aca86e9816abf766193a53f4ddbfa77254962b840b7722cd934ce91e2	0	0	8	0	0	16246	1
0	4	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	\\x64a84c3ec9ab508b38fc0e3de2e7b8fef6d05fcfe39c60bf4556684504c92cd2	\\x80a1e5e70f1a5ecc266937aad93df6e818294b8544058d699af39b6dc3894199b07ee730201eb8cf39fbc6a6c61e606a106693ee00639bb5e376990e454a929b65b3ea190680c5a92ee9d8fa71dad845dc7afa38c554291ddcfe117d348759a7	\\x97c7c8c05c82d63f361df7c109494e524ff8af81c7dcb3e5dcbb5146636f057120d96c7690e96993628e598390b2baff0a8fd8aeacdb63554f1b5611326ce6a42a3d972d7ebadb6320cb1cd48ba6c4969e39d462ad6be97a1b8df9c64c5d286b	\\x74656b752f76302e31322e31342d6465762d3261373539633062000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	11263	1
0	5	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	\\xceb6e40b4209296a65b94fd982441237f3c005296efcb10fd921632c12e96004	\\x89d383df7776954cee8802bdebcb76a48c30e0bd867553716f49965e3dc4610594896d790aca498d5f0934d5933b8dd414b0aaf6896945de68f117efeb409f08f1c3d599c1246908fa119e89e6affd006c64499cbde0222b22387c70e98a6579	\\xa84748e78090293908a827aafbf71759fe1c587022a7e3852ab8a13335771e3ac512625febb484e2d55abbb80d0eeba10c953b8f6143230960ececb54a32d7fea514dfd764d21539ddcaa5e5491fcb55ee52da185cc2d04934af0a706fd4fc93	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	12298	1
0	6	\\x880e0571d7ebb5566d998f767b9355bc45f3393081eeb2a9a68c9cba488cae40	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	\\xd41088feb11591b09cb89d4c0a37130064ead36fbfbb7229f3e617d5c4fb304c	\\xae43bbff70813796c4ca64407775e9048772d261f4d51201c41930b98e456c820c2872e8ee132f3ccc69da4615e2edb70156a1e904dcc70c84f73fb5636a377a02d54c3951cb3e9689b495ea3cfb7e13665a2d4c8d024c1fd7873447cb810335	\\xa928e359b588758fb602cc9433ceca17ac594858bcf89a290652381f722b7591c2c3b6d5a86c166dc51eb30b8b0146f017d6508d014c25773b37b6ff6c59c22907f7f63b41f35a0e57a45f199b555d0c0b24d9c72b8c9fa140863079d9f117f3	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	12602	1
0	2	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	3112	2
0	7	\\x8e335a9cfc31352680b1a366c3563116d8c6ec4e5ff2aa4de9791b162ede631f	\\x880e0571d7ebb5566d998f767b9355bc45f3393081eeb2a9a68c9cba488cae40	\\x17f4ac7a2915a6c2f79879ae6813a693ecb99e91da06ead0feaec8c0fb597d4c	\\x93ec027a09edfc15b41f858426ced6af997c94baddc290671054e63aad6bfc3f4f2fa4ad3a36525f6719637602cd9b4213e189d443abdb13725963fcc3a1c4ad9781abdd14fd373606a2540b8ceef9c1043456f80dfa4c791cddd84d08f4ca7e	\\x884531cf044dd781b1a6e91307065812601d662b3e48013246ed608abd744d15bcf6f2bcf7122971175235ef8295ee2a0f6f136cb53c96980fd87763fc59e03c64b7b72b81bfbe7a425db2693a6325c63910d6ddb06e0a2b6d32875f088b04ad	\\x4e696d6275732f76302e362e302d66356261313337372d686f70650000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xb607308aca86e9816abf766193a53f4ddbfa77254962b840b7722cd934ce91e2	0	0	16	0	0	15443	1
0	8	\\xb0609f8cedcabfac555b91d07d60615f6e345b11f279cf2d6b755087589f52d9	\\x8e335a9cfc31352680b1a366c3563116d8c6ec4e5ff2aa4de9791b162ede631f	\\x7459f08bba0ef5507d6679d6c9ab7e27da498e7a0874229b6b82d4b53e2d4748	\\xb15609bfdba6e8f516da2afe50e9e2711998300d39c92dd4b653fff2826f2fc4d7a7b134ce541388672736f2f9b70f630a39fb02e1569806368a2e66d2f1bce714caf3828a6a9f54481eb12be88ac88f2c20a9d9da99ecdf3046eb95f5a565e3	\\x843b88ce723d86b6a5ba35cdab5ac75122b95746651911f3f27d5381bf5868816baa5ab1247b533a9bbdc71e7cef95a007460cb90bcbc760625e8f4d5b952a1258217f975c2ea274e7dde223a2f484a2208391fb6b305959ba7413ec9837ba67	\\xf09f8d8d2074656b7520f09f8d8d202333320000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	6487	1
0	9	\\x84cfe7cdcf68654ec286d31f21f5e894d36b5ea2a37008ea6cce97e9fe0f9d06	\\xb0609f8cedcabfac555b91d07d60615f6e345b11f279cf2d6b755087589f52d9	\\x2cf9ee3b86be93fb147f87ad1497af64eb4f98e8fc210df7d9eec20828d1859e	\\xa0b62e624a35a6cfb445698595a6b9aa73cc8714237f871c2b1717dab4e03c4bb3adf9666e0567c1f1955c64901beffc14d7eb6412d638763ffaf1b43c28846ce8ff91c1a0e0600df0c153b827a0732a541ef1fe24aeda8e73ca7e82738c30dd	\\xa626bdedd34a0c231c3b5893a299b7409ec532c335f7e64c5f6d701814e1a7230b2924a9cee149b1a49d3aec7b1882a3156cc34cb991e4a4de09e564d7bab15fd730b65a350a332ef6ad7b454a569639756e690710f8da7fbc312bc0618b92ab	\\xf09f8d8d206c69676874686f75736520f09f8d8d202334000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	769	1
0	10	\\x940d5804ff78f14152cc5a6140a54526f586d383d86ed3e412e894eb616fb851	\\x84cfe7cdcf68654ec286d31f21f5e894d36b5ea2a37008ea6cce97e9fe0f9d06	\\xb381c289abc2bf21da33b5acb03434b7db4856a8902768eb18cc2298ecbbcf87	\\x8465c5e8aae0837cb34ba1a2f2b1003069240cc3824168701a2a6950a350a46f99cd775b3df01d6a222cc0168197531c037455f6e19e29d6bbb19115180ebafabe920ec9027308d91fd72b9797b6f18682b132a2e50b548a0a48149da0423b8d	\\xb76338566f4a083c310b3ed5069022a46f08cb2f3bca0a173e2424be9d8c541057cbeedcaf2e6add2ec57d749e41f05817a7af3b56078d2635133e97ceca481db304629429a3ef486387e43709fce3e943f2ace6345b7e9314629b537111d2ee	\\xf09f8d8d2074656b7520f09f8d8d202333370000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	7531	1
0	11	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	\\x940d5804ff78f14152cc5a6140a54526f586d383d86ed3e412e894eb616fb851	\\x6608c51e831ed2745f93f77eb6d606760b6faf7846e02bfbada000c2ea02bede	\\xae2dc93b1a5203a7587a710eb978a164a3ed1a64cf36bab50ca5a416255a2d6d72218c3c5c7a10abe7cce341cea3495a0910d8de50fe458c52bd96832e064caf763113848aef5ca2b73faa5016bcb2b84abcdf4b9617ebd16e6397232a58dd52	\\xb944a3129d39761abc1f7e01efe2615cda48a7fb68698beff1d15342ec20d04bebcb88c28226a8fd672c24ec375ce9140bce2ea18767610b9061ae1fbd06070e2a81207e33443909aa50b65c7150e9da74d66a60dbfac904eb23085dbac704ff	\\x74656b752f76302e31322e31342d6465762d3261373539633062000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	11466	1
0	14	\\x4f483c496a117c227f63cdb5785ddde6ca5d73831e281445627e9b82a3f62b2b	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	\\xb04d49757c276adadf8c110de5bf02aac0856eb9727826355a146a95518d6c39	\\xb73e8e891cd9fdbaa414bbdf60fb6535a65f811d7c425ff325b6c8ac8c832a9ee004e256beb98ddff7261361fc6303f513a6215e8d70ef285e5a5e9dd3bc9b872adf6e3383d946e4d3c927fb0aa43e567d98370af1b3df176188377e8a68851e	\\x99c87e1b69c159742e4aec2fbed70b228a3dbdd9fd61801c1677002a72855a4f111cb2c7da7fe9a248a81c2281757e0602f2e9a94581609a7aa81c415847569c0a159bab9c84f4301af97390fb12831dbfd554ae8ffea88ac7edf7ffa7523e9b	\\x0000000000000000000000000000000000000000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xec2c3b648d3da79afc1ea2fde04550081ac2d725deeb1ff710b734917ce2b390	0	0	12	0	0	9082	1
0	15	\\x5c2619b66754d717f864168eae96ee46fe4870544bdfe20d936e0628097290b1	\\x4f483c496a117c227f63cdb5785ddde6ca5d73831e281445627e9b82a3f62b2b	\\x1a9e4fdc62f0f7270fd687c6688a24805b4f4221ac9a660c77d3fc6237cd0873	\\x8800c4acddb74bfc2421cac5e332a6edda50ced8de0075959f56e5319ee7bd650f82ab702e906b5725da73a7aa1f22fe0714709cf577db7199c63f36bb23117bcd7041a07702777b0d62608668ec6752d784df7ad85efa573edb3b40f59d0c36	\\xb3e08bed98121a811c243dcd40929d511bcc2abc29909606b1e1e9d9eca6eae8659d6130c4b81a083cd7842cd0ed73330d29b2c99a9a87ad8a2684b4b1fa747fb7db18a00efe1f8fc64f5e8dc5c340f290990aaac61fa5bf79f260d0c5b4a2bc	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	12790	1
0	12	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	15968	2
0	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	\\x0000000000000000000000000000000000000000000000000000000000000000	\\x881c6014f67e4ed9fb01eaac1e115a05851c2092e62f0360e8f9deec6f08baa3	\\x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000	\\x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000	\\x0000000000000000000000000000000000000000000000000000000000000000	\N	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	0	0	16384	0	2147483647	1
0	16	\\x6896c2af89443efb024c4583ac6ad4c500ca7b2bf32a9102a92586bc31ca9d67	\\x5c2619b66754d717f864168eae96ee46fe4870544bdfe20d936e0628097290b1	\\xe28a3d3f7343e3da174672caf2cddb8538a495af34b096082b20650799352977	\\xa64412f2b15f3f753459482100aec85f9594ea04f525a2018fbed927c3244467609899c464cdf95a290d724a783feff418b12c94a4c522fabd3d94f061d8f7bad2ab50a79bd9dd919bf46a9eba69532f436dace5d2b20198bed73ef7ca57ec45	\\x8ea5976e777ee7a969c5bfd639258cdc65c42b2441d733f475edd696e02efea74dcc2216478f588d4f494bc822b9bfa00039e638ddcb810862fe7409c8af7bb390501be7b2427d9ccb156ed7307e54c85ef11317be76699df5ba1d0f0e3d86ac	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	12383	1
0	13	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	4793	2
0	18	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	\\xb2ae877ef2b67095d8a01249da2183b3086366f51a03e69fcaa6ce07b6dd9544	\\x339bd084582064407158e2995ca045d84b41c1b104bdb25b573de3a278e3bfc2	\\xb12bd2c498b3a4f3c02a48a863c8a5c789e4ca8b5f948c7d20644cd2680011e078f74ece4ce2a38e586493132b232d8e06fa32a69c344cc491a5af4d98d53161f16f43bda1157f78a22056110d03483cebb15608ff58099a6f8629b5a70685f1	\\xa4dfe2c93c041e893ed6b747ba373cb13f07207e5c64cb8c858d76fcae0c5c1048b49a18aa6550290cc5e97d5916ca14169686d7c3430010fb7cf838e6a6d85c23228187630d1f9c07753eea1389b7ecc9901829484a9bca566fba8caf2929e9	\\x0000000000000000000000000000000000000000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xec2c3b648d3da79afc1ea2fde04550081ac2d725deeb1ff710b734917ce2b390	0	0	4	0	0	8321	1
0	20	\\x2713ca87ce6a7f8906269f72a60860cd48b6b4f2220f801c499cac3edf47144a	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	\\xec18ebb281780d067afb8507ac0a7b8d77cfa558370b5b62c191554a8b9a90ee	\\x91096fc002695b46bc0ccc674187179b07adc0f6995f6deed4b97216e0a920fdd7014c676f461d0f2d7b69abc56ba0a81856d3ae9e82adc57bde80f2bae3c98c6dc68b107d89bb39690f95d145726dd7c6776a30c0bdfce84c448c95d3016f55	\\x94195c43cb2a04bbb3d017ce2a442c79170088ac2db47a5b2ef3cef31fc45867214dcaca5002d64736a7892cb3bc84240a90b746805f7e7f421fe0952101bf516dcf2fbbe5bae1ce594e419290920b1bcd7fac70f1a009a1288ff684135ddb29	\\xf09f8d8d2074656b7520f09f8d8d202334300000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	8011	1
0	22	\\xbb0c3e0a5ca27e2f9ead7d95369daed60682ac96348d1480216d8de58d66c0ba	\\x2713ca87ce6a7f8906269f72a60860cd48b6b4f2220f801c499cac3edf47144a	\\x4c2389ebf4b14640096af750bfab5667a3a3a1b7c0edee8a9b2a28bcd976bc8b	\\xaf98a4be3a34a73df75dafa86259c22c6078b05b046adb39db4136cd914252e93ae2f43a1b66532d91c4b0d6df58aa480c3aa3380874ef8e6437bd5ec008293b7f305c67423d30afa6f5ab3d755685ec407be651ccbf9638a6926f65fd045183	\\xb6428b5c0314fd329d2db4db5d85969d29eb87b725159e840abe871af855409c174ed4d44fd368f8326b3ec3d3cfce01137ca390294bf0a5abe4ae52ca395ad41efa56e474ee41590a7e1a79e50bbe63a7f9022d59551bff413cb7211b3c95f9	\\xf09f8d8d2074656b7520f09f8d8d202333310000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	6329	1
0	19	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	2741	2
0	24	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	5822	2
0	29	\\xce92a1b32539beae18a5ebf951831f5f2b0c24913712506498a6b4f6b261f92d	\\x9da68f5ba4a7ba67cc2e68141278fc04fbebcf41d3de0c5e877f41563ff216b0	\\xedeaa6b835f05cabfdfb2d2bb86305b714d2745f71b7ece12b7321321c41df87	\\x9730f38f45c17403d6ea990895c09a2b6b67a222e92bbb0b5bf4a36c81cdfba74df42e84e013dc65720f3152c03124530ef29000b32f4e82bcf891ee937425a0465d8b2b5a4c35fb87297b8f28adaaa2a33c6886e3f217d1116447ec63d6f92a	\\xb05b69329b4ecc6e76361e0e80ee9a967c7407b53e7368c2c9219a6b4c88d462d657fff39689ef2e575bfa3109df401d0e2db9c8a6616ca69b50349b60288ee3899a17864f017199ef18a3aa499413f249a6a307fe58254afa45ce5282008b47	\\xf09f8d8d2074656b7520f09f8d8d202333340000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	12	0	0	6757	1
0	25	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	4406	2
0	27	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	5417	2
1	33	\\x6a9bb66108d7fc8418d1f69305006c745d040358809cca726221c3111142a5bf	\\xf0eb4c913727a9e84f838dfc52b094727e9f98f2e8b296e527f0ac81184cd716	\\x80b113950112543310ff657a043904f7e67e434c5f3b7796b36591efeb77a066	\\x95526f7962c58bfd2ad8c2057ec19f71fbfc15ef422748874c8cf2e75039c2073666b6dbffdaf193a60719037848ceed04455acf9c326dec5bcefbfcca26dd5a1e6bc3e6cc10030175e49fa49610496c350b0a6f408c2e7e7cec0e9bb790ec59	\\xa9bac490a0c13936c0795bf7742d53988ed4afbea3c60bf9fa11ab74702ceaf9c11c54a72237697d094f310011b6a1a100052e678e901f60c8e237763725b213a899de78d4fcc999a3cbf37a7d6ffa2b83f8fb6d6dfee55ff49a86e1ff535054	\\x0000000000000000000000000000000000000000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xec2c3b648d3da79afc1ea2fde04550081ac2d725deeb1ff710b734917ce2b390	0	0	17	0	0	9743	1
1	34	\\x6459f791339a74e0c927a2c7f19da275e22152df0e3086498d63ab650a99ae46	\\x6a9bb66108d7fc8418d1f69305006c745d040358809cca726221c3111142a5bf	\\x6dc2790f12a075e6d79c35d0bf32fecda964669abab6c49b833a14fb2fe79d9c	\\xb486d7f7634663edac544350868ba214191f804957d95ad7648b50fb8c6f89dfb0d37034803b27a5fdeeed55f491167d11c4f44ea0b328bad92859ecaacd00a2df5e3416ef730a83b25338a7c7cc683075bb22b9f85ec482ec4a7e2c1c17e500	\\x928952fd135558986f5ea537193e1bb8a2ff1587f3c01f4aeac1d9c0504b8f9e8a43a0bf9a2eb3c2bb60b9efc60c204414c9a5c493e8a94345a867cc0fd49935c0a10537e194d09b7e4944f86b553004fa3ac27d5df6cbd839020291f81be370	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	13141	1
0	30	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	14765	2
0	17	\\xb2ae877ef2b67095d8a01249da2183b3086366f51a03e69fcaa6ce07b6dd9544	\\x6896c2af89443efb024c4583ac6ad4c500ca7b2bf32a9102a92586bc31ca9d67	\\x7dcddfc6eb3524d32bf96c18d96622ef71f1ab66f3dc0acf63d4076c37ff369e	\\xb6edb63350dadf0dd0045ba3c3dee1f4e9b0f53842b2e073a81eb2b57f938ffb6c33464535882e2f73240d2d47f1a29614e44318ac04aa0718ae815113daeadf5953e0fd9fe7a13f20d68495d5a0c5728664aa3b5f635584d93a77849931ebbd	\\x82f4675ffdbe6e944e9962758fc982af2e9aed4277818e69c140e254bdee357f066c50ee1278ee654ee13f0150b3084c0061efa8f9637f768123ecda8a37df8510a60ec6777eb6be60a68652359481e0a0e715c370c1582b982bd3132101d520	\\x74656b752f76302e31322e31342d6465762d3261373539633062000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	10328	1
0	21	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	5726	2
0	26	\\x9da68f5ba4a7ba67cc2e68141278fc04fbebcf41d3de0c5e877f41563ff216b0	\\xbb0c3e0a5ca27e2f9ead7d95369daed60682ac96348d1480216d8de58d66c0ba	\\x53d5066aa81cb2cc910eaeb0764dcd3662a88b13e434784eb0346473b693f894	\\x8786c590a4e9592268ee4275a95c9972f504cd6f77164fdbc6ac0bfcfe3e12f44756fd328d14e469c376269cc26238f215fa840c4546f482e727d79f19fe68b584eb33c03f0d78b5aa2d3e6d3c2733c5513b7cd5fee9bb2b04a3ee3248a0d64f	\\xaa22c02d7fb06ac73d615220f053088e8bd12be67d0850b5bcaa5a14590c8c93b10defae842e2c4a3b5561ee9ba0ea1d06939ae75d95fc0c68a4653e958c5b784f7d9d4eba3a1273e246814f5998b50730b5f248df0cc49958cdd9e9f30215ad	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	16	0	0	13944	1
0	23	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	2375	2
0	31	\\x4d4f57a3e9f5e6633becfbbd35f1c0cdd5688f286b95afa3a258589cdadc8ed1	\\xce92a1b32539beae18a5ebf951831f5f2b0c24913712506498a6b4f6b261f92d	\\x2584a961fe3bdf6a90ebb6a94a987d0cf0a976a757d06be16dfd89a71c146008	\\xa42ca77e2b4436b7766174457bcb38e11d9957168645eb8f94819d0d3717d639f26f43aca4dd84d68500b1b456964bef17b2c4d9d6851dd8e708747fedc953e28fabad91228d5d98f14a48b0df532a18770ce61d8f575735b453c8ebd8759b7e	\\xb8d203eb554f49633c9907dd043bf6ef2983063567af67140bbb32e91da8364fb29038f31555be45272bd19919d0702417889f7f740b5ce2833d0a7e7aeda3e402d6f79f82beeb61e4c29270da197004b214d1955b9b08c1049e1085c773a539	\\xf09f8d8d2074656b7520f09f8d8d202333340000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	14	0	0	6905	1
1	32	\\xf0eb4c913727a9e84f838dfc52b094727e9f98f2e8b296e527f0ac81184cd716	\\x4d4f57a3e9f5e6633becfbbd35f1c0cdd5688f286b95afa3a258589cdadc8ed1	\\x5ccfcf85f149c734083cb1e40e0e2f8f05b5f6901fb42cc655317b004b575d66	\\x922020c921097e52c2f110f379d2fc3a9bf0ef2a853d982a102bc39eaedb897685db990629da7d39669b63e8ac18ca8b16119836cc5e4a2fb868c14f4f02d831da6cc4cd1c26267186e6b2d700a3c10934e3cf837c26c205d93adf20c7d96df0	\\xad00a381fac004e50eb8db218cf84e32ab48c07ef58f97807cd9e1376240f30463f5af9dd57669132878d6bb1c4d05bb17928b341ac3d4dfed5074d92f6cdccdac5ccd9e06fcb0b54c88a7f4abee5066cdc1f5fec38d56dbaecde441040e5fa4	\\xf09f8d8d206c69676874686f75736520f09f8d8d202331300000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	9	0	0	1848	1
0	28	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	3547	2
1	35	\\x5e7d79f713ab074b3ac27ab20acee0cdc3c1d4a3dd4d832375ce24ed31610415	\\x6459f791339a74e0c927a2c7f19da275e22152df0e3086498d63ab650a99ae46	\\xb8701367653eea746b87bb6ae04c003ec121ae7ca825d25da31cb41bfb0aac56	\\x977357aab8038fea235a257aed97160c19bbfd27c3eb393e4490e3322164d7966e89d222c889307e8cf1caa13672b022113da56f154dfda3e3f0c5502a6f55f174887124cf73631d4a97dc21336d2bcfd3239966b8e363708441150d3987ecd2	\\x90ebec83fc21c808c03fa896d3b69ca9cee0d8efb32ec099fe61ce931041773671167c3ad63a4e329c7b5cf037b4843808599463c0a91ce8950a46c5813ad70c834247ea9749de4e253dc64bc4bf1fb6169844e6bfa47fa52252c5fbd5e96c93	\\xf09f8d8d206c69676874686f75736520f09f8d8d202335000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	13	0	0	829	1
1	36	\\xd4bfaaae401fd58d7f472e3f07dd015a5a9ab52d6119a9532640042acb9a8f77	\\x5e7d79f713ab074b3ac27ab20acee0cdc3c1d4a3dd4d832375ce24ed31610415	\\xeb45a92f46b7f8000019880563f2c9ab1b79145dea58a2947ed469ad090d4449	\\xac617aabfdaa298cf61460b75c6e3df50714be071485dc964548ecd58c7477539be84fa7668bf0b5015d5470fd4a502601c587792602f5f69a7a951f70db23df8faeb655cb375a7cd39198445fd764e2ce6c8f7ebe3397847d0e1d84788b6aa8	\\x993b10babd61ab3fbdfeba6a0f9f336c521f8ef4f89378f2c80fe905717e0f72abb9866c677318b8141c3789a6d3adbc12c69a9fcf2ac68dc93d2c0c68ce2fe5b7c6972d66c2b4d9e48034096dd2718fc1638c3270398eb63977eba2db858212	\\xf09f8d8d206c69676874686f75736520f09f8d8d202331300000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	7	0	0	1948	1
1	38	\\x7ae97ba16a84c003ebc07cf9af75458d87f2d11563f42fbe5c5b370c6df119db	\\xd4bfaaae401fd58d7f472e3f07dd015a5a9ab52d6119a9532640042acb9a8f77	\\x20e091cf1df4652bf3f78fa4a36de6ef64881a22b389da4e32a148694caa64bf	\\xb4cc72afb72b57b167a1ef2cdeac4453f5937372b8344ee7d504341b1780900ef18e78c411b6130eb33a6cf0349b0fde022209becc30e34c1f45f30044bc56796e2235f6e5f3119fe2c5540df25e0207a7f794b14d75bbf54253f8892eb26110	\\xb90262ebce643d763d56ec1786c886cf7ad824c48e9be4d00c69b181a4499f68897b5c5ad85e4e9fcb2875c10ba7303f130cde52597d45acaa2fe1cc976c229231c95989a3cdaad6b881e3474ec0fb6ca87d3998eed07fdd3355bfc2ae3a291a	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	13039	1
1	39	\\x11e5e94c7f321f11dced1c90d13fdbf87cc09520650ee63ef40a62ea1179b240	\\x7ae97ba16a84c003ebc07cf9af75458d87f2d11563f42fbe5c5b370c6df119db	\\x882860dd1ed299b08a2a27b7c58a7189c4a3ad452be26495295d67dc1878b6ce	\\xb701ef618c1e2a197a26f69cd54e0795f6512241741de4a994f092368122447578ae26e9d2f898a9f48a9b0c5547040f07520a052d462dc7f7a055c465a4e4c56499fafd51f926720a89070c1a8b4567dff3ca159de374e1a7b936dd74d6058e	\\x82585a2994dfb5e1921022b6271bc751cd1f7cd26f184de388201d45501632b2a53feb5eba3b42432c6a6c614bf89ca8034b3d296ea43838642fab9a6ff356243e8695330e680ee10b0ea7626de2f8263998f72e947678c0331c3ac777248b5b	\\xf09f8d8d2074656b7520f09f8d8d202333360000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	7229	1
1	40	\\x47c8bdf97aeef6c23390d9eb5c5feaf1c0321b615109d9be15b2c91b4202b4af	\\x11e5e94c7f321f11dced1c90d13fdbf87cc09520650ee63ef40a62ea1179b240	\\xbc002c8f48cde1954297df1aa597528502fd60aa4e602fa758654a70b2c96cef	\\xa599218c937df7b42b056f4794bb548e4fae1ea86a97991144bec1b01154ed15e815c6c754ef11b3286fe48e9fc10ecc0ed115aa0b9fa20173f7d449fffb658a299dbeda530f00cfaa514e1caae7508c46c0425646a60e51ae3b1644ea371266	\\x92f9c53de406c86fe72c054feac3ed6c04fc29550455f657b949ff3db628bf78bcf500ab1082d7e47994c5ce79142e4a0cba3d7107f26b988f49df55df4a85b2049ecf5a34482eeb5443fe34e17ce5d170d231a8e7102ee3554cce7c0cc64e29	\\xf09f8d8d20707279736d20f09f8d8d2023323400000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	10	0	0	4909	1
1	37	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	15792	2
1	42	\\x0a2c618b39280247e8b7c7843c038cb12a65c3660211e4dcf2a7d2da32b84ac6	\\x47c8bdf97aeef6c23390d9eb5c5feaf1c0321b615109d9be15b2c91b4202b4af	\\x64f98e41f3bfa322f9f555533e667b5dfd7db24fe91d781c4bdce5d942dc0ce4	\\xa50aa7ef8c41e7ce2b7f53d6c58c757b92c8c174e255a2be79d1d558e132526121cb9f5fe4aa6f22ea7b5cfaf7e146630fa4b59eb2c10caf29562940f0a94b9d6d4985fd1c2a2450083de65e6ec93feeded922eba5eba912823a9a2b68f7f67c	\\x85e2f03c56c145cb382b1e6f67c1ce7dc6740d333b1ba66a9fdd5a536511a8c084fe1dfc5cad1948e2455c1b72f6b6d80cd29d0671fb147a32cbe6b21beb4d25c8959905c3c75a2f238e3fef39aa10916be8a8d15f490e81304c8c4f28b5abaa	\\xf09f8d8d206c69676874686f75736520f09f8d8d202335000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	1010	1
1	44	\\xabc97323bf3127b141796e3828eb6d9ac3750b84424f6ea02ffcb8913a43f20e	\\x0a2c618b39280247e8b7c7843c038cb12a65c3660211e4dcf2a7d2da32b84ac6	\\x927c9f5acd53b66ccdbd28379766413e79310f7276f7695ffc357873e2dfdc76	\\xb8ef128b9c914534d39bfe84f0d9dea6206c4cd469ae7b56afe6d53d1ca7bdce8f33022d1071c651c717aa2d2d7a035e16b9b844f4383c6f4620d9ab3ecfad2b5aedb680dab0488cc0959fd058f4a090825a4d9eebcd1a100752f437c2704050	\\x8d93ebb85f0913cbd86403fd450465d3f345498599ebfe55858b540c4775132c5be7417fcaa5f7b7c277db7e4929ebc8152c1fb6bbf4dc5c95c59df4f17816780b0ec93c29ffe535cc7d8f93fe4b40f81a77171952d31b440ee46be597594a47	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	12792	1
1	45	\\x7426a91d26eb7e338e34a86c812da28c18d6521282445dbb1ed171e70e79060d	\\xabc97323bf3127b141796e3828eb6d9ac3750b84424f6ea02ffcb8913a43f20e	\\xe5f5af596f968be065e2b3a1df2cb8cb38d5434a87325af921e29773b7ed9541	\\xb8e8fb19764ae543ed256cc4f81c65e405b6f8bde4d9db24f2f967468eb6d0688c60514b868196c35d18c2ee83cd849000d1c912acd1bde402459597f283a90e90adc0c0f4ab94eb438b3afe54e738de0dba7210b23b9f4cd59e1b78555ad54c	\\xa2dfbedc5dcbdc915380119e66efbd8f1db6348902b407bed616b62629219ce3b35c7f71ffea5be7105e56f87ff9d3ff009b434b12f3cab7b59d09da790ec3c4393a078bc2c07fba2cbb70d6e3ced9013257bc798942f90ade405fdb37c8c995	\\xf09f8d8d20707279736d20f09f8d8d2023323200000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xec2c3b648d3da79afc1ea2fde04550081ac2d725deeb1ff710b734917ce2b390	0	0	5	0	0	4375	1
1	41	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	2911	2
1	43	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	14440	2
1	48	\\xf266a32a22eecf8b09474c4cc1545c5140a5cf944a1290510ed79c6a1af54c48	\\x7426a91d26eb7e338e34a86c812da28c18d6521282445dbb1ed171e70e79060d	\\xb92b2f610e69dbb290493961a2bf65cf894a532ed168998cc44f748e753db2e1	\\xaf78997aaac6a8c5848ba38408797216752c52c6c8afdd1dad8c7b22b624bc4af14a75168613d10d71b65e60ab7575ba15942502d01bca203613a8e0a049ca152bb0a17fad9eb37aa2c3edcf4e1031692dcba34b40e9452cc81a63d5a2097dd3	\\x8a10bd80e1e670ce74a2453366897b9117b159460dc918c71eef3104a8140b78d95d3c2e3b3a67aff3cb720d79d598be02abfab59cad354d472c3ef343c6de75fbd9addc32b88daa63fae79e32e02625c4a7186d5ace0a9f171d74a629277ff3	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	12	0	0	13645	1
1	50	\\xa71d47c8a547b921a1f694fbdc1b7f448086b8bc27dbd0f3c0c6e5e4b84ee3e0	\\xf266a32a22eecf8b09474c4cc1545c5140a5cf944a1290510ed79c6a1af54c48	\\x4837afaa41d39aba65975e8fd097d2733ac20f0a4b92602d9c3185b42cdf0a98	\\x8eb216acf2a50c6a8bd5d3edaa9b9cac9e36e06c233210ce1865a69bd616247bdc1b98d9ef0ff05ce83e2ceb71771d3d185394c56c1e9162e997d4fc7b240fc023ade29884bec0f357073d18e6f2cb022410acf486f9ea601ae8cbc54f79e6ab	\\x932959daf03f72dbcba0b4b8d94f773439de0508dfa9cb5fd29422f7921be989e1ef567862b8d26f95d1e651851b4b0f12b1685871c4e0c53a163d123fe543806c0c0b3405bfe18df0bf1c293a71fdefc37c1f249ca5f94485500f9a539e9db6	\\x74656b752f76302e31322e31342d6465762d3261373539633062000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	11616	1
1	46	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	2402	2
1	51	\\xf84221aea4b17cf9afc497aeebb84de54f77b50c163d690a8a5b50af14861b47	\\xa71d47c8a547b921a1f694fbdc1b7f448086b8bc27dbd0f3c0c6e5e4b84ee3e0	\\x40602d19ff318542a0b59de58a6b322d8cec80f01d2172328ff51553b31a7351	\\x82036743a2fe02adcd56116a24ed16b81dba85f3dda4b31f5abfca4431c33dc822062b55e4e8c5f13115f432009b5ca70deff74f4010174e064c8a024af30fb932ade8a3d25d63990a80d3c259aa706d6380841a96570d5f4b8ce12748af2dfc	\\x8236925b0c85716b8e72b4e3decbbe3b9a0760157be952b22c365a27b95822b9a101f055bc8fdb9cee5079137744b1c508aa71bdfd400082ea2b7b22d928e4a2891ad6b38d762760bb07802232947fe4a8ac797277dbddeccd167fcee898bc19	\\xf09f8d8d206c69676874686f75736520f09f8d8d202331000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	18	1
1	47	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	2506	2
1	53	\\x4bc923f8752c4e0e424d8a06ec2c929e0155b7527666e1c3940616108a7beaac	\\xf84221aea4b17cf9afc497aeebb84de54f77b50c163d690a8a5b50af14861b47	\\x06e71fb44b1c971294d1954309df4ededd997571dfb61be75e238451e5ce5f42	\\xaf4c6b0b3fd1350dce472aae0bcb0700d83a261cef90f79d4d1a3a70a2b7cf422482d3825bd6fb5f42cc0304b8d7e233139756421fc21d6b26df4f7012b37bcc6afa1f1b3516a00aec2e6025ad21837b4138c6035f5e6b4cbe4e005710776e6a	\\x85d2151fc5c78f029faaab950eb3b55dcd5b2fa566ad72315b65dbc2eef3d225cceedbb4d634612831eaaab52d9658db0711797416a67cad59a09b24e456672e964823a77daf8f1cce6673d8ac69e70c28bec0b05a6d6a79da35848c013f01e2	\\x74656b752f76302e31322e31342d6465762d3261373539633062000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	10242	1
1	49	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	3453	2
1	54	\\xec9a6aef0b5ea37835e82b487a316b0e806ad76c63cb2d2d3742117ebdda9cd0	\\x4bc923f8752c4e0e424d8a06ec2c929e0155b7527666e1c3940616108a7beaac	\\x8b2e13c3e11d4176caf19e7711651dba5634579a1eac455cd5f7ef0e2a08a9b5	\\x9057c5f0fb2fe37ac5b9003736a75d60529b954451958f368a8fb5313c78f725b632a75f046a45b960ff86a6a3d3d6890636fb6468d008c6fb7e36b1225442596493ac4f842cbbe9cc78459a691a80c1402438d39a80b11ebee208eae35a9d03	\\xb94668e9515a21154fe67dd7122ad259f012d4ef08d32f1e96770b8a38e5bad335a89d2d183c75c2348faa48430e9b3f12f42add968c6e7cea04ea589f23779b0f518cc2a05cfeca44303cdd2b995d26b14863bf56b8aaa47932f4076263f9c8	\\x0000000000000000000000000000000000000000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xec2c3b648d3da79afc1ea2fde04550081ac2d725deeb1ff710b734917ce2b390	0	0	6	0	0	9590	1
1	55	\\x1aca519556f10e456e61b06ae70093ea9bf253c5dce2d890e8aba6df014484c3	\\xec9a6aef0b5ea37835e82b487a316b0e806ad76c63cb2d2d3742117ebdda9cd0	\\xc2479972539d1d9f680d4fb23292bcd0336cda384a4dce49f9995f139bc5cd41	\\x823b86702acec52fc37eb65121ad0bd8a0df71043f79bb7d9111d3312dc3994aec1899f8f3ebe72b4489d114a35aa15811c6fdbfa0756a56a89890c9ee67cd608d28868955296cbbaa3ea0786190f7c284315b260b1ff4beac0d44c127b24622	\\x84179a5fc75611ec54aa9a16661b427aedd9cfee6f54f1cfc2f3d1605f97207b4820afdd81bbd258c648ec5495c9565318d18d46aae4ca84b4b711cb92a2f9e9449f1df78f52a97940873a576e9496833e14c404a366f5de3e06251ca2efbbb8	\\xf09f8d8d206c69676874686f75736520f09f8d8d202331000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	120	1
1	56	\\xac2e9a5b71389c484d54cbe89027cc31b57ccdc5672d15b2fbd5887b37a92a02	\\x1aca519556f10e456e61b06ae70093ea9bf253c5dce2d890e8aba6df014484c3	\\x0bc0bff3827d9e3f4997ec0dea5b72299e5c16cf763abf59a7c17dcc81aa22b3	\\x82cea0dad5b20a60bd418fee115fe9422f53247ea8dcf6137448c1eaf4a2b4f860060fa278535ba154b24595185c04e6084f835bd2d8ad6a3f23cd853db2b5571eaf6aaf0172752f2eb508d82562860287d00481d8df3f824f2b074e9a31d725	\\xa3a639793fb26bcd25a260c822d6cf6aabec373b2e0abf05a057d0fc5c562323d8e133d198f1a887edeaa8d1120653b706594f5a83d94ca8c92e1367a6517c8a73b132d0bdb712efcb41fb58120a011db1c5026cdf49a1deba2b9ffdd1a5df5e	\\xf09f8d8d206c69676874686f75736520f09f8d8d202335000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	1016	1
1	52	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	2221	2
1	57	\\x88c38edcd25e8dc8c760196399dd165fc0c5b49b3a50ddb0ecf9fb58439bca40	\\xac2e9a5b71389c484d54cbe89027cc31b57ccdc5672d15b2fbd5887b37a92a02	\\x0dd62ef96471cd23cd62ca92960086d541de4987f7555015736eab258e62b483	\\x84781e9227385514d0e4c9c1dd374df327f473cf027d344e1768f4716f17a0456ed1518215133964b0178d0da9d6705d0856dc3babd82b82a5a4f130375c804d36b28e6685c6ed5e42c6872c4e177eb1f303265b37ab70b1afdedb06aaff959c	\\x8b91ed2ff229f1b7eecd604f1a569414c6a9fc2d77a26937399dd5e80555d17532aea6e1567056d88450c883edd208d109100745d4693673e6ef4b41700791e36cab1a171fd29947cdcd54aa4ee792eca94f4dfddcb3578e38440fd9cf87a794	\\xf09f8d8d2074656b7520f09f8d8d202333380000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	7731	1
1	59	\\xc691215e3811d414c75c9527eb5c92870750e650088c0163d47c0b6c96af88dd	\\x88c38edcd25e8dc8c760196399dd165fc0c5b49b3a50ddb0ecf9fb58439bca40	\\xd5f70a52bea5877e73f4aa7c7f3a2c9b0a10430abb3d9da3b54baadee479c9f2	\\x810f307460865608a49039f7da3bbd1c522058e8420635f4f335712d982b2beb96d47c4a00484cb33b5f4e09294cbb0b17879dd298ff6eb87c9d3da0059bb35395a14898d2c2a3d96ad1e3784b697be2f80fd4f7705115d91ae8ce42fd385ba5	\\x942cfba1ed14411e8c3e3a9e29017c8ea1a3055f672e334e5d330aef1a2a04361c6dc358c9fa0c34ff2ad275cb83306b0d5b3f7e68c93282aa622577e42df90ac6079b0fedcaa6c16bae6b8a58928eee45800277e00c5c04a38ebd1b795054f2	\\xf09f8d8d20707279736d20f09f8d8d2023333000000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	8	0	0	5979	1
1	60	\\x0b9cbd570e625e0b2e47d87438a5498b9bff1f7c54277eec7ac9e1925fcb3c3d	\\xc691215e3811d414c75c9527eb5c92870750e650088c0163d47c0b6c96af88dd	\\x9f8c2dd0a06048c174618a97016f9d2d44d49bf5f020cb105dee521f120bb47c	\\x851a6222230a83c9c5284139e2bc581e982c906a92414d877c1dfca61eb97ec369e49e1298026810faf97ba93d41f06a153cdca2553590a59898451910fc5cc6adde015e2795a36cfaeb4ce8dc6c789a0c64b0f4c859c1000cca89ad659a8cb8	\\x815c4cc984dce4e41f4142354df2403de6ec298c60907b5554f5fbcf925b4fd17b18e63d482176d0dcef722d8cc6ccd804d4d3dbb29ab2ef8b69db087ad401176f54f5b8988ef6c9c1eb7148ff0b762953e27a959f400d7b0871719f2daf6191	\\x74656b752f76302e31322e31342d6465762d3261373539633062000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	12009	1
1	61	\\x29632c65b30fd08d8914d158d8c438c4757ebe171e4c8cf1c9553389e3da96cf	\\x0b9cbd570e625e0b2e47d87438a5498b9bff1f7c54277eec7ac9e1925fcb3c3d	\\xd70e9b7f08b818b636cf94063424319e76c622815cc392252f7078f25faebfaf	\\x8687f864c054695144f34a2fdd40c2139fa3d435b1f35c3e381ac24d481d8d725a5b009858e84bc9d8a84aabbe9ecb4c10392e79b5881a51213fcff5e140583bcb04858bc2c8856cee4fd5e8083d4b4abdb300b6179915d545261fbb276ee366	\\x99820aa030476b3234c7b15c4aae38a1085640cf5cc68a9037f96e5c5584259c72d70982274dbbe32a1ef3f0889841631863a337eec89f2bd80d48ad2ee4d1a6f99580b8329150bf1431284431a2ec0ad4039f2449684dca7d9679915be3e8e1	\\xf09f8d8d20707279736d20f09f8d8d2023323900000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	5	0	0	5854	1
1	62	\\x684686605eed2aba6e62cbf3b6fe3013bb6c16a585fc8875f0e545b41a4f9f63	\\x29632c65b30fd08d8914d158d8c438c4757ebe171e4c8cf1c9553389e3da96cf	\\x730ce5f1be49e2a0dd6f012a749a6df4655a8667faba233cc6ca46783001e997	\\xb2bb13a618771be6f300c9e9e10fc4a67c7325d5e2bc7981980311581638b21bb20c23773c5ce10dd4d011072e619d9102ff52701eccfaf40d43c494687ba546288952431d4d4e726bff2d9d568d3d9d3cad3082f2475201bf92a10314f3d443	\\x98283359fa296c3def6ab779db9f0c210885f644f6cdbccc459d6ef1f40b720ae2ff2a28e053bda0f4dbf90c83db5e950258fddb0ad51bdd8794e34d8d67688b6c7f79ce12ef96b41278cf6e23620e938b645915f02dff4dd3f473d3185db95c	\\xf09f8d8d20707279736d20f09f8d8d2023323200000000000000000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xec2c3b648d3da79afc1ea2fde04550081ac2d725deeb1ff710b734917ce2b390	0	0	5	0	0	4372	1
1	58	\\x01	\\x	\\x	\\x	\\x	\\x	\N	\N	0	\N	0	0	0	0	0	3049	2
1	63	\\x1a5cea0a762893640b72ef782cc27c807da416192202f28f16571f936edfd8fc	\\x684686605eed2aba6e62cbf3b6fe3013bb6c16a585fc8875f0e545b41a4f9f63	\\xb7ef7cc034028763801884f7d80f65d973252d251085ebd7f27eaac260620f93	\\x84347bd54b4ba19d6e2712cd5eefcabf49fed78db503f5d2a4d6710ae2d44851babca35e94fbf85dbf1a0bb919e595490cae42f9e96f06646c11b9e80f7757dfb3e72f7db98e246f6b0e29b8dfd1a9413e259da28a30915d35ce4c9c74f16731	\\x825bd684162968bae6c332fa412994bacd40ca90a649ee6ea136fa0338febee10b428cac516acf6fa0086490c47644b50f23506fea330d72fc59215fda605e61a5106eb86abb9420d6410740e8568cd56914551d97696ad9c4348075c3f4e4bc	\\x4c69676874686f7573652f76302e332e332d6364303066303500000000000000	\N	\\x906801a472ec9da027e3924bcb359f9a38329c17d31b76b6ef97b4daf16a8c0a	16384	\\xe58ceef9d4edd68055a03cc98d629409045504d03afbe3fd1c37487914e5a97e	0	0	4	0	0	12759	1
\.
COPY public.blocks_attestations FROM stdin;
1	0	\\x983000a48810201004000108440002	{8769,8904,9142,8564,10198,9427,8781,9069,9136,9684,8776,9687,10134,9162,8593,9191,8656,9605,8913}	\\x853a6cd13e07c84257f8d472d399d0fc101772433fa53375598d7c662f58c7ee1c997e4938a012d691ba3521e62728c902850c01ddbe529e530d5a191c18025c6ccb08330ce1600dcd9a60cd6baa2922e0984e8307e205415bdff66497b46bd2	0	3	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
1	1	\\x00000024002000125404408000401102	{9806,8821,8397,9454,9646,9194,8954,8940,10136,9360,9950,9091,8719,9690,9913}	\\xb7276ca32a14002cca199369333685611f975941a4f09bcd090d50c229fddd58716c7f7e8b07517f451d085a35b9756a16f897f211a1fd07f6c1249096c5274ea8f915f2ce9a30e3cdc5a6f37484449fff3c94b0a7f1cad5bbd2592b81f6af66	0	1	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
1	2	\\x8300040a0020000000d0000044	{9636,10203,9149,8388,9870,10210,8953,9087,8379,9740,9265,8272}	\\x84032fad8dc0077cd8988be06f4a89cae0f37851605cd8a238552ff547b13920106bd2b67607a59e3655cd35002169990c51d332ea3be26f10e2c9b8fe1e2f753eb9bc6b6edaf1c8e0f4f3d85e99e68f5e3a7c5d7bfa39140d4209c8c800d6a6	0	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
1	3	\\x00002210000004000080000014380002	{8563,8784,8975,9159,8228,10103,9696,10189,9270,8548,9600}	\\xacba6146f7a53b4ba5806345444606ae1ac2ad936c08a543e177599af08a708ac23cd0e63ea25928de50967960b5e23c16069384f03b496453e1be24cf1139709bed6d9f984dcacc302fed3b57863d86ba25fd1ef48697aeabeb78ed882a2a61	0	2	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	0	\\x5455bd73786f7326b4b34d02bf1db019	{15402,863,14605,14491,2029,1616,9099,720,15928,1186,13185,15738,10537,10162,13057,15971,108,14677,11130,8623,12434,15029,10413,11317,13529,8456,14934,8773,12261,1928,11334,8929,14428,13799,8985,13672,9805,13237,12659,10545,13228,1874,13903,15327,1174,16227,12273,83,10905,8205,10854,962,8316,1847,12590,13726,8285,314,15305,10845,12636,491,8760,11421,1094,715,10458}	\\x940891f7a202a3c4ff07a778103208116b0fc7d8da63c0dfe64fb5a6ad81798b4d588f0d701f8040ca2581c99e61f6c901c409c757647430f36b5548cd7e1d36e6f03bd7357f4648eb48abd4b379e5ce36e33da33c12b204b48bed5299faec25	2	2	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	1	\\xa66735eaf9bbba7ad8ecd78d3d7a7dca	{103,9743,882,14453,10108,9538,13102,329,15518,156,596,1879,157,8947,14965,8757,10621,8626,8324,13276,8433,12940,2032,10557,15455,12962,8511,1541,950,12138,14202,15901,14760,1759,15475,16256,10404,15994,8800,14253,15615,438,27,10271,14227,9898,1856,130,15529,9803,10000,404,10373,1786,16197,1046,1527,11433,15328,10301,12944,14636,12247,15857,15761,8834,9243,8338,9426,1761,196,12938,12705,11474,14779,13348,12694,15513,15922}	\\xa76c3dc97755429ed7f3f05017f9f92ce1fad19388b101dcd4340e891205e384f8a81ab762c43eb94d23f856e24b4849064b6e922a9d0a442d06ad96a845f0507c6602b93150b1272d115ac3e196fb8113bfa27200252435f15f84ed9b585b2c	2	1	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	2	\\x8eded76ed7bec7fa64fadc77f70b6bef	{13389,15008,8208,10316,15034,12126,15415,13545,8542,210,14099,14279,14132,12358,10799,15714,9144,16054,8299,8458,1900,591,12233,15355,12282,9257,9988,1878,9154,9464,10180,1414,13975,13737,1144,8766,14749,8670,11661,11769,10549,11304,1400,12671,11361,12070,14262,9065,9216,12922,14828,11500,10697,9115,11086,1172,12790,9647,11613,10436,12390,8494,1445,10269,9928,11805,14527,13676,13186,16343,9112,12712,13832,9467,9139,1101,11316,1833,1687,1344,15845,15012,10775,9275,9999,12908}	\\xb8700fe62f02a356bf4b6f8ba423acad8d9bce1f3872cccf41663881ae677cba50a2054a14eedc1bfcae497d9dc2e1c01513ad1639dd816847f411032bca649335058fef942fa4b74e2f4d95a100de3636d931352e72ca6da1db2c5711989ef6	2	0	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	3	\\xfc6de16f3da550b5cfd4297f4f0cc79e	{10768,13586,11731,1918,898,1551,765,10699,12758,8856,15580,349,131,14305,16075,12104,14165,657,15149,14472,11280,13072,11787,15807,8411,13583,1767,8450,12094,11028,1088,8917,9459,11749,15453,938,15568,221,447,15967,227,809,12858,15843,16061,15485,14900,16285,10354,11770,10720,372,8339,15403,1178,15657,8229,14112,8861,10563,9207,11793,16118,9026,9418,12911,283,1822,14987,8647,15056,1521,15989,9536}	\\xb3b9279f28797b77454ac599b95a755e2e572a4a848a18d5578d702d6e52fd3d1f9d0cb37cbc6f4db387b8142a59cbd20f324d71b0991c20833afd617aa68c08cc504204a3721a0f0d2e42f7f950e3df1f9a4ecbe7e4f46951bc4407eba4f0c3	2	3	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	4	\\x6204b85d9def0824a83761331061bbaa	{13898,9868,13217,86,12966,8257,12316,8701,12851,9096,13035,9461,13005,13637,10485,13752,14310,9251,1832,9577,9882,842,14018,12435,9349,14326,12021,14280,12519,10453,11499,14307,1742,9880,12771,8827,8942,11043,8988,11078,9836,11799,12677,10510,13756,10181,60,13429,10703,1960,457,13616,10691,8547,12957,9925,11802}	\\x935cb0238811546fa88b6c07df3500350ea6638434eab7eb99339ea29f3c239600aeadb722c0aa9f1699f36c92ac277b054b7bc79cc3a49ae3a2c5e30325232795c09a6cf08d059026eee7bc09a3ecf38a068b9582eb8e633964a4a296b0fb02	1	0	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	5	\\xfbcbcbe50c929a503ac8a7ffadd3d421	{13949,12016,12122,13329,9422,12631,10448,10923,12381,10822,1368,10786,13014,11378,12160,1119,1277,13297,11251,233,31,12863,11464,11176,13333,299,11253,364,9178,9335,12623,1647,13395,8240,464,14239,11234,13275,9648,1656,13696,10756,318,11902,10486,11839,933,13540,8608,9935,12836,8571,1108,10495,88,10730,9248,13740,666,14302,12945,9503,9272,14315,8,9788,8270,8273,822}	\\x8b0401edb540724910699e4186c94cadfd1459619c43792050e17cb1a4d072fc02a60668d62436c2e6c3c090e77bbe4e0d6a3d545ed632583f52e1b31b1467c76209fe2dff94ed669f24059e04830a0e4b9faea4f4849dd1905a806c5cc7233a	1	3	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	6	\\x7137f571c504fac367c1d7ebc565467e	{9992,760,9771,12744,10531,9185,1184,14246,13260,11867,12388,1064,14301,10637,9830,13603,8392,11134,902,9762,1083,10627,9359,13662,11656,13011,11826,9508,102,14176,10583,11098,1039,8624,13227,475,13723,10157,13592,12472,12719,9517,480,9539,12500,1677,10659,13131,11768,295,11561,1682,2037,104,986,14311,1525,9847,645,10326,1124,12407,12624,9576,8607,11166,1390,9085,8420,11919,9633}	\\x974668831800d4d89b778571e64c75ee31438dff8d648b1623d4681991de7917df0cb4108844450a178c2bec16f54d8219e4fb22241b2e900466edccde02f04a848faf6055e0a3005edb8fbf2ca1ace2d2d08882734dc396095da960d4ad9bfc	1	1	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
3	7	\\xe61c501bfc0a8782b58c1443b18bb1d9	{12256,8573,12543,12561,1009,9413,1026,13477,8893,12954,109,9738,8467,1528,11777,10190,12484,13995,9160,13419,10001,11397,12878,8916,11604,10267,10565,13428,9333,8312,9229,12724,440,9843,1589,9273,14081,12961,11458,11516,8589,12228,13837,1966,11688,9557,13668,8221,13135,8880,68,1265,940,10110,11526,1226,1961,8460}	\\x934de885a56768f4b92fa90fabf63da35a7831b73dd0fa84ca907b57cf0b999424102e376d93902eeaea1787d95255c913e684388688a0a60289f857ac40e4bd2dc5c7e390eb4f3ab1cee03a14dc2a4603be1484085dbd7cbcdb194164d52338	1	2	\\x1791fe4ab3a910858d765c247c672e1e08968d95b6aec5fc3c102c6f45167c2c	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	0	\\x9f5ed97bbdfb22f942f1f947c17dc4d3	{9341,9376,13430,16207,9970,8717,11856,11386,10574,13056,1255,14198,13872,11249,8826,12745,260,13417,11323,10789,466,8703,8220,12148,628,14553,10539,9521,1253,8510,212,11187,10518,10397,11625,11972,1977,15141,12798,448,15267,11577,14567,439,14796,241,8271,1038,10579,9105,9869,12572,1062,10608,13679,15587,9944,1386,9522,9943,1807,8795,9699,1819,11820,12849,14598,271,8892,687,15388,14513,10041,13101,93,14632,15991}	\\xb2f04d5ce585de3e13d571edc13c5eaca8df1e9208d385fdc548893129840dd4f6ca58d3523768827ab3e64ccd151bb407e69613b80ae003202546cd43f84fc056dc8d1cfded6223450ccf8538133ddf1a37325835dceae1924783d7a420f3e5	3	3	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	1	\\xad9b7e9c25dadde3c4efef5feefa8cd7	{14662,13831,14454,8721,11705,10601,15283,12081,11,11804,369,1760,13584,12648,9934,11960,432,9765,275,12742,9923,9472,11305,13466,152,14290,8446,11403,15476,9127,14823,11606,237,8196,11665,8740,11762,1747,9320,1992,11699,12933,1853,12657,8421,14463,14233,13537,9787,13410,12303,15391,8311,10615,8992,14086,16238,8712,15940,894,15263,12118,16164,13023,29,13524,11910,14189,12593,9278,14415,10587,12373,19,10011,14220,12253,14361,11616,1730,10932,14249,699}	\\xb48991cd81d51c017d88d1fc884a909ba08e9d7c70b84b16ffecd0eeb0536f190fa7865fd518acd9d780582dfaac5fc716631d050378398d68102cf6195320f47c03f7d71e6ecd3bfec05059555c701434e302e83bb7b49f870dedef5516382a	3	1	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	2	\\xf7587fd90f6d7c18ed6ff5662dc272e7	{891,1817,15677,587,78,14696,15628,1458,788,15515,15273,261,944,10788,10250,1942,13073,14495,15774,16147,12872,15338,394,9697,10624,10437,15969,12793,13390,14185,14720,14896,14509,16297,11579,11510,1649,9893,15524,1221,11126,242,359,10472,9810,14924,791,13301,13105,10862,10913,11143,808,9833,12375,9010,10914,12976,1619,16250,15871,8305,9017,13533,10377,8788,15813,1781,14416,13206,1162,16206,12499,14351,541,8569,15833}	\\x8f773105786875b9b037da39a1675ba35f1807453f06df522c1a68f28cb48ebcfdb8a691ba69da67ca965b76b58a07c80fbb49a7ec29e835dca654201aca35eed5e9752e588d54238e5a547a002dd2dd75d180e974df04abe397d2d31d3040bd	3	0	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	3	\\x9c5bf8ff6cdbdd59bcf974c7dfa92bbe	{8248,10113,15711,16112,12114,11571,8709,11874,9226,14485,12653,9425,16093,13064,15569,9343,8490,697,13711,11993,15469,12841,10619,14941,1396,13532,10830,9401,15886,203,15122,13852,9931,13896,12066,1207,13024,488,15579,10171,37,8911,12360,15225,1377,14059,9860,14959,1364,253,15232,1970,988,10115,9637,10359,9301,11974,15420,11550,446,13028,11256,14001,9019,1205,1803,10371,14727,15187,15610,14707,13595,14610,15775,1716,8213,13130,13908,9965,16056,1008,13193}	\\x9886bb94398803c7c914380666bbb6a8d64a979789fa78891b90631f87d6ffbb6cbd803a85b7ecb58582d21a595447f9149410bf3c1ad12a689300b2477f6546181e9cfc6df9ee35029b04571e959cfed5d3056474d00855277cf4004ffd6ba7	3	2	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	4	\\x808000080002000010040014000020	{15026,15593,16314,14630,15938,15399,15348,15297,14468}	\\x92b14f075ba38c908a3f9addc66de4545c4908f600c3ef0f2505ccd3626b7b28974bde014282043b877846def228433112588d355144fd442d3e9fcfb8e4df23bb834ceb59024be6a87708e9aa5c153ff4b871495e232cf66f8631db620c0076	1	1	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	5	\\x806202000010000000000c80000a	{15814,15778,16192,15892,15817,14918,14664,16036,15889,15805,14617}	\\xae63534dfc0b459d104fdea8675b6edbc7332af38568dec1df6fa67476734196ee4cad474a78e64d826e6051e2c09aa2196ddf9acb01ffbff9fc132484df4b2c5d14897d992fbebf3d7d32200879e6115e388503ce159a240c052a5635ff9f0f	1	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	6	\\x000089c0018028000a10c020	{15486,16009,16159,15977,14737,14855,14800,15038,15820,15921,15652,14978,16163,16263,14895}	\\x908d3deba46f2cd34a06939095cbcd3692bd71577f7bfd2a65787102d69efde6d90575634ef95b9fca8ce7ad1229ed5a0d0f570666a8add8412024d6d747ff0ce12065cd4cdaa69311fcef09d10e90358633bd8650159535f8ff4e218d27ab09	1	2	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
4	7	\\x00140018124040288021080052200004	{15777,14863,15493,14840,15751,15691,16268,14406,15630,16045,14599,15944,15790,15094,15578,15324,15362,14623,16034}	\\x908e08a18ff3283d9f0712f1291c357de3887ed49cbf3371a2b6b4030de6ff5ec38f07de9057eddfebb67d86de0d061b0455f4388f86cf4ba289d3cbf2765b47200660b2884f7287abb075782983cf1c760ae42c6c7bc2cb1f8f5d851a4c19b6	1	3	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
5	0	\\x36df6efc7f7f13ef6b1bf57697eff99a	{15913,12717,1606,12942,9715,8924,12211,10801,14679,16210,2034,12236,10153,12702,10930,827,530,15743,13450,8378,10023,8611,15337,13858,11528,13303,11918,1303,11477,13567,11154,13369,8325,15720,12861,508,925,14852,12875,8636,1923,10081,8793,13279,12781,13813,11687,15043,9825,14446,10295,1402,11999,1652,9991,11655,11987,15478,11051,15128,12419,10584,11875,12530,15650,12173,11845,10661,8823,13650,9734,14499,10852,12503,10342,12690,1355,12916,13032,1675,8733,8908,9661,10199,14037,10380,13869,9777}	\\xb461c00c189104be86ab32fd004822c64dbfacbf71c6779fa8a4f71716238cf248ac66091ea79d01ae4a6a11dc72af4b1935fe53fbdf50f7fc9ab44cf319fd59f0ab56e14b6bf86f461b586914c49212df176d0608046518bad4b8a0f11ad05d	4	0	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
5	1	\\x1f33dffc597fe89ecb7b913fa57fcdc6	{10797,10068,15545,13490,13807,14120,12412,10410,1018,15414,9295,15460,8574,11158,9394,13652,784,8649,1728,12007,13865,14923,1746,9674,14340,1138,16041,8302,1407,10069,15709,8785,8639,15549,15058,654,15837,8282,12276,744,9841,11953,8381,8805,12254,9051,13978,10769,15210,11210,9415,163,9104,11220,875,1482,16242,10541,10763,10272,12097,15288,74,688,8794,16193,15236,8648,12610,12725,10156,209,10938,8200,601,14616,13338,15086,12311,9429,13241,11964}	\\xafa9651b764540b37a5c8f1f5f7950e2d29deddef63cf67cc41338cc2d7529bfdc5fc396dfc23f4151dc26cb82ee65d0135b767c9dba30592075807963c6710008cc64ef9fadaa2f0454f0115caa6339b00affe7d74bf3970b33ca1412aee168	4	1	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
5	2	\\x977afb16db320add1afea82f6c756f7f	{1300,13037,8836,15407,14750,2025,14194,1447,13517,8672,15715,9582,8523,10903,12508,9200,8745,8764,8596,13523,365,12983,907,1199,8290,11876,10944,9811,14161,14911,13833,13054,14460,11245,8883,13815,9559,11603,13681,15256,15161,1951,15300,422,9482,9361,16223,8849,8514,1125,13964,14032,15963,13196,9239,16293,13252,1524,9212,13484,13715,1056,12190,14666,9812,9966,9430,13162,11660,3,999,11738,12108,13273,9909,16170,12698}	\\x8db2f4ea8f6b9d49385a46fca1ba67ddd44ceb8427587be8f959adf95386c36172cb73a2947511da8d68b3ee419d87fa11e1ab99bb653b73d1e864279bc5c2232568465ef9ef092589866dd2fa43d2f5a6bfc263bfa918d309e4beff040e1b23	4	3	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
5	3	\\x6a26e7adac64db9fcceb7066654e2ebf	{15497,546,9371,11229,10566,1520,15241,10544,1875,9954,1449,11299,14573,14248,9878,11306,11121,1693,12630,90,15437,10502,1934,16241,11420,15271,14897,578,11652,16016,1798,661,11668,16032,14736,10848,9486,11595,12144,11124,14759,9214,10005,1549,15319,8677,8926,13119,927,903,15035,15562,11798,10442,9111,10911,9953,948,912,1638,8980,15869,15849,14504,11620,13223,15442,8244,11834,426,8688,967,12558}	\\xa76d8ca7a738963ca1f555e6ce8c530ce7b99f7549d0909da9fc7cf14c6585cd5c5c3e4c4e677e55d6f69efc05ab16781768362e1cf867edc6fce64ab9b8d5579101a979ab35b0d888dcaf41a199f43b1bbff9c36da1b1bc8115e882d7702004	4	2	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
6	0	\\x58fa3cbfebd9b2ef3fdf56d75fa3adbc	{998,343,8850,1021,15245,10841,8579,13007,14011,10432,1227,15957,11639,13190,690,12547,8508,9591,9535,14634,15261,14984,1190,15826,9885,12418,919,9507,16304,9579,14986,1783,10771,14071,11755,11712,11089,15907,14377,14442,8570,16212,20,10670,8782,13339,16371,15541,12439,12142,13048,8937,1710,10668,8852,9834,8408,1932,14124,814,13144,9250,8727,14430,8914,11352,11631,13612,8444,1924,9141,1585,9718,9502,9864,12534,1646,10893,385,752,9189,13118,11027,14698,39}	\\xb6f84fb20acd5a997be042ff479a8898bd896d764300deeed2e17ae60140aab666614481a6a207dc09aef5d74ede77741345ecbb9c8fb63bc5167a495e3c90244e65f2c1042b0904b54e0ff78a8ee47c73b1d3cd5df101fedbb171d0d3604340	5	1	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
6	1	\\x5be42bbfabaffdd925e2dff8fdb38f7a	{1812,1872,1653,10724,11674,11927,14089,9919,16310,14799,1137,12817,8978,12502,11785,8583,15270,13575,11521,14656,14276,14533,8560,15666,14282,14946,10360,14905,10243,9082,13184,14580,10275,16313,16126,9480,12813,12571,13052,15193,8318,13253,14529,8901,593,14359,8915,9382,8416,12778,8931,10117,14652,13824,16318,1078,10089,13605,1909,10942,10998,13313,13589,10457,14389,9225,10714,13426,563,15082,11157,12855,12061,8684,9711,764,13644,12992,15155,12688,15367,968,9252,311}	\\x8b4495cbe26b205b788229b314f54b3a1cd17e1f61420a42d9c4fcb12a81717b743fc57659eaf34c1f1f5633c9635833023904e81fe0e087ebe9f43d7bea3759a536337e1e2bf70b278b90e5f14cac897ac8ead4a772373e97a38ba5ff340668	5	3	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
6	2	\\xbc79a87e5fcfac7b727d3dde29f74ac7	{11013,15488,14629,819,399,9102,14877,9172,1913,0,14901,9570,11496,1745,514,14390,336,11852,8867,1758,9255,14772,8664,523,11850,10018,1025,10735,15868,15205,1091,291,354,11073,14925,567,15164,11014,14631,701,10793,10278,12383,689,9068,9669,9078,10663,14957,13647,11088,10550,12039,13180,10329,10855,8588,876,10701,9963,1478,1733,10751,8923,8276,187,14837,16315,15357,15341,13354,1607,12026,1986,11872,168,14564,11199,197,12895}	\\x824eecba20132b0799e6eadd159ef795de4737898257fafde69e31016583f2e5050b8e60d5fa43e1d6a41fca737450ad0814d2941ac1713e179f068f9c5ae4322d9a7b48f1e7b5b493304a98eddf0e9a8a39b07d3e3c39659f2a6ea88a744821	5	0	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
6	3	\\x77934e5efb1d0bf35e9fe30a0fca0f77	{862,1941,12107,823,15663,9001,9171,9289,12121,868,11903,8521,14867,11286,15787,11873,9778,9555,12652,11268,14299,11252,13909,9266,15551,11237,10808,15473,13021,16308,1383,1348,12777,11291,8544,8462,1446,8971,16023,12801,9862,14216,11471,11354,16039,9192,13718,15158,8651,13212,15753,11135,13773,15716,13734,8417,10446,16228,476,14008,603,9344,1270,8498,595,8731,9135,16094,9780,9742,11980,12859,15836,11920,1785}	\\x9477952e36901dc3c4bd1d2c2d13485dea1ad8b3dc3b32736117ebba91e5dc5505c62baf66f305e108552ad8e1fc814c0cbbcde2b232cff521e974396001f721308c88f6529db54c3ade66b6e8be901ccf2db85127712d419522908e8a892c0e	5	2	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	0	\\xf9ff47f79d7bd6ec675a1e97224db5d9	{12874,12666,1906,13899,9013,8303,10612,10230,13235,737,15566,9589,10461,9230,12689,8394,15773,15516,1041,217,10835,1603,14360,15185,10494,15118,15556,16221,1222,8737,15672,747,12699,14929,14014,10507,13264,292,9246,1700,1795,12186,11970,12830,12447,703,8693,14091,10434,206,8870,14791,13304,12603,13156,1260,13534,8575,11743,13157,9411,1132,8197,15927,1220,1800,14353,16108,10208,16121,8997,10960,8401,14558,11590,8423,15915,15906,10837,13915}	\\xad713a278697c4a732f76c9620868bcd09dd5ed9efeb06290d3bbd5255a6678f3f82b99eb231ebfa9ecd8f8fcad9b6eb10e5dc23e09bb7c503ea7c154010e045d94cb25c4eb7640a60970109b07661a3de15b76a81e89ce91e1e1ec2a357c98c	6	0	\\x880e0571d7ebb5566d998f767b9355bc45f3393081eeb2a9a68c9cba488cae40	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	1	\\xeee5bbfc68b8697fb16f3f5c1e7added	{12649,13801,9654,15840,1876,15223,12120,15464,14217,518,11346,10873,12012,9103,381,14674,11965,14831,24,11930,14793,11822,16062,9643,9569,15701,15703,13211,8787,14424,1329,12728,8830,12687,9279,13468,10361,14215,12488,9544,11675,9217,10638,13951,14993,15013,973,13385,753,9018,12873,1708,11491,16232,14751,9380,12459,13400,8902,14376,11482,15909,1711,13873,1684,16151,12473,1437,8681,14983,14561,12550,1307,284,14386,1093,1596,14654,13708,15435,11992,9307}	\\x8c6f48a218a970d8f7fccb0e4bc4e83d37c5d0d797d6493bc60a10097bd3375272bf8ec07a028349fa712c5f26820a49099625b976d4d2bee0482e2e3c5f94a2e975154f38549d3ec0a4f9b0b7e72b31ac6f90e9816155cc7a33b994d99ec404	6	3	\\x880e0571d7ebb5566d998f767b9355bc45f3393081eeb2a9a68c9cba488cae40	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	2	\\xd2fb3f77ffad9dbff7cab5f6fd10ed6f	{14321,9305,10921,539,751,13280,15130,12361,47,13079,9608,13856,127,10459,887,9666,2011,1404,12592,1416,10044,1841,10692,1389,15848,14019,15809,10939,9692,15574,13215,644,792,1902,12537,853,9506,16299,14332,932,13527,11161,15733,13688,8741,13246,9457,9324,13747,11847,14330,13639,10792,15510,10070,1072,11335,12515,12289,11573,254,13027,15093,15981,12022,10655,12351,14164,9664,13198,11846,8507,14024,1844,10263,13977,15629,9916,8289,9027,10454,8598,9073,12578,15268,11056,13720,8565,10648,9631}	\\xa50f40246f49c76dda238fe147b3ad24ad0b8ce18c349a566505e77d60f5d71319f154680d5e6dae3265f050e05d445700afb3b673917d96f74da3c617663dba371f20cb0bfa8255fc24d7f2f466121657a540c84350340e7fc7b9f121e80cc5	6	2	\\x880e0571d7ebb5566d998f767b9355bc45f3393081eeb2a9a68c9cba488cae40	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	3	\\x4754fff18fbba126f9d539ff9875fdf2	{15451,15772,1982,14336,8550,335,16186,1085,9347,846,14030,9357,1481,16230,16246,648,13445,12124,1079,13768,1789,96,14976,8546,8306,14098,560,16351,13435,12568,12263,8752,14717,612,9922,15934,12656,431,12510,714,15789,1826,528,9768,14358,8694,15970,8803,308,12341,8949,15996,10802,11455,9469,12757,8635,1593,11724,10142,10392,332,9290,9203,9713,10988,15220,8640,10503,10770,13432,12006,1485,10238,13116,667,12640,1954,15829,8331}	\\x9579dd4d2fed89b82b10c30f59420adc6330f36aaa33501d691956d32d5ad8ab2393f6333b8d1d76635db0a21ba53a3d0f4592ff147a6374dfbb2d789f024758b5f003aef6b44d7e12dfe50d4887974db335c02049c582122cade71fb12c6159	6	1	\\x880e0571d7ebb5566d998f767b9355bc45f3393081eeb2a9a68c9cba488cae40	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	4	\\xbc79a87e5fcfac7b727d3dde29f74ac7	{11013,15488,14629,819,399,9102,14877,9172,1913,0,14901,9570,11496,1745,514,14390,336,11852,8867,1758,9255,14772,8664,523,11850,10018,1025,10735,15868,15205,1091,291,354,11073,14925,567,15164,11014,14631,701,10793,10278,12383,689,9068,9669,9078,10663,14957,13647,11088,10550,12039,13180,10329,10855,8588,876,10701,9963,1478,1733,10751,8923,8276,187,14837,16315,15357,15341,13354,1607,12026,1986,11872,168,14564,11199,197,12895}	\\x824eecba20132b0799e6eadd159ef795de4737898257fafde69e31016583f2e5050b8e60d5fa43e1d6a41fca737450ad0814d2941ac1713e179f068f9c5ae4322d9a7b48f1e7b5b493304a98eddf0e9a8a39b07d3e3c39659f2a6ea88a744821	5	0	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	5	\\x5be42bbfabaffdd925e2dff8fdb38f7a	{1812,1872,1653,10724,11674,11927,14089,9919,16310,14799,1137,12817,8978,12502,11785,8583,15270,13575,11521,14656,14276,14533,8560,15666,14282,14946,10360,14905,10243,9082,13184,14580,10275,16313,16126,9480,12813,12571,13052,15193,8318,13253,14529,8901,593,14359,8915,9382,8416,12778,8931,10117,14652,13824,16318,1078,10089,13605,1909,10942,10998,13313,13589,10457,14389,9225,10714,13426,563,15082,11157,12855,12061,8684,9711,764,13644,12992,15155,12688,15367,968,9252,311}	\\x8b4495cbe26b205b788229b314f54b3a1cd17e1f61420a42d9c4fcb12a81717b743fc57659eaf34c1f1f5633c9635833023904e81fe0e087ebe9f43d7bea3759a536337e1e2bf70b278b90e5f14cac897ac8ead4a772373e97a38ba5ff340668	5	3	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	6	\\x77934e5efb1d0bf35e9fe30a0fca0f77	{862,1941,12107,823,15663,9001,9171,9289,12121,868,11903,8521,14867,11286,15787,11873,9778,9555,12652,11268,14299,11252,13909,9266,15551,11237,10808,15473,13021,16308,1383,1348,12777,11291,8544,8462,1446,8971,16023,12801,9862,14216,11471,11354,16039,9192,13718,15158,8651,13212,15753,11135,13773,15716,13734,8417,10446,16228,476,14008,603,9344,1270,8498,595,8731,9135,16094,9780,9742,11980,12859,15836,11920,1785}	\\x9477952e36901dc3c4bd1d2c2d13485dea1ad8b3dc3b32736117ebba91e5dc5505c62baf66f305e108552ad8e1fc814c0cbbcde2b232cff521e974396001f721308c88f6529db54c3ade66b6e8be901ccf2db85127712d419522908e8a892c0e	5	2	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	7	\\x58fa3cbfebd9b2ef3fdf56d75fa3adbc	{998,343,8850,1021,15245,10841,8579,13007,14011,10432,1227,15957,11639,13190,690,12547,8508,9591,9535,14634,15261,14984,1190,15826,9885,12418,919,9507,16304,9579,14986,1783,10771,14071,11755,11712,11089,15907,14377,14442,8570,16212,20,10670,8782,13339,16371,15541,12439,12142,13048,8937,1710,10668,8852,9834,8408,1932,14124,814,13144,9250,8727,14430,8914,11352,11631,13612,8444,1924,9141,1585,9718,9502,9864,12534,1646,10893,385,752,9189,13118,11027,14698,39}	\\xb6f84fb20acd5a997be042ff479a8898bd896d764300deeed2e17ae60140aab666614481a6a207dc09aef5d74ede77741345ecbb9c8fb63bc5167a495e3c90244e65f2c1042b0904b54e0ff78a8ee47c73b1d3cd5df101fedbb171d0d3604340	5	1	\\x397bb526568624eea3b29e27152d393019c5f9b0f37a3a0636ae6fd78b9143fe	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	8	\\x6a26e7adac64db9fcceb7066654e2ebf	{15497,546,9371,11229,10566,1520,15241,10544,1875,9954,1449,11299,14573,14248,9878,11306,11121,1693,12630,90,15437,10502,1934,16241,11420,15271,14897,578,11652,16016,1798,661,11668,16032,14736,10848,9486,11595,12144,11124,14759,9214,10005,1549,15319,8677,8926,13119,927,903,15035,15562,11798,10442,9111,10911,9953,948,912,1638,8980,15869,15849,14504,11620,13223,15442,8244,11834,426,8688,967,12558}	\\xa76d8ca7a738963ca1f555e6ce8c530ce7b99f7549d0909da9fc7cf14c6585cd5c5c3e4c4e677e55d6f69efc05ab16781768362e1cf867edc6fce64ab9b8d5579101a979ab35b0d888dcaf41a199f43b1bbff9c36da1b1bc8115e882d7702004	4	2	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	9	\\x977afb16db320add1afea82f6c756f7f	{1300,13037,8836,15407,14750,2025,14194,1447,13517,8672,15715,9582,8523,10903,12508,9200,8745,8764,8596,13523,365,12983,907,1199,8290,11876,10944,9811,14161,14911,13833,13054,14460,11245,8883,13815,9559,11603,13681,15256,15161,1951,15300,422,9482,9361,16223,8849,8514,1125,13964,14032,15963,13196,9239,16293,13252,1524,9212,13484,13715,1056,12190,14666,9812,9966,9430,13162,11660,3,999,11738,12108,13273,9909,16170,12698}	\\x8db2f4ea8f6b9d49385a46fca1ba67ddd44ceb8427587be8f959adf95386c36172cb73a2947511da8d68b3ee419d87fa11e1ab99bb653b73d1e864279bc5c2232568465ef9ef092589866dd2fa43d2f5a6bfc263bfa918d309e4beff040e1b23	4	3	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	10	\\x36df6efc7f7f13ef6b1bf57697eff99a	{15913,12717,1606,12942,9715,8924,12211,10801,14679,16210,2034,12236,10153,12702,10930,827,530,15743,13450,8378,10023,8611,15337,13858,11528,13303,11918,1303,11477,13567,11154,13369,8325,15720,12861,508,925,14852,12875,8636,1923,10081,8793,13279,12781,13813,11687,15043,9825,14446,10295,1402,11999,1652,9991,11655,11987,15478,11051,15128,12419,10584,11875,12530,15650,12173,11845,10661,8823,13650,9734,14499,10852,12503,10342,12690,1355,12916,13032,1675,8733,8908,9661,10199,14037,10380,13869,9777}	\\xb461c00c189104be86ab32fd004822c64dbfacbf71c6779fa8a4f71716238cf248ac66091ea79d01ae4a6a11dc72af4b1935fe53fbdf50f7fc9ab44cf319fd59f0ab56e14b6bf86f461b586914c49212df176d0608046518bad4b8a0f11ad05d	4	0	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	11	\\x1f33dffc597fe89ecb7b913fa57fcdc6	{10797,10068,15545,13490,13807,14120,12412,10410,1018,15414,9295,15460,8574,11158,9394,13652,784,8649,1728,12007,13865,14923,1746,9674,14340,1138,16041,8302,1407,10069,15709,8785,8639,15549,15058,654,15837,8282,12276,744,9841,11953,8381,8805,12254,9051,13978,10769,15210,11210,9415,163,9104,11220,875,1482,16242,10541,10763,10272,12097,15288,74,688,8794,16193,15236,8648,12610,12725,10156,209,10938,8200,601,14616,13338,15086,12311,9429,13241,11964}	\\xafa9651b764540b37a5c8f1f5f7950e2d29deddef63cf67cc41338cc2d7529bfdc5fc396dfc23f4151dc26cb82ee65d0135b767c9dba30592075807963c6710008cc64ef9fadaa2f0454f0115caa6339b00affe7d74bf3970b33ca1412aee168	4	1	\\xa8aacebb4b8bb3af15039eb2881202432fc764be367d414806cc7ce04255b4ce	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	12	\\xad9b7e9c25dadde3c4efef5feefa8cd7	{14662,13831,14454,8721,11705,10601,15283,12081,11,11804,369,1760,13584,12648,9934,11960,432,9765,275,12742,9923,9472,11305,13466,152,14290,8446,11403,15476,9127,14823,11606,237,8196,11665,8740,11762,1747,9320,1992,11699,12933,1853,12657,8421,14463,14233,13537,9787,13410,12303,15391,8311,10615,8992,14086,16238,8712,15940,894,15263,12118,16164,13023,29,13524,11910,14189,12593,9278,14415,10587,12373,19,10011,14220,12253,14361,11616,1730,10932,14249,699}	\\xb48991cd81d51c017d88d1fc884a909ba08e9d7c70b84b16ffecd0eeb0536f190fa7865fd518acd9d780582dfaac5fc716631d050378398d68102cf6195320f47c03f7d71e6ecd3bfec05059555c701434e302e83bb7b49f870dedef5516382a	3	1	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	13	\\x9c5bf8ff6cdbdd59bcf974c7dfa92bbe	{8248,10113,15711,16112,12114,11571,8709,11874,9226,14485,12653,9425,16093,13064,15569,9343,8490,697,13711,11993,15469,12841,10619,14941,1396,13532,10830,9401,15886,203,15122,13852,9931,13896,12066,1207,13024,488,15579,10171,37,8911,12360,15225,1377,14059,9860,14959,1364,253,15232,1970,988,10115,9637,10359,9301,11974,15420,11550,446,13028,11256,14001,9019,1205,1803,10371,14727,15187,15610,14707,13595,14610,15775,1716,8213,13130,13908,9965,16056,1008,13193}	\\x9886bb94398803c7c914380666bbb6a8d64a979789fa78891b90631f87d6ffbb6cbd803a85b7ecb58582d21a595447f9149410bf3c1ad12a689300b2477f6546181e9cfc6df9ee35029b04571e959cfed5d3056474d00855277cf4004ffd6ba7	3	2	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	14	\\xf7587fd90f6d7c18ed6ff5662dc272e7	{891,1817,15677,587,78,14696,15628,1458,788,15515,15273,261,944,10788,10250,1942,13073,14495,15774,16147,12872,15338,394,9697,10624,10437,15969,12793,13390,14185,14720,14896,14509,16297,11579,11510,1649,9893,15524,1221,11126,242,359,10472,9810,14924,791,13301,13105,10862,10913,11143,808,9833,12375,9010,10914,12976,1619,16250,15871,8305,9017,13533,10377,8788,15813,1781,14416,13206,1162,16206,12499,14351,541,8569,15833}	\\x8f773105786875b9b037da39a1675ba35f1807453f06df522c1a68f28cb48ebcfdb8a691ba69da67ca965b76b58a07c80fbb49a7ec29e835dca654201aca35eed5e9752e588d54238e5a547a002dd2dd75d180e974df04abe397d2d31d3040bd	3	0	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
7	15	\\x9f5ed97bbdfb22f942f1f947c17dc4d3	{9341,9376,13430,16207,9970,8717,11856,11386,10574,13056,1255,14198,13872,11249,8826,12745,260,13417,11323,10789,466,8703,8220,12148,628,14553,10539,9521,1253,8510,212,11187,10518,10397,11625,11972,1977,15141,12798,448,15267,11577,14567,439,14796,241,8271,1038,10579,9105,9869,12572,1062,10608,13679,15587,9944,1386,9522,9943,1807,8795,9699,1819,11820,12849,14598,271,8892,687,15388,14513,10041,13101,93,14632,15991}	\\xb2f04d5ce585de3e13d571edc13c5eaca8df1e9208d385fdc548893129840dd4f6ca58d3523768827ab3e64ccd151bb407e69613b80ae003202546cd43f84fc056dc8d1cfded6223450ccf8538133ddf1a37325835dceae1924783d7a420f3e5	3	3	\\x5981ab300e74f9132fab1a5ce8ddca3addaf893b56d3b2ed05b72c0301533f44	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
8	0	\\x3fe72b6f7747eb8b9c50fdb776efaebb	{9009,1460,14370,11441,11217,8928,9237,1665,14526,411,8333,992,12196,8491,171,11092,14948,9849,1622,1632,10022,13441,14622,13600,442,15459,1031,9832,211,8404,11239,10497,15336,12542,10126,1251,15175,1045,12164,9634,11024,8288,10079,13336,9451,12814,9720,1825,11064,9921,120,15736,11981,9374,1156,12312,16245,16064,1907,14704,12158,10746,11841,14718,15590,12803,9781,8742,9260,9195,9714,12930,11216,12647,11952,10003,11698,10182,1810,16067,14255,15284,15933,11911}	\\x87ab610c9930d2e3711984c945945c0f1d1d1502df801d63c196d4379a8c3be44487b0b6c0688c1a38c99929f5e99dc9153b14405c52699b0c28071053dd26e93eadac5476ca201ad3a6e27922987cfaab18c55c4c51ca52180d4bde48f5324a	7	0	\\x8e335a9cfc31352680b1a366c3563116d8c6ec4e5ff2aa4de9791b162ede631f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
8	1	\\x67cf1b3e3cbf529f08f56dfbe8713f5b	{13967,16198,16258,14699,12310,9440,14143,12195,8708,11127,14990,10560,128,198,572,14954,10622,10010,10952,12243,15400,10870,8477,8699,12854,13268,12336,15631,15964,309,12676,15949,9352,9979,9705,9042,9888,14289,8907,10516,10728,15329,9549,13068,13224,15737,11394,8857,13620,12990,181,15409,16372,12424,11800,16144,9492,1153,10867,12461,13663,8755,12491,13531,15623,16281,8976,14426,1036,14572,1529,9228,8637,1864,14842,725,10195,11497}	\\xb097cfadfef977f46da5ba8a2b4be7c8e4c2675e2bef4ce17e5098d7bef600b2bd48ad644a38862c48169624f068ff7a1512818364250d8b1d9283c649dd55d2e42929ada6856400d04046bee61d51410a095f75b3e85a446a1f7d1f09599969	7	3	\\x8e335a9cfc31352680b1a366c3563116d8c6ec4e5ff2aa4de9791b162ede631f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
8	2	\\x47f6cffb13a7bffa679b9be9376cd3af	{13739,9794,15091,15445,13242,532,8768,8828,9598,9434,14723,8414,13238,8680,14232,15019,15352,892,16217,9816,10947,9392,8808,213,14692,1160,8890,1254,1936,14187,14420,9089,9431,13478,11565,16353,10695,13245,16226,10347,9287,15914,16027,14618,9264,9,15167,11108,8847,15727,12773,15990,461,41,14961,8810,10957,1103,772,1191,1424,11706,14046,2036,9796,12991,9822,10106,15688,16235,910,12350,8201,12644,12198,13820,10390,8934,12163,834,1024,15365,1048,13809}	\\xb7c51e1c83575005854a8e9f69c48b71c9a261af970fe975346187df73a72561b2e0724def5ab689ae9c1000418e92050165ca40c8ff0ed624d4c4d83eb079c5823aa5cc788c0a4ff3a5689b9e8f5ece56108d8961e7337746172bfc647800fb	7	2	\\x8e335a9cfc31352680b1a366c3563116d8c6ec4e5ff2aa4de9791b162ede631f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
8	3	\\xe01f7abc700def8baae3fa0f0aa5a7de	{11164,13749,10217,362,9299,13176,15617,14297,16086,14804,10394,11277,8692,12553,14769,14436,1192,1542,10861,9125,14945,304,12904,1598,8882,10078,652,8935,13785,13699,8743,15052,9752,14294,10282,14836,12607,8873,10094,13518,1526,9113,11450,9962,9081,15183,1694,14919,8486,9764,11728,15079,12145,1707,14894,14890,16088,14077,12546,9721,12977,1428,16290,1223,9493,8413,14906,16099,12549,15347,389}	\\x83aa8a72ca6a39f26ff9125c58f3bcb1a0190b0b3ce385b69a9a90e12de13763751834cbda7192646d22741f59cb1677048b2348760323b3934d39730a01e21b916de46d5e956025fa39d97a8b115c4fcf5da721ef9fc35d892ce7b52fe2b14e	7	1	\\x8e335a9cfc31352680b1a366c3563116d8c6ec4e5ff2aa4de9791b162ede631f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
9	0	\\x7af4f69bc7fcb6fd632d4bdca4aee462	{12258,1712,16174,15178,15592,1597,10242,8284,14149,13767,1267,852,1327,1004,13171,10197,13843,14042,14163,12272,1326,13181,9984,8811,641,12241,9994,10982,11548,392,9725,243,1645,8967,1586,10256,15888,8292,10239,13421,16378,13522,13470,9795,10012,9164,12157,8765,1225,12820,281,10492,13294,1914,1605,1797,15290,2028,8557,15127,8549,1037,1945,14078,11717,285,1375,1969,13091,1107,659,13796,8621,8472,10825,91}	\\xb30d5b6022e31009a1fe96d1b967572ef8bd853e26177f4cd0fd8e12b220602cc9e6de0f913d05004b2cf96efab7ed8604eb5267227b098075ad5ffd2ad00e8d76286be0a739670f52878e66a8c303f3539e96999bc89b9babe5ea25744598e4	8	1	\\xb0609f8cedcabfac555b91d07d60615f6e345b11f279cf2d6b755087589f52d9	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
9	1	\\x94e7e9ba83017ee68c8f4a5bbefb172b	{11686,935,13655,10705,10606,12956,12774,13539,1824,9452,9942,10809,1347,14546,8577,13001,11865,1269,8851,220,9086,11764,14113,10259,10646,1338,414,8377,8314,11752,13178,13956,463,990,8779,12371,9798,13309,1641,8634,12883,14403,12490,1049,11221,13546,680,14047,14383,9683,9971,12340,9741,1944,12521,374,13069,323,11961,10629,11932,10378,13613,505,154,10150,11165,14214,1236,12337,13780}	\\xb3a4f55155ab5cb2eb50ccdab1ea322dfb913a50b9a6fe12c96aedf9a97822367f6ed889b74cebe4429526bd2c67244d19baa19b550fe728b37e1fdd1bcb554b16179d60ca5c6684ab0eefcc15c0428801238c10a10fd96c2b3d7e0c3e9156f7	8	0	\\xb0609f8cedcabfac555b91d07d60615f6e345b11f279cf2d6b755087589f52d9	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
9	2	\\x9fc12aff1de8334c4a539d58ffabe981	{10542,2039,12095,8370,14579,14797,14141,126,11890,9150,581,370,1109,11363,11669,16157,9838,8259,1563,13580,15425,886,10605,12049,931,11259,13900,202,12579,9874,1591,829,13558,849,11892,11225,13226,8592,837,12093,9853,2012,9642,13698,14191,239,14818,10331,225,15835,2022,15878,1360,9751,13343,9491,10234,10652,8482,8925,8961,10232,11271,11347,9706,10902,13563,12280,11888,12642}	\\xa092ba263ca5a519e6c895a78de7f76049c3a1b934ab5d1d7c79bf62b13511b63055c4890ef291983b9333606138fa8714816cfc3665604cb5ddb4f44cd98c0b2f002b5a195ba7214b0827b9a5fd595f9fd7492d7347b686c4771188844f5345	8	2	\\xb0609f8cedcabfac555b91d07d60615f6e345b11f279cf2d6b755087589f52d9	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
9	3	\\x77a23ec9daf81a322c61dc27fb9ee13e	{8822,13506,10133,12304,9375,13465,13408,13208,10356,12183,11535,12919,10765,10336,11349,16148,14119,13058,15120,11810,14003,11612,11837,12008,15795,11520,9724,15110,10004,28,441,8488,15080,9385,8506,14694,468,1061,10908,9615,8372,13912,10750,838,14335,11881,137,13736,9208,17,9978,9641,8595,13876,980,12234,11591,14891,9820,10900,14714,830,8391,10796,11232,12431,16270,11954,11137,10471}	\\xb5cf3c98e5c92e125deec1a59fa450403c785f1343d325c9a789d8eb48d3c473f388c1051196e2358625d3909940c4a10d747878dae355384001974bd68b94e6151df917cc4f9573c553c742f89a6dc8f8205f5122ae40e0a19919d54fa68880	8	3	\\xb0609f8cedcabfac555b91d07d60615f6e345b11f279cf2d6b755087589f52d9	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
10	0	\\x7a064b3896220834156ae66f7ca09bfa	{10169,13793,9133,8869,14105,393,9688,274,10931,11607,14135,13621,1530,13359,630,8412,8568,11680,12812,12322,11531,8261,9548,9134,12264,748,12569,12024,11854,13371,13094,1582,1766,11284,158,13191,10088,8475,1689,11511,11128,8343,12953,10984,12319,13179,13508,11469,12167,10305,859,10536,517,9770,428,10634,11115,1463,9483,1043}	\\xa40211daec174b78e32334b6b74373ad442210c001ff4ebd0f5b4a667a85d37e19f1a03ed8e73666a5450d12374417f914a5f53d540d5066471279e5b865f7718bc03eb65dfc5c0112f4738038ee9bb234f00b9ffcd7d71067c620dd3f02fa35	9	0	\\x84cfe7cdcf68654ec286d31f21f5e894d36b5ea2a37008ea6cce97e9fe0f9d06	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
10	1	\\x4153152f315571a06fe3be3609dd1f8c	{923,8453,12101,11439,12730,1688,11302,36,10381,8520,12566,160,1376,14061,9023,1723,9511,11732,10632,11258,9487,9981,134,9823,12616,12556,11041,9224,8538,9915,11472,11454,10934,921,9616,8735,10509,12321,1806,1837,739,9819,8567,13926,11263,14160,1034,9146,9201,1956,11475,10918,1548,1200,10200,1210,1869,317,12069,9951,10367,12617,9728,11333}	\\xa5641606e963143784df8ec139e596cc81a9af92d32ed9047672486badf88b3d5ddf731bf3d1b217b2250d5261d9e7b31485405bdae0ff5b5906093ce0d9c274e0f028924c900e671866f3975cf0ef893e5d8ae8dcd723dbc467a62a3147040a	9	2	\\x84cfe7cdcf68654ec286d31f21f5e894d36b5ea2a37008ea6cce97e9fe0f9d06	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
10	2	\\xd10bbefd5a4f53a26bcf40624b08362a	{11318,11146,10821,12786,13250,9179,185,11117,11702,293,634,1146,11215,272,10452,1394,9958,13174,9063,14177,1644,9792,12354,12194,10325,8754,10899,12551,1120,9090,14029,11180,1788,11558,14175,204,13719,14173,10488,12433,95,13213,995,12009,10118,9580,9106,25,11707,376,12509,8535,10445,12772,9036,12199,9355,12132,790,8633,12507,805,10050}	\\xb38c5e052f6fc9cef157292e51ec99b1a92d715417cf7b2d45ee0b8a4912886756ecba4011c635040612aee54bac309e0cfc368f07e132fad6bc5d20c4b7cb42a723ede1486e9e1ddbcf624cce890692c4bc3066d20bbb08bf448d3021e6d6c2	9	3	\\x84cfe7cdcf68654ec286d31f21f5e894d36b5ea2a37008ea6cce97e9fe0f9d06	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
10	3	\\xbb94d2d177fd467b9e3b5bb6e9c50573	{10072,1704,10776,1835,1531,509,10673,11611,8327,1629,1509,406,373,12002,1480,10060,11667,11353,13712,8258,1422,8452,9011,10289,10474,10158,13564,10553,8729,1427,248,13448,12300,13195,10517,1950,9831,11448,8225,9540,745,205,10483,11319,707,1991,8747,166,8298,13391,13948,12192,877,12634,9062,1128,9417,1239,10598,1740,9532,13731,11307,1866,12980,10851,681,8772,149,9442,12342,13841,12969,1514,10222}	\\xb248ed5972f07f51389553c24e2271065a4b418a236d126b94b351d5535da6238c2c26f93b6ddd392671ac27824f03ee0b181df367cff819c5f13bdffc9091d855c4eba5120cf691674bb2a077687cd65bbd8d1054dd380aa3b18cd082444fa6	9	1	\\x84cfe7cdcf68654ec286d31f21f5e894d36b5ea2a37008ea6cce97e9fe0f9d06	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
11	0	\\xbea9c630bf434cd1e4217c7dab9bcdae	{11683,1570,11409,1193,11111,1044,8853,8710,12295,14125,10725,13423,12090,9665,16,12834,9528,10310,14111,1884,14266,10415,8448,1734,403,13315,9444,12265,123,59,573,11248,1987,10047,8209,13823,424,11015,13821,8255,12536,10303,13461,10753,10992,9903,1790,825,13310,8591,10237,10674,11868,1500,1965,147,569,8400,14002,832,9543,9025,1076,11415,1003,12072,5,13714,35,11059}	\\x8a514ff2f6518fd18ba5a575c74e8703797becabaf97914f5b56b74e80c1aa25bbc16b4132c9a6118cb86ec3b7596e210a41df6aaba8b6d826bcf687f9a318db1f411c2b44c7f2bf83316e9e5f6ab186ffd3430f576d91f6ca734fe194570466	10	3	\\x940d5804ff78f14152cc5a6140a54526f586d383d86ed3e412e894eb616fb851	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
11	1	\\x6b50a71a9c34012c46d38e7d1df77506	{674,10358,58,13146,10248,12850,8479,12449,14252,11722,289,1104,14093,10532,1298,8337,9969,12835,10683,13306,512,12477,9723,11825,1312,13158,12063,1358,14167,10034,1214,12754,10056,11813,10521,8493,11529,14260,12436,12215,9408,10086,10820,8470,1066,1301,10779,8354,958,13950,14010,9080,1073,12363,10849,10600,11634,8602,11485,13958,9496,11228}	\\xa894109461543fabb639752a100544aab0447b2c94064dd7e19ea239ec9dae5fea1c11b8aa271544ac32ee25a1d5a1a613caf1332ea7f800ed3d9d41933d86f40124cc7f0bc7e102b9e9f21dd66dda8bf301c9466f6f0df175326bd774f04ec7	10	2	\\x940d5804ff78f14152cc5a6140a54526f586d383d86ed3e412e894eb616fb851	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
11	2	\\x6b54ea9863b1f0228f643929e45e0902	{8871,1696,10826,11940,1678,13188,13819,11342,10178,13078,8696,11513,13772,12862,12517,140,10114,178,179,9182,12821,386,8872,1042,599,9563,10639,1492,13356,917,124,9581,11442,9391,9861,1095,13233,10212,607,8552,353,12809,13332,14146,13538,10950,11949,13255,14025,1216,14007,11939,142,8854,12043,13018,1284}	\\x81e741ffad1497f6f9170b0ad7cb467ad0df1f0c574f43dceba74b6df740d7e9deefd7d74d8a41443bcb8d929df54cbd19d12d5ab80a6ce3af35f9b05bb350219854acedfd8c97cc59227a8d55faf20aa6e01937b2f7e52e691243a8c30f676a	10	1	\\x940d5804ff78f14152cc5a6140a54526f586d383d86ed3e412e894eb616fb851	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
11	3	\\x84e87afea7430a325c51d2c1ad468440	{1213,11937,9504,13251,10431,13420,10963,1539,711,8653,344,1502,564,806,1387,9801,13701,11786,10369,11428,9176,10389,1561,13963,433,11971,13840,1883,13108,10057,14218,533,14151,8254,8689,1328,10990,11250,9108,9358,8457,8585,100,800,1098,10604,11125,10829,12739,8353,12737,8987,12626,11998,11659,590}	\\x8816bd068452958b042f97c12be31f7bdeff0c6cc3ddc862057bee2a7b2a170a4cf7913c0f70419333406ac090dbe7cd056ebe305b8e0abcd03e50d6cc86d26f872d26de0577a74b73a22f61d5ed65ed64071c2a5a9c9001e5992f01b02210cb	10	0	\\x940d5804ff78f14152cc5a6140a54526f586d383d86ed3e412e894eb616fb851	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	0	\\x1155897f787b72635685ee7da3b6a540	{646,1962,11355,501,11445,878,13288,12573,12964,13152,1777,8831,14140,13757,150,10019,8887,12641,8382,21,382,13084,10924,10886,10514,12288,9659,11922,13649,10991,11042,9769,11948,11407,280,9412,12981,8891,810,10731,13519,1461,9964,13414,450,14329,10172,11438,9405,10611,1702,12526,9640,87,10794,8643,14048,1058,9168,8348,10008,8481,13700,1352,12301,8249,13565}	\\xb69529bcea2bc8d7a664d7566fb78281d52538129ec39c00853ce79f7f69727ee17cfe8df57d154a7a425efdcc9e8a1b01113d71da2a111e3cde2bedcdd81aa19eeeec74ab1ce304475fb644b2d67ff62b0bf220b342dca720dfc3261cdc073c	13	2	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	1	\\x7f206b50eae044972efcad734989e769	{14171,12283,11849,8986,11601,1535,1246,110,10897,1608,11093,11375,13346,1523,9924,1249,12986,9210,13597,12683,14275,9501,12996,12413,8373,12597,8517,12946,12731,12824,10054,9602,10051,12999,11663,1518,9021,11061,12681,9274,1558,897,10995,10398,12037,1784,12596,13377,785,1544,8629,12206,1268,13599,9366,12464,13982,13541,1304,9043,13794,11292,1264,161,14058,12696}	\\x80126ec8af4ac504e18a335ddabaa49b96809d0b0e4f73e747a0536aa0ef81b650d284b9dfcfad1b65eca4163656c7850b2a3ad7b83110bc3c7eb51cfb6b671a6b314b69d2ebb2226d094acf829f9e745fbe57ab00e44e01fefc770cd34786c9	13	1	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	2	\\x4c7c48c6a4ea7347533b1a145d4ed50e	{1090,11047,11685,8495,11095,10994,11467,11933,12879,9022,8881,10761,11723,10065,1479,13854,9443,11020,10928,13942,10137,13082,1169,734,400,10213,12505,10748,11772,14121,11490,11523,570,10254,151,11022,10824,13897,631,9004,11586,13847,13830,1917,1741,356,10989,11973,12355,804,1425,8933,2033,12887,10339,11671,13954,8753,13631,10700,138,13406}	\\x83751ebbeae7cc2dec0882fb055482befa90147bd950bbfcfd9f0442ef1f043abd284a457d2232c9bbe4b0c66d15e903173035994567e7ca590da03b345b1d42024c6cb093c0f52e42f92621abcd098086402a5c33b2f5ea787803574acc3a27	13	3	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	3	\\x561bd693cd8231e912479784744c0e92	{1456,850,1717,12156,321,10105,11560,14303,12513,13846,10993,10052,10552,8438,12406,10062,1627,13015,99,1809,8666,10719,9710,13062,10188,9772,13549,251,377,13705,9858,9173,11223,9948,1780,13516,1692,13462,10500,12995,13247,12059,779,9527,9497,11406,11654,8736,14186,588,700,13327,833,11775,1170,10075,10671,11946}	\\xadc47fbd79910093acc44eef967ec295d0c28af162a55892eccccaaa63b7687412b5015db8c558a9e244880e2dffaeb0098b354f1984831d4a92d6f6666804e6102ac3272e0c4f05da9a9d6104c490922d70ff07e713c737b18863ef34690453	13	0	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	4	\\xa39aab245895f8ebde1166dcd4269fcb	{1110,13845,8531,12019,10726,1868,494,669,10834,11193,11400,11329,13870,9748,12952,13330,9632,11725,13862,12633,12943,167,579,11877,1484,1451,8403,1848,10292,11812,1813,1135,9917,8443,14104,1496,8487,1189,8199,9556,1180,13555,13407,768,11001,2044,773,13150,1281,754,13569,9821,11312,12437,13111,10235,13159,888,9545,12975,12014,8437,9423,1007,13587,13733,368,10013}	\\xad109bfc17f95886b65becf05f62c50b2b4fe5304623822a7e1092f5d0e5a732e5c25d75b96f53c6de8955744f981d14026efacedaf6065d7adbf785b7d12be82e062694164f965296c12cbb5de4900dc2994e7b7614e47f22d007a1ff4b4d74	12	2	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	5	\\xf4ba95246b1a5628f4fb3ed2705abf08	{9066,977,11727,8519,44,8955,12046,453,9379,9475,13142,11069,13045,12831,8513,12042,1873,14118,9128,10286,13504,11269,11709,13578,11646,12352,9656,9441,11035,9890,9783,8654,8522,9365,11866,1794,13576,10630,10884,979,11417,11294,10138,8946,9180,13040,1568,2023,1610,9901,8790,13296,10221,13917,11091,14166,12804,807,13798,13675,34,11996,9152,12665,9673,13561}	\\xa27169d075532dc22ae1c8a582cc84dd72d779bcf468f08fbdeea0c3336063b30c606d26314af5beca1afdc04e57d7e90cf9a31810cec00952fe222126e5f0cc924ac1353caca798f02c6f28fb6fac2ed9c6d71d54e079f15d83993722cf33b2	12	1	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	6	\\x9ea467c7bc00246bb6dfd707108719c1	{12763,794,1621,114,8455,861,13729,2002,13630,1166,8532,13660,12829,2020,13122,658,9197,1035,1238,13871,1278,1321,9460,1077,12213,11676,8235,15,13983,9281,10338,10183,14128,9611,165,11618,1744,1898,11759,1089,10506,13085,1286,10966,12084,11373,9204,10922,13013,900,8888,12738,12237,10074,12028,8906,9406,10025,11178,11243,1173,1102}	\\xa83058948256223701ec40fc095e54ba332139a2b93eb08c147c6488cac534cbe04804b21011966368b74b2a69bb26b91989131378f91d9c9e9304f0c4fb7c76fde9b673532ff3f6bd3f5d01a7655a0f3c9494ac51d61c82b3e7fe0b94d44eae	12	3	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	7	\\x3c80d60b3676f7278925a09d7e909418	{139,11489,13816,10551,9158,13693,9101,12845,559,26,13633,13022,14122,1494,9109,10033,13134,1623,11831,8300,12307,13513,9283,675,12780,13817,12396,966,11763,200,8402,11567,8211,1243,13753,12564,11599,8686,13403,13454,1367,9938,11119,10099,8451,13317,8286,12890,14074,11610,1181,10478,8628,8529,1129,9354,10817,8268,11689}	\\xb2f3d3b918b2204a0312a13f2e052ef82c9abe2daca089ca23b4e802932154ea3eceda79dc19745e8813a4cfb5a1b8e604c6e7ea347bd192df7c0749ddd02ad1db24ba33408d1a641c29a7fb02455b05e9f9f5e54b601a77b761e175d99945c5	12	0	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	8	\\x89e86cbff441cd22cf99964496750ddd	{13236,437,12393,9929,192,13104,12219,1431,11138,8806,12570,8725,207,12364,14300,10425,11226,10174,854,12055,89,11985,1690,12871,11338,1211,8866,11008,9530,13594,8594,9655,11871,9060,9800,12588,13308,9679,1063,10287,1904,9606,8216,10349,1229,410,10925,383,11742,10317,535,10035,13347,246,13352,8704,10196,8900,11794,10082,8409,11402,12766,11371,303,1536}	\\x807aff647b9a8142898495859fb6992e4ac6348e647b4ff190242ba2d6dea5eeeaad6e06cc9568e2cecb09fd07b69ead12be144359db134f13759ea049d72cea2c8353bf70121a7394eb990b4abbbd1de49569cbc7f24643ee34aff67accc447	11	3	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	9	\\xef60041e439547aaf1f86d50a5b61e1d	{14317,10937,1218,1556,963,9744,9701,13690,164,10374,13230,9554,10833,11122,11230,10148,10798,9280,1487,12226,11206,9481,14230,40,8630,10185,11882,9987,14116,14268,14097,9327,1560,8644,12632,8251,13387,10773,1999,989,10475,361,10755,953,9700,10571,1118,499,11478,914,11436,498,9575,14070,1828,231,10616,14004,12949,952,10372,12511,1475}	\\xaea835044d330c1110a360c9dc33e6a84c746381fdf0d5fcb871d0d73d0c636a48ef5ef63fc9f9da08441ac4269c82a3058a3064512ed44c9fc4e050012b45911e8479111dadb287dab1aea7ed5a2941ee4bb512a7305e7f1cfca87dc32eb6bd	11	2	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	10	\\x2c1ee5d58d2cf1f82517b015d2a18b74	{9897,245,9477,10160,12753,8706,13075,9708,13993,1834,670,367,13850,9682,434,10836,10333,821,8537,11103,9930,12189,11519,13890,991,1318,13404,9889,8815,348,10926,9703,10214,11382,11133,454,8669,12218,378,8326,10343,9814,11296,11494,11967,8385,14271,1519,11270,1755,10781,11374,10743,10617,11113,1791,9304,12223,11431,11297,11101,8734}	\\x936c99c47f86fcef4bce78df6a79cce36884c23045470f50cf50179e72fe963872bb4354a6d47fcf7a69dcf94eb4384f04cb389dee47ea3fc2b4441e1ea714cb349771ff30497eb1839519b7a365c7bc6ab17f4cf9c95b7f03e6202d028c352b	11	1	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
14	11	\\x12bf69e5ec7511991221683d9901ca0a	{9107,12765,1080,10496,1609,10529,9676,10384,8277,11493,12941,12465,11585,9612,14221,12471,12638,1721,12784,534,1624,13440,13581,1698,1164,8767,12025,13155,49,8386,11183,9478,10366,1731,13436,10375,12168,9206,766,10508,12905,702,8962,1317,1272,14095,1770,12259,13709,9222,10399,8432,12166,10357,14115,10803,1292}	\\xa882e74b7ddd3803e5027aa3ea59374cad77f9833eced734ab1b911a49bc2f2b6e28acc6f7a22799ac8da22949f74f341666d5d75ddde31aae48b9a98ae15db8f4651b64902d438cde771b006493770bbec2ee5b28f9a3dd3764395bda720c88	11	0	\\x5effba63f2499456c65c7259010900bf1b31c0024f3af3a1110344bf5401240e	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
15	0	\\x69645b89e7b5f276ffacf6bee986dfe2	{9048,10253,9596,9912,13826,12516,9818,13812,13177,9297,13789,13918,119,8509,193,9829,14199,8323,8843,1943,11339,14293,8903,305,10856,12797,8485,8245,10838,1167,574,8989,10647,12052,13786,12456,9473,13955,12296,10006,12514,11242,11393,8566,10241,14000,1750,11509,10762,65,1081,9947,571,11141,8833,9974,12468,840,11753,2010,11219,10391,1071,12386,8478,14049,14196,13219,12715,310,9514,10322,9937,11524,8678,10919,2046,12823,12040}	\\xb55968ea3d85bb5ff4462b7df1b75cbfd87b34f7eac7609da8dac882c22c1193bc4aad271c2e73355bf2aed61be5d27f108cf00105dfff9e9a24b81e094e8cf62ff82a330c5c61ac4c18eae2f0641dc1259250831aaf2eb699f0673b779ae824	14	0	\\x4f483c496a117c227f63cdb5785ddde6ca5d73831e281445627e9b82a3f62b2b	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
15	1	\\x4cd2afb46ab18c2d16b1873373817f64	{13494,1576,141,1592,10843,14228,407,8468,8799,1154,9881,12286,12629,1310,12004,10669,14033,976,8198,1571,8233,9973,9072,10645,11007,8584,11501,8287,12179,8269,12667,12230,10314,8281,10152,1100,1429,12446,1751,10757,12245,9512,1725,10049,8739,11994,9262,13474,8218,14106,9016,13602,296,1839,11936,13202,12504,12091,14136,536,9750,655,9088,9084}	\\xb8d4aaaef069ef02eca359b97bd32a8953563cda2f4c92ee93bdf0a7f34b1f50c91548e26fd7a8feaf4fb3069b501a51173f3eec34d88c2438e370f82c2145e642b3a039eec352b800b53ee91891933dc85c7275a97271303ef541811dda513d	14	3	\\x4f483c496a117c227f63cdb5785ddde6ca5d73831e281445627e9b82a3f62b2b	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
15	2	\\xff6e9c9816a1e44f164b97e94308f1c8	{10635,11037,13375,13645,8618,531,11149,9092,8820,13611,8322,13682,10228,11901,812,8313,14090,10917,9031,13258,11321,1581,1149,10607,13300,10464,111,10920,12522,13290,13782,9186,10782,12030,12221,12497,10916,9813,11896,13636,13924,1564,101,10972,11235,2003,10694,1112,1984,12370,12416,9474,159,12457,1203,13183,9716,11582,11527,12971,8454,13259,11389,1830}	\\xa3ccd8b5311177a15a484f5d83be1734876c3b2662bf5c3d564f7bc3a1fb099f5b413dbcc56ed7bce205d70b66474e4201206282a11c707ed329d6b5452a15116e30432eedbf7b04c5f8c5516e9f7aabdf1de67821e656ac1b7703284ea482bc	14	2	\\x4f483c496a117c227f63cdb5785ddde6ca5d73831e281445627e9b82a3f62b2b	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
15	3	\\x584d88a8087ba54e2ab985da76572d92	{10657,1980,9702,12978,13033,12740,8841,1280,10456,9433,1642,12591,14055,10955,9891,10143,12285,12385,12761,915,10891,229,13920,10441,11840,14183,12001,1082,11819,13002,813,9495,10978,13881,11429,43,10522,1931,11898,9956,12423,11640,53,12250,10892,2007,247,672,13795,10144,8483,8756,11159,12323,8518,8368,8597,10664,10258,13263}	\\x8404f91d9492cf6c52387ef81df199f80e12b12922d96e075fd9ea85355a14b219e0077826b78da5b3d30e2b1a4ed16117b2081bc1e1384f44956197bf1c3a7acca7e10695f13433d1a86fd5d1ba382e13fd74de6de7988676f13e754595c5c1	14	1	\\x4f483c496a117c227f63cdb5785ddde6ca5d73831e281445627e9b82a3f62b2b	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
16	0	\\xb14392e04df1cdb909fc3bc7b9dfb8c7	{9629,8858,8878,12326,132,8536,9437,12936,11167,13906,12580,14229,1768,9815,1134,10885,10721,13017,341,12581,8761,8295,13485,10444,1854,13889,12384,12860,12137,12533,11859,9462,10270,10567,13857,1232,9842,8496,12171,11861,316,12291,9610,1158,10168,10330,1894,8956,9420,11459,497,10324,9256,9707,11748,12680,1709,13566,8310,9419,13486,10528,12092,9618,13093,12299,12445,13774,1508,12697,11309}	\\x804aa61c640c43fd2d475c458ebe6c4beb57030474542f14dfc69674d1d96480475fdd07605ce60b527f04957bce93bc09ea75942761092d9c61f157fc64e3685147d28486b76b158d3fa461705b73dd6ec253d7d2c1a5be6250777ec5f668b5	15	1	\\x5c2619b66754d717f864168eae96ee46fe4870544bdfe20d936e0628097290b1	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
16	1	\\x997ce9ed701b06ea7073cb23ccf1f79c	{12684,1111,10976,12897,9590,873,153,1197,12972,709,13615,12426,1131,10682,11664,11650,12224,1477,11554,9824,8778,12397,9196,145,12257,8981,12268,11311,1365,174,10547,14236,8948,11298,13579,1489,11754,1703,10805,12714,38,10262,11106,1916,66,1259,1040,13099,11005,10636,12614,12662,11584,12795,8998,9123,12832,10559,1971,759,12907,9276,13455,8283,11345,12478,8212,13112,11797,61}	\\xaea23f9f80024427be4946f48dbc04b26f4a3cffc0b08b5f6274dce7c4f8544763ae5287b820053e265d0b4780501cdb11cc647ec17f94efaf0916a276d049372bc6c1e4f8d261c2048eb28fe774adc9ee433055578c966804ff591f027cefc4	15	2	\\x5c2619b66754d717f864168eae96ee46fe4870544bdfe20d936e0628097290b1	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
16	2	\\x5969a50c65ff7ee2a67b7c2104449917	{10626,13168,10226,13979,11414,1587,10877,12552,12628,13411,11578,11247,1182,11118,9006,12324,13763,13427,8410,10504,9946,12203,10813,12038,12700,11295,2008,10679,12345,121,1993,10940,1802,12368,1714,12806,12668,11844,14043,1787,9056,1985,9456,13547,12113,418,8786,333,9485,1276,263,10929,10785,14085,14225,294,710,12313,9242,10128,1289,10894,13397,10525}	\\x85eedc9b62e4c1a837356912114f892be3f9e8eb4c9b7815888af50d55dbe121d7225055a6b71aeebf7285603cd9f718052c4126deb8f1fbcf3eec65badd396467dce40af2bf6dbe7ca2e280c47c383f38f35f8164fb3201d1b874a32183f3ca	15	0	\\x5c2619b66754d717f864168eae96ee46fe4870544bdfe20d936e0628097290b1	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
16	3	\\x68e203a0d597d4cc8a2bc47be93be24e	{12294,798,69,13096,11387,10907,9008,1771,12409,1314,10490,9732,12618,965,11408,1975,1493,14211,302,13120,277,9014,1926,1743,1262,14286,13016,879,9839,1964,13667,1324,11750,9735,746,13748,13311,11029,13952,9681,13938,1497,731,10790,11224,10818,13140,9005,12013,13622,10396,9763,1843,12528,12734,1245,11324,1393,12716,8899,11425,11084,9325}	\\xad258015a295805052c9a6c9d7b9334a188f62bb53ccc9f082dc17b756a69852c313fb93e287f9c52a57411ecb933e4e006e949f2b237b05944eacbe4590f228ad47af743f73904c504cded459d33837d782fc06e4313557f7a383c77769b169	15	3	\\x5c2619b66754d717f864168eae96ee46fe4870544bdfe20d936e0628097290b1	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
17	0	\\x0af91aa721912068dce6ce08add64dcb	{12535,10401,10245,13609,11543,12119,12398,1617,12125,8687,2035,313,10127,8243,8263,11488,12523,10260,9851,13557,12127,11544,9531,1010,10429,12328,12438,1727,9993,13334,1651,11583,662,13496,1247,10534,12896,10586,8425,112,1075,12298,11282,11486,10572,10376,11003,8215,13514,11392,12251,10666,521,13727,319,11626,1336,10107,12864,8415}	\\x8b71402a82ff1fecccf6df22fe70950d322732bb51136085ff814b0e2a11a14ffe4111a2aad4a1570081f534a1761a1a162bd60a734a666fc6bd48b4b5e691a6bdb3bdc3c94f22d672a1ef742ae2a169e0a0050dcab2c5b1424ac9d4ea5bde44	16	3	\\x6896c2af89443efb024c4583ac6ad4c500ca7b2bf32a9102a92586bc31ca9d67	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
17	1	\\xffbbe0273924c2618af095fe902354f0	{10816,13965,8365,14015,12428,13472,10951,11569,926,10906,761,8223,10654,13835,10677,256,1565,12736,1313,11039,9424,12287,10311,14334,11899,278,857,12903,14209,9052,13604,14170,11000,12541,12244,10986,9403,9499,13914,10097,8262,13777,928,8231,10202,11703,1455,1569,384,10527,10382,9799,8964,12984,9918,610,11589,8796,11186,9363,8375,9817,831}	\\xafb96467c2bd66fa5392d6207eaaa1fbc9dd63b0ed0dd82849fde43a24727681767df47d12f80d53cf114f1ca583a2ca00dcaee25866795d3c55231e9ae9f30eed430d825c51638954e33a31cabf2fff71a0d7c64042a413553c56a6045ab517	16	2	\\x6896c2af89443efb024c4583ac6ad4c500ca7b2bf32a9102a92586bc31ca9d67	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
17	2	\\x44e413af2ec8c1bfbcf537e748dce352	{13214,12378,906,11026,345,742,412,9188,12985,13919,8894,13097,13970,943,1850,11395,12613,9254,11283,11617,11063,9500,11213,10575,8241,12958,12089,12133,10480,551,352,8840,660,11833,11104,9117,13986,11131,1418,10251,9568,12080,331,10405,9583,9061,9140,13412,10945,13685,9303,10505,11050,12512,835,9809,10493,12713,905,14224,10737,9319,10538,13090,12912,14066,13905,9911,264}	\\x98413d38a78e080a4dc87427b09e8c0fbffa4869de2dc5cbc40e940f65b240494fe56347b58ee669045c760cb7f361a904accab1e6dd8077bdfa97bdc01e9a41b7d7eba37eb6c72fab6dbef1f6b6fc0e851cbb846c0d729793feed70ab9bab4a	16	1	\\x6896c2af89443efb024c4583ac6ad4c500ca7b2bf32a9102a92586bc31ca9d67	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
17	3	\\x7e0af84f60f48a54ffbe73642f72e5b4	{10229,8390,9448,12451,598,13154,565,9784,13646,8461,12867,14017,11830,11279,11358,13378,12010,1939,12562,11947,1067,12262,12455,1562,11142,9094,11107,11082,11824,9322,8267,12442,1499,901,10161,11756,14052,1796,14259,708,13322,12577,1722,9067,10218,495,9644,14062,14154,11016,13521,12749,12076,10423,12060,9190,9519,46,11835,8345,13083,8319,14180,14272,9097,290,10760,9904,1808,9161,11774}	\\xa1729a77d1e7aa58fd0b2cb508d55453f8904e2a7c356f33e31e9bac8369b7eb196397d9ea16fba8d64564bc58bd67000c3bd157ff787c721be31e35bd42129e1b3cb689358d59f9516646b926952255cb63ef53afd3601b8d4cfae4475def0c	16	0	\\x6896c2af89443efb024c4583ac6ad4c500ca7b2bf32a9102a92586bc31ca9d67	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
18	0	\\xded5a87bb5583f9d82528942bc3b4d36	{11790,1988,8551,287,8655,427,1176,14130,14188,1343,12129,1713,9490,537,13383,11327,8896,9121,1392,8214,1880,11422,11236,4,9873,12998,8930,12201,9398,12583,10304,13551,10767,12159,1927,10881,9077,13553,8932,12485,885,9351,320,9949,273,8447,12181,9990,10642,12058,13628,10732,12208,1114,1749,12848,9435,1474,11729,10890,13370,11410,13860,10335,946,12382,1384}	\\xacf9e252b4123f6f4dbddd8a257b37b93df0420d0c44a3d8a3d530cee6fc38892d27122cd01f77d31ba03f11b77212aa12adf9186e092315bc97cc027481a0eb10de489d667507c1b400d55554f300d1913e54c19ac904cf1de10f6bc690ad92	17	1	\\xb2ae877ef2b67095d8a01249da2183b3086366f51a03e69fcaa6ce07b6dd9544	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
18	1	\\x625271baaa1b6e650667d396763960bf	{10293,899,12290,9302,13240,13381,11151,13855,10564,789,449,13089,13694,1356,1925,13382,11737,12674,1060,135,12982,360,8236,11265,13722,8671,11905,9447,11538,8783,13145,11895,13791,9645,13207,8389,9209,13328,543,615,12692,9310,13008,9761,12023,10077,10958,2047,8982,9736,8533,9041,13286,12604,8427,13732,12425,13373,11864,12044,10363,11838,12408,371,12210,10346}	\\x9938beee9e0bdac7d4e31a2eb34327354974dd1a381e1736c76d28b97f552779a602b33b9d702479ec178d775c01aee50ee9b9aeb90d96c84c0b6ff91fbd182318ab87c5bf2969d063d554cb4c2e819459eb87b85040a62b186b018155e2b523	17	0	\\xb2ae877ef2b67095d8a01249da2183b3086366f51a03e69fcaa6ce07b6dd9544	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
18	2	\\xc70d1a78afc9431f7fc6d45c3901b92c	{10597,9837,9670,8347,8294,10874,12356,12764,13497,11053,11551,11897,939,9592,12239,10264,10847,8252,9982,11733,1092,13087,1863,11498,12805,10015,1896,1858,13256,9578,9785,9980,13132,1715,1635,10961,12888,10741,13684,8442,13324,2013,9767,8792,456,9510,14257,1308,12794,9450,9967,13863,673,11570,11884,11807,12756,13123,1699,199,13046,1029,12913,8973,14323}	\\x8b2f466a898fefb9b0035ce514ca88ab0c43064cdea62717b0ac10d4f44f2d78b653c013cd75e6e0910fef1acf2b25a10af4254d60ee0868c2d7b65f26858056630b5ee72c8fe5b5d11eb9785318a64b8fb04add1dc911934d2714de0b0804e7	17	3	\\xb2ae877ef2b67095d8a01249da2183b3086366f51a03e69fcaa6ce07b6dd9544	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
18	3	\\xa8ff6197002bd82cc091e29b9dea41a7	{1215,13453,9370,14267,11809,10206,10662,12866,1658,10981,870,1555,1990,12492,12808,11594,665,1659,13044,9856,452,10783,9282,9479,11200,640,8256,443,11266,10090,1579,351,12682,12833,1895,8307,8938,12143,13996,12606,10313,12209,10179,11623,485,8676,11315,705,8846,9227,13548,545,8720,722,10332,9156,8492,10590,252,12302,1845}	\\xacd624cc36b2f3cdc18b2ceb72cf8ef4daa37c01eaa55bb5aa3f453969cb7960c22112f321b2d737f3e9675f72a0279b0ad53f5a045f18c4b88dbf3aa7b8f4d9fbeabf91825fbb70bb984af25b710f045ce5c8f796e0558c2c12543bc240a0c5	17	2	\\xb2ae877ef2b67095d8a01249da2183b3086366f51a03e69fcaa6ce07b6dd9544	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	0	\\x88019c3c393a37823421428662ea0a13	{13216,9745,11087,12727,12852,12031,1584,13666,9377,12151,11622,11726,10865,957,918,13510,12252,13874,14144,9219,10888,9986,13376,1271,627,9175,9122,11330,8952,1306,515,1349,11997,11962,12275,2021,10355,9623,1660,8842,9905,756,1846,677,13577,9790,12811,11075,51}	\\xb3c72645b809b1f51f61e0f8197b62f2565bf904079f9ca70056a603f7899b38c7dc91a97f9bd07d331f9a959448ae430d3a71287be0efeacd5fded2f353be4d4b1e93aac6293d939a9076113a50af4bd8069b3eae6b2f46db9c6587c6a6e8fe	19	1	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	1	\\xd401f387c50e21900863aa84289fe217	{1150,12723,8555,8359,9466,941,10274,13357,12921,10340,12828,10716,11784,549,13205,955,12136,13061,11534,10811,11398,1957,8320,12182,10759,13593,10618,11734,1159,8210,1654,11123,8605,1552,14278,8227,13481,12249,1440,12598,9730,1155,11563,14277,11388,1086,13498,10016,8380,13299,1772,9446,22}	\\x86432b9b88c6becbed3d92e824d4bcdc88e9ec88a6069dc3272919a22877ee0d7fc5e8a99d4d9595bac4df6dc42d358b0b77cc95258d7d9ef9b880a795b1e742eb1719e79ee02a9e0f11316cd824c0f2ab107e43213ed3b2706b97dbe9928f59	19	2	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	2	\\x8b77d208210b1dccd0ac947ef1f2f740	{12297,10400,11040,10589,11814,1972,12779,10165,692,13640,1421,1434,1443,10443,8366,13937,11418,11757,1363,11963,614,11915,12755,691,1507,172,9350,8418,8503,162,13128,13418,8759,10883,1865,1801,9388,14296,12238,12615,8524,14159,10204,8817,12487,10170,10832,12796,12884,12792,11711,13495,8250,949,9779,8328,10451,13943,10740,845,10640,14222}	\\x90a9f3d77b63641acf989bb5c2ac6cfeea524bc6d5ed49d8bcdd7ff3706e2d3c5816a8ebd686722098cb068ee575011d0323651a7c20708ec9ba52e594320deb31a6a1be95eb42f688b743935746de1cb0c14620cad191f60fd50244c2543805	19	3	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	3	\\xc37d8021e0c960b5b5ce387eea43d1cb	{14212,10294,1450,13706,1540,13626,9693,1729,1599,10660,576,11716,13137,13026,326,8750,1382,13086,11780,10777,30,286,9887,12452,13203,9560,1339,8816,2040,12029,8253,14031,14026,357,13721,969,12776,11766,12947,9362,1380,14134,11530,12348,9945,9247,10498,9253,8387,959,12825,10186,1467,14235,8238,10795,14333,10440,12544,12377,12532,12074,13442}	\\xb37e93d73c24de13a16bc5fad4cd58b162a4c54a1d86209674e6a5c489a9430b8d7ef49cb61aa0bd5db91b804334173a11943bc105e048043c7b28d646c6ad9f175af92c4d69f5582ad27bd765cf0963a0a4be5f77b5953569ef7b85a0750bc8	19	0	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	4	\\xabe31eb5814cf1a7bffde6ee09279e7c	{13505,1448,10868,11140,1823,10291,10216,9712,14153,1337,14127,8321,190,801,13312,12443,9468,12732,11630,11792,1224,9804,11751,12399,527,14308,11109,98,1557,12152,8711,12587,12073,9489,820,11322,14039,13114,954,620,13878,14101,1685,13405,1738,8424,425,8242,12430,13248,795,608,13139,9639,493,10328,10393,13596,9079,13284,222,1175,704,10460,13331,11858,8576,10736,10764,473,9900,14126,1989,8641,981}	\\xaedd52ffc4cea463eb72953fa0b219f51db2c72089b4121f60b0334ad69ccb82a9d1dc176c384a2a439d663bd3e385aa0b8995024647d335f21a57122fd9dd155c871b15be85faa85a78d4e65a98edccca9bd1c969a6aab696f28610f510c0fc	18	0	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	5	\\xa4746ce9a6595ebe865c0978c33bf4c5	{1417,325,13601,13136,9516,12611,9367,9630,11372,1047,11432,974,14298,13471,10383,11209,741,12601,8371,605,13451,1244,12602,8969,12673,11246,12880,9187,13892,10523,10581,10987,12,11129,12703,11495,11735,1234,9314,1285,11811,10224,1282,1472,9957,12308,1908,12394,14131,13318,9620,10281,1885,11801,9983,10470,11308,13335,8738,851,11110,11889,8407,11150,9037,11435}	\\x8998c785b6f9138bff3d05bfb21414c96b3616ce67ea721c0b17361ac5bad5558820b2e5b86a75ac646c45db8c3e98690cf71d2b34ea048b386a447102b12fc623c6200f63b01e4b0b91b23f3540b278e81bbb6fcba15735eea1261e0f655565	18	1	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	6	\\x83043a4487fa7f30813754e9a8a0a8d3	{1580,8910,9002,13197,11817,1705,12180,11894,11552,8813,13368,1554,12838,423,13341,8897,12331,575,11615,922,8516,11085,10489,8884,1279,10594,14219,10649,1374,1141,9233,11444,1209,13960,143,12035,8994,8440,11648,14009,13957,1161,9729,11049,10999,719,8713,8217,9296,481,1341,10956,8358,9624,2005,924,9663}	\\x86150bf02fa08c11a9b01c83e3a23ac89877ac19e6e6148b12bdba7a9e117686b8a03da4f515448f21017a8958e41bd40d01e2d6d5f1d4408276117260a1aab442a95f5ed1a3718f3f693e6d33b49128c6f450b46c4574a19837c1d75a4d7497	18	3	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
20	7	\\xf62e345bc429909f68147ef610002680	{8504,13409,9908,8675,479,9267,8429,13374,14226,13452,1891,9939,12540,14287,12454,696,1636,11272,14243,11796,11504,10672,13493,10693,524,12279,8960,13818,8957,173,14200,9313,11404,1643,10298,733,566,1115,843,11969,11483,8274,9138,1583,1732,1231,257,8541,11605,617,10562,175,12979,460}	\\x8d1c69f04ef2f1b8b79903e6e17352ad14663beb69d927f549e20c5bf03b314b0f178f43f95d5cf313f13dc432ee596f0966fcfffaf93317ebeda82efd3226070d59fd7dd4abc9f024629b395e00824a3ba9e1953acbad441247c32fc53409bb	18	2	\\x8e5ee6639b683649e3b5055ff8468474c9c781453d7b06bcaa0e79821d34824f	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
22	0	\\x009ba1cf255d65482c6088ef8323ae3e	{8984,1765,1769,11545,1867,12229,391,12729,12646,13582,9553,12789,459,13316,13006,9317,9894,8889,13724,10277,13680,14251,14240,80,1123,14067,11290,10084,12937,9920,13261,8818,9263,12099,11778,11275,9245,8526,11515,11099,13844,13344,9386,10465,1196,12965,647,11331,10123,420,13109,8340,10681,11596,13911,12155,9677,9802}	\\x8be2d96ce94e7205dac8afe77802b0f8245e411803c18aaa289cccac0219c2f60973b29472fadedd045804f6c675f7141761d0b90f1a79713ba229462ee9932d58f589ace312c08136c076b64e6db897e22861b98841ed21e1356b1696bfcc01	21	3	\\x2713ca87ce6a7f8906269f72a60860cd48b6b4f2220f801c499cac3edf47144a	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
22	1	\\xd7cc8dc8c1bd101ecbde25bfec771097	{13030,9151,9438,12893,9782,11377,10319,8660,10487,8344,13388,8512,12856,996,8945,11162,13482,12791,11196,1979,8317,10599,8572,8819,11368,8698,11065,1143,1595,11789,10341,10512,13254,13811,13149,11079,14324,9030,11679,11860,11693,9601,13913,10163,13204,10042,11968,1022,9757,8476,13350,14012,1816,12117,13041,12483,8812,10827,13792,12141,9157,9791,12200,8690,1831,706,12586,14292,10201}	\\xaf1cba57bef82444d13effb36ef9659a884bf3d4eec74fb93bdfdd5f51e559906c95ee6c3e518a2c523c7e217f5be24b0f8da337ca9c1fe164897dd4c3f779ba7117f220075670b0efa1934af7e401f2cee01409c00d0c83fa09416e71f11c97	21	0	\\x2713ca87ce6a7f8906269f72a60860cd48b6b4f2220f801c499cac3edf47144a	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
22	2	\\xbc2b8a0bd2ed0a1b4f5b78af1cb84601	{13886,10863,14304,12325,1899,1283,811,10176,12470,228,10742,8194,10747,13292,13025,14147,12411,1829,10953,13746,783,11690,14273,10141,10064,11083,8431,1840,11369,836,11411,148,558,762,13319,10479,14083,1679,1887,13415,11194,1998,13160,9754,13065,8609,2038,10974,337,14331,13805,1674,12762,10880,10407,11880,13326,13103,11714,8645,694,1462}	\\x95de122ea54ea3aaaca1b95d572aef00264822cdba5061144d9dfd90c0d8fa4d2cdfaac2c8ade02704310a04734039050be2a39bb842d0b679161076b9673c0bb503e94fb1e801bc106305a522d354b61bdcacd13b4285cef601a8fa1d11af2e	21	1	\\x2713ca87ce6a7f8906269f72a60860cd48b6b4f2220f801c499cac3edf47144a	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
22	3	\\x2ba03884e0aab923155c7ddbf6375f9b	{486,1023,12344,13320,11096,84,1849,330,13342,10569,9959,9232,10139,12204,12266,189,9733,2045,10193,416,8632,11147,267,993,14041,8406,8905,11473,340,13556,14036,625,1017,1370,10602,13117,1827,11746,14327,8612,10806,11025,1438,14023,11480,1027,11152,12637,12654,9020,13366,1179,8237,455,9698,10625,13063,653,8707,14206,13659,8459,8774,1361,848,10685,1412}	\\x89a23d659ec5f6b70aa6be4d5bde415e5af4e8bbad88d2d830d62080a81820144bf0cc533e2b2ec9666b07ce62323707154126c4541af5e928a49d5acc00bda71e9197606584431a3844a47af595eb761b01991e076d97bdc5cb88ddfcbd0729	21	2	\\x2713ca87ce6a7f8906269f72a60860cd48b6b4f2220f801c499cac3edf47144a	0	\\x0000000000000000000000000000000000000000000000000000000000000000	0	\\xa7ad6516d0f223d3e9035c6434aa5ab5496f631c42401f812e428004e14533d3
\.
COPY public.blocks_attesterslashings FROM stdin;
\.
COPY public.blocks_deposits FROM stdin;
0	0	\N	\\xb1391379360cb6fa63c4d4d90b116a2104640c69887193b985beee07546816a6a02074b1ffaca63c9853d3f2f85c98dd	\\x009a716695c1c6b19c75db730aa3cb92b64fae5ca257a81a918696bdf5277570	32000000000	\\xa78d5d520619102ebe808e38e2d0de3f6d7a616b1ae2b2c54d4d7117b183cb2172cba3db620f4049c8b0326b923ab6090cf2307959cc9b82d74c5140910c64d456d2adfc92cd60d44ca4a6a2626e1b9e81095091dd08129071cab274e2396c85
0	1	\N	\\xb1fbbc356aa240f7d4a825e0e51eff0bdc700de3929df296b6e49a1d85130b62bcfa00a6912acad2f8ea55a0c2b00c75	\\x00b7e59f3808089f25eba12a6fa66a5c0fde9ff73fa5805e644b83fa272d79ac	32000000000	\\x81e166d4fa1dcc730a8bc7ed38189cd092bf0e09ddb751b5db61423e146197404931743fadd4f2e4a396c90cc2356ac2079e2994623030b10634eb52f4b6dc5201d5200053322bab53b1210f92718189e339fe0c6dc94abf2725ef6611436137
0	2	\N	\\x8bf10faad775298f089941b54072081933430077efa7edb8800a5ad67b910d1d4906a59aa09a2a82c4bbb2eb572ec41f	\\x006525c13a1be61a018ad639e181648d58a1647c16cb7eccf77213e0f8251797	32000000000	\\xb7a32454c6363c9c4f54ef80643ab0840769d6b5b4da62a79a3d146f484adcebeeda802620db45849fafd1810964607610f336e229b8d4d383df52e00f423968b207d9e0e388248403f05891b4ade9c1f4736c77d91617975ef6ec06fd691dc4
0	3	\N	\\xa4e8f56301594c4ef7544eac805c2c701e62124f6bcc2ce39ae242fb34a01795c9f96f3d5667773b0329dfb9b5ec3e20	\\x0087b8efff1e5494b3f4a85ea9b82385671b4b2855457d5ea656f863da457f56	32000000000	\\x8ad6469eec0bc881369c28ea214a601407953fd5b0343e1f033524506d12ead7153ad0328f31bfe81af96f8f2602d22012a2d854cd848131c6342bd88a3c9bfda9adc08d09774b53bf5f7715224d5c40cc25a5aaef7386765fd6cd1fb4628c5e
0	4	\N	\\xb54268644802909635c3ac0630394c8608a705d335277a3a0e130b86d23300312afd7e5d52e4e44aaa67f2810bb1ea60	\\x004b211838ed05a579a285aed07ee9b5b353d6c3d67fd0aa29292dd4266346a7	32000000000	\\x89d00cb3e91ff90f863c16e9ee1fbcf3cf0d41c445916b8c770bacfed39bc33119ab82767a0f755d4508262a8333a7c01333ad2489e58ff0d3e44a023e5165bba4ea52d8e861161a29aa67c82cc532e3c09618a78466d5ffcb43c59997c24dff
0	5	\N	\\x887e202bf6cbd63da61fd95b3b42406d47f91bc159a70e6697701afdcd64193e5ebf91a526b027e004fad8c600cd818a	\\x00a54f28489d114c335d9208b9955e6c4b399287f23946a1e692580e2bcea72d	32000000000	\\xb803908c6e859657f6e809139443a5c45ebdef61f782968abde749d2fd0d0f13cc63904595910702d48a1e6801afc4d50a5880f9d273d02483b563e87a85bddd7644bc4df6cf125aaffc4d3742d2c152dec2f1d8a258968ff226b44f0d26f594
0	6	\N	\\xb3cf4e3f286c3d07406154a509e12065adfe0b2decc00fc55de4adc77900feabe24e2ca095221834dfb9ef768c629d33	\\x00de417ba822a4d19ff89cceabbcfa5b88450019e1e16d37937fd32aa219b82d	32000000000	\\x991dadcdff493466045a4f8aeb12ab04e09a9c5da9ac5e61e0fa859c597e6351754cd4f7f2413381ff9d80d4115e1f6615f3781f032a6feb477d855e9701914021ca53eab93e1bb8354780d58cd4bc393fd7170353ed6985d642aef61b122d5b
0	7	\N	\\xa32a6208028023dc022a896421cfae19d901aa49943b099499b252957819048525184085e3b650b5315804940b44598b	\\x009192f1645b0cb7e047b33b5ed1589d480c6d5680c3ea74eaae0c1650ae9654	32000000000	\\xa4fba4ab172cbca1678c3824212324bab4c8e4e89cfff848b844d4fca37861542121aca1a9c71dc676ca23ae7e48bc6f12a88e2c4f0b25cd0878eb7b7ab86417f61c50f921129a4e8848935c496afb429cdeac82cb4d1078859fe350375efef0
0	8	\N	\\x83562589ade60f551f6b1f19ead319bc6031dee66f46e173e05f1cb06d85e7e0b1960e8b263a7b96d15916e9806446f6	\\x00990af45751941807ac88896701b0fe33f62e87b02f8f4729d7b6e71689c5ef	32000000000	\\xb2a7e35a4fcad6eae6246650665013fe8e8a6d96adab574de3404838ad31d4dafa925e768250d38f52b109b93e41d26f19eb7383db8525d592f632544d832ec3002d8f3929f54b3dd18ade98b2f48a82b08accd7160e8c64e71530b083978d2f
0	9	\N	\\x8bb4551148eba7433c302ea318e7c18ea44a1b8ebf9555447d03e0c80e77d3b745ae8f517eb954ca0dc52d46fc6296e0	\\x008763b26fd367d4ed8edc1bb58aa2838ff1045a1af82f851da9dc6e34b49f5c	32000000000	\\xb204c7850c57bc5191f683cf9bd8f4d9c011183136e02df8e690955a12c7f7c1782b98b042b9fcf4a5d7c6448daae23006c2974c5ab7bcbea7e6ba59752e6aa83bee563c3286ddf530fd6702bf3c36433655b79cfa91e30f54abc471dbe63c3a
0	10	\N	\\xa9e7e48d0c3539f260c01cb93b6a2e9c7ec94d2b5a6134d4b09561413043a8e686e1fa2e545a981146e4686d99b8df3d	\\x00cc3fc841cd79655f08fba3c8a6524b0fa945447874a76b6de348349b9ab390	32000000000	\\xaab3d67003bf7da0a73c13ec9c86772157f851a1de36bd18ca28ac51e381bc8db434e190b3202b1444fbc8b23f5dd9170d80007d67a5ecf891a4f10938339a3890d1f9577586724aee10cdc581041ebb15e0cee236169be26081a07d44511449
0	11	\N	\\x86accb54629ceec44a4c06436ce639fb5d0ef4503abc0d3958a5007c6f3c4b2eb2b4f8155cdce8452f2f01d537c41328	\\x003e97a4e3d984d983c5023aba79e2abfb765fe58ab59cd55f85aa2ddb0f4b92	32000000000	\\x967442d3ee08b01bb6d62b3a72a3986ec67807336de3553adfddf9fcfff50f93ed6dbbe25a637afa0fccc35737e775dd1454efc47ea5b54ec326d84ea226599a4710a503b4f6758339176be01527a234fc2411a26291d6dcd69e78092e053f6a
0	12	\N	\\x9946c1c8da6d97b9387262f43762befdf0ba714c4b3884051411a0e7f18e5c5dbefb99099c737fe9f54f124d02ed1707	\\x0022d4d5f14bbf6cb0fa60f58e9befe6444de903821af0dd7b7e996b2da1e77e	32000000000	\\xab93f453b8ffdb8356193671d01095955498028d2c3c5ddf49911cd91f06ab127bc35c604d21f17a3227f06a8a6bdb0d1476f27e8dd7072aa18e810469d47a0073318be8c2bdd0072ed61a272968bcd67e640aa0d8ac96fc7af05d6de83fc0c0
0	13	\N	\\x977c217ebe82ed92c00aa4f6f11adab18c6c585d1199612409ce8f24bf5cf665161dc5b3c991835f4d2d59c6a14ae218	\\x0077b4ecac02b3dc1d752dcb66c34ee675d719d29d3770ede0d7ecdc92b588ee	32000000000	\\x99aa4668e4916511bc8cb1cea8794976c4695342b9ad0d963d726280d21ba3ba42ec4fd0103fbbcaf9466aaa56e1eabe17aa75d747bfef558941f7fb8b524c06365eecbbe77d5d93022216b15679b5b5ddcec90fb026c83b09a0e90dc17ea332
0	14	\N	\\xabfb6896a5bf3105adb4e03ca49c487cd75f0cf7570c917a396fa8211e933fd9297fb1def46020a03e21c1956262563d	\\x00beb39f317480db5397db86a708d4606cbdc1d03767a5b28dd5ddfeeed7ff94	32000000000	\\x83f56575ca5d93bd28ef1b359179239acaaeb915126348a889b08419913ed3012c61b19d000fe6338dd21b342ca0481e12dad7b30e5d275f65999ff41033923dfd3de9e801a4d12b617d1263bf21a0ae939400c9742f1bafb6aebea6f37f52f1
0	15	\N	\\xb18943c3b4a71068a436e877bff53a8cd2de552708f9e8ebe81e1ebca8ed19f5f577044884edf4f700949b480e622cfe	\\x0012b57157e16f8fbbe5e18dae4a8bc7dac61a73998bdc1f6a51c9cd534bc8e2	32000000000	\\xa6045f581d89bd83b7abf491856db89e7d733fd2764f04b14c4224e7ba93ef348d9e33109f6a4c6ed14442b181a8eb9306ae706326f2ed92bd1b0e350b829b51bef56d87d0f9236adfe64c0c067dbfd5588b9b56e386ddad4b44fedebe26a5ef
0	16	\N	\\x8b4ff884f0825d8368f4077ad53149759332554cbe3655c34518a30e75365857ae787b208301be502249bf4e06e03273	\\x00a78a401ca1d5c49c62aa99ee38e2052527a090723474492aa47f2361bd88e6	32000000000	\\xac55f66329dd396905c64ad4c7b73a58c341f3fde4b7859670fa7ed77cd6b291baad139c9c54f6079ec73885a6bc7d9f1625bc0e3aa6fb06eda426fb4a94634f178422dc53ded7bf00d65e0d7ea3461b272444a9df14a4eda554ec36e56734b9
0	17	\N	\\xb528a635883d81c4cd4b0af451765471114df70eae4c91d974a8b0dae9c470656fcdea4135b3c9013b601cd6463ba16f	\\x0086f4ae865356daf55112079b84ba696d1cdacaa4e47237a8f780a1a4d9f4d9	32000000000	\\xa28abcae9291723b7a83c14cf96e3983b2c93b2969840db8fe1596f5bfbce9112a0cb1c857944f29c90c315d536675c8095fdf5c436ce4639f484a9a95154525e8acab8c7f038e0eaf5753bd90af4c15ab53da1dae8e076dc54f8818cee44c04
0	18	\N	\\xae94ec1191683f51784b6e6201edac5a6d1365e97271cb6ab048d722c770de76eae0f901d83ea4ef1e1ed84cd3c80aa1	\\x0047d061828d0a5139c5e1aad92b58b7b0bf8e836f16c36322e95c7dc90a4d40	32000000000	\\x884637e39ed722476bd39a37a16e462144a14f4dbbf2a38cd2425817cede480e3508ed436b1aebc8610bb4b1c395793e09e45156b9db8758c4f13b16324b521b3896a51e3c7fbac7828265ffbe0cc854ffeb214fb9165cf9e198d276e1fdc916
0	19	\N	\\xb05b4fe1b803df50e7f275f399b629c77b98037492d35038121c48a258df5a8dec9532836a37e481d28623b7b5ba57a9	\\x00bdf5b24df66925ffb6f5b5ced49d3e558d0ff096488dc4bab45c647f511ad9	32000000000	\\xa3710b802a5acc07abe5441f1ae1d2b30f4350a8d79fe634e2eee79ed3b3e2ef45622516f179c0c90774e6b12a66b8ff030de95507ecd047255fa8ebc931110d46a3530cf4b5d3b7b6bb048dd79c224ccdb0a398731edcecf68d9632cfdc35d3
0	20	\N	\\x98bf2d673a152213fff14f66827978aaa04a5a1f0fcbd95fe6fd521af08e470ff74d1b6667ed1a607dbb6b597ae7c573	\\x00d2d24d652c1ded1c9453b663ea35d2b1964b5581fb78a1fd745c49659f9276	32000000000	\\xa2de94027b2af9b2197fc112eeaa41bf3620bde0cd7306ccc42d543c3b133f8c79de66f328fa28f2a26b3c6c91d7517e15de894e64fbc5c1c44c0cc29e5b9cefdc9862b6621465f76dd1a68fbd329006db308f2f642357c7c702e676683c1f14
0	21	\N	\\x90f0e853deef7f089213ee5083d716145a6390a7c87d635bbee808af93f75e87367c4372e01bf9643c1ce63b393420c9	\\x00d35a6cb2e97151a50ad1fa054907e6869cb6b11429fcc7fb406ffb8eed82f9	32000000000	\\xa0d2ef56668674d707ab2cb865feb66e63fa613e0fa8189819b7cacb79085da711b992d54d00decf93bc6c620bb8cc691936ffcbf86440868fe9e2afc19f10709b3e02eedc362f217be147bc67b5337b3ecdda162a1ce1020067411794c555ee
0	22	\N	\\x84861ddaa7690c07e1c97f5fcf5e2c681b1fcb34f052273f959de26e145d27a10ae77e33cfcc3cbe73b9920b27655e4d	\\x00cde83b04f0b875474732fbba2cffc91c68bdd2deaed4891c157a3ad14b69c9	32000000000	\\x8829b5bfe0aeda6e0c904af4cc4cf20d5b014d8e0afdda24e35bfe14bdab1f1ccfb2b90f6dbac97dbf82bb97695b52080df9668ecbc4748b0d17ade44dbfef8ff955ac9cfd6095a706bbe497f79bc64fab3dd54a076d9208a97ff9c5b3915cf6
0	23	\N	\\xaec6abe9f5b90e7e27f1e1c6f497581968339f33eb3e70b2a561ac7b93259502dbc7d372e34f5f96f7e1087f9a999ad0	\\x00d65aaac0e78311e6cf698b04b120a0a240fb55cc2d68455faf193e3db9ba33	32000000000	\\x89a025f41804430c4c9c98c456f435ad5059834edbdedbd65881e563cbbac999e0abe7ff96b46b4568feabda13d74b9d197155b3e1679c1287e06490850ce68c0d1ab93d836410a64f183b878966656cc0dcd610698bc6c02e0168b44124523a
0	24	\N	\\xb0726c1fd5187cf9452384c0cf1a5bb5275539f42a1234770f640b8e1e5a65a42e12624e46009d9562b125ffb43c9d51	\\x007a5842359d55625d38f6d7722240179ec8b183b92f079d36b2764b25cce3ee	32000000000	\\xa3cd49837c5bb65c55d2827a5eddd4039455e253623930c94eda6707c48d3703a23a71bcdddfb70f0fdac35dfb404a7b0feb7e525865e17a3f64e78ae71b178cfd08320cc37b3ef2760b9c99d24beeee47b42ffca9c9b4dc209e9fb36d65ab4c
0	25	\N	\\xa239d39a222ccdcd839bddd9e9f1cb5a4d21c573961e113124f061f892d5ae373154a7a2f75b89a16a6a8377b7340577	\\x004ecf641de477642dc37470a6ce15e638cb8611de1d25f4b26341077ab81bda	32000000000	\\xa11e9d78bbc884778a32a42af3bb40d827adb968f55b5e2944db57d8d056eddb560ac6c1f2c2ad6e63f4a7a6a97eb3cc02b3cf19402e9fb5e2c68738576374f3289f977ac21cd54f1354d71303192c46ed421c2fa4e6ad478bd04868d83a2678
0	26	\N	\\x987f7e4af595a34439eb2f87721e8efdb60cbed30b3908960b37112aab6aad5aacce9f6be4df4646f8592f1e8e573924	\\x00d1d394bab365274d7c05994c9234cf7e707d3447b258ec009fd009bd9fc4b2	32000000000	\\x920d5f89967dd051e404037fc18379990636d60425746673ae9809e6177f05c45e70e8bd34a380ea47767fe9a44af531016a1190c669593800ad6ab6c46f840f02d461a6a172a74cbcca2075f0b68b0bbc7d20a3b5f707c85709707e4414de19
0	27	\N	\\xa4e6b4771ab3fa4e7c6730744957f5c437722e214d888f441620caaf750e9acd3c0ac65c158a9f2dd9e09e89a79c25af	\\x0049c3a5dd7dae42e82f6793aba7002fefe8ea4cd788af16ec200ac499a8dce1	32000000000	\\x86a38ede992c96a2d51d0ca6b0e783366562f78207cd7ff8288122a464eb78c48859371a518743e8e1510d85bf4ac1e71378854f3b6c4409120b71e637570c76829b39a855a63329a751ec8fa9116faeea4de4c371a4259323418288bee5bce2
0	28	\N	\\x90aa1cffa2198638bf6a8ada0e8377e855ba225edab0cfe355bc1348e1cc1a4ccb0af4548abb52122d94f364bb7a84c3	\\x005e3166ab4573556e13b12af19c0e3fb2326ede5f94a4b4ec318d4f68956a13	32000000000	\\x9579eeaff4067bc86130b5d3f3f67cb15e27f75200acb7a03fda445e11ebd4cc057733038c4c7de270973557a1553926124dd6d73fb62ad0e284edc961acbebb0d3a468861f7c1724d7fc1c78311006753b160dcb67f29cbe685581f38b6d03a
0	29	\N	\\xb740bda6f2a1103de870bed80c73f0f4f02f6950ec8b57822ce593acbea510992d4e65ee0d20f472a4daf91f56965dc5	\\x007ed1c9c6b03b06c8b01f76b5aacaa95ce3c50ff7b681054544be5c45e30c0f	32000000000	\\x838ca52dc7f49b59e69ced839da2af2a41bd1366b7b3bd4117eeaf7348a6173362f577b2e60d9920df78f11d1c1cfa4a0cedcb10e51af9c017bfee742d0cf6e201ef80067a0a734b5ccf909f8be741d16a30f333a84b1b935946eb47dcfd19e9
0	30	\N	\\x94a1111c0eab1bb5666eacb9742efbef79042ca3ba384ecece98ca2027c2452a7dc0467c16e45ad9f7bc7c89b6ca8737	\\x00ecd9abb1a74651a9d23426c3c5619d7fd8c119c636a2120b8d1c12745f72cd	32000000000	\\xa5adf6c12dfb27965ebaf98f225885c315e21d06e5e7bced093d6e02c1f20d983f7717caba3f744e8db0f6cbf00ae0f2054a1655271e413bb9a701ea3a6ce61a915624c62460a6296be4a26a788e20abc986529181b0853848b69692fb63c025
0	31	\N	\\xa7068aa763b441ffaec30fa6c051046ef7f5c3ecb3a7d86eeac2a34f92a5e9e74a049184358c900f82b63dbd56bb5ba5	\\x00b9ba4ff15dd8b0d0c57781724bc73ac19be0064dfe5720921b6cea20cf1e2b	32000000000	\\x91757e21810ee83cb63255afe60f98737f747a5abc796ba60ca4624c53c7e69b5321978eb1e83b735081d2a8f3702b0b0f4587b8c8a36e1923a5483dee54f86e3dbd9990da4f172cfe2a5715a0b5aaee6750469ba5307c2dd5b411b6486a2445
0	32	\N	\\xaddfc94c84125c2547cea7c83af983ac1e10738aa1bb6cff10ba0cca43e0e4b5f38c02803383a39fa21e25e619f61061	\\x0076600df9c724e0204736cf71d7242412f8d27f55b0587eaf3ebb853b7fc4bb	32000000000	\\x8a8829f7309a499ba3abb0b1f402bcee02a273d7fbc46243cd4ef1179a88df448ab255a0df777547e6c5e15094f66ae004de58b845b919da8208abecc5c2db52c0a5f13648ce2259574aeeff00d847c792da73669c3388fdad85a28652c02283
0	33	\N	\\xa8b2f27e1005db3b973b61d9bbccd7cea9fe0a299d804cba0014ee103a4a658f1c5ede680b4bf5ff01beb4055e1a7b2d	\\x00811b6f0aed81bb67fd104358a8b1364fd4524c960352872d1b20f079d4cfce	32000000000	\\x8549775b73dd7dd4a8ccb6ae33b404c482252c6fb90be2e635b22fc900f924d856631b7a4b3234251ae4377574a3c9bc105d6c0da8f3cb68e89e2cd2abc2f6549184f3b3cc1a1772b6648dfd6def7f43b08e16b96522c7c05c371e34d8cc934b
0	34	\N	\\xb424af8a10fda0eaec704f3a1b28fc66ea6d504007d3361fba95ec2f9a3e2113ace14770ab9cf5434b8081e588ad2394	\\x0040b77d6daa6dfa5f25a7a06a38a0571f307ebd78737bfe02700e9b4b78ebed	32000000000	\\xa2f6d3e43f886cd798e71d0d65434b20435db3f833b1d3f5b9635b5bff7b99e84f54cf797333d90c484c1cf9af11669907985a1834be923ba88d6750f0a030cb08d6dcad2fdc1419175766708e41214f83af150be4fc611da5428e6a300a824a
0	35	\N	\\xb043fb18bf5394897f5492bcf5760e33d27996b6a343bad70afd83c45c41dacc3edbdad3f7dc88a6c158a847856d7909	\\x00f58e7183f20ae8e2a2013e1b9fade6e9bcbab479ae11d9aec2a596ad382b32	32000000000	\\xb7e35ddd2819a008d176a9b613056a629f11c6328b05058aa50c4a9a14fe78cd2338094c1c323aeb86efd6be0e7d4fee0c0268918ec7427bc3fc38043c8a8e97848727731f7ade4c8a5dc69fa44787609e206b30674c4d77111288f04f1df637
0	36	\N	\\x939e5cb88723ab173d622b11f1197b0c3e1d2c5b0d32a53ec0bf08effd1131ce382239080334c5ccaff5ace2b34c56ff	\\x00955377a9bfb6016cb6b94b1a26d68b14cb3213dbd4718902d804c1a698f896	32000000000	\\x971312cd73d19379a192828a588e7b7058f90b6949d8208c2dce7fbf050c02eccad5ff6c924295249bcf2045f36de9031591cc96dc8a4db9211b41e4295729faefcaeb6247d76bb3ff02c690376836e5ee65d9d0eea01e8423c4f5e781c16d32
0	37	\N	\\xb2a530285fd462d4b3c22981942c147d22cb1cded09116e8dc060ccae29305ce3ded8e19243d475009b6e77ff830e129	\\x00a9db877f6ab66672e4d41449b534122289b1ba406ea409eac3afed9254bfd7	32000000000	\\x8e91b1852d9116d069b8302f1ab9256062e03aa72daa24b57f2938bcf0071d3b2d1d2b5df56236fbdc268a6abb8d3a9918d1edda6e5ede8b7f46a43f53e214b603c1529df831ad0a7e0556c7309971c4d7ac3f4974259608e07cfa926165bcde
0	38	\N	\\x975c9686d05ee1b7b494fb432b1a1589e0001270bf51ecadb6aed8de8aae1f53ed0c8f7f1bf60fc2dde9ea7aed5f28fd	\\x006ba6aae2712bb20e80f6bb4d40bfa129f588a3acd046e513cfc4c7d0098268	32000000000	\\xa2f81a6fac530e4ef8ed202131a5d3886b6b94c985f1c4b01ae7519db516dde675133c12ab844f6465521ee410ef561611fef272e2f31fa3215d37ed956215695637ae3fb371d203cb919494ed6411fa3418d6045785ab56e5963aa40266de09
0	39	\N	\\x80c526e2bc1208f95cad144990c312f8493d223623a7adcf80e19c7943e107b6b3e62bb3acbb3ef3c32cb026b71e728c	\\x00af9004ee5af2e07c23a6cc8490623689b838b1f9aa941c4ef874b38c1dd643	32000000000	\\xb060cb3a63055407c29e3e2618cf11bb577fc1dbd63cb94f84cf3419f526bbab89ea84e0bb95a0328f21717c48b9d913174ed7c4b0d9383cf6a45fd86b585dab5cc6fcde69597876e8b4ed5c06070fdf823db9f3d88d1226814fe23f189c2100
0	40	\N	\\x80371785befe39cc94cad0b5d7f6dacf5628b43754c1a8d26303be5604e522358c75347987f153e7885e8a9952c0f354	\\x00cd9d0954a4a90dcb90834faa9fb108ba401fb0dcc80032b9192fa0b0787753	32000000000	\\x81c21884cd0e9311704a77675987f0bf6583281c364c6797542a8dc924ffb5fa000bdd09c70f0339a93e1ee3345b3b991958eee9c3bd075985c3c95652fa734d5fd87a5dae2c61a443d70dd0599b244c3b1a5cc8598c85260dae2d91549bff86
0	41	\N	\\xb7747a6497f53d9b065673c314c40087ea30364d87cf7bba66d3168d6a373b9c760af02778145f75a421611250ab94d0	\\x00b5e64eccecf4cfa15fc558f619a0226d91ae22eb6c9620403274440f818222	32000000000	\\x8bd4fdb6e74a4199b989f7e33f208322b7c5ff2cddf22755edc1daceadd05b0e1de3c99a8ebd2aefde44b528cd8ac27d09cfcf2b1db062e87330dfe8eca81efd9dcd7b9874078d8399c7afa16b473296108e6ce10ae21fb26dc60f357c533175
0	42	\N	\\x9704a2c3ab29d5992aaf8e644e666176c6d51c0f959829c63526e6cce514836ff4e889d5f83538ac2a06a1338501e5fc	\\x0019edb47a56c7e6baa2596aebbe14b88f0a6be9c30ee766608e844629c630b9	32000000000	\\xa2d9156b57a0515012a98a965002261b6020088e5f6968da6588b362f624d1557643ae6aec3141ff22509d0696e2a85208f0c0e9c382216b666a19e1d796076f174164790e44906ffdde3a9de8d796e5f873700518bfa75544d2d70716aae330
0	43	\N	\\xb5db47c31dc8a52360beed1e0a97f285dc0c9f71dcdf499b8c385eb6689175886ddc36f0a9a2de58b70663b031769dfc	\\x00877d03b97f89c8428f6f8297a7b5172be4a7cbf990a48e1b42eb26bb2e41b6	32000000000	\\x98a010875ab0282d881542c850b8b7169b962fea748fbf823b90bfa8d31d44496b3a1bbc2fb59e21dd9ec34573f7c76609a69f0e33b1f3354f55d266e27c6333956e331b988e7668eff1cc90e9cb5a06b96e60dcb987ede2cc32f2776241c78d
0	44	\N	\\x852ebcc4483656f0c93b98ce6ab2a79fc9f8c7d7c560c146d9fff3e0b1eabd26c9dab54cf74adbfd5cdaca70d8807b37	\\x00995b3dda3ac3ea15f1ad4190045da7b9716527174bad2fd241c33d058c2215	32000000000	\\xa9470869ae2c6b0ed4cf7c3aa3f2bdb690c63a05b5683f2e69c19eaa00dbef7601f940370911fdf038b668412130515c1831ec3a74d72c4bc5b276130a256dd348fe5b832826e9bee959eb5f9f29c36b1e7141efd6e06d4e26726e4acb5656f8
0	45	\N	\\x91a43c9bc21dd27b15ae62017adb5a8c3e5224c5f07af6abd925fac2f8c978c81fd584639f4a7a88ad8f885b04aba63d	\\x00aad32ab570fc0d91bea2b185691d3556a1004f7df1535b3d3d7a0ad8810e7e	32000000000	\\xa4e7f1b75f29da00c0cd0e85190826a332335f63010b46e4a410506b5e1eab86e60c221a79ca9197ddf576fb8a22eeb8052b55f0aa864a341d1590f6200bfbff15f068cde40018e89f35b3c0d35c02eed22e5dab83d14284f73df32f16ad32f5
0	46	\N	\\xa00798fc1a2671615704a2fb650d43368dee54f60e2c1fb19a8a75f7474702dfdde2f05d51667405675d04e59890d690	\\x00b7a80722aa25daca7def5eafb8e5f1559d52fa0c40f45d9d0072eb0313bba1	32000000000	\\xb9cd681ade324840de685e5fc5d09027d7f22f5bd89669c631b43241baa5a78916ed0b258a8600e8ed454d62a15d120a01ed42013a321e9e2ae90fbe2e68b0c486082403bd5b2de503256badffc9999d3a7078a60b88c416ef0ba4e74a3795bb
0	47	\N	\\xadf7ade854a3c6cb0e675db842cc52d7cddfa51a7bb5611bb302a1196471fe0e3ff8523901a9f19ed28cb3bee7425381	\\x0090df29ae1d3f18b94e8a250299a48a4bf1f7ae6b4c807222fc58517659522e	32000000000	\\x9104b7072baba2b7dc813d8dcc17661401ce07ad74fa936d52b102943943e918e6a91226475622b3d79c8968954672b017d91ac2d157749268993ca626f9922b1d2640f0c9fd6333c7403d3b1e736a2c45bb0ec93ae93c6ae55a97ba5087c5d0
0	48	\N	\\xa0491939d33b503684ce69b876979eb3e12d23957219c2e02c4af0767f7ffc03fcac36454ae1a9700eac3493d6c71808	\\x00b73c95fb6e50ad58904b149b89915f244554948e0c201524cc8a53615fcb67	32000000000	\\x85bcf976cdaa76ac9f3f6f73fafa7b28fdefec1fdfa913102fa05f8c2201b186cad1b6e82366c2ea0d9b13c1afc6460912683991128d9ddbb1084211ea0482742f9bc0fdc414f9c9ddcc2bb7e6b2e748c871ba399293e5431d72ee4281e97bcd
0	49	\N	\\xb4eec498e4baf6f0db0987db72cd05776855e3267cdc2e35008bffe2759e9d86c6b67f1b8cbbef7c5a58031af4951b35	\\x00e1ae4d3509356c41a393c52511db9ddcac9b41c9da5ed3f11b39b0097638a6	32000000000	\\x9706d0838d76436babc03ccf1102999e7e67acd3f1eea104e68a392f98ef6ba4b6f6cdcb766cd2751bbbdb43fdd4f35319219209ba066d84eab6000d85e7a232d29c2134a711f4794ee897fb923fcc1da3dc37084115c35366864dcf7244155f
0	50	\N	\\xa7aa7c0d2209e0483499d12ac4e236cd77a0583ab3686399bf0177b5472d53b88ee877c09eb91aa76d3ecba124264b2c	\\x00b00c0f5336621d5c8e8ab416195bd110cea4e7cffca6b4cacf34d626736de9	32000000000	\\x8fe13a3364cff0f79936ff6c283325327bda5cf3c5ece2d5b084a1fd605c35bc7fe23810ab6ef10e4f46959432d5f4bf0a838b8ed560d543568a427bcc91ea9257ef090ac5c34697163496a5c6fee875b7aa7401eeb7544df2273cdbf9c7b2b5
0	51	\N	\\x8e4326fede8adecf5c265479a0c8b5b6ef2949f9df9f2fb0804be6beb05f7a1337027578c161b709a7880ad6f1fc44c3	\\x0041a1e5b4e0834ad4f34962aa38c0c5fc830dbc5b69b2ad6a68f802ee155bdd	32000000000	\\xa8752f2b6355e76e6ce35dda43ae877b9b90b1841c24dc6d5194c9196e946a00da74d761d9e5bb197f062325e361a7fc086fa982b837316adaa6523fcef607a4ce9a07c48589c01b3432346e4bc95162adcc3a168b2bd75f68ebe8f39991d64f
0	52	\N	\\x910558df0bb4bbf30d47fd565376377c22749f0ce215cb1c37ec06cfdf4656f2a2a4eaa5c69c1dcf103b672f397eb29b	\\x00fc904f2401a9d0fa30e0cd7f5f7ff87dfac17278d961ef53de652fd9017a98	32000000000	\\xa0fcb02921557eade7a9c593118557593ae42684fe207bd8f4319f7b403a445d15437033eadb0ebb89ea4ce357345c14086805c12ed8e76f275fb486f62a0d343cd138b0c24f6988f6b698a8310d13d5570b827af97c6f33702bcbdd8fe4eb46
0	53	\N	\\x8c2e89ac29ab28d1f9ad9101741f4a0e02bf02abcb8e4d922d2dc534f5a4ec0769d202c92a7d2b3d505f4f26fe1568ec	\\x0033a6ca2c4f9b7f9d2199ce43c500d001182eba10d27c9a6c9664ba70707737	32000000000	\\xb77a2c6824162411fae3ec3b623f6c12bbb8d8a4b518f18855c10833c2a0d94aa836c7f76815bab840dc183643d163fc08d6e8373cf876ab88f1cc8c418080edcc93a9f884c7733cde7dc1a02a18dd8532c8fb92a42e8111c4e32d142c8e08b6
0	54	\N	\\xb1c7c48d98063b73dd69af7500e8a293107ddd326ccfd0f115c6c69281ea539a4703519f9d79936c60d6565f127b255b	\\x007de621de43d3af51fc2b190b67a619d669b36477908f5611c9dbf5866ec83a	32000000000	\\xb1de39612999ca610c48111ea753e37cedcb80808142b0a2d76dea2a048e3870451c805f0df651fd7e68b57ba372c58715c35a08d6424f0106e6aa80e6436f6f97ff1ba8f8a3bb7423b615df775c3fefc5500ada5a1dfb00aa29f7dc802aa269
0	55	\N	\\xb8adde3b7c6b894ec8ea3cceedc0d84f8502825a6d681417f18a52f13445aa79d16d38463950c535ba196fca21bad4d6	\\x005c935f8eb0985b94d49dabf03a9acd77b8d0efdc1558a285cf259e2946c3c9	32000000000	\\x8980a0282b4d2dd172f49bea6378fb1f91dfdf713b2326eda6a7228c6634c8b8f5f1a05d6c0a26aa0fecfa29397f73771081f197972d28fbcf52381400e2f8ed44aa295822d3733ddc2dbdf45e8d7ebca139b165ebab09e1b1adcc9319e9a03e
0	56	\N	\\x826b214e8b2f90149b28ad26cca709ab5fb09edcf573aa0041df1950add3509abf1a3af2632cb824eea3327e7e3ac25e	\\x007a33cf3a3475670605abcd92410357b74d8d8e65c7eb878fcbeec85ea9acf4	32000000000	\\x9436db08658259be5b867768d9d312c1e68b04545f1c9127b30ca992b003c751b2b346827d5aec67402441fc4251989d0af1bd3e36fced8339a6639994d030b8aafa6e15d209e1e66895c23f9c23b31b1a0e2bfc38315052dac6bd44d3262775
0	57	\N	\\xa0d9aaa26584dee9be29bc2684d9c565a2e2f807b6d82c72400925e12b98904f8e75337cb24651bae306e7e03a89830c	\\x0082845d3cb358397ea932ae03c7f017a1db744c8286acba5bb417ad67483475	32000000000	\\x856e8336913fb865418a3e1116b1e0a8d065c67a3a81b1fa87ad4fe5d8404193d08254492f2494eee20ddf7f4075dcbe0e843cb82ff26af29d13a5e22323b3c181002199754db591b911e9817713a85dd4f88d2b23fd620aaf809d255f5c63f0
0	58	\N	\\x93ceb84f34c3ac783537916014a14b81d5e4aa674076162d87dac6d058bd21e58fb934b90d73f270f7884f31cfabeed3	\\x00463b05ce8dc11e89a628b59777f789bc7225edf20c940914fa9a1bca6a3ff9	32000000000	\\xb033d94520a38248dc9c6ccca7046d3e0ce2179e86083cc45d94ad1f4447f953836f567923649ab46894ab3c712fc4e00ed8f7575d48bfdebef18fb44157f35aba5702828d638cd673e6955907504893a2e0014012d9cc3372c72e873b30a353
0	59	\N	\\xa0c75d5c10aa0ce0369e84d99ae186ff4ca5338fb0a9221d73c46a544e61b450b22d0d9dddcae11a16456e3d58e02928	\\x001d6fb1b18623d209cc323cf056b284ca2ddb9c9d759801968f20bfff3a2410	32000000000	\\x83f40076a301e270e856c6c75b920716dc6b0022be4b3516b4f5801a2ef51a03dddcca8072e8fa50041629691420809109a0fedf1d91ed49791100d3bb8539041cecc7f3df61a7c04182692eab5f05ebe9d31314bb370cac59f360a7266d44ab
0	60	\N	\\x938b0df86bd30d3569067008690fca01cad17286ffbd3c972c32310872b41de1f76c44e13e23b8939c6633a934d4506f	\\x0089a43556070f645f6fbe08756d575453bec54cecea016c9d92153a980de92e	32000000000	\\x9160f1e13a0448b7f5f638adbd309f43b28da9a79e921524fe0fcf8c59ee3984c4b24f18128099090b1d8c7352676f8913ff34b5a2a04c8b2aca2b8c7cfaa8fd3e8c876b14acb37aa1db05b1ab5cc22441c2b9b557e04f20d6ac6f0b16632856
0	61	\N	\\x83130ed10c6bc79f07d8c4d73f0a5bf6407b1ecf47f7d760f1fcf746bc2cba855f9160adc948e93954ce4752c2a82041	\\x0053a23df1b77e47ab234fd2559e9e7ff77a9d120a5575b8625a246d97521e2f	32000000000	\\x8748c24f22dc81d87f36929bc200f70bec7e2fed52043f6378e4d0818e1104c738c28c0c5d3f8f58dc7cca9f1ed326d70f85c50374720292c9024838eefb6356fb557f7521fe609d9d265bf6bc9b9b5d5a035a74f759e989011a0841cf510e96
0	62	\N	\\x8c8e9dee0d5a334640ae786294322a11542f9f2aa967be42a12eebc7957b8f8fbb873be54cd50fb9aac7c8b58889a686	\\x003b32f6bb6f3a76cdd4dda57a6757ea7c98120f0877e17748d1fa1efbd54e68	32000000000	\\x83e9f9dc154bdb3d16dca47374403a6559661df45f55cbd18eb27a97c92fe786a256ad3a822ad342cdafa7211971bc720b4d5c3e0f101852067d0d9bd42336fd6205d0c16407a4826f6d418e2b4068d6ff7bae3089591f624c821d411d39c074
0	63	\N	\\x8d9b9b5eb57c035d1c566c968fcabb371de823d72ebae21a03462f3c6f4faed1bf93355837cffd94ef0deea8e95c78bb	\\x003642c3d4ed9cea0dbfe4bd31c26a77af14952517c0ff4bef12eab2f841b622	32000000000	\\x999ef67137b53b65eddf48756658851e43a51d2195a5f8cd0de935795b325269f86a8afc893636f23082b6e7f5ff62ac0bd258d96ba69db6810e3be180214c01cdcc537c38c864c4789543e20e13d63b16b2aa4a53e74a5f3b7a6a21b9bbeafa
0	64	\N	\\x82e1c3754bae5e17a5707acc4b299804d567b50a8509018fd4cf0f4c9872056934cc158ce658947f0d8ad4869e3e2acd	\\x00ce17b843578be8d0cc53c1771da3b4e609e22965ba102b3d891b5d61948b6e	32000000000	\\x82889e81bdb471ff008b78205aeec95eeec6830082257ec2ca67561a9cc0108eb1984681c0701343543bad18bfe76b5d112c73a89623058defec6c95565828ebf5b884dac10ddae9b31eac02111f59478f12985e109f8e061fa6ce1db643e6bd
0	65	\N	\\xb80a1ade6ef3cf2931b14773161aa2174c96ff8f88b5aaec80b863d321306bfbd3873f9beba785b4c6d50d5d621d946a	\\x009ecb1fd0b47f571eab96157b96cc290acb20c8dab6e2cf8c0d4f11b079749d	32000000000	\\xb4fe255599b994005cb7319e77d7cf2a70fdd5156e82893a286ceaf831632a80177146eb9723c3e988f69e94923f025b19c0301dbba4cc41ddeaad96914b400768225658c8aea7c0385053c1e2c28874ebde10ae7cee596257d08fa0af45cb0b
0	66	\N	\\xb8898dca419f19338fa0e527a5085bef0f944ef066df3a5f1f61fb7bd3c759d17888d01923fd023bbf5591b3ded2ff62	\\x006493022cb4582b6ae5e1525c8fd404f412d4fe0a0424777e03d847d530bcb3	32000000000	\\xaae24fd69063d09aa999690b176efad95a057838409b235ad2909197e37d91f87125f7d79ad0513c187d366e8c924d1e17a6a80ed758df2b40fcc576e10f494ecafc534870134cf3c81a043250a543ecae02035068961e222846ef5df3dd79f6
0	67	\N	\\xb30a4b31bdd96a9b8b9283174e4cc1513f4e1bbb9f4d2d551db6426850b3ea909090ffc18f411f94a6e2f19a50f36d51	\\x0031747a6f2ab541711ffba739171aac3eb49f2c9edda048a4949e53c3c75898	32000000000	\\x8173c380441f968997c47b6b827a24a16ce6367d3a4c41218ec6048b6065f1424b5e6b5c4e600c2d35af9b4608f592b7198238e698ec965b18c89eec2f3ee8ba291316c004d026e8a3ac32180b3dd100b77304cf623b53e58b3b5641317fc86d
0	68	\N	\\x8f20450c0a7db2d0e7123248ec5f16b9f417c38f0f6f868793c7d0c3b672fb14e78aa3fe5c0b3f75ee4f1b19630a5dd7	\\x00797ebc5637771f0ffbba5de132e7e4547575fd67bc820256e37cce6a32b83e	32000000000	\\xb1d802d083ca97d4faabbf33fc72a5d9b7187b242d5ffe952539e6270a22355171bb58d77b78701f8305422cc8407315142390a562b04486907340158d52e482566f604c20094ca1feb0921c88276a593411f7b17f17f4574c327172e8fc04b7
0	69	\N	\\xb4ff5e68271b333325d7b2bb8123577e63cdb6dfa4a5758613feb6b80716ae10c083e199db2f4ca379ab87523b817c74	\\x0046ebb6fdb2c4f631d55774b300fb35df777d6ac9c0e6c685dc8a650f03ca8d	32000000000	\\xa4b9023c1813343ee06d6770ca94b6ca674472d3460f1b7ba5f8c6753292f9170cf273042ddc996bfb7726d42756099505bafa57cfd62cd3654b47ac66a8d77a2740ed725612604146f7e4ee8583a18a3e62d918b28a814619674d532bb21549
0	70	\N	\\xb879a89cf9b9c87dd0e9775b16cb4e977ededf9b3743a0639a7c6c54049a26d9475b29ab3f16c88d813cc2498063cfa5	\\x006698d3a399132c05b7fc610c7013673414e79ded961415be6cbd6c14fcd94d	32000000000	\\x9942fa56dfe54408289640800402f5369cd16a4a79dff7c189b1e6935d8a21582aeaf77b4fec97cdb803d26058faa98e13057e2fc1fb7f19da8976adf57523c69148daa8dd032736d16699439e04d6c1fdc673fa864a4c24d477059b8eea4c3f
0	71	\N	\\x844ec23c577acedb9fec25c92c153ca05a64cfe70ba72627915f7377b3d28d6290b6f909e47e4c97abb8df6e97250fa9	\\x0087b400cf6b1eeed04302e58404bcabd75fc62eddebf0734a1f04e4811b77ef	32000000000	\\x84edbe155b3dd6d41da99cf54400b9a3eb7505b449aaaec809f12f12f905fdfe6dc8b23ca27a0966628ee3be4e61fcaa09e57530b7a89af2958b03e353a35d22d1ac4c10366c875fb67eb79fbc22a5f15f370d748fae53fe886e95048052ae5a
0	72	\N	\\xb2aa98650e2fc5d3cedf36ff5a9a1539f6279b397661e39e4f82db203f6ecce9a942303e2ed68168dc34b378b8a5feb4	\\x00b9b1fc54a348c7d47023bb44c2b6d8f10790177818105773dc23854b44988c	32000000000	\\x968af33e2bc0cc77768664cb48c6c4afa39566bc7166361aadbe575d2f215ddc9b17bf43317cbb7b5bdd757afd64fda10ba0ea3795cd059c023a2df5dbc4fae8de15b4bced07da501031382b5ec4f1e715c71fcdfd8caf67cbfd9517b1bca540
0	73	\N	\\xb16cf4aba3cd4a1935c2f9c94c785e4c9ad4cdb4062a63f249fe56f67392d20be214f3a583cba22868bb971c7ee2a1cb	\\x00762d7145c6391cdcc5e2d05d5b19c44b12b21df2122ecc7f184f6c9c888a41	32000000000	\\xa04484aacd8e39a0e23aa43aebcac1e9571e454807c76d4155a9e23ece11b56845c5a4ec2b72569236b0251c35e31de301e6c60371175cca1cf8e93268b2b59093542dd9cf533741ccab6a7ce712613edd94cd00066f31b004d129ed38d48a7b
0	74	\N	\\xb24178a145b8b49c7f3e84f32e3304cfdf6fa41d1121a362ccd032fcda382994ae24901e999f9f6c16a938631e7d3e43	\\x000556ddd74e20d22b0c9931e61146e494796dfcc2f10c23f559247a99b832fe	32000000000	\\xb1b3989318fec8e43035e2afce821437a88a3bf7195896a99bf78a773a7ecb969435c4bf01ee653f49030cfe976efe6b03a02a7b2d9ea8f5219ca96e432f3fa23de190e96533540dee77c9a3c586b08a57f6a48c7ba0b4a566f19dac8880aced
0	75	\N	\\xa0cc44291c1fa6de9bdffb6f687c137c72cab0abd6b3281c88050cb478e394a18196b2dc44f64fd0866ac1f4e4ebed21	\\x00e6e0a2fc50526b405d5c06eb63d0e3d2464f5116afae90cec9fa030dd8ddb5	32000000000	\\x86c73b7711d69c936cadfeae8ba2e304a503bf68ceb3cd37b3650ede6c19d8a049b6c8c59d4e1a190daf60bcee742b26138f393d315fb33e1b3b8b5e53e17a46321e32650516f4a2dc0018db49e1727af7fc11e671891cfd05263f23db4348c8
0	76	\N	\\x8ace11b7126a6812188465fad6588854aee300a35ed948e51f09b65f72d162c3fd41b2aa3fda15a731d9152a1a88459c	\\x008f5fe72525ff5cf659a7e706823b41a0bd6a4f1b1537c243b2d49f1014f11d	32000000000	\\x8c60b8e3634f5468f70010614938947852b7d9e91753e4955d6b9268a23fc7e246dd19014880f6dba7265e78fca2225604fa6427a5fc2c2e5aee3893f85e1fb33c269b8d3490dd2966b3aca7884f57b4913d36b2778bb3e9c9e6aac8a90f6193
0	77	\N	\\xb1f80875d11f58324df319ce9674f1b84ffb06a7a380b9e110743fe3ab13a96160658baca5cae2a022439c2579a6a22b	\\x003a237b693b7971b691fdbed59a20d669b7db4ac044e3a4df91ed2ed08db62b	32000000000	\\x8552b9538617b380aac517f2db99672e785cba74233809e9ccf3b81a8a491c9ad7dcccd2e2e04ee04716edb75cb6c61413461c60a290240e0ff8c1c97b8e7926a58751264349e8a2f7f49872132e1c606f5f750ea5677c7bb2dc1fa0f4d0fa2a
0	78	\N	\\x94caa466dd3d9399cd597af7ae2fb944f1d27fa619d7a3b7c852ecda4f2489c9e6ed3c74abe02089452a90882d212b88	\\x002c1e1ed20f6e28c1897d1e484d6613845b25253482a6d3fcdab7fce7b1d90a	32000000000	\\x8c0bfebf8751d62d4c3f8f7fc958d2cb54b617ed878304914d5c826f6d79032e8112a32849516f8d3467bd7d6fbdef76119da01e9f7762143a594c03d1d7b6b79c12ee292bce41d847281f41432f2f008bcaddd915711fabb39bb2d5fb66145e
0	79	\N	\\x82c461c6b1e311571fd67c16e87d78d1d2f28455fc34e00c6300d70dc5e8b22752c76306b2ec2bf8e0aa96053b8ce2b3	\\x00cdee43bbba59c4ab8a2f916b7048de2b8980cc86e3c3ce8ad552766313f584	32000000000	\\xa333cc78d57dd50036c6b4a0948b3cac10470d34369188bf23fd79ffcc17d42c47dd0170c734582ada81e1ac77dfd542105809274ca17222071b85fb79a2a166fdadc8230367f4bdd6604e6176477626a694e7f18ee5b127eda7cf7eb42694be
0	80	\N	\\x84ed2a69fff2a9f0081d01d6b2bfef22ad3d27c0edf21570e93ffa0d754913f6b4749cea76542f38bc0df95a4b90d1b8	\\x002b1df3816bd417639237e5b5676969bc8fa3c570f436a7e9238f133f0f05ce	32000000000	\\xb3a0b72af798668853c0563c49c0a62b82157754814a52ce2cf776c5167e8cdae64a6dae2cc9df896d16a0c4c74dc3f719c2765dbdf75330e467a308500ab0df5c2f054815b916d834cfc842699e68c236e935f3c95278c630cfdf0f8d0a957b
0	81	\N	\\x8f3d901d2484e30150a2befbdb72f637ae12b4ad8d39d190364967f95bafb4dcc0f906eb999b2c1cdd70478905f1f928	\\x00b50c195256e77379e65732b663773b6a7b7a9c7c9b6541784195af6721fe80	32000000000	\\x843febb8a88bbe0c686e8a18b93996bf27ace593190907513805b141e66e2e3120c513cf07929509624e8f324c9c900c01c17ab004ad76e3fa1e1a0f3cf0745af16a0c6f082ee158d50d394a7068e5da21a96b1f355f04bcb4b6391fcb637dce
0	82	\N	\\xaa4d06ca85fcfa442f5f777606f1ddfbdbfa11bbaf8cac2ea06b03a1ca8e3b73b46b4519d62349d45c52a1e58ff4a758	\\x008a51294d1d52855b91efa35fe66a1c33478c9524d9ba3b5c68b8cd0f66eb2b	32000000000	\\xb12926600d7726e90dcc4e536bda242414c425e32bc2efee209f13a4132f56b01ef0086f382c86a426b62fded389fe8f06aeb759041fde8b88ba7235776fba6c1d6c022f3fa890519abae3964800b4e5767beed16d483464b2bd0a66e9f63a97
0	83	\N	\\x97fa3618bd698a8e34ba37672bb68ed4e5618b12f0aa94f5a73380bfdec431ebbd88340ab817fa7184259eeaa89eac79	\\x0013ac23fdb5d63778d7e68ad6bef2364940a598c135fd54ca2ef8c3e2eb3b93	32000000000	\\x98a85afea8c5783c114a824afd0c26bafc1b4b4bb59643ac3f07a86ac7c495d7ba3dafc226033da22d4941e9e2e4cbf20d4cf7687248332ce09b13dde0338093787e856dfbf4106c0196a83278cfc16231d6752fa431b9d64a772f053d05ad94
0	84	\N	\\xa40c385b86e965919861d415698cd9b608a5d19f0f25d4fd8dba394dd974fb57eb7fe7a1be188237b14b9c19d288f6e2	\\x007bbe63e0f6554b6a4e89d5828fac6559e3d50e98e438e7a0d20793415d499d	32000000000	\\xa2e023e408b889c7a492eb077a0d6730b26021c4bbc9909825737e436ab3678d6b05125157beeb8c46e054960476c4670bfa8ec9f4a352b4418bb23a7d2768efb3cf942236daccf594e074d1501e60df4d1448dc4029b9d2973112bf3ac736cb
0	85	\N	\\xaada169f71fda55e8d25a082e13783daaf63e96beccd1016a19b74e6cd63ef732962df0bcb99843cf4e66951883ccee6	\\x0079e1c3b6afc33e89bc5de1ffb821efd438c5d27041be849645aea244d8e831	32000000000	\\xaf67c8ce942d69d482f122fa3b3c2cb1a769af01ffddb9b00400e2638b4d0ebf5b351685e774f5bca80fffc485b9b3590c358b9707f199664026e3e5751219f9e4c2314cdc3694ea7cdcc18ef00092b2e1b2945b1685265274ecc1c9e3e256e7
0	86	\N	\\xa2825c3810fe12d7b82e31345dbe397dd382bb806668af56dde384e5a5a24bf488bd2ff02a2b42bc192399d77e868ccb	\\x0097f1d4ce24a466628a2c7ec4d6e97a9fe43e241b97c936743596c8b4539d56	32000000000	\\x91b966fd7a25cd79c3cdfdd05fd16aee572d8bf83b5780776a13cd44e685d157e8dafc3672a31fe73e74ca166e3dae1402e284fb0c4995e8eb8071719bf8565f888c2c791be913cff7466e828a4b4bb2c4b5ca2df02c1d1e99c95e8fde3257f8
0	87	\N	\\xae00791909d3fa62690d0419d7525322e4381d601448f76d181af8398fe67c550a27560dfaa443a654092cb8d0f89adf	\\x00e7e501c212d05267fa30da5bdae257b409a70c53801ea4027579752cd9f255	32000000000	\\x8de36e1fbf69046ef6bb27f04134e9ac2d494a929473b72f01aa47bf5b23e1743addb0569a4184939d66678a4e1ad0d50c26e50b37ab615bbf044fca9ba1e0cb835df361eccff34401116fe777a42895377e8ccc5d64694163409f930406d223
0	88	\N	\\x85220e0846d40cbfdc54102f4c23644ffd303e9a46f0ba92a83514bdc59530f11456a320f9fb8734c2a2bbc3978c2871	\\x0070c438423261873067be189c1ec22f979156edf1ae93e665d44efafc7279b7	32000000000	\\x96f1e6be8b56b40a6c42b826d442ee0df1a61187bccafdd2b59e0c7f0917207007e97f6d261eaceb5e072343f0d5e2510852cdc97fa4a354b6ba2dffd86de6482cec635b62902ffd44b611ef3d94931d858a3f145c743c2c392a3afa5e511f7c
0	89	\N	\\x8c4c8155183cafd5d4d0816d8fdac488d489e71e33e0e340e423bd7a9e4e768524d0367f84b5c59a4f61bd27285f1fcc	\\x00ea52abdc3d375552949be57e2836e236489c4a82483eb029294abf61ae3b17	32000000000	\\x8b731fc13b108fa47c5c8d823c990d94881c792d0f63a0bad966d721b9d2ff065aa87b9ad0ce2ea2a255845edf3c992d02d1f00ab27b81d1a97ded02fab47a18bc4ca40dea10c077114c74d06e01f3dc932099643f0e27fed07ec09e0ea155c8
0	90	\N	\\x9752a6a79b2ef3502d9491f6ef0dd242d609999233895268edadb1e1f431b88785d43be47019028cc8a6152867229d7d	\\x008507c9f5f110b02ab3f61abdfb7a436860eeac039789d8a04b252ce11ea420	32000000000	\\xb4700e150638c23a480c32c56dbdaf1c5776bc22a09dafd05b8cb14452aec6d8d51272d9acdf48cdf6c8d7a368f793bc13cd75632f097a34bde8c0353917ff9259be4f3986347aaeed9b2d64112113b7c529c4603bd195b7fc1bf2de71dc1ee5
0	91	\N	\\xae60a59b127d624c781e3a3f8df0dcc28a93916df163cc0ca9a3b701058c908475eae2eda4166613c4a80b4cdd34ce55	\\x0049528d595eb4a4056f5a413a455f78db90bfd3059036b0f81381cf83bfada5	32000000000	\\x8f372ef8d7d1a39717952ee369b97103f333a0428a1211dfe48cf576cce6232e6d8682bc98b06f995c86548811d3830e1560ba0391bea3a81e5a7285e2d1e79e93e76694d0cd121d9c96c370b82e29215e8f767b35851dba83168ce1d667aed6
0	92	\N	\\x972b37c3878959b6b678ff7a39eb3f66b18b712b08cbbaef4d9deebd20a62ee051fd4688dfef1569e15171ddac436e06	\\x0019d4420e19795c6f8fc26fa6056391ece5b6adfff0e770f6557c09ce8012ac	32000000000	\\xa31372e71f45e277c8c98a099396d23987cd42b9d703d26e2b4ee5643f6e5251bbec5e6d4dfc1fd2121e2297a219b8d5139eb8b10693edd3479628227a7f95fe79a6a18f8d2c56b9eae17d90e3e1a9d4be3117f74778a2d9180034eae5a4aa2f
0	93	\N	\\xa6181e9935a67c439dca406d03a3bbbd1949bcf7825145364ffd6762c8825874d06b4f7d01c6809d4ab8d795c8c0fda8	\\x005167b8c20fbd241318b69a56dce576f47301a7e36e77ec93c30584d036f230	32000000000	\\xabe8db90ffc8300333b6cb15a9c78afaa926b197aa835a2f44825273e71f87dff3e7482f3d99203f441b76b7e2a15c15090d660763b2263ba615fa1c2603f4e19af946095ee2dd21b5e5ad17df360d9b8a6d4ccb9ea685223e5d2b27fe19f16f
0	94	\N	\\xadd895d337ff435378452a5679c8fe17651dba4954780014ef77b978bad5fb29d2d2974c88372162d693a400783370c6	\\x001783d5e1644838725a15dd60b1c1df4ea59c712deaf20a746499acde9e045a	32000000000	\\xb64754d42a10c18f4b1398a48fcd26c5b2b5cc64db75b6ac3b832dd14c6f25e509e3d877547857563db5350c50c295cf0c314809b7fcad4a136bee0dd003aac9f4ae2ddbccb0e6a82d7f7895860b91dd84604e6cbf68fa61729bb87dc879250b
0	95	\N	\\x99d58b0547ae38f9d6e3b4cd3caa3c6032d6cb0f6254e1941ac18e9972780a719ed78ac761253e3930db6ce600fb7aa7	\\x00fbf2304d948058224264699342584b3aecc9078f2731f02454ddef07faa95e	32000000000	\\xb29e763eb30b4331a1f788561ea8204ee0903f5f56a8298258f3b9e7decdd26a9b098e908bdc4f833d37c9388ab7b1de02f9ed8dc3a500e07d5740d8e49e9e4fccd44e4d5fce58372a36c85ac564185a97b188787b983579fcec86e5a443c025
0	96	\N	\\xb7622055e39899f48b8133531ae66658b5d098773d60901bd0fb58d8d70f20fd89ab43351a3baf4fef8cb42b53aedb34	\\x007b8e4ac9a82e7cedee7228de5fbe0612ba8478c4ead1f703125d161d9a11e2	32000000000	\\xa48681fe15b00d0930a3edd72f86d553d570cf09ad5745df4c7517965d8a92ac8789a6b979e4353863971e9d57f3e0a80790ac26d9cf7de26844a1acb3d252b5d8638a84fd6873b148baf42cf6addd79ceba725d25e9c4bd513ae467502819a0
0	97	\N	\\xb86f436bd51db1ae4776f5e360b40c9b05f07ba8590a15c0b73ebaf917f979510cc4756cf30a6526733d18eb986b8e4b	\\x001b9043ea9496b321f8f4390bac725cb10d3c533315c5a81b293baefebd6cc1	32000000000	\\x8a6c19a0826f24b45bff59646e9186f0a7ad175fc424618186de0d9f5b7551075121cabf0c1d6b62f2a9ca0c2af21a400ab2d4aacc2e846edb36fa36eafdfc22ff23f3dcf105d1c88994c60423f483383b872a219021f3ba1385ab0faa3f40e7
0	98	\N	\\xaa6be03c9514abe44b49eeb8acdca85a010b68d14e8d5067970db0a9ee7384037dda075d386d8a9c287d1c8b36d502bf	\\x004d717102d53ab7be17d22becadf824491b6a091fbe5841efbd4d88a740787e	32000000000	\\xa3384f4d08248b39686978ca36a14b4b345de307592cd02f3848c9366dd004e3c5c94bb3655d0585ef4f72c70ed9471b0b95760959a7b9e7cc1bebfbf9b3810eca4aca0919da559ff4b73f540f87e5074cdcf8ee73a45d378ba02772dbda19e5
0	99	\N	\\xb131a68b8671d5cb0c44aea145330a218bbc0698f3cd81c3fed88b12780e713254327957f3f953e26996a63e80462815	\\x000459b532b9111968bc20d918d88ca1a71577ba17b97cb338f45f334b9a8493	32000000000	\\xb01ea8ad1a9da783746761bbe860ddda0fb80cbf8224b3742176887edf70227ff0724270a2ab8946c04993200e6645b60e278b02fa540b79047ca7629c4bc2ffc2ce4e1d19e2cc52dc2043b1d84eb756fedfd948efcd33338dc7957eee881529
\.
COPY public.blocks_proposerslashings FROM stdin;
\.
COPY public.blocks_voluntaryexits FROM stdin;
\.
COPY public.epochs FROM stdin;
1	32	0	0	173	0	0	16384	32000000000	524288000000000	t	524288000000000	0.9866943359375	517312000000000
0	32	0	0	146	0	0	16384	32000000000	524288000000000	t	524288000000000	0.9866943359375	517312000000000
\.
COPY public.eth1_deposits FROM stdin;
\\x0212cfa91d5518e154ff20570037c7fde7404f9916d657babbbfe80ed9afefd8	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120ed47b21c49600a4f3cbfb11445be5b2b235418354290705dd59d4b362fd593070000000000000000000000000000000000000000000000000000000000000030a4e8f56301594c4ef7544eac805c2c701e62124f6bcc2ce39ae242fb34a01795c9f96f3d5667773b0329dfb9b5ec3e200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200087b8efff1e5494b3f4a85ea9b82385671b4b2855457d5ea656f863da457f5600000000000000000000000000000000000000000000000000000000000000608ad6469eec0bc881369c28ea214a601407953fd5b0343e1f033524506d12ead7153ad0328f31bfe81af96f8f2602d22012a2d854cd848131c6342bd88a3c9bfda9adc08d09774b53bf5f7715224d5c40cc25a5aaef7386765fd6cd1fb4628c5e	4	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa4e8f56301594c4ef7544eac805c2c701e62124f6bcc2ce39ae242fb34a01795c9f96f3d5667773b0329dfb9b5ec3e20	\\x0087b8efff1e5494b3f4a85ea9b82385671b4b2855457d5ea656f863da457f56	32000000000	\\x8ad6469eec0bc881369c28ea214a601407953fd5b0343e1f033524506d12ead7153ad0328f31bfe81af96f8f2602d22012a2d854cd848131c6342bd88a3c9bfda9adc08d09774b53bf5f7715224d5c40cc25a5aaef7386765fd6cd1fb4628c5e	\\x0300000000000000	f	t
\\x6a0a51cd2d2d67bdb18ea988a3aab15bd643eaf4cf0fc6c5dc05d12d50b05e7b	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120c13db3b3578214217813d5efa6c28da98ced95e3d91fdd155e705279d2fac9170000000000000000000000000000000000000000000000000000000000000030a32a6208028023dc022a896421cfae19d901aa49943b099499b252957819048525184085e3b650b5315804940b44598b000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020009192f1645b0cb7e047b33b5ed1589d480c6d5680c3ea74eaae0c1650ae96540000000000000000000000000000000000000000000000000000000000000060a4fba4ab172cbca1678c3824212324bab4c8e4e89cfff848b844d4fca37861542121aca1a9c71dc676ca23ae7e48bc6f12a88e2c4f0b25cd0878eb7b7ab86417f61c50f921129a4e8848935c496afb429cdeac82cb4d1078859fe350375efef0	8	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa32a6208028023dc022a896421cfae19d901aa49943b099499b252957819048525184085e3b650b5315804940b44598b	\\x009192f1645b0cb7e047b33b5ed1589d480c6d5680c3ea74eaae0c1650ae9654	32000000000	\\xa4fba4ab172cbca1678c3824212324bab4c8e4e89cfff848b844d4fca37861542121aca1a9c71dc676ca23ae7e48bc6f12a88e2c4f0b25cd0878eb7b7ab86417f61c50f921129a4e8848935c496afb429cdeac82cb4d1078859fe350375efef0	\\x0700000000000000	f	t
\\x473eaadfe1585e0a84177587cdebc05c3f98a4aa44dffeae231a356e5efefff2	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120312f1da20b81589ffd946e4c000e14204df628d617fa1bb1570776d29bb199080000000000000000000000000000000000000000000000000000000000000030b54268644802909635c3ac0630394c8608a705d335277a3a0e130b86d23300312afd7e5d52e4e44aaa67f2810bb1ea60000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020004b211838ed05a579a285aed07ee9b5b353d6c3d67fd0aa29292dd4266346a7000000000000000000000000000000000000000000000000000000000000006089d00cb3e91ff90f863c16e9ee1fbcf3cf0d41c445916b8c770bacfed39bc33119ab82767a0f755d4508262a8333a7c01333ad2489e58ff0d3e44a023e5165bba4ea52d8e861161a29aa67c82cc532e3c09618a78466d5ffcb43c59997c24dff	5	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb54268644802909635c3ac0630394c8608a705d335277a3a0e130b86d23300312afd7e5d52e4e44aaa67f2810bb1ea60	\\x004b211838ed05a579a285aed07ee9b5b353d6c3d67fd0aa29292dd4266346a7	32000000000	\\x89d00cb3e91ff90f863c16e9ee1fbcf3cf0d41c445916b8c770bacfed39bc33119ab82767a0f755d4508262a8333a7c01333ad2489e58ff0d3e44a023e5165bba4ea52d8e861161a29aa67c82cc532e3c09618a78466d5ffcb43c59997c24dff	\\x0400000000000000	f	t
\\xf870fcfc401c55ffb85f51bdcd00f073abee683c23d9e89c72c9ac87e9d41bfd	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001204ab234e7471a83ec63e6dc8c7e089987877e04808c0234230849c5fcd09a724100000000000000000000000000000000000000000000000000000000000000308bf10faad775298f089941b54072081933430077efa7edb8800a5ad67b910d1d4906a59aa09a2a82c4bbb2eb572ec41f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020006525c13a1be61a018ad639e181648d58a1647c16cb7eccf77213e0f82517970000000000000000000000000000000000000000000000000000000000000060b7a32454c6363c9c4f54ef80643ab0840769d6b5b4da62a79a3d146f484adcebeeda802620db45849fafd1810964607610f336e229b8d4d383df52e00f423968b207d9e0e388248403f05891b4ade9c1f4736c77d91617975ef6ec06fd691dc4	3	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x8bf10faad775298f089941b54072081933430077efa7edb8800a5ad67b910d1d4906a59aa09a2a82c4bbb2eb572ec41f	\\x006525c13a1be61a018ad639e181648d58a1647c16cb7eccf77213e0f8251797	32000000000	\\xb7a32454c6363c9c4f54ef80643ab0840769d6b5b4da62a79a3d146f484adcebeeda802620db45849fafd1810964607610f336e229b8d4d383df52e00f423968b207d9e0e388248403f05891b4ade9c1f4736c77d91617975ef6ec06fd691dc4	\\x0200000000000000	f	t
\\xc8854704903370af4660a1efc172f6ebaf6e05d9e01d9847191496e4a9089513	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120b1574b7e5d0999b176ccc0eff6a3002ed85f37c3e07d58143c04dfc3294644c4000000000000000000000000000000000000000000000000000000000000003083562589ade60f551f6b1f19ead319bc6031dee66f46e173e05f1cb06d85e7e0b1960e8b263a7b96d15916e9806446f600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000990af45751941807ac88896701b0fe33f62e87b02f8f4729d7b6e71689c5ef0000000000000000000000000000000000000000000000000000000000000060b2a7e35a4fcad6eae6246650665013fe8e8a6d96adab574de3404838ad31d4dafa925e768250d38f52b109b93e41d26f19eb7383db8525d592f632544d832ec3002d8f3929f54b3dd18ade98b2f48a82b08accd7160e8c64e71530b083978d2f	9	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x83562589ade60f551f6b1f19ead319bc6031dee66f46e173e05f1cb06d85e7e0b1960e8b263a7b96d15916e9806446f6	\\x00990af45751941807ac88896701b0fe33f62e87b02f8f4729d7b6e71689c5ef	32000000000	\\xb2a7e35a4fcad6eae6246650665013fe8e8a6d96adab574de3404838ad31d4dafa925e768250d38f52b109b93e41d26f19eb7383db8525d592f632544d832ec3002d8f3929f54b3dd18ade98b2f48a82b08accd7160e8c64e71530b083978d2f	\\x0800000000000000	f	t
\\x59373b539e4e152590e9131db025df70e67e279b8c6b9adffaf9203d00e6be52	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012000df358ac8110a1954468290c5d5d818116c08bd2449a436a5913b5d8cfab63b0000000000000000000000000000000000000000000000000000000000000030b3cf4e3f286c3d07406154a509e12065adfe0b2decc00fc55de4adc77900feabe24e2ca095221834dfb9ef768c629d3300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000de417ba822a4d19ff89cceabbcfa5b88450019e1e16d37937fd32aa219b82d0000000000000000000000000000000000000000000000000000000000000060991dadcdff493466045a4f8aeb12ab04e09a9c5da9ac5e61e0fa859c597e6351754cd4f7f2413381ff9d80d4115e1f6615f3781f032a6feb477d855e9701914021ca53eab93e1bb8354780d58cd4bc393fd7170353ed6985d642aef61b122d5b	7	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb3cf4e3f286c3d07406154a509e12065adfe0b2decc00fc55de4adc77900feabe24e2ca095221834dfb9ef768c629d33	\\x00de417ba822a4d19ff89cceabbcfa5b88450019e1e16d37937fd32aa219b82d	32000000000	\\x991dadcdff493466045a4f8aeb12ab04e09a9c5da9ac5e61e0fa859c597e6351754cd4f7f2413381ff9d80d4115e1f6615f3781f032a6feb477d855e9701914021ca53eab93e1bb8354780d58cd4bc393fd7170353ed6985d642aef61b122d5b	\\x0600000000000000	f	t
\\x0c81e91680a3cdc21b65d049d20e7b2b87623fff2bf73455b68a585e50247708	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120e7db1309313ee743c15e1c5afb48e5a9253ad12f2b04405922990fa5e5ea17e80000000000000000000000000000000000000000000000000000000000000030b1fbbc356aa240f7d4a825e0e51eff0bdc700de3929df296b6e49a1d85130b62bcfa00a6912acad2f8ea55a0c2b00c7500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000b7e59f3808089f25eba12a6fa66a5c0fde9ff73fa5805e644b83fa272d79ac000000000000000000000000000000000000000000000000000000000000006081e166d4fa1dcc730a8bc7ed38189cd092bf0e09ddb751b5db61423e146197404931743fadd4f2e4a396c90cc2356ac2079e2994623030b10634eb52f4b6dc5201d5200053322bab53b1210f92718189e339fe0c6dc94abf2725ef6611436137	2	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb1fbbc356aa240f7d4a825e0e51eff0bdc700de3929df296b6e49a1d85130b62bcfa00a6912acad2f8ea55a0c2b00c75	\\x00b7e59f3808089f25eba12a6fa66a5c0fde9ff73fa5805e644b83fa272d79ac	32000000000	\\x81e166d4fa1dcc730a8bc7ed38189cd092bf0e09ddb751b5db61423e146197404931743fadd4f2e4a396c90cc2356ac2079e2994623030b10634eb52f4b6dc5201d5200053322bab53b1210f92718189e339fe0c6dc94abf2725ef6611436137	\\x0100000000000000	f	t
\\x686eb74f2761a88e993d7591631c50475b8f5e7ba52cabcc53ccdf061974f6d6	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120f88ff2e514d053f3615a62e7502fa5c8b450f615023bb7ee231028a0f571bdd40000000000000000000000000000000000000000000000000000000000000030887e202bf6cbd63da61fd95b3b42406d47f91bc159a70e6697701afdcd64193e5ebf91a526b027e004fad8c600cd818a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000a54f28489d114c335d9208b9955e6c4b399287f23946a1e692580e2bcea72d0000000000000000000000000000000000000000000000000000000000000060b803908c6e859657f6e809139443a5c45ebdef61f782968abde749d2fd0d0f13cc63904595910702d48a1e6801afc4d50a5880f9d273d02483b563e87a85bddd7644bc4df6cf125aaffc4d3742d2c152dec2f1d8a258968ff226b44f0d26f594	6	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x887e202bf6cbd63da61fd95b3b42406d47f91bc159a70e6697701afdcd64193e5ebf91a526b027e004fad8c600cd818a	\\x00a54f28489d114c335d9208b9955e6c4b399287f23946a1e692580e2bcea72d	32000000000	\\xb803908c6e859657f6e809139443a5c45ebdef61f782968abde749d2fd0d0f13cc63904595910702d48a1e6801afc4d50a5880f9d273d02483b563e87a85bddd7644bc4df6cf125aaffc4d3742d2c152dec2f1d8a258968ff226b44f0d26f594	\\x0500000000000000	f	t
\\x1fbb4ae1cb16dd17db524e02fd316e380dac16cb4ec6f7263b83e5b0eeb989bb	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012064a566aaca534c5f06ed9de603a8b1ae6102eb2ff50fe3518e2e6eb019217a510000000000000000000000000000000000000000000000000000000000000030b1391379360cb6fa63c4d4d90b116a2104640c69887193b985beee07546816a6a02074b1ffaca63c9853d3f2f85c98dd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020009a716695c1c6b19c75db730aa3cb92b64fae5ca257a81a918696bdf52775700000000000000000000000000000000000000000000000000000000000000060a78d5d520619102ebe808e38e2d0de3f6d7a616b1ae2b2c54d4d7117b183cb2172cba3db620f4049c8b0326b923ab6090cf2307959cc9b82d74c5140910c64d456d2adfc92cd60d44ca4a6a2626e1b9e81095091dd08129071cab274e2396c85	1	3702501	2020-11-05 20:56:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb1391379360cb6fa63c4d4d90b116a2104640c69887193b985beee07546816a6a02074b1ffaca63c9853d3f2f85c98dd	\\x009a716695c1c6b19c75db730aa3cb92b64fae5ca257a81a918696bdf5277570	32000000000	\\xa78d5d520619102ebe808e38e2d0de3f6d7a616b1ae2b2c54d4d7117b183cb2172cba3db620f4049c8b0326b923ab6090cf2307959cc9b82d74c5140910c64d456d2adfc92cd60d44ca4a6a2626e1b9e81095091dd08129071cab274e2396c85	\\x0000000000000000	f	t
\\xd143fd8aae98a2a9d7cadda3562364600a6f61364c7dbab8daffded0981e74ca	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120665e81ecd260a0e53165b325959a1b1320ea71fb4d6cadf327699318addc35d00000000000000000000000000000000000000000000000000000000000000030abfb6896a5bf3105adb4e03ca49c487cd75f0cf7570c917a396fa8211e933fd9297fb1def46020a03e21c1956262563d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000beb39f317480db5397db86a708d4606cbdc1d03767a5b28dd5ddfeeed7ff94000000000000000000000000000000000000000000000000000000000000006083f56575ca5d93bd28ef1b359179239acaaeb915126348a889b08419913ed3012c61b19d000fe6338dd21b342ca0481e12dad7b30e5d275f65999ff41033923dfd3de9e801a4d12b617d1263bf21a0ae939400c9742f1bafb6aebea6f37f52f1	5	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xabfb6896a5bf3105adb4e03ca49c487cd75f0cf7570c917a396fa8211e933fd9297fb1def46020a03e21c1956262563d	\\x00beb39f317480db5397db86a708d4606cbdc1d03767a5b28dd5ddfeeed7ff94	32000000000	\\x83f56575ca5d93bd28ef1b359179239acaaeb915126348a889b08419913ed3012c61b19d000fe6338dd21b342ca0481e12dad7b30e5d275f65999ff41033923dfd3de9e801a4d12b617d1263bf21a0ae939400c9742f1bafb6aebea6f37f52f1	\\x0e00000000000000	f	t
\\xa93644066615a0bb2e25d894d92b47bd57eaca0942d472c41f9271dfefcf82e4	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120f5aa4dc936975277a3fe37e4e3067ff0828359a9368e549d41fa745bf297736c0000000000000000000000000000000000000000000000000000000000000030977c217ebe82ed92c00aa4f6f11adab18c6c585d1199612409ce8f24bf5cf665161dc5b3c991835f4d2d59c6a14ae2180000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200077b4ecac02b3dc1d752dcb66c34ee675d719d29d3770ede0d7ecdc92b588ee000000000000000000000000000000000000000000000000000000000000006099aa4668e4916511bc8cb1cea8794976c4695342b9ad0d963d726280d21ba3ba42ec4fd0103fbbcaf9466aaa56e1eabe17aa75d747bfef558941f7fb8b524c06365eecbbe77d5d93022216b15679b5b5ddcec90fb026c83b09a0e90dc17ea332	4	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x977c217ebe82ed92c00aa4f6f11adab18c6c585d1199612409ce8f24bf5cf665161dc5b3c991835f4d2d59c6a14ae218	\\x0077b4ecac02b3dc1d752dcb66c34ee675d719d29d3770ede0d7ecdc92b588ee	32000000000	\\x99aa4668e4916511bc8cb1cea8794976c4695342b9ad0d963d726280d21ba3ba42ec4fd0103fbbcaf9466aaa56e1eabe17aa75d747bfef558941f7fb8b524c06365eecbbe77d5d93022216b15679b5b5ddcec90fb026c83b09a0e90dc17ea332	\\x0d00000000000000	f	t
\\xd1e02f71db4519df6f0aa6fdcf5e0aba8cdd0ddc914227ab4905844b4f128433	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001204111a440728f4318d0ecfeddd3b098d9a347163f53c0ca3dd4018b74b06ef0b50000000000000000000000000000000000000000000000000000000000000030b528a635883d81c4cd4b0af451765471114df70eae4c91d974a8b0dae9c470656fcdea4135b3c9013b601cd6463ba16f0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200086f4ae865356daf55112079b84ba696d1cdacaa4e47237a8f780a1a4d9f4d90000000000000000000000000000000000000000000000000000000000000060a28abcae9291723b7a83c14cf96e3983b2c93b2969840db8fe1596f5bfbce9112a0cb1c857944f29c90c315d536675c8095fdf5c436ce4639f484a9a95154525e8acab8c7f038e0eaf5753bd90af4c15ab53da1dae8e076dc54f8818cee44c04	8	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb528a635883d81c4cd4b0af451765471114df70eae4c91d974a8b0dae9c470656fcdea4135b3c9013b601cd6463ba16f	\\x0086f4ae865356daf55112079b84ba696d1cdacaa4e47237a8f780a1a4d9f4d9	32000000000	\\xa28abcae9291723b7a83c14cf96e3983b2c93b2969840db8fe1596f5bfbce9112a0cb1c857944f29c90c315d536675c8095fdf5c436ce4639f484a9a95154525e8acab8c7f038e0eaf5753bd90af4c15ab53da1dae8e076dc54f8818cee44c04	\\x1100000000000000	f	t
\\xc5fae7cbf0fb5db3d5e5b18e5344ed30eee4d1735d092560bfa9f4a972a99998	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120760837d346e1f6579a8c1562a08b8d66690483da1a523d1cb3a122cbd3b4a521000000000000000000000000000000000000000000000000000000000000003086accb54629ceec44a4c06436ce639fb5d0ef4503abc0d3958a5007c6f3c4b2eb2b4f8155cdce8452f2f01d537c41328000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020003e97a4e3d984d983c5023aba79e2abfb765fe58ab59cd55f85aa2ddb0f4b920000000000000000000000000000000000000000000000000000000000000060967442d3ee08b01bb6d62b3a72a3986ec67807336de3553adfddf9fcfff50f93ed6dbbe25a637afa0fccc35737e775dd1454efc47ea5b54ec326d84ea226599a4710a503b4f6758339176be01527a234fc2411a26291d6dcd69e78092e053f6a	2	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x86accb54629ceec44a4c06436ce639fb5d0ef4503abc0d3958a5007c6f3c4b2eb2b4f8155cdce8452f2f01d537c41328	\\x003e97a4e3d984d983c5023aba79e2abfb765fe58ab59cd55f85aa2ddb0f4b92	32000000000	\\x967442d3ee08b01bb6d62b3a72a3986ec67807336de3553adfddf9fcfff50f93ed6dbbe25a637afa0fccc35737e775dd1454efc47ea5b54ec326d84ea226599a4710a503b4f6758339176be01527a234fc2411a26291d6dcd69e78092e053f6a	\\x0b00000000000000	f	t
\\xbaed5faf83d1b9d8c81da38f077e6f0468f51fcf1812f158465f7d6b558033ef	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001204d973e5b31080ed9219f371d2f41bcdb26818c47143b212eb5b4b0644741e5770000000000000000000000000000000000000000000000000000000000000030a9e7e48d0c3539f260c01cb93b6a2e9c7ec94d2b5a6134d4b09561413043a8e686e1fa2e545a981146e4686d99b8df3d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000cc3fc841cd79655f08fba3c8a6524b0fa945447874a76b6de348349b9ab3900000000000000000000000000000000000000000000000000000000000000060aab3d67003bf7da0a73c13ec9c86772157f851a1de36bd18ca28ac51e381bc8db434e190b3202b1444fbc8b23f5dd9170d80007d67a5ecf891a4f10938339a3890d1f9577586724aee10cdc581041ebb15e0cee236169be26081a07d44511449	1	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa9e7e48d0c3539f260c01cb93b6a2e9c7ec94d2b5a6134d4b09561413043a8e686e1fa2e545a981146e4686d99b8df3d	\\x00cc3fc841cd79655f08fba3c8a6524b0fa945447874a76b6de348349b9ab390	32000000000	\\xaab3d67003bf7da0a73c13ec9c86772157f851a1de36bd18ca28ac51e381bc8db434e190b3202b1444fbc8b23f5dd9170d80007d67a5ecf891a4f10938339a3890d1f9577586724aee10cdc581041ebb15e0cee236169be26081a07d44511449	\\x0a00000000000000	f	t
\\xced2dfdfc0b682f6b309eaa3e65b4fc86a9130d5f17368044e563d6a4eba836c	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012035d09dd484d38389217422a00d05fd92e8ee13d89b76aa0586e8d6ba2ee740ef00000000000000000000000000000000000000000000000000000000000000309946c1c8da6d97b9387262f43762befdf0ba714c4b3884051411a0e7f18e5c5dbefb99099c737fe9f54f124d02ed17070000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200022d4d5f14bbf6cb0fa60f58e9befe6444de903821af0dd7b7e996b2da1e77e0000000000000000000000000000000000000000000000000000000000000060ab93f453b8ffdb8356193671d01095955498028d2c3c5ddf49911cd91f06ab127bc35c604d21f17a3227f06a8a6bdb0d1476f27e8dd7072aa18e810469d47a0073318be8c2bdd0072ed61a272968bcd67e640aa0d8ac96fc7af05d6de83fc0c0	3	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x9946c1c8da6d97b9387262f43762befdf0ba714c4b3884051411a0e7f18e5c5dbefb99099c737fe9f54f124d02ed1707	\\x0022d4d5f14bbf6cb0fa60f58e9befe6444de903821af0dd7b7e996b2da1e77e	32000000000	\\xab93f453b8ffdb8356193671d01095955498028d2c3c5ddf49911cd91f06ab127bc35c604d21f17a3227f06a8a6bdb0d1476f27e8dd7072aa18e810469d47a0073318be8c2bdd0072ed61a272968bcd67e640aa0d8ac96fc7af05d6de83fc0c0	\\x0c00000000000000	f	t
\\x7225434aa3a2096880252980c8429cfb7da3a952c822e12bd569a54d813502f7	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120051841e5a0212f5c63ae9563b8fae880c28da05ad1dff32fbd8d9523b411f90500000000000000000000000000000000000000000000000000000000000000308bb4551148eba7433c302ea318e7c18ea44a1b8ebf9555447d03e0c80e77d3b745ae8f517eb954ca0dc52d46fc6296e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020008763b26fd367d4ed8edc1bb58aa2838ff1045a1af82f851da9dc6e34b49f5c0000000000000000000000000000000000000000000000000000000000000060b204c7850c57bc5191f683cf9bd8f4d9c011183136e02df8e690955a12c7f7c1782b98b042b9fcf4a5d7c6448daae23006c2974c5ab7bcbea7e6ba59752e6aa83bee563c3286ddf530fd6702bf3c36433655b79cfa91e30f54abc471dbe63c3a	0	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x8bb4551148eba7433c302ea318e7c18ea44a1b8ebf9555447d03e0c80e77d3b745ae8f517eb954ca0dc52d46fc6296e0	\\x008763b26fd367d4ed8edc1bb58aa2838ff1045a1af82f851da9dc6e34b49f5c	32000000000	\\xb204c7850c57bc5191f683cf9bd8f4d9c011183136e02df8e690955a12c7f7c1782b98b042b9fcf4a5d7c6448daae23006c2974c5ab7bcbea7e6ba59752e6aa83bee563c3286ddf530fd6702bf3c36433655b79cfa91e30f54abc471dbe63c3a	\\x0900000000000000	f	t
\\xe20775d1c37efbc3f4560f25c6c9b1a33e26d204be5d21bd24b7cff78b683756	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120b7c7cc40b75c82e54dba2684662122eaf57fdcb71c2f6048a4917c29f09f740200000000000000000000000000000000000000000000000000000000000000308b4ff884f0825d8368f4077ad53149759332554cbe3655c34518a30e75365857ae787b208301be502249bf4e06e0327300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000a78a401ca1d5c49c62aa99ee38e2052527a090723474492aa47f2361bd88e60000000000000000000000000000000000000000000000000000000000000060ac55f66329dd396905c64ad4c7b73a58c341f3fde4b7859670fa7ed77cd6b291baad139c9c54f6079ec73885a6bc7d9f1625bc0e3aa6fb06eda426fb4a94634f178422dc53ded7bf00d65e0d7ea3461b272444a9df14a4eda554ec36e56734b9	7	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x8b4ff884f0825d8368f4077ad53149759332554cbe3655c34518a30e75365857ae787b208301be502249bf4e06e03273	\\x00a78a401ca1d5c49c62aa99ee38e2052527a090723474492aa47f2361bd88e6	32000000000	\\xac55f66329dd396905c64ad4c7b73a58c341f3fde4b7859670fa7ed77cd6b291baad139c9c54f6079ec73885a6bc7d9f1625bc0e3aa6fb06eda426fb4a94634f178422dc53ded7bf00d65e0d7ea3461b272444a9df14a4eda554ec36e56734b9	\\x1000000000000000	f	t
\\x40bd1e2e6e2e8c460a233b5307f0cd22f6c95812f5df538e6feb63d6d33924bc	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001205f5cfc6a522b80de4e0059263f2390b055372cd5ca6d21557f01c871801506a70000000000000000000000000000000000000000000000000000000000000030ae94ec1191683f51784b6e6201edac5a6d1365e97271cb6ab048d722c770de76eae0f901d83ea4ef1e1ed84cd3c80aa10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200047d061828d0a5139c5e1aad92b58b7b0bf8e836f16c36322e95c7dc90a4d400000000000000000000000000000000000000000000000000000000000000060884637e39ed722476bd39a37a16e462144a14f4dbbf2a38cd2425817cede480e3508ed436b1aebc8610bb4b1c395793e09e45156b9db8758c4f13b16324b521b3896a51e3c7fbac7828265ffbe0cc854ffeb214fb9165cf9e198d276e1fdc916	9	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xae94ec1191683f51784b6e6201edac5a6d1365e97271cb6ab048d722c770de76eae0f901d83ea4ef1e1ed84cd3c80aa1	\\x0047d061828d0a5139c5e1aad92b58b7b0bf8e836f16c36322e95c7dc90a4d40	32000000000	\\x884637e39ed722476bd39a37a16e462144a14f4dbbf2a38cd2425817cede480e3508ed436b1aebc8610bb4b1c395793e09e45156b9db8758c4f13b16324b521b3896a51e3c7fbac7828265ffbe0cc854ffeb214fb9165cf9e198d276e1fdc916	\\x1200000000000000	f	t
\\xee524f34a28b0417a29cf8bdee0134b33df59573e956bc057d9eab0cfe40804f	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120fcfbc89cb72dbc7641d28adb811d69e79224bbb48ea2675b855ffd313baf10bf0000000000000000000000000000000000000000000000000000000000000030b18943c3b4a71068a436e877bff53a8cd2de552708f9e8ebe81e1ebca8ed19f5f577044884edf4f700949b480e622cfe0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200012b57157e16f8fbbe5e18dae4a8bc7dac61a73998bdc1f6a51c9cd534bc8e20000000000000000000000000000000000000000000000000000000000000060a6045f581d89bd83b7abf491856db89e7d733fd2764f04b14c4224e7ba93ef348d9e33109f6a4c6ed14442b181a8eb9306ae706326f2ed92bd1b0e350b829b51bef56d87d0f9236adfe64c0c067dbfd5588b9b56e386ddad4b44fedebe26a5ef	6	3702502	2020-11-05 20:56:38	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb18943c3b4a71068a436e877bff53a8cd2de552708f9e8ebe81e1ebca8ed19f5f577044884edf4f700949b480e622cfe	\\x0012b57157e16f8fbbe5e18dae4a8bc7dac61a73998bdc1f6a51c9cd534bc8e2	32000000000	\\xa6045f581d89bd83b7abf491856db89e7d733fd2764f04b14c4224e7ba93ef348d9e33109f6a4c6ed14442b181a8eb9306ae706326f2ed92bd1b0e350b829b51bef56d87d0f9236adfe64c0c067dbfd5588b9b56e386ddad4b44fedebe26a5ef	\\x0f00000000000000	f	t
\\x4c7501ab16c3dcba6db5f993a7686a25facad5988eb831f216d2756b0ccee5b8	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001200109a39061b60da75db4c530123961ab955212483928b88c51f976f99e1e8d4c000000000000000000000000000000000000000000000000000000000000003098bf2d673a152213fff14f66827978aaa04a5a1f0fcbd95fe6fd521af08e470ff74d1b6667ed1a607dbb6b597ae7c57300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000d2d24d652c1ded1c9453b663ea35d2b1964b5581fb78a1fd745c49659f92760000000000000000000000000000000000000000000000000000000000000060a2de94027b2af9b2197fc112eeaa41bf3620bde0cd7306ccc42d543c3b133f8c79de66f328fa28f2a26b3c6c91d7517e15de894e64fbc5c1c44c0cc29e5b9cefdc9862b6621465f76dd1a68fbd329006db308f2f642357c7c702e676683c1f14	1	3702503	2020-11-05 20:56:53	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x98bf2d673a152213fff14f66827978aaa04a5a1f0fcbd95fe6fd521af08e470ff74d1b6667ed1a607dbb6b597ae7c573	\\x00d2d24d652c1ded1c9453b663ea35d2b1964b5581fb78a1fd745c49659f9276	32000000000	\\xa2de94027b2af9b2197fc112eeaa41bf3620bde0cd7306ccc42d543c3b133f8c79de66f328fa28f2a26b3c6c91d7517e15de894e64fbc5c1c44c0cc29e5b9cefdc9862b6621465f76dd1a68fbd329006db308f2f642357c7c702e676683c1f14	\\x1400000000000000	f	t
\\xb57b8462e849bd575854f2940442faeab770c5ad801d2bd2298a6f581f292fae	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120f980bb1ab1b1e93fa5cfe64649c4145cddbf9e502eed6d915dc309feec4eb5690000000000000000000000000000000000000000000000000000000000000030b05b4fe1b803df50e7f275f399b629c77b98037492d35038121c48a258df5a8dec9532836a37e481d28623b7b5ba57a900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000bdf5b24df66925ffb6f5b5ced49d3e558d0ff096488dc4bab45c647f511ad90000000000000000000000000000000000000000000000000000000000000060a3710b802a5acc07abe5441f1ae1d2b30f4350a8d79fe634e2eee79ed3b3e2ef45622516f179c0c90774e6b12a66b8ff030de95507ecd047255fa8ebc931110d46a3530cf4b5d3b7b6bb048dd79c224ccdb0a398731edcecf68d9632cfdc35d3	0	3702503	2020-11-05 20:56:53	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb05b4fe1b803df50e7f275f399b629c77b98037492d35038121c48a258df5a8dec9532836a37e481d28623b7b5ba57a9	\\x00bdf5b24df66925ffb6f5b5ced49d3e558d0ff096488dc4bab45c647f511ad9	32000000000	\\xa3710b802a5acc07abe5441f1ae1d2b30f4350a8d79fe634e2eee79ed3b3e2ef45622516f179c0c90774e6b12a66b8ff030de95507ecd047255fa8ebc931110d46a3530cf4b5d3b7b6bb048dd79c224ccdb0a398731edcecf68d9632cfdc35d3	\\x1300000000000000	f	t
\\x64bc37dac9e0cefae031a30d3e50fec179fcd1c247cff79cdd93072bba7e41c5	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012032768c580fb07a4489b66dc243fc50fa740ec506e76e02d9f97ad7cb84be0c8a000000000000000000000000000000000000000000000000000000000000003090f0e853deef7f089213ee5083d716145a6390a7c87d635bbee808af93f75e87367c4372e01bf9643c1ce63b393420c900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000d35a6cb2e97151a50ad1fa054907e6869cb6b11429fcc7fb406ffb8eed82f90000000000000000000000000000000000000000000000000000000000000060a0d2ef56668674d707ab2cb865feb66e63fa613e0fa8189819b7cacb79085da711b992d54d00decf93bc6c620bb8cc691936ffcbf86440868fe9e2afc19f10709b3e02eedc362f217be147bc67b5337b3ecdda162a1ce1020067411794c555ee	2	3702503	2020-11-05 20:56:53	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x90f0e853deef7f089213ee5083d716145a6390a7c87d635bbee808af93f75e87367c4372e01bf9643c1ce63b393420c9	\\x00d35a6cb2e97151a50ad1fa054907e6869cb6b11429fcc7fb406ffb8eed82f9	32000000000	\\xa0d2ef56668674d707ab2cb865feb66e63fa613e0fa8189819b7cacb79085da711b992d54d00decf93bc6c620bb8cc691936ffcbf86440868fe9e2afc19f10709b3e02eedc362f217be147bc67b5337b3ecdda162a1ce1020067411794c555ee	\\x1500000000000000	f	t
\\x038b75056250f248141a15b9a1649ca95f8cfc8ab7fdff6ff75de2c7dba851f5	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001207e0a9e84b1059da4d1cab3afa4d60fa25c2a982de16df8b9e76e9d14ed525bd70000000000000000000000000000000000000000000000000000000000000030b740bda6f2a1103de870bed80c73f0f4f02f6950ec8b57822ce593acbea510992d4e65ee0d20f472a4daf91f56965dc5000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020007ed1c9c6b03b06c8b01f76b5aacaa95ce3c50ff7b681054544be5c45e30c0f0000000000000000000000000000000000000000000000000000000000000060838ca52dc7f49b59e69ced839da2af2a41bd1366b7b3bd4117eeaf7348a6173362f577b2e60d9920df78f11d1c1cfa4a0cedcb10e51af9c017bfee742d0cf6e201ef80067a0a734b5ccf909f8be741d16a30f333a84b1b935946eb47dcfd19e9	8	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb740bda6f2a1103de870bed80c73f0f4f02f6950ec8b57822ce593acbea510992d4e65ee0d20f472a4daf91f56965dc5	\\x007ed1c9c6b03b06c8b01f76b5aacaa95ce3c50ff7b681054544be5c45e30c0f	32000000000	\\x838ca52dc7f49b59e69ced839da2af2a41bd1366b7b3bd4117eeaf7348a6173362f577b2e60d9920df78f11d1c1cfa4a0cedcb10e51af9c017bfee742d0cf6e201ef80067a0a734b5ccf909f8be741d16a30f333a84b1b935946eb47dcfd19e9	\\x1d00000000000000	f	t
\\xebb9eb16d944d9520b6d6343a7cfdf99c27e2ab5eb505107b6613ab610ebe733	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001209bd2485e6e5148615b26c3fdadc66abda752634f11056294172c172cc86592e1000000000000000000000000000000000000000000000000000000000000003090aa1cffa2198638bf6a8ada0e8377e855ba225edab0cfe355bc1348e1cc1a4ccb0af4548abb52122d94f364bb7a84c3000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020005e3166ab4573556e13b12af19c0e3fb2326ede5f94a4b4ec318d4f68956a1300000000000000000000000000000000000000000000000000000000000000609579eeaff4067bc86130b5d3f3f67cb15e27f75200acb7a03fda445e11ebd4cc057733038c4c7de270973557a1553926124dd6d73fb62ad0e284edc961acbebb0d3a468861f7c1724d7fc1c78311006753b160dcb67f29cbe685581f38b6d03a	7	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x90aa1cffa2198638bf6a8ada0e8377e855ba225edab0cfe355bc1348e1cc1a4ccb0af4548abb52122d94f364bb7a84c3	\\x005e3166ab4573556e13b12af19c0e3fb2326ede5f94a4b4ec318d4f68956a13	32000000000	\\x9579eeaff4067bc86130b5d3f3f67cb15e27f75200acb7a03fda445e11ebd4cc057733038c4c7de270973557a1553926124dd6d73fb62ad0e284edc961acbebb0d3a468861f7c1724d7fc1c78311006753b160dcb67f29cbe685581f38b6d03a	\\x1c00000000000000	f	t
\\xad85e9a7fc8b968f86eab5b9368d4280f1eff04893c3ecf27ae43b25d620ef26	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001202e7b3ded2e4f7c4bb8402f4fb94135fc61a06a48292e8e60bc6024e2cfef7cb40000000000000000000000000000000000000000000000000000000000000030a4e6b4771ab3fa4e7c6730744957f5c437722e214d888f441620caaf750e9acd3c0ac65c158a9f2dd9e09e89a79c25af0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200049c3a5dd7dae42e82f6793aba7002fefe8ea4cd788af16ec200ac499a8dce1000000000000000000000000000000000000000000000000000000000000006086a38ede992c96a2d51d0ca6b0e783366562f78207cd7ff8288122a464eb78c48859371a518743e8e1510d85bf4ac1e71378854f3b6c4409120b71e637570c76829b39a855a63329a751ec8fa9116faeea4de4c371a4259323418288bee5bce2	6	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa4e6b4771ab3fa4e7c6730744957f5c437722e214d888f441620caaf750e9acd3c0ac65c158a9f2dd9e09e89a79c25af	\\x0049c3a5dd7dae42e82f6793aba7002fefe8ea4cd788af16ec200ac499a8dce1	32000000000	\\x86a38ede992c96a2d51d0ca6b0e783366562f78207cd7ff8288122a464eb78c48859371a518743e8e1510d85bf4ac1e71378854f3b6c4409120b71e637570c76829b39a855a63329a751ec8fa9116faeea4de4c371a4259323418288bee5bce2	\\x1b00000000000000	f	t
\\x35ab7e720029c20928a0586e20e1dae2b4dded6efebef5eec1418982debe8ac2	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120d9d84642010280b1608a7654adda985cae21d1bdfad7a739c7e327c0c4ffa8160000000000000000000000000000000000000000000000000000000000000030987f7e4af595a34439eb2f87721e8efdb60cbed30b3908960b37112aab6aad5aacce9f6be4df4646f8592f1e8e57392400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000d1d394bab365274d7c05994c9234cf7e707d3447b258ec009fd009bd9fc4b20000000000000000000000000000000000000000000000000000000000000060920d5f89967dd051e404037fc18379990636d60425746673ae9809e6177f05c45e70e8bd34a380ea47767fe9a44af531016a1190c669593800ad6ab6c46f840f02d461a6a172a74cbcca2075f0b68b0bbc7d20a3b5f707c85709707e4414de19	5	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x987f7e4af595a34439eb2f87721e8efdb60cbed30b3908960b37112aab6aad5aacce9f6be4df4646f8592f1e8e573924	\\x00d1d394bab365274d7c05994c9234cf7e707d3447b258ec009fd009bd9fc4b2	32000000000	\\x920d5f89967dd051e404037fc18379990636d60425746673ae9809e6177f05c45e70e8bd34a380ea47767fe9a44af531016a1190c669593800ad6ab6c46f840f02d461a6a172a74cbcca2075f0b68b0bbc7d20a3b5f707c85709707e4414de19	\\x1a00000000000000	f	t
\\x84d3278d41f6b8ccd028de582a43b1bcfef2ac04a42c35ef23bd3df72df2a8c4	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120e739fab83e5f7799b1c7b806a1b98847e3e2018240890381128e225575069f1c0000000000000000000000000000000000000000000000000000000000000030a239d39a222ccdcd839bddd9e9f1cb5a4d21c573961e113124f061f892d5ae373154a7a2f75b89a16a6a8377b7340577000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020004ecf641de477642dc37470a6ce15e638cb8611de1d25f4b26341077ab81bda0000000000000000000000000000000000000000000000000000000000000060a11e9d78bbc884778a32a42af3bb40d827adb968f55b5e2944db57d8d056eddb560ac6c1f2c2ad6e63f4a7a6a97eb3cc02b3cf19402e9fb5e2c68738576374f3289f977ac21cd54f1354d71303192c46ed421c2fa4e6ad478bd04868d83a2678	4	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa239d39a222ccdcd839bddd9e9f1cb5a4d21c573961e113124f061f892d5ae373154a7a2f75b89a16a6a8377b7340577	\\x004ecf641de477642dc37470a6ce15e638cb8611de1d25f4b26341077ab81bda	32000000000	\\xa11e9d78bbc884778a32a42af3bb40d827adb968f55b5e2944db57d8d056eddb560ac6c1f2c2ad6e63f4a7a6a97eb3cc02b3cf19402e9fb5e2c68738576374f3289f977ac21cd54f1354d71303192c46ed421c2fa4e6ad478bd04868d83a2678	\\x1900000000000000	f	t
\\xe265027a5a1d0126e4ca464787fe928f53a7a58a96c8f3e8cf3d8fc249b034c4	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001204eca807aa8c2baa6f2a9a8d976e5e34375b1a59ade956f5dc9aa5b2cc53bafd40000000000000000000000000000000000000000000000000000000000000030aec6abe9f5b90e7e27f1e1c6f497581968339f33eb3e70b2a561ac7b93259502dbc7d372e34f5f96f7e1087f9a999ad000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000d65aaac0e78311e6cf698b04b120a0a240fb55cc2d68455faf193e3db9ba33000000000000000000000000000000000000000000000000000000000000006089a025f41804430c4c9c98c456f435ad5059834edbdedbd65881e563cbbac999e0abe7ff96b46b4568feabda13d74b9d197155b3e1679c1287e06490850ce68c0d1ab93d836410a64f183b878966656cc0dcd610698bc6c02e0168b44124523a	2	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xaec6abe9f5b90e7e27f1e1c6f497581968339f33eb3e70b2a561ac7b93259502dbc7d372e34f5f96f7e1087f9a999ad0	\\x00d65aaac0e78311e6cf698b04b120a0a240fb55cc2d68455faf193e3db9ba33	32000000000	\\x89a025f41804430c4c9c98c456f435ad5059834edbdedbd65881e563cbbac999e0abe7ff96b46b4568feabda13d74b9d197155b3e1679c1287e06490850ce68c0d1ab93d836410a64f183b878966656cc0dcd610698bc6c02e0168b44124523a	\\x1700000000000000	f	t
\\x010a1f7f441c6eb6b8e359df70279ba78d795e43930b038e1270f95dcd211a89	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120fcbbc18b18fdcab02d59eac38703e37f3cc6d9d8b875b93130f73d34ed0d6bbe0000000000000000000000000000000000000000000000000000000000000030b0726c1fd5187cf9452384c0cf1a5bb5275539f42a1234770f640b8e1e5a65a42e12624e46009d9562b125ffb43c9d51000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020007a5842359d55625d38f6d7722240179ec8b183b92f079d36b2764b25cce3ee0000000000000000000000000000000000000000000000000000000000000060a3cd49837c5bb65c55d2827a5eddd4039455e253623930c94eda6707c48d3703a23a71bcdddfb70f0fdac35dfb404a7b0feb7e525865e17a3f64e78ae71b178cfd08320cc37b3ef2760b9c99d24beeee47b42ffca9c9b4dc209e9fb36d65ab4c	3	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb0726c1fd5187cf9452384c0cf1a5bb5275539f42a1234770f640b8e1e5a65a42e12624e46009d9562b125ffb43c9d51	\\x007a5842359d55625d38f6d7722240179ec8b183b92f079d36b2764b25cce3ee	32000000000	\\xa3cd49837c5bb65c55d2827a5eddd4039455e253623930c94eda6707c48d3703a23a71bcdddfb70f0fdac35dfb404a7b0feb7e525865e17a3f64e78ae71b178cfd08320cc37b3ef2760b9c99d24beeee47b42ffca9c9b4dc209e9fb36d65ab4c	\\x1800000000000000	f	t
\\xba0ff82706db01094a1d34c4d309c53d366fcb06b2519cbb2e01f798937df337	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001206a124f12035a91dfd00c35da95df60c2efc8217ccaee104b4181475d15cbfef3000000000000000000000000000000000000000000000000000000000000003094a1111c0eab1bb5666eacb9742efbef79042ca3ba384ecece98ca2027c2452a7dc0467c16e45ad9f7bc7c89b6ca873700000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000ecd9abb1a74651a9d23426c3c5619d7fd8c119c636a2120b8d1c12745f72cd0000000000000000000000000000000000000000000000000000000000000060a5adf6c12dfb27965ebaf98f225885c315e21d06e5e7bced093d6e02c1f20d983f7717caba3f744e8db0f6cbf00ae0f2054a1655271e413bb9a701ea3a6ce61a915624c62460a6296be4a26a788e20abc986529181b0853848b69692fb63c025	9	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x94a1111c0eab1bb5666eacb9742efbef79042ca3ba384ecece98ca2027c2452a7dc0467c16e45ad9f7bc7c89b6ca8737	\\x00ecd9abb1a74651a9d23426c3c5619d7fd8c119c636a2120b8d1c12745f72cd	32000000000	\\xa5adf6c12dfb27965ebaf98f225885c315e21d06e5e7bced093d6e02c1f20d983f7717caba3f744e8db0f6cbf00ae0f2054a1655271e413bb9a701ea3a6ce61a915624c62460a6296be4a26a788e20abc986529181b0853848b69692fb63c025	\\x1e00000000000000	f	t
\\x2b75d49fa8d72efa16ec0c5e344d1d4e7b1b7513564a9f7e3d4e8a6caeea0ddd	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120a6e1a371c8a131a531d228e1e5a70270de9f1bf46e1b15c458ed1a80c4cfe5370000000000000000000000000000000000000000000000000000000000000030a7068aa763b441ffaec30fa6c051046ef7f5c3ecb3a7d86eeac2a34f92a5e9e74a049184358c900f82b63dbd56bb5ba500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000b9ba4ff15dd8b0d0c57781724bc73ac19be0064dfe5720921b6cea20cf1e2b000000000000000000000000000000000000000000000000000000000000006091757e21810ee83cb63255afe60f98737f747a5abc796ba60ca4624c53c7e69b5321978eb1e83b735081d2a8f3702b0b0f4587b8c8a36e1923a5483dee54f86e3dbd9990da4f172cfe2a5715a0b5aaee6750469ba5307c2dd5b411b6486a2445	10	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa7068aa763b441ffaec30fa6c051046ef7f5c3ecb3a7d86eeac2a34f92a5e9e74a049184358c900f82b63dbd56bb5ba5	\\x00b9ba4ff15dd8b0d0c57781724bc73ac19be0064dfe5720921b6cea20cf1e2b	32000000000	\\x91757e21810ee83cb63255afe60f98737f747a5abc796ba60ca4624c53c7e69b5321978eb1e83b735081d2a8f3702b0b0f4587b8c8a36e1923a5483dee54f86e3dbd9990da4f172cfe2a5715a0b5aaee6750469ba5307c2dd5b411b6486a2445	\\x1f00000000000000	f	t
\\x4a368ed2171e9b8e065c1074459c912d1b1dbe004e4d0b6a496c27f2c1bc179b	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012094c5dfe570d910d9f50a0a69f8f4402698eeeeeb65c35ea9470302b5fcdda5ef000000000000000000000000000000000000000000000000000000000000003084861ddaa7690c07e1c97f5fcf5e2c681b1fcb34f052273f959de26e145d27a10ae77e33cfcc3cbe73b9920b27655e4d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000cde83b04f0b875474732fbba2cffc91c68bdd2deaed4891c157a3ad14b69c900000000000000000000000000000000000000000000000000000000000000608829b5bfe0aeda6e0c904af4cc4cf20d5b014d8e0afdda24e35bfe14bdab1f1ccfb2b90f6dbac97dbf82bb97695b52080df9668ecbc4748b0d17ade44dbfef8ff955ac9cfd6095a706bbe497f79bc64fab3dd54a076d9208a97ff9c5b3915cf6	0	3702504	2020-11-05 20:57:08	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x84861ddaa7690c07e1c97f5fcf5e2c681b1fcb34f052273f959de26e145d27a10ae77e33cfcc3cbe73b9920b27655e4d	\\x00cde83b04f0b875474732fbba2cffc91c68bdd2deaed4891c157a3ad14b69c9	32000000000	\\x8829b5bfe0aeda6e0c904af4cc4cf20d5b014d8e0afdda24e35bfe14bdab1f1ccfb2b90f6dbac97dbf82bb97695b52080df9668ecbc4748b0d17ade44dbfef8ff955ac9cfd6095a706bbe497f79bc64fab3dd54a076d9208a97ff9c5b3915cf6	\\x1600000000000000	f	t
\\x31d58188383dc6ef6a7946a982c99f7148aefdac1b48b8d1b74acfb08f522104	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120ef3babda616eeb6e310224c14b7009749e5a76c74191e3b1364b4ceb7e2f86ea0000000000000000000000000000000000000000000000000000000000000030b4eec498e4baf6f0db0987db72cd05776855e3267cdc2e35008bffe2759e9d86c6b67f1b8cbbef7c5a58031af4951b3500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000e1ae4d3509356c41a393c52511db9ddcac9b41c9da5ed3f11b39b0097638a600000000000000000000000000000000000000000000000000000000000000609706d0838d76436babc03ccf1102999e7e67acd3f1eea104e68a392f98ef6ba4b6f6cdcb766cd2751bbbdb43fdd4f35319219209ba066d84eab6000d85e7a232d29c2134a711f4794ee897fb923fcc1da3dc37084115c35366864dcf7244155f	20	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb4eec498e4baf6f0db0987db72cd05776855e3267cdc2e35008bffe2759e9d86c6b67f1b8cbbef7c5a58031af4951b35	\\x00e1ae4d3509356c41a393c52511db9ddcac9b41c9da5ed3f11b39b0097638a6	32000000000	\\x9706d0838d76436babc03ccf1102999e7e67acd3f1eea104e68a392f98ef6ba4b6f6cdcb766cd2751bbbdb43fdd4f35319219209ba066d84eab6000d85e7a232d29c2134a711f4794ee897fb923fcc1da3dc37084115c35366864dcf7244155f	\\x3100000000000000	f	t
\\x67f05d63ca79070b115a1f11c191415024084af7d920c234bfb57993b83996dd	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001207d552f1ee2c14a0d3472924d39e74f3c3797834323ceb1d98c348e4f3444462b0000000000000000000000000000000000000000000000000000000000000030a0491939d33b503684ce69b876979eb3e12d23957219c2e02c4af0767f7ffc03fcac36454ae1a9700eac3493d6c7180800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000b73c95fb6e50ad58904b149b89915f244554948e0c201524cc8a53615fcb67000000000000000000000000000000000000000000000000000000000000006085bcf976cdaa76ac9f3f6f73fafa7b28fdefec1fdfa913102fa05f8c2201b186cad1b6e82366c2ea0d9b13c1afc6460912683991128d9ddbb1084211ea0482742f9bc0fdc414f9c9ddcc2bb7e6b2e748c871ba399293e5431d72ee4281e97bcd	19	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa0491939d33b503684ce69b876979eb3e12d23957219c2e02c4af0767f7ffc03fcac36454ae1a9700eac3493d6c71808	\\x00b73c95fb6e50ad58904b149b89915f244554948e0c201524cc8a53615fcb67	32000000000	\\x85bcf976cdaa76ac9f3f6f73fafa7b28fdefec1fdfa913102fa05f8c2201b186cad1b6e82366c2ea0d9b13c1afc6460912683991128d9ddbb1084211ea0482742f9bc0fdc414f9c9ddcc2bb7e6b2e748c871ba399293e5431d72ee4281e97bcd	\\x3000000000000000	f	t
\\x7e9aedf54729f32e4e41bdf4ab4d468115719a5acff92ec205553593ba16faee	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001208d6fca26ec5cf0556c7b4988358a2a470fccf9c5beabf11595d6648369b78acb0000000000000000000000000000000000000000000000000000000000000030adf7ade854a3c6cb0e675db842cc52d7cddfa51a7bb5611bb302a1196471fe0e3ff8523901a9f19ed28cb3bee74253810000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200090df29ae1d3f18b94e8a250299a48a4bf1f7ae6b4c807222fc58517659522e00000000000000000000000000000000000000000000000000000000000000609104b7072baba2b7dc813d8dcc17661401ce07ad74fa936d52b102943943e918e6a91226475622b3d79c8968954672b017d91ac2d157749268993ca626f9922b1d2640f0c9fd6333c7403d3b1e736a2c45bb0ec93ae93c6ae55a97ba5087c5d0	18	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xadf7ade854a3c6cb0e675db842cc52d7cddfa51a7bb5611bb302a1196471fe0e3ff8523901a9f19ed28cb3bee7425381	\\x0090df29ae1d3f18b94e8a250299a48a4bf1f7ae6b4c807222fc58517659522e	32000000000	\\x9104b7072baba2b7dc813d8dcc17661401ce07ad74fa936d52b102943943e918e6a91226475622b3d79c8968954672b017d91ac2d157749268993ca626f9922b1d2640f0c9fd6333c7403d3b1e736a2c45bb0ec93ae93c6ae55a97ba5087c5d0	\\x2f00000000000000	f	t
\\xa2b69cf763434a995239d0e15e51745642a50550c29004db601091c2a9e9b348	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001205aaae81be7037c186b1a1d657af2c13db334d50b7d2b1832964cf3508af07877000000000000000000000000000000000000000000000000000000000000003091a43c9bc21dd27b15ae62017adb5a8c3e5224c5f07af6abd925fac2f8c978c81fd584639f4a7a88ad8f885b04aba63d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000aad32ab570fc0d91bea2b185691d3556a1004f7df1535b3d3d7a0ad8810e7e0000000000000000000000000000000000000000000000000000000000000060a4e7f1b75f29da00c0cd0e85190826a332335f63010b46e4a410506b5e1eab86e60c221a79ca9197ddf576fb8a22eeb8052b55f0aa864a341d1590f6200bfbff15f068cde40018e89f35b3c0d35c02eed22e5dab83d14284f73df32f16ad32f5	16	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x91a43c9bc21dd27b15ae62017adb5a8c3e5224c5f07af6abd925fac2f8c978c81fd584639f4a7a88ad8f885b04aba63d	\\x00aad32ab570fc0d91bea2b185691d3556a1004f7df1535b3d3d7a0ad8810e7e	32000000000	\\xa4e7f1b75f29da00c0cd0e85190826a332335f63010b46e4a410506b5e1eab86e60c221a79ca9197ddf576fb8a22eeb8052b55f0aa864a341d1590f6200bfbff15f068cde40018e89f35b3c0d35c02eed22e5dab83d14284f73df32f16ad32f5	\\x2d00000000000000	f	t
\\xd69c8974604050d6acdbd6315cdb455ab040d9742e284add11a5257360f5b882	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012070b1fac171a27f77f8aa98e9c1f59a009c3018998cc61e9f586ab453b0b3395d0000000000000000000000000000000000000000000000000000000000000030852ebcc4483656f0c93b98ce6ab2a79fc9f8c7d7c560c146d9fff3e0b1eabd26c9dab54cf74adbfd5cdaca70d8807b3700000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000995b3dda3ac3ea15f1ad4190045da7b9716527174bad2fd241c33d058c22150000000000000000000000000000000000000000000000000000000000000060a9470869ae2c6b0ed4cf7c3aa3f2bdb690c63a05b5683f2e69c19eaa00dbef7601f940370911fdf038b668412130515c1831ec3a74d72c4bc5b276130a256dd348fe5b832826e9bee959eb5f9f29c36b1e7141efd6e06d4e26726e4acb5656f8	15	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x852ebcc4483656f0c93b98ce6ab2a79fc9f8c7d7c560c146d9fff3e0b1eabd26c9dab54cf74adbfd5cdaca70d8807b37	\\x00995b3dda3ac3ea15f1ad4190045da7b9716527174bad2fd241c33d058c2215	32000000000	\\xa9470869ae2c6b0ed4cf7c3aa3f2bdb690c63a05b5683f2e69c19eaa00dbef7601f940370911fdf038b668412130515c1831ec3a74d72c4bc5b276130a256dd348fe5b832826e9bee959eb5f9f29c36b1e7141efd6e06d4e26726e4acb5656f8	\\x2c00000000000000	f	t
\\xb2ed932ea85d180273a334a2b7ec0da49078f72fd1881ef431169184c7a8529e	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001202e7b1ba7ff4f884173bb064920819f3f7c8e7b4ad6b8e49104202da33193ca510000000000000000000000000000000000000000000000000000000000000030b5db47c31dc8a52360beed1e0a97f285dc0c9f71dcdf499b8c385eb6689175886ddc36f0a9a2de58b70663b031769dfc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000877d03b97f89c8428f6f8297a7b5172be4a7cbf990a48e1b42eb26bb2e41b6000000000000000000000000000000000000000000000000000000000000006098a010875ab0282d881542c850b8b7169b962fea748fbf823b90bfa8d31d44496b3a1bbc2fb59e21dd9ec34573f7c76609a69f0e33b1f3354f55d266e27c6333956e331b988e7668eff1cc90e9cb5a06b96e60dcb987ede2cc32f2776241c78d	14	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb5db47c31dc8a52360beed1e0a97f285dc0c9f71dcdf499b8c385eb6689175886ddc36f0a9a2de58b70663b031769dfc	\\x00877d03b97f89c8428f6f8297a7b5172be4a7cbf990a48e1b42eb26bb2e41b6	32000000000	\\x98a010875ab0282d881542c850b8b7169b962fea748fbf823b90bfa8d31d44496b3a1bbc2fb59e21dd9ec34573f7c76609a69f0e33b1f3354f55d266e27c6333956e331b988e7668eff1cc90e9cb5a06b96e60dcb987ede2cc32f2776241c78d	\\x2b00000000000000	f	t
\\x163477d3703434c6e094bf67ed6d3a682f4f33d98a6b62e4546c9e33b3a160b5	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001208037666cc330d760b25b534754fab536550cc618e2b8c1af33016a007eb5d0df00000000000000000000000000000000000000000000000000000000000000309704a2c3ab29d5992aaf8e644e666176c6d51c0f959829c63526e6cce514836ff4e889d5f83538ac2a06a1338501e5fc0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200019edb47a56c7e6baa2596aebbe14b88f0a6be9c30ee766608e844629c630b90000000000000000000000000000000000000000000000000000000000000060a2d9156b57a0515012a98a965002261b6020088e5f6968da6588b362f624d1557643ae6aec3141ff22509d0696e2a85208f0c0e9c382216b666a19e1d796076f174164790e44906ffdde3a9de8d796e5f873700518bfa75544d2d70716aae330	13	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x9704a2c3ab29d5992aaf8e644e666176c6d51c0f959829c63526e6cce514836ff4e889d5f83538ac2a06a1338501e5fc	\\x0019edb47a56c7e6baa2596aebbe14b88f0a6be9c30ee766608e844629c630b9	32000000000	\\xa2d9156b57a0515012a98a965002261b6020088e5f6968da6588b362f624d1557643ae6aec3141ff22509d0696e2a85208f0c0e9c382216b666a19e1d796076f174164790e44906ffdde3a9de8d796e5f873700518bfa75544d2d70716aae330	\\x2a00000000000000	f	t
\\x4b2e930148114511a17f14a90516006580e3f831aa3717dd3ce9e8927c2f165c	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120f19665b295be1d9498cb2fb9cec192e6f5a36ab725bba967f83899415234dfcb0000000000000000000000000000000000000000000000000000000000000030b7747a6497f53d9b065673c314c40087ea30364d87cf7bba66d3168d6a373b9c760af02778145f75a421611250ab94d000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000b5e64eccecf4cfa15fc558f619a0226d91ae22eb6c9620403274440f81822200000000000000000000000000000000000000000000000000000000000000608bd4fdb6e74a4199b989f7e33f208322b7c5ff2cddf22755edc1daceadd05b0e1de3c99a8ebd2aefde44b528cd8ac27d09cfcf2b1db062e87330dfe8eca81efd9dcd7b9874078d8399c7afa16b473296108e6ce10ae21fb26dc60f357c533175	12	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb7747a6497f53d9b065673c314c40087ea30364d87cf7bba66d3168d6a373b9c760af02778145f75a421611250ab94d0	\\x00b5e64eccecf4cfa15fc558f619a0226d91ae22eb6c9620403274440f818222	32000000000	\\x8bd4fdb6e74a4199b989f7e33f208322b7c5ff2cddf22755edc1daceadd05b0e1de3c99a8ebd2aefde44b528cd8ac27d09cfcf2b1db062e87330dfe8eca81efd9dcd7b9874078d8399c7afa16b473296108e6ce10ae21fb26dc60f357c533175	\\x2900000000000000	f	t
\\x39bdfd210f1affaba293d42a549b8b925d06a71fe8a1dbfc26e7c5d7bcdb9862	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001207e0bdb4d2bd0a202640740ed6f409179746b4548640c59d8c5cea926762ea180000000000000000000000000000000000000000000000000000000000000003080371785befe39cc94cad0b5d7f6dacf5628b43754c1a8d26303be5604e522358c75347987f153e7885e8a9952c0f35400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000cd9d0954a4a90dcb90834faa9fb108ba401fb0dcc80032b9192fa0b0787753000000000000000000000000000000000000000000000000000000000000006081c21884cd0e9311704a77675987f0bf6583281c364c6797542a8dc924ffb5fa000bdd09c70f0339a93e1ee3345b3b991958eee9c3bd075985c3c95652fa734d5fd87a5dae2c61a443d70dd0599b244c3b1a5cc8598c85260dae2d91549bff86	11	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x80371785befe39cc94cad0b5d7f6dacf5628b43754c1a8d26303be5604e522358c75347987f153e7885e8a9952c0f354	\\x00cd9d0954a4a90dcb90834faa9fb108ba401fb0dcc80032b9192fa0b0787753	32000000000	\\x81c21884cd0e9311704a77675987f0bf6583281c364c6797542a8dc924ffb5fa000bdd09c70f0339a93e1ee3345b3b991958eee9c3bd075985c3c95652fa734d5fd87a5dae2c61a443d70dd0599b244c3b1a5cc8598c85260dae2d91549bff86	\\x2800000000000000	f	t
\\xf2a7eb492b704b4d80b2cb698e7313d6f1d46807c1a3b574af9ee8c03024fefb	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001209d67d00f8598d12d664153c68bd676c4e3df3d51968eede83edf2ade03733dcf000000000000000000000000000000000000000000000000000000000000003080c526e2bc1208f95cad144990c312f8493d223623a7adcf80e19c7943e107b6b3e62bb3acbb3ef3c32cb026b71e728c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000af9004ee5af2e07c23a6cc8490623689b838b1f9aa941c4ef874b38c1dd6430000000000000000000000000000000000000000000000000000000000000060b060cb3a63055407c29e3e2618cf11bb577fc1dbd63cb94f84cf3419f526bbab89ea84e0bb95a0328f21717c48b9d913174ed7c4b0d9383cf6a45fd86b585dab5cc6fcde69597876e8b4ed5c06070fdf823db9f3d88d1226814fe23f189c2100	9	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x80c526e2bc1208f95cad144990c312f8493d223623a7adcf80e19c7943e107b6b3e62bb3acbb3ef3c32cb026b71e728c	\\x00af9004ee5af2e07c23a6cc8490623689b838b1f9aa941c4ef874b38c1dd643	32000000000	\\xb060cb3a63055407c29e3e2618cf11bb577fc1dbd63cb94f84cf3419f526bbab89ea84e0bb95a0328f21717c48b9d913174ed7c4b0d9383cf6a45fd86b585dab5cc6fcde69597876e8b4ed5c06070fdf823db9f3d88d1226814fe23f189c2100	\\x2700000000000000	f	t
\\x616bc74eddbdc493ee5f6238c066530188ab623677cc9a224f6bc93c29441a10	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001208f4bdec6a803dbdb2488ed5355e7b9460338fe72a768009c79dc0512a027bc620000000000000000000000000000000000000000000000000000000000000030975c9686d05ee1b7b494fb432b1a1589e0001270bf51ecadb6aed8de8aae1f53ed0c8f7f1bf60fc2dde9ea7aed5f28fd000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020006ba6aae2712bb20e80f6bb4d40bfa129f588a3acd046e513cfc4c7d00982680000000000000000000000000000000000000000000000000000000000000060a2f81a6fac530e4ef8ed202131a5d3886b6b94c985f1c4b01ae7519db516dde675133c12ab844f6465521ee410ef561611fef272e2f31fa3215d37ed956215695637ae3fb371d203cb919494ed6411fa3418d6045785ab56e5963aa40266de09	8	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x975c9686d05ee1b7b494fb432b1a1589e0001270bf51ecadb6aed8de8aae1f53ed0c8f7f1bf60fc2dde9ea7aed5f28fd	\\x006ba6aae2712bb20e80f6bb4d40bfa129f588a3acd046e513cfc4c7d0098268	32000000000	\\xa2f81a6fac530e4ef8ed202131a5d3886b6b94c985f1c4b01ae7519db516dde675133c12ab844f6465521ee410ef561611fef272e2f31fa3215d37ed956215695637ae3fb371d203cb919494ed6411fa3418d6045785ab56e5963aa40266de09	\\x2600000000000000	f	t
\\x73c962709d8dfb909689bd7382556239efc8291efdd20619f058130a47aa83c4	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012075b43e440d40e6244df810faf15fb7995b38ce62856a38633aa120908b63d30b0000000000000000000000000000000000000000000000000000000000000030b2a530285fd462d4b3c22981942c147d22cb1cded09116e8dc060ccae29305ce3ded8e19243d475009b6e77ff830e12900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000a9db877f6ab66672e4d41449b534122289b1ba406ea409eac3afed9254bfd700000000000000000000000000000000000000000000000000000000000000608e91b1852d9116d069b8302f1ab9256062e03aa72daa24b57f2938bcf0071d3b2d1d2b5df56236fbdc268a6abb8d3a9918d1edda6e5ede8b7f46a43f53e214b603c1529df831ad0a7e0556c7309971c4d7ac3f4974259608e07cfa926165bcde	7	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb2a530285fd462d4b3c22981942c147d22cb1cded09116e8dc060ccae29305ce3ded8e19243d475009b6e77ff830e129	\\x00a9db877f6ab66672e4d41449b534122289b1ba406ea409eac3afed9254bfd7	32000000000	\\x8e91b1852d9116d069b8302f1ab9256062e03aa72daa24b57f2938bcf0071d3b2d1d2b5df56236fbdc268a6abb8d3a9918d1edda6e5ede8b7f46a43f53e214b603c1529df831ad0a7e0556c7309971c4d7ac3f4974259608e07cfa926165bcde	\\x2500000000000000	f	t
\\x03e675db84a9b4194c404e97af631c96cdd684fce3f9967300c013385f450779	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001202d6c67aa8c974a0a61d6e0e1411968e495398bcfe9e07c67e645098455603f0a0000000000000000000000000000000000000000000000000000000000000030939e5cb88723ab173d622b11f1197b0c3e1d2c5b0d32a53ec0bf08effd1131ce382239080334c5ccaff5ace2b34c56ff00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000955377a9bfb6016cb6b94b1a26d68b14cb3213dbd4718902d804c1a698f8960000000000000000000000000000000000000000000000000000000000000060971312cd73d19379a192828a588e7b7058f90b6949d8208c2dce7fbf050c02eccad5ff6c924295249bcf2045f36de9031591cc96dc8a4db9211b41e4295729faefcaeb6247d76bb3ff02c690376836e5ee65d9d0eea01e8423c4f5e781c16d32	6	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\x939e5cb88723ab173d622b11f1197b0c3e1d2c5b0d32a53ec0bf08effd1131ce382239080334c5ccaff5ace2b34c56ff	\\x00955377a9bfb6016cb6b94b1a26d68b14cb3213dbd4718902d804c1a698f896	32000000000	\\x971312cd73d19379a192828a588e7b7058f90b6949d8208c2dce7fbf050c02eccad5ff6c924295249bcf2045f36de9031591cc96dc8a4db9211b41e4295729faefcaeb6247d76bb3ff02c690376836e5ee65d9d0eea01e8423c4f5e781c16d32	\\x2400000000000000	f	t
\\x503f0a346cd47ceef588f6787e98f99e1601df237605e7596d6a818cbca17481	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120bf8e2fb23ea118b061b03cc5e15c84a5ba5265bb1142624a6cf16f42bcab64d70000000000000000000000000000000000000000000000000000000000000030a00798fc1a2671615704a2fb650d43368dee54f60e2c1fb19a8a75f7474702dfdde2f05d51667405675d04e59890d69000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000b7a80722aa25daca7def5eafb8e5f1559d52fa0c40f45d9d0072eb0313bba10000000000000000000000000000000000000000000000000000000000000060b9cd681ade324840de685e5fc5d09027d7f22f5bd89669c631b43241baa5a78916ed0b258a8600e8ed454d62a15d120a01ed42013a321e9e2ae90fbe2e68b0c486082403bd5b2de503256badffc9999d3a7078a60b88c416ef0ba4e74a3795bb	17	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa00798fc1a2671615704a2fb650d43368dee54f60e2c1fb19a8a75f7474702dfdde2f05d51667405675d04e59890d690	\\x00b7a80722aa25daca7def5eafb8e5f1559d52fa0c40f45d9d0072eb0313bba1	32000000000	\\xb9cd681ade324840de685e5fc5d09027d7f22f5bd89669c631b43241baa5a78916ed0b258a8600e8ed454d62a15d120a01ed42013a321e9e2ae90fbe2e68b0c486082403bd5b2de503256badffc9999d3a7078a60b88c416ef0ba4e74a3795bb	\\x2e00000000000000	f	t
\\xb521e92475cb3012ffc42f2386568f231d1be5c9557300081e31d446ad9364b5	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012018bd1f8e439967da8a3b844d4ef3e66d41d0b0bee649751b99c3f9a8185ef9af0000000000000000000000000000000000000000000000000000000000000030b043fb18bf5394897f5492bcf5760e33d27996b6a343bad70afd83c45c41dacc3edbdad3f7dc88a6c158a847856d790900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000f58e7183f20ae8e2a2013e1b9fade6e9bcbab479ae11d9aec2a596ad382b320000000000000000000000000000000000000000000000000000000000000060b7e35ddd2819a008d176a9b613056a629f11c6328b05058aa50c4a9a14fe78cd2338094c1c323aeb86efd6be0e7d4fee0c0268918ec7427bc3fc38043c8a8e97848727731f7ade4c8a5dc69fa44787609e206b30674c4d77111288f04f1df637	5	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb043fb18bf5394897f5492bcf5760e33d27996b6a343bad70afd83c45c41dacc3edbdad3f7dc88a6c158a847856d7909	\\x00f58e7183f20ae8e2a2013e1b9fade6e9bcbab479ae11d9aec2a596ad382b32	32000000000	\\xb7e35ddd2819a008d176a9b613056a629f11c6328b05058aa50c4a9a14fe78cd2338094c1c323aeb86efd6be0e7d4fee0c0268918ec7427bc3fc38043c8a8e97848727731f7ade4c8a5dc69fa44787609e206b30674c4d77111288f04f1df637	\\x2300000000000000	f	t
\\x6102ed38047c5c43f5290476f5cbbb7af3e03e10e9e42eadcb1179337d9680eb	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001207a38818db7e9e2b4a0aa08fc0859015324e120c4dea263d58e31c7faa34939f40000000000000000000000000000000000000000000000000000000000000030b424af8a10fda0eaec704f3a1b28fc66ea6d504007d3361fba95ec2f9a3e2113ace14770ab9cf5434b8081e588ad23940000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200040b77d6daa6dfa5f25a7a06a38a0571f307ebd78737bfe02700e9b4b78ebed0000000000000000000000000000000000000000000000000000000000000060a2f6d3e43f886cd798e71d0d65434b20435db3f833b1d3f5b9635b5bff7b99e84f54cf797333d90c484c1cf9af11669907985a1834be923ba88d6750f0a030cb08d6dcad2fdc1419175766708e41214f83af150be4fc611da5428e6a300a824a	4	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xb424af8a10fda0eaec704f3a1b28fc66ea6d504007d3361fba95ec2f9a3e2113ace14770ab9cf5434b8081e588ad2394	\\x0040b77d6daa6dfa5f25a7a06a38a0571f307ebd78737bfe02700e9b4b78ebed	32000000000	\\xa2f6d3e43f886cd798e71d0d65434b20435db3f833b1d3f5b9635b5bff7b99e84f54cf797333d90c484c1cf9af11669907985a1834be923ba88d6750f0a030cb08d6dcad2fdc1419175766708e41214f83af150be4fc611da5428e6a300a824a	\\x2200000000000000	f	t
\\xc846260a78a71c0f82c2150604755f796a60363f5f7c5e5fe10bceaea1fe8e0c	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000001201a309f056864c053847b222f4e40c502c716534aca0268abb9fca08a08079bb10000000000000000000000000000000000000000000000000000000000000030a8b2f27e1005db3b973b61d9bbccd7cea9fe0a299d804cba0014ee103a4a658f1c5ede680b4bf5ff01beb4055e1a7b2d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000811b6f0aed81bb67fd104358a8b1364fd4524c960352872d1b20f079d4cfce00000000000000000000000000000000000000000000000000000000000000608549775b73dd7dd4a8ccb6ae33b404c482252c6fb90be2e635b22fc900f924d856631b7a4b3234251ae4377574a3c9bc105d6c0da8f3cb68e89e2cd2abc2f6549184f3b3cc1a1772b6648dfd6def7f43b08e16b96522c7c05c371e34d8cc934b	3	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xa8b2f27e1005db3b973b61d9bbccd7cea9fe0a299d804cba0014ee103a4a658f1c5ede680b4bf5ff01beb4055e1a7b2d	\\x00811b6f0aed81bb67fd104358a8b1364fd4524c960352872d1b20f079d4cfce	32000000000	\\x8549775b73dd7dd4a8ccb6ae33b404c482252c6fb90be2e635b22fc900f924d856631b7a4b3234251ae4377574a3c9bc105d6c0da8f3cb68e89e2cd2abc2f6549184f3b3cc1a1772b6648dfd6def7f43b08e16b96522c7c05c371e34d8cc934b	\\x2100000000000000	f	t
\\x57758c775fd79223d318731546e2ce08bee7e494d0499f15d9cb5ca13c906613	\\x22895118000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000000000000000000000000000000000000000012087367a174b0243b0fbc4dd7d4ba7c981ade5927352c8ba62c07351e575624a110000000000000000000000000000000000000000000000000000000000000030addfc94c84125c2547cea7c83af983ac1e10738aa1bb6cff10ba0cca43e0e4b5f38c02803383a39fa21e25e619f610610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200076600df9c724e0204736cf71d7242412f8d27f55b0587eaf3ebb853b7fc4bb00000000000000000000000000000000000000000000000000000000000000608a8829f7309a499ba3abb0b1f402bcee02a273d7fbc46243cd4ef1179a88df448ab255a0df777547e6c5e15094f66ae004de58b845b919da8208abecc5c2db52c0a5f13648ce2259574aeeff00d847c792da73669c3388fdad85a28652c02283	2	3702505	2020-11-05 20:57:23	\\x28aa7d30eb27b8955930bee3bc72255ab6a574d9	\\xaddfc94c84125c2547cea7c83af983ac1e10738aa1bb6cff10ba0cca43e0e4b5f38c02803383a39fa21e25e619f61061	\\x0076600df9c724e0204736cf71d7242412f8d27f55b0587eaf3ebb853b7fc4bb	32000000000	\\x8a8829f7309a499ba3abb0b1f402bcee02a273d7fbc46243cd4ef1179a88df448ab255a0df777547e6c5e15094f66ae004de58b845b919da8208abecc5c2db52c0a5f13648ce2259574aeeff00d847c792da73669c3388fdad85a28652c02283	\\x2000000000000000	f	t
\.
COPY public.proposal_assignments FROM stdin;
0	0	0	1
0	9220	1	1
0	3112	2	2
0	16246	3	1
0	11263	4	1
0	12298	5	1
0	12602	6	1
0	15443	7	1
0	6487	8	1
0	769	9	1
0	7531	10	1
0	11466	11	1
0	15968	12	2
0	4793	13	2
0	9082	14	1
0	12790	15	1
0	12383	16	1
0	10328	17	1
0	8321	18	1
0	2741	19	2
0	8011	20	1
0	5726	21	2
0	6329	22	1
0	2375	23	2
0	5822	24	2
0	4406	25	2
0	13944	26	1
0	5417	27	2
0	3547	28	2
0	6757	29	1
0	14765	30	2
0	6905	31	1
1	1848	32	1
1	9743	33	1
1	13141	34	1
1	829	35	1
1	1948	36	1
1	15792	37	2
1	13039	38	1
1	7229	39	1
1	4909	40	1
1	2911	41	2
1	1010	42	1
1	14440	43	2
1	12792	44	1
1	4375	45	1
1	2402	46	2
1	2506	47	2
1	13645	48	1
1	3453	49	2
1	11616	50	1
1	18	51	1
1	2221	52	2
1	10242	53	1
1	9590	54	1
1	120	55	1
1	1016	56	1
1	7731	57	1
1	3049	58	2
1	5979	59	1
1	12009	60	1
1	5854	61	1
1	4372	62	1
1	12759	63	1
\.
COPY public.queue FROM stdin;
2020-11-16 16:00:00	0	0
2020-11-16 15:00:00	0	0
2020-11-16 14:00:00	0	0
2020-11-16 13:00:00	0	0
2020-11-16 12:00:00	0	0
2020-11-16 11:00:00	0	0
2020-11-16 10:00:00	0	0
2020-11-16 09:00:00	0	0
2020-11-16 08:00:00	0	0
2020-11-16 07:00:00	0	0
2020-11-16 06:00:00	0	0
2020-11-16 05:00:00	0	0
2020-11-16 04:00:00	0	0
2020-11-16 03:00:00	0	0
2020-11-16 02:00:00	0	0
2020-11-16 01:00:00	0	0
2020-11-16 00:00:00	0	0
2020-11-15 23:00:00	0	0
2020-11-15 22:00:00	0	0
2020-11-15 21:00:00	0	0
2020-11-15 20:00:00	0	0
2020-11-15 19:00:00	0	0
2020-11-15 18:00:00	0	0
2020-11-15 17:00:00	0	0
2020-11-15 16:00:00	0	0
2020-11-15 15:00:00	0	0
2020-11-15 14:00:00	0	0
2020-11-15 13:00:00	0	0
2020-11-15 12:00:00	0	0
2020-11-15 11:00:00	0	0
2020-11-15 10:00:00	0	0
2020-11-15 09:00:00	0	0
2020-11-15 08:00:00	0	0
2020-11-15 07:00:00	0	0
2020-11-15 06:00:00	0	0
2020-11-15 05:00:00	0	0
2020-11-15 04:00:00	0	0
2020-11-15 03:00:00	0	0
2020-11-15 02:00:00	0	0
2020-11-15 01:00:00	0	0
2020-11-15 00:00:00	0	0
2020-11-14 23:00:00	0	0
2020-11-14 22:00:00	0	0
2020-11-14 21:00:00	0	0
2020-11-14 20:00:00	0	0
2020-11-14 19:00:00	0	0
2020-11-14 18:00:00	0	0
2020-11-14 17:00:00	0	0
2020-11-14 16:00:00	0	0
2020-11-14 15:00:00	0	0
2020-11-14 14:00:00	0	0
2020-11-14 13:00:00	0	0
2020-11-14 12:00:00	0	0
2020-11-14 11:00:00	0	0
2020-11-14 10:00:00	0	0
2020-11-14 09:00:00	0	0
2020-11-14 08:00:00	0	0
2020-11-14 07:00:00	0	0
2020-11-14 06:00:00	0	0
2020-11-14 05:00:00	0	0
2020-11-14 04:00:00	0	0
2020-11-14 03:00:00	0	0
2020-11-14 02:00:00	0	0
2020-11-14 01:00:00	0	0
2020-11-14 00:00:00	0	0
2020-11-13 23:00:00	0	0
2020-11-13 22:00:00	0	0
2020-11-13 21:00:00	0	0
2020-11-13 20:00:00	0	0
2020-11-13 19:00:00	0	0
2020-11-13 18:00:00	0	0
2020-11-13 17:00:00	0	0
2020-11-13 16:00:00	0	0
2020-11-13 15:00:00	0	0
2020-11-13 14:00:00	0	0
2020-11-13 13:00:00	0	0
2020-11-13 12:00:00	0	0
2020-11-13 11:00:00	0	0
2020-11-13 10:00:00	0	0
2020-11-13 09:00:00	0	0
2020-11-13 08:00:00	0	0
2020-11-13 07:00:00	0	0
2020-11-13 06:00:00	0	0
2020-11-13 05:00:00	0	0
2020-11-13 04:00:00	0	0
2020-11-13 03:00:00	0	0
2020-11-13 02:00:00	0	0
2020-11-13 01:00:00	0	0
2020-11-13 00:00:00	0	0
2020-11-12 23:00:00	0	0
2020-11-12 22:00:00	0	0
2020-11-12 21:00:00	0	0
2020-11-12 20:00:00	0	0
2020-11-12 19:00:00	0	0
2020-11-12 18:00:00	0	0
2020-11-12 17:00:00	0	0
2020-11-12 16:00:00	0	0
2020-11-12 15:00:00	0	0
2020-11-12 14:00:00	0	0
2020-11-12 13:00:00	0	0
\.
COPY public.validator_balances FROM stdin;
0	0	32000000000	32000000000
0	1	32000000000	32000000000
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
COPY public.validator_performance FROM stdin;
6546	32122393887	22971494	122393887	122393887	122393887
10391	32116510620	18847054	116510620	116510620	116510620
11022	32116852944	18827708	116852944	116852944	116852944
13277	32111478309	16213825	111478309	111478309	111478309
13708	32113590114	16334459	113590114	113590114	113590114
855	32118291790	19049516	118291790	118291790	118291790
13792	32116347784	21736445	116347784	116347784	116347784
13	32119657047	19001210	119657047	119657047	119657047
1541	32116529777	18745481	116529777	116529777	116529777
8124	32117607867	18460483	117607867	117607867	117607867
11771	32118397666	18836042	118397666	118397666	118397666
11897	32115457138	18997843	115457138	115457138	115457138
15793	32117068125	21748049	117068125	117068125	117068125
7061	32115988717	18726029	115988717	115988717	115988717
7325	32120600149	22772660	120600149	120600149	120600149
8514	32123283135	19134478	123283135	123283135	123283135
8871	32126549072	21759311	126549072	126549072	126549072
10473	32124101235	19037032	124101235	124101235	124101235
3092	32116650099	18757771	116650099	116650099	116650099
3130	32116197817	19017312	116197817	116197817	116197817
13302	32113829114	18338176	113829114	113829114	113829114
13776	32117187366	16905803	117187366	117187366	117187366
2591	32119959864	20248855	119959864	119959864	119959864
7005	32118172690	18702176	118172690	118172690	118172690
7587	32118150389	20257681	118150389	118150389	118150389
10574	32116691522	18965182	116691522	116691522	116691522
14578	32115460454	20324760	115460454	115460454	115460454
10035	32125025768	20290658	125025768	125025768	125025768
14777	32121656901	18917626	121656901	121656901	121656901
586	32117029739	19099625	117029739	117029739	117029739
743	32120926850	20240315	120926850	120926850	120926850
1663	32120613744	18946147	120613744	120613744	120613744
6364	32115975364	20337239	115975364	115975364	115975364
7740	32118339236	20253877	118339236	118339236	118339236
2741	32116145194	18748789	116145194	116145194	116145194
4211	32117914194	19010767	117914194	117914194	117914194
4264	32114893921	19040628	114893921	114893921	114893921
7467	32115276788	18644550	115276788	115276788	115276788
3767	32114899449	18732052	114899449	114899449	114899449
8423	32122396298	22122975	122396298	122396298	122396298
12825	32115047035	18707629	115047035	115047035	115047035
2676	32116703501	18831697	116703501	116703501	116703501
15456	32119293689	18540944	119293689	119293689	119293689
9086	32123742124	18859783	123742124	123742124	123742124
11023	32122222214	18888924	122222214	122222214	122222214
13734	32111460096	17146363	111460096	111460096	111460096
53	32119563000	20373831	119563000	119563000	119563000
5004	32116741202	18898696	116741202	116741202	116741202
5287	32107570294	20549219	107570294	107570294	107570294
6898	32115435341	19986527	115435341	115435341	115435341
7066	32117123988	19106642	117123988	117123988	117123988
6530	32118288497	18876986	118288497	118288497	118288497
9879	32118380813	18878406	118380813	118380813	118380813
11864	32116903280	18923932	116903280	116903280	116903280
12963	32113156451	18830754	113156451	113156451	113156451
16352	32115345687	20172256	115345687	115345687	115345687
368	32122057413	18851875	122057413	122057413	122057413
1023	32117859135	18921622	117859135	117859135	117859135
4028	32118305888	18771546	118305888	118305888	118305888
8335	32119563757	18989983	119563757	119563757	119563757
14275	32118915551	17383610	118915551	118915551	118915551
13749	32113912058	16823510	113912058	113912058	113912058
14207	32112391215	17106221	112391215	112391215	112391215
3204	32120310836	18925305	120310836	120310836	120310836
4918	32111630899	18987336	111630899	111630899	111630899
4941	32116143412	18887360	116143412	116143412	116143412
6290	32117122028	20300422	117122028	117122028	117122028
9826	32124008034	18950110	124008034	124008034	124008034
935	32120845568	18901871	120845568	120845568	120845568
2413	32115978738	19030106	115978738	115978738	115978738
3565	32117977497	18280025	117977497	117977497	117977497
7831	32117594057	20186614	117594057	117594057	117594057
3403	32117119920	19036648	117119920	117119920	117119920
4344	32115046678	19081966	115046678	115046678	115046678
7787	32120403999	21726333	120403999	120403999	120403999
10630	32118141447	18934781	118141447	118141447	118141447
11459	32118437004	18839515	118437004	118437004	118437004
5279	32105329706	18777026	105329706	105329706	105329706
6803	32115934896	19992670	115934896	115934896	115934896
8263	32125140775	18862480	125140775	125140775	125140775
15542	32118939776	18967903	118939776	118939776	118939776
72	32123470562	18771480	123470562	123470562	123470562
2194	32078778361	20094993	78778361	78778361	78778361
4515	32105418446	19036563	105418446	105418446	105418446
6856	32123439915	20048288	123439915	123439915	123439915
14995	32115503518	20247065	115503518	115503518	115503518
34	32123539829	21665908	123539829	123539829	123539829
1896	32119399695	20384529	119399695	119399695	119399695
2019	32125133741	20414626	125133741	125133741	125133741
4296	32114062099	18984178	114062099	114062099	114062099
9958	32119544650	18947632	119544650	119544650	119544650
977	32118001332	20328066	118001332	118001332	118001332
1296	32119528812	19047432	119528812	119528812	119528812
4725	32107286558	18794959	107286558	107286558	107286558
12272	32118601649	18579288	118601649	118601649	118601649
13317	32112949102	16160308	112949102	112949102	112949102
3198	32123085896	19966624	123085896	123085896	123085896
4287	32116125812	20489426	116125812	116125812	116125812
6849	32119748984	20884304	119748984	119748984	119748984
9425	32117903243	18880285	117903243	117903243	117903243
\.
COPY public.validators FROM stdin;
2	\\x8bf10faad775298f089941b54072081933430077efa7edb8800a5ad67b910d1d4906a59aa09a2a82c4bbb2eb572ec41f	9223372036854775807	\\x006525c13a1be61a018ad639e181648d58a1647c16cb7eccf77213e0f8251797	32119833120	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
3	\\xa4e8f56301594c4ef7544eac805c2c701e62124f6bcc2ce39ae242fb34a01795c9f96f3d5667773b0329dfb9b5ec3e20	9223372036854775807	\\x0087b8efff1e5494b3f4a85ea9b82385671b4b2855457d5ea656f863da457f56	32126374074	32000000000	f	0	0	9223372036854775807	44608	proto_lighthouse
4	\\xb54268644802909635c3ac0630394c8608a705d335277a3a0e130b86d23300312afd7e5d52e4e44aaa67f2810bb1ea60	9223372036854775807	\\x004b211838ed05a579a285aed07ee9b5b353d6c3d67fd0aa29292dd4266346a7	32117028354	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
5	\\x887e202bf6cbd63da61fd95b3b42406d47f91bc159a70e6697701afdcd64193e5ebf91a526b027e004fad8c600cd818a	9223372036854775807	\\x00a54f28489d114c335d9208b9955e6c4b399287f23946a1e692580e2bcea72d	32118797730	32000000000	f	0	0	9223372036854775807	44622	proto_lighthouse
6	\\xb3cf4e3f286c3d07406154a509e12065adfe0b2decc00fc55de4adc77900feabe24e2ca095221834dfb9ef768c629d33	9223372036854775807	\\x00de417ba822a4d19ff89cceabbcfa5b88450019e1e16d37937fd32aa219b82d	32125536268	32000000000	f	0	0	9223372036854775807	44594	proto_lighthouse
7	\\xa32a6208028023dc022a896421cfae19d901aa49943b099499b252957819048525184085e3b650b5315804940b44598b	9223372036854775807	\\x009192f1645b0cb7e047b33b5ed1589d480c6d5680c3ea74eaae0c1650ae9654	32121076151	32000000000	f	0	0	9223372036854775807	44593	proto_lighthouse
8	\\x83562589ade60f551f6b1f19ead319bc6031dee66f46e173e05f1cb06d85e7e0b1960e8b263a7b96d15916e9806446f6	9223372036854775807	\\x00990af45751941807ac88896701b0fe33f62e87b02f8f4729d7b6e71689c5ef	32121131815	32000000000	f	0	0	9223372036854775807	44599	proto_lighthouse
9	\\x8bb4551148eba7433c302ea318e7c18ea44a1b8ebf9555447d03e0c80e77d3b745ae8f517eb954ca0dc52d46fc6296e0	9223372036854775807	\\x008763b26fd367d4ed8edc1bb58aa2838ff1045a1af82f851da9dc6e34b49f5c	32122282425	32000000000	f	0	0	9223372036854775807	44606	proto_lighthouse
10	\\xa9e7e48d0c3539f260c01cb93b6a2e9c7ec94d2b5a6134d4b09561413043a8e686e1fa2e545a981146e4686d99b8df3d	9223372036854775807	\\x00cc3fc841cd79655f08fba3c8a6524b0fa945447874a76b6de348349b9ab390	32122573905	32000000000	f	0	0	9223372036854775807	44597	proto_lighthouse
11	\\x86accb54629ceec44a4c06436ce639fb5d0ef4503abc0d3958a5007c6f3c4b2eb2b4f8155cdce8452f2f01d537c41328	9223372036854775807	\\x003e97a4e3d984d983c5023aba79e2abfb765fe58ab59cd55f85aa2ddb0f4b92	32121229636	32000000000	f	0	0	9223372036854775807	44617	proto_lighthouse
12	\\x9946c1c8da6d97b9387262f43762befdf0ba714c4b3884051411a0e7f18e5c5dbefb99099c737fe9f54f124d02ed1707	9223372036854775807	\\x0022d4d5f14bbf6cb0fa60f58e9befe6444de903821af0dd7b7e996b2da1e77e	32121032683	32000000000	f	0	0	9223372036854775807	44603	proto_lighthouse
13	\\x977c217ebe82ed92c00aa4f6f11adab18c6c585d1199612409ce8f24bf5cf665161dc5b3c991835f4d2d59c6a14ae218	9223372036854775807	\\x0077b4ecac02b3dc1d752dcb66c34ee675d719d29d3770ede0d7ecdc92b588ee	32120001174	32000000000	f	0	0	9223372036854775807	44617	proto_lighthouse
14	\\xabfb6896a5bf3105adb4e03ca49c487cd75f0cf7570c917a396fa8211e933fd9297fb1def46020a03e21c1956262563d	9223372036854775807	\\x00beb39f317480db5397db86a708d4606cbdc1d03767a5b28dd5ddfeeed7ff94	32119779045	32000000000	f	0	0	9223372036854775807	44611	proto_lighthouse
15	\\xb18943c3b4a71068a436e877bff53a8cd2de552708f9e8ebe81e1ebca8ed19f5f577044884edf4f700949b480e622cfe	9223372036854775807	\\x0012b57157e16f8fbbe5e18dae4a8bc7dac61a73998bdc1f6a51c9cd534bc8e2	32117001002	32000000000	f	0	0	9223372036854775807	44584	proto_lighthouse
16	\\x8b4ff884f0825d8368f4077ad53149759332554cbe3655c34518a30e75365857ae787b208301be502249bf4e06e03273	9223372036854775807	\\x00a78a401ca1d5c49c62aa99ee38e2052527a090723474492aa47f2361bd88e6	32118367599	32000000000	f	0	0	9223372036854775807	44623	proto_lighthouse
17	\\xb528a635883d81c4cd4b0af451765471114df70eae4c91d974a8b0dae9c470656fcdea4135b3c9013b601cd6463ba16f	9223372036854775807	\\x0086f4ae865356daf55112079b84ba696d1cdacaa4e47237a8f780a1a4d9f4d9	32119625392	32000000000	f	0	0	9223372036854775807	44623	proto_lighthouse
18	\\xae94ec1191683f51784b6e6201edac5a6d1365e97271cb6ab048d722c770de76eae0f901d83ea4ef1e1ed84cd3c80aa1	9223372036854775807	\\x0047d061828d0a5139c5e1aad92b58b7b0bf8e836f16c36322e95c7dc90a4d40	32122368065	32000000000	f	0	0	9223372036854775807	44604	proto_lighthouse
19	\\xb05b4fe1b803df50e7f275f399b629c77b98037492d35038121c48a258df5a8dec9532836a37e481d28623b7b5ba57a9	9223372036854775807	\\x00bdf5b24df66925ffb6f5b5ced49d3e558d0ff096488dc4bab45c647f511ad9	32119212602	32000000000	f	0	0	9223372036854775807	44591	proto_lighthouse
20	\\x98bf2d673a152213fff14f66827978aaa04a5a1f0fcbd95fe6fd521af08e470ff74d1b6667ed1a607dbb6b597ae7c573	9223372036854775807	\\x00d2d24d652c1ded1c9453b663ea35d2b1964b5581fb78a1fd745c49659f9276	32121149864	32000000000	f	0	0	9223372036854775807	44589	proto_lighthouse
21	\\x90f0e853deef7f089213ee5083d716145a6390a7c87d635bbee808af93f75e87367c4372e01bf9643c1ce63b393420c9	9223372036854775807	\\x00d35a6cb2e97151a50ad1fa054907e6869cb6b11429fcc7fb406ffb8eed82f9	32118123458	32000000000	f	0	0	9223372036854775807	44586	proto_lighthouse
22	\\x84861ddaa7690c07e1c97f5fcf5e2c681b1fcb34f052273f959de26e145d27a10ae77e33cfcc3cbe73b9920b27655e4d	9223372036854775807	\\x00cde83b04f0b875474732fbba2cffc91c68bdd2deaed4891c157a3ad14b69c9	32119909669	32000000000	f	0	0	9223372036854775807	44610	proto_lighthouse
23	\\xaec6abe9f5b90e7e27f1e1c6f497581968339f33eb3e70b2a561ac7b93259502dbc7d372e34f5f96f7e1087f9a999ad0	9223372036854775807	\\x00d65aaac0e78311e6cf698b04b120a0a240fb55cc2d68455faf193e3db9ba33	32122207927	32000000000	f	0	0	9223372036854775807	44610	proto_lighthouse
24	\\xb0726c1fd5187cf9452384c0cf1a5bb5275539f42a1234770f640b8e1e5a65a42e12624e46009d9562b125ffb43c9d51	9223372036854775807	\\x007a5842359d55625d38f6d7722240179ec8b183b92f079d36b2764b25cce3ee	32122535014	32000000000	f	0	0	9223372036854775807	44619	proto_lighthouse
25	\\xa239d39a222ccdcd839bddd9e9f1cb5a4d21c573961e113124f061f892d5ae373154a7a2f75b89a16a6a8377b7340577	9223372036854775807	\\x004ecf641de477642dc37470a6ce15e638cb8611de1d25f4b26341077ab81bda	32119926039	32000000000	f	0	0	9223372036854775807	44603	proto_lighthouse
26	\\x987f7e4af595a34439eb2f87721e8efdb60cbed30b3908960b37112aab6aad5aacce9f6be4df4646f8592f1e8e573924	9223372036854775807	\\x00d1d394bab365274d7c05994c9234cf7e707d3447b258ec009fd009bd9fc4b2	32128165731	32000000000	f	0	0	9223372036854775807	44581	proto_lighthouse
27	\\xa4e6b4771ab3fa4e7c6730744957f5c437722e214d888f441620caaf750e9acd3c0ac65c158a9f2dd9e09e89a79c25af	9223372036854775807	\\x0049c3a5dd7dae42e82f6793aba7002fefe8ea4cd788af16ec200ac499a8dce1	32122508434	32000000000	f	0	0	9223372036854775807	44581	proto_lighthouse
28	\\x90aa1cffa2198638bf6a8ada0e8377e855ba225edab0cfe355bc1348e1cc1a4ccb0af4548abb52122d94f364bb7a84c3	9223372036854775807	\\x005e3166ab4573556e13b12af19c0e3fb2326ede5f94a4b4ec318d4f68956a13	32120012648	32000000000	f	0	0	9223372036854775807	44618	proto_lighthouse
29	\\xb740bda6f2a1103de870bed80c73f0f4f02f6950ec8b57822ce593acbea510992d4e65ee0d20f472a4daf91f56965dc5	9223372036854775807	\\x007ed1c9c6b03b06c8b01f76b5aacaa95ce3c50ff7b681054544be5c45e30c0f	32123372909	32000000000	f	0	0	9223372036854775807	44617	proto_lighthouse
30	\\x94a1111c0eab1bb5666eacb9742efbef79042ca3ba384ecece98ca2027c2452a7dc0467c16e45ad9f7bc7c89b6ca8737	9223372036854775807	\\x00ecd9abb1a74651a9d23426c3c5619d7fd8c119c636a2120b8d1c12745f72cd	32119708886	32000000000	f	0	0	9223372036854775807	44595	proto_lighthouse
31	\\xa7068aa763b441ffaec30fa6c051046ef7f5c3ecb3a7d86eeac2a34f92a5e9e74a049184358c900f82b63dbd56bb5ba5	9223372036854775807	\\x00b9ba4ff15dd8b0d0c57781724bc73ac19be0064dfe5720921b6cea20cf1e2b	32118444364	32000000000	f	0	0	9223372036854775807	44592	proto_lighthouse
32	\\xaddfc94c84125c2547cea7c83af983ac1e10738aa1bb6cff10ba0cca43e0e4b5f38c02803383a39fa21e25e619f61061	9223372036854775807	\\x0076600df9c724e0204736cf71d7242412f8d27f55b0587eaf3ebb853b7fc4bb	32122550612	32000000000	f	0	0	9223372036854775807	44595	proto_lighthouse
33	\\xa8b2f27e1005db3b973b61d9bbccd7cea9fe0a299d804cba0014ee103a4a658f1c5ede680b4bf5ff01beb4055e1a7b2d	9223372036854775807	\\x00811b6f0aed81bb67fd104358a8b1364fd4524c960352872d1b20f079d4cfce	32118429567	32000000000	f	0	0	9223372036854775807	44598	proto_lighthouse
34	\\xb424af8a10fda0eaec704f3a1b28fc66ea6d504007d3361fba95ec2f9a3e2113ace14770ab9cf5434b8081e588ad2394	9223372036854775807	\\x0040b77d6daa6dfa5f25a7a06a38a0571f307ebd78737bfe02700e9b4b78ebed	32123883956	32000000000	f	0	0	9223372036854775807	44576	proto_lighthouse
35	\\xb043fb18bf5394897f5492bcf5760e33d27996b6a343bad70afd83c45c41dacc3edbdad3f7dc88a6c158a847856d7909	9223372036854775807	\\x00f58e7183f20ae8e2a2013e1b9fade6e9bcbab479ae11d9aec2a596ad382b32	32124119065	32000000000	f	0	0	9223372036854775807	44607	proto_lighthouse
36	\\x939e5cb88723ab173d622b11f1197b0c3e1d2c5b0d32a53ec0bf08effd1131ce382239080334c5ccaff5ace2b34c56ff	9223372036854775807	\\x00955377a9bfb6016cb6b94b1a26d68b14cb3213dbd4718902d804c1a698f896	32121319413	32000000000	f	0	0	9223372036854775807	44605	proto_lighthouse
37	\\xb2a530285fd462d4b3c22981942c147d22cb1cded09116e8dc060ccae29305ce3ded8e19243d475009b6e77ff830e129	9223372036854775807	\\x00a9db877f6ab66672e4d41449b534122289b1ba406ea409eac3afed9254bfd7	32122510420	32000000000	f	0	0	9223372036854775807	44615	proto_lighthouse
38	\\x975c9686d05ee1b7b494fb432b1a1589e0001270bf51ecadb6aed8de8aae1f53ed0c8f7f1bf60fc2dde9ea7aed5f28fd	9223372036854775807	\\x006ba6aae2712bb20e80f6bb4d40bfa129f588a3acd046e513cfc4c7d0098268	32120209451	32000000000	f	0	0	9223372036854775807	44611	proto_lighthouse
39	\\x80c526e2bc1208f95cad144990c312f8493d223623a7adcf80e19c7943e107b6b3e62bb3acbb3ef3c32cb026b71e728c	9223372036854775807	\\x00af9004ee5af2e07c23a6cc8490623689b838b1f9aa941c4ef874b38c1dd643	32125103192	32000000000	f	0	0	9223372036854775807	44570	proto_lighthouse
40	\\x80371785befe39cc94cad0b5d7f6dacf5628b43754c1a8d26303be5604e522358c75347987f153e7885e8a9952c0f354	9223372036854775807	\\x00cd9d0954a4a90dcb90834faa9fb108ba401fb0dcc80032b9192fa0b0787753	32118154228	32000000000	f	0	0	9223372036854775807	44587	proto_lighthouse
41	\\xb7747a6497f53d9b065673c314c40087ea30364d87cf7bba66d3168d6a373b9c760af02778145f75a421611250ab94d0	9223372036854775807	\\x00b5e64eccecf4cfa15fc558f619a0226d91ae22eb6c9620403274440f818222	32119775326	32000000000	f	0	0	9223372036854775807	44615	proto_lighthouse
42	\\x9704a2c3ab29d5992aaf8e644e666176c6d51c0f959829c63526e6cce514836ff4e889d5f83538ac2a06a1338501e5fc	9223372036854775807	\\x0019edb47a56c7e6baa2596aebbe14b88f0a6be9c30ee766608e844629c630b9	32122760767	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
43	\\xb5db47c31dc8a52360beed1e0a97f285dc0c9f71dcdf499b8c385eb6689175886ddc36f0a9a2de58b70663b031769dfc	9223372036854775807	\\x00877d03b97f89c8428f6f8297a7b5172be4a7cbf990a48e1b42eb26bb2e41b6	32122720241	32000000000	f	0	0	9223372036854775807	44622	proto_lighthouse
44	\\x852ebcc4483656f0c93b98ce6ab2a79fc9f8c7d7c560c146d9fff3e0b1eabd26c9dab54cf74adbfd5cdaca70d8807b37	9223372036854775807	\\x00995b3dda3ac3ea15f1ad4190045da7b9716527174bad2fd241c33d058c2215	32124984575	32000000000	f	0	0	9223372036854775807	44580	proto_lighthouse
45	\\x91a43c9bc21dd27b15ae62017adb5a8c3e5224c5f07af6abd925fac2f8c978c81fd584639f4a7a88ad8f885b04aba63d	9223372036854775807	\\x00aad32ab570fc0d91bea2b185691d3556a1004f7df1535b3d3d7a0ad8810e7e	32122573111	32000000000	f	0	0	9223372036854775807	44603	proto_lighthouse
46	\\xa00798fc1a2671615704a2fb650d43368dee54f60e2c1fb19a8a75f7474702dfdde2f05d51667405675d04e59890d690	9223372036854775807	\\x00b7a80722aa25daca7def5eafb8e5f1559d52fa0c40f45d9d0072eb0313bba1	32119556541	32000000000	f	0	0	9223372036854775807	44576	proto_lighthouse
47	\\xadf7ade854a3c6cb0e675db842cc52d7cddfa51a7bb5611bb302a1196471fe0e3ff8523901a9f19ed28cb3bee7425381	9223372036854775807	\\x0090df29ae1d3f18b94e8a250299a48a4bf1f7ae6b4c807222fc58517659522e	32119396408	32000000000	f	0	0	9223372036854775807	44589	proto_lighthouse
48	\\xa0491939d33b503684ce69b876979eb3e12d23957219c2e02c4af0767f7ffc03fcac36454ae1a9700eac3493d6c71808	9223372036854775807	\\x00b73c95fb6e50ad58904b149b89915f244554948e0c201524cc8a53615fcb67	32122961553	32000000000	f	0	0	9223372036854775807	44612	proto_lighthouse
49	\\xb4eec498e4baf6f0db0987db72cd05776855e3267cdc2e35008bffe2759e9d86c6b67f1b8cbbef7c5a58031af4951b35	9223372036854775807	\\x00e1ae4d3509356c41a393c52511db9ddcac9b41c9da5ed3f11b39b0097638a6	32120155169	32000000000	f	0	0	9223372036854775807	44614	proto_lighthouse
50	\\xa7aa7c0d2209e0483499d12ac4e236cd77a0583ab3686399bf0177b5472d53b88ee877c09eb91aa76d3ecba124264b2c	9223372036854775807	\\x00b00c0f5336621d5c8e8ab416195bd110cea4e7cffca6b4cacf34d626736de9	32123585993	32000000000	f	0	0	9223372036854775807	44597	proto_lighthouse
51	\\x8e4326fede8adecf5c265479a0c8b5b6ef2949f9df9f2fb0804be6beb05f7a1337027578c161b709a7880ad6f1fc44c3	9223372036854775807	\\x0041a1e5b4e0834ad4f34962aa38c0c5fc830dbc5b69b2ad6a68f802ee155bdd	32117195018	32000000000	f	0	0	9223372036854775807	44579	proto_lighthouse
52	\\x910558df0bb4bbf30d47fd565376377c22749f0ce215cb1c37ec06cfdf4656f2a2a4eaa5c69c1dcf103b672f397eb29b	9223372036854775807	\\x00fc904f2401a9d0fa30e0cd7f5f7ff87dfac17278d961ef53de652fd9017a98	32128687720	32000000000	f	0	0	9223372036854775807	44584	proto_lighthouse
53	\\x8c2e89ac29ab28d1f9ad9101741f4a0e02bf02abcb8e4d922d2dc534f5a4ec0769d202c92a7d2b3d505f4f26fe1568ec	9223372036854775807	\\x0033a6ca2c4f9b7f9d2199ce43c500d001182eba10d27c9a6c9664ba70707737	32119907127	32000000000	f	0	0	9223372036854775807	44623	proto_lighthouse
54	\\xb1c7c48d98063b73dd69af7500e8a293107ddd326ccfd0f115c6c69281ea539a4703519f9d79936c60d6565f127b255b	9223372036854775807	\\x007de621de43d3af51fc2b190b67a619d669b36477908f5611c9dbf5866ec83a	32122699824	32000000000	f	0	0	9223372036854775807	44581	proto_lighthouse
55	\\xb8adde3b7c6b894ec8ea3cceedc0d84f8502825a6d681417f18a52f13445aa79d16d38463950c535ba196fca21bad4d6	9223372036854775807	\\x005c935f8eb0985b94d49dabf03a9acd77b8d0efdc1558a285cf259e2946c3c9	32121659802	32000000000	f	0	0	9223372036854775807	44608	proto_lighthouse
56	\\x826b214e8b2f90149b28ad26cca709ab5fb09edcf573aa0041df1950add3509abf1a3af2632cb824eea3327e7e3ac25e	9223372036854775807	\\x007a33cf3a3475670605abcd92410357b74d8d8e65c7eb878fcbeec85ea9acf4	32119670496	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
57	\\xa0d9aaa26584dee9be29bc2684d9c565a2e2f807b6d82c72400925e12b98904f8e75337cb24651bae306e7e03a89830c	9223372036854775807	\\x0082845d3cb358397ea932ae03c7f017a1db744c8286acba5bb417ad67483475	32122786532	32000000000	f	0	0	9223372036854775807	44589	proto_lighthouse
58	\\x93ceb84f34c3ac783537916014a14b81d5e4aa674076162d87dac6d058bd21e58fb934b90d73f270f7884f31cfabeed3	9223372036854775807	\\x00463b05ce8dc11e89a628b59777f789bc7225edf20c940914fa9a1bca6a3ff9	32122696098	32000000000	f	0	0	9223372036854775807	44591	proto_lighthouse
59	\\xa0c75d5c10aa0ce0369e84d99ae186ff4ca5338fb0a9221d73c46a544e61b450b22d0d9dddcae11a16456e3d58e02928	9223372036854775807	\\x001d6fb1b18623d209cc323cf056b284ca2ddb9c9d759801968f20bfff3a2410	32119646118	32000000000	f	0	0	9223372036854775807	44596	proto_lighthouse
60	\\x938b0df86bd30d3569067008690fca01cad17286ffbd3c972c32310872b41de1f76c44e13e23b8939c6633a934d4506f	9223372036854775807	\\x0089a43556070f645f6fbe08756d575453bec54cecea016c9d92153a980de92e	32118356743	32000000000	f	0	0	9223372036854775807	44618	proto_lighthouse
61	\\x83130ed10c6bc79f07d8c4d73f0a5bf6407b1ecf47f7d760f1fcf746bc2cba855f9160adc948e93954ce4752c2a82041	9223372036854775807	\\x0053a23df1b77e47ab234fd2559e9e7ff77a9d120a5575b8625a246d97521e2f	32125426679	32000000000	f	0	0	9223372036854775807	44596	proto_lighthouse
62	\\x8c8e9dee0d5a334640ae786294322a11542f9f2aa967be42a12eebc7957b8f8fbb873be54cd50fb9aac7c8b58889a686	9223372036854775807	\\x003b32f6bb6f3a76cdd4dda57a6757ea7c98120f0877e17748d1fa1efbd54e68	32122097123	32000000000	f	0	0	9223372036854775807	44582	proto_lighthouse
63	\\x8d9b9b5eb57c035d1c566c968fcabb371de823d72ebae21a03462f3c6f4faed1bf93355837cffd94ef0deea8e95c78bb	9223372036854775807	\\x003642c3d4ed9cea0dbfe4bd31c26a77af14952517c0ff4bef12eab2f841b622	32120829005	32000000000	f	0	0	9223372036854775807	44610	proto_lighthouse
64	\\x82e1c3754bae5e17a5707acc4b299804d567b50a8509018fd4cf0f4c9872056934cc158ce658947f0d8ad4869e3e2acd	9223372036854775807	\\x00ce17b843578be8d0cc53c1771da3b4e609e22965ba102b3d891b5d61948b6e	32122849890	32000000000	f	0	0	9223372036854775807	44588	proto_lighthouse
65	\\xb80a1ade6ef3cf2931b14773161aa2174c96ff8f88b5aaec80b863d321306bfbd3873f9beba785b4c6d50d5d621d946a	9223372036854775807	\\x009ecb1fd0b47f571eab96157b96cc290acb20c8dab6e2cf8c0d4f11b079749d	32116791021	32000000000	f	0	0	9223372036854775807	44590	proto_lighthouse
66	\\xb8898dca419f19338fa0e527a5085bef0f944ef066df3a5f1f61fb7bd3c759d17888d01923fd023bbf5591b3ded2ff62	9223372036854775807	\\x006493022cb4582b6ae5e1525c8fd404f412d4fe0a0424777e03d847d530bcb3	32123817498	32000000000	f	0	0	9223372036854775807	44587	proto_lighthouse
67	\\xb30a4b31bdd96a9b8b9283174e4cc1513f4e1bbb9f4d2d551db6426850b3ea909090ffc18f411f94a6e2f19a50f36d51	9223372036854775807	\\x0031747a6f2ab541711ffba739171aac3eb49f2c9edda048a4949e53c3c75898	32121354547	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
68	\\x8f20450c0a7db2d0e7123248ec5f16b9f417c38f0f6f868793c7d0c3b672fb14e78aa3fe5c0b3f75ee4f1b19630a5dd7	9223372036854775807	\\x00797ebc5637771f0ffbba5de132e7e4547575fd67bc820256e37cce6a32b83e	32122395366	32000000000	f	0	0	9223372036854775807	44607	proto_lighthouse
69	\\xb4ff5e68271b333325d7b2bb8123577e63cdb6dfa4a5758613feb6b80716ae10c083e199db2f4ca379ab87523b817c74	9223372036854775807	\\x0046ebb6fdb2c4f631d55774b300fb35df777d6ac9c0e6c685dc8a650f03ca8d	32124477878	32000000000	f	0	0	9223372036854775807	44618	proto_lighthouse
70	\\xb879a89cf9b9c87dd0e9775b16cb4e977ededf9b3743a0639a7c6c54049a26d9475b29ab3f16c88d813cc2498063cfa5	9223372036854775807	\\x006698d3a399132c05b7fc610c7013673414e79ded961415be6cbd6c14fcd94d	32124000737	32000000000	f	0	0	9223372036854775807	44622	proto_lighthouse
71	\\x844ec23c577acedb9fec25c92c153ca05a64cfe70ba72627915f7377b3d28d6290b6f909e47e4c97abb8df6e97250fa9	9223372036854775807	\\x0087b400cf6b1eeed04302e58404bcabd75fc62eddebf0734a1f04e4811b77ef	32119654550	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
72	\\xb2aa98650e2fc5d3cedf36ff5a9a1539f6279b397661e39e4f82db203f6ecce9a942303e2ed68168dc34b378b8a5feb4	9223372036854775807	\\x00b9b1fc54a348c7d47023bb44c2b6d8f10790177818105773dc23854b44988c	32123814689	32000000000	f	0	0	9223372036854775807	44620	proto_lighthouse
73	\\xb16cf4aba3cd4a1935c2f9c94c785e4c9ad4cdb4062a63f249fe56f67392d20be214f3a583cba22868bb971c7ee2a1cb	9223372036854775807	\\x00762d7145c6391cdcc5e2d05d5b19c44b12b21df2122ecc7f184f6c9c888a41	32123443172	32000000000	f	0	0	9223372036854775807	44605	proto_lighthouse
74	\\xb24178a145b8b49c7f3e84f32e3304cfdf6fa41d1121a362ccd032fcda382994ae24901e999f9f6c16a938631e7d3e43	9223372036854775807	\\x000556ddd74e20d22b0c9931e61146e494796dfcc2f10c23f559247a99b832fe	32120177367	32000000000	f	0	0	9223372036854775807	44614	proto_lighthouse
75	\\xa0cc44291c1fa6de9bdffb6f687c137c72cab0abd6b3281c88050cb478e394a18196b2dc44f64fd0866ac1f4e4ebed21	9223372036854775807	\\x00e6e0a2fc50526b405d5c06eb63d0e3d2464f5116afae90cec9fa030dd8ddb5	32119821852	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
76	\\x8ace11b7126a6812188465fad6588854aee300a35ed948e51f09b65f72d162c3fd41b2aa3fda15a731d9152a1a88459c	9223372036854775807	\\x008f5fe72525ff5cf659a7e706823b41a0bd6a4f1b1537c243b2d49f1014f11d	32118373511	32000000000	f	0	0	9223372036854775807	44623	proto_lighthouse
77	\\xb1f80875d11f58324df319ce9674f1b84ffb06a7a380b9e110743fe3ab13a96160658baca5cae2a022439c2579a6a22b	9223372036854775807	\\x003a237b693b7971b691fdbed59a20d669b7db4ac044e3a4df91ed2ed08db62b	32124053590	32000000000	f	0	0	9223372036854775807	44584	proto_lighthouse
78	\\x94caa466dd3d9399cd597af7ae2fb944f1d27fa619d7a3b7c852ecda4f2489c9e6ed3c74abe02089452a90882d212b88	9223372036854775807	\\x002c1e1ed20f6e28c1897d1e484d6613845b25253482a6d3fcdab7fce7b1d90a	32120280530	32000000000	f	0	0	9223372036854775807	44595	proto_lighthouse
79	\\x82c461c6b1e311571fd67c16e87d78d1d2f28455fc34e00c6300d70dc5e8b22752c76306b2ec2bf8e0aa96053b8ce2b3	9223372036854775807	\\x00cdee43bbba59c4ab8a2f916b7048de2b8980cc86e3c3ce8ad552766313f584	32126975795	32000000000	f	0	0	9223372036854775807	44623	proto_lighthouse
80	\\x84ed2a69fff2a9f0081d01d6b2bfef22ad3d27c0edf21570e93ffa0d754913f6b4749cea76542f38bc0df95a4b90d1b8	9223372036854775807	\\x002b1df3816bd417639237e5b5676969bc8fa3c570f436a7e9238f133f0f05ce	32124782599	32000000000	f	0	0	9223372036854775807	44588	proto_lighthouse
81	\\x8f3d901d2484e30150a2befbdb72f637ae12b4ad8d39d190364967f95bafb4dcc0f906eb999b2c1cdd70478905f1f928	9223372036854775807	\\x00b50c195256e77379e65732b663773b6a7b7a9c7c9b6541784195af6721fe80	32121259037	32000000000	f	0	0	9223372036854775807	44612	proto_lighthouse
82	\\xaa4d06ca85fcfa442f5f777606f1ddfbdbfa11bbaf8cac2ea06b03a1ca8e3b73b46b4519d62349d45c52a1e58ff4a758	9223372036854775807	\\x008a51294d1d52855b91efa35fe66a1c33478c9524d9ba3b5c68b8cd0f66eb2b	32122285616	32000000000	f	0	0	9223372036854775807	44576	proto_lighthouse
83	\\x97fa3618bd698a8e34ba37672bb68ed4e5618b12f0aa94f5a73380bfdec431ebbd88340ab817fa7184259eeaa89eac79	9223372036854775807	\\x0013ac23fdb5d63778d7e68ad6bef2364940a598c135fd54ca2ef8c3e2eb3b93	32121201974	32000000000	f	0	0	9223372036854775807	44607	proto_lighthouse
84	\\xa40c385b86e965919861d415698cd9b608a5d19f0f25d4fd8dba394dd974fb57eb7fe7a1be188237b14b9c19d288f6e2	9223372036854775807	\\x007bbe63e0f6554b6a4e89d5828fac6559e3d50e98e438e7a0d20793415d499d	32121236115	32000000000	f	0	0	9223372036854775807	44619	proto_lighthouse
85	\\xaada169f71fda55e8d25a082e13783daaf63e96beccd1016a19b74e6cd63ef732962df0bcb99843cf4e66951883ccee6	9223372036854775807	\\x0079e1c3b6afc33e89bc5de1ffb821efd438c5d27041be849645aea244d8e831	32121059241	32000000000	f	0	0	9223372036854775807	44589	proto_lighthouse
86	\\xa2825c3810fe12d7b82e31345dbe397dd382bb806668af56dde384e5a5a24bf488bd2ff02a2b42bc192399d77e868ccb	9223372036854775807	\\x0097f1d4ce24a466628a2c7ec4d6e97a9fe43e241b97c936743596c8b4539d56	32120904504	32000000000	f	0	0	9223372036854775807	44609	proto_lighthouse
87	\\xae00791909d3fa62690d0419d7525322e4381d601448f76d181af8398fe67c550a27560dfaa443a654092cb8d0f89adf	9223372036854775807	\\x00e7e501c212d05267fa30da5bdae257b409a70c53801ea4027579752cd9f255	32119944309	32000000000	f	0	0	9223372036854775807	44582	proto_lighthouse
88	\\x85220e0846d40cbfdc54102f4c23644ffd303e9a46f0ba92a83514bdc59530f11456a320f9fb8734c2a2bbc3978c2871	9223372036854775807	\\x0070c438423261873067be189c1ec22f979156edf1ae93e665d44efafc7279b7	32129288874	32000000000	f	0	0	9223372036854775807	44606	proto_lighthouse
89	\\x8c4c8155183cafd5d4d0816d8fdac488d489e71e33e0e340e423bd7a9e4e768524d0367f84b5c59a4f61bd27285f1fcc	9223372036854775807	\\x00ea52abdc3d375552949be57e2836e236489c4a82483eb029294abf61ae3b17	32122920857	32000000000	f	0	0	9223372036854775807	44576	proto_lighthouse
90	\\x9752a6a79b2ef3502d9491f6ef0dd242d609999233895268edadb1e1f431b88785d43be47019028cc8a6152867229d7d	9223372036854775807	\\x008507c9f5f110b02ab3f61abdfb7a436860eeac039789d8a04b252ce11ea420	32119647568	32000000000	f	0	0	9223372036854775807	44616	proto_lighthouse
91	\\xae60a59b127d624c781e3a3f8df0dcc28a93916df163cc0ca9a3b701058c908475eae2eda4166613c4a80b4cdd34ce55	9223372036854775807	\\x0049528d595eb4a4056f5a413a455f78db90bfd3059036b0f81381cf83bfada5	32120155019	32000000000	f	0	0	9223372036854775807	44593	proto_lighthouse
92	\\x972b37c3878959b6b678ff7a39eb3f66b18b712b08cbbaef4d9deebd20a62ee051fd4688dfef1569e15171ddac436e06	9223372036854775807	\\x0019d4420e19795c6f8fc26fa6056391ece5b6adfff0e770f6557c09ce8012ac	32122848805	32000000000	f	0	0	9223372036854775807	44610	proto_lighthouse
93	\\xa6181e9935a67c439dca406d03a3bbbd1949bcf7825145364ffd6762c8825874d06b4f7d01c6809d4ab8d795c8c0fda8	9223372036854775807	\\x005167b8c20fbd241318b69a56dce576f47301a7e36e77ec93c30584d036f230	32121347613	32000000000	f	0	0	9223372036854775807	44623	proto_lighthouse
94	\\xadd895d337ff435378452a5679c8fe17651dba4954780014ef77b978bad5fb29d2d2974c88372162d693a400783370c6	9223372036854775807	\\x001783d5e1644838725a15dd60b1c1df4ea59c712deaf20a746499acde9e045a	32122532974	32000000000	f	0	0	9223372036854775807	44583	proto_lighthouse
95	\\x99d58b0547ae38f9d6e3b4cd3caa3c6032d6cb0f6254e1941ac18e9972780a719ed78ac761253e3930db6ce600fb7aa7	9223372036854775807	\\x00fbf2304d948058224264699342584b3aecc9078f2731f02454ddef07faa95e	32120928748	32000000000	f	0	0	9223372036854775807	44596	proto_lighthouse
96	\\xb7622055e39899f48b8133531ae66658b5d098773d60901bd0fb58d8d70f20fd89ab43351a3baf4fef8cb42b53aedb34	9223372036854775807	\\x007b8e4ac9a82e7cedee7228de5fbe0612ba8478c4ead1f703125d161d9a11e2	32118082178	32000000000	f	0	0	9223372036854775807	44613	proto_lighthouse
97	\\xb86f436bd51db1ae4776f5e360b40c9b05f07ba8590a15c0b73ebaf917f979510cc4756cf30a6526733d18eb986b8e4b	9223372036854775807	\\x001b9043ea9496b321f8f4390bac725cb10d3c533315c5a81b293baefebd6cc1	32116978030	32000000000	f	0	0	9223372036854775807	44621	proto_lighthouse
98	\\xaa6be03c9514abe44b49eeb8acdca85a010b68d14e8d5067970db0a9ee7384037dda075d386d8a9c287d1c8b36d502bf	9223372036854775807	\\x004d717102d53ab7be17d22becadf824491b6a091fbe5841efbd4d88a740787e	32121309779	32000000000	f	0	0	9223372036854775807	44612	proto_lighthouse
99	\\xb131a68b8671d5cb0c44aea145330a218bbc0698f3cd81c3fed88b12780e713254327957f3f953e26996a63e80462815	9223372036854775807	\\x000459b532b9111968bc20d918d88ca1a71577ba17b97cb338f45f334b9a8493	32118593146	32000000000	f	0	0	9223372036854775807	44594	proto_lighthouse
100	\\xb44e41514a23489f69d1b959498fde80a15082ec960aa76a55dec1dfb159faaf0a4a21abfb0b69544b86d24639f2525a	9223372036854775807	\\x00926c2149cd9af48e6476c34dd17539432420306116065183e6d19e1e70c6fa	32118507294	32000000000	f	0	0	9223372036854775807	44588	proto_lighthouse
101	\\xa9a264aae20bc5aff10baffac195adf0f94b6f1aaccbb964b6964bd2f9910632876239d6227584dcbc8fbd27d88937fa	9223372036854775807	\\x00f90664caeb146768a6c3c518582555fcfb2e1b507d28aa1eb6f9c0c5b7886f	32117997159	32000000000	f	0	0	9223372036854775807	44586	proto_lighthouse
\.

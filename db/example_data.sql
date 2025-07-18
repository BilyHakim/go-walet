--
-- PostgreSQL database dump
--

-- Dumped from database version 15.13 (Debian 15.13-1.pgdg120+1)
-- Dumped by pg_dump version 15.13 (Debian 15.13-1.pgdg120+1)

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
-- Name: transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transactions (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    target_user_id uuid,
    type text NOT NULL,
    amount numeric NOT NULL,
    remarks text,
    balance_before numeric,
    balance_after numeric,
    status text DEFAULT 'PENDING'::text,
    created_at timestamp with time zone
);


ALTER TABLE public.transactions OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    phone_number text NOT NULL,
    address text NOT NULL,
    pin text NOT NULL,
    balance numeric DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.transactions (id, user_id, target_user_id, type, amount, remarks, balance_before, balance_after, status, created_at) FROM stdin;
61d0109a-dbdd-43ec-9c0d-15c12b79d590	7fa38278-b7b0-42f9-a0a4-2ce5657d37b8	\N	CREDIT	100000	Top Up	0	100000	SUCCESS	2025-07-18 13:35:54.892124+00
89fa7f46-fd16-441a-8856-1e411edd4ea9	7fa38278-b7b0-42f9-a0a4-2ce5657d37b8	\N	DEBIT	25000	Pembayaran makanan	100000	75000	SUCCESS	2025-07-18 13:36:44.489864+00
746c4df3-3e8f-4839-b868-f333c4eeef81	531551f5-c34c-4c76-969a-302d83187560	7fa38278-b7b0-42f9-a0a4-2ce5657d37b8	CREDIT	5000	Transfer ke teman (from transfer ID: f92c52dd-c4d8-473e-807e-3adc145d03ca)	0	5000	SUCCESS	2025-07-18 14:07:37.441662+00
f92c52dd-c4d8-473e-807e-3adc145d03ca	7fa38278-b7b0-42f9-a0a4-2ce5657d37b8	531551f5-c34c-4c76-969a-302d83187560	DEBIT	5000	Transfer ke teman	75000	70000	SUCCESS	2025-07-18 14:07:37.406206+00
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, first_name, last_name, phone_number, address, pin, balance, created_at, updated_at) FROM stdin;
7fa38278-b7b0-42f9-a0a4-2ce5657d37b8	Bily Updated	Hakim Updated	081234567890	Jl. Thamrin No. 456, Jakarta	$2a$10$UCBq1V.u5q/qVj6yGVUquOE.RJXJkxLp5Dt3UEA6aOWWcCBBjfYTq	70000	2025-07-18 13:13:46.73611+00	2025-07-18 14:07:37.390352+00
531551f5-c34c-4c76-969a-302d83187560	Amal	Solihat	082298222515	Jl. Pekanbaru No.12 tangerang	$2a$10$ZzXvGZPQzxgXq3ZfCYzo0.dVxsX7RevfJ.8yQGeqTOFxL1Fz26.32	5000	2025-07-18 14:06:38.90304+00	2025-07-18 14:07:37.437928+00
\.


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: users uni_users_phone_number; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_phone_number UNIQUE (phone_number);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--


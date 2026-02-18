--
-- PostgreSQL database dump
--

\restrict Hy7yAtgLy2F98prSdIipTvt5fxndGVuRSEfeHXzEfJn6AtvDC1XlWfp7inP8IS7

-- Dumped from database version 14.20
-- Dumped by pg_dump version 14.20

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

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: get_order_status_id(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_order_status_id(status_name text) RETURNS integer
    LANGUAGE sql
    AS $$
	SELECT id FROM order_statuses WHERE status = status_name;
	$$;


ALTER FUNCTION public.get_order_status_id(status_name text) OWNER TO postgres;

--
-- Name: update_cart_total_price(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_cart_total_price() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
	BEGIN 
	IF (TG_OP = 'DELETE') THEN 
		UPDATE carts c 
			SET total_price = ( 
				SELECT COALESCE ( SUM ( CASE WHEN pi.discount_price > 0 THEN pi.discount_price * ci.qty ELSE pi.price * ci.qty END), 0)::bigint 
				FROM cart_items ci INNER JOIN product_items pi ON ci.product_item_id = pi.id 
				WHERE ci.cart_id = OLD.cart_id  
			), applied_coupon_id = 0, discount_amount = 0   
		WHERE c.id = OLD.cart_id; 
		RETURN NEW; 
	ELSE 
		UPDATE carts c 
			SET total_price = (
				SELECT SUM (CASE WHEN pi.discount_price > 0 THEN pi.discount_price * ci.qty ELSE pi.price * ci.qty END) 
				FROM cart_items ci INNER JOIN product_items pi ON ci.product_item_id = pi.id 
				WHERE ci.cart_id = NEW.cart_id 
			), applied_coupon_id = 0, discount_amount = 0 
			WHERE c.id = NEW.cart_id;
	
	END IF; 
	RETURN NEW; 
	END; 
	$$;


ALTER FUNCTION public.update_cart_total_price() OWNER TO postgres;

--
-- Name: update_product_quantity(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_product_quantity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ 
	BEGIN 
		IF (TG_OP = 'INSERT') THEN 
			UPDATE product_items pi 
			SET qty_in_stock = pi.qty_in_stock - NEW.qty 
			WHERE pi.id = NEW.product_item_id; 
	
		END IF; 
		RETURN NEW; 
	END; 
	$$;


ALTER FUNCTION public.update_product_quantity() OWNER TO postgres;

--
-- Name: update_product_quantity_on_return(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_product_quantity_on_return() RETURNS trigger
    LANGUAGE plpgsql
    AS $_$
	BEGIN
	  IF (TG_OP = 'UPDATE') THEN 
		EXECUTE format('UPDATE product_items pi
						SET qty_in_stock = qty_in_stock + ol.qty
						FROM %I ol
						WHERE pi.id = ol.product_item_id
						AND ol.shop_order_id = $1.id',
						'order_lines')
		USING NEW;
	  
		RETURN NEW;
	  ELSE
		RETURN NULL;
	  END IF;
	END;
	$_$;


ALTER FUNCTION public.update_product_quantity_on_return() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: addresses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.addresses (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    addtess_line1 text NOT NULL,
    addtess_line2 text NOT NULL,
    area character varying(255) DEFAULT NULL::character varying,
    land_mark text NOT NULL,
    city text NOT NULL,
    pincode bigint NOT NULL,
    country_id bigint NOT NULL,
    latitude numeric(10,7),
    longitude numeric(10,7),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    name text NOT NULL,
    phone_number text NOT NULL,
    house text NOT NULL,
    address_line1 text NOT NULL,
    address_line2 text NOT NULL,
    address_type text NOT NULL,
    is_default boolean
);


ALTER TABLE public.addresses OWNER TO postgres;

--
-- Name: addresses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.addresses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.addresses_id_seq OWNER TO postgres;

--
-- Name: addresses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.addresses_id_seq OWNED BY public.addresses.id;


--
-- Name: admin_refresh_sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.admin_refresh_sessions (
    token_id text NOT NULL,
    user_id bigint,
    admin_id bigint,
    user_type text,
    refresh_token text NOT NULL,
    expire_at timestamp with time zone NOT NULL,
    is_blocked boolean DEFAULT false NOT NULL
);


ALTER TABLE public.admin_refresh_sessions OWNER TO postgres;

--
-- Name: admins; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.admins (
    id bigint NOT NULL,
    full_name text,
    email text,
    password text NOT NULL,
    address_line1 character varying(255),
    address_line2 character varying(255),
    city character varying(50),
    state character varying(50),
    country character varying(50),
    pincode character varying(50),
    mobile character varying(50),
    latitude numeric(10,7),
    longitude numeric(10,7),
    payment_status boolean DEFAULT false NOT NULL,
    payment_type character varying(50),
    payment_date timestamp with time zone,
    start_date timestamp with time zone,
    expiry_date timestamp with time zone,
    bank_account_number character varying(50),
    bank_ifsc character varying(20),
    pan character varying(20),
    aadhar character varying(20),
    agree_to_terms boolean,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    verified_seller boolean DEFAULT false NOT NULL,
    status character varying(50),
    profile_image_url character varying(255),
    department_type integer
);


ALTER TABLE public.admins OWNER TO postgres;

--
-- Name: admins_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.admins_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.admins_id_seq OWNER TO postgres;

--
-- Name: admins_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.admins_id_seq OWNED BY public.admins.id;


--
-- Name: advertisements; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.advertisements (
    id bigint NOT NULL,
    title character varying(100),
    content text,
    image_url character varying(255),
    target_url character varying(255),
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    created_by_admin bigint NOT NULL,
    admin_id text NOT NULL,
    area_targeted character varying(255),
    pincode_targeted character varying(20),
    latitude numeric(10,7),
    longitude numeric(10,7),
    distance_km numeric(10,2),
    status character varying(50),
    priority character varying(20)
);


ALTER TABLE public.advertisements OWNER TO postgres;

--
-- Name: advertisements_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.advertisements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.advertisements_id_seq OWNER TO postgres;

--
-- Name: advertisements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.advertisements_id_seq OWNED BY public.advertisements.id;


--
-- Name: banners; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.banners (
    id bigint NOT NULL,
    title character varying(255) NOT NULL,
    description character varying(500),
    image_url character varying(500),
    link character varying(500),
    active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.banners OWNER TO postgres;

--
-- Name: banners_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.banners_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.banners_id_seq OWNER TO postgres;

--
-- Name: banners_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.banners_id_seq OWNED BY public.banners.id;


--
-- Name: brands; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.brands (
    id bigint NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.brands OWNER TO postgres;

--
-- Name: brands_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.brands_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.brands_id_seq OWNER TO postgres;

--
-- Name: brands_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.brands_id_seq OWNED BY public.brands.id;


--
-- Name: cart_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cart_items (
    id bigint NOT NULL,
    cart_id bigint,
    product_item_id bigint NOT NULL,
    qty bigint NOT NULL
);


ALTER TABLE public.cart_items OWNER TO postgres;

--
-- Name: cart_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cart_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cart_items_id_seq OWNER TO postgres;

--
-- Name: cart_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cart_items_id_seq OWNED BY public.cart_items.id;


--
-- Name: carts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.carts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total_price bigint NOT NULL,
    applied_coupon_id bigint,
    discount_amount bigint
);


ALTER TABLE public.carts OWNER TO postgres;

--
-- Name: carts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.carts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.carts_id_seq OWNER TO postgres;

--
-- Name: carts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.carts_id_seq OWNED BY public.carts.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.categories (
    id bigint NOT NULL,
    department_id bigint NOT NULL,
    name text NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    image_url text NOT NULL
);


ALTER TABLE public.categories OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.categories_id_seq OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: category_images; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.category_images (
    id bigint NOT NULL,
    category_id bigint NOT NULL,
    image_url text NOT NULL,
    alt_text text,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.category_images OWNER TO postgres;

--
-- Name: category_images_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.category_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.category_images_id_seq OWNER TO postgres;

--
-- Name: category_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.category_images_id_seq OWNED BY public.category_images.id;


--
-- Name: countries; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.countries (
    id bigint NOT NULL,
    country_name text NOT NULL,
    iso_code text NOT NULL
);


ALTER TABLE public.countries OWNER TO postgres;

--
-- Name: countries_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.countries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.countries_id_seq OWNER TO postgres;

--
-- Name: countries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.countries_id_seq OWNED BY public.countries.id;


--
-- Name: coupon_uses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.coupon_uses (
    coupon_uses_id bigint NOT NULL,
    coupon_id bigint NOT NULL,
    user_id bigint NOT NULL,
    used_at timestamp with time zone NOT NULL
);


ALTER TABLE public.coupon_uses OWNER TO postgres;

--
-- Name: coupon_uses_coupon_uses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.coupon_uses_coupon_uses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.coupon_uses_coupon_uses_id_seq OWNER TO postgres;

--
-- Name: coupon_uses_coupon_uses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.coupon_uses_coupon_uses_id_seq OWNED BY public.coupon_uses.coupon_uses_id;


--
-- Name: coupons; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.coupons (
    coupon_id bigint NOT NULL,
    coupon_name text NOT NULL,
    coupon_code text NOT NULL,
    expire_date timestamp with time zone NOT NULL,
    description text NOT NULL,
    discount_rate bigint NOT NULL,
    minimum_cart_price bigint NOT NULL,
    image text,
    block_status boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone
);


ALTER TABLE public.coupons OWNER TO postgres;

--
-- Name: coupons_coupon_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.coupons_coupon_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.coupons_coupon_id_seq OWNER TO postgres;

--
-- Name: coupons_coupon_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.coupons_coupon_id_seq OWNED BY public.coupons.coupon_id;


--
-- Name: departments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.departments (
    id bigint NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    image_url text NOT NULL
);


ALTER TABLE public.departments OWNER TO postgres;

--
-- Name: departments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.departments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.departments_id_seq OWNER TO postgres;

--
-- Name: departments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.departments_id_seq OWNED BY public.departments.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    sender_type character varying(50) NOT NULL,
    receiver_type character varying(50) NOT NULL,
    type character varying(100) NOT NULL,
    sender_id bigint NOT NULL,
    title character varying(255) NOT NULL,
    message text NOT NULL,
    body text NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    receiver_id bigint NOT NULL,
    category_id bigint NOT NULL,
    product_id bigint NOT NULL,
    variation_id bigint NOT NULL,
    shop_id bigint NOT NULL,
    user_id bigint NOT NULL,
    admin_id bigint NOT NULL,
    order_id bigint NOT NULL,
    offer_id bigint NOT NULL,
    notification_meta_data text,
    status character varying(50) NOT NULL,
    created_at character varying(50) NOT NULL,
    updated_at character varying(50) NOT NULL
);


ALTER TABLE public.notifications OWNER TO postgres;

--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notifications_id_seq OWNER TO postgres;

--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: offer_categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.offer_categories (
    id bigint NOT NULL,
    offer_id bigint NOT NULL,
    category_id bigint NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.offer_categories OWNER TO postgres;

--
-- Name: offer_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.offer_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.offer_categories_id_seq OWNER TO postgres;

--
-- Name: offer_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.offer_categories_id_seq OWNED BY public.offer_categories.id;


--
-- Name: offer_products; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.offer_products (
    id bigint NOT NULL,
    offer_id bigint NOT NULL,
    product_item_id bigint NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.offer_products OWNER TO postgres;

--
-- Name: offer_products_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.offer_products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.offer_products_id_seq OWNER TO postgres;

--
-- Name: offer_products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.offer_products_id_seq OWNED BY public.offer_products.id;


--
-- Name: offers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.offers (
    id bigint NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    discount_rate bigint NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    offer_type text NOT NULL,
    image text NOT NULL,
    thumbnail text
);


ALTER TABLE public.offers OWNER TO postgres;

--
-- Name: offers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.offers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.offers_id_seq OWNER TO postgres;

--
-- Name: offers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.offers_id_seq OWNED BY public.offers.id;


--
-- Name: order_lines; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_lines (
    id bigint NOT NULL,
    product_item_id bigint NOT NULL,
    shop_order_id bigint NOT NULL,
    qty bigint NOT NULL,
    price bigint NOT NULL
);


ALTER TABLE public.order_lines OWNER TO postgres;

--
-- Name: order_lines_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_lines_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.order_lines_id_seq OWNER TO postgres;

--
-- Name: order_lines_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.order_lines_id_seq OWNED BY public.order_lines.id;


--
-- Name: order_returns; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_returns (
    id bigint NOT NULL,
    shop_order_id bigint NOT NULL,
    request_date timestamp with time zone NOT NULL,
    return_reason text NOT NULL,
    refund_amount bigint NOT NULL,
    is_approved boolean,
    return_date timestamp with time zone,
    approval_date timestamp with time zone,
    admin_comment text
);


ALTER TABLE public.order_returns OWNER TO postgres;

--
-- Name: order_returns_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_returns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.order_returns_id_seq OWNER TO postgres;

--
-- Name: order_returns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.order_returns_id_seq OWNED BY public.order_returns.id;


--
-- Name: order_statuses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_statuses (
    id bigint NOT NULL,
    status text NOT NULL
);


ALTER TABLE public.order_statuses OWNER TO postgres;

--
-- Name: order_statuses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_statuses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.order_statuses_id_seq OWNER TO postgres;

--
-- Name: order_statuses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.order_statuses_id_seq OWNED BY public.order_statuses.id;


--
-- Name: otp_sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.otp_sessions (
    id bigint NOT NULL,
    otp_id text NOT NULL,
    user_id bigint NOT NULL,
    admin_id bigint,
    user_type text,
    phone text NOT NULL,
    expire_at timestamp with time zone NOT NULL
);


ALTER TABLE public.otp_sessions OWNER TO postgres;

--
-- Name: otp_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.otp_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.otp_sessions_id_seq OWNER TO postgres;

--
-- Name: otp_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.otp_sessions_id_seq OWNED BY public.otp_sessions.id;


--
-- Name: payment_methods; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.payment_methods (
    id bigint NOT NULL,
    name text NOT NULL,
    block_status boolean DEFAULT false NOT NULL,
    maximum_amount bigint NOT NULL
);


ALTER TABLE public.payment_methods OWNER TO postgres;

--
-- Name: payment_methods_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.payment_methods_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.payment_methods_id_seq OWNER TO postgres;

--
-- Name: payment_methods_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.payment_methods_id_seq OWNED BY public.payment_methods.id;


--
-- Name: product_configurations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_configurations (
    product_item_id bigint NOT NULL,
    variation_option_id bigint NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.product_configurations OWNER TO postgres;

--
-- Name: product_images; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_images (
    id bigint NOT NULL,
    product_item_id bigint NOT NULL,
    image text NOT NULL,
    image_url text[],
    shop_id bigint NOT NULL,
    product_id bigint NOT NULL,
    alt_text character varying(255),
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone
);


ALTER TABLE public.product_images OWNER TO postgres;

--
-- Name: product_images_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_images_id_seq OWNER TO postgres;

--
-- Name: product_images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_images_id_seq OWNED BY public.product_images.id;


--
-- Name: product_item_filter_types; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_item_filter_types (
    id bigint NOT NULL,
    filter_name text NOT NULL,
    shop_id bigint,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone
);


ALTER TABLE public.product_item_filter_types OWNER TO postgres;

--
-- Name: product_item_filter_types_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_item_filter_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_item_filter_types_id_seq OWNER TO postgres;

--
-- Name: product_item_filter_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_item_filter_types_id_seq OWNED BY public.product_item_filter_types.id;


--
-- Name: product_item_views; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_item_views (
    id bigint NOT NULL,
    product_item_id bigint NOT NULL,
    shop_id bigint NOT NULL,
    admin_id jsonb NOT NULL,
    viewed_at timestamp with time zone NOT NULL,
    view_count bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone
);


ALTER TABLE public.product_item_views OWNER TO postgres;

--
-- Name: product_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_items (
    id bigint NOT NULL,
    sub_category_name text NOT NULL,
    dynamic_fields jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    product_item_images text[],
    admin_id jsonb NOT NULL,
    sub_category_id bigint,
    category_id bigint,
    department_id bigint,
    shop_id bigint,
    stock boolean DEFAULT true
);


ALTER TABLE public.product_items OWNER TO postgres;

--
-- Name: product_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_items_id_seq OWNER TO postgres;

--
-- Name: product_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_items_id_seq OWNED BY public.product_items.id;


--
-- Name: product_views_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_views_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.product_views_id_seq OWNER TO postgres;

--
-- Name: product_views_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_views_id_seq OWNED BY public.product_item_views.id;


--
-- Name: products; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    category_id bigint,
    image text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone,
    stock bigint,
    department_id bigint,
    shop_id bigint NOT NULL,
    price bigint DEFAULT 0,
    discount_price bigint DEFAULT 0,
    brand_id bigint
);


ALTER TABLE public.products OWNER TO postgres;

--
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.products_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.products_id_seq OWNER TO postgres;

--
-- Name: products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;


--
-- Name: promotion_categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.promotion_categories (
    id bigint NOT NULL,
    name text NOT NULL,
    shop_id bigint,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    icon_path text
);


ALTER TABLE public.promotion_categories OWNER TO postgres;

--
-- Name: promotion_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.promotion_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.promotion_categories_id_seq OWNER TO postgres;

--
-- Name: promotion_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.promotion_categories_id_seq OWNED BY public.promotion_categories.id;


--
-- Name: promotions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.promotions (
    id bigint NOT NULL,
    promotion_category_id bigint,
    promotion_type_id bigint,
    offer_name text,
    description text,
    discount_rate numeric,
    start_date text,
    end_date text,
    minimum_purchase_amount numeric,
    tier_quantity bigint,
    bogo_get_quantity bigint,
    bogo_buy_quantity bigint,
    bogo_combination_enabled boolean,
    gift_description text,
    shop_id bigint,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone
);


ALTER TABLE public.promotions OWNER TO postgres;

--
-- Name: promotions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.promotions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.promotions_id_seq OWNER TO postgres;

--
-- Name: promotions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.promotions_id_seq OWNED BY public.promotions.id;


--
-- Name: promotions_types; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.promotions_types (
    id bigint NOT NULL,
    name text,
    is_active boolean,
    shop_id text,
    promotion_category_id bigint,
    promotion_offer_details jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    icon_path text,
    type character varying(255)
);


ALTER TABLE public.promotions_types OWNER TO postgres;

--
-- Name: promotions_types_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.promotions_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.promotions_types_id_seq OWNER TO postgres;

--
-- Name: promotions_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.promotions_types_id_seq OWNED BY public.promotions_types.id;


--
-- Name: service_providers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.service_providers (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    business_name character varying(255),
    shop_id character varying(255),
    admin_id character varying(255),
    profile_photo text,
    bio text,
    phone character varying(20) NOT NULL,
    whatsapp character varying(20),
    email character varying(255),
    base_address text,
    serviceable_pincodes text[],
    service_radius_km integer,
    categories text[],
    sub_services text[],
    experience_years integer,
    tools_brought text[],
    pricing_model character varying(50),
    base_charge numeric(10,2),
    min_job_charge numeric(10,2),
    rate_card jsonb,
    working_days text[],
    time_slots text[],
    advance_notice_hours integer,
    kyc_status character varying(50) DEFAULT 'pending'::character varying,
    license_number character varying(100),
    insurance boolean DEFAULT false,
    police_verification boolean DEFAULT false,
    cancellation_hours integer,
    warranty_days integer,
    rating numeric(3,2) DEFAULT 0.0,
    total_jobs integer DEFAULT 0,
    response_time_min integer,
    portfolio_images text[],
    account_status character varying(50) DEFAULT 'active'::character varying,
    payout_upi character varying(100),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.service_providers OWNER TO postgres;

--
-- Name: service_providers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.service_providers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.service_providers_id_seq OWNER TO postgres;

--
-- Name: service_providers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.service_providers_id_seq OWNED BY public.service_providers.id;


--
-- Name: shop_details; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shop_details (
    id bigint NOT NULL,
    admin_id bigint,
    shop_name character varying(100),
    owner_name character varying(100),
    email character varying(100),
    phone character varying(50),
    address_line1 character varying(255),
    address_line2 character varying(255),
    city character varying(50),
    state character varying(50),
    country character varying(50),
    pincode character varying(50),
    latitude numeric(10,7),
    longitude numeric(10,7),
    shop_description text,
    shop_verification_docs text,
    document_type character varying(50),
    document_value text,
    pan_number character varying(20),
    itr_documents text,
    shop_status character varying(50),
    bank_account_number character varying(50),
    bank_ifsc character varying(20),
    shop_image_url character varying(255),
    shop_verification_status boolean DEFAULT false NOT NULL,
    shop_verification_remarks text DEFAULT 'false'::text NOT NULL,
    photo_shop_verification boolean DEFAULT false NOT NULL,
    business_doc_verification boolean DEFAULT false NOT NULL,
    identity_doc_verification boolean DEFAULT false NOT NULL,
    address_proof_verification boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    has_offers boolean,
    shop_type character varying(50)
);


ALTER TABLE public.shop_details OWNER TO postgres;

--
-- Name: shop_details_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.shop_details_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shop_details_id_seq OWNER TO postgres;

--
-- Name: shop_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.shop_details_id_seq OWNED BY public.shop_details.id;


--
-- Name: shop_offers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shop_offers (
    id bigint NOT NULL,
    shop_id bigint NOT NULL,
    offer_id bigint NOT NULL,
    admin_id text NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone
);


ALTER TABLE public.shop_offers OWNER TO postgres;

--
-- Name: shop_offers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.shop_offers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shop_offers_id_seq OWNER TO postgres;

--
-- Name: shop_offers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.shop_offers_id_seq OWNED BY public.shop_offers.id;


--
-- Name: shop_orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shop_orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    order_date timestamp with time zone NOT NULL,
    address_id bigint NOT NULL,
    order_total_price bigint NOT NULL,
    discount bigint NOT NULL,
    order_status_id bigint NOT NULL,
    payment_method_id bigint,
    shop_id bigint
);


ALTER TABLE public.shop_orders OWNER TO postgres;

--
-- Name: shop_orders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.shop_orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shop_orders_id_seq OWNER TO postgres;

--
-- Name: shop_orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.shop_orders_id_seq OWNED BY public.shop_orders.id;


--
-- Name: shop_times; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shop_times (
    id bigint NOT NULL,
    shop_id bigint NOT NULL,
    status character varying(20) NOT NULL,
    open_time text NOT NULL,
    close_time text NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.shop_times OWNER TO postgres;

--
-- Name: shop_times_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.shop_times_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shop_times_id_seq OWNER TO postgres;

--
-- Name: shop_times_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.shop_times_id_seq OWNED BY public.shop_times.id;


--
-- Name: shop_verification_histories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shop_verification_histories (
    id bigint NOT NULL,
    admin_id text NOT NULL,
    shop_id bigint NOT NULL,
    verification_status text NOT NULL,
    remarks character varying(255),
    changed_at timestamp with time zone NOT NULL
);


ALTER TABLE public.shop_verification_histories OWNER TO postgres;

--
-- Name: shop_verification_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.shop_verification_histories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shop_verification_histories_id_seq OWNER TO postgres;

--
-- Name: shop_verification_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.shop_verification_histories_id_seq OWNED BY public.shop_verification_histories.id;


--
-- Name: shop_verifications; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.shop_verifications (
    id bigint NOT NULL,
    admin_id text,
    shop_id bigint,
    shop_name text,
    verification_status boolean DEFAULT false NOT NULL,
    remarks text,
    agent_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.shop_verifications OWNER TO postgres;

--
-- Name: shop_verifications_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.shop_verifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.shop_verifications_id_seq OWNER TO postgres;

--
-- Name: shop_verifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.shop_verifications_id_seq OWNED BY public.shop_verifications.id;


--
-- Name: sub_categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sub_categories (
    id bigint NOT NULL,
    department_id bigint NOT NULL,
    category_id bigint NOT NULL,
    name text NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    image_url text
);


ALTER TABLE public.sub_categories OWNER TO postgres;

--
-- Name: sub_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sub_categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.sub_categories_id_seq OWNER TO postgres;

--
-- Name: sub_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sub_categories_id_seq OWNED BY public.sub_categories.id;


--
-- Name: sub_category_details; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sub_category_details (
    id bigint NOT NULL,
    sub_category_id bigint,
    sub_category_image_url text
);


ALTER TABLE public.sub_category_details OWNER TO postgres;

--
-- Name: sub_category_details_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sub_category_details_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.sub_category_details_id_seq OWNER TO postgres;

--
-- Name: sub_category_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sub_category_details_id_seq OWNED BY public.sub_category_details.id;


--
-- Name: sub_type_attribute_options; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sub_type_attribute_options (
    id bigint NOT NULL,
    sub_type_attribute_id bigint NOT NULL,
    option_value character varying(50),
    sort_order bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.sub_type_attribute_options OWNER TO postgres;

--
-- Name: sub_type_attribute_options_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sub_type_attribute_options_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.sub_type_attribute_options_id_seq OWNER TO postgres;

--
-- Name: sub_type_attribute_options_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sub_type_attribute_options_id_seq OWNED BY public.sub_type_attribute_options.id;


--
-- Name: sub_type_attributes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sub_type_attributes (
    id bigint NOT NULL,
    sub_category_id bigint NOT NULL,
    field_name character varying(50),
    field_type character varying(20),
    is_required boolean DEFAULT true NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL
);


ALTER TABLE public.sub_type_attributes OWNER TO postgres;

--
-- Name: sub_type_attributes_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sub_type_attributes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.sub_type_attributes_id_seq OWNER TO postgres;

--
-- Name: sub_type_attributes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sub_type_attributes_id_seq OWNED BY public.sub_type_attributes.id;


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.transactions (
    transaction_id bigint NOT NULL,
    wallet_id bigint NOT NULL,
    transaction_date timestamp with time zone NOT NULL,
    amount bigint NOT NULL,
    transaction_type text NOT NULL
);


ALTER TABLE public.transactions OWNER TO postgres;

--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.transactions_transaction_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.transactions_transaction_id_seq OWNER TO postgres;

--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.transactions_transaction_id_seq OWNED BY public.transactions.transaction_id;


--
-- Name: user_addresses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_addresses (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    address_id bigint NOT NULL,
    is_default boolean
);


ALTER TABLE public.user_addresses OWNER TO postgres;

--
-- Name: user_addresses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_addresses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_addresses_id_seq OWNER TO postgres;

--
-- Name: user_addresses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_addresses_id_seq OWNED BY public.user_addresses.id;


--
-- Name: user_refresh_sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_refresh_sessions (
    token_id text NOT NULL,
    user_id bigint,
    admin_id bigint,
    user_type text,
    refresh_token text NOT NULL,
    expire_at timestamp with time zone NOT NULL,
    is_blocked boolean DEFAULT false NOT NULL
);


ALTER TABLE public.user_refresh_sessions OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    age bigint,
    first_name text,
    last_name text,
    email text,
    phone text,
    password text,
    verified boolean DEFAULT false,
    block_status boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: variation_options; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.variation_options (
    id bigint NOT NULL,
    variation_id bigint NOT NULL,
    value text NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.variation_options OWNER TO postgres;

--
-- Name: variation_options_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.variation_options_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.variation_options_id_seq OWNER TO postgres;

--
-- Name: variation_options_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.variation_options_id_seq OWNED BY public.variation_options.id;


--
-- Name: variations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.variations (
    id bigint NOT NULL,
    sub_category_id bigint NOT NULL,
    name text NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.variations OWNER TO postgres;

--
-- Name: variations_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.variations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.variations_id_seq OWNER TO postgres;

--
-- Name: variations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.variations_id_seq OWNED BY public.variations.id;


--
-- Name: wallets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wallets (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total_amount bigint NOT NULL
);


ALTER TABLE public.wallets OWNER TO postgres;

--
-- Name: wallets_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.wallets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.wallets_id_seq OWNER TO postgres;

--
-- Name: wallets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.wallets_id_seq OWNED BY public.wallets.id;


--
-- Name: wish_lists; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wish_lists (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    shop_id bigint NOT NULL,
    admin_id bigint NOT NULL,
    product_item_id bigint NOT NULL
);


ALTER TABLE public.wish_lists OWNER TO postgres;

--
-- Name: wish_lists_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.wish_lists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.wish_lists_id_seq OWNER TO postgres;

--
-- Name: wish_lists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.wish_lists_id_seq OWNED BY public.wish_lists.id;


--
-- Name: addresses id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.addresses ALTER COLUMN id SET DEFAULT nextval('public.addresses_id_seq'::regclass);


--
-- Name: admins id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admins ALTER COLUMN id SET DEFAULT nextval('public.admins_id_seq'::regclass);


--
-- Name: advertisements id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.advertisements ALTER COLUMN id SET DEFAULT nextval('public.advertisements_id_seq'::regclass);


--
-- Name: banners id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.banners ALTER COLUMN id SET DEFAULT nextval('public.banners_id_seq'::regclass);


--
-- Name: brands id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.brands ALTER COLUMN id SET DEFAULT nextval('public.brands_id_seq'::regclass);


--
-- Name: cart_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cart_items ALTER COLUMN id SET DEFAULT nextval('public.cart_items_id_seq'::regclass);


--
-- Name: carts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.carts ALTER COLUMN id SET DEFAULT nextval('public.carts_id_seq'::regclass);


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: category_images id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category_images ALTER COLUMN id SET DEFAULT nextval('public.category_images_id_seq'::regclass);


--
-- Name: countries id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.countries ALTER COLUMN id SET DEFAULT nextval('public.countries_id_seq'::regclass);


--
-- Name: coupon_uses coupon_uses_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupon_uses ALTER COLUMN coupon_uses_id SET DEFAULT nextval('public.coupon_uses_coupon_uses_id_seq'::regclass);


--
-- Name: coupons coupon_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupons ALTER COLUMN coupon_id SET DEFAULT nextval('public.coupons_coupon_id_seq'::regclass);


--
-- Name: departments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.departments ALTER COLUMN id SET DEFAULT nextval('public.departments_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: offer_categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offer_categories ALTER COLUMN id SET DEFAULT nextval('public.offer_categories_id_seq'::regclass);


--
-- Name: offer_products id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offer_products ALTER COLUMN id SET DEFAULT nextval('public.offer_products_id_seq'::regclass);


--
-- Name: offers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offers ALTER COLUMN id SET DEFAULT nextval('public.offers_id_seq'::regclass);


--
-- Name: order_lines id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_lines ALTER COLUMN id SET DEFAULT nextval('public.order_lines_id_seq'::regclass);


--
-- Name: order_returns id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_returns ALTER COLUMN id SET DEFAULT nextval('public.order_returns_id_seq'::regclass);


--
-- Name: order_statuses id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_statuses ALTER COLUMN id SET DEFAULT nextval('public.order_statuses_id_seq'::regclass);


--
-- Name: otp_sessions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.otp_sessions ALTER COLUMN id SET DEFAULT nextval('public.otp_sessions_id_seq'::regclass);


--
-- Name: payment_methods id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_methods ALTER COLUMN id SET DEFAULT nextval('public.payment_methods_id_seq'::regclass);


--
-- Name: product_images id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_images ALTER COLUMN id SET DEFAULT nextval('public.product_images_id_seq'::regclass);


--
-- Name: product_item_filter_types id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_item_filter_types ALTER COLUMN id SET DEFAULT nextval('public.product_item_filter_types_id_seq'::regclass);


--
-- Name: product_item_views id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_item_views ALTER COLUMN id SET DEFAULT nextval('public.product_views_id_seq'::regclass);


--
-- Name: product_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_items ALTER COLUMN id SET DEFAULT nextval('public.product_items_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);


--
-- Name: promotion_categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_categories ALTER COLUMN id SET DEFAULT nextval('public.promotion_categories_id_seq'::regclass);


--
-- Name: promotions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions ALTER COLUMN id SET DEFAULT nextval('public.promotions_id_seq'::regclass);


--
-- Name: promotions_types id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions_types ALTER COLUMN id SET DEFAULT nextval('public.promotions_types_id_seq'::regclass);


--
-- Name: service_providers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_providers ALTER COLUMN id SET DEFAULT nextval('public.service_providers_id_seq'::regclass);


--
-- Name: shop_details id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_details ALTER COLUMN id SET DEFAULT nextval('public.shop_details_id_seq'::regclass);


--
-- Name: shop_offers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_offers ALTER COLUMN id SET DEFAULT nextval('public.shop_offers_id_seq'::regclass);


--
-- Name: shop_orders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_orders ALTER COLUMN id SET DEFAULT nextval('public.shop_orders_id_seq'::regclass);


--
-- Name: shop_times id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_times ALTER COLUMN id SET DEFAULT nextval('public.shop_times_id_seq'::regclass);


--
-- Name: shop_verification_histories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_verification_histories ALTER COLUMN id SET DEFAULT nextval('public.shop_verification_histories_id_seq'::regclass);


--
-- Name: shop_verifications id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_verifications ALTER COLUMN id SET DEFAULT nextval('public.shop_verifications_id_seq'::regclass);


--
-- Name: sub_categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_categories ALTER COLUMN id SET DEFAULT nextval('public.sub_categories_id_seq'::regclass);


--
-- Name: sub_category_details id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_category_details ALTER COLUMN id SET DEFAULT nextval('public.sub_category_details_id_seq'::regclass);


--
-- Name: sub_type_attribute_options id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_type_attribute_options ALTER COLUMN id SET DEFAULT nextval('public.sub_type_attribute_options_id_seq'::regclass);


--
-- Name: sub_type_attributes id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_type_attributes ALTER COLUMN id SET DEFAULT nextval('public.sub_type_attributes_id_seq'::regclass);


--
-- Name: transactions transaction_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions ALTER COLUMN transaction_id SET DEFAULT nextval('public.transactions_transaction_id_seq'::regclass);


--
-- Name: user_addresses id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_addresses ALTER COLUMN id SET DEFAULT nextval('public.user_addresses_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: variation_options id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.variation_options ALTER COLUMN id SET DEFAULT nextval('public.variation_options_id_seq'::regclass);


--
-- Name: variations id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.variations ALTER COLUMN id SET DEFAULT nextval('public.variations_id_seq'::regclass);


--
-- Name: wallets id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wallets ALTER COLUMN id SET DEFAULT nextval('public.wallets_id_seq'::regclass);


--
-- Name: wish_lists id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wish_lists ALTER COLUMN id SET DEFAULT nextval('public.wish_lists_id_seq'::regclass);


--
-- Data for Name: addresses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.addresses (id, user_id, addtess_line1, addtess_line2, area, land_mark, city, pincode, country_id, latitude, longitude, created_at, updated_at, name, phone_number, house, address_line1, address_line2, address_type, is_default) FROM stdin;
\.


--
-- Data for Name: admin_refresh_sessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.admin_refresh_sessions (token_id, user_id, admin_id, user_type, refresh_token, expire_at, is_blocked) FROM stdin;
c6fe4e70-f2c9-442d-a68e-9e3de116e9fc	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYzZmZTRlNzAtZjJjOS00NDJkLWE2OGUtOWUzZGUxMTZlOWZjIiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjJUMTA6MjI6MzcuMDQxNDg5NiswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.HAALq09y7WeaQXhXrwuGwtDT2Y7yqqvhS5g7B4QSJiA	2026-02-22 04:52:37+00	f
ae4d80f6-ebb7-47dd-bc41-696b96116f21	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYWU0ZDgwZjYtZWJiNy00N2RkLWJjNDEtNjk2Yjk2MTE2ZjIxIiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjNUMjM6NDU6NTguMTE4OTI2MyswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.AHKz23668bRdXFZRw8BtDCJJFV5J-x8-mSZeKmlgxsw	2026-02-23 18:15:58+00	f
29099133-28ed-4d0f-952e-8870c67b2242	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiMjkwOTkxMzMtMjhlZC00ZDBmLTk1MmUtODg3MGM2N2IyMjQyIiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjNUMjM6NDY6MTIuOTc4MzE1MiswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.x-qf0sv2_zoelmnefciSRwad0E4o_O2O2y-LFV8cu3o	2026-02-23 18:16:12+00	f
a953a742-c20a-4e2c-be8d-ed72dafce043	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYTk1M2E3NDItYzIwYS00ZTJjLWJlOGQtZWQ3MmRhZmNlMDQzIiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjNUMjM6NDY6NDYuNDc5NjEwNCswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.K5WDe2LjtOXfN_zEFIg-GaSPtbJnFp1Ld9FwJ31JboA	2026-02-23 18:16:46+00	f
1fb3912d-4f0c-4541-b983-1444eff473d2	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiMWZiMzkxMmQtNGYwYy00NTQxLWI5ODMtMTQ0NGVmZjQ3M2QyIiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjNUMjM6NDY6NDkuMDM3MTkyMyswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.48_acHYaqL5CbvcO7x1RtsJFcWbAMDiXNy-MCuuqxfQ	2026-02-23 18:16:49+00	f
207365c3-05ae-4ac7-a287-646dd6d48cbf	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiMjA3MzY1YzMtMDVhZS00YWM3LWEyODctNjQ2ZGQ2ZDQ4Y2JmIiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjNUMjM6NDY6NTkuMTM0NTkxMiswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.TnZdmpHRZm4A46N6j11JuCx3C3tn4x2gIRQYshCZ2uk	2026-02-23 18:16:59+00	f
c689b231-de0a-4412-b3da-6f355239d88e	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYzY4OWIyMzEtZGUwYS00NDEyLWIzZGEtNmYzNTUyMzlkODhlIiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjRUMDg6NDM6NTAuMTc0NjUyOCswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.FXrEFPe1_3vZG3COF7s8cI6iGJ3n-ivRBgjrdzS28NM	2026-02-24 03:13:50+00	f
a89ef783-c701-43a5-8c72-4628710efd64	1	\N	admin	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYTg5ZWY3ODMtYzcwMS00M2E1LThjNzItNDYyODcxMGVmZDY0IiwiVXNlcklEIjoiMSIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMjRUMDg6NDM6NTUuMzYxNDI0MyswNTozMCIsIlVzZWRGb3IiOiJhZG1pbiJ9.tvUaOuxIgagayR-KtY7E_Fkr5nI4i5BVPtR7oJZ9nwM	2026-02-24 03:13:55+00	f
\.


--
-- Data for Name: admins; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.admins (id, full_name, email, password, address_line1, address_line2, city, state, country, pincode, mobile, latitude, longitude, payment_status, payment_type, payment_date, start_date, expiry_date, bank_account_number, bank_ifsc, pan, aadhar, agree_to_terms, created_at, updated_at, verified_seller, status, profile_image_url, department_type) FROM stdin;
2	Rohit		$2a$10$Y8FzRHnyEfrmxr4OCz2rru1EBX0U4jgrEFy341.zFEOj.sAV6JFZy							9886569963	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2025-12-16 15:16:09.76078+00	2025-12-16 15:16:09.76078+00	f		\N	\N
3	Rohit		$2a$10$QvEkxrmu77y7krS83Owgb.pbPOpRUiR0mI1zZvZwgMqQ7rGxLZ4hW							9886569961	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2025-12-18 12:46:48.930434+00	2025-12-18 12:46:48.930434+00	f		\N	\N
8	Rohit		$2a$10$C5V8X65ZOyDZIlD.wTJEle1OEA.7.RScIlVonljnzArIBIDZCTQu2							jangid.rohit70@gmail.com	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-01-23 19:09:12.875395+00	2026-01-23 19:09:12.875395+00	f		\N	\N
9	Rohit		$2a$10$eg8Ah9GUKJ2kHNquSWnu0ejZ/6cvUKELu7Ps7pqGwicNClCNZskw.							9549115670	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-01-23 19:09:46.263607+00	2026-01-23 19:09:46.263607+00	f		\N	\N
10	rohit		$2a$10$QWLE3qNOhGQZ0TiAJkGfm.U8Mwzv0pIdHlOr61s91TB0GsiqJYcum							jangid.rohit70@gmail.com	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-01-25 19:51:33.661902+00	2026-01-25 19:51:33.661902+00	f		\N	\N
11	rohit		$2a$10$cQM8gGbLSMZXK8QgmIVvZusA0XSayaVS6eyvbglFlcbksy2XjEPNe							jangid.rohit70@gmail.com	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-01-25 19:52:47.963259+00	2026-01-25 19:52:47.963259+00	f		\N	\N
1	Rohit Jangid		$2a$10$72l8k6qmI02kLKFXv8DVtOYAoBxPppN1XIrlPSxh9ZbYVZq2Z6E16							9886569962	0.0000000	0.0000000	f	\N	\N	\N	\N			pan	aadhar	f	2025-12-01 18:14:22.325133+00	2026-01-26 10:04:09.406614+00	t		uploads/admin-profiles/admin_1_1764836706.jpg	\N
12	Rohit Jangid		$2a$10$LEQKVHW63h6whmEOPr/h.O1EtOLCB9jqNw2sKF58aNdTnO21ttnYu							9549115670	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-01-26 10:07:57.705436+00	2026-01-26 10:07:57.705436+00	f		\N	\N
13	rohit		$2a$10$6nMTJvqDYmeoKOsd0YiRG.10vbptwYlSA27s4Fgn.KJD5b3YStAGC							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 09:48:09.03176+00	2026-02-01 09:48:09.03176+00	f		\N	\N
14	Dhruv		$2a$10$wRTP8wtnEeSZ/nDEOk3qUu5NAo/h9GeYAByj5YF5xxnsD/VHtVg9G							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:15:32.359743+00	2026-02-01 11:15:32.359743+00	f		\N	\N
15	Dhruv		$2a$10$ag0RNjd1uUJE2lRyc1P7O.NKy7XH5XgtML5DY7s0SKvFAxzShrNVq							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:24:16.680538+00	2026-02-01 11:24:16.680538+00	f		\N	\N
16	Dhruv		$2a$10$qrwnbzl3m5OPgk0U4GTAJ.ZUwRs6cETdj6p0cKk0rILFlC4VPvpVi							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:25:03.78216+00	2026-02-01 11:25:03.78216+00	f		\N	\N
17	Dhruv		$2a$10$msqaXzW017a1TGfNKKQK..M.dGYIx89ti17uN7O/0b0v.iX7vD7U.							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:26:33.858363+00	2026-02-01 11:26:33.858363+00	f		\N	\N
18	Dhruv		$2a$10$gSLvbTxO95uO.eaoCGu3w.CqOFgOj1PgdS5pCQVAEXeK60FfzD.1a							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:43:12.693989+00	2026-02-01 11:43:12.693989+00	f		\N	\N
19	rohit		$2a$10$eEjLY1b2Ydac8so.4KeXveSBwHkN0Rl7tIlR05YqR1E0cj3B5QKYK							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:45:31.661863+00	2026-02-01 11:45:31.661863+00	f		\N	\N
20	Dhruv		$2a$10$7/aO9TOYuH4Xvn47IfCjj.M2/GW7OxA6ZP1CKLGJf6FVMR3jbs8KK							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:54:50.524058+00	2026-02-01 11:54:50.524058+00	f		\N	\N
21	rohit		$2a$10$gJf5Yw.K079WldQAiaPlDugiCm2DV1Kf/1I16ix56NErlRS/H8dO6							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 11:55:48.805581+00	2026-02-01 11:55:48.805581+00	f		\N	\N
22	Dhruv		$2a$10$MZ9Vpwm/TJER1TP/Xr1Gl.00AHJVSiX/ncBdnZv/lfBhpHkxTr.Om							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 12:00:37.984766+00	2026-02-01 12:00:37.984766+00	f		\N	\N
23	Dhruv		$2a$10$WGI.FoKdPkh6rh9ffp1c8uXInpjDTSkVMgTFncnk1ttrGg5IHvuam							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 12:04:22.73959+00	2026-02-01 12:04:22.73959+00	f		\N	\N
24	Dhruv		$2a$10$biidFJHe.vgwXSZM0VxpDu4uJKeECaF7L1aI1bR2WMUbwAclQvi82							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 12:24:15.135144+00	2026-02-01 12:24:15.135144+00	f		\N	\N
25	Dhruv		$2a$10$cR2WIzE1R.FDn.ZnZNGoe.jjzhndzRh2XFMYzn/9T0CorVQ5XYon6							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 12:53:48.262594+00	2026-02-01 12:53:48.262594+00	f		\N	\N
26	Dhruv		$2a$10$NCjghvKgqRiaNGsYGmu9w.JGMdNfA9YaXg6dUayG1ktxhBFOkL38O							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:02:44.567181+00	2026-02-01 13:02:44.567181+00	f		\N	\N
27	Dhruv		$2a$10$ML7RtlhBTQGeyI3bmn9Ng.hjlXA8rVXYDT5yZMOy3JQJEh8mmzxiK							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:05:19.632758+00	2026-02-01 13:05:19.632758+00	f		\N	\N
28	Dhruv		$2a$10$GwTNyh9F6jkM431jboCO8OoafIqWQDM0TLbeE7roS23Cj.egRw5f6							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:05:46.977921+00	2026-02-01 13:05:46.977921+00	f		\N	\N
29	Dhruv		$2a$10$cjIN.E3YeqfLAylipo5z3OBeNVdqN4ejK4MfjYG.fGXBegNuBl7Oi							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:07:16.982151+00	2026-02-01 13:07:16.982151+00	f		\N	\N
30	Dhruv		$2a$10$7y5utRaBKuwfYcTRliuEqOYMPxzzxEuEfcTrR5ahXUF9ofyS/i3Tu							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:13:17.577779+00	2026-02-01 13:13:17.577779+00	f		\N	\N
31	rohit		$2a$10$vOZ53DTOXPjcrfl2oUmCr.WQQ4uX97llkJjDCYl3URKheSXge3Uj2							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:17:27.806061+00	2026-02-01 13:17:27.806061+00	f		\N	\N
32	Rohit		$2a$10$W9blEg5Z14L8k8FpC1ULveaknGBQ3GsEkgLz5u5gS.3RbZQwXiHla							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:21:45.941401+00	2026-02-01 13:21:45.941401+00	f		\N	\N
6	Dhruv		$2a$10$0rsscQ2UdeL/AQd69leFZOKTfaghFJfqxiqfL0ZtUmUCIm4h.cu1q							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-01-05 13:11:17.17757+00	2026-01-05 13:11:17.17757+00	t		\N	\N
33	Rohit		$2a$10$NBFThxiz02/W13wiuN2CE.iBG9A0IS/qzdNjP4GwAyT0rFPO4EFGC							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-01 13:23:27.169195+00	2026-02-01 13:23:27.169195+00	f		\N	\N
7	Dhruv		$2a$10$alqWEasFoFNbm0V29C4p5eYN0evDU41FqCqpZvZ8mXNr/1cUAPeKy							8302827722	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-01-06 05:06:16.430365+00	2026-01-06 05:06:16.430365+00	t		\N	\N
34	Rohit		$2a$10$/.9TVmeY8tcBjXjZEn.qduc8YV.Yw18bOAXae4oIvaBZik4JO5cIK							9549115670	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-16 18:01:40.069281+00	2026-02-16 18:01:40.069281+00	f		\N	\N
35	Rohit		$2a$10$9/jQh87XyDj1.Epr.p4Or..KyP36t0nOefJ1u5t4hnr8wJYf.to/y							9549115670	0.0000000	0.0000000	f	\N	\N	\N	\N					f	2026-02-16 18:02:40.417245+00	2026-02-16 18:02:40.417245+00	f		\N	\N
\.


--
-- Data for Name: advertisements; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.advertisements (id, title, content, image_url, target_url, start_date, end_date, created_at, updated_at, created_by_admin, admin_id, area_targeted, pincode_targeted, latitude, longitude, distance_km, status, priority) FROM stdin;
\.


--
-- Data for Name: banners; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.banners (id, title, description, image_url, link, active, created_at, updated_at) FROM stdin;
2	apply_offer	to sell the product apply the offer on that	uploads/banners/apply_offer_banner.jpg	\N	t	\N	\N
3	Shop Digital	To make shop digital	uploads/banners/shop_digital_banner.jpg	\N	t	\N	\N
4	User Communication	To communicate with user	uploads/banners/communicate_banner.jpg	\N	t	\N	\N
\.


--
-- Data for Name: brands; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.brands (id, name, slug, sort_order, is_active) FROM stdin;
\.


--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.cart_items (id, cart_id, product_item_id, qty) FROM stdin;
\.


--
-- Data for Name: carts; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.carts (id, user_id, total_price, applied_coupon_id, discount_amount) FROM stdin;
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.categories (id, department_id, name, sort_order, is_active, image_url) FROM stdin;
1	1	Fasteners	1	t	uploads/category-images/hardware_1.jpg
2	1	Plumbing	2	t	uploads/category-images/hardware_2.jpg
13	1	Welding Supplies	13	t	uploads/category-images/hardware_13.jpg
14	1	Adhesives & Sealants	14	t	uploads/category-images/hardware_14.jpg
15	1	Others	16	t	uploads/category-images/hardware_15.jpg
16	2	Building Materials	1	t	uploads/category-images/material_1.jpg
31	2	Others	16	t	uploads/category-images/material_17.jpg
48	4	Mobiles	1	t	uploads/category-images/electronic_1.jpg
18	2	Wood & Timber	3	t	uploads/category-images/material_7.jpg
19	2	Roofing	4	t	uploads/category-images/material_5.jpg
20	2	Civil Supplies	5	t	uploads/category-images/material_8.jpg
21	2	Cement & Concrete	6	t	uploads/category-images/material_9.jpg
22	2	Sand & Aggregates	7	t	uploads/category-images/material_9.jpg
17	2	Steel & Metals	2	t	uploads/category-images/material_4.jpg
23	2	Bricks & Blocks	8	t	uploads/category-images/material_2.jpg
24	2	Doors & Windows	9	t	uploads/category-images/material_24.jpg
25	2	Plywood & Laminates	10	t	uploads/category-images/material_20.jpg
26	2	Gypsum & Plaster	11	t	uploads/category-images/material_18.jpg
27	2	Tiles & Flooring	12	t	uploads/category-images/material_15.jpg
3	1	Electrical	3	t	uploads/category-images/hardware_3.jpg
4	1	Paints	4	t	uploads/category-images/hardware_4.jpg
6	1	Hand Tools	6	t	uploads/category-images/hardware_7.jpg
5	1	Power Tools	5	t	uploads/category-images/hardware_6.jpg
7	1	Hardware Accessories	7	t	uploads/category-images/hardware_16.jpg
8	1	Ladders & Platforms	8	t	uploads/category-images/hardware_9.jpg
10	1	Cutting Tools	10	t	uploads/category-images/hardware_10.jpg
9	1	Measuring Tools	9	t	uploads/category-images/hardware_17.jpg
11	1	Safety Equipment	11	t	uploads/category-images/hardware_11.jpg
12	1	Garden Tools	12	t	uploads/category-images/hardware_12.jpg
28	2	Structural Steel	13	t	uploads/category-images/material_22.jpg
29	2	Pipes & Fittings	14	t	uploads/category-images/material_21.jpg
30	2	Construction Chemicals	15	t	uploads/category-images/material_23.jpg
49	4	Tablets	2	t	uploads/category-images/electronic_2.jpg
108	7	Accessory	4	t	uploads/category-images/clothe_5.jpg
50	4	Laptops	3	t	uploads/category-images/electronic_3.jpg
51	4	Desktops	4	t	uploads/category-images/electronic_4.jpg
52	4	Televisions	5	t	uploads/category-images/electronic_4.jpg
53	4	Home Appliances	6	t	uploads/category-images/electronic_5.jpg
54	4	Kitchen Appliances	7	t	uploads/category-images/electronic_6.jpg
55	4	Audio Systems	8	t	uploads/category-images/electronic_13.jpg
109	7	Dresses	5	t	uploads/category-images/clothe_4.jpg
56	4	Cameras	9	t	uploads/category-images/electronic_7.jpg
57	4	Gaming Consoles	10	t	uploads/category-images/electronic_14.jpg
58	4	Accessories	11	t	uploads/category-images/electronic_15.jpg
59	4	Networking	12	t	uploads/category-images/electronic_16.jpg
60	4	Smart Home	13	t	uploads/category-images/electronic_11.jpg
65	4	Headphones	18	t	uploads/category-images/electronic_17.jpg
62	4	Printers & Scanners	15	t	uploads/category-images/electronic_18.jpg
63	4	Storage Devices	16	t	uploads/category-images/electronic_19.jpg
64	4	Speakers	17	t	uploads/category-images/electronic_20.jpg
61	4	Wearables	14	t	uploads/category-images/electronic_21.jpg
66	4	Power Banks	19	t	uploads/category-images/electronic_9.jpg
67	4	Fans & Coolers	20	t	uploads/category-images/electronic_22.jpg
68	4	Others	21	t	uploads/category-images/electronic_23.jpg
69	6	Staples	1	t	uploads/category-images/stationary_15.jpg
70	5	Snacks	2	t	uploads/category-images/grocery_2.jpg
72	5	Household	4	t	uploads/category-images/grocery_4.jpg
71	5	Beverages	3	t	uploads/category-images/grocery_1.jpg
74	5	Bakery	6	t	uploads/category-images/grocery_5.jpg
75	5	Spices & Masalas	7	t	uploads/category-images/grocery_6.jpg
77	5	Personal Care	9	t	uploads/category-images/grocery_7.jpg
73	5	Dairy	5	t	uploads/category-images/grocery_13.jpg
76	5	Cooking Essentials	8	t	uploads/category-images/grocery_15.jpg
78	5	Baby Products	10	t	uploads/category-images/grocery_14.jpg
79	5	Health & Wellness	11	t	uploads/category-images/grocery_16.jpg
80	5	Frozen Foods	12	t	uploads/category-images/grocery_17.jpg
81	5	Ready to Eat	13	t	uploads/category-images/grocery_18.jpg
82	5	Organic Products	14	t	uploads/category-images/grocery_19.jpg
83	5	Fruits & Vegetables	15	t	uploads/category-images/grocery_20.jpg
84	5	Sweets & Mithai	16	t	uploads/category-images/grocery_11.jpg
107	7	Suits	3	t	uploads/category-images/clothe_3.jpg
34	3	Kitchen	3	t	uploads/category-images/image_3.jpg
35	3	Office	4	t	uploads/category-images/image_4.jpg
40	3	Beds	9	t	uploads/category-images/image_9.jpg
38	3	Mattresses	7	t	uploads/category-images/image_7.jpg
37	3	Dining Room	6	t	uploads/category-images/image_8.jpg
41	3	Wardrobes	10	t	uploads/category-images/image_10.jpg
45	3	TV Units & Entertainment	14	t	uploads/category-images/image_16.jpg
43	3	Chairs	12	t	uploads/category-images/image_13.jpg
46	3	Bookcases & Shelves	15	t	uploads/category-images/image_18.jpg
33	3	Bedroom	2	t	uploads/category-images/image_2.jpg
36	3	Outdoor	5	t	uploads/category-images/image_5.jpg
39	3	Sofas & Recliners	8	t	uploads/category-images/image_1.jpg
42	3	Tables	11	t	uploads/category-images/image_11.jpg
44	3	Cabinets & Storage	13	t	uploads/category-images/image_14.jpg
32	3	Living Room	1	t	uploads/category-images/image_6.jpg
47	3	Others	16	t	uploads/category-images/image_20.jpg
85	5	Pet Supplies	17	t	uploads/category-images/grocery_10.jpg
86	5	Others	18	t	uploads/category-images/grocery_21.jpg
87	6	Office Supplies	1	t	uploads/category-images/stationary_2.jpg
99	6	Calculators	13	t	uploads/category-images/stationary_23.jpg
100	6	Presentation Supplies	14	t	uploads/category-images/stationary_9.jpg
88	6	School Supplies	2	t	uploads/category-images/stationary_19.jpg
89	6	Art Materials	3	t	uploads/category-images/stationary_24.jpg
120	7	Ethnic Wear	16	t	uploads/category-images/clothe_32.jpg
90	6	Packaging	4	t	uploads/category-images/stationary_4.jpg
91	6	Writing Instruments	5	t	uploads/category-images/stationary_5.jpg
92	6	Notebooks & Diaries	6	t	uploads/category-images/stationary_6.jpg
93	6	Files & Folders	7	t	uploads/category-images/stationary_7.jpg
94	6	Desk Accessories	8	t	uploads/category-images/stationary_8.jpg
95	6	Printer Supplies	9	t	uploads/category-images/stationary_20.jpg
96	6	Drawing & Drafting	10	t	uploads/category-images/stationary_10.jpg
97	6	Craft Supplies	11	t	uploads/category-images/stationary_21.jpg
98	6	Whiteboards & Boards	12	t	uploads/category-images/stationary_22.jpg
101	6	Correction Supplies	15	t	uploads/category-images/stationary_25.jpg
102	6	Sticky Notes & Tapes	16	t	uploads/category-images/stationary_26.jpg
103	6	Envelopes & Labels	17	t	uploads/category-images/stationary_27.jpg
104	6	Others	18	t	uploads/category-images/stationary_28.jpg
105	7	Shirts	1	t	uploads/category-images/clothe_1.jpg
106	7	Pants	2	t	uploads/category-images/clothe_2.jpg
110	7	Tops	6	t	uploads/category-images/clothe_9.jpg
111	7	Bottoms	7	t	uploads/category-images/clothe_7.jpg
112	7	Lingerie	8	t	uploads/category-images/clothe_6.jpg
121	7	Formal Wear	17	t	uploads/category-images/clothe_33.jpg
113	7	Infants	9	t	uploads/category-images/clothe_28.jpg
114	7	Toddlers	10	t	uploads/category-images/clothe_29.jpg
133	7	Bridal Wear	29	t	uploads/category-images/clothe_40.jpg
122	7	Casual Wear	18	t	uploads/category-images/clothe_34.jpg
116	7	Gym	12	t	uploads/category-images/clothe_31.jpg
134	7	Party Wear	30	t	uploads/category-images/clothe_41.jpg
118	7	Women Sportswear	14	t	uploads/category-images/clothe_10.jpg
119	7	Mens Sportswear	15	t	uploads/category-images/clothe_30.jpg
123	7	Sleepwear	19	t	uploads/category-images/clothe_19.jpg
124	7	Swimwear	20	t	uploads/category-images/clothe_18.jpg
125	7	Winter Wear	21	t	uploads/category-images/clothe_11.jpg
126	7	Footwear	22	t	uploads/category-images/clothe_21.jpg
127	7	Bags	23	t	uploads/category-images/clothe_16.jpg
115	7	Juniors	11	t	uploads/category-images/clothe_27.jpg
129	7	Sunglasses	25	t	uploads/category-images/clothe_36.jpg
130	7	Watches	26	t	uploads/category-images/clothe_25.jpg
128	7	Jewelry	24	t	uploads/category-images/clothe_37.jpg
117	7	Outerwear	13	t	uploads/category-images/clothe_38.jpg
131	7	Innerwear	27	t	uploads/category-images/clothe_39.jpg
135	7	Others	31	t	uploads/category-images/clothe_42.jpg
132	7	Plus Size	28	t	uploads/category-images/clothe_43.jpg
136	9	Delivery	1	t	uploads/departments/services/delivery.png
143	12	Makeup	1	t	
137	9	Home Service	2	t	uploads/departments/services/delivery/medicine.png
138	9	Beauty & Wellness	3	t	uploads/departments/services/delivery/bakery.png
140	9	Food & Dining	3	t	uploads/departments/services/food/food.jpeg
142	9	Other Essentials	3	t	uploads/departments/services/other/other_services.jpeg
139	9	Repair & Maintenance	3	t	uploads/departments/services/repair/repair.jpeg
141	9	Professional Services	3	t	uploads/departments/services/professional_service/professional_services.jpeg
151	15	Pet Food	1	t	/images/baby/sterilizer.jpg
152	15	Pet Toys	2	t	/images/baby/sterilizer.jpg
153	15	Grooming	3	t	/images/baby/sterilizer.jpg
154	15	Accessories	4	t	/images/baby/sterilizer.jpg
155	15	Beds & Travel	5	t	/images/baby/sterilizer.jpg
156	16	Fitness Equipment	1	t	/images/baby/sterilizer.jpg
157	16	Team Sports	2	t	/images/baby/sterilizer.jpg
158	16	Outdoor Sports	3	t	/images/baby/sterilizer.jpg
159	16	Sportswear	4	t	/images/baby/sterilizer.jpg
161	17	Organic Foods	1	t	/images/sports/track-pants.jpg
162	17	Personal Care	2	t	/images/sports/track-pants.jpg
163	17	Health Products	3	t	/images/sports/track-pants.jpg
164	17	Baby Care	4	t	/images/sports/track-pants.jpg
146	14	Feeding & Nursing	1	t	uploads/category-images/feeding.jpg
147	14	Diapering	2	t	uploads/category-images/diaper.jpg
148	14	Clothing	3	t	uploads/category-images/clothes.jpg
149	14	Bath & Care	4	t	uploads/category-images/bath_care.jpg
150	14	Baby Gear	5	t	uploads/category-images/baby_gear.jpg
\.


--
-- Data for Name: category_images; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.category_images (id, category_id, image_url, alt_text, sort_order, is_active, created_at, updated_at) FROM stdin;
1	1	/uploads/category-images/category_1.png	Fasteners Category	0	t	2025-12-10 16:12:44.092221	2025-12-10 16:12:44.092221
2	2	/uploads/category-images/category_2.png	Plumbing Category	0	t	2025-12-10 16:12:44.097468	2025-12-10 16:12:44.097468
3	3	/uploads/category-images/category_3.png	Electrical Category	0	t	2025-12-10 16:12:44.09767	2025-12-10 16:12:44.09767
4	4	/uploads/category-images/category_4.png	Paints Category	0	t	2025-12-10 16:12:44.097809	2025-12-10 16:12:44.097809
5	5	/uploads/category-images/category_5.png	Power Tools Category	0	t	2025-12-10 16:12:44.097988	2025-12-10 16:12:44.097988
6	6	/uploads/category-images/category_6.png	Hand Tools Category	0	t	2025-12-10 16:12:44.098109	2025-12-10 16:12:44.098109
7	7	/uploads/category-images/category_7.png	Hardware Accessories Category	0	t	2025-12-10 16:12:44.098217	2025-12-10 16:12:44.098217
8	8	/uploads/category-images/category_8.png	Ladders & Platforms Category	0	t	2025-12-10 16:12:44.098321	2025-12-10 16:12:44.098321
9	9	/uploads/category-images/category_9.png	Measuring Tools Category	0	t	2025-12-10 16:12:44.098415	2025-12-10 16:12:44.098415
10	10	/uploads/category-images/category_10.png	Cutting Tools Category	0	t	2025-12-10 16:12:44.098506	2025-12-10 16:12:44.098506
11	11	/uploads/category-images/category_11.png	Safety Equipment Category	0	t	2025-12-10 16:12:44.098643	2025-12-10 16:12:44.098643
12	12	/uploads/category-images/category_12.png	Garden Tools Category	0	t	2025-12-10 16:12:44.098789	2025-12-10 16:12:44.098789
13	13	/uploads/category-images/category_13.png	Welding Supplies Category	0	t	2025-12-10 16:12:44.098918	2025-12-10 16:12:44.098918
14	14	/uploads/category-images/category_14.png	Adhesives & Sealants Category	0	t	2025-12-10 16:12:44.099056	2025-12-10 16:12:44.099056
15	15	/uploads/category-images/category_15.png	Others Category	0	t	2025-12-10 16:12:44.099179	2025-12-10 16:12:44.099179
16	16	/uploads/category-images/category_16.png	Building Materials Category	0	t	2025-12-10 16:12:44.099305	2025-12-10 16:12:44.099305
17	17	/uploads/category-images/category_17.png	Steel & Metals Category	0	t	2025-12-10 16:12:44.099427	2025-12-10 16:12:44.099427
18	18	/uploads/category-images/category_18.png	Wood & Timber Category	0	t	2025-12-10 16:12:44.099582	2025-12-10 16:12:44.099582
19	19	/uploads/category-images/category_19.png	Roofing Category	0	t	2025-12-10 16:12:44.099719	2025-12-10 16:12:44.099719
20	20	/uploads/category-images/category_20.png	Civil Supplies Category	0	t	2025-12-10 16:12:44.099882	2025-12-10 16:12:44.099882
21	21	/uploads/category-images/category_21.png	Cement & Concrete Category	0	t	2025-12-10 16:12:44.100067	2025-12-10 16:12:44.100067
22	22	/uploads/category-images/category_22.png	Sand & Aggregates Category	0	t	2025-12-10 16:12:44.100219	2025-12-10 16:12:44.100219
23	23	/uploads/category-images/category_23.png	Bricks & Blocks Category	0	t	2025-12-10 16:12:44.100336	2025-12-10 16:12:44.100336
24	24	/uploads/category-images/category_24.png	Doors & Windows Category	0	t	2025-12-10 16:12:44.100453	2025-12-10 16:12:44.100453
25	25	/uploads/category-images/category_25.png	Plywood & Laminates Category	0	t	2025-12-10 16:12:44.100575	2025-12-10 16:12:44.100575
26	26	/uploads/category-images/category_26.png	Gypsum & Plaster Category	0	t	2025-12-10 16:12:44.100716	2025-12-10 16:12:44.100716
27	27	/uploads/category-images/category_27.png	Tiles & Flooring Category	0	t	2025-12-10 16:12:44.101017	2025-12-10 16:12:44.101017
28	28	/uploads/category-images/category_28.png	Structural Steel Category	0	t	2025-12-10 16:12:44.101142	2025-12-10 16:12:44.101142
29	29	/uploads/category-images/category_29.png	Pipes & Fittings Category	0	t	2025-12-10 16:12:44.101259	2025-12-10 16:12:44.101259
30	30	/uploads/category-images/category_30.png	Construction Chemicals Category	0	t	2025-12-10 16:12:44.101376	2025-12-10 16:12:44.101376
31	31	/uploads/category-images/category_31.png	Others Category	0	t	2025-12-10 16:12:44.101493	2025-12-10 16:12:44.101493
32	32	/uploads/category-images/category_32.png	Living Room Category	0	t	2025-12-10 16:12:44.101628	2025-12-10 16:12:44.101628
33	33	/uploads/category-images/category_33.png	Bedroom Category	0	t	2025-12-10 16:12:44.101744	2025-12-10 16:12:44.101744
34	34	/uploads/category-images/category_34.png	Kitchen Category	0	t	2025-12-10 16:12:44.101908	2025-12-10 16:12:44.101908
35	35	/uploads/category-images/category_35.png	Office Category	0	t	2025-12-10 16:12:44.102043	2025-12-10 16:12:44.102043
36	36	/uploads/category-images/category_36.png	Outdoor Category	0	t	2025-12-10 16:12:44.102142	2025-12-10 16:12:44.102142
37	37	/uploads/category-images/category_37.png	Dining Room Category	0	t	2025-12-10 16:12:44.10224	2025-12-10 16:12:44.10224
38	38	/uploads/category-images/category_38.png	Mattresses Category	0	t	2025-12-10 16:12:44.102403	2025-12-10 16:12:44.102403
39	39	/uploads/category-images/category_39.png	Sofas & Recliners Category	0	t	2025-12-10 16:12:44.102512	2025-12-10 16:12:44.102512
40	40	/uploads/category-images/category_40.png	Beds Category	0	t	2025-12-10 16:12:44.102652	2025-12-10 16:12:44.102652
41	41	/uploads/category-images/category_41.png	Wardrobes Category	0	t	2025-12-10 16:12:44.102801	2025-12-10 16:12:44.102801
42	42	/uploads/category-images/category_42.png	Tables Category	0	t	2025-12-10 16:12:44.102957	2025-12-10 16:12:44.102957
43	43	/uploads/category-images/category_43.png	Chairs Category	0	t	2025-12-10 16:12:44.103097	2025-12-10 16:12:44.103097
44	44	/uploads/category-images/category_44.png	Cabinets & Storage Category	0	t	2025-12-10 16:12:44.103235	2025-12-10 16:12:44.103235
45	45	/uploads/category-images/category_45.png	TV Units & Entertainment Category	0	t	2025-12-10 16:12:44.103382	2025-12-10 16:12:44.103382
46	46	/uploads/category-images/category_46.png	Bookcases & Shelves Category	0	t	2025-12-10 16:12:44.103512	2025-12-10 16:12:44.103512
47	47	/uploads/category-images/category_47.png	Others Category	0	t	2025-12-10 16:12:44.103644	2025-12-10 16:12:44.103644
48	48	/uploads/category-images/category_48.png	Mobiles Category	0	t	2025-12-10 16:12:44.103795	2025-12-10 16:12:44.103795
49	49	/uploads/category-images/category_49.png	Tablets Category	0	t	2025-12-10 16:12:44.103948	2025-12-10 16:12:44.103948
50	50	/uploads/category-images/category_50.png	Laptops Category	0	t	2025-12-10 16:12:44.10409	2025-12-10 16:12:44.10409
51	51	/uploads/category-images/category_51.png	Desktops Category	0	t	2025-12-10 16:12:44.104232	2025-12-10 16:12:44.104232
52	52	/uploads/category-images/category_52.png	Televisions Category	0	t	2025-12-10 16:12:44.104372	2025-12-10 16:12:44.104372
53	53	/uploads/category-images/category_53.png	Home Appliances Category	0	t	2025-12-10 16:12:44.104539	2025-12-10 16:12:44.104539
54	54	/uploads/category-images/category_54.png	Kitchen Appliances Category	0	t	2025-12-10 16:12:44.104836	2025-12-10 16:12:44.104836
55	55	/uploads/category-images/category_55.png	Audio Systems Category	0	t	2025-12-10 16:12:44.104994	2025-12-10 16:12:44.104994
56	56	/uploads/category-images/category_56.png	Cameras Category	0	t	2025-12-10 16:12:44.105138	2025-12-10 16:12:44.105138
57	57	/uploads/category-images/category_57.png	Gaming Consoles Category	0	t	2025-12-10 16:12:44.105275	2025-12-10 16:12:44.105275
58	58	/uploads/category-images/category_58.png	Accessories Category	0	t	2025-12-10 16:12:44.105724	2025-12-10 16:12:44.105724
59	59	/uploads/category-images/category_59.png	Networking Category	0	t	2025-12-10 16:12:44.105903	2025-12-10 16:12:44.105903
60	60	/uploads/category-images/category_60.png	Smart Home Category	0	t	2025-12-10 16:12:44.106038	2025-12-10 16:12:44.106038
61	61	/uploads/category-images/category_61.png	Wearables Category	0	t	2025-12-10 16:12:44.106167	2025-12-10 16:12:44.106167
62	62	/uploads/category-images/category_62.png	Printers & Scanners Category	0	t	2025-12-10 16:12:44.106311	2025-12-10 16:12:44.106311
63	63	/uploads/category-images/category_63.png	Storage Devices Category	0	t	2025-12-10 16:12:44.106479	2025-12-10 16:12:44.106479
64	64	/uploads/category-images/category_64.png	Speakers Category	0	t	2025-12-10 16:12:44.106601	2025-12-10 16:12:44.106601
65	65	/uploads/category-images/category_65.png	Headphones Category	0	t	2025-12-10 16:12:44.106717	2025-12-10 16:12:44.106717
66	66	/uploads/category-images/category_66.png	Power Banks Category	0	t	2025-12-10 16:12:44.106836	2025-12-10 16:12:44.106836
67	67	/uploads/category-images/category_67.png	Fans & Coolers Category	0	t	2025-12-10 16:12:44.106966	2025-12-10 16:12:44.106966
68	68	/uploads/category-images/category_68.png	Others Category	0	t	2025-12-10 16:12:44.107081	2025-12-10 16:12:44.107081
69	69	/uploads/category-images/category_69.png	Staples Category	0	t	2025-12-10 16:12:44.107195	2025-12-10 16:12:44.107195
70	70	/uploads/category-images/category_70.png	Snacks Category	0	t	2025-12-10 16:12:44.107306	2025-12-10 16:12:44.107306
71	71	/uploads/category-images/category_71.png	Beverages Category	0	t	2025-12-10 16:12:44.107418	2025-12-10 16:12:44.107418
72	72	/uploads/category-images/category_72.png	Household Category	0	t	2025-12-10 16:12:44.107533	2025-12-10 16:12:44.107533
73	73	/uploads/category-images/category_73.png	Dairy Category	0	t	2025-12-10 16:12:44.10766	2025-12-10 16:12:44.10766
74	74	/uploads/category-images/category_74.png	Bakery Category	0	t	2025-12-10 16:12:44.107787	2025-12-10 16:12:44.107787
75	75	/uploads/category-images/category_75.png	Spices & Masalas Category	0	t	2025-12-10 16:12:44.107915	2025-12-10 16:12:44.107915
76	76	/uploads/category-images/category_76.png	Cooking Essentials Category	0	t	2025-12-10 16:12:44.108042	2025-12-10 16:12:44.108042
77	77	/uploads/category-images/category_77.png	Personal Care Category	0	t	2025-12-10 16:12:44.108167	2025-12-10 16:12:44.108167
78	78	/uploads/category-images/category_78.png	Baby Products Category	0	t	2025-12-10 16:12:44.108426	2025-12-10 16:12:44.108426
79	79	/uploads/category-images/category_79.png	Health & Wellness Category	0	t	2025-12-10 16:12:44.108555	2025-12-10 16:12:44.108555
80	80	/uploads/category-images/category_80.png	Frozen Foods Category	0	t	2025-12-10 16:12:44.110005	2025-12-10 16:12:44.110005
81	81	/uploads/category-images/category_81.png	Ready to Eat Category	0	t	2025-12-10 16:12:44.110188	2025-12-10 16:12:44.110188
82	82	/uploads/category-images/category_82.png	Organic Products Category	0	t	2025-12-10 16:12:44.110312	2025-12-10 16:12:44.110312
83	83	/uploads/category-images/category_83.png	Fruits & Vegetables Category	0	t	2025-12-10 16:12:44.110451	2025-12-10 16:12:44.110451
84	84	/uploads/category-images/category_84.png	Sweets & Mithai Category	0	t	2025-12-10 16:12:44.110568	2025-12-10 16:12:44.110568
85	85	/uploads/category-images/category_85.png	Pet Supplies Category	0	t	2025-12-10 16:12:44.110674	2025-12-10 16:12:44.110674
86	86	/uploads/category-images/category_86.png	Others Category	0	t	2025-12-10 16:12:44.110774	2025-12-10 16:12:44.110774
87	87	/uploads/category-images/category_87.png	Office Supplies Category	0	t	2025-12-10 16:12:44.11087	2025-12-10 16:12:44.11087
88	88	/uploads/category-images/category_88.png	School Supplies Category	0	t	2025-12-10 16:12:44.110965	2025-12-10 16:12:44.110965
89	89	/uploads/category-images/category_89.png	Art Materials Category	0	t	2025-12-10 16:12:44.111058	2025-12-10 16:12:44.111058
90	90	/uploads/category-images/category_90.png	Packaging Category	0	t	2025-12-10 16:12:44.11115	2025-12-10 16:12:44.11115
91	91	/uploads/category-images/category_91.png	Writing Instruments Category	0	t	2025-12-10 16:12:44.111242	2025-12-10 16:12:44.111242
92	92	/uploads/category-images/category_92.png	Notebooks & Diaries Category	0	t	2025-12-10 16:12:44.111341	2025-12-10 16:12:44.111341
93	93	/uploads/category-images/category_93.png	Files & Folders Category	0	t	2025-12-10 16:12:44.111439	2025-12-10 16:12:44.111439
94	94	/uploads/category-images/category_94.png	Desk Accessories Category	0	t	2025-12-10 16:12:44.111536	2025-12-10 16:12:44.111536
95	95	/uploads/category-images/category_95.png	Printer Supplies Category	0	t	2025-12-10 16:12:44.111633	2025-12-10 16:12:44.111633
96	96	/uploads/category-images/category_96.png	Drawing & Drafting Category	0	t	2025-12-10 16:12:44.111742	2025-12-10 16:12:44.111742
97	97	/uploads/category-images/category_97.png	Craft Supplies Category	0	t	2025-12-10 16:12:44.111864	2025-12-10 16:12:44.111864
98	98	/uploads/category-images/category_98.png	Whiteboards & Boards Category	0	t	2025-12-10 16:12:44.111975	2025-12-10 16:12:44.111975
99	99	/uploads/category-images/category_99.png	Calculators Category	0	t	2025-12-10 16:12:44.112075	2025-12-10 16:12:44.112075
100	100	/uploads/category-images/category_100.png	Presentation Supplies Category	0	t	2025-12-10 16:12:44.112175	2025-12-10 16:12:44.112175
101	101	/uploads/category-images/category_101.png	Correction Supplies Category	0	t	2025-12-10 16:12:44.112272	2025-12-10 16:12:44.112272
102	102	/uploads/category-images/category_102.png	Sticky Notes & Tapes Category	0	t	2025-12-10 16:12:44.11237	2025-12-10 16:12:44.11237
103	103	/uploads/category-images/category_103.png	Envelopes & Labels Category	0	t	2025-12-10 16:12:44.112467	2025-12-10 16:12:44.112467
104	104	/uploads/category-images/category_104.png	Others Category	0	t	2025-12-10 16:12:44.112563	2025-12-10 16:12:44.112563
105	105	/uploads/category-images/category_105.png	Shirts Category	0	t	2025-12-10 16:12:44.112658	2025-12-10 16:12:44.112658
106	106	/uploads/category-images/category_106.png	Pants Category	0	t	2025-12-10 16:12:44.112885	2025-12-10 16:12:44.112885
107	107	/uploads/category-images/category_107.png	Suits Category	0	t	2025-12-10 16:12:44.112979	2025-12-10 16:12:44.112979
108	108	/uploads/category-images/category_108.png	Accessory Category	0	t	2025-12-10 16:12:44.113086	2025-12-10 16:12:44.113086
109	109	/uploads/category-images/category_109.png	Dresses Category	0	t	2025-12-10 16:12:44.113184	2025-12-10 16:12:44.113184
110	110	/uploads/category-images/category_110.png	Tops Category	0	t	2025-12-10 16:12:44.113279	2025-12-10 16:12:44.113279
111	111	/uploads/category-images/category_111.png	Bottoms Category	0	t	2025-12-10 16:12:44.11338	2025-12-10 16:12:44.11338
112	112	/uploads/category-images/category_112.png	Lingerie Category	0	t	2025-12-10 16:12:44.113494	2025-12-10 16:12:44.113494
113	113	/uploads/category-images/category_113.png	Infants Category	0	t	2025-12-10 16:12:44.113589	2025-12-10 16:12:44.113589
114	114	/uploads/category-images/category_114.png	Toddlers Category	0	t	2025-12-10 16:12:44.113688	2025-12-10 16:12:44.113688
115	115	/uploads/category-images/category_115.png	Juniors Category	0	t	2025-12-10 16:12:44.113857	2025-12-10 16:12:44.113857
116	116	/uploads/category-images/category_116.png	Gym Category	0	t	2025-12-10 16:12:44.113964	2025-12-10 16:12:44.113964
117	117	/uploads/category-images/category_117.png	Outerwear Category	0	t	2025-12-10 16:12:44.114063	2025-12-10 16:12:44.114063
118	118	/uploads/category-images/category_118.png	Women Sportswear Category	0	t	2025-12-10 16:12:44.114158	2025-12-10 16:12:44.114158
119	119	/uploads/category-images/category_119.png	Mens Sportswear Category	0	t	2025-12-10 16:12:44.114254	2025-12-10 16:12:44.114254
120	120	/uploads/category-images/category_120.png	Ethnic Wear Category	0	t	2025-12-10 16:12:44.114356	2025-12-10 16:12:44.114356
121	121	/uploads/category-images/category_121.png	Formal Wear Category	0	t	2025-12-10 16:12:44.114452	2025-12-10 16:12:44.114452
122	122	/uploads/category-images/category_122.png	Casual Wear Category	0	t	2025-12-10 16:12:44.11466	2025-12-10 16:12:44.11466
123	123	/uploads/category-images/category_123.png	Sleepwear Category	0	t	2025-12-10 16:12:44.114759	2025-12-10 16:12:44.114759
124	124	/uploads/category-images/category_124.png	Swimwear Category	0	t	2025-12-10 16:12:44.114859	2025-12-10 16:12:44.114859
125	125	/uploads/category-images/category_125.png	Winter Wear Category	0	t	2025-12-10 16:12:44.114972	2025-12-10 16:12:44.114972
126	126	/uploads/category-images/category_126.png	Footwear Category	0	t	2025-12-10 16:12:44.115069	2025-12-10 16:12:44.115069
127	127	/uploads/category-images/category_127.png	Bags Category	0	t	2025-12-10 16:12:44.115165	2025-12-10 16:12:44.115165
128	128	/uploads/category-images/category_128.png	Jewelry Category	0	t	2025-12-10 16:12:44.115564	2025-12-10 16:12:44.115564
129	129	/uploads/category-images/category_129.png	Sunglasses Category	0	t	2025-12-10 16:12:44.115759	2025-12-10 16:12:44.115759
130	130	/uploads/category-images/category_130.png	Watches Category	0	t	2025-12-10 16:12:44.115904	2025-12-10 16:12:44.115904
131	131	/uploads/category-images/category_131.png	Innerwear Category	0	t	2025-12-10 16:12:44.116058	2025-12-10 16:12:44.116058
132	132	/uploads/category-images/category_132.png	Plus Size Category	0	t	2025-12-10 16:12:44.116195	2025-12-10 16:12:44.116195
133	133	/uploads/category-images/category_133.png	Bridal Wear Category	0	t	2025-12-10 16:12:44.116323	2025-12-10 16:12:44.116323
134	134	/uploads/category-images/category_134.png	Party Wear Category	0	t	2025-12-10 16:12:44.116521	2025-12-10 16:12:44.116521
135	135	/uploads/category-images/category_135.png	Others Category	0	t	2025-12-10 16:12:44.11665	2025-12-10 16:12:44.11665
\.


--
-- Data for Name: countries; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.countries (id, country_name, iso_code) FROM stdin;
1	India	IN
2	United States	US
3	United Kingdom	GB
4	Canada	CA
5	Australia	AU
6	Germany	DE
7	France	FR
8	Japan	JP
9	China	CN
10	Brazil	BR
\.


--
-- Data for Name: coupon_uses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.coupon_uses (coupon_uses_id, coupon_id, user_id, used_at) FROM stdin;
\.


--
-- Data for Name: coupons; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.coupons (coupon_id, coupon_name, coupon_code, expire_date, description, discount_rate, minimum_cart_price, image, block_status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: departments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.departments (id, name, slug, sort_order, is_active, image_url) FROM stdin;
1	Hardware	hardware	1	t	uploads/category-images/generated_image_departments_1.jpg
2	Material Service	material-service	2	t	uploads/category-images/generated_image_departments_3.jpg
3	Furniture	furniture	3	t	uploads/category-images/generated_image_departments_5.jpg
4	Electronics	electronics	4	t	uploads/category-images/generated_image_departments_4.jpg
5	Grocery	grocery	5	t	uploads/category-images/generated_image_departments_8.jpg
6	Stationery	stationery	6	t	uploads/category-images/generated_image_departments_11.jpg
7	Clothing	clothing	7	t	uploads/category-images/generated_image_departments_12.jpg
9	Services	service	8	t	uploads/departments/services/services.jpeg
14	Baby Products	baby	11	t	uploads/departments/baby_products.jpg
15	Pet Supplies	pet	12	t	uploads/departments/pet_supplies.jpg
16	Sports	sports	13	t	uploads/departments/sports.jpg
17	Organic Products	organic	14	t	uploads/departments/organic_products.jpg
12	Beauty	beauty	9	f	uploads/departments/beauty_products.jpg
13	Veg & Fruits	fruits	10	f	uploads/departments/veg_fruits.jpg
18	Sweet Shop	sweet	15	f	uploads/departments/sweet_shop.jpg
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.notifications (id, sender_type, receiver_type, type, sender_id, title, message, body, is_read, receiver_id, category_id, product_id, variation_id, shop_id, user_id, admin_id, order_id, offer_id, notification_meta_data, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: offer_categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.offer_categories (id, offer_id, category_id, sort_order, is_active) FROM stdin;
\.


--
-- Data for Name: offer_products; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.offer_products (id, offer_id, product_item_id, sort_order, is_active) FROM stdin;
16	1	36	0	t
17	1	53	0	t
25	7	61	0	t
27	7	63	0	t
33	5	65	0	t
34	15	66	0	t
35	6	79	0	t
36	7	72	0	t
38	26	71	0	t
\.


--
-- Data for Name: offers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.offers (id, name, description, discount_rate, start_date, end_date, created_at, updated_at, sort_order, is_active, offer_type, image, thumbnail) FROM stdin;
5	Eid Special Savings	Celebrate Eid with generous discounts	19	2026-03-18 00:00:00+00	2026-03-22 23:59:59+00	\N	\N	5	t	percentage	uploads/offers/offer_5.png	uploads/offers/thumbnail/sale_icon.png
8	Independence Day Mega Sale	Big savings for Independence Day shoppers	26	2026-08-13 01:00:00+01	2026-08-17 00:59:59+01	\N	\N	8	t	percentage	uploads/offers/offer_8.png	uploads/offers/thumbnail/sale_icon.png
9	Diwali Lights and Deals	Festival of Lights and offers	30	2026-11-05 00:00:00+00	2026-11-10 23:59:59+00	\N	\N	9	t	percentage	uploads/offers/offer_9.png	uploads/offers/thumbnail/sale_icon.png
21	Christmas Offer	Christmas Offer	20	2025-12-25 07:03:00+00	2025-12-31 20:03:06.789912+00	2025-12-24 20:15:30.354559+00	2025-12-24 20:15:30.354559+00	0	t	percentage	uploads/offers/offer_27c3bc23-9bb8-4a42-b209-948be0a0ae64.png	uploads/offers/thumbnail/offer_thumb_27c3bc23-9bb8-4a42-b209-948be0a0ae64.png
\.


--
-- Data for Name: order_lines; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_lines (id, product_item_id, shop_order_id, qty, price) FROM stdin;
\.


--
-- Data for Name: order_returns; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_returns (id, shop_order_id, request_date, return_reason, refund_amount, is_approved, return_date, approval_date, admin_comment) FROM stdin;
\.


--
-- Data for Name: order_statuses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_statuses (id, status) FROM stdin;
1	payment pending
2	order placed
3	order cancelled
4	order delivered
5	return requested
6	return approved
7	return cancelled
8	order returned
\.


--
-- Data for Name: otp_sessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.otp_sessions (id, otp_id, user_id, admin_id, user_type, phone, expire_at) FROM stdin;
1	6e197d4d-fc9e-4f01-907d-dfc7cd85805e	0	1	Seller	9886569962	2026-04-05 19:14:23.402042+01
2	f9afeec0-607e-4797-8c01-69c89e97f696	13	0		9549115789	2026-04-21 09:25:45.161376+01
3	77d7fc73-fc08-4a04-8a14-911ebc09768d	13	0		9886569962	2026-04-21 09:27:59.897594+01
4	3d799113-e0ed-4673-a74e-ec6dc1eca741	1	0		9886569962	2026-04-21 10:09:11.889422+01
5	af5f6062-2539-46c1-a84f-851af4bb5689	0	4	Seller	9886569962	2026-04-22 13:50:08.734901+01
6	976bd0c5-f4b4-44e2-b8ea-7861db16d65b	0	4	Seller	9886569962	2026-04-22 13:50:09.659673+01
7	e05138a9-8729-40af-ad2d-280976baba06	2	0		stringstri	2026-04-24 11:14:49.640332+01
8	e1d93844-f821-4a3f-9cbc-5ad54a6e3b04	3	0		8302827722	2026-04-24 11:21:31.7473+01
9	266f1796-58bd-49cb-a0f2-c1a815b5f4f3	4	0		83028277229	2026-04-24 14:27:18.681144+01
10	330b72c9-f19f-444a-8258-5af7c9706be8	5	0		83028276722	2026-04-24 14:47:20.899694+01
11	3f3e8238-b231-48c2-9976-834ec16af0a8	6	0		8302827723	2026-04-24 14:48:13.523427+01
12	03bf4a27-c87f-492d-95e6-7b57287d21a0	7	0		8302827722	2026-04-24 14:53:45.214584+01
13	1152c8e5-c3a5-4882-8436-45268c3bb683	7	0		8302827722	2026-04-24 14:54:49.420407+01
14	95ad2d30-8269-433f-bd8e-e37476ba9208	7	0		8302827722	2026-04-24 14:55:21.342027+01
15	d04a5562-65a0-43ca-b283-3b6086601d75	7	0		8302827722	2026-04-24 14:56:08.896769+01
16	6705cfce-f2e2-46a4-a273-9fce4550873b	7	0		8302827722	2026-04-24 14:59:39.879983+01
17	fdc71fe7-1b83-4340-b5c3-210fe532c182	7	0		8302827722	2026-04-24 15:02:53.273368+01
18	bb735892-613b-488d-9101-5e9e9b637645	7	0		8302827722	2026-04-24 15:28:04.59991+01
19	9ae15a59-e54e-4ced-bdda-058add0b224a	8	0		8302827722	2026-04-25 06:29:51.048884+01
20	aa6b789c-e62e-4d34-93b6-fb0bebff3a24	8	0		8302827722	2026-04-25 07:30:05.492991+01
21	20cd3cc9-45d3-41ef-b8da-062fafc12553	9	0		8302827722	2026-04-25 16:11:31.95313+01
22	d769c69f-7798-4485-9198-54d39cd9e7db	10	0		8302827722	2026-04-25 17:36:12.999452+01
23	546be764-6d0c-4fa1-b231-16d0acef2d6f	11	0		8302827722	2026-04-25 18:13:02.272384+01
24	b0cb61b2-d748-4f05-a873-848998503358	12	0		8302827722	2026-04-25 18:37:49.225411+01
25	633ab0d3-6a3a-4e5e-85b2-ca0fe3e1d46e	13	0		8302827722	2026-04-25 20:08:02.581949+01
26	9ba95352-18d5-4318-a49c-20267a414bc7	13	0		8302827722	2026-04-25 20:11:09.979261+01
27	14ad37d0-cedd-49c2-8d39-07acb91a2a21	13	0		8302827722	2026-04-25 21:13:08.885031+01
28	88d8962b-0ef4-4ee8-9e14-3378c319b158	14	0		8302827722	2026-04-25 22:38:19.827561+01
29	e558853c-0e42-41ea-a3d5-77c8f1d3b70f	14	0		8302827722	2026-04-25 22:49:29.199359+01
30	5b8ac4d6-79cc-4197-9f04-d8f17b428520	14	0		8302827722	2026-04-25 22:49:43.544446+01
31	b366ae44-229c-4319-abc0-2b95378cbe11	15	0		9886569962	2026-04-28 13:13:36.35748+01
32	3e3e3899-1afe-4722-9ffd-35061e0b0eb9	15	0		9886569962	2026-04-28 13:14:12.132977+01
33	8db794ef-076b-456a-a6f3-f426e9754869	15	0		9886569962	2026-04-28 13:18:19.496375+01
34	f4972d3c-38dc-4264-b82d-82b282beb1dd	15	0		9886569962	2026-04-28 13:25:45.014773+01
35	c9261ae4-b2cd-45cf-bddd-4c16e5ec36b0	15	0		9886569962	2026-04-28 13:28:13.210281+01
36	2c25dd80-36fc-46f4-bc5b-ce4cb84eaae7	16	0		stringstri	2026-04-29 08:12:10.295113+01
37	5dfc48c7-578d-4d89-b3be-0c4691de35fb	17	0		8343434343	2026-04-29 08:13:24.948094+01
38	353d5c68-733d-4dd0-ac1b-0689686911b6	14	0		8302827722	2026-04-29 08:13:39.782822+01
39	bded6f8e-db6f-4a01-9853-195159f9ed7b	14	0		8302827722	2026-04-29 08:15:10.664163+01
40	84a14426-da20-4ad8-90a8-8dd5bf08685a	14	0		8302827722	2026-04-29 08:15:56.816603+01
41	afbd6f61-2566-4d5d-a1d1-c81a757a5f96	14	0		8302827722	2026-04-29 08:20:12.516038+01
42	c386d0c7-357f-4183-8f59-2b36166738a4	14	0		8302827722	2026-04-29 08:41:45.59852+01
43	e6ba792a-df4a-4574-9cfb-ab5ff858ceb5	14	0		8302827722	2026-04-29 09:17:24.998265+01
44	f4788edb-d55d-4786-9def-e420f09fbb45	15	0		9886569962	2026-05-04 09:35:50.679348+01
45	1fde9153-25cb-41ce-90f7-bdcc3fd4eb56	15	0		9886569962	2026-05-04 09:47:34.971375+01
46	ff33301e-ffbe-4627-a06f-43393a0408a1	15	0		9886569962	2026-05-04 09:55:21.309619+01
47	30150dd5-c4b4-48af-8bbc-56e1cc3c1e13	15	0		9886569962	2026-05-07 18:27:54.210144+01
48	4d2be370-c002-4640-8d16-b1e21fa7df13	15	0		9886569962	2026-05-08 07:28:13.440369+01
49	cd943022-bd8f-4636-a3a4-2e2ffd40bc7f	15	0		9886569962	2026-05-08 07:31:11.935111+01
50	bb603781-17ed-4020-8e4e-14bce323c663	15	0		9886569962	2026-05-08 07:35:30.565077+01
51	5e05e8ac-5664-48f6-9dca-632e29bac299	15	0		9886569962	2026-05-08 07:41:40.849587+01
52	1d63478c-1429-4efc-84c6-1bad23821c2e	15	0		9886569962	2026-05-08 08:59:49.465215+01
53	c68fedf6-c277-4bf1-8c7f-9bed358cfa2f	0	6	Seller	8302827722	2026-05-10 14:11:18.381276+01
54	7ac1a15d-b8cc-4dae-bf85-4df1b39fa33d	0	6	Seller	8302827722	2026-05-11 06:06:17.704222+01
55	61cfe2ca-cf75-48ea-a1fa-caed496a6957	15	0		9886569962	2026-05-24 19:26:26.278103+01
56	3a8a158a-c7a2-47ef-830c-e160cb7f62df	15	0		9886569962	2026-05-26 11:47:26.894179+01
57	f65695bb-23a0-419a-a908-9d801fe83105	0	6	Seller	8302827722	2026-06-06 10:48:10.352892+01
58	65a4c826-6571-44f1-be34-3e4529e2fdcd	0	6	Seller	8302827722	2026-06-06 12:15:33.179056+01
59	95ce40dc-ca56-4614-b732-b38736960add	0	6	Seller	8302827722	2026-06-06 12:24:17.395082+01
60	23c44e38-15d4-4263-bb07-408ea84606ed	0	6	Seller	8302827722	2026-06-06 12:25:04.453658+01
61	ad2dba5b-1a82-4a6c-a933-2ca4bb0ae312	0	6	Seller	8302827722	2026-06-06 12:26:34.409812+01
62	f1ab2383-4a0d-4b6d-a30a-7f7f5f162c45	0	6	Seller	8302827722	2026-02-01 11:45:13.945519+00
63	da89a3e0-baed-437d-89f0-1ce55e892f4f	0	6	Seller	8302827722	2026-02-01 11:47:32.605365+00
64	a0bdbc73-1208-4dd7-8cf1-29d28c183814	0	6	Seller	8302827722	2027-02-01 11:54:51.800375+00
65	721bdf6c-345f-4a62-b942-13eeb7709a59	0	6	Seller	8302827722	2027-02-01 11:55:49.578779+00
66	4309711c-1ede-4b82-afd6-d58336c17ccf	0	6	Seller	8302827722	2027-02-01 12:00:39.211894+00
67	3077e23b-3792-4c75-806d-6c68b691e011	0	6	Seller	8302827722	2027-02-01 12:04:23.928695+00
68	a6201b9d-4a81-4dc6-8931-8168f376d501	0	6	Seller	8302827722	2027-02-01 12:24:16.108098+00
69	153bf18d-0cf5-4688-8873-6c06319816cc	0	6	Seller	8302827722	2027-02-01 12:53:49.445085+00
70	c1e479d4-6d47-4e4f-9eae-725906fa5908	0	6	Seller	8302827722	2027-02-01 13:02:45.54444+00
71	1b42decd-afe4-4172-9422-92d188871aef	0	6	Seller	8302827722	2027-02-01 13:05:20.51133+00
72	3a18c6aa-b05f-47e9-814e-1eb89340896a	0	6	Seller	8302827722	2027-02-01 13:05:47.728932+00
73	1799ad82-82f1-487b-bd7a-73c4918b69dd	0	6	Seller	8302827722	2027-02-01 13:07:17.777663+00
74	6446f859-0118-4820-910a-aaf97b2b1f9c	0	6	Seller	8302827722	2027-02-01 13:13:18.81556+00
75	e7c73499-0750-4416-8f52-e19265102858	0	6	Seller	8302827722	2027-02-01 13:17:29.020968+00
76	4d5c2b90-d025-4afe-bb50-f7833b6aa158	0	6	Seller	8302827722	2027-02-01 13:21:47.317235+00
77	f4ab7648-79de-48c2-ab70-03ec3a8e1683	0	7	Seller	8302827722	2027-02-01 13:23:28.003264+00
78	a10d52d7-eec3-45b1-8db8-33ccaa86b309	0	9	Seller	9549115670	2027-02-16 18:01:41.253635+00
79	08c44afb-656d-4987-a0ab-b8faa272b0d5	0	9	Seller	9549115670	2027-02-16 18:02:41.187424+00
\.


--
-- Data for Name: payment_methods; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.payment_methods (id, name, block_status, maximum_amount) FROM stdin;
1	cod	f	20000
2	razor pay	f	50000
3	stripe	f	50000
\.


--
-- Data for Name: product_configurations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product_configurations (product_item_id, variation_option_id, sort_order, is_active) FROM stdin;
\.


--
-- Data for Name: product_images; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product_images (id, product_item_id, image, image_url, shop_id, product_id, alt_text, sort_order, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: product_item_filter_types; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product_item_filter_types (id, filter_name, shop_id, created_at, updated_at) FROM stdin;
1	All	20	2026-01-03 19:48:41.87941+00	2026-01-03 19:48:41.87941+00
2	Offer	20	2026-01-03 19:48:41.87941+00	2026-01-03 19:48:41.87941+00
\.


--
-- Data for Name: product_item_views; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product_item_views (id, product_item_id, shop_id, admin_id, viewed_at, view_count, created_at, updated_at) FROM stdin;
41	67	20	1	2026-02-04 18:37:53.230452+00	5	2026-02-04 18:13:58.714565+00	\N
42	68	20	1	2026-02-07 22:02:32.66004+00	4	2026-02-07 21:03:06.946383+00	\N
3	39	20	1	2026-01-22 13:05:20.820863+00	29	2025-12-17 13:34:56.415129+00	2025-12-17 13:34:56.415129+00
16	49	20	1	2026-01-22 13:46:04.269896+00	32	2026-01-03 06:38:01.772728+00	\N
20	47	20	1	2026-01-11 19:23:01.913125+00	5	2026-01-03 06:38:40.918349+00	\N
18	48	20	1	2026-01-22 13:49:05.063807+00	18	2026-01-03 06:38:20.924548+00	\N
4	38	20	1	2026-01-22 14:22:39.862455+00	85	2025-12-17 13:34:56.415129+00	2025-12-17 13:34:56.415129+00
25	51	20	1	2026-01-22 15:04:45.571668+00	2	2026-01-20 09:36:51.661572+00	\N
43	10	20	1	2026-02-10 03:07:01.581651+00	5	2026-02-09 14:15:10.976054+00	\N
39	66	20	1	2026-02-15 04:53:07.245139+00	4	2026-02-03 19:00:11.548837+00	\N
17	37	20	1	2026-01-11 19:23:32.243694+00	5	2026-01-03 06:38:10.159449+00	\N
21	46	20	1	2026-01-14 14:07:49.359803+00	8	2026-01-03 19:45:12.286277+00	\N
22	45	20	1	2026-01-14 18:55:11.879382+00	7	2026-01-05 12:51:53.664616+00	\N
26	52	20	1	2026-01-23 06:12:35.113949+00	22	2026-01-22 04:28:16.464799+00	\N
24	53	20	1	2026-01-23 09:45:55.158484+00	40	2026-01-20 04:59:36.109151+00	\N
23	50	20	1	2026-01-20 04:20:31.98948+00	1	2026-01-20 04:20:31.98948+00	\N
19	36	20	1	2026-01-23 09:46:13.602038+00	44	2026-01-03 06:38:29.825882+00	\N
2	40	20	1	2026-01-22 05:39:49.840793+00	199	2025-12-17 13:34:56.415129+00	2025-12-17 13:34:56.415129+00
36	62	20	1	2026-01-31 08:30:38.488355+00	2	2026-01-31 08:30:26.107458+00	\N
37	63	113	13	2026-02-01 17:41:52.131549+00	2	2026-02-01 17:40:28.91301+00	\N
38	65	20	1	2026-02-04 15:32:31.130291+00	7	2026-02-03 18:50:53.274141+00	\N
40	64	20	1	2026-02-04 15:40:01.04036+00	1	2026-02-04 15:40:01.04036+00	\N
35	61	20	1	2026-02-04 16:08:13.65273+00	3	2026-01-31 08:29:05.13332+00	\N
\.


--
-- Data for Name: product_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product_items (id, sub_category_name, dynamic_fields, created_at, updated_at, product_item_images, admin_id, sub_category_id, category_id, department_id, shop_id, stock) FROM stdin;
46	Anchor Screws	{"Size": "M8", "Finish": "Plain", "Length": "50mm", "Material": "Zinc Alloy", "Quantity": "10", "Head Type": "Flat Head", "Thread Type": "Expansion Type"}	2025-12-29 03:58:03.149884+00	2025-12-29 03:58:03.149884+00	{uploads/products/afe2bff8-7708-4d1d-8b48-09ba93c61b19.jpg}	1	5	1	1	\N	t
47	Eye Bolts	{"Size": "XS", "Finish": "Zinc Plated", "Length": "Medium", "Material": "Polyester", "Quantity": "10", "Head Type": "Round Head", "Thread Type": "Coarse Thread"}	2025-12-31 12:32:34.554002+00	2025-12-31 12:32:34.554002+00	{uploads/products/deb63cf0-0994-4ebf-922f-ae4c6b462013.jpg,uploads/products/4ece0941-f055-4ba8-8aff-f5dbd9c2213e.jpg,uploads/products/38ecf67c-a8cc-4151-9065-f19545290eea.jpg}	1	11	1	1	\N	t
48	Finishing Nails	{"Size": "S", "Finish": "Zinc Plated", "Length": "Medium", "Material": "Polyester", "Quantity": "10", "Head Type": "Flat Head", "Thread Type": "Coarse Thread"}	2026-01-01 12:24:24.399786+00	2026-01-01 12:24:24.399786+00	{uploads/products/78308415-06a3-46f4-a0b7-34cff0b89e4b.jpg}	1	7	1	1	\N	t
49	Anchor Screws	{"Size": "M10", "Finish": "Plain", "Length": "60mm", "Material": "Stainless Steel", "Quantity": "1", "Head Type": "Flat Head", "Thread Type": "Coarse Thread"}	2026-01-01 12:25:17.310126+00	2026-01-01 12:25:17.310126+00	{uploads/products/4119d550-150d-4376-91bd-9d88b4386707.jpg}	1	5	1	1	\N	t
50	Hex Bolts	{"Size": ["XS", "M"], "Finish": ["Stainless Steel", "Zinc Plated", "Black Oxide"], "Length": ["Short", "Medium"], "Material": ["Silk", "Polyester"], "Quantity": "11", "Head Type": ["Round Head", "Flat Head"], "Thread Type": ["Fine Thread", "Coarse Thread"]}	2026-01-20 04:20:23.375025+00	2026-01-20 04:20:23.375025+00	{"uploads\\\\products\\\\7f609ad3-2869-4d50-801d-0d1da93e5413.jpg","uploads\\\\products\\\\386fc659-5947-4e61-a554-b31f9753ef0c.jpg"}	1	9	1	1	\N	t
51	PVC Pipes 1/2"	{"Length": "32", "Diameter": ["6mm", "8mm", "10mm"], "Material": ["Cotton", "Polyester"], "Quantity": "20", "Pressure Rating": ["Medium Pressure", "Low Pressure"]}	2026-01-20 04:31:36.067214+00	2026-01-20 04:31:36.067214+00	{"uploads\\\\products\\\\402990aa-1043-45b8-9ced-86eb8ba69f27.jpg"}	1	16	2	1	\N	t
52	Garden Forks	{"Material": ["Silk", "Polyester", "Cotton"], "Quantity": "10", "Tine Count": ["Small", "Large"], "Handle Length": ["Short (10-20cm)", "Medium (20-40cm)"]}	2026-01-20 04:36:24.825259+00	2026-01-20 04:36:24.825259+00	{"uploads\\\\products\\\\4b0f6c5a-cbba-47e5-a317-1826e4119362.jpg"}	1	97	12	1	\N	t
37	Dining Chairs	{"Armrest": "No", "Material": "Polyester", "Quantity": "10", "Stackable": "Yes", "Upholstery": "Leather"}	2025-12-17 13:49:19.252913+00	2025-12-17 13:49:19.252913+00	{uploads/products/527e619f-1b16-4513-93b1-d197a14e0b28.jpg,uploads/products/29cb5ab6-4cab-4660-8950-0bebdf417a16.jpg}	1	210	37	3	102	t
38	Dining Chairs	{"Armrest": "No", "Material": "Wool", "Quantity": "10", "Stackable": "No", "Upholstery": "Leatherette"}	2025-12-17 13:51:52.717785+00	2025-12-17 13:51:52.717785+00	{uploads/products/792b39dd-713b-443b-8048-1f4d594dc35d.jpg}	1	210	37	3	103	t
39	6 Seater Dining Sets	{"Finish": "Stainless Steel", "Quantity": "10", "Table Size": "Large", "Chair Material": "Large", "Table Material": "Medium"}	2025-12-17 14:31:00.345377+00	2025-12-17 14:31:00.345377+00	{uploads/products/ae55da25-db01-49fc-b024-a6182e45bd7c.jpg,uploads/products/3fbb98e4-7f5c-41cf-8e70-5f480f315519.jpg,uploads/products/75a16a03-f40b-48c2-a009-969454b5e320.jpg,uploads/products/53ff379b-e7ac-415b-bca7-411c84f0cf9d.jpg,uploads/products/22ae746f-3189-44df-a2d4-4bb77a45941f.jpg,uploads/products/713ababc-dbb5-434b-8edc-3736d3cb264e.jpg}	1	207	37	3	104	t
40	Dining Chairs	{"Armrest": "Yes", "Material": "Silk", "Quantity": "10", "Stackable": "Yes", "Upholstery": "Leather"}	2025-12-18 12:28:52.071826+00	2025-12-18 12:28:52.071826+00	{uploads/products/1a3dea66-8d4d-40ed-a14c-5167b5d62475.webp,uploads/products/96afe1e9-8202-4d18-b373-96302c1f2cdf.webp,uploads/products/8adf4c62-b2c6-4cc8-ae44-86132ba37508.webp,uploads/products/a5dc9c91-90fd-4d2b-bea3-ba97e5dfcaae.webp}	1	210	37	3	105	t
45	Self Tapping Screws	{"Size": "#6", "Finish": "Black Oxide", "Length": "12mm", "Material": "Carbon Steel", "Quantity": "11", "Head Type": "Flat Head", "Thread Type": "Type B"}	2025-12-28 20:00:50.445158+00	2025-12-28 20:00:50.445158+00	{uploads/products/590d2b15-38d5-4dac-ae9b-8e142c747344.jpg}	1	2	1	1	110	t
70	Araldite	null	2026-02-06 16:16:58.068657+00	2026-02-06 16:16:58.068657+00	{"uploads\\\\products\\\\6e12aed7-6fb4-4249-86c5-8dc3e05a4fb6.jpg"}	1	106	14	1	20	t
69	Araldite	{"Type": ["Premium", "Deluxe"], "Quantity": "2", "Cure Time": ["Medium", "Large"], "Mix Ratio": ["Medium", "Large"]}	2026-02-06 16:16:27.638047+00	2026-02-07 09:52:22.478068+00	{"uploads\\\\products\\\\5d22a384-03fc-4a23-94e7-1efef0aa3444.jpg"}	1	106	14	1	20	t
72	Silicone Sealants	{"Type": ["Premium", "Deluxe"], "Color": ["Yellow", "Red", "Purple", "Black"], "Quantity": "3", "Cure Time": ["Small", "Large", "Medium"], "Cartridge Size": ["Large", "Medium"]}	2026-02-06 16:21:39.980289+00	2026-02-07 22:01:14.23475+00	{"\\"uploads\\\\\\\\products\\\\\\\\0c3e6dce-0f61-4c3b-be23-94bc41ac8b13.jpg\\""}	1	108	14	1	20	t
71	Araldite	{"Type": ["Premium"], "Quantity": "2", "Cure Time": ["Small", "Medium", "Large"], "Mix Ratio": ["Small", "Medium", "Large"]}	2026-02-06 16:17:54.924301+00	2026-02-13 15:18:54.990511+00	{"uploads\\\\products\\\\329ff414-6367-452c-b353-57371c70d2d2.jpg",uploads/products/919fc99b-422c-477a-98ac-51cede81c8b6.jpg}	1	106	14	1	20	t
61	Sideboards	{"Finish": ["Zinc Plated", "Stainless Steel", "Black Oxide"], "Material": ["Cotton", "Polyester", "Silk"], "Quantity": "3", "Dimensions": "30*30", "Doors/Drawers": "1"}	2026-01-31 08:26:44.440887+00	2026-01-31 08:26:44.440887+00	{"uploads\\\\products\\\\e75ed470-bdd5-457c-9324-8f113db2d86e.jpg","uploads\\\\products\\\\d07dcb55-8952-42e4-8e70-7e30d5674dea.jpg"}	1	211	37	3	20	t
62	Hammers	{"Type": ["Standard"], "Length": ["Long", "Medium"], "Quantity": "5", "Head Weight": ["Medium", "Small"], "Handle Material": ["Wood"]}	2026-01-31 08:30:22.873894+00	2026-01-31 08:30:22.873894+00	{"uploads\\\\products\\\\637b672f-9e8f-4b0e-b84d-83819dba1616.jpg"}	1	64	6	1	20	t
63	Drills	{"Brand": "Moter", "Weight": "23", "Quantity": "22", "Chuck Size": ["Medium", "Large"], "Power Type": ["Battery", "DC", "AC"], "Power Rating": "23"}	2026-02-01 17:17:06.041083+00	2026-02-01 17:17:06.041083+00	{"uploads\\\\products\\\\a39a1668-0c00-4acf-93b7-43ac327ed48c.jpg"}	13	56	5	1	113	t
64	Concrete Nails	{"Size": ["S", "M"], "Finish": ["Stainless Steel", "Zinc Plated"], "Length": ["Short", "Medium", "Long"], "Material": ["Cotton", "Polyester", "Silk"], "Quantity": "10", "Head Type": ["Round Head", "Flat Head"], "Thread Type": ["Fine Thread", "Coarse Thread"]}	2026-02-02 18:47:36.856471+00	2026-02-02 18:47:36.856471+00	{"uploads\\\\products\\\\9d21d92b-ef2d-4ea5-80da-3218b819c3f2.jpg"}	1	8	1	1	20	t
65	Casual Shirts	{"Size": ["XS", "S", "M"], "Fabric": ["Polyester", "Cotton"], "Pattern": ["Checkered", "Striped", "Solid"], "Quantity": "2", "Sleeve Length": ["Half Sleeve", "Full Sleeve", "Three Quarter Sleeve"]}	2026-02-03 18:50:49.554435+00	2026-02-03 18:50:49.554435+00	{"uploads\\\\products\\\\8f16de56-c6e1-4cac-bd47-599098ab35c6.jpg"}	1	468	105	7	20	t
66	Formal Shirts	{"Color": ["Yellow", "Green", "Black"], "Fit Type": ["Slim Fit", "Regular Fit"], "Quantity": "3", "Sleeve Length": ["Half Sleeve", "Full Sleeve"]}	2026-02-03 19:00:07.830358+00	2026-02-03 19:00:07.830358+00	{"uploads\\\\products\\\\5da4b213-9604-4756-a5b1-2510d5ddc428.jpg"}	1	467	105	7	20	t
67	Araldite	{"Type": ["Premium", "Standard"], "Quantity": "2", "Cure Time": ["Large", "Medium", "Small"], "Mix Ratio": ["Large", "Small", "Medium"]}	2026-02-04 18:13:52.394372+00	2026-02-04 18:13:52.394372+00	{"uploads\\\\products\\\\9bb9f81b-33da-4f13-8e85-bcf498468b7c.jpg","uploads\\\\products\\\\1424b8af-a318-49e7-977b-df7a9abae617.jpg"}	1	106	14	1	20	t
75	Bricks	{"Size": ["M", "XS", "L"], "Type": ["Premium"], "Quantity": "2"}	2026-02-06 19:34:51.536931+00	2026-02-07 05:32:08.903442+00	{"uploads\\\\products\\\\26a6ea77-7948-4b82-bc86-725d18fd6de9.jpg",uploads/products/473df627-cb6c-4315-8d29-fd3b24067955.jpg}	1	111	16	2	20	t
81	Safety Gloves	{"Size": ["M", "XS", "S", "L"], "Type": ["Standard", "Premium"], "Material": ["Polyester", "Cotton", "Silk"], "Quantity": "6", "Cut Resistance": ["Medium", "Small", "Large", "Extra Large"]}	2026-02-07 10:12:18.888685+00	2026-02-07 10:12:46.417938+00	{uploads/products/fa6b148d-f334-45a7-8b30-23f56af43f66.jpg}	1	92	11	1	20	t
79	TMT Bars 16mm	{"Grade": ["Grade B", "Grade C"], "Length": ["Long", "Short", "Medium"], "Diameter": ["8mm", "5mm", "6mm"], "Quantity": "5"}	2026-02-06 19:44:28.423968+00	2026-02-15 04:52:53.601687+00	{"uploads\\\\products\\\\7766b22b-ed2e-4a25-9e1e-50ab1e0d4da8.jpg"}	1	117	17	2	20	t
76	Welding Rods	{"Length": ["Medium", "Long"], "Diameter": ["8mm", "6mm"], "Quantity": "3", "AWS Grade": "Nice"}	2026-02-06 19:38:34.437417+00	2026-02-06 19:38:34.437417+00	{"uploads\\\\products\\\\99a1741c-a3f0-4d76-ba8e-c5f5a94849fe.jpg"}	1	102	13	1	1	t
77	Epoxy Adhesives	{"Type": ["Premium", "Deluxe"], "Quantity": "2", "Strength": ["Medium", "Large"], "Mix Ratio": ["Medium", "Large"], "Viscosity": ["Large", "Medium"]}	2026-02-06 19:39:59.733058+00	2026-02-06 19:39:59.733058+00	{"uploads\\\\products\\\\607bd44b-4910-46fe-a723-7745a2c0650c.jpg"}	1	107	14	1	1	t
74	PVC Pipes 1"	{"Length": ["11"], "Diameter": ["8mm", "5mm", "10mm"], "Material": ["Polyester", "Silk"], "Quantity": "3", "Pressure Rating": ["Medium Pressure"]}	2026-02-06 18:27:16.755633+00	2026-02-15 04:53:30.81436+00	{uploads/products/6c81cdda-9059-42bb-8eac-1fa8b7ac9c85.jpg}	1	17	2	1	20	t
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.products (id, name, description, category_id, image, created_at, updated_at, stock, department_id, shop_id, price, discount_price, brand_id) FROM stdin;
37	Personal Care	Medical	77	uploads/products/d97469e7-4873-4056-a1c7-a00f94f4028f.jpg	2025-12-12 08:42:13.76698+00	\N	\N	5	1	0	0	\N
38	Cabinets & Storage	Furniture Pen Holder	44	uploads/products/8e9d4e1d-cef6-4218-8112-1e6e6249ab02.jpg	2025-12-12 09:00:56.938698+00	\N	\N	3	1	0	0	\N
39	Dresses	Eyes wear	109	uploads/products/514e00c4-512f-4bf0-bda5-c4d157bc6640.jpg	2025-12-12 18:44:33.212951+00	\N	\N	7	1	0	0	\N
40	Bedroom	Bed	33	uploads/products/6c7a3807-e3e5-455b-86f4-ba7c780fb5f3.jpg	2025-12-12 20:09:10.613184+00	\N	\N	3	1	0	0	\N
41	Laptops	Laptop	50	uploads/products/08f7ff0e-b33e-4b5b-bcb1-0e7058fd7705.jpg	2025-12-12 20:09:49.353185+00	\N	\N	4	1	0	0	\N
42	Desktops	Tv	51	uploads/products/fb2fb257-05dd-4f10-86cb-08ae297f6024.jpg	2025-12-12 20:10:26.384112+00	\N	\N	4	1	0	0	\N
43	Home Appliances	Cooler	53	uploads/products/e94ab71d-d9ff-4239-8472-ad0df8dc1dec.jpg	2025-12-12 20:11:47.503113+00	\N	\N	4	1	0	0	\N
44	Mobiles	Mobile	48	uploads/products/2139d5a5-bf36-484d-a4c8-20a286c6a60c.jpg	2025-12-12 20:12:45.347484+00	\N	\N	4	1	0	0	\N
45	Accessories	Watches	58	uploads/products/d1fb841f-24b2-4ce3-8c77-b82b7a8011c3.jpg	2025-12-13 08:07:34.810224+00	\N	\N	4	1	0	0	\N
46	Ethnic Wear	wzc	120	uploads/products/6843e0a3-3683-45f5-8548-3e71044045e7.jpg	2025-12-15 13:00:44.460725+00	\N	\N	7	20	0	0	\N
\.


--
-- Data for Name: promotion_categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.promotion_categories (id, name, shop_id, is_active, created_at, updated_at, icon_path) FROM stdin;
3	Loyalty	\N	t	2026-01-07 05:41:57.155072+00	\N	uploads/icon/Loyalty.png
2	Value-Add	\N	t	2026-01-07 05:41:22.695655+00	\N	uploads/icon/ValueAdd.png
1	Price-Based	\N	t	2026-01-07 05:40:31.268827+00	\N	uploads/icon/Price_Based.png
\.


--
-- Data for Name: promotions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.promotions (id, promotion_category_id, promotion_type_id, offer_name, description, discount_rate, start_date, end_date, minimum_purchase_amount, tier_quantity, bogo_get_quantity, bogo_buy_quantity, bogo_combination_enabled, gift_description, shop_id, is_active, created_at, updated_at) FROM stdin;
5	1	2	5% Off	Small spend incentive	5	2026-01-28	2026-07-28	500	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
6	1	2	10% Off	Moderate spend	10	2026-01-28	2026-07-28	1000	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
7	1	2	15% Off	High spend	15	2026-01-28	2026-07-28	1500	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
8	1	5	₹50 Off	Small fixed discount	50	2026-01-28	2026-07-28	400	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
9	1	5	₹100 Off	Moderate fixed discount	100	2026-01-28	2026-07-28	800	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
10	1	6	₹150 Off Clearance	Clearance sale	150	2026-01-28	2026-07-28	1200	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
11	1	7	Buy 1 Save 5%	Single item	5	2026-01-28	2026-07-28	\N	1	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
12	1	7	Buy 2 Save 10%	Pair purchase	10	2026-01-28	2026-07-28	\N	2	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
13	1	7	Buy 3 Save 15%	Triple purchase	15	2026-01-28	2026-07-28	\N	3	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
14	1	7	Buy 5 Save 25%	Bulk purchase	25	2026-01-28	2026-07-28	\N	5	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
15	2	9	Buy 1 Get 1 Free	Starter BOGO	0	2026-01-28	2026-07-28	\N	1	1	1	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
16	2	9	Buy 1 Get 2 Free	Starter BOGO+	0	2026-01-28	2026-07-28	\N	1	2	1	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
17	2	9	Buy 2 Get 1 Free	Classic BOGO	0	2026-01-28	2026-07-28	\N	2	1	2	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
18	2	9	Buy 2 Get 2 Free	Double BOGO	0	2026-01-28	2026-07-28	\N	2	2	2	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
19	2	9	Buy 3 Get 1 Free	Medium BOGO	0	2026-01-28	2026-07-28	\N	3	1	3	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
20	2	9	Buy 5 Get 2 Free	Bulk BOGO	0	2026-01-28	2026-07-28	\N	5	2	5	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
21	2	8	Free Shipping	Orders above ₹500	0	2026-01-28	2026-07-28	500	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
22	2	8	Free Shipping Premium	Orders above ₹1000	0	2026-01-28	2026-07-28	1000	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
23	3	13	Welcome Offer	New user discount	15	2026-01-28	2026-07-28	\N	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
24	3	14	Loyalty Reward	Returning customer	10	2026-01-28	2026-07-28	\N	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
25	3	15	Referral Bonus	Refer a friend	10	2026-01-28	2026-07-28	\N	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
26	3	16	Birthday Special	Birthday discount	15	2026-01-28	2026-07-28	\N	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
27	3	17	Cart Recovery	Abandoned cart offer	10	2026-01-28	2026-07-28	\N	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
28	3	18	Default Offer	General promotion	20	2026-01-28	2026-07-28	\N	\N	\N	\N	f	\N	\N	t	2026-01-28 08:27:02.99833+00	2026-01-28 08:27:02.99833+00
\.


--
-- Data for Name: promotions_types; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.promotions_types (id, name, is_active, shop_id, promotion_category_id, promotion_offer_details, created_at, updated_at, icon_path, type) FROM stdin;
13	Welcome	t	\N	3	{"showGift": true, "valueHint": "e.g., 15", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/loyalty/welcome.jpg	welcome
9	Buy & Get	t	\N	2	{"showGift": false, "valueHint": "e.g., 1", "valueLabel": "Buy Quantity", "showQuantity": true, "showMinPurchase": false}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/value_add/Bogo.png	bogo
10	Bulk Sale	t	\N	2	{"showGift": false, "valueHint": "e.g., 25", "valueLabel": "Bundle Discount (%)", "showQuantity": true, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/value_add/Bundle.png	bundle
2	Percentage	t	\N	1	{"showGift": false, "valueHint": "e.g., 20", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/price_based/Percentage.png	percentage
11	Free Gift	t	\N	2	{"showGift": false, "valueHint": "e.g., 25", "valueLabel": "Bundle Discount (%)", "showQuantity": true, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/value_add/Free_Gift.png	free_gift
14	Loyalty	t	\N	3	{"showGift": false, "valueHint": "e.g., 15", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": false}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/loyalty/loyalty.jpg	loyalty
15	Referral	t	\N	3	{"showGift": false, "valueHint": "e.g., 15", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": false}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/loyalty/referal.jpg	referral
16	Birthday	t	\N	3	{"showGift": false, "valueHint": "e.g., 15", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": false}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/loyalty/birthday.jpg	birthday
3	Flash Sale	t	\N	1	{"showGift": false, "valueHint": "e.g., 20", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/price_based/Flash_Sales.png	flash_sale
17	Cart Recovery	t	\N	3	{"showGift": false, "valueHint": "e.g., 10", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/loyalty/cart_recovery.jpg	cart_recovery
18	Default	t	\N	3	{"showGift": false, "valueHint": "20", "valueLabel": "Value", "showQuantity": false, "showMinPurchase": false}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/loyalty/default.jpg	default
12	Cashback	t	\N	2	{"showGift": true, "valueHint": "e.g., 15", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/value_add/Cashback.png	cashback
4	Seasonal	t	\N	1	{"showGift": false, "valueHint": "e.g., 20", "valueLabel": "Discount (%)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/price_based/Seasonal.png	seasonal
5	Fixed	t	\N	1	{"showGift": false, "valueHint": "e.g., 100", "valueLabel": "Discount Amount (₹)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/price_based/Fixed.png	fixed
6	Clearance	t	\N	1	{"showGift": false, "valueHint": "e.g., 100", "valueLabel": "Discount Amount (₹)", "showQuantity": false, "showMinPurchase": true}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/price_based/Clearance.png	clearance
7	Volume Sale	t	\N	1	{"showGift": false, "valueHint": "e.g., 15", "valueLabel": "Discount (%)", "showQuantity": true, "showMinPurchase": false}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/price_based/Volume.png	volume
8	Free Shipping	t	\N	2	{"showGift": false, "valueHint": "e.g., 500", "valueLabel": "Min Purchase (₹)", "showQuantity": false, "showMinPurchase": false}	2026-01-07 10:06:58.954119+00	\N	uploads/promotions/value_add/Free_Shipping.png	free_shipping
\.


--
-- Data for Name: service_providers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.service_providers (id, name, business_name, shop_id, admin_id, profile_photo, bio, phone, whatsapp, email, base_address, serviceable_pincodes, service_radius_km, categories, sub_services, experience_years, tools_brought, pricing_model, base_charge, min_job_charge, rate_card, working_days, time_slots, advance_notice_hours, kyc_status, license_number, insurance, police_verification, cancellation_hours, warranty_days, rating, total_jobs, response_time_min, portfolio_images, account_status, payout_upi, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: shop_details; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shop_details (id, admin_id, shop_name, owner_name, email, phone, address_line1, address_line2, city, state, country, pincode, latitude, longitude, shop_description, shop_verification_docs, document_type, document_value, pan_number, itr_documents, shop_status, bank_account_number, bank_ifsc, shop_image_url, shop_verification_status, shop_verification_remarks, photo_shop_verification, business_doc_verification, identity_doc_verification, address_proof_verification, created_at, updated_at, has_offers, shop_type) FROM stdin;
102	2	Daily Needs Mart	Rohit	shop102@gmail.com	9001000002	Hoodi Circle	Near Bus Stop	Bengaluru	Karnataka	India	560048	12.9718890	77.7012450	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_2.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
103	3	Green Basket	Rohit	shop103@gmail.com	9001000003	AECS Layout	Near Park	Bengaluru	Karnataka	India	560037	12.9732140	77.7041180	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_3.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
104	4	Neighborhood Store	Rohit	shop104@gmail.com	9001000004	Brookefield Road	Near Axis Bank	Bengaluru	Karnataka	India	560037	12.9764210	77.7035020	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_4.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
105	5	Fresh Mart	Rohit	shop105@gmail.com	9001000005	Varthur Road	Near Signal	Bengaluru	Karnataka	India	560066	12.9725410	77.7069810	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_5.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
106	6	Smart Kirana	Rohit	shop106@gmail.com	9001000006	Doddanekundi	Near Lake	Bengaluru	Karnataka	India	560037	12.9708890	77.6987420	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_6.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
107	7	Urban Grocery	Rohit	shop107@gmail.com	9001000007	Marathahalli	Near Bridge	Bengaluru	Karnataka	India	560037	12.9697110	77.7023140	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_7.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
108	8	Family Store	Rohit	shop108@gmail.com	9001000008	Kundalahalli	Near BDA Complex	Bengaluru	Karnataka	India	560037	12.9771120	77.7001180	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_8.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
109	9	Budget Mart	Rohit	shop109@gmail.com	9001000009	HAL Main Road	Near Petrol Pump	Bengaluru	Karnataka	India	560008	12.9721180	77.6978120	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_9.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
110	10	Quick Buy	Rohit	shop110@gmail.com	9001000010	Graphite India Road	Near Tech Park	Bengaluru	Karnataka	India	560048	12.9768890	77.7054120	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_10.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
111	11	Local Mart	Rohit	shop111@gmail.com	9001000011	EPIP Zone	Near Gate	Bengaluru	Karnataka	India	560066	12.9748120	77.7071150	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_11.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
112	12	City Kirana	Rohit	shop112@gmail.com	9001000012	BEML Layout	Near School	Bengaluru	Karnataka	India	560066	12.9780150	77.6993180	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_12.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
113	13	Everyday Needs	Rohit	shop113@gmail.com	9001000013	Ramagondanahalli	Near Temple	Bengaluru	Karnataka	India	560066	12.9712140	77.7049980	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_13.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
114	14	Prime Store	Rohit	shop114@gmail.com	9001000014	Hope Farm	Near Metro	Bengaluru	Karnataka	India	560066	12.9765410	77.6984110	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_14.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
115	15	Mega Mart	Rohit	shop115@gmail.com	9001000015	Kadugodi	Near Railway Station	Bengaluru	Karnataka	India	560067	12.9871120	77.7078120	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_15.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
116	16	Corner Shop	Rohit	shop116@gmail.com	9001000016	Seegehalli	Near Bus Stand	Bengaluru	Karnataka	India	560049	12.9821180	77.7123140	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_16.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
117	17	Express Kirana	Rohit	shop117@gmail.com	9001000017	Channasandra	Near Lake	Bengaluru	Karnataka	India	560067	12.9857410	77.7015120	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_17.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
118	18	Smart Basket	Rohit	shop118@gmail.com	9001000018	Varthur Main Rd	Near Signal	Bengaluru	Karnataka	India	560087	12.9698120	77.7089140	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_18.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
119	19	Town Store	Rohit	shop119@gmail.com	9001000019	Munnekollal	Near School	Bengaluru	Karnataka	India	560037	12.9687410	77.6991140	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_19.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
120	20	Nearby Mart	Rohit	shop120@gmail.com	9001000020	KR Puram	Near Flyover	Bengaluru	Karnataka	India	560036	12.9711120	77.6958120	\N	manual	manual	\N	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_20.jpg	t	false	t	t	t	t	2026-01-14 19:08:56.957431+00	2026-01-14 19:08:56.957431+00	\N	\N
20	1	Ajay Mobile shop	Rohit	rohit.jangid.social@gmail.com	9886569962	The Greens apartment	Doddanekundi	Bangalore	Karnataka	India	560037	12.9742340	77.7014990	\N	\N	manual	manual	\N	\N	\N	\N	\N	uploads/admin-profiles/admin_1_1769421966.jpg	t	false	t	t	t	t	2025-12-14 19:29:54.535661+00	2026-02-10 05:03:53.337768+00	\N	3
\.


--
-- Data for Name: shop_offers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shop_offers (id, shop_id, offer_id, admin_id, start_date, end_date, created_at, updated_at) FROM stdin;
1	20	2	1	2026-01-24 00:00:00+00	2026-01-27 23:59:59+00	2025-12-24 19:17:09.82279+00	2025-12-24 19:17:09.82279+00
2	113	0	13	0001-12-31 23:58:45-00:01:15 BC	0001-12-31 23:58:45-00:01:15 BC	2026-02-01 18:25:38.571257+00	2026-02-01 18:25:38.571257+00
3	113	0	13	0001-12-31 23:58:45-00:01:15 BC	0001-12-31 23:58:45-00:01:15 BC	2026-02-01 18:35:07.799918+00	2026-02-01 18:35:07.799918+00
\.


--
-- Data for Name: shop_orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shop_orders (id, user_id, order_date, address_id, order_total_price, discount, order_status_id, payment_method_id, shop_id) FROM stdin;
\.


--
-- Data for Name: shop_times; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shop_times (id, shop_id, status, open_time, close_time, created_at, updated_at) FROM stdin;
1	20	open	09:15	23:30	0001-12-31 23:58:45-00:01:15 BC	2026-02-13 16:13:55.062811+00
\.


--
-- Data for Name: shop_verification_histories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shop_verification_histories (id, admin_id, shop_id, verification_status, remarks, changed_at) FROM stdin;
\.


--
-- Data for Name: shop_verifications; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.shop_verifications (id, admin_id, shop_id, shop_name, verification_status, remarks, agent_id, created_at, updated_at) FROM stdin;
1	1	\N	\N	t	Shop registration under review	\N	2025-12-01 18:14:22.326484+00	2025-12-04 08:25:41.750184+00
10	2	\N	\N	f	Shop registration under review	\N	2025-12-16 15:16:09.79106+00	2025-12-16 15:16:09.79106+00
11	3	\N	\N	f	Shop registration under review	\N	2025-12-18 12:46:48.944094+00	2025-12-18 12:46:48.944096+00
12	4	\N	\N	f	Shop registration under review	\N	2025-12-18 12:50:03.981641+00	2025-12-18 12:50:03.981641+00
13	5	\N	\N	f	Shop registration under review	\N	2025-12-18 12:50:08.954761+00	2025-12-18 12:50:08.954761+00
14	6	\N	\N	f	Shop registration under review	\N	2026-01-05 13:11:17.195526+00	2026-01-05 13:11:17.195526+00
15	7	\N	\N	f	Shop registration under review	\N	2026-01-06 05:06:16.431395+00	2026-01-06 05:06:16.431395+00
16	8	\N	\N	f	Shop registration under review	\N	2026-01-23 19:09:12.914149+00	2026-01-23 19:09:12.914149+00
17	9	\N	\N	f	Shop registration under review	\N	2026-01-23 19:09:46.263607+00	2026-01-23 19:09:46.263607+00
18	10	\N	\N	f	Shop registration under review	\N	2026-01-25 19:51:33.704816+00	2026-01-25 19:51:33.704816+00
19	11	\N	\N	f	Shop registration under review	\N	2026-01-25 19:52:47.965066+00	2026-01-25 19:52:47.965066+00
20	12	\N	\N	f	Shop registration under review	\N	2026-01-26 10:07:57.74411+00	2026-01-26 10:07:57.74411+00
21	13	\N	\N	f	Shop registration under review	\N	2026-02-01 09:48:09.054097+00	2026-02-01 09:48:09.054097+00
22	14	\N	\N	f	Shop registration under review	\N	2026-02-01 11:15:32.381869+00	2026-02-01 11:15:32.381869+00
23	15	\N	\N	f	Shop registration under review	\N	2026-02-01 11:24:16.689206+00	2026-02-01 11:24:16.689206+00
24	16	\N	\N	f	Shop registration under review	\N	2026-02-01 11:25:03.783163+00	2026-02-01 11:25:03.783163+00
25	17	\N	\N	f	Shop registration under review	\N	2026-02-01 11:26:33.86267+00	2026-02-01 11:26:33.86267+00
26	18	\N	\N	f	Shop registration under review	\N	2026-02-01 11:43:12.724306+00	2026-02-01 11:43:12.724306+00
27	19	\N	\N	f	Shop registration under review	\N	2026-02-01 11:45:31.662822+00	2026-02-01 11:45:31.662822+00
28	20	\N	\N	f	Shop registration under review	\N	2026-02-01 11:54:50.561432+00	2026-02-01 11:54:50.561432+00
29	21	\N	\N	f	Shop registration under review	\N	2026-02-01 11:55:48.805581+00	2026-02-01 11:55:48.805581+00
30	22	\N	\N	f	Shop registration under review	\N	2026-02-01 12:00:38.012066+00	2026-02-01 12:00:38.012066+00
31	23	\N	\N	f	Shop registration under review	\N	2026-02-01 12:04:22.762082+00	2026-02-01 12:04:22.762082+00
32	24	\N	\N	f	Shop registration under review	\N	2026-02-01 12:24:15.181316+00	2026-02-01 12:24:15.181316+00
33	25	\N	\N	f	Shop registration under review	\N	2026-02-01 12:53:48.286786+00	2026-02-01 12:53:48.286786+00
34	26	\N	\N	f	Shop registration under review	\N	2026-02-01 13:02:44.578572+00	2026-02-01 13:02:44.578572+00
35	27	\N	\N	f	Shop registration under review	\N	2026-02-01 13:05:19.632758+00	2026-02-01 13:05:19.632758+00
36	28	\N	\N	f	Shop registration under review	\N	2026-02-01 13:05:46.977921+00	2026-02-01 13:05:46.977921+00
37	29	\N	\N	f	Shop registration under review	\N	2026-02-01 13:07:16.982151+00	2026-02-01 13:07:16.982151+00
38	30	\N	\N	f	Shop registration under review	\N	2026-02-01 13:13:17.578779+00	2026-02-01 13:13:17.578779+00
39	31	\N	\N	f	Shop registration under review	\N	2026-02-01 13:17:27.806061+00	2026-02-01 13:17:27.806061+00
40	32	\N	\N	f	Shop registration under review	\N	2026-02-01 13:21:45.941401+00	2026-02-01 13:21:45.941401+00
41	33	\N	\N	f	Shop registration under review	\N	2026-02-01 13:23:27.169195+00	2026-02-01 13:23:27.169195+00
42	34	\N	\N	f	Shop registration under review	\N	2026-02-16 18:01:40.626052+00	2026-02-16 18:01:40.626052+00
43	35	\N	\N	f	Shop registration under review	\N	2026-02-16 18:02:40.424304+00	2026-02-16 18:02:40.424304+00
\.


--
-- Data for Name: sub_categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sub_categories (id, department_id, category_id, name, sort_order, is_active, image_url) FROM stdin;
36	1	4	Asian Paints	1	t	\N
37	1	4	Nerolac Paints	2	t	\N
58	1	5	Circular Saws	3	t	\N
81	1	8	Scaffolding	5	t	\N
83	1	9	Spirit Levels	2	t	\N
84	1	9	Vernier Calipers	3	t	\N
85	1	9	Laser Distance Meters	4	t	\N
16	1	2	PVC Pipes 1/2"	1	t	uploads/sub-category-images/hardware/plumbing/pvc_pipes_half.png
17	1	2	PVC Pipes 1"	2	t	uploads/sub-category-images/hardware/plumbing/pvc_pipes.png
18	1	2	CPVC Pipes	3	t	uploads/sub-category-images/hardware/plumbing/cpvc_pipes.png
19	1	2	GI Pipes	4	t	uploads/sub-category-images/hardware/plumbing/gi_pipes.png
20	1	2	Elbow Fittings	5	t	uploads/sub-category-images/hardware/plumbing/elbow_fittings.png
21	1	2	Tee Fittings	6	t	uploads/sub-category-images/hardware/plumbing/tee_fittings.png
22	1	2	Gate Valves	7	t	uploads/sub-category-images/hardware/plumbing/gate_valves.png
23	1	2	Ball Valves	8	t	uploads/sub-category-images/hardware/plumbing/ball_valves.png
24	1	2	Bathroom Fittings	9	t	uploads/sub-category-images/hardware/plumbing/bathroom_fittings.png
25	1	2	Wash Basins	10	t	uploads/sub-category-images/hardware/plumbing/wash_basins.png
26	1	3	LED Bulbs 5W	1	t	uploads/sub-category-images/hardware/electricle/bulbs_5w.png
27	1	3	LED Bulbs 9W	2	t	uploads/sub-category-images/hardware/electricle/led_bulbs_9w.png
28	1	3	LED Tubes 20W	3	t	uploads/sub-category-images/hardware/electricle/led_tubes.png
29	1	3	Copper Wires	4	t	uploads/sub-category-images/hardware/electricle/copper_wires.png
35	1	3	MCB Boxes	10	t	uploads/sub-category-images/hardware/electricle/mcb_boxes.png
31	1	3	Modular Switches	6	t	uploads/sub-category-images/hardware/electricle/modular_switches.png
30	1	3	MCB 16A	5	t	uploads/sub-category-images/hardware/electricle/MCB_16A.png
34	1	3	Exhaust Fans	9	t	uploads/sub-category-images/hardware/electricle/exhaust_fans.png
33	1	3	Ceiling Fans	8	t	uploads/sub-category-images/hardware/electricle/celling_fans.png
38	1	4	Interior Wall Paints	3	t	uploads/sub-category-images/hardware/paint/interior_wall_paint.png
40	1	4	Enamel Paints	5	t	uploads/sub-category-images/hardware/paint/enamel_paints.png
42	1	4	Wood Primer	7	t	uploads/sub-category-images/hardware/paint/wood_primer.png
43	1	4	Putty	8	t	uploads/sub-category-images/hardware/paint/putty.png
45	1	4	Paint Rollers	10	t	uploads/sub-category-images/hardware/paint/rollers.png
56	1	5	Drills	1	t	uploads/sub-category-images/hardware/power_tools/drills.png
57	1	5	Angle Grinders	2	t	uploads/sub-category-images/hardware/power_tools/angle_grinder.png
59	1	5	Jigsaws	4	t	uploads/sub-category-images/hardware/power_tools/jigsaws.png
60	1	5	Sanders	5	t	uploads/sub-category-images/hardware/power_tools/sanders.png
61	1	5	Impact Wrenches	6	t	uploads/sub-category-images/hardware/power_tools/impact_wrenches.png
62	1	5	Rotary Hammers	7	t	uploads/sub-category-images/hardware/power_tools/rotary_hammer.png
63	1	5	Cordless Tools	8	t	uploads/sub-category-images/hardware/power_tools/coardless_tool.png
64	1	6	Hammers	1	t	uploads/sub-category-images/hardware/power_tools/hammer.png
65	1	6	Screwdrivers	2	t	uploads/sub-category-images/hardware/power_tools/screwdrivers.png
66	1	6	Wrenches	3	t	uploads/sub-category-images/hardware/power_tools/wrenches.png
67	1	6	Pliers	4	t	uploads/sub-category-images/hardware/power_tools/pillers.png
68	1	6	Chisels	5	t	uploads/sub-category-images/hardware/power_tools/chisels.png
69	1	6	Hand Saws	6	t	uploads/sub-category-images/hardware/power_tools/hand_saws.png
71	1	6	Tool Kits	8	t	uploads/sub-category-images/hardware/power_tools/tool_kit.png
73	1	7	Work Benches	2	t	uploads/sub-category-images/hardware/hardware_accessories/work_banches.png
75	1	7	Extension Cords	4	t	uploads/sub-category-images/hardware/hardware_accessories/extention_cords.png
77	1	8	Aluminum Ladders	1	t	uploads/sub-category-images/hardware/Ladders/Aluminum_ladders.png
78	1	8	Steel Ladders	2	t	uploads/sub-category-images/hardware/Ladders/steel_ladder.png
80	1	8	Extension Ladders	4	t	uploads/sub-category-images/hardware/Ladders/extention_ladders.png
86	1	10	Hacksaws	1	t	uploads/sub-category-images/hardware/cutting_tools/hacksaws.png
88	1	10	Bolt Cutters	3	t	uploads/sub-category-images/hardware/cutting_tools/bolt_cutters.png
90	1	10	Utility Knives	5	t	uploads/sub-category-images/hardware/cutting_tools/utility_knives.png
92	1	11	Safety Gloves	2	t	uploads/sub-category-images/hardware/safety_equipment/safety_gloves.png
94	1	11	Ear Protection	4	t	uploads/sub-category-images/hardware/safety_equipment/ear_protection.png
96	1	12	Shovels & Spades	1	t	uploads/sub-category-images/hardware/garden_tools/shovels.png
110	2	16	Sand Bags	2	t	uploads/sub-category-images/material/building/sand_bags.jpg
111	2	16	Bricks	3	t	uploads/sub-category-images/material/building/bricks.jpg
113	2	16	Reinforcement Bars	5	t	uploads/sub-category-images/material/building/reinforcement_bars.jpg
115	2	17	TMT Bars 10mm	2	t	uploads/sub-category-images/material/steel_metals/tmt_bars_10mm.jpg
117	2	17	TMT Bars 16mm	4	t	uploads/sub-category-images/material/steel_metals/tmt_bars_16mm.jpg
118	2	17	MS Round Bars	5	t	uploads/sub-category-images/material/steel_metals/ms_round_bars.jpg
121	2	17	GI Sheets	8	t	\N
122	2	18	Teak Wood	1	t	\N
123	2	18	Sal Wood	2	t	\N
124	2	18	Pine Wood	3	t	\N
125	2	18	Plywood 4x8	4	t	\N
126	2	18	Flush Doors	5	t	\N
127	2	18	Timber Logs	6	t	\N
128	2	19	GI Roofing Sheets	1	t	\N
129	2	19	Color Coated Sheets	2	t	\N
130	2	19	Polycarbonate Sheets	3	t	\N
131	2	19	Asbestos Sheets	4	t	\N
132	2	19	Roofing Screws	5	t	\N
133	2	20	Binding Wire	1	t	\N
134	2	20	Construction Barricades	2	t	\N
135	2	20	Safety Nets	3	t	\N
136	2	20	Formwork Accessories	4	t	\N
137	2	21	OPC 43 Grade	1	t	\N
138	2	21	OPC 53 Grade	2	t	\N
139	2	21	PPC Cement	3	t	\N
140	2	21	Ready Mix Concrete	4	t	\N
141	2	22	River Sand	1	t	\N
142	2	22	M-Sand	2	t	\N
143	2	22	10mm Jelly	3	t	\N
144	2	22	20mm Jelly	4	t	\N
145	2	22	40mm Jelly	5	t	\N
146	2	23	Red Bricks	1	t	\N
147	2	23	Fly Ash Bricks	2	t	\N
148	2	23	Concrete Blocks	3	t	\N
149	2	23	AAC Blocks	4	t	\N
150	2	24	Flush Doors	1	t	\N
151	2	24	Membrane Doors	2	t	\N
152	2	24	PVC Doors	3	t	\N
153	2	24	Aluminum Windows	4	t	\N
154	2	24	UPVC Windows	5	t	\N
155	2	25	Commercial Plywood	1	t	\N
156	2	25	Marine Plywood	2	t	\N
157	2	25	BWP Plywood	3	t	\N
158	2	25	Laminates 1mm	4	t	\N
159	2	26	Gypsum Boards	1	t	\N
160	2	26	POP Ceiling	2	t	\N
161	2	26	Wall Plaster	3	t	\N
162	2	26	Plaster of Paris	4	t	\N
163	2	27	Ceramic Tiles	1	t	\N
164	2	27	Vitrified Tiles	2	t	\N
165	2	27	Porcelain Tiles	3	t	\N
166	2	27	Marble	4	t	\N
167	2	27	Granite	5	t	\N
168	2	28	I-Beams	1	t	\N
169	2	28	H-Beams	2	t	\N
170	2	28	Channels	3	t	\N
171	2	28	Angles	4	t	\N
172	2	28	MS Plates	5	t	\N
173	2	29	PVC Pipes	1	t	\N
174	2	29	CPVC Pipes	2	t	\N
175	2	29	GI Pipes	3	t	\N
176	2	29	PVC Fittings	4	t	\N
177	2	30	Waterproofing Chemicals	1	t	\N
178	2	30	Concrete Admixtures	2	t	\N
179	2	30	Wall Putty	3	t	\N
180	3	32	3 Seater Sofas	1	t	\N
181	3	32	L-Shaped Sofas	2	t	\N
182	3	32	Recliner Sofas	3	t	\N
183	3	32	Sofa Cum Beds	4	t	\N
184	3	32	Coffee Tables	5	t	\N
185	3	32	Side Tables	6	t	\N
186	3	32	Wall Shelves	7	t	\N
187	3	33	Single Beds	1	t	\N
188	3	33	Double Beds	2	t	\N
189	3	33	King Size Beds	3	t	\N
190	3	33	Queen Size Beds	4	t	\N
191	3	33	Wardrobes 3 Door	5	t	\N
192	3	33	Dressing Tables	6	t	\N
99	1	12	Garden Rakes	4	t	uploads/sub-category-images/hardware/garden_tools/garden_rakes.png
101	1	13	Welding Electrodes	1	t	uploads/sub-category-images/hardware/welding/welding_electrodes.png
102	1	13	Welding Rods	2	t	uploads/sub-category-images/hardware/welding/welding_rods.png
104	1	13	Welding Gloves	4	t	uploads/sub-category-images/hardware/welding/welding_gloves.png
106	1	14	Araldite	2	t	uploads/sub-category-images/hardware/sealants/araldite.png
108	1	14	Silicone Sealants	4	t	uploads/sub-category-images/hardware/sealants/silicone.png
566	14	1	Bottles & Nipples	1	t	/images/baby/bottles.jpg
567	14	1	Formula Milk	2	t	/images/baby/formula.jpg
193	3	33	Bedside Tables	7	t	\N
194	3	33	Study Tables	8	t	\N
195	3	34	Dining Tables 6 Seater	1	t	\N
196	3	34	Dining Sets	2	t	\N
197	3	34	Kitchen Cabinets	3	t	\N
198	3	34	Trolleys	4	t	\N
199	3	35	Office Tables	1	t	\N
200	3	35	Office Chairs	2	t	\N
201	3	35	Filing Cabinets	3	t	\N
202	3	35	Office Storage	4	t	\N
203	3	36	Outdoor Sofas	1	t	\N
204	3	36	Garden Chairs	2	t	\N
205	3	36	Outdoor Tables	3	t	\N
206	3	36	Hammocks	4	t	\N
212	3	38	Orthopedic Mattresses	1	t	\N
213	3	38	Memory Foam Mattresses	2	t	\N
214	3	38	Spring Mattresses	3	t	\N
215	3	38	Single Size	4	t	\N
216	3	38	Queen Size	5	t	\N
217	3	38	King Size	6	t	\N
218	3	39	Fabric Sofas	1	t	\N
219	3	39	Leatherette Sofas	2	t	\N
220	3	39	Recliner Sofas	3	t	\N
221	3	39	Sectional Sofas	4	t	\N
222	3	40	Single Beds	1	t	\N
223	3	40	Double Beds	2	t	\N
224	3	40	Queen Beds	3	t	\N
225	3	40	King Beds	4	t	\N
226	3	40	Bunk Beds	5	t	\N
227	3	41	2 Door Wardrobes	1	t	\N
228	3	41	3 Door Wardrobes	2	t	\N
229	3	41	4 Door Wardrobes	3	t	\N
230	3	41	Sliding Door Wardrobes	4	t	\N
231	3	42	Coffee Tables	1	t	\N
232	3	42	Center Tables	2	t	\N
233	3	42	Dining Tables	3	t	\N
234	3	42	Study Tables	4	t	\N
235	3	43	Dining Chairs	1	t	\N
236	3	43	Office Chairs	2	t	\N
237	3	43	Recliner Chairs	3	t	\N
238	3	43	Rocking Chairs	4	t	\N
239	3	44	Shoe Racks	1	t	\N
240	3	44	Kitchen Cabinets	2	t	\N
241	3	44	Wall Cabinets	3	t	\N
242	3	45	Wall Mount TV Units	1	t	\N
243	3	45	Floor Standing TV Units	2	t	\N
244	3	45	Entertainment Units	3	t	\N
245	3	46	Wall Shelves	1	t	\N
246	3	46	Floor Bookcases	2	t	\N
247	3	46	Corner Shelves	3	t	\N
248	3	46	Floating Shelves	4	t	\N
249	4	48	Samsung Galaxy	1	t	\N
250	4	48	Apple iPhone	2	t	\N
251	4	48	OnePlus Nord	3	t	\N
252	4	48	Realme Series	4	t	\N
253	4	48	Vivo Y Series	5	t	\N
254	4	48	Mi Redmi	6	t	\N
255	4	48	Oppo Reno	7	t	\N
256	4	49	Samsung Tablets	1	t	\N
257	4	49	Apple iPads	2	t	\N
258	4	49	Lenovo Tablets	3	t	\N
259	4	49	10-inch Tablets	4	t	\N
260	4	50	Gaming Laptops	1	t	\N
261	4	50	Business Laptops	2	t	\N
262	4	50	Ultrabooks	3	t	\N
263	4	50	HP Laptops	4	t	\N
264	4	50	Dell Laptops	5	t	\N
265	4	50	Lenovo Laptops	6	t	\N
266	4	51	Gaming PCs	1	t	\N
267	4	51	Office PCs	2	t	\N
268	4	51	All-in-One PCs	3	t	\N
269	4	51	Desktop Towers	4	t	\N
270	4	52	Samsung LED TVs	1	t	\N
271	4	52	LG Smart TVs	2	t	\N
272	4	52	Sony Bravia	3	t	\N
273	4	52	32 inch TVs	4	t	\N
274	4	52	43 inch TVs	5	t	\N
275	4	52	55 inch TVs	6	t	\N
276	4	52	65 inch TVs	7	t	\N
277	4	53	Air Conditioners 1 Ton	1	t	\N
278	4	53	Air Conditioners 1.5 Ton	2	t	\N
279	4	53	Refrigerators 200L	3	t	\N
280	4	53	Washing Machines 7kg	4	t	\N
281	4	53	Water Purifiers	5	t	\N
282	4	54	Microwave Ovens	1	t	\N
283	4	54	Induction Cooktops	2	t	\N
284	4	54	Mixer Grinders	3	t	\N
285	4	54	Juicers	4	t	\N
286	4	54	OTGs	5	t	\N
287	4	55	Soundbars	1	t	\N
209	3	37	Dining Tables	3	t	uploads/sub-category-images/dining_room_sub_category_3.jpg
211	3	37	Sideboards	5	t	uploads/sub-category-images/dining_room_sub_category_6.jpg
288	4	55	Home Theatre Systems	2	t	\N
289	4	55	Bluetooth Speakers	3	t	\N
290	4	56	DSLR Cameras	1	t	\N
291	4	56	Mirrorless Cameras	2	t	\N
292	4	56	Point & Shoot	3	t	\N
293	4	56	CCTV Cameras	4	t	\N
294	4	57	PlayStation 5	1	t	\N
295	4	57	Xbox Series X	2	t	\N
296	4	57	Nintendo Switch	3	t	\N
297	4	57	Gaming Accessories	4	t	\N
298	4	58	Mobile Covers	1	t	\N
299	4	58	Chargers	2	t	\N
300	4	58	Screen Guards	3	t	\N
301	4	58	Earphones	4	t	\N
302	4	59	Routers	1	t	\N
303	4	59	WiFi Extenders	2	t	\N
304	4	59	LAN Cables	3	t	\N
305	4	59	Network Switches	4	t	\N
306	4	60	Smart Bulbs	1	t	\N
307	4	60	Smart Plugs	2	t	\N
308	4	60	Smart Cameras	3	t	\N
309	4	60	Smart Locks	4	t	\N
310	4	61	Smartwatches	1	t	\N
311	4	61	Fitness Bands	2	t	\N
312	4	61	Smart Rings	3	t	\N
313	4	62	Inkjet Printers	1	t	\N
314	4	62	Laser Printers	2	t	\N
315	4	62	All-in-One Printers	3	t	\N
316	4	62	Scanners	4	t	\N
317	4	63	Pen Drives 32GB	1	t	\N
318	4	63	Pen Drives 64GB	2	t	\N
319	4	63	External HDD 1TB	3	t	\N
320	4	63	SSD Drives	4	t	\N
321	4	64	Bluetooth Speakers	1	t	\N
322	4	64	Portable Speakers	2	t	\N
323	4	64	Tower Speakers	3	t	\N
324	4	65	Wired Headphones	1	t	\N
325	4	65	Wireless Headphones	2	t	\N
326	4	65	Gaming Headsets	3	t	\N
327	4	66	10000mAh Power Banks	1	t	\N
328	4	66	20000mAh Power Banks	2	t	\N
329	4	66	Wireless Power Banks	3	t	\N
330	4	67	Ceiling Fans	1	t	\N
331	4	67	Table Fans	2	t	\N
332	4	67	Pedestal Fans	3	t	\N
333	4	67	Air Coolers	4	t	\N
334	5	69	Basmati Rice	1	t	\N
335	5	69	Sona Masoori Rice	2	t	\N
336	5	69	Wheat Flour	3	t	\N
337	5	69	Toor Dal	4	t	\N
338	5	69	Cooking Oils	5	t	\N
339	5	69	Sugar	6	t	\N
340	5	69	Salt	7	t	\N
341	5	70	Haldiram Namkeen	1	t	\N
342	5	70	Lays Chips	2	t	\N
343	5	70	Parle-G Biscuits	3	t	\N
344	5	70	Marie Gold	4	t	\N
345	5	70	Good Day Cookies	5	t	\N
346	5	71	Thums Up 2L	1	t	\N
347	5	71	Sprite 2L	2	t	\N
348	5	71	Maaza Mango	3	t	\N
349	5	71	Mineral Water	4	t	\N
350	5	71	Tea Powder	5	t	\N
351	5	71	Coffee Powder	6	t	\N
352	5	72	Surf Excel	1	t	\N
353	5	72	Rin Detergent	2	t	\N
354	5	72	Vim Liquid	3	t	\N
355	5	72	Lizol Floor Cleaner	4	t	\N
356	5	72	Odonil Air Freshener	5	t	\N
357	5	73	Amul Milk 500ml	1	t	\N
358	5	73	Amul Butter	2	t	\N
359	5	73	Amul Ghee	3	t	\N
360	5	73	Paneer 200g	4	t	\N
361	5	73	Curd 500g	5	t	\N
362	5	74	Britannia Bread	1	t	\N
363	5	74	Milk Bread	2	t	\N
364	5	74	Pav Buns	3	t	\N
365	5	74	Rusks	4	t	\N
366	5	74	Eggless Cookies	5	t	\N
367	5	75	Everest Haldi Powder	1	t	\N
368	5	75	Everest Mirchi Powder	2	t	\N
369	5	75	Everest Dhaniya Powder	3	t	\N
370	5	75	Everest Garam Masala	4	t	\N
371	5	75	MDH Chana Masala	5	t	\N
372	5	76	Fortune Sunflower Oil	1	t	\N
373	5	76	Sundrop Oil	2	t	\N
374	5	76	Mustard Oil	3	t	\N
375	5	76	Rice Bran Oil	4	t	\N
376	5	77	Clinic Plus Shampoo	1	t	\N
377	5	77	Lifebuoy Soap	2	t	\N
378	5	77	Pepsodent Toothpaste	3	t	\N
379	5	77	Dettol Handwash	4	t	\N
380	5	78	Pampers Diapers	1	t	\N
381	5	78	Johnson Baby Soap	2	t	\N
382	5	78	Cerelac Baby Food	3	t	\N
383	5	79	Protein Powder	1	t	\N
384	5	79	Multivitamins	2	t	\N
385	5	79	Honey	3	t	\N
386	5	80	Frozen Paratha	1	t	\N
387	5	80	Frozen Chicken	2	t	\N
388	5	80	Frozen Vegetables	3	t	\N
389	5	81	Maggi Noodles	1	t	\N
390	5	81	Yippee Noodles	2	t	\N
391	5	81	Ready to Cook Pasta	3	t	\N
392	5	82	Organic Rice	1	t	\N
393	5	82	Organic Dal	2	t	\N
394	5	82	Organic Honey	3	t	\N
395	5	83	Potato 1kg	1	t	\N
396	5	83	Onion 1kg	2	t	\N
397	5	83	Tomato 1kg	3	t	\N
398	5	83	Fresh Fruits	4	t	\N
399	5	84	Gulab Jamun	1	t	\N
400	5	84	Barfi	2	t	\N
401	5	84	Ladoo	3	t	\N
402	5	84	Haldiram Sweets	4	t	\N
403	5	85	Dog Food	1	t	\N
404	5	85	Cat Food	2	t	\N
405	5	85	Pet Toys	3	t	\N
406	6	87	Cello Pens	1	t	\N
407	6	87	Pilot Pens	2	t	\N
408	6	87	A4 Paper	3	t	\N
409	6	87	Staplers	4	t	\N
410	6	87	Paper Clips	5	t	\N
411	6	88	Natraj Pencils	1	t	\N
412	6	88	Camlin Geometry Box	2	t	\N
413	6	88	Classmate Notebooks	3	t	\N
414	6	88	Erasers	4	t	\N
415	6	88	Sharpeners	5	t	\N
416	6	89	Camlin Color Pencils	1	t	\N
417	6	89	FeviCryl Colors	2	t	\N
418	6	89	Sketch Pens	3	t	\N
419	6	89	Drawing Books	4	t	\N
420	6	90	Brown Envelopes	1	t	\N
421	6	90	White Envelopes	2	t	\N
422	6	90	Bubble Wrap	3	t	\N
423	6	90	Packing Tape	4	t	\N
424	6	91	Gel Pens	1	t	\N
425	6	91	Ball Pens	2	t	\N
426	6	91	Fountain Pens	3	t	\N
427	6	91	Markers	4	t	\N
428	6	92	A4 Notebooks	1	t	\N
429	6	92	A5 Notebooks	2	t	\N
430	6	92	Spiral Notebooks	3	t	\N
431	6	92	Diaries	4	t	\N
432	6	93	Lever Arch Files	1	t	\N
433	6	93	Box Files	2	t	\N
434	6	93	Hanging Folders	3	t	\N
435	6	93	Document Folders	4	t	\N
436	6	94	Pen Holders	1	t	\N
437	6	94	Paper Weights	2	t	\N
438	6	94	Desk Organizers	3	t	\N
439	6	94	Clipboards	4	t	\N
440	6	95	Ink Cartridges	1	t	\N
441	6	95	Toner Cartridges	2	t	\N
442	6	95	Printer Paper	3	t	\N
443	6	96	Set Squares	1	t	\N
444	6	96	Protractors	2	t	\N
445	6	96	Drawing Boards	3	t	\N
446	6	96	T-Squares	4	t	\N
447	6	97	Craft Paper	1	t	\N
448	6	97	Gluesticks	2	t	\N
449	6	97	Craft Scissors	3	t	\N
450	6	98	Whiteboard Markers	1	t	\N
451	6	98	Whiteboard Erasers	2	t	\N
452	6	98	Magnetic Whiteboards	3	t	\N
453	6	99	Basic Calculators	1	t	\N
454	6	99	Scientific Calculators	2	t	\N
455	6	100	Pointers	1	t	\N
456	6	100	Flip Charts	2	t	\N
457	6	100	Presentation Boards	3	t	\N
458	6	101	Correction Pens	1	t	\N
459	6	101	Correction Tapes	2	t	\N
460	6	101	Erasers	3	t	\N
461	6	102	Post-it Notes	1	t	\N
462	6	102	Transparent Tape	2	t	\N
463	6	102	Masking Tape	3	t	\N
464	6	103	A4 Envelopes	1	t	\N
465	6	103	C6 Envelopes	2	t	\N
466	6	103	Address Labels	3	t	\N
467	7	105	Formal Shirts	1	t	\N
468	7	105	Casual Shirts	2	t	\N
469	7	105	T-Shirts	3	t	\N
470	7	105	Polo Shirts	4	t	\N
471	7	105	Kurtas	5	t	\N
472	7	106	Formal Trousers	1	t	\N
473	7	106	Jeans	2	t	\N
474	7	106	Chinos	3	t	\N
475	7	106	Cargo Pants	4	t	\N
476	7	106	Track Pants	5	t	\N
477	7	107	Men Suits	1	t	\N
478	7	107	Sherwanis	2	t	\N
479	7	107	Suit Pieces	3	t	\N
480	7	108	Belts	1	t	\N
481	7	108	Ties	2	t	\N
482	7	108	Handkerchiefs	3	t	\N
483	7	108	Wallets	4	t	\N
484	7	109	Maxi Dresses	1	t	\N
485	7	109	Party Dresses	2	t	\N
486	7	109	Casual Dresses	3	t	\N
487	7	110	T-Shirts	1	t	\N
488	7	110	Crop Tops	2	t	\N
489	7	110	Tank Tops	3	t	\N
490	7	110	Blouses	4	t	\N
491	7	111	Leggings	1	t	\N
492	7	111	Skirts	2	t	\N
493	7	111	Palazzos	3	t	\N
494	7	111	Shorts	4	t	\N
495	7	112	Bras	1	t	\N
496	7	112	Panties	2	t	\N
497	7	112	Nightwear	3	t	\N
498	7	113	Baby Frocks	1	t	\N
499	7	113	Rompers	2	t	\N
500	7	114	Kids T-Shirts	1	t	\N
501	7	114	Kids Jeans	2	t	\N
502	7	114	Kids Frocks	3	t	\N
503	7	115	Junior Tops	1	t	\N
504	7	115	Junior Bottoms	2	t	\N
505	7	116	Gym T-Shirts	1	t	\N
506	7	116	Gym Track Pants	2	t	\N
507	7	116	Sports Bras	3	t	\N
508	7	117	Jackets	1	t	\N
509	7	117	Blazers	2	t	\N
510	7	117	Coats	3	t	\N
511	7	118	Women Track Suits	1	t	\N
512	7	118	Women Leggings	2	t	\N
513	7	119	Men Track Suits	1	t	\N
514	7	119	Men Sports Shorts	2	t	\N
515	7	120	Sarees	1	t	\N
516	7	120	Salwar Suits	2	t	\N
517	7	120	Lehenga Cholis	3	t	\N
518	7	120	Kurtas Pyjamas	4	t	\N
519	7	121	Formal Shirts	1	t	\N
520	7	121	Formal Trousers	2	t	\N
521	7	121	Blazers	3	t	\N
522	7	122	Casual Shirts	1	t	\N
523	7	122	T-Shirts	2	t	\N
524	7	122	Jeans	3	t	\N
525	7	122	Shorts	4	t	\N
526	7	123	Nighties	1	t	\N
527	7	123	Pyjamas	2	t	\N
528	7	124	Swim Suits	1	t	\N
529	7	124	Bikinis	2	t	\N
530	7	125	Sweaters	1	t	\N
531	7	125	Jackets	2	t	\N
532	7	125	Mufflers	3	t	\N
533	7	126	Mens Formal Shoes	1	t	\N
534	7	126	Mens Casual Shoes	2	t	\N
535	7	126	Sports Shoes	3	t	\N
536	7	126	Sandals	4	t	\N
537	7	126	Ladies Heels	5	t	\N
538	7	126	Kids Shoes	6	t	\N
539	7	127	Backpacks	1	t	\N
540	7	127	Handbags	2	t	\N
541	7	127	Tote Bags	3	t	\N
542	7	128	Gold Jewelry	1	t	\N
543	7	128	Silver Jewelry	2	t	\N
544	7	128	Artificial Jewelry	3	t	\N
545	7	129	Mens Sunglasses	1	t	\N
546	7	129	Women Sunglasses	2	t	\N
547	7	130	Mens Watches	1	t	\N
548	7	130	Ladies Watches	2	t	\N
549	7	130	Smart Watches	3	t	\N
550	7	131	Mens Vests	1	t	\N
551	7	131	Mens Briefs	2	t	\N
552	7	131	Ladies Innerwear	3	t	\N
553	7	132	Plus Size Tops	1	t	\N
554	7	132	Plus Size Bottoms	2	t	\N
555	7	133	Bridal Sarees	1	t	\N
556	7	133	Bridal Lehengas	2	t	\N
557	7	133	Groom Sherwanis	3	t	\N
558	7	134	Party Sarees	1	t	\N
559	7	134	Party Salwars	2	t	\N
207	3	37	6 Seater Dining Sets	1	t	uploads/sub-category-images/dining_room_sub_category_1.jpg
208	3	37	8 Seater Dining Sets	2	t	uploads/sub-category-images/dining_room_sub_category_2.jpg
210	3	37	Dining Chairs	4	t	uploads/sub-category-images/dining_room_sub_category_5.jpg
109	2	16	Cement Bags	1	t	uploads/sub-category-images/material/building/cement_bags.jpg
112	2	16	Concrete Blocks	4	t	uploads/sub-category-images/material/building/concrete_blocks.jpg
114	2	17	TMT Bars 8mm	1	t	uploads/sub-category-images/material/steel_metals/tmt_bars_8mm.jpg
116	2	17	TMT Bars 12mm	3	t	uploads/sub-category-images/material/steel_metals/tmt_bars_12mm.jpg
119	2	17	MS Angles	6	t	uploads/sub-category-images/material/steel_metals/ms_angles.jpg
568	14	1	Nursing Pillows	3	t	/images/baby/nursing-pillow.jpg
569	14	1	Sterilizers	4	t	/images/baby/sterilizer.jpg
1	1	1	Machine Screws	1	t	uploads/sub-category-images/hardware/fastners/machine_screws.png
2	1	1	Self Tapping Screws	2	t	uploads/sub-category-images/hardware/fastners/self_taping_screws.png
3	1	1	Wood Screws	3	t	uploads/sub-category-images/hardware/fastners/wood_screws.png
4	1	1	Coach Screws	4	t	uploads/sub-category-images/hardware/fastners/coach_screws.png
5	1	1	Anchor Screws	5	t	uploads/sub-category-images/hardware/fastners/anchor_screws.png
6	1	1	Common Nails	6	t	uploads/sub-category-images/hardware/fastners/common_nails.png
7	1	1	Finishing Nails	7	t	uploads/sub-category-images/hardware/fastners/finishing_nails.png
8	1	1	Concrete Nails	8	t	uploads/sub-category-images/hardware/fastners/concrete_nails.png
120	2	17	MS Channels	7	t	uploads/sub-category-images/material/steel_metals/ms_channels.jpg
10	1	1	Carriage Bolts	10	t	uploads/sub-category-images/hardware/fastners/wing_nuts.png
11	1	1	Eye Bolts	11	t	uploads/sub-category-images/hardware/fastners/plain_washers.png
561	9	139	Plumbing Services	2	t	uploads/sub-category-images/repair_service/plumbing_service.jpg
563	9	139	Appliance Repairs	4	t	uploads/sub-category-images/repair_service/appliance_service.jpg
564	9	139	Home Civil Work	5	t	uploads/sub-category-images/repair_service/home_civil_work.jpg
9	1	1	Hex Bolts	9	t	uploads/sub-category-images/hardware/fastners/hex_bolts.png
565	9	139	Electronics works	6	t	uploads/sub-category-images/repair_service/electronics_gadget_service.jpg
12	1	1	Hex Nuts	12	t	uploads/sub-category-images/hardware/fastners/hex_nuts.png
13	1	1	Wing Nuts	13	t	uploads/sub-category-images/hardware/fastners/wing_nuts.png
14	1	1	Lock Washers	14	t	uploads/sub-category-images/hardware/fastners/lock_washers.png
15	1	1	Plain Washers	15	t	uploads/sub-category-images/hardware/fastners/plain_washers.png
32	1	3	Electrical Sockets	7	t	uploads/sub-category-images/hardware/electricle/electricle_switches.png
39	1	4	Exterior Wall Paints	4	t	uploads/sub-category-images/hardware/paint/external_wall_paints.png
41	1	4	Primer Sealer	6	t	uploads/sub-category-images/hardware/paint/primer_sealer.png
44	1	4	Paint Thinner	9	t	uploads/sub-category-images/hardware/paint/paint_thinner.png
70	1	6	Files	7	t	uploads/sub-category-images/hardware/power_tools/files.png
72	1	7	Tool Boxes	1	t	uploads/sub-category-images/hardware/hardware_accessories/tool_boxes.png
74	1	7	Tool Holders	3	t	uploads/sub-category-images/hardware/hardware_accessories/tool_holders.png
76	1	7	Tool Bags	5	t	uploads/sub-category-images/hardware/hardware_accessories/tool_bags.png
79	1	8	Step Ladders	3	t	uploads/sub-category-images/hardware/Ladders/step_ledders.png
82	1	9	Measuring Tapes	1	t	uploads/sub-category-images/hardware/measuring_tools/measuring_tapes.png
87	1	10	Pipe Cutters	2	t	uploads/sub-category-images/hardware/cutting_tools/pipe_cutters.png
89	1	10	Aviation Snips	4	t	uploads/sub-category-images/hardware/cutting_tools/aviation_snips.png
91	1	11	Safety Helmets	1	t	uploads/sub-category-images/hardware/safety_equipment/safety_helmets.png
93	1	11	Safety Goggles	3	t	uploads/sub-category-images/hardware/safety_equipment/safety_goggles.png
95	1	11	Face Masks	5	t	uploads/sub-category-images/hardware/safety_equipment/face_masks.png
97	1	12	Garden Forks	2	t	uploads/sub-category-images/hardware/garden_tools/garden_forks.png
98	1	12	Pruning Shears	3	t	uploads/sub-category-images/hardware/garden_tools/pruning_shears.png
100	1	12	Watering Cans	5	t	uploads/sub-category-images/hardware/garden_tools/watering_cans.png
103	1	13	Welding Helmets	3	t	uploads/sub-category-images/hardware/welding/welding_helmets.png
105	1	14	FeviKwik	1	t	uploads/sub-category-images/hardware/sealants/fevikwilk.png
107	1	14	Epoxy Adhesives	3	t	uploads/sub-category-images/hardware/sealants/epoxy.png
560	9	139	Electrical Repairs	1	t	uploads/sub-category-images/repair_service/carpentry_service.jpg
562	9	139	Carpentry Services	3	t	uploads/sub-category-images/repair_service/electricle_service.jpg
570	14	2	Newborn Diapers	1	t	/images/baby/diapers-newborn.jpg
571	14	2	Baby Wipes	2	t	/images/baby/wipes.jpg
572	14	2	Rash Creams	3	t	/images/baby/rash-cream.jpg
573	14	3	Newborn Onesies	1	t	/images/baby/onesies.jpg
574	14	3	Sleepwear Sets	2	t	/images/baby/sleepwear.jpg
575	14	3	Rompers	3	t	/images/baby/rompers.jpg
576	14	4	Baby Shampoo	1	t	/images/baby/shampoo.jpg
577	14	4	Body Lotion	2	t	/images/baby/lotion.jpg
578	14	4	Baby Bath Tub	3	t	/images/baby/bathtub.jpg
579	15	1	Dry Dog Food	1	t	/images/pet/dry-dog-food.jpg
580	15	1	Dry Cat Food	2	t	/images/pet/dry-cat-food.jpg
581	15	1	Wet Dog Food	3	t	/images/pet/wet-dog-food.jpg
582	15	1	Wet Cat Food	4	t	/images/pet/wet-cat-food.jpg
583	15	1	Dog Treats	5	t	/images/pet/dog-treats.jpg
584	15	1	Cat Treats	6	t	/images/pet/cat-treats.jpg
585	15	1	Bird Food	7	t	/images/pet/bird-food.jpg
586	15	1	Fish Food	8	t	/images/pet/fish-food.jpg
587	15	2	Chew Toys	1	t	/images/pet/chew-toys.jpg
588	15	2	Interactive Toys	2	t	/images/pet/interactive-toys.jpg
589	15	2	Plush Toys	3	t	/images/pet/plush-toys.jpg
590	15	2	Fetch Balls	4	t	/images/pet/fetch-balls.jpg
591	15	2	Laser Toys	5	t	/images/pet/laser-toys.jpg
592	15	2	Catnip Toys	6	t	/images/pet/catnip-toys.jpg
593	15	3	Pet Shampoo	1	t	/images/pet/pet-shampoo.jpg
594	15	3	Flea Treatment	2	t	/images/pet/flea-treatment.jpg
595	15	3	Pet Brushes	3	t	/images/pet/pet-brush.jpg
596	15	3	Nail Clippers	4	t	/images/pet/nail-clippers.jpg
597	15	3	Pet Wipes	5	t	/images/pet/pet-wipes.jpg
598	15	3	Ear Cleaners	6	t	/images/pet/ear-cleaner.jpg
599	15	4	Dog Collars	1	t	/images/pet/dog-collar.jpg
600	15	4	Cat Collars	2	t	/images/pet/cat-collar.jpg
601	15	4	Leashes	3	t	/images/pet/leash.jpg
602	15	4	Pet Bowls	4	t	/images/pet/pet-bowl.jpg
603	15	4	ID Tags	5	t	/images/pet/id-tag.jpg
604	15	4	Harnesses	6	t	/images/pet/harness.jpg
605	15	5	Dog Beds	1	t	/images/pet/dog-bed.jpg
606	15	5	Cat Beds	2	t	/images/pet/cat-bed.jpg
607	15	5	Pet Carriers	3	t	/images/pet/pet-carrier.jpg
608	15	5	Travel Bowls	4	t	/images/pet/travel-bowl.jpg
609	15	5	Pet Crates	5	t	/images/pet/pet-crate.jpg
610	16	1	Dumbbells	1	t	/images/sports/dumbbells.jpg
611	16	1	Yoga Mats	2	t	/images/sports/yoga-mat.jpg
612	16	1	Resistance Bands	3	t	/images/sports/resistance-bands.jpg
613	16	1	Jump Ropes	4	t	/images/sports/jump-rope.jpg
614	16	2	Cricket Bats	1	t	/images/sports/cricket-bat.jpg
615	16	2	Football	2	t	/images/sports/football.jpg
616	16	2	Badminton Rackets	3	t	/images/sports/badminton-racket.jpg
617	16	3	Cycling Helmets	1	t	/images/sports/cycling-helmet.jpg
618	16	3	Camping Tents	2	t	/images/sports/camping-tent.jpg
619	16	3	Hiking Shoes	3	t	/images/sports/hiking-shoes.jpg
620	16	4	Track Pants	1	t	/images/sports/track-pants.jpg
621	16	4	Sports T-Shirts	2	t	/images/sports/sports-tshirt.jpg
622	16	4	Gym Gloves	3	t	/images/sports/gym-gloves.jpg
623	14	146	Bottles & Nipples	1	t	/images/baby/bottles.jpg
624	14	146	Formula Milk	2	t	/images/baby/formula.jpg
625	14	146	Nursing Pillows	3	t	/images/baby/nursing-pillow.jpg
626	14	146	Sterilizers	4	t	/images/baby/sterilizer.jpg
627	14	147	Newborn Diapers	1	t	/images/baby/diapers-newborn.jpg
628	14	147	Baby Wipes	2	t	/images/baby/wipes.jpg
629	14	147	Rash Creams	3	t	/images/baby/rash-cream.jpg
630	14	148	Newborn Onesies	1	t	/images/baby/onesies.jpg
631	14	148	Sleepwear Sets	2	t	/images/baby/sleepwear.jpg
632	14	148	Rompers	3	t	/images/baby/rompers.jpg
633	14	149	Baby Shampoo	1	t	/images/baby/shampoo.jpg
634	14	149	Body Lotion	2	t	/images/baby/lotion.jpg
635	14	149	Baby Bath Tub	3	t	/images/baby/bathtub.jpg
636	15	151	Dry Dog Food	1	t	/images/pet/dry-dog-food.jpg
637	15	151	Dry Cat Food	2	t	/images/pet/dry-cat-food.jpg
638	15	151	Wet Dog Food	3	t	/images/pet/wet-dog-food.jpg
639	15	151	Wet Cat Food	4	t	/images/pet/wet-cat-food.jpg
640	15	151	Dog Treats	5	t	/images/pet/dog-treats.jpg
641	15	151	Cat Treats	6	t	/images/pet/cat-treats.jpg
642	15	151	Bird Food	7	t	/images/pet/bird-food.jpg
643	15	151	Fish Food	8	t	/images/pet/fish-food.jpg
644	15	152	Chew Toys	1	t	/images/pet/chew-toys.jpg
645	15	152	Interactive Toys	2	t	/images/pet/interactive-toys.jpg
646	15	152	Plush Toys	3	t	/images/pet/plush-toys.jpg
647	15	152	Fetch Balls	4	t	/images/pet/fetch-balls.jpg
648	15	152	Laser Toys	5	t	/images/pet/laser-toys.jpg
649	15	152	Catnip Toys	6	t	/images/pet/catnip-toys.jpg
650	15	153	Pet Shampoo	1	t	/images/pet/pet-shampoo.jpg
651	15	153	Flea Treatment	2	t	/images/pet/flea-treatment.jpg
652	15	153	Pet Brushes	3	t	/images/pet/pet-brush.jpg
653	15	153	Nail Clippers	4	t	/images/pet/nail-clippers.jpg
654	15	153	Pet Wipes	5	t	/images/pet/pet-wipes.jpg
655	15	153	Ear Cleaners	6	t	/images/pet/ear-cleaner.jpg
656	15	154	Dog Collars	1	t	/images/pet/dog-collar.jpg
657	15	154	Cat Collars	2	t	/images/pet/cat-collar.jpg
658	15	154	Leashes	3	t	/images/pet/leash.jpg
659	15	154	Pet Bowls	4	t	/images/pet/pet-bowl.jpg
660	15	154	ID Tags	5	t	/images/pet/id-tag.jpg
661	15	154	Harnesses	6	t	/images/pet/harness.jpg
662	15	155	Dog Beds	1	t	/images/pet/dog-bed.jpg
663	15	155	Cat Beds	2	t	/images/pet/cat-bed.jpg
664	15	155	Pet Carriers	3	t	/images/pet/pet-carrier.jpg
665	15	155	Travel Bowls	4	t	/images/pet/travel-bowl.jpg
666	15	155	Pet Crates	5	t	/images/pet/pet-crate.jpg
667	16	156	Dumbbells	1	t	/images/sports/dumbbells.jpg
668	16	156	Yoga Mats	2	t	/images/sports/yoga-mat.jpg
669	16	156	Resistance Bands	3	t	/images/sports/resistance-bands.jpg
670	16	156	Jump Ropes	4	t	/images/sports/jump-rope.jpg
671	16	157	Cricket Bats	1	t	/images/sports/cricket-bat.jpg
672	16	157	Football	2	t	/images/sports/football.jpg
673	16	157	Badminton Rackets	3	t	/images/sports/badminton-racket.jpg
674	16	158	Cycling Helmets	1	t	/images/sports/cycling-helmet.jpg
675	16	158	Camping Tents	2	t	/images/sports/camping-tent.jpg
676	16	158	Hiking Shoes	3	t	/images/sports/hiking-shoes.jpg
677	16	159	Track Pants	1	t	/images/sports/track-pants.jpg
678	16	159	Sports T-Shirts	2	t	/images/sports/sports-tshirt.jpg
679	16	159	Gym Gloves	3	t	/images/sports/gym-gloves.jpg
680	17	161	Organic Rice	1	t	/images/organic/rice.jpg
681	17	161	Organic Millets	2	t	/images/organic/millets.jpg
682	17	161	Organic Pulses	3	t	/images/organic/pulses.jpg
683	17	161	Organic Spices	4	t	/images/organic/spices.jpg
684	17	162	Organic Soap	1	t	/images/organic/soap.jpg
685	17	162	Organic Shampoo	2	t	/images/organic/shampoo.jpg
686	17	162	Organic Toothpaste	3	t	/images/organic/toothpaste.jpg
687	17	163	Organic Honey	1	t	/images/organic/honey.jpg
688	17	163	Organic Oils	2	t	/images/organic/oils.jpg
689	17	163	Herbal Teas	3	t	/images/organic/herbal-tea.jpg
690	17	164	Organic Baby Lotion	1	t	/images/organic/baby-lotion.jpg
691	17	164	Organic Diapers	2	t	/images/organic/diapers.jpg
692	17	164	Organic Baby Food	3	t	/images/organic/baby-food.jpg
\.


--
-- Data for Name: sub_category_details; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sub_category_details (id, sub_category_id, sub_category_image_url) FROM stdin;
\.


--
-- Data for Name: sub_type_attribute_options; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sub_type_attribute_options (id, sub_type_attribute_id, option_value, sort_order) FROM stdin;
1	1	M2	1
2	1	M3	2
3	1	M4	3
4	1	M5	4
5	1	M6	5
6	1	M8	6
7	1	M10	7
8	1	M12	8
9	1	M16	9
10	1	M20	10
11	2	10mm	1
12	2	12mm	2
13	2	16mm	3
14	2	20mm	4
15	2	25mm	5
16	2	30mm	6
17	2	35mm	7
18	2	40mm	8
19	2	50mm	9
20	2	60mm	10
21	2	70mm	11
22	2	80mm	12
23	2	100mm	13
24	3	Stainless Steel	1
25	3	Carbon Steel	2
26	3	Galvanized Steel	3
27	3	Brass	4
28	3	Aluminum	5
29	3	Titanium	6
30	4	Zinc Plated	1
31	4	Black Oxide	2
32	4	Chrome Plated	3
33	4	Nickel Plated	4
34	4	Plain	5
35	4	Hot Dip Galvanized	6
36	5	Pan Head	1
37	5	Flat Head	2
38	5	Button Head	3
39	5	Cheese Head	4
40	5	Countersunk Head	5
41	5	Hex Head	6
42	5	Socket Head	7
43	6	Coarse Thread	1
44	6	Fine Thread	2
45	6	Extra Fine Thread	3
46	8	#2	1
47	8	#4	2
48	8	#6	3
49	8	#8	4
50	8	#10	5
51	8	#12	6
52	8	#14	7
53	9	10mm	1
54	9	12mm	2
55	9	16mm	3
56	9	20mm	4
57	9	25mm	5
58	9	30mm	6
59	9	40mm	7
60	9	50mm	8
61	9	75mm	9
62	9	100mm	10
63	10	Stainless Steel	1
64	10	Carbon Steel	2
65	10	Zinc Plated Steel	3
66	10	Brass	4
67	11	Zinc Plated	1
68	11	Black Oxide	2
69	11	Chrome Plated	3
70	11	Plain	4
71	12	Pan Head	1
72	12	Flat Head	2
73	12	Hex Head	3
74	12	Button Head	4
75	13	Type A	1
76	13	Type B	2
77	13	Type AB	3
78	13	Type 17	4
79	13	Type 23	5
80	15	#4	1
81	15	#6	2
82	15	#8	3
83	15	#10	4
84	15	#12	5
85	15	#14	6
86	16	12mm	1
87	16	16mm	2
88	16	20mm	3
89	16	25mm	4
90	16	30mm	5
91	16	40mm	6
92	16	50mm	7
93	16	60mm	8
94	16	75mm	9
95	16	100mm	10
96	17	Steel	1
97	17	Stainless Steel	2
98	17	Brass	3
99	17	Bronze	4
100	18	Zinc Plated	1
101	18	Black Oxide	2
102	18	Chrome Plated	3
103	18	Plain	4
104	18	Brass Plated	5
105	19	Flat Head	1
106	19	Round Head	2
107	19	Oval Head	3
108	19	Pan Head	4
109	20	Coarse Thread	1
110	20	Fine Thread	2
111	22	M6	1
112	22	M8	2
113	22	M10	3
114	22	M12	4
115	22	M16	5
116	22	M20	6
117	23	30mm	1
118	23	40mm	2
119	23	50mm	3
120	23	60mm	4
121	23	80mm	5
122	23	100mm	6
123	23	120mm	7
124	23	150mm	8
125	23	200mm	9
126	24	Steel	1
127	24	Stainless Steel	2
128	24	Galvanized Steel	3
129	25	Zinc Plated	1
130	25	Hot Dip Galvanized	2
131	25	Plain	3
132	26	Hex Head	1
133	26	Square Head	2
134	27	Coarse Thread	1
135	27	Wood Thread	2
136	29	M6	1
137	29	M8	2
138	29	M10	3
139	29	M12	4
140	29	M14	5
141	29	M16	6
142	30	40mm	1
143	30	50mm	2
144	30	60mm	3
145	30	80mm	4
146	30	100mm	5
147	30	120mm	6
148	31	Steel	1
149	31	Stainless Steel	2
150	31	Zinc Alloy	3
151	32	Zinc Plated	1
152	32	Plain	2
153	32	Black Oxide	3
154	33	Hex Head	1
155	33	Flat Head	2
156	33	Round Head	3
157	34	Coarse Thread	1
158	34	Expansion Type	2
159	36	2d	1
160	36	4d	2
161	36	6d	3
162	36	8d	4
163	36	10d	5
164	36	12d	6
165	36	16d	7
166	36	20d	8
167	37	25mm	1
168	37	30mm	2
169	37	40mm	3
170	37	50mm	4
171	37	65mm	5
172	37	75mm	6
173	37	90mm	7
174	37	100mm	8
175	38	Steel	1
176	38	Stainless Steel	2
177	38	Galvanized Steel	3
178	39	Bright	1
179	39	Galvanized	2
180	39	Hot Dip Galvanized	3
181	39	Vinyl Coated	4
182	43	XS	1
183	43	S	2
184	43	M	3
185	43	L	4
186	43	XL	5
187	43	XXL	6
188	43	XXXL	7
189	43	28	8
190	43	30	9
191	43	32	10
192	43	34	11
193	43	36	12
194	43	38	13
195	43	40	14
196	43	42	15
197	43	Free Size	16
198	50	XS	1
199	50	S	2
200	50	M	3
201	50	L	4
202	50	XL	5
203	50	XXL	6
204	50	XXXL	7
205	50	28	8
206	50	30	9
207	50	32	10
208	50	34	11
209	50	36	12
210	50	38	13
211	50	40	14
212	50	42	15
213	50	Free Size	16
214	57	XS	1
215	57	S	2
216	57	M	3
217	57	L	4
218	57	XL	5
219	57	XXL	6
220	57	XXXL	7
221	57	28	8
222	57	30	9
223	57	32	10
224	57	34	11
225	57	36	12
226	57	38	13
227	57	40	14
228	57	42	15
229	57	Free Size	16
230	64	XS	1
231	64	S	2
232	64	M	3
233	64	L	4
234	64	XL	5
235	64	XXL	6
236	64	XXXL	7
237	64	28	8
238	64	30	9
239	64	32	10
240	64	34	11
241	64	36	12
242	64	38	13
243	64	40	14
244	64	42	15
245	64	Free Size	16
246	71	XS	1
247	71	S	2
248	71	M	3
249	71	L	4
250	71	XL	5
251	71	XXL	6
252	71	XXXL	7
253	71	28	8
254	71	30	9
255	71	32	10
256	71	34	11
257	71	36	12
258	71	38	13
259	71	40	14
260	71	42	15
261	71	Free Size	16
262	78	XS	1
263	78	S	2
264	78	M	3
265	78	L	4
266	78	XL	5
267	78	XXL	6
268	78	XXXL	7
269	78	28	8
270	78	30	9
271	78	32	10
272	78	34	11
273	78	36	12
274	78	38	13
275	78	40	14
276	78	42	15
277	78	Free Size	16
278	83	XS	1
279	83	S	2
280	83	M	3
281	83	L	4
282	83	XL	5
283	83	XXL	6
284	83	XXXL	7
285	83	28	8
286	83	30	9
287	83	32	10
288	83	34	11
289	83	36	12
290	83	38	13
291	83	40	14
292	83	42	15
293	83	Free Size	16
294	88	XS	1
295	88	S	2
296	88	M	3
297	88	L	4
298	88	XL	5
299	88	XXL	6
300	88	XXXL	7
301	88	28	8
302	88	30	9
303	88	32	10
304	88	34	11
305	88	36	12
306	88	38	13
307	88	40	14
308	88	42	15
309	88	Free Size	16
310	92	XS	1
311	92	S	2
312	92	M	3
313	92	L	4
314	92	XL	5
315	92	XXL	6
316	92	XXXL	7
317	92	28	8
318	92	30	9
319	92	32	10
320	92	34	11
321	92	36	12
322	92	38	13
323	92	40	14
324	92	42	15
325	92	Free Size	16
326	116	XS	1
327	116	S	2
328	116	M	3
329	116	L	4
330	116	XL	5
331	116	XXL	6
332	116	XXXL	7
333	116	28	8
334	116	30	9
335	116	32	10
336	116	34	11
337	116	36	12
338	116	38	13
339	116	40	14
340	116	42	15
341	116	Free Size	16
342	120	XS	1
343	120	S	2
344	120	M	3
345	120	L	4
346	120	XL	5
347	120	XXL	6
348	120	XXXL	7
349	120	28	8
350	120	30	9
351	120	32	10
352	120	34	11
353	120	36	12
354	120	38	13
355	120	40	14
356	120	42	15
357	120	Free Size	16
358	123	XS	1
359	123	S	2
360	123	M	3
361	123	L	4
362	123	XL	5
363	123	XXL	6
364	123	XXXL	7
365	123	28	8
366	123	30	9
367	123	32	10
368	123	34	11
369	123	36	12
370	123	38	13
371	123	40	14
372	123	42	15
373	123	Free Size	16
374	127	XS	1
375	127	S	2
376	127	M	3
377	127	L	4
378	127	XL	5
379	127	XXL	6
380	127	XXXL	7
381	127	28	8
382	127	30	9
383	127	32	10
384	127	34	11
385	127	36	12
386	127	38	13
387	127	40	14
388	127	42	15
389	127	Free Size	16
390	135	XS	1
391	135	S	2
392	135	M	3
393	135	L	4
394	135	XL	5
395	135	XXL	6
396	135	XXXL	7
397	135	28	8
398	135	30	9
399	135	32	10
400	135	34	11
401	135	36	12
402	135	38	13
403	135	40	14
404	135	42	15
405	135	Free Size	16
406	181	XS	1
407	181	S	2
408	181	M	3
409	181	L	4
410	181	XL	5
411	181	XXL	6
412	181	XXXL	7
413	181	28	8
414	181	30	9
415	181	32	10
416	181	34	11
417	181	36	12
418	181	38	13
419	181	40	14
420	181	42	15
421	181	Free Size	16
422	232	XS	1
423	232	S	2
424	232	M	3
425	232	L	4
426	232	XL	5
427	232	XXL	6
428	232	XXXL	7
429	232	28	8
430	232	30	9
431	232	32	10
432	232	34	11
433	232	36	12
434	232	38	13
435	232	40	14
436	232	42	15
437	232	Free Size	16
438	276	XS	1
439	276	S	2
440	276	M	3
441	276	L	4
442	276	XL	5
443	276	XXL	6
444	276	XXXL	7
445	276	28	8
446	276	30	9
447	276	32	10
448	276	34	11
449	276	36	12
450	276	38	13
451	276	40	14
452	276	42	15
453	276	Free Size	16
454	369	XS	1
455	369	S	2
456	369	M	3
457	369	L	4
458	369	XL	5
459	369	XXL	6
460	369	XXXL	7
461	369	28	8
462	369	30	9
463	369	32	10
464	369	34	11
465	369	36	12
466	369	38	13
467	369	40	14
468	369	42	15
469	369	Free Size	16
470	459	XS	1
471	459	S	2
472	459	M	3
473	459	L	4
474	459	XL	5
475	459	XXL	6
476	459	XXXL	7
477	459	28	8
478	459	30	9
479	459	32	10
480	459	34	11
481	459	36	12
482	459	38	13
483	459	40	14
484	459	42	15
485	459	Free Size	16
486	463	XS	1
487	463	S	2
488	463	M	3
489	463	L	4
490	463	XL	5
491	463	XXL	6
492	463	XXXL	7
493	463	28	8
494	463	30	9
495	463	32	10
496	463	34	11
497	463	36	12
498	463	38	13
499	463	40	14
500	463	42	15
501	463	Free Size	16
502	474	XS	1
503	474	S	2
504	474	M	3
505	474	L	4
506	474	XL	5
507	474	XXL	6
508	474	XXXL	7
509	474	28	8
510	474	30	9
511	474	32	10
512	474	34	11
513	474	36	12
514	474	38	13
515	474	40	14
516	474	42	15
517	474	Free Size	16
518	479	XS	1
519	479	S	2
520	479	M	3
521	479	L	4
522	479	XL	5
523	479	XXL	6
524	479	XXXL	7
525	479	28	8
526	479	30	9
527	479	32	10
528	479	34	11
529	479	36	12
530	479	38	13
531	479	40	14
532	479	42	15
533	479	Free Size	16
534	517	XS	1
535	517	S	2
536	517	M	3
537	517	L	4
538	517	XL	5
539	517	XXL	6
540	517	XXXL	7
541	517	28	8
542	517	30	9
543	517	32	10
544	517	34	11
545	517	36	12
546	517	38	13
547	517	40	14
548	517	42	15
549	517	Free Size	16
550	520	XS	1
551	520	S	2
552	520	M	3
553	520	L	4
554	520	XL	5
555	520	XXL	6
556	520	XXXL	7
557	520	28	8
558	520	30	9
559	520	32	10
560	520	34	11
561	520	36	12
562	520	38	13
563	520	40	14
564	520	42	15
565	520	Free Size	16
566	523	XS	1
567	523	S	2
568	523	M	3
569	523	L	4
570	523	XL	5
571	523	XXL	6
572	523	XXXL	7
573	523	28	8
574	523	30	9
575	523	32	10
576	523	34	11
577	523	36	12
578	523	38	13
579	523	40	14
580	523	42	15
581	523	Free Size	16
582	548	XS	1
583	548	S	2
584	548	M	3
585	548	L	4
586	548	XL	5
587	548	XXL	6
588	548	XXXL	7
589	548	28	8
590	548	30	9
591	548	32	10
592	548	34	11
593	548	36	12
594	548	38	13
595	548	40	14
596	548	42	15
597	548	Free Size	16
598	550	XS	1
599	550	S	2
600	550	M	3
601	550	L	4
602	550	XL	5
603	550	XXL	6
604	550	XXXL	7
605	550	28	8
606	550	30	9
607	550	32	10
608	550	34	11
609	550	36	12
610	550	38	13
611	550	40	14
612	550	42	15
613	550	Free Size	16
614	578	XS	1
615	578	S	2
616	578	M	3
617	578	L	4
618	578	XL	5
619	578	XXL	6
620	578	XXXL	7
621	578	28	8
622	578	30	9
623	578	32	10
624	578	34	11
625	578	36	12
626	578	38	13
627	578	40	14
628	578	42	15
629	578	Free Size	16
630	582	XS	1
631	582	S	2
632	582	M	3
633	582	L	4
634	582	XL	5
635	582	XXL	6
636	582	XXXL	7
637	582	28	8
638	582	30	9
639	582	32	10
640	582	34	11
641	582	36	12
642	582	38	13
643	582	40	14
644	582	42	15
645	582	Free Size	16
646	587	XS	1
647	587	S	2
648	587	M	3
649	587	L	4
650	587	XL	5
651	587	XXL	6
652	587	XXXL	7
653	587	28	8
654	587	30	9
655	587	32	10
656	587	34	11
657	587	36	12
658	587	38	13
659	587	40	14
660	587	42	15
661	587	Free Size	16
662	609	XS	1
663	609	S	2
664	609	M	3
665	609	L	4
666	609	XL	5
667	609	XXL	6
668	609	XXXL	7
669	609	28	8
670	609	30	9
671	609	32	10
672	609	34	11
673	609	36	12
674	609	38	13
675	609	40	14
676	609	42	15
677	609	Free Size	16
678	630	XS	1
679	630	S	2
680	630	M	3
681	630	L	4
682	630	XL	5
683	630	XXL	6
684	630	XXXL	7
685	630	28	8
686	630	30	9
687	630	32	10
688	630	34	11
689	630	36	12
690	630	38	13
691	630	40	14
692	630	42	15
693	630	Free Size	16
694	634	XS	1
695	634	S	2
696	634	M	3
697	634	L	4
698	634	XL	5
699	634	XXL	6
700	634	XXXL	7
701	634	28	8
702	634	30	9
703	634	32	10
704	634	34	11
705	634	36	12
706	634	38	13
707	634	40	14
708	634	42	15
709	634	Free Size	16
710	637	XS	1
711	637	S	2
712	637	M	3
713	637	L	4
714	637	XL	5
715	637	XXL	6
716	637	XXXL	7
717	637	28	8
718	637	30	9
719	637	32	10
720	637	34	11
721	637	36	12
722	637	38	13
723	637	40	14
724	637	42	15
725	637	Free Size	16
726	656	XS	1
727	656	S	2
728	656	M	3
729	656	L	4
730	656	XL	5
731	656	XXL	6
732	656	XXXL	7
733	656	28	8
734	656	30	9
735	656	32	10
736	656	34	11
737	656	36	12
738	656	38	13
739	656	40	14
740	656	42	15
741	656	Free Size	16
742	675	XS	1
743	675	S	2
744	675	M	3
745	675	L	4
746	675	XL	5
747	675	XXL	6
748	675	XXXL	7
749	675	28	8
750	675	30	9
751	675	32	10
752	675	34	11
753	675	36	12
754	675	38	13
755	675	40	14
756	675	42	15
757	675	Free Size	16
758	678	XS	1
759	678	S	2
760	678	M	3
761	678	L	4
762	678	XL	5
763	678	XXL	6
764	678	XXXL	7
765	678	28	8
766	678	30	9
767	678	32	10
768	678	34	11
769	678	36	12
770	678	38	13
771	678	40	14
772	678	42	15
773	678	Free Size	16
774	681	XS	1
775	681	S	2
776	681	M	3
777	681	L	4
778	681	XL	5
779	681	XXL	6
780	681	XXXL	7
781	681	28	8
782	681	30	9
783	681	32	10
784	681	34	11
785	681	36	12
786	681	38	13
787	681	40	14
788	681	42	15
789	681	Free Size	16
790	684	XS	1
791	684	S	2
792	684	M	3
793	684	L	4
794	684	XL	5
795	684	XXL	6
796	684	XXXL	7
797	684	28	8
798	684	30	9
799	684	32	10
800	684	34	11
801	684	36	12
802	684	38	13
803	684	40	14
804	684	42	15
805	684	Free Size	16
806	687	XS	1
807	687	S	2
808	687	M	3
809	687	L	4
810	687	XL	5
811	687	XXL	6
812	687	XXXL	7
813	687	28	8
814	687	30	9
815	687	32	10
816	687	34	11
817	687	36	12
818	687	38	13
819	687	40	14
820	687	42	15
821	687	Free Size	16
822	689	XS	1
823	689	S	2
824	689	M	3
825	689	L	4
826	689	XL	5
827	689	XXL	6
828	689	XXXL	7
829	689	28	8
830	689	30	9
831	689	32	10
832	689	34	11
833	689	36	12
834	689	38	13
835	689	40	14
836	689	42	15
837	689	Free Size	16
838	692	XS	1
839	692	S	2
840	692	M	3
841	692	L	4
842	692	XL	5
843	692	XXL	6
844	692	XXXL	7
845	692	28	8
846	692	30	9
847	692	32	10
848	692	34	11
849	692	36	12
850	692	38	13
851	692	40	14
852	692	42	15
853	692	Free Size	16
854	695	XS	1
855	695	S	2
856	695	M	3
857	695	L	4
858	695	XL	5
859	695	XXL	6
860	695	XXXL	7
861	695	28	8
862	695	30	9
863	695	32	10
864	695	34	11
865	695	36	12
866	695	38	13
867	695	40	14
868	695	42	15
869	695	Free Size	16
870	698	XS	1
871	698	S	2
872	698	M	3
873	698	L	4
874	698	XL	5
875	698	XXL	6
876	698	XXXL	7
877	698	28	8
878	698	30	9
879	698	32	10
880	698	34	11
881	698	36	12
882	698	38	13
883	698	40	14
884	698	42	15
885	698	Free Size	16
886	703	XS	1
887	703	S	2
888	703	M	3
889	703	L	4
890	703	XL	5
891	703	XXL	6
892	703	XXXL	7
893	703	28	8
894	703	30	9
895	703	32	10
896	703	34	11
897	703	36	12
898	703	38	13
899	703	40	14
900	703	42	15
901	703	Free Size	16
902	708	XS	1
903	708	S	2
904	708	M	3
905	708	L	4
906	708	XL	5
907	708	XXL	6
908	708	XXXL	7
909	708	28	8
910	708	30	9
911	708	32	10
912	708	34	11
913	708	36	12
914	708	38	13
915	708	40	14
916	708	42	15
917	708	Free Size	16
918	713	XS	1
919	713	S	2
920	713	M	3
921	713	L	4
922	713	XL	5
923	713	XXL	6
924	713	XXXL	7
925	713	28	8
926	713	30	9
927	713	32	10
928	713	34	11
929	713	36	12
930	713	38	13
931	713	40	14
932	713	42	15
933	713	Free Size	16
934	718	XS	1
935	718	S	2
936	718	M	3
937	718	L	4
938	718	XL	5
939	718	XXL	6
940	718	XXXL	7
941	718	28	8
942	718	30	9
943	718	32	10
944	718	34	11
945	718	36	12
946	718	38	13
947	718	40	14
948	718	42	15
949	718	Free Size	16
950	724	XS	1
951	724	S	2
952	724	M	3
953	724	L	4
954	724	XL	5
955	724	XXL	6
956	724	XXXL	7
957	724	28	8
958	724	30	9
959	724	32	10
960	724	34	11
961	724	36	12
962	724	38	13
963	724	40	14
964	724	42	15
965	724	Free Size	16
966	728	XS	1
967	728	S	2
968	728	M	3
969	728	L	4
970	728	XL	5
971	728	XXL	6
972	728	XXXL	7
973	728	28	8
974	728	30	9
975	728	32	10
976	728	34	11
977	728	36	12
978	728	38	13
979	728	40	14
980	728	42	15
981	728	Free Size	16
982	732	XS	1
983	732	S	2
984	732	M	3
985	732	L	4
986	732	XL	5
987	732	XXL	6
988	732	XXXL	7
989	732	28	8
990	732	30	9
991	732	32	10
992	732	34	11
993	732	36	12
994	732	38	13
995	732	40	14
996	732	42	15
997	732	Free Size	16
998	736	XS	1
999	736	S	2
1000	736	M	3
1001	736	L	4
1002	736	XL	5
1003	736	XXL	6
1004	736	XXXL	7
1005	736	28	8
1006	736	30	9
1007	736	32	10
1008	736	34	11
1009	736	36	12
1010	736	38	13
1011	736	40	14
1012	736	42	15
1013	736	Free Size	16
1014	741	XS	1
1015	741	S	2
1016	741	M	3
1017	741	L	4
1018	741	XL	5
1019	741	XXL	6
1020	741	XXXL	7
1021	741	28	8
1022	741	30	9
1023	741	32	10
1024	741	34	11
1025	741	36	12
1026	741	38	13
1027	741	40	14
1028	741	42	15
1029	741	Free Size	16
1030	755	XS	1
1031	755	S	2
1032	755	M	3
1033	755	L	4
1034	755	XL	5
1035	755	XXL	6
1036	755	XXXL	7
1037	755	28	8
1038	755	30	9
1039	755	32	10
1040	755	34	11
1041	755	36	12
1042	755	38	13
1043	755	40	14
1044	755	42	15
1045	755	Free Size	16
1046	759	XS	1
1047	759	S	2
1048	759	M	3
1049	759	L	4
1050	759	XL	5
1051	759	XXL	6
1052	759	XXXL	7
1053	759	28	8
1054	759	30	9
1055	759	32	10
1056	759	34	11
1057	759	36	12
1058	759	38	13
1059	759	40	14
1060	759	42	15
1061	759	Free Size	16
1062	763	XS	1
1063	763	S	2
1064	763	M	3
1065	763	L	4
1066	763	XL	5
1067	763	XXL	6
1068	763	XXXL	7
1069	763	28	8
1070	763	30	9
1071	763	32	10
1072	763	34	11
1073	763	36	12
1074	763	38	13
1075	763	40	14
1076	763	42	15
1077	763	Free Size	16
1078	776	XS	1
1079	776	S	2
1080	776	M	3
1081	776	L	4
1082	776	XL	5
1083	776	XXL	6
1084	776	XXXL	7
1085	776	28	8
1086	776	30	9
1087	776	32	10
1088	776	34	11
1089	776	36	12
1090	776	38	13
1091	776	40	14
1092	776	42	15
1093	776	Free Size	16
1094	781	XS	1
1095	781	S	2
1096	781	M	3
1097	781	L	4
1098	781	XL	5
1099	781	XXL	6
1100	781	XXXL	7
1101	781	28	8
1102	781	30	9
1103	781	32	10
1104	781	34	11
1105	781	36	12
1106	781	38	13
1107	781	40	14
1108	781	42	15
1109	781	Free Size	16
1110	786	XS	1
1111	786	S	2
1112	786	M	3
1113	786	L	4
1114	786	XL	5
1115	786	XXL	6
1116	786	XXXL	7
1117	786	28	8
1118	786	30	9
1119	786	32	10
1120	786	34	11
1121	786	36	12
1122	786	38	13
1123	786	40	14
1124	786	42	15
1125	786	Free Size	16
1126	791	XS	1
1127	791	S	2
1128	791	M	3
1129	791	L	4
1130	791	XL	5
1131	791	XXL	6
1132	791	XXXL	7
1133	791	28	8
1134	791	30	9
1135	791	32	10
1136	791	34	11
1137	791	36	12
1138	791	38	13
1139	791	40	14
1140	791	42	15
1141	791	Free Size	16
1142	815	XS	1
1143	815	S	2
1144	815	M	3
1145	815	L	4
1146	815	XL	5
1147	815	XXL	6
1148	815	XXXL	7
1149	815	28	8
1150	815	30	9
1151	815	32	10
1152	815	34	11
1153	815	36	12
1154	815	38	13
1155	815	40	14
1156	815	42	15
1157	815	Free Size	16
1158	867	XS	1
1159	867	S	2
1160	867	M	3
1161	867	L	4
1162	867	XL	5
1163	867	XXL	6
1164	867	XXXL	7
1165	867	28	8
1166	867	30	9
1167	867	32	10
1168	867	34	11
1169	867	36	12
1170	867	38	13
1171	867	40	14
1172	867	42	15
1173	867	Free Size	16
1174	872	XS	1
1175	872	S	2
1176	872	M	3
1177	872	L	4
1178	872	XL	5
1179	872	XXL	6
1180	872	XXXL	7
1181	872	28	8
1182	872	30	9
1183	872	32	10
1184	872	34	11
1185	872	36	12
1186	872	38	13
1187	872	40	14
1188	872	42	15
1189	872	Free Size	16
1190	903	XS	1
1191	903	S	2
1192	903	M	3
1193	903	L	4
1194	903	XL	5
1195	903	XXL	6
1196	903	XXXL	7
1197	903	28	8
1198	903	30	9
1199	903	32	10
1200	903	34	11
1201	903	36	12
1202	903	38	13
1203	903	40	14
1204	903	42	15
1205	903	Free Size	16
1206	926	XS	1
1207	926	S	2
1208	926	M	3
1209	926	L	4
1210	926	XL	5
1211	926	XXL	6
1212	926	XXXL	7
1213	926	28	8
1214	926	30	9
1215	926	32	10
1216	926	34	11
1217	926	36	12
1218	926	38	13
1219	926	40	14
1220	926	42	15
1221	926	Free Size	16
1222	962	XS	1
1223	962	S	2
1224	962	M	3
1225	962	L	4
1226	962	XL	5
1227	962	XXL	6
1228	962	XXXL	7
1229	962	28	8
1230	962	30	9
1231	962	32	10
1232	962	34	11
1233	962	36	12
1234	962	38	13
1235	962	40	14
1236	962	42	15
1237	962	Free Size	16
1238	976	XS	1
1239	976	S	2
1240	976	M	3
1241	976	L	4
1242	976	XL	5
1243	976	XXL	6
1244	976	XXXL	7
1245	976	28	8
1246	976	30	9
1247	976	32	10
1248	976	34	11
1249	976	36	12
1250	976	38	13
1251	976	40	14
1252	976	42	15
1253	976	Free Size	16
1254	991	XS	1
1255	991	S	2
1256	991	M	3
1257	991	L	4
1258	991	XL	5
1259	991	XXL	6
1260	991	XXXL	7
1261	991	28	8
1262	991	30	9
1263	991	32	10
1264	991	34	11
1265	991	36	12
1266	991	38	13
1267	991	40	14
1268	991	42	15
1269	991	Free Size	16
1270	996	XS	1
1271	996	S	2
1272	996	M	3
1273	996	L	4
1274	996	XL	5
1275	996	XXL	6
1276	996	XXXL	7
1277	996	28	8
1278	996	30	9
1279	996	32	10
1280	996	34	11
1281	996	36	12
1282	996	38	13
1283	996	40	14
1284	996	42	15
1285	996	Free Size	16
1286	1001	XS	1
1287	1001	S	2
1288	1001	M	3
1289	1001	L	4
1290	1001	XL	5
1291	1001	XXL	6
1292	1001	XXXL	7
1293	1001	28	8
1294	1001	30	9
1295	1001	32	10
1296	1001	34	11
1297	1001	36	12
1298	1001	38	13
1299	1001	40	14
1300	1001	42	15
1301	1001	Free Size	16
1302	1039	XS	1
1303	1039	S	2
1304	1039	M	3
1305	1039	L	4
1306	1039	XL	5
1307	1039	XXL	6
1308	1039	XXXL	7
1309	1039	28	8
1310	1039	30	9
1311	1039	32	10
1312	1039	34	11
1313	1039	36	12
1314	1039	38	13
1315	1039	40	14
1316	1039	42	15
1317	1039	Free Size	16
1318	1044	XS	1
1319	1044	S	2
1320	1044	M	3
1321	1044	L	4
1322	1044	XL	5
1323	1044	XXL	6
1324	1044	XXXL	7
1325	1044	28	8
1326	1044	30	9
1327	1044	32	10
1328	1044	34	11
1329	1044	36	12
1330	1044	38	13
1331	1044	40	14
1332	1044	42	15
1333	1044	Free Size	16
1334	1095	XS	1
1335	1095	S	2
1336	1095	M	3
1337	1095	L	4
1338	1095	XL	5
1339	1095	XXL	6
1340	1095	XXXL	7
1341	1095	28	8
1342	1095	30	9
1343	1095	32	10
1344	1095	34	11
1345	1095	36	12
1346	1095	38	13
1347	1095	40	14
1348	1095	42	15
1349	1095	Free Size	16
1350	1102	XS	1
1351	1102	S	2
1352	1102	M	3
1353	1102	L	4
1354	1102	XL	5
1355	1102	XXL	6
1356	1102	XXXL	7
1357	1102	28	8
1358	1102	30	9
1359	1102	32	10
1360	1102	34	11
1361	1102	36	12
1362	1102	38	13
1363	1102	40	14
1364	1102	42	15
1365	1102	Free Size	16
1366	1506	XS	1
1367	1506	S	2
1368	1506	M	3
1369	1506	L	4
1370	1506	XL	5
1371	1506	XXL	6
1372	1506	XXXL	7
1373	1506	28	8
1374	1506	30	9
1375	1506	32	10
1376	1506	34	11
1377	1506	36	12
1378	1506	38	13
1379	1506	40	14
1380	1506	42	15
1381	1506	Free Size	16
1382	1800	XS	1
1383	1800	S	2
1384	1800	M	3
1385	1800	L	4
1386	1800	XL	5
1387	1800	XXL	6
1388	1800	XXXL	7
1389	1800	28	8
1390	1800	30	9
1391	1800	32	10
1392	1800	34	11
1393	1800	36	12
1394	1800	38	13
1395	1800	40	14
1396	1800	42	15
1397	1800	Free Size	16
1398	1920	XS	1
1399	1920	S	2
1400	1920	M	3
1401	1920	L	4
1402	1920	XL	5
1403	1920	XXL	6
1404	1920	XXXL	7
1405	1920	28	8
1406	1920	30	9
1407	1920	32	10
1408	1920	34	11
1409	1920	36	12
1410	1920	38	13
1411	1920	40	14
1412	1920	42	15
1413	1920	Free Size	16
1414	1960	XS	1
1415	1960	S	2
1416	1960	M	3
1417	1960	L	4
1418	1960	XL	5
1419	1960	XXL	6
1420	1960	XXXL	7
1421	1960	28	8
1422	1960	30	9
1423	1960	32	10
1424	1960	34	11
1425	1960	36	12
1426	1960	38	13
1427	1960	40	14
1428	1960	42	15
1429	1960	Free Size	16
1430	1968	XS	1
1431	1968	S	2
1432	1968	M	3
1433	1968	L	4
1434	1968	XL	5
1435	1968	XXL	6
1436	1968	XXXL	7
1437	1968	28	8
1438	1968	30	9
1439	1968	32	10
1440	1968	34	11
1441	1968	36	12
1442	1968	38	13
1443	1968	40	14
1444	1968	42	15
1445	1968	Free Size	16
1446	1972	XS	1
1447	1972	S	2
1448	1972	M	3
1449	1972	L	4
1450	1972	XL	5
1451	1972	XXL	6
1452	1972	XXXL	7
1453	1972	28	8
1454	1972	30	9
1455	1972	32	10
1456	1972	34	11
1457	1972	36	12
1458	1972	38	13
1459	1972	40	14
1460	1972	42	15
1461	1972	Free Size	16
1462	1976	XS	1
1463	1976	S	2
1464	1976	M	3
1465	1976	L	4
1466	1976	XL	5
1467	1976	XXL	6
1468	1976	XXXL	7
1469	1976	28	8
1470	1976	30	9
1471	1976	32	10
1472	1976	34	11
1473	1976	36	12
1474	1976	38	13
1475	1976	40	14
1476	1976	42	15
1477	1976	Free Size	16
1478	1980	XS	1
1479	1980	S	2
1480	1980	M	3
1481	1980	L	4
1482	1980	XL	5
1483	1980	XXL	6
1484	1980	XXXL	7
1485	1980	28	8
1486	1980	30	9
1487	1980	32	10
1488	1980	34	11
1489	1980	36	12
1490	1980	38	13
1491	1980	40	14
1492	1980	42	15
1493	1980	Free Size	16
1494	1989	XS	1
1495	1989	S	2
1496	1989	M	3
1497	1989	L	4
1498	1989	XL	5
1499	1989	XXL	6
1500	1989	XXXL	7
1501	1989	28	8
1502	1989	30	9
1503	1989	32	10
1504	1989	34	11
1505	1989	36	12
1506	1989	38	13
1507	1989	40	14
1508	1989	42	15
1509	1989	Free Size	16
1510	1993	XS	1
1511	1993	S	2
1512	1993	M	3
1513	1993	L	4
1514	1993	XL	5
1515	1993	XXL	6
1516	1993	XXXL	7
1517	1993	28	8
1518	1993	30	9
1519	1993	32	10
1520	1993	34	11
1521	1993	36	12
1522	1993	38	13
1523	1993	40	14
1524	1993	42	15
1525	1993	Free Size	16
1526	2005	XS	1
1527	2005	S	2
1528	2005	M	3
1529	2005	L	4
1530	2005	XL	5
1531	2005	XXL	6
1532	2005	XXXL	7
1533	2005	28	8
1534	2005	30	9
1535	2005	32	10
1536	2005	34	11
1537	2005	36	12
1538	2005	38	13
1539	2005	40	14
1540	2005	42	15
1541	2005	Free Size	16
1542	2008	XS	1
1543	2008	S	2
1544	2008	M	3
1545	2008	L	4
1546	2008	XL	5
1547	2008	XXL	6
1548	2008	XXXL	7
1549	2008	28	8
1550	2008	30	9
1551	2008	32	10
1552	2008	34	11
1553	2008	36	12
1554	2008	38	13
1555	2008	40	14
1556	2008	42	15
1557	2008	Free Size	16
1558	2024	XS	1
1559	2024	S	2
1560	2024	M	3
1561	2024	L	4
1562	2024	XL	5
1563	2024	XXL	6
1564	2024	XXXL	7
1565	2024	28	8
1566	2024	30	9
1567	2024	32	10
1568	2024	34	11
1569	2024	36	12
1570	2024	38	13
1571	2024	40	14
1572	2024	42	15
1573	2024	Free Size	16
1574	2028	XS	1
1575	2028	S	2
1576	2028	M	3
1577	2028	L	4
1578	2028	XL	5
1579	2028	XXL	6
1580	2028	XXXL	7
1581	2028	28	8
1582	2028	30	9
1583	2028	32	10
1584	2028	34	11
1585	2028	36	12
1586	2028	38	13
1587	2028	40	14
1588	2028	42	15
1589	2028	Free Size	16
1590	2045	XS	1
1591	2045	S	2
1592	2045	M	3
1593	2045	L	4
1594	2045	XL	5
1595	2045	XXL	6
1596	2045	XXXL	7
1597	2045	28	8
1598	2045	30	9
1599	2045	32	10
1600	2045	34	11
1601	2045	36	12
1602	2045	38	13
1603	2045	40	14
1604	2045	42	15
1605	2045	Free Size	16
1606	2057	XS	1
1607	2057	S	2
1608	2057	M	3
1609	2057	L	4
1610	2057	XL	5
1611	2057	XXL	6
1612	2057	XXXL	7
1613	2057	28	8
1614	2057	30	9
1615	2057	32	10
1616	2057	34	11
1617	2057	36	12
1618	2057	38	13
1619	2057	40	14
1620	2057	42	15
1621	2057	Free Size	16
1622	2061	XS	1
1623	2061	S	2
1624	2061	M	3
1625	2061	L	4
1626	2061	XL	5
1627	2061	XXL	6
1628	2061	XXXL	7
1629	2061	28	8
1630	2061	30	9
1631	2061	32	10
1632	2061	34	11
1633	2061	36	12
1634	2061	38	13
1635	2061	40	14
1636	2061	42	15
1637	2061	Free Size	16
1638	2069	XS	1
1639	2069	S	2
1640	2069	M	3
1641	2069	L	4
1642	2069	XL	5
1643	2069	XXL	6
1644	2069	XXXL	7
1645	2069	28	8
1646	2069	30	9
1647	2069	32	10
1648	2069	34	11
1649	2069	36	12
1650	2069	38	13
1651	2069	40	14
1652	2069	42	15
1653	2069	Free Size	16
1654	2075	XS	1
1655	2075	S	2
1656	2075	M	3
1657	2075	L	4
1658	2075	XL	5
1659	2075	XXL	6
1660	2075	XXXL	7
1661	2075	28	8
1662	2075	30	9
1663	2075	32	10
1664	2075	34	11
1665	2075	36	12
1666	2075	38	13
1667	2075	40	14
1668	2075	42	15
1669	2075	Free Size	16
1670	2081	XS	1
1671	2081	S	2
1672	2081	M	3
1673	2081	L	4
1674	2081	XL	5
1675	2081	XXL	6
1676	2081	XXXL	7
1677	2081	28	8
1678	2081	30	9
1679	2081	32	10
1680	2081	34	11
1681	2081	36	12
1682	2081	38	13
1683	2081	40	14
1684	2081	42	15
1685	2081	Free Size	16
1686	2087	XS	1
1687	2087	S	2
1688	2087	M	3
1689	2087	L	4
1690	2087	XL	5
1691	2087	XXL	6
1692	2087	XXXL	7
1693	2087	28	8
1694	2087	30	9
1695	2087	32	10
1696	2087	34	11
1697	2087	36	12
1698	2087	38	13
1699	2087	40	14
1700	2087	42	15
1701	2087	Free Size	16
1702	2093	XS	1
1703	2093	S	2
1704	2093	M	3
1705	2093	L	4
1706	2093	XL	5
1707	2093	XXL	6
1708	2093	XXXL	7
1709	2093	28	8
1710	2093	30	9
1711	2093	32	10
1712	2093	34	11
1713	2093	36	12
1714	2093	38	13
1715	2093	40	14
1716	2093	42	15
1717	2093	Free Size	16
1718	2099	XS	1
1719	2099	S	2
1720	2099	M	3
1721	2099	L	4
1722	2099	XL	5
1723	2099	XXL	6
1724	2099	XXXL	7
1725	2099	28	8
1726	2099	30	9
1727	2099	32	10
1728	2099	34	11
1729	2099	36	12
1730	2099	38	13
1731	2099	40	14
1732	2099	42	15
1733	2099	Free Size	16
1734	2105	XS	1
1735	2105	S	2
1736	2105	M	3
1737	2105	L	4
1738	2105	XL	5
1739	2105	XXL	6
1740	2105	XXXL	7
1741	2105	28	8
1742	2105	30	9
1743	2105	32	10
1744	2105	34	11
1745	2105	36	12
1746	2105	38	13
1747	2105	40	14
1748	2105	42	15
1749	2105	Free Size	16
1750	2111	XS	1
1751	2111	S	2
1752	2111	M	3
1753	2111	L	4
1754	2111	XL	5
1755	2111	XXL	6
1756	2111	XXXL	7
1757	2111	28	8
1758	2111	30	9
1759	2111	32	10
1760	2111	34	11
1761	2111	36	12
1762	2111	38	13
1763	2111	40	14
1764	2111	42	15
1765	2111	Free Size	16
1766	2117	XS	1
1767	2117	S	2
1768	2117	M	3
1769	2117	L	4
1770	2117	XL	5
1771	2117	XXL	6
1772	2117	XXXL	7
1773	2117	28	8
1774	2117	30	9
1775	2117	32	10
1776	2117	34	11
1777	2117	36	12
1778	2117	38	13
1779	2117	40	14
1780	2117	42	15
1781	2117	Free Size	16
1782	2123	XS	1
1783	2123	S	2
1784	2123	M	3
1785	2123	L	4
1786	2123	XL	5
1787	2123	XXL	6
1788	2123	XXXL	7
1789	2123	28	8
1790	2123	30	9
1791	2123	32	10
1792	2123	34	11
1793	2123	36	12
1794	2123	38	13
1795	2123	40	14
1796	2123	42	15
1797	2123	Free Size	16
1798	2129	XS	1
1799	2129	S	2
1800	2129	M	3
1801	2129	L	4
1802	2129	XL	5
1803	2129	XXL	6
1804	2129	XXXL	7
1805	2129	28	8
1806	2129	30	9
1807	2129	32	10
1808	2129	34	11
1809	2129	36	12
1810	2129	38	13
1811	2129	40	14
1812	2129	42	15
1813	2129	Free Size	16
1814	2135	XS	1
1815	2135	S	2
1816	2135	M	3
1817	2135	L	4
1818	2135	XL	5
1819	2135	XXL	6
1820	2135	XXXL	7
1821	2135	28	8
1822	2135	30	9
1823	2135	32	10
1824	2135	34	11
1825	2135	36	12
1826	2135	38	13
1827	2135	40	14
1828	2135	42	15
1829	2135	Free Size	16
1830	2142	XS	1
1831	2142	S	2
1832	2142	M	3
1833	2142	L	4
1834	2142	XL	5
1835	2142	XXL	6
1836	2142	XXXL	7
1837	2142	28	8
1838	2142	30	9
1839	2142	32	10
1840	2142	34	11
1841	2142	36	12
1842	2142	38	13
1843	2142	40	14
1844	2142	42	15
1845	2142	Free Size	16
1846	2146	XS	1
1847	2146	S	2
1848	2146	M	3
1849	2146	L	4
1850	2146	XL	5
1851	2146	XXL	6
1852	2146	XXXL	7
1853	2146	28	8
1854	2146	30	9
1855	2146	32	10
1856	2146	34	11
1857	2146	36	12
1858	2146	38	13
1859	2146	40	14
1860	2146	42	15
1861	2146	Free Size	16
1862	2163	XS	1
1863	2163	S	2
1864	2163	M	3
1865	2163	L	4
1866	2163	XL	5
1867	2163	XXL	6
1868	2163	XXXL	7
1869	2163	28	8
1870	2163	30	9
1871	2163	32	10
1872	2163	34	11
1873	2163	36	12
1874	2163	38	13
1875	2163	40	14
1876	2163	42	15
1877	2163	Free Size	16
1878	2165	XS	1
1879	2165	S	2
1880	2165	M	3
1881	2165	L	4
1882	2165	XL	5
1883	2165	XXL	6
1884	2165	XXXL	7
1885	2165	28	8
1886	2165	30	9
1887	2165	32	10
1888	2165	34	11
1889	2165	36	12
1890	2165	38	13
1891	2165	40	14
1892	2165	42	15
1893	2165	Free Size	16
1894	2171	XS	1
1895	2171	S	2
1896	2171	M	3
1897	2171	L	4
1898	2171	XL	5
1899	2171	XXL	6
1900	2171	XXXL	7
1901	2171	28	8
1902	2171	30	9
1903	2171	32	10
1904	2171	34	11
1905	2171	36	12
1906	2171	38	13
1907	2171	40	14
1908	2171	42	15
1909	2171	Free Size	16
1910	2177	XS	1
1911	2177	S	2
1912	2177	M	3
1913	2177	L	4
1914	2177	XL	5
1915	2177	XXL	6
1916	2177	XXXL	7
1917	2177	28	8
1918	2177	30	9
1919	2177	32	10
1920	2177	34	11
1921	2177	36	12
1922	2177	38	13
1923	2177	40	14
1924	2177	42	15
1925	2177	Free Size	16
1926	2183	XS	1
1927	2183	S	2
1928	2183	M	3
1929	2183	L	4
1930	2183	XL	5
1931	2183	XXL	6
1932	2183	XXXL	7
1933	2183	28	8
1934	2183	30	9
1935	2183	32	10
1936	2183	34	11
1937	2183	36	12
1938	2183	38	13
1939	2183	40	14
1940	2183	42	15
1941	2183	Free Size	16
1942	2189	XS	1
1943	2189	S	2
1944	2189	M	3
1945	2189	L	4
1946	2189	XL	5
1947	2189	XXL	6
1948	2189	XXXL	7
1949	2189	28	8
1950	2189	30	9
1951	2189	32	10
1952	2189	34	11
1953	2189	36	12
1954	2189	38	13
1955	2189	40	14
1956	2189	42	15
1957	2189	Free Size	16
1958	2195	XS	1
1959	2195	S	2
1960	2195	M	3
1961	2195	L	4
1962	2195	XL	5
1963	2195	XXL	6
1964	2195	XXXL	7
1965	2195	28	8
1966	2195	30	9
1967	2195	32	10
1968	2195	34	11
1969	2195	36	12
1970	2195	38	13
1971	2195	40	14
1972	2195	42	15
1973	2195	Free Size	16
1974	2201	XS	1
1975	2201	S	2
1976	2201	M	3
1977	2201	L	4
1978	2201	XL	5
1979	2201	XXL	6
1980	2201	XXXL	7
1981	2201	28	8
1982	2201	30	9
1983	2201	32	10
1984	2201	34	11
1985	2201	36	12
1986	2201	38	13
1987	2201	40	14
1988	2201	42	15
1989	2201	Free Size	16
1990	2207	XS	1
1991	2207	S	2
1992	2207	M	3
1993	2207	L	4
1994	2207	XL	5
1995	2207	XXL	6
1996	2207	XXXL	7
1997	2207	28	8
1998	2207	30	9
1999	2207	32	10
2000	2207	34	11
2001	2207	36	12
2002	2207	38	13
2003	2207	40	14
2004	2207	42	15
2005	2207	Free Size	16
2006	2213	XS	1
2007	2213	S	2
2008	2213	M	3
2009	2213	L	4
2010	2213	XL	5
2011	2213	XXL	6
2012	2213	XXXL	7
2013	2213	28	8
2014	2213	30	9
2015	2213	32	10
2016	2213	34	11
2017	2213	36	12
2018	2213	38	13
2019	2213	40	14
2020	2213	42	15
2021	2213	Free Size	16
2022	2219	XS	1
2023	2219	S	2
2024	2219	M	3
2025	2219	L	4
2026	2219	XL	5
2027	2219	XXL	6
2028	2219	XXXL	7
2029	2219	28	8
2030	2219	30	9
2031	2219	32	10
2032	2219	34	11
2033	2219	36	12
2034	2219	38	13
2035	2219	40	14
2036	2219	42	15
2037	2219	Free Size	16
2038	2225	XS	1
2039	2225	S	2
2040	2225	M	3
2041	2225	L	4
2042	2225	XL	5
2043	2225	XXL	6
2044	2225	XXXL	7
2045	2225	28	8
2046	2225	30	9
2047	2225	32	10
2048	2225	34	11
2049	2225	36	12
2050	2225	38	13
2051	2225	40	14
2052	2225	42	15
2053	2225	Free Size	16
2054	2231	XS	1
2055	2231	S	2
2056	2231	M	3
2057	2231	L	4
2058	2231	XL	5
2059	2231	XXL	6
2060	2231	XXXL	7
2061	2231	28	8
2062	2231	30	9
2063	2231	32	10
2064	2231	34	11
2065	2231	36	12
2066	2231	38	13
2067	2231	40	14
2068	2231	42	15
2069	2231	Free Size	16
2070	2236	XS	1
2071	2236	S	2
2072	2236	M	3
2073	2236	L	4
2074	2236	XL	5
2075	2236	XXL	6
2076	2236	XXXL	7
2077	2236	28	8
2078	2236	30	9
2079	2236	32	10
2080	2236	34	11
2081	2236	36	12
2082	2236	38	13
2083	2236	40	14
2084	2236	42	15
2085	2236	Free Size	16
2086	2241	XS	1
2087	2241	S	2
2088	2241	M	3
2089	2241	L	4
2090	2241	XL	5
2091	2241	XXL	6
2092	2241	XXXL	7
2093	2241	28	8
2094	2241	30	9
2095	2241	32	10
2096	2241	34	11
2097	2241	36	12
2098	2241	38	13
2099	2241	40	14
2100	2241	42	15
2101	2241	Free Size	16
2102	2246	XS	1
2103	2246	S	2
2104	2246	M	3
2105	2246	L	4
2106	2246	XL	5
2107	2246	XXL	6
2108	2246	XXXL	7
2109	2246	28	8
2110	2246	30	9
2111	2246	32	10
2112	2246	34	11
2113	2246	36	12
2114	2246	38	13
2115	2246	40	14
2116	2246	42	15
2117	2246	Free Size	16
2118	2251	XS	1
2119	2251	S	2
2120	2251	M	3
2121	2251	L	4
2122	2251	XL	5
2123	2251	XXL	6
2124	2251	XXXL	7
2125	2251	28	8
2126	2251	30	9
2127	2251	32	10
2128	2251	34	11
2129	2251	36	12
2130	2251	38	13
2131	2251	40	14
2132	2251	42	15
2133	2251	Free Size	16
2134	2256	XS	1
2135	2256	S	2
2136	2256	M	3
2137	2256	L	4
2138	2256	XL	5
2139	2256	XXL	6
2140	2256	XXXL	7
2141	2256	28	8
2142	2256	30	9
2143	2256	32	10
2144	2256	34	11
2145	2256	36	12
2146	2256	38	13
2147	2256	40	14
2148	2256	42	15
2149	2256	Free Size	16
2150	2262	XS	1
2151	2262	S	2
2152	2262	M	3
2153	2262	L	4
2154	2262	XL	5
2155	2262	XXL	6
2156	2262	XXXL	7
2157	2262	28	8
2158	2262	30	9
2159	2262	32	10
2160	2262	34	11
2161	2262	36	12
2162	2262	38	13
2163	2262	40	14
2164	2262	42	15
2165	2262	Free Size	16
2166	2268	XS	1
2167	2268	S	2
2168	2268	M	3
2169	2268	L	4
2170	2268	XL	5
2171	2268	XXL	6
2172	2268	XXXL	7
2173	2268	28	8
2174	2268	30	9
2175	2268	32	10
2176	2268	34	11
2177	2268	36	12
2178	2268	38	13
2179	2268	40	14
2180	2268	42	15
2181	2268	Free Size	16
2182	2274	XS	1
2183	2274	S	2
2184	2274	M	3
2185	2274	L	4
2186	2274	XL	5
2187	2274	XXL	6
2188	2274	XXXL	7
2189	2274	28	8
2190	2274	30	9
2191	2274	32	10
2192	2274	34	11
2193	2274	36	12
2194	2274	38	13
2195	2274	40	14
2196	2274	42	15
2197	2274	Free Size	16
2198	2280	XS	1
2199	2280	S	2
2200	2280	M	3
2201	2280	L	4
2202	2280	XL	5
2203	2280	XXL	6
2204	2280	XXXL	7
2205	2280	28	8
2206	2280	30	9
2207	2280	32	10
2208	2280	34	11
2209	2280	36	12
2210	2280	38	13
2211	2280	40	14
2212	2280	42	15
2213	2280	Free Size	16
2214	2286	XS	1
2215	2286	S	2
2216	2286	M	3
2217	2286	L	4
2218	2286	XL	5
2219	2286	XXL	6
2220	2286	XXXL	7
2221	2286	28	8
2222	2286	30	9
2223	2286	32	10
2224	2286	34	11
2225	2286	36	12
2226	2286	38	13
2227	2286	40	14
2228	2286	42	15
2229	2286	Free Size	16
2230	2291	XS	1
2231	2291	S	2
2232	2291	M	3
2233	2291	L	4
2234	2291	XL	5
2235	2291	XXL	6
2236	2291	XXXL	7
2237	2291	28	8
2238	2291	30	9
2239	2291	32	10
2240	2291	34	11
2241	2291	36	12
2242	2291	38	13
2243	2291	40	14
2244	2291	42	15
2245	2291	Free Size	16
2246	2296	XS	1
2247	2296	S	2
2248	2296	M	3
2249	2296	L	4
2250	2296	XL	5
2251	2296	XXL	6
2252	2296	XXXL	7
2253	2296	28	8
2254	2296	30	9
2255	2296	32	10
2256	2296	34	11
2257	2296	36	12
2258	2296	38	13
2259	2296	40	14
2260	2296	42	15
2261	2296	Free Size	16
2262	2301	XS	1
2263	2301	S	2
2264	2301	M	3
2265	2301	L	4
2266	2301	XL	5
2267	2301	XXL	6
2268	2301	XXXL	7
2269	2301	28	8
2270	2301	30	9
2271	2301	32	10
2272	2301	34	11
2273	2301	36	12
2274	2301	38	13
2275	2301	40	14
2276	2301	42	15
2277	2301	Free Size	16
2278	2307	XS	1
2279	2307	S	2
2280	2307	M	3
2281	2307	L	4
2282	2307	XL	5
2283	2307	XXL	6
2284	2307	XXXL	7
2285	2307	28	8
2286	2307	30	9
2287	2307	32	10
2288	2307	34	11
2289	2307	36	12
2290	2307	38	13
2291	2307	40	14
2292	2307	42	15
2293	2307	Free Size	16
2294	2313	XS	1
2295	2313	S	2
2296	2313	M	3
2297	2313	L	4
2298	2313	XL	5
2299	2313	XXL	6
2300	2313	XXXL	7
2301	2313	28	8
2302	2313	30	9
2303	2313	32	10
2304	2313	34	11
2305	2313	36	12
2306	2313	38	13
2307	2313	40	14
2308	2313	42	15
2309	2313	Free Size	16
2310	2319	XS	1
2311	2319	S	2
2312	2319	M	3
2313	2319	L	4
2314	2319	XL	5
2315	2319	XXL	6
2316	2319	XXXL	7
2317	2319	28	8
2318	2319	30	9
2319	2319	32	10
2320	2319	34	11
2321	2319	36	12
2322	2319	38	13
2323	2319	40	14
2324	2319	42	15
2325	2319	Free Size	16
2326	2324	XS	1
2327	2324	S	2
2328	2324	M	3
2329	2324	L	4
2330	2324	XL	5
2331	2324	XXL	6
2332	2324	XXXL	7
2333	2324	28	8
2334	2324	30	9
2335	2324	32	10
2336	2324	34	11
2337	2324	36	12
2338	2324	38	13
2339	2324	40	14
2340	2324	42	15
2341	2324	Free Size	16
2342	2329	XS	1
2343	2329	S	2
2344	2329	M	3
2345	2329	L	4
2346	2329	XL	5
2347	2329	XXL	6
2348	2329	XXXL	7
2349	2329	28	8
2350	2329	30	9
2351	2329	32	10
2352	2329	34	11
2353	2329	36	12
2354	2329	38	13
2355	2329	40	14
2356	2329	42	15
2357	2329	Free Size	16
2358	2334	XS	1
2359	2334	S	2
2360	2334	M	3
2361	2334	L	4
2362	2334	XL	5
2363	2334	XXL	6
2364	2334	XXXL	7
2365	2334	28	8
2366	2334	30	9
2367	2334	32	10
2368	2334	34	11
2369	2334	36	12
2370	2334	38	13
2371	2334	40	14
2372	2334	42	15
2373	2334	Free Size	16
2374	2339	XS	1
2375	2339	S	2
2376	2339	M	3
2377	2339	L	4
2378	2339	XL	5
2379	2339	XXL	6
2380	2339	XXXL	7
2381	2339	28	8
2382	2339	30	9
2383	2339	32	10
2384	2339	34	11
2385	2339	36	12
2386	2339	38	13
2387	2339	40	14
2388	2339	42	15
2389	2339	Free Size	16
2390	2345	XS	1
2391	2345	S	2
2392	2345	M	3
2393	2345	L	4
2394	2345	XL	5
2395	2345	XXL	6
2396	2345	XXXL	7
2397	2345	28	8
2398	2345	30	9
2399	2345	32	10
2400	2345	34	11
2401	2345	36	12
2402	2345	38	13
2403	2345	40	14
2404	2345	42	15
2405	2345	Free Size	16
2406	2351	XS	1
2407	2351	S	2
2408	2351	M	3
2409	2351	L	4
2410	2351	XL	5
2411	2351	XXL	6
2412	2351	XXXL	7
2413	2351	28	8
2414	2351	30	9
2415	2351	32	10
2416	2351	34	11
2417	2351	36	12
2418	2351	38	13
2419	2351	40	14
2420	2351	42	15
2421	2351	Free Size	16
2422	2357	XS	1
2423	2357	S	2
2424	2357	M	3
2425	2357	L	4
2426	2357	XL	5
2427	2357	XXL	6
2428	2357	XXXL	7
2429	2357	28	8
2430	2357	30	9
2431	2357	32	10
2432	2357	34	11
2433	2357	36	12
2434	2357	38	13
2435	2357	40	14
2436	2357	42	15
2437	2357	Free Size	16
2438	2363	XS	1
2439	2363	S	2
2440	2363	M	3
2441	2363	L	4
2442	2363	XL	5
2443	2363	XXL	6
2444	2363	XXXL	7
2445	2363	28	8
2446	2363	30	9
2447	2363	32	10
2448	2363	34	11
2449	2363	36	12
2450	2363	38	13
2451	2363	40	14
2452	2363	42	15
2453	2363	Free Size	16
2454	2369	XS	1
2455	2369	S	2
2456	2369	M	3
2457	2369	L	4
2458	2369	XL	5
2459	2369	XXL	6
2460	2369	XXXL	7
2461	2369	28	8
2462	2369	30	9
2463	2369	32	10
2464	2369	34	11
2465	2369	36	12
2466	2369	38	13
2467	2369	40	14
2468	2369	42	15
2469	2369	Free Size	16
2470	2375	XS	1
2471	2375	S	2
2472	2375	M	3
2473	2375	L	4
2474	2375	XL	5
2475	2375	XXL	6
2476	2375	XXXL	7
2477	2375	28	8
2478	2375	30	9
2479	2375	32	10
2480	2375	34	11
2481	2375	36	12
2482	2375	38	13
2483	2375	40	14
2484	2375	42	15
2485	2375	Free Size	16
2486	2381	XS	1
2487	2381	S	2
2488	2381	M	3
2489	2381	L	4
2490	2381	XL	5
2491	2381	XXL	6
2492	2381	XXXL	7
2493	2381	28	8
2494	2381	30	9
2495	2381	32	10
2496	2381	34	11
2497	2381	36	12
2498	2381	38	13
2499	2381	40	14
2500	2381	42	15
2501	2381	Free Size	16
2502	2387	XS	1
2503	2387	S	2
2504	2387	M	3
2505	2387	L	4
2506	2387	XL	5
2507	2387	XXL	6
2508	2387	XXXL	7
2509	2387	28	8
2510	2387	30	9
2511	2387	32	10
2512	2387	34	11
2513	2387	36	12
2514	2387	38	13
2515	2387	40	14
2516	2387	42	15
2517	2387	Free Size	16
2518	2393	XS	1
2519	2393	S	2
2520	2393	M	3
2521	2393	L	4
2522	2393	XL	5
2523	2393	XXL	6
2524	2393	XXXL	7
2525	2393	28	8
2526	2393	30	9
2527	2393	32	10
2528	2393	34	11
2529	2393	36	12
2530	2393	38	13
2531	2393	40	14
2532	2393	42	15
2533	2393	Free Size	16
2534	2399	XS	1
2535	2399	S	2
2536	2399	M	3
2537	2399	L	4
2538	2399	XL	5
2539	2399	XXL	6
2540	2399	XXXL	7
2541	2399	28	8
2542	2399	30	9
2543	2399	32	10
2544	2399	34	11
2545	2399	36	12
2546	2399	38	13
2547	2399	40	14
2548	2399	42	15
2549	2399	Free Size	16
2550	2405	XS	1
2551	2405	S	2
2552	2405	M	3
2553	2405	L	4
2554	2405	XL	5
2555	2405	XXL	6
2556	2405	XXXL	7
2557	2405	28	8
2558	2405	30	9
2559	2405	32	10
2560	2405	34	11
2561	2405	36	12
2562	2405	38	13
2563	2405	40	14
2564	2405	42	15
2565	2405	Free Size	16
2566	2411	XS	1
2567	2411	S	2
2568	2411	M	3
2569	2411	L	4
2570	2411	XL	5
2571	2411	XXL	6
2572	2411	XXXL	7
2573	2411	28	8
2574	2411	30	9
2575	2411	32	10
2576	2411	34	11
2577	2411	36	12
2578	2411	38	13
2579	2411	40	14
2580	2411	42	15
2581	2411	Free Size	16
2582	2416	XS	1
2583	2416	S	2
2584	2416	M	3
2585	2416	L	4
2586	2416	XL	5
2587	2416	XXL	6
2588	2416	XXXL	7
2589	2416	28	8
2590	2416	30	9
2591	2416	32	10
2592	2416	34	11
2593	2416	36	12
2594	2416	38	13
2595	2416	40	14
2596	2416	42	15
2597	2416	Free Size	16
2598	2421	XS	1
2599	2421	S	2
2600	2421	M	3
2601	2421	L	4
2602	2421	XL	5
2603	2421	XXL	6
2604	2421	XXXL	7
2605	2421	28	8
2606	2421	30	9
2607	2421	32	10
2608	2421	34	11
2609	2421	36	12
2610	2421	38	13
2611	2421	40	14
2612	2421	42	15
2613	2421	Free Size	16
2614	2426	XS	1
2615	2426	S	2
2616	2426	M	3
2617	2426	L	4
2618	2426	XL	5
2619	2426	XXL	6
2620	2426	XXXL	7
2621	2426	28	8
2622	2426	30	9
2623	2426	32	10
2624	2426	34	11
2625	2426	36	12
2626	2426	38	13
2627	2426	40	14
2628	2426	42	15
2629	2426	Free Size	16
2630	2432	XS	1
2631	2432	S	2
2632	2432	M	3
2633	2432	L	4
2634	2432	XL	5
2635	2432	XXL	6
2636	2432	XXXL	7
2637	2432	28	8
2638	2432	30	9
2639	2432	32	10
2640	2432	34	11
2641	2432	36	12
2642	2432	38	13
2643	2432	40	14
2644	2432	42	15
2645	2432	Free Size	16
2646	2438	XS	1
2647	2438	S	2
2648	2438	M	3
2649	2438	L	4
2650	2438	XL	5
2651	2438	XXL	6
2652	2438	XXXL	7
2653	2438	28	8
2654	2438	30	9
2655	2438	32	10
2656	2438	34	11
2657	2438	36	12
2658	2438	38	13
2659	2438	40	14
2660	2438	42	15
2661	2438	Free Size	16
2662	2443	XS	1
2663	2443	S	2
2664	2443	M	3
2665	2443	L	4
2666	2443	XL	5
2667	2443	XXL	6
2668	2443	XXXL	7
2669	2443	28	8
2670	2443	30	9
2671	2443	32	10
2672	2443	34	11
2673	2443	36	12
2674	2443	38	13
2675	2443	40	14
2676	2443	42	15
2677	2443	Free Size	16
2678	2448	XS	1
2679	2448	S	2
2680	2448	M	3
2681	2448	L	4
2682	2448	XL	5
2683	2448	XXL	6
2684	2448	XXXL	7
2685	2448	28	8
2686	2448	30	9
2687	2448	32	10
2688	2448	34	11
2689	2448	36	12
2690	2448	38	13
2691	2448	40	14
2692	2448	42	15
2693	2448	Free Size	16
2694	2453	XS	1
2695	2453	S	2
2696	2453	M	3
2697	2453	L	4
2698	2453	XL	5
2699	2453	XXL	6
2700	2453	XXXL	7
2701	2453	28	8
2702	2453	30	9
2703	2453	32	10
2704	2453	34	11
2705	2453	36	12
2706	2453	38	13
2707	2453	40	14
2708	2453	42	15
2709	2453	Free Size	16
2710	2458	XS	1
2711	2458	S	2
2712	2458	M	3
2713	2458	L	4
2714	2458	XL	5
2715	2458	XXL	6
2716	2458	XXXL	7
2717	2458	28	8
2718	2458	30	9
2719	2458	32	10
2720	2458	34	11
2721	2458	36	12
2722	2458	38	13
2723	2458	40	14
2724	2458	42	15
2725	2458	Free Size	16
2726	2463	XS	1
2727	2463	S	2
2728	2463	M	3
2729	2463	L	4
2730	2463	XL	5
2731	2463	XXL	6
2732	2463	XXXL	7
2733	2463	28	8
2734	2463	30	9
2735	2463	32	10
2736	2463	34	11
2737	2463	36	12
2738	2463	38	13
2739	2463	40	14
2740	2463	42	15
2741	2463	Free Size	16
2742	2468	XS	1
2743	2468	S	2
2744	2468	M	3
2745	2468	L	4
2746	2468	XL	5
2747	2468	XXL	6
2748	2468	XXXL	7
2749	2468	28	8
2750	2468	30	9
2751	2468	32	10
2752	2468	34	11
2753	2468	36	12
2754	2468	38	13
2755	2468	40	14
2756	2468	42	15
2757	2468	Free Size	16
2758	2473	XS	1
2759	2473	S	2
2760	2473	M	3
2761	2473	L	4
2762	2473	XL	5
2763	2473	XXL	6
2764	2473	XXXL	7
2765	2473	28	8
2766	2473	30	9
2767	2473	32	10
2768	2473	34	11
2769	2473	36	12
2770	2473	38	13
2771	2473	40	14
2772	2473	42	15
2773	2473	Free Size	16
2774	2478	XS	1
2775	2478	S	2
2776	2478	M	3
2777	2478	L	4
2778	2478	XL	5
2779	2478	XXL	6
2780	2478	XXXL	7
2781	2478	28	8
2782	2478	30	9
2783	2478	32	10
2784	2478	34	11
2785	2478	36	12
2786	2478	38	13
2787	2478	40	14
2788	2478	42	15
2789	2478	Free Size	16
2790	2483	XS	1
2791	2483	S	2
2792	2483	M	3
2793	2483	L	4
2794	2483	XL	5
2795	2483	XXL	6
2796	2483	XXXL	7
2797	2483	28	8
2798	2483	30	9
2799	2483	32	10
2800	2483	34	11
2801	2483	36	12
2802	2483	38	13
2803	2483	40	14
2804	2483	42	15
2805	2483	Free Size	16
2806	2528	XS	1
2807	2528	S	2
2808	2528	M	3
2809	2528	L	4
2810	2528	XL	5
2811	2528	XXL	6
2812	2528	XXXL	7
2813	2528	28	8
2814	2528	30	9
2815	2528	32	10
2816	2528	34	11
2817	2528	36	12
2818	2528	38	13
2819	2528	40	14
2820	2528	42	15
2821	2528	Free Size	16
2822	2532	XS	1
2823	2532	S	2
2824	2532	M	3
2825	2532	L	4
2826	2532	XL	5
2827	2532	XXL	6
2828	2532	XXXL	7
2829	2532	28	8
2830	2532	30	9
2831	2532	32	10
2832	2532	34	11
2833	2532	36	12
2834	2532	38	13
2835	2532	40	14
2836	2532	42	15
2837	2532	Free Size	16
2838	2538	XS	1
2839	2538	S	2
2840	2538	M	3
2841	2538	L	4
2842	2538	XL	5
2843	2538	XXL	6
2844	2538	XXXL	7
2845	2538	28	8
2846	2538	30	9
2847	2538	32	10
2848	2538	34	11
2849	2538	36	12
2850	2538	38	13
2851	2538	40	14
2852	2538	42	15
2853	2538	Free Size	16
2854	2542	XS	1
2855	2542	S	2
2856	2542	M	3
2857	2542	L	4
2858	2542	XL	5
2859	2542	XXL	6
2860	2542	XXXL	7
2861	2542	28	8
2862	2542	30	9
2863	2542	32	10
2864	2542	34	11
2865	2542	36	12
2866	2542	38	13
2867	2542	40	14
2868	2542	42	15
2869	2542	Free Size	16
2870	2547	XS	1
2871	2547	S	2
2872	2547	M	3
2873	2547	L	4
2874	2547	XL	5
2875	2547	XXL	6
2876	2547	XXXL	7
2877	2547	28	8
2878	2547	30	9
2879	2547	32	10
2880	2547	34	11
2881	2547	36	12
2882	2547	38	13
2883	2547	40	14
2884	2547	42	15
2885	2547	Free Size	16
2886	2565	XS	1
2887	2565	S	2
2888	2565	M	3
2889	2565	L	4
2890	2565	XL	5
2891	2565	XXL	6
2892	2565	XXXL	7
2893	2565	28	8
2894	2565	30	9
2895	2565	32	10
2896	2565	34	11
2897	2565	36	12
2898	2565	38	13
2899	2565	40	14
2900	2565	42	15
2901	2565	Free Size	16
2902	45	Cotton	1
2903	45	Polyester	2
2904	45	Silk	3
2905	45	Wool	4
2906	45	Leather	5
2907	45	Wood	6
2908	45	Metal	7
2909	45	Plastic	8
2910	45	Glass	9
2911	45	Steel	10
2912	45	Stainless Steel	11
2913	45	Aluminum	12
2914	45	Brass	13
2915	45	Bronze	14
2916	45	Fabric	15
2917	45	Synthetic	16
2918	45	Nylon	17
2919	45	Canvas	18
2920	52	Cotton	1
2921	52	Polyester	2
2922	52	Silk	3
2923	52	Wool	4
2924	52	Leather	5
2925	52	Wood	6
2926	52	Metal	7
2927	52	Plastic	8
2928	52	Glass	9
2929	52	Steel	10
2930	52	Stainless Steel	11
2931	52	Aluminum	12
2932	52	Brass	13
2933	52	Bronze	14
2934	52	Fabric	15
2935	52	Synthetic	16
2936	52	Nylon	17
2937	52	Canvas	18
2938	59	Cotton	1
2939	59	Polyester	2
2940	59	Silk	3
2941	59	Wool	4
2942	59	Leather	5
2943	59	Wood	6
2944	59	Metal	7
2945	59	Plastic	8
2946	59	Glass	9
2947	59	Steel	10
2948	59	Stainless Steel	11
2949	59	Aluminum	12
2950	59	Brass	13
2951	59	Bronze	14
2952	59	Fabric	15
2953	59	Synthetic	16
2954	59	Nylon	17
2955	59	Canvas	18
2956	66	Cotton	1
2957	66	Polyester	2
2958	66	Silk	3
2959	66	Wool	4
2960	66	Leather	5
2961	66	Wood	6
2962	66	Metal	7
2963	66	Plastic	8
2964	66	Glass	9
2965	66	Steel	10
2966	66	Stainless Steel	11
2967	66	Aluminum	12
2968	66	Brass	13
2969	66	Bronze	14
2970	66	Fabric	15
2971	66	Synthetic	16
2972	66	Nylon	17
2973	66	Canvas	18
2974	73	Cotton	1
2975	73	Polyester	2
2976	73	Silk	3
2977	73	Wool	4
2978	73	Leather	5
2979	73	Wood	6
2980	73	Metal	7
2981	73	Plastic	8
2982	73	Glass	9
2983	73	Steel	10
2984	73	Stainless Steel	11
2985	73	Aluminum	12
2986	73	Brass	13
2987	73	Bronze	14
2988	73	Fabric	15
2989	73	Synthetic	16
2990	73	Nylon	17
2991	73	Canvas	18
2992	79	Cotton	1
2993	79	Polyester	2
2994	79	Silk	3
2995	79	Wool	4
2996	79	Leather	5
2997	79	Wood	6
2998	79	Metal	7
2999	79	Plastic	8
3000	79	Glass	9
3001	79	Steel	10
3002	79	Stainless Steel	11
3003	79	Aluminum	12
3004	79	Brass	13
3005	79	Bronze	14
3006	79	Fabric	15
3007	79	Synthetic	16
3008	79	Nylon	17
3009	79	Canvas	18
3010	84	Cotton	1
3011	84	Polyester	2
3012	84	Silk	3
3013	84	Wool	4
3014	84	Leather	5
3015	84	Wood	6
3016	84	Metal	7
3017	84	Plastic	8
3018	84	Glass	9
3019	84	Steel	10
3020	84	Stainless Steel	11
3021	84	Aluminum	12
3022	84	Brass	13
3023	84	Bronze	14
3024	84	Fabric	15
3025	84	Synthetic	16
3026	84	Nylon	17
3027	84	Canvas	18
3028	89	Cotton	1
3029	89	Polyester	2
3030	89	Silk	3
3031	89	Wool	4
3032	89	Leather	5
3033	89	Wood	6
3034	89	Metal	7
3035	89	Plastic	8
3036	89	Glass	9
3037	89	Steel	10
3038	89	Stainless Steel	11
3039	89	Aluminum	12
3040	89	Brass	13
3041	89	Bronze	14
3042	89	Fabric	15
3043	89	Synthetic	16
3044	89	Nylon	17
3045	89	Canvas	18
3046	93	Cotton	1
3047	93	Polyester	2
3048	93	Silk	3
3049	93	Wool	4
3050	93	Leather	5
3051	93	Wood	6
3052	93	Metal	7
3053	93	Plastic	8
3054	93	Glass	9
3055	93	Steel	10
3056	93	Stainless Steel	11
3057	93	Aluminum	12
3058	93	Brass	13
3059	93	Bronze	14
3060	93	Fabric	15
3061	93	Synthetic	16
3062	93	Nylon	17
3063	93	Canvas	18
3064	98	Cotton	1
3065	98	Polyester	2
3066	98	Silk	3
3067	98	Wool	4
3068	98	Leather	5
3069	98	Wood	6
3070	98	Metal	7
3071	98	Plastic	8
3072	98	Glass	9
3073	98	Steel	10
3074	98	Stainless Steel	11
3075	98	Aluminum	12
3076	98	Brass	13
3077	98	Bronze	14
3078	98	Fabric	15
3079	98	Synthetic	16
3080	98	Nylon	17
3081	98	Canvas	18
3082	103	Cotton	1
3083	103	Polyester	2
3084	103	Silk	3
3085	103	Wool	4
3086	103	Leather	5
3087	103	Wood	6
3088	103	Metal	7
3089	103	Plastic	8
3090	103	Glass	9
3091	103	Steel	10
3092	103	Stainless Steel	11
3093	103	Aluminum	12
3094	103	Brass	13
3095	103	Bronze	14
3096	103	Fabric	15
3097	103	Synthetic	16
3098	103	Nylon	17
3099	103	Canvas	18
3100	108	Cotton	1
3101	108	Polyester	2
3102	108	Silk	3
3103	108	Wool	4
3104	108	Leather	5
3105	108	Wood	6
3106	108	Metal	7
3107	108	Plastic	8
3108	108	Glass	9
3109	108	Steel	10
3110	108	Stainless Steel	11
3111	108	Aluminum	12
3112	108	Brass	13
3113	108	Bronze	14
3114	108	Fabric	15
3115	108	Synthetic	16
3116	108	Nylon	17
3117	108	Canvas	18
3118	114	Cotton	1
3119	114	Polyester	2
3120	114	Silk	3
3121	114	Wool	4
3122	114	Leather	5
3123	114	Wood	6
3124	114	Metal	7
3125	114	Plastic	8
3126	114	Glass	9
3127	114	Steel	10
3128	114	Stainless Steel	11
3129	114	Aluminum	12
3130	114	Brass	13
3131	114	Bronze	14
3132	114	Fabric	15
3133	114	Synthetic	16
3134	114	Nylon	17
3135	114	Canvas	18
3136	118	Cotton	1
3137	118	Polyester	2
3138	118	Silk	3
3139	118	Wool	4
3140	118	Leather	5
3141	118	Wood	6
3142	118	Metal	7
3143	118	Plastic	8
3144	118	Glass	9
3145	118	Steel	10
3146	118	Stainless Steel	11
3147	118	Aluminum	12
3148	118	Brass	13
3149	118	Bronze	14
3150	118	Fabric	15
3151	118	Synthetic	16
3152	118	Nylon	17
3153	118	Canvas	18
3154	121	Cotton	1
3155	121	Polyester	2
3156	121	Silk	3
3157	121	Wool	4
3158	121	Leather	5
3159	121	Wood	6
3160	121	Metal	7
3161	121	Plastic	8
3162	121	Glass	9
3163	121	Steel	10
3164	121	Stainless Steel	11
3165	121	Aluminum	12
3166	121	Brass	13
3167	121	Bronze	14
3168	121	Fabric	15
3169	121	Synthetic	16
3170	121	Nylon	17
3171	121	Canvas	18
3172	125	Cotton	1
3173	125	Polyester	2
3174	125	Silk	3
3175	125	Wool	4
3176	125	Leather	5
3177	125	Wood	6
3178	125	Metal	7
3179	125	Plastic	8
3180	125	Glass	9
3181	125	Steel	10
3182	125	Stainless Steel	11
3183	125	Aluminum	12
3184	125	Brass	13
3185	125	Bronze	14
3186	125	Fabric	15
3187	125	Synthetic	16
3188	125	Nylon	17
3189	125	Canvas	18
3190	129	Cotton	1
3191	129	Polyester	2
3192	129	Silk	3
3193	129	Wool	4
3194	129	Leather	5
3195	129	Wood	6
3196	129	Metal	7
3197	129	Plastic	8
3198	129	Glass	9
3199	129	Steel	10
3200	129	Stainless Steel	11
3201	129	Aluminum	12
3202	129	Brass	13
3203	129	Bronze	14
3204	129	Fabric	15
3205	129	Synthetic	16
3206	129	Nylon	17
3207	129	Canvas	18
3208	132	Cotton	1
3209	132	Polyester	2
3210	132	Silk	3
3211	132	Wool	4
3212	132	Leather	5
3213	132	Wood	6
3214	132	Metal	7
3215	132	Plastic	8
3216	132	Glass	9
3217	132	Steel	10
3218	132	Stainless Steel	11
3219	132	Aluminum	12
3220	132	Brass	13
3221	132	Bronze	14
3222	132	Fabric	15
3223	132	Synthetic	16
3224	132	Nylon	17
3225	132	Canvas	18
3226	136	Cotton	1
3227	136	Polyester	2
3228	136	Silk	3
3229	136	Wool	4
3230	136	Leather	5
3231	136	Wood	6
3232	136	Metal	7
3233	136	Plastic	8
3234	136	Glass	9
3235	136	Steel	10
3236	136	Stainless Steel	11
3237	136	Aluminum	12
3238	136	Brass	13
3239	136	Bronze	14
3240	136	Fabric	15
3241	136	Synthetic	16
3242	136	Nylon	17
3243	136	Canvas	18
3244	189	Cotton	1
3245	189	Polyester	2
3246	189	Silk	3
3247	189	Wool	4
3248	189	Leather	5
3249	189	Wood	6
3250	189	Metal	7
3251	189	Plastic	8
3252	189	Glass	9
3253	189	Steel	10
3254	189	Stainless Steel	11
3255	189	Aluminum	12
3256	189	Brass	13
3257	189	Bronze	14
3258	189	Fabric	15
3259	189	Synthetic	16
3260	189	Nylon	17
3261	189	Canvas	18
3262	233	Cotton	1
3263	233	Polyester	2
3264	233	Silk	3
3265	233	Wool	4
3266	233	Leather	5
3267	233	Wood	6
3268	233	Metal	7
3269	233	Plastic	8
3270	233	Glass	9
3271	233	Steel	10
3272	233	Stainless Steel	11
3273	233	Aluminum	12
3274	233	Brass	13
3275	233	Bronze	14
3276	233	Fabric	15
3277	233	Synthetic	16
3278	233	Nylon	17
3279	233	Canvas	18
3280	277	Cotton	1
3281	277	Polyester	2
3282	277	Silk	3
3283	277	Wool	4
3284	277	Leather	5
3285	277	Wood	6
3286	277	Metal	7
3287	277	Plastic	8
3288	277	Glass	9
3289	277	Steel	10
3290	277	Stainless Steel	11
3291	277	Aluminum	12
3292	277	Brass	13
3293	277	Bronze	14
3294	277	Fabric	15
3295	277	Synthetic	16
3296	277	Nylon	17
3297	277	Canvas	18
3298	339	Cotton	1
3299	339	Polyester	2
3300	339	Silk	3
3301	339	Wool	4
3302	339	Leather	5
3303	339	Wood	6
3304	339	Metal	7
3305	339	Plastic	8
3306	339	Glass	9
3307	339	Steel	10
3308	339	Stainless Steel	11
3309	339	Aluminum	12
3310	339	Brass	13
3311	339	Bronze	14
3312	339	Fabric	15
3313	339	Synthetic	16
3314	339	Nylon	17
3315	339	Canvas	18
3316	343	Cotton	1
3317	343	Polyester	2
3318	343	Silk	3
3319	343	Wool	4
3320	343	Leather	5
3321	343	Wood	6
3322	343	Metal	7
3323	343	Plastic	8
3324	343	Glass	9
3325	343	Steel	10
3326	343	Stainless Steel	11
3327	343	Aluminum	12
3328	343	Brass	13
3329	343	Bronze	14
3330	343	Fabric	15
3331	343	Synthetic	16
3332	343	Nylon	17
3333	343	Canvas	18
3334	365	Cotton	1
3335	365	Polyester	2
3336	365	Silk	3
3337	365	Wool	4
3338	365	Leather	5
3339	365	Wood	6
3340	365	Metal	7
3341	365	Plastic	8
3342	365	Glass	9
3343	365	Steel	10
3344	365	Stainless Steel	11
3345	365	Aluminum	12
3346	365	Brass	13
3347	365	Bronze	14
3348	365	Fabric	15
3349	365	Synthetic	16
3350	365	Nylon	17
3351	365	Canvas	18
3352	371	Cotton	1
3353	371	Polyester	2
3354	371	Silk	3
3355	371	Wool	4
3356	371	Leather	5
3357	371	Wood	6
3358	371	Metal	7
3359	371	Plastic	8
3360	371	Glass	9
3361	371	Steel	10
3362	371	Stainless Steel	11
3363	371	Aluminum	12
3364	371	Brass	13
3365	371	Bronze	14
3366	371	Fabric	15
3367	371	Synthetic	16
3368	371	Nylon	17
3369	371	Canvas	18
3370	376	Cotton	1
3371	376	Polyester	2
3372	376	Silk	3
3373	376	Wool	4
3374	376	Leather	5
3375	376	Wood	6
3376	376	Metal	7
3377	376	Plastic	8
3378	376	Glass	9
3379	376	Steel	10
3380	376	Stainless Steel	11
3381	376	Aluminum	12
3382	376	Brass	13
3383	376	Bronze	14
3384	376	Fabric	15
3385	376	Synthetic	16
3386	376	Nylon	17
3387	376	Canvas	18
3388	385	Cotton	1
3389	385	Polyester	2
3390	385	Silk	3
3391	385	Wool	4
3392	385	Leather	5
3393	385	Wood	6
3394	385	Metal	7
3395	385	Plastic	8
3396	385	Glass	9
3397	385	Steel	10
3398	385	Stainless Steel	11
3399	385	Aluminum	12
3400	385	Brass	13
3401	385	Bronze	14
3402	385	Fabric	15
3403	385	Synthetic	16
3404	385	Nylon	17
3405	385	Canvas	18
3406	402	Cotton	1
3407	402	Polyester	2
3408	402	Silk	3
3409	402	Wool	4
3410	402	Leather	5
3411	402	Wood	6
3412	402	Metal	7
3413	402	Plastic	8
3414	402	Glass	9
3415	402	Steel	10
3416	402	Stainless Steel	11
3417	402	Aluminum	12
3418	402	Brass	13
3419	402	Bronze	14
3420	402	Fabric	15
3421	402	Synthetic	16
3422	402	Nylon	17
3423	402	Canvas	18
3424	411	Cotton	1
3425	411	Polyester	2
3426	411	Silk	3
3427	411	Wool	4
3428	411	Leather	5
3429	411	Wood	6
3430	411	Metal	7
3431	411	Plastic	8
3432	411	Glass	9
3433	411	Steel	10
3434	411	Stainless Steel	11
3435	411	Aluminum	12
3436	411	Brass	13
3437	411	Bronze	14
3438	411	Fabric	15
3439	411	Synthetic	16
3440	411	Nylon	17
3441	411	Canvas	18
3442	416	Cotton	1
3443	416	Polyester	2
3444	416	Silk	3
3445	416	Wool	4
3446	416	Leather	5
3447	416	Wood	6
3448	416	Metal	7
3449	416	Plastic	8
3450	416	Glass	9
3451	416	Steel	10
3452	416	Stainless Steel	11
3453	416	Aluminum	12
3454	416	Brass	13
3455	416	Bronze	14
3456	416	Fabric	15
3457	416	Synthetic	16
3458	416	Nylon	17
3459	416	Canvas	18
3460	421	Cotton	1
3461	421	Polyester	2
3462	421	Silk	3
3463	421	Wool	4
3464	421	Leather	5
3465	421	Wood	6
3466	421	Metal	7
3467	421	Plastic	8
3468	421	Glass	9
3469	421	Steel	10
3470	421	Stainless Steel	11
3471	421	Aluminum	12
3472	421	Brass	13
3473	421	Bronze	14
3474	421	Fabric	15
3475	421	Synthetic	16
3476	421	Nylon	17
3477	421	Canvas	18
3478	426	Cotton	1
3479	426	Polyester	2
3480	426	Silk	3
3481	426	Wool	4
3482	426	Leather	5
3483	426	Wood	6
3484	426	Metal	7
3485	426	Plastic	8
3486	426	Glass	9
3487	426	Steel	10
3488	426	Stainless Steel	11
3489	426	Aluminum	12
3490	426	Brass	13
3491	426	Bronze	14
3492	426	Fabric	15
3493	426	Synthetic	16
3494	426	Nylon	17
3495	426	Canvas	18
3496	440	Cotton	1
3497	440	Polyester	2
3498	440	Silk	3
3499	440	Wool	4
3500	440	Leather	5
3501	440	Wood	6
3502	440	Metal	7
3503	440	Plastic	8
3504	440	Glass	9
3505	440	Steel	10
3506	440	Stainless Steel	11
3507	440	Aluminum	12
3508	440	Brass	13
3509	440	Bronze	14
3510	440	Fabric	15
3511	440	Synthetic	16
3512	440	Nylon	17
3513	440	Canvas	18
3514	457	Cotton	1
3515	457	Polyester	2
3516	457	Silk	3
3517	457	Wool	4
3518	457	Leather	5
3519	457	Wood	6
3520	457	Metal	7
3521	457	Plastic	8
3522	457	Glass	9
3523	457	Steel	10
3524	457	Stainless Steel	11
3525	457	Aluminum	12
3526	457	Brass	13
3527	457	Bronze	14
3528	457	Fabric	15
3529	457	Synthetic	16
3530	457	Nylon	17
3531	457	Canvas	18
3532	462	Cotton	1
3533	462	Polyester	2
3534	462	Silk	3
3535	462	Wool	4
3536	462	Leather	5
3537	462	Wood	6
3538	462	Metal	7
3539	462	Plastic	8
3540	462	Glass	9
3541	462	Steel	10
3542	462	Stainless Steel	11
3543	462	Aluminum	12
3544	462	Brass	13
3545	462	Bronze	14
3546	462	Fabric	15
3547	462	Synthetic	16
3548	462	Nylon	17
3549	462	Canvas	18
3550	473	Cotton	1
3551	473	Polyester	2
3552	473	Silk	3
3553	473	Wool	4
3554	473	Leather	5
3555	473	Wood	6
3556	473	Metal	7
3557	473	Plastic	8
3558	473	Glass	9
3559	473	Steel	10
3560	473	Stainless Steel	11
3561	473	Aluminum	12
3562	473	Brass	13
3563	473	Bronze	14
3564	473	Fabric	15
3565	473	Synthetic	16
3566	473	Nylon	17
3567	473	Canvas	18
3568	478	Cotton	1
3569	478	Polyester	2
3570	478	Silk	3
3571	478	Wool	4
3572	478	Leather	5
3573	478	Wood	6
3574	478	Metal	7
3575	478	Plastic	8
3576	478	Glass	9
3577	478	Steel	10
3578	478	Stainless Steel	11
3579	478	Aluminum	12
3580	478	Brass	13
3581	478	Bronze	14
3582	478	Fabric	15
3583	478	Synthetic	16
3584	478	Nylon	17
3585	478	Canvas	18
3586	488	Cotton	1
3587	488	Polyester	2
3588	488	Silk	3
3589	488	Wool	4
3590	488	Leather	5
3591	488	Wood	6
3592	488	Metal	7
3593	488	Plastic	8
3594	488	Glass	9
3595	488	Steel	10
3596	488	Stainless Steel	11
3597	488	Aluminum	12
3598	488	Brass	13
3599	488	Bronze	14
3600	488	Fabric	15
3601	488	Synthetic	16
3602	488	Nylon	17
3603	488	Canvas	18
3604	498	Cotton	1
3605	498	Polyester	2
3606	498	Silk	3
3607	498	Wool	4
3608	498	Leather	5
3609	498	Wood	6
3610	498	Metal	7
3611	498	Plastic	8
3612	498	Glass	9
3613	498	Steel	10
3614	498	Stainless Steel	11
3615	498	Aluminum	12
3616	498	Brass	13
3617	498	Bronze	14
3618	498	Fabric	15
3619	498	Synthetic	16
3620	498	Nylon	17
3621	498	Canvas	18
3622	501	Cotton	1
3623	501	Polyester	2
3624	501	Silk	3
3625	501	Wool	4
3626	501	Leather	5
3627	501	Wood	6
3628	501	Metal	7
3629	501	Plastic	8
3630	501	Glass	9
3631	501	Steel	10
3632	501	Stainless Steel	11
3633	501	Aluminum	12
3634	501	Brass	13
3635	501	Bronze	14
3636	501	Fabric	15
3637	501	Synthetic	16
3638	501	Nylon	17
3639	501	Canvas	18
3640	516	Cotton	1
3641	516	Polyester	2
3642	516	Silk	3
3643	516	Wool	4
3644	516	Leather	5
3645	516	Wood	6
3646	516	Metal	7
3647	516	Plastic	8
3648	516	Glass	9
3649	516	Steel	10
3650	516	Stainless Steel	11
3651	516	Aluminum	12
3652	516	Brass	13
3653	516	Bronze	14
3654	516	Fabric	15
3655	516	Synthetic	16
3656	516	Nylon	17
3657	516	Canvas	18
3658	519	Cotton	1
3659	519	Polyester	2
3660	519	Silk	3
3661	519	Wool	4
3662	519	Leather	5
3663	519	Wood	6
3664	519	Metal	7
3665	519	Plastic	8
3666	519	Glass	9
3667	519	Steel	10
3668	519	Stainless Steel	11
3669	519	Aluminum	12
3670	519	Brass	13
3671	519	Bronze	14
3672	519	Fabric	15
3673	519	Synthetic	16
3674	519	Nylon	17
3675	519	Canvas	18
3676	612	Cotton	1
3677	612	Polyester	2
3678	612	Silk	3
3679	612	Wool	4
3680	612	Leather	5
3681	612	Wood	6
3682	612	Metal	7
3683	612	Plastic	8
3684	612	Glass	9
3685	612	Steel	10
3686	612	Stainless Steel	11
3687	612	Aluminum	12
3688	612	Brass	13
3689	612	Bronze	14
3690	612	Fabric	15
3691	612	Synthetic	16
3692	612	Nylon	17
3693	612	Canvas	18
3694	639	Cotton	1
3695	639	Polyester	2
3696	639	Silk	3
3697	639	Wool	4
3698	639	Leather	5
3699	639	Wood	6
3700	639	Metal	7
3701	639	Plastic	8
3702	639	Glass	9
3703	639	Steel	10
3704	639	Stainless Steel	11
3705	639	Aluminum	12
3706	639	Brass	13
3707	639	Bronze	14
3708	639	Fabric	15
3709	639	Synthetic	16
3710	639	Nylon	17
3711	639	Canvas	18
3712	648	Cotton	1
3713	648	Polyester	2
3714	648	Silk	3
3715	648	Wool	4
3716	648	Leather	5
3717	648	Wood	6
3718	648	Metal	7
3719	648	Plastic	8
3720	648	Glass	9
3721	648	Steel	10
3722	648	Stainless Steel	11
3723	648	Aluminum	12
3724	648	Brass	13
3725	648	Bronze	14
3726	648	Fabric	15
3727	648	Synthetic	16
3728	648	Nylon	17
3729	648	Canvas	18
3730	653	Cotton	1
3731	653	Polyester	2
3732	653	Silk	3
3733	653	Wool	4
3734	653	Leather	5
3735	653	Wood	6
3736	653	Metal	7
3737	653	Plastic	8
3738	653	Glass	9
3739	653	Steel	10
3740	653	Stainless Steel	11
3741	653	Aluminum	12
3742	653	Brass	13
3743	653	Bronze	14
3744	653	Fabric	15
3745	653	Synthetic	16
3746	653	Nylon	17
3747	653	Canvas	18
3748	657	Cotton	1
3749	657	Polyester	2
3750	657	Silk	3
3751	657	Wool	4
3752	657	Leather	5
3753	657	Wood	6
3754	657	Metal	7
3755	657	Plastic	8
3756	657	Glass	9
3757	657	Steel	10
3758	657	Stainless Steel	11
3759	657	Aluminum	12
3760	657	Brass	13
3761	657	Bronze	14
3762	657	Fabric	15
3763	657	Synthetic	16
3764	657	Nylon	17
3765	657	Canvas	18
3766	701	Cotton	1
3767	701	Polyester	2
3768	701	Silk	3
3769	701	Wool	4
3770	701	Leather	5
3771	701	Wood	6
3772	701	Metal	7
3773	701	Plastic	8
3774	701	Glass	9
3775	701	Steel	10
3776	701	Stainless Steel	11
3777	701	Aluminum	12
3778	701	Brass	13
3779	701	Bronze	14
3780	701	Fabric	15
3781	701	Synthetic	16
3782	701	Nylon	17
3783	701	Canvas	18
3784	706	Cotton	1
3785	706	Polyester	2
3786	706	Silk	3
3787	706	Wool	4
3788	706	Leather	5
3789	706	Wood	6
3790	706	Metal	7
3791	706	Plastic	8
3792	706	Glass	9
3793	706	Steel	10
3794	706	Stainless Steel	11
3795	706	Aluminum	12
3796	706	Brass	13
3797	706	Bronze	14
3798	706	Fabric	15
3799	706	Synthetic	16
3800	706	Nylon	17
3801	706	Canvas	18
3802	816	Cotton	1
3803	816	Polyester	2
3804	816	Silk	3
3805	816	Wool	4
3806	816	Leather	5
3807	816	Wood	6
3808	816	Metal	7
3809	816	Plastic	8
3810	816	Glass	9
3811	816	Steel	10
3812	816	Stainless Steel	11
3813	816	Aluminum	12
3814	816	Brass	13
3815	816	Bronze	14
3816	816	Fabric	15
3817	816	Synthetic	16
3818	816	Nylon	17
3819	816	Canvas	18
3820	830	Cotton	1
3821	830	Polyester	2
3822	830	Silk	3
3823	830	Wool	4
3824	830	Leather	5
3825	830	Wood	6
3826	830	Metal	7
3827	830	Plastic	8
3828	830	Glass	9
3829	830	Steel	10
3830	830	Stainless Steel	11
3831	830	Aluminum	12
3832	830	Brass	13
3833	830	Bronze	14
3834	830	Fabric	15
3835	830	Synthetic	16
3836	830	Nylon	17
3837	830	Canvas	18
3838	836	Cotton	1
3839	836	Polyester	2
3840	836	Silk	3
3841	836	Wool	4
3842	836	Leather	5
3843	836	Wood	6
3844	836	Metal	7
3845	836	Plastic	8
3846	836	Glass	9
3847	836	Steel	10
3848	836	Stainless Steel	11
3849	836	Aluminum	12
3850	836	Brass	13
3851	836	Bronze	14
3852	836	Fabric	15
3853	836	Synthetic	16
3854	836	Nylon	17
3855	836	Canvas	18
3856	842	Cotton	1
3857	842	Polyester	2
3858	842	Silk	3
3859	842	Wool	4
3860	842	Leather	5
3861	842	Wood	6
3862	842	Metal	7
3863	842	Plastic	8
3864	842	Glass	9
3865	842	Steel	10
3866	842	Stainless Steel	11
3867	842	Aluminum	12
3868	842	Brass	13
3869	842	Bronze	14
3870	842	Fabric	15
3871	842	Synthetic	16
3872	842	Nylon	17
3873	842	Canvas	18
3874	847	Cotton	1
3875	847	Polyester	2
3876	847	Silk	3
3877	847	Wool	4
3878	847	Leather	5
3879	847	Wood	6
3880	847	Metal	7
3881	847	Plastic	8
3882	847	Glass	9
3883	847	Steel	10
3884	847	Stainless Steel	11
3885	847	Aluminum	12
3886	847	Brass	13
3887	847	Bronze	14
3888	847	Fabric	15
3889	847	Synthetic	16
3890	847	Nylon	17
3891	847	Canvas	18
3892	852	Cotton	1
3893	852	Polyester	2
3894	852	Silk	3
3895	852	Wool	4
3896	852	Leather	5
3897	852	Wood	6
3898	852	Metal	7
3899	852	Plastic	8
3900	852	Glass	9
3901	852	Steel	10
3902	852	Stainless Steel	11
3903	852	Aluminum	12
3904	852	Brass	13
3905	852	Bronze	14
3906	852	Fabric	15
3907	852	Synthetic	16
3908	852	Nylon	17
3909	852	Canvas	18
3910	857	Cotton	1
3911	857	Polyester	2
3912	857	Silk	3
3913	857	Wool	4
3914	857	Leather	5
3915	857	Wood	6
3916	857	Metal	7
3917	857	Plastic	8
3918	857	Glass	9
3919	857	Steel	10
3920	857	Stainless Steel	11
3921	857	Aluminum	12
3922	857	Brass	13
3923	857	Bronze	14
3924	857	Fabric	15
3925	857	Synthetic	16
3926	857	Nylon	17
3927	857	Canvas	18
3928	862	Cotton	1
3929	862	Polyester	2
3930	862	Silk	3
3931	862	Wool	4
3932	862	Leather	5
3933	862	Wood	6
3934	862	Metal	7
3935	862	Plastic	8
3936	862	Glass	9
3937	862	Steel	10
3938	862	Stainless Steel	11
3939	862	Aluminum	12
3940	862	Brass	13
3941	862	Bronze	14
3942	862	Fabric	15
3943	862	Synthetic	16
3944	862	Nylon	17
3945	862	Canvas	18
3946	868	Cotton	1
3947	868	Polyester	2
3948	868	Silk	3
3949	868	Wool	4
3950	868	Leather	5
3951	868	Wood	6
3952	868	Metal	7
3953	868	Plastic	8
3954	868	Glass	9
3955	868	Steel	10
3956	868	Stainless Steel	11
3957	868	Aluminum	12
3958	868	Brass	13
3959	868	Bronze	14
3960	868	Fabric	15
3961	868	Synthetic	16
3962	868	Nylon	17
3963	868	Canvas	18
3964	873	Cotton	1
3965	873	Polyester	2
3966	873	Silk	3
3967	873	Wool	4
3968	873	Leather	5
3969	873	Wood	6
3970	873	Metal	7
3971	873	Plastic	8
3972	873	Glass	9
3973	873	Steel	10
3974	873	Stainless Steel	11
3975	873	Aluminum	12
3976	873	Brass	13
3977	873	Bronze	14
3978	873	Fabric	15
3979	873	Synthetic	16
3980	873	Nylon	17
3981	873	Canvas	18
3982	877	Cotton	1
3983	877	Polyester	2
3984	877	Silk	3
3985	877	Wool	4
3986	877	Leather	5
3987	877	Wood	6
3988	877	Metal	7
3989	877	Plastic	8
3990	877	Glass	9
3991	877	Steel	10
3992	877	Stainless Steel	11
3993	877	Aluminum	12
3994	877	Brass	13
3995	877	Bronze	14
3996	877	Fabric	15
3997	877	Synthetic	16
3998	877	Nylon	17
3999	877	Canvas	18
4000	882	Cotton	1
4001	882	Polyester	2
4002	882	Silk	3
4003	882	Wool	4
4004	882	Leather	5
4005	882	Wood	6
4006	882	Metal	7
4007	882	Plastic	8
4008	882	Glass	9
4009	882	Steel	10
4010	882	Stainless Steel	11
4011	882	Aluminum	12
4012	882	Brass	13
4013	882	Bronze	14
4014	882	Fabric	15
4015	882	Synthetic	16
4016	882	Nylon	17
4017	882	Canvas	18
4018	887	Cotton	1
4019	887	Polyester	2
4020	887	Silk	3
4021	887	Wool	4
4022	887	Leather	5
4023	887	Wood	6
4024	887	Metal	7
4025	887	Plastic	8
4026	887	Glass	9
4027	887	Steel	10
4028	887	Stainless Steel	11
4029	887	Aluminum	12
4030	887	Brass	13
4031	887	Bronze	14
4032	887	Fabric	15
4033	887	Synthetic	16
4034	887	Nylon	17
4035	887	Canvas	18
4036	892	Cotton	1
4037	892	Polyester	2
4038	892	Silk	3
4039	892	Wool	4
4040	892	Leather	5
4041	892	Wood	6
4042	892	Metal	7
4043	892	Plastic	8
4044	892	Glass	9
4045	892	Steel	10
4046	892	Stainless Steel	11
4047	892	Aluminum	12
4048	892	Brass	13
4049	892	Bronze	14
4050	892	Fabric	15
4051	892	Synthetic	16
4052	892	Nylon	17
4053	892	Canvas	18
4054	897	Cotton	1
4055	897	Polyester	2
4056	897	Silk	3
4057	897	Wool	4
4058	897	Leather	5
4059	897	Wood	6
4060	897	Metal	7
4061	897	Plastic	8
4062	897	Glass	9
4063	897	Steel	10
4064	897	Stainless Steel	11
4065	897	Aluminum	12
4066	897	Brass	13
4067	897	Bronze	14
4068	897	Fabric	15
4069	897	Synthetic	16
4070	897	Nylon	17
4071	897	Canvas	18
4072	902	Cotton	1
4073	902	Polyester	2
4074	902	Silk	3
4075	902	Wool	4
4076	902	Leather	5
4077	902	Wood	6
4078	902	Metal	7
4079	902	Plastic	8
4080	902	Glass	9
4081	902	Steel	10
4082	902	Stainless Steel	11
4083	902	Aluminum	12
4084	902	Brass	13
4085	902	Bronze	14
4086	902	Fabric	15
4087	902	Synthetic	16
4088	902	Nylon	17
4089	902	Canvas	18
4090	907	Cotton	1
4091	907	Polyester	2
4092	907	Silk	3
4093	907	Wool	4
4094	907	Leather	5
4095	907	Wood	6
4096	907	Metal	7
4097	907	Plastic	8
4098	907	Glass	9
4099	907	Steel	10
4100	907	Stainless Steel	11
4101	907	Aluminum	12
4102	907	Brass	13
4103	907	Bronze	14
4104	907	Fabric	15
4105	907	Synthetic	16
4106	907	Nylon	17
4107	907	Canvas	18
4108	916	Cotton	1
4109	916	Polyester	2
4110	916	Silk	3
4111	916	Wool	4
4112	916	Leather	5
4113	916	Wood	6
4114	916	Metal	7
4115	916	Plastic	8
4116	916	Glass	9
4117	916	Steel	10
4118	916	Stainless Steel	11
4119	916	Aluminum	12
4120	916	Brass	13
4121	916	Bronze	14
4122	916	Fabric	15
4123	916	Synthetic	16
4124	916	Nylon	17
4125	916	Canvas	18
4126	921	Cotton	1
4127	921	Polyester	2
4128	921	Silk	3
4129	921	Wool	4
4130	921	Leather	5
4131	921	Wood	6
4132	921	Metal	7
4133	921	Plastic	8
4134	921	Glass	9
4135	921	Steel	10
4136	921	Stainless Steel	11
4137	921	Aluminum	12
4138	921	Brass	13
4139	921	Bronze	14
4140	921	Fabric	15
4141	921	Synthetic	16
4142	921	Nylon	17
4143	921	Canvas	18
4144	927	Cotton	1
4145	927	Polyester	2
4146	927	Silk	3
4147	927	Wool	4
4148	927	Leather	5
4149	927	Wood	6
4150	927	Metal	7
4151	927	Plastic	8
4152	927	Glass	9
4153	927	Steel	10
4154	927	Stainless Steel	11
4155	927	Aluminum	12
4156	927	Brass	13
4157	927	Bronze	14
4158	927	Fabric	15
4159	927	Synthetic	16
4160	927	Nylon	17
4161	927	Canvas	18
4162	932	Cotton	1
4163	932	Polyester	2
4164	932	Silk	3
4165	932	Wool	4
4166	932	Leather	5
4167	932	Wood	6
4168	932	Metal	7
4169	932	Plastic	8
4170	932	Glass	9
4171	932	Steel	10
4172	932	Stainless Steel	11
4173	932	Aluminum	12
4174	932	Brass	13
4175	932	Bronze	14
4176	932	Fabric	15
4177	932	Synthetic	16
4178	932	Nylon	17
4179	932	Canvas	18
4180	936	Cotton	1
4181	936	Polyester	2
4182	936	Silk	3
4183	936	Wool	4
4184	936	Leather	5
4185	936	Wood	6
4186	936	Metal	7
4187	936	Plastic	8
4188	936	Glass	9
4189	936	Steel	10
4190	936	Stainless Steel	11
4191	936	Aluminum	12
4192	936	Brass	13
4193	936	Bronze	14
4194	936	Fabric	15
4195	936	Synthetic	16
4196	936	Nylon	17
4197	936	Canvas	18
4198	942	Cotton	1
4199	942	Polyester	2
4200	942	Silk	3
4201	942	Wool	4
4202	942	Leather	5
4203	942	Wood	6
4204	942	Metal	7
4205	942	Plastic	8
4206	942	Glass	9
4207	942	Steel	10
4208	942	Stainless Steel	11
4209	942	Aluminum	12
4210	942	Brass	13
4211	942	Bronze	14
4212	942	Fabric	15
4213	942	Synthetic	16
4214	942	Nylon	17
4215	942	Canvas	18
4216	946	Cotton	1
4217	946	Polyester	2
4218	946	Silk	3
4219	946	Wool	4
4220	946	Leather	5
4221	946	Wood	6
4222	946	Metal	7
4223	946	Plastic	8
4224	946	Glass	9
4225	946	Steel	10
4226	946	Stainless Steel	11
4227	946	Aluminum	12
4228	946	Brass	13
4229	946	Bronze	14
4230	946	Fabric	15
4231	946	Synthetic	16
4232	946	Nylon	17
4233	946	Canvas	18
4234	951	Cotton	1
4235	951	Polyester	2
4236	951	Silk	3
4237	951	Wool	4
4238	951	Leather	5
4239	951	Wood	6
4240	951	Metal	7
4241	951	Plastic	8
4242	951	Glass	9
4243	951	Steel	10
4244	951	Stainless Steel	11
4245	951	Aluminum	12
4246	951	Brass	13
4247	951	Bronze	14
4248	951	Fabric	15
4249	951	Synthetic	16
4250	951	Nylon	17
4251	951	Canvas	18
4252	956	Cotton	1
4253	956	Polyester	2
4254	956	Silk	3
4255	956	Wool	4
4256	956	Leather	5
4257	956	Wood	6
4258	956	Metal	7
4259	956	Plastic	8
4260	956	Glass	9
4261	956	Steel	10
4262	956	Stainless Steel	11
4263	956	Aluminum	12
4264	956	Brass	13
4265	956	Bronze	14
4266	956	Fabric	15
4267	956	Synthetic	16
4268	956	Nylon	17
4269	956	Canvas	18
4270	961	Cotton	1
4271	961	Polyester	2
4272	961	Silk	3
4273	961	Wool	4
4274	961	Leather	5
4275	961	Wood	6
4276	961	Metal	7
4277	961	Plastic	8
4278	961	Glass	9
4279	961	Steel	10
4280	961	Stainless Steel	11
4281	961	Aluminum	12
4282	961	Brass	13
4283	961	Bronze	14
4284	961	Fabric	15
4285	961	Synthetic	16
4286	961	Nylon	17
4287	961	Canvas	18
4288	977	Cotton	1
4289	977	Polyester	2
4290	977	Silk	3
4291	977	Wool	4
4292	977	Leather	5
4293	977	Wood	6
4294	977	Metal	7
4295	977	Plastic	8
4296	977	Glass	9
4297	977	Steel	10
4298	977	Stainless Steel	11
4299	977	Aluminum	12
4300	977	Brass	13
4301	977	Bronze	14
4302	977	Fabric	15
4303	977	Synthetic	16
4304	977	Nylon	17
4305	977	Canvas	18
4306	981	Cotton	1
4307	981	Polyester	2
4308	981	Silk	3
4309	981	Wool	4
4310	981	Leather	5
4311	981	Wood	6
4312	981	Metal	7
4313	981	Plastic	8
4314	981	Glass	9
4315	981	Steel	10
4316	981	Stainless Steel	11
4317	981	Aluminum	12
4318	981	Brass	13
4319	981	Bronze	14
4320	981	Fabric	15
4321	981	Synthetic	16
4322	981	Nylon	17
4323	981	Canvas	18
4324	986	Cotton	1
4325	986	Polyester	2
4326	986	Silk	3
4327	986	Wool	4
4328	986	Leather	5
4329	986	Wood	6
4330	986	Metal	7
4331	986	Plastic	8
4332	986	Glass	9
4333	986	Steel	10
4334	986	Stainless Steel	11
4335	986	Aluminum	12
4336	986	Brass	13
4337	986	Bronze	14
4338	986	Fabric	15
4339	986	Synthetic	16
4340	986	Nylon	17
4341	986	Canvas	18
4342	994	Cotton	1
4343	994	Polyester	2
4344	994	Silk	3
4345	994	Wool	4
4346	994	Leather	5
4347	994	Wood	6
4348	994	Metal	7
4349	994	Plastic	8
4350	994	Glass	9
4351	994	Steel	10
4352	994	Stainless Steel	11
4353	994	Aluminum	12
4354	994	Brass	13
4355	994	Bronze	14
4356	994	Fabric	15
4357	994	Synthetic	16
4358	994	Nylon	17
4359	994	Canvas	18
4360	1018	Cotton	1
4361	1018	Polyester	2
4362	1018	Silk	3
4363	1018	Wool	4
4364	1018	Leather	5
4365	1018	Wood	6
4366	1018	Metal	7
4367	1018	Plastic	8
4368	1018	Glass	9
4369	1018	Steel	10
4370	1018	Stainless Steel	11
4371	1018	Aluminum	12
4372	1018	Brass	13
4373	1018	Bronze	14
4374	1018	Fabric	15
4375	1018	Synthetic	16
4376	1018	Nylon	17
4377	1018	Canvas	18
4378	1023	Cotton	1
4379	1023	Polyester	2
4380	1023	Silk	3
4381	1023	Wool	4
4382	1023	Leather	5
4383	1023	Wood	6
4384	1023	Metal	7
4385	1023	Plastic	8
4386	1023	Glass	9
4387	1023	Steel	10
4388	1023	Stainless Steel	11
4389	1023	Aluminum	12
4390	1023	Brass	13
4391	1023	Bronze	14
4392	1023	Fabric	15
4393	1023	Synthetic	16
4394	1023	Nylon	17
4395	1023	Canvas	18
4396	1029	Cotton	1
4397	1029	Polyester	2
4398	1029	Silk	3
4399	1029	Wool	4
4400	1029	Leather	5
4401	1029	Wood	6
4402	1029	Metal	7
4403	1029	Plastic	8
4404	1029	Glass	9
4405	1029	Steel	10
4406	1029	Stainless Steel	11
4407	1029	Aluminum	12
4408	1029	Brass	13
4409	1029	Bronze	14
4410	1029	Fabric	15
4411	1029	Synthetic	16
4412	1029	Nylon	17
4413	1029	Canvas	18
4414	1033	Cotton	1
4415	1033	Polyester	2
4416	1033	Silk	3
4417	1033	Wool	4
4418	1033	Leather	5
4419	1033	Wood	6
4420	1033	Metal	7
4421	1033	Plastic	8
4422	1033	Glass	9
4423	1033	Steel	10
4424	1033	Stainless Steel	11
4425	1033	Aluminum	12
4426	1033	Brass	13
4427	1033	Bronze	14
4428	1033	Fabric	15
4429	1033	Synthetic	16
4430	1033	Nylon	17
4431	1033	Canvas	18
4432	1040	Cotton	1
4433	1040	Polyester	2
4434	1040	Silk	3
4435	1040	Wool	4
4436	1040	Leather	5
4437	1040	Wood	6
4438	1040	Metal	7
4439	1040	Plastic	8
4440	1040	Glass	9
4441	1040	Steel	10
4442	1040	Stainless Steel	11
4443	1040	Aluminum	12
4444	1040	Brass	13
4445	1040	Bronze	14
4446	1040	Fabric	15
4447	1040	Synthetic	16
4448	1040	Nylon	17
4449	1040	Canvas	18
4450	1045	Cotton	1
4451	1045	Polyester	2
4452	1045	Silk	3
4453	1045	Wool	4
4454	1045	Leather	5
4455	1045	Wood	6
4456	1045	Metal	7
4457	1045	Plastic	8
4458	1045	Glass	9
4459	1045	Steel	10
4460	1045	Stainless Steel	11
4461	1045	Aluminum	12
4462	1045	Brass	13
4463	1045	Bronze	14
4464	1045	Fabric	15
4465	1045	Synthetic	16
4466	1045	Nylon	17
4467	1045	Canvas	18
4468	1049	Cotton	1
4469	1049	Polyester	2
4470	1049	Silk	3
4471	1049	Wool	4
4472	1049	Leather	5
4473	1049	Wood	6
4474	1049	Metal	7
4475	1049	Plastic	8
4476	1049	Glass	9
4477	1049	Steel	10
4478	1049	Stainless Steel	11
4479	1049	Aluminum	12
4480	1049	Brass	13
4481	1049	Bronze	14
4482	1049	Fabric	15
4483	1049	Synthetic	16
4484	1049	Nylon	17
4485	1049	Canvas	18
4486	1054	Cotton	1
4487	1054	Polyester	2
4488	1054	Silk	3
4489	1054	Wool	4
4490	1054	Leather	5
4491	1054	Wood	6
4492	1054	Metal	7
4493	1054	Plastic	8
4494	1054	Glass	9
4495	1054	Steel	10
4496	1054	Stainless Steel	11
4497	1054	Aluminum	12
4498	1054	Brass	13
4499	1054	Bronze	14
4500	1054	Fabric	15
4501	1054	Synthetic	16
4502	1054	Nylon	17
4503	1054	Canvas	18
4504	1060	Cotton	1
4505	1060	Polyester	2
4506	1060	Silk	3
4507	1060	Wool	4
4508	1060	Leather	5
4509	1060	Wood	6
4510	1060	Metal	7
4511	1060	Plastic	8
4512	1060	Glass	9
4513	1060	Steel	10
4514	1060	Stainless Steel	11
4515	1060	Aluminum	12
4516	1060	Brass	13
4517	1060	Bronze	14
4518	1060	Fabric	15
4519	1060	Synthetic	16
4520	1060	Nylon	17
4521	1060	Canvas	18
4522	1064	Cotton	1
4523	1064	Polyester	2
4524	1064	Silk	3
4525	1064	Wool	4
4526	1064	Leather	5
4527	1064	Wood	6
4528	1064	Metal	7
4529	1064	Plastic	8
4530	1064	Glass	9
4531	1064	Steel	10
4532	1064	Stainless Steel	11
4533	1064	Aluminum	12
4534	1064	Brass	13
4535	1064	Bronze	14
4536	1064	Fabric	15
4537	1064	Synthetic	16
4538	1064	Nylon	17
4539	1064	Canvas	18
4540	1069	Cotton	1
4541	1069	Polyester	2
4542	1069	Silk	3
4543	1069	Wool	4
4544	1069	Leather	5
4545	1069	Wood	6
4546	1069	Metal	7
4547	1069	Plastic	8
4548	1069	Glass	9
4549	1069	Steel	10
4550	1069	Stainless Steel	11
4551	1069	Aluminum	12
4552	1069	Brass	13
4553	1069	Bronze	14
4554	1069	Fabric	15
4555	1069	Synthetic	16
4556	1069	Nylon	17
4557	1069	Canvas	18
4558	1074	Cotton	1
4559	1074	Polyester	2
4560	1074	Silk	3
4561	1074	Wool	4
4562	1074	Leather	5
4563	1074	Wood	6
4564	1074	Metal	7
4565	1074	Plastic	8
4566	1074	Glass	9
4567	1074	Steel	10
4568	1074	Stainless Steel	11
4569	1074	Aluminum	12
4570	1074	Brass	13
4571	1074	Bronze	14
4572	1074	Fabric	15
4573	1074	Synthetic	16
4574	1074	Nylon	17
4575	1074	Canvas	18
4576	1079	Cotton	1
4577	1079	Polyester	2
4578	1079	Silk	3
4579	1079	Wool	4
4580	1079	Leather	5
4581	1079	Wood	6
4582	1079	Metal	7
4583	1079	Plastic	8
4584	1079	Glass	9
4585	1079	Steel	10
4586	1079	Stainless Steel	11
4587	1079	Aluminum	12
4588	1079	Brass	13
4589	1079	Bronze	14
4590	1079	Fabric	15
4591	1079	Synthetic	16
4592	1079	Nylon	17
4593	1079	Canvas	18
4594	1084	Cotton	1
4595	1084	Polyester	2
4596	1084	Silk	3
4597	1084	Wool	4
4598	1084	Leather	5
4599	1084	Wood	6
4600	1084	Metal	7
4601	1084	Plastic	8
4602	1084	Glass	9
4603	1084	Steel	10
4604	1084	Stainless Steel	11
4605	1084	Aluminum	12
4606	1084	Brass	13
4607	1084	Bronze	14
4608	1084	Fabric	15
4609	1084	Synthetic	16
4610	1084	Nylon	17
4611	1084	Canvas	18
4612	1090	Cotton	1
4613	1090	Polyester	2
4614	1090	Silk	3
4615	1090	Wool	4
4616	1090	Leather	5
4617	1090	Wood	6
4618	1090	Metal	7
4619	1090	Plastic	8
4620	1090	Glass	9
4621	1090	Steel	10
4622	1090	Stainless Steel	11
4623	1090	Aluminum	12
4624	1090	Brass	13
4625	1090	Bronze	14
4626	1090	Fabric	15
4627	1090	Synthetic	16
4628	1090	Nylon	17
4629	1090	Canvas	18
4630	1097	Cotton	1
4631	1097	Polyester	2
4632	1097	Silk	3
4633	1097	Wool	4
4634	1097	Leather	5
4635	1097	Wood	6
4636	1097	Metal	7
4637	1097	Plastic	8
4638	1097	Glass	9
4639	1097	Steel	10
4640	1097	Stainless Steel	11
4641	1097	Aluminum	12
4642	1097	Brass	13
4643	1097	Bronze	14
4644	1097	Fabric	15
4645	1097	Synthetic	16
4646	1097	Nylon	17
4647	1097	Canvas	18
4648	1101	Cotton	1
4649	1101	Polyester	2
4650	1101	Silk	3
4651	1101	Wool	4
4652	1101	Leather	5
4653	1101	Wood	6
4654	1101	Metal	7
4655	1101	Plastic	8
4656	1101	Glass	9
4657	1101	Steel	10
4658	1101	Stainless Steel	11
4659	1101	Aluminum	12
4660	1101	Brass	13
4661	1101	Bronze	14
4662	1101	Fabric	15
4663	1101	Synthetic	16
4664	1101	Nylon	17
4665	1101	Canvas	18
4666	1106	Cotton	1
4667	1106	Polyester	2
4668	1106	Silk	3
4669	1106	Wool	4
4670	1106	Leather	5
4671	1106	Wood	6
4672	1106	Metal	7
4673	1106	Plastic	8
4674	1106	Glass	9
4675	1106	Steel	10
4676	1106	Stainless Steel	11
4677	1106	Aluminum	12
4678	1106	Brass	13
4679	1106	Bronze	14
4680	1106	Fabric	15
4681	1106	Synthetic	16
4682	1106	Nylon	17
4683	1106	Canvas	18
4684	1111	Cotton	1
4685	1111	Polyester	2
4686	1111	Silk	3
4687	1111	Wool	4
4688	1111	Leather	5
4689	1111	Wood	6
4690	1111	Metal	7
4691	1111	Plastic	8
4692	1111	Glass	9
4693	1111	Steel	10
4694	1111	Stainless Steel	11
4695	1111	Aluminum	12
4696	1111	Brass	13
4697	1111	Bronze	14
4698	1111	Fabric	15
4699	1111	Synthetic	16
4700	1111	Nylon	17
4701	1111	Canvas	18
4702	1116	Cotton	1
4703	1116	Polyester	2
4704	1116	Silk	3
4705	1116	Wool	4
4706	1116	Leather	5
4707	1116	Wood	6
4708	1116	Metal	7
4709	1116	Plastic	8
4710	1116	Glass	9
4711	1116	Steel	10
4712	1116	Stainless Steel	11
4713	1116	Aluminum	12
4714	1116	Brass	13
4715	1116	Bronze	14
4716	1116	Fabric	15
4717	1116	Synthetic	16
4718	1116	Nylon	17
4719	1116	Canvas	18
4720	1120	Cotton	1
4721	1120	Polyester	2
4722	1120	Silk	3
4723	1120	Wool	4
4724	1120	Leather	5
4725	1120	Wood	6
4726	1120	Metal	7
4727	1120	Plastic	8
4728	1120	Glass	9
4729	1120	Steel	10
4730	1120	Stainless Steel	11
4731	1120	Aluminum	12
4732	1120	Brass	13
4733	1120	Bronze	14
4734	1120	Fabric	15
4735	1120	Synthetic	16
4736	1120	Nylon	17
4737	1120	Canvas	18
4738	1124	Cotton	1
4739	1124	Polyester	2
4740	1124	Silk	3
4741	1124	Wool	4
4742	1124	Leather	5
4743	1124	Wood	6
4744	1124	Metal	7
4745	1124	Plastic	8
4746	1124	Glass	9
4747	1124	Steel	10
4748	1124	Stainless Steel	11
4749	1124	Aluminum	12
4750	1124	Brass	13
4751	1124	Bronze	14
4752	1124	Fabric	15
4753	1124	Synthetic	16
4754	1124	Nylon	17
4755	1124	Canvas	18
4756	1129	Cotton	1
4757	1129	Polyester	2
4758	1129	Silk	3
4759	1129	Wool	4
4760	1129	Leather	5
4761	1129	Wood	6
4762	1129	Metal	7
4763	1129	Plastic	8
4764	1129	Glass	9
4765	1129	Steel	10
4766	1129	Stainless Steel	11
4767	1129	Aluminum	12
4768	1129	Brass	13
4769	1129	Bronze	14
4770	1129	Fabric	15
4771	1129	Synthetic	16
4772	1129	Nylon	17
4773	1129	Canvas	18
4774	1134	Cotton	1
4775	1134	Polyester	2
4776	1134	Silk	3
4777	1134	Wool	4
4778	1134	Leather	5
4779	1134	Wood	6
4780	1134	Metal	7
4781	1134	Plastic	8
4782	1134	Glass	9
4783	1134	Steel	10
4784	1134	Stainless Steel	11
4785	1134	Aluminum	12
4786	1134	Brass	13
4787	1134	Bronze	14
4788	1134	Fabric	15
4789	1134	Synthetic	16
4790	1134	Nylon	17
4791	1134	Canvas	18
4792	1139	Cotton	1
4793	1139	Polyester	2
4794	1139	Silk	3
4795	1139	Wool	4
4796	1139	Leather	5
4797	1139	Wood	6
4798	1139	Metal	7
4799	1139	Plastic	8
4800	1139	Glass	9
4801	1139	Steel	10
4802	1139	Stainless Steel	11
4803	1139	Aluminum	12
4804	1139	Brass	13
4805	1139	Bronze	14
4806	1139	Fabric	15
4807	1139	Synthetic	16
4808	1139	Nylon	17
4809	1139	Canvas	18
4810	1144	Cotton	1
4811	1144	Polyester	2
4812	1144	Silk	3
4813	1144	Wool	4
4814	1144	Leather	5
4815	1144	Wood	6
4816	1144	Metal	7
4817	1144	Plastic	8
4818	1144	Glass	9
4819	1144	Steel	10
4820	1144	Stainless Steel	11
4821	1144	Aluminum	12
4822	1144	Brass	13
4823	1144	Bronze	14
4824	1144	Fabric	15
4825	1144	Synthetic	16
4826	1144	Nylon	17
4827	1144	Canvas	18
4828	1149	Cotton	1
4829	1149	Polyester	2
4830	1149	Silk	3
4831	1149	Wool	4
4832	1149	Leather	5
4833	1149	Wood	6
4834	1149	Metal	7
4835	1149	Plastic	8
4836	1149	Glass	9
4837	1149	Steel	10
4838	1149	Stainless Steel	11
4839	1149	Aluminum	12
4840	1149	Brass	13
4841	1149	Bronze	14
4842	1149	Fabric	15
4843	1149	Synthetic	16
4844	1149	Nylon	17
4845	1149	Canvas	18
4846	1154	Cotton	1
4847	1154	Polyester	2
4848	1154	Silk	3
4849	1154	Wool	4
4850	1154	Leather	5
4851	1154	Wood	6
4852	1154	Metal	7
4853	1154	Plastic	8
4854	1154	Glass	9
4855	1154	Steel	10
4856	1154	Stainless Steel	11
4857	1154	Aluminum	12
4858	1154	Brass	13
4859	1154	Bronze	14
4860	1154	Fabric	15
4861	1154	Synthetic	16
4862	1154	Nylon	17
4863	1154	Canvas	18
4864	1159	Cotton	1
4865	1159	Polyester	2
4866	1159	Silk	3
4867	1159	Wool	4
4868	1159	Leather	5
4869	1159	Wood	6
4870	1159	Metal	7
4871	1159	Plastic	8
4872	1159	Glass	9
4873	1159	Steel	10
4874	1159	Stainless Steel	11
4875	1159	Aluminum	12
4876	1159	Brass	13
4877	1159	Bronze	14
4878	1159	Fabric	15
4879	1159	Synthetic	16
4880	1159	Nylon	17
4881	1159	Canvas	18
4882	1164	Cotton	1
4883	1164	Polyester	2
4884	1164	Silk	3
4885	1164	Wool	4
4886	1164	Leather	5
4887	1164	Wood	6
4888	1164	Metal	7
4889	1164	Plastic	8
4890	1164	Glass	9
4891	1164	Steel	10
4892	1164	Stainless Steel	11
4893	1164	Aluminum	12
4894	1164	Brass	13
4895	1164	Bronze	14
4896	1164	Fabric	15
4897	1164	Synthetic	16
4898	1164	Nylon	17
4899	1164	Canvas	18
4900	1169	Cotton	1
4901	1169	Polyester	2
4902	1169	Silk	3
4903	1169	Wool	4
4904	1169	Leather	5
4905	1169	Wood	6
4906	1169	Metal	7
4907	1169	Plastic	8
4908	1169	Glass	9
4909	1169	Steel	10
4910	1169	Stainless Steel	11
4911	1169	Aluminum	12
4912	1169	Brass	13
4913	1169	Bronze	14
4914	1169	Fabric	15
4915	1169	Synthetic	16
4916	1169	Nylon	17
4917	1169	Canvas	18
4918	1431	Cotton	1
4919	1431	Polyester	2
4920	1431	Silk	3
4921	1431	Wool	4
4922	1431	Leather	5
4923	1431	Wood	6
4924	1431	Metal	7
4925	1431	Plastic	8
4926	1431	Glass	9
4927	1431	Steel	10
4928	1431	Stainless Steel	11
4929	1431	Aluminum	12
4930	1431	Brass	13
4931	1431	Bronze	14
4932	1431	Fabric	15
4933	1431	Synthetic	16
4934	1431	Nylon	17
4935	1431	Canvas	18
4936	1441	Cotton	1
4937	1441	Polyester	2
4938	1441	Silk	3
4939	1441	Wool	4
4940	1441	Leather	5
4941	1441	Wood	6
4942	1441	Metal	7
4943	1441	Plastic	8
4944	1441	Glass	9
4945	1441	Steel	10
4946	1441	Stainless Steel	11
4947	1441	Aluminum	12
4948	1441	Brass	13
4949	1441	Bronze	14
4950	1441	Fabric	15
4951	1441	Synthetic	16
4952	1441	Nylon	17
4953	1441	Canvas	18
4954	1446	Cotton	1
4955	1446	Polyester	2
4956	1446	Silk	3
4957	1446	Wool	4
4958	1446	Leather	5
4959	1446	Wood	6
4960	1446	Metal	7
4961	1446	Plastic	8
4962	1446	Glass	9
4963	1446	Steel	10
4964	1446	Stainless Steel	11
4965	1446	Aluminum	12
4966	1446	Brass	13
4967	1446	Bronze	14
4968	1446	Fabric	15
4969	1446	Synthetic	16
4970	1446	Nylon	17
4971	1446	Canvas	18
4972	1598	Cotton	1
4973	1598	Polyester	2
4974	1598	Silk	3
4975	1598	Wool	4
4976	1598	Leather	5
4977	1598	Wood	6
4978	1598	Metal	7
4979	1598	Plastic	8
4980	1598	Glass	9
4981	1598	Steel	10
4982	1598	Stainless Steel	11
4983	1598	Aluminum	12
4984	1598	Brass	13
4985	1598	Bronze	14
4986	1598	Fabric	15
4987	1598	Synthetic	16
4988	1598	Nylon	17
4989	1598	Canvas	18
4990	1902	Cotton	1
4991	1902	Polyester	2
4992	1902	Silk	3
4993	1902	Wool	4
4994	1902	Leather	5
4995	1902	Wood	6
4996	1902	Metal	7
4997	1902	Plastic	8
4998	1902	Glass	9
4999	1902	Steel	10
5000	1902	Stainless Steel	11
5001	1902	Aluminum	12
5002	1902	Brass	13
5003	1902	Bronze	14
5004	1902	Fabric	15
5005	1902	Synthetic	16
5006	1902	Nylon	17
5007	1902	Canvas	18
5008	1922	Cotton	1
5009	1922	Polyester	2
5010	1922	Silk	3
5011	1922	Wool	4
5012	1922	Leather	5
5013	1922	Wood	6
5014	1922	Metal	7
5015	1922	Plastic	8
5016	1922	Glass	9
5017	1922	Steel	10
5018	1922	Stainless Steel	11
5019	1922	Aluminum	12
5020	1922	Brass	13
5021	1922	Bronze	14
5022	1922	Fabric	15
5023	1922	Synthetic	16
5024	1922	Nylon	17
5025	1922	Canvas	18
5026	1973	Cotton	1
5027	1973	Polyester	2
5028	1973	Silk	3
5029	1973	Wool	4
5030	1973	Leather	5
5031	1973	Wood	6
5032	1973	Metal	7
5033	1973	Plastic	8
5034	1973	Glass	9
5035	1973	Steel	10
5036	1973	Stainless Steel	11
5037	1973	Aluminum	12
5038	1973	Brass	13
5039	1973	Bronze	14
5040	1973	Fabric	15
5041	1973	Synthetic	16
5042	1973	Nylon	17
5043	1973	Canvas	18
5044	1977	Cotton	1
5045	1977	Polyester	2
5046	1977	Silk	3
5047	1977	Wool	4
5048	1977	Leather	5
5049	1977	Wood	6
5050	1977	Metal	7
5051	1977	Plastic	8
5052	1977	Glass	9
5053	1977	Steel	10
5054	1977	Stainless Steel	11
5055	1977	Aluminum	12
5056	1977	Brass	13
5057	1977	Bronze	14
5058	1977	Fabric	15
5059	1977	Synthetic	16
5060	1977	Nylon	17
5061	1977	Canvas	18
5062	1981	Cotton	1
5063	1981	Polyester	2
5064	1981	Silk	3
5065	1981	Wool	4
5066	1981	Leather	5
5067	1981	Wood	6
5068	1981	Metal	7
5069	1981	Plastic	8
5070	1981	Glass	9
5071	1981	Steel	10
5072	1981	Stainless Steel	11
5073	1981	Aluminum	12
5074	1981	Brass	13
5075	1981	Bronze	14
5076	1981	Fabric	15
5077	1981	Synthetic	16
5078	1981	Nylon	17
5079	1981	Canvas	18
5080	1985	Cotton	1
5081	1985	Polyester	2
5082	1985	Silk	3
5083	1985	Wool	4
5084	1985	Leather	5
5085	1985	Wood	6
5086	1985	Metal	7
5087	1985	Plastic	8
5088	1985	Glass	9
5089	1985	Steel	10
5090	1985	Stainless Steel	11
5091	1985	Aluminum	12
5092	1985	Brass	13
5093	1985	Bronze	14
5094	1985	Fabric	15
5095	1985	Synthetic	16
5096	1985	Nylon	17
5097	1985	Canvas	18
5098	1997	Cotton	1
5099	1997	Polyester	2
5100	1997	Silk	3
5101	1997	Wool	4
5102	1997	Leather	5
5103	1997	Wood	6
5104	1997	Metal	7
5105	1997	Plastic	8
5106	1997	Glass	9
5107	1997	Steel	10
5108	1997	Stainless Steel	11
5109	1997	Aluminum	12
5110	1997	Brass	13
5111	1997	Bronze	14
5112	1997	Fabric	15
5113	1997	Synthetic	16
5114	1997	Nylon	17
5115	1997	Canvas	18
5116	2004	Cotton	1
5117	2004	Polyester	2
5118	2004	Silk	3
5119	2004	Wool	4
5120	2004	Leather	5
5121	2004	Wood	6
5122	2004	Metal	7
5123	2004	Plastic	8
5124	2004	Glass	9
5125	2004	Steel	10
5126	2004	Stainless Steel	11
5127	2004	Aluminum	12
5128	2004	Brass	13
5129	2004	Bronze	14
5130	2004	Fabric	15
5131	2004	Synthetic	16
5132	2004	Nylon	17
5133	2004	Canvas	18
5134	2055	Cotton	1
5135	2055	Polyester	2
5136	2055	Silk	3
5137	2055	Wool	4
5138	2055	Leather	5
5139	2055	Wood	6
5140	2055	Metal	7
5141	2055	Plastic	8
5142	2055	Glass	9
5143	2055	Steel	10
5144	2055	Stainless Steel	11
5145	2055	Aluminum	12
5146	2055	Brass	13
5147	2055	Bronze	14
5148	2055	Fabric	15
5149	2055	Synthetic	16
5150	2055	Nylon	17
5151	2055	Canvas	18
5152	2147	Cotton	1
5153	2147	Polyester	2
5154	2147	Silk	3
5155	2147	Wool	4
5156	2147	Leather	5
5157	2147	Wood	6
5158	2147	Metal	7
5159	2147	Plastic	8
5160	2147	Glass	9
5161	2147	Steel	10
5162	2147	Stainless Steel	11
5163	2147	Aluminum	12
5164	2147	Brass	13
5165	2147	Bronze	14
5166	2147	Fabric	15
5167	2147	Synthetic	16
5168	2147	Nylon	17
5169	2147	Canvas	18
5170	2152	Cotton	1
5171	2152	Polyester	2
5172	2152	Silk	3
5173	2152	Wool	4
5174	2152	Leather	5
5175	2152	Wood	6
5176	2152	Metal	7
5177	2152	Plastic	8
5178	2152	Glass	9
5179	2152	Steel	10
5180	2152	Stainless Steel	11
5181	2152	Aluminum	12
5182	2152	Brass	13
5183	2152	Bronze	14
5184	2152	Fabric	15
5185	2152	Synthetic	16
5186	2152	Nylon	17
5187	2152	Canvas	18
5188	2156	Cotton	1
5189	2156	Polyester	2
5190	2156	Silk	3
5191	2156	Wool	4
5192	2156	Leather	5
5193	2156	Wood	6
5194	2156	Metal	7
5195	2156	Plastic	8
5196	2156	Glass	9
5197	2156	Steel	10
5198	2156	Stainless Steel	11
5199	2156	Aluminum	12
5200	2156	Brass	13
5201	2156	Bronze	14
5202	2156	Fabric	15
5203	2156	Synthetic	16
5204	2156	Nylon	17
5205	2156	Canvas	18
5206	2160	Cotton	1
5207	2160	Polyester	2
5208	2160	Silk	3
5209	2160	Wool	4
5210	2160	Leather	5
5211	2160	Wood	6
5212	2160	Metal	7
5213	2160	Plastic	8
5214	2160	Glass	9
5215	2160	Steel	10
5216	2160	Stainless Steel	11
5217	2160	Aluminum	12
5218	2160	Brass	13
5219	2160	Bronze	14
5220	2160	Fabric	15
5221	2160	Synthetic	16
5222	2160	Nylon	17
5223	2160	Canvas	18
5224	2233	Cotton	1
5225	2233	Polyester	2
5226	2233	Silk	3
5227	2233	Wool	4
5228	2233	Leather	5
5229	2233	Wood	6
5230	2233	Metal	7
5231	2233	Plastic	8
5232	2233	Glass	9
5233	2233	Steel	10
5234	2233	Stainless Steel	11
5235	2233	Aluminum	12
5236	2233	Brass	13
5237	2233	Bronze	14
5238	2233	Fabric	15
5239	2233	Synthetic	16
5240	2233	Nylon	17
5241	2233	Canvas	18
5242	2237	Cotton	1
5243	2237	Polyester	2
5244	2237	Silk	3
5245	2237	Wool	4
5246	2237	Leather	5
5247	2237	Wood	6
5248	2237	Metal	7
5249	2237	Plastic	8
5250	2237	Glass	9
5251	2237	Steel	10
5252	2237	Stainless Steel	11
5253	2237	Aluminum	12
5254	2237	Brass	13
5255	2237	Bronze	14
5256	2237	Fabric	15
5257	2237	Synthetic	16
5258	2237	Nylon	17
5259	2237	Canvas	18
5260	2445	Cotton	1
5261	2445	Polyester	2
5262	2445	Silk	3
5263	2445	Wool	4
5264	2445	Leather	5
5265	2445	Wood	6
5266	2445	Metal	7
5267	2445	Plastic	8
5268	2445	Glass	9
5269	2445	Steel	10
5270	2445	Stainless Steel	11
5271	2445	Aluminum	12
5272	2445	Brass	13
5273	2445	Bronze	14
5274	2445	Fabric	15
5275	2445	Synthetic	16
5276	2445	Nylon	17
5277	2445	Canvas	18
5278	2450	Cotton	1
5279	2450	Polyester	2
5280	2450	Silk	3
5281	2450	Wool	4
5282	2450	Leather	5
5283	2450	Wood	6
5284	2450	Metal	7
5285	2450	Plastic	8
5286	2450	Glass	9
5287	2450	Steel	10
5288	2450	Stainless Steel	11
5289	2450	Aluminum	12
5290	2450	Brass	13
5291	2450	Bronze	14
5292	2450	Fabric	15
5293	2450	Synthetic	16
5294	2450	Nylon	17
5295	2450	Canvas	18
5296	2460	Cotton	1
5297	2460	Polyester	2
5298	2460	Silk	3
5299	2460	Wool	4
5300	2460	Leather	5
5301	2460	Wood	6
5302	2460	Metal	7
5303	2460	Plastic	8
5304	2460	Glass	9
5305	2460	Steel	10
5306	2460	Stainless Steel	11
5307	2460	Aluminum	12
5308	2460	Brass	13
5309	2460	Bronze	14
5310	2460	Fabric	15
5311	2460	Synthetic	16
5312	2460	Nylon	17
5313	2460	Canvas	18
5314	2466	Cotton	1
5315	2466	Polyester	2
5316	2466	Silk	3
5317	2466	Wool	4
5318	2466	Leather	5
5319	2466	Wood	6
5320	2466	Metal	7
5321	2466	Plastic	8
5322	2466	Glass	9
5323	2466	Steel	10
5324	2466	Stainless Steel	11
5325	2466	Aluminum	12
5326	2466	Brass	13
5327	2466	Bronze	14
5328	2466	Fabric	15
5329	2466	Synthetic	16
5330	2466	Nylon	17
5331	2466	Canvas	18
5332	2471	Cotton	1
5333	2471	Polyester	2
5334	2471	Silk	3
5335	2471	Wool	4
5336	2471	Leather	5
5337	2471	Wood	6
5338	2471	Metal	7
5339	2471	Plastic	8
5340	2471	Glass	9
5341	2471	Steel	10
5342	2471	Stainless Steel	11
5343	2471	Aluminum	12
5344	2471	Brass	13
5345	2471	Bronze	14
5346	2471	Fabric	15
5347	2471	Synthetic	16
5348	2471	Nylon	17
5349	2471	Canvas	18
5350	2474	Cotton	1
5351	2474	Polyester	2
5352	2474	Silk	3
5353	2474	Wool	4
5354	2474	Leather	5
5355	2474	Wood	6
5356	2474	Metal	7
5357	2474	Plastic	8
5358	2474	Glass	9
5359	2474	Steel	10
5360	2474	Stainless Steel	11
5361	2474	Aluminum	12
5362	2474	Brass	13
5363	2474	Bronze	14
5364	2474	Fabric	15
5365	2474	Synthetic	16
5366	2474	Nylon	17
5367	2474	Canvas	18
5368	2479	Cotton	1
5369	2479	Polyester	2
5370	2479	Silk	3
5371	2479	Wool	4
5372	2479	Leather	5
5373	2479	Wood	6
5374	2479	Metal	7
5375	2479	Plastic	8
5376	2479	Glass	9
5377	2479	Steel	10
5378	2479	Stainless Steel	11
5379	2479	Aluminum	12
5380	2479	Brass	13
5381	2479	Bronze	14
5382	2479	Fabric	15
5383	2479	Synthetic	16
5384	2479	Nylon	17
5385	2479	Canvas	18
5386	2484	Cotton	1
5387	2484	Polyester	2
5388	2484	Silk	3
5389	2484	Wool	4
5390	2484	Leather	5
5391	2484	Wood	6
5392	2484	Metal	7
5393	2484	Plastic	8
5394	2484	Glass	9
5395	2484	Steel	10
5396	2484	Stainless Steel	11
5397	2484	Aluminum	12
5398	2484	Brass	13
5399	2484	Bronze	14
5400	2484	Fabric	15
5401	2484	Synthetic	16
5402	2484	Nylon	17
5403	2484	Canvas	18
5404	2499	Cotton	1
5405	2499	Polyester	2
5406	2499	Silk	3
5407	2499	Wool	4
5408	2499	Leather	5
5409	2499	Wood	6
5410	2499	Metal	7
5411	2499	Plastic	8
5412	2499	Glass	9
5413	2499	Steel	10
5414	2499	Stainless Steel	11
5415	2499	Aluminum	12
5416	2499	Brass	13
5417	2499	Bronze	14
5418	2499	Fabric	15
5419	2499	Synthetic	16
5420	2499	Nylon	17
5421	2499	Canvas	18
5422	2070	Cotton	1
5423	2070	Polyester	2
5424	2070	Silk	3
5425	2070	Linen	4
5426	2070	Wool	5
5427	2070	Rayon	6
5428	2070	Denim	7
5429	2070	Velvet	8
5430	2070	Chiffon	9
5431	2070	Georgette	10
5432	2070	Crepe	11
5433	2070	Satin	12
5434	2070	Net	13
5435	2070	Jersey	14
5436	2070	Lycra	15
5437	2076	Cotton	1
5438	2076	Polyester	2
5439	2076	Silk	3
5440	2076	Linen	4
5441	2076	Wool	5
5442	2076	Rayon	6
5443	2076	Denim	7
5444	2076	Velvet	8
5445	2076	Chiffon	9
5446	2076	Georgette	10
5447	2076	Crepe	11
5448	2076	Satin	12
5449	2076	Net	13
5450	2076	Jersey	14
5451	2076	Lycra	15
5452	2082	Cotton	1
5453	2082	Polyester	2
5454	2082	Silk	3
5455	2082	Linen	4
5456	2082	Wool	5
5457	2082	Rayon	6
5458	2082	Denim	7
5459	2082	Velvet	8
5460	2082	Chiffon	9
5461	2082	Georgette	10
5462	2082	Crepe	11
5463	2082	Satin	12
5464	2082	Net	13
5465	2082	Jersey	14
5466	2082	Lycra	15
5467	2088	Cotton	1
5468	2088	Polyester	2
5469	2088	Silk	3
5470	2088	Linen	4
5471	2088	Wool	5
5472	2088	Rayon	6
5473	2088	Denim	7
5474	2088	Velvet	8
5475	2088	Chiffon	9
5476	2088	Georgette	10
5477	2088	Crepe	11
5478	2088	Satin	12
5479	2088	Net	13
5480	2088	Jersey	14
5481	2088	Lycra	15
5482	2094	Cotton	1
5483	2094	Polyester	2
5484	2094	Silk	3
5485	2094	Linen	4
5486	2094	Wool	5
5487	2094	Rayon	6
5488	2094	Denim	7
5489	2094	Velvet	8
5490	2094	Chiffon	9
5491	2094	Georgette	10
5492	2094	Crepe	11
5493	2094	Satin	12
5494	2094	Net	13
5495	2094	Jersey	14
5496	2094	Lycra	15
5497	2101	Cotton	1
5498	2101	Polyester	2
5499	2101	Silk	3
5500	2101	Linen	4
5501	2101	Wool	5
5502	2101	Rayon	6
5503	2101	Denim	7
5504	2101	Velvet	8
5505	2101	Chiffon	9
5506	2101	Georgette	10
5507	2101	Crepe	11
5508	2101	Satin	12
5509	2101	Net	13
5510	2101	Jersey	14
5511	2101	Lycra	15
5512	2113	Cotton	1
5513	2113	Polyester	2
5514	2113	Silk	3
5515	2113	Linen	4
5516	2113	Wool	5
5517	2113	Rayon	6
5518	2113	Denim	7
5519	2113	Velvet	8
5520	2113	Chiffon	9
5521	2113	Georgette	10
5522	2113	Crepe	11
5523	2113	Satin	12
5524	2113	Net	13
5525	2113	Jersey	14
5526	2113	Lycra	15
5527	2119	Cotton	1
5528	2119	Polyester	2
5529	2119	Silk	3
5530	2119	Linen	4
5531	2119	Wool	5
5532	2119	Rayon	6
5533	2119	Denim	7
5534	2119	Velvet	8
5535	2119	Chiffon	9
5536	2119	Georgette	10
5537	2119	Crepe	11
5538	2119	Satin	12
5539	2119	Net	13
5540	2119	Jersey	14
5541	2119	Lycra	15
5542	2125	Cotton	1
5543	2125	Polyester	2
5544	2125	Silk	3
5545	2125	Linen	4
5546	2125	Wool	5
5547	2125	Rayon	6
5548	2125	Denim	7
5549	2125	Velvet	8
5550	2125	Chiffon	9
5551	2125	Georgette	10
5552	2125	Crepe	11
5553	2125	Satin	12
5554	2125	Net	13
5555	2125	Jersey	14
5556	2125	Lycra	15
5557	2130	Cotton	1
5558	2130	Polyester	2
5559	2130	Silk	3
5560	2130	Linen	4
5561	2130	Wool	5
5562	2130	Rayon	6
5563	2130	Denim	7
5564	2130	Velvet	8
5565	2130	Chiffon	9
5566	2130	Georgette	10
5567	2130	Crepe	11
5568	2130	Satin	12
5569	2130	Net	13
5570	2130	Jersey	14
5571	2130	Lycra	15
5572	2136	Cotton	1
5573	2136	Polyester	2
5574	2136	Silk	3
5575	2136	Linen	4
5576	2136	Wool	5
5577	2136	Rayon	6
5578	2136	Denim	7
5579	2136	Velvet	8
5580	2136	Chiffon	9
5581	2136	Georgette	10
5582	2136	Crepe	11
5583	2136	Satin	12
5584	2136	Net	13
5585	2136	Jersey	14
5586	2136	Lycra	15
5587	2143	Cotton	1
5588	2143	Polyester	2
5589	2143	Silk	3
5590	2143	Linen	4
5591	2143	Wool	5
5592	2143	Rayon	6
5593	2143	Denim	7
5594	2143	Velvet	8
5595	2143	Chiffon	9
5596	2143	Georgette	10
5597	2143	Crepe	11
5598	2143	Satin	12
5599	2143	Net	13
5600	2143	Jersey	14
5601	2143	Lycra	15
5602	2166	Cotton	1
5603	2166	Polyester	2
5604	2166	Silk	3
5605	2166	Linen	4
5606	2166	Wool	5
5607	2166	Rayon	6
5608	2166	Denim	7
5609	2166	Velvet	8
5610	2166	Chiffon	9
5611	2166	Georgette	10
5612	2166	Crepe	11
5613	2166	Satin	12
5614	2166	Net	13
5615	2166	Jersey	14
5616	2166	Lycra	15
5617	2172	Cotton	1
5618	2172	Polyester	2
5619	2172	Silk	3
5620	2172	Linen	4
5621	2172	Wool	5
5622	2172	Rayon	6
5623	2172	Denim	7
5624	2172	Velvet	8
5625	2172	Chiffon	9
5626	2172	Georgette	10
5627	2172	Crepe	11
5628	2172	Satin	12
5629	2172	Net	13
5630	2172	Jersey	14
5631	2172	Lycra	15
5632	2178	Cotton	1
5633	2178	Polyester	2
5634	2178	Silk	3
5635	2178	Linen	4
5636	2178	Wool	5
5637	2178	Rayon	6
5638	2178	Denim	7
5639	2178	Velvet	8
5640	2178	Chiffon	9
5641	2178	Georgette	10
5642	2178	Crepe	11
5643	2178	Satin	12
5644	2178	Net	13
5645	2178	Jersey	14
5646	2178	Lycra	15
5647	2184	Cotton	1
5648	2184	Polyester	2
5649	2184	Silk	3
5650	2184	Linen	4
5651	2184	Wool	5
5652	2184	Rayon	6
5653	2184	Denim	7
5654	2184	Velvet	8
5655	2184	Chiffon	9
5656	2184	Georgette	10
5657	2184	Crepe	11
5658	2184	Satin	12
5659	2184	Net	13
5660	2184	Jersey	14
5661	2184	Lycra	15
5662	2190	Cotton	1
5663	2190	Polyester	2
5664	2190	Silk	3
5665	2190	Linen	4
5666	2190	Wool	5
5667	2190	Rayon	6
5668	2190	Denim	7
5669	2190	Velvet	8
5670	2190	Chiffon	9
5671	2190	Georgette	10
5672	2190	Crepe	11
5673	2190	Satin	12
5674	2190	Net	13
5675	2190	Jersey	14
5676	2190	Lycra	15
5677	2196	Cotton	1
5678	2196	Polyester	2
5679	2196	Silk	3
5680	2196	Linen	4
5681	2196	Wool	5
5682	2196	Rayon	6
5683	2196	Denim	7
5684	2196	Velvet	8
5685	2196	Chiffon	9
5686	2196	Georgette	10
5687	2196	Crepe	11
5688	2196	Satin	12
5689	2196	Net	13
5690	2196	Jersey	14
5691	2196	Lycra	15
5692	2202	Cotton	1
5693	2202	Polyester	2
5694	2202	Silk	3
5695	2202	Linen	4
5696	2202	Wool	5
5697	2202	Rayon	6
5698	2202	Denim	7
5699	2202	Velvet	8
5700	2202	Chiffon	9
5701	2202	Georgette	10
5702	2202	Crepe	11
5703	2202	Satin	12
5704	2202	Net	13
5705	2202	Jersey	14
5706	2202	Lycra	15
5707	2208	Cotton	1
5708	2208	Polyester	2
5709	2208	Silk	3
5710	2208	Linen	4
5711	2208	Wool	5
5712	2208	Rayon	6
5713	2208	Denim	7
5714	2208	Velvet	8
5715	2208	Chiffon	9
5716	2208	Georgette	10
5717	2208	Crepe	11
5718	2208	Satin	12
5719	2208	Net	13
5720	2208	Jersey	14
5721	2208	Lycra	15
5722	2214	Cotton	1
5723	2214	Polyester	2
5724	2214	Silk	3
5725	2214	Linen	4
5726	2214	Wool	5
5727	2214	Rayon	6
5728	2214	Denim	7
5729	2214	Velvet	8
5730	2214	Chiffon	9
5731	2214	Georgette	10
5732	2214	Crepe	11
5733	2214	Satin	12
5734	2214	Net	13
5735	2214	Jersey	14
5736	2214	Lycra	15
5737	2220	Cotton	1
5738	2220	Polyester	2
5739	2220	Silk	3
5740	2220	Linen	4
5741	2220	Wool	5
5742	2220	Rayon	6
5743	2220	Denim	7
5744	2220	Velvet	8
5745	2220	Chiffon	9
5746	2220	Georgette	10
5747	2220	Crepe	11
5748	2220	Satin	12
5749	2220	Net	13
5750	2220	Jersey	14
5751	2220	Lycra	15
5752	2226	Cotton	1
5753	2226	Polyester	2
5754	2226	Silk	3
5755	2226	Linen	4
5756	2226	Wool	5
5757	2226	Rayon	6
5758	2226	Denim	7
5759	2226	Velvet	8
5760	2226	Chiffon	9
5761	2226	Georgette	10
5762	2226	Crepe	11
5763	2226	Satin	12
5764	2226	Net	13
5765	2226	Jersey	14
5766	2226	Lycra	15
5767	2242	Cotton	1
5768	2242	Polyester	2
5769	2242	Silk	3
5770	2242	Linen	4
5771	2242	Wool	5
5772	2242	Rayon	6
5773	2242	Denim	7
5774	2242	Velvet	8
5775	2242	Chiffon	9
5776	2242	Georgette	10
5777	2242	Crepe	11
5778	2242	Satin	12
5779	2242	Net	13
5780	2242	Jersey	14
5781	2242	Lycra	15
5782	2247	Cotton	1
5783	2247	Polyester	2
5784	2247	Silk	3
5785	2247	Linen	4
5786	2247	Wool	5
5787	2247	Rayon	6
5788	2247	Denim	7
5789	2247	Velvet	8
5790	2247	Chiffon	9
5791	2247	Georgette	10
5792	2247	Crepe	11
5793	2247	Satin	12
5794	2247	Net	13
5795	2247	Jersey	14
5796	2247	Lycra	15
5797	2252	Cotton	1
5798	2252	Polyester	2
5799	2252	Silk	3
5800	2252	Linen	4
5801	2252	Wool	5
5802	2252	Rayon	6
5803	2252	Denim	7
5804	2252	Velvet	8
5805	2252	Chiffon	9
5806	2252	Georgette	10
5807	2252	Crepe	11
5808	2252	Satin	12
5809	2252	Net	13
5810	2252	Jersey	14
5811	2252	Lycra	15
5812	2257	Cotton	1
5813	2257	Polyester	2
5814	2257	Silk	3
5815	2257	Linen	4
5816	2257	Wool	5
5817	2257	Rayon	6
5818	2257	Denim	7
5819	2257	Velvet	8
5820	2257	Chiffon	9
5821	2257	Georgette	10
5822	2257	Crepe	11
5823	2257	Satin	12
5824	2257	Net	13
5825	2257	Jersey	14
5826	2257	Lycra	15
5827	2264	Cotton	1
5828	2264	Polyester	2
5829	2264	Silk	3
5830	2264	Linen	4
5831	2264	Wool	5
5832	2264	Rayon	6
5833	2264	Denim	7
5834	2264	Velvet	8
5835	2264	Chiffon	9
5836	2264	Georgette	10
5837	2264	Crepe	11
5838	2264	Satin	12
5839	2264	Net	13
5840	2264	Jersey	14
5841	2264	Lycra	15
5842	2269	Cotton	1
5843	2269	Polyester	2
5844	2269	Silk	3
5845	2269	Linen	4
5846	2269	Wool	5
5847	2269	Rayon	6
5848	2269	Denim	7
5849	2269	Velvet	8
5850	2269	Chiffon	9
5851	2269	Georgette	10
5852	2269	Crepe	11
5853	2269	Satin	12
5854	2269	Net	13
5855	2269	Jersey	14
5856	2269	Lycra	15
5857	2275	Cotton	1
5858	2275	Polyester	2
5859	2275	Silk	3
5860	2275	Linen	4
5861	2275	Wool	5
5862	2275	Rayon	6
5863	2275	Denim	7
5864	2275	Velvet	8
5865	2275	Chiffon	9
5866	2275	Georgette	10
5867	2275	Crepe	11
5868	2275	Satin	12
5869	2275	Net	13
5870	2275	Jersey	14
5871	2275	Lycra	15
5872	2282	Cotton	1
5873	2282	Polyester	2
5874	2282	Silk	3
5875	2282	Linen	4
5876	2282	Wool	5
5877	2282	Rayon	6
5878	2282	Denim	7
5879	2282	Velvet	8
5880	2282	Chiffon	9
5881	2282	Georgette	10
5882	2282	Crepe	11
5883	2282	Satin	12
5884	2282	Net	13
5885	2282	Jersey	14
5886	2282	Lycra	15
5887	2287	Cotton	1
5888	2287	Polyester	2
5889	2287	Silk	3
5890	2287	Linen	4
5891	2287	Wool	5
5892	2287	Rayon	6
5893	2287	Denim	7
5894	2287	Velvet	8
5895	2287	Chiffon	9
5896	2287	Georgette	10
5897	2287	Crepe	11
5898	2287	Satin	12
5899	2287	Net	13
5900	2287	Jersey	14
5901	2287	Lycra	15
5902	2293	Cotton	1
5903	2293	Polyester	2
5904	2293	Silk	3
5905	2293	Linen	4
5906	2293	Wool	5
5907	2293	Rayon	6
5908	2293	Denim	7
5909	2293	Velvet	8
5910	2293	Chiffon	9
5911	2293	Georgette	10
5912	2293	Crepe	11
5913	2293	Satin	12
5914	2293	Net	13
5915	2293	Jersey	14
5916	2293	Lycra	15
5917	2299	Cotton	1
5918	2299	Polyester	2
5919	2299	Silk	3
5920	2299	Linen	4
5921	2299	Wool	5
5922	2299	Rayon	6
5923	2299	Denim	7
5924	2299	Velvet	8
5925	2299	Chiffon	9
5926	2299	Georgette	10
5927	2299	Crepe	11
5928	2299	Satin	12
5929	2299	Net	13
5930	2299	Jersey	14
5931	2299	Lycra	15
5932	2302	Cotton	1
5933	2302	Polyester	2
5934	2302	Silk	3
5935	2302	Linen	4
5936	2302	Wool	5
5937	2302	Rayon	6
5938	2302	Denim	7
5939	2302	Velvet	8
5940	2302	Chiffon	9
5941	2302	Georgette	10
5942	2302	Crepe	11
5943	2302	Satin	12
5944	2302	Net	13
5945	2302	Jersey	14
5946	2302	Lycra	15
5947	2308	Cotton	1
5948	2308	Polyester	2
5949	2308	Silk	3
5950	2308	Linen	4
5951	2308	Wool	5
5952	2308	Rayon	6
5953	2308	Denim	7
5954	2308	Velvet	8
5955	2308	Chiffon	9
5956	2308	Georgette	10
5957	2308	Crepe	11
5958	2308	Satin	12
5959	2308	Net	13
5960	2308	Jersey	14
5961	2308	Lycra	15
5962	2314	Cotton	1
5963	2314	Polyester	2
5964	2314	Silk	3
5965	2314	Linen	4
5966	2314	Wool	5
5967	2314	Rayon	6
5968	2314	Denim	7
5969	2314	Velvet	8
5970	2314	Chiffon	9
5971	2314	Georgette	10
5972	2314	Crepe	11
5973	2314	Satin	12
5974	2314	Net	13
5975	2314	Jersey	14
5976	2314	Lycra	15
5977	2320	Cotton	1
5978	2320	Polyester	2
5979	2320	Silk	3
5980	2320	Linen	4
5981	2320	Wool	5
5982	2320	Rayon	6
5983	2320	Denim	7
5984	2320	Velvet	8
5985	2320	Chiffon	9
5986	2320	Georgette	10
5987	2320	Crepe	11
5988	2320	Satin	12
5989	2320	Net	13
5990	2320	Jersey	14
5991	2320	Lycra	15
5992	2325	Cotton	1
5993	2325	Polyester	2
5994	2325	Silk	3
5995	2325	Linen	4
5996	2325	Wool	5
5997	2325	Rayon	6
5998	2325	Denim	7
5999	2325	Velvet	8
6000	2325	Chiffon	9
6001	2325	Georgette	10
6002	2325	Crepe	11
6003	2325	Satin	12
6004	2325	Net	13
6005	2325	Jersey	14
6006	2325	Lycra	15
6007	2330	Cotton	1
6008	2330	Polyester	2
6009	2330	Silk	3
6010	2330	Linen	4
6011	2330	Wool	5
6012	2330	Rayon	6
6013	2330	Denim	7
6014	2330	Velvet	8
6015	2330	Chiffon	9
6016	2330	Georgette	10
6017	2330	Crepe	11
6018	2330	Satin	12
6019	2330	Net	13
6020	2330	Jersey	14
6021	2330	Lycra	15
6022	2336	Cotton	1
6023	2336	Polyester	2
6024	2336	Silk	3
6025	2336	Linen	4
6026	2336	Wool	5
6027	2336	Rayon	6
6028	2336	Denim	7
6029	2336	Velvet	8
6030	2336	Chiffon	9
6031	2336	Georgette	10
6032	2336	Crepe	11
6033	2336	Satin	12
6034	2336	Net	13
6035	2336	Jersey	14
6036	2336	Lycra	15
6037	2340	Cotton	1
6038	2340	Polyester	2
6039	2340	Silk	3
6040	2340	Linen	4
6041	2340	Wool	5
6042	2340	Rayon	6
6043	2340	Denim	7
6044	2340	Velvet	8
6045	2340	Chiffon	9
6046	2340	Georgette	10
6047	2340	Crepe	11
6048	2340	Satin	12
6049	2340	Net	13
6050	2340	Jersey	14
6051	2340	Lycra	15
6052	2346	Cotton	1
6053	2346	Polyester	2
6054	2346	Silk	3
6055	2346	Linen	4
6056	2346	Wool	5
6057	2346	Rayon	6
6058	2346	Denim	7
6059	2346	Velvet	8
6060	2346	Chiffon	9
6061	2346	Georgette	10
6062	2346	Crepe	11
6063	2346	Satin	12
6064	2346	Net	13
6065	2346	Jersey	14
6066	2346	Lycra	15
6067	2352	Cotton	1
6068	2352	Polyester	2
6069	2352	Silk	3
6070	2352	Linen	4
6071	2352	Wool	5
6072	2352	Rayon	6
6073	2352	Denim	7
6074	2352	Velvet	8
6075	2352	Chiffon	9
6076	2352	Georgette	10
6077	2352	Crepe	11
6078	2352	Satin	12
6079	2352	Net	13
6080	2352	Jersey	14
6081	2352	Lycra	15
6082	2358	Cotton	1
6083	2358	Polyester	2
6084	2358	Silk	3
6085	2358	Linen	4
6086	2358	Wool	5
6087	2358	Rayon	6
6088	2358	Denim	7
6089	2358	Velvet	8
6090	2358	Chiffon	9
6091	2358	Georgette	10
6092	2358	Crepe	11
6093	2358	Satin	12
6094	2358	Net	13
6095	2358	Jersey	14
6096	2358	Lycra	15
6097	2364	Cotton	1
6098	2364	Polyester	2
6099	2364	Silk	3
6100	2364	Linen	4
6101	2364	Wool	5
6102	2364	Rayon	6
6103	2364	Denim	7
6104	2364	Velvet	8
6105	2364	Chiffon	9
6106	2364	Georgette	10
6107	2364	Crepe	11
6108	2364	Satin	12
6109	2364	Net	13
6110	2364	Jersey	14
6111	2364	Lycra	15
6112	2371	Cotton	1
6113	2371	Polyester	2
6114	2371	Silk	3
6115	2371	Linen	4
6116	2371	Wool	5
6117	2371	Rayon	6
6118	2371	Denim	7
6119	2371	Velvet	8
6120	2371	Chiffon	9
6121	2371	Georgette	10
6122	2371	Crepe	11
6123	2371	Satin	12
6124	2371	Net	13
6125	2371	Jersey	14
6126	2371	Lycra	15
6127	2376	Cotton	1
6128	2376	Polyester	2
6129	2376	Silk	3
6130	2376	Linen	4
6131	2376	Wool	5
6132	2376	Rayon	6
6133	2376	Denim	7
6134	2376	Velvet	8
6135	2376	Chiffon	9
6136	2376	Georgette	10
6137	2376	Crepe	11
6138	2376	Satin	12
6139	2376	Net	13
6140	2376	Jersey	14
6141	2376	Lycra	15
6142	2382	Cotton	1
6143	2382	Polyester	2
6144	2382	Silk	3
6145	2382	Linen	4
6146	2382	Wool	5
6147	2382	Rayon	6
6148	2382	Denim	7
6149	2382	Velvet	8
6150	2382	Chiffon	9
6151	2382	Georgette	10
6152	2382	Crepe	11
6153	2382	Satin	12
6154	2382	Net	13
6155	2382	Jersey	14
6156	2382	Lycra	15
6157	2388	Cotton	1
6158	2388	Polyester	2
6159	2388	Silk	3
6160	2388	Linen	4
6161	2388	Wool	5
6162	2388	Rayon	6
6163	2388	Denim	7
6164	2388	Velvet	8
6165	2388	Chiffon	9
6166	2388	Georgette	10
6167	2388	Crepe	11
6168	2388	Satin	12
6169	2388	Net	13
6170	2388	Jersey	14
6171	2388	Lycra	15
6172	2401	Cotton	1
6173	2401	Polyester	2
6174	2401	Silk	3
6175	2401	Linen	4
6176	2401	Wool	5
6177	2401	Rayon	6
6178	2401	Denim	7
6179	2401	Velvet	8
6180	2401	Chiffon	9
6181	2401	Georgette	10
6182	2401	Crepe	11
6183	2401	Satin	12
6184	2401	Net	13
6185	2401	Jersey	14
6186	2401	Lycra	15
6187	2406	Cotton	1
6188	2406	Polyester	2
6189	2406	Silk	3
6190	2406	Linen	4
6191	2406	Wool	5
6192	2406	Rayon	6
6193	2406	Denim	7
6194	2406	Velvet	8
6195	2406	Chiffon	9
6196	2406	Georgette	10
6197	2406	Crepe	11
6198	2406	Satin	12
6199	2406	Net	13
6200	2406	Jersey	14
6201	2406	Lycra	15
6202	2412	Cotton	1
6203	2412	Polyester	2
6204	2412	Silk	3
6205	2412	Linen	4
6206	2412	Wool	5
6207	2412	Rayon	6
6208	2412	Denim	7
6209	2412	Velvet	8
6210	2412	Chiffon	9
6211	2412	Georgette	10
6212	2412	Crepe	11
6213	2412	Satin	12
6214	2412	Net	13
6215	2412	Jersey	14
6216	2412	Lycra	15
6217	2417	Cotton	1
6218	2417	Polyester	2
6219	2417	Silk	3
6220	2417	Linen	4
6221	2417	Wool	5
6222	2417	Rayon	6
6223	2417	Denim	7
6224	2417	Velvet	8
6225	2417	Chiffon	9
6226	2417	Georgette	10
6227	2417	Crepe	11
6228	2417	Satin	12
6229	2417	Net	13
6230	2417	Jersey	14
6231	2417	Lycra	15
6232	2422	Cotton	1
6233	2422	Polyester	2
6234	2422	Silk	3
6235	2422	Linen	4
6236	2422	Wool	5
6237	2422	Rayon	6
6238	2422	Denim	7
6239	2422	Velvet	8
6240	2422	Chiffon	9
6241	2422	Georgette	10
6242	2422	Crepe	11
6243	2422	Satin	12
6244	2422	Net	13
6245	2422	Jersey	14
6246	2422	Lycra	15
6247	2427	Cotton	1
6248	2427	Polyester	2
6249	2427	Silk	3
6250	2427	Linen	4
6251	2427	Wool	5
6252	2427	Rayon	6
6253	2427	Denim	7
6254	2427	Velvet	8
6255	2427	Chiffon	9
6256	2427	Georgette	10
6257	2427	Crepe	11
6258	2427	Satin	12
6259	2427	Net	13
6260	2427	Jersey	14
6261	2427	Lycra	15
6262	2433	Cotton	1
6263	2433	Polyester	2
6264	2433	Silk	3
6265	2433	Linen	4
6266	2433	Wool	5
6267	2433	Rayon	6
6268	2433	Denim	7
6269	2433	Velvet	8
6270	2433	Chiffon	9
6271	2433	Georgette	10
6272	2433	Crepe	11
6273	2433	Satin	12
6274	2433	Net	13
6275	2433	Jersey	14
6276	2433	Lycra	15
6277	2439	Cotton	1
6278	2439	Polyester	2
6279	2439	Silk	3
6280	2439	Linen	4
6281	2439	Wool	5
6282	2439	Rayon	6
6283	2439	Denim	7
6284	2439	Velvet	8
6285	2439	Chiffon	9
6286	2439	Georgette	10
6287	2439	Crepe	11
6288	2439	Satin	12
6289	2439	Net	13
6290	2439	Jersey	14
6291	2439	Lycra	15
6292	2529	Cotton	1
6293	2529	Polyester	2
6294	2529	Silk	3
6295	2529	Linen	4
6296	2529	Wool	5
6297	2529	Rayon	6
6298	2529	Denim	7
6299	2529	Velvet	8
6300	2529	Chiffon	9
6301	2529	Georgette	10
6302	2529	Crepe	11
6303	2529	Satin	12
6304	2529	Net	13
6305	2529	Jersey	14
6306	2529	Lycra	15
6307	2533	Cotton	1
6308	2533	Polyester	2
6309	2533	Silk	3
6310	2533	Linen	4
6311	2533	Wool	5
6312	2533	Rayon	6
6313	2533	Denim	7
6314	2533	Velvet	8
6315	2533	Chiffon	9
6316	2533	Georgette	10
6317	2533	Crepe	11
6318	2533	Satin	12
6319	2533	Net	13
6320	2533	Jersey	14
6321	2533	Lycra	15
6322	2539	Cotton	1
6323	2539	Polyester	2
6324	2539	Silk	3
6325	2539	Linen	4
6326	2539	Wool	5
6327	2539	Rayon	6
6328	2539	Denim	7
6329	2539	Velvet	8
6330	2539	Chiffon	9
6331	2539	Georgette	10
6332	2539	Crepe	11
6333	2539	Satin	12
6334	2539	Net	13
6335	2539	Jersey	14
6336	2539	Lycra	15
6337	2543	Cotton	1
6338	2543	Polyester	2
6339	2543	Silk	3
6340	2543	Linen	4
6341	2543	Wool	5
6342	2543	Rayon	6
6343	2543	Denim	7
6344	2543	Velvet	8
6345	2543	Chiffon	9
6346	2543	Georgette	10
6347	2543	Crepe	11
6348	2543	Satin	12
6349	2543	Net	13
6350	2543	Jersey	14
6351	2543	Lycra	15
6352	2549	Cotton	1
6353	2549	Polyester	2
6354	2549	Silk	3
6355	2549	Linen	4
6356	2549	Wool	5
6357	2549	Rayon	6
6358	2549	Denim	7
6359	2549	Velvet	8
6360	2549	Chiffon	9
6361	2549	Georgette	10
6362	2549	Crepe	11
6363	2549	Satin	12
6364	2549	Net	13
6365	2549	Jersey	14
6366	2549	Lycra	15
6367	2552	Cotton	1
6368	2552	Polyester	2
6369	2552	Silk	3
6370	2552	Linen	4
6371	2552	Wool	5
6372	2552	Rayon	6
6373	2552	Denim	7
6374	2552	Velvet	8
6375	2552	Chiffon	9
6376	2552	Georgette	10
6377	2552	Crepe	11
6378	2552	Satin	12
6379	2552	Net	13
6380	2552	Jersey	14
6381	2552	Lycra	15
6382	2557	Cotton	1
6383	2557	Polyester	2
6384	2557	Silk	3
6385	2557	Linen	4
6386	2557	Wool	5
6387	2557	Rayon	6
6388	2557	Denim	7
6389	2557	Velvet	8
6390	2557	Chiffon	9
6391	2557	Georgette	10
6392	2557	Crepe	11
6393	2557	Satin	12
6394	2557	Net	13
6395	2557	Jersey	14
6396	2557	Lycra	15
6397	2562	Cotton	1
6398	2562	Polyester	2
6399	2562	Silk	3
6400	2562	Linen	4
6401	2562	Wool	5
6402	2562	Rayon	6
6403	2562	Denim	7
6404	2562	Velvet	8
6405	2562	Chiffon	9
6406	2562	Georgette	10
6407	2562	Crepe	11
6408	2562	Satin	12
6409	2562	Net	13
6410	2562	Jersey	14
6411	2562	Lycra	15
6412	2567	Cotton	1
6413	2567	Polyester	2
6414	2567	Silk	3
6415	2567	Linen	4
6416	2567	Wool	5
6417	2567	Rayon	6
6418	2567	Denim	7
6419	2567	Velvet	8
6420	2567	Chiffon	9
6421	2567	Georgette	10
6422	2567	Crepe	11
6423	2567	Satin	12
6424	2567	Net	13
6425	2567	Jersey	14
6426	2567	Lycra	15
6427	2572	Cotton	1
6428	2572	Polyester	2
6429	2572	Silk	3
6430	2572	Linen	4
6431	2572	Wool	5
6432	2572	Rayon	6
6433	2572	Denim	7
6434	2572	Velvet	8
6435	2572	Chiffon	9
6436	2572	Georgette	10
6437	2572	Crepe	11
6438	2572	Satin	12
6439	2572	Net	13
6440	2572	Jersey	14
6441	2572	Lycra	15
6442	1616	50g	1
6443	1616	100g	2
6444	1616	200g	3
6445	1616	250g	4
6446	1616	500g	5
6447	1616	1kg	6
6448	1616	2kg	5
6449	1616	5kg	8
6450	1616	10kg	9
6451	1616	Single	10
6452	1616	Pack of 2	11
6453	1616	Pack of 3	12
6454	1616	Pack of 5	13
6455	1616	Pack of 10	14
6456	1620	50g	1
6457	1620	100g	2
6458	1620	200g	3
6459	1620	250g	4
6460	1620	500g	5
6461	1620	1kg	6
6462	1620	2kg	5
6463	1620	5kg	8
6464	1620	10kg	9
6465	1620	Single	10
6466	1620	Pack of 2	11
6467	1620	Pack of 3	12
6468	1620	Pack of 5	13
6469	1620	Pack of 10	14
6470	1624	50g	1
6471	1624	100g	2
6472	1624	200g	3
6473	1624	250g	4
6474	1624	500g	5
6475	1624	1kg	6
6476	1624	2kg	5
6477	1624	5kg	8
6478	1624	10kg	9
6479	1624	Single	10
6480	1624	Pack of 2	11
6481	1624	Pack of 3	12
6482	1624	Pack of 5	13
6483	1624	Pack of 10	14
6484	1628	50g	1
6485	1628	100g	2
6486	1628	200g	3
6487	1628	250g	4
6488	1628	500g	5
6489	1628	1kg	6
6490	1628	2kg	5
6491	1628	5kg	8
6492	1628	10kg	9
6493	1628	Single	10
6494	1628	Pack of 2	11
6495	1628	Pack of 3	12
6496	1628	Pack of 5	13
6497	1628	Pack of 10	14
6498	1633	50g	1
6499	1633	100g	2
6500	1633	200g	3
6501	1633	250g	4
6502	1633	500g	5
6503	1633	1kg	6
6504	1633	2kg	5
6505	1633	5kg	8
6506	1633	10kg	9
6507	1633	Single	10
6508	1633	Pack of 2	11
6509	1633	Pack of 3	12
6510	1633	Pack of 5	13
6511	1633	Pack of 10	14
6512	1637	50g	1
6513	1637	100g	2
6514	1637	200g	3
6515	1637	250g	4
6516	1637	500g	5
6517	1637	1kg	6
6518	1637	2kg	5
6519	1637	5kg	8
6520	1637	10kg	9
6521	1637	Single	10
6522	1637	Pack of 2	11
6523	1637	Pack of 3	12
6524	1637	Pack of 5	13
6525	1637	Pack of 10	14
6526	1641	50g	1
6527	1641	100g	2
6528	1641	200g	3
6529	1641	250g	4
6530	1641	500g	5
6531	1641	1kg	6
6532	1641	2kg	5
6533	1641	5kg	8
6534	1641	10kg	9
6535	1641	Single	10
6536	1641	Pack of 2	11
6537	1641	Pack of 3	12
6538	1641	Pack of 5	13
6539	1641	Pack of 10	14
6540	1644	50g	1
6541	1644	100g	2
6542	1644	200g	3
6543	1644	250g	4
6544	1644	500g	5
6545	1644	1kg	6
6546	1644	2kg	5
6547	1644	5kg	8
6548	1644	10kg	9
6549	1644	Single	10
6550	1644	Pack of 2	11
6551	1644	Pack of 3	12
6552	1644	Pack of 5	13
6553	1644	Pack of 10	14
6554	1648	50g	1
6555	1648	100g	2
6556	1648	200g	3
6557	1648	250g	4
6558	1648	500g	5
6559	1648	1kg	6
6560	1648	2kg	5
6561	1648	5kg	8
6562	1648	10kg	9
6563	1648	Single	10
6564	1648	Pack of 2	11
6565	1648	Pack of 3	12
6566	1648	Pack of 5	13
6567	1648	Pack of 10	14
6568	1652	50g	1
6569	1652	100g	2
6570	1652	200g	3
6571	1652	250g	4
6572	1652	500g	5
6573	1652	1kg	6
6574	1652	2kg	5
6575	1652	5kg	8
6576	1652	10kg	9
6577	1652	Single	10
6578	1652	Pack of 2	11
6579	1652	Pack of 3	12
6580	1652	Pack of 5	13
6581	1652	Pack of 10	14
6582	1656	50g	1
6583	1656	100g	2
6584	1656	200g	3
6585	1656	250g	4
6586	1656	500g	5
6587	1656	1kg	6
6588	1656	2kg	5
6589	1656	5kg	8
6590	1656	10kg	9
6591	1656	Single	10
6592	1656	Pack of 2	11
6593	1656	Pack of 3	12
6594	1656	Pack of 5	13
6595	1656	Pack of 10	14
6596	1660	50g	1
6597	1660	100g	2
6598	1660	200g	3
6599	1660	250g	4
6600	1660	500g	5
6601	1660	1kg	6
6602	1660	2kg	5
6603	1660	5kg	8
6604	1660	10kg	9
6605	1660	Single	10
6606	1660	Pack of 2	11
6607	1660	Pack of 3	12
6608	1660	Pack of 5	13
6609	1660	Pack of 10	14
6610	1664	50g	1
6611	1664	100g	2
6612	1664	200g	3
6613	1664	250g	4
6614	1664	500g	5
6615	1664	1kg	6
6616	1664	2kg	5
6617	1664	5kg	8
6618	1664	10kg	9
6619	1664	Single	10
6620	1664	Pack of 2	11
6621	1664	Pack of 3	12
6622	1664	Pack of 5	13
6623	1664	Pack of 10	14
6624	1668	50g	1
6625	1668	100g	2
6626	1668	200g	3
6627	1668	250g	4
6628	1668	500g	5
6629	1668	1kg	6
6630	1668	2kg	5
6631	1668	5kg	8
6632	1668	10kg	9
6633	1668	Single	10
6634	1668	Pack of 2	11
6635	1668	Pack of 3	12
6636	1668	Pack of 5	13
6637	1668	Pack of 10	14
6638	1672	50g	1
6639	1672	100g	2
6640	1672	200g	3
6641	1672	250g	4
6642	1672	500g	5
6643	1672	1kg	6
6644	1672	2kg	5
6645	1672	5kg	8
6646	1672	10kg	9
6647	1672	Single	10
6648	1672	Pack of 2	11
6649	1672	Pack of 3	12
6650	1672	Pack of 5	13
6651	1672	Pack of 10	14
6652	1676	50g	1
6653	1676	100g	2
6654	1676	200g	3
6655	1676	250g	4
6656	1676	500g	5
6657	1676	1kg	6
6658	1676	2kg	5
6659	1676	5kg	8
6660	1676	10kg	9
6661	1676	Single	10
6662	1676	Pack of 2	11
6663	1676	Pack of 3	12
6664	1676	Pack of 5	13
6665	1676	Pack of 10	14
6666	1680	50g	1
6667	1680	100g	2
6668	1680	200g	3
6669	1680	250g	4
6670	1680	500g	5
6671	1680	1kg	6
6672	1680	2kg	5
6673	1680	5kg	8
6674	1680	10kg	9
6675	1680	Single	10
6676	1680	Pack of 2	11
6677	1680	Pack of 3	12
6678	1680	Pack of 5	13
6679	1680	Pack of 10	14
6680	1684	50g	1
6681	1684	100g	2
6682	1684	200g	3
6683	1684	250g	4
6684	1684	500g	5
6685	1684	1kg	6
6686	1684	2kg	5
6687	1684	5kg	8
6688	1684	10kg	9
6689	1684	Single	10
6690	1684	Pack of 2	11
6691	1684	Pack of 3	12
6692	1684	Pack of 5	13
6693	1684	Pack of 10	14
6694	1688	50g	1
6695	1688	100g	2
6696	1688	200g	3
6697	1688	250g	4
6698	1688	500g	5
6699	1688	1kg	6
6700	1688	2kg	5
6701	1688	5kg	8
6702	1688	10kg	9
6703	1688	Single	10
6704	1688	Pack of 2	11
6705	1688	Pack of 3	12
6706	1688	Pack of 5	13
6707	1688	Pack of 10	14
6708	1692	50g	1
6709	1692	100g	2
6710	1692	200g	3
6711	1692	250g	4
6712	1692	500g	5
6713	1692	1kg	6
6714	1692	2kg	5
6715	1692	5kg	8
6716	1692	10kg	9
6717	1692	Single	10
6718	1692	Pack of 2	11
6719	1692	Pack of 3	12
6720	1692	Pack of 5	13
6721	1692	Pack of 10	14
6722	1696	50g	1
6723	1696	100g	2
6724	1696	200g	3
6725	1696	250g	4
6726	1696	500g	5
6727	1696	1kg	6
6728	1696	2kg	5
6729	1696	5kg	8
6730	1696	10kg	9
6731	1696	Single	10
6732	1696	Pack of 2	11
6733	1696	Pack of 3	12
6734	1696	Pack of 5	13
6735	1696	Pack of 10	14
6736	1700	50g	1
6737	1700	100g	2
6738	1700	200g	3
6739	1700	250g	4
6740	1700	500g	5
6741	1700	1kg	6
6742	1700	2kg	5
6743	1700	5kg	8
6744	1700	10kg	9
6745	1700	Single	10
6746	1700	Pack of 2	11
6747	1700	Pack of 3	12
6748	1700	Pack of 5	13
6749	1700	Pack of 10	14
6750	1704	50g	1
6751	1704	100g	2
6752	1704	200g	3
6753	1704	250g	4
6754	1704	500g	5
6755	1704	1kg	6
6756	1704	2kg	5
6757	1704	5kg	8
6758	1704	10kg	9
6759	1704	Single	10
6760	1704	Pack of 2	11
6761	1704	Pack of 3	12
6762	1704	Pack of 5	13
6763	1704	Pack of 10	14
6764	1708	50g	1
6765	1708	100g	2
6766	1708	200g	3
6767	1708	250g	4
6768	1708	500g	5
6769	1708	1kg	6
6770	1708	2kg	5
6771	1708	5kg	8
6772	1708	10kg	9
6773	1708	Single	10
6774	1708	Pack of 2	11
6775	1708	Pack of 3	12
6776	1708	Pack of 5	13
6777	1708	Pack of 10	14
6778	1712	50g	1
6779	1712	100g	2
6780	1712	200g	3
6781	1712	250g	4
6782	1712	500g	5
6783	1712	1kg	6
6784	1712	2kg	5
6785	1712	5kg	8
6786	1712	10kg	9
6787	1712	Single	10
6788	1712	Pack of 2	11
6789	1712	Pack of 3	12
6790	1712	Pack of 5	13
6791	1712	Pack of 10	14
6792	1716	50g	1
6793	1716	100g	2
6794	1716	200g	3
6795	1716	250g	4
6796	1716	500g	5
6797	1716	1kg	6
6798	1716	2kg	5
6799	1716	5kg	8
6800	1716	10kg	9
6801	1716	Single	10
6802	1716	Pack of 2	11
6803	1716	Pack of 3	12
6804	1716	Pack of 5	13
6805	1716	Pack of 10	14
6806	1720	50g	1
6807	1720	100g	2
6808	1720	200g	3
6809	1720	250g	4
6810	1720	500g	5
6811	1720	1kg	6
6812	1720	2kg	5
6813	1720	5kg	8
6814	1720	10kg	9
6815	1720	Single	10
6816	1720	Pack of 2	11
6817	1720	Pack of 3	12
6818	1720	Pack of 5	13
6819	1720	Pack of 10	14
6820	1724	50g	1
6821	1724	100g	2
6822	1724	200g	3
6823	1724	250g	4
6824	1724	500g	5
6825	1724	1kg	6
6826	1724	2kg	5
6827	1724	5kg	8
6828	1724	10kg	9
6829	1724	Single	10
6830	1724	Pack of 2	11
6831	1724	Pack of 3	12
6832	1724	Pack of 5	13
6833	1724	Pack of 10	14
6834	1728	50g	1
6835	1728	100g	2
6836	1728	200g	3
6837	1728	250g	4
6838	1728	500g	5
6839	1728	1kg	6
6840	1728	2kg	5
6841	1728	5kg	8
6842	1728	10kg	9
6843	1728	Single	10
6844	1728	Pack of 2	11
6845	1728	Pack of 3	12
6846	1728	Pack of 5	13
6847	1728	Pack of 10	14
6848	1732	50g	1
6849	1732	100g	2
6850	1732	200g	3
6851	1732	250g	4
6852	1732	500g	5
6853	1732	1kg	6
6854	1732	2kg	5
6855	1732	5kg	8
6856	1732	10kg	9
6857	1732	Single	10
6858	1732	Pack of 2	11
6859	1732	Pack of 3	12
6860	1732	Pack of 5	13
6861	1732	Pack of 10	14
6862	1736	50g	1
6863	1736	100g	2
6864	1736	200g	3
6865	1736	250g	4
6866	1736	500g	5
6867	1736	1kg	6
6868	1736	2kg	5
6869	1736	5kg	8
6870	1736	10kg	9
6871	1736	Single	10
6872	1736	Pack of 2	11
6873	1736	Pack of 3	12
6874	1736	Pack of 5	13
6875	1736	Pack of 10	14
6876	1740	50g	1
6877	1740	100g	2
6878	1740	200g	3
6879	1740	250g	4
6880	1740	500g	5
6881	1740	1kg	6
6882	1740	2kg	5
6883	1740	5kg	8
6884	1740	10kg	9
6885	1740	Single	10
6886	1740	Pack of 2	11
6887	1740	Pack of 3	12
6888	1740	Pack of 5	13
6889	1740	Pack of 10	14
6890	1744	50g	1
6891	1744	100g	2
6892	1744	200g	3
6893	1744	250g	4
6894	1744	500g	5
6895	1744	1kg	6
6896	1744	2kg	5
6897	1744	5kg	8
6898	1744	10kg	9
6899	1744	Single	10
6900	1744	Pack of 2	11
6901	1744	Pack of 3	12
6902	1744	Pack of 5	13
6903	1744	Pack of 10	14
6904	1748	50g	1
6905	1748	100g	2
6906	1748	200g	3
6907	1748	250g	4
6908	1748	500g	5
6909	1748	1kg	6
6910	1748	2kg	5
6911	1748	5kg	8
6912	1748	10kg	9
6913	1748	Single	10
6914	1748	Pack of 2	11
6915	1748	Pack of 3	12
6916	1748	Pack of 5	13
6917	1748	Pack of 10	14
6918	1752	50g	1
6919	1752	100g	2
6920	1752	200g	3
6921	1752	250g	4
6922	1752	500g	5
6923	1752	1kg	6
6924	1752	2kg	5
6925	1752	5kg	8
6926	1752	10kg	9
6927	1752	Single	10
6928	1752	Pack of 2	11
6929	1752	Pack of 3	12
6930	1752	Pack of 5	13
6931	1752	Pack of 10	14
6932	1756	50g	1
6933	1756	100g	2
6934	1756	200g	3
6935	1756	250g	4
6936	1756	500g	5
6937	1756	1kg	6
6938	1756	2kg	5
6939	1756	5kg	8
6940	1756	10kg	9
6941	1756	Single	10
6942	1756	Pack of 2	11
6943	1756	Pack of 3	12
6944	1756	Pack of 5	13
6945	1756	Pack of 10	14
6946	1760	50g	1
6947	1760	100g	2
6948	1760	200g	3
6949	1760	250g	4
6950	1760	500g	5
6951	1760	1kg	6
6952	1760	2kg	5
6953	1760	5kg	8
6954	1760	10kg	9
6955	1760	Single	10
6956	1760	Pack of 2	11
6957	1760	Pack of 3	12
6958	1760	Pack of 5	13
6959	1760	Pack of 10	14
6960	1764	50g	1
6961	1764	100g	2
6962	1764	200g	3
6963	1764	250g	4
6964	1764	500g	5
6965	1764	1kg	6
6966	1764	2kg	5
6967	1764	5kg	8
6968	1764	10kg	9
6969	1764	Single	10
6970	1764	Pack of 2	11
6971	1764	Pack of 3	12
6972	1764	Pack of 5	13
6973	1764	Pack of 10	14
6974	1768	50g	1
6975	1768	100g	2
6976	1768	200g	3
6977	1768	250g	4
6978	1768	500g	5
6979	1768	1kg	6
6980	1768	2kg	5
6981	1768	5kg	8
6982	1768	10kg	9
6983	1768	Single	10
6984	1768	Pack of 2	11
6985	1768	Pack of 3	12
6986	1768	Pack of 5	13
6987	1768	Pack of 10	14
6988	1772	50g	1
6989	1772	100g	2
6990	1772	200g	3
6991	1772	250g	4
6992	1772	500g	5
6993	1772	1kg	6
6994	1772	2kg	5
6995	1772	5kg	8
6996	1772	10kg	9
6997	1772	Single	10
6998	1772	Pack of 2	11
6999	1772	Pack of 3	12
7000	1772	Pack of 5	13
7001	1772	Pack of 10	14
7002	1776	50g	1
7003	1776	100g	2
7004	1776	200g	3
7005	1776	250g	4
7006	1776	500g	5
7007	1776	1kg	6
7008	1776	2kg	5
7009	1776	5kg	8
7010	1776	10kg	9
7011	1776	Single	10
7012	1776	Pack of 2	11
7013	1776	Pack of 3	12
7014	1776	Pack of 5	13
7015	1776	Pack of 10	14
7016	1780	50g	1
7017	1780	100g	2
7018	1780	200g	3
7019	1780	250g	4
7020	1780	500g	5
7021	1780	1kg	6
7022	1780	2kg	5
7023	1780	5kg	8
7024	1780	10kg	9
7025	1780	Single	10
7026	1780	Pack of 2	11
7027	1780	Pack of 3	12
7028	1780	Pack of 5	13
7029	1780	Pack of 10	14
7030	1784	50g	1
7031	1784	100g	2
7032	1784	200g	3
7033	1784	250g	4
7034	1784	500g	5
7035	1784	1kg	6
7036	1784	2kg	5
7037	1784	5kg	8
7038	1784	10kg	9
7039	1784	Single	10
7040	1784	Pack of 2	11
7041	1784	Pack of 3	12
7042	1784	Pack of 5	13
7043	1784	Pack of 10	14
7044	1788	50g	1
7045	1788	100g	2
7046	1788	200g	3
7047	1788	250g	4
7048	1788	500g	5
7049	1788	1kg	6
7050	1788	2kg	5
7051	1788	5kg	8
7052	1788	10kg	9
7053	1788	Single	10
7054	1788	Pack of 2	11
7055	1788	Pack of 3	12
7056	1788	Pack of 5	13
7057	1788	Pack of 10	14
7058	1792	50g	1
7059	1792	100g	2
7060	1792	200g	3
7061	1792	250g	4
7062	1792	500g	5
7063	1792	1kg	6
7064	1792	2kg	5
7065	1792	5kg	8
7066	1792	10kg	9
7067	1792	Single	10
7068	1792	Pack of 2	11
7069	1792	Pack of 3	12
7070	1792	Pack of 5	13
7071	1792	Pack of 10	14
7072	1796	50g	1
7073	1796	100g	2
7074	1796	200g	3
7075	1796	250g	4
7076	1796	500g	5
7077	1796	1kg	6
7078	1796	2kg	5
7079	1796	5kg	8
7080	1796	10kg	9
7081	1796	Single	10
7082	1796	Pack of 2	11
7083	1796	Pack of 3	12
7084	1796	Pack of 5	13
7085	1796	Pack of 10	14
7086	1801	50g	1
7087	1801	100g	2
7088	1801	200g	3
7089	1801	250g	4
7090	1801	500g	5
7091	1801	1kg	6
7092	1801	2kg	5
7093	1801	5kg	8
7094	1801	10kg	9
7095	1801	Single	10
7096	1801	Pack of 2	11
7097	1801	Pack of 3	12
7098	1801	Pack of 5	13
7099	1801	Pack of 10	14
7100	1804	50g	1
7101	1804	100g	2
7102	1804	200g	3
7103	1804	250g	4
7104	1804	500g	5
7105	1804	1kg	6
7106	1804	2kg	5
7107	1804	5kg	8
7108	1804	10kg	9
7109	1804	Single	10
7110	1804	Pack of 2	11
7111	1804	Pack of 3	12
7112	1804	Pack of 5	13
7113	1804	Pack of 10	14
7114	1809	50g	1
7115	1809	100g	2
7116	1809	200g	3
7117	1809	250g	4
7118	1809	500g	5
7119	1809	1kg	6
7120	1809	2kg	5
7121	1809	5kg	8
7122	1809	10kg	9
7123	1809	Single	10
7124	1809	Pack of 2	11
7125	1809	Pack of 3	12
7126	1809	Pack of 5	13
7127	1809	Pack of 10	14
7128	1813	50g	1
7129	1813	100g	2
7130	1813	200g	3
7131	1813	250g	4
7132	1813	500g	5
7133	1813	1kg	6
7134	1813	2kg	5
7135	1813	5kg	8
7136	1813	10kg	9
7137	1813	Single	10
7138	1813	Pack of 2	11
7139	1813	Pack of 3	12
7140	1813	Pack of 5	13
7141	1813	Pack of 10	14
7142	1818	50g	1
7143	1818	100g	2
7144	1818	200g	3
7145	1818	250g	4
7146	1818	500g	5
7147	1818	1kg	6
7148	1818	2kg	5
7149	1818	5kg	8
7150	1818	10kg	9
7151	1818	Single	10
7152	1818	Pack of 2	11
7153	1818	Pack of 3	12
7154	1818	Pack of 5	13
7155	1818	Pack of 10	14
7156	1822	50g	1
7157	1822	100g	2
7158	1822	200g	3
7159	1822	250g	4
7160	1822	500g	5
7161	1822	1kg	6
7162	1822	2kg	5
7163	1822	5kg	8
7164	1822	10kg	9
7165	1822	Single	10
7166	1822	Pack of 2	11
7167	1822	Pack of 3	12
7168	1822	Pack of 5	13
7169	1822	Pack of 10	14
7170	1826	50g	1
7171	1826	100g	2
7172	1826	200g	3
7173	1826	250g	4
7174	1826	500g	5
7175	1826	1kg	6
7176	1826	2kg	5
7177	1826	5kg	8
7178	1826	10kg	9
7179	1826	Single	10
7180	1826	Pack of 2	11
7181	1826	Pack of 3	12
7182	1826	Pack of 5	13
7183	1826	Pack of 10	14
7184	1838	50g	1
7185	1838	100g	2
7186	1838	200g	3
7187	1838	250g	4
7188	1838	500g	5
7189	1838	1kg	6
7190	1838	2kg	5
7191	1838	5kg	8
7192	1838	10kg	9
7193	1838	Single	10
7194	1838	Pack of 2	11
7195	1838	Pack of 3	12
7196	1838	Pack of 5	13
7197	1838	Pack of 10	14
7198	1842	50g	1
7199	1842	100g	2
7200	1842	200g	3
7201	1842	250g	4
7202	1842	500g	5
7203	1842	1kg	6
7204	1842	2kg	5
7205	1842	5kg	8
7206	1842	10kg	9
7207	1842	Single	10
7208	1842	Pack of 2	11
7209	1842	Pack of 3	12
7210	1842	Pack of 5	13
7211	1842	Pack of 10	14
7212	1850	50g	1
7213	1850	100g	2
7214	1850	200g	3
7215	1850	250g	4
7216	1850	500g	5
7217	1850	1kg	6
7218	1850	2kg	5
7219	1850	5kg	8
7220	1850	10kg	9
7221	1850	Single	10
7222	1850	Pack of 2	11
7223	1850	Pack of 3	12
7224	1850	Pack of 5	13
7225	1850	Pack of 10	14
7226	1854	50g	1
7227	1854	100g	2
7228	1854	200g	3
7229	1854	250g	4
7230	1854	500g	5
7231	1854	1kg	6
7232	1854	2kg	5
7233	1854	5kg	8
7234	1854	10kg	9
7235	1854	Single	10
7236	1854	Pack of 2	11
7237	1854	Pack of 3	12
7238	1854	Pack of 5	13
7239	1854	Pack of 10	14
7240	1858	50g	1
7241	1858	100g	2
7242	1858	200g	3
7243	1858	250g	4
7244	1858	500g	5
7245	1858	1kg	6
7246	1858	2kg	5
7247	1858	5kg	8
7248	1858	10kg	9
7249	1858	Single	10
7250	1858	Pack of 2	11
7251	1858	Pack of 3	12
7252	1858	Pack of 5	13
7253	1858	Pack of 10	14
7254	2157	50g	1
7255	2157	100g	2
7256	2157	200g	3
7257	2157	250g	4
7258	2157	500g	5
7259	2157	1kg	6
7260	2157	2kg	5
7261	2157	5kg	8
7262	2157	10kg	9
7263	2157	Single	10
7264	2157	Pack of 2	11
7265	2157	Pack of 3	12
7266	2157	Pack of 5	13
7267	2157	Pack of 10	14
7268	2239	50g	1
7269	2239	100g	2
7270	2239	200g	3
7271	2239	250g	4
7272	2239	500g	5
7273	2239	1kg	6
7274	2239	2kg	5
7275	2239	5kg	8
7276	2239	10kg	9
7277	2239	Single	10
7278	2239	Pack of 2	11
7279	2239	Pack of 3	12
7280	2239	Pack of 5	13
7281	2239	Pack of 10	14
7282	2530	50g	1
7283	2530	100g	2
7284	2530	200g	3
7285	2530	250g	4
7286	2530	500g	5
7287	2530	1kg	6
7288	2530	2kg	5
7289	2530	5kg	8
7290	2530	10kg	9
7291	2530	Single	10
7292	2530	Pack of 2	11
7293	2530	Pack of 3	12
7294	2530	Pack of 5	13
7295	2530	Pack of 10	14
7296	2534	50g	1
7297	2534	100g	2
7298	2534	200g	3
7299	2534	250g	4
7300	2534	500g	5
7301	2534	1kg	6
7302	2534	2kg	5
7303	2534	5kg	8
7304	2534	10kg	9
7305	2534	Single	10
7306	2534	Pack of 2	11
7307	2534	Pack of 3	12
7308	2534	Pack of 5	13
7309	2534	Pack of 10	14
7310	2540	50g	1
7311	2540	100g	2
7312	2540	200g	3
7313	2540	250g	4
7314	2540	500g	5
7315	2540	1kg	6
7316	2540	2kg	5
7317	2540	5kg	8
7318	2540	10kg	9
7319	2540	Single	10
7320	2540	Pack of 2	11
7321	2540	Pack of 3	12
7322	2540	Pack of 5	13
7323	2540	Pack of 10	14
7324	131	Standard	1
7325	131	Premium	2
7326	131	Deluxe	3
7327	131	Basic	4
7328	131	Professional	5
7329	131	Regular	6
7330	131	Classic	7
7331	131	Modern	8
7332	131	Traditional	9
7333	131	Contemporary	10
7334	162	Standard	1
7335	162	Premium	2
7336	162	Deluxe	3
7337	162	Basic	4
7338	162	Professional	5
7339	162	Regular	6
7340	162	Classic	7
7341	162	Modern	8
7342	162	Traditional	9
7343	162	Contemporary	10
7344	165	Standard	1
7345	165	Premium	2
7346	165	Deluxe	3
7347	165	Basic	4
7348	165	Professional	5
7349	165	Regular	6
7350	165	Classic	7
7351	165	Modern	8
7352	165	Traditional	9
7353	165	Contemporary	10
7354	170	Standard	1
7355	170	Premium	2
7356	170	Deluxe	3
7357	170	Basic	4
7358	170	Professional	5
7359	170	Regular	6
7360	170	Classic	7
7361	170	Modern	8
7362	170	Traditional	9
7363	170	Contemporary	10
7364	188	Standard	1
7365	188	Premium	2
7366	188	Deluxe	3
7367	188	Basic	4
7368	188	Professional	5
7369	188	Regular	6
7370	188	Classic	7
7371	188	Modern	8
7372	188	Traditional	9
7373	188	Contemporary	10
7374	217	Standard	1
7375	217	Premium	2
7376	217	Deluxe	3
7377	217	Basic	4
7378	217	Professional	5
7379	217	Regular	6
7380	217	Classic	7
7381	217	Modern	8
7382	217	Traditional	9
7383	217	Contemporary	10
7384	221	Standard	1
7385	221	Premium	2
7386	221	Deluxe	3
7387	221	Basic	4
7388	221	Professional	5
7389	221	Regular	6
7390	221	Classic	7
7391	221	Modern	8
7392	221	Traditional	9
7393	221	Contemporary	10
7394	225	Standard	1
7395	225	Premium	2
7396	225	Deluxe	3
7397	225	Basic	4
7398	225	Professional	5
7399	225	Regular	6
7400	225	Classic	7
7401	225	Modern	8
7402	225	Traditional	9
7403	225	Contemporary	10
7404	229	Standard	1
7405	229	Premium	2
7406	229	Deluxe	3
7407	229	Basic	4
7408	229	Professional	5
7409	229	Regular	6
7410	229	Classic	7
7411	229	Modern	8
7412	229	Traditional	9
7413	229	Contemporary	10
7414	261	Standard	1
7415	261	Premium	2
7416	261	Deluxe	3
7417	261	Basic	4
7418	261	Professional	5
7419	261	Regular	6
7420	261	Classic	7
7421	261	Modern	8
7422	261	Traditional	9
7423	261	Contemporary	10
7424	265	Standard	1
7425	265	Premium	2
7426	265	Deluxe	3
7427	265	Basic	4
7428	265	Professional	5
7429	265	Regular	6
7430	265	Classic	7
7431	265	Modern	8
7432	265	Traditional	9
7433	265	Contemporary	10
7434	269	Standard	1
7435	269	Premium	2
7436	269	Deluxe	3
7437	269	Basic	4
7438	269	Professional	5
7439	269	Regular	6
7440	269	Classic	7
7441	269	Modern	8
7442	269	Traditional	9
7443	269	Contemporary	10
7444	273	Standard	1
7445	273	Premium	2
7446	273	Deluxe	3
7447	273	Basic	4
7448	273	Professional	5
7449	273	Regular	6
7450	273	Classic	7
7451	273	Modern	8
7452	273	Traditional	9
7453	273	Contemporary	10
7454	327	Standard	1
7455	327	Premium	2
7456	327	Deluxe	3
7457	327	Basic	4
7458	327	Professional	5
7459	327	Regular	6
7460	327	Classic	7
7461	327	Modern	8
7462	327	Traditional	9
7463	327	Contemporary	10
7464	332	Standard	1
7465	332	Premium	2
7466	332	Deluxe	3
7467	332	Basic	4
7468	332	Professional	5
7469	332	Regular	6
7470	332	Classic	7
7471	332	Modern	8
7472	332	Traditional	9
7473	332	Contemporary	10
7474	337	Standard	1
7475	337	Premium	2
7476	337	Deluxe	3
7477	337	Basic	4
7478	337	Professional	5
7479	337	Regular	6
7480	337	Classic	7
7481	337	Modern	8
7482	337	Traditional	9
7483	337	Contemporary	10
7484	341	Standard	1
7485	341	Premium	2
7486	341	Deluxe	3
7487	341	Basic	4
7488	341	Professional	5
7489	341	Regular	6
7490	341	Classic	7
7491	341	Modern	8
7492	341	Traditional	9
7493	341	Contemporary	10
7494	355	Standard	1
7495	355	Premium	2
7496	355	Deluxe	3
7497	355	Basic	4
7498	355	Professional	5
7499	355	Regular	6
7500	355	Classic	7
7501	355	Modern	8
7502	355	Traditional	9
7503	355	Contemporary	10
7504	374	Standard	1
7505	374	Premium	2
7506	374	Deluxe	3
7507	374	Basic	4
7508	374	Professional	5
7509	374	Regular	6
7510	374	Classic	7
7511	374	Modern	8
7512	374	Traditional	9
7513	374	Contemporary	10
7514	392	Standard	1
7515	392	Premium	2
7516	392	Deluxe	3
7517	392	Basic	4
7518	392	Professional	5
7519	392	Regular	6
7520	392	Classic	7
7521	392	Modern	8
7522	392	Traditional	9
7523	392	Contemporary	10
7524	397	Standard	1
7525	397	Premium	2
7526	397	Deluxe	3
7527	397	Basic	4
7528	397	Professional	5
7529	397	Regular	6
7530	397	Classic	7
7531	397	Modern	8
7532	397	Traditional	9
7533	397	Contemporary	10
7534	427	Standard	1
7535	427	Premium	2
7536	427	Deluxe	3
7537	427	Basic	4
7538	427	Professional	5
7539	427	Regular	6
7540	427	Classic	7
7541	427	Modern	8
7542	427	Traditional	9
7543	427	Contemporary	10
7544	456	Standard	1
7545	456	Premium	2
7546	456	Deluxe	3
7547	456	Basic	4
7548	456	Professional	5
7549	456	Regular	6
7550	456	Classic	7
7551	456	Modern	8
7552	456	Traditional	9
7553	456	Contemporary	10
7554	461	Standard	1
7555	461	Premium	2
7556	461	Deluxe	3
7557	461	Basic	4
7558	461	Professional	5
7559	461	Regular	6
7560	461	Classic	7
7561	461	Modern	8
7562	461	Traditional	9
7563	461	Contemporary	10
7564	466	Standard	1
7565	466	Premium	2
7566	466	Deluxe	3
7567	466	Basic	4
7568	466	Professional	5
7569	466	Regular	6
7570	466	Classic	7
7571	466	Modern	8
7572	466	Traditional	9
7573	466	Contemporary	10
7574	471	Standard	1
7575	471	Premium	2
7576	471	Deluxe	3
7577	471	Basic	4
7578	471	Professional	5
7579	471	Regular	6
7580	471	Classic	7
7581	471	Modern	8
7582	471	Traditional	9
7583	471	Contemporary	10
7584	476	Standard	1
7585	476	Premium	2
7586	476	Deluxe	3
7587	476	Basic	4
7588	476	Professional	5
7589	476	Regular	6
7590	476	Classic	7
7591	476	Modern	8
7592	476	Traditional	9
7593	476	Contemporary	10
7594	481	Standard	1
7595	481	Premium	2
7596	481	Deluxe	3
7597	481	Basic	4
7598	481	Professional	5
7599	481	Regular	6
7600	481	Classic	7
7601	481	Modern	8
7602	481	Traditional	9
7603	481	Contemporary	10
7604	492	Standard	1
7605	492	Premium	2
7606	492	Deluxe	3
7607	492	Basic	4
7608	492	Professional	5
7609	492	Regular	6
7610	492	Classic	7
7611	492	Modern	8
7612	492	Traditional	9
7613	492	Contemporary	10
7614	507	Standard	1
7615	507	Premium	2
7616	507	Deluxe	3
7617	507	Basic	4
7618	507	Professional	5
7619	507	Regular	6
7620	507	Classic	7
7621	507	Modern	8
7622	507	Traditional	9
7623	507	Contemporary	10
7624	514	Standard	1
7625	514	Premium	2
7626	514	Deluxe	3
7627	514	Basic	4
7628	514	Professional	5
7629	514	Regular	6
7630	514	Classic	7
7631	514	Modern	8
7632	514	Traditional	9
7633	514	Contemporary	10
7634	527	Standard	1
7635	527	Premium	2
7636	527	Deluxe	3
7637	527	Basic	4
7638	527	Professional	5
7639	527	Regular	6
7640	527	Classic	7
7641	527	Modern	8
7642	527	Traditional	9
7643	527	Contemporary	10
7644	531	Standard	1
7645	531	Premium	2
7646	531	Deluxe	3
7647	531	Basic	4
7648	531	Professional	5
7649	531	Regular	6
7650	531	Classic	7
7651	531	Modern	8
7652	531	Traditional	9
7653	531	Contemporary	10
7654	536	Standard	1
7655	536	Premium	2
7656	536	Deluxe	3
7657	536	Basic	4
7658	536	Professional	5
7659	536	Regular	6
7660	536	Classic	7
7661	536	Modern	8
7662	536	Traditional	9
7663	536	Contemporary	10
7664	544	Standard	1
7665	544	Premium	2
7666	544	Deluxe	3
7667	544	Basic	4
7668	544	Professional	5
7669	544	Regular	6
7670	544	Classic	7
7671	544	Modern	8
7672	544	Traditional	9
7673	544	Contemporary	10
7674	547	Standard	1
7675	547	Premium	2
7676	547	Deluxe	3
7677	547	Basic	4
7678	547	Professional	5
7679	547	Regular	6
7680	547	Classic	7
7681	547	Modern	8
7682	547	Traditional	9
7683	547	Contemporary	10
7684	551	Standard	1
7685	551	Premium	2
7686	551	Deluxe	3
7687	551	Basic	4
7688	551	Professional	5
7689	551	Regular	6
7690	551	Classic	7
7691	551	Modern	8
7692	551	Traditional	9
7693	551	Contemporary	10
7694	607	Standard	1
7695	607	Premium	2
7696	607	Deluxe	3
7697	607	Basic	4
7698	607	Professional	5
7699	607	Regular	6
7700	607	Classic	7
7701	607	Modern	8
7702	607	Traditional	9
7703	607	Contemporary	10
7704	611	Standard	1
7705	611	Premium	2
7706	611	Deluxe	3
7707	611	Basic	4
7708	611	Professional	5
7709	611	Regular	6
7710	611	Classic	7
7711	611	Modern	8
7712	611	Traditional	9
7713	611	Contemporary	10
7714	631	Standard	1
7715	631	Premium	2
7716	631	Deluxe	3
7717	631	Basic	4
7718	631	Professional	5
7719	631	Regular	6
7720	631	Classic	7
7721	631	Modern	8
7722	631	Traditional	9
7723	631	Contemporary	10
7724	655	Standard	1
7725	655	Premium	2
7726	655	Deluxe	3
7727	655	Basic	4
7728	655	Professional	5
7729	655	Regular	6
7730	655	Classic	7
7731	655	Modern	8
7732	655	Traditional	9
7733	655	Contemporary	10
7734	672	Standard	1
7735	672	Premium	2
7736	672	Deluxe	3
7737	672	Basic	4
7738	672	Professional	5
7739	672	Regular	6
7740	672	Classic	7
7741	672	Modern	8
7742	672	Traditional	9
7743	672	Contemporary	10
7744	693	Standard	1
7745	693	Premium	2
7746	693	Deluxe	3
7747	693	Basic	4
7748	693	Professional	5
7749	693	Regular	6
7750	693	Classic	7
7751	693	Modern	8
7752	693	Traditional	9
7753	693	Contemporary	10
7754	700	Standard	1
7755	700	Premium	2
7756	700	Deluxe	3
7757	700	Basic	4
7758	700	Professional	5
7759	700	Regular	6
7760	700	Classic	7
7761	700	Modern	8
7762	700	Traditional	9
7763	700	Contemporary	10
7764	742	Standard	1
7765	742	Premium	2
7766	742	Deluxe	3
7767	742	Basic	4
7768	742	Professional	5
7769	742	Regular	6
7770	742	Classic	7
7771	742	Modern	8
7772	742	Traditional	9
7773	742	Contemporary	10
7774	748	Standard	1
7775	748	Premium	2
7776	748	Deluxe	3
7777	748	Basic	4
7778	748	Professional	5
7779	748	Regular	6
7780	748	Classic	7
7781	748	Modern	8
7782	748	Traditional	9
7783	748	Contemporary	10
7784	814	Standard	1
7785	814	Premium	2
7786	814	Deluxe	3
7787	814	Basic	4
7788	814	Professional	5
7789	814	Regular	6
7790	814	Classic	7
7791	814	Modern	8
7792	814	Traditional	9
7793	814	Contemporary	10
7794	818	Standard	1
7795	818	Premium	2
7796	818	Deluxe	3
7797	818	Basic	4
7798	818	Professional	5
7799	818	Regular	6
7800	818	Classic	7
7801	818	Modern	8
7802	818	Traditional	9
7803	818	Contemporary	10
7804	822	Standard	1
7805	822	Premium	2
7806	822	Deluxe	3
7807	822	Basic	4
7808	822	Professional	5
7809	822	Regular	6
7810	822	Classic	7
7811	822	Modern	8
7812	822	Traditional	9
7813	822	Contemporary	10
7814	826	Standard	1
7815	826	Premium	2
7816	826	Deluxe	3
7817	826	Basic	4
7818	826	Professional	5
7819	826	Regular	6
7820	826	Classic	7
7821	826	Modern	8
7822	826	Traditional	9
7823	826	Contemporary	10
7824	931	Standard	1
7825	931	Premium	2
7826	931	Deluxe	3
7827	931	Basic	4
7828	931	Professional	5
7829	931	Regular	6
7830	931	Classic	7
7831	931	Modern	8
7832	931	Traditional	9
7833	931	Contemporary	10
7834	941	Standard	1
7835	941	Premium	2
7836	941	Deluxe	3
7837	941	Basic	4
7838	941	Professional	5
7839	941	Regular	6
7840	941	Classic	7
7841	941	Modern	8
7842	941	Traditional	9
7843	941	Contemporary	10
7844	1007	Standard	1
7845	1007	Premium	2
7846	1007	Deluxe	3
7847	1007	Basic	4
7848	1007	Professional	5
7849	1007	Regular	6
7850	1007	Classic	7
7851	1007	Modern	8
7852	1007	Traditional	9
7853	1007	Contemporary	10
7854	1011	Standard	1
7855	1011	Premium	2
7856	1011	Deluxe	3
7857	1011	Basic	4
7858	1011	Professional	5
7859	1011	Regular	6
7860	1011	Classic	7
7861	1011	Modern	8
7862	1011	Traditional	9
7863	1011	Contemporary	10
7864	1015	Standard	1
7865	1015	Premium	2
7866	1015	Deluxe	3
7867	1015	Basic	4
7868	1015	Professional	5
7869	1015	Regular	6
7870	1015	Classic	7
7871	1015	Modern	8
7872	1015	Traditional	9
7873	1015	Contemporary	10
7874	1112	Standard	1
7875	1112	Premium	2
7876	1112	Deluxe	3
7877	1112	Basic	4
7878	1112	Professional	5
7879	1112	Regular	6
7880	1112	Classic	7
7881	1112	Modern	8
7882	1112	Traditional	9
7883	1112	Contemporary	10
7884	1450	Standard	1
7885	1450	Premium	2
7886	1450	Deluxe	3
7887	1450	Basic	4
7888	1450	Professional	5
7889	1450	Regular	6
7890	1450	Classic	7
7891	1450	Modern	8
7892	1450	Traditional	9
7893	1450	Contemporary	10
7894	1625	Standard	1
7895	1625	Premium	2
7896	1625	Deluxe	3
7897	1625	Basic	4
7898	1625	Professional	5
7899	1625	Regular	6
7900	1625	Classic	7
7901	1625	Modern	8
7902	1625	Traditional	9
7903	1625	Contemporary	10
7904	1629	Standard	1
7905	1629	Premium	2
7906	1629	Deluxe	3
7907	1629	Basic	4
7908	1629	Professional	5
7909	1629	Regular	6
7910	1629	Classic	7
7911	1629	Modern	8
7912	1629	Traditional	9
7913	1629	Contemporary	10
7914	1653	Standard	1
7915	1653	Premium	2
7916	1653	Deluxe	3
7917	1653	Basic	4
7918	1653	Professional	5
7919	1653	Regular	6
7920	1653	Classic	7
7921	1653	Modern	8
7922	1653	Traditional	9
7923	1653	Contemporary	10
7924	1657	Standard	1
7925	1657	Premium	2
7926	1657	Deluxe	3
7927	1657	Basic	4
7928	1657	Professional	5
7929	1657	Regular	6
7930	1657	Classic	7
7931	1657	Modern	8
7932	1657	Traditional	9
7933	1657	Contemporary	10
7934	1681	Standard	1
7935	1681	Premium	2
7936	1681	Deluxe	3
7937	1681	Basic	4
7938	1681	Professional	5
7939	1681	Regular	6
7940	1681	Classic	7
7941	1681	Modern	8
7942	1681	Traditional	9
7943	1681	Contemporary	10
7944	1685	Standard	1
7945	1685	Premium	2
7946	1685	Deluxe	3
7947	1685	Basic	4
7948	1685	Professional	5
7949	1685	Regular	6
7950	1685	Classic	7
7951	1685	Modern	8
7952	1685	Traditional	9
7953	1685	Contemporary	10
7954	1689	Standard	1
7955	1689	Premium	2
7956	1689	Deluxe	3
7957	1689	Basic	4
7958	1689	Professional	5
7959	1689	Regular	6
7960	1689	Classic	7
7961	1689	Modern	8
7962	1689	Traditional	9
7963	1689	Contemporary	10
7964	1693	Standard	1
7965	1693	Premium	2
7966	1693	Deluxe	3
7967	1693	Basic	4
7968	1693	Professional	5
7969	1693	Regular	6
7970	1693	Classic	7
7971	1693	Modern	8
7972	1693	Traditional	9
7973	1693	Contemporary	10
7974	1709	Standard	1
7975	1709	Premium	2
7976	1709	Deluxe	3
7977	1709	Basic	4
7978	1709	Professional	5
7979	1709	Regular	6
7980	1709	Classic	7
7981	1709	Modern	8
7982	1709	Traditional	9
7983	1709	Contemporary	10
7984	1713	Standard	1
7985	1713	Premium	2
7986	1713	Deluxe	3
7987	1713	Basic	4
7988	1713	Professional	5
7989	1713	Regular	6
7990	1713	Classic	7
7991	1713	Modern	8
7992	1713	Traditional	9
7993	1713	Contemporary	10
7994	1721	Standard	1
7995	1721	Premium	2
7996	1721	Deluxe	3
7997	1721	Basic	4
7998	1721	Professional	5
7999	1721	Regular	6
8000	1721	Classic	7
8001	1721	Modern	8
8002	1721	Traditional	9
8003	1721	Contemporary	10
8004	1729	Standard	1
8005	1729	Premium	2
8006	1729	Deluxe	3
8007	1729	Basic	4
8008	1729	Professional	5
8009	1729	Regular	6
8010	1729	Classic	7
8011	1729	Modern	8
8012	1729	Traditional	9
8013	1729	Contemporary	10
8014	1733	Standard	1
8015	1733	Premium	2
8016	1733	Deluxe	3
8017	1733	Basic	4
8018	1733	Professional	5
8019	1733	Regular	6
8020	1733	Classic	7
8021	1733	Modern	8
8022	1733	Traditional	9
8023	1733	Contemporary	10
8024	1741	Standard	1
8025	1741	Premium	2
8026	1741	Deluxe	3
8027	1741	Basic	4
8028	1741	Professional	5
8029	1741	Regular	6
8030	1741	Classic	7
8031	1741	Modern	8
8032	1741	Traditional	9
8033	1741	Contemporary	10
8034	1749	Standard	1
8035	1749	Premium	2
8036	1749	Deluxe	3
8037	1749	Basic	4
8038	1749	Professional	5
8039	1749	Regular	6
8040	1749	Classic	7
8041	1749	Modern	8
8042	1749	Traditional	9
8043	1749	Contemporary	10
8044	1753	Standard	1
8045	1753	Premium	2
8046	1753	Deluxe	3
8047	1753	Basic	4
8048	1753	Professional	5
8049	1753	Regular	6
8050	1753	Classic	7
8051	1753	Modern	8
8052	1753	Traditional	9
8053	1753	Contemporary	10
8054	1757	Standard	1
8055	1757	Premium	2
8056	1757	Deluxe	3
8057	1757	Basic	4
8058	1757	Professional	5
8059	1757	Regular	6
8060	1757	Classic	7
8061	1757	Modern	8
8062	1757	Traditional	9
8063	1757	Contemporary	10
8064	1761	Standard	1
8065	1761	Premium	2
8066	1761	Deluxe	3
8067	1761	Basic	4
8068	1761	Professional	5
8069	1761	Regular	6
8070	1761	Classic	7
8071	1761	Modern	8
8072	1761	Traditional	9
8073	1761	Contemporary	10
8074	1765	Standard	1
8075	1765	Premium	2
8076	1765	Deluxe	3
8077	1765	Basic	4
8078	1765	Professional	5
8079	1765	Regular	6
8080	1765	Classic	7
8081	1765	Modern	8
8082	1765	Traditional	9
8083	1765	Contemporary	10
8084	1817	Standard	1
8085	1817	Premium	2
8086	1817	Deluxe	3
8087	1817	Basic	4
8088	1817	Professional	5
8089	1817	Regular	6
8090	1817	Classic	7
8091	1817	Modern	8
8092	1817	Traditional	9
8093	1817	Contemporary	10
8094	1821	Standard	1
8095	1821	Premium	2
8096	1821	Deluxe	3
8097	1821	Basic	4
8098	1821	Professional	5
8099	1821	Regular	6
8100	1821	Classic	7
8101	1821	Modern	8
8102	1821	Traditional	9
8103	1821	Contemporary	10
8104	1830	Standard	1
8105	1830	Premium	2
8106	1830	Deluxe	3
8107	1830	Basic	4
8108	1830	Professional	5
8109	1830	Regular	6
8110	1830	Classic	7
8111	1830	Modern	8
8112	1830	Traditional	9
8113	1830	Contemporary	10
8114	1849	Standard	1
8115	1849	Premium	2
8116	1849	Deluxe	3
8117	1849	Basic	4
8118	1849	Professional	5
8119	1849	Regular	6
8120	1849	Classic	7
8121	1849	Modern	8
8122	1849	Traditional	9
8123	1849	Contemporary	10
8124	1857	Standard	1
8125	1857	Premium	2
8126	1857	Deluxe	3
8127	1857	Basic	4
8128	1857	Professional	5
8129	1857	Regular	6
8130	1857	Classic	7
8131	1857	Modern	8
8132	1857	Traditional	9
8133	1857	Contemporary	10
8134	1878	Standard	1
8135	1878	Premium	2
8136	1878	Deluxe	3
8137	1878	Basic	4
8138	1878	Professional	5
8139	1878	Regular	6
8140	1878	Classic	7
8141	1878	Modern	8
8142	1878	Traditional	9
8143	1878	Contemporary	10
8144	1881	Standard	1
8145	1881	Premium	2
8146	1881	Deluxe	3
8147	1881	Basic	4
8148	1881	Professional	5
8149	1881	Regular	6
8150	1881	Classic	7
8151	1881	Modern	8
8152	1881	Traditional	9
8153	1881	Contemporary	10
8154	1885	Standard	1
8155	1885	Premium	2
8156	1885	Deluxe	3
8157	1885	Basic	4
8158	1885	Professional	5
8159	1885	Regular	6
8160	1885	Classic	7
8161	1885	Modern	8
8162	1885	Traditional	9
8163	1885	Contemporary	10
8164	1917	Standard	1
8165	1917	Premium	2
8166	1917	Deluxe	3
8167	1917	Basic	4
8168	1917	Professional	5
8169	1917	Regular	6
8170	1917	Classic	7
8171	1917	Modern	8
8172	1917	Traditional	9
8173	1917	Contemporary	10
8174	1937	Standard	1
8175	1937	Premium	2
8176	1937	Deluxe	3
8177	1937	Basic	4
8178	1937	Professional	5
8179	1937	Regular	6
8180	1937	Classic	7
8181	1937	Modern	8
8182	1937	Traditional	9
8183	1937	Contemporary	10
8184	1941	Standard	1
8185	1941	Premium	2
8186	1941	Deluxe	3
8187	1941	Basic	4
8188	1941	Professional	5
8189	1941	Regular	6
8190	1941	Classic	7
8191	1941	Modern	8
8192	1941	Traditional	9
8193	1941	Contemporary	10
8194	2020	Standard	1
8195	2020	Premium	2
8196	2020	Deluxe	3
8197	2020	Basic	4
8198	2020	Professional	5
8199	2020	Regular	6
8200	2020	Classic	7
8201	2020	Modern	8
8202	2020	Traditional	9
8203	2020	Contemporary	10
8204	2041	Standard	1
8205	2041	Premium	2
8206	2041	Deluxe	3
8207	2041	Basic	4
8208	2041	Professional	5
8209	2041	Regular	6
8210	2041	Classic	7
8211	2041	Modern	8
8212	2041	Traditional	9
8213	2041	Contemporary	10
8214	2151	Standard	1
8215	2151	Premium	2
8216	2151	Deluxe	3
8217	2151	Basic	4
8218	2151	Professional	5
8219	2151	Regular	6
8220	2151	Classic	7
8221	2151	Modern	8
8222	2151	Traditional	9
8223	2151	Contemporary	10
8224	2234	Standard	1
8225	2234	Premium	2
8226	2234	Deluxe	3
8227	2234	Basic	4
8228	2234	Professional	5
8229	2234	Regular	6
8230	2234	Classic	7
8231	2234	Modern	8
8232	2234	Traditional	9
8233	2234	Contemporary	10
8234	2238	Standard	1
8235	2238	Premium	2
8236	2238	Deluxe	3
8237	2238	Basic	4
8238	2238	Professional	5
8239	2238	Regular	6
8240	2238	Classic	7
8241	2238	Modern	8
8242	2238	Traditional	9
8243	2238	Contemporary	10
8244	2244	Standard	1
8245	2244	Premium	2
8246	2244	Deluxe	3
8247	2244	Basic	4
8248	2244	Professional	5
8249	2244	Regular	6
8250	2244	Classic	7
8251	2244	Modern	8
8252	2244	Traditional	9
8253	2244	Contemporary	10
8254	2304	Standard	1
8255	2304	Premium	2
8256	2304	Deluxe	3
8257	2304	Basic	4
8258	2304	Professional	5
8259	2304	Regular	6
8260	2304	Classic	7
8261	2304	Modern	8
8262	2304	Traditional	9
8263	2304	Contemporary	10
8264	2322	Standard	1
8265	2322	Premium	2
8266	2322	Deluxe	3
8267	2322	Basic	4
8268	2322	Professional	5
8269	2322	Regular	6
8270	2322	Classic	7
8271	2322	Modern	8
8272	2322	Traditional	9
8273	2322	Contemporary	10
8274	2332	Standard	1
8275	2332	Premium	2
8276	2332	Deluxe	3
8277	2332	Basic	4
8278	2332	Professional	5
8279	2332	Regular	6
8280	2332	Classic	7
8281	2332	Modern	8
8282	2332	Traditional	9
8283	2332	Contemporary	10
8284	2414	Standard	1
8285	2414	Premium	2
8286	2414	Deluxe	3
8287	2414	Basic	4
8288	2414	Professional	5
8289	2414	Regular	6
8290	2414	Classic	7
8291	2414	Modern	8
8292	2414	Traditional	9
8293	2414	Contemporary	10
8294	2419	Standard	1
8295	2419	Premium	2
8296	2419	Deluxe	3
8297	2419	Basic	4
8298	2419	Professional	5
8299	2419	Regular	6
8300	2419	Classic	7
8301	2419	Modern	8
8302	2419	Traditional	9
8303	2419	Contemporary	10
8304	2435	Standard	1
8305	2435	Premium	2
8306	2435	Deluxe	3
8307	2435	Basic	4
8308	2435	Professional	5
8309	2435	Regular	6
8310	2435	Classic	7
8311	2435	Modern	8
8312	2435	Traditional	9
8313	2435	Contemporary	10
8314	2446	Standard	1
8315	2446	Premium	2
8316	2446	Deluxe	3
8317	2446	Basic	4
8318	2446	Professional	5
8319	2446	Regular	6
8320	2446	Classic	7
8321	2446	Modern	8
8322	2446	Traditional	9
8323	2446	Contemporary	10
8324	2451	Standard	1
8325	2451	Premium	2
8326	2451	Deluxe	3
8327	2451	Basic	4
8328	2451	Professional	5
8329	2451	Regular	6
8330	2451	Classic	7
8331	2451	Modern	8
8332	2451	Traditional	9
8333	2451	Contemporary	10
8334	2455	Standard	1
8335	2455	Premium	2
8336	2455	Deluxe	3
8337	2455	Basic	4
8338	2455	Professional	5
8339	2455	Regular	6
8340	2455	Classic	7
8341	2455	Modern	8
8342	2455	Traditional	9
8343	2455	Contemporary	10
8344	2470	Standard	1
8345	2470	Premium	2
8346	2470	Deluxe	3
8347	2470	Basic	4
8348	2470	Professional	5
8349	2470	Regular	6
8350	2470	Classic	7
8351	2470	Modern	8
8352	2470	Traditional	9
8353	2470	Contemporary	10
8354	2481	Standard	1
8355	2481	Premium	2
8356	2481	Deluxe	3
8357	2481	Basic	4
8358	2481	Professional	5
8359	2481	Regular	6
8360	2481	Classic	7
8361	2481	Modern	8
8362	2481	Traditional	9
8363	2481	Contemporary	10
8364	2488	Standard	1
8365	2488	Premium	2
8366	2488	Deluxe	3
8367	2488	Basic	4
8368	2488	Professional	5
8369	2488	Regular	6
8370	2488	Classic	7
8371	2488	Modern	8
8372	2488	Traditional	9
8373	2488	Contemporary	10
8374	2493	Standard	1
8375	2493	Premium	2
8376	2493	Deluxe	3
8377	2493	Basic	4
8378	2493	Professional	5
8379	2493	Regular	6
8380	2493	Classic	7
8381	2493	Modern	8
8382	2493	Traditional	9
8383	2493	Contemporary	10
8384	2498	Standard	1
8385	2498	Premium	2
8386	2498	Deluxe	3
8387	2498	Basic	4
8388	2498	Professional	5
8389	2498	Regular	6
8390	2498	Classic	7
8391	2498	Modern	8
8392	2498	Traditional	9
8393	2498	Contemporary	10
8394	2537	Standard	1
8395	2537	Premium	2
8396	2537	Deluxe	3
8397	2537	Basic	4
8398	2537	Professional	5
8399	2537	Regular	6
8400	2537	Classic	7
8401	2537	Modern	8
8402	2537	Traditional	9
8403	2537	Contemporary	10
8404	44	Short	1
8405	44	Medium	2
8406	44	Long	3
8407	44	Extra Long	4
8408	44	10cm	5
8409	44	20cm	6
8410	44	30cm	7
8411	44	40cm	8
8412	44	50cm	9
8413	44	100cm	10
8414	51	Short	1
8415	51	Medium	2
8416	51	Long	3
8417	51	Extra Long	4
8418	51	10cm	5
8419	51	20cm	6
8420	51	30cm	7
8421	51	40cm	8
8422	51	50cm	9
8423	51	100cm	10
8424	58	Short	1
8425	58	Medium	2
8426	58	Long	3
8427	58	Extra Long	4
8428	58	10cm	5
8429	58	20cm	6
8430	58	30cm	7
8431	58	40cm	8
8432	58	50cm	9
8433	58	100cm	10
8434	65	Short	1
8435	65	Medium	2
8436	65	Long	3
8437	65	Extra Long	4
8438	65	10cm	5
8439	65	20cm	6
8440	65	30cm	7
8441	65	40cm	8
8442	65	50cm	9
8443	65	100cm	10
8444	72	Short	1
8445	72	Medium	2
8446	72	Long	3
8447	72	Extra Long	4
8448	72	10cm	5
8449	72	20cm	6
8450	72	30cm	7
8451	72	40cm	8
8452	72	50cm	9
8453	72	100cm	10
8454	152	Short	1
8455	152	Medium	2
8456	152	Long	3
8457	152	Extra Long	4
8458	152	10cm	5
8459	152	20cm	6
8460	152	30cm	7
8461	152	40cm	8
8462	152	50cm	9
8463	152	100cm	10
8464	330	Short	1
8465	330	Medium	2
8466	330	Long	3
8467	330	Extra Long	4
8468	330	10cm	5
8469	330	20cm	6
8470	330	30cm	7
8471	330	40cm	8
8472	330	50cm	9
8473	330	100cm	10
8474	335	Short	1
8475	335	Medium	2
8476	335	Long	3
8477	335	Extra Long	4
8478	335	10cm	5
8479	335	20cm	6
8480	335	30cm	7
8481	335	40cm	8
8482	335	50cm	9
8483	335	100cm	10
8484	344	Short	1
8485	344	Medium	2
8486	344	Long	3
8487	344	Extra Long	4
8488	344	10cm	5
8489	344	20cm	6
8490	344	30cm	7
8491	344	40cm	8
8492	344	50cm	9
8493	344	100cm	10
8494	356	Short	1
8495	356	Medium	2
8496	356	Long	3
8497	356	Extra Long	4
8498	356	10cm	5
8499	356	20cm	6
8500	356	30cm	7
8501	356	40cm	8
8502	356	50cm	9
8503	356	100cm	10
8504	379	Short	1
8505	379	Medium	2
8506	379	Long	3
8507	379	Extra Long	4
8508	379	10cm	5
8509	379	20cm	6
8510	379	30cm	7
8511	379	40cm	8
8512	379	50cm	9
8513	379	100cm	10
8514	414	Short	1
8515	414	Medium	2
8516	414	Long	3
8517	414	Extra Long	4
8518	414	10cm	5
8519	414	20cm	6
8520	414	30cm	7
8521	414	40cm	8
8522	414	50cm	9
8523	414	100cm	10
8524	419	Short	1
8525	419	Medium	2
8526	419	Long	3
8527	419	Extra Long	4
8528	419	10cm	5
8529	419	20cm	6
8530	419	30cm	7
8531	419	40cm	8
8532	419	50cm	9
8533	419	100cm	10
8534	450	Short	1
8535	450	Medium	2
8536	450	Long	3
8537	450	Extra Long	4
8538	450	10cm	5
8539	450	20cm	6
8540	450	30cm	7
8541	450	40cm	8
8542	450	50cm	9
8543	450	100cm	10
8544	506	Short	1
8545	506	Medium	2
8546	506	Long	3
8547	506	Extra Long	4
8548	506	10cm	5
8549	506	20cm	6
8550	506	30cm	7
8551	506	40cm	8
8552	506	50cm	9
8553	506	100cm	10
8554	511	Short	1
8555	511	Medium	2
8556	511	Long	3
8557	511	Extra Long	4
8558	511	10cm	5
8559	511	20cm	6
8560	511	30cm	7
8561	511	40cm	8
8562	511	50cm	9
8563	511	100cm	10
8564	555	Short	1
8565	555	Medium	2
8566	555	Long	3
8567	555	Extra Long	4
8568	555	10cm	5
8569	555	20cm	6
8570	555	30cm	7
8571	555	40cm	8
8572	555	50cm	9
8573	555	100cm	10
8574	559	Short	1
8575	559	Medium	2
8576	559	Long	3
8577	559	Extra Long	4
8578	559	10cm	5
8579	559	20cm	6
8580	559	30cm	7
8581	559	40cm	8
8582	559	50cm	9
8583	559	100cm	10
8584	563	Short	1
8585	563	Medium	2
8586	563	Long	3
8587	563	Extra Long	4
8588	563	10cm	5
8589	563	20cm	6
8590	563	30cm	7
8591	563	40cm	8
8592	563	50cm	9
8593	563	100cm	10
8594	567	Short	1
8595	567	Medium	2
8596	567	Long	3
8597	567	Extra Long	4
8598	567	10cm	5
8599	567	20cm	6
8600	567	30cm	7
8601	567	40cm	8
8602	567	50cm	9
8603	567	100cm	10
8604	571	Short	1
8605	571	Medium	2
8606	571	Long	3
8607	571	Extra Long	4
8608	571	10cm	5
8609	571	20cm	6
8610	571	30cm	7
8611	571	40cm	8
8612	571	50cm	9
8613	571	100cm	10
8614	575	Short	1
8615	575	Medium	2
8616	575	Long	3
8617	575	Extra Long	4
8618	575	10cm	5
8619	575	20cm	6
8620	575	30cm	7
8621	575	40cm	8
8622	575	50cm	9
8623	575	100cm	10
8624	580	Short	1
8625	580	Medium	2
8626	580	Long	3
8627	580	Extra Long	4
8628	580	10cm	5
8629	580	20cm	6
8630	580	30cm	7
8631	580	40cm	8
8632	580	50cm	9
8633	580	100cm	10
8634	584	Short	1
8635	584	Medium	2
8636	584	Long	3
8637	584	Extra Long	4
8638	584	10cm	5
8639	584	20cm	6
8640	584	30cm	7
8641	584	40cm	8
8642	584	50cm	9
8643	584	100cm	10
8644	592	Short	1
8645	592	Medium	2
8646	592	Long	3
8647	592	Extra Long	4
8648	592	10cm	5
8649	592	20cm	6
8650	592	30cm	7
8651	592	40cm	8
8652	592	50cm	9
8653	592	100cm	10
8654	597	Short	1
8655	597	Medium	2
8656	597	Long	3
8657	597	Extra Long	4
8658	597	10cm	5
8659	597	20cm	6
8660	597	30cm	7
8661	597	40cm	8
8662	597	50cm	9
8663	597	100cm	10
8664	602	Short	1
8665	602	Medium	2
8666	602	Long	3
8667	602	Extra Long	4
8668	602	10cm	5
8669	602	20cm	6
8670	602	30cm	7
8671	602	40cm	8
8672	602	50cm	9
8673	602	100cm	10
8674	615	Short	1
8675	615	Medium	2
8676	615	Long	3
8677	615	Extra Long	4
8678	615	10cm	5
8679	615	20cm	6
8680	615	30cm	7
8681	615	40cm	8
8682	615	50cm	9
8683	615	100cm	10
8684	620	Short	1
8685	620	Medium	2
8686	620	Long	3
8687	620	Extra Long	4
8688	620	10cm	5
8689	620	20cm	6
8690	620	30cm	7
8691	620	40cm	8
8692	620	50cm	9
8693	620	100cm	10
8694	626	Short	1
8695	626	Medium	2
8696	626	Long	3
8697	626	Extra Long	4
8698	626	10cm	5
8699	626	20cm	6
8700	626	30cm	7
8701	626	40cm	8
8702	626	50cm	9
8703	626	100cm	10
8704	638	Short	1
8705	638	Medium	2
8706	638	Long	3
8707	638	Extra Long	4
8708	638	10cm	5
8709	638	20cm	6
8710	638	30cm	7
8711	638	40cm	8
8712	638	50cm	9
8713	638	100cm	10
8714	643	Short	1
8715	643	Medium	2
8716	643	Long	3
8717	643	Extra Long	4
8718	643	10cm	5
8719	643	20cm	6
8720	643	30cm	7
8721	643	40cm	8
8722	643	50cm	9
8723	643	100cm	10
8724	777	Short	1
8725	777	Medium	2
8726	777	Long	3
8727	777	Extra Long	4
8728	777	10cm	5
8729	777	20cm	6
8730	777	30cm	7
8731	777	40cm	8
8732	777	50cm	9
8733	777	100cm	10
8734	782	Short	1
8735	782	Medium	2
8736	782	Long	3
8737	782	Extra Long	4
8738	782	10cm	5
8739	782	20cm	6
8740	782	30cm	7
8741	782	40cm	8
8742	782	50cm	9
8743	782	100cm	10
8744	787	Short	1
8745	787	Medium	2
8746	787	Long	3
8747	787	Extra Long	4
8748	787	10cm	5
8749	787	20cm	6
8750	787	30cm	7
8751	787	40cm	8
8752	787	50cm	9
8753	787	100cm	10
8754	793	Short	1
8755	793	Medium	2
8756	793	Long	3
8757	793	Extra Long	4
8758	793	10cm	5
8759	793	20cm	6
8760	793	30cm	7
8761	793	40cm	8
8762	793	50cm	9
8763	793	100cm	10
8764	798	Short	1
8765	798	Medium	2
8766	798	Long	3
8767	798	Extra Long	4
8768	798	10cm	5
8769	798	20cm	6
8770	798	30cm	7
8771	798	40cm	8
8772	798	50cm	9
8773	798	100cm	10
8774	1466	Short	1
8775	1466	Medium	2
8776	1466	Long	3
8777	1466	Extra Long	4
8778	1466	10cm	5
8779	1466	20cm	6
8780	1466	30cm	7
8781	1466	40cm	8
8782	1466	50cm	9
8783	1466	100cm	10
8784	1984	Short	1
8785	1984	Medium	2
8786	1984	Long	3
8787	1984	Extra Long	4
8788	1984	10cm	5
8789	1984	20cm	6
8790	1984	30cm	7
8791	1984	40cm	8
8792	1984	50cm	9
8793	1984	100cm	10
8794	2038	Short	1
8795	2038	Medium	2
8796	2038	Long	3
8797	2038	Extra Long	4
8798	2038	10cm	5
8799	2038	20cm	6
8800	2038	30cm	7
8801	2038	40cm	8
8802	2038	50cm	9
8803	2038	100cm	10
8804	2050	Short	1
8805	2050	Medium	2
8806	2050	Long	3
8807	2050	Extra Long	4
8808	2050	10cm	5
8809	2050	20cm	6
8810	2050	30cm	7
8811	2050	40cm	8
8812	2050	50cm	9
8813	2050	100cm	10
8814	2054	Short	1
8815	2054	Medium	2
8816	2054	Long	3
8817	2054	Extra Long	4
8818	2054	10cm	5
8819	2054	20cm	6
8820	2054	30cm	7
8821	2054	40cm	8
8822	2054	50cm	9
8823	2054	100cm	10
8824	2107	Short	1
8825	2107	Medium	2
8826	2107	Long	3
8827	2107	Extra Long	4
8828	2107	10cm	5
8829	2107	20cm	6
8830	2107	30cm	7
8831	2107	40cm	8
8832	2107	50cm	9
8833	2107	100cm	10
8834	2181	Short	1
8835	2181	Medium	2
8836	2181	Long	3
8837	2181	Extra Long	4
8838	2181	10cm	5
8839	2181	20cm	6
8840	2181	30cm	7
8841	2181	40cm	8
8842	2181	50cm	9
8843	2181	100cm	10
8844	2210	Short	1
8845	2210	Medium	2
8846	2210	Long	3
8847	2210	Extra Long	4
8848	2210	10cm	5
8849	2210	20cm	6
8850	2210	30cm	7
8851	2210	40cm	8
8852	2210	50cm	9
8853	2210	100cm	10
8854	2216	Short	1
8855	2216	Medium	2
8856	2216	Long	3
8857	2216	Extra Long	4
8858	2216	10cm	5
8859	2216	20cm	6
8860	2216	30cm	7
8861	2216	40cm	8
8862	2216	50cm	9
8863	2216	100cm	10
8864	2228	Short	1
8865	2228	Medium	2
8866	2228	Long	3
8867	2228	Extra Long	4
8868	2228	10cm	5
8869	2228	20cm	6
8870	2228	30cm	7
8871	2228	40cm	8
8872	2228	50cm	9
8873	2228	100cm	10
8874	2284	Short	1
8875	2284	Medium	2
8876	2284	Long	3
8877	2284	Extra Long	4
8878	2284	10cm	5
8879	2284	20cm	6
8880	2284	30cm	7
8881	2284	40cm	8
8882	2284	50cm	9
8883	2284	100cm	10
8884	2316	Short	1
8885	2316	Medium	2
8886	2316	Long	3
8887	2316	Extra Long	4
8888	2316	10cm	5
8889	2316	20cm	6
8890	2316	30cm	7
8891	2316	40cm	8
8892	2316	50cm	9
8893	2316	100cm	10
8894	2395	Short	1
8895	2395	Medium	2
8896	2395	Long	3
8897	2395	Extra Long	4
8898	2395	10cm	5
8899	2395	20cm	6
8900	2395	30cm	7
8901	2395	40cm	8
8902	2395	50cm	9
8903	2395	100cm	10
8904	2403	Short	1
8905	2403	Medium	2
8906	2403	Long	3
8907	2403	Extra Long	4
8908	2403	10cm	5
8909	2403	20cm	6
8910	2403	30cm	7
8911	2403	40cm	8
8912	2403	50cm	9
8913	2403	100cm	10
8914	2408	Short	1
8915	2408	Medium	2
8916	2408	Long	3
8917	2408	Extra Long	4
8918	2408	10cm	5
8919	2408	20cm	6
8920	2408	30cm	7
8921	2408	40cm	8
8922	2408	50cm	9
8923	2408	100cm	10
8924	2550	Short	1
8925	2550	Medium	2
8926	2550	Long	3
8927	2550	Extra Long	4
8928	2550	10cm	5
8929	2550	20cm	6
8930	2550	30cm	7
8931	2550	40cm	8
8932	2550	50cm	9
8933	2550	100cm	10
8934	860	No Storage	1
8935	860	With Drawers	2
8936	860	With Shelves	3
8937	860	With Cabinet	4
8938	860	Multiple Compartments	5
8939	860	Compact Storage	6
8940	860	Large Storage	7
8941	870	No Storage	1
8942	870	With Drawers	2
8943	870	With Shelves	3
8944	870	With Cabinet	4
8945	870	Multiple Compartments	5
8946	870	Compact Storage	6
8947	870	Large Storage	7
8948	875	No Storage	1
8949	875	With Drawers	2
8950	875	With Shelves	3
8951	875	With Cabinet	4
8952	875	Multiple Compartments	5
8953	875	Compact Storage	6
8954	875	Large Storage	7
8955	879	No Storage	1
8956	879	With Drawers	2
8957	879	With Shelves	3
8958	879	With Cabinet	4
8959	879	Multiple Compartments	5
8960	879	Compact Storage	6
8961	879	Large Storage	7
8962	884	No Storage	1
8963	884	With Drawers	2
8964	884	With Shelves	3
8965	884	With Cabinet	4
8966	884	Multiple Compartments	5
8967	884	Compact Storage	6
8968	884	Large Storage	7
8969	929	No Storage	1
8970	929	With Drawers	2
8971	929	With Shelves	3
8972	929	With Cabinet	4
8973	929	Multiple Compartments	5
8974	929	Compact Storage	6
8975	929	Large Storage	7
8976	1042	No Storage	1
8977	1042	With Drawers	2
8978	1042	With Shelves	3
8979	1042	With Cabinet	4
8980	1042	Multiple Compartments	5
8981	1042	Compact Storage	6
8982	1042	Large Storage	7
8983	1047	No Storage	1
8984	1047	With Drawers	2
8985	1047	With Shelves	3
8986	1047	With Cabinet	4
8987	1047	Multiple Compartments	5
8988	1047	Compact Storage	6
8989	1047	Large Storage	7
8990	1051	No Storage	1
8991	1051	With Drawers	2
8992	1051	With Shelves	3
8993	1051	With Cabinet	4
8994	1051	Multiple Compartments	5
8995	1051	Compact Storage	6
8996	1051	Large Storage	7
8997	1056	No Storage	1
8998	1056	With Drawers	2
8999	1056	With Shelves	3
9000	1056	With Cabinet	4
9001	1056	Multiple Compartments	5
9002	1056	Compact Storage	6
9003	1056	Large Storage	7
9004	1088	No Storage	1
9005	1088	With Drawers	2
9006	1088	With Shelves	3
9007	1088	With Cabinet	4
9008	1088	Multiple Compartments	5
9009	1088	Compact Storage	6
9010	1088	Large Storage	7
9011	1142	No Storage	1
9012	1142	With Drawers	2
9013	1142	With Shelves	3
9014	1142	With Cabinet	4
9015	1142	Multiple Compartments	5
9016	1142	Compact Storage	6
9017	1142	Large Storage	7
9018	1147	No Storage	1
9019	1147	With Drawers	2
9020	1147	With Shelves	3
9021	1147	With Cabinet	4
9022	1147	Multiple Compartments	5
9023	1147	Compact Storage	6
9024	1147	Large Storage	7
9025	1176	No Storage	1
9026	1176	With Drawers	2
9027	1176	With Shelves	3
9028	1176	With Cabinet	4
9029	1176	Multiple Compartments	5
9030	1176	Compact Storage	6
9031	1176	Large Storage	7
9032	1180	No Storage	1
9033	1180	With Drawers	2
9034	1180	With Shelves	3
9035	1180	With Cabinet	4
9036	1180	Multiple Compartments	5
9037	1180	Compact Storage	6
9038	1180	Large Storage	7
9039	1186	No Storage	1
9040	1186	With Drawers	2
9041	1186	With Shelves	3
9042	1186	With Cabinet	4
9043	1186	Multiple Compartments	5
9044	1186	Compact Storage	6
9045	1186	Large Storage	7
9046	1191	No Storage	1
9047	1191	With Drawers	2
9048	1191	With Shelves	3
9049	1191	With Cabinet	4
9050	1191	Multiple Compartments	5
9051	1191	Compact Storage	6
9052	1191	Large Storage	7
9053	1196	No Storage	1
9054	1196	With Drawers	2
9055	1196	With Shelves	3
9056	1196	With Cabinet	4
9057	1196	Multiple Compartments	5
9058	1196	Compact Storage	6
9059	1196	Large Storage	7
9060	1201	No Storage	1
9061	1201	With Drawers	2
9062	1201	With Shelves	3
9063	1201	With Cabinet	4
9064	1201	Multiple Compartments	5
9065	1201	Compact Storage	6
9066	1201	Large Storage	7
9067	1206	No Storage	1
9068	1206	With Drawers	2
9069	1206	With Shelves	3
9070	1206	With Cabinet	4
9071	1206	Multiple Compartments	5
9072	1206	Compact Storage	6
9073	1206	Large Storage	7
9074	1211	No Storage	1
9075	1211	With Drawers	2
9076	1211	With Shelves	3
9077	1211	With Cabinet	4
9078	1211	Multiple Compartments	5
9079	1211	Compact Storage	6
9080	1211	Large Storage	7
9081	1215	No Storage	1
9082	1215	With Drawers	2
9083	1215	With Shelves	3
9084	1215	With Cabinet	4
9085	1215	Multiple Compartments	5
9086	1215	Compact Storage	6
9087	1215	Large Storage	7
9088	1221	No Storage	1
9089	1221	With Drawers	2
9090	1221	With Shelves	3
9091	1221	With Cabinet	4
9092	1221	Multiple Compartments	5
9093	1221	Compact Storage	6
9094	1221	Large Storage	7
9095	1226	No Storage	1
9096	1226	With Drawers	2
9097	1226	With Shelves	3
9098	1226	With Cabinet	4
9099	1226	Multiple Compartments	5
9100	1226	Compact Storage	6
9101	1226	Large Storage	7
9102	1231	No Storage	1
9103	1231	With Drawers	2
9104	1231	With Shelves	3
9105	1231	With Cabinet	4
9106	1231	Multiple Compartments	5
9107	1231	Compact Storage	6
9108	1231	Large Storage	7
9109	1238	No Storage	1
9110	1238	With Drawers	2
9111	1238	With Shelves	3
9112	1238	With Cabinet	4
9113	1238	Multiple Compartments	5
9114	1238	Compact Storage	6
9115	1238	Large Storage	7
9116	1244	No Storage	1
9117	1244	With Drawers	2
9118	1244	With Shelves	3
9119	1244	With Cabinet	4
9120	1244	Multiple Compartments	5
9121	1244	Compact Storage	6
9122	1244	Large Storage	7
9123	1251	No Storage	1
9124	1251	With Drawers	2
9125	1251	With Shelves	3
9126	1251	With Cabinet	4
9127	1251	Multiple Compartments	5
9128	1251	Compact Storage	6
9129	1251	Large Storage	7
9130	1257	No Storage	1
9131	1257	With Drawers	2
9132	1257	With Shelves	3
9133	1257	With Cabinet	4
9134	1257	Multiple Compartments	5
9135	1257	Compact Storage	6
9136	1257	Large Storage	7
9137	1263	No Storage	1
9138	1263	With Drawers	2
9139	1263	With Shelves	3
9140	1263	With Cabinet	4
9141	1263	Multiple Compartments	5
9142	1263	Compact Storage	6
9143	1263	Large Storage	7
9144	1269	No Storage	1
9145	1269	With Drawers	2
9146	1269	With Shelves	3
9147	1269	With Cabinet	4
9148	1269	Multiple Compartments	5
9149	1269	Compact Storage	6
9150	1269	Large Storage	7
9151	1275	No Storage	1
9152	1275	With Drawers	2
9153	1275	With Shelves	3
9154	1275	With Cabinet	4
9155	1275	Multiple Compartments	5
9156	1275	Compact Storage	6
9157	1275	Large Storage	7
9158	1280	No Storage	1
9159	1280	With Drawers	2
9160	1280	With Shelves	3
9161	1280	With Cabinet	4
9162	1280	Multiple Compartments	5
9163	1280	Compact Storage	6
9164	1280	Large Storage	7
9165	1286	No Storage	1
9166	1286	With Drawers	2
9167	1286	With Shelves	3
9168	1286	With Cabinet	4
9169	1286	Multiple Compartments	5
9170	1286	Compact Storage	6
9171	1286	Large Storage	7
9172	1415	No Storage	1
9173	1415	With Drawers	2
9174	1415	With Shelves	3
9175	1415	With Cabinet	4
9176	1415	Multiple Compartments	5
9177	1415	Compact Storage	6
9178	1415	Large Storage	7
9179	1419	No Storage	1
9180	1419	With Drawers	2
9181	1419	With Shelves	3
9182	1419	With Cabinet	4
9183	1419	Multiple Compartments	5
9184	1419	Compact Storage	6
9185	1419	Large Storage	7
9186	1423	No Storage	1
9187	1423	With Drawers	2
9188	1423	With Shelves	3
9189	1423	With Cabinet	4
9190	1423	Multiple Compartments	5
9191	1423	Compact Storage	6
9192	1423	Large Storage	7
9193	541	Grade A	1
9194	541	Grade B	2
9195	541	Grade C	3
9196	541	Premium	4
9197	541	Standard	5
9198	541	Economy	6
9199	541	Superior	7
9200	556	Grade A	1
9201	556	Grade B	2
9202	556	Grade C	3
9203	556	Premium	4
9204	556	Standard	5
9205	556	Economy	6
9206	556	Superior	7
9207	560	Grade A	1
9208	560	Grade B	2
9209	560	Grade C	3
9210	560	Premium	4
9211	560	Standard	5
9212	560	Economy	6
9213	560	Superior	7
9214	564	Grade A	1
9215	564	Grade B	2
9216	564	Grade C	3
9217	564	Premium	4
9218	564	Standard	5
9219	564	Economy	6
9220	564	Superior	7
9221	568	Grade A	1
9222	568	Grade B	2
9223	568	Grade C	3
9224	568	Premium	4
9225	568	Standard	5
9226	568	Economy	6
9227	568	Superior	7
9228	572	Grade A	1
9229	572	Grade B	2
9230	572	Grade C	3
9231	572	Premium	4
9232	572	Standard	5
9233	572	Economy	6
9234	572	Superior	7
9235	576	Grade A	1
9236	576	Grade B	2
9237	576	Grade C	3
9238	576	Premium	4
9239	576	Standard	5
9240	576	Economy	6
9241	576	Superior	7
9242	588	Grade A	1
9243	588	Grade B	2
9244	588	Grade C	3
9245	588	Premium	4
9246	588	Standard	5
9247	588	Economy	6
9248	588	Superior	7
9249	593	Grade A	1
9250	593	Grade B	2
9251	593	Grade C	3
9252	593	Premium	4
9253	593	Standard	5
9254	593	Economy	6
9255	593	Superior	7
9256	598	Grade A	1
9257	598	Grade B	2
9258	598	Grade C	3
9259	598	Premium	4
9260	598	Standard	5
9261	598	Economy	6
9262	598	Superior	7
9263	603	Grade A	1
9264	603	Grade B	2
9265	603	Grade C	3
9266	603	Premium	4
9267	603	Standard	5
9268	603	Economy	6
9269	603	Superior	7
9270	606	Grade A	1
9271	606	Grade B	2
9272	606	Grade C	3
9273	606	Premium	4
9274	606	Standard	5
9275	606	Economy	6
9276	606	Superior	7
9277	616	Grade A	1
9278	616	Grade B	2
9279	616	Grade C	3
9280	616	Premium	4
9281	616	Standard	5
9282	616	Economy	6
9283	616	Superior	7
9284	621	Grade A	1
9285	621	Grade B	2
9286	621	Grade C	3
9287	621	Premium	4
9288	621	Standard	5
9289	621	Economy	6
9290	621	Superior	7
9291	635	Grade A	1
9292	635	Grade B	2
9293	635	Grade C	3
9294	635	Premium	4
9295	635	Standard	5
9296	635	Economy	6
9297	635	Superior	7
9298	668	Grade A	1
9299	668	Grade B	2
9300	668	Grade C	3
9301	668	Premium	4
9302	668	Standard	5
9303	668	Economy	6
9304	668	Superior	7
9305	690	Grade A	1
9306	690	Grade B	2
9307	690	Grade C	3
9308	690	Premium	4
9309	690	Standard	5
9310	690	Economy	6
9311	690	Superior	7
9312	725	Grade A	1
9313	725	Grade B	2
9314	725	Grade C	3
9315	725	Premium	4
9316	725	Standard	5
9317	725	Economy	6
9318	725	Superior	7
9319	729	Grade A	1
9320	729	Grade B	2
9321	729	Grade C	3
9322	729	Premium	4
9323	729	Standard	5
9324	729	Economy	6
9325	729	Superior	7
9326	733	Grade A	1
9327	733	Grade B	2
9328	733	Grade C	3
9329	733	Premium	4
9330	733	Standard	5
9331	733	Economy	6
9332	733	Superior	7
9333	752	Grade A	1
9334	752	Grade B	2
9335	752	Grade C	3
9336	752	Premium	4
9337	752	Standard	5
9338	752	Economy	6
9339	752	Superior	7
9340	779	Grade A	1
9341	779	Grade B	2
9342	779	Grade C	3
9343	779	Premium	4
9344	779	Standard	5
9345	779	Economy	6
9346	779	Superior	7
9347	784	Grade A	1
9348	784	Grade B	2
9349	784	Grade C	3
9350	784	Premium	4
9351	784	Standard	5
9352	784	Economy	6
9353	784	Superior	7
9354	789	Grade A	1
9355	789	Grade B	2
9356	789	Grade C	3
9357	789	Premium	4
9358	789	Standard	5
9359	789	Economy	6
9360	789	Superior	7
9361	794	Grade A	1
9362	794	Grade B	2
9363	794	Grade C	3
9364	794	Premium	4
9365	794	Standard	5
9366	794	Economy	6
9367	794	Superior	7
9368	799	Grade A	1
9369	799	Grade B	2
9370	799	Grade C	3
9371	799	Premium	4
9372	799	Standard	5
9373	799	Economy	6
9374	799	Superior	7
9375	812	Grade A	1
9376	812	Grade B	2
9377	812	Grade C	3
9378	812	Premium	4
9379	812	Standard	5
9380	812	Economy	6
9381	812	Superior	7
9382	1617	Grade A	1
9383	1617	Grade B	2
9384	1617	Grade C	3
9385	1617	Premium	4
9386	1617	Standard	5
9387	1617	Economy	6
9388	1617	Superior	7
9389	1621	Grade A	1
9390	1621	Grade B	2
9391	1621	Grade C	3
9392	1621	Premium	4
9393	1621	Standard	5
9394	1621	Economy	6
9395	1621	Superior	7
9396	1862	Grade A	1
9397	1862	Grade B	2
9398	1862	Grade C	3
9399	1862	Premium	4
9400	1862	Standard	5
9401	1862	Economy	6
9402	1862	Superior	7
9403	1866	Grade A	1
9404	1866	Grade B	2
9405	1866	Grade C	3
9406	1866	Premium	4
9407	1866	Standard	5
9408	1866	Economy	6
9409	1866	Superior	7
9410	1870	Grade A	1
9411	1870	Grade B	2
9412	1870	Grade C	3
9413	1870	Premium	4
9414	1870	Standard	5
9415	1870	Economy	6
9416	1870	Superior	7
9417	1874	Grade A	1
9418	1874	Grade B	2
9419	1874	Grade C	3
9420	1874	Premium	4
9421	1874	Standard	5
9422	1874	Economy	6
9423	1874	Superior	7
9424	1925	Grade A	1
9425	1925	Grade B	2
9426	1925	Grade C	3
9427	1925	Premium	4
9428	1925	Standard	5
9429	1925	Economy	6
9430	1925	Superior	7
9431	113	4 inch	1
9432	113	5 inch	2
9433	113	6 inch	3
9434	113	8 inch	4
9435	113	10 inch	5
9436	113	Thin	6
9437	113	Medium	7
9438	113	Thick	8
9439	113	Extra Thick	9
9440	579	4 inch	1
9441	579	5 inch	2
9442	579	6 inch	3
9443	579	8 inch	4
9444	579	10 inch	5
9445	579	Thin	6
9446	579	Medium	7
9447	579	Thick	8
9448	579	Extra Thick	9
9449	583	4 inch	1
9450	583	5 inch	2
9451	583	6 inch	3
9452	583	8 inch	4
9453	583	10 inch	5
9454	583	Thin	6
9455	583	Medium	7
9456	583	Thick	8
9457	583	Extra Thick	9
9458	586	4 inch	1
9459	586	5 inch	2
9460	586	6 inch	3
9461	586	8 inch	4
9462	586	10 inch	5
9463	586	Thin	6
9464	586	Medium	7
9465	586	Thick	8
9466	586	Extra Thick	9
9467	590	4 inch	1
9468	590	5 inch	2
9469	590	6 inch	3
9470	590	8 inch	4
9471	590	10 inch	5
9472	590	Thin	6
9473	590	Medium	7
9474	590	Thick	8
9475	590	Extra Thick	9
9476	595	4 inch	1
9477	595	5 inch	2
9478	595	6 inch	3
9479	595	8 inch	4
9480	595	10 inch	5
9481	595	Thin	6
9482	595	Medium	7
9483	595	Thick	8
9484	595	Extra Thick	9
9485	600	4 inch	1
9486	600	5 inch	2
9487	600	6 inch	3
9488	600	8 inch	4
9489	600	10 inch	5
9490	600	Thin	6
9491	600	Medium	7
9492	600	Thick	8
9493	600	Extra Thick	9
9494	605	4 inch	1
9495	605	5 inch	2
9496	605	6 inch	3
9497	605	8 inch	4
9498	605	10 inch	5
9499	605	Thin	6
9500	605	Medium	7
9501	605	Thick	8
9502	605	Extra Thick	9
9503	610	4 inch	1
9504	610	5 inch	2
9505	610	6 inch	3
9506	610	8 inch	4
9507	610	10 inch	5
9508	610	Thin	6
9509	610	Medium	7
9510	610	Thick	8
9511	610	Extra Thick	9
9512	618	4 inch	1
9513	618	5 inch	2
9514	618	6 inch	3
9515	618	8 inch	4
9516	618	10 inch	5
9517	618	Thin	6
9518	618	Medium	7
9519	618	Thick	8
9520	618	Extra Thick	9
9521	624	4 inch	1
9522	624	5 inch	2
9523	624	6 inch	3
9524	624	8 inch	4
9525	624	10 inch	5
9526	624	Thin	6
9527	624	Medium	7
9528	624	Thick	8
9529	624	Extra Thick	9
9530	628	4 inch	1
9531	628	5 inch	2
9532	628	6 inch	3
9533	628	8 inch	4
9534	628	10 inch	5
9535	628	Thin	6
9536	628	Medium	7
9537	628	Thick	8
9538	628	Extra Thick	9
9539	633	4 inch	1
9540	633	5 inch	2
9541	633	6 inch	3
9542	633	8 inch	4
9543	633	10 inch	5
9544	633	Thin	6
9545	633	Medium	7
9546	633	Thick	8
9547	633	Extra Thick	9
9548	699	4 inch	1
9549	699	5 inch	2
9550	699	6 inch	3
9551	699	8 inch	4
9552	699	10 inch	5
9553	699	Thin	6
9554	699	Medium	7
9555	699	Thick	8
9556	699	Extra Thick	9
9557	704	4 inch	1
9558	704	5 inch	2
9559	704	6 inch	3
9560	704	8 inch	4
9561	704	10 inch	5
9562	704	Thin	6
9563	704	Medium	7
9564	704	Thick	8
9565	704	Extra Thick	9
9566	709	4 inch	1
9567	709	5 inch	2
9568	709	6 inch	3
9569	709	8 inch	4
9570	709	10 inch	5
9571	709	Thin	6
9572	709	Medium	7
9573	709	Thick	8
9574	709	Extra Thick	9
9575	723	4 inch	1
9576	723	5 inch	2
9577	723	6 inch	3
9578	723	8 inch	4
9579	723	10 inch	5
9580	723	Thin	6
9581	723	Medium	7
9582	723	Thick	8
9583	723	Extra Thick	9
9584	727	4 inch	1
9585	727	5 inch	2
9586	727	6 inch	3
9587	727	8 inch	4
9588	727	10 inch	5
9589	727	Thin	6
9590	727	Medium	7
9591	727	Thick	8
9592	727	Extra Thick	9
9593	731	4 inch	1
9594	731	5 inch	2
9595	731	6 inch	3
9596	731	8 inch	4
9597	731	10 inch	5
9598	731	Thin	6
9599	731	Medium	7
9600	731	Thick	8
9601	731	Extra Thick	9
9602	735	4 inch	1
9603	735	5 inch	2
9604	735	6 inch	3
9605	735	8 inch	4
9606	735	10 inch	5
9607	735	Thin	6
9608	735	Medium	7
9609	735	Thick	8
9610	735	Extra Thick	9
9611	740	4 inch	1
9612	740	5 inch	2
9613	740	6 inch	3
9614	740	8 inch	4
9615	740	10 inch	5
9616	740	Thin	6
9617	740	Medium	7
9618	740	Thick	8
9619	740	Extra Thick	9
9620	744	4 inch	1
9621	744	5 inch	2
9622	744	6 inch	3
9623	744	8 inch	4
9624	744	10 inch	5
9625	744	Thin	6
9626	744	Medium	7
9627	744	Thick	8
9628	744	Extra Thick	9
9629	765	4 inch	1
9630	765	5 inch	2
9631	765	6 inch	3
9632	765	8 inch	4
9633	765	10 inch	5
9634	765	Thin	6
9635	765	Medium	7
9636	765	Thick	8
9637	765	Extra Thick	9
9638	768	4 inch	1
9639	768	5 inch	2
9640	768	6 inch	3
9641	768	8 inch	4
9642	768	10 inch	5
9643	768	Thin	6
9644	768	Medium	7
9645	768	Thick	8
9646	768	Extra Thick	9
9647	772	4 inch	1
9648	772	5 inch	2
9649	772	6 inch	3
9650	772	8 inch	4
9651	772	10 inch	5
9652	772	Thin	6
9653	772	Medium	7
9654	772	Thick	8
9655	772	Extra Thick	9
9656	792	4 inch	1
9657	792	5 inch	2
9658	792	6 inch	3
9659	792	8 inch	4
9660	792	10 inch	5
9661	792	Thin	6
9662	792	Medium	7
9663	792	Thick	8
9664	792	Extra Thick	9
9665	796	4 inch	1
9666	796	5 inch	2
9667	796	6 inch	3
9668	796	8 inch	4
9669	796	10 inch	5
9670	796	Thin	6
9671	796	Medium	7
9672	796	Thick	8
9673	796	Extra Thick	9
9674	810	4 inch	1
9675	810	5 inch	2
9676	810	6 inch	3
9677	810	8 inch	4
9678	810	10 inch	5
9679	810	Thin	6
9680	810	Medium	7
9681	810	Thick	8
9682	810	Extra Thick	9
9683	992	4 inch	1
9684	992	5 inch	2
9685	992	6 inch	3
9686	992	8 inch	4
9687	992	10 inch	5
9688	992	Thin	6
9689	992	Medium	7
9690	992	Thick	8
9691	992	Extra Thick	9
9692	997	4 inch	1
9693	997	5 inch	2
9694	997	6 inch	3
9695	997	8 inch	4
9696	997	10 inch	5
9697	997	Thin	6
9698	997	Medium	7
9699	997	Thick	8
9700	997	Extra Thick	9
9701	1003	4 inch	1
9702	1003	5 inch	2
9703	1003	6 inch	3
9704	1003	8 inch	4
9705	1003	10 inch	5
9706	1003	Thin	6
9707	1003	Medium	7
9708	1003	Thick	8
9709	1003	Extra Thick	9
9710	1006	4 inch	1
9711	1006	5 inch	2
9712	1006	6 inch	3
9713	1006	8 inch	4
9714	1006	10 inch	5
9715	1006	Thin	6
9716	1006	Medium	7
9717	1006	Thick	8
9718	1006	Extra Thick	9
9719	1010	4 inch	1
9720	1010	5 inch	2
9721	1010	6 inch	3
9722	1010	8 inch	4
9723	1010	10 inch	5
9724	1010	Thin	6
9725	1010	Medium	7
9726	1010	Thick	8
9727	1010	Extra Thick	9
9728	1014	4 inch	1
9729	1014	5 inch	2
9730	1014	6 inch	3
9731	1014	8 inch	4
9732	1014	10 inch	5
9733	1014	Thin	6
9734	1014	Medium	7
9735	1014	Thick	8
9736	1014	Extra Thick	9
9737	1175	2GB	1
9738	1185	2GB	1
9739	1190	2GB	1
9740	1195	2GB	1
9741	1200	2GB	1
9742	1205	2GB	1
9743	1210	2GB	1
9744	1220	2GB	1
9745	1225	2GB	1
9746	1230	2GB	1
9747	1237	2GB	1
9748	1243	2GB	1
9749	1250	2GB	1
9750	1256	2GB	1
9751	1262	2GB	1
9752	1268	2GB	1
9753	1274	2GB	1
9754	1279	2GB	1
9755	1285	2GB	1
9756	1175	3GB	2
9757	1185	3GB	2
9758	1190	3GB	2
9759	1195	3GB	2
9760	1200	3GB	2
9761	1205	3GB	2
9762	1210	3GB	2
9763	1220	3GB	2
9764	1225	3GB	2
9765	1230	3GB	2
9766	1237	3GB	2
9767	1243	3GB	2
9768	1250	3GB	2
9769	1256	3GB	2
9770	1262	3GB	2
9771	1268	3GB	2
9772	1274	3GB	2
9773	1279	3GB	2
9774	1285	3GB	2
9775	1175	4GB	3
9776	1185	4GB	3
9777	1190	4GB	3
9778	1195	4GB	3
9779	1200	4GB	3
9780	1205	4GB	3
9781	1210	4GB	3
9782	1220	4GB	3
9783	1225	4GB	3
9784	1230	4GB	3
9785	1237	4GB	3
9786	1243	4GB	3
9787	1250	4GB	3
9788	1256	4GB	3
9789	1262	4GB	3
9790	1268	4GB	3
9791	1274	4GB	3
9792	1279	4GB	3
9793	1285	4GB	3
9794	1175	6GB	4
9795	1185	6GB	4
9796	1190	6GB	4
9797	1195	6GB	4
9798	1200	6GB	4
9799	1205	6GB	4
9800	1210	6GB	4
9801	1220	6GB	4
9802	1225	6GB	4
9803	1230	6GB	4
9804	1237	6GB	4
9805	1243	6GB	4
9806	1250	6GB	4
9807	1256	6GB	4
9808	1262	6GB	4
9809	1268	6GB	4
9810	1274	6GB	4
9811	1279	6GB	4
9812	1285	6GB	4
9813	1175	8GB	5
9814	1185	8GB	5
9815	1190	8GB	5
9816	1195	8GB	5
9817	1200	8GB	5
9818	1205	8GB	5
9819	1210	8GB	5
9820	1220	8GB	5
9821	1225	8GB	5
9822	1230	8GB	5
9823	1237	8GB	5
9824	1243	8GB	5
9825	1250	8GB	5
9826	1256	8GB	5
9827	1262	8GB	5
9828	1268	8GB	5
9829	1274	8GB	5
9830	1279	8GB	5
9831	1285	8GB	5
9832	1175	12GB	6
9833	1185	12GB	6
9834	1190	12GB	6
9835	1195	12GB	6
9836	1200	12GB	6
9837	1205	12GB	6
9838	1210	12GB	6
9839	1220	12GB	6
9840	1225	12GB	6
9841	1230	12GB	6
9842	1237	12GB	6
9843	1243	12GB	6
9844	1250	12GB	6
9845	1256	12GB	6
9846	1262	12GB	6
9847	1268	12GB	6
9848	1274	12GB	6
9849	1279	12GB	6
9850	1285	12GB	6
9851	1175	16GB	7
9852	1185	16GB	7
9853	1190	16GB	7
9854	1195	16GB	7
9855	1200	16GB	7
9856	1205	16GB	7
9857	1210	16GB	7
9858	1220	16GB	7
9859	1225	16GB	7
9860	1230	16GB	7
9861	1237	16GB	7
9862	1243	16GB	7
9863	1250	16GB	7
9864	1256	16GB	7
9865	1262	16GB	7
9866	1268	16GB	7
9867	1274	16GB	7
9868	1279	16GB	7
9869	1285	16GB	7
9870	1175	32GB	8
9871	1185	32GB	8
9872	1190	32GB	8
9873	1195	32GB	8
9874	1200	32GB	8
9875	1205	32GB	8
9876	1210	32GB	8
9877	1220	32GB	8
9878	1225	32GB	8
9879	1230	32GB	8
9880	1237	32GB	8
9881	1243	32GB	8
9882	1250	32GB	8
9883	1256	32GB	8
9884	1262	32GB	8
9885	1268	32GB	8
9886	1274	32GB	8
9887	1279	32GB	8
9888	1285	32GB	8
9889	1175	64GB	9
9890	1185	64GB	9
9891	1190	64GB	9
9892	1195	64GB	9
9893	1200	64GB	9
9894	1205	64GB	9
9895	1210	64GB	9
9896	1220	64GB	9
9897	1225	64GB	9
9898	1230	64GB	9
9899	1237	64GB	9
9900	1243	64GB	9
9901	1250	64GB	9
9902	1256	64GB	9
9903	1262	64GB	9
9904	1268	64GB	9
9905	1274	64GB	9
9906	1279	64GB	9
9907	1285	64GB	9
9908	1212	32 inch	1
9909	1216	32 inch	1
9910	1222	32 inch	1
9911	1233	32 inch	1
9912	1239	32 inch	1
9913	1245	32 inch	1
9914	1252	32 inch	1
9915	1258	32 inch	1
9916	1264	32 inch	1
9917	1281	32 inch	1
9918	1290	32 inch	1
9919	1296	32 inch	1
9920	1302	32 inch	1
9921	1323	32 inch	1
9922	1496	32 inch	1
9923	2524	32 inch	1
9924	1212	40 inch	2
9925	1216	40 inch	2
9926	1222	40 inch	2
9927	1233	40 inch	2
9928	1239	40 inch	2
9929	1245	40 inch	2
9930	1252	40 inch	2
9931	1258	40 inch	2
9932	1264	40 inch	2
9933	1281	40 inch	2
9934	1290	40 inch	2
9935	1296	40 inch	2
9936	1302	40 inch	2
9937	1323	40 inch	2
9938	1496	40 inch	2
9939	2524	40 inch	2
9940	1212	43 inch	3
9941	1216	43 inch	3
9942	1222	43 inch	3
9943	1233	43 inch	3
9944	1239	43 inch	3
9945	1245	43 inch	3
9946	1252	43 inch	3
9947	1258	43 inch	3
9948	1264	43 inch	3
9949	1281	43 inch	3
9950	1290	43 inch	3
9951	1296	43 inch	3
9952	1302	43 inch	3
9953	1323	43 inch	3
9954	1496	43 inch	3
9955	2524	43 inch	3
9956	1212	50 inch	4
9957	1216	50 inch	4
9958	1222	50 inch	4
9959	1233	50 inch	4
9960	1239	50 inch	4
9961	1245	50 inch	4
9962	1252	50 inch	4
9963	1258	50 inch	4
9964	1264	50 inch	4
9965	1281	50 inch	4
9966	1290	50 inch	4
9967	1296	50 inch	4
9968	1302	50 inch	4
9969	1323	50 inch	4
9970	1496	50 inch	4
9971	2524	50 inch	4
9972	1212	55 inch	5
9973	1216	55 inch	5
9974	1222	55 inch	5
9975	1233	55 inch	5
9976	1239	55 inch	5
9977	1245	55 inch	5
9978	1252	55 inch	5
9979	1258	55 inch	5
9980	1264	55 inch	5
9981	1281	55 inch	5
9982	1290	55 inch	5
9983	1296	55 inch	5
9984	1302	55 inch	5
9985	1323	55 inch	5
9986	1496	55 inch	5
9987	2524	55 inch	5
9988	1212	65 inch	6
9989	1216	65 inch	6
9990	1222	65 inch	6
9991	1233	65 inch	6
9992	1239	65 inch	6
9993	1245	65 inch	6
9994	1252	65 inch	6
9995	1258	65 inch	6
9996	1264	65 inch	6
9997	1281	65 inch	6
9998	1290	65 inch	6
9999	1296	65 inch	6
10000	1302	65 inch	6
10001	1323	65 inch	6
10002	1496	65 inch	6
10003	2524	65 inch	6
10004	1212	75 inch	7
10005	1216	75 inch	7
10006	1222	75 inch	7
10007	1233	75 inch	7
10008	1239	75 inch	7
10009	1245	75 inch	7
10010	1252	75 inch	7
10011	1258	75 inch	7
10012	1264	75 inch	7
10013	1281	75 inch	7
10014	1290	75 inch	7
10015	1296	75 inch	7
10016	1302	75 inch	7
10017	1323	75 inch	7
10018	1496	75 inch	7
10019	2524	75 inch	7
10020	1212	85 inch	8
10021	1216	85 inch	8
10022	1222	85 inch	8
10023	1233	85 inch	8
10024	1239	85 inch	8
10025	1245	85 inch	8
10026	1252	85 inch	8
10027	1258	85 inch	8
10028	1264	85 inch	8
10029	1281	85 inch	8
10030	1290	85 inch	8
10031	1296	85 inch	8
10032	1302	85 inch	8
10033	1323	85 inch	8
10034	1496	85 inch	8
10035	2524	85 inch	8
10036	1212	10 inch	9
10037	1216	10 inch	9
10038	1222	10 inch	9
10039	1233	10 inch	9
10040	1239	10 inch	9
10041	1245	10 inch	9
10042	1252	10 inch	9
10043	1258	10 inch	9
10044	1264	10 inch	9
10045	1281	10 inch	9
10046	1290	10 inch	9
10047	1296	10 inch	9
10048	1302	10 inch	9
10049	1323	10 inch	9
10050	1496	10 inch	9
10051	2524	10 inch	9
10052	1212	13 inch	10
10053	1216	13 inch	10
10054	1222	13 inch	10
10055	1233	13 inch	10
10056	1239	13 inch	10
10057	1245	13 inch	10
10058	1252	13 inch	10
10059	1258	13 inch	10
10060	1264	13 inch	10
10061	1281	13 inch	10
10062	1290	13 inch	10
10063	1296	13 inch	10
10064	1302	13 inch	10
10065	1323	13 inch	10
10066	1496	13 inch	10
10067	2524	13 inch	10
10068	1212	15 inch	11
10069	1216	15 inch	11
10070	1222	15 inch	11
10071	1233	15 inch	11
10072	1239	15 inch	11
10073	1245	15 inch	11
10074	1252	15 inch	11
10075	1258	15 inch	11
10076	1264	15 inch	11
10077	1281	15 inch	11
10078	1290	15 inch	11
10079	1296	15 inch	11
10080	1302	15 inch	11
10081	1323	15 inch	11
10082	1496	15 inch	11
10083	2524	15 inch	11
10084	1212	17 inch	12
10085	1216	17 inch	12
10086	1222	17 inch	12
10087	1233	17 inch	12
10088	1239	17 inch	12
10089	1245	17 inch	12
10090	1252	17 inch	12
10091	1258	17 inch	12
10092	1264	17 inch	12
10093	1281	17 inch	12
10094	1290	17 inch	12
10095	1296	17 inch	12
10096	1302	17 inch	12
10097	1323	17 inch	12
10098	1496	17 inch	12
10099	2524	17 inch	12
10100	425	HD	1
10101	1291	HD	1
10102	1297	HD	1
10103	1303	HD	1
10104	1308	HD	1
10105	1313	HD	1
10106	1318	HD	1
10107	1324	HD	1
10108	1410	HD	1
10109	1485	HD	1
10110	425	Full HD	2
10111	1291	Full HD	2
10112	1297	Full HD	2
10113	1303	Full HD	2
10114	1308	Full HD	2
10115	1313	Full HD	2
10116	1318	Full HD	2
10117	1324	Full HD	2
10118	1410	Full HD	2
10119	1485	Full HD	2
10120	425	4K	3
10121	1291	4K	3
10122	1297	4K	3
10123	1303	4K	3
10124	1308	4K	3
10125	1313	4K	3
10126	1318	4K	3
10127	1324	4K	3
10128	1410	4K	3
10129	1485	4K	3
10130	425	8K	4
10131	1291	8K	4
10132	1297	8K	4
10133	1303	8K	4
10134	1308	8K	4
10135	1313	8K	4
10136	1318	8K	4
10137	1324	8K	4
10138	1410	8K	4
10139	1485	8K	4
10140	425	2K	5
10141	1291	2K	5
10142	1297	2K	5
10143	1303	2K	5
10144	1308	2K	5
10145	1313	2K	5
10146	1318	2K	5
10147	1324	2K	5
10148	1410	2K	5
10149	1485	2K	5
10150	425	1080p	6
10151	1291	1080p	6
10152	1297	1080p	6
10153	1303	1080p	6
10154	1308	1080p	6
10155	1313	1080p	6
10156	1318	1080p	6
10157	1324	1080p	6
10158	1410	1080p	6
10159	1485	1080p	6
10160	425	720p	7
10161	1291	720p	7
10162	1297	720p	7
10163	1303	720p	7
10164	1308	720p	7
10165	1313	720p	7
10166	1318	720p	7
10167	1324	720p	7
10168	1410	720p	7
10169	1485	720p	7
10170	425	4K UHD	8
10171	1291	4K UHD	8
10172	1297	4K UHD	8
10173	1303	4K UHD	8
10174	1308	4K UHD	8
10175	1313	4K UHD	8
10176	1318	4K UHD	8
10177	1324	4K UHD	8
10178	1410	4K UHD	8
10179	1485	4K UHD	8
10180	1182	Bluetooth	1
10181	1217	Bluetooth	1
10182	1227	Bluetooth	1
10183	1381	Bluetooth	1
10184	1386	Bluetooth	1
10185	1392	Bluetooth	1
10186	1452	Bluetooth	1
10187	1475	Bluetooth	1
10188	1480	Bluetooth	1
10189	1486	Bluetooth	1
10190	1491	Bluetooth	1
10191	1513	Bluetooth	1
10192	1523	Bluetooth	1
10193	1528	Bluetooth	1
10194	1558	Bluetooth	1
10195	1577	Bluetooth	1
10196	1182	WiFi	2
10197	1217	WiFi	2
10198	1227	WiFi	2
10199	1381	WiFi	2
10200	1386	WiFi	2
10201	1392	WiFi	2
10202	1452	WiFi	2
10203	1475	WiFi	2
10204	1480	WiFi	2
10205	1486	WiFi	2
10206	1491	WiFi	2
10207	1513	WiFi	2
10208	1523	WiFi	2
10209	1528	WiFi	2
10210	1558	WiFi	2
10211	1577	WiFi	2
10212	1182	USB	3
10213	1217	USB	3
10214	1227	USB	3
10215	1381	USB	3
10216	1386	USB	3
10217	1392	USB	3
10218	1452	USB	3
10219	1475	USB	3
10220	1480	USB	3
10221	1486	USB	3
10222	1491	USB	3
10223	1513	USB	3
10224	1523	USB	3
10225	1528	USB	3
10226	1558	USB	3
10227	1577	USB	3
10228	1182	HDMI	4
10229	1217	HDMI	4
10230	1227	HDMI	4
10231	1381	HDMI	4
10232	1386	HDMI	4
10233	1392	HDMI	4
10234	1452	HDMI	4
10235	1475	HDMI	4
10236	1480	HDMI	4
10237	1486	HDMI	4
10238	1491	HDMI	4
10239	1513	HDMI	4
10240	1523	HDMI	4
10241	1528	HDMI	4
10242	1558	HDMI	4
10243	1577	HDMI	4
10244	1182	Wired	5
10245	1217	Wired	5
10246	1227	Wired	5
10247	1381	Wired	5
10248	1386	Wired	5
10249	1392	Wired	5
10250	1452	Wired	5
10251	1475	Wired	5
10252	1480	Wired	5
10253	1486	Wired	5
10254	1491	Wired	5
10255	1513	Wired	5
10256	1523	Wired	5
10257	1528	Wired	5
10258	1558	Wired	5
10259	1577	Wired	5
10260	1182	Wireless	6
10261	1217	Wireless	6
10262	1227	Wireless	6
10263	1381	Wireless	6
10264	1386	Wireless	6
10265	1392	Wireless	6
10266	1452	Wireless	6
10267	1475	Wireless	6
10268	1480	Wireless	6
10269	1486	Wireless	6
10270	1491	Wireless	6
10271	1513	Wireless	6
10272	1523	Wireless	6
10273	1528	Wireless	6
10274	1558	Wireless	6
10275	1577	Wireless	6
10276	1182	3.5mm Jack	7
10277	1217	3.5mm Jack	7
10278	1227	3.5mm Jack	7
10279	1381	3.5mm Jack	7
10280	1386	3.5mm Jack	7
10281	1392	3.5mm Jack	7
10282	1452	3.5mm Jack	7
10283	1475	3.5mm Jack	7
10284	1480	3.5mm Jack	7
10285	1486	3.5mm Jack	7
10286	1491	3.5mm Jack	7
10287	1513	3.5mm Jack	7
10288	1523	3.5mm Jack	7
10289	1528	3.5mm Jack	7
10290	1558	3.5mm Jack	7
10291	1577	3.5mm Jack	7
10292	1182	Type-C	8
10293	1217	Type-C	8
10294	1227	Type-C	8
10295	1381	Type-C	8
10296	1386	Type-C	8
10297	1392	Type-C	8
10298	1452	Type-C	8
10299	1475	Type-C	8
10300	1480	Type-C	8
10301	1486	Type-C	8
10302	1491	Type-C	8
10303	1513	Type-C	8
10304	1523	Type-C	8
10305	1528	Type-C	8
10306	1558	Type-C	8
10307	1577	Type-C	8
10308	1182	NFC	9
10309	1217	NFC	9
10310	1227	NFC	9
10311	1381	NFC	9
10312	1386	NFC	9
10313	1392	NFC	9
10314	1452	NFC	9
10315	1475	NFC	9
10316	1480	NFC	9
10317	1486	NFC	9
10318	1491	NFC	9
10319	1513	NFC	9
10320	1523	NFC	9
10321	1528	NFC	9
10322	1558	NFC	9
10323	1577	NFC	9
10324	1229	Intel Core i3	1
10325	1236	Intel Core i3	1
10326	1242	Intel Core i3	1
10327	1249	Intel Core i3	1
10328	1255	Intel Core i3	1
10329	1261	Intel Core i3	1
10330	1267	Intel Core i3	1
10331	1273	Intel Core i3	1
10332	1278	Intel Core i3	1
10333	1284	Intel Core i3	1
10334	1229	Intel Core i5	2
10335	1236	Intel Core i5	2
10336	1242	Intel Core i5	2
10337	1249	Intel Core i5	2
10338	1255	Intel Core i5	2
10339	1261	Intel Core i5	2
10340	1267	Intel Core i5	2
10341	1273	Intel Core i5	2
10342	1278	Intel Core i5	2
10343	1284	Intel Core i5	2
10344	1229	Intel Core i7	3
10345	1236	Intel Core i7	3
10346	1242	Intel Core i7	3
10347	1249	Intel Core i7	3
10348	1255	Intel Core i7	3
10349	1261	Intel Core i7	3
10350	1267	Intel Core i7	3
10351	1273	Intel Core i7	3
10352	1278	Intel Core i7	3
10353	1284	Intel Core i7	3
10354	1229	Intel Core i9	4
10355	1236	Intel Core i9	4
10356	1242	Intel Core i9	4
10357	1249	Intel Core i9	4
10358	1255	Intel Core i9	4
10359	1261	Intel Core i9	4
10360	1267	Intel Core i9	4
10361	1273	Intel Core i9	4
10362	1278	Intel Core i9	4
10363	1284	Intel Core i9	4
10364	1229	AMD Ryzen 3	5
10365	1236	AMD Ryzen 3	5
10366	1242	AMD Ryzen 3	5
10367	1249	AMD Ryzen 3	5
10368	1255	AMD Ryzen 3	5
10369	1261	AMD Ryzen 3	5
10370	1267	AMD Ryzen 3	5
10371	1273	AMD Ryzen 3	5
10372	1278	AMD Ryzen 3	5
10373	1284	AMD Ryzen 3	5
10374	1229	AMD Ryzen 5	6
10375	1236	AMD Ryzen 5	6
10376	1242	AMD Ryzen 5	6
10377	1249	AMD Ryzen 5	6
10378	1255	AMD Ryzen 5	6
10379	1261	AMD Ryzen 5	6
10380	1267	AMD Ryzen 5	6
10381	1273	AMD Ryzen 5	6
10382	1278	AMD Ryzen 5	6
10383	1284	AMD Ryzen 5	6
10384	1229	AMD Ryzen 7	7
10385	1236	AMD Ryzen 7	7
10386	1242	AMD Ryzen 7	7
10387	1249	AMD Ryzen 7	7
10388	1255	AMD Ryzen 7	7
10389	1261	AMD Ryzen 7	7
10390	1267	AMD Ryzen 7	7
10391	1273	AMD Ryzen 7	7
10392	1278	AMD Ryzen 7	7
10393	1284	AMD Ryzen 7	7
10394	1229	AMD Ryzen 9	8
10395	1236	AMD Ryzen 9	8
10396	1242	AMD Ryzen 9	8
10397	1249	AMD Ryzen 9	8
10398	1255	AMD Ryzen 9	8
10399	1261	AMD Ryzen 9	8
10400	1267	AMD Ryzen 9	8
10401	1273	AMD Ryzen 9	8
10402	1278	AMD Ryzen 9	8
10403	1284	AMD Ryzen 9	8
10404	1229	Apple M1	9
10405	1236	Apple M1	9
10406	1242	Apple M1	9
10407	1249	Apple M1	9
10408	1255	Apple M1	9
10409	1261	Apple M1	9
10410	1267	Apple M1	9
10411	1273	Apple M1	9
10412	1278	Apple M1	9
10413	1284	Apple M1	9
10414	1229	Apple M2	10
10415	1236	Apple M2	10
10416	1242	Apple M2	10
10417	1249	Apple M2	10
10418	1255	Apple M2	10
10419	1261	Apple M2	10
10420	1267	Apple M2	10
10421	1273	Apple M2	10
10422	1278	Apple M2	10
10423	1284	Apple M2	10
10424	1229	Snapdragon	11
10425	1236	Snapdragon	11
10426	1242	Snapdragon	11
10427	1249	Snapdragon	11
10428	1255	Snapdragon	11
10429	1261	Snapdragon	11
10430	1267	Snapdragon	11
10431	1273	Snapdragon	11
10432	1278	Snapdragon	11
10433	1284	Snapdragon	11
10434	1229	MediaTek	12
10435	1236	MediaTek	12
10436	1242	MediaTek	12
10437	1249	MediaTek	12
10438	1255	MediaTek	12
10439	1261	MediaTek	12
10440	1267	MediaTek	12
10441	1273	MediaTek	12
10442	1278	MediaTek	12
10443	1284	MediaTek	12
10444	1234	Windows 11	1
10445	1240	Windows 11	1
10446	1247	Windows 11	1
10447	1253	Windows 11	1
10448	1259	Windows 11	1
10449	1265	Windows 11	1
10450	1271	Windows 11	1
10451	1276	Windows 11	1
10452	1282	Windows 11	1
10453	1288	Windows 11	1
10454	1234	Windows 10	2
10455	1240	Windows 10	2
10456	1247	Windows 10	2
10457	1253	Windows 10	2
10458	1259	Windows 10	2
10459	1265	Windows 10	2
10460	1271	Windows 10	2
10461	1276	Windows 10	2
10462	1282	Windows 10	2
10463	1288	Windows 10	2
10464	1234	macOS	3
10465	1240	macOS	3
10466	1247	macOS	3
10467	1253	macOS	3
10468	1259	macOS	3
10469	1265	macOS	3
10470	1271	macOS	3
10471	1276	macOS	3
10472	1282	macOS	3
10473	1288	macOS	3
10474	1234	Linux	4
10475	1240	Linux	4
10476	1247	Linux	4
10477	1253	Linux	4
10478	1259	Linux	4
10479	1265	Linux	4
10480	1271	Linux	4
10481	1276	Linux	4
10482	1282	Linux	4
10483	1288	Linux	4
10484	1234	Chrome OS	5
10485	1240	Chrome OS	5
10486	1247	Chrome OS	5
10487	1253	Chrome OS	5
10488	1259	Chrome OS	5
10489	1265	Chrome OS	5
10490	1271	Chrome OS	5
10491	1276	Chrome OS	5
10492	1282	Chrome OS	5
10493	1288	Chrome OS	5
10494	1234	DOS	6
10495	1240	DOS	6
10496	1247	DOS	6
10497	1253	DOS	6
10498	1259	DOS	6
10499	1265	DOS	6
10500	1271	DOS	6
10501	1276	DOS	6
10502	1282	DOS	6
10503	1288	DOS	6
10504	1234	Ubuntu	7
10505	1240	Ubuntu	7
10506	1247	Ubuntu	7
10507	1253	Ubuntu	7
10508	1259	Ubuntu	7
10509	1265	Ubuntu	7
10510	1271	Ubuntu	7
10511	1276	Ubuntu	7
10512	1282	Ubuntu	7
10513	1288	Ubuntu	7
10514	1232	Integrated	1
10515	1232	NVIDIA GTX 1650	2
10516	1232	NVIDIA RTX 3050	3
10517	1232	NVIDIA RTX 3060	4
10518	1232	NVIDIA RTX 3070	5
10519	1232	NVIDIA RTX 4060	6
10520	1232	AMD Radeon	7
10521	1232	Intel Iris	8
10522	1270	Integrated	1
10523	1270	NVIDIA GTX 1650	2
10524	1270	NVIDIA RTX 3050	3
10525	1270	NVIDIA RTX 3060	4
10526	1270	NVIDIA RTX 3070	5
10527	1270	NVIDIA RTX 4060	6
10528	1270	AMD Radeon	7
10529	1270	Intel Iris	8
10530	1287	Integrated	1
10531	1287	NVIDIA GTX 1650	2
10532	1287	NVIDIA RTX 3050	3
10533	1287	NVIDIA RTX 3060	4
10534	1287	NVIDIA RTX 3070	5
10535	1287	NVIDIA RTX 4060	6
10536	1287	AMD Radeon	7
10537	1287	Intel Iris	8
10538	364	5L	1
10539	384	5L	1
10540	500	5L	1
10541	1339	5L	1
10542	1345	5L	1
10543	1355	5L	1
10544	1374	5L	1
10545	1530	5L	1
10546	1535	5L	1
10547	1540	5L	1
10548	1545	5L	1
10549	1580	5L	1
10550	1585	5L	1
10551	1590	5L	1
10552	1918	5L	1
10553	364	10L	2
10554	384	10L	2
10555	500	10L	2
10556	1339	10L	2
10557	1345	10L	2
10558	1355	10L	2
10559	1374	10L	2
10560	1530	10L	2
10561	1535	10L	2
10562	1540	10L	2
10563	1545	10L	2
10564	1580	10L	2
10565	1585	10L	2
10566	1590	10L	2
10567	1918	10L	2
10568	364	15L	3
10569	384	15L	3
10570	500	15L	3
10571	1339	15L	3
10572	1345	15L	3
10573	1355	15L	3
10574	1374	15L	3
10575	1530	15L	3
10576	1535	15L	3
10577	1540	15L	3
10578	1545	15L	3
10579	1580	15L	3
10580	1585	15L	3
10581	1590	15L	3
10582	1918	15L	3
10583	364	20L	4
10584	384	20L	4
10585	500	20L	4
10586	1339	20L	4
10587	1345	20L	4
10588	1355	20L	4
10589	1374	20L	4
10590	1530	20L	4
10591	1535	20L	4
10592	1540	20L	4
10593	1545	20L	4
10594	1580	20L	4
10595	1585	20L	4
10596	1590	20L	4
10597	1918	20L	4
10598	364	25L	5
10599	384	25L	5
10600	500	25L	5
10601	1339	25L	5
10602	1345	25L	5
10603	1355	25L	5
10604	1374	25L	5
10605	1530	25L	5
10606	1535	25L	5
10607	1540	25L	5
10608	1545	25L	5
10609	1580	25L	5
10610	1585	25L	5
10611	1590	25L	5
10612	1918	25L	5
10613	364	30L	6
10614	384	30L	6
10615	500	30L	6
10616	1339	30L	6
10617	1345	30L	6
10618	1355	30L	6
10619	1374	30L	6
10620	1530	30L	6
10621	1535	30L	6
10622	1540	30L	6
10623	1545	30L	6
10624	1580	30L	6
10625	1585	30L	6
10626	1590	30L	6
10627	1918	30L	6
10628	364	100L	7
10629	384	100L	7
10630	500	100L	7
10631	1339	100L	7
10632	1345	100L	7
10633	1355	100L	7
10634	1374	100L	7
10635	1530	100L	7
10636	1535	100L	7
10637	1540	100L	7
10638	1545	100L	7
10639	1580	100L	7
10640	1585	100L	7
10641	1590	100L	7
10642	1918	100L	7
10643	364	200L	8
10644	384	200L	8
10645	500	200L	8
10646	1339	200L	8
10647	1345	200L	8
10648	1355	200L	8
10649	1374	200L	8
10650	1530	200L	8
10651	1535	200L	8
10652	1540	200L	8
10653	1545	200L	8
10654	1580	200L	8
10655	1585	200L	8
10656	1590	200L	8
10657	1918	200L	8
10658	364	250L	9
10659	384	250L	9
10660	500	250L	9
10661	1339	250L	9
10662	1345	250L	9
10663	1355	250L	9
10664	1374	250L	9
10665	1530	250L	9
10666	1535	250L	9
10667	1540	250L	9
10668	1545	250L	9
10669	1580	250L	9
10670	1585	250L	9
10671	1590	250L	9
10672	1918	250L	9
10673	364	300L	10
10674	384	300L	10
10675	500	300L	10
10676	1339	300L	10
10677	1345	300L	10
10678	1355	300L	10
10679	1374	300L	10
10680	1530	300L	10
10681	1535	300L	10
10682	1540	300L	10
10683	1545	300L	10
10684	1580	300L	10
10685	1585	300L	10
10686	1590	300L	10
10687	1918	300L	10
10688	364	500L	11
10689	384	500L	11
10690	500	500L	11
10691	1339	500L	11
10692	1345	500L	11
10693	1355	500L	11
10694	1374	500L	11
10695	1530	500L	11
10696	1535	500L	11
10697	1540	500L	11
10698	1545	500L	11
10699	1580	500L	11
10700	1585	500L	11
10701	1590	500L	11
10702	1918	500L	11
10703	364	Small	12
10704	384	Small	12
10705	500	Small	12
10706	1339	Small	12
10707	1345	Small	12
10708	1355	Small	12
10709	1374	Small	12
10710	1530	Small	12
10711	1535	Small	12
10712	1540	Small	12
10713	1545	Small	12
10714	1580	Small	12
10715	1585	Small	12
10716	1590	Small	12
10717	1918	Small	12
10718	364	Medium	13
10719	384	Medium	13
10720	500	Medium	13
10721	1339	Medium	13
10722	1345	Medium	13
10723	1355	Medium	13
10724	1374	Medium	13
10725	1530	Medium	13
10726	1535	Medium	13
10727	1540	Medium	13
10728	1545	Medium	13
10729	1580	Medium	13
10730	1585	Medium	13
10731	1590	Medium	13
10732	1918	Medium	13
10733	364	Large	14
10734	384	Large	14
10735	500	Large	14
10736	1339	Large	14
10737	1345	Large	14
10738	1355	Large	14
10739	1374	Large	14
10740	1530	Large	14
10741	1535	Large	14
10742	1540	Large	14
10743	1545	Large	14
10744	1580	Large	14
10745	1585	Large	14
10746	1590	Large	14
10747	1918	Large	14
10748	1330	1 Star	1
10749	1330	2 Star	2
10750	1330	3 Star	3
10751	1330	4 Star	4
10752	1330	5 Star	5
10753	1335	1 Star	1
10754	1335	2 Star	2
10755	1335	3 Star	3
10756	1335	4 Star	4
10757	1335	5 Star	5
10758	1341	1 Star	1
10759	1341	2 Star	2
10760	1341	3 Star	3
10761	1341	4 Star	4
10762	1341	5 Star	5
10763	1346	1 Star	1
10764	1346	2 Star	2
10765	1346	3 Star	3
10766	1346	4 Star	4
10767	1346	5 Star	5
10768	629	Black	1
10769	629	White	2
10770	629	Red	3
10771	629	Blue	4
10772	629	Green	5
10773	629	Yellow	6
10774	629	Pink	7
10775	629	Purple	8
10776	629	Orange	9
10777	629	Brown	10
10778	629	Grey	11
10779	629	Navy	12
10780	629	Maroon	13
10781	629	Beige	14
10782	629	Multicolor	15
10783	1965	Black	1
10784	1965	White	2
10785	1965	Red	3
10786	1965	Blue	4
10787	1965	Green	5
10788	1965	Yellow	6
10789	1965	Pink	7
10790	1965	Purple	8
10791	1965	Orange	9
10792	1965	Brown	10
10793	1965	Grey	11
10794	1965	Navy	12
10795	1965	Maroon	13
10796	1965	Beige	14
10797	1965	Multicolor	15
10798	1988	Black	1
10799	1988	White	2
10800	1988	Red	3
10801	1988	Blue	4
10802	1988	Green	5
10803	1988	Yellow	6
10804	1988	Pink	7
10805	1988	Purple	8
10806	1988	Orange	9
10807	1988	Brown	10
10808	1988	Grey	11
10809	1988	Navy	12
10810	1988	Maroon	13
10811	1988	Beige	14
10812	1988	Multicolor	15
10813	2047	Black	1
10814	2047	White	2
10815	2047	Red	3
10816	2047	Blue	4
10817	2047	Green	5
10818	2047	Yellow	6
10819	2047	Pink	7
10820	2047	Purple	8
10821	2047	Orange	9
10822	2047	Brown	10
10823	2047	Grey	11
10824	2047	Navy	12
10825	2047	Maroon	13
10826	2047	Beige	14
10827	2047	Multicolor	15
10828	2079	Solid	1
10829	2132	Solid	1
10830	2154	Solid	1
10831	2168	Solid	1
10832	2180	Solid	1
10833	2222	Solid	1
10834	2254	Solid	1
10835	2272	Solid	1
10836	2311	Solid	1
10837	2379	Solid	1
10838	2385	Solid	1
10839	2441	Solid	1
10840	2079	Striped	2
10841	2132	Striped	2
10842	2154	Striped	2
10843	2168	Striped	2
10844	2180	Striped	2
10845	2222	Striped	2
10846	2254	Striped	2
10847	2272	Striped	2
10848	2311	Striped	2
10849	2379	Striped	2
10850	2385	Striped	2
10851	2441	Striped	2
10852	2079	Checkered	3
10853	2132	Checkered	3
10854	2154	Checkered	3
10855	2168	Checkered	3
10856	2180	Checkered	3
10857	2222	Checkered	3
10858	2254	Checkered	3
10859	2272	Checkered	3
10860	2311	Checkered	3
10861	2379	Checkered	3
10862	2385	Checkered	3
10863	2441	Checkered	3
10864	2079	Printed	4
10865	2132	Printed	4
10866	2154	Printed	4
10867	2168	Printed	4
10868	2180	Printed	4
10869	2222	Printed	4
10870	2254	Printed	4
10871	2272	Printed	4
10872	2311	Printed	4
10873	2379	Printed	4
10874	2385	Printed	4
10875	2441	Printed	4
10876	2079	Floral	5
10877	2132	Floral	5
10878	2154	Floral	5
10879	2168	Floral	5
10880	2180	Floral	5
10881	2222	Floral	5
10882	2254	Floral	5
10883	2272	Floral	5
10884	2311	Floral	5
10885	2379	Floral	5
10886	2385	Floral	5
10887	2441	Floral	5
10888	2079	Polka Dots	6
10889	2132	Polka Dots	6
10890	2154	Polka Dots	6
10891	2168	Polka Dots	6
10892	2180	Polka Dots	6
10893	2222	Polka Dots	6
10894	2254	Polka Dots	6
10895	2272	Polka Dots	6
10896	2311	Polka Dots	6
10897	2379	Polka Dots	6
10898	2385	Polka Dots	6
10899	2441	Polka Dots	6
10900	2079	Abstract	7
10901	2132	Abstract	7
10902	2154	Abstract	7
10903	2168	Abstract	7
10904	2180	Abstract	7
10905	2222	Abstract	7
10906	2254	Abstract	7
10907	2272	Abstract	7
10908	2311	Abstract	7
10909	2379	Abstract	7
10910	2385	Abstract	7
10911	2441	Abstract	7
10912	2079	Geometric	8
10913	2132	Geometric	8
10914	2154	Geometric	8
10915	2168	Geometric	8
10916	2180	Geometric	8
10917	2222	Geometric	8
10918	2254	Geometric	8
10919	2272	Geometric	8
10920	2311	Geometric	8
10921	2379	Geometric	8
10922	2385	Geometric	8
10923	2441	Geometric	8
10924	2079	Plain	9
10925	2132	Plain	9
10926	2154	Plain	9
10927	2168	Plain	9
10928	2180	Plain	9
10929	2222	Plain	9
10930	2254	Plain	9
10931	2272	Plain	9
10932	2311	Plain	9
10933	2379	Plain	9
10934	2385	Plain	9
10935	2441	Plain	9
10936	2073	Slim Fit	1
10937	2097	Slim Fit	1
10938	2103	Slim Fit	1
10939	2109	Slim Fit	1
10940	2115	Slim Fit	1
10941	2133	Slim Fit	1
10942	2139	Slim Fit	1
10943	2199	Slim Fit	1
10944	2229	Slim Fit	1
10945	2266	Slim Fit	1
10946	2310	Slim Fit	1
10947	2361	Slim Fit	1
10948	2367	Slim Fit	1
10949	2373	Slim Fit	1
10950	2378	Slim Fit	1
10951	2397	Slim Fit	1
10952	2073	Regular Fit	2
10953	2097	Regular Fit	2
10954	2103	Regular Fit	2
10955	2109	Regular Fit	2
10956	2115	Regular Fit	2
10957	2133	Regular Fit	2
10958	2139	Regular Fit	2
10959	2199	Regular Fit	2
10960	2229	Regular Fit	2
10961	2266	Regular Fit	2
10962	2310	Regular Fit	2
10963	2361	Regular Fit	2
10964	2367	Regular Fit	2
10965	2373	Regular Fit	2
10966	2378	Regular Fit	2
10967	2397	Regular Fit	2
10968	2073	Relaxed Fit	3
10969	2097	Relaxed Fit	3
10970	2103	Relaxed Fit	3
10971	2109	Relaxed Fit	3
10972	2115	Relaxed Fit	3
10973	2133	Relaxed Fit	3
10974	2139	Relaxed Fit	3
10975	2199	Relaxed Fit	3
10976	2229	Relaxed Fit	3
10977	2266	Relaxed Fit	3
10978	2310	Relaxed Fit	3
10979	2361	Relaxed Fit	3
10980	2367	Relaxed Fit	3
10981	2373	Relaxed Fit	3
10982	2378	Relaxed Fit	3
10983	2397	Relaxed Fit	3
10984	2073	Loose Fit	4
10985	2097	Loose Fit	4
10986	2103	Loose Fit	4
10987	2109	Loose Fit	4
10988	2115	Loose Fit	4
10989	2133	Loose Fit	4
10990	2139	Loose Fit	4
10991	2199	Loose Fit	4
10992	2229	Loose Fit	4
10993	2266	Loose Fit	4
10994	2310	Loose Fit	4
10995	2361	Loose Fit	4
10996	2367	Loose Fit	4
10997	2373	Loose Fit	4
10998	2378	Loose Fit	4
10999	2397	Loose Fit	4
11000	2073	Athletic Fit	5
11001	2097	Athletic Fit	5
11002	2103	Athletic Fit	5
11003	2109	Athletic Fit	5
11004	2115	Athletic Fit	5
11005	2133	Athletic Fit	5
11006	2139	Athletic Fit	5
11007	2199	Athletic Fit	5
11008	2229	Athletic Fit	5
11009	2266	Athletic Fit	5
11010	2310	Athletic Fit	5
11011	2361	Athletic Fit	5
11012	2367	Athletic Fit	5
11013	2373	Athletic Fit	5
11014	2378	Athletic Fit	5
11015	2397	Athletic Fit	5
11016	2073	Tailored Fit	6
11017	2097	Tailored Fit	6
11018	2103	Tailored Fit	6
11019	2109	Tailored Fit	6
11020	2115	Tailored Fit	6
11021	2133	Tailored Fit	6
11022	2139	Tailored Fit	6
11023	2199	Tailored Fit	6
11024	2229	Tailored Fit	6
11025	2266	Tailored Fit	6
11026	2310	Tailored Fit	6
11027	2361	Tailored Fit	6
11028	2367	Tailored Fit	6
11029	2373	Tailored Fit	6
11030	2378	Tailored Fit	6
11031	2397	Tailored Fit	6
11032	2073	Comfort Fit	7
11033	2097	Comfort Fit	7
11034	2103	Comfort Fit	7
11035	2109	Comfort Fit	7
11036	2115	Comfort Fit	7
11037	2133	Comfort Fit	7
11038	2139	Comfort Fit	7
11039	2199	Comfort Fit	7
11040	2229	Comfort Fit	7
11041	2266	Comfort Fit	7
11042	2310	Comfort Fit	7
11043	2361	Comfort Fit	7
11044	2367	Comfort Fit	7
11045	2373	Comfort Fit	7
11046	2378	Comfort Fit	7
11047	2397	Comfort Fit	7
11048	2072	Full Sleeve	1
11049	2078	Full Sleeve	1
11050	2091	Full Sleeve	1
11051	2175	Full Sleeve	1
11052	2204	Full Sleeve	1
11053	2249	Full Sleeve	1
11054	2271	Full Sleeve	1
11055	2366	Full Sleeve	1
11056	2384	Full Sleeve	1
11057	2430	Full Sleeve	1
11058	2545	Full Sleeve	1
11059	2072	Half Sleeve	2
11060	2078	Half Sleeve	2
11061	2091	Half Sleeve	2
11062	2175	Half Sleeve	2
11063	2204	Half Sleeve	2
11064	2249	Half Sleeve	2
11065	2271	Half Sleeve	2
11066	2366	Half Sleeve	2
11067	2384	Half Sleeve	2
11068	2430	Half Sleeve	2
11069	2545	Half Sleeve	2
11070	2072	Three Quarter Sleeve	3
11071	2078	Three Quarter Sleeve	3
11072	2091	Three Quarter Sleeve	3
11073	2175	Three Quarter Sleeve	3
11074	2204	Three Quarter Sleeve	3
11075	2249	Three Quarter Sleeve	3
11076	2271	Three Quarter Sleeve	3
11077	2366	Three Quarter Sleeve	3
11078	2384	Three Quarter Sleeve	3
11079	2430	Three Quarter Sleeve	3
11080	2545	Three Quarter Sleeve	3
11081	2072	Sleeveless	4
11082	2078	Sleeveless	4
11083	2091	Sleeveless	4
11084	2175	Sleeveless	4
11085	2204	Sleeveless	4
11086	2249	Sleeveless	4
11087	2271	Sleeveless	4
11088	2366	Sleeveless	4
11089	2384	Sleeveless	4
11090	2430	Sleeveless	4
11091	2545	Sleeveless	4
11092	2072	Short Sleeve	5
11093	2078	Short Sleeve	5
11094	2091	Short Sleeve	5
11095	2175	Short Sleeve	5
11096	2204	Short Sleeve	5
11097	2249	Short Sleeve	5
11098	2271	Short Sleeve	5
11099	2366	Short Sleeve	5
11100	2384	Short Sleeve	5
11101	2430	Short Sleeve	5
11102	2545	Short Sleeve	5
11103	2072	Long Sleeve	6
11104	2078	Long Sleeve	6
11105	2091	Long Sleeve	6
11106	2175	Long Sleeve	6
11107	2204	Long Sleeve	6
11108	2249	Long Sleeve	6
11109	2271	Long Sleeve	6
11110	2366	Long Sleeve	6
11111	2384	Long Sleeve	6
11112	2430	Long Sleeve	6
11113	2545	Long Sleeve	6
11114	2084	Round Neck	1
11115	2169	Round Neck	1
11116	2186	Round Neck	1
11117	2192	Round Neck	1
11118	2205	Round Neck	1
11119	2259	Round Neck	1
11120	2278	Round Neck	1
11121	2390	Round Neck	1
11122	2429	Round Neck	1
11123	2084	V Neck	2
11124	2169	V Neck	2
11125	2186	V Neck	2
11126	2192	V Neck	2
11127	2205	V Neck	2
11128	2259	V Neck	2
11129	2278	V Neck	2
11130	2390	V Neck	2
11131	2429	V Neck	2
11132	2084	Collar	3
11133	2169	Collar	3
11134	2186	Collar	3
11135	2192	Collar	3
11136	2205	Collar	3
11137	2259	Collar	3
11138	2278	Collar	3
11139	2390	Collar	3
11140	2429	Collar	3
11141	2084	Boat Neck	4
11142	2169	Boat Neck	4
11143	2186	Boat Neck	4
11144	2192	Boat Neck	4
11145	2205	Boat Neck	4
11146	2259	Boat Neck	4
11147	2278	Boat Neck	4
11148	2390	Boat Neck	4
11149	2429	Boat Neck	4
11150	2084	High Neck	5
11151	2169	High Neck	5
11152	2186	High Neck	5
11153	2192	High Neck	5
11154	2205	High Neck	5
11155	2259	High Neck	5
11156	2278	High Neck	5
11157	2390	High Neck	5
11158	2429	High Neck	5
11159	2084	Scoop Neck	6
11160	2169	Scoop Neck	6
11161	2186	Scoop Neck	6
11162	2192	Scoop Neck	6
11163	2205	Scoop Neck	6
11164	2259	Scoop Neck	6
11165	2278	Scoop Neck	6
11166	2390	Scoop Neck	6
11167	2429	Scoop Neck	6
11168	2084	Square Neck	7
11169	2169	Square Neck	7
11170	2186	Square Neck	7
11171	2192	Square Neck	7
11172	2205	Square Neck	7
11173	2259	Square Neck	7
11174	2278	Square Neck	7
11175	2390	Square Neck	7
11176	2429	Square Neck	7
11177	2084	Turtle Neck	8
11178	2169	Turtle Neck	8
11179	2186	Turtle Neck	8
11180	2192	Turtle Neck	8
11181	2205	Turtle Neck	8
11182	2259	Turtle Neck	8
11183	2278	Turtle Neck	8
11184	2390	Turtle Neck	8
11185	2429	Turtle Neck	8
11186	2100	26	1
11187	2106	26	1
11188	2112	26	1
11189	2118	26	1
11190	2124	26	1
11191	2263	26	1
11192	2281	26	1
11193	2292	26	1
11194	2335	26	1
11195	2370	26	1
11196	2394	26	1
11197	2400	26	1
11198	2548	26	1
11199	2100	28	2
11200	2106	28	2
11201	2112	28	2
11202	2118	28	2
11203	2124	28	2
11204	2263	28	2
11205	2281	28	2
11206	2292	28	2
11207	2335	28	2
11208	2370	28	2
11209	2394	28	2
11210	2400	28	2
11211	2548	28	2
11212	2100	30	3
11213	2106	30	3
11214	2112	30	3
11215	2118	30	3
11216	2124	30	3
11217	2263	30	3
11218	2281	30	3
11219	2292	30	3
11220	2335	30	3
11221	2370	30	3
11222	2394	30	3
11223	2400	30	3
11224	2548	30	3
11225	2100	32	4
11226	2106	32	4
11227	2112	32	4
11228	2118	32	4
11229	2124	32	4
11230	2263	32	4
11231	2281	32	4
11232	2292	32	4
11233	2335	32	4
11234	2370	32	4
11235	2394	32	4
11236	2400	32	4
11237	2548	32	4
11238	2100	34	5
11239	2106	34	5
11240	2112	34	5
11241	2118	34	5
11242	2124	34	5
11243	2263	34	5
11244	2281	34	5
11245	2292	34	5
11246	2335	34	5
11247	2370	34	5
11248	2394	34	5
11249	2400	34	5
11250	2548	34	5
11251	2100	36	6
11252	2106	36	6
11253	2112	36	6
11254	2118	36	6
11255	2124	36	6
11256	2263	36	6
11257	2281	36	6
11258	2292	36	6
11259	2335	36	6
11260	2370	36	6
11261	2394	36	6
11262	2400	36	6
11263	2548	36	6
11264	2100	38	7
11265	2106	38	7
11266	2112	38	7
11267	2118	38	7
11268	2124	38	7
11269	2263	38	7
11270	2281	38	7
11271	2292	38	7
11272	2335	38	7
11273	2370	38	7
11274	2394	38	7
11275	2400	38	7
11276	2548	38	7
11277	2100	40	8
11278	2106	40	8
11279	2112	40	8
11280	2118	40	8
11281	2124	40	8
11282	2263	40	8
11283	2281	40	8
11284	2292	40	8
11285	2335	40	8
11286	2370	40	8
11287	2394	40	8
11288	2400	40	8
11289	2548	40	8
11290	2100	42	9
11291	2106	42	9
11292	2112	42	9
11293	2118	42	9
11294	2124	42	9
11295	2263	42	9
11296	2281	42	9
11297	2292	42	9
11298	2335	42	9
11299	2370	42	9
11300	2394	42	9
11301	2400	42	9
11302	2548	42	9
11303	2100	44	10
11304	2106	44	10
11305	2112	44	10
11306	2118	44	10
11307	2124	44	10
11308	2263	44	10
11309	2281	44	10
11310	2292	44	10
11311	2335	44	10
11312	2370	44	10
11313	2394	44	10
11314	2400	44	10
11315	2548	44	10
11316	2100	46	11
11317	2106	46	11
11318	2112	46	11
11319	2118	46	11
11320	2124	46	11
11321	2263	46	11
11322	2281	46	11
11323	2292	46	11
11324	2335	46	11
11325	2370	46	11
11326	2394	46	11
11327	2400	46	11
11328	2548	46	11
11329	839	1 Seater	1
11330	844	1 Seater	1
11331	909	1 Seater	1
11332	912	1 Seater	1
11333	948	1 Seater	1
11334	958	1 Seater	1
11335	978	1 Seater	1
11336	1019	1 Seater	1
11337	1024	1 Seater	1
11338	1030	1 Seater	1
11339	1035	1 Seater	1
11340	1096	1 Seater	1
11341	839	2 Seater	2
11342	844	2 Seater	2
11343	909	2 Seater	2
11344	912	2 Seater	2
11345	948	2 Seater	2
11346	958	2 Seater	2
11347	978	2 Seater	2
11348	1019	2 Seater	2
11349	1024	2 Seater	2
11350	1030	2 Seater	2
11351	1035	2 Seater	2
11352	1096	2 Seater	2
11353	839	3 Seater	3
11354	844	3 Seater	3
11355	909	3 Seater	3
11356	912	3 Seater	3
11357	948	3 Seater	3
11358	958	3 Seater	3
11359	978	3 Seater	3
11360	1019	3 Seater	3
11361	1024	3 Seater	3
11362	1030	3 Seater	3
11363	1035	3 Seater	3
11364	1096	3 Seater	3
11365	839	4 Seater	4
11366	844	4 Seater	4
11367	909	4 Seater	4
11368	912	4 Seater	4
11369	948	4 Seater	4
11370	958	4 Seater	4
11371	978	4 Seater	4
11372	1019	4 Seater	4
11373	1024	4 Seater	4
11374	1030	4 Seater	4
11375	1035	4 Seater	4
11376	1096	4 Seater	4
11377	839	5 Seater	5
11378	844	5 Seater	5
11379	909	5 Seater	5
11380	912	5 Seater	5
11381	948	5 Seater	5
11382	958	5 Seater	5
11383	978	5 Seater	5
11384	1019	5 Seater	5
11385	1024	5 Seater	5
11386	1030	5 Seater	5
11387	1035	5 Seater	5
11388	1096	5 Seater	5
11389	839	6 Seater	6
11390	844	6 Seater	6
11391	909	6 Seater	6
11392	912	6 Seater	6
11393	948	6 Seater	6
11394	958	6 Seater	6
11395	978	6 Seater	6
11396	1019	6 Seater	6
11397	1024	6 Seater	6
11398	1030	6 Seater	6
11399	1035	6 Seater	6
11400	1096	6 Seater	6
11401	839	8 Seater	7
11402	844	8 Seater	7
11403	909	8 Seater	7
11404	912	8 Seater	7
11405	948	8 Seater	7
11406	958	8 Seater	7
11407	978	8 Seater	7
11408	1019	8 Seater	7
11409	1024	8 Seater	7
11410	1030	8 Seater	7
11411	1035	8 Seater	7
11412	1096	8 Seater	7
11413	839	10+ Seater	8
11414	844	10+ Seater	8
11415	909	10+ Seater	8
11416	912	10+ Seater	8
11417	948	10+ Seater	8
11418	958	10+ Seater	8
11419	978	10+ Seater	8
11420	1019	10+ Seater	8
11421	1024	10+ Seater	8
11422	1030	10+ Seater	8
11423	1035	10+ Seater	8
11424	1096	10+ Seater	8
11425	855	Round	1
11426	859	Round	1
11427	979	Round	1
11428	1087	Round	1
11429	1093	Round	1
11430	1098	Round	1
11431	855	Square	2
11432	859	Square	2
11433	979	Square	2
11434	1087	Square	2
11435	1093	Square	2
11436	1098	Square	2
11437	855	Rectangle	3
11438	859	Rectangle	3
11439	979	Rectangle	3
11440	1087	Rectangle	3
11441	1093	Rectangle	3
11442	1098	Rectangle	3
11443	855	Oval	4
11444	859	Oval	4
11445	979	Oval	4
11446	1087	Oval	4
11447	1093	Oval	4
11448	1098	Oval	4
11449	855	Hexagonal	5
11450	859	Hexagonal	5
11451	979	Hexagonal	5
11452	1087	Hexagonal	5
11453	1093	Hexagonal	5
11454	1098	Hexagonal	5
11455	855	L-Shaped	6
11456	859	L-Shaped	6
11457	979	L-Shaped	6
11458	1087	L-Shaped	6
11459	1093	L-Shaped	6
11460	1098	L-Shaped	6
11461	855	Irregular	7
11462	859	Irregular	7
11463	979	Irregular	7
11464	1087	Irregular	7
11465	1093	Irregular	7
11466	1098	Irregular	7
11467	993	Soft	1
11468	1004	Soft	1
11469	1008	Soft	1
11470	1012	Soft	1
11471	1016	Soft	1
11472	993	Medium	2
11473	1004	Medium	2
11474	1008	Medium	2
11475	1012	Medium	2
11476	1016	Medium	2
11477	993	Firm	3
11478	1004	Firm	3
11479	1008	Firm	3
11480	1012	Firm	3
11481	1016	Firm	3
11482	993	Extra Firm	4
11483	1004	Extra Firm	4
11484	1008	Extra Firm	4
11485	1012	Extra Firm	4
11486	1016	Extra Firm	4
11487	993	Plush	5
11488	1004	Plush	5
11489	1008	Plush	5
11490	1012	Plush	5
11491	1016	Plush	5
11492	869	No Headboard	1
11493	874	No Headboard	1
11494	878	No Headboard	1
11495	883	No Headboard	1
11496	1041	No Headboard	1
11497	1046	No Headboard	1
11498	1050	No Headboard	1
11499	1055	No Headboard	1
11500	869	Wooden Headboard	2
11501	874	Wooden Headboard	2
11502	878	Wooden Headboard	2
11503	883	Wooden Headboard	2
11504	1041	Wooden Headboard	2
11505	1046	Wooden Headboard	2
11506	1050	Wooden Headboard	2
11507	1055	Wooden Headboard	2
11508	869	Upholstered	3
11509	874	Upholstered	3
11510	878	Upholstered	3
11511	883	Upholstered	3
11512	1041	Upholstered	3
11513	1046	Upholstered	3
11514	1050	Upholstered	3
11515	1055	Upholstered	3
11516	869	Metal Headboard	4
11517	874	Metal Headboard	4
11518	878	Metal Headboard	4
11519	883	Metal Headboard	4
11520	1041	Metal Headboard	4
11521	1046	Metal Headboard	4
11522	1050	Metal Headboard	4
11523	1055	Metal Headboard	4
11524	869	Storage Headboard	5
11525	874	Storage Headboard	5
11526	878	Storage Headboard	5
11527	883	Storage Headboard	5
11528	1041	Storage Headboard	5
11529	1046	Storage Headboard	5
11530	1050	Storage Headboard	5
11531	1055	Storage Headboard	5
11532	869	Bookcase Headboard	6
11533	874	Bookcase Headboard	6
11534	878	Bookcase Headboard	6
11535	883	Bookcase Headboard	6
11536	1041	Bookcase Headboard	6
11537	1046	Bookcase Headboard	6
11538	1050	Bookcase Headboard	6
11539	1055	Bookcase Headboard	6
11540	890	With Mirror	1
11541	890	Without Mirror	2
11542	1066	With Mirror	1
11543	1066	Without Mirror	2
11544	1071	With Mirror	1
11545	1071	Without Mirror	2
11546	1076	With Mirror	1
11547	1076	Without Mirror	2
11548	840	Fabric	1
11549	840	Leather	2
11550	840	Leatherette	3
11551	840	Velvet	4
11552	840	Suede	5
11553	840	No Upholstery	6
11554	982	Fabric	1
11555	982	Leather	2
11556	982	Leatherette	3
11557	982	Velvet	4
11558	982	Suede	5
11559	982	No Upholstery	6
11560	1107	Fabric	1
11561	1107	Leather	2
11562	1107	Leatherette	3
11563	1107	Velvet	4
11564	1107	Suede	5
11565	1107	No Upholstery	6
11566	1645	Original	1
11567	1649	Original	1
11568	1661	Original	1
11569	1665	Original	1
11570	1669	Original	1
11571	1673	Original	1
11572	1745	Original	1
11573	1808	Original	1
11574	1812	Original	1
11575	1825	Original	1
11576	1837	Original	1
11577	1841	Original	1
11578	1645	Chocolate	2
11579	1649	Chocolate	2
11580	1661	Chocolate	2
11581	1665	Chocolate	2
11582	1669	Chocolate	2
11583	1673	Chocolate	2
11584	1745	Chocolate	2
11585	1808	Chocolate	2
11586	1812	Chocolate	2
11587	1825	Chocolate	2
11588	1837	Chocolate	2
11589	1841	Chocolate	2
11590	1645	Vanilla	3
11591	1649	Vanilla	3
11592	1661	Vanilla	3
11593	1665	Vanilla	3
11594	1669	Vanilla	3
11595	1673	Vanilla	3
11596	1745	Vanilla	3
11597	1808	Vanilla	3
11598	1812	Vanilla	3
11599	1825	Vanilla	3
11600	1837	Vanilla	3
11601	1841	Vanilla	3
11602	1645	Strawberry	4
11603	1649	Strawberry	4
11604	1661	Strawberry	4
11605	1665	Strawberry	4
11606	1669	Strawberry	4
11607	1673	Strawberry	4
11608	1745	Strawberry	4
11609	1808	Strawberry	4
11610	1812	Strawberry	4
11611	1825	Strawberry	4
11612	1837	Strawberry	4
11613	1841	Strawberry	4
11614	1645	Mango	5
11615	1649	Mango	5
11616	1661	Mango	5
11617	1665	Mango	5
11618	1669	Mango	5
11619	1673	Mango	5
11620	1745	Mango	5
11621	1808	Mango	5
11622	1812	Mango	5
11623	1825	Mango	5
11624	1837	Mango	5
11625	1841	Mango	5
11626	1645	Butter	6
11627	1649	Butter	6
11628	1661	Butter	6
11629	1665	Butter	6
11630	1669	Butter	6
11631	1673	Butter	6
11632	1745	Butter	6
11633	1808	Butter	6
11634	1812	Butter	6
11635	1825	Butter	6
11636	1837	Butter	6
11637	1841	Butter	6
11638	1645	Salted	7
11639	1649	Salted	7
11640	1661	Salted	7
11641	1665	Salted	7
11642	1669	Salted	7
11643	1673	Salted	7
11644	1745	Salted	7
11645	1808	Salted	7
11646	1812	Salted	7
11647	1825	Salted	7
11648	1837	Salted	7
11649	1841	Salted	7
11650	1645	Cheese	8
11651	1649	Cheese	8
11652	1661	Cheese	8
11653	1665	Cheese	8
11654	1669	Cheese	8
11655	1673	Cheese	8
11656	1745	Cheese	8
11657	1808	Cheese	8
11658	1812	Cheese	8
11659	1825	Cheese	8
11660	1837	Cheese	8
11661	1841	Cheese	8
11662	1645	Masala	9
11663	1649	Masala	9
11664	1661	Masala	9
11665	1665	Masala	9
11666	1669	Masala	9
11667	1673	Masala	9
11668	1745	Masala	9
11669	1808	Masala	9
11670	1812	Masala	9
11671	1825	Masala	9
11672	1837	Masala	9
11673	1841	Masala	9
11674	1645	Tomato	10
11675	1649	Tomato	10
11676	1661	Tomato	10
11677	1665	Tomato	10
11678	1669	Tomato	10
11679	1673	Tomato	10
11680	1745	Tomato	10
11681	1808	Tomato	10
11682	1812	Tomato	10
11683	1825	Tomato	10
11684	1837	Tomato	10
11685	1841	Tomato	10
11686	1645	Spicy	11
11687	1649	Spicy	11
11688	1661	Spicy	11
11689	1665	Spicy	11
11690	1669	Spicy	11
11691	1673	Spicy	11
11692	1745	Spicy	11
11693	1808	Spicy	11
11694	1812	Spicy	11
11695	1825	Spicy	11
11696	1837	Spicy	11
11697	1841	Spicy	11
11698	1645	Plain	12
11699	1649	Plain	12
11700	1661	Plain	12
11701	1665	Plain	12
11702	1669	Plain	12
11703	1673	Plain	12
11704	1745	Plain	12
11705	1808	Plain	12
11706	1812	Plain	12
11707	1825	Plain	12
11708	1837	Plain	12
11709	1841	Plain	12
11710	1697	Regular	1
11711	1701	Regular	1
11712	1785	Regular	1
11713	1789	Regular	1
11714	1793	Regular	1
11715	1797	Regular	1
11716	1805	Regular	1
11717	1697	Fresh	2
11718	1701	Fresh	2
11719	1785	Fresh	2
11720	1789	Fresh	2
11721	1793	Fresh	2
11722	1797	Fresh	2
11723	1805	Fresh	2
11724	1697	Cool	3
11725	1701	Cool	3
11726	1785	Cool	3
11727	1789	Cool	3
11728	1793	Cool	3
11729	1797	Cool	3
11730	1805	Cool	3
11731	1697	Strong	4
11732	1701	Strong	4
11733	1785	Strong	4
11734	1789	Strong	4
11735	1793	Strong	4
11736	1797	Strong	4
11737	1805	Strong	4
11738	1697	Mild	5
11739	1701	Mild	5
11740	1785	Mild	5
11741	1789	Mild	5
11742	1793	Mild	5
11743	1797	Mild	5
11744	1805	Mild	5
11745	1697	Sensitive	6
11746	1701	Sensitive	6
11747	1785	Sensitive	6
11748	1789	Sensitive	6
11749	1793	Sensitive	6
11750	1797	Sensitive	6
11751	1805	Sensitive	6
11752	1697	Natural	7
11753	1701	Natural	7
11754	1785	Natural	7
11755	1789	Natural	7
11756	1793	Natural	7
11757	1797	Natural	7
11758	1805	Natural	7
11759	1697	Active	8
11760	1701	Active	8
11761	1785	Active	8
11762	1789	Active	8
11763	1793	Active	8
11764	1797	Active	8
11765	1805	Active	8
11766	1697	Original	9
11767	1701	Original	9
11768	1785	Original	9
11769	1789	Original	9
11770	1793	Original	9
11771	1797	Original	9
11772	1805	Original	9
11773	1717	100% Pure	1
11774	1717	99.9% Pure	2
11775	1717	Pure	3
11776	1717	Organic	4
11777	2489	100% Pure	1
11778	2489	99.9% Pure	2
11779	2489	Pure	3
11780	2489	Organic	4
11781	2494	100% Pure	1
11782	2494	99.9% Pure	2
11783	2494	Pure	3
11784	2494	Organic	4
11785	1631	Sunflower	1
11786	1769	Sunflower	1
11787	1773	Sunflower	1
11788	1777	Sunflower	1
11789	1781	Sunflower	1
11790	1631	Mustard	2
11791	1769	Mustard	2
11792	1773	Mustard	2
11793	1777	Mustard	2
11794	1781	Mustard	2
11795	1631	Coconut	3
11796	1769	Coconut	3
11797	1773	Coconut	3
11798	1777	Coconut	3
11799	1781	Coconut	3
11800	1631	Olive	4
11801	1769	Olive	4
11802	1773	Olive	4
11803	1777	Olive	4
11804	1781	Olive	4
11805	1631	Groundnut	5
11806	1769	Groundnut	5
11807	1773	Groundnut	5
11808	1777	Groundnut	5
11809	1781	Groundnut	5
11810	1631	Rice Bran	6
11811	1769	Rice Bran	6
11812	1773	Rice Bran	6
11813	1777	Rice Bran	6
11814	1781	Rice Bran	6
11815	1631	Palm	7
11816	1769	Palm	7
11817	1773	Palm	7
11818	1777	Palm	7
11819	1781	Palm	7
11820	1631	Soybean	8
11821	1769	Soybean	8
11822	1773	Soybean	8
11823	1777	Soybean	8
11824	1781	Soybean	8
11825	1390	1-2 hours	1
11826	1453	1-2 hours	1
11827	1493	1-2 hours	1
11828	1497	1-2 hours	1
11829	1502	1-2 hours	1
11830	1507	1-2 hours	1
11831	1551	1-2 hours	1
11832	1556	1-2 hours	1
11833	1572	1-2 hours	1
11834	2525	1-2 hours	1
11835	1390	3-4 hours	2
11836	1453	3-4 hours	2
11837	1493	3-4 hours	2
11838	1497	3-4 hours	2
11839	1502	3-4 hours	2
11840	1507	3-4 hours	2
11841	1551	3-4 hours	2
11842	1556	3-4 hours	2
11843	1572	3-4 hours	2
11844	2525	3-4 hours	2
11845	1390	5-6 hours	3
11846	1453	5-6 hours	3
11847	1493	5-6 hours	3
11848	1497	5-6 hours	3
11849	1502	5-6 hours	3
11850	1507	5-6 hours	3
11851	1551	5-6 hours	3
11852	1556	5-6 hours	3
11853	1572	5-6 hours	3
11854	2525	5-6 hours	3
11855	1390	7-8 hours	4
11856	1453	7-8 hours	4
11857	1493	7-8 hours	4
11858	1497	7-8 hours	4
11859	1502	7-8 hours	4
11860	1507	7-8 hours	4
11861	1551	7-8 hours	4
11862	1556	7-8 hours	4
11863	1572	7-8 hours	4
11864	2525	7-8 hours	4
11865	1390	1 day	5
11866	1453	1 day	5
11867	1493	1 day	5
11868	1497	1 day	5
11869	1502	1 day	5
11870	1507	1 day	5
11871	1551	1 day	5
11872	1556	1 day	5
11873	1572	1 day	5
11874	2525	1 day	5
11875	1390	2 days	6
11876	1453	2 days	6
11877	1493	2 days	6
11878	1497	2 days	6
11879	1502	2 days	6
11880	1507	2 days	6
11881	1551	2 days	6
11882	1556	2 days	6
11883	1572	2 days	6
11884	2525	2 days	6
11885	1390	5 days	7
11886	1453	5 days	7
11887	1493	5 days	7
11888	1497	5 days	7
11889	1502	5 days	7
11890	1507	5 days	7
11891	1551	5 days	7
11892	1556	5 days	7
11893	1572	5 days	7
11894	2525	5 days	7
11895	1390	1 week	8
11896	1453	1 week	8
11897	1493	1 week	8
11898	1497	1 week	8
11899	1502	1 week	8
11900	1507	1 week	8
11901	1551	1 week	8
11902	1556	1 week	8
11903	1572	1 week	8
11904	2525	1 week	8
11905	1390	2 weeks	9
11906	1453	2 weeks	9
11907	1493	2 weeks	9
11908	1497	2 weeks	9
11909	1502	2 weeks	9
11910	1507	2 weeks	9
11911	1551	2 weeks	9
11912	1556	2 weeks	9
11913	1572	2 weeks	9
11914	2525	2 weeks	9
11915	1391	Not Water Resistant	1
11916	1391	Splash Proof	2
11917	1391	IPX4	3
11918	1391	IPX5	4
11919	1391	IPX6	5
11920	1391	IPX7	6
11921	1391	IP67	7
11922	1391	IP68	8
11923	1503	Not Water Resistant	1
11924	1503	Splash Proof	2
11925	1503	IPX4	3
11926	1503	IPX5	4
11927	1503	IPX6	5
11928	1503	IPX7	6
11929	1503	IP67	7
11930	1503	IP68	8
11931	1552	Not Water Resistant	1
11932	1552	Splash Proof	2
11933	1552	IPX4	3
11934	1552	IPX5	4
11935	1552	IPX6	5
11936	1552	IPX7	6
11937	1552	IP67	7
11938	1552	IP68	8
11939	2516	Not Water Resistant	1
11940	2516	Splash Proof	2
11941	2516	IPX4	3
11942	2516	IPX5	4
11943	2516	IPX6	5
11944	2516	IPX7	6
11945	2516	IP67	7
11946	2516	IP68	8
11947	1331	Inverter	1
11948	1331	Non-Inverter	2
11949	1336	Inverter	1
11950	1336	Non-Inverter	2
11951	918	Single Door	1
11952	1127	Single Door	1
11953	1132	Single Door	1
11954	1137	Single Door	1
11955	1340	Single Door	1
11956	918	Double Door	2
11957	1127	Double Door	2
11958	1132	Double Door	2
11959	1137	Double Door	2
11960	1340	Double Door	2
11961	918	Triple Door	3
11962	1127	Triple Door	3
11963	1132	Triple Door	3
11964	1137	Triple Door	3
11965	1340	Triple Door	3
11966	918	French Door	4
11967	1127	French Door	4
11968	1132	French Door	4
11969	1137	French Door	4
11970	1340	French Door	4
11971	918	Side by Side	5
11972	1127	Side by Side	5
11973	1132	Side by Side	5
11974	1137	Side by Side	5
11975	1340	Side by Side	5
11976	918	Bottom Freezer	6
11977	1127	Bottom Freezer	6
11978	1132	Bottom Freezer	6
11979	1137	Bottom Freezer	6
11980	1340	Bottom Freezer	6
11981	933	Yes	1
11982	933	No	2
11983	953	Yes	1
11984	953	No	2
11985	954	Yes	1
11986	954	No	2
11987	983	Yes	1
11988	983	No	2
11989	984	Yes	1
11990	984	No	2
11991	1099	Yes	1
11992	1099	No	2
11993	1113	Yes	1
11994	1113	No	2
11995	1114	Yes	1
11996	1114	No	2
11997	1608	Yes	1
11998	1608	No	2
11999	1292	Smart TV	1
12000	1292	Non-Smart TV	2
12001	1304	Smart TV	1
12002	1304	Non-Smart TV	2
12003	1309	Smart TV	1
12004	1309	Non-Smart TV	2
12005	1314	Smart TV	1
12006	1314	Non-Smart TV	2
12007	1319	Smart TV	1
12008	1319	Non-Smart TV	2
12009	1325	Smart TV	1
12010	1325	Non-Smart TV	2
12011	2096	Embroidery	1
12012	2138	Embroidery	1
12013	2174	Embroidery	1
12014	2342	Embroidery	1
12015	2348	Embroidery	1
12016	2354	Embroidery	1
12017	2360	Embroidery	1
12018	2553	Embroidery	1
12019	2558	Embroidery	1
12020	2563	Embroidery	1
12021	2568	Embroidery	1
12022	2573	Embroidery	1
12023	2096	Hand Work	2
12024	2138	Hand Work	2
12025	2174	Hand Work	2
12026	2342	Hand Work	2
12027	2348	Hand Work	2
12028	2354	Hand Work	2
12029	2360	Hand Work	2
12030	2553	Hand Work	2
12031	2558	Hand Work	2
12032	2563	Hand Work	2
12033	2568	Hand Work	2
12034	2573	Hand Work	2
12035	2096	Machine Work	3
12036	2138	Machine Work	3
12037	2174	Machine Work	3
12038	2342	Machine Work	3
12039	2348	Machine Work	3
12040	2354	Machine Work	3
12041	2360	Machine Work	3
12042	2553	Machine Work	3
12043	2558	Machine Work	3
12044	2563	Machine Work	3
12045	2568	Machine Work	3
12046	2573	Machine Work	3
12047	2096	Zari Work	4
12048	2138	Zari Work	4
12049	2174	Zari Work	4
12050	2342	Zari Work	4
12051	2348	Zari Work	4
12052	2354	Zari Work	4
12053	2360	Zari Work	4
12054	2553	Zari Work	4
12055	2558	Zari Work	4
12056	2563	Zari Work	4
12057	2568	Zari Work	4
12058	2573	Zari Work	4
12059	2096	Stone Work	5
12060	2138	Stone Work	5
12061	2174	Stone Work	5
12062	2342	Stone Work	5
12063	2348	Stone Work	5
12064	2354	Stone Work	5
12065	2360	Stone Work	5
12066	2553	Stone Work	5
12067	2558	Stone Work	5
12068	2563	Stone Work	5
12069	2568	Stone Work	5
12070	2573	Stone Work	5
12071	2096	Plain	6
12072	2138	Plain	6
12073	2174	Plain	6
12074	2342	Plain	6
12075	2348	Plain	6
12076	2354	Plain	6
12077	2360	Plain	6
12078	2553	Plain	6
12079	2558	Plain	6
12080	2563	Plain	6
12081	2568	Plain	6
12082	2573	Plain	6
12083	2096	Printed	7
12084	2138	Printed	7
12085	2174	Printed	7
12086	2342	Printed	7
12087	2348	Printed	7
12088	2354	Printed	7
12089	2360	Printed	7
12090	2553	Printed	7
12091	2558	Printed	7
12092	2563	Printed	7
12093	2568	Printed	7
12094	2573	Printed	7
12095	2096	Beaded	8
12096	2138	Beaded	8
12097	2174	Beaded	8
12098	2342	Beaded	8
12099	2348	Beaded	8
12100	2354	Beaded	8
12101	2360	Beaded	8
12102	2553	Beaded	8
12103	2558	Beaded	8
12104	2563	Beaded	8
12105	2568	Beaded	8
12106	2573	Beaded	8
12107	1122	Yes	1
12108	1122	No	2
12109	1438	Yes	1
12110	1438	No	2
12111	2343	Yes	1
12112	2343	No	2
12113	2349	Yes	1
12114	2349	No	2
12115	2355	Yes	1
12116	2355	No	2
12117	2555	Yes	1
12118	2555	No	2
12119	2560	Yes	1
12120	2560	No	2
12121	2570	Yes	1
12122	2570	No	2
12123	2575	Yes	1
12124	2575	No	2
12125	2085	Regular Sleeve	1
12126	2187	Regular Sleeve	1
12127	2193	Regular Sleeve	1
12128	2260	Regular Sleeve	1
12129	2277	Regular Sleeve	1
12130	2391	Regular Sleeve	1
12131	2409	Regular Sleeve	1
12132	2085	Cap Sleeve	2
12133	2187	Cap Sleeve	2
12134	2193	Cap Sleeve	2
12135	2260	Cap Sleeve	2
12136	2277	Cap Sleeve	2
12137	2391	Cap Sleeve	2
12138	2409	Cap Sleeve	2
12139	2085	Puff Sleeve	3
12140	2187	Puff Sleeve	3
12141	2193	Puff Sleeve	3
12142	2260	Puff Sleeve	3
12143	2277	Puff Sleeve	3
12144	2391	Puff Sleeve	3
12145	2409	Puff Sleeve	3
12146	2085	Bell Sleeve	4
12147	2187	Bell Sleeve	4
12148	2193	Bell Sleeve	4
12149	2260	Bell Sleeve	4
12150	2277	Bell Sleeve	4
12151	2391	Bell Sleeve	4
12152	2409	Bell Sleeve	4
12153	2085	Raglan Sleeve	5
12154	2187	Raglan Sleeve	5
12155	2193	Raglan Sleeve	5
12156	2260	Raglan Sleeve	5
12157	2277	Raglan Sleeve	5
12158	2391	Raglan Sleeve	5
12159	2409	Raglan Sleeve	5
12160	2085	Drop Sleeve	6
12161	2187	Drop Sleeve	6
12162	2193	Drop Sleeve	6
12163	2260	Drop Sleeve	6
12164	2277	Drop Sleeve	6
12165	2391	Drop Sleeve	6
12166	2409	Drop Sleeve	6
12167	2085	Sleeveless	7
12168	2187	Sleeveless	7
12169	2193	Sleeveless	7
12170	2260	Sleeveless	7
12171	2277	Sleeveless	7
12172	2391	Sleeveless	7
12173	2409	Sleeveless	7
12174	2232	A	1
12175	2232	B	2
12176	2232	C	3
12177	2232	D	4
12178	2232	DD	5
12179	2232	E	6
12180	2297	A	1
12181	2297	B	2
12182	2297	C	3
12183	2297	D	4
12184	2297	DD	5
12185	2297	E	6
12186	2108	Light Wash	1
12187	2108	Medium Wash	2
12188	2108	Dark Wash	3
12189	2108	Stone Wash	4
12190	2108	Acid Wash	5
12191	2108	Clean Look	6
12192	2396	Light Wash	1
12193	2396	Medium Wash	2
12194	2396	Dark Wash	3
12195	2396	Stone Wash	4
12196	2396	Acid Wash	5
12197	2396	Clean Look	6
12198	2464	Flat	1
12199	2464	1 inch	2
12200	2464	2 inch	3
12201	2464	3 inch	4
12202	2464	4 inch	5
12203	2464	5+ inch	6
12204	2198	Ankle Strap	1
12205	2198	T-Strap	2
12206	2198	Sling Back	3
12207	2198	Backstrap	4
12208	2198	No Strap	5
12209	2198	Adjustable Strap	6
12210	2461	Ankle Strap	1
12211	2461	T-Strap	2
12212	2461	Sling Back	3
12213	2461	Backstrap	4
12214	2461	No Strap	5
12215	2461	Adjustable Strap	6
12216	2149	Pin Buckle	1
12217	2149	Automatic Buckle	2
12218	2149	Reversible Buckle	3
12219	2149	Plate Buckle	4
12220	2149	No Buckle	5
12221	1541	USB 2.0	1
12222	1541	USB 3.0	2
12223	1541	USB 3.1	3
12224	1541	USB-C	4
12225	1541	SATA	5
12226	1541	NVMe	6
12227	1541	M.2	7
12228	1541	Thunderbolt	8
12229	1546	USB 2.0	1
12230	1546	USB 3.0	2
12231	1546	USB 3.1	3
12232	1546	USB-C	4
12233	1546	SATA	5
12234	1546	NVMe	6
12235	1546	M.2	7
12236	1546	Thunderbolt	8
12237	1437	Type-A	1
12238	1468	Type-A	1
12239	1533	Type-A	1
12240	1538	Type-A	1
12241	1568	Type-A	1
12242	1437	Type-C	2
12243	1468	Type-C	2
12244	1533	Type-C	2
12245	1538	Type-C	2
12246	1568	Type-C	2
12247	1437	Micro USB	3
12248	1468	Micro USB	3
12249	1533	Micro USB	3
12250	1538	Micro USB	3
12251	1568	Micro USB	3
12252	1437	Lightning	4
12253	1468	Lightning	4
12254	1533	Lightning	4
12255	1538	Lightning	4
12256	1568	Lightning	4
12257	1437	3.5mm Jack	5
12258	1468	3.5mm Jack	5
12259	1533	3.5mm Jack	5
12260	1538	3.5mm Jack	5
12261	1568	3.5mm Jack	5
12262	1437	USB-B	6
12263	1468	USB-B	6
12264	1533	USB-B	6
12265	1538	USB-B	6
12266	1568	USB-B	6
12267	1437	Mini USB	7
12268	1468	Mini USB	7
12269	1533	Mini USB	7
12270	1538	Mini USB	7
12271	1568	Mini USB	7
12272	1455	802.11ac	1
12273	1455	802.11ax (WiFi 6)	2
12274	1455	802.11n	3
12275	1455	802.11g	4
12276	1455	WiFi 6E	5
12277	1455	Dual Band	6
12278	1455	Tri Band	7
12279	1460	802.11ac	1
12280	1460	802.11ax (WiFi 6)	2
12281	1460	802.11n	3
12282	1460	802.11g	4
12283	1460	WiFi 6E	5
12284	1460	Dual Band	6
12285	1460	Tri Band	7
12286	1456	150 Mbps	1
12287	1456	300 Mbps	2
12288	1456	450 Mbps	3
12289	1456	600 Mbps	4
12290	1456	1200 Mbps	5
12291	1456	1800 Mbps	6
12292	1456	3000 Mbps	7
12293	1456	5400 Mbps	8
12294	1462	150 Mbps	1
12295	1462	300 Mbps	2
12296	1462	450 Mbps	3
12297	1462	600 Mbps	4
12298	1462	1200 Mbps	5
12299	1462	1800 Mbps	6
12300	1462	3000 Mbps	7
12301	1462	5400 Mbps	8
12302	1471	150 Mbps	1
12303	1471	300 Mbps	2
12304	1471	450 Mbps	3
12305	1471	600 Mbps	4
12306	1471	1200 Mbps	5
12307	1471	1800 Mbps	6
12308	1471	3000 Mbps	7
12309	1471	5400 Mbps	8
12310	1428	Wired	1
12311	1428	Wireless	2
12312	1428	Both	3
12313	1427	PlayStation 5	1
12314	1427	PlayStation 4	2
12315	1427	Xbox Series X/S	3
12316	1427	Xbox One	4
12317	1427	Nintendo Switch	5
12318	1427	PC	6
12319	1427	Multi-Platform	7
12320	1414	Standard Edition	1
12321	1414	Deluxe Edition	2
12322	1414	Ultimate Edition	3
12323	1414	Collector Edition	4
12324	1414	GOTY Edition	5
12325	1418	Standard Edition	1
12326	1418	Deluxe Edition	2
12327	1418	Ultimate Edition	3
12328	1418	Collector Edition	4
12329	1418	GOTY Edition	5
12330	1416	Region Free	1
12331	1416	NTSC	2
12332	1416	PAL	3
12333	1416	Region 1	4
12334	1416	Region 2	5
12335	1416	Region 3	6
12336	1420	Region Free	1
12337	1420	NTSC	2
12338	1420	PAL	3
12339	1420	Region 1	4
12340	1420	Region 2	5
12341	1420	Region 3	6
12342	1424	Region Free	1
12343	1424	NTSC	2
12344	1424	PAL	3
12345	1424	Region 1	4
12346	1424	Region 2	5
12347	1424	Region 3	6
12348	1435	Wall Charger	1
12349	1435	Car Charger	2
12350	1435	Wireless Charger	3
12351	1435	Power Bank	4
12352	1435	Fast Charger	5
12353	1435	Multi-Port Charger	6
12354	1436	5W	1
12355	1436	10W	2
12356	1436	15W	3
12357	1436	18W	4
12358	1436	20W	5
12359	1436	25W	6
12360	1436	30W	7
12361	1436	45W	8
12362	1436	65W	9
12363	1436	100W	10
12364	1451	6mm	1
12365	1557	6mm	1
12366	1565	6mm	1
12367	1570	6mm	1
12368	1575	6mm	1
12369	1451	8mm	2
12370	1557	8mm	2
12371	1565	8mm	2
12372	1570	8mm	2
12373	1575	8mm	2
12374	1451	10mm	3
12375	1557	10mm	3
12376	1565	10mm	3
12377	1570	10mm	3
12378	1575	10mm	3
12379	1451	13mm	4
12380	1557	13mm	4
12381	1565	13mm	4
12382	1570	13mm	4
12383	1575	13mm	4
12384	1451	30mm	5
12385	1557	30mm	5
12386	1565	30mm	5
12387	1570	30mm	5
12388	1575	30mm	5
12389	1451	40mm	6
12390	1557	40mm	6
12391	1565	40mm	6
12392	1570	40mm	6
12393	1575	40mm	6
12394	1451	50mm	7
12395	1557	50mm	7
12396	1565	50mm	7
12397	1570	50mm	7
12398	1575	50mm	7
12399	1478	All Devices	1
12400	1478	Android	2
12401	1478	iOS	3
12402	1478	Windows	4
12403	1478	MacOS	5
12404	1478	Linux	6
12405	1478	Smart TV	7
12406	1482	All Devices	1
12407	1482	Android	2
12408	1482	iOS	3
12409	1482	Windows	4
12410	1482	MacOS	5
12411	1482	Linux	6
12412	1482	Smart TV	7
12413	1034	Single	1
12414	1034	Double	2
12415	1034	Queen	3
12416	1034	King	4
12417	1034	L-Shaped	5
12418	1034	U-Shaped	6
12419	1034	Sectional	7
12420	1034	Modular	8
12421	1059	Twin over Twin	1
12422	1059	Twin over Full	2
12423	1059	Full over Full	3
12424	1059	L-Shaped Bunk	4
12425	1059	Triple Bunk	5
12426	1061	Both Sides	1
12427	1061	One Side	2
12428	1061	Removable	3
12429	1061	No Rails	4
12430	1062	Straight Ladder	1
12431	1062	Angled Ladder	2
12432	1062	Built-in Steps	3
12433	1062	No Ladder	4
12434	1067	With Drawers	1
12435	1067	With Shelves	2
12436	1067	With Hanging Rod	3
12437	1067	Full Hanging	4
12438	1067	Customizable	5
12439	1072	With Drawers	1
12440	1072	With Shelves	2
12441	1072	With Hanging Rod	3
12442	1072	Full Hanging	4
12443	1072	Customizable	5
12444	1077	With Drawers	1
12445	1077	With Shelves	2
12446	1077	With Hanging Rod	3
12447	1077	Full Hanging	4
12448	1077	Customizable	5
12449	1080	Hinged Door	1
12450	1080	Sliding Door	2
12451	1080	Folding Door	3
12452	1080	Push to Open	4
12453	1082	All Doors	1
12454	1082	Partial Doors	2
12455	1082	No Mirror	3
12456	853	Laminate	1
12457	908	Laminate	1
12458	928	Laminate	1
12459	957	Laminate	1
12460	1085	Laminate	1
12461	1091	Laminate	1
12462	853	Glass	2
12463	908	Glass	2
12464	928	Glass	2
12465	957	Glass	2
12466	1085	Glass	2
12467	1091	Glass	2
12468	853	Marble	3
12469	908	Marble	3
12470	928	Marble	3
12471	957	Marble	3
12472	1085	Marble	3
12473	1091	Marble	3
12474	853	Granite	4
12475	908	Granite	4
12476	928	Granite	4
12477	957	Granite	4
12478	1085	Granite	4
12479	1091	Granite	4
12480	853	Wood	5
12481	908	Wood	5
12482	928	Wood	5
12483	957	Wood	5
12484	1085	Wood	5
12485	1091	Wood	5
12486	853	MDF	6
12487	908	MDF	6
12488	928	MDF	6
12489	957	MDF	6
12490	1085	MDF	6
12491	1091	MDF	6
12492	853	Metal	7
12493	908	Metal	7
12494	928	Metal	7
12495	957	Metal	7
12496	1085	Metal	7
12497	1091	Metal	7
12498	905	Fixed Shelves	1
12499	905	Adjustable Shelves	2
12500	905	Pull-out Shelves	3
12501	905	No Shelves	4
12502	1104	Fixed Shelves	1
12503	1104	Adjustable Shelves	2
12504	1104	Pull-out Shelves	3
12505	1104	No Shelves	4
12506	1117	Manual	1
12507	1117	Push Back	2
12508	1117	Lever	3
12509	1117	Electric	4
12510	1117	Rocker	5
12511	1853	Toor Dal	1
12512	1853	Moong Dal	2
12513	1853	Urad Dal	3
12514	1853	Chana Dal	4
12515	1853	Masoor Dal	5
12516	1853	Mixed Dal	6
12517	1635	White Sugar	1
12518	1635	Brown Sugar	2
12519	1635	Jaggery Powder	3
12520	1635	Rock Sugar	4
12521	1635	Organic Sugar	5
12522	1635	Sugar Free	6
12523	1639	Iodized Salt	1
12524	1639	Rock Salt	2
12525	1639	Sea Salt	3
12526	1639	Black Salt	4
12527	1639	Himalayan Pink Salt	5
12528	1639	Low Sodium	6
12529	1725	Full Cream	1
12530	1725	Toned	2
12531	1725	Double Toned	3
12532	1725	Skimmed	4
12533	1725	Low Fat	5
12534	1725	Fat Free	6
12535	1814	Whey Protein	1
12536	1814	Casein Protein	2
12537	1814	Plant Protein	3
12538	1814	Soy Protein	4
12539	1814	Pea Protein	5
12540	1814	Mixed Protein	6
12541	357	Boneless	1
12542	357	With Bone	2
12543	357	Mince	3
12544	357	Curry Cut	4
12545	357	Steak Cut	5
12546	357	Whole	6
12547	448	Boneless	1
12548	448	With Bone	2
12549	448	Mince	3
12550	448	Curry Cut	4
12551	448	Steak Cut	5
12552	448	Whole	6
12553	1828	Boneless	1
12554	1828	With Bone	2
12555	1828	Mince	3
12556	1828	Curry Cut	4
12557	1828	Steak Cut	5
12558	1828	Whole	6
12559	1832	Leafy Greens	1
12560	1832	Root Vegetables	2
12561	1832	Cruciferous	3
12562	1832	Gourd Family	4
12563	1832	Podded Vegetables	5
12564	1832	Nightshades	6
12565	1832	Mixed	7
12566	1872	Citrus	1
12567	1872	Berries	2
12568	1872	Tropical	3
12569	1872	Stone Fruits	4
12570	1872	Pomaceous	5
12571	1872	Melons	6
12572	1872	Exotic	7
12573	1889	Milk Based	1
12574	1889	Dry Sweets	2
12575	1889	Syrup Based	3
12576	1889	Chocolate Based	4
12577	1889	Sugar Free	5
12578	1889	Traditional	6
12579	1860	Basmati	1
12580	1860	Non-Basmati	2
12581	1860	Jasmine	3
12582	1860	Sona Masoori	4
12583	1860	Brown Rice	5
12584	1860	Red Rice	6
12585	1860	Black Rice	7
12586	1860	Organic	8
12587	1864	Basmati	1
12588	1864	Non-Basmati	2
12589	1864	Jasmine	3
12590	1864	Sona Masoori	4
12591	1864	Brown Rice	5
12592	1864	Red Rice	6
12593	1864	Black Rice	7
12594	1864	Organic	8
12595	1868	Basmati	1
12596	1868	Non-Basmati	2
12597	1868	Jasmine	3
12598	1868	Sona Masoori	4
12599	1868	Brown Rice	5
12600	1868	Red Rice	6
12601	1868	Black Rice	7
12602	1868	Organic	8
12603	1905	Ball Pen	1
12604	1905	Gel Pen	2
12605	1905	Fountain Pen	3
12606	1905	Marker Pen	4
12607	1905	Sketch Pen	5
12608	1905	Highlighter	6
12609	1909	Ball Pen	1
12610	1909	Gel Pen	2
12611	1909	Fountain Pen	3
12612	1909	Marker Pen	4
12613	1909	Sketch Pen	5
12614	1909	Highlighter	6
12615	1954	Fine (0.5mm)	1
12616	1954	Medium (0.7mm)	2
12617	1954	Bold (1.0mm)	3
12618	1954	Extra Fine (0.3mm)	4
12619	1954	Chisel Tip	5
12620	1954	Brush Tip	6
12621	2000	Fine (0.5mm)	1
12622	2000	Medium (0.7mm)	2
12623	2000	Bold (1.0mm)	3
12624	2000	Extra Fine (0.3mm)	4
12625	2000	Chisel Tip	5
12626	2000	Brush Tip	6
12627	1906	1 Piece	1
12628	1910	1 Piece	1
12629	1921	1 Piece	1
12630	1926	1 Piece	1
12631	1938	1 Piece	1
12632	1945	1 Piece	1
12633	1961	1 Piece	1
12634	1994	1 Piece	1
12635	2002	1 Piece	1
12636	2042	1 Piece	1
12637	2046	1 Piece	1
12638	2058	1 Piece	1
12639	2062	1 Piece	1
12640	1906	2 Pieces	2
12641	1910	2 Pieces	2
12642	1921	2 Pieces	2
12643	1926	2 Pieces	2
12644	1938	2 Pieces	2
12645	1945	2 Pieces	2
12646	1961	2 Pieces	2
12647	1994	2 Pieces	2
12648	2002	2 Pieces	2
12649	2042	2 Pieces	2
12650	2046	2 Pieces	2
12651	2058	2 Pieces	2
12652	2062	2 Pieces	2
12653	1906	5 Pieces	3
12654	1910	5 Pieces	3
12655	1921	5 Pieces	3
12656	1926	5 Pieces	3
12657	1938	5 Pieces	3
12658	1945	5 Pieces	3
12659	1961	5 Pieces	3
12660	1994	5 Pieces	3
12661	2002	5 Pieces	3
12662	2042	5 Pieces	3
12663	2046	5 Pieces	3
12664	2058	5 Pieces	3
12665	2062	5 Pieces	3
12666	1906	10 Pieces	4
12667	1910	10 Pieces	4
12668	1921	10 Pieces	4
12669	1926	10 Pieces	4
12670	1938	10 Pieces	4
12671	1945	10 Pieces	4
12672	1961	10 Pieces	4
12673	1994	10 Pieces	4
12674	2002	10 Pieces	4
12675	2042	10 Pieces	4
12676	2046	10 Pieces	4
12677	2058	10 Pieces	4
12678	2062	10 Pieces	4
12679	1906	12 Pieces	5
12680	1910	12 Pieces	5
12681	1921	12 Pieces	5
12682	1926	12 Pieces	5
12683	1938	12 Pieces	5
12684	1945	12 Pieces	5
12685	1961	12 Pieces	5
12686	1994	12 Pieces	5
12687	2002	12 Pieces	5
12688	2042	12 Pieces	5
12689	2046	12 Pieces	5
12690	2058	12 Pieces	5
12691	2062	12 Pieces	5
12692	1906	20 Pieces	6
12693	1910	20 Pieces	6
12694	1921	20 Pieces	6
12695	1926	20 Pieces	6
12696	1938	20 Pieces	6
12697	1945	20 Pieces	6
12698	1961	20 Pieces	6
12699	1994	20 Pieces	6
12700	2002	20 Pieces	6
12701	2042	20 Pieces	6
12702	2046	20 Pieces	6
12703	2058	20 Pieces	6
12704	2062	20 Pieces	6
12705	1906	50 Pieces	7
12706	1910	50 Pieces	7
12707	1921	50 Pieces	7
12708	1926	50 Pieces	7
12709	1938	50 Pieces	7
12710	1945	50 Pieces	7
12711	1961	50 Pieces	7
12712	1994	50 Pieces	7
12713	2002	50 Pieces	7
12714	2042	50 Pieces	7
12715	2046	50 Pieces	7
12716	2058	50 Pieces	7
12717	2062	50 Pieces	7
12718	1906	100 Pieces	8
12719	1910	100 Pieces	8
12720	1921	100 Pieces	8
12721	1926	100 Pieces	8
12722	1938	100 Pieces	8
12723	1945	100 Pieces	8
12724	1961	100 Pieces	8
12725	1994	100 Pieces	8
12726	2002	100 Pieces	8
12727	2042	100 Pieces	8
12728	2046	100 Pieces	8
12729	2058	100 Pieces	8
12730	2062	100 Pieces	8
12731	1913	70 GSM	1
12732	1958	70 GSM	1
12733	1962	70 GSM	1
12734	1969	70 GSM	1
12735	1990	70 GSM	1
12736	2026	70 GSM	1
12737	2059	70 GSM	1
12738	2063	70 GSM	1
12739	1913	75 GSM	2
12740	1958	75 GSM	2
12741	1962	75 GSM	2
12742	1969	75 GSM	2
12743	1990	75 GSM	2
12744	2026	75 GSM	2
12745	2059	75 GSM	2
12746	2063	75 GSM	2
12747	1913	80 GSM	3
12748	1958	80 GSM	3
12749	1962	80 GSM	3
12750	1969	80 GSM	3
12751	1990	80 GSM	3
12752	2026	80 GSM	3
12753	2059	80 GSM	3
12754	2063	80 GSM	3
12755	1913	90 GSM	4
12756	1958	90 GSM	4
12757	1962	90 GSM	4
12758	1969	90 GSM	4
12759	1990	90 GSM	4
12760	2026	90 GSM	4
12761	2059	90 GSM	4
12762	2063	90 GSM	4
12763	1913	100 GSM	5
12764	1958	100 GSM	5
12765	1962	100 GSM	5
12766	1969	100 GSM	5
12767	1990	100 GSM	5
12768	2026	100 GSM	5
12769	2059	100 GSM	5
12770	2063	100 GSM	5
12771	1913	120 GSM	6
12772	1958	120 GSM	6
12773	1962	120 GSM	6
12774	1969	120 GSM	6
12775	1990	120 GSM	6
12776	2026	120 GSM	6
12777	2059	120 GSM	6
12778	2063	120 GSM	6
12779	1913	150 GSM	7
12780	1958	150 GSM	7
12781	1962	150 GSM	7
12782	1969	150 GSM	7
12783	1990	150 GSM	7
12784	2026	150 GSM	7
12785	2059	150 GSM	7
12786	2063	150 GSM	7
12787	1914	25 Sheets	1
12788	1914	50 Sheets	2
12789	1914	100 Sheets	3
12790	1914	200 Sheets	4
12791	1914	500 Sheets	5
12792	1914	1000 Sheets	6
12793	1970	25 Sheets	1
12794	1970	50 Sheets	2
12795	1970	100 Sheets	3
12796	1970	200 Sheets	4
12797	1970	500 Sheets	5
12798	1970	1000 Sheets	6
12799	2066	25 Sheets	1
12800	2066	50 Sheets	2
12801	2066	100 Sheets	3
12802	2066	200 Sheets	4
12803	2066	500 Sheets	5
12804	2066	1000 Sheets	6
12805	1934	Single Line	1
12806	1934	Double Line	2
12807	1934	Four Line	3
12808	1934	Square Grid	4
12809	1934	Blank	5
12810	1934	Dotted	6
12811	1933	50 Pages	1
12812	1933	100 Pages	2
12813	1933	120 Pages	3
12814	1933	160 Pages	4
12815	1933	200 Pages	5
12816	1933	240 Pages	6
12817	1933	300 Pages	7
12818	1957	50 Pages	1
12819	1957	100 Pages	2
12820	1957	120 Pages	3
12821	1957	160 Pages	4
12822	1957	200 Pages	5
12823	1957	240 Pages	6
12824	1957	300 Pages	7
12825	2025	50 Pages	1
12826	2025	100 Pages	2
12827	2025	120 Pages	3
12828	2025	160 Pages	4
12829	2025	200 Pages	5
12830	2025	240 Pages	6
12831	2025	300 Pages	7
12832	347	9mm	1
12833	350	9mm	1
12834	434	9mm	1
12835	454	9mm	1
12836	490	9mm	1
12837	1996	9mm	1
12838	347	12mm	2
12839	350	12mm	2
12840	434	12mm	2
12841	454	12mm	2
12842	490	12mm	2
12843	1996	12mm	2
12844	347	18mm	3
12845	350	18mm	3
12846	434	18mm	3
12847	454	18mm	3
12848	490	18mm	3
12849	1996	18mm	3
12850	347	25mm	4
12851	350	25mm	4
12852	434	25mm	4
12853	454	25mm	4
12854	490	25mm	4
12855	1996	25mm	4
12856	2037	12mm	1
12857	2037	18mm	2
12858	2037	24mm	3
12859	2037	36mm	4
12860	2037	48mm	5
12861	2051	1 inch	1
12862	2051	2 inch	2
12863	2051	3 inch	3
12864	41	Coarse Thread	1
12865	48	Coarse Thread	1
12866	55	Coarse Thread	1
12867	62	Coarse Thread	1
12868	69	Coarse Thread	1
12869	76	Coarse Thread	1
12870	81	Coarse Thread	1
12871	86	Coarse Thread	1
12872	41	Fine Thread	2
12873	48	Fine Thread	2
12874	55	Fine Thread	2
12875	62	Fine Thread	2
12876	69	Fine Thread	2
12877	76	Fine Thread	2
12878	81	Fine Thread	2
12879	86	Fine Thread	2
12880	41	Machine Thread	3
12881	48	Machine Thread	3
12882	55	Machine Thread	3
12883	62	Machine Thread	3
12884	69	Machine Thread	3
12885	76	Machine Thread	3
12886	81	Machine Thread	3
12887	86	Machine Thread	3
12888	41	Wood Thread	4
12889	48	Wood Thread	4
12890	55	Wood Thread	4
12891	62	Wood Thread	4
12892	69	Wood Thread	4
12893	76	Wood Thread	4
12894	81	Wood Thread	4
12895	86	Wood Thread	4
12896	41	Self-Tapping	5
12897	48	Self-Tapping	5
12898	55	Self-Tapping	5
12899	62	Self-Tapping	5
12900	69	Self-Tapping	5
12901	76	Self-Tapping	5
12902	81	Self-Tapping	5
12903	86	Self-Tapping	5
12904	41	No Thread	6
12905	48	No Thread	6
12906	55	No Thread	6
12907	62	No Thread	6
12908	69	No Thread	6
12909	76	No Thread	6
12910	81	No Thread	6
12911	86	No Thread	6
12912	40	Flat Head	1
12913	47	Flat Head	1
12914	54	Flat Head	1
12915	61	Flat Head	1
12916	68	Flat Head	1
12917	75	Flat Head	1
12918	640	Flat Head	1
12919	40	Round Head	2
12920	47	Round Head	2
12921	54	Round Head	2
12922	61	Round Head	2
12923	68	Round Head	2
12924	75	Round Head	2
12925	640	Round Head	2
12926	40	Pan Head	3
12927	47	Pan Head	3
12928	54	Pan Head	3
12929	61	Pan Head	3
12930	68	Pan Head	3
12931	75	Pan Head	3
12932	640	Pan Head	3
12933	40	Hex Head	4
12934	47	Hex Head	4
12935	54	Hex Head	4
12936	61	Hex Head	4
12937	68	Hex Head	4
12938	75	Hex Head	4
12939	640	Hex Head	4
12940	40	Truss Head	5
12941	47	Truss Head	5
12942	54	Truss Head	5
12943	61	Truss Head	5
12944	68	Truss Head	5
12945	75	Truss Head	5
12946	640	Truss Head	5
12947	40	Button Head	6
12948	47	Button Head	6
12949	54	Button Head	6
12950	61	Button Head	6
12951	68	Button Head	6
12952	75	Button Head	6
12953	640	Button Head	6
12954	40	Countersunk	7
12955	47	Countersunk	7
12956	54	Countersunk	7
12957	61	Countersunk	7
12958	68	Countersunk	7
12959	75	Countersunk	7
12960	640	Countersunk	7
12961	46	Zinc Plated	1
12962	46	Stainless Steel	2
12963	46	Black Oxide	3
12964	46	Brass	4
12965	46	Galvanized	5
12966	46	Chrome Plated	6
12967	46	Plain	7
12968	46	Nickel Plated	8
12969	53	Zinc Plated	1
12970	53	Stainless Steel	2
12971	53	Black Oxide	3
12972	53	Brass	4
12973	53	Galvanized	5
12974	53	Chrome Plated	6
12975	53	Plain	7
12976	53	Nickel Plated	8
12977	60	Zinc Plated	1
12978	60	Stainless Steel	2
12979	60	Black Oxide	3
12980	60	Brass	4
12981	60	Galvanized	5
12982	60	Chrome Plated	6
12983	60	Plain	7
12984	60	Nickel Plated	8
12985	67	Zinc Plated	1
12986	67	Stainless Steel	2
12987	67	Black Oxide	3
12988	67	Brass	4
12989	67	Galvanized	5
12990	67	Chrome Plated	6
12991	67	Plain	7
12992	67	Nickel Plated	8
12993	74	Zinc Plated	1
12994	74	Stainless Steel	2
12995	74	Black Oxide	3
12996	74	Brass	4
12997	74	Galvanized	5
12998	74	Chrome Plated	6
12999	74	Plain	7
13000	74	Nickel Plated	8
13001	80	Zinc Plated	1
13002	80	Stainless Steel	2
13003	80	Black Oxide	3
13004	80	Brass	4
13005	80	Galvanized	5
13006	80	Chrome Plated	6
13007	80	Plain	7
13008	80	Nickel Plated	8
13009	85	Zinc Plated	1
13010	85	Stainless Steel	2
13011	85	Black Oxide	3
13012	85	Brass	4
13013	85	Galvanized	5
13014	85	Chrome Plated	6
13015	85	Plain	7
13016	85	Nickel Plated	8
13017	90	Zinc Plated	1
13018	90	Stainless Steel	2
13019	90	Black Oxide	3
13020	90	Brass	4
13021	90	Galvanized	5
13022	90	Chrome Plated	6
13023	90	Plain	7
13024	90	Nickel Plated	8
13025	94	Zinc Plated	1
13026	94	Stainless Steel	2
13027	94	Black Oxide	3
13028	94	Brass	4
13029	94	Galvanized	5
13030	94	Chrome Plated	6
13031	94	Plain	7
13032	94	Nickel Plated	8
13033	133	Zinc Plated	1
13034	133	Stainless Steel	2
13035	133	Black Oxide	3
13036	133	Brass	4
13037	133	Galvanized	5
13038	133	Chrome Plated	6
13039	133	Plain	7
13040	133	Nickel Plated	8
13041	756	Zinc Plated	1
13042	756	Stainless Steel	2
13043	756	Black Oxide	3
13044	756	Brass	4
13045	756	Galvanized	5
13046	756	Chrome Plated	6
13047	756	Plain	7
13048	756	Nickel Plated	8
13049	760	Zinc Plated	1
13050	760	Stainless Steel	2
13051	760	Black Oxide	3
13052	760	Brass	4
13053	760	Galvanized	5
13054	760	Chrome Plated	6
13055	760	Plain	7
13056	760	Nickel Plated	8
13057	764	Zinc Plated	1
13058	764	Stainless Steel	2
13059	764	Black Oxide	3
13060	764	Brass	4
13061	764	Galvanized	5
13062	764	Chrome Plated	6
13063	764	Plain	7
13064	764	Nickel Plated	8
13065	769	Zinc Plated	1
13066	769	Stainless Steel	2
13067	769	Black Oxide	3
13068	769	Brass	4
13069	769	Galvanized	5
13070	769	Chrome Plated	6
13071	769	Plain	7
13072	769	Nickel Plated	8
13073	773	Zinc Plated	1
13074	773	Stainless Steel	2
13075	773	Black Oxide	3
13076	773	Brass	4
13077	773	Galvanized	5
13078	773	Chrome Plated	6
13079	773	Plain	7
13080	773	Nickel Plated	8
13081	919	Zinc Plated	1
13082	919	Stainless Steel	2
13083	919	Black Oxide	3
13084	919	Brass	4
13085	919	Galvanized	5
13086	919	Chrome Plated	6
13087	919	Plain	7
13088	919	Nickel Plated	8
13089	969	Zinc Plated	1
13090	969	Stainless Steel	2
13091	969	Black Oxide	3
13092	969	Brass	4
13093	969	Galvanized	5
13094	969	Chrome Plated	6
13095	969	Plain	7
13096	969	Nickel Plated	8
13097	974	Zinc Plated	1
13098	974	Stainless Steel	2
13099	974	Black Oxide	3
13100	974	Brass	4
13101	974	Galvanized	5
13102	974	Chrome Plated	6
13103	974	Plain	7
13104	974	Nickel Plated	8
13105	989	Zinc Plated	1
13106	989	Stainless Steel	2
13107	989	Black Oxide	3
13108	989	Brass	4
13109	989	Galvanized	5
13110	989	Chrome Plated	6
13111	989	Plain	7
13112	989	Nickel Plated	8
13113	1121	Zinc Plated	1
13114	1121	Stainless Steel	2
13115	1121	Black Oxide	3
13116	1121	Brass	4
13117	1121	Galvanized	5
13118	1121	Chrome Plated	6
13119	1121	Plain	7
13120	1121	Nickel Plated	8
13121	1131	Zinc Plated	1
13122	1131	Stainless Steel	2
13123	1131	Black Oxide	3
13124	1131	Brass	4
13125	1131	Galvanized	5
13126	1131	Chrome Plated	6
13127	1131	Plain	7
13128	1131	Nickel Plated	8
13129	939	Keyed Lock	1
13130	939	Combination Lock	2
13131	939	Smart Lock	3
13132	939	Biometric Lock	4
13133	939	Padlock	5
13134	939	Deadbolt	6
13135	939	Electronic Lock	7
13136	1490	Keyed Lock	1
13137	1490	Combination Lock	2
13138	1490	Smart Lock	3
13139	1490	Biometric Lock	4
13140	1490	Padlock	5
13141	1490	Deadbolt	6
13142	1490	Electronic Lock	7
13143	2014	Battery Powered	1
13144	2014	Rechargeable	2
13145	2014	Corded Electric	3
13146	2014	Cordless	4
13147	2014	Manual	5
13148	2014	Solar Powered	6
13149	2014	USB Powered	7
13150	1998	Blade Guard	1
13151	1998	Safety Switch	2
13152	1998	Lock-Off Button	3
13153	1998	Dual Switch	4
13154	1998	Thermal Overload	5
13155	1998	Automatic Shutoff	6
13156	2018	Basic Functions	1
13157	2018	Scientific	2
13158	2018	Financial	3
13159	2018	Graphing	4
13160	2018	Programmable	5
13161	2018	Statistical	6
13162	1501	LCD	1
13163	1501	LED	2
13164	1501	Digital	3
13165	1501	Analog	4
13166	1501	Touch Screen	5
13167	1501	E-Ink	6
13168	1501	OLED	7
13169	1705	Lavender	1
13170	1705	Rose	2
13171	1705	Jasmine	3
13172	1705	Sandalwood	4
13173	1705	Citrus	5
13174	1705	Ocean Breeze	6
13175	1705	Vanilla	7
13176	1705	Fresh Linen	8
13177	1705	Unscented	9
13178	96	5mm	1
13179	101	5mm	1
13180	106	5mm	1
13181	111	5mm	1
13182	505	5mm	1
13183	510	5mm	1
13184	554	5mm	1
13185	558	5mm	1
13186	562	5mm	1
13187	566	5mm	1
13188	570	5mm	1
13189	574	5mm	1
13190	614	5mm	1
13191	801	5mm	1
13192	805	5mm	1
13193	809	5mm	1
13194	96	6mm	2
13195	101	6mm	2
13196	106	6mm	2
13197	111	6mm	2
13198	505	6mm	2
13199	510	6mm	2
13200	554	6mm	2
13201	558	6mm	2
13202	562	6mm	2
13203	566	6mm	2
13204	570	6mm	2
13205	574	6mm	2
13206	614	6mm	2
13207	801	6mm	2
13208	805	6mm	2
13209	809	6mm	2
13210	96	8mm	3
13211	101	8mm	3
13212	106	8mm	3
13213	111	8mm	3
13214	505	8mm	3
13215	510	8mm	3
13216	554	8mm	3
13217	558	8mm	3
13218	562	8mm	3
13219	566	8mm	3
13220	570	8mm	3
13221	574	8mm	3
13222	614	8mm	3
13223	801	8mm	3
13224	805	8mm	3
13225	809	8mm	3
13226	96	10mm	4
13227	101	10mm	4
13228	106	10mm	4
13229	111	10mm	4
13230	505	10mm	4
13231	510	10mm	4
13232	554	10mm	4
13233	558	10mm	4
13234	562	10mm	4
13235	566	10mm	4
13236	570	10mm	4
13237	574	10mm	4
13238	614	10mm	4
13239	801	10mm	4
13240	805	10mm	4
13241	809	10mm	4
13242	96	12mm	5
13243	101	12mm	5
13244	106	12mm	5
13245	111	12mm	5
13246	505	12mm	5
13247	510	12mm	5
13248	554	12mm	5
13249	558	12mm	5
13250	562	12mm	5
13251	566	12mm	5
13252	570	12mm	5
13253	574	12mm	5
13254	614	12mm	5
13255	801	12mm	5
13256	805	12mm	5
13257	809	12mm	5
13258	96	16mm	6
13259	101	16mm	6
13260	106	16mm	6
13261	111	16mm	6
13262	505	16mm	6
13263	510	16mm	6
13264	554	16mm	6
13265	558	16mm	6
13266	562	16mm	6
13267	566	16mm	6
13268	570	16mm	6
13269	574	16mm	6
13270	614	16mm	6
13271	801	16mm	6
13272	805	16mm	6
13273	809	16mm	6
13274	96	20mm	7
13275	101	20mm	7
13276	106	20mm	7
13277	111	20mm	7
13278	505	20mm	7
13279	510	20mm	7
13280	554	20mm	7
13281	558	20mm	7
13282	562	20mm	7
13283	566	20mm	7
13284	570	20mm	7
13285	574	20mm	7
13286	614	20mm	7
13287	801	20mm	7
13288	805	20mm	7
13289	809	20mm	7
13290	96	25mm	8
13291	101	25mm	8
13292	106	25mm	8
13293	111	25mm	8
13294	505	25mm	8
13295	510	25mm	8
13296	554	25mm	8
13297	558	25mm	8
13298	562	25mm	8
13299	566	25mm	8
13300	570	25mm	8
13301	574	25mm	8
13302	614	25mm	8
13303	801	25mm	8
13304	805	25mm	8
13305	809	25mm	8
13306	96	32mm	9
13307	101	32mm	9
13308	106	32mm	9
13309	111	32mm	9
13310	505	32mm	9
13311	510	32mm	9
13312	554	32mm	9
13313	558	32mm	9
13314	562	32mm	9
13315	566	32mm	9
13316	570	32mm	9
13317	574	32mm	9
13318	614	32mm	9
13319	801	32mm	9
13320	805	32mm	9
13321	809	32mm	9
13322	96	40mm	10
13323	101	40mm	10
13324	106	40mm	10
13325	111	40mm	10
13326	505	40mm	10
13327	510	40mm	10
13328	554	40mm	10
13329	558	40mm	10
13330	562	40mm	10
13331	566	40mm	10
13332	570	40mm	10
13333	574	40mm	10
13334	614	40mm	10
13335	801	40mm	10
13336	805	40mm	10
13337	809	40mm	10
13338	194	100ml	1
13339	199	100ml	1
13340	204	100ml	1
13341	209	100ml	1
13342	214	100ml	1
13343	218	100ml	1
13344	222	100ml	1
13345	230	100ml	1
13346	238	100ml	1
13347	243	100ml	1
13348	248	100ml	1
13349	253	100ml	1
13350	258	100ml	1
13351	262	100ml	1
13352	266	100ml	1
13353	274	100ml	1
13354	194	250ml	2
13355	199	250ml	2
13356	204	250ml	2
13357	209	250ml	2
13358	214	250ml	2
13359	218	250ml	2
13360	222	250ml	2
13361	230	250ml	2
13362	238	250ml	2
13363	243	250ml	2
13364	248	250ml	2
13365	253	250ml	2
13366	258	250ml	2
13367	262	250ml	2
13368	266	250ml	2
13369	274	250ml	2
13370	194	500ml	3
13371	199	500ml	3
13372	204	500ml	3
13373	209	500ml	3
13374	214	500ml	3
13375	218	500ml	3
13376	222	500ml	3
13377	230	500ml	3
13378	238	500ml	3
13379	243	500ml	3
13380	248	500ml	3
13381	253	500ml	3
13382	258	500ml	3
13383	262	500ml	3
13384	266	500ml	3
13385	274	500ml	3
13386	194	750ml	4
13387	199	750ml	4
13388	204	750ml	4
13389	209	750ml	4
13390	214	750ml	4
13391	218	750ml	4
13392	222	750ml	4
13393	230	750ml	4
13394	238	750ml	4
13395	243	750ml	4
13396	248	750ml	4
13397	253	750ml	4
13398	258	750ml	4
13399	262	750ml	4
13400	266	750ml	4
13401	274	750ml	4
13402	194	1L	5
13403	199	1L	5
13404	204	1L	5
13405	209	1L	5
13406	214	1L	5
13407	218	1L	5
13408	222	1L	5
13409	230	1L	5
13410	238	1L	5
13411	243	1L	5
13412	248	1L	5
13413	253	1L	5
13414	258	1L	5
13415	262	1L	5
13416	266	1L	5
13417	274	1L	5
13418	194	1.5L	6
13419	199	1.5L	6
13420	204	1.5L	6
13421	209	1.5L	6
13422	214	1.5L	6
13423	218	1.5L	6
13424	222	1.5L	6
13425	230	1.5L	6
13426	238	1.5L	6
13427	243	1.5L	6
13428	248	1.5L	6
13429	253	1.5L	6
13430	258	1.5L	6
13431	262	1.5L	6
13432	266	1.5L	6
13433	274	1.5L	6
13434	194	2L	7
13435	199	2L	7
13436	204	2L	7
13437	209	2L	7
13438	214	2L	7
13439	218	2L	7
13440	222	2L	7
13441	230	2L	7
13442	238	2L	7
13443	243	2L	7
13444	248	2L	7
13445	253	2L	7
13446	258	2L	7
13447	262	2L	7
13448	266	2L	7
13449	274	2L	7
13450	194	5L	8
13451	199	5L	8
13452	204	5L	8
13453	209	5L	8
13454	214	5L	8
13455	218	5L	8
13456	222	5L	8
13457	230	5L	8
13458	238	5L	8
13459	243	5L	8
13460	248	5L	8
13461	253	5L	8
13462	258	5L	8
13463	262	5L	8
13464	266	5L	8
13465	274	5L	8
13466	194	10L	9
13467	199	10L	9
13468	204	10L	9
13469	209	10L	9
13470	214	10L	9
13471	218	10L	9
13472	222	10L	9
13473	230	10L	9
13474	238	10L	9
13475	243	10L	9
13476	248	10L	9
13477	253	10L	9
13478	258	10L	9
13479	262	10L	9
13480	266	10L	9
13481	274	10L	9
13482	194	20L	10
13483	199	20L	10
13484	204	20L	10
13485	209	20L	10
13486	214	20L	10
13487	218	20L	10
13488	222	20L	10
13489	230	20L	10
13490	238	20L	10
13491	243	20L	10
13492	248	20L	10
13493	253	20L	10
13494	258	20L	10
13495	262	20L	10
13496	266	20L	10
13497	274	20L	10
13498	542	100g	1
13499	545	100g	1
13500	659	100g	1
13501	662	100g	1
13502	665	100g	1
13503	673	100g	1
13504	676	100g	1
13505	679	100g	1
13506	682	100g	1
13507	685	100g	1
13508	745	100g	1
13509	749	100g	1
13510	753	100g	1
13511	827	100g	1
13512	1829	100g	1
13513	1833	100g	1
13514	1846	100g	1
13515	1861	100g	1
13516	1865	100g	1
13517	1869	100g	1
13518	1873	100g	1
13519	1877	100g	1
13520	1882	100g	1
13521	1886	100g	1
13522	1890	100g	1
13523	1894	100g	1
13524	1898	100g	1
13525	542	200g	2
13526	545	200g	2
13527	659	200g	2
13528	662	200g	2
13529	665	200g	2
13530	673	200g	2
13531	676	200g	2
13532	679	200g	2
13533	682	200g	2
13534	685	200g	2
13535	745	200g	2
13536	749	200g	2
13537	753	200g	2
13538	827	200g	2
13539	1829	200g	2
13540	1833	200g	2
13541	1846	200g	2
13542	1861	200g	2
13543	1865	200g	2
13544	1869	200g	2
13545	1873	200g	2
13546	1877	200g	2
13547	1882	200g	2
13548	1886	200g	2
13549	1890	200g	2
13550	1894	200g	2
13551	1898	200g	2
13552	542	250g	3
13553	545	250g	3
13554	659	250g	3
13555	662	250g	3
13556	665	250g	3
13557	673	250g	3
13558	676	250g	3
13559	679	250g	3
13560	682	250g	3
13561	685	250g	3
13562	745	250g	3
13563	749	250g	3
13564	753	250g	3
13565	827	250g	3
13566	1829	250g	3
13567	1833	250g	3
13568	1846	250g	3
13569	1861	250g	3
13570	1865	250g	3
13571	1869	250g	3
13572	1873	250g	3
13573	1877	250g	3
13574	1882	250g	3
13575	1886	250g	3
13576	1890	250g	3
13577	1894	250g	3
13578	1898	250g	3
13579	542	500g	4
13580	545	500g	4
13581	659	500g	4
13582	662	500g	4
13583	665	500g	4
13584	673	500g	4
13585	676	500g	4
13586	679	500g	4
13587	682	500g	4
13588	685	500g	4
13589	745	500g	4
13590	749	500g	4
13591	753	500g	4
13592	827	500g	4
13593	1829	500g	4
13594	1833	500g	4
13595	1846	500g	4
13596	1861	500g	4
13597	1865	500g	4
13598	1869	500g	4
13599	1873	500g	4
13600	1877	500g	4
13601	1882	500g	4
13602	1886	500g	4
13603	1890	500g	4
13604	1894	500g	4
13605	1898	500g	4
13606	542	1kg	5
13607	545	1kg	5
13608	659	1kg	5
13609	662	1kg	5
13610	665	1kg	5
13611	673	1kg	5
13612	676	1kg	5
13613	679	1kg	5
13614	682	1kg	5
13615	685	1kg	5
13616	745	1kg	5
13617	749	1kg	5
13618	753	1kg	5
13619	827	1kg	5
13620	1829	1kg	5
13621	1833	1kg	5
13622	1846	1kg	5
13623	1861	1kg	5
13624	1865	1kg	5
13625	1869	1kg	5
13626	1873	1kg	5
13627	1877	1kg	5
13628	1882	1kg	5
13629	1886	1kg	5
13630	1890	1kg	5
13631	1894	1kg	5
13632	1898	1kg	5
13633	542	2kg	6
13634	545	2kg	6
13635	659	2kg	6
13636	662	2kg	6
13637	665	2kg	6
13638	673	2kg	6
13639	676	2kg	6
13640	679	2kg	6
13641	682	2kg	6
13642	685	2kg	6
13643	745	2kg	6
13644	749	2kg	6
13645	753	2kg	6
13646	827	2kg	6
13647	1829	2kg	6
13648	1833	2kg	6
13649	1846	2kg	6
13650	1861	2kg	6
13651	1865	2kg	6
13652	1869	2kg	6
13653	1873	2kg	6
13654	1877	2kg	6
13655	1882	2kg	6
13656	1886	2kg	6
13657	1890	2kg	6
13658	1894	2kg	6
13659	1898	2kg	6
13660	542	5kg	7
13661	545	5kg	7
13662	659	5kg	7
13663	662	5kg	7
13664	665	5kg	7
13665	673	5kg	7
13666	676	5kg	7
13667	679	5kg	7
13668	682	5kg	7
13669	685	5kg	7
13670	745	5kg	7
13671	749	5kg	7
13672	753	5kg	7
13673	827	5kg	7
13674	1829	5kg	7
13675	1833	5kg	7
13676	1846	5kg	7
13677	1861	5kg	7
13678	1865	5kg	7
13679	1869	5kg	7
13680	1873	5kg	7
13681	1877	5kg	7
13682	1882	5kg	7
13683	1886	5kg	7
13684	1890	5kg	7
13685	1894	5kg	7
13686	1898	5kg	7
13687	542	10kg	8
13688	545	10kg	8
13689	659	10kg	8
13690	662	10kg	8
13691	665	10kg	8
13692	673	10kg	8
13693	676	10kg	8
13694	679	10kg	8
13695	682	10kg	8
13696	685	10kg	8
13697	745	10kg	8
13698	749	10kg	8
13699	753	10kg	8
13700	827	10kg	8
13701	1829	10kg	8
13702	1833	10kg	8
13703	1846	10kg	8
13704	1861	10kg	8
13705	1865	10kg	8
13706	1869	10kg	8
13707	1873	10kg	8
13708	1877	10kg	8
13709	1882	10kg	8
13710	1886	10kg	8
13711	1890	10kg	8
13712	1894	10kg	8
13713	1898	10kg	8
13714	542	25kg	9
13715	545	25kg	9
13716	659	25kg	9
13717	662	25kg	9
13718	665	25kg	9
13719	673	25kg	9
13720	676	25kg	9
13721	679	25kg	9
13722	682	25kg	9
13723	685	25kg	9
13724	745	25kg	9
13725	749	25kg	9
13726	753	25kg	9
13727	827	25kg	9
13728	1829	25kg	9
13729	1833	25kg	9
13730	1846	25kg	9
13731	1861	25kg	9
13732	1865	25kg	9
13733	1869	25kg	9
13734	1873	25kg	9
13735	1877	25kg	9
13736	1882	25kg	9
13737	1886	25kg	9
13738	1890	25kg	9
13739	1894	25kg	9
13740	1898	25kg	9
13741	542	50kg	10
13742	545	50kg	10
13743	659	50kg	10
13744	662	50kg	10
13745	665	50kg	10
13746	673	50kg	10
13747	676	50kg	10
13748	679	50kg	10
13749	682	50kg	10
13750	685	50kg	10
13751	745	50kg	10
13752	749	50kg	10
13753	753	50kg	10
13754	827	50kg	10
13755	1829	50kg	10
13756	1833	50kg	10
13757	1846	50kg	10
13758	1861	50kg	10
13759	1865	50kg	10
13760	1869	50kg	10
13761	1873	50kg	10
13762	1877	50kg	10
13763	1882	50kg	10
13764	1886	50kg	10
13765	1890	50kg	10
13766	1894	50kg	10
13767	1898	50kg	10
13768	591	10cm	1
13769	596	10cm	1
13770	601	10cm	1
13771	619	10cm	1
13772	625	10cm	1
13773	647	10cm	1
13774	797	10cm	1
13775	2049	10cm	1
13776	2053	10cm	1
13777	591	15cm	2
13778	596	15cm	2
13779	601	15cm	2
13780	619	15cm	2
13781	625	15cm	2
13782	647	15cm	2
13783	797	15cm	2
13784	2049	15cm	2
13785	2053	15cm	2
13786	591	20cm	3
13787	596	20cm	3
13788	601	20cm	3
13789	619	20cm	3
13790	625	20cm	3
13791	647	20cm	3
13792	797	20cm	3
13793	2049	20cm	3
13794	2053	20cm	3
13795	591	25cm	4
13796	596	25cm	4
13797	601	25cm	4
13798	619	25cm	4
13799	625	25cm	4
13800	647	25cm	4
13801	797	25cm	4
13802	2049	25cm	4
13803	2053	25cm	4
13804	591	30cm	5
13805	596	30cm	5
13806	601	30cm	5
13807	619	30cm	5
13808	625	30cm	5
13809	647	30cm	5
13810	797	30cm	5
13811	2049	30cm	5
13812	2053	30cm	5
13813	591	40cm	6
13814	596	40cm	6
13815	601	40cm	6
13816	619	40cm	6
13817	625	40cm	6
13818	647	40cm	6
13819	797	40cm	6
13820	2049	40cm	6
13821	2053	40cm	6
13822	591	50cm	7
13823	596	50cm	7
13824	601	50cm	7
13825	619	50cm	7
13826	625	50cm	7
13827	647	50cm	7
13828	797	50cm	7
13829	2049	50cm	7
13830	2053	50cm	7
13831	591	60cm	8
13832	596	60cm	8
13833	601	60cm	8
13834	619	60cm	8
13835	625	60cm	8
13836	647	60cm	8
13837	797	60cm	8
13838	2049	60cm	8
13839	2053	60cm	8
13840	591	80cm	9
13841	596	80cm	9
13842	601	80cm	9
13843	619	80cm	9
13844	625	80cm	9
13845	647	80cm	9
13846	797	80cm	9
13847	2049	80cm	9
13848	2053	80cm	9
13849	591	100cm	10
13850	596	100cm	10
13851	601	100cm	10
13852	619	100cm	10
13853	625	100cm	10
13854	647	100cm	10
13855	797	100cm	10
13856	2049	100cm	10
13857	2053	100cm	10
13858	329	Plastic	1
13859	334	Plastic	1
13860	348	Plastic	1
13861	449	Plastic	1
13862	453	Plastic	1
13863	484	Plastic	1
13864	493	Plastic	1
13865	329	Rubber	2
13866	334	Rubber	2
13867	348	Rubber	2
13868	449	Rubber	2
13869	453	Rubber	2
13870	484	Rubber	2
13871	493	Rubber	2
13872	329	Wood	3
13873	334	Wood	3
13874	348	Wood	3
13875	449	Wood	3
13876	453	Wood	3
13877	484	Wood	3
13878	493	Wood	3
13879	329	Metal	4
13880	334	Metal	4
13881	348	Metal	4
13882	449	Metal	4
13883	453	Metal	4
13884	484	Metal	4
13885	493	Metal	4
13886	329	Leather	5
13887	334	Leather	5
13888	348	Leather	5
13889	449	Leather	5
13890	453	Leather	5
13891	484	Leather	5
13892	493	Leather	5
13893	329	Silicone	6
13894	334	Silicone	6
13895	348	Silicone	6
13896	449	Silicone	6
13897	453	Silicone	6
13898	484	Silicone	6
13899	493	Silicone	6
13900	329	Foam	7
13901	334	Foam	7
13902	348	Foam	7
13903	449	Foam	7
13904	453	Foam	7
13905	484	Foam	7
13906	493	Foam	7
13907	281	AC	1
13908	287	AC	1
13909	293	AC	1
13910	299	AC	1
13911	305	AC	1
13912	311	AC	1
13913	317	AC	1
13914	281	DC	2
13915	287	DC	2
13916	293	DC	2
13917	299	DC	2
13918	305	DC	2
13919	311	DC	2
13920	317	DC	2
13921	281	Battery	3
13922	287	Battery	3
13923	293	Battery	3
13924	299	Battery	3
13925	305	Battery	3
13926	311	Battery	3
13927	317	Battery	3
13928	281	Rechargeable	4
13929	287	Rechargeable	4
13930	293	Rechargeable	4
13931	299	Rechargeable	4
13932	305	Rechargeable	4
13933	311	Rechargeable	4
13934	317	Rechargeable	4
13935	281	Solar	5
13936	287	Solar	5
13937	293	Solar	5
13938	299	Solar	5
13939	305	Solar	5
13940	311	Solar	5
13941	317	Solar	5
13942	281	Dual Power	6
13943	287	Dual Power	6
13944	293	Dual Power	6
13945	299	Dual Power	6
13946	305	Dual Power	6
13947	311	Dual Power	6
13948	317	Dual Power	6
13949	138	Wall Mount	1
13950	377	Wall Mount	1
13951	865	Wall Mount	1
13952	1157	Wall Mount	1
13953	1167	Wall Mount	1
13954	1352	Wall Mount	1
13955	2010	Wall Mount	1
13956	138	Ceiling Mount	2
13957	377	Ceiling Mount	2
13958	865	Ceiling Mount	2
13959	1157	Ceiling Mount	2
13960	1167	Ceiling Mount	2
13961	1352	Ceiling Mount	2
13962	2010	Ceiling Mount	2
13963	138	Floor Stand	3
13964	377	Floor Stand	3
13965	865	Floor Stand	3
13966	1157	Floor Stand	3
13967	1167	Floor Stand	3
13968	1352	Floor Stand	3
13969	2010	Floor Stand	3
13970	138	Table Mount	4
13971	377	Table Mount	4
13972	865	Table Mount	4
13973	1157	Table Mount	4
13974	1167	Table Mount	4
13975	1352	Table Mount	4
13976	2010	Table Mount	4
13977	138	Pole Mount	5
13978	377	Pole Mount	5
13979	865	Pole Mount	5
13980	1157	Pole Mount	5
13981	1167	Pole Mount	5
13982	1352	Pole Mount	5
13983	2010	Pole Mount	5
13984	138	Flush Mount	6
13985	377	Flush Mount	6
13986	865	Flush Mount	6
13987	1157	Flush Mount	6
13988	1167	Flush Mount	6
13989	1352	Flush Mount	6
13990	2010	Flush Mount	6
13991	138	Recessed	7
13992	377	Recessed	7
13993	865	Recessed	7
13994	1157	Recessed	7
13995	1167	Recessed	7
13996	1352	Recessed	7
13997	2010	Recessed	7
13998	1293	Monocrystalline	1
13999	1299	Monocrystalline	1
14000	1305	Monocrystalline	1
14001	1310	Monocrystalline	1
14002	1315	Monocrystalline	1
14003	1320	Monocrystalline	1
14004	1326	Monocrystalline	1
14005	1293	Polycrystalline	2
14006	1299	Polycrystalline	2
14007	1305	Polycrystalline	2
14008	1310	Polycrystalline	2
14009	1315	Polycrystalline	2
14010	1320	Polycrystalline	2
14011	1326	Polycrystalline	2
14012	1293	Thin Film	3
14013	1299	Thin Film	3
14014	1305	Thin Film	3
14015	1310	Thin Film	3
14016	1315	Thin Film	3
14017	1320	Thin Film	3
14018	1326	Thin Film	3
14019	1293	PERC	4
14020	1299	PERC	4
14021	1305	PERC	4
14022	1310	PERC	4
14023	1315	PERC	4
14024	1320	PERC	4
14025	1326	PERC	4
14026	1293	Bifacial	5
14027	1299	Bifacial	5
14028	1305	Bifacial	5
14029	1310	Bifacial	5
14030	1315	Bifacial	5
14031	1320	Bifacial	5
14032	1326	Bifacial	5
14033	141	5V	1
14034	146	5V	1
14035	151	5V	1
14036	167	5V	1
14037	172	5V	1
14038	177	5V	1
14039	183	5V	1
14040	141	12V	2
14041	146	12V	2
14042	151	12V	2
14043	167	12V	2
14044	172	12V	2
14045	177	12V	2
14046	183	12V	2
14047	141	24V	3
14048	146	24V	3
14049	151	24V	3
14050	167	24V	3
14051	172	24V	3
14052	177	24V	3
14053	183	24V	3
14054	141	48V	4
14055	146	48V	4
14056	151	48V	4
14057	167	48V	4
14058	172	48V	4
14059	177	48V	4
14060	183	48V	4
14061	141	110V	5
14062	146	110V	5
14063	151	110V	5
14064	167	110V	5
14065	172	110V	5
14066	177	110V	5
14067	183	110V	5
14068	141	220V	6
14069	146	220V	6
14070	151	220V	6
14071	167	220V	6
14072	172	220V	6
14073	177	220V	6
14074	183	220V	6
14075	141	240V	7
14076	146	240V	7
14077	151	240V	7
14078	167	240V	7
14079	172	240V	7
14080	177	240V	7
14081	183	240V	7
14082	141	Multi-Voltage	8
14083	146	Multi-Voltage	8
14084	151	Multi-Voltage	8
14085	167	Multi-Voltage	8
14086	172	Multi-Voltage	8
14087	177	Multi-Voltage	8
14088	183	Multi-Voltage	8
14089	140	5W	1
14090	145	5W	1
14091	150	5W	1
14092	176	5W	1
14093	182	5W	1
14094	1476	5W	1
14095	140	10W	2
14096	145	10W	2
14097	150	10W	2
14098	176	10W	2
14099	182	10W	2
14100	1476	10W	2
14101	140	15W	3
14102	145	15W	3
14103	150	15W	3
14104	176	15W	3
14105	182	15W	3
14106	1476	15W	3
14107	140	20W	4
14108	145	20W	4
14109	150	20W	4
14110	176	20W	4
14111	182	20W	4
14112	1476	20W	4
14113	140	40W	5
14114	145	40W	5
14115	150	40W	5
14116	176	40W	5
14117	182	40W	5
14118	1476	40W	5
14119	140	60W	6
14120	145	60W	6
14121	150	60W	6
14122	176	60W	6
14123	182	60W	6
14124	1476	60W	6
14125	140	100W	7
14126	145	100W	7
14127	150	100W	7
14128	176	100W	7
14129	182	100W	7
14130	1476	100W	7
14131	140	150W	8
14132	145	150W	8
14133	150	150W	8
14134	176	150W	8
14135	182	150W	8
14136	1476	150W	8
14137	140	200W	9
14138	145	200W	9
14139	150	200W	9
14140	176	200W	9
14141	182	200W	9
14142	1476	200W	9
14143	142	E27	1
14144	147	E27	1
14145	195	E27	1
14146	200	E27	1
14147	239	E27	1
14148	244	E27	1
14149	142	E14	2
14150	147	E14	2
14151	195	E14	2
14152	200	E14	2
14153	239	E14	2
14154	244	E14	2
14155	142	B22	3
14156	147	B22	3
14157	195	B22	3
14158	200	B22	3
14159	239	B22	3
14160	244	B22	3
14161	142	GU10	4
14162	147	GU10	4
14163	195	GU10	4
14164	200	GU10	4
14165	239	GU10	4
14166	244	GU10	4
14167	142	G9	5
14168	147	G9	5
14169	195	G9	5
14170	200	G9	5
14171	239	G9	5
14172	244	G9	5
14173	142	MR16	6
14174	147	MR16	6
14175	195	MR16	6
14176	200	MR16	6
14177	239	MR16	6
14178	244	MR16	6
14179	389	10cm	1
14180	394	10cm	1
14181	399	10cm	1
14182	409	10cm	1
14183	646	10cm	1
14184	389	20cm	2
14185	394	20cm	2
14186	399	20cm	2
14187	409	20cm	2
14188	646	20cm	2
14189	389	30cm	3
14190	394	30cm	3
14191	399	30cm	3
14192	409	30cm	3
14193	646	30cm	3
14194	389	50cm	4
14195	394	50cm	4
14196	399	50cm	4
14197	409	50cm	4
14198	646	50cm	4
14199	389	75cm	5
14200	394	75cm	5
14201	399	75cm	5
14202	409	75cm	5
14203	646	75cm	5
14204	389	100cm	6
14205	394	100cm	6
14206	399	100cm	6
14207	409	100cm	6
14208	646	100cm	6
14209	389	150cm	7
14210	394	150cm	7
14211	399	150cm	7
14212	409	150cm	7
14213	646	150cm	7
14214	389	200cm	8
14215	394	200cm	8
14216	399	200cm	8
14217	409	200cm	8
14218	646	200cm	8
14219	193	Glossy	1
14220	198	Glossy	1
14221	237	Glossy	1
14222	242	Glossy	1
14223	737	Glossy	1
14224	193	Matte	2
14225	198	Matte	2
14226	237	Matte	2
14227	242	Matte	2
14228	737	Matte	2
14229	193	Satin	3
14230	198	Satin	3
14231	237	Satin	3
14232	242	Satin	3
14233	737	Satin	3
14234	193	Metallic	4
14235	198	Metallic	4
14236	237	Metallic	4
14237	242	Metallic	4
14238	737	Metallic	4
14239	193	Brushed	5
14240	198	Brushed	5
14241	237	Brushed	5
14242	242	Brushed	5
14243	737	Brushed	5
14244	193	Polished	6
14245	198	Polished	6
14246	237	Polished	6
14247	242	Polished	6
14248	737	Polished	6
14249	193	Textured	7
14250	198	Textured	7
14251	237	Textured	7
14252	242	Textured	7
14253	737	Textured	7
14254	175	900mm (36")	1
14255	482	900mm (36")	1
14256	1595	900mm (36")	1
14257	1600	900mm (36")	1
14258	1605	900mm (36")	1
14259	175	1050mm (42")	2
14260	482	1050mm (42")	2
14261	1595	1050mm (42")	2
14262	1600	1050mm (42")	2
14263	1605	1050mm (42")	2
14264	175	1200mm (48")	3
14265	482	1200mm (48")	3
14266	1595	1200mm (48")	3
14267	1600	1200mm (48")	3
14268	1605	1200mm (48")	3
14269	175	1400mm (56")	4
14270	482	1400mm (56")	4
14271	1595	1400mm (56")	4
14272	1600	1400mm (56")	4
14273	1605	1400mm (56")	4
14274	175	1800mm (72")	5
14275	482	1800mm (72")	5
14276	1595	1800mm (72")	5
14277	1600	1800mm (72")	5
14278	1605	1800mm (72")	5
14279	99	Low Pressure	1
14280	104	Low Pressure	1
14281	109	Low Pressure	1
14282	124	Low Pressure	1
14283	128	Low Pressure	1
14284	99	Medium Pressure	2
14285	104	Medium Pressure	2
14286	109	Medium Pressure	2
14287	124	Medium Pressure	2
14288	128	Medium Pressure	2
14289	99	High Pressure	3
14290	104	High Pressure	3
14291	109	High Pressure	3
14292	124	High Pressure	3
14293	128	High Pressure	3
14294	99	10 Bar	4
14295	104	10 Bar	4
14296	109	10 Bar	4
14297	124	10 Bar	4
14298	128	10 Bar	4
14299	99	15 Bar	5
14300	104	15 Bar	5
14301	109	15 Bar	5
14302	124	15 Bar	5
14303	128	15 Bar	5
14304	99	20 Bar	6
14305	104	20 Bar	6
14306	109	20 Bar	6
14307	124	20 Bar	6
14308	128	20 Bar	6
14309	208	Indoor Only	1
14310	208	Weather Resistant	2
14311	208	Waterproof	3
14312	208	IP44	4
14313	208	IP55	5
14314	208	IP65	6
14315	208	IP67	7
14316	252	Indoor Only	1
14317	252	Weather Resistant	2
14318	252	Waterproof	3
14319	252	IP44	4
14320	252	IP55	5
14321	252	IP65	6
14322	252	IP67	7
14323	947	Indoor Only	1
14324	947	Weather Resistant	2
14325	947	Waterproof	3
14326	947	IP44	4
14327	947	IP55	5
14328	947	IP65	6
14329	947	IP67	7
14330	952	Indoor Only	1
14331	952	Weather Resistant	2
14332	952	Waterproof	3
14333	952	IP44	4
14334	952	IP55	5
14335	952	IP65	6
14336	952	IP67	7
14337	234	Short (10-20cm)	1
14338	278	Short (10-20cm)	1
14339	444	Short (10-20cm)	1
14340	483	Short (10-20cm)	1
14341	487	Short (10-20cm)	1
14342	496	Short (10-20cm)	1
14343	234	Medium (20-40cm)	2
14344	278	Medium (20-40cm)	2
14345	444	Medium (20-40cm)	2
14346	483	Medium (20-40cm)	2
14347	487	Medium (20-40cm)	2
14348	496	Medium (20-40cm)	2
14349	234	Long (40-60cm)	3
14350	278	Long (40-60cm)	3
14351	444	Long (40-60cm)	3
14352	483	Long (40-60cm)	3
14353	487	Long (40-60cm)	3
14354	496	Long (40-60cm)	3
14355	234	Extra Long (60cm+)	4
14356	278	Extra Long (60cm+)	4
14357	444	Extra Long (60cm+)	4
14358	483	Extra Long (60cm+)	4
14359	487	Extra Long (60cm+)	4
14360	496	Extra Long (60cm+)	4
14361	117	Small	1
14362	117	Medium	2
14363	117	Large	3
14364	117	Extra Large	4
14365	143	Small	1
14366	143	Medium	2
14367	143	Large	3
14368	143	Extra Large	4
14369	148	Small	1
14370	148	Medium	2
14371	148	Large	3
14372	148	Extra Large	4
14373	153	Small	1
14374	153	Medium	2
14375	153	Large	3
14376	153	Extra Large	4
14377	155	Small	1
14378	155	Medium	2
14379	155	Large	3
14380	155	Extra Large	4
14381	157	Small	1
14382	157	Medium	2
14383	157	Large	3
14384	157	Extra Large	4
14385	158	Small	1
14386	158	Medium	2
14387	158	Large	3
14388	158	Extra Large	4
14389	160	Small	1
14390	160	Medium	2
14391	160	Large	3
14392	160	Extra Large	4
14393	161	Small	1
14394	161	Medium	2
14395	161	Large	3
14396	161	Extra Large	4
14397	166	Small	1
14398	166	Medium	2
14399	166	Large	3
14400	166	Extra Large	4
14401	171	Small	1
14402	171	Medium	2
14403	171	Large	3
14404	171	Extra Large	4
14405	178	Small	1
14406	178	Medium	2
14407	178	Large	3
14408	178	Extra Large	4
14409	187	Small	1
14410	187	Medium	2
14411	187	Large	3
14412	187	Extra Large	4
14413	203	Small	1
14414	203	Medium	2
14415	203	Large	3
14416	203	Extra Large	4
14417	213	Small	1
14418	213	Medium	2
14419	213	Large	3
14420	213	Extra Large	4
14421	215	Small	1
14422	215	Medium	2
14423	215	Large	3
14424	215	Extra Large	4
14425	226	Small	1
14426	226	Medium	2
14427	226	Large	3
14428	226	Extra Large	4
14429	247	Small	1
14430	247	Medium	2
14431	247	Large	3
14432	247	Extra Large	4
14433	257	Small	1
14434	257	Medium	2
14435	257	Large	3
14436	257	Extra Large	4
14437	259	Small	1
14438	259	Medium	2
14439	259	Large	3
14440	259	Extra Large	4
14441	270	Small	1
14442	270	Medium	2
14443	270	Large	3
14444	270	Extra Large	4
14445	283	Small	1
14446	283	Medium	2
14447	283	Large	3
14448	283	Extra Large	4
14449	289	Small	1
14450	289	Medium	2
14451	289	Large	3
14452	289	Extra Large	4
14453	295	Small	1
14454	295	Medium	2
14455	295	Large	3
14456	295	Extra Large	4
14457	301	Small	1
14458	301	Medium	2
14459	301	Large	3
14460	301	Extra Large	4
14461	307	Small	1
14462	307	Medium	2
14463	307	Large	3
14464	307	Extra Large	4
14465	313	Small	1
14466	313	Medium	2
14467	313	Large	3
14468	313	Extra Large	4
14469	319	Small	1
14470	319	Medium	2
14471	319	Large	3
14472	319	Extra Large	4
14473	323	Small	1
14474	323	Medium	2
14475	323	Large	3
14476	323	Extra Large	4
14477	324	Small	1
14478	324	Medium	2
14479	324	Large	3
14480	324	Extra Large	4
14481	325	Small	1
14482	325	Medium	2
14483	325	Large	3
14484	325	Extra Large	4
14485	328	Small	1
14486	328	Medium	2
14487	328	Large	3
14488	328	Extra Large	4
14489	333	Small	1
14490	333	Medium	2
14491	333	Large	3
14492	333	Extra Large	4
14493	342	Small	1
14494	342	Medium	2
14495	342	Large	3
14496	342	Extra Large	4
14497	346	Small	1
14498	346	Medium	2
14499	346	Large	3
14500	346	Extra Large	4
14501	351	Small	1
14502	351	Medium	2
14503	351	Large	3
14504	351	Extra Large	4
14505	352	Small	1
14506	352	Medium	2
14507	352	Large	3
14508	352	Extra Large	4
14509	353	Small	1
14510	353	Medium	2
14511	353	Large	3
14512	353	Extra Large	4
14513	359	Small	1
14514	359	Medium	2
14515	359	Large	3
14516	359	Extra Large	4
14517	361	Small	1
14518	361	Medium	2
14519	361	Large	3
14520	361	Extra Large	4
14521	366	Small	1
14522	366	Medium	2
14523	366	Large	3
14524	366	Extra Large	4
14525	380	Small	1
14526	380	Medium	2
14527	380	Large	3
14528	380	Extra Large	4
14529	381	Small	1
14530	381	Medium	2
14531	381	Large	3
14532	381	Extra Large	4
14533	382	Small	1
14534	382	Medium	2
14535	382	Large	3
14536	382	Extra Large	4
14537	404	Small	1
14538	404	Medium	2
14539	404	Large	3
14540	404	Extra Large	4
14541	405	Small	1
14542	405	Medium	2
14543	405	Large	3
14544	405	Extra Large	4
14545	415	Small	1
14546	415	Medium	2
14547	415	Large	3
14548	415	Extra Large	4
14549	420	Small	1
14550	420	Medium	2
14551	420	Large	3
14552	420	Extra Large	4
14553	424	Small	1
14554	424	Medium	2
14555	424	Large	3
14556	424	Extra Large	4
14557	429	Small	1
14558	429	Medium	2
14559	429	Large	3
14560	429	Extra Large	4
14561	431	Small	1
14562	431	Medium	2
14563	431	Large	3
14564	431	Extra Large	4
14565	435	Small	1
14566	435	Medium	2
14567	435	Large	3
14568	435	Extra Large	4
14569	436	Small	1
14570	436	Medium	2
14571	436	Large	3
14572	436	Extra Large	4
14573	437	Small	1
14574	437	Medium	2
14575	437	Large	3
14576	437	Extra Large	4
14577	439	Small	1
14578	439	Medium	2
14579	439	Large	3
14580	439	Extra Large	4
14581	441	Small	1
14582	441	Medium	2
14583	441	Large	3
14584	441	Extra Large	4
14585	443	Small	1
14586	443	Medium	2
14587	443	Large	3
14588	443	Extra Large	4
14589	445	Small	1
14590	445	Medium	2
14591	445	Large	3
14592	445	Extra Large	4
14593	447	Small	1
14594	447	Medium	2
14595	447	Large	3
14596	447	Extra Large	4
14597	452	Small	1
14598	452	Medium	2
14599	452	Large	3
14600	452	Extra Large	4
14601	464	Small	1
14602	464	Medium	2
14603	464	Large	3
14604	464	Extra Large	4
14605	467	Small	1
14606	467	Medium	2
14607	467	Large	3
14608	467	Extra Large	4
14609	468	Small	1
14610	468	Medium	2
14611	468	Large	3
14612	468	Extra Large	4
14613	469	Small	1
14614	469	Medium	2
14615	469	Large	3
14616	469	Extra Large	4
14617	472	Small	1
14618	472	Medium	2
14619	472	Large	3
14620	472	Extra Large	4
14621	477	Small	1
14622	477	Medium	2
14623	477	Large	3
14624	477	Extra Large	4
14625	486	Small	1
14626	486	Medium	2
14627	486	Large	3
14628	486	Extra Large	4
14629	491	Small	1
14630	491	Medium	2
14631	491	Large	3
14632	491	Extra Large	4
14633	495	Small	1
14634	495	Medium	2
14635	495	Large	3
14636	495	Extra Large	4
14637	497	Small	1
14638	497	Medium	2
14639	497	Large	3
14640	497	Extra Large	4
14641	502	Small	1
14642	502	Medium	2
14643	502	Large	3
14644	502	Extra Large	4
14645	503	Small	1
14646	503	Medium	2
14647	503	Large	3
14648	503	Extra Large	4
14649	508	Small	1
14650	508	Medium	2
14651	508	Large	3
14652	508	Extra Large	4
14653	515	Small	1
14654	515	Medium	2
14655	515	Large	3
14656	515	Extra Large	4
14657	521	Small	1
14658	521	Medium	2
14659	521	Large	3
14660	521	Extra Large	4
14661	524	Small	1
14662	524	Medium	2
14663	524	Large	3
14664	524	Extra Large	4
14665	525	Small	1
14666	525	Medium	2
14667	525	Large	3
14668	525	Extra Large	4
14669	528	Small	1
14670	528	Medium	2
14671	528	Large	3
14672	528	Extra Large	4
14673	529	Small	1
14674	529	Medium	2
14675	529	Large	3
14676	529	Extra Large	4
14677	532	Small	1
14678	532	Medium	2
14679	532	Large	3
14680	532	Extra Large	4
14681	533	Small	1
14682	533	Medium	2
14683	533	Large	3
14684	533	Extra Large	4
14685	537	Small	1
14686	537	Medium	2
14687	537	Large	3
14688	537	Extra Large	4
14689	538	Small	1
14690	538	Medium	2
14691	538	Large	3
14692	538	Extra Large	4
14693	642	Small	1
14694	642	Medium	2
14695	642	Large	3
14696	642	Extra Large	4
14697	644	Small	1
14698	644	Medium	2
14699	644	Large	3
14700	644	Extra Large	4
14701	649	Small	1
14702	649	Medium	2
14703	649	Large	3
14704	649	Extra Large	4
14705	651	Small	1
14706	651	Medium	2
14707	651	Large	3
14708	651	Extra Large	4
14709	670	Small	1
14710	670	Medium	2
14711	670	Large	3
14712	670	Extra Large	4
14713	696	Small	1
14714	696	Medium	2
14715	696	Large	3
14716	696	Extra Large	4
14717	705	Small	1
14718	705	Medium	2
14719	705	Large	3
14720	705	Extra Large	4
14721	711	Small	1
14722	711	Medium	2
14723	711	Large	3
14724	711	Extra Large	4
14725	714	Small	1
14726	714	Medium	2
14727	714	Large	3
14728	714	Extra Large	4
14729	715	Small	1
14730	715	Medium	2
14731	715	Large	3
14732	715	Extra Large	4
14733	716	Small	1
14734	716	Medium	2
14735	716	Large	3
14736	716	Extra Large	4
14737	719	Small	1
14738	719	Medium	2
14739	719	Large	3
14740	719	Extra Large	4
14741	720	Small	1
14742	720	Medium	2
14743	720	Large	3
14744	720	Extra Large	4
14745	761	Small	1
14746	761	Medium	2
14747	761	Large	3
14748	761	Extra Large	4
14749	767	Small	1
14750	767	Medium	2
14751	767	Large	3
14752	767	Extra Large	4
14753	771	Small	1
14754	771	Medium	2
14755	771	Large	3
14756	771	Extra Large	4
14757	803	Small	1
14758	803	Medium	2
14759	803	Large	3
14760	803	Extra Large	4
14761	807	Small	1
14762	807	Medium	2
14763	807	Large	3
14764	807	Extra Large	4
14765	820	Small	1
14766	820	Medium	2
14767	820	Large	3
14768	820	Extra Large	4
14769	824	Small	1
14770	824	Medium	2
14771	824	Large	3
14772	824	Extra Large	4
14773	833	Small	1
14774	833	Medium	2
14775	833	Large	3
14776	833	Extra Large	4
14777	834	Small	1
14778	834	Medium	2
14779	834	Large	3
14780	834	Extra Large	4
14781	843	Small	1
14782	843	Medium	2
14783	843	Large	3
14784	843	Extra Large	4
14785	848	Small	1
14786	848	Medium	2
14787	848	Large	3
14788	848	Extra Large	4
14789	893	Small	1
14790	893	Medium	2
14791	893	Large	3
14792	893	Extra Large	4
14793	900	Small	1
14794	900	Medium	2
14795	900	Large	3
14796	900	Extra Large	4
14797	913	Small	1
14798	913	Medium	2
14799	913	Large	3
14800	913	Extra Large	4
14801	914	Small	1
14802	914	Medium	2
14803	914	Large	3
14804	914	Extra Large	4
14805	922	Small	1
14806	922	Medium	2
14807	922	Large	3
14808	922	Extra Large	4
14809	934	Small	1
14810	934	Medium	2
14811	934	Large	3
14812	934	Extra Large	4
14813	949	Small	1
14814	949	Medium	2
14815	949	Large	3
14816	949	Extra Large	4
14817	966	Small	1
14818	966	Medium	2
14819	966	Large	3
14820	966	Extra Large	4
14821	967	Small	1
14822	967	Medium	2
14823	967	Large	3
14824	967	Extra Large	4
14825	968	Small	1
14826	968	Medium	2
14827	968	Large	3
14828	968	Extra Large	4
14829	971	Small	1
14830	971	Medium	2
14831	971	Large	3
14832	971	Extra Large	4
14833	972	Small	1
14834	972	Medium	2
14835	972	Large	3
14836	972	Extra Large	4
14837	973	Small	1
14838	973	Medium	2
14839	973	Large	3
14840	973	Extra Large	4
14841	998	Small	1
14842	998	Medium	2
14843	998	Large	3
14844	998	Extra Large	4
14845	1028	Small	1
14846	1028	Medium	2
14847	1028	Large	3
14848	1028	Extra Large	4
14849	1031	Small	1
14850	1031	Medium	2
14851	1031	Large	3
14852	1031	Extra Large	4
14853	1108	Small	1
14854	1108	Medium	2
14855	1108	Large	3
14856	1108	Extra Large	4
14857	1140	Small	1
14858	1140	Medium	2
14859	1140	Large	3
14860	1140	Extra Large	4
14861	1145	Small	1
14862	1145	Medium	2
14863	1145	Large	3
14864	1145	Extra Large	4
14865	1162	Small	1
14866	1162	Medium	2
14867	1162	Large	3
14868	1162	Extra Large	4
14869	1298	Small	1
14870	1298	Medium	2
14871	1298	Large	3
14872	1298	Extra Large	4
14873	1329	Small	1
14874	1329	Medium	2
14875	1329	Large	3
14876	1329	Extra Large	4
14877	1332	Small	1
14878	1332	Medium	2
14879	1332	Large	3
14880	1332	Extra Large	4
14881	1334	Small	1
14882	1334	Medium	2
14883	1334	Large	3
14884	1334	Extra Large	4
14885	1337	Small	1
14886	1337	Medium	2
14887	1337	Large	3
14888	1337	Extra Large	4
14889	1342	Small	1
14890	1342	Medium	2
14891	1342	Large	3
14892	1342	Extra Large	4
14893	1344	Small	1
14894	1344	Medium	2
14895	1344	Large	3
14896	1344	Extra Large	4
14897	1347	Small	1
14898	1347	Medium	2
14899	1347	Large	3
14900	1347	Extra Large	4
14901	1349	Small	1
14902	1349	Medium	2
14903	1349	Large	3
14904	1349	Extra Large	4
14905	1351	Small	1
14906	1351	Medium	2
14907	1351	Large	3
14908	1351	Extra Large	4
14909	1354	Small	1
14910	1354	Medium	2
14911	1354	Large	3
14912	1354	Extra Large	4
14913	1357	Small	1
14914	1357	Medium	2
14915	1357	Large	3
14916	1357	Extra Large	4
14917	1361	Small	1
14918	1361	Medium	2
14919	1361	Large	3
14920	1361	Extra Large	4
14921	1362	Small	1
14922	1362	Medium	2
14923	1362	Large	3
14924	1362	Extra Large	4
14925	1366	Small	1
14926	1366	Medium	2
14927	1366	Large	3
14928	1366	Extra Large	4
14929	1369	Small	1
14930	1369	Medium	2
14931	1369	Large	3
14932	1369	Extra Large	4
14933	1371	Small	1
14934	1371	Medium	2
14935	1371	Large	3
14936	1371	Extra Large	4
14937	1372	Small	1
14938	1372	Medium	2
14939	1372	Large	3
14940	1372	Extra Large	4
14941	1377	Small	1
14942	1377	Medium	2
14943	1377	Large	3
14944	1377	Extra Large	4
14945	1380	Small	1
14946	1380	Medium	2
14947	1380	Large	3
14948	1380	Extra Large	4
14949	1382	Small	1
14950	1382	Medium	2
14951	1382	Large	3
14952	1382	Extra Large	4
14953	1385	Small	1
14954	1385	Medium	2
14955	1385	Large	3
14956	1385	Extra Large	4
14957	1394	Small	1
14958	1394	Medium	2
14959	1394	Large	3
14960	1394	Extra Large	4
14961	1395	Small	1
14962	1395	Medium	2
14963	1395	Large	3
14964	1395	Extra Large	4
14965	1396	Small	1
14966	1396	Medium	2
14967	1396	Large	3
14968	1396	Extra Large	4
14969	1399	Small	1
14970	1399	Medium	2
14971	1399	Large	3
14972	1399	Extra Large	4
14973	1400	Small	1
14974	1400	Medium	2
14975	1400	Large	3
14976	1400	Extra Large	4
14977	1401	Small	1
14978	1401	Medium	2
14979	1401	Large	3
14980	1401	Extra Large	4
14981	1402	Small	1
14982	1402	Medium	2
14983	1402	Large	3
14984	1402	Extra Large	4
14985	1404	Small	1
14986	1404	Medium	2
14987	1404	Large	3
14988	1404	Extra Large	4
14989	1405	Small	1
14990	1405	Medium	2
14991	1405	Large	3
14992	1405	Extra Large	4
14993	1406	Small	1
14994	1406	Medium	2
14995	1406	Large	3
14996	1406	Extra Large	4
14997	1407	Small	1
14998	1407	Medium	2
14999	1407	Large	3
15000	1407	Extra Large	4
15001	1409	Small	1
15002	1409	Medium	2
15003	1409	Large	3
15004	1409	Extra Large	4
15005	1411	Small	1
15006	1411	Medium	2
15007	1411	Large	3
15008	1411	Extra Large	4
15009	1412	Small	1
15010	1412	Medium	2
15011	1412	Large	3
15012	1412	Extra Large	4
15013	1422	Small	1
15014	1422	Medium	2
15015	1422	Large	3
15016	1422	Extra Large	4
15017	1426	Small	1
15018	1426	Medium	2
15019	1426	Large	3
15020	1426	Extra Large	4
15021	1432	Small	1
15022	1432	Medium	2
15023	1432	Large	3
15024	1432	Extra Large	4
15025	1442	Small	1
15026	1442	Medium	2
15027	1442	Large	3
15028	1442	Extra Large	4
15029	1447	Small	1
15030	1447	Medium	2
15031	1447	Large	3
15032	1447	Extra Large	4
15033	1465	Small	1
15034	1465	Medium	2
15035	1465	Large	3
15036	1465	Extra Large	4
15037	1467	Small	1
15038	1467	Medium	2
15039	1467	Large	3
15040	1467	Extra Large	4
15041	1472	Small	1
15042	1472	Medium	2
15043	1472	Large	3
15044	1472	Extra Large	4
15045	1473	Small	1
15046	1473	Medium	2
15047	1473	Large	3
15048	1473	Extra Large	4
15049	1477	Small	1
15050	1477	Medium	2
15051	1477	Large	3
15052	1477	Extra Large	4
15053	1481	Small	1
15054	1481	Medium	2
15055	1481	Large	3
15056	1481	Extra Large	4
15057	1483	Small	1
15058	1483	Medium	2
15059	1483	Large	3
15060	1483	Extra Large	4
15061	1487	Small	1
15062	1487	Medium	2
15063	1487	Large	3
15064	1487	Extra Large	4
15065	1488	Small	1
15066	1488	Medium	2
15067	1488	Large	3
15068	1488	Extra Large	4
15069	1498	Small	1
15070	1498	Medium	2
15071	1498	Large	3
15072	1498	Extra Large	4
15073	1510	Small	1
15074	1510	Medium	2
15075	1510	Large	3
15076	1510	Extra Large	4
15077	1511	Small	1
15078	1511	Medium	2
15079	1511	Large	3
15080	1511	Extra Large	4
15081	1512	Small	1
15082	1512	Medium	2
15083	1512	Large	3
15084	1512	Extra Large	4
15085	1515	Small	1
15086	1515	Medium	2
15087	1515	Large	3
15088	1515	Extra Large	4
15089	1516	Small	1
15090	1516	Medium	2
15091	1516	Large	3
15092	1516	Extra Large	4
15093	1517	Small	1
15094	1517	Medium	2
15095	1517	Large	3
15096	1517	Extra Large	4
15097	1518	Small	1
15098	1518	Medium	2
15099	1518	Large	3
15100	1518	Extra Large	4
15101	1521	Small	1
15102	1521	Medium	2
15103	1521	Large	3
15104	1521	Extra Large	4
15105	1522	Small	1
15106	1522	Medium	2
15107	1522	Large	3
15108	1522	Extra Large	4
15109	1525	Small	1
15110	1525	Medium	2
15111	1525	Large	3
15112	1525	Extra Large	4
15113	1526	Small	1
15114	1526	Medium	2
15115	1526	Large	3
15116	1526	Extra Large	4
15117	1527	Small	1
15118	1527	Medium	2
15119	1527	Large	3
15120	1527	Extra Large	4
15121	1531	Small	1
15122	1531	Medium	2
15123	1531	Large	3
15124	1531	Extra Large	4
15125	1532	Small	1
15126	1532	Medium	2
15127	1532	Large	3
15128	1532	Extra Large	4
15129	1536	Small	1
15130	1536	Medium	2
15131	1536	Large	3
15132	1536	Extra Large	4
15133	1537	Small	1
15134	1537	Medium	2
15135	1537	Large	3
15136	1537	Extra Large	4
15137	1542	Small	1
15138	1542	Medium	2
15139	1542	Large	3
15140	1542	Extra Large	4
15141	1543	Small	1
15142	1543	Medium	2
15143	1543	Large	3
15144	1543	Extra Large	4
15145	1547	Small	1
15146	1547	Medium	2
15147	1547	Large	3
15148	1547	Extra Large	4
15149	1548	Small	1
15150	1548	Medium	2
15151	1548	Large	3
15152	1548	Extra Large	4
15153	1553	Small	1
15154	1553	Medium	2
15155	1553	Large	3
15156	1553	Extra Large	4
15157	1563	Small	1
15158	1563	Medium	2
15159	1563	Large	3
15160	1563	Extra Large	4
15161	1567	Small	1
15162	1567	Medium	2
15163	1567	Large	3
15164	1567	Extra Large	4
15165	1571	Small	1
15166	1571	Medium	2
15167	1571	Large	3
15168	1571	Extra Large	4
15169	1573	Small	1
15170	1573	Medium	2
15171	1573	Large	3
15172	1573	Extra Large	4
15173	1576	Small	1
15174	1576	Medium	2
15175	1576	Large	3
15176	1576	Extra Large	4
15177	1578	Small	1
15178	1578	Medium	2
15179	1578	Large	3
15180	1578	Extra Large	4
15181	1582	Small	1
15182	1582	Medium	2
15183	1582	Large	3
15184	1582	Extra Large	4
15185	1583	Small	1
15186	1583	Medium	2
15187	1583	Large	3
15188	1583	Extra Large	4
15189	1587	Small	1
15190	1587	Medium	2
15191	1587	Large	3
15192	1587	Extra Large	4
15193	1588	Small	1
15194	1588	Medium	2
15195	1588	Large	3
15196	1588	Extra Large	4
15197	1591	Small	1
15198	1591	Medium	2
15199	1591	Large	3
15200	1591	Extra Large	4
15201	1592	Small	1
15202	1592	Medium	2
15203	1592	Large	3
15204	1592	Extra Large	4
15205	1603	Small	1
15206	1603	Medium	2
15207	1603	Large	3
15208	1603	Extra Large	4
15209	1610	Small	1
15210	1610	Medium	2
15211	1610	Large	3
15212	1610	Extra Large	4
15213	1611	Small	1
15214	1611	Medium	2
15215	1611	Large	3
15216	1611	Extra Large	4
15217	1612	Small	1
15218	1612	Medium	2
15219	1612	Large	3
15220	1612	Extra Large	4
15221	1677	Small	1
15222	1677	Medium	2
15223	1677	Large	3
15224	1677	Extra Large	4
15225	1834	Small	1
15226	1834	Medium	2
15227	1834	Large	3
15228	1834	Extra Large	4
15229	1845	Small	1
15230	1845	Medium	2
15231	1845	Large	3
15232	1845	Extra Large	4
15233	1893	Small	1
15234	1893	Medium	2
15235	1893	Large	3
15236	1893	Extra Large	4
15237	1897	Small	1
15238	1897	Medium	2
15239	1897	Large	3
15240	1897	Extra Large	4
15241	1900	Small	1
15242	1900	Medium	2
15243	1900	Large	3
15244	1900	Extra Large	4
15245	1901	Small	1
15246	1901	Medium	2
15247	1901	Large	3
15248	1901	Extra Large	4
15249	1929	Small	1
15250	1929	Medium	2
15251	1929	Large	3
15252	1929	Extra Large	4
15253	1942	Small	1
15254	1942	Medium	2
15255	1942	Large	3
15256	1942	Extra Large	4
15257	1946	Small	1
15258	1946	Medium	2
15259	1946	Large	3
15260	1946	Extra Large	4
15261	1949	Small	1
15262	1949	Medium	2
15263	1949	Large	3
15264	1949	Extra Large	4
15265	1950	Small	1
15266	1950	Medium	2
15267	1950	Large	3
15268	1950	Extra Large	4
15269	1953	Small	1
15270	1953	Medium	2
15271	1953	Large	3
15272	1953	Extra Large	4
15273	1966	Small	1
15274	1966	Medium	2
15275	1966	Large	3
15276	1966	Extra Large	4
15277	1978	Small	1
15278	1978	Medium	2
15279	1978	Large	3
15280	1978	Extra Large	4
15281	1982	Small	1
15282	1982	Medium	2
15283	1982	Large	3
15284	1982	Extra Large	4
15285	1986	Small	1
15286	1986	Medium	2
15287	1986	Large	3
15288	1986	Extra Large	4
15289	2001	Small	1
15290	2001	Medium	2
15291	2001	Large	3
15292	2001	Extra Large	4
15293	2006	Small	1
15294	2006	Medium	2
15295	2006	Large	3
15296	2006	Extra Large	4
15297	2009	Small	1
15298	2009	Medium	2
15299	2009	Large	3
15300	2009	Extra Large	4
15301	2013	Small	1
15302	2013	Medium	2
15303	2013	Large	3
15304	2013	Extra Large	4
15305	2017	Small	1
15306	2017	Medium	2
15307	2017	Large	3
15308	2017	Extra Large	4
15309	2021	Small	1
15310	2021	Medium	2
15311	2021	Large	3
15312	2021	Extra Large	4
15313	2022	Small	1
15314	2022	Medium	2
15315	2022	Large	3
15316	2022	Extra Large	4
15317	2029	Small	1
15318	2029	Medium	2
15319	2029	Large	3
15320	2029	Extra Large	4
15321	2030	Small	1
15322	2030	Medium	2
15323	2030	Large	3
15324	2030	Extra Large	4
15325	2033	Small	1
15326	2033	Medium	2
15327	2033	Large	3
15328	2033	Extra Large	4
15329	2034	Small	1
15330	2034	Medium	2
15331	2034	Large	3
15332	2034	Extra Large	4
15333	2065	Small	1
15334	2065	Medium	2
15335	2065	Large	3
15336	2065	Extra Large	4
15337	2090	Small	1
15338	2090	Medium	2
15339	2090	Large	3
15340	2090	Extra Large	4
15341	2121	Small	1
15342	2121	Medium	2
15343	2121	Large	3
15344	2121	Extra Large	4
15345	2127	Small	1
15346	2127	Medium	2
15347	2127	Large	3
15348	2127	Extra Large	4
15349	2141	Small	1
15350	2141	Medium	2
15351	2141	Large	3
15352	2141	Extra Large	4
15353	2211	Small	1
15354	2211	Medium	2
15355	2211	Large	3
15356	2211	Extra Large	4
15357	2217	Small	1
15358	2217	Medium	2
15359	2217	Large	3
15360	2217	Extra Large	4
15361	2223	Small	1
15362	2223	Medium	2
15363	2223	Large	3
15364	2223	Extra Large	4
15365	2289	Small	1
15366	2289	Medium	2
15367	2289	Large	3
15368	2289	Extra Large	4
15369	2298	Small	1
15370	2298	Medium	2
15371	2298	Large	3
15372	2298	Extra Large	4
15373	2305	Small	1
15374	2305	Medium	2
15375	2305	Large	3
15376	2305	Extra Large	4
15377	2317	Small	1
15378	2317	Medium	2
15379	2317	Large	3
15380	2317	Extra Large	4
15381	2327	Small	1
15382	2327	Medium	2
15383	2327	Large	3
15384	2327	Extra Large	4
15385	2424	Small	1
15386	2424	Medium	2
15387	2424	Large	3
15388	2424	Extra Large	4
15389	2436	Small	1
15390	2436	Medium	2
15391	2436	Large	3
15392	2436	Extra Large	4
15393	2456	Small	1
15394	2456	Medium	2
15395	2456	Large	3
15396	2456	Extra Large	4
15397	2486	Small	1
15398	2486	Medium	2
15399	2486	Large	3
15400	2486	Extra Large	4
15401	2503	Small	1
15402	2503	Medium	2
15403	2503	Large	3
15404	2503	Extra Large	4
15405	2504	Small	1
15406	2504	Medium	2
15407	2504	Large	3
15408	2504	Extra Large	4
15409	2505	Small	1
15410	2505	Medium	2
15411	2505	Large	3
15412	2505	Extra Large	4
15413	2506	Small	1
15414	2506	Medium	2
15415	2506	Large	3
15416	2506	Extra Large	4
15417	2508	Small	1
15418	2508	Medium	2
15419	2508	Large	3
15420	2508	Extra Large	4
15421	2509	Small	1
15422	2509	Medium	2
15423	2509	Large	3
15424	2509	Extra Large	4
15425	2510	Small	1
15426	2510	Medium	2
15427	2510	Large	3
15428	2510	Extra Large	4
15429	2511	Small	1
15430	2511	Medium	2
15431	2511	Large	3
15432	2511	Extra Large	4
15433	2513	Small	1
15434	2513	Medium	2
15435	2513	Large	3
15436	2513	Extra Large	4
15437	2514	Small	1
15438	2514	Medium	2
15439	2514	Large	3
15440	2514	Extra Large	4
15441	2515	Small	1
15442	2515	Medium	2
15443	2515	Large	3
15444	2515	Extra Large	4
15445	2518	Small	1
15446	2518	Medium	2
15447	2518	Large	3
15448	2518	Extra Large	4
15449	2519	Small	1
15450	2519	Medium	2
15451	2519	Large	3
15452	2519	Extra Large	4
15453	2520	Small	1
15454	2520	Medium	2
15455	2520	Large	3
15456	2520	Extra Large	4
15457	2521	Small	1
15458	2521	Medium	2
15459	2521	Large	3
15460	2521	Extra Large	4
15461	2526	Small	1
15462	2526	Medium	2
15463	2526	Large	3
15464	2526	Extra Large	4
15465	2535	Small	1
15466	2535	Medium	2
15467	2535	Large	3
15468	2535	Extra Large	4
\.


--
-- Data for Name: sub_type_attributes; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sub_type_attributes (id, sub_category_id, field_name, field_type, is_required, sort_order) FROM stdin;
1	1	Size	dropdown	t	1
2	1	Length	dropdown	t	2
3	1	Material	dropdown	t	3
4	1	Finish	dropdown	t	4
5	1	Head Type	dropdown	t	5
6	1	Thread Type	dropdown	t	6
7	1	Quantity	number	t	7
8	2	Size	dropdown	t	1
9	2	Length	dropdown	t	2
10	2	Material	dropdown	t	3
11	2	Finish	dropdown	t	4
12	2	Head Type	dropdown	t	5
13	2	Thread Type	dropdown	t	6
14	2	Quantity	number	t	7
15	3	Size	dropdown	t	1
16	3	Length	dropdown	t	2
17	3	Material	dropdown	t	3
18	3	Finish	dropdown	t	4
19	3	Head Type	dropdown	t	5
20	3	Thread Type	dropdown	t	6
21	3	Quantity	number	t	7
22	4	Size	dropdown	t	1
23	4	Length	dropdown	t	2
24	4	Material	dropdown	t	3
25	4	Finish	dropdown	t	4
26	4	Head Type	dropdown	t	5
27	4	Thread Type	dropdown	t	6
28	4	Quantity	number	t	7
29	5	Size	dropdown	t	1
30	5	Length	dropdown	t	2
31	5	Material	dropdown	t	3
32	5	Finish	dropdown	t	4
33	5	Head Type	dropdown	t	5
34	5	Thread Type	dropdown	t	6
35	5	Quantity	number	t	7
36	6	Size	dropdown	t	1
37	6	Length	dropdown	t	2
38	6	Material	dropdown	t	3
39	6	Finish	dropdown	t	4
40	6	Head Type	dropdown	t	5
41	6	Thread Type	dropdown	t	6
42	6	Quantity	number	t	7
43	7	Size	dropdown	t	1
44	7	Length	dropdown	t	2
45	7	Material	dropdown	t	3
46	7	Finish	dropdown	t	4
47	7	Head Type	dropdown	t	5
48	7	Thread Type	dropdown	t	6
49	7	Quantity	number	t	7
50	8	Size	dropdown	t	1
51	8	Length	dropdown	t	2
52	8	Material	dropdown	t	3
53	8	Finish	dropdown	t	4
54	8	Head Type	dropdown	t	5
55	8	Thread Type	dropdown	t	6
56	8	Quantity	number	t	7
57	9	Size	dropdown	t	1
58	9	Length	dropdown	t	2
59	9	Material	dropdown	t	3
60	9	Finish	dropdown	t	4
61	9	Head Type	dropdown	t	5
62	9	Thread Type	dropdown	t	6
63	9	Quantity	number	t	7
64	10	Size	dropdown	t	1
65	10	Length	dropdown	t	2
66	10	Material	dropdown	t	3
67	10	Finish	dropdown	t	4
68	10	Head Type	dropdown	t	5
69	10	Thread Type	dropdown	t	6
70	10	Quantity	number	t	7
71	11	Size	dropdown	t	1
72	11	Length	dropdown	t	2
73	11	Material	dropdown	t	3
74	11	Finish	dropdown	t	4
75	11	Head Type	dropdown	t	5
76	11	Thread Type	dropdown	t	6
77	11	Quantity	number	t	7
78	12	Size	dropdown	t	1
79	12	Material	dropdown	t	2
80	12	Finish	dropdown	t	3
81	12	Thread Type	dropdown	t	4
82	12	Quantity	number	t	5
83	13	Size	dropdown	t	1
84	13	Material	dropdown	t	2
85	13	Finish	dropdown	t	3
86	13	Thread Type	dropdown	t	4
87	13	Quantity	number	t	5
88	14	Size	dropdown	t	1
89	14	Material	dropdown	t	2
90	14	Finish	dropdown	t	3
91	14	Quantity	number	t	4
92	15	Size	dropdown	t	1
93	15	Material	dropdown	t	2
94	15	Finish	dropdown	t	3
95	15	Quantity	number	t	4
96	16	Diameter	dropdown	t	1
97	16	Length	number	t	2
98	16	Material	dropdown	t	3
99	16	Pressure Rating	dropdown	t	4
100	16	Quantity	number	t	5
101	17	Diameter	dropdown	t	1
102	17	Length	number	t	2
103	17	Material	dropdown	t	3
104	17	Pressure Rating	dropdown	t	4
105	17	Quantity	number	t	5
106	18	Diameter	dropdown	t	1
107	18	Length	number	t	2
108	18	Material	dropdown	t	3
109	18	Pressure Rating	dropdown	t	4
110	18	Quantity	number	t	5
111	19	Diameter	dropdown	t	1
112	19	Length	number	t	2
113	19	Thickness	dropdown	t	3
114	19	Material	dropdown	t	4
115	19	Quantity	number	t	5
116	20	Size	dropdown	t	1
117	20	Angle	dropdown	t	2
118	20	Material	dropdown	t	3
119	20	Quantity	number	t	4
120	21	Size	dropdown	t	1
121	21	Material	dropdown	t	2
122	21	Quantity	number	t	3
123	22	Size	dropdown	t	1
124	22	Pressure Rating	dropdown	t	2
125	22	Material	dropdown	t	3
126	22	Quantity	number	t	4
127	23	Size	dropdown	t	1
128	23	Pressure Rating	dropdown	t	2
129	23	Material	dropdown	t	3
130	23	Quantity	number	t	4
131	24	Type	dropdown	t	1
132	24	Material	dropdown	t	2
133	24	Finish	dropdown	t	3
134	24	Quantity	number	t	4
135	25	Size	dropdown	t	1
136	25	Material	dropdown	t	2
137	25	Color	text	t	3
138	25	Mount Type	dropdown	t	4
139	25	Quantity	number	t	5
140	26	Wattage	dropdown	t	1
141	26	Voltage	dropdown	t	2
142	26	Base Type	dropdown	t	3
143	26	Color Temp	dropdown	t	4
144	26	Quantity	number	t	5
145	27	Wattage	dropdown	t	1
146	27	Voltage	dropdown	t	2
147	27	Base Type	dropdown	t	3
148	27	Color Temp	dropdown	t	4
149	27	Quantity	number	t	5
150	28	Wattage	dropdown	t	1
151	28	Voltage	dropdown	t	2
152	28	Length	dropdown	t	3
153	28	Color Temp	dropdown	t	4
154	28	Quantity	number	t	5
155	29	Gauge	dropdown	t	1
156	29	Length	number	t	2
157	29	Core Type	dropdown	t	3
158	29	Insulation	dropdown	t	4
159	29	Quantity	number	t	5
160	30	Rating	dropdown	t	1
161	30	Poles	dropdown	t	2
162	30	Type	dropdown	t	3
163	30	Brand	text	t	4
164	30	Quantity	number	t	5
165	31	Type	dropdown	t	1
166	31	Modules	dropdown	t	2
167	31	Voltage	dropdown	t	3
168	31	Brand	text	t	4
169	31	Quantity	number	t	5
170	32	Type	dropdown	t	1
171	32	Modules	dropdown	t	2
172	32	Voltage	dropdown	t	3
173	32	Brand	text	t	4
174	32	Quantity	number	t	5
175	33	Blade Size	dropdown	t	1
176	33	Wattage	dropdown	t	2
177	33	Voltage	dropdown	t	3
178	33	Sweep	dropdown	t	4
179	33	Brand	text	t	5
180	33	Quantity	number	t	6
181	34	Size	dropdown	t	1
182	34	Wattage	dropdown	t	2
183	34	Voltage	dropdown	t	3
184	34	Air Delivery	number	t	4
185	34	Brand	text	t	5
186	34	Quantity	number	t	6
187	35	Ways	dropdown	t	1
188	35	Type	dropdown	t	2
189	35	Material	dropdown	t	3
190	35	Brand	text	t	4
191	35	Quantity	number	t	5
192	36	Color	text	t	1
193	36	Finish Type	dropdown	t	2
194	36	Volume	dropdown	t	3
195	36	Base Type	dropdown	t	4
196	36	Quantity	number	t	5
197	37	Color	text	t	1
198	37	Finish Type	dropdown	t	2
199	37	Volume	dropdown	t	3
200	37	Base Type	dropdown	t	4
201	37	Quantity	number	t	5
202	38	Color	text	t	1
203	38	Sheen	dropdown	t	2
204	38	Volume	dropdown	t	3
205	38	Coverage Area	number	t	4
206	38	Quantity	number	t	5
207	39	Color	text	t	1
208	39	Weather Resistance	dropdown	t	2
209	39	Volume	dropdown	t	3
210	39	Coverage Area	number	t	4
211	39	Quantity	number	t	5
212	40	Color	text	t	1
213	40	Sheen	dropdown	t	2
214	40	Volume	dropdown	t	3
215	40	Surface Type	dropdown	t	4
216	40	Quantity	number	t	5
217	41	Type	dropdown	t	1
218	41	Volume	dropdown	t	2
219	41	Coverage Area	number	t	3
220	41	Quantity	number	t	4
221	42	Type	dropdown	t	1
222	42	Volume	dropdown	t	2
223	42	Coverage Area	number	t	3
224	42	Quantity	number	t	4
225	43	Type	dropdown	t	1
226	43	Weight	dropdown	t	2
227	43	Coverage Area	number	t	3
228	43	Quantity	number	t	4
229	44	Type	dropdown	t	1
230	44	Volume	dropdown	t	2
231	44	Quantity	number	t	3
232	45	Size	dropdown	t	1
233	45	Material	dropdown	t	2
234	45	Handle Length	dropdown	t	3
235	45	Quantity	number	t	4
236	46	Color	text	t	1
237	46	Finish Type	dropdown	t	2
238	46	Volume	dropdown	t	3
239	46	Base Type	dropdown	t	4
240	46	Quantity	number	t	5
241	47	Color	text	t	1
242	47	Finish Type	dropdown	t	2
243	47	Volume	dropdown	t	3
244	47	Base Type	dropdown	t	4
245	47	Quantity	number	t	5
246	48	Color	text	t	1
247	48	Sheen	dropdown	t	2
248	48	Volume	dropdown	t	3
249	48	Coverage Area	number	t	4
250	48	Quantity	number	t	5
251	49	Color	text	t	1
252	49	Weather Resistance	dropdown	t	2
253	49	Volume	dropdown	t	3
254	49	Coverage Area	number	t	4
255	49	Quantity	number	t	5
256	50	Color	text	t	1
257	50	Sheen	dropdown	t	2
258	50	Volume	dropdown	t	3
259	50	Surface Type	dropdown	t	4
260	50	Quantity	number	t	5
261	51	Type	dropdown	t	1
262	51	Volume	dropdown	t	2
263	51	Coverage Area	number	t	3
264	51	Quantity	number	t	4
265	52	Type	dropdown	t	1
266	52	Volume	dropdown	t	2
267	52	Coverage Area	number	t	3
268	52	Quantity	number	t	4
269	53	Type	dropdown	t	1
270	53	Weight	dropdown	t	2
271	53	Coverage Area	number	t	3
272	53	Quantity	number	t	4
273	54	Type	dropdown	t	1
274	54	Volume	dropdown	t	2
275	54	Quantity	number	t	3
276	55	Size	dropdown	t	1
277	55	Material	dropdown	t	2
278	55	Handle Length	dropdown	t	3
279	55	Quantity	number	t	4
280	56	Brand	text	t	1
281	56	Power Type	dropdown	t	2
282	56	Power Rating	number	t	3
283	56	Chuck Size	dropdown	t	4
284	56	Weight	number	t	5
285	56	Quantity	number	t	6
286	57	Brand	text	t	1
287	57	Power Type	dropdown	t	2
288	57	Power Rating	number	t	3
289	57	Disc Diameter	dropdown	t	4
290	57	Weight	number	t	5
291	57	Quantity	number	t	6
292	58	Brand	text	t	1
293	58	Power Type	dropdown	t	2
294	58	Power Rating	number	t	3
295	58	Blade Diameter	dropdown	t	4
296	58	Cutting Depth	number	t	5
297	58	Quantity	number	t	6
298	59	Brand	text	t	1
299	59	Power Type	dropdown	t	2
300	59	Power Rating	number	t	3
301	59	Stroke Length	dropdown	t	4
302	59	Weight	number	t	5
303	59	Quantity	number	t	6
304	60	Brand	text	t	1
305	60	Power Type	dropdown	t	2
306	60	Power Rating	number	t	3
307	60	Disc Size	dropdown	t	4
308	60	Weight	number	t	5
309	60	Quantity	number	t	6
310	61	Brand	text	t	1
311	61	Power Type	dropdown	t	2
312	61	Torque Rating	number	t	3
313	61	Drive Size	dropdown	t	4
314	61	Weight	number	t	5
315	61	Quantity	number	t	6
316	62	Brand	text	t	1
317	62	Power Type	dropdown	t	2
318	62	Power Rating	number	t	3
319	62	Chuck Size	dropdown	t	4
320	62	Impact Energy	number	t	5
321	62	Quantity	number	t	6
322	63	Brand	text	t	1
323	63	Battery Voltage	dropdown	t	2
324	63	Battery Capacity	dropdown	t	3
325	63	Tool Type	dropdown	t	4
326	63	Quantity	number	t	5
327	64	Type	dropdown	t	1
328	64	Head Weight	dropdown	t	2
329	64	Handle Material	dropdown	t	3
330	64	Length	dropdown	t	4
331	64	Quantity	number	t	5
332	65	Type	dropdown	t	1
333	65	Tip Size	dropdown	t	2
334	65	Handle Material	dropdown	t	3
335	65	Length	dropdown	t	4
336	65	Quantity	number	t	5
337	66	Type	dropdown	t	1
338	66	Size Range	text	t	2
339	66	Material	dropdown	t	3
340	66	Quantity	number	t	4
341	67	Type	dropdown	t	1
342	67	Jaw Size	dropdown	t	2
343	67	Material	dropdown	t	3
344	67	Length	dropdown	t	4
345	67	Quantity	number	t	5
346	68	Blade Width	dropdown	t	1
347	68	Blade Length	dropdown	t	2
348	68	Handle Material	dropdown	t	3
349	68	Quantity	number	t	4
350	69	Blade Length	dropdown	t	1
351	69	Teeth Per Inch	dropdown	t	2
352	69	Blade Material	dropdown	t	3
353	69	Handle Type	dropdown	t	4
354	69	Quantity	number	t	5
355	70	Type	dropdown	t	1
356	70	Length	dropdown	t	2
357	70	Cut Type	dropdown	t	3
358	70	Quantity	number	t	4
359	71	Kit Size	dropdown	t	1
360	71	Tool Count	number	t	2
361	71	Case Type	dropdown	t	3
362	71	Brand	text	t	4
363	71	Quantity	number	t	5
364	72	Capacity	dropdown	t	1
365	72	Material	dropdown	t	2
366	72	Compartments	dropdown	t	3
367	72	Dimensions	text	t	4
368	72	Quantity	number	t	5
369	73	Size	dropdown	t	1
370	73	Weight Capacity	number	t	2
371	73	Material	dropdown	t	3
372	73	Height	number	t	4
373	73	Quantity	number	t	5
374	74	Type	dropdown	t	1
375	74	Capacity	number	t	2
376	74	Material	dropdown	t	3
377	74	Mount Type	dropdown	t	4
378	74	Quantity	number	t	5
379	75	Length	dropdown	t	1
380	75	Gauge	dropdown	t	2
381	75	Rating	dropdown	t	3
382	75	Sockets	dropdown	t	4
383	75	Quantity	number	t	5
384	76	Capacity	dropdown	t	1
385	76	Material	dropdown	t	2
386	76	Dimensions	text	t	3
387	76	Pockets	number	t	4
388	76	Quantity	number	t	5
389	77	Height	dropdown	t	1
390	77	Weight Capacity	number	t	2
391	77	Steps	number	t	3
392	77	Type	dropdown	t	4
393	77	Quantity	number	t	5
394	78	Height	dropdown	t	1
395	78	Weight Capacity	number	t	2
396	78	Steps	number	t	3
397	78	Type	dropdown	t	4
398	78	Quantity	number	t	5
399	79	Height	dropdown	t	1
400	79	Weight Capacity	number	t	2
401	79	Steps	number	t	3
402	79	Material	dropdown	t	4
403	79	Quantity	number	t	5
404	80	Extended Length	dropdown	t	1
405	80	Closed Length	dropdown	t	2
406	80	Weight Capacity	number	t	3
407	80	Sections	number	t	4
408	80	Quantity	number	t	5
409	81	Height	dropdown	t	1
410	81	Load Capacity	number	t	2
411	81	Material	dropdown	t	3
412	81	Platform Size	text	t	4
413	81	Quantity	number	t	5
414	82	Length	dropdown	t	1
415	82	Blade Width	dropdown	t	2
416	82	Material	dropdown	t	3
417	82	Accuracy	text	t	4
418	82	Quantity	number	t	5
419	83	Length	dropdown	t	1
420	83	Vials	dropdown	t	2
421	83	Material	dropdown	t	3
422	83	Accuracy	text	t	4
423	83	Quantity	number	t	5
424	84	Range	dropdown	t	1
425	84	Resolution	dropdown	t	2
426	84	Material	dropdown	t	3
427	84	Type	dropdown	t	4
428	84	Quantity	number	t	5
429	85	Range	dropdown	t	1
430	85	Accuracy	text	t	2
431	85	Battery Type	dropdown	t	3
432	85	Brand	text	t	4
433	85	Quantity	number	t	5
434	86	Blade Length	dropdown	t	1
435	86	Teeth Per Inch	dropdown	t	2
436	86	Frame Material	dropdown	t	3
437	86	Handle Type	dropdown	t	4
438	86	Quantity	number	t	5
439	87	Pipe Capacity	dropdown	t	1
440	87	Material	dropdown	t	2
441	87	Wheel Type	dropdown	t	3
442	87	Quantity	number	t	4
443	88	Cutting Capacity	dropdown	t	1
444	88	Handle Length	dropdown	t	2
445	88	Jaw Material	dropdown	t	3
446	88	Quantity	number	t	4
447	89	Cutting Capacity	dropdown	t	1
448	89	Cut Type	dropdown	t	2
449	89	Handle Material	dropdown	t	3
450	89	Length	dropdown	t	4
451	89	Quantity	number	t	5
452	90	Blade Type	dropdown	t	1
453	90	Handle Material	dropdown	t	2
454	90	Blade Length	dropdown	t	3
455	90	Quantity	number	t	4
456	91	Type	dropdown	t	1
457	91	Material	dropdown	t	2
458	91	Color	text	t	3
459	91	Size	dropdown	t	4
460	91	Quantity	number	t	5
461	92	Type	dropdown	t	1
462	92	Material	dropdown	t	2
463	92	Size	dropdown	t	3
464	92	Cut Resistance	dropdown	t	4
465	92	Quantity	number	t	5
466	93	Type	dropdown	t	1
467	93	Lens Material	dropdown	t	2
468	93	Frame Material	dropdown	t	3
469	93	Anti-Fog	dropdown	t	4
470	93	Quantity	number	t	5
471	94	Type	dropdown	t	1
472	94	NRR Rating	dropdown	t	2
473	94	Material	dropdown	t	3
474	94	Size	dropdown	t	4
475	94	Quantity	number	t	5
476	95	Type	dropdown	t	1
477	95	Filter Rating	dropdown	t	2
478	95	Material	dropdown	t	3
479	95	Size	dropdown	t	4
480	95	Quantity	number	t	5
481	96	Type	dropdown	t	1
482	96	Blade Size	dropdown	t	2
483	96	Handle Length	dropdown	t	3
484	96	Handle Material	dropdown	t	4
485	96	Quantity	number	t	5
486	97	Tine Count	dropdown	t	1
487	97	Handle Length	dropdown	t	2
488	97	Material	dropdown	t	3
489	97	Quantity	number	t	4
490	98	Blade Length	dropdown	t	1
491	98	Cut Capacity	dropdown	t	2
492	98	Type	dropdown	t	3
493	98	Handle Material	dropdown	t	4
494	98	Quantity	number	t	5
495	99	Tine Count	dropdown	t	1
496	99	Handle Length	dropdown	t	2
497	99	Head Width	dropdown	t	3
498	99	Material	dropdown	t	4
499	99	Quantity	number	t	5
500	100	Capacity	dropdown	t	1
501	100	Material	dropdown	t	2
502	100	Spout Type	dropdown	t	3
503	100	Handle Type	dropdown	t	4
504	100	Quantity	number	t	5
505	101	Diameter	dropdown	t	1
506	101	Length	dropdown	t	2
507	101	Type	dropdown	t	3
508	101	Current Type	dropdown	t	4
509	101	Quantity	number	t	5
510	102	Diameter	dropdown	t	1
511	102	Length	dropdown	t	2
512	102	AWS Grade	text	t	3
513	102	Quantity	number	t	4
514	103	Type	dropdown	t	1
515	103	Shade Range	dropdown	t	2
516	103	Material	dropdown	t	3
517	103	Size	dropdown	t	4
518	103	Quantity	number	t	5
519	104	Material	dropdown	t	1
520	104	Size	dropdown	t	2
521	104	Heat Resistance	dropdown	t	3
522	104	Quantity	number	t	4
523	105	Size	dropdown	t	1
524	105	Application Type	dropdown	t	2
525	105	Setting Time	dropdown	t	3
526	105	Quantity	number	t	4
527	106	Type	dropdown	t	1
528	106	Mix Ratio	dropdown	t	2
529	106	Cure Time	dropdown	t	3
530	106	Quantity	number	t	4
531	107	Type	dropdown	t	1
532	107	Viscosity	dropdown	t	2
533	107	Strength	dropdown	t	3
534	107	Quantity	number	t	4
535	108	Color	text	t	1
536	108	Type	dropdown	t	2
537	108	Cure Time	dropdown	t	3
538	108	Cartridge Size	dropdown	t	4
539	108	Quantity	number	t	5
540	109	Brand	text	t	1
541	109	Grade	dropdown	t	2
542	109	Bag Weight	dropdown	t	3
543	109	Quantity	number	t	4
544	110	Type	dropdown	t	1
545	110	Bag Weight	dropdown	t	2
546	110	Quantity	number	t	3
547	111	Type	dropdown	t	1
548	111	Size	dropdown	t	2
549	111	Quantity	number	t	3
550	112	Size	dropdown	t	1
551	112	Type	dropdown	t	2
552	112	Weight	number	t	3
553	112	Quantity	number	t	4
554	113	Diameter	dropdown	t	1
555	113	Length	dropdown	t	2
556	113	Grade	dropdown	t	3
557	113	Quantity	number	t	4
558	114	Diameter	dropdown	t	1
559	114	Length	dropdown	t	2
560	114	Grade	dropdown	t	3
561	114	Quantity	number	t	4
562	115	Diameter	dropdown	t	1
563	115	Length	dropdown	t	2
564	115	Grade	dropdown	t	3
565	115	Quantity	number	t	4
566	116	Diameter	dropdown	t	1
567	116	Length	dropdown	t	2
568	116	Grade	dropdown	t	3
569	116	Quantity	number	t	4
570	117	Diameter	dropdown	t	1
571	117	Length	dropdown	t	2
572	117	Grade	dropdown	t	3
573	117	Quantity	number	t	4
574	118	Diameter	dropdown	t	1
575	118	Length	dropdown	t	2
576	118	Grade	dropdown	t	3
577	118	Quantity	number	t	4
578	119	Size	dropdown	t	1
579	119	Thickness	dropdown	t	2
580	119	Length	dropdown	t	3
581	119	Quantity	number	t	4
582	120	Size	dropdown	t	1
583	120	Thickness	dropdown	t	2
584	120	Length	dropdown	t	3
585	120	Quantity	number	t	4
586	121	Thickness	dropdown	t	1
587	121	Size	dropdown	t	2
588	121	Grade	dropdown	t	3
589	121	Quantity	number	t	4
590	122	Thickness	dropdown	t	1
591	122	Width	dropdown	t	2
592	122	Length	dropdown	t	3
593	122	Grade	dropdown	t	4
594	122	Quantity	number	t	5
595	123	Thickness	dropdown	t	1
596	123	Width	dropdown	t	2
597	123	Length	dropdown	t	3
598	123	Grade	dropdown	t	4
599	123	Quantity	number	t	5
600	124	Thickness	dropdown	t	1
601	124	Width	dropdown	t	2
602	124	Length	dropdown	t	3
603	124	Grade	dropdown	t	4
604	124	Quantity	number	t	5
605	125	Thickness	dropdown	t	1
606	125	Grade	dropdown	t	2
607	125	Type	dropdown	t	3
608	125	Quantity	number	t	4
609	126	Size	dropdown	t	1
610	126	Thickness	dropdown	t	2
611	126	Type	dropdown	t	3
612	126	Material	dropdown	t	4
613	126	Quantity	number	t	5
614	127	Diameter	dropdown	t	1
615	127	Length	dropdown	t	2
616	127	Grade	dropdown	t	3
617	127	Quantity	number	t	4
618	128	Thickness	dropdown	t	1
619	128	Width	dropdown	t	2
620	128	Length	dropdown	t	3
621	128	Grade	dropdown	t	4
622	128	Quantity	number	t	5
623	129	Color	text	t	1
624	129	Thickness	dropdown	t	2
625	129	Width	dropdown	t	3
626	129	Length	dropdown	t	4
627	129	Quantity	number	t	5
628	130	Thickness	dropdown	t	1
629	130	Color	dropdown	t	2
630	130	Size	dropdown	t	3
631	130	Type	dropdown	t	4
632	130	Quantity	number	t	5
633	131	Thickness	dropdown	t	1
634	131	Size	dropdown	t	2
635	131	Grade	dropdown	t	3
636	131	Quantity	number	t	4
637	132	Size	dropdown	t	1
638	132	Length	dropdown	t	2
639	132	Material	dropdown	t	3
640	132	Head Type	dropdown	t	4
641	132	Quantity	number	t	5
642	133	Gauge	dropdown	t	1
643	133	Length	dropdown	t	2
644	133	Coil Weight	dropdown	t	3
645	133	Quantity	number	t	4
646	134	Height	dropdown	t	1
647	134	Width	dropdown	t	2
648	134	Material	dropdown	t	3
649	134	Reflective Type	dropdown	t	4
650	134	Quantity	number	t	5
651	135	Mesh Size	dropdown	t	1
652	135	Size	text	t	2
653	135	Material	dropdown	t	3
654	135	Quantity	number	t	4
655	136	Type	dropdown	t	1
656	136	Size	dropdown	t	2
657	136	Material	dropdown	t	3
658	136	Quantity	number	t	4
659	137	Bag Weight	dropdown	t	1
660	137	Brand	text	t	2
661	137	Quantity	number	t	3
662	138	Bag Weight	dropdown	t	1
663	138	Brand	text	t	2
664	138	Quantity	number	t	3
665	139	Bag Weight	dropdown	t	1
666	139	Brand	text	t	2
667	139	Quantity	number	t	3
668	140	Grade	dropdown	t	1
669	140	Cubic Meter	number	t	2
670	140	Slump	dropdown	t	3
671	140	Quantity	number	t	4
672	141	Type	dropdown	t	1
673	141	Bag Weight	dropdown	t	2
674	141	Quantity	number	t	3
675	142	Size	dropdown	t	1
676	142	Bag Weight	dropdown	t	2
677	142	Quantity	number	t	3
678	143	Size	dropdown	t	1
679	143	Bag Weight	dropdown	t	2
680	143	Quantity	number	t	3
681	144	Size	dropdown	t	1
682	144	Bag Weight	dropdown	t	2
683	144	Quantity	number	t	3
684	145	Size	dropdown	t	1
685	145	Bag Weight	dropdown	t	2
686	145	Quantity	number	t	3
687	146	Size	dropdown	t	1
688	146	Quantity	number	t	2
689	147	Size	dropdown	t	1
690	147	Grade	dropdown	t	2
691	147	Quantity	number	t	3
692	148	Size	dropdown	t	1
693	148	Type	dropdown	t	2
694	148	Quantity	number	t	3
695	149	Size	dropdown	t	1
696	149	Density	dropdown	t	2
697	149	Quantity	number	t	3
698	150	Size	dropdown	t	1
699	150	Thickness	dropdown	t	2
700	150	Type	dropdown	t	3
701	150	Material	dropdown	t	4
702	150	Quantity	number	t	5
703	151	Size	dropdown	t	1
704	151	Thickness	dropdown	t	2
705	151	Design Type	dropdown	t	3
706	151	Material	dropdown	t	4
707	151	Quantity	number	t	5
708	152	Size	dropdown	t	1
709	152	Thickness	dropdown	t	2
710	152	Color	text	t	3
711	152	Frame Type	dropdown	t	4
712	152	Quantity	number	t	5
713	153	Size	dropdown	t	1
714	153	Frame Thickness	dropdown	t	2
715	153	Glass Type	dropdown	t	3
716	153	Opening Type	dropdown	t	4
717	153	Quantity	number	t	5
718	154	Size	dropdown	t	1
719	154	Profile Type	dropdown	t	2
720	154	Glass Type	dropdown	t	3
721	154	Color	text	t	4
722	154	Quantity	number	t	5
723	155	Thickness	dropdown	t	1
724	155	Size	dropdown	t	2
725	155	Grade	dropdown	t	3
726	155	Quantity	number	t	4
727	156	Thickness	dropdown	t	1
728	156	Size	dropdown	t	2
729	156	Grade	dropdown	t	3
730	156	Quantity	number	t	4
731	157	Thickness	dropdown	t	1
732	157	Size	dropdown	t	2
733	157	Grade	dropdown	t	3
734	157	Quantity	number	t	4
735	158	Thickness	dropdown	t	1
736	158	Size	dropdown	t	2
737	158	Finish Type	dropdown	t	3
738	158	Design	text	t	4
739	158	Quantity	number	t	5
740	159	Thickness	dropdown	t	1
741	159	Size	dropdown	t	2
742	159	Type	dropdown	t	3
743	159	Quantity	number	t	4
744	160	Thickness	dropdown	t	1
745	160	Bag Weight	dropdown	t	2
746	160	Coverage Area	number	t	3
747	160	Quantity	number	t	4
748	161	Type	dropdown	t	1
749	161	Bag Weight	dropdown	t	2
750	161	Coverage Area	number	t	3
751	161	Quantity	number	t	4
752	162	Grade	dropdown	t	1
753	162	Bag Weight	dropdown	t	2
754	162	Quantity	number	t	3
755	163	Size	dropdown	t	1
756	163	Finish	dropdown	t	2
757	163	Color	text	t	3
758	163	Quantity	number	t	4
759	164	Size	dropdown	t	1
760	164	Finish	dropdown	t	2
761	164	Rectified	dropdown	t	3
762	164	Quantity	number	t	4
763	165	Size	dropdown	t	1
764	165	Finish	dropdown	t	2
765	165	Thickness	dropdown	t	3
766	165	Quantity	number	t	4
767	166	Slab Size	dropdown	t	1
768	166	Thickness	dropdown	t	2
769	166	Finish	dropdown	t	3
770	166	Quantity	number	t	4
771	167	Slab Size	dropdown	t	1
772	167	Thickness	dropdown	t	2
773	167	Finish	dropdown	t	3
774	167	Color	text	t	4
775	167	Quantity	number	t	5
776	168	Size	dropdown	t	1
777	168	Length	dropdown	t	2
778	168	Weight per Meter	number	t	3
779	168	Grade	dropdown	t	4
780	168	Quantity	number	t	5
781	169	Size	dropdown	t	1
782	169	Length	dropdown	t	2
783	169	Weight per Meter	number	t	3
784	169	Grade	dropdown	t	4
785	169	Quantity	number	t	5
786	170	Size	dropdown	t	1
787	170	Length	dropdown	t	2
788	170	Weight per Meter	number	t	3
789	170	Grade	dropdown	t	4
790	170	Quantity	number	t	5
791	171	Size	dropdown	t	1
792	171	Thickness	dropdown	t	2
793	171	Length	dropdown	t	3
794	171	Grade	dropdown	t	4
795	171	Quantity	number	t	5
796	172	Thickness	dropdown	t	1
797	172	Width	dropdown	t	2
798	172	Length	dropdown	t	3
799	172	Grade	dropdown	t	4
800	172	Quantity	number	t	5
801	173	Diameter	dropdown	t	1
802	173	Length	number	t	2
803	173	Pressure Class	dropdown	t	3
804	173	Quantity	number	t	4
805	174	Diameter	dropdown	t	1
806	174	Length	number	t	2
807	174	Pressure Class	dropdown	t	3
808	174	Quantity	number	t	4
809	175	Diameter	dropdown	t	1
810	175	Thickness	dropdown	t	2
811	175	Length	number	t	3
812	175	Grade	dropdown	t	4
813	175	Quantity	number	t	5
814	176	Type	dropdown	t	1
815	176	Size	dropdown	t	2
816	176	Material	dropdown	t	3
817	176	Quantity	number	t	4
818	177	Type	dropdown	t	1
819	177	Coverage Area	number	t	2
820	177	Packaging	dropdown	t	3
821	177	Quantity	number	t	4
822	178	Type	dropdown	t	1
823	178	Dosage	text	t	2
824	178	Packaging	dropdown	t	3
825	178	Quantity	number	t	4
826	179	Type	dropdown	t	1
827	179	Bag Weight	dropdown	t	2
828	179	Coverage Area	number	t	3
829	179	Quantity	number	t	4
830	180	Material	dropdown	t	1
831	180	Color	text	t	2
832	180	Dimensions	text	t	3
833	180	Upholstery Type	dropdown	t	4
834	180	Recliner	dropdown	t	5
835	180	Quantity	number	t	6
836	181	Material	dropdown	t	1
837	181	Color	text	t	2
838	181	Dimensions	text	t	3
839	181	Seating Capacity	dropdown	t	4
840	181	Upholstery	dropdown	t	5
841	181	Quantity	number	t	6
842	182	Material	dropdown	t	1
843	182	Recliner Type	dropdown	t	2
844	182	Seating Capacity	dropdown	t	3
845	182	Color	text	t	4
846	182	Quantity	number	t	5
847	183	Material	dropdown	t	1
848	183	Mattress Size	dropdown	t	2
849	183	Dimensions	text	t	3
850	183	Color	text	t	4
851	183	Quantity	number	t	5
852	184	Material	dropdown	t	1
853	184	Top Material	dropdown	t	2
854	184	Dimensions	text	t	3
855	184	Shape	dropdown	t	4
856	184	Quantity	number	t	5
857	185	Material	dropdown	t	1
858	185	Dimensions	text	t	2
859	185	Shape	dropdown	t	3
860	185	Storage	dropdown	t	4
861	185	Quantity	number	t	5
862	186	Material	dropdown	t	1
863	186	Dimensions	text	t	2
864	186	No. of Shelves	number	t	3
865	186	Mount Type	dropdown	t	4
866	186	Quantity	number	t	5
867	187	Size	dropdown	t	1
868	187	Material	dropdown	t	2
869	187	Headboard Type	dropdown	t	3
870	187	Storage	dropdown	t	4
871	187	Quantity	number	t	5
872	188	Size	dropdown	t	1
873	188	Material	dropdown	t	2
874	188	Headboard Type	dropdown	t	3
875	188	Storage	dropdown	t	4
876	188	Quantity	number	t	5
877	189	Material	dropdown	t	1
878	189	Headboard Type	dropdown	t	2
879	189	Storage	dropdown	t	3
880	189	Dimensions	text	t	4
881	189	Quantity	number	t	5
882	190	Material	dropdown	t	1
883	190	Headboard Type	dropdown	t	2
884	190	Storage	dropdown	t	3
885	190	Dimensions	text	t	4
886	190	Quantity	number	t	5
887	191	Material	dropdown	t	1
888	191	Color	text	t	2
889	191	Dimensions	text	t	3
890	191	Mirror	dropdown	t	4
891	191	Quantity	number	t	5
892	192	Material	dropdown	t	1
893	192	Mirror Size	dropdown	t	2
894	192	Drawers	number	t	3
895	192	Dimensions	text	t	4
896	192	Quantity	number	t	5
897	193	Material	dropdown	t	1
898	193	Drawers	number	t	2
899	193	Dimensions	text	t	3
900	193	Storage Type	dropdown	t	4
901	193	Quantity	number	t	5
902	194	Material	dropdown	t	1
903	194	Size	dropdown	t	2
904	194	Drawers	number	t	3
905	194	Shelf Type	dropdown	t	4
906	194	Quantity	number	t	5
907	195	Material	dropdown	t	1
908	195	Top Material	dropdown	t	2
909	195	Seating Capacity	dropdown	t	3
910	195	Dimensions	text	t	4
911	195	Quantity	number	t	5
912	196	Seating Capacity	dropdown	t	1
913	196	Table Material	dropdown	t	2
914	196	Chair Material	dropdown	t	3
915	196	Quantity	number	t	4
916	197	Material	dropdown	t	1
917	197	Dimensions	text	t	2
918	197	Door Type	dropdown	t	3
919	197	Finish	dropdown	t	4
920	197	Quantity	number	t	5
921	198	Material	dropdown	t	1
922	198	Wheels	dropdown	t	2
923	198	Shelves	number	t	3
924	198	Dimensions	text	t	4
925	198	Quantity	number	t	5
926	199	Size	dropdown	t	1
927	199	Material	dropdown	t	2
928	199	Top Material	dropdown	t	3
929	199	Storage	dropdown	t	4
930	199	Quantity	number	t	5
931	200	Type	dropdown	t	1
932	200	Material	dropdown	t	2
933	200	Adjustable Height	dropdown	t	3
934	200	Armrest Type	dropdown	t	4
935	200	Quantity	number	t	5
936	201	Material	dropdown	t	1
937	201	No. of Drawers	number	t	2
938	201	Dimensions	text	t	3
939	201	Lock Type	dropdown	t	4
940	201	Quantity	number	t	5
941	202	Type	dropdown	t	1
942	202	Material	dropdown	t	2
943	202	Shelves/Drawers	number	t	3
944	202	Dimensions	text	t	4
945	202	Quantity	number	t	5
946	203	Material	dropdown	t	1
947	203	Weather Resistance	dropdown	t	2
948	203	Seating Capacity	dropdown	t	3
949	203	Cushion Type	dropdown	t	4
950	203	Quantity	number	t	5
951	204	Material	dropdown	t	1
952	204	Weather Resistance	dropdown	t	2
953	204	Stackable	dropdown	t	3
954	204	Armrest	dropdown	t	4
955	204	Quantity	number	t	5
956	205	Material	dropdown	t	1
957	205	Top Material	dropdown	t	2
958	205	Seating Capacity	dropdown	t	3
959	205	Dimensions	text	t	4
960	205	Quantity	number	t	5
961	206	Material	dropdown	t	1
962	206	Size	dropdown	t	2
963	206	Weight Capacity	number	t	3
964	206	Color	text	t	4
965	206	Quantity	number	t	5
966	207	Table Material	dropdown	t	1
967	207	Chair Material	dropdown	t	2
968	207	Table Size	dropdown	t	3
969	207	Finish	dropdown	t	4
970	207	Quantity	number	t	5
971	208	Table Material	dropdown	t	1
972	208	Chair Material	dropdown	t	2
973	208	Table Size	dropdown	t	3
974	208	Finish	dropdown	t	4
975	208	Quantity	number	t	5
976	209	Size	dropdown	t	1
977	209	Material	dropdown	t	2
978	209	Seating Capacity	dropdown	t	3
979	209	Shape	dropdown	t	4
980	209	Quantity	number	t	5
981	210	Material	dropdown	t	1
982	210	Upholstery	dropdown	t	2
983	210	Armrest	dropdown	t	3
984	210	Stackable	dropdown	t	4
985	210	Quantity	number	t	5
986	211	Material	dropdown	t	1
987	211	Doors/Drawers	number	t	2
988	211	Dimensions	text	t	3
989	211	Finish	dropdown	t	4
990	211	Quantity	number	t	5
991	212	Size	dropdown	t	1
992	212	Thickness	dropdown	t	2
993	212	Firmness	dropdown	t	3
994	212	Material	dropdown	t	4
995	212	Quantity	number	t	5
996	213	Size	dropdown	t	1
997	213	Thickness	dropdown	t	2
998	213	Density	dropdown	t	3
999	213	Layers	number	t	4
1000	213	Quantity	number	t	5
1001	214	Size	dropdown	t	1
1002	214	Spring Count	number	t	2
1003	214	Thickness	dropdown	t	3
1004	214	Firmness	dropdown	t	4
1005	214	Quantity	number	t	5
1006	215	Thickness	dropdown	t	1
1007	215	Type	dropdown	t	2
1008	215	Firmness	dropdown	t	3
1009	215	Quantity	number	t	4
1010	216	Thickness	dropdown	t	1
1011	216	Type	dropdown	t	2
1012	216	Firmness	dropdown	t	3
1013	216	Quantity	number	t	4
1014	217	Thickness	dropdown	t	1
1015	217	Type	dropdown	t	2
1016	217	Firmness	dropdown	t	3
1017	217	Quantity	number	t	4
1018	218	Material	dropdown	t	1
1019	218	Seating Capacity	dropdown	t	2
1020	218	Dimensions	text	t	3
1021	218	Color	text	t	4
1022	218	Quantity	number	t	5
1023	219	Material	dropdown	t	1
1024	219	Seating Capacity	dropdown	t	2
1025	219	Dimensions	text	t	3
1026	219	Color	text	t	4
1027	219	Quantity	number	t	5
1028	220	Recliner Type	dropdown	t	1
1029	220	Material	dropdown	t	2
1030	220	Seating Capacity	dropdown	t	3
1031	220	Motorized	dropdown	t	4
1032	220	Quantity	number	t	5
1033	221	Material	dropdown	t	1
1034	221	Configuration	dropdown	t	2
1035	221	Seating Capacity	dropdown	t	3
1036	221	Dimensions	text	t	4
1037	221	Color	text	t	5
1038	221	Quantity	number	t	6
1039	222	Size	dropdown	t	1
1040	222	Material	dropdown	t	2
1041	222	Headboard Type	dropdown	t	3
1042	222	Storage	dropdown	t	4
1043	222	Quantity	number	t	5
1044	223	Size	dropdown	t	1
1045	223	Material	dropdown	t	2
1046	223	Headboard Type	dropdown	t	3
1047	223	Storage	dropdown	t	4
1048	223	Quantity	number	t	5
1049	224	Material	dropdown	t	1
1050	224	Headboard Type	dropdown	t	2
1051	224	Storage	dropdown	t	3
1052	224	Dimensions	text	t	4
1053	224	Quantity	number	t	5
1054	225	Material	dropdown	t	1
1055	225	Headboard Type	dropdown	t	2
1056	225	Storage	dropdown	t	3
1057	225	Dimensions	text	t	4
1058	225	Quantity	number	t	5
1059	226	Bunk Type	dropdown	t	1
1060	226	Material	dropdown	t	2
1061	226	Safety Rails	dropdown	t	3
1062	226	Ladder Type	dropdown	t	4
1063	226	Quantity	number	t	5
1064	227	Material	dropdown	t	1
1065	227	Dimensions	text	t	2
1066	227	Mirror	dropdown	t	3
1067	227	Internal Layout	dropdown	t	4
1068	227	Quantity	number	t	5
1069	228	Material	dropdown	t	1
1070	228	Dimensions	text	t	2
1071	228	Mirror	dropdown	t	3
1072	228	Internal Layout	dropdown	t	4
1073	228	Quantity	number	t	5
1074	229	Material	dropdown	t	1
1075	229	Dimensions	text	t	2
1076	229	Mirror	dropdown	t	3
1077	229	Internal Layout	dropdown	t	4
1078	229	Quantity	number	t	5
1079	230	Material	dropdown	t	1
1080	230	Door Mechanism	dropdown	t	2
1081	230	Dimensions	text	t	3
1082	230	Mirror Doors	dropdown	t	4
1083	230	Quantity	number	t	5
1084	231	Material	dropdown	t	1
1085	231	Top Material	dropdown	t	2
1086	231	Dimensions	text	t	3
1087	231	Shape	dropdown	t	4
1088	231	Storage	dropdown	t	5
1089	231	Quantity	number	t	6
1090	232	Material	dropdown	t	1
1091	232	Top Material	dropdown	t	2
1092	232	Dimensions	text	t	3
1093	232	Shape	dropdown	t	4
1094	232	Quantity	number	t	5
1095	233	Size	dropdown	t	1
1096	233	Seating Capacity	dropdown	t	2
1097	233	Material	dropdown	t	3
1098	233	Shape	dropdown	t	4
1099	233	Extendable	dropdown	t	5
1100	233	Quantity	number	t	6
1101	234	Material	dropdown	t	1
1102	234	Size	dropdown	t	2
1103	234	Drawers	number	t	3
1104	234	Shelf Type	dropdown	t	4
1105	234	Quantity	number	t	5
1106	235	Material	dropdown	t	1
1107	235	Upholstery	dropdown	t	2
1108	235	Back Type	dropdown	t	3
1109	235	Color	text	t	4
1110	235	Quantity	number	t	5
1111	236	Material	dropdown	t	1
1112	236	Type	dropdown	t	2
1113	236	Adjustable Height	dropdown	t	3
1114	236	Armrest	dropdown	t	4
1115	236	Quantity	number	t	5
1116	237	Material	dropdown	t	1
1117	237	Mechanism	dropdown	t	2
1118	237	Color	text	t	3
1119	237	Quantity	number	t	4
1120	238	Material	dropdown	t	1
1121	238	Finish	dropdown	t	2
1122	238	Cushion Included	dropdown	t	3
1123	238	Quantity	number	t	4
1124	239	Material	dropdown	t	1
1125	239	Capacity (Pairs)	number	t	2
1126	239	Dimensions	text	t	3
1127	239	Door Type	dropdown	t	4
1128	239	Quantity	number	t	5
1129	240	Material	dropdown	t	1
1130	240	Dimensions	text	t	2
1131	240	Finish	dropdown	t	3
1132	240	Door Type	dropdown	t	4
1133	240	Quantity	number	t	5
1134	241	Material	dropdown	t	1
1135	241	Dimensions	text	t	2
1136	241	No. of Shelves	number	t	3
1137	241	Door Type	dropdown	t	4
1138	241	Quantity	number	t	5
1139	242	Material	dropdown	t	1
1140	242	TV Size Compatibility	dropdown	t	2
1141	242	Dimensions	text	t	3
1142	242	Storage	dropdown	t	4
1143	242	Quantity	number	t	5
1144	243	Material	dropdown	t	1
1145	243	TV Size Compatibility	dropdown	t	2
1146	243	Dimensions	text	t	3
1147	243	Storage	dropdown	t	4
1148	243	Quantity	number	t	5
1149	244	Material	dropdown	t	1
1150	244	Dimensions	text	t	2
1151	244	Features	text	t	3
1152	244	Color	text	t	4
1153	244	Quantity	number	t	5
1154	245	Material	dropdown	t	1
1155	245	Dimensions	text	t	2
1156	245	Weight Capacity	number	t	3
1157	245	Mount Type	dropdown	t	4
1158	245	Quantity	number	t	5
1159	246	Material	dropdown	t	1
1160	246	No. of Shelves	number	t	2
1161	246	Dimensions	text	t	3
1162	246	Open/Closed	dropdown	t	4
1163	246	Quantity	number	t	5
1164	247	Material	dropdown	t	1
1165	247	Dimensions	text	t	2
1166	247	No. of Tiers	number	t	3
1167	247	Mount Type	dropdown	t	4
1168	247	Quantity	number	t	5
1169	248	Material	dropdown	t	1
1170	248	Dimensions	text	t	2
1171	248	Weight Capacity	number	t	3
1172	248	Color	text	t	4
1173	248	Quantity	number	t	5
1174	249	Model Name	text	t	1
1175	249	RAM	dropdown	t	2
1176	249	Storage	dropdown	t	3
1177	249	Color	text	t	4
1178	249	Quantity	number	t	5
1179	250	Model Name	text	t	1
1180	250	Storage	dropdown	t	2
1181	250	Color	text	t	3
1182	250	Connectivity	dropdown	t	4
1183	250	Quantity	number	t	5
1184	251	Model Name	text	t	1
1185	251	RAM	dropdown	t	2
1186	251	Storage	dropdown	t	3
1187	251	Color	text	t	4
1188	251	Quantity	number	t	5
1189	252	Model Name	text	t	1
1190	252	RAM	dropdown	t	2
1191	252	Storage	dropdown	t	3
1192	252	Color	text	t	4
1193	252	Quantity	number	t	5
1194	253	Model Name	text	t	1
1195	253	RAM	dropdown	t	2
1196	253	Storage	dropdown	t	3
1197	253	Color	text	t	4
1198	253	Quantity	number	t	5
1199	254	Model Name	text	t	1
1200	254	RAM	dropdown	t	2
1201	254	Storage	dropdown	t	3
1202	254	Color	text	t	4
1203	254	Quantity	number	t	5
1204	255	Model Name	text	t	1
1205	255	RAM	dropdown	t	2
1206	255	Storage	dropdown	t	3
1207	255	Color	text	t	4
1208	255	Quantity	number	t	5
1209	256	Model Name	text	t	1
1210	256	RAM	dropdown	t	2
1211	256	Storage	dropdown	t	3
1212	256	Screen Size	dropdown	t	4
1213	256	Quantity	number	t	5
1214	257	Model Name	text	t	1
1215	257	Storage	dropdown	t	2
1216	257	Screen Size	dropdown	t	3
1217	257	Connectivity	dropdown	t	4
1218	257	Quantity	number	t	5
1219	258	Model Name	text	t	1
1220	258	RAM	dropdown	t	2
1221	258	Storage	dropdown	t	3
1222	258	Screen Size	dropdown	t	4
1223	258	Quantity	number	t	5
1224	259	Brand	text	t	1
1225	259	RAM	dropdown	t	2
1226	259	Storage	dropdown	t	3
1227	259	Connectivity	dropdown	t	4
1228	259	Quantity	number	t	5
1229	260	Processor	dropdown	t	1
1230	260	RAM	dropdown	t	2
1231	260	Storage	dropdown	t	3
1232	260	Graphics Card	dropdown	t	4
1233	260	Screen Size	dropdown	t	5
1234	260	Operating System	dropdown	t	6
1235	260	Quantity	number	t	7
1236	261	Processor	dropdown	t	1
1237	261	RAM	dropdown	t	2
1238	261	Storage	dropdown	t	3
1239	261	Screen Size	dropdown	t	4
1240	261	Operating System	dropdown	t	5
1241	261	Quantity	number	t	6
1242	262	Processor	dropdown	t	1
1243	262	RAM	dropdown	t	2
1244	262	Storage	dropdown	t	3
1245	262	Screen Size	dropdown	t	4
1246	262	Weight	number	t	5
1247	262	Operating System	dropdown	t	6
1248	262	Quantity	number	t	7
1249	263	Processor	dropdown	t	1
1250	263	RAM	dropdown	t	2
1251	263	Storage	dropdown	t	3
1252	263	Screen Size	dropdown	t	4
1253	263	Operating System	dropdown	t	5
1254	263	Quantity	number	t	6
1255	264	Processor	dropdown	t	1
1256	264	RAM	dropdown	t	2
1257	264	Storage	dropdown	t	3
1258	264	Screen Size	dropdown	t	4
1259	264	Operating System	dropdown	t	5
1260	264	Quantity	number	t	6
1261	265	Processor	dropdown	t	1
1262	265	RAM	dropdown	t	2
1263	265	Storage	dropdown	t	3
1264	265	Screen Size	dropdown	t	4
1265	265	Operating System	dropdown	t	5
1266	265	Quantity	number	t	6
1267	266	Processor	dropdown	t	1
1268	266	RAM	dropdown	t	2
1269	266	Storage	dropdown	t	3
1270	266	Graphics Card	dropdown	t	4
1271	266	Operating System	dropdown	t	5
1272	266	Quantity	number	t	6
1273	267	Processor	dropdown	t	1
1274	267	RAM	dropdown	t	2
1275	267	Storage	dropdown	t	3
1276	267	Operating System	dropdown	t	4
1277	267	Quantity	number	t	5
1278	268	Processor	dropdown	t	1
1279	268	RAM	dropdown	t	2
1280	268	Storage	dropdown	t	3
1281	268	Screen Size	dropdown	t	4
1282	268	Operating System	dropdown	t	5
1283	268	Quantity	number	t	6
1284	269	Processor	dropdown	t	1
1285	269	RAM	dropdown	t	2
1286	269	Storage	dropdown	t	3
1287	269	Graphics Card	dropdown	t	4
1288	269	Operating System	dropdown	t	5
1289	269	Quantity	number	t	6
1290	270	Screen Size	dropdown	t	1
1291	270	Resolution	dropdown	t	2
1292	270	Smart TV	dropdown	t	3
1293	270	Panel Type	dropdown	t	4
1294	270	HDMI Ports	number	t	5
1295	270	Quantity	number	t	6
1296	271	Screen Size	dropdown	t	1
1297	271	Resolution	dropdown	t	2
1298	271	Smart Platform	dropdown	t	3
1299	271	Panel Type	dropdown	t	4
1300	271	HDMI Ports	number	t	5
1301	271	Quantity	number	t	6
1302	272	Screen Size	dropdown	t	1
1303	272	Resolution	dropdown	t	2
1304	272	Smart TV	dropdown	t	3
1305	272	Panel Type	dropdown	t	4
1306	272	HDMI Ports	number	t	5
1307	272	Quantity	number	t	6
1308	273	Resolution	dropdown	t	1
1309	273	Smart TV	dropdown	t	2
1310	273	Panel Type	dropdown	t	3
1311	273	HDMI Ports	number	t	4
1312	273	Quantity	number	t	5
1313	274	Resolution	dropdown	t	1
1314	274	Smart TV	dropdown	t	2
1315	274	Panel Type	dropdown	t	3
1316	274	HDMI Ports	number	t	4
1317	274	Quantity	number	t	5
1318	275	Resolution	dropdown	t	1
1319	275	Smart TV	dropdown	t	2
1320	275	Panel Type	dropdown	t	3
1321	275	HDMI Ports	number	t	4
1322	275	Quantity	number	t	5
1323	276	Screen Size	dropdown	t	1
1324	276	Resolution	dropdown	t	2
1325	276	Smart TV	dropdown	t	3
1326	276	Panel Type	dropdown	t	4
1327	276	HDMI Ports	number	t	5
1328	276	Quantity	number	t	6
1329	277	AC Type	dropdown	t	1
1330	277	Star Rating	dropdown	t	2
1331	277	Inverter	dropdown	t	3
1332	277	Cooling Capacity	dropdown	t	4
1333	277	Quantity	number	t	5
1334	278	AC Type	dropdown	t	1
1335	278	Star Rating	dropdown	t	2
1336	278	Inverter	dropdown	t	3
1337	278	Cooling Capacity	dropdown	t	4
1338	278	Quantity	number	t	5
1339	279	Capacity	dropdown	t	1
1340	279	Door Type	dropdown	t	2
1341	279	Star Rating	dropdown	t	3
1342	279	Defrost Type	dropdown	t	4
1343	279	Quantity	number	t	5
1344	280	Machine Type	dropdown	t	1
1345	280	Capacity	dropdown	t	2
1346	280	Star Rating	dropdown	t	3
1347	280	Wash Technology	dropdown	t	4
1348	280	Quantity	number	t	5
1349	281	Purifier Type	dropdown	t	1
1350	281	Stages of Purification	number	t	2
1351	281	Storage Capacity	dropdown	t	3
1352	281	Mount Type	dropdown	t	4
1353	281	Quantity	number	t	5
1354	282	Oven Type	dropdown	t	1
1355	282	Capacity	dropdown	t	2
1356	282	Power	number	t	3
1357	282	Control Type	dropdown	t	4
1358	282	Quantity	number	t	5
1359	283	Power	number	t	1
1360	283	Preset Menus	number	t	2
1361	283	Control Type	dropdown	t	3
1362	283	Body Material	dropdown	t	4
1363	283	Quantity	number	t	5
1364	284	Power	number	t	1
1365	284	No. of Jars	number	t	2
1366	284	Jar Material	dropdown	t	3
1367	284	Speed Settings	number	t	4
1368	284	Quantity	number	t	5
1369	285	Juicer Type	dropdown	t	1
1370	285	Power	number	t	2
1371	285	Jar Capacity	dropdown	t	3
1372	285	Body Material	dropdown	t	4
1373	285	Quantity	number	t	5
1374	286	Capacity	dropdown	t	1
1375	286	Power	number	t	2
1376	286	Functions	text	t	3
1377	286	Control Type	dropdown	t	4
1378	286	Quantity	number	t	5
1379	287	Power Output	number	t	1
1380	287	Channels	dropdown	t	2
1381	287	Connectivity	dropdown	t	3
1382	287	Subwoofer Included	dropdown	t	4
1383	287	Quantity	number	t	5
1384	288	Power Output	number	t	1
1385	288	Channels	dropdown	t	2
1386	288	Connectivity	dropdown	t	3
1387	288	Speaker Configuration	text	t	4
1388	288	Quantity	number	t	5
1389	289	Power Output	number	t	1
1390	289	Battery Backup	dropdown	t	2
1391	289	Water Resistance	dropdown	t	3
1392	289	Connectivity	dropdown	t	4
1393	289	Quantity	number	t	5
1394	290	Sensor Resolution	dropdown	t	1
1395	290	Lens Mount	dropdown	t	2
1396	290	Video Resolution	dropdown	t	3
1397	290	Card Slots	number	t	4
1398	290	Quantity	number	t	5
1399	291	Sensor Resolution	dropdown	t	1
1400	291	Lens Mount	dropdown	t	2
1401	291	Video Resolution	dropdown	t	3
1402	291	Stabilization	dropdown	t	4
1403	291	Quantity	number	t	5
1404	292	Sensor Resolution	dropdown	t	1
1405	292	Optical Zoom	dropdown	t	2
1406	292	Video Resolution	dropdown	t	3
1407	292	Battery Type	dropdown	t	4
1408	292	Quantity	number	t	5
1409	293	Camera Type	dropdown	t	1
1410	293	Resolution	dropdown	t	2
1411	293	Lens Type	dropdown	t	3
1412	293	Night Vision	dropdown	t	4
1413	293	Quantity	number	t	5
1414	294	Edition	dropdown	t	1
1415	294	Storage	dropdown	t	2
1416	294	Region	dropdown	t	3
1417	294	Quantity	number	t	4
1418	295	Edition	dropdown	t	1
1419	295	Storage	dropdown	t	2
1420	295	Region	dropdown	t	3
1421	295	Quantity	number	t	4
1422	296	Model	dropdown	t	1
1423	296	Storage	dropdown	t	2
1424	296	Region	dropdown	t	3
1425	296	Quantity	number	t	4
1426	297	Accessory Type	dropdown	t	1
1427	297	Platform	dropdown	t	2
1428	297	Wired/Wireless	dropdown	t	3
1429	297	Quantity	number	t	4
1430	298	Phone Model	text	t	1
1431	298	Material	dropdown	t	2
1432	298	Cover Type	dropdown	t	3
1433	298	Color	text	t	4
1434	298	Quantity	number	t	5
1435	299	Charger Type	dropdown	t	1
1436	299	Output Power	dropdown	t	2
1437	299	Connector Type	dropdown	t	3
1438	299	Cable Included	dropdown	t	4
1439	299	Quantity	number	t	5
1440	300	Phone Model	text	t	1
1441	300	Material	dropdown	t	2
1442	300	Guard Type	dropdown	t	3
1443	300	Pack Size	number	t	4
1444	300	Quantity	number	t	5
1445	300	Phone Model	text	t	1
1446	300	Material	dropdown	t	2
1447	300	Guard Type	dropdown	t	3
1448	300	Pack Size	number	t	4
1449	300	Quantity	number	t	5
1450	301	Type	dropdown	t	1
1451	301	Driver Size	dropdown	t	2
1452	301	Connectivity	dropdown	t	3
1453	301	Battery Backup	dropdown	t	4
1454	301	Quantity	number	t	5
1455	302	WiFi Standard	dropdown	t	1
1456	302	Speed	dropdown	t	2
1457	302	Antennas	number	t	3
1458	302	Ports	number	t	4
1459	302	Quantity	number	t	5
1460	303	WiFi Standard	dropdown	t	1
1461	303	Coverage Area	number	t	2
1462	303	Speed	dropdown	t	3
1463	303	Ports	number	t	4
1464	303	Quantity	number	t	5
1465	304	Category	dropdown	t	1
1466	304	Length	dropdown	t	2
1467	304	Shielding	dropdown	t	3
1468	304	Connector Type	dropdown	t	4
1469	304	Quantity	number	t	5
1470	305	Ports	number	t	1
1471	305	Speed	dropdown	t	2
1472	305	Managed/Unmanaged	dropdown	t	3
1473	305	Power over Ethernet	dropdown	t	4
1474	305	Quantity	number	t	5
1475	306	Connectivity	dropdown	t	1
1476	306	Wattage	dropdown	t	2
1477	306	Color Temperature	dropdown	t	3
1478	306	Compatible With	dropdown	t	4
1479	306	Quantity	number	t	5
1480	307	Connectivity	dropdown	t	1
1481	307	Max Load	dropdown	t	2
1482	307	Compatible With	dropdown	t	3
1483	307	Energy Monitoring	dropdown	t	4
1484	307	Quantity	number	t	5
1485	308	Resolution	dropdown	t	1
1486	308	Connectivity	dropdown	t	2
1487	308	Night Vision	dropdown	t	3
1488	308	Motion Detection	dropdown	t	4
1489	308	Quantity	number	t	5
1490	309	Lock Type	dropdown	t	1
1491	309	Connectivity	dropdown	t	2
1492	309	Unlock Methods	text	t	3
1493	309	Battery Backup	dropdown	t	4
1494	309	Quantity	number	t	5
1495	310	Brand	text	t	1
1496	310	Display Size	dropdown	t	2
1497	310	Battery Backup	dropdown	t	3
1498	310	Compatible Phone	dropdown	t	4
1499	310	Quantity	number	t	5
1500	311	Brand	text	t	1
1501	311	Display Type	dropdown	t	2
1502	311	Battery Backup	dropdown	t	3
1503	311	Water Resistance	dropdown	t	4
1504	311	Quantity	number	t	5
1505	312	Brand	text	t	1
1506	312	Size	dropdown	t	2
1507	312	Battery Backup	dropdown	t	3
1508	312	Features	text	t	4
1509	312	Quantity	number	t	5
1510	313	Print Technology	dropdown	t	1
1511	313	Print Speed (B&W)	dropdown	t	2
1512	313	Print Speed (Color)	dropdown	t	3
1513	313	Connectivity	dropdown	t	4
1514	313	Quantity	number	t	5
1515	314	Print Technology	dropdown	t	1
1516	314	Print Speed	dropdown	t	2
1517	314	Print Resolution	dropdown	t	3
1518	314	Duplex Printing	dropdown	t	4
1519	314	Quantity	number	t	5
1520	315	Functions	text	t	1
1521	315	Print Speed	dropdown	t	2
1522	315	Scan Resolution	dropdown	t	3
1523	315	Connectivity	dropdown	t	4
1524	315	Quantity	number	t	5
1525	316	Scanner Type	dropdown	t	1
1526	316	Scan Resolution	dropdown	t	2
1527	316	Scan Speed	dropdown	t	3
1528	316	Connectivity	dropdown	t	4
1529	316	Quantity	number	t	5
1530	317	Capacity	dropdown	t	1
1531	317	Read Speed	dropdown	t	2
1532	317	Write Speed	dropdown	t	3
1533	317	Connector Type	dropdown	t	4
1534	317	Quantity	number	t	5
1535	318	Capacity	dropdown	t	1
1536	318	Read Speed	dropdown	t	2
1537	318	Write Speed	dropdown	t	3
1538	318	Connector Type	dropdown	t	4
1539	318	Quantity	number	t	5
1540	319	Capacity	dropdown	t	1
1541	319	Interface	dropdown	t	2
1542	319	Rotation Speed	dropdown	t	3
1543	319	Transfer Rate	dropdown	t	4
1544	319	Quantity	number	t	5
1545	320	Capacity	dropdown	t	1
1546	320	Interface	dropdown	t	2
1547	320	Read Speed	dropdown	t	3
1548	320	Write Speed	dropdown	t	4
1549	320	Quantity	number	t	5
1550	321	Power Output	number	t	1
1551	321	Battery Backup	dropdown	t	2
1552	321	Water Resistance	dropdown	t	3
1553	321	Bluetooth Version	dropdown	t	4
1554	321	Quantity	number	t	5
1555	322	Power Output	number	t	1
1556	322	Battery Backup	dropdown	t	2
1557	322	Driver Size	dropdown	t	3
1558	322	Connectivity	dropdown	t	4
1559	322	Quantity	number	t	5
1560	323	Power Output	number	t	1
1561	323	Driver Configuration	text	t	2
1562	323	Frequency Response	text	t	3
1563	323	Impedance	dropdown	t	4
1564	323	Quantity	number	t	5
1565	324	Driver Size	dropdown	t	1
1566	324	Frequency Response	text	t	2
1567	324	Impedance	dropdown	t	3
1568	324	Connector Type	dropdown	t	4
1569	324	Quantity	number	t	5
1570	325	Driver Size	dropdown	t	1
1571	325	Bluetooth Version	dropdown	t	2
1572	325	Battery Backup	dropdown	t	3
1573	325	Active Noise Cancellation	dropdown	t	4
1574	325	Quantity	number	t	5
1575	326	Driver Size	dropdown	t	1
1576	326	Microphone Type	dropdown	t	2
1577	326	Connectivity	dropdown	t	3
1578	326	Surround Sound	dropdown	t	4
1579	326	Quantity	number	t	5
1580	327	Capacity	dropdown	t	1
1581	327	Output Ports	number	t	2
1582	327	Fast Charging	dropdown	t	3
1583	327	Charging Cable Included	dropdown	t	4
1584	327	Quantity	number	t	5
1585	328	Capacity	dropdown	t	1
1586	328	Output Ports	number	t	2
1587	328	Fast Charging	dropdown	t	3
1588	328	Charging Cable Included	dropdown	t	4
1589	328	Quantity	number	t	5
1590	329	Capacity	dropdown	t	1
1591	329	Wireless Charging Standard	dropdown	t	2
1592	329	Fast Charging	dropdown	t	3
1593	329	Output Ports	number	t	4
1594	329	Quantity	number	t	5
1595	330	Blade Size	dropdown	t	1
1596	330	Power Consumption	number	t	2
1597	330	Speed Settings	number	t	3
1598	330	Material	dropdown	t	4
1599	330	Quantity	number	t	5
1600	331	Blade Size	dropdown	t	1
1601	331	Power Consumption	number	t	2
1602	331	Speed Settings	number	t	3
1603	331	Oscillation	dropdown	t	4
1604	331	Quantity	number	t	5
1605	332	Blade Size	dropdown	t	1
1606	332	Power Consumption	number	t	2
1607	332	Speed Settings	number	t	3
1608	332	Height Adjustable	dropdown	t	4
1609	332	Quantity	number	t	5
1610	333	Cooling Capacity	dropdown	t	1
1611	333	Tank Capacity	dropdown	t	2
1612	333	Cooling Pads	dropdown	t	3
1613	333	Air Throw	number	t	4
1614	333	Quantity	number	t	5
1615	334	Brand	text	t	1
1616	334	Pack Size	dropdown	t	2
1617	334	Grade	dropdown	t	3
1618	334	Quantity	number	t	4
1619	335	Brand	text	t	1
1620	335	Pack Size	dropdown	t	2
1621	335	Grade	dropdown	t	3
1622	335	Quantity	number	t	4
1623	336	Brand	text	t	1
1624	336	Pack Size	dropdown	t	2
1625	336	Type	dropdown	t	3
1848	392	Brand	text	t	1
1626	336	Quantity	number	t	4
1627	337	Brand	text	t	1
1628	337	Pack Size	dropdown	t	2
1629	337	Type	dropdown	t	3
1630	337	Quantity	number	t	4
1631	338	Oil Type	dropdown	t	1
1632	338	Brand	text	t	2
1633	338	Pack Size	dropdown	t	3
1634	338	Quantity	number	t	4
1635	339	Sugar Type	dropdown	t	1
1636	339	Brand	text	t	2
1637	339	Pack Size	dropdown	t	3
1638	339	Quantity	number	t	4
1639	340	Salt Type	dropdown	t	1
1640	340	Brand	text	t	2
1641	340	Pack Size	dropdown	t	3
1642	340	Quantity	number	t	4
1643	341	Brand	text	t	1
1644	341	Pack Size	dropdown	t	2
1645	341	Flavor	dropdown	t	3
1646	341	Quantity	number	t	4
1647	342	Brand	text	t	1
1648	342	Pack Size	dropdown	t	2
1649	342	Flavor	dropdown	t	3
1650	342	Quantity	number	t	4
1651	343	Brand	text	t	1
1652	343	Pack Size	dropdown	t	2
1653	343	Type	dropdown	t	3
1654	343	Quantity	number	t	4
1655	344	Brand	text	t	1
1656	344	Pack Size	dropdown	t	2
1657	344	Type	dropdown	t	3
1658	344	Quantity	number	t	4
1659	345	Brand	text	t	1
1660	345	Pack Size	dropdown	t	2
1661	345	Flavor	dropdown	t	3
1662	345	Quantity	number	t	4
1663	346	Brand	text	t	1
1664	346	Pack Size	dropdown	t	2
1665	346	Flavor	dropdown	t	3
1666	346	Quantity	number	t	4
1667	347	Brand	text	t	1
1668	347	Pack Size	dropdown	t	2
1669	347	Flavor	dropdown	t	3
1670	347	Quantity	number	t	4
1671	348	Brand	text	t	1
1672	348	Pack Size	dropdown	t	2
1673	348	Flavor	dropdown	t	3
1674	348	Quantity	number	t	4
1675	349	Brand	text	t	1
1676	349	Pack Size	dropdown	t	2
1677	349	Bottle Size	dropdown	t	3
1678	349	Quantity	number	t	4
1679	350	Brand	text	t	1
1680	350	Pack Size	dropdown	t	2
1681	350	Type	dropdown	t	3
1682	350	Quantity	number	t	4
1683	351	Brand	text	t	1
1684	351	Pack Size	dropdown	t	2
1685	351	Type	dropdown	t	3
1686	351	Quantity	number	t	4
1687	352	Brand	text	t	1
1688	352	Pack Size	dropdown	t	2
1689	352	Type	dropdown	t	3
1690	352	Quantity	number	t	4
1691	353	Brand	text	t	1
1692	353	Pack Size	dropdown	t	2
1693	353	Type	dropdown	t	3
1694	353	Quantity	number	t	4
1695	354	Brand	text	t	1
1696	354	Pack Size	dropdown	t	2
1697	354	Variant	dropdown	t	3
1698	354	Quantity	number	t	4
1699	355	Brand	text	t	1
1700	355	Pack Size	dropdown	t	2
1701	355	Variant	dropdown	t	3
1702	355	Quantity	number	t	4
1703	356	Brand	text	t	1
1704	356	Pack Size	dropdown	t	2
1705	356	Fragrance	dropdown	t	3
1706	356	Quantity	number	t	4
1707	357	Brand	text	t	1
1708	357	Pack Size	dropdown	t	2
1709	357	Type	dropdown	t	3
1710	357	Quantity	number	t	4
1711	358	Brand	text	t	1
1712	358	Pack Size	dropdown	t	2
1713	358	Type	dropdown	t	3
1714	358	Quantity	number	t	4
1715	359	Brand	text	t	1
1716	359	Pack Size	dropdown	t	2
1717	359	Purity	dropdown	t	3
1718	359	Quantity	number	t	4
1719	360	Brand	text	t	1
1720	360	Pack Size	dropdown	t	2
1721	360	Type	dropdown	t	3
1722	360	Quantity	number	t	4
1723	361	Brand	text	t	1
1724	361	Pack Size	dropdown	t	2
1725	361	Fat Content	dropdown	t	3
1726	361	Quantity	number	t	4
1727	362	Brand	text	t	1
1728	362	Pack Size	dropdown	t	2
1729	362	Type	dropdown	t	3
1730	362	Quantity	number	t	4
1731	363	Brand	text	t	1
1732	363	Pack Size	dropdown	t	2
1733	363	Type	dropdown	t	3
1734	363	Quantity	number	t	4
1735	364	Brand	text	t	1
1736	364	Pack Size	dropdown	t	2
1737	364	Pieces Per Pack	number	t	3
1738	364	Quantity	number	t	4
1739	365	Brand	text	t	1
1740	365	Pack Size	dropdown	t	2
1741	365	Type	dropdown	t	3
1742	365	Quantity	number	t	4
1743	366	Brand	text	t	1
1744	366	Pack Size	dropdown	t	2
1745	366	Flavor	dropdown	t	3
1746	366	Quantity	number	t	4
1747	367	Brand	text	t	1
1748	367	Pack Size	dropdown	t	2
1749	367	Type	dropdown	t	3
1750	367	Quantity	number	t	4
1751	368	Brand	text	t	1
1752	368	Pack Size	dropdown	t	2
1753	368	Type	dropdown	t	3
1754	368	Quantity	number	t	4
1755	369	Brand	text	t	1
1756	369	Pack Size	dropdown	t	2
1757	369	Type	dropdown	t	3
1758	369	Quantity	number	t	4
1759	370	Brand	text	t	1
1760	370	Pack Size	dropdown	t	2
1761	370	Type	dropdown	t	3
1762	370	Quantity	number	t	4
1763	371	Brand	text	t	1
1764	371	Pack Size	dropdown	t	2
1765	371	Type	dropdown	t	3
1766	371	Quantity	number	t	4
1767	372	Brand	text	t	1
1768	372	Pack Size	dropdown	t	2
1769	372	Oil Type	dropdown	t	3
1770	372	Quantity	number	t	4
1771	373	Brand	text	t	1
1772	373	Pack Size	dropdown	t	2
1773	373	Oil Type	dropdown	t	3
1774	373	Quantity	number	t	4
1775	374	Brand	text	t	1
1776	374	Pack Size	dropdown	t	2
1777	374	Oil Type	dropdown	t	3
1778	374	Quantity	number	t	4
1779	375	Brand	text	t	1
1780	375	Pack Size	dropdown	t	2
1781	375	Oil Type	dropdown	t	3
1782	375	Quantity	number	t	4
1783	376	Brand	text	t	1
1784	376	Pack Size	dropdown	t	2
1785	376	Variant	dropdown	t	3
1786	376	Quantity	number	t	4
1787	377	Brand	text	t	1
1788	377	Pack Size	dropdown	t	2
1789	377	Variant	dropdown	t	3
1790	377	Quantity	number	t	4
1791	378	Brand	text	t	1
1792	378	Pack Size	dropdown	t	2
1793	378	Variant	dropdown	t	3
1794	378	Quantity	number	t	4
1795	379	Brand	text	t	1
1796	379	Pack Size	dropdown	t	2
1797	379	Variant	dropdown	t	3
1798	379	Quantity	number	t	4
1799	380	Brand	text	t	1
1800	380	Size	dropdown	t	2
1801	380	Pack Size	dropdown	t	3
1802	380	Quantity	number	t	4
1803	381	Brand	text	t	1
1804	381	Pack Size	dropdown	t	2
1805	381	Variant	dropdown	t	3
1806	381	Quantity	number	t	4
1807	382	Brand	text	t	1
1808	382	Flavor	dropdown	t	2
1809	382	Pack Size	dropdown	t	3
1810	382	Quantity	number	t	4
1811	383	Brand	text	t	1
1812	383	Flavor	dropdown	t	2
1813	383	Pack Size	dropdown	t	3
1814	383	Protein Type	dropdown	t	4
1815	383	Quantity	number	t	5
1816	384	Brand	text	t	1
1817	384	Type	dropdown	t	2
1818	384	Pack Size	dropdown	t	3
1819	384	Quantity	number	t	4
1820	385	Brand	text	t	1
1821	385	Type	dropdown	t	2
1822	385	Pack Size	dropdown	t	3
1823	385	Quantity	number	t	4
1824	386	Brand	text	t	1
1825	386	Flavor	dropdown	t	2
1826	386	Pack Size	dropdown	t	3
1827	386	Quantity	number	t	4
1828	387	Cut Type	dropdown	t	1
1829	387	Pack Weight	dropdown	t	2
1830	387	Type	dropdown	t	3
1831	387	Quantity	number	t	4
1832	388	Vegetable Type	dropdown	t	1
1833	388	Pack Weight	dropdown	t	2
1834	388	Mixed/Single	dropdown	t	3
1835	388	Quantity	number	t	4
1836	389	Brand	text	t	1
1837	389	Flavor	dropdown	t	2
1838	389	Pack Size	dropdown	t	3
1839	389	Quantity	number	t	4
1840	390	Brand	text	t	1
1841	390	Flavor	dropdown	t	2
1842	390	Pack Size	dropdown	t	3
1843	390	Quantity	number	t	4
1844	391	Brand	text	t	1
1845	391	Pasta Type	dropdown	t	2
1846	391	Pack Weight	dropdown	t	3
1847	391	Quantity	number	t	4
1849	392	Type	dropdown	t	2
1850	392	Pack Size	dropdown	t	3
1851	392	Quantity	number	t	4
1852	393	Brand	text	t	1
1853	393	Dal Type	dropdown	t	2
1854	393	Pack Size	dropdown	t	3
1855	393	Quantity	number	t	4
1856	394	Brand	text	t	1
1857	394	Type	dropdown	t	2
1858	394	Pack Size	dropdown	t	3
1859	394	Quantity	number	t	4
1860	395	Variety	dropdown	t	1
1861	395	Pack Weight	dropdown	t	2
1862	395	Grade	dropdown	t	3
1863	395	Quantity	number	t	4
1864	396	Variety	dropdown	t	1
1865	396	Pack Weight	dropdown	t	2
1866	396	Grade	dropdown	t	3
1867	396	Quantity	number	t	4
1868	397	Variety	dropdown	t	1
1869	397	Pack Weight	dropdown	t	2
1870	397	Grade	dropdown	t	3
1871	397	Quantity	number	t	4
1872	398	Fruit Type	dropdown	t	1
1873	398	Pack Weight	dropdown	t	2
1874	398	Grade	dropdown	t	3
1875	398	Quantity	number	t	4
1876	399	Brand	text	t	1
1877	399	Pack Weight	dropdown	t	2
1878	399	Type	dropdown	t	3
1879	399	Quantity	number	t	4
1880	400	Brand	text	t	1
1881	400	Type	dropdown	t	2
1882	400	Pack Weight	dropdown	t	3
1883	400	Quantity	number	t	4
1884	401	Brand	text	t	1
1885	401	Type	dropdown	t	2
1886	401	Pack Weight	dropdown	t	3
1887	401	Quantity	number	t	4
1888	402	Brand	text	t	1
1889	402	Sweet Type	dropdown	t	2
1890	402	Pack Weight	dropdown	t	3
1891	402	Quantity	number	t	4
1892	403	Brand	text	t	1
1893	403	Food Type	dropdown	t	2
1894	403	Pack Weight	dropdown	t	3
1895	403	Quantity	number	t	4
1896	404	Brand	text	t	1
1897	404	Food Type	dropdown	t	2
1898	404	Pack Weight	dropdown	t	3
1899	404	Quantity	number	t	4
1900	405	Toy Type	dropdown	t	1
1901	405	Pet Category	dropdown	t	2
1902	405	Material	dropdown	t	3
1903	405	Quantity	number	t	4
1904	406	Brand	text	t	1
1905	406	Pen Type	dropdown	t	2
1906	406	Pack Count	dropdown	t	3
1907	406	Quantity	number	t	4
1908	407	Brand	text	t	1
1909	407	Pen Type	dropdown	t	2
1910	407	Pack Count	dropdown	t	3
1911	407	Quantity	number	t	4
1912	408	Brand	text	t	1
1913	408	GSM	dropdown	t	2
1914	408	Sheets Per Pack	dropdown	t	3
1915	408	Quantity	number	t	4
1916	409	Brand	text	t	1
1917	409	Type	dropdown	t	2
1918	409	Capacity	dropdown	t	3
1919	409	Quantity	number	t	4
1920	410	Size	dropdown	t	1
1921	410	Pack Count	dropdown	t	2
1922	410	Material	dropdown	t	3
1923	410	Quantity	number	t	4
1924	411	Brand	text	t	1
1925	411	Grade	dropdown	t	2
1926	411	Pack Count	dropdown	t	3
1927	411	Quantity	number	t	4
1928	412	Brand	text	t	1
1929	412	Set Type	dropdown	t	2
1930	412	Items Count	number	t	3
1931	412	Quantity	number	t	4
1932	413	Brand	text	t	1
1933	413	Pages	dropdown	t	2
1934	413	Ruling	dropdown	t	3
1935	413	Quantity	number	t	4
1936	414	Brand	text	t	1
1937	414	Type	dropdown	t	2
1938	414	Pack Count	dropdown	t	3
1939	414	Quantity	number	t	4
1940	415	Brand	text	t	1
1941	415	Type	dropdown	t	2
1942	415	Holes	dropdown	t	3
1943	415	Quantity	number	t	4
1944	416	Brand	text	t	1
1945	416	Pack Count	dropdown	t	2
1946	416	Shade Type	dropdown	t	3
1947	416	Quantity	number	t	4
1948	417	Brand	text	t	1
1949	417	Color Count	dropdown	t	2
1950	417	Volume Per Bottle	dropdown	t	3
1951	417	Quantity	number	t	4
1952	418	Brand	text	t	1
1953	418	Color Count	dropdown	t	2
1954	418	Tip Type	dropdown	t	3
1955	418	Quantity	number	t	4
1956	419	Brand	text	t	1
1957	419	Pages	dropdown	t	2
1958	419	Paper GSM	dropdown	t	3
1959	419	Quantity	number	t	4
1960	420	Size	dropdown	t	1
1961	420	Pack Count	dropdown	t	2
1962	420	GSM	dropdown	t	3
1963	420	Quantity	number	t	4
1964	441	Printer Model	text	t	1
1965	441	Color	dropdown	t	2
1966	441	Page Yield	dropdown	t	3
1967	441	Quantity	number	t	4
1968	442	Size	dropdown	t	1
1969	442	GSM	dropdown	t	2
1970	442	Sheets Per Pack	dropdown	t	3
1971	442	Quantity	number	t	4
1972	443	Size	dropdown	t	1
1973	443	Material	dropdown	t	2
1974	443	Angles	text	t	3
1975	443	Quantity	number	t	4
1976	444	Size	dropdown	t	1
1977	444	Material	dropdown	t	2
1978	444	Accuracy	dropdown	t	3
1979	444	Quantity	number	t	4
1980	445	Size	dropdown	t	1
1981	445	Material	dropdown	t	2
1982	445	Surface Type	dropdown	t	3
1983	445	Quantity	number	t	4
1984	446	Length	dropdown	t	1
1985	446	Material	dropdown	t	2
1986	446	Blade Width	dropdown	t	3
1987	446	Quantity	number	t	4
1988	447	Color	dropdown	t	1
1989	447	Size	dropdown	t	2
1990	447	GSM	dropdown	t	3
1991	447	Quantity	number	t	4
1992	448	Brand	text	t	1
1993	448	Size	dropdown	t	2
1994	448	Pack Count	dropdown	t	3
1995	448	Quantity	number	t	4
1996	449	Blade Length	dropdown	t	1
1997	449	Material	dropdown	t	2
1998	449	Safety Type	dropdown	t	3
1999	449	Quantity	number	t	4
2000	450	Tip Type	dropdown	t	1
2001	450	Color Count	dropdown	t	2
2002	450	Pack Count	dropdown	t	3
2003	450	Quantity	number	t	4
2004	451	Material	dropdown	t	1
2005	451	Size	dropdown	t	2
2006	451	Magnetic	dropdown	t	3
2007	451	Quantity	number	t	4
2008	452	Size	dropdown	t	1
2009	452	Frame Material	dropdown	t	2
2010	452	Mount Type	dropdown	t	3
2011	452	Quantity	number	t	4
2012	453	Brand	text	t	1
2013	453	Display Digits	dropdown	t	2
2014	453	Power Source	dropdown	t	3
2015	453	Quantity	number	t	4
2016	454	Brand	text	t	1
2017	454	Display Digits	dropdown	t	2
2018	454	Functions	dropdown	t	3
2019	454	Quantity	number	t	4
2020	455	Type	dropdown	t	1
2021	455	Range	dropdown	t	2
2022	455	Laser Color	dropdown	t	3
2023	455	Quantity	number	t	4
2024	456	Size	dropdown	t	1
2025	456	Pages	dropdown	t	2
2026	456	Paper GSM	dropdown	t	3
2027	456	Quantity	number	t	4
2028	457	Size	dropdown	t	1
2029	457	Surface Type	dropdown	t	2
2030	457	Frame Material	dropdown	t	3
2031	457	Quantity	number	t	4
2032	458	Brand	text	t	1
2033	458	Tip Size	dropdown	t	2
2034	458	Dry Time	dropdown	t	3
2035	458	Quantity	number	t	4
2036	459	Brand	text	t	1
2037	459	Tape Width	dropdown	t	2
2038	459	Length	dropdown	t	3
2039	459	Quantity	number	t	4
2040	460	Brand	text	t	1
2041	460	Type	dropdown	t	2
2042	460	Pack Count	dropdown	t	3
2043	460	Quantity	number	t	4
2044	461	Brand	text	t	1
2045	461	Size	dropdown	t	2
2046	461	Pack Count	dropdown	t	3
2047	461	Color	dropdown	t	4
2048	461	Quantity	number	t	5
2049	462	Width	dropdown	t	1
2050	462	Length	dropdown	t	2
2051	462	Core Size	dropdown	t	3
2052	462	Quantity	number	t	4
2053	463	Width	dropdown	t	1
2054	463	Length	dropdown	t	2
2055	463	Material	dropdown	t	3
2056	463	Quantity	number	t	4
2057	464	Size	dropdown	t	1
2058	464	Pack Count	dropdown	t	2
2059	464	GSM	dropdown	t	3
2060	464	Quantity	number	t	4
2061	465	Size	dropdown	t	1
2062	465	Pack Count	dropdown	t	2
2063	465	GSM	dropdown	t	3
2064	465	Quantity	number	t	4
2065	466	Label Size	dropdown	t	1
2066	466	Sheets Per Pack	dropdown	t	2
2067	466	Labels Per Sheet	number	t	3
2068	466	Quantity	number	t	4
2069	467	Size	dropdown	t	1
2070	467	Fabric	dropdown	t	2
2071	467	Color	text	t	3
2072	467	Sleeve Length	dropdown	t	4
2073	467	Fit Type	dropdown	t	5
2074	467	Quantity	number	t	6
2075	468	Size	dropdown	t	1
2076	468	Fabric	dropdown	t	2
2077	468	Color	text	t	3
2078	468	Sleeve Length	dropdown	t	4
2079	468	Pattern	dropdown	t	5
2080	468	Quantity	number	t	6
2081	469	Size	dropdown	t	1
2082	469	Fabric	dropdown	t	2
2083	469	Color	text	t	3
2084	469	Neck Type	dropdown	t	4
2085	469	Sleeve Type	dropdown	t	5
2086	469	Quantity	number	t	6
2087	470	Size	dropdown	t	1
2088	470	Fabric	dropdown	t	2
2089	470	Color	text	t	3
2090	470	Collar Type	dropdown	t	4
2091	470	Sleeve Length	dropdown	t	5
2092	470	Quantity	number	t	6
2093	471	Size	dropdown	t	1
2094	471	Fabric	dropdown	t	2
2095	471	Color	text	t	3
2096	471	Work Type	dropdown	t	4
2097	471	Fit Type	dropdown	t	5
2098	471	Quantity	number	t	6
2099	472	Size	dropdown	t	1
2100	472	Waist Size	dropdown	t	2
2101	472	Fabric	dropdown	t	3
2102	472	Color	text	t	4
2103	472	Fit Type	dropdown	t	5
2104	472	Quantity	number	t	6
2105	473	Size	dropdown	t	1
2106	473	Waist Size	dropdown	t	2
2107	473	Length	dropdown	t	3
2108	473	Wash Type	dropdown	t	4
2109	473	Fit Type	dropdown	t	5
2110	473	Quantity	number	t	6
2111	474	Size	dropdown	t	1
2112	474	Waist Size	dropdown	t	2
2113	474	Fabric	dropdown	t	3
2114	474	Color	text	t	4
2115	474	Fit Type	dropdown	t	5
2116	474	Quantity	number	t	6
2117	475	Size	dropdown	t	1
2118	475	Waist Size	dropdown	t	2
2119	475	Fabric	dropdown	t	3
2120	475	Color	text	t	4
2121	475	Pockets	dropdown	t	5
2122	475	Quantity	number	t	6
2123	476	Size	dropdown	t	1
2124	476	Waist Size	dropdown	t	2
2125	476	Fabric	dropdown	t	3
2126	476	Color	text	t	4
2127	476	Drawstring	dropdown	t	5
2128	476	Quantity	number	t	6
2129	477	Size	dropdown	t	1
2130	477	Fabric	dropdown	t	2
2131	477	Color	text	t	3
2132	477	Pattern	dropdown	t	4
2133	477	Fit Type	dropdown	t	5
2134	477	Quantity	number	t	6
2135	478	Size	dropdown	t	1
2136	478	Fabric	dropdown	t	2
2137	478	Color	text	t	3
2138	478	Work Type	dropdown	t	4
2139	478	Fit Type	dropdown	t	5
2140	478	Quantity	number	t	6
2141	479	Pieces	dropdown	t	1
2142	479	Size	dropdown	t	2
2143	479	Fabric	dropdown	t	3
2144	479	Color	text	t	4
2145	479	Quantity	number	t	5
2146	480	Size	dropdown	t	1
2147	480	Material	dropdown	t	2
2148	480	Color	text	t	3
2149	480	Buckle Type	dropdown	t	4
2150	480	Quantity	number	t	5
2151	481	Type	dropdown	t	1
2152	481	Material	dropdown	t	2
2153	481	Color	text	t	3
2154	481	Pattern	dropdown	t	4
2155	481	Quantity	number	t	5
2156	482	Material	dropdown	t	1
2157	482	Pack Size	dropdown	t	2
2158	482	Color	text	t	3
2159	482	Quantity	number	t	4
2160	483	Material	dropdown	t	1
2161	483	Color	text	t	2
2162	483	Slots Count	number	t	3
2163	483	Size	dropdown	t	4
2164	483	Quantity	number	t	5
2165	484	Size	dropdown	t	1
2166	484	Fabric	dropdown	t	2
2167	484	Color	text	t	3
2168	484	Pattern	dropdown	t	4
2169	484	Neck Type	dropdown	t	5
2170	484	Quantity	number	t	6
2171	485	Size	dropdown	t	1
2172	485	Fabric	dropdown	t	2
2173	485	Color	text	t	3
2174	485	Work Type	dropdown	t	4
2175	485	Sleeve Length	dropdown	t	5
2176	485	Quantity	number	t	6
2177	486	Size	dropdown	t	1
2178	486	Fabric	dropdown	t	2
2179	486	Color	text	t	3
2180	486	Pattern	dropdown	t	4
2181	486	Length	dropdown	t	5
2182	486	Quantity	number	t	6
2183	487	Size	dropdown	t	1
2184	487	Fabric	dropdown	t	2
2185	487	Color	text	t	3
2186	487	Neck Type	dropdown	t	4
2187	487	Sleeve Type	dropdown	t	5
2188	487	Quantity	number	t	6
2189	488	Size	dropdown	t	1
2190	488	Fabric	dropdown	t	2
2191	488	Color	text	t	3
2192	488	Neck Type	dropdown	t	4
2193	488	Sleeve Type	dropdown	t	5
2194	488	Quantity	number	t	6
2195	489	Size	dropdown	t	1
2196	489	Fabric	dropdown	t	2
2197	489	Color	text	t	3
2198	489	Strap Type	dropdown	t	4
2199	489	Fit Type	dropdown	t	5
2200	489	Quantity	number	t	6
2201	490	Size	dropdown	t	1
2202	490	Fabric	dropdown	t	2
2203	490	Color	text	t	3
2204	490	Sleeve Length	dropdown	t	4
2205	490	Neck Type	dropdown	t	5
2206	490	Quantity	number	t	6
2207	491	Size	dropdown	t	1
2208	491	Fabric	dropdown	t	2
2209	491	Color	text	t	3
2210	491	Length	dropdown	t	4
2211	491	Waist Type	dropdown	t	5
2212	491	Quantity	number	t	6
2213	492	Size	dropdown	t	1
2214	492	Fabric	dropdown	t	2
2215	492	Color	text	t	3
2216	492	Length	dropdown	t	4
2217	492	Waist Type	dropdown	t	5
2218	492	Quantity	number	t	6
2219	493	Size	dropdown	t	1
2220	493	Fabric	dropdown	t	2
2221	493	Color	text	t	3
2222	493	Pattern	dropdown	t	4
2223	493	Waist Type	dropdown	t	5
2224	493	Quantity	number	t	6
2225	494	Size	dropdown	t	1
2226	494	Fabric	dropdown	t	2
2227	494	Color	text	t	3
2228	494	Length	dropdown	t	4
2229	494	Fit Type	dropdown	t	5
2230	494	Quantity	number	t	6
2231	495	Size	dropdown	t	1
2232	495	Cup Size	dropdown	t	2
2233	495	Material	dropdown	t	3
2234	495	Type	dropdown	t	4
2235	495	Quantity	number	t	5
2236	496	Size	dropdown	t	1
2237	496	Material	dropdown	t	2
2238	496	Type	dropdown	t	3
2239	496	Pack Size	dropdown	t	4
2240	496	Quantity	number	t	5
2241	497	Size	dropdown	t	1
2242	497	Fabric	dropdown	t	2
2243	497	Color	text	t	3
2244	497	Type	dropdown	t	4
2245	497	Quantity	number	t	5
2246	498	Size	dropdown	t	1
2247	498	Fabric	dropdown	t	2
2248	498	Color	text	t	3
2249	498	Sleeve Length	dropdown	t	4
2250	498	Quantity	number	t	5
2251	499	Size	dropdown	t	1
2252	499	Fabric	dropdown	t	2
2253	499	Color	text	t	3
2254	499	Pattern	dropdown	t	4
2255	499	Quantity	number	t	5
2256	500	Size	dropdown	t	1
2257	500	Fabric	dropdown	t	2
2258	500	Color	text	t	3
2259	500	Neck Type	dropdown	t	4
2260	500	Sleeve Type	dropdown	t	5
2261	500	Quantity	number	t	6
2262	501	Size	dropdown	t	1
2263	501	Waist Size	dropdown	t	2
2264	501	Fabric	dropdown	t	3
2265	501	Color	text	t	4
2266	501	Fit Type	dropdown	t	5
2267	501	Quantity	number	t	6
2268	502	Size	dropdown	t	1
2269	502	Fabric	dropdown	t	2
2270	502	Color	text	t	3
2271	502	Sleeve Length	dropdown	t	4
2272	502	Pattern	dropdown	t	5
2273	502	Quantity	number	t	6
2274	503	Size	dropdown	t	1
2275	503	Fabric	dropdown	t	2
2276	503	Color	text	t	3
2277	503	Sleeve Type	dropdown	t	4
2278	503	Neck Type	dropdown	t	5
2279	503	Quantity	number	t	6
2280	504	Size	dropdown	t	1
2281	504	Waist Size	dropdown	t	2
2282	504	Fabric	dropdown	t	3
2283	504	Color	text	t	4
2284	504	Length	dropdown	t	5
2285	504	Quantity	number	t	6
2286	505	Size	dropdown	t	1
2287	505	Fabric	dropdown	t	2
2288	505	Color	text	t	3
2289	505	Moisture Wicking	dropdown	t	4
2290	505	Quantity	number	t	5
2291	506	Size	dropdown	t	1
2292	506	Waist Size	dropdown	t	2
2293	506	Fabric	dropdown	t	3
2294	506	Color	text	t	4
2295	506	Quantity	number	t	5
2296	507	Size	dropdown	t	1
2297	507	Cup Size	dropdown	t	2
2298	507	Support Level	dropdown	t	3
2299	507	Fabric	dropdown	t	4
2300	507	Quantity	number	t	5
2301	508	Size	dropdown	t	1
2302	508	Fabric	dropdown	t	2
2303	508	Color	text	t	3
2304	508	Type	dropdown	t	4
2305	508	Waterproof	dropdown	t	5
2306	508	Quantity	number	t	6
2307	509	Size	dropdown	t	1
2308	509	Fabric	dropdown	t	2
2309	509	Color	text	t	3
2310	509	Fit Type	dropdown	t	4
2311	509	Pattern	dropdown	t	5
2312	509	Quantity	number	t	6
2313	510	Size	dropdown	t	1
2314	510	Fabric	dropdown	t	2
2315	510	Color	text	t	3
2316	510	Length	dropdown	t	4
2317	510	Lining	dropdown	t	5
2318	510	Quantity	number	t	6
2319	511	Size	dropdown	t	1
2320	511	Fabric	dropdown	t	2
2321	511	Color	text	t	3
2322	511	Type	dropdown	t	4
2323	511	Quantity	number	t	5
2324	512	Size	dropdown	t	1
2325	512	Fabric	dropdown	t	2
2326	512	Color	text	t	3
2327	512	Compression	dropdown	t	4
2328	512	Quantity	number	t	5
2329	513	Size	dropdown	t	1
2330	513	Fabric	dropdown	t	2
2331	513	Color	text	t	3
2332	513	Type	dropdown	t	4
2333	513	Quantity	number	t	5
2334	514	Size	dropdown	t	1
2335	514	Waist Size	dropdown	t	2
2336	514	Fabric	dropdown	t	3
2337	514	Color	text	t	4
2338	514	Quantity	number	t	5
2339	515	Size	dropdown	t	1
2340	515	Fabric	dropdown	t	2
2341	515	Color	text	t	3
2342	515	Work Type	dropdown	t	4
2343	515	Blouse Piece	dropdown	t	5
2344	515	Quantity	number	t	6
2345	516	Size	dropdown	t	1
2346	516	Fabric	dropdown	t	2
2347	516	Color	text	t	3
2348	516	Work Type	dropdown	t	4
2349	516	Dupatta Included	dropdown	t	5
2350	516	Quantity	number	t	6
2351	517	Size	dropdown	t	1
2352	517	Fabric	dropdown	t	2
2353	517	Color	text	t	3
2354	517	Work Type	dropdown	t	4
2355	517	Dupatta Included	dropdown	t	5
2356	517	Quantity	number	t	6
2357	518	Size	dropdown	t	1
2358	518	Fabric	dropdown	t	2
2359	518	Color	text	t	3
2360	518	Work Type	dropdown	t	4
2361	518	Fit Type	dropdown	t	5
2362	518	Quantity	number	t	6
2363	519	Size	dropdown	t	1
2364	519	Fabric	dropdown	t	2
2365	519	Color	text	t	3
2366	519	Sleeve Length	dropdown	t	4
2367	519	Fit Type	dropdown	t	5
2368	519	Quantity	number	t	6
2369	520	Size	dropdown	t	1
2370	520	Waist Size	dropdown	t	2
2371	520	Fabric	dropdown	t	3
2372	520	Color	text	t	4
2373	520	Fit Type	dropdown	t	5
2374	520	Quantity	number	t	6
2375	521	Size	dropdown	t	1
2376	521	Fabric	dropdown	t	2
2377	521	Color	text	t	3
2378	521	Fit Type	dropdown	t	4
2379	521	Pattern	dropdown	t	5
2380	521	Quantity	number	t	6
2381	522	Size	dropdown	t	1
2382	522	Fabric	dropdown	t	2
2383	522	Color	text	t	3
2384	522	Sleeve Length	dropdown	t	4
2385	522	Pattern	dropdown	t	5
2386	522	Quantity	number	t	6
2387	523	Size	dropdown	t	1
2388	523	Fabric	dropdown	t	2
2389	523	Color	text	t	3
2390	523	Neck Type	dropdown	t	4
2391	523	Sleeve Type	dropdown	t	5
2392	523	Quantity	number	t	6
2393	524	Size	dropdown	t	1
2394	524	Waist Size	dropdown	t	2
2395	524	Length	dropdown	t	3
2396	524	Wash Type	dropdown	t	4
2397	524	Fit Type	dropdown	t	5
2398	524	Quantity	number	t	6
2399	525	Size	dropdown	t	1
2400	525	Waist Size	dropdown	t	2
2401	525	Fabric	dropdown	t	3
2402	525	Color	text	t	4
2403	525	Length	dropdown	t	5
2404	525	Quantity	number	t	6
2405	526	Size	dropdown	t	1
2406	526	Fabric	dropdown	t	2
2407	526	Color	text	t	3
2408	526	Length	dropdown	t	4
2409	526	Sleeve Type	dropdown	t	5
2410	526	Quantity	number	t	6
2411	527	Size	dropdown	t	1
2412	527	Fabric	dropdown	t	2
2413	527	Color	text	t	3
2414	527	Type	dropdown	t	4
2415	527	Quantity	number	t	5
2416	528	Size	dropdown	t	1
2417	528	Fabric	dropdown	t	2
2418	528	Color	text	t	3
2419	528	Type	dropdown	t	4
2420	528	Quantity	number	t	5
2421	529	Size	dropdown	t	1
2422	529	Fabric	dropdown	t	2
2423	529	Color	text	t	3
2424	529	Top Type	dropdown	t	4
2425	529	Quantity	number	t	5
2426	530	Size	dropdown	t	1
2427	530	Fabric	dropdown	t	2
2428	530	Color	text	t	3
2429	530	Neck Type	dropdown	t	4
2430	530	Sleeve Length	dropdown	t	5
2431	530	Quantity	number	t	6
2432	531	Size	dropdown	t	1
2433	531	Fabric	dropdown	t	2
2434	531	Color	text	t	3
2435	531	Type	dropdown	t	4
2436	531	Waterproof	dropdown	t	5
2437	531	Quantity	number	t	6
2438	532	Size	dropdown	t	1
2439	532	Fabric	dropdown	t	2
2440	532	Color	text	t	3
2441	532	Pattern	dropdown	t	4
2442	532	Quantity	number	t	5
2443	533	Size	dropdown	t	1
2444	533	Color	text	t	2
2445	533	Material	dropdown	t	3
2446	533	Type	dropdown	t	4
2447	533	Quantity	number	t	5
2448	534	Size	dropdown	t	1
2449	534	Color	text	t	2
2450	534	Material	dropdown	t	3
2451	534	Type	dropdown	t	4
2452	534	Quantity	number	t	5
2453	535	Size	dropdown	t	1
2454	535	Color	text	t	2
2455	535	Type	dropdown	t	3
2456	535	Cushioning	dropdown	t	4
2457	535	Quantity	number	t	5
2458	536	Size	dropdown	t	1
2459	536	Color	text	t	2
2460	536	Material	dropdown	t	3
2461	536	Strap Type	dropdown	t	4
2462	536	Quantity	number	t	5
2463	537	Size	dropdown	t	1
2464	537	Heel Height	dropdown	t	2
2465	537	Color	text	t	3
2466	537	Material	dropdown	t	4
2467	537	Quantity	number	t	5
2468	538	Size	dropdown	t	1
2469	538	Color	text	t	2
2470	538	Type	dropdown	t	3
2471	538	Material	dropdown	t	4
2472	538	Quantity	number	t	5
2473	539	Size	dropdown	t	1
2474	539	Material	dropdown	t	2
2475	539	Color	text	t	3
2476	539	Compartments	number	t	4
2477	539	Quantity	number	t	5
2478	540	Size	dropdown	t	1
2479	540	Material	dropdown	t	2
2480	540	Color	text	t	3
2481	540	Type	dropdown	t	4
2482	540	Quantity	number	t	5
2483	541	Size	dropdown	t	1
2484	541	Material	dropdown	t	2
2485	541	Color	text	t	3
2486	541	Handle Type	dropdown	t	4
2487	541	Quantity	number	t	5
2488	542	Type	dropdown	t	1
2489	542	Purity	dropdown	t	2
2490	542	Weight	number	t	3
2491	542	Design	text	t	4
2492	542	Quantity	number	t	5
2493	543	Type	dropdown	t	1
2494	543	Purity	dropdown	t	2
2495	543	Weight	number	t	3
2496	543	Design	text	t	4
2497	543	Quantity	number	t	5
2498	544	Type	dropdown	t	1
2499	544	Material	dropdown	t	2
2500	544	Color	text	t	3
2501	544	Design	text	t	4
2502	544	Quantity	number	t	5
2503	545	Frame Material	dropdown	t	1
2504	545	Lens Color	dropdown	t	2
2505	545	Frame Size	dropdown	t	3
2506	545	Polarization	dropdown	t	4
2507	545	Quantity	number	t	5
2508	546	Frame Material	dropdown	t	1
2509	546	Lens Color	dropdown	t	2
2510	546	Frame Size	dropdown	t	3
2511	546	Style	dropdown	t	4
2512	546	Quantity	number	t	5
2513	547	Dial Size	dropdown	t	1
2514	547	Strap Material	dropdown	t	2
2515	547	Movement Type	dropdown	t	3
2516	547	Water Resistance	dropdown	t	4
2517	547	Quantity	number	t	5
2518	548	Dial Size	dropdown	t	1
2519	548	Strap Material	dropdown	t	2
2520	548	Movement Type	dropdown	t	3
2521	548	Style	dropdown	t	4
2522	548	Quantity	number	t	5
2523	549	Brand	text	t	1
2524	549	Display Size	dropdown	t	2
2525	549	Battery Backup	dropdown	t	3
2526	549	Compatible Phone	dropdown	t	4
2527	549	Quantity	number	t	5
2528	550	Size	dropdown	t	1
2529	550	Fabric	dropdown	t	2
2530	550	Pack Size	dropdown	t	3
2531	550	Quantity	number	t	4
2532	551	Size	dropdown	t	1
2533	551	Fabric	dropdown	t	2
2534	551	Pack Size	dropdown	t	3
2535	551	Waistband Type	dropdown	t	4
2536	551	Quantity	number	t	5
2537	552	Type	dropdown	t	1
2538	552	Size	dropdown	t	2
2539	552	Fabric	dropdown	t	3
2540	552	Pack Size	dropdown	t	4
2541	552	Quantity	number	t	5
2542	553	Size	dropdown	t	1
2543	553	Fabric	dropdown	t	2
2544	553	Color	text	t	3
2545	553	Sleeve Length	dropdown	t	4
2546	553	Quantity	number	t	5
2547	554	Size	dropdown	t	1
2548	554	Waist Size	dropdown	t	2
2549	554	Fabric	dropdown	t	3
2550	554	Length	dropdown	t	4
2551	554	Quantity	number	t	5
2552	555	Fabric	dropdown	t	1
2553	555	Work Type	dropdown	t	2
2554	555	Color	text	t	3
2555	555	Blouse Piece	dropdown	t	4
2556	555	Quantity	number	t	5
2557	556	Fabric	dropdown	t	1
2558	556	Work Type	dropdown	t	2
2559	556	Color	text	t	3
2560	556	Dupatta Included	dropdown	t	4
2561	556	Quantity	number	t	5
2562	557	Fabric	dropdown	t	1
2563	557	Work Type	dropdown	t	2
2564	557	Color	text	t	3
2565	557	Size	dropdown	t	4
2566	557	Quantity	number	t	5
2567	558	Fabric	dropdown	t	1
2568	558	Work Type	dropdown	t	2
2569	558	Color	text	t	3
2570	558	Blouse Piece	dropdown	t	4
2571	558	Quantity	number	t	5
2572	559	Fabric	dropdown	t	1
2573	559	Work Type	dropdown	t	2
2574	559	Color	text	t	3
2575	559	Dupatta Included	dropdown	t	4
2576	559	Quantity	number	t	5
2578	560	Fee per day	number	t	2
2579	560	Fee per hour	number	t	3
2580	560	Booking Fee	number	t	4
2577	560	Visiting/Inspection Fee	number	t	1
2581	561	Inspection Fee	number	t	1
2582	561	Fee per day	number	t	2
2583	561	Fee per hour	number	t	3
2584	561	Booking Fee	number	t	4
2585	562	Inspection Fee	number	t	1
2586	562	Fee per day	number	t	2
2587	562	Fee per hour	number	t	3
2588	562	Booking Fee	number	t	4
2589	563	Inspection Fee	number	t	1
2590	563	Fee per day	number	t	2
2591	563	Fee per hour	number	t	3
2592	563	Booking Fee	number	t	4
2593	564	Inspection Fee	number	t	1
2594	564	Fee per day	number	t	2
2595	564	Fee per hour	number	t	3
2596	564	Booking Fee	number	t	4
2597	565	Inspection Fee	number	t	1
2598	565	Fee per day	number	t	2
2599	565	Fee per hour	number	t	3
2600	565	Booking Fee	number	t	4
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.transactions (transaction_id, wallet_id, transaction_date, amount, transaction_type) FROM stdin;
\.


--
-- Data for Name: user_addresses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_addresses (id, user_id, address_id, is_default) FROM stdin;
\.


--
-- Data for Name: user_refresh_sessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_refresh_sessions (token_id, user_id, admin_id, user_type, refresh_token, expire_at, is_blocked) FROM stdin;
6401c1a8-7848-480f-85ec-2c5c21c3080c	9	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiNjQwMWMxYTgtNzg0OC00ODBmLTg1ZWMtMmM1YzIxYzMwODBjIiwiVXNlcklEIjoiOSIsIkV4cGlyZXNBdCI6IjIwMjUtMTItMjhUMTU6MTE6NTQuOTc2MzY1M1oiLCJVc2VkRm9yIjoidXNlciJ9.ibp2IGRpal4L567JfKnxBSVl3l9AI3WZogczG3vaQlk	2025-12-28 15:11:54+00	f
58b889bf-8620-49b2-ab51-1b51ddcc5fce	10	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiNThiODg5YmYtODYyMC00OWIyLWFiNTEtMWI1MWRkY2M1ZmNlIiwiVXNlcklEIjoiMTAiLCJFeHBpcmVzQXQiOiIyMDI1LTEyLTI4VDE2OjM2OjMyLjczODgxOTRaIiwiVXNlZEZvciI6InVzZXIifQ.VAnInmn5mrHYHNALyOdos6jCUASOa_UAS1j0NNZ7DaE	2025-12-28 16:36:32+00	f
969e4f32-b457-4213-9ac1-c7363dd701e3	11	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiOTY5ZTRmMzItYjQ1Ny00MjEzLTlhYzEtYzczNjNkZDcwMWUzIiwiVXNlcklEIjoiMTEiLCJFeHBpcmVzQXQiOiIyMDI1LTEyLTI4VDE3OjEzOjI0LjA0NjQzODFaIiwiVXNlZEZvciI6InVzZXIifQ.7o2_D4HWTLmh3cMF3w5ZcUGIjxAnYHgwqpLVgn7PJl8	2025-12-28 17:13:24+00	f
759f3995-31e6-4210-8141-63b05be4feff	12	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiNzU5ZjM5OTUtMzFlNi00MjEwLTgxNDEtNjNiMDViZTRmZWZmIiwiVXNlcklEIjoiMTIiLCJFeHBpcmVzQXQiOiIyMDI1LTEyLTI4VDE3OjM4OjExLjI5OTA3NTlaIiwiVXNlZEZvciI6InVzZXIifQ.96FUlQVe9HAX7Qub3GGACp05CO3Hg_buzIyCFUYoZcc	2025-12-28 17:38:11+00	f
3ad87a55-6ebb-41ec-a94f-0385f1a3aa75	13	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiM2FkODdhNTUtNmViYi00MWVjLWE5NGYtMDM4NWYxYTNhYTc1IiwiVXNlcklEIjoiMTMiLCJFeHBpcmVzQXQiOiIyMDI1LTEyLTI4VDE5OjExOjI1LjcyNTgxMzVaIiwiVXNlZEZvciI6InVzZXIifQ.XBwIqYfv9QmXMKxVBTIlnkHp1ZQHm0GRRgDC7Tpt1UA	2025-12-28 19:11:25+00	f
dcc9ae41-414b-4a66-acfa-a715c987c239	15	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiZGNjOWFlNDEtNDE0Yi00YTY2LWFjZmEtYTcxNWM5ODdjMjM5IiwiVXNlcklEIjoiMTUiLCJFeHBpcmVzQXQiOiIyMDI2LTAxLTA2VDA4OjU2OjI2LjI4MTUyMzlaIiwiVXNlZEZvciI6InVzZXIifQ.2KwfO0OzFXJgzFGZAptuoK69lKVoLE-uv0qaF2E1W1w	2026-01-06 08:56:26+00	f
c8ed3e5d-7682-4197-aef9-089431b83239	15	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYzhlZDNlNWQtNzY4Mi00MTk3LWFlZjktMDg5NDMxYjgzMjM5IiwiVXNlcklEIjoiMTUiLCJFeHBpcmVzQXQiOiIyMDI2LTAxLTEwVDA4OjAxOjEyLjAwMTA0MloiLCJVc2VkRm9yIjoidXNlciJ9.1_HrkRtekAeV2JWcIEBq8RS7VKha5sjL5cIqZ0JA2UE	2026-01-10 08:01:12+00	f
bfa5eb20-3cad-4bf4-bb5b-3616b0820ba0	15	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYmZhNWViMjAtM2NhZC00YmY0LWJiNWItMzYxNmIwODIwYmEwIiwiVXNlcklEIjoiMTUiLCJFeHBpcmVzQXQiOiIyMDI2LTAxLTI2VDIzOjU3OjQyLjY0MjU3NzUrMDU6MzAiLCJVc2VkRm9yIjoidXNlciJ9.nYd9sUxhWfRORhpc_cxzT2OFsDA-SMIFZA_ISUv0KKI	2026-01-26 18:27:42+00	f
f4539655-c9d3-4d23-9c78-ae1e2974d5d9	15	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiZjQ1Mzk2NTUtYzlkMy00ZDIzLTljNzgtYWUxZTI5NzRkNWQ5IiwiVXNlcklEIjoiMTUiLCJFeHBpcmVzQXQiOiIyMDI2LTAxLTI4VDE2OjE4OjIyLjg5NjIzMTcrMDU6MzAiLCJVc2VkRm9yIjoidXNlciJ9.QJY-j7Fus0fv5m7H6ow-wrzJoyRmNj_hqocT_V6O2AI	2026-01-28 10:48:22+00	f
7d3109bc-91d0-4767-9fcd-67ff9f14710b	6	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiN2QzMTA5YmMtOTFkMC00NzY3LTlmY2QtNjdmZjlmMTQ3MTBiIiwiVXNlcklEIjoiNiIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMDhUMTg6NTI6MDMuMjU3MzM0NCswNTozMCIsIlVzZWRGb3IiOiJ1c2VyIn0.oJ7ovD6viZUyIzCKVPe2oXefLEAawAU7B2R_ShBXMak	2026-02-08 13:22:03+00	f
b50c4531-4f58-4e4e-9256-cda7f5270875	7	\N	user	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbklEIjoiYjUwYzQ1MzEtNGY1OC00ZTRlLTkyNTYtY2RhN2Y1MjcwODc1IiwiVXNlcklEIjoiNyIsIkV4cGlyZXNBdCI6IjIwMjYtMDItMDhUMTg6NTM6NDAuOTQ2MDMzMyswNTozMCIsIlVzZWRGb3IiOiJ1c2VyIn0.borYR32du5hgND1aoi1AVQdjuzrKM3jWP7M3oFmAqvU	2026-02-08 13:23:40+00	f
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, age, first_name, last_name, email, phone, password, verified, block_status, created_at, updated_at) FROM stdin;
14	\N	\N	\N	\N	8302827722	\N	f	f	2025-12-21 21:38:19.823836+00	\N
16	\N	\N	\N	string	stringstri	\N	f	f	2025-12-25 07:12:10.288172+00	\N
17	\N	\N	\N	\N	8343434343	\N	f	f	2025-12-25 07:13:24.940676+00	\N
15	\N	\N	\N	\N	9886569962	\N	t	f	2025-12-24 12:13:36.344953+00	\N
\.


--
-- Data for Name: variation_options; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.variation_options (id, variation_id, value, sort_order, is_active) FROM stdin;
\.


--
-- Data for Name: variations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.variations (id, sub_category_id, name, sort_order, is_active) FROM stdin;
\.


--
-- Data for Name: wallets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wallets (id, user_id, total_amount) FROM stdin;
\.


--
-- Data for Name: wish_lists; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wish_lists (id, user_id, shop_id, admin_id, product_item_id) FROM stdin;
\.


--
-- Name: addresses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.addresses_id_seq', 1, false);


--
-- Name: admins_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.admins_id_seq', 35, true);


--
-- Name: advertisements_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.advertisements_id_seq', 1, false);


--
-- Name: banners_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.banners_id_seq', 4, true);


--
-- Name: brands_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.brands_id_seq', 1, false);


--
-- Name: cart_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cart_items_id_seq', 1, false);


--
-- Name: carts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.carts_id_seq', 1, false);


--
-- Name: categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.categories_id_seq', 164, true);


--
-- Name: category_images_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.category_images_id_seq', 135, true);


--
-- Name: countries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.countries_id_seq', 1, false);


--
-- Name: coupon_uses_coupon_uses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.coupon_uses_coupon_uses_id_seq', 1, false);


--
-- Name: coupons_coupon_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.coupons_coupon_id_seq', 1, false);


--
-- Name: departments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.departments_id_seq', 18, true);


--
-- Name: notifications_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notifications_id_seq', 1, false);


--
-- Name: offer_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.offer_categories_id_seq', 1, false);


--
-- Name: offer_products_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.offer_products_id_seq', 38, true);


--
-- Name: offers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.offers_id_seq', 30, true);


--
-- Name: order_lines_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_lines_id_seq', 1, false);


--
-- Name: order_returns_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_returns_id_seq', 1, false);


--
-- Name: order_statuses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_statuses_id_seq', 8, true);


--
-- Name: otp_sessions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.otp_sessions_id_seq', 79, true);


--
-- Name: payment_methods_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.payment_methods_id_seq', 3, true);


--
-- Name: product_images_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_images_id_seq', 1, false);


--
-- Name: product_item_filter_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_item_filter_types_id_seq', 2, true);


--
-- Name: product_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_items_id_seq', 81, true);


--
-- Name: product_views_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_views_id_seq', 43, true);


--
-- Name: products_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.products_id_seq', 46, true);


--
-- Name: promotion_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.promotion_categories_id_seq', 1, true);


--
-- Name: promotions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.promotions_id_seq', 28, true);


--
-- Name: promotions_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.promotions_types_id_seq', 18, true);


--
-- Name: service_providers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.service_providers_id_seq', 1, false);


--
-- Name: shop_details_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.shop_details_id_seq', 23, true);


--
-- Name: shop_offers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.shop_offers_id_seq', 3, true);


--
-- Name: shop_orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.shop_orders_id_seq', 1, false);


--
-- Name: shop_times_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.shop_times_id_seq', 1, true);


--
-- Name: shop_verification_histories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.shop_verification_histories_id_seq', 1, false);


--
-- Name: shop_verifications_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.shop_verifications_id_seq', 43, true);


--
-- Name: sub_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sub_categories_id_seq', 692, true);


--
-- Name: sub_category_details_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sub_category_details_id_seq', 1, false);


--
-- Name: sub_type_attribute_options_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sub_type_attribute_options_id_seq', 15468, true);


--
-- Name: sub_type_attributes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sub_type_attributes_id_seq', 2600, true);


--
-- Name: transactions_transaction_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.transactions_transaction_id_seq', 1, false);


--
-- Name: user_addresses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_addresses_id_seq', 1, false);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 17, true);


--
-- Name: variation_options_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.variation_options_id_seq', 1, false);


--
-- Name: variations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.variations_id_seq', 1, false);


--
-- Name: wallets_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.wallets_id_seq', 1, false);


--
-- Name: wish_lists_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.wish_lists_id_seq', 1, false);


--
-- Name: addresses addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);


--
-- Name: admin_refresh_sessions admin_refresh_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_refresh_sessions
    ADD CONSTRAINT admin_refresh_sessions_pkey PRIMARY KEY (token_id);


--
-- Name: admins admins_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_pkey PRIMARY KEY (id);


--
-- Name: advertisements advertisements_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.advertisements
    ADD CONSTRAINT advertisements_pkey PRIMARY KEY (id);


--
-- Name: banners banners_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.banners
    ADD CONSTRAINT banners_pkey PRIMARY KEY (id);


--
-- Name: brands brands_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_name_key UNIQUE (name);


--
-- Name: brands brands_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (id);


--
-- Name: brands brands_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_slug_key UNIQUE (slug);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: carts carts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: category_images category_images_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category_images
    ADD CONSTRAINT category_images_pkey PRIMARY KEY (id);


--
-- Name: countries countries_country_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_country_name_key UNIQUE (country_name);


--
-- Name: countries countries_iso_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_iso_code_key UNIQUE (iso_code);


--
-- Name: countries countries_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_pkey PRIMARY KEY (id);


--
-- Name: coupon_uses coupon_uses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupon_uses
    ADD CONSTRAINT coupon_uses_pkey PRIMARY KEY (coupon_uses_id);


--
-- Name: coupons coupons_coupon_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupons
    ADD CONSTRAINT coupons_coupon_code_key UNIQUE (coupon_code);


--
-- Name: coupons coupons_coupon_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupons
    ADD CONSTRAINT coupons_coupon_name_key UNIQUE (coupon_name);


--
-- Name: coupons coupons_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupons
    ADD CONSTRAINT coupons_pkey PRIMARY KEY (coupon_id);


--
-- Name: departments departments_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_name_key UNIQUE (name);


--
-- Name: departments departments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_pkey PRIMARY KEY (id);


--
-- Name: departments departments_slug_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_slug_key UNIQUE (slug);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: offer_categories offer_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offer_categories
    ADD CONSTRAINT offer_categories_pkey PRIMARY KEY (id);


--
-- Name: offer_products offer_products_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offer_products
    ADD CONSTRAINT offer_products_pkey PRIMARY KEY (id);


--
-- Name: offers offers_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offers
    ADD CONSTRAINT offers_name_key UNIQUE (name);


--
-- Name: offers offers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offers
    ADD CONSTRAINT offers_pkey PRIMARY KEY (id);


--
-- Name: order_lines order_lines_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_lines
    ADD CONSTRAINT order_lines_pkey PRIMARY KEY (id);


--
-- Name: order_returns order_returns_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_returns
    ADD CONSTRAINT order_returns_pkey PRIMARY KEY (id);


--
-- Name: order_returns order_returns_shop_order_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_returns
    ADD CONSTRAINT order_returns_shop_order_id_key UNIQUE (shop_order_id);


--
-- Name: order_statuses order_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_statuses
    ADD CONSTRAINT order_statuses_pkey PRIMARY KEY (id);


--
-- Name: order_statuses order_statuses_status_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_statuses
    ADD CONSTRAINT order_statuses_status_key UNIQUE (status);


--
-- Name: otp_sessions otp_sessions_otp_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.otp_sessions
    ADD CONSTRAINT otp_sessions_otp_id_key UNIQUE (otp_id);


--
-- Name: otp_sessions otp_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.otp_sessions
    ADD CONSTRAINT otp_sessions_pkey PRIMARY KEY (id);


--
-- Name: payment_methods payment_methods_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT payment_methods_name_key UNIQUE (name);


--
-- Name: payment_methods payment_methods_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payment_methods
    ADD CONSTRAINT payment_methods_pkey PRIMARY KEY (id);


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


--
-- Name: product_item_filter_types product_item_filter_types_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_item_filter_types
    ADD CONSTRAINT product_item_filter_types_pkey PRIMARY KEY (id);


--
-- Name: product_items product_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_items
    ADD CONSTRAINT product_items_pkey PRIMARY KEY (id);


--
-- Name: product_item_views product_views_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_item_views
    ADD CONSTRAINT product_views_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: promotion_categories promotion_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotion_categories
    ADD CONSTRAINT promotion_categories_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: promotions_types promotions_types_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions_types
    ADD CONSTRAINT promotions_types_pkey PRIMARY KEY (id);


--
-- Name: service_providers service_providers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.service_providers
    ADD CONSTRAINT service_providers_pkey PRIMARY KEY (id);


--
-- Name: shop_details shop_details_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_details
    ADD CONSTRAINT shop_details_pkey PRIMARY KEY (id);


--
-- Name: shop_offers shop_offers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_offers
    ADD CONSTRAINT shop_offers_pkey PRIMARY KEY (id);


--
-- Name: shop_orders shop_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_orders
    ADD CONSTRAINT shop_orders_pkey PRIMARY KEY (id);


--
-- Name: shop_times shop_times_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_times
    ADD CONSTRAINT shop_times_pkey PRIMARY KEY (id);


--
-- Name: shop_verification_histories shop_verification_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_verification_histories
    ADD CONSTRAINT shop_verification_histories_pkey PRIMARY KEY (id);


--
-- Name: shop_verifications shop_verifications_admin_id_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_verifications
    ADD CONSTRAINT shop_verifications_admin_id_unique UNIQUE (admin_id);


--
-- Name: shop_verifications shop_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_verifications
    ADD CONSTRAINT shop_verifications_pkey PRIMARY KEY (id);


--
-- Name: sub_categories sub_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_categories
    ADD CONSTRAINT sub_categories_pkey PRIMARY KEY (id);


--
-- Name: sub_category_details sub_category_details_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_category_details
    ADD CONSTRAINT sub_category_details_pkey PRIMARY KEY (id);


--
-- Name: sub_type_attribute_options sub_type_attribute_options_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_type_attribute_options
    ADD CONSTRAINT sub_type_attribute_options_pkey PRIMARY KEY (id);


--
-- Name: sub_type_attributes sub_type_attributes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_type_attributes
    ADD CONSTRAINT sub_type_attributes_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (transaction_id);


--
-- Name: category_images unique_category_image; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category_images
    ADD CONSTRAINT unique_category_image UNIQUE (category_id, image_url);


--
-- Name: user_addresses user_addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT user_addresses_pkey PRIMARY KEY (id);


--
-- Name: user_refresh_sessions user_refresh_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_refresh_sessions
    ADD CONSTRAINT user_refresh_sessions_pkey PRIMARY KEY (token_id);


--
-- Name: users users_phone_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_key UNIQUE (phone);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: variation_options variation_options_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.variation_options
    ADD CONSTRAINT variation_options_pkey PRIMARY KEY (id);


--
-- Name: variations variations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.variations
    ADD CONSTRAINT variations_pkey PRIMARY KEY (id);


--
-- Name: wallets wallets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT wallets_pkey PRIMARY KEY (id);


--
-- Name: wish_lists wish_lists_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wish_lists
    ADD CONSTRAINT wish_lists_pkey PRIMARY KEY (id);


--
-- Name: idx_admins_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_admins_search ON public.admins USING gin (to_tsvector('english'::regconfig, ((((full_name || ' '::text) || email) || ' '::text) || (mobile)::text)));


--
-- Name: idx_brands_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_brands_search ON public.brands USING gin (to_tsvector('english'::regconfig, name));


--
-- Name: idx_categories_dept_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_categories_dept_id ON public.categories USING btree (department_id);


--
-- Name: idx_categories_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_categories_search ON public.categories USING gin (to_tsvector('english'::regconfig, name));


--
-- Name: idx_categories_sort_order; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_categories_sort_order ON public.categories USING btree (sort_order);


--
-- Name: idx_category_images_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_category_images_active ON public.category_images USING btree (is_active);


--
-- Name: idx_category_images_category_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_category_images_category_id ON public.category_images USING btree (category_id);


--
-- Name: idx_departments_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_departments_search ON public.departments USING gin (to_tsvector('english'::regconfig, name));


--
-- Name: idx_offers_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_offers_search ON public.offers USING gin (to_tsvector('english'::regconfig, ((name || ' '::text) || description)));


--
-- Name: idx_product_items_cat_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_product_items_cat_id ON public.product_items USING btree (category_id);


--
-- Name: idx_product_items_dept_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_product_items_dept_id ON public.product_items USING btree (department_id);


--
-- Name: idx_product_items_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_product_items_search ON public.product_items USING gin (to_tsvector('english'::regconfig, sub_category_name));


--
-- Name: idx_product_items_sub_cat_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_product_items_sub_cat_id ON public.product_items USING btree (sub_category_id);


--
-- Name: idx_products_description_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_description_trgm ON public.products USING gin (description public.gin_trgm_ops);


--
-- Name: idx_products_name_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_name_trgm ON public.products USING gin (name public.gin_trgm_ops);


--
-- Name: idx_products_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_search ON public.products USING gin (to_tsvector('english'::regconfig, ((name || ' '::text) || description)));


--
-- Name: idx_service_providers_categories; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_service_providers_categories ON public.service_providers USING gin (categories);


--
-- Name: idx_service_providers_location; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_service_providers_location ON public.service_providers USING btree (service_radius_km);


--
-- Name: idx_service_providers_pincodes; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_service_providers_pincodes ON public.service_providers USING gin (serviceable_pincodes);


--
-- Name: idx_service_providers_rating; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_service_providers_rating ON public.service_providers USING btree (rating DESC);


--
-- Name: idx_shop_details_admin_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_shop_details_admin_id ON public.shop_details USING btree (admin_id);


--
-- Name: idx_shop_details_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_shop_details_email ON public.shop_details USING btree (email);


--
-- Name: idx_sub_categories_cat_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sub_categories_cat_id ON public.sub_categories USING btree (category_id);


--
-- Name: idx_sub_categories_dept_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sub_categories_dept_id ON public.sub_categories USING btree (department_id);


--
-- Name: idx_sub_categories_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sub_categories_search ON public.sub_categories USING gin (to_tsvector('english'::regconfig, name));


--
-- Name: idx_sub_type_attribute_options_attr_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sub_type_attribute_options_attr_id ON public.sub_type_attribute_options USING btree (sub_type_attribute_id);


--
-- Name: idx_sub_type_attribute_options_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sub_type_attribute_options_search ON public.sub_type_attribute_options USING gin (to_tsvector('english'::regconfig, (option_value)::text));


--
-- Name: idx_sub_type_attributes_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sub_type_attributes_search ON public.sub_type_attributes USING gin (to_tsvector('english'::regconfig, (field_name)::text));


--
-- Name: idx_sub_type_attributes_sub_cat_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sub_type_attributes_sub_cat_id ON public.sub_type_attributes USING btree (sub_category_id);


--
-- Name: idx_users_search; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_search ON public.users USING gin (to_tsvector('english'::regconfig, ((((((first_name || ' '::text) || last_name) || ' '::text) || email) || ' '::text) || phone)));


--
-- Name: cart_items update_cart_total_price; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_cart_total_price AFTER INSERT OR DELETE OR UPDATE ON public.cart_items FOR EACH ROW EXECUTE FUNCTION public.update_cart_total_price();


--
-- Name: shop_orders update_product_qty_on_order_return; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_product_qty_on_order_return AFTER UPDATE OF order_status_id ON public.shop_orders FOR EACH ROW WHEN ((new.order_status_id = public.get_order_status_id('order returned'::text))) EXECUTE FUNCTION public.update_product_quantity_on_return();


--
-- Name: order_lines update_product_quantity; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_product_quantity AFTER INSERT ON public.order_lines FOR EACH ROW EXECUTE FUNCTION public.update_product_quantity();


--
-- Name: addresses fk_addresses_country; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.addresses
    ADD CONSTRAINT fk_addresses_country FOREIGN KEY (country_id) REFERENCES public.countries(id);


--
-- Name: cart_items fk_cart_items_cart; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT fk_cart_items_cart FOREIGN KEY (cart_id) REFERENCES public.carts(id);


--
-- Name: cart_items fk_cart_items_product_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT fk_cart_items_product_item FOREIGN KEY (product_item_id) REFERENCES public.product_items(id);


--
-- Name: category_images fk_category_images_category; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.category_images
    ADD CONSTRAINT fk_category_images_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE CASCADE;


--
-- Name: coupon_uses fk_coupon_uses_coupon; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupon_uses
    ADD CONSTRAINT fk_coupon_uses_coupon FOREIGN KEY (coupon_id) REFERENCES public.coupons(coupon_id);


--
-- Name: coupon_uses fk_coupon_uses_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.coupon_uses
    ADD CONSTRAINT fk_coupon_uses_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: offer_categories fk_offer_categories_category; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offer_categories
    ADD CONSTRAINT fk_offer_categories_category FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: offer_categories fk_offer_categories_offer; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.offer_categories
    ADD CONSTRAINT fk_offer_categories_offer FOREIGN KEY (offer_id) REFERENCES public.offers(id);


--
-- Name: order_lines fk_order_lines_shop_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_lines
    ADD CONSTRAINT fk_order_lines_shop_order FOREIGN KEY (shop_order_id) REFERENCES public.shop_orders(id);


--
-- Name: order_returns fk_order_returns_shop_order; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_returns
    ADD CONSTRAINT fk_order_returns_shop_order FOREIGN KEY (shop_order_id) REFERENCES public.shop_orders(id);


--
-- Name: product_configurations fk_product_configurations_product_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_configurations
    ADD CONSTRAINT fk_product_configurations_product_item FOREIGN KEY (product_item_id) REFERENCES public.product_items(id);


--
-- Name: product_configurations fk_product_configurations_variation_option; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_configurations
    ADD CONSTRAINT fk_product_configurations_variation_option FOREIGN KEY (variation_option_id) REFERENCES public.variation_options(id);


--
-- Name: product_images fk_product_images_product_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT fk_product_images_product_item FOREIGN KEY (product_item_id) REFERENCES public.product_items(id);


--
-- Name: products fk_products_category; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: products fk_products_department; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT fk_products_department FOREIGN KEY (department_id) REFERENCES public.departments(id);


--
-- Name: promotions fk_promotions_promotion_category; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_promotion_category FOREIGN KEY (promotion_category_id) REFERENCES public.promotion_categories(id);


--
-- Name: promotions fk_promotions_promotion_type; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT fk_promotions_promotion_type FOREIGN KEY (promotion_type_id) REFERENCES public.promotions_types(id);


--
-- Name: shop_orders fk_shop_orders_address; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_orders
    ADD CONSTRAINT fk_shop_orders_address FOREIGN KEY (address_id) REFERENCES public.addresses(id);


--
-- Name: shop_orders fk_shop_orders_order_status; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_orders
    ADD CONSTRAINT fk_shop_orders_order_status FOREIGN KEY (order_status_id) REFERENCES public.order_statuses(id);


--
-- Name: shop_orders fk_shop_orders_payment_method; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_orders
    ADD CONSTRAINT fk_shop_orders_payment_method FOREIGN KEY (payment_method_id) REFERENCES public.payment_methods(id);


--
-- Name: shop_orders fk_shop_orders_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.shop_orders
    ADD CONSTRAINT fk_shop_orders_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: sub_type_attribute_options fk_sub_type_attribute_options_sub_type_attribute; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sub_type_attribute_options
    ADD CONSTRAINT fk_sub_type_attribute_options_sub_type_attribute FOREIGN KEY (sub_type_attribute_id) REFERENCES public.sub_type_attributes(id);


--
-- Name: transactions fk_transactions_wallet; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT fk_transactions_wallet FOREIGN KEY (wallet_id) REFERENCES public.wallets(id);


--
-- Name: user_addresses fk_user_addresses_address; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_addresses
    ADD CONSTRAINT fk_user_addresses_address FOREIGN KEY (address_id) REFERENCES public.addresses(id);


--
-- Name: variation_options fk_variation_options_variation; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.variation_options
    ADD CONSTRAINT fk_variation_options_variation FOREIGN KEY (variation_id) REFERENCES public.variations(id);


--
-- Name: variations fk_variations_sub_category; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.variations
    ADD CONSTRAINT fk_variations_sub_category FOREIGN KEY (sub_category_id) REFERENCES public.categories(id);


--
-- Name: wallets fk_wallets_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wallets
    ADD CONSTRAINT fk_wallets_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: wish_lists fk_wish_lists_product_item; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wish_lists
    ADD CONSTRAINT fk_wish_lists_product_item FOREIGN KEY (product_item_id) REFERENCES public.product_items(id);


--
-- Name: wish_lists fk_wish_lists_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wish_lists
    ADD CONSTRAINT fk_wish_lists_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

\unrestrict Hy7yAtgLy2F98prSdIipTvt5fxndGVuRSEfeHXzEfJn6AtvDC1XlWfp7inP8IS7


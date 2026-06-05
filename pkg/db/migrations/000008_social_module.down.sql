-- 000002_social_module.down.sql

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS product_rating_summary CASCADE;
DROP TABLE IF EXISTS shop_rating_summary    CASCADE;
DROP TABLE IF EXISTS feedbacks              CASCADE;
DROP TABLE IF EXISTS review_reactions       CASCADE;
DROP TABLE IF EXISTS review_comments        CASCADE;
DROP TABLE IF EXISTS review_images          CASCADE;
DROP TABLE IF EXISTS product_reviews        CASCADE;
DROP TABLE IF EXISTS shop_reviews           CASCADE;

-- Drop ENUM types
DROP TYPE IF EXISTS feedback_status CASCADE;
DROP TYPE IF EXISTS review_status   CASCADE;
DROP TYPE IF EXISTS reaction_type   CASCADE;
DROP TYPE IF EXISTS feedback_type   CASCADE;
DROP TYPE IF EXISTS review_type     CASCADE;
